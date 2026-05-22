package advisor

import (
	"time"

	"github.com/0xmhha/buddy/internal/notify"
)

// Advisory satisfies notify.Notifiable so the dispatcher can carry
// advisor output without importing this package. The interface is
// defined on the consumer side (notify), and the producer (advisor)
// adapts to it — the layer direction matches "delivery layer depends
// on nothing; generation layers attach as producers". Future
// producers (daemon health monitor, agent run failures, manual user
// injections) implement the same six methods and route through the
// same dispatcher.

// NotifyID returns the row id used by notification_log for dedup +
// audit. Zero-value advisories (in-memory only, never persisted)
// surface as 0; the dispatcher tolerates that because notification_log
// stores the id verbatim alongside the channel name + kind.
func (a Advisory) NotifyID() int64 { return a.ID }

// NotifyKind returns the rule / source identifier the dispatcher uses
// as the dedup key (channel, kind).
func (a Advisory) NotifyKind() string { return a.Kind }

// NotifySeverity returns the raw severity string the producer stores.
// The dispatcher converts it into its own typed Severity at the edge,
// so this side stays free of notify imports for the severity type.
func (a Advisory) NotifySeverity() string { return string(a.Severity) }

// NotifyTitle returns the short title channels render (glyph + kind).
// Rendering goes through notify.RenderTitle so the format stays
// consistent across producers.
func (a Advisory) NotifyTitle() string {
	return notify.RenderTitle(a.Kind, string(a.Severity))
}

// NotifyBody returns the long-form prose channels render as the
// notification body. The advisor stores friend-tone Korean here.
func (a Advisory) NotifyBody() string { return a.Message }

// NotifyCreatedAt returns the moment the advisory was generated.
// Channels render this as a relative timestamp ("3분 전") or absolute
// ISO depending on the surface.
func (a Advisory) NotifyCreatedAt() time.Time { return a.CreatedAt }

// NotifyMuted lets a producer suppress dispatch without removing the
// record (the advisor exposes a mute UX so users can silence specific
// kinds without losing the history).
func (a Advisory) NotifyMuted() bool { return a.Muted }
