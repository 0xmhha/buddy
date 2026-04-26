package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/config"
)

// TestLoadEffectiveConfig_MissingFile_ReturnsDefaults locks the M5 T3
// "first-run UX": when ~/.buddy/config.json doesn't exist, doctor / daemon
// must fall back to the spec defaults silently — not error out.
func TestLoadEffectiveConfig_MissingFile_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	// File deliberately not created.
	cfgPath := filepath.Join(dir, "config.json")

	eff, err := loadEffectiveConfig(cfgPath)
	require.NoError(t, err)

	// Effective view must equal Defaults() exactly. If a field were
	// accidentally zeroed or set to a different default, this assertion
	// catches it.
	assert.Equal(t, config.Defaults(), eff)
}

// TestLoadEffectiveConfig_WithOverride_AppliesOverlay locks the integration
// path: a user override in config.json must surface as Effective.<Field>.
// Two fields chosen — one int (HookSlowMs) and one duration (PollInterval)
// — to exercise both the JSON unmarshal and the Duration custom marshaller.
func TestLoadEffectiveConfig_WithOverride_AppliesOverlay(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	c := config.Config{}
	require.NoError(t, setFieldForTest(&c, "hookSlowMs", "3000"))
	require.NoError(t, setFieldForTest(&c, "pollInterval", "250ms"))
	require.NoError(t, config.Save(cfgPath, c))

	eff, err := loadEffectiveConfig(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, int64(3000), eff.HookSlowMs)
	assert.Equal(t, 250*time.Millisecond, eff.PollInterval)
	// Untouched fields keep their defaults.
	assert.Equal(t, config.Defaults().HookTimeoutMs, eff.HookTimeoutMs)
	assert.Equal(t, config.Defaults().BatchSize, eff.BatchSize)
}

// TestLoadEffectiveConfig_InvalidConfig_ReturnsFriendlyError locks the error
// path: a hand-edited config with bad values must surface as a friendError
// (so main() prints the message verbatim and exits 1) carrying the
// per-field bullet shape.
func TestLoadEffectiveConfig_InvalidConfig_ReturnsFriendlyError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// Two bad fields at once → MultiError, the harder branch.
	raw := []byte(`{"hookFailRatePct": 200, "personaLocale": "fr"}`)
	require.NoError(t, os.WriteFile(cfgPath, raw, 0o644))

	_, err := loadEffectiveConfig(cfgPath)
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe), "want friendError, got %T: %v", err, err)
	assert.Contains(t, fe.msg, "buddy: 설정이 잘못됐어:")
	assert.Contains(t, fe.msg, "  - hookFailRatePct:")
	assert.Contains(t, fe.msg, "  - personaLocale:")
}

// TestBuildDoctorOptions_MapsAllFourThresholds locks the Effective →
// diagnose.Thresholds wiring. If a future Effective field gains doctor-side
// meaning and someone forgets to wire it in buildDoctorOptions, this test
// fails when they bump the assertion (or quietly passes — and the integration
// test below catches the runtime gap).
func TestBuildDoctorOptions_MapsAllFourThresholds(t *testing.T) {
	eff := config.Effective{
		HookTimeoutMs:   12345,
		HookSlowMs:      2000,
		HookFailRatePct: 42,
		OutboxBacklog:   77,
		// Other Effective fields irrelevant for doctor.
	}
	opts := buildDoctorOptions("/tmp/buddy.db", "/tmp/daemon.pid", eff)

	assert.Equal(t, "/tmp/buddy.db", opts.DBPath)
	assert.Equal(t, "/tmp/daemon.pid", opts.PIDFile)
	assert.Equal(t, int64(12345), opts.Thresholds.HookTimeoutMs)
	assert.Equal(t, int64(2000), opts.Thresholds.HookSlowMs)
	assert.Equal(t, 42, opts.Thresholds.HookFailRatePct)
	assert.Equal(t, 77, opts.Thresholds.OutboxBacklog)
}

// TestDoctor_UsesConfiguredHookSlowMs is the user-visible acceptance test
// from M5 T3: after `buddy config set hookSlowMs 2000`, doctor must build
// its diagnose.Options with HookSlowMs=2000. We exercise the precise path
// the cobra RunE takes (loadEffectiveConfig → buildDoctorOptions) without
// reaching for sqlite — the threshold mapping is the contract under test.
func TestDoctor_UsesConfiguredHookSlowMs(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	c := config.Config{}
	require.NoError(t, setFieldForTest(&c, "hookSlowMs", "2000"))
	require.NoError(t, config.Save(cfgPath, c))

	eff, err := loadEffectiveConfig(cfgPath)
	require.NoError(t, err)

	opts := buildDoctorOptions("/tmp/buddy.db", "/tmp/daemon.pid", eff)
	assert.Equal(t, int64(2000), opts.Thresholds.HookSlowMs,
		"acceptance: `buddy config set hookSlowMs 2000` must reach diagnose.Thresholds")
	// Other thresholds keep their defaults — locks that the override is
	// scoped to the field the user set, not a partial-clobber of the rest.
	assert.Equal(t, config.Defaults().HookTimeoutMs, opts.Thresholds.HookTimeoutMs)
	assert.Equal(t, config.Defaults().HookFailRatePct, opts.Thresholds.HookFailRatePct)
	assert.Equal(t, config.Defaults().OutboxBacklog, opts.Thresholds.OutboxBacklog)
}

// TestResolveDaemonRunConfig_PrecedenceMatrix locks the three-tier precedence
// for daemon.Config: explicit flag > config file > spec default. Each row
// covers a distinct case.
func TestResolveDaemonRunConfig_PrecedenceMatrix(t *testing.T) {
	defaults := config.Defaults()

	cases := []struct {
		name       string
		pollFlag   time.Duration
		batchFlag  int
		eff        config.Effective
		wantPoll   time.Duration
		wantBatch  int
		wantDBPath string
	}{
		{
			name:      "all defaults: zero flag + Defaults eff",
			pollFlag:  0,
			batchFlag: 0,
			eff:       defaults,
			wantPoll:  defaults.PollInterval,
			wantBatch: defaults.BatchSize,
		},
		{
			name:      "config wins when flag is zero sentinel",
			pollFlag:  0,
			batchFlag: 0,
			eff: config.Effective{
				PollInterval: 250 * time.Millisecond,
				BatchSize:    100,
			},
			wantPoll:  250 * time.Millisecond,
			wantBatch: 100,
		},
		{
			name:      "flag wins over config",
			pollFlag:  500 * time.Millisecond,
			batchFlag: 42,
			eff: config.Effective{
				PollInterval: 250 * time.Millisecond,
				BatchSize:    100,
			},
			wantPoll:  500 * time.Millisecond,
			wantBatch: 42,
		},
		{
			name:      "mixed: poll from flag, batch from config",
			pollFlag:  500 * time.Millisecond,
			batchFlag: 0,
			eff: config.Effective{
				PollInterval: 250 * time.Millisecond,
				BatchSize:    100,
			},
			wantPoll:  500 * time.Millisecond,
			wantBatch: 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := resolveDaemonRunConfig("/tmp/db", "/tmp/pid", tc.pollFlag, tc.batchFlag, tc.eff)
			assert.Equal(t, tc.wantPoll, cfg.PollInterval)
			assert.Equal(t, tc.wantBatch, cfg.BatchSize)
			assert.Equal(t, "/tmp/db", cfg.DBPath)
			assert.Equal(t, "/tmp/pid", cfg.PIDFile)
		})
	}
}

// TestDaemonRun_UsesConfiguredPollAndBatch exercises the full chain that
// `buddy daemon run` (with no --poll / --batch flags) takes when a config
// file overrides those knobs. Mirrors TestDoctor_UsesConfiguredHookSlowMs.
func TestDaemonRun_UsesConfiguredPollAndBatch(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	c := config.Config{}
	require.NoError(t, setFieldForTest(&c, "pollInterval", "200ms"))
	require.NoError(t, setFieldForTest(&c, "batchSize", "100"))
	require.NoError(t, config.Save(cfgPath, c))

	eff, err := loadEffectiveConfig(cfgPath)
	require.NoError(t, err)

	// Sentinel zero flags → "use config".
	cfg := resolveDaemonRunConfig("/tmp/db", "/tmp/pid", 0, 0, eff)
	assert.Equal(t, 200*time.Millisecond, cfg.PollInterval,
		"acceptance: pollInterval=200ms in config must reach daemon.Config")
	assert.Equal(t, 100, cfg.BatchSize,
		"acceptance: batchSize=100 in config must reach daemon.Config")
}
