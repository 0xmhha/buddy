package queries_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/queries"
)

// fixtures -----------------------------------------------------------------

// insertEvent inserts one hook_events row directly. ts is unix ms.
func insertEvent(t *testing.T, conn *sql.DB,
	ts int64, event, hookName, toolName, sessionID string,
	durationMs int64, exitCode int,
) {
	t.Helper()
	var toolArg, sessArg any
	if toolName != "" {
		toolArg = toolName
	}
	if sessionID != "" {
		sessArg = sessionID
	}
	_, err := conn.Exec(`
		INSERT INTO hook_events
		  (ts, event, hook_name, duration_ms, exit_code, session_id, tool_name)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ts, event, hookName, durationMs, exitCode, sessArg, toolArg)
	require.NoError(t, err)
}

// RunEvents ----------------------------------------------------------------

func TestRunEvents_EmptyTable_NoErrorNoRows(t *testing.T) {
	dbPath := newTempDBPath(t)
	res, err := queries.RunEvents(queries.EventsOptions{DBPath: dbPath})
	require.NoError(t, err)
	assert.Empty(t, res.Events)

	var buf bytes.Buffer
	res.RenderLines(&buf)
	assert.Empty(t, buf.String())
}

func TestRunEvents_DefaultLimitIs20(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	base := time.Now().UnixMilli()
	for i := range 25 {
		insertEvent(t, conn, base+int64(i), "PreToolUse", "lint", "Bash", "abcd1234efgh", 100, 0)
	}
	res, err := queries.RunEvents(queries.EventsOptions{DBPath: dbPath})
	require.NoError(t, err)
	assert.Len(t, res.Events, 20)
}

func TestRunEvents_LimitReturnsMostRecentInChronologicalOrder(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	base := time.Now().UnixMilli()
	for i := range 5 {
		insertEvent(t, conn, base+int64(i*1000), "PreToolUse", "lint", "Bash", "s", int64(i*100), 0)
	}
	res, err := queries.RunEvents(queries.EventsOptions{DBPath: dbPath, Limit: 3})
	require.NoError(t, err)
	require.Len(t, res.Events, 3)
	// Most recent 3 are i=2,3,4 → ts asc after reverse.
	assert.Equal(t, base+2000, res.Events[0].Ts)
	assert.Equal(t, base+3000, res.Events[1].Ts)
	assert.Equal(t, base+4000, res.Events[2].Ts)
}

func TestRunEvents_HookFilter_IsCaseInsensitive(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	base := time.Now().UnixMilli()
	insertEvent(t, conn, base+1, "PreToolUse", "PreToolUse", "Bash", "s", 100, 0)
	insertEvent(t, conn, base+2, "PreToolUse", "lint", "Bash", "s", 100, 0)

	res, err := queries.RunEvents(queries.EventsOptions{
		DBPath:     dbPath,
		HookFilter: "pretooluse",
	})
	require.NoError(t, err)
	require.Len(t, res.Events, 1)
	assert.Equal(t, "PreToolUse", res.Events[0].HookName)
}

func TestRunEvents_SinceFilterExcludesOlderEvents(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	base := time.Now().UnixMilli()
	insertEvent(t, conn, base+1000, "PreToolUse", "lint", "Bash", "s", 100, 0)
	insertEvent(t, conn, base+2000, "PreToolUse", "lint", "Bash", "s", 100, 0)
	insertEvent(t, conn, base+3000, "PreToolUse", "lint", "Bash", "s", 100, 0)

	res, err := queries.RunEvents(queries.EventsOptions{
		DBPath: dbPath,
		Since:  base + 1500,
	})
	require.NoError(t, err)
	require.Len(t, res.Events, 2)
	assert.Equal(t, base+2000, res.Events[0].Ts)
	assert.Equal(t, base+3000, res.Events[1].Ts)
}

// RenderLines --------------------------------------------------------------

func TestRenderLines_EmptyToolNameRendersAsDash(t *testing.T) {
	r := queries.EventsResult{Events: []queries.Event{
		{Ts: 0, Event: "Stop", HookName: "Stop", DurationMs: 50, ExitCode: 0, SessionID: "abcd1234efgh"},
	}}
	var buf bytes.Buffer
	r.RenderLines(&buf)
	assert.Contains(t, buf.String(), "  Stop  Stop  -  exit=0")
}

func TestRenderLines_EmptySessionIDRendersAsDash(t *testing.T) {
	r := queries.EventsResult{Events: []queries.Event{
		{Ts: 0, Event: "PreToolUse", HookName: "lint", ToolName: "Bash", DurationMs: 100, ExitCode: 0, SessionID: ""},
	}}
	var buf bytes.Buffer
	r.RenderLines(&buf)
	assert.Contains(t, buf.String(), "sess=-")
}

func TestRenderLines_SessionIDTruncatedTo8Chars(t *testing.T) {
	r := queries.EventsResult{Events: []queries.Event{
		{Ts: 0, Event: "PreToolUse", HookName: "lint", ToolName: "Bash", SessionID: "abcd1234XXXXX"},
	}}
	var buf bytes.Buffer
	r.RenderLines(&buf)
	out := buf.String()
	assert.Contains(t, out, "sess=abcd1234")
	assert.NotContains(t, out, "abcd1234X", "must truncate, not pass-through")
}

func TestRenderLines_SessionIDShorterThan8IsKeptVerbatim(t *testing.T) {
	r := queries.EventsResult{Events: []queries.Event{
		{Ts: 0, Event: "PreToolUse", HookName: "lint", ToolName: "Bash", SessionID: "abc"},
	}}
	var buf bytes.Buffer
	r.RenderLines(&buf)
	assert.Contains(t, buf.String(), "sess=abc")
}

func TestRenderLines_DurationUsesFormatHelper(t *testing.T) {
	cases := []struct {
		ms       int64
		contains string
	}{
		{120, "dur=120ms"},
		{5500, "dur=5.5s"},
		{60_000, "dur=1.0m"},
	}
	for _, c := range cases {
		t.Run(c.contains, func(t *testing.T) {
			r := queries.EventsResult{Events: []queries.Event{
				{Ts: 0, Event: "X", HookName: "x", ToolName: "Bash", DurationMs: c.ms, SessionID: "s"},
			}}
			var buf bytes.Buffer
			r.RenderLines(&buf)
			assert.Contains(t, buf.String(), c.contains)
		})
	}
}

func TestRenderLines_TimestampIsRFC3339UTC(t *testing.T) {
	// 2026-04-24T11:23:45Z = 1777288425000 ms
	ts := time.Date(2026, 4, 24, 11, 23, 45, 0, time.UTC).UnixMilli()
	r := queries.EventsResult{Events: []queries.Event{
		{Ts: ts, Event: "PostToolUse", HookName: "pre-commit", ToolName: "Bash",
			DurationMs: 1200, ExitCode: 0, SessionID: "abc1234XYZ"},
	}}
	var buf bytes.Buffer
	r.RenderLines(&buf)
	assert.Contains(t, buf.String(), "2026-04-24T11:23:45Z")
	// Spec example line shape (2-space separators).
	assert.Contains(t, buf.String(), "pre-commit  PostToolUse  Bash  exit=0  dur=1.2s  sess=abc1234")
}

func TestRenderLines_OneLinePerEvent(t *testing.T) {
	r := queries.EventsResult{Events: []queries.Event{
		{Ts: 0, Event: "A", HookName: "h", ToolName: "T", SessionID: "s"},
		{Ts: 1000, Event: "B", HookName: "h", ToolName: "T", SessionID: "s"},
		{Ts: 2000, Event: "C", HookName: "h", ToolName: "T", SessionID: "s"},
	}}
	var buf bytes.Buffer
	r.RenderLines(&buf)
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	assert.Len(t, lines, 3)
}

// Follow -------------------------------------------------------------------

// Synthetic Follow test: insert events, run Follow with a very short poll
// interval, cancel the context after a short while. Asserts the initial fetch
// hits stdout, friend-tone markers hit stderr, and cancel returns nil cleanly.
func TestFollow_InitialFetchAndCleanCancel(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	base := time.Now().UnixMilli()
	insertEvent(t, conn, base+1, "PreToolUse", "lint", "Bash", "abcd1234EXTRA", 100, 0)
	insertEvent(t, conn, base+2, "PostToolUse", "lint", "Bash", "abcd1234EXTRA", 200, 0)

	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- queries.Follow(ctx, queries.EventsOptions{
			DBPath:       dbPath,
			PollInterval: 50 * time.Millisecond,
			Stderr:       &stderr,
		}, &stdout)
	}()

	// Give Follow time to do the initial fetch + at least one tick.
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err, "Follow should return nil on ctx cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("Follow did not return after ctx cancel")
	}

	// Both initial events should be in stdout (chronological order).
	assert.Contains(t, stdout.String(), "lint  PreToolUse  Bash")
	assert.Contains(t, stdout.String(), "lint  PostToolUse  Bash")
	// SessionID truncated to 8 chars in render.
	assert.Contains(t, stdout.String(), "sess=abcd1234")
	assert.NotContains(t, stdout.String(), "abcd1234E")

	// Friend-tone markers on stderr only.
	assert.Contains(t, stderr.String(), "(따라가는 중. Ctrl-C로 멈춰.)")
	assert.Contains(t, stderr.String(), "(끝.)")
}

// Follow's polling loop should pick up an event inserted after the initial
// fetch, using the Since cursor maintained internally. Without this guarantee
// follow would silently drop new rows.
func TestFollow_PicksUpNewEventAfterInitialFetch(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	base := time.Now().UnixMilli()
	insertEvent(t, conn, base, "PreToolUse", "first", "Bash", "s1XXXXXXX", 100, 0)

	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- queries.Follow(ctx, queries.EventsOptions{
			DBPath:       dbPath,
			PollInterval: 30 * time.Millisecond,
			Stderr:       &stderr,
		}, &stdout)
	}()

	// Let the initial fetch happen.
	time.Sleep(50 * time.Millisecond)
	insertEvent(t, conn, base+10_000, "PostToolUse", "second", "Read", "s2XXXXXXX", 200, 1)
	// Wait long enough for the poll to pick it up.
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Follow did not return after ctx cancel")
	}

	out := stdout.String()
	assert.Contains(t, out, "first  PreToolUse")
	assert.Contains(t, out, "second  PostToolUse")
	assert.Contains(t, out, "exit=1", "second event's non-zero exit code should appear")
}

// Same-millisecond cursor regression: with a strict `WHERE ts > ?` cursor,
// Follow would silently drop the second of two rows landing in the same
// unix-ms because the cursor advanced past that ts. The lexicographic
// (ts, id) cursor must keep both rows visible across ticks.
func TestFollow_PicksUpEventsWithSameMillisecondTs(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)

	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- queries.Follow(ctx, queries.EventsOptions{
			DBPath:       dbPath,
			PollInterval: 30 * time.Millisecond,
			Stderr:       &stderr,
		}, &stdout)
	}()

	// Let the initial fetch complete on the empty table.
	time.Sleep(50 * time.Millisecond)

	// Insert three rows at the same ms — the daemon batches inserts so this
	// is realistic. With a strict ts > cursor, the 2nd and 3rd would drop.
	sameMs := time.Now().UnixMilli()
	insertEvent(t, conn, sameMs, "PreToolUse", "h", "Bash", "s1XXXXXXX", 100, 0)
	insertEvent(t, conn, sameMs, "PostToolUse", "h", "Bash", "s1XXXXXXX", 200, 0)
	insertEvent(t, conn, sameMs, "Stop", "h", "Bash", "s1XXXXXXX", 300, 0)

	// Wait long enough for ≥2 ticks at 30ms cadence.
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Follow did not return after ctx cancel")
	}

	out := stdout.String()
	// All three same-ms events must appear.
	assert.Contains(t, out, "PreToolUse")
	assert.Contains(t, out, "PostToolUse")
	assert.Contains(t, out, "Stop")
	// Sanity: count distinct event lines (one per row, no duplicates from
	// the cursor re-fetching the same id).
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	assert.Len(t, lines, 3, "exactly three events expected, got %d:\n%s", len(lines), out)
}

// Permanent DB failure must abort instead of spinning forever at 1Hz. We
// inject a fake tail func that always errors and assert Follow returns once
// the consecutive-error cap is hit.
func TestFollow_AbortsAfterConsecutiveErrors(t *testing.T) {
	dbPath := newTempDBPath(t)

	var calls atomic.Int32
	failErr := errors.New("synthetic db failure")
	restore := queries.SwapFollowTailForTest(func(_ context.Context, _ *sql.DB,
		_ string, _, _ int64, _ int) ([]queries.Event, error) {
		calls.Add(1)
		return nil, failErr
	})
	defer restore()

	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := queries.Follow(ctx, queries.EventsOptions{
		DBPath:       dbPath,
		PollInterval: 10 * time.Millisecond,
		Stderr:       &stderr,
	}, &stdout)

	// Initial fetch already errored once, which is also a "tail" call. The
	// initial-fetch error path returns immediately — verify that path too.
	require.Error(t, err)
	assert.ErrorIs(t, err, failErr)
	// Initial-fetch failure aborts before the polling loop, so we expect
	// exactly one call. That's stricter than "≥5" but matches the
	// fail-fast invariant for the entry-point tail.
	assert.Equal(t, int32(1), calls.Load(),
		"initial-fetch error must abort immediately, not enter the poll loop")
	// End marker still printed even on error exit.
	assert.Contains(t, stderr.String(), "(끝.)")
}

// Variant: initial fetch succeeds, then every poll fails. Must abort after
// followMaxConsecutiveErrors (5) ticks with the underlying error wrapped.
func TestFollow_AbortsAfter5ConsecutivePollErrors(t *testing.T) {
	dbPath := newTempDBPath(t)

	var calls atomic.Int32
	failErr := errors.New("synthetic poll failure")
	restore := queries.SwapFollowTailForTest(func(_ context.Context, _ *sql.DB,
		_ string, _, _ int64, _ int) ([]queries.Event, error) {
		// First call (initial fetch) succeeds with no rows; subsequent
		// poll calls fail.
		if calls.Add(1) == 1 {
			return nil, nil
		}
		return nil, failErr
	})
	defer restore()

	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := queries.Follow(ctx, queries.EventsOptions{
		DBPath:       dbPath,
		PollInterval: 5 * time.Millisecond,
		Stderr:       &stderr,
	}, &stdout)

	require.Error(t, err)
	assert.ErrorIs(t, err, failErr)
	// 1 initial + 5 poll attempts = 6 total calls expected.
	assert.Equal(t, int32(6), calls.Load(),
		"expected 1 initial + 5 consecutive poll failures before abort")
	// The poll error message should have been logged at least once on the
	// way to the abort threshold.
	assert.Contains(t, stderr.String(), "events poll 실패")
	assert.Contains(t, stderr.String(), "(끝.)")
}

// Empty DB at start should not crash Follow; subsequent inserts must surface
// in chronological order across ticks.
func TestFollow_StartsEmptyThenPicksUpNewEvents(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)

	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- queries.Follow(ctx, queries.EventsOptions{
			DBPath:       dbPath,
			PollInterval: 50 * time.Millisecond,
			Stderr:       &stderr,
		}, &stdout)
	}()

	// Let the initial fetch complete on the empty table.
	time.Sleep(80 * time.Millisecond)
	assert.Empty(t, stdout.String(),
		"initial fetch on empty DB must produce no stdout")

	base := time.Now().UnixMilli()
	insertEvent(t, conn, base, "PreToolUse", "lint", "Bash", "s1XXXXXXX", 100, 0)
	insertEvent(t, conn, base+5, "PostToolUse", "lint", "Bash", "s1XXXXXXX", 200, 0)

	// Wait for ~2 ticks to ensure both rows are picked up.
	time.Sleep(180 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Follow did not return after ctx cancel")
	}

	out := stdout.String()
	assert.Contains(t, out, "PreToolUse")
	assert.Contains(t, out, "PostToolUse")
	preIdx := strings.Index(out, "PreToolUse")
	postIdx := strings.Index(out, "PostToolUse")
	require.GreaterOrEqual(t, preIdx, 0)
	require.GreaterOrEqual(t, postIdx, 0)
	assert.Less(t, preIdx, postIdx,
		"events must render in chronological order")
	// Friend-tone start + end markers on stderr.
	assert.Contains(t, stderr.String(), "(따라가는 중. Ctrl-C로 멈춰.)")
	assert.Contains(t, stderr.String(), "(끝.)")
}

// --limit boundary: negative is rejected with ErrInvalidLimit; zero defaults.
func TestRunEvents_NegativeLimitReturnsErrInvalidLimit(t *testing.T) {
	dbPath := newTempDBPath(t)
	_, err := queries.RunEvents(queries.EventsOptions{DBPath: dbPath, Limit: -1})
	require.Error(t, err)
	assert.ErrorIs(t, err, queries.ErrInvalidLimit)
}

func TestRunEvents_ZeroLimitDefaultsTo20(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	base := time.Now().UnixMilli()
	for i := range 25 {
		insertEvent(t, conn, base+int64(i), "PreToolUse", "lint", "Bash", "s", 100, 0)
	}
	res, err := queries.RunEvents(queries.EventsOptions{DBPath: dbPath, Limit: 0})
	require.NoError(t, err)
	assert.Len(t, res.Events, 20, "Limit=0 must default to defaultEventsLimit (20)")
}

func TestFollow_NegativeLimitReturnsErrInvalidLimit(t *testing.T) {
	dbPath := newTempDBPath(t)
	var stdout, stderr bytes.Buffer
	err := queries.Follow(context.Background(), queries.EventsOptions{
		DBPath: dbPath,
		Limit:  -5,
		Stderr: &stderr,
	}, &stdout)
	require.Error(t, err)
	assert.ErrorIs(t, err, queries.ErrInvalidLimit)
}
