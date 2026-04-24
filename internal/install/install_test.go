package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/install"
)

const fakeBuddy = "/usr/local/bin/buddy"

// settingsWithMixedHooks returns a JSON document containing every shape we
// care about: multiple events, multiple matchers per event, multi-command
// inner arrays, plus an unrelated top-level field that we must preserve.
func settingsWithMixedHooks() string {
	return `{
  "permissions": {
    "allow": ["Read", "Write"]
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "tool == \"Bash\"",
        "hooks": [
          {"type": "command", "command": "/bin/foo --bar"}
        ]
      },
      {
        "matcher": "tool == \"Edit\"",
        "hooks": [
          {"type": "command", "command": "echo edit"},
          {"type": "command", "command": "/usr/bin/lint"}
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {"type": "command", "command": "report.sh"}
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {"type": "command", "command": "cleanup"}
        ]
      }
    ]
  }
}
`
}

func writeSettings(t *testing.T, claudeDir, contents string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))
	path := filepath.Join(claudeDir, install.SettingsFileName)
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	return path
}

func newTempDirs(t *testing.T) (claudeDir, buddyDir string) {
	t.Helper()
	root := t.TempDir()
	claudeDir = filepath.Join(root, ".claude")
	buddyDir = filepath.Join(root, ".buddy")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))
	return
}

// extractCommands walks the doc and returns every hooks[].hooks[].command
// across all events, in deterministic order: event name (alphabetical) then
// entry index then inner index.
func extractCommands(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw, &doc))
	hooks, ok := doc["hooks"].(map[string]any)
	require.True(t, ok, "hooks must be a map")

	events := make([]string, 0, len(hooks))
	for k := range hooks {
		events = append(events, k)
	}
	// sort for determinism
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[j] < events[i] {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	var out []string
	for _, ev := range events {
		entries, ok := hooks[ev].([]any)
		if !ok {
			continue
		}
		for _, entry := range entries {
			em := entry.(map[string]any)
			inner, ok := em["hooks"].([]any)
			if !ok {
				continue
			}
			for _, h := range inner {
				hm := h.(map[string]any)
				if cmd, ok := hm["command"].(string); ok {
					out = append(out, cmd)
				}
			}
		}
	}
	return out
}

func TestInstall_WrapsAllHookCommands_AndCreatesBackup(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	settingsPath := writeSettings(t, claudeDir, settingsWithMixedHooks())
	originalBytes, _ := os.ReadFile(settingsPath)

	res, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, 5, res.Wrapped, "five commands across the fixture")
	assert.Equal(t, 0, res.AlreadyWrapped)
	assert.True(t, res.BackupCreated, "first install must create backup")
	assert.False(t, res.NoOp)

	// Backup byte-equals original.
	backup, err := os.ReadFile(settingsPath + install.BackupSuffix)
	require.NoError(t, err)
	assert.Equal(t, string(originalBytes), string(backup))

	// Every command now starts with the buddy wrap prefix.
	for _, cmd := range extractCommands(t, settingsPath) {
		assert.True(t, strings.HasPrefix(cmd, fakeBuddy+" hook-wrap "),
			"expected wrap prefix, got %q", cmd)
		assert.Contains(t, cmd, " -- ", "wrap must contain -- separator")
	}
}

func TestInstall_Idempotent_NoSecondBackupNoChange(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	_, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)

	settingsPath := filepath.Join(claudeDir, install.SettingsFileName)
	afterFirst, err := os.ReadFile(settingsPath)
	require.NoError(t, err)
	backupAfterFirst, err := os.ReadFile(settingsPath + install.BackupSuffix)
	require.NoError(t, err)

	// Tamper with backup so we can detect any (forbidden) overwrite on the
	// second install: write a sentinel and confirm it survives.
	sentinel := []byte("SENTINEL_BACKUP\n")
	require.NoError(t, os.WriteFile(settingsPath+install.BackupSuffix, sentinel, 0o644))

	res2, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, res2.Wrapped, "no commands should need wrapping")
	assert.Equal(t, 5, res2.AlreadyWrapped)
	assert.False(t, res2.BackupCreated)
	assert.True(t, res2.NoOp)

	afterSecond, err := os.ReadFile(settingsPath)
	require.NoError(t, err)
	assert.Equal(t, string(afterFirst), string(afterSecond), "second install must not modify settings")

	// Backup must NOT be overwritten — sentinel still there.
	currentBackup, err := os.ReadFile(settingsPath + install.BackupSuffix)
	require.NoError(t, err)
	assert.Equal(t, string(sentinel), string(currentBackup), "backup overwrite forbidden")
	_ = backupAfterFirst
}

func TestInstall_BackupAlreadyExists_NotOverwritten(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	settingsPath := writeSettings(t, claudeDir, settingsWithMixedHooks())

	preexisting := []byte(`{"user":"hand-edited backup"}`)
	require.NoError(t, os.WriteFile(settingsPath+install.BackupSuffix, preexisting, 0o644))

	res, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)
	assert.False(t, res.BackupCreated)

	bk, err := os.ReadFile(settingsPath + install.BackupSuffix)
	require.NoError(t, err)
	assert.Equal(t, string(preexisting), string(bk))
}

func TestInstallUninstall_Roundtrip_RestoresOriginalBytes(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	settingsPath := writeSettings(t, claudeDir, settingsWithMixedHooks())
	original, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	_, err = install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)

	res, err := install.Uninstall(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)
	assert.True(t, res.RestoredFromBackup)

	restored, err := os.ReadFile(settingsPath)
	require.NoError(t, err)
	assert.Equal(t, string(original), string(restored), "round-trip must be byte-equal")

	// Backup should be gone after successful restore.
	_, err = os.Stat(settingsPath + install.BackupSuffix)
	assert.True(t, os.IsNotExist(err), "backup should be removed after restore")
}

func TestUninstall_NoBackup_UnwrapsByJSONWalk(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	// Hand-craft a wrapped settings.json with NO backup — simulates the
	// "user moved/deleted the backup" recovery path.
	wrapped := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {"type": "command", "command": "/usr/local/bin/buddy hook-wrap PreToolUse -- /bin/foo --bar"}
        ]
      }
    ]
  }
}
`
	settingsPath := writeSettings(t, claudeDir, wrapped)

	res, err := install.Uninstall(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)
	assert.False(t, res.RestoredFromBackup)
	assert.Equal(t, 1, res.Unwrapped)

	cmds := extractCommands(t, settingsPath)
	require.Len(t, cmds, 1)
	assert.Equal(t, "/bin/foo --bar", cmds[0])
}

func TestInstall_WithCliwrap_WritesValidYAML(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	res, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
		WithCliwrap: true,
		DBPath:      "/var/lib/buddy/buddy.db",
	})
	require.NoError(t, err)
	require.True(t, res.CliwrapWritten)

	yaml, err := os.ReadFile(filepath.Join(buddyDir, install.CliwrapFileName))
	require.NoError(t, err)
	body := string(yaml)

	assert.Contains(t, body, "buddy-daemon")
	assert.Contains(t, body, `"daemon", "run"`)
	assert.Contains(t, body, fakeBuddy)
	assert.Contains(t, body, "/var/lib/buddy/buddy.db")
}

func TestInstall_MissingSettingsJSON_ReturnsSentinelError(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	// Note: we deliberately do NOT write settings.json.

	_, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, install.ErrSettingsMissing)
}

func TestInstall_NoHookFields_NoOps(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	settingsPath := writeSettings(t, claudeDir, `{"permissions":{"allow":["Read"]}}`+"\n")
	original, _ := os.ReadFile(settingsPath)

	res, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)
	assert.True(t, res.NoOp)
	assert.Equal(t, 0, res.Wrapped)
	assert.False(t, res.BackupCreated)

	// File untouched.
	after, _ := os.ReadFile(settingsPath)
	assert.Equal(t, string(original), string(after))
	// And no backup either.
	_, statErr := os.Stat(settingsPath + install.BackupSuffix)
	assert.True(t, os.IsNotExist(statErr))
}

func TestInstall_EmptyHooksObject_NoOps(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, `{"hooks":{}}`+"\n")

	res, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)
	assert.True(t, res.NoOp)
}

func TestUninstall_MissingSettingsJSON_ReturnsSentinelError(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	_, err := install.Uninstall(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, install.ErrSettingsMissing)
}

// Table-driven check on the wrapping rule itself, going through Install.
func TestInstall_WrapTransformShape(t *testing.T) {
	cases := []struct {
		name  string
		event string
		input string
		want  string
	}{
		{"PreToolUse simple", "PreToolUse", "/bin/foo --bar",
			fakeBuddy + " hook-wrap PreToolUse -- /bin/foo --bar"},
		{"Stop with spaces", "Stop", "echo hello world",
			fakeBuddy + " hook-wrap Stop -- echo hello world"},
		{"PostToolUse with quotes", "PostToolUse", `sh -c "echo $X"`,
			fakeBuddy + ` hook-wrap PostToolUse -- sh -c "echo $X"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claudeDir, buddyDir := newTempDirs(t)
			doc := map[string]any{
				"hooks": map[string]any{
					tc.event: []any{
						map[string]any{
							"hooks": []any{
								map[string]any{"type": "command", "command": tc.input},
							},
						},
					},
				},
			}
			raw, err := json.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			writeSettings(t, claudeDir, string(raw)+"\n")

			_, err = install.Install(install.Options{
				ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
			})
			require.NoError(t, err)

			cmds := extractCommands(t, filepath.Join(claudeDir, install.SettingsFileName))
			require.Len(t, cmds, 1)
			assert.Equal(t, tc.want, cmds[0])
		})
	}
}
