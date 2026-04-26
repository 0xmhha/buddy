package main

// config_cmd.go owns the `buddy config {show, get, set, unset}` subtree.
// Lives next to main.go (rather than inside it) so main.go doesn't keep
// growing — see HANDOFF.md §11 watch list.
//
// Design contract:
//   - SUCCESS is silent for set/unset (spec §6.3 "침묵 default" — no
//     "saved!" hype). show and get write to stdout; everything else stays
//     quiet on success.
//   - FAILURE flows through friendError (already defined in main.go) so the
//     user gets the friend-tone message verbatim and the process exits 1.
//   - The actual parsing/formatting per field is delegated to the Field
//     registry in internal/config. This file is pure CLI plumbing.

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wm-it-22-00661/buddy/internal/config"
	"github.com/wm-it-22-00661/buddy/internal/persona"
)

// newConfigCmd returns the `buddy config` parent command. The four subcommands
// share a `--config <path>` flag (default ~/.buddy/config.json) so tests can
// point them at t.TempDir() without setting HOME.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "buddy 설정 보기/수정 (~/.buddy/config.json)",
	}
	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigUnsetCmd(),
	)
	return cmd
}

// configPath returns either the explicit --config flag value, or the spec'd
// default path (~/.buddy/config.json). Centralised so all four subcommands
// resolve the path identically.
func configPath(cmd *cobra.Command) (string, error) {
	flag, _ := cmd.Flags().GetString("config")
	if flag != "" {
		return flag, nil
	}
	p, err := config.DefaultPath()
	if err != nil {
		return "", newFriendError(persona.M(persona.KeyConfigPathUnknown, err))
	}
	return p, nil
}

// loadForCLI loads the config file with friend-tone error mapping. A missing
// file is NOT an error — it just means "no overrides yet" and we return a
// zero Config.
func loadForCLI(path string) (config.Config, error) {
	c, err := config.Load(path)
	if err != nil {
		return config.Config{}, translateConfigError(err)
	}
	return c, nil
}

// translateConfigError wraps a config-package error in the friend-tone shell
// the user expects. Validation errors get a per-field bullet list; everything
// else gets a generic "설정 못 읽었어" wrapper.
//
// TODO(M5/v0.2): the bullet's Reason is still the English string emitted by
// internal/config (e.g. `must be 1..100 (got 200)`). The persona catalog
// already declares Korean replacements (KeyConfigReasonHookTimeoutOutOfRange
// & friends); v0.2's i18n sweep will route ValidationError → persona Key by
// adding a Code/Args field on ValidationError. Out of scope for T5 — see
// docs/roadmap.md M5 T2 deferred Important #2.
func translateConfigError(err error) error {
	var ve *config.ValidationError
	var multi *config.MultiError
	switch {
	case errors.As(err, &multi):
		var sb strings.Builder
		sb.WriteString(persona.M(persona.KeyConfigInvalid))
		sb.WriteString("\n")
		for _, e := range multi.Errors {
			sb.WriteString(persona.M(persona.KeyConfigInvalidField, e.Field, e.Reason))
			sb.WriteString("\n")
		}
		return newFriendError(strings.TrimRight(sb.String(), "\n"))
	case errors.As(err, &ve):
		return newFriendError(
			persona.M(persona.KeyConfigInvalid) + "\n" +
				persona.M(persona.KeyConfigInvalidField, ve.Field, ve.Reason))
	}
	return newFriendError(persona.M(persona.KeyConfigReadFailed, err))
}

// unknownFieldError is the friend-tone message for "you typed a name that
// isn't a config knob". Used by get/set/unset.
func unknownFieldError(name string) error {
	return newFriendError(persona.M(persona.KeyConfigUnknownField, name))
}

// friendParseError translates the internal/config parser error into the
// friend-tone Korean wording the user sees on `buddy config set` failures.
// The cmd layer owns the locale; internal/config stays string-locale-free.
func friendParseError(name, raw string, kind config.FieldKind, err error) string {
	switch kind {
	case config.KindInt, config.KindInt64:
		return persona.M(persona.KeyConfigSetExpectInt, name, raw)
	case config.KindDuration:
		return persona.M(persona.KeyConfigSetExpectDuration, name, raw)
	default:
		// KindString never errors in the current registry, but stay
		// defensive in case a future kind needs a generic fallback.
		return persona.M(persona.KeyConfigSetParseFailed, name, err)
	}
}

// --- show -------------------------------------------------------------------

func newConfigShowCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "현재 effective 설정 출력 (defaults + overrides)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			c, err := loadForCLI(path)
			if err != nil {
				return err
			}
			if jsonOut {
				return showJSON(cmd, c)
			}
			return showText(cmd, c)
		},
	}
	cmd.Flags().String("config", "", "config 파일 경로 (기본: ~/.buddy/config.json)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON 형식으로 출력 (tooling 용)")
	return cmd
}

// showText prints "<key>=<value> (default|override)" lines, one per field,
// alphabetically sorted. Designed for human reading + grep/awk; not friend
// tone (it's a structured surface).
func showText(cmd *cobra.Command, c config.Config) error {
	eff := c.Effective()
	fields := config.Fields()
	// Fields() already returns alphabetical order, but sort defensively in
	// case a future change reorders the registry.
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })

	out := cmd.OutOrStdout()
	for _, f := range fields {
		_, hasOverride := f.GetOverride(c)
		marker := "default"
		if hasOverride {
			marker = "override"
		}
		fmt.Fprintf(out, "%s=%s (%s)\n", f.Name, f.Get(eff), marker)
	}
	return nil
}

// showJSON marshals the Effective view as a pretty-printed JSON object with
// the same camelCase keys the on-disk schema uses. Per-field JSON shaping
// lives on Field.JSONValue so adding a new Config knob can't accidentally
// emit `null` here — the registry is the single source of truth.
func showJSON(cmd *cobra.Command, c config.Config) error {
	eff := c.Effective()
	view := map[string]any{}
	for _, f := range config.Fields() {
		view[f.Name] = f.JSONValue(eff)
	}
	raw, err := json.MarshalIndent(view, "", "  ")
	if err != nil {
		return newFriendError(persona.M(persona.KeyConfigJSONFailed, err))
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(raw))
	return nil
}

// --- get --------------------------------------------------------------------

func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <field>",
		Short: "한 설정값 출력 (shell 치환용; 마커/접두어 없이)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			f, ok := config.FieldByName(name)
			if !ok {
				return unknownFieldError(name)
			}
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			c, err := loadForCLI(path)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), f.Get(c.Effective()))
			return nil
		},
	}
	cmd.Flags().String("config", "", "config 파일 경로 (기본: ~/.buddy/config.json)")
	return cmd
}

// --- set --------------------------------------------------------------------

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <field> <value>",
		Short: "설정값 갱신 (성공시 침묵)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, raw := args[0], args[1]
			f, ok := config.FieldByName(name)
			if !ok {
				return unknownFieldError(name)
			}
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			c, err := loadForCLI(path)
			if err != nil {
				return err
			}
			if err := f.Set(&c, raw); err != nil {
				// Parser-level error. internal/config returns English /
				// machine-shaped messages; the CLI is the i18n boundary, so
				// translate to friend-tone Korean here per Field.Kind. The
				// underlying err is intentionally swallowed in the user-
				// facing message — its English form would only confuse
				// non-English readers, and the kind already implies the
				// expected shape ("숫자", "duration").
				return newFriendError(friendParseError(name, raw, f.Kind, err))
			}
			if err := c.Validate(); err != nil {
				return translateConfigError(err)
			}
			if err := config.Save(path, c); err != nil {
				return newFriendError(persona.M(persona.KeyConfigSaveFailed, err))
			}
			// Silent on success — see file-level comment.
			return nil
		},
	}
	cmd.Flags().String("config", "", "config 파일 경로 (기본: ~/.buddy/config.json)")
	return cmd
}

// --- unset ------------------------------------------------------------------

func newConfigUnsetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset <field>",
		Short: "override 제거 (기본값으로 되돌림, 성공시 침묵)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			f, ok := config.FieldByName(name)
			if !ok {
				return unknownFieldError(name)
			}
			path, err := configPath(cmd)
			if err != nil {
				return err
			}
			c, err := loadForCLI(path)
			if err != nil {
				return err
			}
			f.Unset(&c)
			// Defaults always validate, but run anyway as a sanity guard
			// against future field additions that change semantics.
			if err := c.Validate(); err != nil {
				return translateConfigError(err)
			}
			if err := config.Save(path, c); err != nil {
				return newFriendError(persona.M(persona.KeyConfigSaveFailed, err))
			}
			return nil
		},
	}
	cmd.Flags().String("config", "", "config 파일 경로 (기본: ~/.buddy/config.json)")
	return cmd
}
