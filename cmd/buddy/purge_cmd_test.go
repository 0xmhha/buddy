package main

import (
	"bytes"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/db"
)

// runPurge wires a fresh `buddy purge` subcommand and returns (stdout, stderr,
// err). Mirrors runConfig from config_cmd_test.go: cobra.SetArgs avoids
// shelling out to ./bin/buddy, keeping the suite fast.
func runPurge(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newPurgeCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

// preparePurgeDB returns a temp DB path with the schema migrated and a writer
// connection registered for cleanup. Caller seeds rows, then closes the
// connection BEFORE invoking the CLI (which opens its own writer connection).
func preparePurgeDB(t *testing.T) (path string, conn *sql.DB) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "buddy.db")
	c, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	return path, c
}

func seedEvent(t *testing.T, conn *sql.DB, ts int64) {
	t.Helper()
	_, err := conn.Exec(`
		INSERT INTO hook_events
		  (ts, event, hook_name, duration_ms, exit_code)
		VALUES (?, 'PreToolUse', 'lint', 10, 0)`, ts)
	require.NoError(t, err)
}

func seedStat(t *testing.T, conn *sql.DB, ts int64) {
	t.Helper()
	_, err := conn.Exec(`
		INSERT INTO hook_stats
		  (hook_name, tool_name, window_min, ts_bucket,
		   count, failures, p50_ms, p95_ms, p99_ms)
		VALUES ('lint', '', 60, ?, 1, 0, 10, 20, 30)`, ts)
	require.NoError(t, err)
}

func seedOutbox(t *testing.T, conn *sql.DB, ts int64) {
	t.Helper()
	_, err := conn.Exec(`INSERT INTO hook_outbox (ts, payload) VALUES (?, '{}')`, ts)
	require.NoError(t, err)
}

func countTable(t *testing.T, path, table string) int64 {
	t.Helper()
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	defer conn.Close()
	var n int64
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&n))
	return n
}

// TestPurge_NoBeforeFlag_FriendlyError: omitting --before is a user error and
// must surface friend-tone, not a cobra flag dump.
func TestPurge_NoBeforeFlag_FriendlyError(t *testing.T) {
	path, conn := preparePurgeDB(t)
	require.NoError(t, conn.Close())

	_, _, err := runPurge(t, "--db", path)
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe), "expected friendError, got %T: %v", err, err)
	assert.Contains(t, fe.msg, "--before")
}

// TestPurge_InvalidBeforeFormat_FriendlyError: an unrecognised --before should
// translate the parser's English error into the friend-tone shell.
func TestPurge_InvalidBeforeFormat_FriendlyError(t *testing.T) {
	path, conn := preparePurgeDB(t)
	require.NoError(t, conn.Close())

	_, _, err := runPurge(t, "--db", path, "--before", "tomorrow")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe))
	assert.Contains(t, fe.msg, "--before")
	// Friend-tone hint should mention at least one accepted form so the user
	// doesn't have to guess.
	assert.True(t,
		strings.Contains(fe.msg, "30d") || strings.Contains(fe.msg, "2026-"),
		"expected an example form in the error, got: %s", fe.msg)
}

// TestPurge_DryRunByDefault_CountsAndPreserves: the default-safe path must
// preserve every row and report what *would* be deleted.
func TestPurge_DryRunByDefault_CountsAndPreserves(t *testing.T) {
	path, conn := preparePurgeDB(t)
	for _, ts := range []int64{1, 2, 3} {
		seedEvent(t, conn, ts)
	}
	for _, ts := range []int64{10, 20} {
		seedStat(t, conn, ts)
	}
	require.NoError(t, conn.Close())

	_, stderr, err := runPurge(t, "--db", path, "--before", "2099-01-01")
	require.NoError(t, err)

	// Friend-tone count summary on stderr.
	assert.Contains(t, stderr, "3")
	assert.Contains(t, stderr, "2")
	assert.Contains(t, stderr, "dry-run")
	assert.Contains(t, stderr, "--apply")
	// Outbox is mentioned ("we don't touch it") so users know it's intentional.
	assert.Contains(t, strings.ToLower(stderr), "outbox")

	// Nothing actually deleted.
	assert.Equal(t, int64(3), countTable(t, path, "hook_events"))
	assert.Equal(t, int64(2), countTable(t, path, "hook_stats"))
}

// TestPurge_Apply_DeletesAndPreservesOutbox is the headline acceptance:
// --apply deletes events/stats AND leaves hook_outbox completely untouched.
func TestPurge_Apply_DeletesAndPreservesOutbox(t *testing.T) {
	path, conn := preparePurgeDB(t)
	for _, ts := range []int64{1, 2, 3, 4, 5} {
		seedEvent(t, conn, ts)
	}
	for _, ts := range []int64{10, 20, 30} {
		seedStat(t, conn, ts)
	}
	for _, ts := range []int64{1, 2, 99} { // far below cutoff
		seedOutbox(t, conn, ts)
	}
	require.NoError(t, conn.Close())

	outboxBefore := countTable(t, path, "hook_outbox")
	require.Equal(t, int64(3), outboxBefore)

	_, stderr, err := runPurge(t, "--db", path, "--before", "2099-01-01", "--apply")
	require.NoError(t, err)
	assert.Contains(t, stderr, "5")
	assert.Contains(t, stderr, "3")
	assert.Contains(t, strings.ToLower(stderr), "outbox")

	assert.Equal(t, int64(0), countTable(t, path, "hook_events"))
	assert.Equal(t, int64(0), countTable(t, path, "hook_stats"))
	// Headline invariant: outbox unchanged.
	assert.Equal(t, outboxBefore, countTable(t, path, "hook_outbox"),
		"hook_outbox must not be touched by purge")
}

// TestPurge_DBMissing_FriendlyError: a nonexistent --db path triggers the T8
// sentinel translation, NOT a leaked SQLite OOM(14) message.
func TestPurge_DBMissing_FriendlyError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir", "buddy.db")
	_, _, err := runPurge(t, "--db", missing, "--before", "30d")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe))
	// Match the T8 wording vibe — "DB가 아직 없어".
	assert.Contains(t, fe.msg, "DB")
	assert.NotContains(t, strings.ToLower(fe.msg), "out of memory")
}

// TestPurge_RelativeBefore_Works ensures the relative-day form actually
// drives the cutoff. Seed a row at ts=1 (very old) and one at "now-ish";
// purge with --before 1d --apply should drop only the old one.
func TestPurge_RelativeBefore_Works(t *testing.T) {
	path, conn := preparePurgeDB(t)
	// Tiny ts (≈ 1970) and a near-future ts that's clearly newer than 1d ago.
	seedEvent(t, conn, 1)
	// 5_000_000_000_000 ≈ year 2128, so it'll always survive --before 1d.
	seedEvent(t, conn, 5_000_000_000_000)
	require.NoError(t, conn.Close())

	_, _, err := runPurge(t, "--db", path, "--before", "1d", "--apply")
	require.NoError(t, err)
	assert.Equal(t, int64(1), countTable(t, path, "hook_events"))
}
