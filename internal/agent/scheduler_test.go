package agent

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// safeBuf is a thread-safe wrapper around bytes.Buffer for the scheduler's
// log output. Cron job callbacks run on their own goroutines, so the test
// main goroutine cannot read from a raw bytes.Buffer without racing.
type safeBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// compile-time check that *safeBuf still satisfies io.Writer.
var _ io.Writer = (*safeBuf)(nil)

// scheduledSpecYAML uses a 1-second period because robfig/cron's
// ConstantDelaySchedule.Next() rounds nanoseconds to the next whole
// second — sub-second @every values fire erratically. Anything that
// needs sub-second granularity tests the in-flight guard directly
// (see TestScheduler_MakeJob_OverlapGuardDropsConcurrentTick).
const scheduledSpecYAML = `
id: scheduler-test
name: "Scheduler test agent"
schedule: "@every 1s"
chain:
  - command: status
`

// TestScheduler_LoadRegistersAgentsWithSchedule asserts that Load only picks
// up agents whose schedule field is non-empty + parseable. On-demand agents
// (empty schedule) stay invisible to the cron tick.
func TestScheduler_LoadRegistersAgentsWithSchedule(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	// Two agents: one scheduled, one on-demand.
	scheduledSpec, err := ParseSpec([]byte(scheduledSpecYAML))
	require.NoError(t, err)
	_, err = store.Create(ctx, scheduledSpec, scheduledSpecYAML)
	require.NoError(t, err)

	onDemandSpec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	_, err = store.Create(ctx, onDemandSpec, minimalSpecYAML)
	require.NoError(t, err)

	rt := NewRuntime(store, NewMockExecutor())
	sched := NewScheduler(store, rt, SchedulerOptions{})
	loaded, skipped, err := sched.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, loaded, "only the agent with non-empty schedule should load")
	require.Empty(t, skipped)
	require.Len(t, sched.Entries(), 1)
}

// TestScheduler_LoadReportsInvalidCron pins the contract that a malformed
// schedule does not abort the rest — it's reported via SkippedAgent so the
// user can fix the typo without blocking other agents.
func TestScheduler_LoadReportsInvalidCron(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	badYAML := `
id: bad-cron-agent
name: "Bad cron"
schedule: "totally not cron"
chain:
  - command: status
`
	spec, err := ParseSpec([]byte(badYAML))
	require.NoError(t, err)
	_, err = store.Create(ctx, spec, badYAML)
	require.NoError(t, err)

	goodSpec, err := ParseSpec([]byte(scheduledSpecYAML))
	require.NoError(t, err)
	_, err = store.Create(ctx, goodSpec, scheduledSpecYAML)
	require.NoError(t, err)

	rt := NewRuntime(store, NewMockExecutor())
	sched := NewScheduler(store, rt, SchedulerOptions{})
	loaded, skipped, err := sched.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, loaded, "good agent must still load")
	require.Len(t, skipped, 1)
	require.Equal(t, "bad-cron-agent", skipped[0].ID)
	require.Equal(t, "totally not cron", skipped[0].Schedule)
}

// TestScheduler_StartFiresRunsAndStops drives the full Start → tick → Stop
// lifecycle with a 200ms cron expression. After 500ms we expect ≥ 1 run
// recorded and Start to return after ctx cancellation.
func TestScheduler_StartFiresRunsAndStops(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, conn := newTestStore(t)

	spec, err := ParseSpec([]byte(scheduledSpecYAML))
	require.NoError(t, err)
	_, err = store.Create(ctx, spec, scheduledSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: "ok", ExitCode: 0}
	rt := NewRuntime(store, mock)

	logs := &safeBuf{}
	sched := NewScheduler(store, rt, SchedulerOptions{Logger: logs})
	loaded, _, err := sched.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, loaded)

	// @every 1s + 2.5s ctx → expect ≥1 tick before stop. We allow up to
	// 5s for cron's whole-second rounding to align with wall-clock seconds
	// when this test starts mid-second.
	runCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- sched.Start(runCtx) }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler.Start did not return within 5s of ctx cancel")
	}

	var runCount int64
	require.NoError(t, conn.QueryRow(`SELECT COUNT(*) FROM agent_runs WHERE agent_id = ?`, "scheduler-test").Scan(&runCount))
	require.GreaterOrEqual(t, runCount, int64(1), "expected ≥1 scheduled run, logs:\n%s", logs.String())
	require.True(t, strings.Contains(logs.String(), "tick starting"), "scheduler should log tick starting; got:\n%s", logs.String())
}

// TestScheduler_MakeJob_OverlapGuardDropsConcurrentTick exercises the
// per-agent in-flight guard directly. We invoke the job func twice in
// parallel goroutines — the second call's CompareAndSwap must fail and
// the call must return early, leaving exactly one agent_runs row.
//
// Direct-job test rather than via cron because robfig/cron's
// ConstantDelaySchedule rounds Next() to whole seconds, which makes
// fast (sub-second) overlap testing erratic. The guard itself is
// timing-independent.
func TestScheduler_MakeJob_OverlapGuardDropsConcurrentTick(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, conn := newTestStore(t)

	spec, err := ParseSpec([]byte(scheduledSpecYAML))
	require.NoError(t, err)
	created, err := store.Create(ctx, spec, scheduledSpecYAML)
	require.NoError(t, err)

	// Slow enough that two concurrent jobs *would* overlap if the guard
	// weren't there. 300ms is plenty for goroutines to enter the closure.
	slow := newSlowExecutor(300 * time.Millisecond)
	rt := NewRuntime(store, slow)

	logs := &safeBuf{}
	sched := NewScheduler(store, rt, SchedulerOptions{Logger: logs})
	job := sched.makeJob(created)

	// Fire two ticks back-to-back, in parallel.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); job() }()
	// Brief stagger so goroutine 1 wins the CAS deterministically.
	time.Sleep(20 * time.Millisecond)
	go func() { defer wg.Done(); job() }()
	wg.Wait()

	var runCount int64
	require.NoError(t, conn.QueryRow(`SELECT COUNT(*) FROM agent_runs`).Scan(&runCount))
	require.Equal(t, int64(1), runCount,
		"overlap guard must drop the second concurrent tick; logs:\n%s", logs.String())
	require.True(t, strings.Contains(logs.String(), "skipped"),
		"expected overlap-skip log line; got:\n%s", logs.String())
	require.Equal(t, int64(1), slow.calls.Load(),
		"slow executor should only see one call; guard dropped the other")
}

// TestScheduler_StartReturnsImmediatelyOnCancelledContext sanity-checks the
// shutdown path: an already-cancelled ctx must not deadlock Start.
func TestScheduler_StartReturnsImmediatelyOnCancelledContext(t *testing.T) {
	t.Parallel()
	store, _ := newTestStore(t)
	rt := NewRuntime(store, NewMockExecutor())
	sched := NewScheduler(store, rt, SchedulerOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, sched.Start(ctx))
}

// ─── live refresh (Tier 1.6) ────────────────────────────────────────────

// schedSpecYAMLWith builds a unique scheduler-test YAML with a custom ID +
// schedule so refresh tests can register multiple agents in one store.
func schedSpecYAMLWith(id, schedule string) string {
	return "id: " + id + "\nname: \"refresh test\"\nschedule: \"" + schedule + "\"\nchain:\n  - command: status\n"
}

// waitForEntries polls sched.Entries() until len() == want, or fails the
// test after deadline. Used by refresh tests to avoid sleeping a fixed
// long duration when the scheduler usually picks up the change in 2-3
// poll ticks.
func waitForEntries(t *testing.T, sched *Scheduler, want int, deadline time.Duration) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		if len(sched.Entries()) == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected %d entries within %s, got %d", want, deadline, len(sched.Entries()))
}

// TestScheduler_RefreshPicksUpNewAgent verifies that an agent created in
// the store *after* Start launches the polling loop ends up in the cron
// table on the next refresh tick — no scheduler restart required.
func TestScheduler_RefreshPicksUpNewAgent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, _ := newTestStore(t)
	rt := NewRuntime(store, NewMockExecutor())

	logs := &safeBuf{}
	sched := NewScheduler(store, rt, SchedulerOptions{
		Logger:          logs,
		RefreshInterval: 20 * time.Millisecond,
	})

	// Start with zero agents — Entries() must be 0 at startup.
	_, _, err := sched.Load(ctx)
	require.NoError(t, err)
	require.Empty(t, sched.Entries())

	startDone := make(chan struct{})
	go func() { _ = sched.Start(ctx); close(startDone) }()

	// Now create a scheduled agent. Use @every 1h so the entry is
	// registered but never actually fires inside the test window — we
	// care about registration, not execution.
	spec, err := ParseSpec([]byte(schedSpecYAMLWith("late-agent", "@every 1h")))
	require.NoError(t, err)
	_, err = store.Create(ctx, spec, schedSpecYAMLWith("late-agent", "@every 1h"))
	require.NoError(t, err)

	waitForEntries(t, sched, 1, 500*time.Millisecond)
	require.Contains(t, logs.String(), `agent "late-agent" scheduled`,
		"refresh poll should log the new agent registration")

	cancel()
	<-startDone
}

// TestScheduler_RefreshRemovesDeletedAgent verifies that deleting an
// agent in the store while the scheduler is running drops its cron
// entry on the next refresh tick.
func TestScheduler_RefreshRemovesDeletedAgent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, _ := newTestStore(t)
	rt := NewRuntime(store, NewMockExecutor())

	spec, err := ParseSpec([]byte(schedSpecYAMLWith("doomed-agent", "@every 1h")))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, schedSpecYAMLWith("doomed-agent", "@every 1h"))
	require.NoError(t, err)

	logs := &safeBuf{}
	sched := NewScheduler(store, rt, SchedulerOptions{
		Logger:          logs,
		RefreshInterval: 20 * time.Millisecond,
	})
	_, _, err = sched.Load(ctx)
	require.NoError(t, err)
	require.Len(t, sched.Entries(), 1, "agent must be registered at startup")

	startDone := make(chan struct{})
	go func() { _ = sched.Start(ctx); close(startDone) }()

	require.NoError(t, store.Delete(ctx, agent.ID))

	waitForEntries(t, sched, 0, 500*time.Millisecond)
	require.Contains(t, logs.String(), `agent "doomed-agent" unscheduled`,
		"refresh poll should log the removal")

	cancel()
	<-startDone
}

// TestScheduler_RefreshDetectsScheduleChange verifies that updating an
// agent's schedule (via direct DB UPDATE — Store does not expose
// UpdateSchedule yet) replaces the cron entry on the next refresh tick
// rather than leaving the stale entry firing on the previous cadence.
func TestScheduler_RefreshDetectsScheduleChange(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, conn := newTestStore(t)
	rt := NewRuntime(store, NewMockExecutor())

	spec, err := ParseSpec([]byte(schedSpecYAMLWith("mutable-agent", "@every 1h")))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, schedSpecYAMLWith("mutable-agent", "@every 1h"))
	require.NoError(t, err)

	logs := &safeBuf{}
	sched := NewScheduler(store, rt, SchedulerOptions{
		Logger:          logs,
		RefreshInterval: 20 * time.Millisecond,
	})
	_, _, err = sched.Load(ctx)
	require.NoError(t, err)
	beforeID := sched.Entries()[0].ID
	require.NotZero(t, beforeID)

	startDone := make(chan struct{})
	go func() { _ = sched.Start(ctx); close(startDone) }()

	// Change the schedule out-of-band — simulates another process editing
	// the spec via a future `buddy agent edit` subcommand or a manual
	// sqlite3 edit during local dev.
	_, err = conn.ExecContext(ctx,
		`UPDATE agents SET schedule = ? WHERE id = ?`,
		"@every 2h", agent.ID)
	require.NoError(t, err)

	// Poll for the entry-id swap. We can't compare entries() count (still
	// 1) so we check the tracked entry id changed.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		entries := sched.Entries()
		if len(entries) == 1 && entries[0].ID != beforeID {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	require.Len(t, sched.Entries(), 1)
	require.NotEqual(t, beforeID, sched.Entries()[0].ID,
		"schedule change should re-register the entry under a new cron entry id")
	require.Contains(t, logs.String(), `agent "mutable-agent" reschedule`,
		"refresh poll should log the reschedule")

	cancel()
	<-startDone
}

// TestScheduler_RefreshDisabledKeepsInitialEntries confirms the opt-out:
// with RefreshDisabled=true the scheduler ignores DB changes after
// Start and never spawns the polling goroutine.
func TestScheduler_RefreshDisabledKeepsInitialEntries(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, _ := newTestStore(t)
	rt := NewRuntime(store, NewMockExecutor())

	logs := &safeBuf{}
	sched := NewScheduler(store, rt, SchedulerOptions{
		Logger:          logs,
		RefreshDisabled: true,
	})
	_, _, err := sched.Load(ctx)
	require.NoError(t, err)
	require.Empty(t, sched.Entries())

	startDone := make(chan struct{})
	go func() { _ = sched.Start(ctx); close(startDone) }()

	// Create an agent that *would* be picked up if refresh were active.
	spec, err := ParseSpec([]byte(schedSpecYAMLWith("ghost-agent", "@every 1h")))
	require.NoError(t, err)
	_, err = store.Create(ctx, spec, schedSpecYAMLWith("ghost-agent", "@every 1h"))
	require.NoError(t, err)

	// Give the scheduler more than enough time to *not* refresh.
	time.Sleep(80 * time.Millisecond)
	require.Empty(t, sched.Entries(),
		"with RefreshDisabled the scheduler must ignore post-startup store changes")
	require.Contains(t, logs.String(), "refresh=disabled",
		"startup log should reflect the disabled state")

	cancel()
	<-startDone
}

// ─── slowExecutor — Executor that blocks for a fixed duration ─────────────

type slowExecutor struct {
	delay time.Duration
	calls atomic.Int64
}

func newSlowExecutor(delay time.Duration) *slowExecutor {
	return &slowExecutor{delay: delay}
}

func (s *slowExecutor) Run(ctx context.Context, _ string, _ string, _ LogSink) (string, string, int, error) {
	s.calls.Add(1)
	select {
	case <-ctx.Done():
		return "", "", -1, ctx.Err()
	case <-time.After(s.delay):
		return "ok", "", 0, nil
	}
}
