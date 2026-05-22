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
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/aggregator"
	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/knowledge"
	"github.com/0xmhha/buddy/internal/notify"
	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/usage"
)

// syncWriter serialises Write calls across the daemon's goroutines so
// concurrent fmt.Fprintf calls into Config.LogTo (the main Run loop +
// runSessionMonitor + runAdvisorMonitor) never overlap. io.Writer is
// not required to be concurrent-safe; os.Stderr happens to be on POSIX
// but a test's bytes.Buffer is not, and the race detector caught the
// gap during dogfood.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

// Config governs how the daemon polls.
type Config struct {
	DBPath       string
	PollInterval time.Duration
	BatchSize    int
	PIDFile      string    // default: <dirname(DBPath)>/daemon.pid
	LogTo        io.Writer // structured info/errors. default os.Stderr.

	// SessionMonitor governs the sessionMonitor goroutine (ADR-012).
	// Disabled=true skips the goroutine entirely (matches the v0.1 daemon
	// behavior for users who don't want session observation). When
	// enabled, PollInterval governs cadence (default 30s) and
	// EndedThreshold marks a session as ended once last_active is older
	// than the threshold (default 1h).
	SessionMonitor SessionMonitorConfig

	// Advisor governs the advisorMonitor goroutine (ADR-015).
	// Disabled=true skips entirely; otherwise the goroutine runs at
	// PollInterval (default 1h) and inserts fired advisories into the
	// advisories table (with dedup window applied per-kind).
	Advisor AdvisorMonitorConfig
}

// SessionMonitorConfig — ADR-012 daemon role config.
type SessionMonitorConfig struct {
	Disabled        bool
	PollInterval    time.Duration
	EndedThreshold  time.Duration
	ProjectsRoot    string // override for tests; empty → $HOME/.claude/projects
}

// AdvisorMonitorConfig — ADR-015 daemon role config. Mirrors the same
// shape as SessionMonitorConfig for consistency.
type AdvisorMonitorConfig struct {
	Disabled   bool
	Thresholds advisor.Thresholds

	// NotifyChannels are ADR-016 channel specs the daemon turns
	// into a *notify.Dispatcher at Run() time (it needs the live
	// *sql.DB the daemon already owns). Empty slice = no out-of-band
	// delivery; advisories still persist and can be read via CLI /
	// TUI / MCP.
	NotifyChannels []NotifyChannelSpec
}

// NotifyChannelSpec is a transport-agnostic descriptor the daemon
// uses to construct concrete notify.Channel instances. The cmd
// (loadconfig) builds these from config; the daemon converts them
// because Dispatcher needs a *sql.DB and we don't want loadconfig to
// open its own connection.
type NotifyChannelSpec struct {
	Kind        string // notify.ChannelDesktop / Webhook / TUIBanner / Shell
	SeverityMin string
	DedupWindow time.Duration

	// Webhook-only fields.
	URL     string
	Method  string
	Headers map[string]string
	Timeout time.Duration
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
	// Wrap the writer so the Run loop + every monitor goroutine share a
	// mutex-protected sink. Re-wrapping an already-wrapped writer is OK
	// — the outer mutex serialises calls through to the inner one,
	// which is itself a no-op extra acquisition.
	if _, alreadyWrapped := c.LogTo.(*syncWriter); !alreadyWrapped {
		c.LogTo = &syncWriter{w: c.LogTo}
	}
	if c.PIDFile == "" && c.DBPath != "" {
		c.PIDFile = filepath.Join(filepath.Dir(c.DBPath), "daemon.pid")
	}
	if c.SessionMonitor.PollInterval == 0 {
		c.SessionMonitor.PollInterval = 30 * time.Second
	}
	if c.SessionMonitor.EndedThreshold == 0 {
		c.SessionMonitor.EndedThreshold = 1 * time.Hour
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

	// sessionMonitor goroutine (ADR-012). Runs alongside the outbox
	// aggregator. Disabled=true skips entirely; otherwise tick at
	// SessionMonitor.PollInterval (default 30s).
	if !cfg.SessionMonitor.Disabled {
		store := sessions.NewStore(conn)
		go runSessionMonitor(ctx, store, cfg.SessionMonitor, cfg.LogTo)
		fmt.Fprintf(cfg.LogTo, "buddy: session monitor up (poll=%s ended_threshold=%s)\n",
			cfg.SessionMonitor.PollInterval, cfg.SessionMonitor.EndedThreshold)
	}

	// advisorMonitor goroutine (ADR-015). Polls at
	// Thresholds.PollInterval (default 1h) and persists fresh
	// advisories. Disabled=true skips entirely.
	if !cfg.Advisor.Disabled {
		go runAdvisorMonitor(ctx, conn, cfg.Advisor, cfg.LogTo)
		t := cfg.Advisor.Thresholds.WithDefaults()
		fmt.Fprintf(cfg.LogTo, "buddy: advisor monitor up (poll=%s dedup=%s)\n",
			t.PollInterval, t.DedupWindow)
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

// runSessionMonitor is the session-observer goroutine. It periodically fs-scans
// ~/.claude/projects/ via fsLister, then sweeps existing sessions for
// stale-vs-active state transitions (ended_at marker on / off).
//
// Errors are logged but never fatal — a corrupt transcript shouldn't
// kill the daemon. ctx cancellation exits cleanly.
func runSessionMonitor(ctx context.Context, store *sessions.Store, cfg SessionMonitorConfig, logTo io.Writer) {
	lister := sessions.NewFSLister(store)
	if cfg.ProjectsRoot != "" {
		lister.ProjectsRoot = cfg.ProjectsRoot
	}

	tick := func() {
		// Phase 1: fs scan + upsert.
		if _, err := lister.List(ctx); err != nil {
			fmt.Fprintf(logTo, "buddy: session monitor scan error: %v\n", err)
		}
		// Phase 2: sweep for stale-vs-active transitions.
		all, err := store.List(ctx, sessions.ListOptions{IncludeEnded: true})
		if err != nil {
			fmt.Fprintf(logTo, "buddy: session monitor sweep error: %v\n", err)
			return
		}
		now := time.Now().UTC()
		staleCutoff := now.Add(-cfg.EndedThreshold)
		for _, s := range all {
			isStale := s.LastActive.Before(staleCutoff)
			alreadyMarkedEnded := s.EndedAt != nil
			switch {
			case isStale && !alreadyMarkedEnded:
				// Newly stale → mark ended.
				t := now
				if err := store.SetEndedAt(ctx, s.ID, &t); err != nil {
					fmt.Fprintf(logTo, "buddy: session monitor: set ended_at %s: %v\n", s.ID, err)
				}
			case !isStale && alreadyMarkedEnded:
				// Resumed → clear ended marker.
				if err := store.SetEndedAt(ctx, s.ID, nil); err != nil {
					fmt.Fprintf(logTo, "buddy: session monitor: clear ended_at %s: %v\n", s.ID, err)
				}
			}
		}
	}

	// Initial tick + ticker.
	tick()
	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick()
		}
	}
}

// runAdvisorMonitor is the advisor goroutine. Polls the rule evaluator
// at Thresholds.PollInterval, persisting fired advisories (dedup
// applied via the advisor.Evaluator path). All errors are logged but
// non-fatal — a transient retrieval failure shouldn't kill the daemon.
func runAdvisorMonitor(ctx context.Context, conn *sql.DB, cfg AdvisorMonitorConfig, logTo io.Writer) {
	t := cfg.Thresholds.WithDefaults()
	runner := &advisor.Evaluator{
		Thresholds: t,
		Usage:      usage.NewService(conn),
		Sessions:   sessions.NewStore(conn),
		Knowledge:  knowledge.NewStore(conn),
		Embedder:   knowledge.NewPythonEmbedder(),
		Advisories: advisor.NewStore(conn),
	}

	// Build the notify dispatcher from cfg.NotifyChannels using the
	// daemon's *sql.DB (loadconfig stays connection-free).
	notifyDisp := buildNotifyDispatcher(conn, cfg.NotifyChannels)
	if notifyDisp != nil && len(cfg.NotifyChannels) > 0 {
		fmt.Fprintf(logTo, "buddy: notify dispatcher up (%d channels)\n", len(cfg.NotifyChannels))
	}

	tick := func() {
		advs, err := runner.Persist(ctx)
		if err != nil {
			fmt.Fprintf(logTo, "buddy: advisor monitor tick error: %v\n", err)
			return
		}
		if len(advs) > 0 {
			fmt.Fprintf(logTo, "buddy: advisor monitor wrote %d advisor(y/ies)\n", len(advs))
		}
		// Dispatch through notify channels right (ADR-016).
		// after persist. Returning a count map per channel so the
		// daemon log records "what got delivered where".
		if notifyDisp != nil && len(advs) > 0 {
			items := make([]notify.Notifiable, len(advs))
			for i, a := range advs {
				items[i] = a
			}
			sent := notifyDisp.Dispatch(ctx, items)
			for ch, n := range sent {
				if n > 0 {
					fmt.Fprintf(logTo, "buddy: notify dispatched %d via %s\n", n, ch)
				}
			}
		}
	}

	// Initial tick + ticker. Sleeping the full interval before the
	// first tick would feel laggy — users want immediate signal when
	// they start the daemon.
	tick()
	ticker := time.NewTicker(t.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick()
		}
	}
}

// buildNotifyDispatcher turns config-side NotifyChannelSpec entries
// into a wired *notify.Dispatcher. Returns nil when no channels are
// configured so the advisor tick path stays cheap.
func buildNotifyDispatcher(conn *sql.DB, specs []NotifyChannelSpec) *notify.Dispatcher {
	if len(specs) == 0 {
		return nil
	}
	disp := notify.NewDispatcher(notify.NewStore(conn))
	for _, sp := range specs {
		cc := notify.ChannelConfig{
			Enabled:     true,
			SeverityMin: notify.Severity(sp.SeverityMin),
			DedupWindow: sp.DedupWindow,
		}
		switch sp.Kind {
		case notify.ChannelDesktop:
			disp.AddChannel(notify.NewDesktopChannel(), cc)
		case notify.ChannelTUIBanner:
			disp.AddChannel(notify.NewTUIBannerChannel(), cc)
		case notify.ChannelShell:
			disp.AddChannel(notify.NewShellPromptChannel(), cc)
		case notify.ChannelWebhook:
			disp.AddChannel(notify.NewWebhookChannel(notify.WebhookConfig{
				URL: sp.URL, Method: sp.Method, Headers: sp.Headers, Timeout: sp.Timeout,
			}), cc)
		}
	}
	return disp
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
