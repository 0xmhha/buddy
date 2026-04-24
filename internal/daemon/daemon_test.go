package daemon_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/daemon"
	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

func setupDB(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "buddy.db")
	pidFile := filepath.Join(dir, "daemon.pid")
	return dbPath, pidFile
}

func enqueueOne(t *testing.T, dbPath string) {
	t.Helper()
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	defer conn.Close()
	_, err = db.AppendToOutbox(conn, &schema.HookEventPayload{
		Ts:         time.Now().UnixMilli(),
		Event:      schema.EventPostToolUse,
		HookName:   "h",
		ToolName:   "Bash",
		DurationMs: 5,
		ExitCode:   0,
	})
	require.NoError(t, err)
}

func TestRun_DrainsOutboxThenStopsOnContextCancel(t *testing.T) {
	dbPath, pidFile := setupDB(t)
	enqueueOne(t, dbPath)

	var logbuf bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- daemon.Run(ctx, daemon.Config{
			DBPath:       dbPath,
			PIDFile:      pidFile,
			PollInterval: 50 * time.Millisecond,
			LogTo:        &logbuf,
		})
	}()

	// Wait until at least one tick has run. Polling every 10ms keeps the test fast.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := db.Open(db.Options{Path: dbPath, ReadOnly: true})
		require.NoError(t, err)
		var c int
		_ = conn.QueryRow("SELECT COUNT(*) FROM hook_events").Scan(&c)
		conn.Close()
		if c > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not stop after cancel")
	}

	// hook_events row present
	conn, err := db.Open(db.Options{Path: dbPath, ReadOnly: true})
	require.NoError(t, err)
	defer conn.Close()
	var c int
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM hook_events").Scan(&c))
	assert.GreaterOrEqual(t, c, 1)
}

func TestRun_RefusesToStartWhenAnotherInstanceHoldsPID(t *testing.T) {
	dbPath, pidFile := setupDB(t)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	done1 := make(chan error, 1)
	go func() {
		done1 <- daemon.Run(ctx1, daemon.Config{
			DBPath:       dbPath,
			PIDFile:      pidFile,
			PollInterval: 50 * time.Millisecond,
		})
	}()

	// give first daemon time to grab the lock
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		st, _ := daemon.CheckStatus(pidFile)
		if st.Running {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// second daemon must fail with ErrAlreadyRunning
	err := daemon.Run(context.Background(), daemon.Config{
		DBPath:       dbPath,
		PIDFile:      pidFile,
		PollInterval: 50 * time.Millisecond,
	})
	assert.ErrorIs(t, err, daemon.ErrAlreadyRunning)

	cancel1()
	<-done1
}

func TestStatusAndStop_OnRunningDaemon(t *testing.T) {
	dbPath, pidFile := setupDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- daemon.Run(ctx, daemon.Config{
			DBPath:       dbPath,
			PIDFile:      pidFile,
			PollInterval: 50 * time.Millisecond,
		})
	}()

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		st, _ := daemon.CheckStatus(pidFile)
		if st.Running {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	st, err := daemon.CheckStatus(pidFile)
	require.NoError(t, err)
	require.True(t, st.Running)
	assert.Greater(t, st.PID, 0)

	require.NoError(t, daemon.Stop(pidFile))

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not stop after Stop()")
	}
}

func TestCheckStatus_ReportsNotRunningWhenNoPIDFile(t *testing.T) {
	st, err := daemon.CheckStatus(filepath.Join(t.TempDir(), "no-pid"))
	require.NoError(t, err)
	assert.False(t, st.Running)
}

func TestStop_NoOpWhenNotRunning(t *testing.T) {
	require.NoError(t, daemon.Stop(filepath.Join(t.TempDir(), "no-pid")))
}
