package purge_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/purge"
)

// fixtures -----------------------------------------------------------------

// openTempDB returns a freshly migrated DB and registers a Cleanup. Uses the
// same pattern as queries/stats_test.go so the schema (hook_outbox /
// hook_events / hook_stats) is real.
func openTempDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// insertEvent inserts one hook_events row at the given ts (unix ms).
func insertEvent(t *testing.T, conn *sql.DB, ts int64) {
	t.Helper()
	_, err := conn.Exec(`
		INSERT INTO hook_events
		  (ts, event, hook_name, duration_ms, exit_code)
		VALUES (?, 'PreToolUse', 'lint', 10, 0)`, ts)
	require.NoError(t, err)
}

// insertStat inserts one hook_stats row at the given ts_bucket.
func insertStat(t *testing.T, conn *sql.DB, hookName, toolName string, windowMin int, bucket int64) {
	t.Helper()
	_, err := conn.Exec(`
		INSERT INTO hook_stats
		  (hook_name, tool_name, window_min, ts_bucket,
		   count, failures, p50_ms, p95_ms, p99_ms)
		VALUES (?, ?, ?, ?, 1, 0, 10, 20, 30)`,
		hookName, toolName, windowMin, bucket)
	require.NoError(t, err)
}

// insertOutbox inserts one hook_outbox row. We never delete from this — the
// invariant we are protecting is "purge MUST NOT touch hook_outbox".
func insertOutbox(t *testing.T, conn *sql.DB, ts int64) {
	t.Helper()
	_, err := conn.Exec(`
		INSERT INTO hook_outbox (ts, payload) VALUES (?, ?)`, ts, "{}")
	require.NoError(t, err)
}

func count(t *testing.T, conn *sql.DB, table string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&n))
	return n
}

// tests --------------------------------------------------------------------

func TestRun_DryRun_CountsButDoesNotDelete(t *testing.T) {
	conn := openTempDB(t)

	// 3 old events, 2 new events; cutoff is 100.
	insertEvent(t, conn, 10)
	insertEvent(t, conn, 50)
	insertEvent(t, conn, 99)
	insertEvent(t, conn, 200)
	insertEvent(t, conn, 300)

	// 2 old stats, 1 new.
	insertStat(t, conn, "lint", "", 60, 10)
	insertStat(t, conn, "lint", "Bash", 60, 50)
	insertStat(t, conn, "lint", "Edit", 60, 200)

	res, err := purge.Run(conn, purge.Options{BeforeMillis: 100, DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, int64(3), res.Events)
	assert.Equal(t, int64(2), res.Stats)

	// Nothing actually deleted.
	assert.Equal(t, int64(5), count(t, conn, "hook_events"))
	assert.Equal(t, int64(3), count(t, conn, "hook_stats"))
}

func TestRun_Apply_DeletesRowsBelowCutoff(t *testing.T) {
	conn := openTempDB(t)

	insertEvent(t, conn, 10)
	insertEvent(t, conn, 50)
	insertEvent(t, conn, 99)
	insertEvent(t, conn, 200) // kept
	insertEvent(t, conn, 300) // kept

	insertStat(t, conn, "lint", "", 60, 10)
	insertStat(t, conn, "lint", "Bash", 60, 50)
	insertStat(t, conn, "lint", "Edit", 60, 200) // kept

	res, err := purge.Run(conn, purge.Options{BeforeMillis: 100, DryRun: false})
	require.NoError(t, err)
	assert.Equal(t, int64(3), res.Events)
	assert.Equal(t, int64(2), res.Stats)

	assert.Equal(t, int64(2), count(t, conn, "hook_events"))
	assert.Equal(t, int64(1), count(t, conn, "hook_stats"))

	// Verify the survivors are the post-cutoff rows.
	var minEvtTs int64
	require.NoError(t, conn.QueryRow("SELECT MIN(ts) FROM hook_events").Scan(&minEvtTs))
	assert.GreaterOrEqual(t, minEvtTs, int64(100))

	var minStatTs int64
	require.NoError(t, conn.QueryRow("SELECT MIN(ts_bucket) FROM hook_stats").Scan(&minStatTs))
	assert.GreaterOrEqual(t, minStatTs, int64(100))
}

// TestRun_NeverTouchesOutbox is the headline invariant: hook_outbox is the
// sync-write WAL the hook wrapper relies on; purge must never touch it.
// Verified for BOTH DryRun and Apply paths.
func TestRun_NeverTouchesOutbox(t *testing.T) {
	conn := openTempDB(t)

	// Seed outbox with a mix of ts values, including some FAR below the
	// cutoff so a buggy implementation that filtered hook_outbox would
	// inevitably delete them.
	for _, ts := range []int64{1, 2, 3, 50, 99, 150, 999} {
		insertOutbox(t, conn, ts)
	}
	before := count(t, conn, "hook_outbox")
	require.Equal(t, int64(7), before)

	// Also seed events/stats so the purge actually does something (a no-op
	// purge could trivially "not touch" outbox). The cutoff catches some.
	insertEvent(t, conn, 10)
	insertEvent(t, conn, 200)
	insertStat(t, conn, "lint", "", 60, 10)
	insertStat(t, conn, "lint", "", 60, 200)

	// DryRun.
	_, err := purge.Run(conn, purge.Options{BeforeMillis: 100, DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, before, count(t, conn, "hook_outbox"),
		"DryRun must not change hook_outbox")

	// Apply.
	_, err = purge.Run(conn, purge.Options{BeforeMillis: 100, DryRun: false})
	require.NoError(t, err)
	assert.Equal(t, before, count(t, conn, "hook_outbox"),
		"Apply must not change hook_outbox")
}

func TestRun_EmptyDB_ReturnsZeros(t *testing.T) {
	conn := openTempDB(t)

	dryRes, err := purge.Run(conn, purge.Options{BeforeMillis: 1_000_000, DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, int64(0), dryRes.Events)
	assert.Equal(t, int64(0), dryRes.Stats)

	applyRes, err := purge.Run(conn, purge.Options{BeforeMillis: 1_000_000, DryRun: false})
	require.NoError(t, err)
	assert.Equal(t, int64(0), applyRes.Events)
	assert.Equal(t, int64(0), applyRes.Stats)
}

func TestRun_OnlyOldRows_ReturnsAllAffected(t *testing.T) {
	conn := openTempDB(t)
	insertEvent(t, conn, 1)
	insertEvent(t, conn, 2)
	insertEvent(t, conn, 3)
	insertStat(t, conn, "h", "", 60, 1)
	insertStat(t, conn, "h", "Bash", 60, 2)

	res, err := purge.Run(conn, purge.Options{BeforeMillis: 1_000_000, DryRun: false})
	require.NoError(t, err)
	assert.Equal(t, int64(3), res.Events)
	assert.Equal(t, int64(2), res.Stats)
	assert.Equal(t, int64(0), count(t, conn, "hook_events"))
	assert.Equal(t, int64(0), count(t, conn, "hook_stats"))
}

func TestRun_OnlyNewRows_ReturnsZero(t *testing.T) {
	conn := openTempDB(t)
	insertEvent(t, conn, 5_000)
	insertEvent(t, conn, 10_000)
	insertStat(t, conn, "h", "", 60, 5_000)

	res, err := purge.Run(conn, purge.Options{BeforeMillis: 100, DryRun: false})
	require.NoError(t, err)
	assert.Equal(t, int64(0), res.Events)
	assert.Equal(t, int64(0), res.Stats)
	assert.Equal(t, int64(2), count(t, conn, "hook_events"))
	assert.Equal(t, int64(1), count(t, conn, "hook_stats"))
}

// TestRun_BoundaryIsStrictLessThan ensures the cutoff is `< before`, never
// `<=`. A row at exactly `BeforeMillis` is preserved.
func TestRun_BoundaryIsStrictLessThan(t *testing.T) {
	conn := openTempDB(t)
	insertEvent(t, conn, 99)              // < 100, deleted
	insertEvent(t, conn, 100)             // == 100, kept
	insertEvent(t, conn, 101)             // > 100, kept
	insertStat(t, conn, "h", "", 60, 99)  // deleted
	insertStat(t, conn, "h", "", 60, 100) // kept

	res, err := purge.Run(conn, purge.Options{BeforeMillis: 100, DryRun: false})
	require.NoError(t, err)
	assert.Equal(t, int64(1), res.Events)
	assert.Equal(t, int64(1), res.Stats)
	assert.Equal(t, int64(2), count(t, conn, "hook_events"))
	assert.Equal(t, int64(1), count(t, conn, "hook_stats"))
}
