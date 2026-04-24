package hookwrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/hookwrap"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

func dbPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "buddy.db")
}

func readPayload(t *testing.T, path string, id int64) schema.HookEventPayload {
	t.Helper()
	conn, err := db.Open(db.Options{Path: path, ReadOnly: true})
	require.NoError(t, err)
	defer conn.Close()

	var raw string
	require.NoError(t, conn.QueryRow(
		"SELECT payload FROM hook_outbox WHERE id = ?", id,
	).Scan(&raw))

	var p schema.HookEventPayload
	require.NoError(t, json.Unmarshal([]byte(raw), &p))
	return p
}

func TestHookWrap_RecordsSuccessfulRun(t *testing.T) {
	path := dbPath(t)
	stdin := `{"session_id":"sess-x","cwd":"/tmp/x","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo hi"}}`

	res, err := hookwrap.Run(context.Background(), hookwrap.Options{
		HookName: "pre-commit",
		Command:  []string{"sh", "-c", "exit 0"},
		DBPath:   path,
		Stdin:    strings.NewReader(stdin),
		Stderr:   &bytes.Buffer{},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode)
	assert.Greater(t, res.OutboxID, int64(0))

	p := readPayload(t, path, res.OutboxID)
	assert.Equal(t, schema.EventPreToolUse, p.Event)
	assert.Equal(t, "pre-commit", p.HookName)
	assert.Equal(t, 0, p.ExitCode)
	assert.Equal(t, "Bash", p.ToolName)
	assert.Equal(t, "sess-x", p.SessionID)
	assert.Equal(t, "/tmp/x", p.Cwd)
	assert.Nil(t, p.ToolArgs, "toolArgs must be omitted by default (privacy)")
}

func TestHookWrap_PassesNonZeroExitCode(t *testing.T) {
	path := dbPath(t)
	res, err := hookwrap.Run(context.Background(), hookwrap.Options{
		HookName: "lint",
		Command:  []string{"sh", "-c", "exit 7"},
		DBPath:   path,
		Stdin:    strings.NewReader("{}"),
	})
	require.NoError(t, err)
	assert.Equal(t, 7, res.ExitCode)

	p := readPayload(t, path, res.OutboxID)
	assert.Equal(t, 7, p.ExitCode)
}

func TestHookWrap_RecordsPositiveDurationForSlowChild(t *testing.T) {
	path := dbPath(t)
	res, err := hookwrap.Run(context.Background(), hookwrap.Options{
		HookName: "slow",
		Command:  []string{"sh", "-c", "sleep 0.08"},
		DBPath:   path,
		Stdin:    strings.NewReader("{}"),
	})
	require.NoError(t, err)

	p := readPayload(t, path, res.OutboxID)
	assert.GreaterOrEqual(t, p.DurationMs, int64(50))
}

func TestHookWrap_RecordsToolArgsWhenEnabled(t *testing.T) {
	path := dbPath(t)
	stdin := `{"tool_name":"Bash","tool_input":{"command":"ls -la"}}`

	res, err := hookwrap.Run(context.Background(), hookwrap.Options{
		HookName:       "pre",
		Command:        []string{"sh", "-c", "exit 0"},
		DBPath:         path,
		RecordToolArgs: true,
		Stdin:          strings.NewReader(stdin),
	})
	require.NoError(t, err)

	p := readPayload(t, path, res.OutboxID)
	require.NotNil(t, p.ToolArgs)
	m := p.ToolArgs.(map[string]any)
	assert.Equal(t, "ls -la", m["command"])
}

func TestHookWrap_MonitoringOnlyWhenCommandEmpty(t *testing.T) {
	path := dbPath(t)
	res, err := hookwrap.Run(context.Background(), hookwrap.Options{
		HookName: "noop",
		Command:  nil,
		DBPath:   path,
		Stdin:    strings.NewReader(`{"hook_event_name":"Stop"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode)
	assert.Greater(t, res.OutboxID, int64(0))

	p := readPayload(t, path, res.OutboxID)
	assert.Equal(t, schema.EventStop, p.Event)
}

func TestHookWrap_ReturnsSpawnFailureCode(t *testing.T) {
	path := dbPath(t)
	res, err := hookwrap.Run(context.Background(), hookwrap.Options{
		HookName: "broken",
		Command:  []string{"/nonexistent/binary-xyz123"},
		DBPath:   path,
		Stdin:    strings.NewReader("{}"),
		Stderr:   &bytes.Buffer{},
	})
	require.NoError(t, err)
	assert.Equal(t, 127, res.ExitCode)
}

func TestHookWrap_ForwardsCustomTags(t *testing.T) {
	path := dbPath(t)
	res, err := hookwrap.Run(context.Background(), hookwrap.Options{
		HookName:   "tagged",
		Command:    []string{"sh", "-c", "exit 0"},
		DBPath:     path,
		CustomTags: map[string]string{"branch": "feature/x", "exp": "A"},
		Stdin:      strings.NewReader("{}"),
	})
	require.NoError(t, err)

	p := readPayload(t, path, res.OutboxID)
	assert.Equal(t, "feature/x", p.CustomTags["branch"])
	assert.Equal(t, "A", p.CustomTags["exp"])
}
