package notify

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

// fakeNotifiable is a minimal Notifiable test value. The dispatcher
// test suite uses it instead of the real advisor.Advisory so the
// notify package's tests do not depend on the advisor package — that
// import direction is reserved for the production adapter.
type fakeNotifiable struct {
	id        int64
	kind      string
	severity  Severity
	muted     bool
	createdAt time.Time
}

func (f fakeNotifiable) NotifyID() int64           { return f.id }
func (f fakeNotifiable) NotifyKind() string         { return f.kind }
func (f fakeNotifiable) NotifySeverity() string     { return string(f.severity) }
func (f fakeNotifiable) NotifyTitle() string        { return RenderTitle(f.kind, string(f.severity)) }
func (f fakeNotifiable) NotifyBody() string         { return "msg-" + f.kind }
func (f fakeNotifiable) NotifyCreatedAt() time.Time { return f.createdAt }
func (f fakeNotifiable) NotifyMuted() bool          { return f.muted }

func makeAdv(id int64, kind string, sev Severity) fakeNotifiable {
	return fakeNotifiable{
		id: id, kind: kind, severity: sev,
		createdAt: time.Now().UTC(),
	}
}

// seedAdv inserts a stub advisory row so notification_log.advisory_id
// FK resolves. The dispatcher's Dispatch records every per-channel
// attempt into notification_log, and the FK references an advisories
// row by id.
func seedAdv(t *testing.T, as *advisoryStub, kind string, sev Severity) int64 {
	t.Helper()
	id, err := as.Insert(context.Background(), advisoryStubRow{
		Kind: kind, Severity: string(sev), Message: "msg-" + kind,
		CreatedAt: time.Now().UTC(),
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

	low := makeAdv(seedAdv(t, as, "low", SeverityInfo), "low", SeverityInfo)
	hi := makeAdv(seedAdv(t, as, "hi", SeverityHigh), "hi", SeverityHigh)
	got := d.Dispatch(context.Background(), []Notifiable{low, hi})
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

	first := makeAdv(seedAdv(t, as, "k", SeverityWarn), "k", SeverityWarn)
	_ = d.Dispatch(context.Background(), []Notifiable{first})
	second := makeAdv(seedAdv(t, as, "k", SeverityWarn), "k", SeverityWarn)
	_ = d.Dispatch(context.Background(), []Notifiable{second})
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

	first := makeAdv(seedAdv(t, as, "k", SeverityWarn), "k", SeverityWarn)
	_ = d.Dispatch(context.Background(), []Notifiable{first})

	// Advance the dispatcher's view of "now" past the window.
	d.Now = func() time.Time { return now.Add(time.Hour) }
	second := makeAdv(seedAdv(t, as, "k", SeverityWarn), "k", SeverityWarn)
	_ = d.Dispatch(context.Background(), []Notifiable{second})
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

	adv := makeAdv(seedAdv(t, as, "k", SeverityWarn), "k", SeverityWarn)
	got := d.Dispatch(context.Background(), []Notifiable{adv})
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

	adv := makeAdv(seedAdv(t, as, "k", SeverityWarn), "k", SeverityWarn)
	got := d.Dispatch(context.Background(), []Notifiable{adv})
	require.Empty(t, got)
	require.Zero(t, ch.calls())
}

func TestDispatcher_MutedAdvisorySkipped(t *testing.T) {
	t.Parallel()
	store, as := newTestStore(t)
	d := NewDispatcher(store)
	ch := &recordingChannel{name: ChannelDesktop}
	d.AddChannel(ch, ChannelConfig{Enabled: true, SeverityMin: SeverityInfo})

	muted := makeAdv(seedAdv(t, as, "k", SeverityHigh), "k", SeverityHigh)
	muted.muted = true
	_ = d.Dispatch(context.Background(), []Notifiable{muted})
	require.Zero(t, ch.calls())
}

// TestSkipForDedup_TreatsLookupErrorAsDedupHit — skipForDedup must
// suppress dispatch (return true) when the store's LastSent fails for
// any reason other than ErrNoRows. The earlier default of "no dedup hit"
// turned a transient DB hiccup into a notification storm — a suppressed
// real notification is recoverable on the next tick, a storm is not.
func TestSkipForDedup_TreatsLookupErrorAsDedupHit(t *testing.T) {
	t.Parallel()
	store, _ := newTestStore(t)
	// Closing the underlying DB makes every subsequent LastSent call
	// return a driver-level error (not sql.ErrNoRows), which is the
	// case the inverted default exists to handle. The Store carries
	// the connection as a db.Conn interface, so this test pulls the
	// concrete *sql.DB out via a type assertion before closing.
	require.NoError(t, store.db.(*sql.DB).Close())

	d := &Dispatcher{store: store}
	got := d.skipForDedup(context.Background(), ChannelDesktop, "k",
		time.Hour, time.Now().UTC())
	require.True(t, got,
		"lookup error must default to skip (assume dedup hit)")
}

// TestSkipForDedup_NoRowsAllowsDispatch — the no-prior-row case must
// stay on the fire side; only genuine errors flip to skip. Otherwise
// the very first notification for a (channel, kind) would never fire.
func TestSkipForDedup_NoRowsAllowsDispatch(t *testing.T) {
	t.Parallel()
	store, _ := newTestStore(t)
	d := &Dispatcher{store: store}

	got := d.skipForDedup(context.Background(), ChannelDesktop, "fresh-kind",
		time.Hour, time.Now().UTC())
	require.False(t, got,
		"no prior row must allow dispatch (no dedup hit)")
}

func TestRenderTitle_GlyphPerSeverity(t *testing.T) {
	t.Parallel()
	require.Contains(t, RenderTitle("k", string(SeverityHigh)), "⚠")
	require.Contains(t, RenderTitle("k", string(SeverityWarn)), "!")
	require.Contains(t, RenderTitle("k", string(SeverityInfo)), "i")
}
