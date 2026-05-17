package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestPreviewSchedule_Standard5Field — a typical 5-field cron string ("every
// minute") computes the next fire time at the next whole minute.
func TestPreviewSchedule_Standard5Field(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 15, 12, 34, 17, 0, time.UTC)
	got := PreviewSchedule("* * * * *", now)
	require.NoError(t, got.Err)
	require.Equal(t, "* * * * *", got.Schedule)
	require.Equal(t,
		time.Date(2026, 5, 15, 12, 35, 0, 0, time.UTC),
		got.Next,
		"every-minute cron must round up to the next whole minute")
}

// TestPreviewSchedule_Descriptor — @daily must parse and compute the next
// midnight (cron descriptor support is what the scheduler comment promised).
func TestPreviewSchedule_Descriptor(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 15, 12, 34, 17, 0, time.UTC)
	got := PreviewSchedule("@daily", now)
	require.NoError(t, got.Err)
	require.Equal(t,
		time.Date(2026, 5, 16, 0, 0, 0, 0, time.UTC),
		got.Next,
		"@daily at 12:34 must roll forward to the next 00:00")
}

// TestPreviewSchedule_Empty — an empty schedule is on-demand, not an error.
func TestPreviewSchedule_Empty(t *testing.T) {
	t.Parallel()
	got := PreviewSchedule("", time.Now())
	require.NoError(t, got.Err)
	require.True(t, got.Next.IsZero(), "empty schedule must produce a zero Next time")
}

// TestPreviewSchedule_InvalidStringRetainsErrorAndSchedule — a typo'd cron
// expression returns the offending string in Schedule + a non-nil Err. This
// is what lets the TUI render "(invalid: <msg>)" next to the bad row rather
// than wiping the whole pane.
func TestPreviewSchedule_InvalidStringRetainsErrorAndSchedule(t *testing.T) {
	t.Parallel()
	got := PreviewSchedule("not a cron string", time.Now())
	require.Error(t, got.Err)
	require.Equal(t, "not a cron string", got.Schedule)
	require.True(t, got.Next.IsZero(), "invalid input must not leak a bogus Next time")
}
