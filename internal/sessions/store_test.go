package sessions

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/schema"
)

// newTestStore opens a fresh on-disk SQLite store + runs migrations.
// On-disk (not :memory:) so multi-connection tests can share the same DB.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "buddy-sessions.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return NewStore(conn)
}

func sampleSession(id string, lastActive time.Time) Session {
	return Session{
		ID:             id,
		PID:            12345,
		TranscriptPath: "/tmp/" + id + ".jsonl",
		StartedAt:      lastActive.Add(-30 * time.Minute),
		LastActive:     lastActive,
		Usage: schema.TokenUsage{
			InputTokens:       100,
			OutputTokens:      200,
			CacheReadTokens:   50,
			CacheCreateTokens: 10,
		},
		LastOffset: 2048,
		GoalText:   "make a thing",
		Metadata:   `{"project":"buddy"}`,
	}
}

// TestStore_Upsert_RoundTrip — Upsert + Get round-trips every field.
func TestStore_Upsert_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	want := sampleSession("sess-a", now)
	require.NoError(t, store.Upsert(ctx, want))

	got, err := store.Get(ctx, "sess-a")
	require.NoError(t, err)
	require.Equal(t, want.ID, got.ID)
	require.Equal(t, want.PID, got.PID)
	require.Equal(t, want.TranscriptPath, got.TranscriptPath)
	require.Equal(t, want.StartedAt.UnixMilli(), got.StartedAt.UnixMilli())
	require.Equal(t, want.LastActive.UnixMilli(), got.LastActive.UnixMilli())
	require.Equal(t, want.Usage, got.Usage)
	require.Equal(t, want.LastOffset, got.LastOffset)
	require.Equal(t, want.GoalText, got.GoalText)
	require.Equal(t, want.Metadata, got.Metadata)
	require.Nil(t, got.EndedAt, "Upsert without EndedAt must leave it NULL")
}

// TestStore_Upsert_PreservesStartedAt — a second Upsert with a different
// StartedAt must NOT overwrite the first (started_at is set on first
// observation and is the immutable session-birth timestamp).
func TestStore_Upsert_PreservesStartedAt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	first := sampleSession("sess-b", now)
	require.NoError(t, store.Upsert(ctx, first))

	// Pretend we re-observe the session 5 minutes later, with a
	// "wrong" StartedAt. The store must NOT clobber.
	later := first
	later.StartedAt = now.Add(5 * time.Minute) // simulating a buggy observer
	later.LastActive = now.Add(5 * time.Minute)
	require.NoError(t, store.Upsert(ctx, later))

	got, err := store.Get(ctx, "sess-b")
	require.NoError(t, err)
	require.Equal(t, first.StartedAt.UnixMilli(), got.StartedAt.UnixMilli(),
		"StartedAt must be the first-observation timestamp, immutable")
	require.Equal(t, later.LastActive.UnixMilli(), got.LastActive.UnixMilli(),
		"LastActive must reflect the latest Upsert")
}

// TestStore_Upsert_GoalTextStickyOnFirstWrite — once GoalText is populated
// it doesn't get wiped by a subsequent Upsert that has GoalText="" (the
// fsLister may extract on first tail and never overwrite).
func TestStore_Upsert_GoalTextStickyOnFirstWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	first := sampleSession("sess-c", now)
	first.GoalText = "design a session monitor"
	require.NoError(t, store.Upsert(ctx, first))

	// Subsequent Upsert with empty GoalText (e.g., fsLister noticed
	// activity but didn't re-extract). Must keep the original.
	silent := first
	silent.GoalText = ""
	silent.LastActive = now.Add(5 * time.Minute)
	require.NoError(t, store.Upsert(ctx, silent))

	got, err := store.Get(ctx, "sess-c")
	require.NoError(t, err)
	require.Equal(t, "design a session monitor", got.GoalText,
		"GoalText must stick once non-empty")
}

// TestStore_Get_UnknownReturnsNotFound — typo'd ID gets ErrNotFound,
// not zero-value Session.
func TestStore_Get_UnknownReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	_, err := store.Get(ctx, "missing")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotFound))
}

// TestStore_List_ActiveOnlyByDefault — ListOptions{} returns only
// sessions with ended_at IS NULL.
func TestStore_List_ActiveOnlyByDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, store.Upsert(ctx, sampleSession("active-1", now)))
	require.NoError(t, store.Upsert(ctx, sampleSession("active-2", now.Add(-1*time.Hour))))

	ended := sampleSession("ended-1", now.Add(-2*time.Hour))
	require.NoError(t, store.Upsert(ctx, ended))
	endedAt := now.Add(-1 * time.Hour)
	require.NoError(t, store.SetEndedAt(ctx, "ended-1", &endedAt))

	got, err := store.List(ctx, ListOptions{})
	require.NoError(t, err)
	require.Len(t, got, 2, "default List excludes ended sessions")
	for _, s := range got {
		require.NotEqual(t, "ended-1", s.ID)
	}
}

// TestStore_List_IncludeEndedReturnsAll — IncludeEnded=true brings the
// ended row back.
func TestStore_List_IncludeEndedReturnsAll(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, store.Upsert(ctx, sampleSession("a", now)))
	require.NoError(t, store.Upsert(ctx, sampleSession("b", now)))
	ended := now.Add(-30 * time.Minute)
	require.NoError(t, store.SetEndedAt(ctx, "b", &ended))

	got, err := store.List(ctx, ListOptions{IncludeEnded: true})
	require.NoError(t, err)
	require.Len(t, got, 2)
}

// TestStore_List_SortedByLastActiveDesc — most recent first.
func TestStore_List_SortedByLastActiveDesc(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, store.Upsert(ctx, sampleSession("oldest", now.Add(-2*time.Hour))))
	require.NoError(t, store.Upsert(ctx, sampleSession("newest", now)))
	require.NoError(t, store.Upsert(ctx, sampleSession("middle", now.Add(-1*time.Hour))))

	got, err := store.List(ctx, ListOptions{})
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, "newest", got[0].ID)
	require.Equal(t, "middle", got[1].ID)
	require.Equal(t, "oldest", got[2].ID)
}

// TestStore_SetEndedAt_ClearOnResume — passing nil clears the marker
// (session resumed activity).
func TestStore_SetEndedAt_ClearOnResume(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, store.Upsert(ctx, sampleSession("c", now.Add(-2*time.Hour))))
	endedAt := now.Add(-1 * time.Hour)
	require.NoError(t, store.SetEndedAt(ctx, "c", &endedAt))

	got, err := store.Get(ctx, "c")
	require.NoError(t, err)
	require.NotNil(t, got.EndedAt)

	// Resume.
	require.NoError(t, store.SetEndedAt(ctx, "c", nil))
	got, err = store.Get(ctx, "c")
	require.NoError(t, err)
	require.Nil(t, got.EndedAt, "passing nil to SetEndedAt must clear")
}

// TestStore_PurgeBefore_ActiveSessionsPreserved — active sessions
// (ended_at IS NULL) survive Purge regardless of LastActive.
func TestStore_PurgeBefore_ActiveSessionsPreserved(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	// An *active* session whose last_active is way older than cutoff —
	// must not be purged.
	require.NoError(t, store.Upsert(ctx, sampleSession("stale-active", now.Add(-7*24*time.Hour))))
	// An ended session, old.
	require.NoError(t, store.Upsert(ctx, sampleSession("stale-ended", now.Add(-7*24*time.Hour))))
	endedLongAgo := now.Add(-7 * 24 * time.Hour)
	require.NoError(t, store.SetEndedAt(ctx, "stale-ended", &endedLongAgo))

	count, err := store.PurgeBefore(ctx, now.Add(-24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, int64(1), count, "only the ended session should be purged")

	// Active still there.
	_, err = store.Get(ctx, "stale-active")
	require.NoError(t, err)
	_, err = store.Get(ctx, "stale-ended")
	require.True(t, errors.Is(err, ErrNotFound))
}

// TestStore_CountBefore_MatchesPurge — dry-run preview agrees with
// what Purge would delete.
func TestStore_CountBefore_MatchesPurge(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	for _, id := range []string{"e1", "e2", "e3"} {
		require.NoError(t, store.Upsert(ctx, sampleSession(id, now.Add(-7*24*time.Hour))))
		ended := now.Add(-3 * 24 * time.Hour)
		require.NoError(t, store.SetEndedAt(ctx, id, &ended))
	}

	cutoff := now.Add(-24 * time.Hour)
	count, err := store.CountBefore(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, int64(3), count)

	purged, err := store.PurgeBefore(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, count, purged, "preview must match actual purge")
}

// TestStore_Delete_UnknownReturnsNotFound — Delete on missing id surfaces
// ErrNotFound (not silent success).
func TestStore_Delete_UnknownReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	err := store.Delete(ctx, "ghost")
	require.True(t, errors.Is(err, ErrNotFound))
}

// TestStore_List_SinceFilter — Since cutoff hides older rows.
func TestStore_List_SinceFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	require.NoError(t, store.Upsert(ctx, sampleSession("recent", now)))
	require.NoError(t, store.Upsert(ctx, sampleSession("stale", now.Add(-48*time.Hour))))

	got, err := store.List(ctx, ListOptions{Since: now.Add(-24 * time.Hour)})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "recent", got[0].ID)
}
