package usage

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Service derives metric primitives from the sessions table.
// Stateless — every method reads the live rows on call. Per ADR-013 Q2
// (live aggregation, no derived table) the row count is small enough
// (~1500 rows / 30d at the user's profile) that SQL aggregate is ms-fast.
//
// Concurrency: safe for parallel callers; each method opens its own
// QueryContext.
type Service struct {
	db *sql.DB
	// Now is injectable for tests. Production callers leave nil and the
	// service uses time.Now().UTC(). All window math is UTC.
	Now func() time.Time
}

// NewService binds a *sql.DB. The same pool used by sessions.Store /
// agent.Store / feature.Store works — usage is purely read-only.
func NewService(conn *sql.DB) *Service {
	return &Service{db: conn}
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// resolveWindow fills Until with now() when zero and returns the
// resolved window. Since is left alone — zero means "all time".
func (s *Service) resolveWindow(w TimeWindow) TimeWindow {
	if w.Until.IsZero() {
		w.Until = s.now()
	}
	return w
}

// QueryTokenSpend sums token columns over the window.
func (s *Service) QueryTokenSpend(ctx context.Context, w TimeWindow) (TokenSpend, error) {
	w = s.resolveWindow(w)
	q, args := s.windowQuery(`
		SELECT
			COALESCE(SUM(total_input_tokens),  0),
			COALESCE(SUM(total_output_tokens), 0),
			COALESCE(SUM(total_cache_read),    0),
			COALESCE(SUM(total_cache_create),  0)
		FROM sessions`, w)
	var out TokenSpend
	out.Window = w
	err := s.db.QueryRowContext(ctx, q, args...).Scan(
		&out.InputTokens, &out.OutputTokens,
		&out.CacheReadTokens, &out.CacheCreateTokens,
	)
	if err != nil {
		return TokenSpend{}, fmt.Errorf("usage: query token spend: %w", err)
	}
	return out, nil
}

// QuerySessionStats returns count + duration percentiles + goal-text
// coverage. Duration percentiles use (last_active - started_at) for
// ENDED sessions only — active sessions are still running.
func (s *Service) QuerySessionStats(ctx context.Context, w TimeWindow) (SessionStats, error) {
	w = s.resolveWindow(w)

	out := SessionStats{Window: w}

	// Counts + goal-text ratio in one row — cheaper than three queries.
	// COALESCE wraps every SUM so an empty result set returns 0 rather
	// than NULL (which Go's Scan can't decode into int64).
	countQ, args := s.windowQuery(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN ended_at IS NULL     THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN ended_at IS NOT NULL THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN goal_text != ''      THEN 1 ELSE 0 END), 0)
		FROM sessions`, w)
	var goalCount int64
	err := s.db.QueryRowContext(ctx, countQ, args...).Scan(
		&out.TotalSessions, &out.ActiveSessions, &out.EndedSessions, &goalCount,
	)
	if err != nil {
		return SessionStats{}, fmt.Errorf("usage: query session counts: %w", err)
	}
	if out.TotalSessions > 0 {
		out.GoalTextRatio = float64(goalCount) / float64(out.TotalSessions)
	}

	// Durations — ENDED sessions only. SQLite has no native percentile,
	// so we pull the duration column and compute in Go. With ~30d data
	// the row count is ~1500, well within in-process aggregation budget.
	durQ, durArgs := s.windowQuery(`
		SELECT (last_active - started_at)
		FROM sessions
		WHERE ended_at IS NOT NULL`, w)
	rows, err := s.db.QueryContext(ctx, durQ, durArgs...)
	if err != nil {
		return SessionStats{}, fmt.Errorf("usage: query durations: %w", err)
	}
	defer rows.Close()
	var durs []int64 // milliseconds
	for rows.Next() {
		var d int64
		if err := rows.Scan(&d); err != nil {
			return SessionStats{}, err
		}
		if d > 0 {
			durs = append(durs, d)
		}
	}
	if err := rows.Err(); err != nil {
		return SessionStats{}, err
	}
	out.DurationP50 = percentileDuration(durs, 0.50)
	out.DurationP90 = percentileDuration(durs, 0.90)
	out.DurationMax = maxDuration(durs)

	return out, nil
}

// QueryTimeDistribution buckets sessions by the *local* hour of
// StartedAt. local = host time zone (time.Local). Per ADR-013 Q4 the
// metric exists to surface "peak hours" so user-local is what they
// expect, not UTC.
func (s *Service) QueryTimeDistribution(ctx context.Context, w TimeWindow) (TimeDistribution, error) {
	w = s.resolveWindow(w)
	q, args := s.windowQuery(`SELECT started_at FROM sessions`, w)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return TimeDistribution{}, fmt.Errorf("usage: query time distribution: %w", err)
	}
	defer rows.Close()

	out := TimeDistribution{Window: w}
	for rows.Next() {
		var ms int64
		if err := rows.Scan(&ms); err != nil {
			return TimeDistribution{}, err
		}
		hour := time.UnixMilli(ms).Local().Hour()
		out.HourCounts[hour]++
		out.Total++
	}
	if err := rows.Err(); err != nil {
		return TimeDistribution{}, err
	}
	return out, nil
}

// QueryTopSessions returns up to limit sessions, ordered by total
// tokens DESC. limit ≤ 0 defaults to 10.
func (s *Service) QueryTopSessions(ctx context.Context, w TimeWindow, limit int) ([]TopSession, error) {
	if limit <= 0 {
		limit = 10
	}
	w = s.resolveWindow(w)
	q, args := s.windowQuery(`
		SELECT id, started_at, goal_text,
		       (total_input_tokens + total_output_tokens + total_cache_read + total_cache_create) AS total_tokens
		FROM sessions`, w)
	q += ` ORDER BY total_tokens DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("usage: query top sessions: %w", err)
	}
	defer rows.Close()

	var out []TopSession
	for rows.Next() {
		var t TopSession
		var startedAt int64
		if err := rows.Scan(&t.ID, &startedAt, &t.GoalText, &t.TotalTokens); err != nil {
			return nil, err
		}
		t.StartedAt = time.UnixMilli(startedAt).UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

// QueryOverview composes the four other queries into one snapshot.
// Single-call convenience for the CLI's `buddy usage overview` and the
// TUI Usage pane. Each sub-query runs sequentially against the same
// window — total cost is the sum of four cheap aggregates.
func (s *Service) QueryOverview(ctx context.Context, w TimeWindow, topLimit int) (Overview, error) {
	w = s.resolveWindow(w)
	out := Overview{Window: w}

	spend, err := s.QueryTokenSpend(ctx, w)
	if err != nil {
		return Overview{}, err
	}
	out.Spend = spend

	stats, err := s.QuerySessionStats(ctx, w)
	if err != nil {
		return Overview{}, err
	}
	out.Stats = stats

	td, err := s.QueryTimeDistribution(ctx, w)
	if err != nil {
		return Overview{}, err
	}
	out.TimeDistribution = td

	top, err := s.QueryTopSessions(ctx, w, topLimit)
	if err != nil {
		return Overview{}, err
	}
	out.Top = top

	return out, nil
}

// ─── helpers ─────────────────────────────────────────────────────────

// windowQuery appends a WHERE on started_at based on TimeWindow bounds.
// Returns (query-with-WHERE, args). Zero-value bounds are omitted —
// Since=0 means "all time", Until is always set by resolveWindow.
func (s *Service) windowQuery(base string, w TimeWindow) (string, []any) {
	var (
		conds []string
		args  []any
	)
	if !w.Since.IsZero() {
		conds = append(conds, "started_at >= ?")
		args = append(args, w.Since.UTC().UnixMilli())
	}
	if !w.Until.IsZero() {
		conds = append(conds, "started_at < ?")
		args = append(args, w.Until.UTC().UnixMilli())
	}
	if len(conds) == 0 {
		return base, args
	}
	q := base
	if containsWhere(base) {
		q += " AND "
	} else {
		q += " WHERE "
	}
	q += conds[0]
	for _, c := range conds[1:] {
		q += " AND " + c
	}
	return q, args
}

// containsWhere is a cheap check — sufficient for our hand-built SQL
// where WHERE only appears as a standalone keyword.
func containsWhere(q string) bool {
	for i := 0; i+5 <= len(q); i++ {
		if (q[i] == 'W' || q[i] == 'w') &&
			(q[i+1] == 'H' || q[i+1] == 'h') &&
			(q[i+2] == 'E' || q[i+2] == 'e') &&
			(q[i+3] == 'R' || q[i+3] == 'r') &&
			(q[i+4] == 'E' || q[i+4] == 'e') {
			return true
		}
	}
	return false
}

// percentileDuration returns the p-th percentile of durs (in ms) as a
// time.Duration. p in [0,1]. Empty input returns 0. Linear-interpolation
// not necessary — nearest-rank is plenty for usage metric display.
func percentileDuration(durs []int64, p float64) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	sorted := make([]int64, len(durs))
	copy(sorted, durs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(float64(len(sorted)-1) * p)
	return time.Duration(sorted[idx]) * time.Millisecond
}

func maxDuration(durs []int64) time.Duration {
	var m int64
	for _, d := range durs {
		if d > m {
			m = d
		}
	}
	return time.Duration(m) * time.Millisecond
}
