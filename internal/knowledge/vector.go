package knowledge

import (
	"math"
	"sort"
)

// CosineSimilarity returns the cosine of the angle between a and b.
// Returns 0 when either is empty or any norm is zero. Doesn't allocate.
// Per ADR-014 §Q2 the embedding dim is whatever the Python embedder
// produces (sentence-transformers default 384 for all-MiniLM-L6-v2),
// so the loop body is the hot path on Search calls.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		fa := float64(a[i])
		fb := float64(b[i])
		dot += fa * fb
		normA += fa * fa
		normB += fb * fb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// VectorSearch returns the top-k chunks ranked by cosine similarity
// against query. Chunks without an Embedding are skipped (BM25 still
// covers them in the hybrid combiner). k<=0 → top 10.
//
// Stateless — every call iterates the full chunks slice. For the
// expected ~10k corpus this is sub-100ms; we'd swap to a flat index
// or chromem-go past that, per ADR-014 §Trigger to revisit.
func VectorSearch(chunks []Chunk, query []float32, k int) []ScoredChunk {
	if k <= 0 {
		k = 10
	}
	if len(query) == 0 || len(chunks) == 0 {
		return nil
	}
	type idxScore struct {
		i int
		s float64
	}
	var ranked []idxScore
	for i, c := range chunks {
		if c.Embedding == nil {
			continue
		}
		s := CosineSimilarity(c.Embedding, query)
		if s > 0 {
			ranked = append(ranked, idxScore{i: i, s: s})
		}
	}
	sort.Slice(ranked, func(a, b int) bool {
		if ranked[a].s != ranked[b].s {
			return ranked[a].s > ranked[b].s
		}
		return chunks[ranked[a].i].ID < chunks[ranked[b].i].ID
	})
	if len(ranked) > k {
		ranked = ranked[:k]
	}
	out := make([]ScoredChunk, 0, len(ranked))
	for _, r := range ranked {
		out = append(out, ScoredChunk{
			Chunk:   chunks[r.i],
			Score:   r.s,
			Channel: "vector",
		})
	}
	return out
}
