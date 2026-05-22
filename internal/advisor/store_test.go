package advisor

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "advisor.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return NewStore(conn)
}

func sampleAdvisory(kind string, severity Severity, at time.Time) Advisory {
	return Advisory{
		Kind:     kind,
		Severity: severity,
		Message:  "test message for " + kind,
		Evidence: []EvidenceItem{
			{Type: "metric", Detail: "value=1"},
			{Type: "chunk", Detail: "snip", ChunkID: 42},
		},
		CreatedAt: at,
	}
}

func TestStore_Insert_RoundTrip(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	want := sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now)

	id, err := store.Insert(ctx, want)
	require.NoError(t, err)
	require.NotZero(t, id)

	got, err := store.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, want.Kind, got.Kind)
	require.Equal(t, want.Severity, got.Severity)
	require.Equal(t, want.Message, got.Message)
	require.Len(t, got.Evidence, 2)
	require.Equal(t, int64(42), got.Evidence[1].ChunkID)
	require.False(t, got.Muted)
	require.Equal(t, want.CreatedAt.UnixMilli(), got.CreatedAt.UnixMilli())
}

func TestStore_Get_UnknownReturnsNotFound(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	_, err := store.Get(context.Background(), 9999)
	require.True(t, errors.Is(err, ErrNotFound))
}

func TestStore_List_DefaultExcludesMuted(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	id1, _ := store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now))
	id2, _ := store.Insert(ctx, sampleAdvisory(KindLongSession, SeverityHigh, now))
	require.NoError(t, store.Mute(ctx, id2))

	got, err := store.List(ctx, ListOptions{})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, id1, got[0].ID)
}

func TestStore_List_IncludeMutedReturnsAll(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now))
	id2, _ := store.Insert(ctx, sampleAdvisory(KindLongSession, SeverityHigh, now))
	require.NoError(t, store.Mute(ctx, id2))

	got, err := store.List(ctx, ListOptions{IncludeMuted: true})
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestStore_List_OrderedDesc(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory("a", SeverityInfo, now.Add(-2*time.Hour)))
	_, _ = store.Insert(ctx, sampleAdvisory("b", SeverityInfo, now))
	_, _ = store.Insert(ctx, sampleAdvisory("c", SeverityInfo, now.Add(-1*time.Hour)))

	got, err := store.List(ctx, ListOptions{})
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, "b", got[0].Kind)
	require.Equal(t, "c", got[1].Kind)
	require.Equal(t, "a", got[2].Kind)
}

func TestStore_List_SinceFilter(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory("old", SeverityInfo, now.Add(-48*time.Hour)))
	_, _ = store.Insert(ctx, sampleAdvisory("recent", SeverityInfo, now))
	got, err := store.List(ctx, ListOptions{Since: now.Add(-24 * time.Hour)})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "recent", got[0].Kind)
}

func TestStore_List_KindsFilter(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityInfo, now))
	_, _ = store.Insert(ctx, sampleAdvisory(KindLongSession, SeverityInfo, now))
	_, _ = store.Insert(ctx, sampleAdvisory(KindLowCacheRatio, SeverityInfo, now))

	got, err := store.List(ctx, ListOptions{Kinds: []string{KindTokenSpikeDay, KindLowCacheRatio}})
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestStore_LastInWindow_FindsRecent(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now.Add(-30*time.Minute)))
	got, err := store.LastInWindow(ctx, KindTokenSpikeDay, time.Hour, now)
	require.NoError(t, err)
	require.Equal(t, KindTokenSpikeDay, got.Kind)
}

func TestStore_LastInWindow_TooOldReturnsNotFound(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now.Add(-25*time.Hour)))
	_, err := store.LastInWindow(ctx, KindTokenSpikeDay, 24*time.Hour, now)
	require.True(t, errors.Is(err, ErrNotFound))
}

func TestStore_LastInWindow_IgnoresMuted(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	id, _ := store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now))
	require.NoError(t, store.Mute(ctx, id))
	_, err := store.LastInWindow(ctx, KindTokenSpikeDay, time.Hour, now)
	require.True(t, errors.Is(err, ErrNotFound), "muted row must not count for dedup")
}

func TestStore_MuteKind(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now))
	_, _ = store.Insert(ctx, sampleAdvisory(KindTokenSpikeDay, SeverityWarn, now))
	_, _ = store.Insert(ctx, sampleAdvisory(KindLongSession, SeverityHigh, now))

	n, err := store.MuteKind(ctx, KindTokenSpikeDay)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)

	active, _ := store.List(ctx, ListOptions{})
	require.Len(t, active, 1)
	require.Equal(t, KindLongSession, active[0].Kind)
}

func TestStore_DeleteOlderThan(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	_, _ = store.Insert(ctx, sampleAdvisory("old", SeverityInfo, now.Add(-48*time.Hour)))
	_, _ = store.Insert(ctx, sampleAdvisory("recent", SeverityInfo, now))
	n, err := store.DeleteOlderThan(ctx, now.Add(-24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	total, _ := store.Count(ctx)
	require.Equal(t, int64(1), total)
}

func TestThresholds_WithDefaults(t *testing.T) {
	t.Parallel()
	// Empty thresholds get fully filled.
	full := Thresholds{}.WithDefaults()
	def := DefaultThresholds()
	require.Equal(t, def, full)

	// Partial overrides preserved.
	partial := Thresholds{TokenSpikeRatio: 2.5, LongSessionHours: 8}.WithDefaults()
	require.Equal(t, 2.5, partial.TokenSpikeRatio)
	require.Equal(t, 8, partial.LongSessionHours)
	require.Equal(t, def.LowCachePct, partial.LowCachePct, "non-overridden field falls back")
}
