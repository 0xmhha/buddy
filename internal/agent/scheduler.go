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
// It is the background-tick counterpart to the on-demand `buddy agent run`
// CLI: each tick fires Runtime.Run for one agent. Single-process /
// sequential dispatch ships first; multi-process clustering or parallel
// runs land in later cycles.
//
// Concurrency contract:
//   - The cron library invokes job callbacks in its own goroutines. Our
//     callback wraps Runtime.Run inside a per-agent "is this agent already
//     running?" check (atomic flag) so overlapping ticks are dropped rather
//     than piled up. This matches the typical user expectation of "don't
//     pile up if last run is still going."
//   - The Scheduler does not refresh the agent set live. Adding / deleting an
//     agent while the scheduler is running has no effect until restart.
//     Live refresh is a separate cycle.
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

	// refreshInterval is the cadence at which the scheduler re-reads the
	// store and reconciles its in-memory cron entries against what the DB
	// says. Zero means refresh is disabled and the scheduler keeps the
	// entry set it loaded at startup.
	refreshInterval time.Duration

	// tracked maps agent.ID → the cron entry ID we registered for it.
	// Used by refreshOnce to diff "what's in the DB" against "what's in
	// the cron table" and apply add / remove / update accordingly. Guarded
	// by trackedMu — refresh polls run in a separate goroutine and may
	// race with manual Load calls in tests.
	trackedMu sync.Mutex
	tracked   map[string]trackedEntry
}

// trackedEntry remembers which schedule string maps to which cron entry.
// We need the schedule on hand to detect "the spec stayed in the DB but
// the schedule string changed" — that path removes the old entry and adds
// a new one rather than leaving a stale tick cadence in place.
type trackedEntry struct {
	entryID  cron.EntryID
	schedule string
}

// SchedulerOptions tunes the Scheduler. Zero-value is fine — Location defaults
// to time.Local so users with `schedule: "0 3 * * *"` get 03:00 their wall
// clock. Explicitly set Location for tests / multi-tz deployments.
type SchedulerOptions struct {
	Location *time.Location // nil → time.Local
	Logger   io.Writer      // nil → io.Discard
	// RefreshInterval controls how often the scheduler polls the store to
	// pick up agents added / deleted / re-scheduled by another process
	// (e.g. a separate `buddy agent create` shell). Zero means "use the
	// production default of 60s"; pass a small value in tests to drive the
	// loop fast. To turn live refresh off entirely, set RefreshDisabled.
	RefreshInterval time.Duration
	// RefreshDisabled, when true, loads once at startup and never re-reads
	// the store. Mainly an escape hatch for users on a single shell who
	// edit specs via restart anyway.
	RefreshDisabled bool
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
	// Resolve refresh cadence. RefreshDisabled wins; otherwise 0 means
	// "use the 60s production default" and a positive value is honoured
	// verbatim (tests want 50ms-ish; CLI users typically want 1m).
	refreshInterval := opts.RefreshInterval
	if opts.RefreshDisabled {
		refreshInterval = 0
	} else if refreshInterval <= 0 {
		refreshInterval = 60 * time.Second
	}
	// Tag the cron parser with "Descriptor" so @daily / @every 30s / @hourly
	// all work in addition to standard 5-field expressions. SecondOptional
	// is *not* enabled — the agent schedule field is minute-precision by
	// design (matches every common cron tutorial). Users that need
	// finer-grained cadence should run a wrapper agent on an event hook.
	return &Scheduler{
		store:           store,
		runtime:         runtime,
		cron:            cron.New(cron.WithLocation(loc)),
		location:        loc,
		logger:          logger,
		refreshInterval: refreshInterval,
		tracked:         map[string]trackedEntry{},
	}
}

// Load enumerates the store, validates each agent's schedule, and registers
// the ones with valid cron expressions. Returns the count loaded plus a
// non-fatal "skipped" slice describing rejected agents (the caller logs
// these so users notice a typo without an extra round-trip). When live
// refresh is enabled, Load is also called implicitly by the first
// refreshOnce tick — calling it manually before Start is harmless and
// gives deterministic ordering for tests / CLI code that read
// `Entries()` before launching the cron loop.
func (s *Scheduler) Load(ctx context.Context) (loaded int, skipped []SkippedAgent, err error) {
	diff, err := s.refreshOnce(ctx)
	if err != nil {
		return 0, nil, err
	}
	return diff.added, diff.skipped, nil
}

// refreshDiff is what refreshOnce hands back to its caller — the number
// of agents added / removed / updated in the cron table this pass, plus
// any spec rows that failed validation (bad cron string) so the caller
// can log them at startup vs ignore them on later polls.
type refreshDiff struct {
	added   int
	removed int
	updated int
	skipped []SkippedAgent
}

// refreshOnce reconciles the in-memory cron entries against `store.List()`.
// New scheduled agents get AddFunc'd, deleted ones get Remove'd, and ones
// whose schedule string changed get removed + re-added.
//
// Locks trackedMu for the whole pass so a parallel Start->pollLoop tick
// can't race with a manual Load call. Cron's own Add/Remove are
// internally synchronised so we don't need to wrap them ourselves.
func (s *Scheduler) refreshOnce(ctx context.Context) (refreshDiff, error) {
	agents, err := s.store.List(ctx)
	if err != nil {
		return refreshDiff{}, fmt.Errorf("scheduler: list agents: %w", err)
	}

	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()

	var diff refreshDiff
	seen := make(map[string]struct{}, len(agents))

	for _, a := range agents {
		if a.Schedule == "" {
			continue // on-demand only — explicitly opt-out
		}
		seen[a.ID] = struct{}{}

		prev, alreadyRegistered := s.tracked[a.ID]
		if alreadyRegistered && prev.schedule == a.Schedule {
			continue // unchanged — keep the existing cron entry
		}
		if alreadyRegistered {
			// Schedule changed — drop the old entry before re-adding so
			// stale ticks don't keep firing on the previous cadence.
			s.cron.Remove(prev.entryID)
			delete(s.tracked, a.ID)
		}

		entryID, addErr := s.cron.AddFunc(a.Schedule, s.makeJob(a))
		if addErr != nil {
			diff.skipped = append(diff.skipped, SkippedAgent{ID: a.ID, Schedule: a.Schedule, Err: addErr})
			continue
		}
		s.tracked[a.ID] = trackedEntry{entryID: entryID, schedule: a.Schedule}
		if alreadyRegistered {
			diff.updated++
			fmt.Fprintf(s.logger, "scheduler: agent %q reschedule (cron=%q entry_id=%d)\n", a.ID, a.Schedule, entryID)
		} else {
			diff.added++
			fmt.Fprintf(s.logger, "scheduler: agent %q scheduled (cron=%q entry_id=%d)\n", a.ID, a.Schedule, entryID)
		}
	}

	// Any tracked agent that didn't show up in `seen` was deleted from the
	// store; drop its cron entry too.
	for id, prev := range s.tracked {
		if _, ok := seen[id]; ok {
			continue
		}
		s.cron.Remove(prev.entryID)
		delete(s.tracked, id)
		diff.removed++
		fmt.Fprintf(s.logger, "scheduler: agent %q unscheduled (entry_id=%d removed from cron)\n", id, prev.entryID)
	}

	return diff, nil
}

// SkippedAgent is one rejected entry from Load.
type SkippedAgent struct {
	ID       string
	Schedule string
	Err      error
}

// Start begins the cron loop and blocks until ctx is cancelled. On cancel,
// in-flight jobs are allowed to finish (cron.Stop's contract).
//
// When refresh is enabled (the default), Start also launches a polling
// goroutine that re-reads the store every refreshInterval and reconciles
// the cron entry set against the DB. Adds, deletes, and schedule
// changes made by another process (e.g. a separate `buddy agent create`
// shell) are picked up on the next tick rather than requiring a
// restart.
func (s *Scheduler) Start(ctx context.Context) error {
	s.cron.Start()
	fmt.Fprintf(s.logger, "scheduler: started (location=%s, entries=%d, refresh=%s)\n",
		s.location, len(s.cron.Entries()), formatRefreshLabel(s.refreshInterval))

	if s.refreshInterval > 0 {
		go s.pollLoop(ctx)
	}

	<-ctx.Done()
	stopCtx := s.cron.Stop()
	fmt.Fprintln(s.logger, "scheduler: stopping (waiting for in-flight jobs)...")
	<-stopCtx.Done()
	fmt.Fprintln(s.logger, "scheduler: stopped")
	return nil
}

// pollLoop is the live-refresh goroutine. It fires refreshOnce on every
// tick of refreshInterval and exits when ctx is cancelled. Errors from
// refreshOnce are logged but do not stop the loop — a transient DB read
// failure should not freeze the scheduled-agent set.
func (s *Scheduler) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			diff, err := s.refreshOnce(ctx)
			if err != nil {
				fmt.Fprintf(s.logger, "scheduler: refresh poll errored: %v\n", err)
				continue
			}
			if diff.added > 0 || diff.removed > 0 || diff.updated > 0 {
				fmt.Fprintf(s.logger,
					"scheduler: refresh poll applied (+%d / -%d / ~%d)\n",
					diff.added, diff.removed, diff.updated)
			}
			for _, sk := range diff.skipped {
				fmt.Fprintf(s.logger,
					"scheduler: refresh poll skipped agent %q (cron=%q): %v\n",
					sk.ID, sk.Schedule, sk.Err)
			}
		}
	}
}

// formatRefreshLabel renders the refresh cadence for the startup log.
// Zero is special-cased to "disabled" so users skimming logs see the
// intent rather than "0s".
func formatRefreshLabel(d time.Duration) string {
	if d <= 0 {
		return "disabled"
	}
	return d.String()
}

// Entries exposes the cron library's scheduled entries (for `buddy agent
// scheduler status` and tests). One entry per loaded agent.
func (s *Scheduler) Entries() []cron.Entry {
	return s.cron.Entries()
}

// makeJob captures one agent ID into a cron-compatible func(). The closure
// re-fetches the agent on each tick so spec edits via the CLI take effect
// on the next tick.
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
		// (future-proofing — the CLI does not yet support edit).
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
