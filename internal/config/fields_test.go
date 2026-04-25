package config_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/config"
)

// TestFields_AllFieldsPresent — the registry must enumerate every Config knob.
// If a future change adds a field to the Config struct without updating
// fields.go, this test fails so the CLI doesn't silently lose track of it.
func TestFields_AllFieldsPresent(t *testing.T) {
	wantNames := []string{
		"batchSize",
		"hookFailRatePct",
		"hookSlowMs",
		"hookTimeoutMs",
		"notifyChannel",
		"outboxBacklog",
		"personaLocale",
		"pollInterval",
	}
	got := config.Fields()
	gotNames := make([]string, 0, len(got))
	for _, f := range got {
		gotNames = append(gotNames, f.Name)
	}
	assert.Equal(t, wantNames, gotNames, "Fields() must be alphabetical and complete")
}

func TestFieldByName_Unknown_ReturnsFalse(t *testing.T) {
	_, ok := config.FieldByName("doesnotexist")
	assert.False(t, ok)
}

func TestFieldByName_KnownField_ReturnsField(t *testing.T) {
	f, ok := config.FieldByName("hookSlowMs")
	require.True(t, ok)
	assert.Equal(t, "hookSlowMs", f.Name)
}

// TestFields_AllFieldsRoundTrip — for every registered field, Set parses a
// representative valid value, GetOverride returns the formatted string back,
// Get on the resulting Effective() returns the same string. This is the
// invariant the CLI's get/set/show contract leans on.
func TestFields_AllFieldsRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"batchSize", "250"},
		{"hookFailRatePct", "15"},
		{"hookSlowMs", "3000"},
		{"hookTimeoutMs", "20000"},
		{"notifyChannel", "stderr"},
		{"outboxBacklog", "500"},
		{"personaLocale", "en"},
		{"pollInterval", "750ms"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f, ok := config.FieldByName(tc.name)
			require.True(t, ok)

			var c config.Config
			require.NoError(t, f.Set(&c, tc.input), "Set should accept valid input")

			gotOverride, has := f.GetOverride(c)
			require.True(t, has, "GetOverride should report a value after Set")
			assert.Equal(t, tc.input, gotOverride)

			gotEff := f.Get(c.Effective())
			assert.Equal(t, tc.input, gotEff)
		})
	}
}

func TestFields_Unset_ClearsOverride(t *testing.T) {
	f, ok := config.FieldByName("hookSlowMs")
	require.True(t, ok)

	var c config.Config
	require.NoError(t, f.Set(&c, "3000"))
	_, has := f.GetOverride(c)
	require.True(t, has)

	f.Unset(&c)
	_, has = f.GetOverride(c)
	assert.False(t, has, "GetOverride should report no value after Unset")
}

func TestFields_Get_ReturnsDefaultWhenNoOverride(t *testing.T) {
	var c config.Config
	d := config.Defaults()
	for _, f := range config.Fields() {
		_, has := f.GetOverride(c)
		assert.False(t, has, "%s: zero Config has no overrides", f.Name)

		got := f.Get(c.Effective())
		assert.NotEmpty(t, got, "%s: Get on defaults must return some string", f.Name)
		// Sanity-spot-check a few known defaults.
		switch f.Name {
		case "hookTimeoutMs":
			assert.Equal(t, strconv.FormatInt(d.HookTimeoutMs, 10), got)
		case "pollInterval":
			assert.Equal(t, d.PollInterval.String(), got)
		case "personaLocale":
			assert.Equal(t, d.PersonaLocale, got)
		}
	}
}

func TestFields_Set_BadInt_ReturnsError(t *testing.T) {
	f, ok := config.FieldByName("hookSlowMs")
	require.True(t, ok)

	var c config.Config
	err := f.Set(&c, "not-a-number")
	require.Error(t, err)
}

func TestFields_Set_BadDuration_ReturnsError(t *testing.T) {
	f, ok := config.FieldByName("pollInterval")
	require.True(t, ok)

	var c config.Config
	err := f.Set(&c, "not-a-duration")
	require.Error(t, err)
}

// TestFields_Set_AllowsValueValidationToCatchInvalid — the field-level Set is
// purely a parser; range validation lives on Config.Validate. This locks that
// contract: Set("hookFailRatePct", "200") succeeds, but Validate then fails.
func TestFields_Set_AllowsValueValidationToCatchInvalid(t *testing.T) {
	f, ok := config.FieldByName("hookFailRatePct")
	require.True(t, ok)

	var c config.Config
	require.NoError(t, f.Set(&c, "200"))

	err := c.Validate()
	require.Error(t, err)
}

func TestFields_PollInterval_AcceptsZeroValuePreservation(t *testing.T) {
	// Make sure the duration field formats consistently after a round-trip
	// through time.ParseDuration -> Duration.String().
	f, ok := config.FieldByName("pollInterval")
	require.True(t, ok)

	var c config.Config
	require.NoError(t, f.Set(&c, "1s"))
	got := f.Get(c.Effective())
	// time.Duration(1*time.Second).String() == "1s"
	assert.Equal(t, "1s", got)

	require.NoError(t, f.Set(&c, "1500ms"))
	// 1500ms canonicalises to "1.5s".
	assert.Equal(t, "1.5s", f.Get(c.Effective()))

	// Sanity: that's also what Effective.PollInterval.String() returns.
	assert.Equal(t, (1500 * time.Millisecond).String(), f.Get(c.Effective()))
}
