// Package notify implements the notification delivery layer. It
// surfaces Notifiable items (the advisor is the current producer;
// future producers attach by satisfying the same interface) to
// OS-level / out-of-band channels. Four channels ship out of the box:
//
//   - desktop   : macOS osascript / Linux notify-send.
//   - webhook   : HTTP POST/PUT/PATCH + custom headers (lifted from
//     agent/postWebhook).
//   - tui-banner: TUI ModeList top section reads recent advisories.
//   - shell     : `buddy notify --prompt` emits a single-line string
//     for PS1.
//
// The Dispatcher orchestrates severity filter + per-channel dedup +
// channel calls; each Channel implementation is trivial (one-shot
// fire-and-record). Daemon runAdvisorMonitor invokes Dispatch after
// each Persist.
package notify

import "time"

// Severity mirrors advisor.Severity but lives here too so the notify
// package doesn't have to depend on advisor for the filter constants.
// Channels accept ChannelConfig.SeverityMin as a string ("info" /
// "warn" / "high") to keep the buddy config layer simple.
type Severity string

const (
	SeverityInfo Severity = "info"
	SeverityWarn Severity = "warn"
	SeverityHigh Severity = "high"
)

// severityRank maps each severity to a comparable integer. Used by
// the filter check (advisory.severity >= channel.min).
func severityRank(s Severity) int {
	switch s {
	case SeverityHigh:
		return 3
	case SeverityWarn:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// Channel name constants — also the values written to the
// notification_log.channel column.
const (
	ChannelDesktop   = "desktop"
	ChannelWebhook   = "webhook"
	ChannelTUIBanner = "tui-banner"
	ChannelShell     = "shell"
)

// Outcome values for notification_log.outcome.
const (
	OutcomeSent            = "sent"
	OutcomeSkippedSeverity = "skipped-severity"
	OutcomeSkippedDedup    = "skipped-dedup"
	OutcomeError           = "error"
)

// Notification is the in-memory payload passed from advisor → channel.
// All channels receive the same struct; each formats it differently.
type Notification struct {
	AdvisoryID int64
	Kind       string
	Severity   Severity
	Title      string // short header (rendered in desktop title / TUI banner row)
	Body       string // longer prose (advisor.Advisory.Message verbatim)
	CreatedAt  time.Time
}

// ChannelConfig is the per-channel knobs the buddy config carries.
// Channel implementations consume these — the Dispatcher does the
// severity + dedup gating before calling Send().
type ChannelConfig struct {
	Enabled     bool
	SeverityMin Severity
	DedupWindow time.Duration
}

// WebhookConfig is one entry in notifyWebhooks. Each webhook is an
// independent destination with its own throttles.
type WebhookConfig struct {
	URL         string
	Method      string            // default POST
	Headers     map[string]string // optional
	SeverityMin Severity          // default info
	DedupWindow time.Duration     // default 1h
	Timeout     time.Duration     // default 30s
}

// Notifiable is the contract the dispatcher expects from a single
// alert item. The advisor is the canonical producer (advisor.Advisory
// satisfies it), but the dispatcher does not import the advisor
// package: any future source — daemon health alerts, agent run
// failures, manual user injection — can attach by implementing the
// same six accessors. Severity is exchanged as the raw string the
// producer stores (the dispatcher converts to its own typed Severity
// at the edge) so the producer side stays free of notify imports.
type Notifiable interface {
	NotifyID() int64
	NotifyKind() string
	NotifySeverity() string
	NotifyTitle() string
	NotifyBody() string
	NotifyCreatedAt() time.Time
	NotifyMuted() bool
}

// LogRow is the read-back shape of a notification_log entry.
type LogRow struct {
	ID         int64
	AdvisoryID int64
	Channel    string
	Kind       string
	Severity   Severity
	SentAt     time.Time
	Outcome    string
	Detail     string
}
