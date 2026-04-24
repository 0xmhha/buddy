// events.go produces `buddy events`: a tail of the hook_events table for
// debugging, optionally followed in real time. Output is structured (one line
// per event, space-separated columns) — this surface is debug-style, not
// friend-tone (per v0.1-spec §6.3 and m4-plan Task 4). The only friend-tone
// touches are the start/end markers Follow writes to stderr.
//
// Like stats.go, every function here is pure: open DB read-only, query, shape
// rows. The CLI layer (cmd/buddy) owns process exit codes and IO routing.

package queries

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/format"
)

// defaultEventsLimit is the row cap when EventsOptions.Limit is unset (0).
// Matches `tail`'s habit of printing the last N lines so the user gets useful
// output from a bare `buddy events` invocation.
const defaultEventsLimit = 20

// defaultPollInterval is how often Follow re-queries for new events when
// EventsOptions.PollInterval is unset. 1s matches the daemon poll cadence so
// follow lag stays predictable relative to ingestion.
const defaultPollInterval = time.Second

// EventsOptions configure RunEvents and Follow.
type EventsOptions struct {
	// DBPath is the absolute path to the buddy SQLite store. Empty means
	// db.DefaultPath() (i.e. ~/.buddy/buddy.db).
	DBPath string
	// HookFilter, when non-empty, restricts results to a single hook_name,
	// matched case-insensitively (same convention as stats.HookFilter).
	HookFilter string
	// Limit caps the row count of the initial fetch. Zero means
	// defaultEventsLimit (20).
	Limit int
	// Since, when > 0, restricts results to events with ts strictly greater
	// than this unix-ms value. Used by Follow to fetch only new events on
	// each poll tick.
	Since int64
	// PollInterval is the cadence at which Follow re-queries. Zero means
	// defaultPollInterval (1s). Exposed for fast-test paths.
	PollInterval time.Duration
	// Stderr is where Follow writes friend-tone start/end markers. Zero
	// means os.Stderr. Exposed so tests can capture markers without
	// touching the global stderr.
	Stderr io.Writer
}

// Event is one row in the events output. Empty SessionID/ToolName render as
// "-" placeholders so columns line up regardless of which fields a hook event
// carried.
type Event struct {
	ID         int64
	Ts         int64 // unix ms
	Event      string
	HookName   string
	DurationMs int64
	ExitCode   int
	SessionID  string
	ToolName   string
}

// EventsResult is the full snapshot ready for RenderLines. Events are in
// chronological order (oldest first) so the output reads top-to-bottom like
// `tail` — the SQL query orders DESC for the LIMIT then we reverse.
type EventsResult struct {
	Events []Event
}

// RunEvents validates options, opens the DB read-only, runs the tail query,
// and returns rows in chronological order.
func RunEvents(opts EventsOptions) (EventsResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultEventsLimit
	}

	conn, err := db.Open(db.Options{Path: opts.DBPath, ReadOnly: true})
	if err != nil {
		return EventsResult{}, fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	events, err := queryEvents(conn, opts.HookFilter, opts.Since, limit)
	if err != nil {
		return EventsResult{}, fmt.Errorf("query hook_events: %w", err)
	}
	return EventsResult{Events: events}, nil
}

// queryEvents builds and runs the tail SELECT. We do all filtering in SQL so
// LIMIT is meaningful: a Go-side filter after a LIMIT would silently drop
// matching rows when the limit is hit by non-matching ones first.
//
// ORDER BY ts DESC + LIMIT + reverse-in-Go gives us "the most recent N events
// in chronological order." Reversing 20 rows in Go is cheaper than a subquery.
func queryEvents(conn *sql.DB, hookFilter string, since int64, limit int) ([]Event, error) {
	var (
		clauses []string
		args    []any
	)
	if hookFilter != "" {
		clauses = append(clauses, "LOWER(hook_name) = LOWER(?)")
		args = append(args, hookFilter)
	}
	if since > 0 {
		clauses = append(clauses, "ts > ?")
		args = append(args, since)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	q := `
		SELECT id, ts, event, hook_name, duration_ms, exit_code,
		       COALESCE(session_id, '') AS session_id,
		       COALESCE(tool_name, '')  AS tool_name
		  FROM hook_events
		  ` + where + `
		 ORDER BY ts DESC
		 LIMIT ?`
	args = append(args, limit)

	rs, err := conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	out := make([]Event, 0, limit)
	for rs.Next() {
		var e Event
		if err := rs.Scan(&e.ID, &e.Ts, &e.Event, &e.HookName,
			&e.DurationMs, &e.ExitCode, &e.SessionID, &e.ToolName); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	// Reverse to chronological order (oldest first) so the tail reads
	// top-to-bottom oldest-to-newest, matching the default `tail` behavior.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// RenderLines writes one structured line per event to w. No header, no
// friend-tone — this is a debug surface. Columns are separated by 2 spaces
// (not tabs) so the output stays grep-friendly. Empty SessionID/ToolName
// render as "-" so columns are visually present in every row.
//
// Format:
//
//	<RFC3339 ts UTC>  <hook>  <event>  <tool>  exit=<n>  dur=<human>  sess=<short>
func (r EventsResult) RenderLines(w io.Writer) {
	for _, e := range r.Events {
		ts := time.UnixMilli(e.Ts).UTC().Format(time.RFC3339)
		tool := e.ToolName
		if tool == "" {
			tool = "-"
		}
		sess := shortSession(e.SessionID)
		_, _ = fmt.Fprintf(w, "%s  %s  %s  %s  exit=%d  dur=%s  sess=%s\n",
			ts, e.HookName, e.Event, tool, e.ExitCode,
			format.Duration(e.DurationMs), sess)
	}
}

// shortSession returns the first 8 chars of id, or "-" when id is empty.
// 8 chars is enough to disambiguate sessions in a tail without dominating the
// line; matches the convention used by `git log --abbrev=8`.
func shortSession(id string) string {
	if id == "" {
		return "-"
	}
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// Follow streams events to w until ctx is cancelled. On entry it writes a
// friend-tone start marker to opts.Stderr, fetches the initial tail (same as
// RunEvents), then polls every PollInterval for events newer than the last
// seen ts. On ctx cancel it writes a friend-tone end marker and returns nil.
//
// Polling pattern (vs sqlite update_hook or fts5 triggers): the daemon writes
// a few rows per second at most, and the user's terminal can only consume that
// fast anyway. A simple ticker keeps the implementation single-goroutine and
// race-free; ctx.Done() is the only stop signal we need.
func Follow(ctx context.Context, opts EventsOptions, w io.Writer) error {
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	interval := opts.PollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}

	_, _ = fmt.Fprintln(stderr, "(따라가는 중. Ctrl-C로 멈춰.)")

	// Initial fetch reuses RunEvents so the entry behaviour matches a bare
	// `buddy events` invocation exactly — no drift between one-shot and follow.
	initial, err := RunEvents(opts)
	if err != nil {
		return err
	}
	initial.RenderLines(w)

	lastTs := opts.Since
	for _, e := range initial.Events {
		if e.Ts > lastTs {
			lastTs = e.Ts
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			_, _ = fmt.Fprintln(stderr, "(끝.)")
			return nil
		case <-ticker.C:
			pollLimit := opts.Limit
			if pollLimit <= 0 {
				pollLimit = defaultEventsLimit
			}
			next, err := RunEvents(EventsOptions{
				DBPath:     opts.DBPath,
				HookFilter: opts.HookFilter,
				// Limit on the polling fetch matches the initial limit so a
				// sudden burst doesn't get truncated to a smaller cap; the
				// Since filter keeps duplicates out across ticks.
				Limit: pollLimit,
				Since: lastTs,
			})
			if err != nil {
				// Surface DB blips to stderr but keep following: a daemon
				// restart or transient read error shouldn't kill the tail.
				_, _ = fmt.Fprintf(stderr, "buddy: events poll 실패 (%v)\n", err)
				continue
			}
			next.RenderLines(w)
			for _, e := range next.Events {
				if e.Ts > lastTs {
					lastTs = e.Ts
				}
			}
		}
	}
}
