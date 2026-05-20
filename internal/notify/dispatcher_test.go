package notify

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/advisor"
)

// recordingChannel captures every Send call so tests can assert order.
type recordingChannel struct {
	name string
	mu   sync.Mutex
	got  []Notification
	err  error
}

func (r *recordingChannel) Name() string { return r.name }
func (r *recordingChannel) Send(_ context.Context, n Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.got = append(r.got, n)
	return r.err
}
func (r *recordingChannel) calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.got)
}

func makeAdv(id int64, kind string, sev advisor.Severity) advisor.Advisory {
	return advisor.Advisory{
		ID: id, Kind: kind, Severity: sev,
		Message: "msg-" + kind, CreatedAt: time.Now().UTC(),
	}
}

// seedAdv inserts an Advisory so notification_log.advisory_id FK
// resolves. Test helpers in this file call this for any AdvisoryID
// they want dispatcher.Dispatch to log against.
func seedAdv(t *testing.T, as *advisor.Store, kind string, sev advisor.Severity) int64 {
	t.Helper()
	id, err := as.Insert(context.Background(), advisor.Advisory{
		Kind: kind, Severity: sev, Message: "msg-" + kind, CreatedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
	return id
}

func TestDispatcher_SeverityFilterSkips(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	d := NewDispatcher(store)
	ch := &recordingChannel{name: ChannelDesktop}
	d.AddChannel(ch, ChannelConfig{Enabled: true, SeverityMin: SeverityHigh})

	low := makeAdv(seedAdv(t, as, "low", advisor.SeverityInfo), "low", advisor.SeverityInfo)
	hi := makeAdv(seedAdv(t, as, "hi", advisor.SeverityHigh), "hi", advisor.SeverityHigh)
	got := d.Dispatch(context.Background(), []advisor.Advisory{low, hi})
	require.Equal(t, map[string]int{ChannelDesktop: 1}, got)
	require.Equal(t, 1, ch.calls(), "info advisory skipped, only high reaches channel")
}

func TestDispatcher_DedupSkipsWithinWindow(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	now := time.Now().UTC()
	d := NewDispatcher(store)
	d.Now = func() time.Time { return now }
	ch := &recordingChannel{name: ChannelDesktop}
	d.AddChannel(ch, ChannelConfig{Enabled: true, SeverityMin: SeverityInfo, DedupWindow: time.Hour})

	first := makeAdv(seedAdv(t, as, "k", advisor.SeverityWarn), "k", advisor.SeverityWarn)
	_ = d.Dispatch(context.Background(), []advisor.Advisory{first})
	second := makeAdv(seedAdv(t, as, "k", advisor.SeverityWarn), "k", advisor.SeverityWarn)
	_ = d.Dispatch(context.Background(), []advisor.Advisory{second})
	require.Equal(t, 1, ch.calls(), "dedup window blocks the second send")
}

func TestDispatcher_DedupAllowsAfterWindowExpires(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	now := time.Now().UTC()
	d := NewDispatcher(store)
	d.Now = func() time.Time { return now }
	ch := &recordingChannel{name: ChannelDesktop}
	d.AddChannel(ch, ChannelConfig{Enabled: true, SeverityMin: SeverityInfo, DedupWindow: 30 * time.Minute})

	first := makeAdv(seedAdv(t, as, "k", advisor.SeverityWarn), "k", advisor.SeverityWarn)
	_ = d.Dispatch(context.Background(), []advisor.Advisory{first})

	// Advance the dispatcher's view of "now" past the window.
	d.Now = func() time.Time { return now.Add(time.Hour) }
	second := makeAdv(seedAdv(t, as, "k", advisor.SeverityWarn), "k", advisor.SeverityWarn)
	_ = d.Dispatch(context.Background(), []advisor.Advisory{second})
	require.Equal(t, 2, ch.calls())
}

func TestDispatcher_ChannelErrorRecordedNotFatal(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	d := NewDispatcher(store)
	ch1 := &recordingChannel{name: ChannelDesktop, err: errors.New("boom")}
	ch2 := &recordingChannel{name: ChannelShell}
	d.AddChannel(ch1, ChannelConfig{Enabled: true, SeverityMin: SeverityInfo})
	d.AddChannel(ch2, ChannelConfig{Enabled: true, SeverityMin: SeverityInfo})

	adv := makeAdv(seedAdv(t, as, "k", advisor.SeverityWarn), "k", advisor.SeverityWarn)
	got := d.Dispatch(context.Background(), []advisor.Advisory{adv})
	require.Equal(t, 0, got[ChannelDesktop])
	require.Equal(t, 1, got[ChannelShell], "failure on one channel doesn't stop others")

	rows, _ := store.List(context.Background(), ListOptions{Channel: ChannelDesktop})
	require.Len(t, rows, 1)
	require.Equal(t, OutcomeError, rows[0].Outcome)
	require.Contains(t, rows[0].Detail, "boom")
}

func TestDispatcher_DisabledChannelNoop(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	d := NewDispatcher(store)
	ch := &recordingChannel{name: ChannelDesktop}
	d.AddChannel(ch, ChannelConfig{Enabled: false, SeverityMin: SeverityInfo})

	adv := makeAdv(seedAdv(t, as, "k", advisor.SeverityWarn), "k", advisor.SeverityWarn)
	got := d.Dispatch(context.Background(), []advisor.Advisory{adv})
	require.Empty(t, got)
	require.Zero(t, ch.calls())
}

func TestDispatcher_MutedAdvisorySkipped(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	d := NewDispatcher(store)
	ch := &recordingChannel{name: ChannelDesktop}
	d.AddChannel(ch, ChannelConfig{Enabled: true, SeverityMin: SeverityInfo})

	muted := makeAdv(seedAdv(t, as, "k", advisor.SeverityHigh), "k", advisor.SeverityHigh)
	muted.Muted = true
	_ = d.Dispatch(context.Background(), []advisor.Advisory{muted})
	require.Zero(t, ch.calls())
}

func TestDispatcher_TitleSpan(t *testing.T) {
	t.Parallel()
	// Title format is exposed via Notification body to channels.
	require.Contains(t, titleFor(advisor.Advisory{Kind: "k", Severity: advisor.SeverityHigh}), "⚠")
	require.Contains(t, titleFor(advisor.Advisory{Kind: "k", Severity: advisor.SeverityWarn}), "!")
	require.Contains(t, titleFor(advisor.Advisory{Kind: "k", Severity: advisor.SeverityInfo}), "i")
}
