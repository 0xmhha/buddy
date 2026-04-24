package queries_test

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
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
	for i := 0; i < 25; i++ {
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
	for i := 0; i < 5; i++ {
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
