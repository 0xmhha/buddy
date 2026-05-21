// Package usage derives AI-usage analytic primitives from the sessions
// table. ADR-013 (W7-2 F2.B Usage Analysis) — sessions-only data source,
// live SQL aggregation, no derived table. Consumed by:
//
//   - `buddy usage` CLI subcommand (cmd/buddy/usage_cmd.go)
//   - 5 `usage_query_*` MCP tools (internal/mcp/usage_tool.go)
//   - TUI Usage pane (internal/tui/model.go)
//
// Why a separate package and not `internal/sessions/`: sessions owns
// CRUD on the rows; usage owns *analytic projection* of those rows.
// Splitting keeps store.go from absorbing aggregate SQL that has
// nothing to do with single-row lifecycle.
package usage

import "time"

// TokenSpend is a token-cost projection over a TimeWindow.
// All counts are summed across every session whose StartedAt falls in
// the window (active + ended). InputTokens + OutputTokens + Cache* mirror
// schema.TokenUsage but aggregated.
type TokenSpend struct {
	Window            TimeWindow
	InputTokens       int64
	OutputTokens      int64
	CacheReadTokens   int64
	CacheCreateTokens int64
}

// TotalTokens returns the sum of all four token categories — used by
// the CLI / TUI as the headline "today's spend" number.
func (t TokenSpend) TotalTokens() int64 {
	return t.InputTokens + t.OutputTokens + t.CacheReadTokens + t.CacheCreateTokens
}

// CacheHitRatio returns cache_read / (input + cache_read + cache_create)
// as a fraction 0.0..1.0. Returns 0 when denominator is zero.
// Output tokens are excluded — they're not cacheable input.
func (t TokenSpend) CacheHitRatio() float64 {
	denom := t.InputTokens + t.CacheReadTokens + t.CacheCreateTokens
	if denom == 0 {
		return 0
	}
	return float64(t.CacheReadTokens) / float64(denom)
}

// SessionStats summarises per-session counters in a window: how many
// sessions started, how many are still active, durations distribution.
type SessionStats struct {
	Window         TimeWindow
	TotalSessions  int64         // count(*) where started_at in window
	ActiveSessions int64         // count(*) where ended_at IS NULL
	EndedSessions  int64         // count(*) where ended_at IS NOT NULL
	DurationP50    time.Duration // ended only — uses (last_active - started_at)
	DurationP90    time.Duration
	DurationMax    time.Duration
	GoalTextRatio  float64 // fraction of sessions with non-empty goal_text
}

// TimeDistribution buckets sessions by the hour-of-day of StartedAt
// (local time of the buddy host) over a window. Used to surface "peak
// hours of usage". HourCounts[h] is the count whose StartedAt local
// hour is h (0..23). Total is sum(HourCounts).
type TimeDistribution struct {
	Window     TimeWindow
	HourCounts [24]int64
	Total      int64
}

// TopSession is one row of `usage_query_top_sessions`. Sorted by
// TotalTokens DESC. GoalText may be empty when the original first
// user message was a slash command (fsLister returns "" then).
type TopSession struct {
	ID          string
	StartedAt   time.Time
	GoalText    string
	TotalTokens int64
}

// Overview is the one-shot snapshot CLI / TUI render as the headline
// page. All fields share the same Window. Top is up to N entries.
type Overview struct {
	Window           TimeWindow
	Spend            TokenSpend
	Stats            SessionStats
	TimeDistribution TimeDistribution
	Top              []TopSession
}

// DailySpend is the per-day token breakdown over a multi-day window.
// Date is the UTC start-of-day timestamp (00:00:00). Zero-token days
// are included so callers rendering a chart see every day in the
// requested range without gap-filling at the call site.
type DailySpend struct {
	Date              time.Time
	InputTokens       int64
	OutputTokens      int64
	CacheReadTokens   int64
	CacheCreateTokens int64
}

// TotalTokens mirrors TokenSpend.TotalTokens for a single day.
func (d DailySpend) TotalTokens() int64 {
	return d.InputTokens + d.OutputTokens + d.CacheReadTokens + d.CacheCreateTokens
}

// TimeWindow scopes every query. Open-ended When-To means "now"
// (queried at call time). Implementations clamp Since to the earliest
// observed session start when caller passes a zero value, so callers
// can ask "all-time" by leaving Since = time.Time{}.
type TimeWindow struct {
	Since time.Time // inclusive lower bound on started_at
	Until time.Time // exclusive upper bound on started_at; zero = now
}

// IsAllTime reports whether the window is unbounded (no Since).
// CLI uses this for header rendering ("all time" vs "last 7d").
func (w TimeWindow) IsAllTime() bool {
	return w.Since.IsZero()
}

// Duration returns Until - Since. Zero when Since is zero (all-time)
// or when Until.Before(Since).
func (w TimeWindow) Duration() time.Duration {
	if w.Since.IsZero() || w.Until.IsZero() {
		return 0
	}
	if w.Until.Before(w.Since) {
		return 0
	}
	return w.Until.Sub(w.Since)
}
