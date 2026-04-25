package install_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/diagnose"
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
          {"type": "command", "command": "/usr/local/bin/buddy hook-wrap PreToolUse:* --event PreToolUse -- /bin/foo --bar"}
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

// I2 regression: uninstall must still unwrap when the buddy binary moved
// between install and uninstall. We simulate that by writing the wrap with
// one binary path and uninstalling with a different one.
func TestUninstall_UnwrapsAcrossBinaryPathChanges(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	wrapped := `{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {"type": "command", "command": "/old/path/to/buddy hook-wrap Stop --event Stop -- cleanup.sh"}
        ]
      }
    ]
  }
}
`
	settingsPath := writeSettings(t, claudeDir, wrapped)

	res, err := install.Uninstall(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: "/new/place/buddy", // different from the wrap prefix
	})
	require.NoError(t, err)
	assert.Equal(t, 1, res.Unwrapped, "suffix-based fallback should detect the wrap")

	cmds := extractCommands(t, settingsPath)
	require.Len(t, cmds, 1)
	assert.Equal(t, "cleanup.sh", cmds[0])
}

func TestInstall_WithCliwrap_WritesValidYAML(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	// T6: install pre-creates the DB at this path, so the path must be
	// writable (was a fixed /var/lib/... before T6).
	customDB := filepath.Join(t.TempDir(), "buddy.db")

	res, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
		WithCliwrap: true,
		DBPath:      customDB,
	})
	require.NoError(t, err)
	require.True(t, res.CliwrapWritten)

	yaml, err := os.ReadFile(filepath.Join(buddyDir, install.CliwrapFileName))
	require.NoError(t, err)
	body := string(yaml)

	assert.Contains(t, body, "buddy-daemon")
	assert.Contains(t, body, `"daemon", "run"`)
	assert.Contains(t, body, fakeBuddy)
	assert.Contains(t, body, customDB)
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
// Asserts: (a) --event flag is the source of truth for the schema event,
// (b) hookName is "<event>[:<matcher>]", (c) -- separator is preserved,
// (d) original command text follows verbatim after " -- ".
func TestInstall_WrapTransformShape(t *testing.T) {
	cases := []struct {
		name    string
		event   string
		matcher string
		input   string
		want    string
	}{
		{"PreToolUse with matcher", "PreToolUse", "Bash", "/bin/foo --bar",
			fakeBuddy + " hook-wrap PreToolUse:Bash --event PreToolUse -- /bin/foo --bar"},
		{"Stop no matcher with spaces", "Stop", "", "echo hello world",
			fakeBuddy + " hook-wrap Stop --event Stop -- echo hello world"},
		{"PostToolUse with quotes", "PostToolUse", "*", `sh -c "echo $X"`,
			fakeBuddy + ` hook-wrap PostToolUse:* --event PostToolUse -- sh -c "echo $X"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claudeDir, buddyDir := newTempDirs(t)
			entry := map[string]any{
				"hooks": []any{
					map[string]any{"type": "command", "command": tc.input},
				},
			}
			if tc.matcher != "" {
				entry["matcher"] = tc.matcher
			}
			doc := map[string]any{
				"hooks": map[string]any{
					tc.event: []any{entry},
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
			// Spec-level invariants the C1 bug violated: --event present and
			// the original command survives intact after " -- ".
			assert.Contains(t, cmds[0], "--event "+tc.event)
			assert.Contains(t, cmds[0], " -- "+tc.input)
		})
	}
}

// I4 regression: would have caught C1. After installing a real PostToolUse
// hook, the wrapped command must contain `--event PostToolUse` (not the
// silent fallback PreToolUse) and the original command after " -- ".
func TestInstall_PostToolUse_CarriesEventFlag(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, `{
  "hooks": {
    "PostToolUse": [
      {"matcher": "*", "hooks": [{"type": "command", "command": "report.sh"}]}
    ]
  }
}
`)
	_, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)

	cmds := extractCommands(t, filepath.Join(claudeDir, install.SettingsFileName))
	require.Len(t, cmds, 1)
	tokens := strings.Fields(cmds[0])
	assert.Contains(t, tokens, "--event")
	// --event PostToolUse appears as adjacent tokens.
	for i, tok := range tokens {
		if tok == "--event" {
			require.Less(t, i+1, len(tokens), "--event must have a value")
			assert.Equal(t, "PostToolUse", tokens[i+1],
				"event flag must reflect the actual hook slot, not fall back")
		}
	}
	assert.Contains(t, tokens, "--", "wrap must preserve -- separator")
	assert.True(t, strings.HasSuffix(cmds[0], " -- report.sh"),
		"original command must follow -- verbatim")
}

// I1 regression: install must reject buddy binary paths containing spaces,
// since wrap output is shell-tokenized.
func TestInstall_RejectsBinaryPathWithSpaces(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	_, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: "/Users/me/My Buddy/buddy",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, install.ErrBinaryPathHasSpaces)

	var spaceErr *install.BinaryPathSpaceError
	require.ErrorAs(t, err, &spaceErr)
	assert.Equal(t, "/Users/me/My Buddy/buddy", spaceErr.Path)
}

// assertMigratedDB opens dbPath read-only and verifies that schema_version
// has at least one applied migration AND that hook_outbox is queryable
// without raising a "no such table" error. This is the headline T6 invariant:
// after install, doctor's read-only open + outbox query must succeed.
func assertMigratedDB(t *testing.T, dbPath string) {
	t.Helper()
	conn, err := db.Open(db.Options{Path: dbPath, ReadOnly: true})
	require.NoError(t, err, "read-only open of pre-created DB must succeed")
	t.Cleanup(func() { _ = conn.Close() })

	var version int
	require.NoError(t,
		conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version),
		"schema_version must be queryable after install")
	assert.GreaterOrEqual(t, version, 1, "at least one migration must be applied")

	var count int
	require.NoError(t,
		conn.QueryRow("SELECT COUNT(*) FROM hook_outbox").Scan(&count),
		"hook_outbox must exist (no 'no such table' SQL error)")
	assert.Equal(t, 0, count, "fresh DB has no outbox rows")
}

// T6: install must pre-create buddy-dir and migrate the default DB so that
// `buddy doctor` (run before `buddy daemon start`) does NOT surface a raw
// `no such table: hook_outbox` SQL error. Default path = <BuddyDir>/buddy.db.
func TestInstall_PreCreatesDB_DefaultPath(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	_, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)

	dbPath := filepath.Join(buddyDir, "buddy.db")
	_, statErr := os.Stat(dbPath)
	require.NoError(t, statErr, "default DB must exist after install")
	assertMigratedDB(t, dbPath)
}

// T6: when `--db /custom/path.db` is supplied, install pre-creates exactly
// that path (not the default buddy-dir/buddy.db), so `--db` is honored
// end-to-end across install → doctor.
func TestInstall_PreCreatesDB_ExplicitPath(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	customDir := filepath.Join(t.TempDir(), "custom-state")
	customDB := filepath.Join(customDir, "elsewhere.db")

	_, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
		DBPath:      customDB,
	})
	require.NoError(t, err)

	_, statErr := os.Stat(customDB)
	require.NoError(t, statErr, "explicit --db path must exist after install")
	assertMigratedDB(t, customDB)

	// Default path under buddy-dir must NOT have been created when --db is set.
	_, defaultStatErr := os.Stat(filepath.Join(buddyDir, "buddy.db"))
	assert.True(t, os.IsNotExist(defaultStatErr),
		"explicit --db must not also create the default DB")
}

// T6 acceptance: this is the headline. After install (no daemon started),
// running diagnose.Check must NOT surface a KindDBOpen issue. A KindDaemon
// issue is expected because we did not start the daemon.
func TestInstall_DoctorCleanAfterInstall(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	_, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)

	dbPath := filepath.Join(buddyDir, "buddy.db")
	pidPath := filepath.Join(buddyDir, "no-such.pid") // daemon definitely not running

	rep, err := diagnose.Check(diagnose.Options{
		DBPath:  dbPath,
		PIDFile: pidPath,
	})
	require.NoError(t, err)

	for _, iss := range rep.Issues {
		assert.NotEqual(t, diagnose.KindDBOpen, iss.Kind,
			"install must leave DB readable; got DB-open issue: %s", iss.Message)
		assert.NotContains(t, iss.Message, "no such table",
			"raw SQL 'no such table' must not surface to users")
	}

	// Sanity: doctor still flags the daemon, so the test isn't just passing
	// because Check silently failed. Daemon was never started.
	var sawDaemon bool
	for _, iss := range rep.Issues {
		if iss.Kind == diagnose.KindDaemon {
			sawDaemon = true
			break
		}
	}
	assert.True(t, sawDaemon,
		"daemon-down issue is expected since we never started the daemon")
}

// T6 invariant: pre-create must run BEFORE the settings.json read, so that
// even when settings.json is missing (and Install therefore returns
// ErrSettingsMissing), a subsequent `buddy doctor` still sees a migrated DB.
// This locks in the order so a future refactor can't move initBuddyState
// below the settings read and silently break the contract.
func TestInstall_MissingSettingsJSON_StillPreCreatesDB(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	// deliberately no settings.json in claudeDir

	_, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
	})
	require.ErrorIs(t, err, install.ErrSettingsMissing)

	// Even though install errored at settings, doctor must see a migrated DB.
	assertMigratedDB(t, filepath.Join(buddyDir, "buddy.db"))
}

// T6: installing twice must be idempotent — db.Open's migration loop is
// already idempotent, but verify install() doesn't double-error on a
// pre-existing DB.
func TestInstall_PreCreatesDB_IdempotentOnReinstall(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	_, err := install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err)

	_, err = install.Install(install.Options{
		ClaudeDir: claudeDir, BuddyDir: buddyDir, BuddyBinary: fakeBuddy,
	})
	require.NoError(t, err, "second install must succeed on already-migrated DB")

	assertMigratedDB(t, filepath.Join(buddyDir, "buddy.db"))
}

// T6: when --with-cliwrap and an explicit --db are both set, the rendered
// cliwrap.yaml must reference the same explicit DB path that install
// pre-created. Prevents drift between install pre-create and cliwrap config.
func TestInstall_WithCliwrap_UsesResolvedDBPath(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	customDB := filepath.Join(t.TempDir(), "custom-state", "buddy.db")

	res, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
		WithCliwrap: true,
		DBPath:      customDB,
	})
	require.NoError(t, err)
	require.True(t, res.CliwrapWritten)

	yaml, err := os.ReadFile(filepath.Join(buddyDir, install.CliwrapFileName))
	require.NoError(t, err)
	assert.Contains(t, string(yaml), customDB,
		"cliwrap.yaml must use the same DB path install pre-created")
	assertMigratedDB(t, customDB)
}

// T6: when --with-cliwrap is set without an explicit --db, the rendered
// cliwrap.yaml must reference the default <buddy-dir>/buddy.db that install
// just created — not an empty path. Same default flows from install through
// cliwrap so the daemon spawned by cliwrap reads the same DB doctor reads.
func TestInstall_WithCliwrap_DefaultDBPath(t *testing.T) {
	claudeDir, buddyDir := newTempDirs(t)
	writeSettings(t, claudeDir, settingsWithMixedHooks())

	res, err := install.Install(install.Options{
		ClaudeDir:   claudeDir,
		BuddyDir:    buddyDir,
		BuddyBinary: fakeBuddy,
		WithCliwrap: true,
	})
	require.NoError(t, err)
	require.True(t, res.CliwrapWritten)

	defaultDB := filepath.Join(buddyDir, "buddy.db")
	yaml, err := os.ReadFile(filepath.Join(buddyDir, install.CliwrapFileName))
	require.NoError(t, err)
	assert.Contains(t, string(yaml), defaultDB,
		"cliwrap.yaml must point at the default DB install just created")
	assertMigratedDB(t, defaultDB)
}
