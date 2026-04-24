package db_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

func TestOpen_CreatesAllTables(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	rows, err := conn.Query(
		"SELECT name FROM sqlite_master WHERE type='table' ORDER BY name",
	)
	require.NoError(t, err)
	defer rows.Close()

	got := map[string]bool{}
	for rows.Next() {
		var n string
		require.NoError(t, rows.Scan(&n))
		got[n] = true
	}
	for _, want := range []string{"hook_outbox", "hook_events", "hook_stats", "schema_version"} {
		assert.True(t, got[want], "missing table %s", want)
	}
}

func TestOpen_RecordsSchemaVersion1(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	var v int
	require.NoError(t, conn.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&v))
	assert.Equal(t, 1, v)
}

func TestOpen_IsIdempotentAcrossReopens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buddy.db")

	conn1, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	require.NoError(t, conn1.Close())

	conn2, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn2.Close() })

	var count int
	require.NoError(t, conn2.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count))
	assert.Equal(t, 1, count)
}

func TestOpen_UsesWALMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buddy.db")

	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	var mode string
	require.NoError(t, conn.QueryRow("PRAGMA journal_mode").Scan(&mode))
	assert.Equal(t, "wal", mode)
}

func TestOutbox_AppendAndReadPending(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	p := &schema.HookEventPayload{
		Ts:         1700_000_000_000,
		Event:      schema.EventPreToolUse,
		HookName:   "pre-commit",
		DurationMs: 42,
		ExitCode:   0,
		ToolName:   "Bash",
	}
	id, err := db.AppendToOutbox(conn, p)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	pending, err := db.ReadPendingOutbox(conn, 100)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, id, pending[0].ID)
}

func TestOutbox_MarkConsumed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	p := &schema.HookEventPayload{
		Ts: 1, Event: schema.EventStop, HookName: "x",
		DurationMs: 1, ExitCode: 0,
	}
	id, err := db.AppendToOutbox(conn, p)
	require.NoError(t, err)

	require.NoError(t, db.MarkConsumed(conn, []int64{id}))

	pending, err := db.ReadPendingOutbox(conn, 100)
	require.NoError(t, err)
	assert.Empty(t, pending)
}
