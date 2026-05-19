package knowledge

import "sort"

// rrfK is the RRF (Reciprocal Rank Fusion) damping constant. 60 is the
// standard value from Cormack & Clarke 2009 — kept small enough that
// rank-1 hits dominate but ties at higher ranks still contribute.
const rrfK = 60

// HybridSearch fuses BM25 + vector results via Reciprocal Rank Fusion.
// Each retrieval channel contributes 1/(k+rank) per shared chunk; sum
// across channels is the final score. Pure RRF with no per-channel
// weight — both channels are treated as equal-strength signals.
//
// The query embedding may be nil — then only BM25 is consulted (still
// useful when Python venv isn't installed yet). queryText is mandatory
// — BM25 always runs.
//
// limit<=0 → top 10.
func HybridSearch(idx *BM25Index, chunks []Chunk, queryText string, queryEmbedding []float32, limit int) []ScoredChunk {
	if limit <= 0 {
		limit = 10
	}
	if idx == nil {
		return nil
	}

	// Each channel returns up to 3x limit so RRF has enough overlap
	// candidates. 3x is empirical — big enough to merge usefully,
	// small enough that the sort stays cheap.
	bmHits := idx.Search(queryText, limit*3)

	var vecHits []ScoredChunk
	if queryEmbedding != nil {
		vecHits = VectorSearch(chunks, queryEmbedding, limit*3)
	}

	// Fuse via RRF.
	type acc struct {
		score   float64
		chunk   Chunk
		sources []string
	}
	merged := map[int64]*acc{}
	addRanked := func(hits []ScoredChunk, source string) {
		for rank, h := range hits {
			a, ok := merged[h.ID]
			if !ok {
				a = &acc{chunk: h.Chunk}
				merged[h.ID] = a
			}
			a.score += 1.0 / float64(rrfK+rank+1)
			a.sources = append(a.sources, source)
		}
	}
	addRanked(bmHits, "bm25")
	addRanked(vecHits, "vector")

	out := make([]ScoredChunk, 0, len(merged))
	for _, a := range merged {
		channel := "hybrid"
		if len(a.sources) == 1 {
			channel = a.sources[0]
		}
		out = append(out, ScoredChunk{
			Chunk:   a.chunk,
			Score:   a.score,
			Channel: channel,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].ID < out[j].ID
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
