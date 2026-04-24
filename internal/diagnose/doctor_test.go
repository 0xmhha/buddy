package diagnose_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/aggregator"
	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/diagnose"
)

// fixtures -----------------------------------------------------------------

// newTempDBPath creates a fresh DB file path inside t.TempDir() and runs the
// schema migrations by opening it once writably.
func newTempDBPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	require.NoError(t, conn.Close())
	return path
}

// writeRunningPID writes a PID file containing the current process PID, so
// daemon.CheckStatus reports Running=true. This avoids spawning real daemons
// in unit tests while still exercising the real code path.
func writeRunningPID(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path,
		[]byte(strconv.Itoa(os.Getpid())+"\n"), 0o644))
}

func nonExistentPID(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "no-such.pid")
}

// insertStat inserts one hook_stats row directly. Bucket defaults to the
// current 5-min bucket so freshness queries hit it.
func insertStat(t *testing.T, conn *sql.DB, hookName, toolName string,
	count, failures int64, p95Ms int64,
) {
	t.Helper()
	bucket := aggregator.BucketStartMs(time.Now().UnixMilli(), 5)
	_, err := conn.Exec(`
		INSERT INTO hook_stats
		  (hook_name, tool_name, window_min, ts_bucket,
		   count, failures, p50_ms, p95_ms, p99_ms)
		VALUES (?, ?, 5, ?, ?, ?, ?, ?, ?)`,
		hookName, toolName, bucket, count, failures, p95Ms, p95Ms, p95Ms)
	require.NoError(t, err)
}

func openWritable(t *testing.T, path string) *sql.DB {
	t.Helper()
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// tests --------------------------------------------------------------------

func TestCheck_DBOpenFailure_ReturnsSingleIssueAndBails(t *testing.T) {
	// modernc/sqlite opens lazily, so an open against a directory succeeds and
	// fails on the first query instead. doctor catches that and collapses to a
	// single KindDBOpen issue — same UX as a true open failure.
	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  t.TempDir(), // directory, not a file
		PIDFile: nonExistentPID(t),
	})
	require.NoError(t, err)
	assert.False(t, rep.Healthy)
	require.Len(t, rep.Issues, 1)
	assert.Equal(t, diagnose.KindDBOpen, rep.Issues[0].Kind)
	assert.Contains(t, rep.Issues[0].Message, "DB를 못 열었어")
}

func TestCheck_EmptyDB_DaemonDown_OneIssue(t *testing.T) {
	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  newTempDBPath(t),
		PIDFile: nonExistentPID(t),
	})
	require.NoError(t, err)
	assert.False(t, rep.Healthy)
	require.Len(t, rep.Issues, 1)
	assert.Equal(t, diagnose.KindDaemon, rep.Issues[0].Kind)
	assert.Contains(t, rep.Issues[0].Message, "daemon이 실행 중이 아니야")
}

func TestCheck_EmptyDB_DaemonUp_Healthy(t *testing.T) {
	pid := filepath.Join(t.TempDir(), "daemon.pid")
	writeRunningPID(t, pid)

	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  newTempDBPath(t),
		PIDFile: pid,
	})
	require.NoError(t, err)
	assert.True(t, rep.Healthy)
	assert.Empty(t, rep.Issues)
}

func TestCheck_OutboxBacklog_TriggersWithFormattedCount(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	// 1247 pending rows — comfortably over the default 1000 threshold.
	tx, err := conn.Begin()
	require.NoError(t, err)
	stmt, err := tx.Prepare("INSERT INTO hook_outbox (ts, payload) VALUES (?, ?)")
	require.NoError(t, err)
	for range 1247 {
		_, err := stmt.Exec(time.Now().UnixMilli(), "{}")
		require.NoError(t, err)
	}
	require.NoError(t, stmt.Close())
	require.NoError(t, tx.Commit())

	pid := filepath.Join(t.TempDir(), "daemon.pid")
	writeRunningPID(t, pid)

	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  dbPath,
		PIDFile: pid,
	})
	require.NoError(t, err)
	require.False(t, rep.Healthy)
	require.Len(t, rep.Issues, 1)
	assert.Equal(t, diagnose.KindBacklog, rep.Issues[0].Kind)
	assert.Contains(t, rep.Issues[0].Message, "1,247")
}

func TestCheck_SlowP95_TriggersIssue(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	insertStat(t, conn, "pre-commit", "Bash", 100, 0, 8200) // 8.2s p95

	pid := filepath.Join(t.TempDir(), "daemon.pid")
	writeRunningPID(t, pid)

	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  dbPath,
		PIDFile: pid,
	})
	require.NoError(t, err)
	require.False(t, rep.Healthy)
	require.Len(t, rep.Issues, 1)
	assert.Equal(t, diagnose.KindSlow, rep.Issues[0].Kind)
	assert.Contains(t, rep.Issues[0].Message, "pre-commit:Bash")
	assert.Contains(t, rep.Issues[0].Message, "8.2s")
	assert.Contains(t, rep.Issues[0].Message, "5.0s")
}

func TestCheck_HighFailRate_TriggersIssue(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	// 20/100 = 20% — at threshold, NOT triggering (we use strict >).
	insertStat(t, conn, "lint", "", 100, 20, 100)
	// 35/100 = 35% — over 20%, triggers.
	insertStat(t, conn, "noisy", "Bash", 100, 35, 100)

	pid := filepath.Join(t.TempDir(), "daemon.pid")
	writeRunningPID(t, pid)

	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  dbPath,
		PIDFile: pid,
	})
	require.NoError(t, err)
	require.False(t, rep.Healthy)
	require.Len(t, rep.Issues, 1)
	assert.Equal(t, diagnose.KindFailRate, rep.Issues[0].Kind)
	assert.Contains(t, rep.Issues[0].Message, "noisy:Bash")
	assert.Contains(t, rep.Issues[0].Message, "35%")
	assert.Contains(t, rep.Issues[0].Message, "최근 100번 중 35번 실패")
}

func TestCheck_AllIssuesStack_PreservesOrder(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)

	// backlog
	tx, err := conn.Begin()
	require.NoError(t, err)
	stmt, err := tx.Prepare("INSERT INTO hook_outbox (ts, payload) VALUES (?, ?)")
	require.NoError(t, err)
	for range 1500 {
		_, err := stmt.Exec(time.Now().UnixMilli(), "{}")
		require.NoError(t, err)
	}
	require.NoError(t, stmt.Close())
	require.NoError(t, tx.Commit())

	// slow p95
	insertStat(t, conn, "pre-commit", "Bash", 50, 0, 8200)
	// high failure rate (different hook so it counts as a separate row)
	insertStat(t, conn, "lint", "", 20, 7, 100)

	// daemon down (no PID file written)
	pid := nonExistentPID(t)

	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  dbPath,
		PIDFile: pid,
	})
	require.NoError(t, err)
	require.False(t, rep.Healthy)
	require.Len(t, rep.Issues, 4)

	// Order: daemon → backlog → slow → fail
	assert.Equal(t, diagnose.KindDaemon, rep.Issues[0].Kind)
	assert.Equal(t, diagnose.KindBacklog, rep.Issues[1].Kind)
	assert.Equal(t, diagnose.KindSlow, rep.Issues[2].Kind)
	assert.Equal(t, diagnose.KindFailRate, rep.Issues[3].Kind)
}

func TestCheck_SlowAndFail_LimitedToTopFive(t *testing.T) {
	dbPath := newTempDBPath(t)
	conn := openWritable(t, dbPath)
	// 7 distinct slow hooks; expect only 5 surfaced.
	for i := range 7 {
		insertStat(t, conn,
			fmt.Sprintf("hook-%d", i), "Bash",
			50, 0, int64(6000+100*i))
	}
	pid := filepath.Join(t.TempDir(), "daemon.pid")
	writeRunningPID(t, pid)

	rep, err := diagnose.Check(diagnose.Options{DBPath: dbPath, PIDFile: pid})
	require.NoError(t, err)
	slowCount := 0
	for _, iss := range rep.Issues {
		if iss.Kind == diagnose.KindSlow {
			slowCount++
		}
	}
	assert.Equal(t, 5, slowCount, "should cap slow issues to topN=5")
}

func TestRender_Healthy_ExactString(t *testing.T) {
	r := diagnose.Report{Healthy: true}
	var buf bytes.Buffer
	r.Render(&buf)
	assert.Equal(t, "모두 정상이야.\n", buf.String())
}

func TestRender_NonHealthy_HeaderAndBullets(t *testing.T) {
	r := diagnose.Report{
		Healthy: false,
		Issues: []diagnose.Diagnostic{
			{Kind: diagnose.KindDaemon, Message: "daemon이 실행 중이 아니야. 'buddy daemon start'로 띄울 수 있어."},
			{Kind: diagnose.KindBacklog, Message: "outbox에 1,247개 쌓였어. daemon 한 번 봐줘 (buddy daemon status)."},
		},
	}
	var buf bytes.Buffer
	r.Render(&buf)
	out := buf.String()
	assert.True(t, strings.HasPrefix(out, "어, 몇 가지 봐줄 게 있어.\n"),
		"output should start with header; got %q", out)
	assert.Contains(t, out, "  • daemon이 실행 중이 아니야")
	assert.Contains(t, out, "  • outbox에 1,247개 쌓였어")
}

func TestHumanDur_Boundaries(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0ms"},
		{1, "1ms"},
		{123, "123ms"},
		{999, "999ms"},
		{1000, "1.0s"},
		{1234, "1.2s"},
		{5_000, "5.0s"},
		{59_949, "59.9s"}, // last ms whose tenths-rounding stays under 60.0s
		{59_950, "1.0m"},  // promotes to minutes once tenths-of-second would render 60.0s
		{59_999, "1.0m"},  // boundary fix: previously rendered as "60.0s"
		{60_000, "1.0m"},
		{125_000, "2.1m"},
		{600_000, "10.0m"},
	}
	for _, c := range cases {
		t.Run(strconv.FormatInt(c.in, 10), func(t *testing.T) {
			got := diagnose.HumanDurForTest(c.in)
			assert.Equal(t, c.want, got)
		})
	}
}

func TestDefaultThresholds_LockInValues(t *testing.T) {
	d := diagnose.DefaultThresholds()
	// Decision 2 lock-in (v0.1-spec §6.2). Asserting the exact values guards
	// against a silent default drift breaking the documented contract.
	assert.Equal(t, int64(30_000), d.HookTimeoutMs)
	assert.Equal(t, int64(5_000), d.HookSlowMs)
	assert.Equal(t, 20, d.HookFailRatePct)
	assert.Equal(t, 1_000, d.OutboxBacklog)
}
