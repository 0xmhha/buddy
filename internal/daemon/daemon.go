// Package daemon owns the long-running poll loop that drains the outbox
// into hook_events and refreshes hook_stats.
//
// Design (v0.1):
//   - Single goroutine, poll every PollInterval (default 1s).
//   - Process up to BatchSize outbox rows per tick (default 500).
//   - Graceful exit on SIGTERM / SIGINT: finish current tick, then close DB.
//   - PID file with flock to prevent concurrent daemons over the same DB.
//   - This package never spawns child supervisors itself; if cli-wrapper is
//     desired, M4's `buddy install --with-cliwrap` writes the cliwrap.yaml.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/wm-it-22-00661/buddy/internal/aggregator"
	"github.com/wm-it-22-00661/buddy/internal/db"
)

// Config governs how the daemon polls.
type Config struct {
	DBPath       string
	PollInterval time.Duration
	BatchSize    int
	PIDFile      string    // default: <dirname(DBPath)>/daemon.pid
	LogTo        io.Writer // structured info/errors. default os.Stderr.
}

// Defaults applies sensible defaults to zero-valued fields.
func (c *Config) Defaults() {
	if c.PollInterval == 0 {
		c.PollInterval = 1 * time.Second
	}
	if c.BatchSize == 0 {
		c.BatchSize = 500
	}
	if c.LogTo == nil {
		c.LogTo = os.Stderr
	}
	if c.PIDFile == "" && c.DBPath != "" {
		c.PIDFile = filepath.Join(filepath.Dir(c.DBPath), "daemon.pid")
	}
}

// ErrAlreadyRunning is returned by Run when another daemon already holds the PID file lock.
var ErrAlreadyRunning = errors.New("buddy daemon: another instance already running")

// Run blocks until ctx is cancelled, a fatal error occurs, or SIGTERM/SIGINT arrives.
// Foreground-only. cli-wrapper supervisors should invoke this via `buddy daemon run`.
func Run(ctx context.Context, cfg Config) error {
	cfg.Defaults()

	// Install the signal handler before publishing the PID file. Otherwise a
	// caller that watches for the PID file appearing can race in with SIGTERM
	// before NotifyContext is wired up, and the default Go signal handler
	// terminates the process.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	conn, err := db.Open(db.Options{Path: cfg.DBPath})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	pidF, err := acquirePIDFile(cfg.PIDFile)
	if err != nil {
		return err
	}
	defer releasePIDFile(pidF, cfg.PIDFile)

	fmt.Fprintf(cfg.LogTo, "buddy: daemon up (db=%s poll=%s batch=%d)\n",
		cfg.DBPath, cfg.PollInterval, cfg.BatchSize)

	// Tick once immediately so a fresh start drains anything already pending.
	if n, err := aggregator.ProcessBatch(conn, cfg.BatchSize); err != nil {
		fmt.Fprintf(cfg.LogTo, "buddy: tick error: %v\n", err)
	} else if n > 0 {
		fmt.Fprintf(cfg.LogTo, "buddy: tick processed %d rows\n", n)
	}

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(cfg.LogTo, "buddy: daemon shutting down (signal received)\n")
			return nil
		case <-ticker.C:
			if n, err := aggregator.ProcessBatch(conn, cfg.BatchSize); err != nil {
				fmt.Fprintf(cfg.LogTo, "buddy: tick error: %v\n", err)
			} else if n > 0 {
				fmt.Fprintf(cfg.LogTo, "buddy: tick processed %d rows\n", n)
			}
		}
	}
}

// Status describes a running daemon.
type Status struct {
	Running bool
	PID     int
}

// CheckStatus reads the PID file and reports whether a daemon is running.
// It does not assume ownership of the file (no lock attempt).
func CheckStatus(pidFile string) (Status, error) {
	if pidFile == "" {
		return Status{}, errors.New("pid file path required")
	}
	b, err := os.ReadFile(pidFile)
	if errors.Is(err, os.ErrNotExist) {
		return Status{Running: false}, nil
	}
	if err != nil {
		return Status{}, err
	}
	pid, err := strconv.Atoi(string(trimNewline(b)))
	if err != nil {
		return Status{}, fmt.Errorf("parse pid: %w", err)
	}
	if !processAlive(pid) {
		return Status{Running: false, PID: pid}, nil
	}
	return Status{Running: true, PID: pid}, nil
}

// Stop sends SIGTERM to the daemon recorded in the PID file.
// Returns nil if no daemon is running.
func Stop(pidFile string) error {
	st, err := CheckStatus(pidFile)
	if err != nil {
		return err
	}
	if !st.Running {
		return nil
	}
	proc, err := os.FindProcess(st.PID)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}

// --- helpers ---

func trimNewline(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == ' ') {
		b = b[:len(b)-1]
	}
	return b
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 = liveness probe on POSIX; no actual signal delivered.
	return proc.Signal(syscall.Signal(0)) == nil
}

func acquirePIDFile(path string) (*os.File, error) {
	if path == "" {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("pid dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open pid file: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, ErrAlreadyRunning
	}
	if err := f.Truncate(0); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("truncate pid: %w", err)
	}
	if _, err := f.WriteString(strconv.Itoa(os.Getpid()) + "\n"); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("write pid: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("sync pid: %w", err)
	}
	return f, nil
}

func releasePIDFile(f *os.File, path string) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
	_ = os.Remove(path)
}
