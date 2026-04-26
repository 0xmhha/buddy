package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/config"
)

// helpers ---------------------------------------------------------------------

func intPtr(v int) *int       { return &v }
func int64Ptr(v int64) *int64 { return &v }
func strPtr(v string) *string { return &v }
func durPtr(d time.Duration) *config.Duration {
	dur := config.Duration{Duration: d}
	return &dur
}

// Defaults ---------------------------------------------------------------------

func TestDefaults_MatchSpec(t *testing.T) {
	d := config.Defaults()
	// Per docs/v0.1-spec §6.2 + §6.3.
	assert.Equal(t, int64(30_000), d.HookTimeoutMs, "hookTimeoutMs spec §6.2")
	assert.Equal(t, int64(5_000), d.HookSlowMs, "hookSlowMs spec §6.2")
	assert.Equal(t, 20, d.HookFailRatePct, "hookFailRatePct spec §6.2")
	assert.Equal(t, 1_000, d.OutboxBacklog, "outboxBacklog spec §6.2")
	assert.Equal(t, "stderr", d.NotifyChannel, "notifyChannel spec §6.2")
	assert.Equal(t, time.Second, d.PollInterval, "pollInterval matches daemon default")
	assert.Equal(t, 500, d.BatchSize, "batchSize matches daemon default")
	assert.Equal(t, "ko", d.PersonaLocale, "personaLocale spec §6.3")
}

// Effective -------------------------------------------------------------------

func TestEffective_AllNil_ReturnsDefaults(t *testing.T) {
	var c config.Config
	assert.Equal(t, config.Defaults(), c.Effective())
}

func TestEffective_PartialOverride_MergesWithDefaults(t *testing.T) {
	c := config.Config{HookSlowMs: int64Ptr(3_000)}
	eff := c.Effective()

	// Overridden field.
	assert.Equal(t, int64(3_000), eff.HookSlowMs)

	// All other fields should match defaults.
	d := config.Defaults()
	assert.Equal(t, d.HookTimeoutMs, eff.HookTimeoutMs)
	assert.Equal(t, d.HookFailRatePct, eff.HookFailRatePct)
	assert.Equal(t, d.OutboxBacklog, eff.OutboxBacklog)
	assert.Equal(t, d.NotifyChannel, eff.NotifyChannel)
	assert.Equal(t, d.PollInterval, eff.PollInterval)
	assert.Equal(t, d.BatchSize, eff.BatchSize)
	assert.Equal(t, d.PersonaLocale, eff.PersonaLocale)
}

func TestEffective_AllFieldsOverridden(t *testing.T) {
	c := config.Config{
		HookTimeoutMs:   int64Ptr(20_000),
		HookSlowMs:      int64Ptr(2_000),
		HookFailRatePct: intPtr(10),
		OutboxBacklog:   intPtr(500),
		NotifyChannel:   strPtr("stderr"),
		PollInterval:    durPtr(2 * time.Second),
		BatchSize:       intPtr(250),
		PersonaLocale:   strPtr("en"),
	}
	eff := c.Effective()
	assert.Equal(t, int64(20_000), eff.HookTimeoutMs)
	assert.Equal(t, int64(2_000), eff.HookSlowMs)
	assert.Equal(t, 10, eff.HookFailRatePct)
	assert.Equal(t, 500, eff.OutboxBacklog)
	assert.Equal(t, "stderr", eff.NotifyChannel)
	assert.Equal(t, 2*time.Second, eff.PollInterval)
	assert.Equal(t, 250, eff.BatchSize)
	assert.Equal(t, "en", eff.PersonaLocale)
}

// Validate --------------------------------------------------------------------

func TestValidate_DefaultsAreValid(t *testing.T) {
	var c config.Config
	require.NoError(t, c.Validate(), "zero Config (= Defaults) must always validate")
}

func TestValidate_AggregatesErrors(t *testing.T) {
	c := config.Config{
		HookFailRatePct: intPtr(200),
		NotifyChannel:   strPtr("desktop"),
		PersonaLocale:   strPtr("fr"),
	}
	err := c.Validate()
	require.Error(t, err)

	var multi *config.MultiError
	require.ErrorAs(t, err, &multi, "want MultiError when multiple fields invalid")
	assert.Len(t, multi.Errors, 3)

	fields := map[string]bool{}
	for _, e := range multi.Errors {
		fields[e.Field] = true
	}
	assert.True(t, fields["hookFailRatePct"])
	assert.True(t, fields["notifyChannel"])
	assert.True(t, fields["personaLocale"])
}

func TestValidate_RejectsHookSlowGreaterThanTimeout(t *testing.T) {
	c := config.Config{
		HookTimeoutMs: int64Ptr(5_000),
		HookSlowMs:    int64Ptr(10_000),
	}
	err := c.Validate()
	require.Error(t, err)
	var ve *config.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "hookSlowMs", ve.Field)
}

func TestValidate_RejectsUnknownNotifyChannel(t *testing.T) {
	c := config.Config{NotifyChannel: strPtr("desktop")}
	err := c.Validate()
	require.Error(t, err)
	var ve *config.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "notifyChannel", ve.Field)
}

func TestValidate_RejectsBadLocale(t *testing.T) {
	c := config.Config{PersonaLocale: strPtr("fr")}
	err := c.Validate()
	require.Error(t, err)
	var ve *config.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "personaLocale", ve.Field)
}

func TestValidate_RejectsTinyHookTimeout(t *testing.T) {
	c := config.Config{HookTimeoutMs: int64Ptr(50)}
	err := c.Validate()
	require.Error(t, err)
}

func TestValidate_RejectsHugeHookTimeout(t *testing.T) {
	c := config.Config{HookTimeoutMs: int64Ptr(11 * 60 * 1000)}
	err := c.Validate()
	require.Error(t, err)
}

func TestValidate_RejectsBadPollInterval(t *testing.T) {
	c := config.Config{PollInterval: durPtr(50 * time.Millisecond)}
	err := c.Validate()
	require.Error(t, err)
	var ve *config.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "pollInterval", ve.Field)
}

func TestValidate_RejectsBadBatchSize(t *testing.T) {
	c := config.Config{BatchSize: intPtr(0)}
	err := c.Validate()
	require.Error(t, err)
	var ve *config.ValidationError
	require.ErrorAs(t, err, &ve)
	assert.Equal(t, "batchSize", ve.Field)
}

// Load ------------------------------------------------------------------------

func TestLoad_MissingFile_ReturnsZeroConfigNoError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nope.json")

	c, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, config.Config{}, c)
	// Effective() should yield Defaults.
	assert.Equal(t, config.Defaults(), c.Effective())
}

func TestLoad_PartialJSON_DecodesNilForMissingFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"hookSlowMs": 3000}`), 0o644))

	c, err := config.Load(path)
	require.NoError(t, err)
	require.NotNil(t, c.HookSlowMs)
	assert.Equal(t, int64(3_000), *c.HookSlowMs)

	assert.Nil(t, c.HookTimeoutMs)
	assert.Nil(t, c.HookFailRatePct)
	assert.Nil(t, c.OutboxBacklog)
	assert.Nil(t, c.NotifyChannel)
	assert.Nil(t, c.PollInterval)
	assert.Nil(t, c.BatchSize)
	assert.Nil(t, c.PersonaLocale)
}

func TestLoad_RejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"hookSlow": 3000}`), 0o644))

	_, err := config.Load(path)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "unknown field")
}

func TestLoad_RejectsInvalidValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"hookFailRatePct": 200}`), 0o644))

	_, err := config.Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hookFailRatePct")
}

func TestLoad_MalformedJSON_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{not json`), 0o644))

	_, err := config.Load(path)
	require.Error(t, err)
}

// Save ------------------------------------------------------------------------

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	c := config.Config{
		HookSlowMs:    int64Ptr(2_500),
		PollInterval:  durPtr(750 * time.Millisecond),
		PersonaLocale: strPtr("en"),
	}
	require.NoError(t, config.Save(path, c))

	loaded, err := config.Load(path)
	require.NoError(t, err)

	require.NotNil(t, loaded.HookSlowMs)
	assert.Equal(t, int64(2_500), *loaded.HookSlowMs)
	require.NotNil(t, loaded.PollInterval)
	assert.Equal(t, 750*time.Millisecond, loaded.PollInterval.Duration)
	require.NotNil(t, loaded.PersonaLocale)
	assert.Equal(t, "en", *loaded.PersonaLocale)

	// Untouched fields stay nil.
	assert.Nil(t, loaded.HookTimeoutMs)
	assert.Nil(t, loaded.BatchSize)
}

func TestSave_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Pre-populate.
	require.NoError(t, os.WriteFile(path, []byte(`{"hookSlowMs": 9999}`), 0o644))

	c := config.Config{HookSlowMs: int64Ptr(1_234)}
	require.NoError(t, config.Save(path, c))

	loaded, err := config.Load(path)
	require.NoError(t, err)
	require.NotNil(t, loaded.HookSlowMs)
	assert.Equal(t, int64(1_234), *loaded.HookSlowMs)
}

// Duration --------------------------------------------------------------------

func TestDuration_MarshalUnmarshal_String(t *testing.T) {
	d := config.Duration{Duration: 1500 * time.Millisecond}
	raw, err := json.Marshal(d)
	require.NoError(t, err)
	assert.JSONEq(t, `"1.5s"`, string(raw))

	var back config.Duration
	require.NoError(t, json.Unmarshal(raw, &back))
	assert.Equal(t, d.Duration, back.Duration)
}

func TestDuration_Unmarshal_Integer(t *testing.T) {
	var d config.Duration
	require.NoError(t, json.Unmarshal([]byte(`500`), &d))
	assert.Equal(t, 500*time.Millisecond, d.Duration)
}

func TestDuration_Unmarshal_BadString(t *testing.T) {
	var d config.Duration
	err := json.Unmarshal([]byte(`"not-a-duration"`), &d)
	require.Error(t, err)
}

// MultiError ------------------------------------------------------------------

func TestMultiError_FormatsAllFields(t *testing.T) {
	c := config.Config{
		HookFailRatePct: intPtr(200),
		BatchSize:       intPtr(0),
	}
	err := c.Validate()
	require.Error(t, err)

	msg := err.Error()
	assert.Contains(t, msg, "hookFailRatePct")
	assert.Contains(t, msg, "batchSize")
}

// DefaultPath -----------------------------------------------------------------

func TestDefaultPath_EndsWithBuddyConfig(t *testing.T) {
	p, err := config.DefaultPath()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(p, filepath.Join(".buddy", "config.json")), "got %s", p)
}
