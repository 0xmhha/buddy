package queries_test

import (
	"bytes"
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/aggregator"
	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/queries"
)

// fixtures -----------------------------------------------------------------

func newTempDBPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	require.NoError(t, conn.Close())
	return path
}

func openWritable(t *testing.T, path string) *sql.DB {
	t.Helper()
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// insertStat inserts one hook_stats row directly. The bucket defaults to the
// current bucket for windowMin so freshness queries hit it.
func insertStat(t *testing.T, conn *sql.DB,
	hookName, toolName string, windowMin int,
	count, failures, p50Ms, p95Ms int64,
) {
	t.Helper()
	bucket := aggregator.BucketStartMs(time.Now().UnixMilli(), windowMin)
	_, err := conn.Exec(`
		INSERT INTO hook_stats
		  (hook_name, tool_name, window_min, ts_bucket,
		   count, failures, p50_ms, p95_ms, p99_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		hookName, toolName, windowMin, bucket,
		count, failures, p50Ms, p95Ms, p95Ms)
	require.NoError(t, err)
}

// tests --------------------------------------------------------------------

func TestRun_DefaultWindowIs1h(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	insertStat(t, conn, "lint", "", 60, 10, 0, 100, 200)

	res, err := queries.Run(queries.Options{DBPath: dbPath, Window: ""})
	require.NoError(t, err)
	assert.Equal(t, 60, res.WindowMin)
	assert.Equal(t, "1시간", res.WindowLabel)
	require.Len(t, res.Rows, 1)
	assert.Equal(t, "lint", res.Rows[0].HookName)
}

func TestRun_InvalidWindow_ReturnsSentinelError(t *testing.T) {
	_, err := queries.Run(queries.Options{DBPath: newTempDBPath(t), Window: "10m"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, queries.ErrInvalidWindow),
		"expected ErrInvalidWindow, got %v", err)
	assert.Contains(t, err.Error(), "5m, 1h, 24h")
}

func TestRun_WindowFiltersToMatchingRowsOnly(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	// Three hooks across the three windows.
	insertStat(t, conn, "five", "", 5, 5, 0, 100, 200)
	insertStat(t, conn, "hour", "", 60, 60, 0, 100, 200)
	insertStat(t, conn, "day", "", 1440, 1440, 0, 100, 200)

	res, err := queries.Run(queries.Options{DBPath: dbPath, Window: "5m"})
	require.NoError(t, err)
	require.Len(t, res.Rows, 1)
	assert.Equal(t, "five", res.Rows[0].HookName)
	assert.Equal(t, 5, res.WindowMin)
	assert.Equal(t, "5분", res.WindowLabel)
}

func TestRun_NoByTool_AggregatesCountAndFailuresAcrossTools_MaxP95(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	// Same hook, two tools — counts/failures sum, p95 is the max across tools.
	insertStat(t, conn, "pre-commit", "Bash", 60, 10, 1, 1000, 5000)
	insertStat(t, conn, "pre-commit", "Read", 60, 90, 0, 50, 100)

	res, err := queries.Run(queries.Options{DBPath: dbPath, Window: "1h"})
	require.NoError(t, err)
	require.Len(t, res.Rows, 1)
	r := res.Rows[0]
	assert.Equal(t, "pre-commit", r.HookName)
	assert.Equal(t, "", r.ToolName)
	assert.Equal(t, int64(100), r.Count)
	assert.Equal(t, int64(1), r.Failures)
	assert.Equal(t, int64(1000), r.P50Ms, "p50 = max across tools")
	assert.Equal(t, int64(5000), r.P95Ms, "p95 = max across tools")
}

func TestRun_ByTool_KeepsRowsSeparate(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	insertStat(t, conn, "pre-commit", "Bash", 60, 10, 1, 1000, 5000)
	insertStat(t, conn, "pre-commit", "Read", 60, 90, 0, 50, 100)

	res, err := queries.Run(queries.Options{DBPath: dbPath, Window: "1h", ByTool: true})
	require.NoError(t, err)
	require.Len(t, res.Rows, 2)
	// Sort: hook asc then tool asc → Bash before Read.
	assert.Equal(t, "Bash", res.Rows[0].ToolName)
	assert.Equal(t, "Read", res.Rows[1].ToolName)
	assert.Equal(t, int64(10), res.Rows[0].Count)
	assert.Equal(t, int64(90), res.Rows[1].Count)
}

func TestRun_HookFilter_KeepsOnlyMatching(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	insertStat(t, conn, "lint", "", 60, 50, 0, 100, 200)
	insertStat(t, conn, "pre-commit", "Bash", 60, 25, 5, 500, 8000)
	insertStat(t, conn, "noisy", "Bash", 60, 10, 5, 100, 200)

	res, err := queries.Run(queries.Options{
		DBPath:     dbPath,
		Window:     "1h",
		HookFilter: "pre-commit",
	})
	require.NoError(t, err)
	require.Len(t, res.Rows, 1)
	assert.Equal(t, "pre-commit", res.Rows[0].HookName)
}

func TestRun_HookFilter_MissesAll_EmptyResult(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	insertStat(t, conn, "lint", "", 60, 50, 0, 100, 200)

	res, err := queries.Run(queries.Options{
		DBPath:     dbPath,
		Window:     "1h",
		HookFilter: "nonexistent",
	})
	require.NoError(t, err)
	assert.Empty(t, res.Rows)
}

// Locks the M4-T3 review fix: typing `--hook pretooluse` against a stored
// hook called "PreToolUse" must match. Strict case matching previously
// returned an empty result with no signal that capitalisation was the cause.
func TestRun_HookFilter_IsCaseInsensitive(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	insertStat(t, conn, "PreToolUse", "Bash", 60, 7, 0, 100, 200)

	res, err := queries.Run(queries.Options{
		DBPath:     dbPath,
		Window:     "1h",
		HookFilter: "pretooluse",
	})
	require.NoError(t, err)
	require.Len(t, res.Rows, 1)
	assert.Equal(t, "PreToolUse", res.Rows[0].HookName)
}

func TestRun_PicksLatestBucketPerPair(t *testing.T) {
	// Older bucket should be ignored when a newer bucket exists for the same
	// (hook, tool, window). Insert manually with explicit ts_bucket values.
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	now := aggregator.BucketStartMs(time.Now().UnixMilli(), 60)
	older := now - 60*60_000

	insert := func(bucket int64, count int64) {
		_, err := conn.Exec(`
			INSERT INTO hook_stats
			  (hook_name, tool_name, window_min, ts_bucket,
			   count, failures, p50_ms, p95_ms, p99_ms)
			VALUES (?, '', 60, ?, ?, 0, 100, 200, 300)`,
			"lint", bucket, count)
		require.NoError(t, err)
	}
	insert(older, 999) // stale
	insert(now, 42)    // current

	res, err := queries.Run(queries.Options{DBPath: dbPath, Window: "1h"})
	require.NoError(t, err)
	require.Len(t, res.Rows, 1)
	assert.Equal(t, int64(42), res.Rows[0].Count, "should pick the latest bucket")
}

// Render -------------------------------------------------------------------

func TestRender_EmptyResult_FriendNudge(t *testing.T) {
	r := queries.Result{WindowMin: 60, WindowLabel: "1시간"}
	var buf bytes.Buffer
	r.Render(&buf)
	assert.Equal(t,
		"아직 기록된 hook 통계가 없어. daemon을 띄우고 좀 기다려봐.\n",
		buf.String())
}

func TestRender_Header_IncludesWindowLabel(t *testing.T) {
	r := queries.Result{
		WindowMin:   5,
		WindowLabel: "5분",
		Rows:        []queries.Row{{HookName: "lint", Count: 1, P50Ms: 100, P95Ms: 200}},
	}
	var buf bytes.Buffer
	r.Render(&buf)
	assert.True(t, strings.HasPrefix(buf.String(), "지난 5분 hook 통계\n\n"),
		"got %q", buf.String())
}

func TestRender_NoByTool_HasExpectedColumns(t *testing.T) {
	r := queries.Result{
		WindowMin:   60,
		WindowLabel: "1시간",
		Rows: []queries.Row{
			{HookName: "lint", Count: 1234, Failures: 0, P50Ms: 100, P95Ms: 200},
		},
	}
	var buf bytes.Buffer
	r.Render(&buf)
	out := buf.String()
	assert.Contains(t, out, "hook")
	assert.Contains(t, out, "count")
	assert.Contains(t, out, "p50")
	assert.Contains(t, out, "p95")
	assert.Contains(t, out, "실패율")
	assert.NotContains(t, out, "tool", "no-by-tool layout has no tool column")
	// Thousands separator.
	assert.Contains(t, out, "1,234")
	assert.NotContains(t, out, " 1234 ")
}

func TestRender_ByTool_HasToolColumn(t *testing.T) {
	r := queries.Result{
		WindowMin:   60,
		WindowLabel: "1시간",
		ByTool:      true,
		Rows: []queries.Row{
			{HookName: "pre-commit", ToolName: "Bash", Count: 10, Failures: 1, P50Ms: 1000, P95Ms: 5000},
		},
	}
	var buf bytes.Buffer
	r.Render(&buf)
	out := buf.String()
	assert.Contains(t, out, "tool")
	assert.Contains(t, out, "Bash")
}

// Locks the M4-T3 review fix: when --by-tool is on but the underlying hook
// has no tool name (e.g. Stop), Render must surface "(none)" rather than
// silently dropping the cell. Without ByTool=true on the Result, the renderer
// previously inferred layout from the data and downgraded to no-tool layout,
// hiding the user's explicit choice.
func TestRender_ByTool_EmptyToolNameRendersAsNone(t *testing.T) {
	r := queries.Result{
		WindowMin:   60,
		WindowLabel: "1시간",
		ByTool:      true,
		Rows: []queries.Row{
			{HookName: "Stop", ToolName: "", Count: 5, Failures: 0, P50Ms: 100, P95Ms: 200},
		},
	}
	var buf bytes.Buffer
	r.Render(&buf)
	out := buf.String()
	assert.Contains(t, out, "tool", "ByTool layout should be preserved")
	assert.Contains(t, out, "(none)", "empty ToolName should render as (none)")
	assert.Contains(t, out, "Stop")
}

func TestRender_FailurePercent_ZeroCountIsZero(t *testing.T) {
	r := queries.Result{
		WindowMin:   60,
		WindowLabel: "1시간",
		Rows:        []queries.Row{{HookName: "x", Count: 0, Failures: 0}},
	}
	var buf bytes.Buffer
	r.Render(&buf)
	assert.Contains(t, buf.String(), "0%")
}

func TestRender_ColumnAlignment_LinesShareWidth(t *testing.T) {
	// tabwriter pads every column to the longest cell in that column. We assert
	// that every table line has a runs-of-2-spaces gap so column boundaries are
	// preserved when a longer cell forces width growth. This catches accidental
	// drift in the tab pattern (e.g., a missing column).
	r := queries.Result{
		WindowMin:   60,
		WindowLabel: "1시간",
		Rows: []queries.Row{
			{HookName: "a", Count: 1, P50Ms: 10, P95Ms: 20},
			{HookName: "much-longer-hook", Count: 999_999, P50Ms: 60_000, P95Ms: 125_000},
		},
	}
	var buf bytes.Buffer
	r.Render(&buf)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.GreaterOrEqual(t, len(lines), 4) // header, blank, table-header, 2 data
	tableLines := lines[2:]                  // skip "지난 …" and blank line

	for _, line := range tableLines {
		assert.Contains(t, line, "  ", "tabwriter should pad columns: %q", line)
	}
}

// table-driven helpers -----------------------------------------------------

func TestFailurePercent_HalfUp(t *testing.T) {
	cases := []struct {
		failures, count int64
		want            int
	}{
		{0, 100, 0},
		{20, 100, 20},
		{31, 200, 16},   // 15.5 → 16 (half-up via math.Round)
		{155, 1000, 16}, // 0.155 → 16, the canonical half-up case
		{154, 1000, 15}, // 0.154 → 15, just below
		{1, 3, 33},      // 33.33 → 33
		{2, 3, 67},      // 66.67 → 67
		{0, 0, 0},       // zero count never trips a divide
	}
	for _, c := range cases {
		t.Run(strconv.FormatInt(c.failures, 10)+"/"+strconv.FormatInt(c.count, 10), func(t *testing.T) {
			assert.Equal(t, c.want, queries.FailurePercentForTest(c.failures, c.count))
		})
	}
}
