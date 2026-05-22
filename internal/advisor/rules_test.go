package advisor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/usage"
)

func baseThresholds() Thresholds { return DefaultThresholds() }

func TestRuleTokenSpikeDay_FiresWhenRatioCrossed(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	snap := Snapshot{
		Now:      now,
		Spend24h: usage.TokenSpend{InputTokens: 600_000},
		Spend7d:  usage.TokenSpend{InputTokens: 1_400_000}, // 200k/day avg
	}
	// ratio 600k / 200k = 3 — over default 1.5
	a := ruleTokenSpikeDay(baseThresholds(), snap)
	require.NotNil(t, a)
	require.Equal(t, KindTokenSpikeDay, a.Kind)
	require.Equal(t, SeverityHigh, a.Severity, "≥2x → high")
	require.Contains(t, a.Message, "토큰 사용")
}

func TestRuleTokenSpikeDay_NoSpike(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	snap := Snapshot{
		Now:      now,
		Spend24h: usage.TokenSpend{InputTokens: 250_000},
		Spend7d:  usage.TokenSpend{InputTokens: 1_400_000}, // 200k/day; ratio 1.25
	}
	require.Nil(t, ruleTokenSpikeDay(baseThresholds(), snap))
}

func TestRuleTokenSpikeDay_NoBaselineNoFire(t *testing.T) {
	t.Parallel()
	snap := Snapshot{Now: time.Now().UTC(), Spend24h: usage.TokenSpend{InputTokens: 500_000}}
	require.Nil(t, ruleTokenSpikeDay(baseThresholds(), snap))
}

func TestRuleLongSession_FiresOnActiveOver4h(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	long := sessions.Session{
		ID: "abcdef1234", StartedAt: now.Add(-5 * time.Hour), LastActive: now,
		GoalText: "implement F2.C advisor",
	}
	short := sessions.Session{
		ID: "shortone", StartedAt: now.Add(-30 * time.Minute), LastActive: now,
	}
	snap := Snapshot{Now: now, ActiveSessions: []sessions.Session{short, long}}
	a := ruleLongSession(baseThresholds(), snap)
	require.NotNil(t, a)
	require.Equal(t, KindLongSession, a.Kind)
	require.Contains(t, a.Message, "abcdef12") // short id
	require.Contains(t, a.Message, "advisor")  // goal text
}

func TestRuleLongSession_NoActiveNoFire(t *testing.T) {
	t.Parallel()
	require.Nil(t, ruleLongSession(baseThresholds(), Snapshot{Now: time.Now().UTC()}))
}

func TestRuleLowCacheRatio_FiresWhenBelowThreshold(t *testing.T) {
	t.Parallel()
	snap := Snapshot{
		Now:      time.Now().UTC(),
		Spend24h: usage.TokenSpend{InputTokens: 800, CacheReadTokens: 100, CacheCreateTokens: 100},
		// denom=1000, cache_read=100 → 10% < 30% default
	}
	a := ruleLowCacheRatio(baseThresholds(), snap)
	require.NotNil(t, a)
	require.Equal(t, KindLowCacheRatio, a.Kind)
}

func TestRuleLowCacheRatio_NoFireWhenHealthy(t *testing.T) {
	t.Parallel()
	snap := Snapshot{
		Spend24h: usage.TokenSpend{InputTokens: 100, CacheReadTokens: 500, CacheCreateTokens: 100},
	}
	require.Nil(t, ruleLowCacheRatio(baseThresholds(), snap))
}

func TestRuleLowCacheRatio_NoDataNoFire(t *testing.T) {
	t.Parallel()
	require.Nil(t, ruleLowCacheRatio(baseThresholds(), Snapshot{}))
}

func TestRuleSessionVolumeDay_FiresOverThreshold(t *testing.T) {
	t.Parallel()
	snap := Snapshot{Stats24h: usage.SessionStats{TotalSessions: 25}}
	a := ruleSessionVolumeDay(baseThresholds(), snap)
	require.NotNil(t, a)
	require.Equal(t, KindSessionVolumeDay, a.Kind)
}

func TestRuleSessionVolumeDay_DoubleThresholdRaisesSeverity(t *testing.T) {
	t.Parallel()
	snap := Snapshot{Stats24h: usage.SessionStats{TotalSessions: 50}}
	a := ruleSessionVolumeDay(baseThresholds(), snap)
	require.NotNil(t, a)
	require.Equal(t, SeverityWarn, a.Severity)
}

func TestRuleTokenDailyCap_FiresOverThreshold(t *testing.T) {
	t.Parallel()
	snap := Snapshot{
		Spend24h: usage.TokenSpend{InputTokens: 600_000}, // > 500k default
	}
	a := ruleTokenDailyCap(baseThresholds(), snap)
	require.NotNil(t, a)
	require.Equal(t, KindTokenDailyCap, a.Kind)
}

func TestRuleTokenDailyCap_DoubleThresholdHigh(t *testing.T) {
	t.Parallel()
	snap := Snapshot{
		Spend24h: usage.TokenSpend{InputTokens: 1_100_000},
	}
	a := ruleTokenDailyCap(baseThresholds(), snap)
	require.NotNil(t, a)
	require.Equal(t, SeverityHigh, a.Severity)
}

func TestHumanCount(t *testing.T) {
	t.Parallel()
	require.Equal(t, "42", humanCount(42))
	require.Equal(t, "1.2k", humanCount(1234))
	require.Equal(t, "1.5M", humanCount(1_500_000))
}

func TestShortDur(t *testing.T) {
	t.Parallel()
	require.Equal(t, "30m", shortDur(30*time.Minute))
	require.Equal(t, "2h", shortDur(2*time.Hour))
	require.Equal(t, "2h30m", shortDur(2*time.Hour+30*time.Minute))
}

// ─── ruleGoalDrift (ADR-017) ──────────────────────────────────────────

func TestRuleGoalDrift_DisabledNoFire(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds()
	thr.GoalDriftDisabled = true
	snap := Snapshot{Now: time.Now().UTC(), DriftItems: []SessionDrift{
		{SessionID: "s1", GoalText: "g", Score: 0.05, SampleChunks: 10},
	}}
	require.Nil(t, ruleGoalDrift(thr, snap))
}

func TestRuleGoalDrift_NoDriftItemsNoFire(t *testing.T) {
	t.Parallel()
	require.Nil(t, ruleGoalDrift(DefaultThresholds(), Snapshot{Now: time.Now().UTC()}))
}

func TestRuleGoalDrift_AboveThresholdNoFire(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds() // threshold 0.4
	snap := Snapshot{Now: time.Now().UTC(), DriftItems: []SessionDrift{
		{SessionID: "s1", GoalText: "g", Score: 0.6, SampleChunks: 10},
	}}
	require.Nil(t, ruleGoalDrift(thr, snap))
}

func TestRuleGoalDrift_FiresWhenBelowThreshold(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds()
	snap := Snapshot{Now: time.Now().UTC(), DriftItems: []SessionDrift{
		{SessionID: "sess-abc123", GoalText: "design F2.D", Score: 0.25,
			SampleChunks: 10, WorstChunk: "chunk talking about something else entirely"},
	}}
	a := ruleGoalDrift(thr, snap)
	require.NotNil(t, a)
	require.Equal(t, KindGoalDrift, a.Kind)
	require.Equal(t, SeverityWarn, a.Severity)
	require.Contains(t, a.Message, "sess-abc")
	require.Contains(t, a.Message, "design F2.D")
	// Evidence: metric + chunk.
	require.GreaterOrEqual(t, len(a.Evidence), 2)
}

func TestRuleGoalDrift_HighSeverityWhenFarBelow(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds() // threshold 0.4 → half = 0.2
	snap := Snapshot{Now: time.Now().UTC(), DriftItems: []SessionDrift{
		{SessionID: "s1", GoalText: "g", Score: 0.1, SampleChunks: 10},
	}}
	a := ruleGoalDrift(thr, snap)
	require.NotNil(t, a)
	require.Equal(t, SeverityHigh, a.Severity, "score < threshold/2 → high")
}

func TestRuleGoalDrift_PicksLowestScoredSession(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds()
	snap := Snapshot{Now: time.Now().UTC(), DriftItems: []SessionDrift{
		{SessionID: "mild", GoalText: "g", Score: 0.35, SampleChunks: 10},
		{SessionID: "severe", GoalText: "g", Score: 0.05, SampleChunks: 10},
		{SessionID: "ok", GoalText: "g", Score: 0.9, SampleChunks: 10},
	}}
	a := ruleGoalDrift(thr, snap)
	require.NotNil(t, a)
	require.Contains(t, a.Message, "severe", "worst-drifted session is named")
}

// truncateToRunes must clip on rune boundaries so a Korean character
// never gets split mid-byte (a plain byte slice would yield "" with the
// trailing UTF-8 bytes orphaned). Each Korean glyph is 3 bytes in UTF-8;
// the test exercises an exact-boundary case + a short-input no-op case.
func TestTruncateToRunes_PreservesKoreanGlyphBoundary(t *testing.T) {
	t.Parallel()

	// "한국어로 작성된 긴 목적 텍스트입니다" — 17 Korean characters.
	korean := "한국어로 작성된 긴 목적 텍스트입니다"

	// Truncate to 5 runes: must keep exactly "한국어로 " and append "…",
	// not a partial byte sequence.
	got := truncateToRunes(korean, 5)
	require.Equal(t, "한국어로 …", got)

	// No truncation when input is shorter than the budget.
	require.Equal(t, "짧음", truncateToRunes("짧음", 5))

	// Empty / zero budget edge cases.
	require.Equal(t, "", truncateToRunes("", 5))
	require.Equal(t, "", truncateToRunes("anything", 0))

	// ASCII path stays correct (one rune == one byte).
	require.Equal(t, "abcde…", truncateToRunes("abcdefghij", 5))
}

// Boundary-pair tests for the rule predicates. Each surviving mutant
// from the previous coverage audit corresponded to a one-sided
// threshold check (one side tested, the other never exercised). The
// pairs below pin both sides so a future mutation that flips < to <=
// or > to >= breaks at least one assertion.

func TestRuleTokenSpikeDay_RatioAtExactThresholdFires(t *testing.T) {
	t.Parallel()
	// Default TokenSpikeRatio is 1.5. Choose 24h vs 7d numbers so the
	// computed ratio lands exactly on the threshold — ratio < threshold
	// must NOT fire, ratio == threshold MUST fire.
	thr := DefaultThresholds()
	now := time.Now().UTC()

	// 300k / (1_400_000 / 7) = 300k / 200k = 1.5 exactly.
	at := Snapshot{
		Now:      now,
		Spend24h: usage.TokenSpend{InputTokens: 300_000},
		Spend7d:  usage.TokenSpend{InputTokens: 1_400_000},
	}
	require.NotNil(t, ruleTokenSpikeDay(thr, at),
		"ratio == TokenSpikeRatio is the >= side and must fire")

	// 200k / 200k = 1.0 — strictly below threshold, no fire.
	below := Snapshot{
		Now:      now,
		Spend24h: usage.TokenSpend{InputTokens: 200_000},
		Spend7d:  usage.TokenSpend{InputTokens: 1_400_000},
	}
	require.Nil(t, ruleTokenSpikeDay(thr, below))
}

func TestRuleTokenSpikeDay_SeverityHighBoundaryAtRatio2(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds()
	now := time.Now().UTC()

	// 400k / 200k = 2.0 exactly — the >= 2.0 side, so high.
	at := Snapshot{
		Now:      now,
		Spend24h: usage.TokenSpend{InputTokens: 400_000},
		Spend7d:  usage.TokenSpend{InputTokens: 1_400_000},
	}
	a := ruleTokenSpikeDay(thr, at)
	require.NotNil(t, a)
	require.Equal(t, SeverityHigh, a.Severity, "ratio == 2.0 must be high")

	// 399_999 / 200_000 ≈ 1.999995 — strictly under 2.0, so warn.
	below := Snapshot{
		Now:      now,
		Spend24h: usage.TokenSpend{InputTokens: 399_999},
		Spend7d:  usage.TokenSpend{InputTokens: 1_400_000},
	}
	a = ruleTokenSpikeDay(thr, below)
	require.NotNil(t, a)
	require.Equal(t, SeverityWarn, a.Severity, "ratio < 2.0 must stay warn")
}

func TestRuleLongSession_DurationAtExactThresholdFires(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds()
	threshold := time.Duration(thr.LongSessionHours) * time.Hour
	now := time.Now().UTC()

	atSnap := Snapshot{
		Now: now,
		ActiveSessions: []sessions.Session{{
			ID: "exactly-4h", PID: 1,
			StartedAt:  now.Add(-threshold),
			LastActive: now,
			GoalText:   "g",
		}},
	}
	require.NotNil(t, ruleLongSession(thr, atSnap),
		"dur == LongSessionHours is the >= side and must fire")

	belowSnap := Snapshot{
		Now: now,
		ActiveSessions: []sessions.Session{{
			ID: "just-under", PID: 1,
			StartedAt:  now.Add(-threshold + time.Minute),
			LastActive: now,
			GoalText:   "g",
		}},
	}
	require.Nil(t, ruleLongSession(thr, belowSnap),
		"dur < LongSessionHours must not fire")
}

func TestRuleLongSession_SeverityHighBoundaryAtDoubleThreshold(t *testing.T) {
	t.Parallel()
	thr := DefaultThresholds()
	threshold := time.Duration(thr.LongSessionHours) * time.Hour
	now := time.Now().UTC()

	atDouble := Snapshot{
		Now: now,
		ActiveSessions: []sessions.Session{{
			ID: "exactly-2x", PID: 1,
			StartedAt:  now.Add(-2 * threshold),
			LastActive: now,
			GoalText:   "g",
		}},
	}
	a := ruleLongSession(thr, atDouble)
	require.NotNil(t, a)
	require.Equal(t, SeverityHigh, a.Severity, "dur == 2*threshold must be high")

	belowDouble := Snapshot{
		Now: now,
		ActiveSessions: []sessions.Session{{
			ID: "just-under-2x", PID: 1,
			StartedAt:  now.Add(-2*threshold + time.Minute),
			LastActive: now,
			GoalText:   "g",
		}},
	}
	a = ruleLongSession(thr, belowDouble)
	require.NotNil(t, a)
	require.Equal(t, SeverityWarn, a.Severity, "dur < 2*threshold must stay warn")
}
