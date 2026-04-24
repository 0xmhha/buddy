// Package aggregator drains the outbox into hook_events and refreshes
// rolling-window statistics in hook_stats.
//
// Design choices (v0.1):
//   - SELECT-based percentile computation: O(N) per affected (hook,tool,window).
//     Accurate; cheap below ~10k events per window. Streaming quantiles can
//     replace this in v1.0+ if needed.
//   - Three windows: 5min, 1h, 24h. Same buckets every poll regardless of
//     traffic — keeps SQL simple and deterministic.
//   - One transaction per batch: drain N outbox rows + insert events + refresh
//     stats together. A crash mid-batch leaves outbox rows still pending.
package aggregator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

// StatsWindowsMin lists the rolling-window sizes (minutes) we maintain.
var StatsWindowsMin = []int{5, 60, 1440}

// BucketStartMs returns the inclusive start (unix ms) of the window-aligned
// bucket containing tsMs.
func BucketStartMs(tsMs int64, windowMin int) int64 {
	minutes := tsMs / 60_000
	return (minutes / int64(windowMin)) * int64(windowMin) * 60_000
}

// ProcessBatch drains up to batchSize pending outbox rows. For each it
// inserts a hook_events row and tracks which (hookName, toolName) pairs were
// touched; once all rows are inserted, the affected pairs' stats are recomputed.
//
// Returns the number of outbox rows processed (0 if nothing pending).
func ProcessBatch(conn *sql.DB, batchSize int) (int, error) {
	rows, err := db.ReadPendingOutbox(conn, batchSize)
	if err != nil {
		return 0, fmt.Errorf("read outbox: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}

	type touched struct{ hookName, toolName string }
	affected := map[touched]bool{}
	consumed := make([]int64, 0, len(rows))

	tx, err := conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, r := range rows {
		var p schema.HookEventPayload
		if err := json.Unmarshal([]byte(r.Payload), &p); err != nil {
			// Malformed payload — mark consumed so it stops blocking the queue,
			// but skip the event. (We trust upstream to validate; this is
			// a defense-in-depth catch.)
			consumed = append(consumed, r.ID)
			continue
		}
		if err := insertEvent(tx, &p); err != nil {
			return 0, fmt.Errorf("insert event id=%d: %w", r.ID, err)
		}
		affected[touched{p.HookName, p.ToolName}] = true
		consumed = append(consumed, r.ID)
	}

	for t := range affected {
		for _, w := range StatsWindowsMin {
			if err := refreshStatsBucket(tx, t.hookName, t.toolName, w, currentBucketTsMs(rows, w)); err != nil {
				return 0, fmt.Errorf("refresh stats hook=%s tool=%s w=%d: %w", t.hookName, t.toolName, w, err)
			}
		}
	}

	if err := markConsumedTx(tx, consumed); err != nil {
		return 0, fmt.Errorf("mark consumed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return len(rows), nil
}

func currentBucketTsMs(rows []db.OutboxRow, windowMin int) int64 {
	// All rows in the batch share the same "current" bucket in their respective
	// windows ~99% of the time. We refresh based on the latest row's ts so the
	// just-written events are included.
	maxTs := int64(0)
	for _, r := range rows {
		if r.Ts > maxTs {
			maxTs = r.Ts
		}
	}
	return BucketStartMs(maxTs, windowMin)
}

func insertEvent(tx *sql.Tx, p *schema.HookEventPayload) error {
	toolArgs := jsonOrEmpty(p.ToolArgs)
	tokenUsage := jsonOrEmpty(p.TokenUsage)
	customTags := jsonOrEmpty(p.CustomTags)
	meta := jsonOrEmpty(p.Meta)

	_, err := tx.Exec(`
		INSERT INTO hook_events
			(ts, event, hook_name, duration_ms, exit_code,
			 session_id, pid, cwd,
			 tool_name, tool_args, model_name, token_usage, custom_tags, meta)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Ts, string(p.Event), p.HookName, p.DurationMs, p.ExitCode,
		nullableString(p.SessionID), nullableInt(p.PID), nullableString(p.Cwd),
		nullableString(p.ToolName), toolArgs, nullableString(p.ModelName),
		tokenUsage, customTags, meta,
	)
	return err
}

func refreshStatsBucket(tx *sql.Tx, hookName, toolName string, windowMin int, bucketTsMs int64) error {
	endTsMs := bucketTsMs + int64(windowMin)*60_000

	var (
		count, failures int64
		durations       []int64
	)

	q := `SELECT duration_ms, exit_code FROM hook_events
	      WHERE hook_name = ? AND COALESCE(tool_name, '') = ?
	        AND ts >= ? AND ts < ?`
	rows, err := tx.Query(q, hookName, toolName, bucketTsMs, endTsMs)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var dur int64
		var code int
		if err := rows.Scan(&dur, &code); err != nil {
			return err
		}
		count++
		if code != 0 {
			failures++
		}
		durations = append(durations, dur)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	slices.Sort(durations)
	p50 := percentile(durations, 50)
	p95 := percentile(durations, 95)
	p99 := percentile(durations, 99)

	_, err = tx.Exec(`
		INSERT INTO hook_stats
			(hook_name, tool_name, window_min, ts_bucket,
			 count, failures, p50_ms, p95_ms, p99_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (hook_name, tool_name, window_min, ts_bucket)
		DO UPDATE SET
			count = excluded.count,
			failures = excluded.failures,
			p50_ms = excluded.p50_ms,
			p95_ms = excluded.p95_ms,
			p99_ms = excluded.p99_ms`,
		hookName, toolName, windowMin, bucketTsMs,
		count, failures, p50, p95, p99,
	)
	return err
}

func markConsumedTx(tx *sql.Tx, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	stmt, err := tx.Prepare("UPDATE hook_outbox SET consumed_at = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := nowMs()
	for _, id := range ids {
		if _, err := stmt.Exec(now, id); err != nil {
			return err
		}
	}
	return nil
}

func percentile(sortedAsc []int64, p float64) int64 {
	n := len(sortedAsc)
	if n == 0 {
		return 0
	}
	if p <= 0 {
		return sortedAsc[0]
	}
	if p >= 100 {
		return sortedAsc[n-1]
	}
	rank := (p / 100.0) * float64(n-1)
	idx := int(rank)
	frac := rank - float64(idx)
	if idx >= n-1 {
		return sortedAsc[n-1]
	}
	return sortedAsc[idx] + int64(frac*float64(sortedAsc[idx+1]-sortedAsc[idx]))
}

func jsonOrEmpty(v any) any {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return string(b)
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableInt(n int) any {
	if n <= 0 {
		return nil
	}
	return n
}
