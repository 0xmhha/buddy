package cliwrapcfg_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/cliwrapcfg"
)

func TestRender_RequiresBuddyBinary(t *testing.T) {
	_, err := cliwrapcfg.Render(cliwrapcfg.Spec{})
	require.Error(t, err)
}

func TestRender_AppliesDefaultsAndIncludesEssentials(t *testing.T) {
	out, err := cliwrapcfg.Render(cliwrapcfg.Spec{
		BuddyBinary: "/usr/local/bin/buddy",
	})
	require.NoError(t, err)

	for _, want := range []string{
		`version: "1"`,
		`/usr/local/bin/buddy`,
		`"daemon", "run"`,
		`restart: on_failure`,
		`max_restarts: 5`,
		`restart_backoff: 2s`,
		`stop_timeout: 5s`,
		`${HOME}/.buddy/cliwrap`,
	} {
		assert.True(t, strings.Contains(out, want), "want %q in output\n--\n%s", want, out)
	}
}

func TestRender_IncludesDBPathArgWhenProvided(t *testing.T) {
	out, err := cliwrapcfg.Render(cliwrapcfg.Spec{
		BuddyBinary: "/usr/local/bin/buddy",
		DBPath:      "/var/lib/buddy/buddy.db",
	})
	require.NoError(t, err)
	assert.Contains(t, out, `"--db", "/var/lib/buddy/buddy.db"`)
}

func TestRender_RespectsCustomRestartTuning(t *testing.T) {
	out, err := cliwrapcfg.Render(cliwrapcfg.Spec{
		BuddyBinary:  "/usr/local/bin/buddy",
		MaxRestarts:  10,
		BackoffSecs:  4,
		StopTimeoutS: 30,
	})
	require.NoError(t, err)
	assert.Contains(t, out, `max_restarts: 10`)
	assert.Contains(t, out, `restart_backoff: 4s`)
	assert.Contains(t, out, `stop_timeout: 30s`)
}

func TestRender_RespectsCustomRuntimeDir(t *testing.T) {
	out, err := cliwrapcfg.Render(cliwrapcfg.Spec{
		BuddyBinary: "/usr/local/bin/buddy",
		RuntimeDir:  "/var/run/buddy",
	})
	require.NoError(t, err)
	assert.Contains(t, out, `dir: /var/run/buddy`)
}
