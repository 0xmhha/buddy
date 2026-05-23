// Package config owns the buddy user-level configuration: ~/.buddy/config.json.
//
// Design:
//
//   - Config holds POINTER fields so JSON unmarshalling can distinguish "absent"
//     (nil) from "explicitly set, even to a zero value" (non-nil). Hand-edited
//     config.json files only override the fields they mention.
//
//   - Effective is the resolved view with non-pointer fields. The rest of the
//     codebase consumes Effective.
//
//   - Defaults are not user-facing; they ship hard-coded and a code change is
//     required to alter them.
//
//   - Validate runs on the EFFECTIVE values, so a zero Config (all defaults)
//     always passes. Invalid configs only come from explicit user overrides.
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config is the on-disk schema for ~/.buddy/config.json.
//
// All fields are pointers: a nil field means "not set in the file", which
// causes Effective() to fall back to Defaults(). This is the standard Go idiom
// for partial-override config — it lets us tell apart "user wants 0" from "user
// did not touch this knob".
type Config struct {
	HookTimeoutMs   *int64    `json:"hookTimeoutMs,omitempty"`
	HookSlowMs      *int64    `json:"hookSlowMs,omitempty"`
	HookFailRatePct *int      `json:"hookFailRatePct,omitempty"`
	OutboxBacklog   *int      `json:"outboxBacklog,omitempty"`
	NotifyChannel   *string   `json:"notifyChannel,omitempty"` // currently only "stderr" is accepted
	PollInterval    *Duration `json:"pollInterval,omitempty"`
	BatchSize       *int      `json:"batchSize,omitempty"`
	PersonaLocale   *string   `json:"personaLocale,omitempty"` // "ko" | "en"

	// SessionMonitor* configure daemon-side observation of Claude Code
	// sessions. Disabled=true skips the goroutine entirely.
	SessionMonitorDisabled       *bool     `json:"sessionMonitorDisabled,omitempty"`
	SessionMonitorPollInterval   *Duration `json:"sessionMonitorPollInterval,omitempty"`
	SessionMonitorEndedThreshold *Duration `json:"sessionMonitorEndedThreshold,omitempty"`

	// Advisor* configure the daemon-side advisory generator: rule
	// thresholds + dedup window + poll interval + master toggle.
	AdvisorDisabled              *bool     `json:"advisorDisabled,omitempty"`
	AdvisorTokenSpikeRatio       *float64  `json:"advisorTokenSpikeRatio,omitempty"`
	AdvisorLongSessionHours      *int      `json:"advisorLongSessionHours,omitempty"`
	AdvisorLowCachePct           *int      `json:"advisorLowCachePct,omitempty"`
	AdvisorSessionVolumePerDay   *int      `json:"advisorSessionVolumePerDay,omitempty"`
	AdvisorTokenDailyThreshold   *int64    `json:"advisorTokenDailyThreshold,omitempty"`
	AdvisorDedupWindow           *Duration `json:"advisorDedupWindow,omitempty"`
	AdvisorPollInterval          *Duration `json:"advisorPollInterval,omitempty"`
	AdvisorGoalDriftDisabled     *bool     `json:"advisorGoalDriftDisabled,omitempty"`
	AdvisorGoalDriftThreshold    *float64  `json:"advisorGoalDriftThreshold,omitempty"`
	AdvisorGoalDriftSampleChunks *int      `json:"advisorGoalDriftSampleChunks,omitempty"`

	// Notify* configure the daemon-side notification dispatcher:
	// per-channel toggles + severity floor + dedup window.
	NotifyDesktopEnabled       *bool     `json:"notifyDesktopEnabled,omitempty"`
	NotifyDesktopSeverityMin   *string   `json:"notifyDesktopSeverityMin,omitempty"`
	NotifyDesktopDedup         *Duration `json:"notifyDesktopDedup,omitempty"`
	NotifyTUIBannerEnabled     *bool     `json:"notifyTuiBannerEnabled,omitempty"`
	NotifyTUIBannerSeverityMin *string   `json:"notifyTuiBannerSeverityMin,omitempty"`
	NotifyShellPromptEnabled   *bool     `json:"notifyShellPromptEnabled,omitempty"`
	NotifyShellPromptSeverityMin *string `json:"notifyShellPromptSeverityMin,omitempty"`
	NotifyWebhooks             []WebhookConfigJSON `json:"notifyWebhooks,omitempty"`
}

// WebhookConfigJSON is the on-disk shape of one webhook destination.
// Empty Method defaults to POST; empty SeverityMin defaults to "info";
// zero DedupWindow defaults to 1h; zero Timeout defaults to 30s.
type WebhookConfigJSON struct {
	URL         string            `json:"url"`
	Method      string            `json:"method,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	SeverityMin string            `json:"severityMin,omitempty"`
	DedupWindow *Duration         `json:"dedupWindow,omitempty"`
	Timeout     *Duration         `json:"timeout,omitempty"`
}

// Effective is the resolved configuration with all fields populated.
// Always produced via Config.Effective(); never JSON-decoded directly.
type Effective struct {
	HookTimeoutMs   int64
	HookSlowMs      int64
	HookFailRatePct int
	OutboxBacklog   int
	NotifyChannel   string
	PollInterval    time.Duration
	BatchSize       int
	PersonaLocale   string

	SessionMonitorDisabled       bool
	SessionMonitorPollInterval   time.Duration
	SessionMonitorEndedThreshold time.Duration

	AdvisorDisabled            bool
	AdvisorTokenSpikeRatio     float64
	AdvisorLongSessionHours    int
	AdvisorLowCachePct         int
	AdvisorSessionVolumePerDay int
	AdvisorTokenDailyThreshold int64
	AdvisorDedupWindow         time.Duration
	AdvisorPollInterval        time.Duration
	AdvisorGoalDriftDisabled     bool
	AdvisorGoalDriftThreshold    float64
	AdvisorGoalDriftSampleChunks int

	NotifyDesktopEnabled         bool
	NotifyDesktopSeverityMin     string
	NotifyDesktopDedup           time.Duration
	NotifyTUIBannerEnabled       bool
	NotifyTUIBannerSeverityMin   string
	NotifyShellPromptEnabled     bool
	NotifyShellPromptSeverityMin string
	NotifyWebhooks               []EffectiveWebhook
}

// EffectiveWebhook is the resolved (defaults-applied) shape consumed
// by the notify dispatcher. Defaults: Method=POST, SeverityMin=info,
// DedupWindow=1h, Timeout=30s.
type EffectiveWebhook struct {
	URL         string
	Method      string
	Headers     map[string]string
	SeverityMin string
	DedupWindow time.Duration
	Timeout     time.Duration
}

// Defaults returns the hard-coded defaults for every Effective field. A zero
// Config (no overrides) resolves to exactly this value.
func Defaults() Effective {
	return Effective{
		HookTimeoutMs:   30_000,
		HookSlowMs:      5_000,
		HookFailRatePct: 20,
		OutboxBacklog:   1_000,
		NotifyChannel:   "stderr",
		PollInterval:    time.Second,
		BatchSize:       500,
		PersonaLocale:   "ko",

		SessionMonitorDisabled:       false,
		SessionMonitorPollInterval:   30 * time.Second,
		SessionMonitorEndedThreshold: 1 * time.Hour,

		AdvisorDisabled:            false,
		AdvisorTokenSpikeRatio:     1.5,
		AdvisorLongSessionHours:    4,
		AdvisorLowCachePct:         30,
		AdvisorSessionVolumePerDay: 20,
		AdvisorTokenDailyThreshold: 500_000,
		AdvisorDedupWindow:         24 * time.Hour,
		AdvisorPollInterval:        1 * time.Hour,
		AdvisorGoalDriftDisabled:     false,
		AdvisorGoalDriftThreshold:    0.4,
		AdvisorGoalDriftSampleChunks: 10,

		NotifyDesktopEnabled:         true,
		NotifyDesktopSeverityMin:     "warn",
		NotifyDesktopDedup:           1 * time.Hour,
		NotifyTUIBannerEnabled:       true,
		NotifyTUIBannerSeverityMin:   "info",
		NotifyShellPromptEnabled:     false,
		NotifyShellPromptSeverityMin: "warn",
		NotifyWebhooks:               nil,
	}
}

// Effective merges any non-nil fields in c on top of Defaults().
func (c Config) Effective() Effective {
	eff := Defaults()
	if c.HookTimeoutMs != nil {
		eff.HookTimeoutMs = *c.HookTimeoutMs
	}
	if c.HookSlowMs != nil {
		eff.HookSlowMs = *c.HookSlowMs
	}
	if c.HookFailRatePct != nil {
		eff.HookFailRatePct = *c.HookFailRatePct
	}
	if c.OutboxBacklog != nil {
		eff.OutboxBacklog = *c.OutboxBacklog
	}
	if c.NotifyChannel != nil {
		eff.NotifyChannel = *c.NotifyChannel
	}
	if c.PollInterval != nil {
		eff.PollInterval = c.PollInterval.Duration
	}
	if c.BatchSize != nil {
		eff.BatchSize = *c.BatchSize
	}
	if c.PersonaLocale != nil {
		eff.PersonaLocale = *c.PersonaLocale
	}
	if c.SessionMonitorDisabled != nil {
		eff.SessionMonitorDisabled = *c.SessionMonitorDisabled
	}
	if c.SessionMonitorPollInterval != nil {
		eff.SessionMonitorPollInterval = c.SessionMonitorPollInterval.Duration
	}
	if c.SessionMonitorEndedThreshold != nil {
		eff.SessionMonitorEndedThreshold = c.SessionMonitorEndedThreshold.Duration
	}
	if c.AdvisorDisabled != nil {
		eff.AdvisorDisabled = *c.AdvisorDisabled
	}
	if c.AdvisorTokenSpikeRatio != nil {
		eff.AdvisorTokenSpikeRatio = *c.AdvisorTokenSpikeRatio
	}
	if c.AdvisorLongSessionHours != nil {
		eff.AdvisorLongSessionHours = *c.AdvisorLongSessionHours
	}
	if c.AdvisorLowCachePct != nil {
		eff.AdvisorLowCachePct = *c.AdvisorLowCachePct
	}
	if c.AdvisorSessionVolumePerDay != nil {
		eff.AdvisorSessionVolumePerDay = *c.AdvisorSessionVolumePerDay
	}
	if c.AdvisorTokenDailyThreshold != nil {
		eff.AdvisorTokenDailyThreshold = *c.AdvisorTokenDailyThreshold
	}
	if c.AdvisorDedupWindow != nil {
		eff.AdvisorDedupWindow = c.AdvisorDedupWindow.Duration
	}
	if c.AdvisorPollInterval != nil {
		eff.AdvisorPollInterval = c.AdvisorPollInterval.Duration
	}
	if c.AdvisorGoalDriftDisabled != nil {
		eff.AdvisorGoalDriftDisabled = *c.AdvisorGoalDriftDisabled
	}
	if c.AdvisorGoalDriftThreshold != nil {
		eff.AdvisorGoalDriftThreshold = *c.AdvisorGoalDriftThreshold
	}
	if c.AdvisorGoalDriftSampleChunks != nil {
		eff.AdvisorGoalDriftSampleChunks = *c.AdvisorGoalDriftSampleChunks
	}
	if c.NotifyDesktopEnabled != nil {
		eff.NotifyDesktopEnabled = *c.NotifyDesktopEnabled
	}
	if c.NotifyDesktopSeverityMin != nil {
		eff.NotifyDesktopSeverityMin = *c.NotifyDesktopSeverityMin
	}
	if c.NotifyDesktopDedup != nil {
		eff.NotifyDesktopDedup = c.NotifyDesktopDedup.Duration
	}
	if c.NotifyTUIBannerEnabled != nil {
		eff.NotifyTUIBannerEnabled = *c.NotifyTUIBannerEnabled
	}
	if c.NotifyTUIBannerSeverityMin != nil {
		eff.NotifyTUIBannerSeverityMin = *c.NotifyTUIBannerSeverityMin
	}
	if c.NotifyShellPromptEnabled != nil {
		eff.NotifyShellPromptEnabled = *c.NotifyShellPromptEnabled
	}
	if c.NotifyShellPromptSeverityMin != nil {
		eff.NotifyShellPromptSeverityMin = *c.NotifyShellPromptSeverityMin
	}
	if len(c.NotifyWebhooks) > 0 {
		out := make([]EffectiveWebhook, 0, len(c.NotifyWebhooks))
		for _, w := range c.NotifyWebhooks {
			ew := EffectiveWebhook{
				URL:         w.URL,
				Method:      w.Method,
				Headers:     w.Headers,
				SeverityMin: w.SeverityMin,
			}
			if ew.Method == "" {
				ew.Method = "POST"
			}
			if ew.SeverityMin == "" {
				ew.SeverityMin = "info"
			}
			if w.DedupWindow != nil {
				ew.DedupWindow = w.DedupWindow.Duration
			} else {
				ew.DedupWindow = 1 * time.Hour
			}
			if w.Timeout != nil {
				ew.Timeout = w.Timeout.Duration
			} else {
				ew.Timeout = 30 * time.Second
			}
			out = append(out, ew)
		}
		eff.NotifyWebhooks = out
	}
	return eff
}

// ValidationError describes a single invalid field. It is wrapped in MultiError
// when more than one field is bad. Both types implement error so callers can use
// errors.As to inspect.
//
// Code + Args are the locale-free structured identifier the cmd layer maps
// to a persona.Key for friend-tone rendering. Reason stays as the English
// fallback so Error() — and any non-localized field — remains usable.
type ValidationError struct {
	Field  string
	Reason string
	Code   string
	Args   []any
}

// Error implements error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("config: %s: %s", e.Field, e.Reason)
}

// MultiError aggregates multiple ValidationError so the user sees every problem
// at once instead of fix-and-retry-and-fix-and-retry.
type MultiError struct{ Errors []*ValidationError }

// Error implements error.
func (m *MultiError) Error() string {
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "config: %d invalid fields:\n", len(m.Errors))
	for _, e := range m.Errors {
		sb.WriteString("  - ")
		sb.WriteString(e.Field)
		sb.WriteString(": ")
		sb.WriteString(e.Reason)
		sb.WriteString("\n")
	}
	return sb.String()
}

// Stable Code values for ValidationError. The cmd layer maps these to a
// persona.Key for friend-tone rendering; tests should reference these
// constants rather than the literal strings.
const (
	ReasonHookTimeoutOutOfRange  = "hook_timeout_out_of_range"
	ReasonHookSlowOutOfRange     = "hook_slow_out_of_range"
	ReasonFailRateOutOfRange     = "fail_rate_out_of_range"
	ReasonOutboxBacklogTooSmall  = "outbox_backlog_too_small"
	ReasonNotifyChannelInvalid   = "notify_channel_invalid"
	ReasonPollIntervalOutOfRange = "poll_interval_out_of_range"
	ReasonBatchSizeOutOfRange    = "batch_size_out_of_range"
	ReasonPersonaLocaleInvalid   = "persona_locale_invalid"
)

// Validate checks the EFFECTIVE values against per-field rules.
// Returns nil, *ValidationError (single failure), or *MultiError (multiple).
func (c Config) Validate() error {
	eff := c.Effective()
	var errs []*ValidationError
	add := func(field, reason string) {
		errs = append(errs, &ValidationError{Field: field, Reason: reason})
	}
	addCoded := func(field, code, reason string, args ...any) {
		errs = append(errs, &ValidationError{Field: field, Reason: reason, Code: code, Args: args})
	}

	// hookTimeoutMs: 100ms (paranoid floor) .. 10min (paranoid ceiling).
	// 30s is the spec default; 10min is "if you're hitting this, you have a
	// bigger problem than buddy".
	if eff.HookTimeoutMs < 100 || eff.HookTimeoutMs > 10*60*1000 {
		addCoded("hookTimeoutMs", ReasonHookTimeoutOutOfRange,
			fmt.Sprintf("must be 100..600000 (got %d)", eff.HookTimeoutMs),
			eff.HookTimeoutMs)
	}
	// hookSlowMs must be smaller than hookTimeoutMs — slow comes before timeout.
	if eff.HookSlowMs < 1 || eff.HookSlowMs > eff.HookTimeoutMs {
		addCoded("hookSlowMs", ReasonHookSlowOutOfRange,
			fmt.Sprintf("must be 1..hookTimeoutMs (got %d, timeout %d)", eff.HookSlowMs, eff.HookTimeoutMs),
			eff.HookSlowMs, eff.HookTimeoutMs)
	}
	if eff.HookFailRatePct < 1 || eff.HookFailRatePct > 100 {
		addCoded("hookFailRatePct", ReasonFailRateOutOfRange,
			fmt.Sprintf("must be 1..100 (got %d)", eff.HookFailRatePct),
			eff.HookFailRatePct)
	}
	if eff.OutboxBacklog < 1 {
		addCoded("outboxBacklog", ReasonOutboxBacklogTooSmall,
			fmt.Sprintf("must be >= 1 (got %d)", eff.OutboxBacklog),
			eff.OutboxBacklog)
	}
	if eff.NotifyChannel != "stderr" {
		// "desktop" lands in v0.2; stderr is the only valid value in v0.1.
		addCoded("notifyChannel", ReasonNotifyChannelInvalid,
			fmt.Sprintf("must be \"stderr\" (got %q)", eff.NotifyChannel),
			eff.NotifyChannel)
	}
	if eff.PollInterval < 100*time.Millisecond || eff.PollInterval > 60*time.Second {
		addCoded("pollInterval", ReasonPollIntervalOutOfRange,
			fmt.Sprintf("must be 100ms..60s (got %s)", eff.PollInterval),
			eff.PollInterval)
	}
	if eff.BatchSize < 1 || eff.BatchSize > 100_000 {
		addCoded("batchSize", ReasonBatchSizeOutOfRange,
			fmt.Sprintf("must be 1..100000 (got %d)", eff.BatchSize),
			eff.BatchSize)
	}
	if eff.PersonaLocale != "ko" && eff.PersonaLocale != "en" {
		addCoded("personaLocale", ReasonPersonaLocaleInvalid,
			fmt.Sprintf("must be \"ko\" or \"en\" (got %q)", eff.PersonaLocale),
			eff.PersonaLocale)
	}

	// Advisor thresholds — permissive bounds because these are
	// user-tunable taste knobs, not safety floors.
	if eff.AdvisorTokenSpikeRatio < 1.0 {
		add("advisorTokenSpikeRatio", fmt.Sprintf("must be >= 1.0 (got %g)", eff.AdvisorTokenSpikeRatio))
	}
	if eff.AdvisorLongSessionHours < 1 || eff.AdvisorLongSessionHours > 168 {
		add("advisorLongSessionHours", fmt.Sprintf("must be 1..168 (got %d)", eff.AdvisorLongSessionHours))
	}
	if eff.AdvisorLowCachePct < 0 || eff.AdvisorLowCachePct > 100 {
		add("advisorLowCachePct", fmt.Sprintf("must be 0..100 (got %d)", eff.AdvisorLowCachePct))
	}
	if eff.AdvisorSessionVolumePerDay < 1 {
		add("advisorSessionVolumePerDay", fmt.Sprintf("must be >= 1 (got %d)", eff.AdvisorSessionVolumePerDay))
	}
	if eff.AdvisorTokenDailyThreshold < 1 {
		add("advisorTokenDailyThreshold", fmt.Sprintf("must be >= 1 (got %d)", eff.AdvisorTokenDailyThreshold))
	}
	if eff.AdvisorDedupWindow < time.Minute || eff.AdvisorDedupWindow > 30*24*time.Hour {
		add("advisorDedupWindow", fmt.Sprintf("must be 1m..30d (got %s)", eff.AdvisorDedupWindow))
	}
	if eff.AdvisorPollInterval < time.Minute || eff.AdvisorPollInterval > 24*time.Hour {
		add("advisorPollInterval", fmt.Sprintf("must be 1m..24h (got %s)", eff.AdvisorPollInterval))
	}
	if eff.AdvisorGoalDriftThreshold < 0 || eff.AdvisorGoalDriftThreshold > 1 {
		add("advisorGoalDriftThreshold", fmt.Sprintf("must be 0..1 (got %g)", eff.AdvisorGoalDriftThreshold))
	}
	if eff.AdvisorGoalDriftSampleChunks < 1 {
		add("advisorGoalDriftSampleChunks", fmt.Sprintf("must be >= 1 (got %d)", eff.AdvisorGoalDriftSampleChunks))
	}

	// Notify thresholds — severity strings constrained to the
	// advisor-side enum; webhook entries validated individually.
	validSev := map[string]bool{"info": true, "warn": true, "high": true}
	if !validSev[eff.NotifyDesktopSeverityMin] {
		add("notifyDesktopSeverityMin", fmt.Sprintf("must be info|warn|high (got %q)", eff.NotifyDesktopSeverityMin))
	}
	if !validSev[eff.NotifyTUIBannerSeverityMin] {
		add("notifyTuiBannerSeverityMin", fmt.Sprintf("must be info|warn|high (got %q)", eff.NotifyTUIBannerSeverityMin))
	}
	if !validSev[eff.NotifyShellPromptSeverityMin] {
		add("notifyShellPromptSeverityMin", fmt.Sprintf("must be info|warn|high (got %q)", eff.NotifyShellPromptSeverityMin))
	}
	if eff.NotifyDesktopDedup < 0 {
		add("notifyDesktopDedup", fmt.Sprintf("must be >= 0 (got %s)", eff.NotifyDesktopDedup))
	}
	for i, w := range eff.NotifyWebhooks {
		if w.URL == "" {
			add(fmt.Sprintf("notifyWebhooks[%d].url", i), "URL is required")
		}
		if w.Method != "POST" && w.Method != "PUT" && w.Method != "PATCH" {
			add(fmt.Sprintf("notifyWebhooks[%d].method", i), fmt.Sprintf("must be POST|PUT|PATCH (got %q)", w.Method))
		}
		if !validSev[w.SeverityMin] {
			add(fmt.Sprintf("notifyWebhooks[%d].severityMin", i), fmt.Sprintf("must be info|warn|high (got %q)", w.SeverityMin))
		}
		if w.DedupWindow < 0 {
			add(fmt.Sprintf("notifyWebhooks[%d].dedupWindow", i), fmt.Sprintf("must be >= 0 (got %s)", w.DedupWindow))
		}
		if w.Timeout <= 0 || w.Timeout > 5*time.Minute {
			add(fmt.Sprintf("notifyWebhooks[%d].timeout", i), fmt.Sprintf("must be 0..5m (got %s)", w.Timeout))
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return &MultiError{Errors: errs}
	}
}

// Duration is a time.Duration with JSON support: it marshals as a Go duration
// string (e.g. "1s", "500ms") and unmarshals from either a string or an
// integer (interpreted as milliseconds, for hand-edited config files that
// would otherwise need to remember the duration syntax).
type Duration struct{ time.Duration }

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// UnmarshalJSON implements json.Unmarshaler. Accepts a string ("1s") or an
// integer (milliseconds).
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		parsed, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("parse duration %q: %w", s, err)
		}
		d.Duration = parsed
		return nil
	}
	var ms int64
	if err := json.Unmarshal(b, &ms); err != nil {
		return fmt.Errorf("duration must be string or int ms: %w", err)
	}
	d.Duration = time.Duration(ms) * time.Millisecond
	return nil
}

// DefaultPath returns the standard config location: ~/.buddy/config.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".buddy", "config.json"), nil
}

// Load reads the config from path. Returns:
//
//   - parsed Config, nil error: file exists, parses, and validates
//   - zero Config, nil error: file does not exist (caller falls back to
//     Defaults via Effective())
//   - zero Config, non-nil error: parse failure or validation failure
//
// Pass an empty path to use DefaultPath().
func Load(path string) (Config, error) {
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return Config{}, err
		}
		path = p
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var c Config
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields() // catches typos like "hookSlow" instead of "hookSlowMs"
	if err := dec.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err := c.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return c, nil
}

// Save writes c to path atomically (temp file + rename). Pretty-printed JSON
// for hand-edit friendliness. Caller is responsible for validation before
// calling Save — Save will not refuse an invalid config, since T2's CLI may
// want to write a sentinel-bad file for tests. Pass an empty path to use
// DefaultPath().
func Save(path string, c Config) error {
	if path == "" {
		p, err := DefaultPath()
		if err != nil {
			return err
		}
		path = p
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
