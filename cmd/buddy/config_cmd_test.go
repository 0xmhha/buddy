package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/config"
)

// runConfig is the test harness: build a fresh config root, point it at a
// throwaway --config path, capture stdout/stderr. We deliberately skip
// shelling out to ./bin/buddy — Cobra's SetArgs gives us the same surface
// without slowing the suite down.
func runConfig(t *testing.T, cfgPath string, args ...string) (string, string, error) {
	t.Helper()
	cmd := newConfigCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	full := append([]string{"--config", cfgPath}, args...)
	cmd.SetArgs(full)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

// --- show -------------------------------------------------------------------

func TestConfigShow_PrintsAllFieldsWithMarkers(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// No file → all defaults.
	stdout, _, err := runConfig(t, cfgPath, "show")
	require.NoError(t, err)

	for _, f := range config.Fields() {
		assert.Contains(t, stdout, f.Name+"=")
	}
	// Every line should be marked (default) since there are no overrides.
	for _, line := range strings.Split(strings.TrimRight(stdout, "\n"), "\n") {
		assert.Contains(t, line, "(default)", "line %q lacks (default) marker", line)
	}

	// Now write an override and re-run.
	c := config.Config{}
	require.NoError(t, setFieldForTest(&c, "hookSlowMs", "3000"))
	require.NoError(t, config.Save(cfgPath, c))

	stdout2, _, err := runConfig(t, cfgPath, "show")
	require.NoError(t, err)
	for _, line := range strings.Split(strings.TrimRight(stdout2, "\n"), "\n") {
		if strings.HasPrefix(line, "hookSlowMs=") {
			assert.Contains(t, line, "(override)")
			assert.Contains(t, line, "3000")
		} else {
			assert.Contains(t, line, "(default)")
		}
	}
}

func TestConfigShow_LinesAreSortedAlphabetically(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	stdout, _, err := runConfig(t, cfgPath, "show")
	require.NoError(t, err)

	got := []string{}
	for _, line := range strings.Split(strings.TrimRight(stdout, "\n"), "\n") {
		eq := strings.Index(line, "=")
		require.Greater(t, eq, 0, "malformed line %q", line)
		got = append(got, line[:eq])
	}
	want := []string{}
	for _, f := range config.Fields() {
		want = append(want, f.Name)
	}
	assert.Equal(t, want, got)
}

func TestConfigShow_JSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	stdout, _, err := runConfig(t, cfgPath, "show", "--json")
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	for _, f := range config.Fields() {
		_, ok := out[f.Name]
		assert.True(t, ok, "JSON output missing key %s", f.Name)
	}

	// Lock value types so a future field accidentally returning the wrong
	// shape (e.g. nil, or a struct) gets caught here instead of breaking
	// downstream JSON consumers. JSON numbers unmarshal to float64; strings
	// (including duration's canonical form) stay strings.
	assert.IsType(t, float64(0), out["hookTimeoutMs"], "hookTimeoutMs must be a JSON number")
	assert.IsType(t, float64(0), out["batchSize"], "batchSize must be a JSON number")
	assert.IsType(t, "", out["pollInterval"], "pollInterval must be a JSON string (duration form)")
	assert.IsType(t, "", out["personaLocale"], "personaLocale must be a JSON string")
	assert.IsType(t, "", out["notifyChannel"], "notifyChannel must be a JSON string")
}

// TestConfigShow_MultiError_FormatsAsBullets — when a hand-edited config has
// multiple invalid fields, `buddy config show` must surface ALL of them in
// one friend-tone bullet list (driven by translateConfigError's *MultiError
// branch). Locks the multi-error CLI surface so it can't silently regress to
// "first error only".
func TestConfigShow_MultiError_FormatsAsBullets(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// Hand-write a config with two invalid fields so Validate produces a
	// *MultiError. hookFailRatePct=200 violates the 1..100 range; the
	// personaLocale "fr" violates the ko|en rule.
	raw := []byte(`{"hookFailRatePct": 200, "personaLocale": "fr"}`)
	require.NoError(t, os.WriteFile(cfgPath, raw, 0o644))

	_, _, err := runConfig(t, cfgPath, "show")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe), "want friendError, got %T: %v", err, err)
	// Friend-tone header from translateConfigError.
	assert.Contains(t, fe.msg, "buddy: 설정이 잘못됐어:")
	// Both bad fields show up — the bullet shape ("  - <field>:") plus the
	// field name itself locks the rendering.
	assert.Contains(t, fe.msg, "  - hookFailRatePct:")
	assert.Contains(t, fe.msg, "  - personaLocale:")
}

// --- get --------------------------------------------------------------------

func TestConfigGet_ReturnsEffectiveValue(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// Default first.
	stdout, _, err := runConfig(t, cfgPath, "get", "hookSlowMs")
	require.NoError(t, err)
	assert.Equal(t, "5000\n", stdout)

	// Override and re-fetch.
	c := config.Config{}
	require.NoError(t, setFieldForTest(&c, "hookSlowMs", "3000"))
	require.NoError(t, config.Save(cfgPath, c))

	stdout, _, err = runConfig(t, cfgPath, "get", "hookSlowMs")
	require.NoError(t, err)
	assert.Equal(t, "3000\n", stdout)
}

func TestConfigGet_UnknownField_FriendlyError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "get", "doesnotexist")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe), "want friendError, got %T: %v", err, err)
	assert.Contains(t, fe.msg, "doesnotexist")
	assert.Contains(t, fe.msg, "buddy config show")
}

// --- set --------------------------------------------------------------------

func TestConfigSet_PersistsAndValidates(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	stdout, stderr, err := runConfig(t, cfgPath, "set", "hookSlowMs", "3000")
	require.NoError(t, err)
	// Silent on success.
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)

	// File must exist with the override.
	raw, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"hookSlowMs": 3000`)

	// `get` returns the new value.
	stdout, _, err = runConfig(t, cfgPath, "get", "hookSlowMs")
	require.NoError(t, err)
	assert.Equal(t, "3000\n", stdout)
}

func TestConfigSet_InvalidValue_RejectsAndKeepsFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "set", "hookFailRatePct", "200")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe))
	assert.Contains(t, fe.msg, "hookFailRatePct")

	// File should not have been created.
	_, statErr := os.Stat(cfgPath)
	assert.True(t, os.IsNotExist(statErr), "config file should not exist after rejected set")
}

func TestConfigSet_InvalidValue_PreservesPriorContent(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// Seed with a valid override.
	_, _, err := runConfig(t, cfgPath, "set", "hookSlowMs", "3000")
	require.NoError(t, err)

	before, err := os.ReadFile(cfgPath)
	require.NoError(t, err)

	// Invalid set must leave the file untouched.
	_, _, err = runConfig(t, cfgPath, "set", "hookFailRatePct", "200")
	require.Error(t, err)

	after, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after))
}

func TestConfigSet_DurationField(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "set", "pollInterval", "500ms")
	require.NoError(t, err)

	stdout, _, err := runConfig(t, cfgPath, "get", "pollInterval")
	require.NoError(t, err)
	assert.Equal(t, "500ms\n", stdout)
}

func TestConfigSet_StringField(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "set", "personaLocale", "en")
	require.NoError(t, err)

	stdout, _, err := runConfig(t, cfgPath, "get", "personaLocale")
	require.NoError(t, err)
	assert.Equal(t, "en\n", stdout)
}

func TestConfigSet_UnknownField_FriendlyError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "set", "doesnotexist", "42")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe))
	assert.Contains(t, fe.msg, "doesnotexist")
}

func TestConfigSet_BadDurationParse_FriendlyError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "set", "pollInterval", "not-a-duration")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe))
	assert.Contains(t, fe.msg, "pollInterval")
	// User-visible surface stays Korean — internal/config returns English,
	// the cmd layer translates. Lock the friend-tone wording so the
	// translation can't silently regress to leaking the English message.
	assert.Contains(t, fe.msg, "duration 형식이어야 해")
	assert.Contains(t, fe.msg, "not-a-duration")
}

// TestConfigSet_BadIntParse_FriendlyError mirrors the duration-parse test for
// the int kind: locks that the cmd-layer wrapping speaks Korean for `Atoi`
// failures too, not just durations.
func TestConfigSet_BadIntParse_FriendlyError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "set", "hookSlowMs", "not-a-number")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe))
	assert.Contains(t, fe.msg, "hookSlowMs")
	assert.Contains(t, fe.msg, "숫자여야 해")
	assert.Contains(t, fe.msg, "not-a-number")
}

// --- unset ------------------------------------------------------------------

func TestConfigUnset_RemovesOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "set", "hookSlowMs", "3000")
	require.NoError(t, err)

	stdout, stderr, err := runConfig(t, cfgPath, "unset", "hookSlowMs")
	require.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)

	stdout, _, err = runConfig(t, cfgPath, "get", "hookSlowMs")
	require.NoError(t, err)
	// Default is 5000.
	assert.Equal(t, "5000\n", stdout)
}

func TestConfigUnset_UnknownField_FriendlyError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, err := runConfig(t, cfgPath, "unset", "doesnotexist")
	require.Error(t, err)

	var fe *friendError
	require.True(t, errors.As(err, &fe))
	assert.Contains(t, fe.msg, "doesnotexist")
}

// --- helpers ----------------------------------------------------------------

func setFieldForTest(c *config.Config, name, raw string) error {
	f, ok := config.FieldByName(name)
	if !ok {
		return errors.New("unknown field: " + name)
	}
	return f.Set(c, raw)
}
