package knowledge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeJSONL(t *testing.T, lines []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644))
	return path
}

func TestChunker_ExtractUserAndAssistantPair(t *testing.T) {
	t.Parallel()
	path := writeJSONL(t, []string{
		`{"type":"user","sessionId":"abc","message":{"role":"user","content":"hello buddy"}}`,
		`{"type":"assistant","sessionId":"abc","message":{"role":"assistant","content":"hi friend"}}`,
	})
	got, err := NewChunker().Extract(path)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "abc", got[0].SessionID)
	require.Contains(t, got[0].Content, "[user]")
	require.Contains(t, got[0].Content, "hello buddy")
	require.Contains(t, got[1].Content, "[assistant]")
}

func TestChunker_SkipsMetaAndEmpty(t *testing.T) {
	t.Parallel()
	path := writeJSONL(t, []string{
		`{"type":"system","isMeta":true,"message":{"role":"system","content":"setup"}}`,
		`{"type":"user","message":{"role":"user","content":""}}`,
		`{"type":"user","message":{"role":"user","content":"real prompt"}}`,
	})
	got, err := NewChunker().Extract(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Contains(t, got[0].Content, "real prompt")
}

func TestChunker_TypedPartsExtractsText(t *testing.T) {
	t.Parallel()
	path := writeJSONL(t, []string{
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"first"},{"type":"tool_use","name":"X"},{"type":"text","text":"second"}]}}`,
	})
	got, err := NewChunker().Extract(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Contains(t, got[0].Content, "first")
	require.Contains(t, got[0].Content, "second")
}

func TestChunker_DropsPureToolMessages(t *testing.T) {
	t.Parallel()
	path := writeJSONL(t, []string{
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash"}]}}`,
	})
	got, err := NewChunker().Extract(path)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestChunker_SplitsLongMessageByMaxTokens(t *testing.T) {
	t.Parallel()
	// 1200 tokens body + role prefix (+1) = 1201 tokens. With MaxTokens
	// 300 we expect 5 chunks (4 × 300 + 1 × ≤300). The exact count is
	// not the point — what matters is that no single chunk exceeds the
	// cap, and that the body got split (not one giant chunk).
	words := make([]string, 1200)
	for i := range words {
		words[i] = "tok"
	}
	long := strings.Join(words, " ")
	path := writeJSONL(t, []string{
		`{"type":"user","message":{"role":"user","content":"` + long + `"}}`,
	})
	c := &Chunker{MaxTokens: 300}
	got, err := c.Extract(path)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(got), 4, "long body must be split into >=4 pieces")
	for _, ch := range got {
		require.LessOrEqual(t, ch.TokenCount, 300, "each chunk must respect MaxTokens cap")
	}
}

func TestChunker_SessionIDFromFilenameWhenAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "sess-xyz.jsonl")
	require.NoError(t, os.WriteFile(path,
		[]byte(`{"type":"user","message":{"role":"user","content":"hi"}}`+"\n"), 0o644))
	got, err := NewChunker().Extract(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "sess-xyz", got[0].SessionID)
}

func TestChunker_MalformedLineSkipped(t *testing.T) {
	t.Parallel()
	path := writeJSONL(t, []string{
		`{this is not json`,
		`{"type":"user","message":{"role":"user","content":"good"}}`,
	})
	got, err := NewChunker().Extract(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
}

func TestChunker_MissingPathErrors(t *testing.T) {
	t.Parallel()
	_, err := NewChunker().Extract(filepath.Join(t.TempDir(), "nope.jsonl"))
	require.Error(t, err)
}

func TestSplitByTokens_BoundariesPreserved(t *testing.T) {
	t.Parallel()
	s := "alpha beta gamma delta epsilon"
	out := splitByTokens(s, 2)
	require.Len(t, out, 3)
	require.Equal(t, "alpha beta", out[0])
	require.Equal(t, "gamma delta", out[1])
	require.Equal(t, "epsilon", out[2])
}

func TestSplitByTokens_EmptyInput(t *testing.T) {
	t.Parallel()
	require.Empty(t, splitByTokens("", 10))
	require.Empty(t, splitByTokens("   \t  ", 10))
}
