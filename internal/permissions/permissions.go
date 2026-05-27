// Package permissions manages Claude Code settings.local.json permissions
// for subagent Edit/Write tool access.
//
// Claude Code subagents run in background and cannot prompt for interactive
// permission approval. Tools must be pre-allowed in settings.local.json
// for subagents to use them. This package detects missing permissions and
// injects them with user confirmation.
package permissions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RequiredPermissions lists the tool patterns that subagents need.
// Subagents run in background and cannot prompt for interactive approval,
// so these must be pre-allowed in settings.local.json.
var RequiredPermissions = []string{
	// Edit — file modification (subagent의 가장 빈번한 거부 원인)
	"Edit(*.md)",
	"Edit(*.go)",
	"Edit(*.json)",
	"Edit(*.yaml)",
	"Edit(*.yml)",
	"Edit(*.ts)",
	"Edit(*.tsx)",
	"Edit(*.js)",
	"Edit(*.jsx)",
	"Edit(*.py)",
	"Edit(*.sh)",
	"Edit(*.sql)",
	"Edit(*.html)",
	"Edit(*.css)",

	// Bash — 코드 구현/빌드/테스트에 필요한 명령어
	"Bash(go:*)",
	"Bash(python3:*)",
	"Bash(make:*)",
	"Bash(grep:*)",
	"Bash(sed:*)",
	"Bash(awk:*)",
	"Bash(sort:*)",
	"Bash(tr:*)",
	"Bash(cut:*)",
	"Bash(uniq:*)",
	"Bash(diff:*)",
	"Bash(touch:*)",
	"Bash(which:*)",
	"Bash(gofmt:*)",
	"Bash(goimports:*)",
}

// settingsFile is the structure of ~/.claude/settings.local.json.
// We only parse the fields we care about to avoid dropping unknown keys.
type settingsFile map[string]any

// CheckResult reports which required permissions are present vs missing.
type CheckResult struct {
	Present []string
	Missing []string
	Path    string
}

// Healthy returns true if no required permissions are missing.
func (r CheckResult) Healthy() bool { return len(r.Missing) == 0 }

// SettingsPath returns the path to ~/.claude/settings.local.json.
func SettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".claude", "settings.local.json"), nil
}

// Check reads the settings file and reports which required permissions
// are present and which are missing.
func Check() (CheckResult, error) {
	path, err := SettingsPath()
	if err != nil {
		return CheckResult{}, err
	}

	existing, err := readAllowList(path)
	if err != nil {
		return CheckResult{}, err
	}

	result := CheckResult{Path: path}
	set := make(map[string]bool, len(existing))
	for _, p := range existing {
		set[p] = true
	}

	for _, req := range RequiredPermissions {
		if set[req] {
			result.Present = append(result.Present, req)
		} else {
			result.Missing = append(result.Missing, req)
		}
	}
	return result, nil
}

// Inject adds missing permissions to the settings file. It reads the
// existing file (or creates a new one), merges the missing permissions
// into permissions.allow, and writes back. Existing entries are preserved.
func Inject(missing []string) error {
	path, err := SettingsPath()
	if err != nil {
		return err
	}

	settings, err := readSettings(path)
	if err != nil {
		return err
	}

	perms := getOrCreatePermissions(settings)
	allowList := getStringSlice(perms, "allow")

	set := make(map[string]bool, len(allowList))
	for _, p := range allowList {
		set[p] = true
	}
	for _, m := range missing {
		if !set[m] {
			allowList = append(allowList, m)
		}
	}

	sort.Strings(allowList)
	perms["allow"] = allowList
	settings["permissions"] = perms

	return writeSettings(path, settings)
}

func readAllowList(path string) ([]string, error) {
	settings, err := readSettings(path)
	if err != nil {
		return nil, err
	}
	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		return nil, nil
	}
	return getStringSlice(perms, "allow"), nil
}

func readSettings(path string) (settingsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return settingsFile{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var settings settingsFile
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return settings, nil
}

func writeSettings(path string, settings settingsFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}

func getOrCreatePermissions(settings settingsFile) map[string]any {
	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		perms = map[string]any{}
	}
	return perms
}

func getStringSlice(m map[string]any, key string) []string {
	raw, ok := m[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// FormatMissing returns a human-readable list of missing permissions,
// grouped by tool type (Edit, Write, etc).
func FormatMissing(missing []string) string {
	if len(missing) == 0 {
		return ""
	}
	var b strings.Builder
	for _, m := range missing {
		b.WriteString("  - ")
		b.WriteString(m)
		b.WriteString("\n")
	}
	return b.String()
}
