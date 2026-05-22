package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/0xmhha/buddy/internal/analytics"
)

// TestAnalyticsTools_Registered verifies the 7 analytics_query_* tools land in
// the server's tool list after NewBuddyServer wires them up. Catches regressions
// where addAnalyticsTools is removed from server.go or a tool name drifts.
func TestAnalyticsTools_Registered(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	server := NewBuddyServer(Options{})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	got, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)

	have := make(map[string]bool, len(got.Tools))
	for _, tool := range got.Tools {
		have[tool.Name] = true
	}

	want := []string{
		"analytics_query_funnel",
		"analytics_query_cohort",
		"analytics_query_ab_experiment",
		"analytics_query_actor_failure",
		"analytics_query_cost",
		"analytics_query_slo_burn",
		"analytics_query_feedback_corpus",
	}
	for _, name := range want {
		require.Truef(t, have[name], "analytics tool %q missing from MCP server registration", name)
	}
}

// TestAnalyticsTools_StubReturnsFriendToneText invokes one of the 7 stubs over
// an in-memory MCP session with BUDDY_ANALYTICS_BACKEND unset and verifies the
// response is a friend-tone text body (not a transport error). v0.2.0
// contract: stubs never error out — they return guidance so Claude can
// surface it to the user verbatim. Real adapters preserve this for the
// "backend present but query failed" case.
//
// Not parallel — t.Setenv is mutually exclusive with t.Parallel.
func TestAnalyticsTools_StubReturnsFriendToneText(t *testing.T) {
	t.Setenv("BUDDY_ANALYTICS_BACKEND", "") // explicit: not configured
	ctx := context.Background()

	server := NewBuddyServer(Options{})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "analytics_query_funnel",
		Arguments: map[string]any{
			"stages": []string{"signup", "first_action"},
			"time_range": map[string]string{
				"from": "2026-04-01T00:00:00Z",
				"to":   "2026-05-01T00:00:00Z",
			},
		},
	})
	require.NoError(t, err, "stub must succeed at the transport layer — the friendly error is in the body")
	require.NotNil(t, res)
	require.NotEmpty(t, res.Content, "stub must return text content describing the missing backend")

	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "stub response content must be TextContent (got %T)", res.Content[0])
	require.Contains(t, text.Text, "analytics-mcp/analytics_query_funnel")
	require.Contains(t, text.Text, "BUDDY_ANALYTICS_BACKEND")
}

// TestResolveAnalyticsBackend_RecognisedBackendsAreStubbed pins the v0.2.0
// promise that *valid* backend names also return a stub error (errBackendStubOnly),
// not silent success. When real adapters land, this test should be
// narrowed per-adapter rather than weakened.
//
// Not parallel — t.Setenv mutates a process-global.
func TestResolveAnalyticsBackend_RecognisedBackendsAreStubbed(t *testing.T) {
	for _, backend := range []string{"sql", "mixpanel", "amplitude", "datadog", "stripe", "elasticsearch"} {
		t.Run(backend, func(t *testing.T) {
			t.Setenv("BUDDY_ANALYTICS_BACKEND", backend)
			got, err := resolveAnalyticsBackend()
			require.ErrorIs(t, err, errBackendStubOnly)
			require.Equal(t, backend, got)
		})
	}
}

// TestResolveAnalyticsBackend_UnknownBackendErrors guards against typos
// silently succeeding once adapters land. Unknown values must surface a
// listing of the recognised backends.
func TestResolveAnalyticsBackend_UnknownBackendErrors(t *testing.T) {
	t.Setenv("BUDDY_ANALYTICS_BACKEND", "snowflake-cortex")
	_, err := resolveAnalyticsBackend()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown analytics backend")
	require.Contains(t, err.Error(), "sql")
	require.NotErrorIs(t, err, errBackendStubOnly)
}

// TestResolveAnalyticsBackend_UnsetReturnsNotConfigured matches the documented
// stub behaviour: missing env var = errBackendNotConfigured (different sentinel
// from stub-only) so the handler can word the user guidance accordingly.
func TestResolveAnalyticsBackend_UnsetReturnsNotConfigured(t *testing.T) {
	t.Setenv("BUDDY_ANALYTICS_BACKEND", "")
	_, err := resolveAnalyticsBackend()
	require.ErrorIs(t, err, errBackendNotConfigured)
}

// TestAnalyticsTools_AdapterWiredReturnsJSON proves the adapter wire-up:
// when Options.Analytics is a real adapter (here, the SQL reference impl over
// an in-memory SQLite), the funnel handler returns a JSON body with the typed
// result — not the friend-tone "backend not configured" stub. This is the
// regression guard for "I plugged in an adapter but the MCP layer still falls
// back to the stub."
func TestAnalyticsTools_AdapterWiredReturnsJSON(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite", ":memory:?cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, analytics.Migrate(ctx, db))

	// Seed two signup events + one downstream feature_used event so the
	// funnel returns non-zero counts and pairwise conversion is computable.
	for _, row := range []struct {
		event, user, at string
	}{
		{"signup", "u1", "2026-04-10T00:00:00Z"},
		{"signup", "u2", "2026-04-11T00:00:00Z"},
		{"feature_used", "u1", "2026-04-12T00:00:00Z"},
	} {
		_, err := db.ExecContext(ctx,
			`INSERT INTO events (event_name, user_id, occurred_at) VALUES (?, ?, ?)`,
			row.event, row.user, row.at,
		)
		require.NoError(t, err)
	}

	server := NewBuddyServer(Options{Analytics: analytics.NewSQLAdapter(db)})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "analytics_query_funnel",
		Arguments: map[string]any{
			"stages": []string{"signup", "feature_used"},
			"time_range": map[string]string{
				"from": "2026-04-01T00:00:00Z",
				"to":   "2026-05-01T00:00:00Z",
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Content)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	// The body must be JSON, not the stub guidance.
	require.NotContains(t, text.Text, "BUDDY_ANALYTICS_BACKEND")
	var parsed analytics.FunnelResult
	require.NoError(t, json.Unmarshal([]byte(text.Text), &parsed))
	require.Len(t, parsed.Funnel, 2)
	require.Equal(t, int64(2), parsed.Funnel[0].Count)
	require.Equal(t, int64(1), parsed.Funnel[1].Count)
	require.InDelta(t, 50.0, parsed.Funnel[1].ConversionFromPrior, 0.01)
}

// strings is used implicitly via require.Contains; keep import explicit so the
// test file still compiles if assertions are reshuffled.
var _ = strings.Contains
