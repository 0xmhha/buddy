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
