package advisor

import (
	"fmt"
	"time"

	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/usage"
)

// Snapshot is the fully-fetched input the rules evaluate against. The
// evaluator builds one Snapshot per Run() call (single fetch round-trip)
// so each rule reuses cached metric values instead of hitting the DB.
//
// Window24h / Window7d are the relative windows the rules compare; the
// concrete bounds are filled by the evaluator with respect to `Now`.
type Snapshot struct {
	Now            time.Time
	Window24h      usage.TimeWindow
	Spend24h       usage.TokenSpend
	Stats24h       usage.SessionStats
	Window7d       usage.TimeWindow
	Spend7d        usage.TokenSpend
	ActiveSessions []sessions.Session

	// DriftItems are per-active-session drift scores, populated by
	// the evaluator when Knowledge + Embedder are wired and the
	// session has goal_text + ≥ SampleChunks chunks. Empty slice =
	// drift detection skipped (no embedder, no chunks, or disabled).
	DriftItems []SessionDrift
}

// SessionDrift is one row of goal-vs-activity comparison output.
// Score is cosine similarity in [0,1] — higher means "current activity
// still matches the original goal". WorstChunk holds a short preview
// of the most-distant chunk for Evidence.
type SessionDrift struct {
	SessionID    string
	GoalText     string
	Score        float64
	SampleChunks int
	WorstChunk   string
}

// ruleFn produces an advisory or nil when the rule doesn't trigger.
// Persona-rendered Message text is set here; the evaluator may add
// retrieval evidence afterwards.
type ruleFn func(t Thresholds, snap Snapshot) *Advisory

// allRules returns every rule in display order. Adding a rule = append
// here + provide the impl + (when applicable) add a persona key.
func allRules() []ruleFn {
	return []ruleFn{
		ruleTokenSpikeDay,
		ruleLongSession,
		ruleLowCacheRatio,
		ruleSessionVolumeDay,
		ruleTokenDailyCap,
		ruleGoalDrift,
	}
}

// ruleTokenSpikeDay — fires when last-24h tokens / (last-7d tokens / 7)
// >= TokenSpikeRatio. Severity scales with the ratio: ≥2x → high,
// otherwise warn.
//
// Why divide by 7: Spend7d sums 7 days worth of usage; the ratio compares
// "today" to "average day in the past week". Skipping the divisor
// systematically under-flags spikes for active weeks.
func ruleTokenSpikeDay(t Thresholds, snap Snapshot) *Advisory {
	dayTokens := float64(snap.Spend24h.TotalTokens())
	weekTokens := float64(snap.Spend7d.TotalTokens())
	if weekTokens <= 0 {
		return nil // no baseline yet; can't decide spike
	}
	avg := weekTokens / 7
	if avg <= 0 {
		return nil
	}
	ratio := dayTokens / avg
	if ratio < t.TokenSpikeRatio {
		return nil
	}
	severity := SeverityWarn
	if ratio >= 2.0 {
		severity = SeverityHigh
	}
	return &Advisory{
		Kind:     KindTokenSpikeDay,
		Severity: severity,
		Message: fmt.Sprintf(
			"오늘 토큰 사용이 평소보다 %.1fx야 (오늘 %s / 일평균 %s).",
			ratio,
			humanCount(int64(dayTokens)),
			humanCount(int64(avg)),
		),
		Evidence: []EvidenceItem{
			{Type: "metric", Detail: fmt.Sprintf("ratio=%.2f threshold=%.2f", ratio, t.TokenSpikeRatio)},
			{Type: "metric", Detail: fmt.Sprintf("day=%d week=%d", int64(dayTokens), int64(weekTokens))},
		},
		CreatedAt: snap.Now,
	}
}

// ruleLongSession — fires when any ACTIVE session's runtime exceeds
// LongSessionHours. The longest active session is named in the message;
// only one advisory per Run regardless of how many sessions exceed.
func ruleLongSession(t Thresholds, snap Snapshot) *Advisory {
	threshold := time.Duration(t.LongSessionHours) * time.Hour
	if threshold <= 0 || len(snap.ActiveSessions) == 0 {
		return nil
	}
	var longest sessions.Session
	var longestDur time.Duration
	for _, s := range snap.ActiveSessions {
		dur := s.LastActive.Sub(s.StartedAt)
		if dur > longestDur {
			longestDur = dur
			longest = s
		}
	}
	if longestDur < threshold {
		return nil
	}
	severity := SeverityWarn
	if longestDur >= 2*threshold {
		severity = SeverityHigh
	}
	shortID := longest.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	goal := longest.GoalText
	if goal == "" {
		goal = "(목적 미기록)"
	} else {
		goal = truncateToRunes(goal, 60)
	}
	return &Advisory{
		Kind:     KindLongSession,
		Severity: severity,
		Message: fmt.Sprintf(
			"세션 %s 이 %s 째 진행 중이야. 잠깐 멈추고 정리해볼래? — 목적: %s",
			shortID, shortDur(longestDur), goal,
		),
		Evidence: []EvidenceItem{
			{Type: "metric", Detail: fmt.Sprintf("session=%s duration=%s threshold=%s",
				longest.ID, longestDur.Round(time.Minute), threshold)},
		},
		CreatedAt: snap.Now,
	}
}

// ruleLowCacheRatio — fires when last-24h cache hit ratio % is below
// the configured floor. Severity rises sharply past zero usage.
func ruleLowCacheRatio(t Thresholds, snap Snapshot) *Advisory {
	denom := snap.Spend24h.InputTokens + snap.Spend24h.CacheReadTokens + snap.Spend24h.CacheCreateTokens
	if denom <= 0 {
		return nil // no input tokens today; ratio is undefined
	}
	pct := snap.Spend24h.CacheHitRatio() * 100
	if int(pct) >= t.LowCachePct {
		return nil
	}
	severity := SeverityInfo
	if pct < float64(t.LowCachePct)/2 {
		severity = SeverityWarn
	}
	return &Advisory{
		Kind:     KindLowCacheRatio,
		Severity: severity,
		Message: fmt.Sprintf(
			"오늘 cache hit ratio 가 %.0f%% 야 (기준 %d%% 이상). 프롬프트 재사용 패턴을 검토해봐.",
			pct, t.LowCachePct,
		),
		Evidence: []EvidenceItem{
			{Type: "metric", Detail: fmt.Sprintf("pct=%.1f threshold=%d", pct, t.LowCachePct)},
		},
		CreatedAt: snap.Now,
	}
}

// ruleSessionVolumeDay — fires when last-24h session count exceeds the
// threshold. Suggests work is being fragmented across many short
// sessions (which often means goal-text isn't sticking).
func ruleSessionVolumeDay(t Thresholds, snap Snapshot) *Advisory {
	if int64(t.SessionVolumePerDay) <= 0 {
		return nil
	}
	if snap.Stats24h.TotalSessions <= int64(t.SessionVolumePerDay) {
		return nil
	}
	severity := SeverityInfo
	if snap.Stats24h.TotalSessions >= 2*int64(t.SessionVolumePerDay) {
		severity = SeverityWarn
	}
	return &Advisory{
		Kind:     KindSessionVolumeDay,
		Severity: severity,
		Message: fmt.Sprintf(
			"24시간 안에 세션 %d개 열렸어 (기준 %d 이하 권장). 작업이 너무 잘려있는지 봐.",
			snap.Stats24h.TotalSessions, t.SessionVolumePerDay,
		),
		Evidence: []EvidenceItem{
			{Type: "metric", Detail: fmt.Sprintf("sessions24h=%d threshold=%d",
				snap.Stats24h.TotalSessions, t.SessionVolumePerDay)},
		},
		CreatedAt: snap.Now,
	}
}

// ruleTokenDailyCap — absolute daily token spend cap. Useful for users
// who care about budget regardless of personal baseline.
func ruleTokenDailyCap(t Thresholds, snap Snapshot) *Advisory {
	if t.TokenDailyThreshold <= 0 {
		return nil
	}
	total := snap.Spend24h.TotalTokens()
	if total <= t.TokenDailyThreshold {
		return nil
	}
	severity := SeverityWarn
	if total >= 2*t.TokenDailyThreshold {
		severity = SeverityHigh
	}
	return &Advisory{
		Kind:     KindTokenDailyCap,
		Severity: severity,
		Message: fmt.Sprintf(
			"24시간 토큰 사용량이 %s 넘었어 (기준 %s). 잠시 쉬거나 작업 범위를 좁혀봐.",
			humanCount(total), humanCount(t.TokenDailyThreshold),
		),
		Evidence: []EvidenceItem{
			{Type: "metric", Detail: fmt.Sprintf("total=%d threshold=%d", total, t.TokenDailyThreshold)},
		},
		CreatedAt: snap.Now,
	}
}

// ruleGoalDrift — fires when any active session's goal-vs-activity
// cosine similarity drops below GoalDriftThreshold. Picks the session
// with the LOWEST score (most drifted) as the advisory's target; only
// one drift advisory per Run so the user isn't flooded when multiple
// sessions drift simultaneously.
//
// Returns nil when:
//   - GoalDriftDisabled is true
//   - DriftItems is empty (evaluator skipped drift — no embedder or no chunks)
//   - every DriftItem.Score >= threshold (no session has drifted)
func ruleGoalDrift(t Thresholds, snap Snapshot) *Advisory {
	if t.GoalDriftDisabled || len(snap.DriftItems) == 0 {
		return nil
	}
	threshold := t.GoalDriftThreshold
	var worst SessionDrift
	worst.Score = 1.0 // start at "perfectly aligned" so the first drifted item wins
	var found bool
	for _, d := range snap.DriftItems {
		if d.Score >= threshold {
			continue
		}
		if !found || d.Score < worst.Score {
			worst = d
			found = true
		}
	}
	if !found {
		return nil
	}
	severity := SeverityWarn
	if worst.Score < threshold/2 {
		severity = SeverityHigh
	}
	shortID := worst.SessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	goal := truncateToRunes(worst.GoalText, 60)
	a := &Advisory{
		Kind:     KindGoalDrift,
		Severity: severity,
		Message: fmt.Sprintf(
			"세션 %s 이 처음 목적에서 벗어나는 것 같아 (cosine %.2f, 기준 %.2f). 원래 목적: %s — 의도된 거야?",
			shortID, worst.Score, threshold, goal,
		),
		Evidence: []EvidenceItem{
			{Type: "metric", Detail: fmt.Sprintf("session=%s score=%.3f threshold=%.2f sample_chunks=%d",
				worst.SessionID, worst.Score, threshold, worst.SampleChunks)},
		},
		CreatedAt: snap.Now,
	}
	if worst.WorstChunk != "" {
		a.Evidence = append(a.Evidence, EvidenceItem{
			Type:   "chunk",
			Detail: truncateToRunes(worst.WorstChunk, 120),
		})
	}
	return a
}

// ─── helpers ─────────────────────────────────────────────────────────

// humanCount formats integers with k/M suffix for readability in
// advisory messages. 1234 → "1.2k"; 1_500_000 → "1.5M".
func humanCount(n int64) string {
	switch {
	case n < 1_000:
		return fmt.Sprintf("%d", n)
	case n < 1_000_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
}

// truncateToRunes returns s clipped to at most maxRunes runes with "…"
// appended when truncation actually occurred. Operating on []rune avoids
// splitting Korean glyphs (3 bytes/char in UTF-8) or emoji surrogates at
// a multi-byte boundary, which a plain byte slice would do.
func truncateToRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}

// shortDur is the same idea as humanCount but for durations — "2h30m"
// is more readable than "2h30m0s" Go-default for friend-tone text.
func shortDur(d time.Duration) string {
	d = d.Round(time.Minute)
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) - h*60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
