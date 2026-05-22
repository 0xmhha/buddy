package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestNotifyTools_NotWiredFallback — the notify_status and notify_test
// tools both register even when NotifyOptions is zero (Store /
// Dispatcher both nil). The CLI process must continue to function for
// installs that have not configured a daemon notification channel
// yet; the tools surface the not-wired condition in the response body
// rather than failing the MCP dispatch.
func TestNotifyTools_NotWiredFallback(t *testing.T) {
	t.Parallel()
	server := NewBuddyServer(Options{})
	client := mcp.NewClient(&mcp.Implementation{Name: "notify-test", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	statusRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "notify_status",
		Arguments: map[string]any{"channel": "", "since": "", "limit": 10},
	})
	require.NoError(t, err)
	tc, ok := statusRes.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "연결 안 돼",
		"missing store surfaces the friend-tone not-wired message")

	testRes, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "notify_test",
		Arguments: map[string]any{"channel": "desktop"},
	})
	require.NoError(t, err)
	tc, ok = testRes.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "연결 안 돼")
}
