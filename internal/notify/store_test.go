package notify

import (
	"context"
	"errors"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/db"
)

func newTestStore(t *testing.T) (*Store, *advisor.Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "notify.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return NewStore(conn), advisor.NewStore(conn)
}

func seedAdvisory(t *testing.T, as *advisor.Store) int64 {
	t.Helper()
	id, err := as.Insert(context.Background(), advisor.Advisory{
		Kind: advisor.KindTokenSpikeDay, Severity: advisor.SeverityWarn,
		Message: "test", CreatedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	return id
}

func TestStore_Insert_RoundTrip(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	advID := seedAdvisory(t, as)
	now := time.Now().UTC().Truncate(time.Millisecond)
	id, err := store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k1",
		Severity: SeverityWarn, SentAt: now, Outcome: OutcomeSent, Detail: "ok",
	})
	require.NoError(t, err)
	require.NotZero(t, id)

	rows, err := store.List(context.Background(), ListOptions{})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, ChannelDesktop, rows[0].Channel)
	require.Equal(t, OutcomeSent, rows[0].Outcome)
}

func TestStore_List_FilterByChannel(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	advID := seedAdvisory(t, as)
	now := time.Now().UTC()
	for _, ch := range []string{ChannelDesktop, ChannelWebhook, ChannelShell} {
		_, err := store.Insert(context.Background(), LogRow{
			AdvisoryID: advID, Channel: ch, Kind: "k", Severity: SeverityInfo,
			SentAt: now, Outcome: OutcomeSent,
		})
		require.NoError(t, err)
	}
	rows, err := store.List(context.Background(), ListOptions{Channel: ChannelWebhook})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, ChannelWebhook, rows[0].Channel)
}

func TestStore_List_SinceFilter(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	advID := seedAdvisory(t, as)
	now := time.Now().UTC()
	_, _ = store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k", Severity: SeverityInfo,
		SentAt: now.Add(-48 * time.Hour), Outcome: OutcomeSent,
	})
	_, _ = store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k", Severity: SeverityInfo,
		SentAt: now, Outcome: OutcomeSent,
	})
	rows, err := store.List(context.Background(), ListOptions{Since: now.Add(-24 * time.Hour)})
	require.NoError(t, err)
	require.Len(t, rows, 1)
}

func TestStore_LastSent_HappyPath(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	advID := seedAdvisory(t, as)
	now := time.Now().UTC()
	_, _ = store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k",
		Severity: SeverityWarn, SentAt: now.Add(-30 * time.Minute), Outcome: OutcomeSent,
	})
	got, err := store.LastSent(context.Background(), ChannelDesktop, "k")
	require.NoError(t, err)
	require.Equal(t, ChannelDesktop, got.Channel)
}

func TestStore_LastSent_SkipsNonSentOutcomes(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	advID := seedAdvisory(t, as)
	now := time.Now().UTC()
	// Most recent row has outcome=skipped — LastSent must skip it.
	_, _ = store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k",
		Severity: SeverityInfo, SentAt: now.Add(-1 * time.Hour), Outcome: OutcomeSent,
	})
	_, _ = store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k",
		Severity: SeverityInfo, SentAt: now, Outcome: OutcomeSkippedDedup,
	})
	got, err := store.LastSent(context.Background(), ChannelDesktop, "k")
	require.NoError(t, err)
	// Returned row should be the OLDER one (the only true 'sent').
	require.Equal(t, OutcomeSent, got.Outcome)
}

func TestStore_LastSent_NoRows(t *testing.T) {
	t.Parallel()
	store, _ := newTestStore(t)
	_, err := store.LastSent(context.Background(), ChannelDesktop, "nothing")
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestStore_DeleteOlderThan(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	advID := seedAdvisory(t, as)
	now := time.Now().UTC()
	_, _ = store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k",
		Severity: SeverityInfo, SentAt: now.Add(-48 * time.Hour), Outcome: OutcomeSent,
	})
	_, _ = store.Insert(context.Background(), LogRow{
		AdvisoryID: advID, Channel: ChannelDesktop, Kind: "k",
		Severity: SeverityInfo, SentAt: now, Outcome: OutcomeSent,
	})
	n, err := store.DeleteOlderThan(context.Background(), now.Add(-24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	left, _ := store.Count(context.Background())
	require.Equal(t, int64(1), left)
}

func TestSeverityRank_Ordering(t *testing.T) {
	t.Parallel()
	require.Less(t, severityRank(SeverityInfo), severityRank(SeverityWarn))
	require.Less(t, severityRank(SeverityWarn), severityRank(SeverityHigh))
	require.Equal(t, 0, severityRank(Severity("unknown")))
}
