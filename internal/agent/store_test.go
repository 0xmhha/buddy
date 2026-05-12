package agent

import (
	"context"
	"errors"
	"testing"

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
