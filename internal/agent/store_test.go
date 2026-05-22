package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestStore_LatestRun_NoRunsReturnsNotFound covers the empty-state path:
// an agent that exists but has never been run yields ErrNotFound, not a
// zero-value AgentRun (so `buddy agent log <id>` can distinguish "never
// ran" from "ran but no logs").
func TestStore_LatestRun_NoRunsReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	_, err = store.LatestRun(ctx, agent.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotFound), "expected ErrNotFound, got %v", err)
}

// TestStore_LatestRun_ReturnsMostRecentByID confirms LatestRun resolves
// multiple historical runs to the highest-id row, matching the
// SQLite rowid-alias monotonicity guarantee that the implementation
// relies on. Without this, `buddy agent log <id>` could surface stale
// logs from a prior run.
func TestStore_LatestRun_ReturnsMostRecentByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	first, err := store.StartRun(ctx, agent.ID)
	require.NoError(t, err)
	require.NoError(t, store.FinishRun(ctx, first, 0, nil, map[string]string{"tag": "first"}))

	second, err := store.StartRun(ctx, agent.ID)
	require.NoError(t, err)
	require.NoError(t, store.FinishRun(ctx, second, 0, nil, map[string]string{"tag": "second"}))

	got, err := store.LatestRun(ctx, agent.ID)
	require.NoError(t, err)
	require.Equal(t, second, got.ID, "LatestRun must pick the highest run id")
	require.Equal(t, agent.ID, got.AgentID)
	require.NotNil(t, got.EndedAt, "FinishRun should populate EndedAt")
	require.Contains(t, got.ResultJSON, `"tag":"second"`)
}

// TestStore_LatestRun_UnknownAgentReturnsNotFound makes sure a typo'd
// agent name doesn't return some other agent's latest run. The query
// is keyed by agent_id; an empty result must surface as ErrNotFound.
func TestStore_LatestRun_UnknownAgentReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	_, err := store.LatestRun(ctx, "no-such-agent")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotFound))
}

// ─── UpdateSpec (used by TUI in-app edit) ─────────────────────────────

// TestStore_UpdateSpec_SuccessReplacesEditableFields — name / schedule /
// spec_yaml are swapped in one UPDATE; status, created_at, last_run_at
// are preserved verbatim.
func TestStore_UpdateSpec_SuccessReplacesEditableFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	specOld, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	created, err := store.Create(ctx, specOld, minimalSpecYAML)
	require.NoError(t, err)

	const newYAML = `
id: hello-agent
name: "Hello agent (edited)"
schedule: "@hourly"
chain:
  - command: status
`
	specNew, err := ParseSpec([]byte(newYAML))
	require.NoError(t, err)
	require.NoError(t, store.UpdateSpec(ctx, specNew, newYAML))

	got, err := store.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "Hello agent (edited)", got.Name, "name must reflect the edit")
	require.Equal(t, "@hourly", got.Schedule, "schedule must reflect the edit")
	require.Contains(t, got.SpecYAML, "Hello agent (edited)", "spec_yaml must be the new bytes")
	require.Equal(t, StatusIdle, got.Status, "status must NOT be reset by an editorial change")
	require.WithinDuration(t, created.CreatedAt, got.CreatedAt, time.Second,
		"created_at must be preserved (the row was not re-created)")
	require.False(t, got.UpdatedAt.Before(created.UpdatedAt),
		"updated_at must advance (or at minimum not go backwards)")
}

// TestStore_UpdateSpec_UnknownIDReturnsNotFound — a typo'd id (or one
// that was deleted between the user pressing 'e' and the save round-trip)
// must surface as ErrNotFound, not a silent zero-row UPDATE.
func TestStore_UpdateSpec_UnknownIDReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	// Note: not creating the agent in the store first.
	err = store.UpdateSpec(ctx, spec, minimalSpecYAML)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotFound), "expected ErrNotFound, got %v", err)
}

// TestStore_UpdateSpec_PreservesLastRunAt — an agent with run history must
// keep its last_run_at after a spec edit (otherwise the list would lose
// its "X minutes ago" cue for editable rows).
func TestStore_UpdateSpec_PreservesLastRunAt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	// Manufacture a run + finish so last_run_at gets set.
	runID, err := store.StartRun(ctx, a.ID)
	require.NoError(t, err)
	require.NoError(t, store.FinishRun(ctx, runID, 0, nil, nil))
	require.NoError(t, store.UpdateStatus(ctx, a.ID, StatusDone))

	before, err := store.Get(ctx, a.ID)
	require.NoError(t, err)
	require.NotNil(t, before.LastRunAt, "precondition: last_run_at populated")

	// Edit the spec.
	const editedYAML = `
id: hello-agent
name: "Hello (edit 2)"
chain:
  - command: status
`
	editedSpec, err := ParseSpec([]byte(editedYAML))
	require.NoError(t, err)
	require.NoError(t, store.UpdateSpec(ctx, editedSpec, editedYAML))

	after, err := store.Get(ctx, a.ID)
	require.NoError(t, err)
	require.NotNil(t, after.LastRunAt, "last_run_at must survive UpdateSpec")
	require.Equal(t, before.LastRunAt.UnixMilli(), after.LastRunAt.UnixMilli(),
		"last_run_at value must be identical (preserved verbatim)")
}

// ─── LogsSince (used by TUI live log-tail) ────────────────────────────

// TestStore_LogsSince_ReturnsOnlyNewerLines covers the incremental-poll
// path: a sinceLogID > 0 must skip everything at or below that id and
// return only rows with strictly greater ids. Sort order is id ASC.
func TestStore_LogsSince_ReturnsOnlyNewerLines(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)
	runID, err := store.StartRun(ctx, a.ID)
	require.NoError(t, err)
	for _, msg := range []string{"first", "second", "third", "fourth"} {
		require.NoError(t, store.AppendLog(ctx, runID, "info", msg))
	}

	all, err := store.LogsSince(ctx, runID, 0)
	require.NoError(t, err)
	require.Len(t, all, 4, "sinceLogID=0 must fetch every line")
	cutoff := all[1].ID // skip first two

	got, err := store.LogsSince(ctx, runID, cutoff)
	require.NoError(t, err)
	require.Len(t, got, 2, "sinceLogID must be strictly less-than")
	require.Equal(t, "third", got[0].Message)
	require.Equal(t, "fourth", got[1].Message)
	require.Greater(t, got[0].ID, cutoff)
}

// TestStore_LogsSince_NoLogsReturnsEmpty — a run that hasn't produced any
// logs yet (or has been fully consumed) returns an empty slice + nil err.
// The TUI relies on this to render "(no log lines yet — run hasn't
// produced any)" without a separate error path.
func TestStore_LogsSince_NoLogsReturnsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)
	runID, err := store.StartRun(ctx, a.ID)
	require.NoError(t, err)

	got, err := store.LogsSince(ctx, runID, 0)
	require.NoError(t, err)
	require.Empty(t, got)
}

// TestStore_LogsSince_ScopedByRunID — log lines belonging to a *different*
// run must never leak into the result, even if their ids happen to fall
// above sinceLogID. Forgotten WHERE clauses on the run_id filter would
// surface here.
func TestStore_LogsSince_ScopedByRunID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)
	runA, err := store.StartRun(ctx, a.ID)
	require.NoError(t, err)
	require.NoError(t, store.AppendLog(ctx, runA, "info", "run A first"))

	runB, err := store.StartRun(ctx, a.ID)
	require.NoError(t, err)
	require.NoError(t, store.AppendLog(ctx, runB, "info", "run B first"))
	require.NoError(t, store.AppendLog(ctx, runB, "info", "run B second"))

	got, err := store.LogsSince(ctx, runA, 0)
	require.NoError(t, err)
	require.Len(t, got, 1, "must not include lines from other runs")
	require.Equal(t, "run A first", got[0].Message)
}

// ─── retention (F5 follow-on) ──────────────────────────────────────────

// TestStore_CountRunsBefore_OnlyCountsFinishedAndOldEnough verifies the
// retention preview path: in-flight runs (ended_at IS NULL) are never
// included, and the cutoff is strictly less-than (a run that ended
// exactly at the cutoff is preserved).
func TestStore_CountRunsBefore_OnlyCountsFinishedAndOldEnough(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, conn := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	// Three runs:
	//   r1 — finished 10 minutes ago     (old, finished)
	//   r2 — finished 1 minute ago       (recent, finished)
	//   r3 — still running (ended_at NULL, even if started 10m ago)
	now := time.Now()
	insertRun := func(endedAt *time.Time) int64 {
		t.Helper()
		var endedMs any
		if endedAt != nil {
			endedMs = endedAt.UnixMilli()
		}
		res, err := conn.ExecContext(ctx,
			`INSERT INTO agent_runs (agent_id, started_at, ended_at, exit_code) VALUES (?, ?, ?, 0)`,
			agent.ID, now.Add(-15*time.Minute).UnixMilli(), endedMs)
		require.NoError(t, err)
		id, err := res.LastInsertId()
		require.NoError(t, err)
		return id
	}
	r1End := now.Add(-10 * time.Minute)
	insertRun(&r1End)
	r2End := now.Add(-1 * time.Minute)
	insertRun(&r2End)
	insertRun(nil) // still running

	cutoff := now.Add(-5 * time.Minute)
	count, err := store.CountRunsBefore(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, int64(1), count,
		"only the run that ended >5m ago should be counted; the in-flight run and the recent finish are excluded")
}

// TestStore_PurgeRunsBefore_CascadesToLogs verifies the actual delete
// path: rows are removed, the FK ON DELETE CASCADE drops the matching
// agent_logs rows transactionally, and in-flight runs survive.
func TestStore_PurgeRunsBefore_CascadesToLogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, conn := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	now := time.Now()
	insertRunWithLogs := func(endedAt *time.Time, logCount int) int64 {
		t.Helper()
		var endedMs any
		if endedAt != nil {
			endedMs = endedAt.UnixMilli()
		}
		res, err := conn.ExecContext(ctx,
			`INSERT INTO agent_runs (agent_id, started_at, ended_at, exit_code) VALUES (?, ?, ?, 0)`,
			agent.ID, now.Add(-15*time.Minute).UnixMilli(), endedMs)
		require.NoError(t, err)
		id, err := res.LastInsertId()
		require.NoError(t, err)
		for i := 0; i < logCount; i++ {
			require.NoError(t, store.AppendLog(ctx, id, "info", "test log line"))
		}
		return id
	}
	oldEnd := now.Add(-30 * time.Minute)
	oldID := insertRunWithLogs(&oldEnd, 5)
	recentEnd := now.Add(-1 * time.Minute)
	recentID := insertRunWithLogs(&recentEnd, 3)
	inflightID := insertRunWithLogs(nil, 2)

	// Sanity — three runs + 10 log rows before purge.
	var logsBefore int64
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM agent_logs`).Scan(&logsBefore))
	require.Equal(t, int64(10), logsBefore)

	cutoff := now.Add(-5 * time.Minute)
	deleted, err := store.PurgeRunsBefore(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted, "only the old finished run should be deleted")

	// The old run is gone, the recent finish + in-flight survive.
	var oldExists, recentExists, inflightExists int64
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM agent_runs WHERE id = ?`, oldID).Scan(&oldExists))
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM agent_runs WHERE id = ?`, recentID).Scan(&recentExists))
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM agent_runs WHERE id = ?`, inflightID).Scan(&inflightExists))
	require.Equal(t, int64(0), oldExists)
	require.Equal(t, int64(1), recentExists)
	require.Equal(t, int64(1), inflightExists)

	// FK CASCADE dropped the 5 log rows for the old run; 3 + 2 remain.
	var logsAfter int64
	require.NoError(t, conn.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM agent_logs`).Scan(&logsAfter))
	require.Equal(t, int64(5), logsAfter,
		"FK CASCADE should have dropped the 5 log rows belonging to the deleted run")
}

// TestStore_PurgeRunsBefore_EmptyDBIsNoop confirms purging an empty
// store returns (0, nil) — not an error — so callers can call it on a
// fresh install without special-casing.
func TestStore_PurgeRunsBefore_EmptyDBIsNoop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	deleted, err := store.PurgeRunsBefore(ctx, time.Now())
	require.NoError(t, err)
	require.Equal(t, int64(0), deleted)
}
