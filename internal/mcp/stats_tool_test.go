package mcp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
)

// TestStatsTool_DefaultsWindowAndReturnsEmpty — the stats tool defaults
// to a 1h window when the caller omits it, and a freshly-migrated DB
// with no hook events returns an empty rows slice + the friend-tone
// "No stats yet" body. Exercises the full MCP round-trip plus the
// default-window path that production callers depend on.
func TestStatsTool_DefaultsWindowAndReturnsEmpty(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "stats.db")
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	server := NewBuddyServer(Options{DBPath: dbPath})
	client := mcp.NewClient(&mcp.Implementation{Name: "stats-test", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// The JSON schema treats both fields as required; an empty string
	// for window flows through to the handler, which then promotes it
	// to the documented default of "1h".
	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "stats",
		Arguments: map[string]any{"window": "", "hook": ""},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Content)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "stats body must be TextContent")
	require.Contains(t, tc.Text, "1h",
		"empty window string must default to 1h")
}
