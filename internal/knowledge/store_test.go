package knowledge

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/sessions"
)

func newTestStore(t *testing.T) (*Store, *sessions.Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "knowledge.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return NewStore(conn), sessions.NewStore(conn)
}

func seedSession(t *testing.T, ss *sessions.Store, id string) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, ss.Upsert(context.Background(), sessions.Session{
		ID: id, PID: 1, TranscriptPath: "/tmp/" + id + ".jsonl",
		StartedAt: now, LastActive: now, Metadata: "{}",
	}))
}

func TestStore_Insert_RoundTripWithoutEmbedding(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "s1")
	id, err := store.Insert(ctx, Chunk{
		SessionID: "s1", Content: "hello world", TokenCount: 2,
	})
	require.NoError(t, err)
	require.NotZero(t, id)
	got, err := store.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "s1", got.SessionID)
	require.Equal(t, "hello world", got.Content)
	require.Equal(t, 2, got.TokenCount)
	require.Nil(t, got.Embedding, "nil embedding round-trips as nil")
	require.False(t, got.CreatedAt.IsZero())
}

func TestStore_Insert_RoundTripWithEmbedding(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "s2")
	emb := []float32{0.1, -0.2, 0.3, -0.4, 1.0}
	id, err := store.Insert(ctx, Chunk{
		SessionID: "s2", Content: "chunk-x", TokenCount: 1,
		Embedding: emb,
	})
	require.NoError(t, err)
	got, err := store.Get(ctx, id)
	require.NoError(t, err)
	require.Len(t, got.Embedding, len(emb))
	for i := range emb {
		require.InDelta(t, float64(emb[i]), float64(got.Embedding[i]), 1e-6)
	}
}

func TestStore_Get_UnknownReturnsNotFound(t *testing.T) {
	t.Parallel()
	store, _ := newTestStore(t)
	_, err := store.Get(context.Background(), 9999)
	require.True(t, errors.Is(err, ErrNotFound))
}

func TestStore_UpdateEmbedding_FlipsNilToValue(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "s3")
	id, err := store.Insert(ctx, Chunk{SessionID: "s3", Content: "x", TokenCount: 1})
	require.NoError(t, err)
	require.NoError(t, store.UpdateEmbedding(ctx, id, []float32{1, 2, 3}))
	got, err := store.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, []float32{1, 2, 3}, got.Embedding)
}

func TestStore_UpdateEmbedding_UnknownNotFound(t *testing.T) {
	t.Parallel()
	store, _ := newTestStore(t)
	err := store.UpdateEmbedding(context.Background(), 9999, []float32{1})
	require.True(t, errors.Is(err, ErrNotFound))
}

func TestStore_ListBySession_OrderedAndScoped(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "alpha")
	seedSession(t, ss, "beta")
	for _, c := range []string{"a1", "a2", "a3"} {
		_, err := store.Insert(ctx, Chunk{SessionID: "alpha", Content: c, TokenCount: 1})
		require.NoError(t, err)
	}
	_, err := store.Insert(ctx, Chunk{SessionID: "beta", Content: "b1", TokenCount: 1})
	require.NoError(t, err)
	got, err := store.ListBySession(ctx, "alpha")
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, "a1", got[0].Content)
	require.Equal(t, "a3", got[2].Content)
}

func TestStore_CountAndCountWithEmbedding(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "s4")
	_, err := store.Insert(ctx, Chunk{SessionID: "s4", Content: "no-emb", TokenCount: 1})
	require.NoError(t, err)
	id2, err := store.Insert(ctx, Chunk{SessionID: "s4", Content: "with-emb", TokenCount: 1,
		Embedding: []float32{1, 2}})
	require.NoError(t, err)
	_ = id2
	c, err := store.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2), c)
	cw, err := store.CountWithEmbedding(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), cw)
}

func TestStore_DeleteBySession(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "doomed")
	seedSession(t, ss, "kept")
	_, _ = store.Insert(ctx, Chunk{SessionID: "doomed", Content: "x", TokenCount: 1})
	_, _ = store.Insert(ctx, Chunk{SessionID: "doomed", Content: "y", TokenCount: 1})
	_, _ = store.Insert(ctx, Chunk{SessionID: "kept", Content: "z", TokenCount: 1})
	n, err := store.DeleteBySession(ctx, "doomed")
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
	left, _ := store.Count(ctx)
	require.Equal(t, int64(1), left)
}

func TestStore_DeleteAll(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "s5")
	_, _ = store.Insert(ctx, Chunk{SessionID: "s5", Content: "a", TokenCount: 1})
	_, _ = store.Insert(ctx, Chunk{SessionID: "s5", Content: "b", TokenCount: 1})
	n, err := store.DeleteAll(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
	c, _ := store.Count(ctx)
	require.Zero(t, c)
}

// TestStore_FKCascadeFromSessions — deleting the parent session row
// cascades to chunks (migration v6 FK declaration).
func TestStore_FKCascadeFromSessions(t *testing.T) {
	t.Parallel()
	store, ss := newTestStore(t)
	ctx := context.Background()
	seedSession(t, ss, "victim")
	_, err := store.Insert(ctx, Chunk{SessionID: "victim", Content: "k", TokenCount: 1})
	require.NoError(t, err)
	require.NoError(t, ss.Delete(ctx, "victim"))
	got, err := store.ListBySession(ctx, "victim")
	require.NoError(t, err)
	require.Empty(t, got, "chunks must cascade-delete with parent session row")
}

func TestEncodeDecodeEmbedding_RoundTrip(t *testing.T) {
	t.Parallel()
	cases := [][]float32{
		nil,
		{},
		{0.0, 1.0, -1.0},
		{0.1234, -0.5678, 3.14159, -2.71828},
	}
	for _, in := range cases {
		blob, err := encodeEmbedding(in)
		require.NoError(t, err)
		if in == nil {
			require.Nil(t, blob)
			continue
		}
		raw, ok := blob.([]byte)
		require.True(t, ok)
		out, err := decodeEmbedding(raw)
		require.NoError(t, err)
		require.Len(t, out, len(in))
		for i := range in {
			require.InDelta(t, float64(in[i]), float64(out[i]), 1e-6)
		}
	}
}
