package purge_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/purge"
)

func TestParseBefore_DateForm(t *testing.T) {
	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	got, err := purge.ParseBefore("2026-04-01", now)
	require.NoError(t, err)
	want := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, got.Equal(want), "got %v want %v", got, want)
}

func TestParseBefore_RFC3339(t *testing.T) {
	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	got, err := purge.ParseBefore("2026-04-01T12:34:56Z", now)
	require.NoError(t, err)
	want := time.Date(2026, 4, 1, 12, 34, 56, 0, time.UTC)
	assert.True(t, got.Equal(want), "got %v want %v", got, want)
}

func TestParseBefore_RelativeDays(t *testing.T) {
	now := time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC)
	got, err := purge.ParseBefore("7d", now)
	require.NoError(t, err)
	want := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, got.Equal(want), "got %v want %v", got, want)
}

// TestParseBefore_RelativeDays_Zero — "0d" is technically allowed by the regex
// and means "now". Useful for tests / dry-runs that want to count everything;
// not blocked.
func TestParseBefore_RelativeDays_Zero(t *testing.T) {
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	got, err := purge.ParseBefore("0d", now)
	require.NoError(t, err)
	assert.True(t, got.Equal(now))
}

func TestParseBefore_RelativeDays_Multidigit(t *testing.T) {
	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	got, err := purge.ParseBefore("90d", now)
	require.NoError(t, err)
	want := now.AddDate(0, 0, -90)
	assert.True(t, got.Equal(want), "got %v want %v", got, want)
}

func TestParseBefore_InvalidForm_ReturnsError(t *testing.T) {
	now := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)
	for _, bad := range []string{
		"",
		"tomorrow",
		"yesterday",
		"-7d",
		"7days",
		"d7",
		"2026/04/01",
		"2026-04-01T00:00:00", // missing tz, RFC3339 requires Z
	} {
		t.Run(bad, func(t *testing.T) {
			_, err := purge.ParseBefore(bad, now)
			assert.Error(t, err, "expected %q to be rejected", bad)
		})
	}
}
