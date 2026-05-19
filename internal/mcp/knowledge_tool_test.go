package mcp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/knowledge"
	"github.com/0xmhha/buddy/internal/sessions"
)

func TestKnowledgeTool_Registered(t *testing.T) {
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
	have := map[string]bool{}
	for _, tool := range got.Tools {
		have[tool.Name] = true
	}
	require.True(t, have["knowledge_query"], "knowledge_query missing")
}

func TestKnowledgeTool_NotWiredFallback(t *testing.T) {
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
		Name: "knowledge_query",
		Arguments: map[string]any{
			"query": "anything",
		},
	})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "knowledge.Store 가 연결 안 돼 있어")
}

func TestKnowledgeTool_EmptyCorpusReturnsHint(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "knowledge-mcp.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	server := NewBuddyServer(Options{
		Knowledge: KnowledgeOptions{Store: knowledge.NewStore(conn)},
	})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "knowledge_query", Arguments: map[string]any{"query": "x"},
	})
	require.NoError(t, err)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "chunks 가 비어 있어")
}

func TestKnowledgeTool_HitsViaBM25Only(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "knowledge-mcp.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	ss := sessions.NewStore(conn)
	require.NoError(t, ss.Upsert(ctx, sessions.Session{
		ID: "s1", PID: 1, TranscriptPath: "/tmp/s.jsonl",
		Metadata: "{}",
	}))
	ks := knowledge.NewStore(conn)
	_, err = ks.Insert(ctx, knowledge.Chunk{
		SessionID: "s1", Content: "buddy mcp knowledge query test",
		TokenCount: 5,
	})
	require.NoError(t, err)

	// No embedder wired → BM25 only.
	server := NewBuddyServer(Options{
		Knowledge: KnowledgeOptions{Store: ks},
	})
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	st, ct := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer serverSession.Close()
	clientSession, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "knowledge_query",
		Arguments: map[string]any{"query": "buddy mcp"},
	})
	require.NoError(t, err)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, tc.Text, "buddy mcp knowledge query test")
	require.Contains(t, tc.Text, "bm25") // channel tag since no embedder
}
