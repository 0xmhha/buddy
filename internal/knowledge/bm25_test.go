package knowledge

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBM25_EmptyCorpus(t *testing.T) {
	t.Parallel()
	idx := NewBM25Index(nil)
	require.Empty(t, idx.Search("anything", 5))
}

func TestBM25_FindsRelevantDocByKeyword(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "the quick brown fox jumps over the lazy dog"},
		{ID: 2, Content: "buddy is a friend for claude code users"},
		{ID: 3, Content: "the fox is sly and quick"},
	}
	idx := NewBM25Index(chunks)
	got := idx.Search("fox", 3)
	require.NotEmpty(t, got)
	// Top hit should mention fox; chunk 1 or 3 both contain it,
	// shorter doc (3) wins by length-norm.
	require.True(t, got[0].ID == 1 || got[0].ID == 3)
	require.Equal(t, "bm25", got[0].Channel)
}

func TestBM25_RankingPrefersTermFrequency(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "buddy buddy buddy buddy buddy is great"},
		{ID: 2, Content: "buddy is great"},
		{ID: 3, Content: "nothing here matches"},
	}
	idx := NewBM25Index(chunks)
	got := idx.Search("buddy", 3)
	require.Len(t, got, 2, "only 2 chunks mention buddy")
	require.Equal(t, int64(1), got[0].ID, "high TF chunk ranks first")
}

func TestBM25_LimitRespected(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "alpha"}, {ID: 2, Content: "alpha"}, {ID: 3, Content: "alpha"},
	}
	idx := NewBM25Index(chunks)
	got := idx.Search("alpha", 2)
	require.Len(t, got, 2)
}

func TestBM25_NoHitsReturnsEmpty(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{{ID: 1, Content: "alpha beta gamma"}}
	idx := NewBM25Index(chunks)
	require.Empty(t, idx.Search("zeta", 5))
}

func TestBM25_KoreanContent(t *testing.T) {
	t.Parallel()
	// Whitespace-split is the documented v0.10.0 limit (per
	// bm25.go's "Future work: language-aware tokenisation" note):
	// query must match a whole whitespace-bounded token. So we test
	// against an exact word boundary; morpheme matching ("버디" inside
	// "버디는") is a v0.10.x trigger.
	chunks := []Chunk{
		{ID: 1, Content: "버디는 클로드 코드 친구"},
		{ID: 2, Content: "오늘 날씨 좋다"},
	}
	idx := NewBM25Index(chunks)
	got := idx.Search("친구", 5)
	require.NotEmpty(t, got)
	require.Equal(t, int64(1), got[0].ID)
}

func TestBM25_IDFLowersScoreForCommonTerms(t *testing.T) {
	t.Parallel()
	chunks := []Chunk{
		{ID: 1, Content: "the buddy"},
		{ID: 2, Content: "the cat"},
		{ID: 3, Content: "the dog"},
	}
	idx := NewBM25Index(chunks)
	// "the" appears in every doc → idf ~0; "buddy" rare → high idf.
	gotBuddy := idx.Search("buddy", 3)
	gotThe := idx.Search("the", 3)
	require.NotEmpty(t, gotBuddy)
	// "the" appears everywhere; its idf rounds to ~0, so score may be ~0.
	// Either no hits or single low-score hit. The contract: rare term
	// scores strictly higher than common term.
	if len(gotThe) > 0 {
		require.Greater(t, gotBuddy[0].Score, gotThe[0].Score,
			"rare term must outscore common term")
	}
}

func TestTokenizeBM25_DropsShortTokens(t *testing.T) {
	t.Parallel()
	// Documented threshold: drop tokens of <= 1 rune. So "a" and "I"
	// vanish; "am" survives (it is a stopword candidate but v0.10.0
	// doesn't have a stoplist — that's a future enhancement).
	got := tokenizeBM25("a I am the cat")
	require.NotContains(t, got, "a")
	require.NotContains(t, got, "i") // lowercased
	require.Contains(t, got, "am")
	require.Contains(t, got, "the")
	require.Contains(t, got, "cat")
}

func TestTokenizeBM25_LowercasesAndStripsPunct(t *testing.T) {
	t.Parallel()
	got := tokenizeBM25("Hello, World! Buddy-CLI.")
	require.Equal(t, []string{"hello", "world", "buddy", "cli"}, got)
}
