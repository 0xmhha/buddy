package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/persona"
)

// TestRoot_PersistentPreRunE_HonorsSubcommandConfigFlag locks the contract
// that a subcommand's --config flag drives the active persona locale. The
// PersistentPreRunE runs on the resolved leaf, so cmd.Flags().Lookup("config")
// reaches a non-persistent flag declared by the subcommand. Without this,
// `buddy <cmd> --config <path>` would render messages in the locale of
// ~/.buddy/config.json instead of the explicit file.
func TestRoot_PersistentPreRunE_HonorsSubcommandConfigFlag(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{"personaLocale":"en"}`), 0o644))

	// Locale is package-global; reset before and after so this test cannot
	// leak state into siblings.
	require.NoError(t, persona.SetLocale(persona.LocaleKO))
	t.Cleanup(func() { _ = persona.SetLocale(persona.LocaleKO) })

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"config", "show", "--config", cfgPath})
	require.NoError(t, root.Execute())

	assert.Equal(t, persona.LocaleEN, persona.ActiveLocale(),
		"--config on the subcommand must drive root's locale resolution")
}

// TestRoot_PersistentPreRunE_FallsBackToDefaultPath guards the legacy
// behavior: when the subcommand's --config flag isn't set, locale
// resolution silently uses config.DefaultPath(). A missing default file is
// fine and leaves locale at its prior value. Sandboxes HOME so a developer
// machine's real ~/.buddy/config.json cannot make this test flaky.
func TestRoot_PersistentPreRunE_FallsBackToDefaultPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	require.NoError(t, persona.SetLocale(persona.LocaleKO))
	t.Cleanup(func() { _ = persona.SetLocale(persona.LocaleKO) })

	root := newRootCmd()
	cfgCmd, _, err := root.Find([]string{"config", "show"})
	require.NoError(t, err)
	// Empty Args means --config was never set; PersistentPreRunE should
	// resolve to DefaultPath() (a non-existent file under the sandboxed
	// HOME) and leave locale untouched at ko.
	require.NoError(t, root.PersistentPreRunE(cfgCmd, nil))

	assert.Equal(t, persona.LocaleKO, persona.ActiveLocale(),
		"no --config means fall back to default path; ko stays ko")
}
