package notify

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"database/sql"
)

// Dispatcher orchestrates the per-channel filter + dedup + send loop.
// One Dispatcher per daemon process; ChannelEntries are added at
// startup based on user config.
//
// Dispatch is intentionally synchronous within a single advisor batch:
// channels are tried in registration order (desktop / webhook(s) /
// tui-banner / shell), each respecting its own dedup window. We don't
// fan out goroutines because the channel set is tiny and HTTP timeouts
// already cap latency.
type Dispatcher struct {
	store    *Store
	entries  []channelEntry

	// Now is injectable for tests. Production callers leave nil.
	Now func() time.Time
}

// channelEntry pairs a Channel with its per-channel config.
type channelEntry struct {
	ch     Channel
	config ChannelConfig
}

// NewDispatcher constructs the empty dispatcher; callers add channels
// via AddChannel before Dispatch.
func NewDispatcher(store *Store) *Dispatcher {
	return &Dispatcher{store: store}
}

// AddChannel registers a channel + its per-channel config. Repeated
// AddChannel of the same Name() is allowed — useful for multiple
// webhook destinations.
func (d *Dispatcher) AddChannel(ch Channel, c ChannelConfig) {
	d.entries = append(d.entries, channelEntry{ch: ch, config: c})
}

func (d *Dispatcher) now() time.Time {
	if d.Now != nil {
		return d.Now().UTC()
	}
	return time.Now().UTC()
}

// Dispatch fans the advisories out through every registered channel.
// Per-channel severity floor + dedup window are enforced; outcomes
// (sent / skipped-severity / skipped-dedup / error) are recorded in
// notification_log so users can debug "why didn't I get pinged".
//
// Errors on individual channels are swallowed (logged via outcome).
// Returns a count summary keyed by channel name → sent count.
func (d *Dispatcher) Dispatch(ctx context.Context, items []Notifiable) map[string]int {
	sentCounts := map[string]int{}
	if d.store == nil || len(items) == 0 || len(d.entries) == 0 {
		return sentCounts
	}
	now := d.now()
	for _, it := range items {
		if it.NotifyMuted() {
			continue
		}
		id := it.NotifyID()
		kind := it.NotifyKind()
		sev := Severity(it.NotifySeverity())
		n := Notification{
			AdvisoryID: id,
			Kind:       kind,
			Severity:   sev,
			Title:      it.NotifyTitle(),
			Body:       it.NotifyBody(),
			CreatedAt:  it.NotifyCreatedAt(),
		}
		for _, e := range d.entries {
			if !e.config.Enabled {
				continue
			}
			if severityRank(n.Severity) < severityRank(e.config.SeverityMin) {
				_, _ = d.store.Insert(ctx, LogRow{
					AdvisoryID: id, Channel: e.ch.Name(), Kind: kind,
					Severity: n.Severity, SentAt: now,
					Outcome: OutcomeSkippedSeverity,
					Detail:  fmt.Sprintf("min=%s adv=%s", e.config.SeverityMin, n.Severity),
				})
				continue
			}
			if d.skipForDedup(ctx, e.ch.Name(), kind, e.config.DedupWindow, now) {
				_, _ = d.store.Insert(ctx, LogRow{
					AdvisoryID: id, Channel: e.ch.Name(), Kind: kind,
					Severity: n.Severity, SentAt: now,
					Outcome: OutcomeSkippedDedup,
					Detail:  fmt.Sprintf("window=%s", e.config.DedupWindow),
				})
				continue
			}
			outcome := OutcomeSent
			detail := ""
			if err := e.ch.Send(ctx, n); err != nil {
				outcome = OutcomeError
				detail = err.Error()
			} else {
				sentCounts[e.ch.Name()]++
			}
			_, _ = d.store.Insert(ctx, LogRow{
				AdvisoryID: id, Channel: e.ch.Name(), Kind: kind,
				Severity: n.Severity, SentAt: now,
				Outcome: outcome, Detail: detail,
			})
		}
	}
	return sentCounts
}

// skipForDedup looks up the most recent "sent" row for (channel,kind)
// and returns true when its sent_at is within the dedup window.
// Window <= 0 disables dedup for this channel.
//
// On lookup error the default is "skip" — assume a duplicate hit
// rather than risk a notification storm if the store is unhappy. A
// suppressed real notification is recoverable (the next tick re-fires
// once the window passes); spamming the user during a transient DB
// hiccup is not. The error is logged so the cause stays visible.
func (d *Dispatcher) skipForDedup(ctx context.Context, channel, kind string, window time.Duration, now time.Time) bool {
	if window <= 0 {
		return false
	}
	last, err := d.store.LastSent(ctx, channel, kind)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		log.Printf("notify: dedup lookup failed for channel=%s kind=%s: %v — assuming dedup hit", channel, kind, err)
		return true
	}
	return now.Sub(last.SentAt) < window
}

// RenderTitle builds the short title channels use (desktop title,
// banner row, webhook payload). Renders the severity glyph + kind so
// the user can scan at a glance. Producers call this from their
// NotifyTitle() implementation so every channel sees the same shape
// regardless of which producer emitted the item.
func RenderTitle(kind, severity string) string {
	glyph := "·"
	switch Severity(severity) {
	case SeverityHigh:
		glyph = "⚠"
	case SeverityWarn:
		glyph = "!"
	case SeverityInfo:
		glyph = "i"
	}
	return fmt.Sprintf("%s buddy %s", glyph, kind)
}
