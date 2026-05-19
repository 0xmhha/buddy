package knowledge

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockEmbedder_HappyPath(t *testing.T) {
	t.Parallel()
	m := &MockEmbedder{
		Vectors: map[string][]float32{
			"hello": {0.1, 0.2},
			"world": {0.3, 0.4},
		},
	}
	got, err := m.Embed(context.Background(), []EmbedRequest{
		{ID: 1, Text: "hello"},
		{ID: 2, Text: "world"},
		{ID: 3, Text: "unmapped"}, // dropped
	})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, int64(1), got[0].ID)
	require.Equal(t, int64(2), got[1].ID)
}

func TestMockEmbedder_ErrPropagated(t *testing.T) {
	t.Parallel()
	bang := errors.New("simulated venv missing")
	m := &MockEmbedder{Err: bang}
	_, err := m.Embed(context.Background(), []EmbedRequest{{ID: 1, Text: "x"}})
	require.ErrorIs(t, err, bang)
}

func TestPythonEmbedder_MissingScript_ReturnsUnavailable(t *testing.T) {
	t.Parallel()
	e := &PythonEmbedder{ScriptPath: filepath.Join(t.TempDir(), "missing.py")}
	_, err := e.Embed(context.Background(), []EmbedRequest{{ID: 1, Text: "x"}})
	require.ErrorIs(t, err, ErrEmbedderUnavailable)
}

func TestPythonEmbedder_EmptyReqsReturnsNilWithoutSpawn(t *testing.T) {
	t.Parallel()
	// Even with no script path, an empty batch should short-circuit.
	e := &PythonEmbedder{ScriptPath: "/nonexistent/path/to/embed.py"}
	got, err := e.Embed(context.Background(), nil)
	require.NoError(t, err)
	require.Nil(t, got)
}
