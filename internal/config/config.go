// Package config owns the buddy user-level configuration: ~/.buddy/config.json.
//
// Design:
//
//   - Config holds POINTER fields so JSON unmarshalling can distinguish "absent"
//     (nil) from "explicitly set, even to a zero value" (non-nil). Hand-edited
//     config.json files only override the fields they mention.
//
//   - Effective is the resolved view with non-pointer fields. The rest of the
//     codebase (M5 T3: doctor / aggregator / daemon) consumes Effective.
//
//   - Defaults are spec-locked (v0.1-spec §6.2 + §6.3). Changing them requires a
//     spec update — they are not user-facing.
//
//   - Validate runs on the EFFECTIVE values, so a zero Config (all defaults)
//     always passes. Invalid configs only come from explicit user overrides.
//
// T1 scope is the schema, Defaults, Effective, Validate, Load, Save. The
// `buddy config get/set/unset/show` CLI is T2; doctor/aggregator/daemon
// integration is T3.
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
	NotifyChannel   *string   `json:"notifyChannel,omitempty"` // v0.1: only "stderr"
	PollInterval    *Duration `json:"pollInterval,omitempty"`
	BatchSize       *int      `json:"batchSize,omitempty"`
	PersonaLocale   *string   `json:"personaLocale,omitempty"` // "ko" | "en"
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
}

// Defaults returns the spec-locked defaults from v0.1-spec §6.2 + §6.3 plus
// the daemon's own defaults (PollInterval, BatchSize) which were never
// user-tunable until M5.
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
	return eff
}

// ValidationError describes a single invalid field. It is wrapped in MultiError
// when more than one field is bad. Both types implement error so callers can use
// errors.As to inspect.
type ValidationError struct {
	Field  string
	Reason string
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

// Validate checks the EFFECTIVE values against per-field rules.
// Returns nil, *ValidationError (single failure), or *MultiError (multiple).
func (c Config) Validate() error {
	eff := c.Effective()
	var errs []*ValidationError
	add := func(field, reason string) {
		errs = append(errs, &ValidationError{Field: field, Reason: reason})
	}

	// hookTimeoutMs: 100ms (paranoid floor) .. 10min (paranoid ceiling).
	// 30s is the spec default; 10min is "if you're hitting this, you have a
	// bigger problem than buddy".
	if eff.HookTimeoutMs < 100 || eff.HookTimeoutMs > 10*60*1000 {
		add("hookTimeoutMs", fmt.Sprintf("must be 100..600000 (got %d)", eff.HookTimeoutMs))
	}
	// hookSlowMs must be smaller than hookTimeoutMs — slow comes before timeout.
	if eff.HookSlowMs < 1 || eff.HookSlowMs > eff.HookTimeoutMs {
		add("hookSlowMs", fmt.Sprintf("must be 1..hookTimeoutMs (got %d, timeout %d)", eff.HookSlowMs, eff.HookTimeoutMs))
	}
	if eff.HookFailRatePct < 1 || eff.HookFailRatePct > 100 {
		add("hookFailRatePct", fmt.Sprintf("must be 1..100 (got %d)", eff.HookFailRatePct))
	}
	if eff.OutboxBacklog < 1 {
		add("outboxBacklog", fmt.Sprintf("must be >= 1 (got %d)", eff.OutboxBacklog))
	}
	if eff.NotifyChannel != "stderr" {
		// "desktop" lands in v0.2; stderr is the only valid value in v0.1.
		add("notifyChannel", fmt.Sprintf("must be \"stderr\" (got %q)", eff.NotifyChannel))
	}
	if eff.PollInterval < 100*time.Millisecond || eff.PollInterval > 60*time.Second {
		add("pollInterval", fmt.Sprintf("must be 100ms..60s (got %s)", eff.PollInterval))
	}
	if eff.BatchSize < 1 || eff.BatchSize > 100_000 {
		add("batchSize", fmt.Sprintf("must be 1..100000 (got %d)", eff.BatchSize))
	}
	if eff.PersonaLocale != "ko" && eff.PersonaLocale != "en" {
		add("personaLocale", fmt.Sprintf("must be \"ko\" or \"en\" (got %q)", eff.PersonaLocale))
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
