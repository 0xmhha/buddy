package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestNewBuddyServer_RegistersExpectedToolset — the server constructor
// wires eight families: doctor, stats, the five feature_* CRUD tools,
// seven analytics_query_*, five usage_query_*, knowledge_query,
// usage_advise, and the two notify_* tools. Twenty-three tools in
// total. A future tool removal must be deliberate (delete the addX
// call AND update this list); a future tool addition is observable
// because this test will fail with the missing assertion.
func TestNewBuddyServer_RegistersExpectedToolset(t *testing.T) {
	t.Parallel()
	server := NewBuddyServer(Options{})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	got, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)

	expected := []string{
		"doctor",
		"stats",
		// feature CRUD
		"feature_list", "feature_get", "feature_upsert",
		"feature_delete", "feature_search",
		// analytics
		"analytics_query_funnel", "analytics_query_cohort",
		"analytics_query_ab_experiment", "analytics_query_actor_failure",
		"analytics_query_cost", "analytics_query_slo_burn",
		"analytics_query_feedback_corpus",
		// usage
		"usage_query_token_spend", "usage_query_session_stats",
		"usage_query_time_distribution", "usage_query_top_sessions",
		"usage_query_overview",
		// knowledge + advise
		"knowledge_query", "usage_advise",
		// notify
		"notify_status", "notify_test",
	}

	have := make(map[string]bool, len(got.Tools))
	for _, tool := range got.Tools {
		have[tool.Name] = true
	}
	for _, name := range expected {
		require.True(t, have[name], "expected tool %q missing from ListTools result", name)
	}
	require.Equal(t, len(expected), len(got.Tools),
		"unexpected tool count — adjust the expected slice if a tool was intentionally added or removed")
}
