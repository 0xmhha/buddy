// Package install wraps and unwraps Claude Code's settings.json hook entries
// with `buddy hook-wrap`, and optionally writes a cliwrap.yaml supervising the
// buddy daemon.
//
// All operations are pure (no os.Exit, no cobra). The CLI layer in
// cmd/buddy/main.go owns user-facing printing and exit codes.
package install

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/wm-it-22-00661/buddy/internal/cliwrapcfg"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

// SettingsFileName is the Claude Code settings filename inside ClaudeDir.
const SettingsFileName = "settings.json"

// BackupSuffix is appended to the settings file path for first-install backup.
const BackupSuffix = ".buddy.bak"

// CliwrapFileName is the cliwrap.yaml written under BuddyDir on --with-cliwrap.
const CliwrapFileName = "cliwrap.yaml"

// hookEvents is the canonical set of Claude Code hook events buddy wraps.
// Derived from schema.KnownEvents so the two lists can never drift; we still
// iterate this fixed list (rather than scanning unknown keys) so non-hook
// settings such as `permissions` or `env` stay untouched.
var hookEvents = func() []string {
	out := make([]string, len(schema.KnownEvents))
	for i, e := range schema.KnownEvents {
		out[i] = string(e)
	}
	return out
}()

// Options configure Install / Uninstall. Paths are absolute.
type Options struct {
	// ClaudeDir is the directory containing settings.json. Defaults to ~/.claude.
	ClaudeDir string
	// BuddyDir is where cliwrap.yaml lands. Defaults to ~/.buddy.
	BuddyDir string
	// BuddyBinary is the absolute path to the buddy binary, used as the wrap prefix.
	// Defaults to os.Executable().
	BuddyBinary string
	// WithCliwrap, if true on Install, also writes cliwrap.yaml.
	WithCliwrap bool
	// DBPath, if set, is forwarded to cliwrapcfg.Render for the daemon's --db flag.
	DBPath string
}

// Result describes what an Install/Uninstall did. Used by the CLI layer
// to choose the right friend-tone message.
type Result struct {
	SettingsPath string
	BackupPath   string
	CliwrapPath  string

	// Wrapped is the count of hook commands transformed in this run.
	Wrapped int
	// Unwrapped is the count of hook commands restored to their original form.
	Unwrapped int
	// AlreadyWrapped is the count of hook commands skipped (already buddy-wrapped on Install).
	AlreadyWrapped int

	// BackupCreated is true only when this run wrote a brand-new backup file.
	BackupCreated bool
	// RestoredFromBackup is true when Uninstall restored from a .buddy.bak.
	RestoredFromBackup bool
	// CliwrapWritten is true when --with-cliwrap produced a file.
	CliwrapWritten bool
	// NoOp is true when nothing changed (idempotent re-install or empty hooks).
	NoOp bool
}

// ErrSettingsMissing is returned when ~/.claude/settings.json does not exist.
// CLI layer translates this to a friend-tone message.
var ErrSettingsMissing = errors.New("settings.json not found")

// ErrBinaryPathHasSpaces is returned (wrapped in BinaryPathSpaceError) when
// the resolved buddy binary path contains a space character. settings.json
// wrap commands are shell-tokenized by Claude Code, so a space in the binary
// path produces malformed invocations. CLI layer translates this to a
// friend-tone message.
var ErrBinaryPathHasSpaces = errors.New("buddy binary path contains spaces")

// BinaryPathSpaceError carries the offending path so the CLI layer can render
// a friend-tone message naming exactly which path is wrong.
type BinaryPathSpaceError struct{ Path string }

func (e *BinaryPathSpaceError) Error() string {
	return fmt.Sprintf("%s: %q", ErrBinaryPathHasSpaces.Error(), e.Path)
}

func (e *BinaryPathSpaceError) Unwrap() error { return ErrBinaryPathHasSpaces }

// Install wraps every command in settings.json hooks with `buddy hook-wrap`.
// Idempotent: already-wrapped commands are detected and skipped.
// On first run, writes a one-time backup at <settings>.buddy.bak.
func Install(opts Options) (*Result, error) {
	resolved, err := resolve(opts)
	if err != nil {
		return nil, err
	}
	res := &Result{
		SettingsPath: resolved.settingsPath,
		BackupPath:   resolved.backupPath,
	}

	raw, err := os.ReadFile(resolved.settingsPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return res, ErrSettingsMissing
		}
		return res, fmt.Errorf("read settings: %w", err)
	}

	doc, err := parseJSON(raw)
	if err != nil {
		return res, fmt.Errorf("parse settings.json: %w", err)
	}

	wrapped, already := transformHooks(doc, resolved.binary, wrapCommand)
	res.Wrapped = wrapped
	res.AlreadyWrapped = already

	if wrapped == 0 {
		// No-op for idempotency: also skip backup/write so the file stays
		// byte-equal to its current state.
		res.NoOp = true
	} else {
		// Write backup before mutating, but only if it doesn't exist yet.
		created, err := writeBackupOnce(resolved.settingsPath, resolved.backupPath, raw)
		if err != nil {
			return res, err
		}
		res.BackupCreated = created
		if err := writeJSONAtomic(resolved.settingsPath, doc); err != nil {
			return res, err
		}
	}

	if opts.WithCliwrap {
		path, err := writeCliwrap(resolved.buddyDir, resolved.binary, opts.DBPath)
		if err != nil {
			return res, err
		}
		res.CliwrapPath = path
		res.CliwrapWritten = true
	}

	return res, nil
}

// Uninstall reverses Install. If a .buddy.bak backup exists, restores from it
// (preserving original byte-for-byte). Otherwise walks the JSON and unwraps
// any buddy-wrapped command back to its original form.
func Uninstall(opts Options) (*Result, error) {
	resolved, err := resolve(opts)
	if err != nil {
		return nil, err
	}
	res := &Result{
		SettingsPath: resolved.settingsPath,
		BackupPath:   resolved.backupPath,
	}

	if _, err := os.Stat(resolved.settingsPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return res, ErrSettingsMissing
		}
		return res, fmt.Errorf("stat settings: %w", err)
	}

	if backup, err := os.ReadFile(resolved.backupPath); err == nil {
		if err := writeBytesAtomic(resolved.settingsPath, backup); err != nil {
			return res, err
		}
		// The backup did its job; remove it so a future install starts fresh.
		if err := os.Remove(resolved.backupPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return res, fmt.Errorf("remove backup: %w", err)
		}
		res.RestoredFromBackup = true
		return res, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return res, fmt.Errorf("read backup: %w", err)
	}

	// No backup — fall back to JSON walk + unwrap.
	raw, err := os.ReadFile(resolved.settingsPath)
	if err != nil {
		return res, fmt.Errorf("read settings: %w", err)
	}
	doc, err := parseJSON(raw)
	if err != nil {
		return res, fmt.Errorf("parse settings.json: %w", err)
	}
	unwrapped, _ := transformHooks(doc, resolved.binary, unwrapCommand)
	res.Unwrapped = unwrapped
	if unwrapped == 0 {
		res.NoOp = true
		return res, nil
	}
	if err := writeJSONAtomic(resolved.settingsPath, doc); err != nil {
		return res, err
	}
	return res, nil
}

// --- internals ---

type resolvedPaths struct {
	settingsPath string
	backupPath   string
	buddyDir     string
	binary       string
}

func resolve(opts Options) (resolvedPaths, error) {
	out := resolvedPaths{}

	claudeDir := opts.ClaudeDir
	if claudeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return out, fmt.Errorf("user home dir: %w", err)
		}
		claudeDir = filepath.Join(home, ".claude")
	}
	out.settingsPath = filepath.Join(claudeDir, SettingsFileName)
	out.backupPath = out.settingsPath + BackupSuffix

	buddyDir := opts.BuddyDir
	if buddyDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return out, fmt.Errorf("user home dir: %w", err)
		}
		buddyDir = filepath.Join(home, ".buddy")
	}
	out.buddyDir = buddyDir

	binary := opts.BuddyBinary
	if binary == "" {
		exe, err := os.Executable()
		if err != nil {
			return out, fmt.Errorf("locate self: %w", err)
		}
		binary = exe
	}
	if strings.ContainsRune(binary, ' ') {
		// settings.json hook commands are shell-tokenized; a path with spaces
		// would split into multiple argv slots and break invocation.
		return out, &BinaryPathSpaceError{Path: binary}
	}
	out.binary = binary
	return out, nil
}

// parseJSON decodes the settings file into a generic map. We round-trip via
// json.RawMessage at the top level only when needed for ordering — for v0.1,
// stable enough output ordering comes from json.MarshalIndent's keys-sorted
// behavior on map[string]any (alphabetical), which is acceptable since we
// always rewrite on install.
func parseJSON(raw []byte) (map[string]any, error) {
	var doc map[string]any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// transformHooks walks doc.hooks.<EventName>[].hooks[].command and applies fn.
// Returns (changed, alreadyMatched) where:
//   - changed is the count of commands fn actually rewrote
//   - alreadyMatched is the count of commands fn returned unchanged because
//     they were already in the target state (e.g. wrap of an already-wrapped cmd).
func transformHooks(doc map[string]any, binary string, fn commandFn) (changed, already int) {
	rawHooks, ok := doc["hooks"]
	if !ok {
		return 0, 0
	}
	hooks, ok := rawHooks.(map[string]any)
	if !ok {
		return 0, 0
	}
	for _, event := range hookEvents {
		entries, ok := hooks[event].([]any)
		if !ok {
			continue
		}
		for _, entry := range entries {
			entryMap, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			matcher, _ := entryMap["matcher"].(string)
			inner, ok := entryMap["hooks"].([]any)
			if !ok {
				continue
			}
			for _, h := range inner {
				hMap, ok := h.(map[string]any)
				if !ok {
					continue
				}
				cmdRaw, ok := hMap["command"].(string)
				if !ok {
					continue
				}
				newCmd, didChange := fn(event, matcher, cmdRaw, binary)
				if didChange {
					hMap["command"] = newCmd
					changed++
				} else {
					already++
				}
			}
		}
	}
	return changed, already
}

// commandFn returns (newCommand, changed). changed=false means the input was
// already in the desired form (or didn't apply) and was left untouched.
type commandFn func(event, matcher, command, binary string) (string, bool)

// wrapCommand transforms `<cmd>` into:
//
//	<binary> hook-wrap <hookName> --event <Event> -- <cmd>
//
// where `<hookName>` is `"<event>:<matcher>"` when a matcher is present (e.g.
// `"PreToolUse:Bash"`) and `<event>` alone otherwise. The `--event` flag (not
// the positional hook-name) decides the schema event when stdin lacks
// `hook_event_name`, so each wrapped slot reports the event it actually lives
// in instead of silently falling back to PreToolUse.
func wrapCommand(event, matcher, command, binary string) (string, bool) {
	prefix := binary + " hook-wrap "
	if strings.HasPrefix(command, prefix) {
		return command, false
	}
	hookName := event
	if matcher != "" {
		hookName = event + ":" + matcher
	}
	return fmt.Sprintf("%s hook-wrap %s --event %s -- %s",
		binary, hookName, event, command), true
}

// unwrapCommand reverses wrapCommand by stripping everything up to and
// including the ` -- ` separator. We accept two prefix shapes so that
// uninstall still matches even when the buddy binary has moved between
// install and uninstall:
//  1. exact `<binary> hook-wrap ` (current binary path)
//  2. any `…/buddy hook-wrap ` substring (handles relocations). The leading
//     slash on the suffix avoids matching a bare `buddy` in cwd, which would
//     be ambiguous.
//
// In both cases we require a literal ` -- ` separator after the prefix; if
// it's missing we return the input untouched rather than guess and silently
// corrupt user data.
func unwrapCommand(_, _, command, binary string) (string, bool) {
	exactPrefix := binary + " hook-wrap "
	if strings.HasPrefix(command, exactPrefix) {
		_, after, ok := strings.Cut(command[len(exactPrefix):], " -- ")
		if !ok {
			return command, false
		}
		return after, true
	}
	const marker = "/buddy hook-wrap "
	idx := strings.Index(command, marker)
	if idx < 0 {
		return command, false
	}
	_, after, ok := strings.Cut(command[idx+len(marker):], " -- ")
	if !ok {
		return command, false
	}
	return after, true
}

func writeBackupOnce(settingsPath, backupPath string, original []byte) (bool, error) {
	if _, err := os.Stat(backupPath); err == nil {
		// Backup already exists — preserve user state. write-once invariant.
		return false, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return false, fmt.Errorf("stat backup: %w", err)
	}
	if err := writeBytesAtomic(backupPath, original); err != nil {
		return false, fmt.Errorf("write backup: %w", err)
	}
	_ = settingsPath // reserved for future cross-checks
	return true, nil
}

func writeJSONAtomic(path string, doc map[string]any) error {
	enc, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	// Preserve trailing newline that humans expect.
	enc = append(enc, '\n')
	return writeBytesAtomic(path, enc)
}

func writeBytesAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open tmp: %w", err)
	}
	if _, err := io.Copy(f, strings.NewReader(string(data))); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename tmp: %w", err)
	}
	return nil
}

func writeCliwrap(buddyDir, binary, dbPath string) (string, error) {
	out, err := cliwrapcfg.Render(cliwrapcfg.Spec{
		BuddyBinary: binary,
		DBPath:      dbPath,
	})
	if err != nil {
		return "", fmt.Errorf("render cliwrap: %w", err)
	}
	dest := filepath.Join(buddyDir, CliwrapFileName)
	if err := writeBytesAtomic(dest, []byte(out)); err != nil {
		return "", err
	}
	return dest, nil
}
