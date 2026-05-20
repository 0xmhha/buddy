package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Channel is the single-method delivery contract every transport
// implements. Send is best-effort — error returns are non-fatal to
// the dispatcher (just get recorded as outcome="error" rows in the
// notification_log).
type Channel interface {
	Name() string
	Send(ctx context.Context, n Notification) error
}

// ─── desktop ─────────────────────────────────────────────────────────

// DesktopChannel invokes the host OS notification tool. macOS uses
// `osascript -e 'display notification ...'`; Linux uses `notify-send`.
// Other OSes register but Send returns ErrUnsupported (dispatcher
// records this as outcome="error" once and skips on subsequent ticks).
type DesktopChannel struct {
	// Bin overrides the auto-detected binary path. Empty → defaults
	// (osascript on darwin, notify-send on linux).
	Bin string
	// Runner is injectable so tests can intercept exec without going
	// to the real OS. Defaults to exec.CommandContext.
	Runner func(ctx context.Context, name string, args ...string) *exec.Cmd
}

// ErrUnsupported indicates the desktop channel doesn't support the
// current OS (Windows / other Unix). Dispatcher logs this once and
// keeps on dispatching to the remaining channels.
var ErrUnsupported = errors.New("notify: desktop channel not supported on this OS")

// NewDesktopChannel returns a DesktopChannel with platform-aware
// defaults applied.
func NewDesktopChannel() *DesktopChannel {
	d := &DesktopChannel{}
	if runtime.GOOS == "darwin" {
		d.Bin = "osascript"
	} else if runtime.GOOS == "linux" {
		d.Bin = "notify-send"
	}
	return d
}

// Name implements Channel.
func (d *DesktopChannel) Name() string { return ChannelDesktop }

// Send dispatches the notification via the chosen OS tool. Best-effort.
func (d *DesktopChannel) Send(ctx context.Context, n Notification) error {
	if d.Bin == "" {
		return ErrUnsupported
	}
	if _, err := exec.LookPath(d.Bin); err != nil {
		return fmt.Errorf("notify desktop: %s not in PATH: %w", d.Bin, err)
	}
	runner := d.Runner
	if runner == nil {
		runner = exec.CommandContext
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(
			`display notification %q with title %q`,
			n.Body, n.Title,
		)
		cmd = runner(ctx, d.Bin, "-e", script)
	case "linux":
		cmd = runner(ctx, d.Bin, n.Title, n.Body)
	default:
		return ErrUnsupported
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("notify desktop: %s: %w (output: %s)", d.Bin, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ─── webhook ─────────────────────────────────────────────────────────

// WebhookChannel POSTs (or PUT/PATCH) the notification as JSON to a
// single destination. One WebhookChannel per destination; the
// Dispatcher iterates them. Lifts the agent/postWebhook pattern with
// notification-specific JSON shape.
type WebhookChannel struct {
	Config WebhookConfig
}

// NewWebhookChannel constructs the channel with defaults applied to
// the missing config fields.
func NewWebhookChannel(c WebhookConfig) *WebhookChannel {
	if c.Method == "" {
		c.Method = http.MethodPost
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	return &WebhookChannel{Config: c}
}

// Name implements Channel.
func (w *WebhookChannel) Name() string { return ChannelWebhook }

// payload is the on-wire shape posted to each webhook endpoint.
// Keeping it small + flat so downstream integrations (Slack
// Incoming Webhook, Discord, custom) can map fields trivially.
type webhookPayload struct {
	AdvisoryID int64     `json:"advisory_id"`
	Kind       string    `json:"kind"`
	Severity   string    `json:"severity"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

// Send implements Channel. Errors are surface-level — the dispatcher
// records the failure but doesn't fail the whole dispatch loop.
func (w *WebhookChannel) Send(ctx context.Context, n Notification) error {
	if strings.TrimSpace(w.Config.URL) == "" {
		return errors.New("notify webhook: URL empty")
	}
	body, err := json.Marshal(webhookPayload{
		AdvisoryID: n.AdvisoryID, Kind: n.Kind, Severity: string(n.Severity),
		Title: n.Title, Body: n.Body, CreatedAt: n.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("notify webhook: marshal: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, w.Config.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, w.Config.Method, w.Config.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify webhook: request: %w", err)
	}
	if _, hasCT := w.Config.Headers["Content-Type"]; !hasCT {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range w.Config.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: w.Config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("notify webhook: %s %s: %w", w.Config.Method, w.Config.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("notify webhook: %s %s returned %s: %s",
			w.Config.Method, w.Config.URL, resp.Status, strings.TrimSpace(string(preview)))
	}
	return nil
}

// ─── shell prompt ────────────────────────────────────────────────────

// ShellPromptChannel is unusual — it doesn't proactively deliver. The
// real delivery happens when the user's PS1 calls `buddy notify
// --prompt` and the CLI reads recent notifications. So Send here is
// a no-op success — the dispatch is logged as "sent" so the dedup
// window still applies (preventing the shell-channel from re-reading
// the same advisory on the next tick).
type ShellPromptChannel struct{}

// NewShellPromptChannel constructor for symmetry.
func NewShellPromptChannel() *ShellPromptChannel { return &ShellPromptChannel{} }

// Name implements Channel.
func (s *ShellPromptChannel) Name() string { return ChannelShell }

// Send is a no-op success. The real surface is `buddy notify --prompt`.
func (s *ShellPromptChannel) Send(_ context.Context, _ Notification) error {
	return nil
}

// ─── tui banner ──────────────────────────────────────────────────────

// TUIBannerChannel is also indirect — the TUI ModeList renderer reads
// notification_log on each frame and surfaces unmuted recent entries.
// Send marks the dispatch as logged so the dedup window applies; the
// TUI's NotifyFetcher does the actual rendering.
type TUIBannerChannel struct{}

// NewTUIBannerChannel constructor for symmetry.
func NewTUIBannerChannel() *TUIBannerChannel { return &TUIBannerChannel{} }

// Name implements Channel.
func (t *TUIBannerChannel) Name() string { return ChannelTUIBanner }

// Send no-op (see channel doc).
func (t *TUIBannerChannel) Send(_ context.Context, _ Notification) error {
	return nil
}
