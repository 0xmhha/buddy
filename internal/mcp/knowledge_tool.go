package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/0xmhha/buddy/internal/knowledge"
)

// knowledge_tool.go wires the single `knowledge_query` MCP tool.
// Read-only over the chunks table. The advisor consumes this as its
// retrieval primitive.

type knowledgeQueryArgs struct {
	Query string `json:"query" jsonschema:"Natural-language query text to retrieve relevant session chunks."`
	K     int    `json:"k,omitempty" jsonschema:"Max chunks to return (default 5)."`
}

// knowledgeQueryResult wraps []ScoredChunk so the MCP SDK has an object
// root (array-at-root output schemas are rejected, same constraint we
// hit in usage_query_top_sessions).
type knowledgeQueryResult struct {
	Hits []knowledge.ScoredChunk `json:"hits"`
}

// KnowledgeOptions configures the retrieval side of knowledge_query.
// Splitting Store from Embedder lets the MCP server load the cheap
// store eagerly while leaving the Python sub-process strictly optional:
// if Embedder is nil, the handler falls back to BM25-only retrieval
// and tags hits as such.
type KnowledgeOptions struct {
	Store    *knowledge.Store
	Embedder knowledge.Embedder
}

func addKnowledgeTools(s *mcp.Server, opts Options) {
	k := opts.Knowledge

	mcp.AddTool(s, &mcp.Tool{
		Name:        "knowledge_query",
		Description: "Retrieve top-K chunks of past Claude Code session content for a natural-language query. Hybrid BM25 + vector retrieval over the local chunks table populated by `buddy knowledge ingest`. When the Python embedder is unavailable, falls back to BM25-only and tags hits with channel='bm25'.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args knowledgeQueryArgs) (*mcp.CallToolResult, knowledgeQueryResult, error) {
		if k.Store == nil {
			return knowledgeNotWired(), knowledgeQueryResult{}, nil
		}
		limit := args.K
		if limit <= 0 {
			limit = 5
		}
		all, err := k.Store.All(ctx)
		if err != nil {
			return nil, knowledgeQueryResult{}, err
		}
		if len(all) == 0 {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{
				Text: "knowledge-mcp: chunks 가 비어 있어. `buddy knowledge ingest --all` 먼저 실행해줘.",
			}}}, knowledgeQueryResult{}, nil
		}
		idx := knowledge.NewBM25Index(all)

		var emb []float32
		if k.Embedder != nil {
			res, err := k.Embedder.Embed(ctx, []knowledge.EmbedRequest{{ID: 0, Text: args.Query}})
			// Best-effort: embedder errors don't fail the whole call; the
			// caller still gets BM25 hits. The friend-tone reason lands in
			// a sibling response field via the channel tag.
			if err == nil && len(res) > 0 {
				emb = res[0].Embedding
			}
		}
		hits := knowledge.HybridSearch(idx, all, args.Query, emb, limit)
		out := knowledgeQueryResult{Hits: hits}
		return jsonContent(out), out, nil
	})
}

func knowledgeNotWired() *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{
		Text: "knowledge-mcp: knowledge.Store 가 연결 안 돼 있어. " +
			"`buddy mcp serve` 가 chunks 테이블이 있는 buddy.db 를 가리키는지 확인해줘.",
	}}}
}
