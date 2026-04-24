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
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/wm-it-22-00661/buddy/internal/db"
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
	Rows        []Row
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

// filterByHook keeps only rows whose HookName matches name. Linear scan; the
// stats table is small enough that pushing the filter to SQL would not pay off.
func filterByHook(rows []Row, name string) []Row {
	out := rows[:0]
	for _, r := range rows {
		if r.HookName == name {
			out = append(out, r)
		}
	}
	return out
}

// aggregateByHook collapses per-(hook,tool) rows into per-hook rows. count and
// failures sum; p50/p95 take the per-hook max — see Run's note for the rationale.
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
	byTool := r.hasTool()
	if byTool {
		_, _ = fmt.Fprintln(tw, "  hook\ttool\tcount\tp50\tp95\t실패율")
	} else {
		_, _ = fmt.Fprintln(tw, "  hook\tcount\tp50\tp95\t실패율")
	}
	for _, row := range r.Rows {
		failPct := failurePercent(row.Failures, row.Count)
		if byTool {
			_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%d%%\n",
				row.HookName, row.ToolName,
				formatThousands(row.Count), humanDur(row.P50Ms),
				humanDur(row.P95Ms), failPct)
		} else {
			_, _ = fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%d%%\n",
				row.HookName,
				formatThousands(row.Count), humanDur(row.P50Ms),
				humanDur(row.P95Ms), failPct)
		}
	}
	_ = tw.Flush()
}

// hasTool reports whether the result was produced with ByTool=true. We infer
// from the data instead of carrying a flag so the Result struct stays small;
// any non-empty ToolName implies the by-tool layout.
func (r Result) hasTool() bool {
	for _, row := range r.Rows {
		if row.ToolName != "" {
			return true
		}
	}
	return false
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

// humanDur renders ms as a human-friendly duration. Behaviour mirrors
// diagnose.humanDur (kept as a package-local copy to avoid cross-importing
// internal/diagnose, which would couple report packages together).
//
//	[0, 1000)      → "<n>ms"
//	[1000, 60000)  → "<n.n>s"   (rounded to nearest 0.1s)
//	[60000, ∞)     → "<n.n>m"
//
// Bucket selection runs on integer-rounded tenths-of-a-second, so the [s → m]
// boundary cannot straddle units. e.g. 59,999ms → "1.0m", not "60.0s".
func humanDur(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	if ms < 1000 {
		return strconv.FormatInt(ms, 10) + "ms"
	}
	tenthsOfSec := (ms + 50) / 100
	if tenthsOfSec < 600 {
		return strconv.FormatFloat(float64(tenthsOfSec)/10.0, 'f', 1, 64) + "s"
	}
	return strconv.FormatFloat(float64(ms)/60_000.0, 'f', 1, 64) + "m"
}

// formatThousands inserts commas into a non-negative integer. Mirrors the
// helper in internal/diagnose; copied for the same package-independence reason
// as humanDur. v0.1 ships ko/en, both fine with comma separators.
func formatThousands(n int64) string {
	if n < 0 {
		return "-" + formatThousands(-n)
	}
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if len(s) > pre {
			b.WriteByte(',')
		}
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}
