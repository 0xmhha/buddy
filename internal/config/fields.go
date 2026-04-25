package config

// fields.go is the registry that powers `buddy config get/set/unset`. It maps
// each Config knob to:
//
//   - its CLI/JSON name (the same string used in config.json keys),
//   - a parser that writes the user-supplied raw string into *Config,
//   - getters that read either the override (Config) or the resolved
//     (Effective) value back out as a display string,
//   - an unsetter that clears the override.
//
// Putting this here (instead of in cmd/buddy/) keeps the CLI thin and lets
// `internal/config` stay the single source of truth for what a config field
// IS — its name, its kind, how to parse it, how to format it. If a future
// task adds a Config field, the CLI picks it up automatically once the new
// entry lands in Fields().

import (
	"fmt"
	"strconv"
	"time"
)

// FieldKind enumerates the parse types supported by the config CLI.
type FieldKind int

const (
	// KindInt64 — fields stored as *int64 in Config (e.g. hookTimeoutMs).
	KindInt64 FieldKind = iota
	// KindInt — fields stored as *int in Config (e.g. batchSize).
	KindInt
	// KindString — fields stored as *string in Config (e.g. notifyChannel).
	KindString
	// KindDuration — fields stored as *Duration in Config (e.g. pollInterval).
	KindDuration
)

// Field describes one tunable Config knob. It is the unit the CLI iterates over.
type Field struct {
	// Name is the JSON / CLI key (e.g. "hookSlowMs"). Lower-camel to match
	// the JSON tag on Config.
	Name string

	// Kind is the parse type. Useful for help text / future tab-completion;
	// the actual parsing is encapsulated in Set.
	Kind FieldKind

	// Get returns the field's effective value formatted for `config get`
	// output. It always succeeds (every Effective value has a string form).
	Get func(eff Effective) string

	// GetOverride returns the field's override value formatted for display
	// when it is non-nil on Config, plus a boolean flag. When the field is
	// nil (= not overridden), it returns ("", false).
	GetOverride func(c Config) (string, bool)

	// Set parses raw and writes it as an override on c. Returns an error
	// describing the parse failure (e.g. "expected integer for batchSize")
	// when raw is unparseable. Set does NOT range-validate — that is
	// Config.Validate's job, run by the CLI after Set so the user sees the
	// canonical validation message rather than a parser-specific one.
	//
	// Error messages are intentionally English/machine-shaped here. The CLI
	// layer (cmd/buddy/config_cmd.go) wraps them with friend-tone Korean
	// before showing the user. Keeping internal/config string-locale-free
	// means a future i18n switch lives in one place (cmd/buddy/) instead of
	// being scattered through the registry.
	Set func(c *Config, raw string) error

	// Unset clears the override (sets the pointer back to nil on c).
	Unset func(c *Config)

	// JSONValue returns the field's effective value as a Go value suitable
	// for direct JSON marshalling: int / int64 / string for primitives, and
	// the duration's canonical string form (e.g. "1s") for Duration fields.
	// Each builder fills this in at registration time, so adding a new field
	// can't accidentally produce a `null` in `buddy config show --json` —
	// forward-compat is enforced structurally, not by a separate switch.
	JSONValue func(eff Effective) any
}

// Fields returns every registered field, sorted alphabetically by Name. The
// slice is fresh each call (cheap; eight entries) so callers can't accidentally
// mutate the registry.
func Fields() []Field {
	// Order is the alphabetical key order the CLI's `show` output relies on.
	// Adding a new field: keep alphabetical, update FieldByName implicitly
	// (it scans this slice), and ensure fields_test.go TestFields_AllFields*
	// covers it.
	return []Field{
		intField("batchSize",
			func(c *Config) **int { return &c.BatchSize },
			func(eff Effective) int { return eff.BatchSize },
		),
		intField("hookFailRatePct",
			func(c *Config) **int { return &c.HookFailRatePct },
			func(eff Effective) int { return eff.HookFailRatePct },
		),
		int64Field("hookSlowMs",
			func(c *Config) **int64 { return &c.HookSlowMs },
			func(eff Effective) int64 { return eff.HookSlowMs },
		),
		int64Field("hookTimeoutMs",
			func(c *Config) **int64 { return &c.HookTimeoutMs },
			func(eff Effective) int64 { return eff.HookTimeoutMs },
		),
		stringField("notifyChannel",
			func(c *Config) **string { return &c.NotifyChannel },
			func(eff Effective) string { return eff.NotifyChannel },
		),
		intField("outboxBacklog",
			func(c *Config) **int { return &c.OutboxBacklog },
			func(eff Effective) int { return eff.OutboxBacklog },
		),
		stringField("personaLocale",
			func(c *Config) **string { return &c.PersonaLocale },
			func(eff Effective) string { return eff.PersonaLocale },
		),
		durationField("pollInterval",
			func(c *Config) **Duration { return &c.PollInterval },
			func(eff Effective) time.Duration { return eff.PollInterval },
		),
	}
}

// FieldByName looks up the registered field with the given name. Returns the
// zero Field and false if no such field exists. Callers typically wrap the
// false case in a friend-tone error before bubbling it up to the user.
func FieldByName(name string) (Field, bool) {
	for _, f := range Fields() {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// --- private builders --------------------------------------------------------
//
// Each builder closes over a "ptr" accessor (returning **T so we can both read
// and assign the underlying *T pointer field on Config) and an "eff" accessor
// (the Effective view). This keeps the per-field code generic without
// resorting to reflection.

func intField(
	name string,
	ptr func(*Config) **int,
	eff func(Effective) int,
) Field {
	return Field{
		Name: name,
		Kind: KindInt,
		Get: func(e Effective) string {
			return strconv.Itoa(eff(e))
		},
		GetOverride: func(c Config) (string, bool) {
			p := *ptr(&c)
			if p == nil {
				return "", false
			}
			return strconv.Itoa(*p), true
		},
		Set: func(c *Config, raw string) error {
			v, err := strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("%s: expected integer, got %q", name, raw)
			}
			*ptr(c) = &v
			return nil
		},
		Unset: func(c *Config) {
			*ptr(c) = nil
		},
		JSONValue: func(e Effective) any {
			return eff(e)
		},
	}
}

func int64Field(
	name string,
	ptr func(*Config) **int64,
	eff func(Effective) int64,
) Field {
	return Field{
		Name: name,
		Kind: KindInt64,
		Get: func(e Effective) string {
			return strconv.FormatInt(eff(e), 10)
		},
		GetOverride: func(c Config) (string, bool) {
			p := *ptr(&c)
			if p == nil {
				return "", false
			}
			return strconv.FormatInt(*p, 10), true
		},
		Set: func(c *Config, raw string) error {
			v, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("%s: expected integer, got %q", name, raw)
			}
			*ptr(c) = &v
			return nil
		},
		Unset: func(c *Config) {
			*ptr(c) = nil
		},
		JSONValue: func(e Effective) any {
			return eff(e)
		},
	}
}

func stringField(
	name string,
	ptr func(*Config) **string,
	eff func(Effective) string,
) Field {
	return Field{
		Name: name,
		Kind: KindString,
		Get: func(e Effective) string {
			return eff(e)
		},
		GetOverride: func(c Config) (string, bool) {
			p := *ptr(&c)
			if p == nil {
				return "", false
			}
			return *p, true
		},
		Set: func(c *Config, raw string) error {
			v := raw
			*ptr(c) = &v
			return nil
		},
		Unset: func(c *Config) {
			*ptr(c) = nil
		},
		JSONValue: func(e Effective) any {
			return eff(e)
		},
	}
}

func durationField(
	name string,
	ptr func(*Config) **Duration,
	eff func(Effective) time.Duration,
) Field {
	return Field{
		Name: name,
		Kind: KindDuration,
		Get: func(e Effective) string {
			return eff(e).String()
		},
		GetOverride: func(c Config) (string, bool) {
			p := *ptr(&c)
			if p == nil {
				return "", false
			}
			return p.Duration.String(), true
		},
		Set: func(c *Config, raw string) error {
			d, err := time.ParseDuration(raw)
			if err != nil {
				return fmt.Errorf("%s: expected duration like 1s or 500ms, got %q", name, raw)
			}
			v := Duration{Duration: d}
			*ptr(c) = &v
			return nil
		},
		Unset: func(c *Config) {
			*ptr(c) = nil
		},
		// Duration emits its canonical string ("1s", "500ms") rather than a
		// nanosecond integer, matching Duration.MarshalJSON. Keeping the
		// `show --json` view byte-equivalent to what `Save` would write.
		JSONValue: func(e Effective) any {
			return eff(e).String()
		},
	}
}
