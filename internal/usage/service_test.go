package usage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/schema"
	"github.com/0xmhha/buddy/internal/sessions"
)

// newTestService creates a fresh on-disk SQLite DB (so the sessions
// table exists via migrations) and returns a Service alongside the
// sessions.Store helper for seeding.
func newTestService(t *testing.T) (*Service, *sessions.Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "usage-test.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return NewService(conn), sessions.NewStore(conn)
}

func seedSession(t *testing.T, store *sessions.Store, id string, started time.Time, durMs int64, usage schema.TokenUsage, goal string, ended bool) {
	t.Helper()
	sess := sessions.Session{
		ID:             id,
		PID:            1,
		TranscriptPath: "/tmp/" + id + ".jsonl",
		StartedAt:      started,
		LastActive:     started.Add(time.Duration(durMs) * time.Millisecond),
		Usage:          usage,
		LastOffset:     1024,
		GoalText:       goal,
		Metadata:       "{}",
	}
	require.NoError(t, store.Upsert(context.Background(), sess))
	if ended {
		endedAt := sess.LastActive
		require.NoError(t, store.SetEndedAt(context.Background(), id, &endedAt))
	}
}

// TestService_QueryTokenSpend_AllTime — empty Since means sum all.
func TestService_QueryTokenSpend_AllTime(t *testing.T) {
	t.Parallel()
	svc, store := newTestService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	seedSession(t, store, "a", now.Add(-2*time.Hour), 600_000,
		schema.TokenUsage{InputTokens: 100, OutputTokens: 200, CacheReadTokens: 50, CacheCreateTokens: 10}, "g", false)
	seedSession(t, store, "b", now.Add(-1*time.Hour), 300_000,
		schema.TokenUsage{InputTokens: 300, OutputTokens: 400, CacheReadTokens: 100, CacheCreateTokens: 20}, "g", false)

	got, err := svc.QueryTokenSpend(context.Background(), TimeWindow{})
	require.NoError(t, err)
	require.Equal(t, int64(400), got.InputTokens)
	require.Equal(t, int64(600), got.OutputTokens)
	require.Equal(t, int64(150), got.CacheReadTokens)
	require.Equal(t, int64(30), got.CacheCreateTokens)
	require.Equal(t, int64(1180), got.TotalTokens())
}

// TestService_QueryTokenSpend_WindowFilters — Since excludes earlier
// sessions, Until excludes later ones.
func TestService_QueryTokenSpend_WindowFilters(t *testing.T) {
	t.Parallel()
	svc, store := newTestService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	seedSession(t, store, "old", now.Add(-48*time.Hour), 0,
		schema.TokenUsage{InputTokens: 1000}, "", false)
	seedSession(t, store, "recent", now.Add(-1*time.Hour), 0,
		schema.TokenUsage{InputTokens: 7}, "", false)

	got, err := svc.QueryTokenSpend(context.Background(), TimeWindow{Since: now.Add(-24 * time.Hour)})
	require.NoError(t, err)
	require.Equal(t, int64(7), got.InputTokens, "Since=24h filters out the 48h session")
}

// TestTokenSpend_CacheHitRatio — math sanity.
func TestTokenSpend_CacheHitRatio(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   TokenSpend
		want float64
	}{
		{"empty", TokenSpend{}, 0},
		{"half-cache", TokenSpend{InputTokens: 50, CacheReadTokens: 50}, 0.5},
		{"all-input", TokenSpend{InputTokens: 100}, 0},
		{"all-cache", TokenSpend{CacheReadTokens: 100}, 1.0},
		{"output-excluded", TokenSpend{InputTokens: 50, CacheReadTokens: 50, OutputTokens: 1000}, 0.5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.InDelta(t, tc.want, tc.in.CacheHitRatio(), 0.0001)
		})
	}
}

// TestService_QuerySessionStats_Counts — total/active/ended/goal.
func TestService_QuerySessionStats_Counts(t *testing.T) {
	t.Parallel()
	svc, store := newTestService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	seedSession(t, store, "act-1", now.Add(-1*time.Hour), 1000, schema.TokenUsage{}, "make a plan", false)
	seedSession(t, store, "act-2", now.Add(-30*time.Minute), 1000, schema.TokenUsage{}, "", false)
	seedSession(t, store, "end-1", now.Add(-3*time.Hour), 600_000, schema.TokenUsage{}, "shipped", true)

	stats, err := svc.QuerySessionStats(context.Background(), TimeWindow{})
	require.NoError(t, err)
	require.Equal(t, int64(3), stats.TotalSessions)
	require.Equal(t, int64(2), stats.ActiveSessions)
	require.Equal(t, int64(1), stats.EndedSessions)
	require.InDelta(t, 2.0/3.0, stats.GoalTextRatio, 0.0001)
}

// TestService_QuerySessionStats_Durations — percentiles over ended sessions.
func TestService_QuerySessionStats_Durations(t *testing.T) {
	t.Parallel()
	svc, store := newTestService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	// 10 ended sessions, durations 1m..10m
	for i := 1; i <= 10; i++ {
		id := "ended-" + string(rune('a'+i-1))
		seedSession(t, store, id, now.Add(-time.Duration(i)*time.Hour),
			int64(i)*60_000, schema.TokenUsage{}, "g", true)
	}
	stats, err := svc.QuerySessionStats(context.Background(), TimeWindow{})
	require.NoError(t, err)
	require.Equal(t, int64(10), stats.EndedSessions)
	// p50 of [1m..10m] nearest-rank at idx (10-1)*0.5 = 4 → durs[4] = 5m
	require.Equal(t, 5*time.Minute, stats.DurationP50)
	// p90 at idx 8 → 9m
	require.Equal(t, 9*time.Minute, stats.DurationP90)
	require.Equal(t, 10*time.Minute, stats.DurationMax)
}

// TestService_QueryTimeDistribution_Buckets — hour buckets local time.
func TestService_QueryTimeDistribution_Buckets(t *testing.T) {
	t.Parallel()
	svc, store := newTestService(t)

	// Pick a known local instant (hour 10 local).
	loc := time.Local
	base := time.Date(2026, 5, 10, 10, 30, 0, 0, loc).UTC()
	seedSession(t, store, "s1", base, 0, schema.TokenUsage{}, "g", false)
	seedSession(t, store, "s2", base.Add(2*time.Hour), 0, schema.TokenUsage{}, "g", false) // hour 12 local
	seedSession(t, store, "s3", base.Add(2*time.Hour), 0, schema.TokenUsage{}, "g", false)

	td, err := svc.QueryTimeDistribution(context.Background(), TimeWindow{})
	require.NoError(t, err)
	require.Equal(t, int64(3), td.Total)
	hourA := base.Local().Hour()
	hourB := base.Add(2 * time.Hour).Local().Hour()
	require.Equal(t, int64(1), td.HourCounts[hourA])
	require.Equal(t, int64(2), td.HourCounts[hourB])
}

// TestService_QueryTopSessions_OrderedAndLimited — DESC + limit.
func TestService_QueryTopSessions_OrderedAndLimited(t *testing.T) {
	t.Parallel()
	svc, store := newTestService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	seedSession(t, store, "small", now, 0,
		schema.TokenUsage{InputTokens: 10}, "small", false)
	seedSession(t, store, "huge", now, 0,
		schema.TokenUsage{InputTokens: 10000, OutputTokens: 5000}, "huge", false)
	seedSession(t, store, "mid", now, 0,
		schema.TokenUsage{InputTokens: 1000}, "mid", false)

	top, err := svc.QueryTopSessions(context.Background(), TimeWindow{}, 2)
	require.NoError(t, err)
	require.Len(t, top, 2)
	require.Equal(t, "huge", top[0].ID)
	require.Equal(t, "mid", top[1].ID)
	require.Equal(t, int64(15000), top[0].TotalTokens)
}

// TestService_QueryOverview_Composition — overview hits all 4 metrics.
func TestService_QueryOverview_Composition(t *testing.T) {
	t.Parallel()
	svc, store := newTestService(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	seedSession(t, store, "s1", now.Add(-1*time.Hour), 600_000,
		schema.TokenUsage{InputTokens: 100, OutputTokens: 200}, "build a thing", true)

	ov, err := svc.QueryOverview(context.Background(), TimeWindow{}, 5)
	require.NoError(t, err)
	require.Equal(t, int64(300), ov.Spend.TotalTokens())
	require.Equal(t, int64(1), ov.Stats.TotalSessions)
	require.Equal(t, int64(1), ov.Stats.EndedSessions)
	require.Equal(t, int64(1), ov.TimeDistribution.Total)
	require.Len(t, ov.Top, 1)
	require.Equal(t, "s1", ov.Top[0].ID)
}

// TestTimeWindow_Helpers — IsAllTime + Duration sanity.
func TestTimeWindow_Helpers(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	require.True(t, TimeWindow{}.IsAllTime())
	require.False(t, TimeWindow{Since: now}.IsAllTime())
	require.Equal(t, time.Duration(0), TimeWindow{}.Duration())
	require.Equal(t, time.Hour, TimeWindow{Since: now.Add(-time.Hour), Until: now}.Duration())
	require.Equal(t, time.Duration(0), TimeWindow{Since: now, Until: now.Add(-time.Hour)}.Duration())
}
