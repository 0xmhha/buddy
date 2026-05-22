package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/usage"
)

// The CLI render helpers were 0% covered before this file landed: they
// run only when a real DB and live data are present. Each test below
// passes a deterministic fixture and asserts the salient field names +
// values + formatting markers are present, without locking the exact
// byte-for-byte layout — a future copy edit to the headers should not
// invalidate the contract. The intent is regression detection on the
// data → string mapping, not pixel-level pinning.

func TestRenderTokenSpend_PopulatesAllSlots(t *testing.T) {
	t.Parallel()
	spend := usage.TokenSpend{
		InputTokens:       1000,
		OutputTokens:      500,
		CacheReadTokens:   200,
		CacheCreateTokens: 100,
	}
	got := renderTokenSpend(spend)

	require.Contains(t, got, "토큰 사용량", "header must render")
	require.Contains(t, got, "1,800", "total = input + output + read + create")
	require.Contains(t, got, "1,000", "input renders with separator")
	require.Contains(t, got, "500", "output present")
	require.Contains(t, got, "200", "cache_read present")
	require.Contains(t, got, "100", "cache_create present")
	require.Contains(t, got, "cache hit:", "cache hit ratio line present")
}

func TestRenderSessionStats_DurationLineOnlyWhenEnded(t *testing.T) {
	t.Parallel()
	// With ended sessions: p50/p90/max line should render.
	withEnded := usage.SessionStats{
		TotalSessions:  10,
		ActiveSessions: 3,
		EndedSessions:  7,
		DurationP50:    30 * time.Minute,
		DurationP90:    90 * time.Minute,
		DurationMax:    2*time.Hour + 15*time.Minute,
		GoalTextRatio:  0.6,
	}
	got := renderSessionStats(withEnded)
	require.Contains(t, got, "세션 통계")
	require.Contains(t, got, "10")
	require.Contains(t, got, "active 3")
	require.Contains(t, got, "ended 7")
	require.Contains(t, got, "duration:", "duration line renders when any sessions ended")
	require.Contains(t, got, "p50=")
	require.Contains(t, got, "60%", "goal ratio formatted as %0.f%%")

	// Without any ended session: duration line must be omitted entirely.
	allActive := usage.SessionStats{
		TotalSessions:  3,
		ActiveSessions: 3,
		EndedSessions:  0,
		GoalTextRatio:  0,
	}
	got = renderSessionStats(allActive)
	require.NotContains(t, got, "duration:",
		"all-active case skips duration percentile line")
	require.Contains(t, got, "0%", "zero goal ratio still rendered")
}

func TestRenderTopSessions_TruncatesIDAndGoal(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 22, 10, 30, 0, 0, time.UTC)
	top := []usage.TopSession{
		{
			ID:          "abcdefghijklmnop", // 16 chars; render truncates to 8
			StartedAt:   now,
			GoalText:    strings.Repeat("가", 80), // 80 Korean chars; render byte-slices at 60
			TotalTokens: 5000,
		},
	}
	got := renderTopSessions(top)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	require.Len(t, lines, 2, "header + 1 row")
	require.Contains(t, lines[0], "ID")
	require.Contains(t, lines[0], "TOKENS")
	require.Contains(t, lines[0], "GOAL")
	require.Contains(t, lines[1], "abcdefgh",
		"ID truncates to first 8 characters")
	require.NotContains(t, lines[1], "abcdefghi",
		"truncation removes character 9 onward")
	require.Contains(t, lines[1], "5,000")
}

func TestRenderTimeDistribution_AllZeroPath(t *testing.T) {
	t.Parallel()
	// Total 0 → header only, no per-hour rows.
	empty := usage.TimeDistribution{Total: 0}
	got := renderTimeDistribution(empty)
	require.Contains(t, got, "시간대 분포")
	require.Contains(t, got, "총 0 세션")
	require.NotContains(t, got, "00시:",
		"empty distribution stops before the per-hour loop")

	// Non-empty path renders 24 hour rows.
	hot := usage.TimeDistribution{Total: 5}
	hot.HourCounts[14] = 3
	hot.HourCounts[15] = 2
	got = renderTimeDistribution(hot)
	for h := 0; h < 24; h++ {
		require.Contains(t, got,
			// %02d format pads the hour with a leading zero.
			"  "+twoDigit(h)+"시:",
			"every hour from 00 to 23 renders")
	}
}

func TestRenderAdvisories_BlankSliceProducesEmptyString(t *testing.T) {
	t.Parallel()
	require.Equal(t, "", renderAdvisories(nil),
		"no advisories → empty output (no spurious header)")

	advs := []advisor.Advisory{
		{Kind: "token-spike-day", Severity: advisor.SeverityHigh,
			Message: "오늘 토큰이 평소보다 3x"},
		{Kind: "long-session", Severity: advisor.SeverityWarn,
			Message: "세션이 6시간째 진행",
			Evidence: []advisor.EvidenceItem{
				{Type: "metric", Detail: "duration=6h"},
			}},
	}
	got := renderAdvisories(advs)
	require.Contains(t, got, "[token-spike-day · high]")
	require.Contains(t, got, "오늘 토큰이 평소보다 3x")
	require.Contains(t, got, "[long-session · warn]")
	require.Contains(t, got, "근거:", "evidence block surfaces under the advisory")
	require.Contains(t, got, "[metric] duration=6h")
}

func TestRenderAdvisoryHistory_TruncatesLongMessage(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", 120) // 120 bytes; renderer cuts at 60 + "…"
	advs := []advisor.Advisory{
		{ID: 42, Kind: "k", Severity: advisor.SeverityWarn,
			CreatedAt: time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC),
			Message:   long},
		{ID: 7, Kind: "m", Severity: advisor.SeverityInfo,
			CreatedAt: time.Date(2026, 5, 22, 11, 0, 0, 0, time.UTC),
			Message:   "short", Muted: true},
	}
	got := renderAdvisoryHistory(advs)
	require.Contains(t, got, "KIND")
	require.Contains(t, got, "CREATED")
	require.Contains(t, got, "42")
	require.Contains(t, got, strings.Repeat("a", 60)+"…",
		"message > 60 chars is truncated with ellipsis")
	require.Contains(t, got, "(muted)",
		"muted advisories carry a (muted) suffix")
}

func TestRenderDailyBarChart_EmptyAndAllZeroProduceEmpty(t *testing.T) {
	t.Parallel()
	require.Equal(t, "", renderDailyBarChart(nil),
		"nil series → empty string, no header noise")

	allZero := []usage.DailySpend{{Date: time.Now()}, {Date: time.Now().Add(24 * time.Hour)}}
	require.Equal(t, "", renderDailyBarChart(allZero),
		"all-zero series → empty (chart has no signal)")

	mixed := []usage.DailySpend{
		{Date: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC), InputTokens: 1000},
		{Date: time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC), InputTokens: 500},
	}
	got := renderDailyBarChart(mixed)
	require.Contains(t, got, "일별 토큰 추이")
	require.Contains(t, got, "peak", "header surfaces the peak label")
	require.Contains(t, got, "█",
		"non-zero day produces at least one bar block")
}

// twoDigit formats h with a leading zero like %02d for the per-hour
// substring assertion. Used inline to keep the test self-contained.
func twoDigit(h int) string {
	if h < 10 {
		return "0" + string(rune('0'+h))
	}
	return string(rune('0'+h/10)) + string(rune('0'+h%10))
}
