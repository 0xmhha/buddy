package knowledge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCosineSimilarity_Sanity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a, b []float32
		want float64
	}{
		{"identical", []float32{1, 0}, []float32{1, 0}, 1.0},
		{"orthogonal", []float32{1, 0}, []float32{0, 1}, 0.0},
		{"opposite", []float32{1, 0}, []float32{-1, 0}, -1.0},
		{"empty", nil, nil, 0.0},
		{"diff_len", []float32{1, 0}, []float32{1, 0, 0}, 0.0},
		{"zero_norm", []float32{0, 0}, []float32{1, 1}, 0.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.InDelta(t, tc.want, CosineSimilarity(tc.a, tc.b), 1e-6)
		})
	}
}

func TestVectorSearch_SkipsNilEmbedding(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "no emb"},
		{ID: 2, Content: "with emb", Embedding: []float32{1, 0, 0}},
	}
	got := VectorSearch(chunks, []float32{1, 0, 0}, 5)
	require.Len(t, got, 1)
	require.Equal(t, int64(2), got[0].ID)
	require.Equal(t, "vector", got[0].Channel)
}

func TestVectorSearch_RanksByCosineDesc(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Embedding: []float32{1, 0, 0}},
		{ID: 2, Embedding: []float32{0.7, 0.7, 0}},
		{ID: 3, Embedding: []float32{0, 1, 0}},
	}
	got := VectorSearch(chunks, []float32{1, 0, 0}, 3)
	require.Equal(t, []int64{1, 2}, []int64{got[0].ID, got[1].ID})
	// chunk 3 orthogonal → score 0, dropped.
	require.Len(t, got, 2)
}

func TestVectorSearch_LimitRespected(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Embedding: []float32{1, 1, 1}},
		{ID: 2, Embedding: []float32{1, 1, 1}},
		{ID: 3, Embedding: []float32{1, 1, 1}},
	}
	got := VectorSearch(chunks, []float32{1, 1, 1}, 2)
	require.Len(t, got, 2)
}

func TestVectorSearch_EmptyQueryReturnsNil(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{{ID: 1, Embedding: []float32{1, 0}}}
	require.Nil(t, VectorSearch(chunks, nil, 5))
	require.Nil(t, VectorSearch(nil, []float32{1, 0}, 5))
}

func TestHybridSearch_FusesBothChannels(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "buddy alpha", Embedding: []float32{1, 0, 0}},
		{ID: 2, Content: "claude beta", Embedding: []float32{0, 1, 0}},
		{ID: 3, Content: "buddy claude gamma", Embedding: []float32{0.5, 0.5, 0}},
	}
	idx := NewBM25Index(chunks)
	got := HybridSearch(idx, chunks, "buddy", []float32{1, 0, 0}, 5)
	require.NotEmpty(t, got)
	// Chunk 1 surfaces in both channels → highest RRF.
	require.Equal(t, int64(1), got[0].ID)
	require.Equal(t, "hybrid", got[0].Channel,
		"chunk surfaced by both channels should be marked hybrid")
}

func TestHybridSearch_NilQueryEmbeddingFallsBackToBM25(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "buddy is great"},
		{ID: 2, Content: "claude code rocks"},
	}
	idx := NewBM25Index(chunks)
	got := HybridSearch(idx, chunks, "buddy", nil, 5)
	require.NotEmpty(t, got)
	require.Equal(t, int64(1), got[0].ID)
	require.Equal(t, "bm25", got[0].Channel,
		"vector channel inactive → result tagged bm25 only")
}

func TestHybridSearch_LimitTrims(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "buddy"},
		{ID: 2, Content: "buddy"},
		{ID: 3, Content: "buddy"},
	}
	idx := NewBM25Index(chunks)
	got := HybridSearch(idx, chunks, "buddy", nil, 2)
	require.Len(t, got, 2)
}

func TestHybridSearch_NilIndex(t *testing.T) {
	t.Parallel()
	require.Nil(t, HybridSearch(nil, nil, "x", nil, 5))
}
