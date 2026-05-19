package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/schema"
	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/usage"
)

// TestUsageTools_Registered — the 5 usage_query_* tools land in the
// server's tool list. Regression catch for addUsageTools / server.go.
func TestUsageTools_Registered(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	server := NewBuddyServer(Options{})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	got, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)
	have := make(map[string]bool, len(got.Tools))
	for _, tool := range got.Tools {
		have[tool.Name] = true
	}
	want := []string{
		"usage_query_token_spend",
		"usage_query_session_stats",
		"usage_query_time_distribution",
		"usage_query_top_sessions",
		"usage_query_overview",
	}
	for _, name := range want {
		require.Truef(t, have[name], "usage tool %q missing from MCP server registration", name)
	}
}

// TestUsageTools_NotWiredFallback — without Options.Usage the tools
// respond with a friend-tone hint instead of failing the call.
func TestUsageTools_NotWiredFallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	server := NewBuddyServer(Options{})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "usage_query_token_spend",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	require.Contains(t, tc.Text, "sessions store 연결이 안 돼 있어")
}

// TestUsageTools_TokenSpendWithService — happy path: Options.Usage wired
// over a real on-disk SQLite returns the seeded counts.
func TestUsageTools_TokenSpendWithService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	path := filepath.Join(t.TempDir(), "usage-mcp-test.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	store := sessions.NewStore(conn)
	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, store.Upsert(ctx, sessions.Session{
		ID: "s1", PID: 1, TranscriptPath: "/tmp/s1.jsonl",
		StartedAt: now.Add(-1 * time.Hour), LastActive: now,
		Usage:    schema.TokenUsage{InputTokens: 11, OutputTokens: 22, CacheReadTokens: 33, CacheCreateTokens: 44},
		Metadata: "{}",
	}))

	server := NewBuddyServer(Options{Usage: usage.NewService(conn)})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "usage_query_token_spend",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.True(t, strings.Contains(tc.Text, "\"InputTokens\": 11") || strings.Contains(tc.Text, "\"InputTokens\":11"),
		"token spend JSON should include the seeded input count, got: %s", tc.Text)

	var spend usage.TokenSpend
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &spend))
	require.Equal(t, int64(110), spend.TotalTokens())
}
