package db_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// TestOpen_ReadOnly_MissingFile_ReturnsErrDBMissing covers the case where the
// parent directory exists but the DB file itself does not. Without the sentinel
// translation, modernc.org/sqlite would lazily create an empty file in mode=ro
// (or surface "no such table: hook_outbox" on first query) — neither helpful.
// (M5 T8.)
func TestOpen_ReadOnly_MissingFile_ReturnsErrDBMissing(t *testing.T) {
	parent := t.TempDir() // exists
	dbPath := filepath.Join(parent, "missing.db")

	conn, err := db.Open(db.Options{Path: dbPath, ReadOnly: true})
	if conn != nil {
		_ = conn.Close()
	}
	require.Error(t, err)
	assert.True(t, errors.Is(err, db.ErrDBMissing),
		"expected ErrDBMissing for read-only open of nonexistent file; got %v", err)
	assert.NotContains(t, err.Error(), "out of memory",
		"sentinel must replace SQLite's cryptic OOM (14) message")
}

// TestOpen_ReadOnly_MissingParent_ReturnsErrDBMissing covers the headline T8
// failure: a user passing --db <nowhere>/buddy.db with the parent dir absent
// previously saw "out of memory (14)". The stat-first check converts that to
// the sentinel that the CLI translates to a friend-tone message.
func TestOpen_ReadOnly_MissingParent_ReturnsErrDBMissing(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "no-such-subdir")
	dbPath := filepath.Join(missingDir, "buddy.db")

	conn, err := db.Open(db.Options{Path: dbPath, ReadOnly: true})
	if conn != nil {
		_ = conn.Close()
	}
	require.Error(t, err)
	assert.True(t, errors.Is(err, db.ErrDBMissing),
		"expected ErrDBMissing for read-only open with missing parent; got %v", err)
	// Regression seal: T8's whole point is to never let "out of memory (14)"
	// reach the CLI again.
	assert.NotContains(t, strings.ToLower(err.Error()), "out of memory",
		"must not leak SQLite OOM(14) wording to callers")
	// Read-only opens still must not have a write side effect.
	_, statErr := os.Stat(missingDir)
	assert.True(t, os.IsNotExist(statErr),
		"sentinel translation must not create parent dir %q (stat err: %v)",
		missingDir, statErr)
}

// TestOpen_Writable_DeepNestedMissingParent_Creates is a regression test:
// the writable path's existing os.MkdirAll(<dir>, 0755) must keep working for
// deeply nested missing parents. T8 must not regress install/daemon/hook-wrap.
func TestOpen_Writable_DeepNestedMissingParent_Creates(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "a", "b", "c", "buddy.db")

	conn, err := db.Open(db.Options{Path: dbPath}) // writable
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr, "writable open should create the DB file even with deeply nested missing parents")
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
