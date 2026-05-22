package mcp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
)

// TestFeatureTools_UpsertGetListDeleteRoundTrip — drives the four
// state-changing feature tools through their natural lifecycle and
// asserts each transition shows up at the next read. The intent is
// regression detection on the wire-up between the MCP handler and
// internal/feature, not exhaustive coverage of the feature package
// itself (which has its own unit tests).
func TestFeatureTools_UpsertGetListDeleteRoundTrip(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "feature.db")
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	server := NewBuddyServer(Options{DBPath: dbPath})
	client := mcp.NewClient(&mcp.Implementation{Name: "feature-test", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Insert.
	upsertRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "feature_upsert",
		Arguments: map[string]any{
			"feature_id":          "test-feat",
			"name":                "Test Feature",
			"status":              "draft",
			"summary":             "round-trip fixture",
			"actors":              []any{},
			"acceptance_criteria": []any{},
			"test_plan": map[string]any{
				"unit": []any{}, "integration": []any{}, "e2e": []any{},
			},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, upsertRes.Content)

	// Get → text body should mention the feature id.
	getRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "feature_get",
		Arguments: map[string]any{"feature_id": "test-feat"},
	})
	require.NoError(t, err)
	tc, ok := getRes.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "test-feat")
	require.Contains(t, tc.Text, "Test Feature")

	// List → must include the upserted row.
	listRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "feature_list",
		Arguments: map[string]any{"status": ""},
	})
	require.NoError(t, err)
	tc, ok = listRes.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "test-feat")

	// Delete is idempotent — it must succeed once.
	delRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "feature_delete",
		Arguments: map[string]any{"feature_id": "test-feat"},
	})
	require.NoError(t, err)
	tc, ok = delRes.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "deleted")

	// Subsequent get → "feature not found" (friend-tone, not an error).
	getRes, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "feature_get",
		Arguments: map[string]any{"feature_id": "test-feat"},
	})
	require.NoError(t, err, "missing feature is reported in body, not as an MCP error")
	tc, ok = getRes.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "not found")
}

// TestFeatureSearch_RegistersAndAcceptsQuery — feature_search is one of
// the five feature tools the server registers. The full round-trip on
// matching/non-matching rows is exercised by the underlying feature
// package's own tests; here we just confirm the MCP envelope accepts a
// query argument and returns a text body without error.
func TestFeatureSearch_RegistersAndAcceptsQuery(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "feature-search.db")
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	server := NewBuddyServer(Options{DBPath: dbPath})
	client := mcp.NewClient(&mcp.Implementation{Name: "feature-search-test", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	ctx := context.Background()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "feature_search",
		Arguments: map[string]any{"query": "nothing-stored"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Content, "search returns a body even when result is empty")
	_, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "feature_search body is TextContent")
}
