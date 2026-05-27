package permissions_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/permissions"
)

func TestInject_CreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.local.json")

	t.Setenv("HOME", dir)

	err := permissions.Inject([]string{"Edit(*.md)", "Edit(*.go)"})
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	perms := settings["permissions"].(map[string]any)
	allow := perms["allow"].([]any)
	assert.Contains(t, allow, "Edit(*.go)")
	assert.Contains(t, allow, "Edit(*.md)")
}

func TestInject_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	existing := map[string]any{
		"permissions": map[string]any{
			"allow": []any{"Bash(git:*)", "Write(*.md)"},
		},
		"someOtherKey": "preserved",
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), data, 0o644))

	t.Setenv("HOME", dir)

	err := permissions.Inject([]string{"Edit(*.md)"})
	require.NoError(t, err)

	readBack, err := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(readBack, &settings))

	assert.Equal(t, "preserved", settings["someOtherKey"])

	perms := settings["permissions"].(map[string]any)
	allow := perms["allow"].([]any)
	assert.Contains(t, allow, "Bash(git:*)")
	assert.Contains(t, allow, "Write(*.md)")
	assert.Contains(t, allow, "Edit(*.md)")
}

func TestInject_Idempotent(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	existing := map[string]any{
		"permissions": map[string]any{
			"allow": []any{"Edit(*.md)", "Edit(*.go)"},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), data, 0o644))

	t.Setenv("HOME", dir)

	err := permissions.Inject([]string{"Edit(*.md)", "Edit(*.go)"})
	require.NoError(t, err)

	readBack, err := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(readBack, &settings))

	perms := settings["permissions"].(map[string]any)
	allow := perms["allow"].([]any)

	count := 0
	for _, v := range allow {
		if v == "Edit(*.md)" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Edit(*.md) should appear exactly once")
}

func TestCheck_DetectsMissing(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	existing := map[string]any{
		"permissions": map[string]any{
			"allow": []any{"Edit(*.md)"},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	require.NoError(t, os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), data, 0o644))

	t.Setenv("HOME", dir)

	result, err := permissions.Check()
	require.NoError(t, err)

	assert.Contains(t, result.Present, "Edit(*.md)")
	assert.Contains(t, result.Missing, "Edit(*.go)")
	assert.False(t, result.Healthy())
}

func TestCheck_NoFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	result, err := permissions.Check()
	require.NoError(t, err)

	assert.Equal(t, len(permissions.RequiredPermissions), len(result.Missing))
	assert.False(t, result.Healthy())
}
