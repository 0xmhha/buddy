// Package knowledge implements F2.C Phase 1 (W7-3a, ADR-014) — the
// knowledge-retrieval foundation that feeds the future F2.C Advisor
// (Phase 2, W7-3b, v0.11.0). Components:
//
//   - Chunk + Store : SQLite-backed chunk records (migration v6).
//   - Chunker       : transcript JSONL → chunks (user/assistant pair,
//                     ~500-token cap).
//   - BM25          : sparse keyword retrieval over chunk content.
//   - Vector        : cosine similarity over float32 embedding BLOBs.
//   - Hybrid        : reciprocal-rank-fusion combiner.
//   - Embedder      : Python sub-process invoker (sentence-transformers).
//
// v0.10.0 ships the primitive; v0.11.0 (Phase 2) builds the advisory
// generator on top via retrieval + rule. Scope strictly retrieval-only
// here per ADR-014 §"v0.10.0 (Phase 1) design choices".
package knowledge

import "time"

// Chunk is one record in the `chunks` table (migration v6). Embedding
// is NULL until the Python embedder fills it — retrieval gracefully
// skips chunks with nil Embedding for the vector channel but uses them
// for BM25 (text is sufficient there).
type Chunk struct {
	ID         int64
	SessionID  string
	Content    string
	TokenCount int
	Embedding  []float32 // nil until embedding pass runs
	CreatedAt  time.Time
}

// ScoredChunk pairs a chunk with the retrieval score that surfaced it.
// Channel is the retrieval source ("bm25" / "vector" / "hybrid") so
// the CLI / MCP renderer can show why a given chunk ranked where it did.
type ScoredChunk struct {
	Chunk
	Score   float64
	Channel string
}
