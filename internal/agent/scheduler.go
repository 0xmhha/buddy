package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler drives agents that have a non-empty `schedule` cron expression.
// It is the W3-3 follow-on to the on-demand `buddy agent run` CLI: each tick
// fires Runtime.Run for one agent. v0.3.x ships single-process / sequential
// dispatch — multi-process clustering or parallel runs land in later cycles.
//
// Concurrency contract:
//   - The cron library invokes job callbacks in its own goroutines. Our
//     callback wraps Runtime.Run inside a per-agent "is this agent already
//     running?" check (atomic flag) so overlapping ticks are dropped rather
//     than piled up. This matches the typical user expectation of "don't
//     pile up if last run is still going."
//   - The Scheduler does not refresh the agent set live. Adding / deleting an
//     agent while the scheduler is running has no effect until restart.
//     Live refresh (W3-3 follow-on follow-on) is a separate cycle.
type Scheduler struct {
	store    *Store
	runtime  *Runtime
	cron     *cron.Cron
	location *time.Location
	// running tracks which agent IDs currently have a Run in flight so a
	// tick fired while the previous run is still going is dropped.
	running sync.Map // map[string]*atomic.Bool
	// logger is where lifecycle events (load N agents, schedule X with cron Y,
	// skipped overlap, etc.) go. Nil = io.Discard. CLI passes os.Stderr.
	logger io.Writer
}

// SchedulerOptions tunes the Scheduler. Zero-value is fine — Location defaults
// to time.Local so users with `schedule: "0 3 * * *"` get 03:00 their wall
// clock. Explicitly set Location for tests / multi-tz deployments.
type SchedulerOptions struct {
	Location *time.Location // nil → time.Local
	Logger   io.Writer      // nil → io.Discard
}

// NewScheduler wires a Scheduler against an open Store + Runtime.
func NewScheduler(store *Store, runtime *Runtime, opts SchedulerOptions) *Scheduler {
	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}
	logger := opts.Logger
	if logger == nil {
		logger = io.Discard
	}
	// Tag the cron parser with "Descriptor" so @daily / @every 30s / @hourly
	// all work in addition to standard 5-field expressions. SecondOptional
	// is *not* enabled — the agent schedule field is minute-precision by
	// design (matches every common cron tutorial). Users that need
	// finer-grained cadence should run a wrapper agent on an event hook.
	return &Scheduler{
		store:    store,
		runtime:  runtime,
		cron:     cron.New(cron.WithLocation(loc)),
		location: loc,
		logger:   logger,
	}
}

// Load enumerates the store, validates each agent's schedule, and registers
// the ones with valid cron expressions. Returns the count loaded plus a
// non-fatal "skipped" slice describing rejected agents (the caller logs
// these so users notice a typo without an extra round-trip).
func (s *Scheduler) Load(ctx context.Context) (loaded int, skipped []SkippedAgent, err error) {
	agents, err := s.store.List(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("scheduler: list agents: %w", err)
	}
	for _, a := range agents {
		if a.Schedule == "" {
			continue // on-demand only — explicitly opt-out
		}
		entryID, addErr := s.cron.AddFunc(a.Schedule, s.makeJob(a))
		if addErr != nil {
			skipped = append(skipped, SkippedAgent{ID: a.ID, Schedule: a.Schedule, Err: addErr})
			continue
		}
		loaded++
		fmt.Fprintf(s.logger, "scheduler: agent %q scheduled (cron=%q entry_id=%d)\n", a.ID, a.Schedule, entryID)
	}
	return loaded, skipped, nil
}

// SkippedAgent is one rejected entry from Load.
type SkippedAgent struct {
	ID       string
	Schedule string
	Err      error
}

// Start begins the cron loop and blocks until ctx is cancelled. On cancel,
// in-flight jobs are allowed to finish (cron.Stop's contract).
func (s *Scheduler) Start(ctx context.Context) error {
	s.cron.Start()
	fmt.Fprintf(s.logger, "scheduler: started (location=%s, entries=%d)\n", s.location, len(s.cron.Entries()))
	<-ctx.Done()
	stopCtx := s.cron.Stop()
	fmt.Fprintln(s.logger, "scheduler: stopping (waiting for in-flight jobs)...")
	<-stopCtx.Done()
	fmt.Fprintln(s.logger, "scheduler: stopped")
	return nil
}

// Entries exposes the cron library's scheduled entries (for `buddy agent
// scheduler status` and tests). One entry per loaded agent.
func (s *Scheduler) Entries() []cron.Entry {
	return s.cron.Entries()
}

// makeJob captures one agent ID into a cron-compatible func(). The closure
// re-fetches the agent on each tick so spec edits via the CLI (after
// restart-and-reload) take effect on the next tick — though in v0.3 the
// Scheduler does not auto-reload between ticks, this still helps when a
// future live-refresh implementation lands.
func (s *Scheduler) makeJob(initial Agent) func() {
	return func() {
		// Per-agent "is this agent already running?" guard.
		flagAny, _ := s.running.LoadOrStore(initial.ID, &atomic.Bool{})
		flag := flagAny.(*atomic.Bool)
		if !flag.CompareAndSwap(false, true) {
			fmt.Fprintf(s.logger, "scheduler: agent %q tick skipped — previous run still in flight\n", initial.ID)
			return
		}
		defer flag.Store(false)

		// Fresh context per job, decoupled from Start's ctx. The scheduler
		// only honours ctx for *waking up*; in-flight jobs use their own
		// timeout so a stop-then-restart cycle does not cut a job in half.
		// 10-minute default — long enough for a multi-step chain but
		// finite for safety.
		jobCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// Re-fetch the agent in case the spec was updated since startup
		// (future-proofing — v0.3 CLI does not yet support edit).
		current, err := s.store.Get(jobCtx, initial.ID)
		if errors.Is(err, ErrNotFound) {
			fmt.Fprintf(s.logger, "scheduler: agent %q vanished — tick dropped\n", initial.ID)
			return
		}
		if err != nil {
			fmt.Fprintf(s.logger, "scheduler: agent %q reload failed: %v\n", initial.ID, err)
			return
		}

		fmt.Fprintf(s.logger, "scheduler: agent %q tick starting\n", current.ID)
		res, runErr := s.runtime.Run(jobCtx, current)
		if runErr != nil {
			fmt.Fprintf(s.logger, "scheduler: agent %q tick errored: %v\n", current.ID, runErr)
			return
		}
		fmt.Fprintf(s.logger, "scheduler: agent %q tick finished run_id=%d exit=%d\n", current.ID, res.RunID, res.ExitCode)
	}
}
