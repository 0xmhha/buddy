// Package queries owns read-only, user-facing reports over the buddy DB.
//
// Each report is a pure function: open DB read-only, run a focused SELECT,
// shape rows into a renderable Result. No os.Exit, no cobra, no globals.
// The CLI layer (cmd/buddy) decides how to print and how to exit.
//
// stats.go produces `buddy stats`: a snapshot of hook_stats aggregated over
// the latest bucket of a chosen window (5m / 1h / 24h), optionally split by
// tool, optionally filtered to a single hook. Output is friend-tone Korean
// with tabwriter alignment.
package queries

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/format"
)

// ErrInvalidWindow is the sentinel returned when Options.Window is not one of
// the accepted values. Carries a friend-tone Korean message that the CLI layer
// surfaces verbatim — see cmd/buddy/main.go's friendError pattern.
var ErrInvalidWindow = errors.New("--window 은 5m, 1h, 24h 중 하나야.")

// Options configure a Run call.
type Options struct {
	// DBPath is the absolute path to the buddy SQLite store. Empty means
	// db.DefaultPath() (i.e. ~/.buddy/buddy.db).
	DBPath string
	// Window is one of "5m", "1h", "24h". Empty defaults to "1h".
	// Anything else returns ErrInvalidWindow at the boundary.
	Window string
	// ByTool, when true, returns one row per (hook, tool) pair. When false,
	// counts/failures sum across tools per hook; p50/p95 take the max across
	// the hook's tools (see Run's comment for the trade-off).
	ByTool bool
	// HookFilter, when non-empty, restricts results to a single hook_name.
	HookFilter string
}

// Row is one rendered line in the stats output. ToolName is "" when the row
// represents a hook-level aggregate (ByTool=false).
type Row struct {
	HookName string
	ToolName string
	Count    int64
	Failures int64
	P50Ms    int64
	P95Ms    int64
}

// Result is the full snapshot ready for Render.
type Result struct {
	WindowMin   int
	WindowLabel string // "5분" | "1시간" | "24시간"
	// ByTool mirrors Options.ByTool so Render can pick the right layout even
	// when the user explicitly asked for per-tool breakdown but the data has
	// no tool names (e.g. Stop hook). Inferring layout from "any non-empty
	// ToolName?" silently downgraded to the no-tool layout in that case;
	// carrying the flag makes the user's intent authoritative.
	ByTool bool
	Rows   []Row
}

// Run validates options, opens the DB read-only, queries the latest hook_stats
// bucket per (hook, tool) for the chosen window, and shapes rows according to
// ByTool / HookFilter.
//
// Cross-tool aggregation note (ByTool=false): count and failures are exact
// sums, but percentiles cannot be combined from per-tool aggregates without
// the underlying samples. For v0.1 we surface MAX(p50) / MAX(p95) across tools
// as a deliberate over-approximation: it surfaces the worst-behaving tool's
// tail, which matches the user's mental model of "is this hook misbehaving?".
// A true cross-tool percentile would require re-querying hook_events; we defer
// that until someone asks for it.
func Run(opts Options) (Result, error) {
	wmin, label, err := resolveWindow(opts.Window)
	if err != nil {
		return Result{}, err
	}

	conn, err := db.Open(db.Options{Path: opts.DBPath, ReadOnly: true})
	if err != nil {
		return Result{}, fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	rows, err := queryLatestBucket(conn, wmin)
	if err != nil {
		return Result{}, fmt.Errorf("query hook_stats: %w", err)
	}

	if opts.HookFilter != "" {
		rows = filterByHook(rows, opts.HookFilter)
	}

	if !opts.ByTool {
		rows = aggregateByHook(rows)
	}

	sortRows(rows, opts.ByTool)

	return Result{
		WindowMin:   wmin,
		WindowLabel: label,
		ByTool:      opts.ByTool,
		Rows:        rows,
	}, nil
}

// resolveWindow maps the user-facing flag value to (windowMin, label). Empty
// defaults to "1h" so a bare `buddy stats` is the most common case.
func resolveWindow(w string) (int, string, error) {
	if w == "" {
		w = "1h"
	}
	switch w {
	case "5m":
		return 5, "5분", nil
	case "1h":
		return 60, "1시간", nil
	case "24h":
		return 1440, "24시간", nil
	}
	return 0, "", ErrInvalidWindow
}

// queryLatestBucket returns one Row per (hook_name, tool_name) for window
// `windowMin`, picking the latest ts_bucket via correlated subquery — same
// pattern as diagnose.slowHookIssues. ToolName comes back as "" when null/empty.
func queryLatestBucket(conn *sql.DB, windowMin int) ([]Row, error) {
	q := `
		SELECT hook_name,
		       COALESCE(tool_name, '') AS tool_name,
		       count, failures,
		       COALESCE(p50_ms, 0) AS p50_ms,
		       COALESCE(p95_ms, 0) AS p95_ms
		  FROM hook_stats h
		 WHERE window_min = ?
		   AND ts_bucket = (
		         SELECT MAX(ts_bucket) FROM hook_stats
		          WHERE hook_name = h.hook_name
		            AND tool_name = h.tool_name
		            AND window_min = ?
		       )`
	rs, err := conn.Query(q, windowMin, windowMin)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	var out []Row
	for rs.Next() {
		var r Row
		if err := rs.Scan(&r.HookName, &r.ToolName, &r.Count, &r.Failures, &r.P50Ms, &r.P95Ms); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// filterByHook keeps only rows whose HookName matches name, case-insensitively.
// Linear scan; the stats table is small enough that pushing the filter to SQL
// would not pay off. Case-insensitive because hook names are user-facing labels
// (e.g. "PreToolUse" vs "pretooluse"); a strict match silently empties the
// result with no signal back to the user that they typo'd capitalisation.
func filterByHook(rows []Row, name string) []Row {
	out := rows[:0]
	for _, r := range rows {
		if strings.EqualFold(r.HookName, name) {
			out = append(out, r)
		}
	}
	return out
}

// aggregateByHook collapses per-(hook,tool) rows into per-hook rows. count and
// failures sum; p50/p95 take the per-hook max — see Run's note for the rationale.
//
// We deliberately drop ToolName when aggregating by hook — even when a hook has
// only one tool, the user's intent (no --by-tool) is "show me hook-level
// totals." Use --by-tool to see per-tool detail. Doctor uses a different display
// (`hookName:toolName`) because its mental model is "alert me on a misbehaving
// pair," not "summarise traffic for this hook."
func aggregateByHook(rows []Row) []Row {
	if len(rows) == 0 {
		return rows
	}
	byHook := map[string]*Row{}
	order := make([]string, 0)
	for _, r := range rows {
		acc, ok := byHook[r.HookName]
		if !ok {
			cp := Row{HookName: r.HookName}
			byHook[r.HookName] = &cp
			acc = byHook[r.HookName]
			order = append(order, r.HookName)
		}
		acc.Count += r.Count
		acc.Failures += r.Failures
		if r.P50Ms > acc.P50Ms {
			acc.P50Ms = r.P50Ms
		}
		if r.P95Ms > acc.P95Ms {
			acc.P95Ms = r.P95Ms
		}
	}
	out := make([]Row, 0, len(order))
	for _, name := range order {
		out = append(out, *byHook[name])
	}
	return out
}

// sortRows enforces a deterministic display order regardless of SQLite version.
func sortRows(rows []Row, byTool bool) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].HookName != rows[j].HookName {
			return rows[i].HookName < rows[j].HookName
		}
		if byTool {
			return rows[i].ToolName < rows[j].ToolName
		}
		return false
	})
}

// Render writes a friend-tone Korean table to w. An empty Result produces a
// one-line nudge; a populated Result renders a header plus a tabwriter-aligned
// table with hook, optional tool, count, p50, p95, failure rate.
func (r Result) Render(w io.Writer) {
	if len(r.Rows) == 0 {
		_, _ = io.WriteString(w, "아직 기록된 hook 통계가 없어. daemon을 띄우고 좀 기다려봐.\n")
		return
	}
	_, _ = fmt.Fprintf(w, "지난 %s hook 통계\n\n", r.WindowLabel)

	// 2-space leading indent matches doctor's bullet-list alignment so the two
	// commands feel visually consistent. tabwriter pads each column to the
	// longest cell; we choose padding=2 (gap between columns) and minwidth=0
	// so single-row outputs don't gain awkward trailing whitespace.
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if r.ByTool {
		_, _ = fmt.Fprintln(tw, "  hook\ttool\tcount\tp50\tp95\t실패율")
	} else {
		_, _ = fmt.Fprintln(tw, "  hook\tcount\tp50\tp95\t실패율")
	}
	for _, row := range r.Rows {
		failPct := failurePercent(row.Failures, row.Count)
		if r.ByTool {
			// Empty ToolName under --by-tool means the underlying hook
			// (e.g. Stop) doesn't carry tool info. Render "(none)" instead
			// of an empty cell so the user sees an explicit "no tool data"
			// signal rather than a silently dropped column.
			tool := row.ToolName
			if tool == "" {
				tool = "(none)"
			}
			_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%d%%\n",
				row.HookName, tool,
				format.Thousands(row.Count), format.Duration(row.P50Ms),
				format.Duration(row.P95Ms), failPct)
		} else {
			_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%d%%\n",
				row.HookName,
				format.Thousands(row.Count), format.Duration(row.P50Ms),
				format.Duration(row.P95Ms), failPct)
		}
	}
	_ = tw.Flush()
}

// failurePercent rounds half-up to the nearest integer. e.g. 0.155 → 16%.
// Using math.Round gives half-away-from-zero, which matches half-up for
// non-negative inputs (count/failures are always non-negative here).
func failurePercent(failures, count int64) int {
	if count <= 0 {
		return 0
	}
	return int(math.Round(float64(failures) / float64(count) * 100))
}
