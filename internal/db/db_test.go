package db_test

import (
	"os"
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

func TestOpen_ReadOnly_DoesNotCreateMissingParentDir(t *testing.T) {
	// Read-only opens must not have a write side effect. A missing parent dir
	// should surface as an open/query failure (doctor collapses it to a single
	// KindDBOpen issue) rather than being silently created behind the user's back.
	missingDir := filepath.Join(t.TempDir(), "no-such-subdir")
	dbPath := filepath.Join(missingDir, "buddy.db")

	conn, err := db.Open(db.Options{Path: dbPath, ReadOnly: true})
	// Open itself may succeed (modernc/sqlite is lazy); the parent dir must
	// still be absent regardless. Force first I/O to surface the failure.
	if err == nil {
		t.Cleanup(func() { _ = conn.Close() })
		_, queryErr := conn.Exec("SELECT 1")
		assert.Error(t, queryErr, "expected first query to fail on missing DB dir")
	}
	_, statErr := os.Stat(missingDir)
	assert.True(t, os.IsNotExist(statErr),
		"read-only open must not create parent dir %q (stat err: %v)", missingDir, statErr)
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
