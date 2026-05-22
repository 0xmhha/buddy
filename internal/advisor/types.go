// Package advisor implements the rule-driven advisory generator
// (ADR-015). Walks usage metric over configurable
// thresholds; enriches triggered advisories with retrieval evidence
//. 4-surface ship: CLI buddy advise, TUI Usage advisory
// section, MCP usage_advise, daemon advisorMonitor.
//
// Composition:
//
//   Evaluator{Thresholds, Usage, Knowledge}.Run(ctx) → []Advisory
//                  ↓
//   Store (advisories table, migration v7) for persistence + dedup
//                  ↓
//   CLI / TUI / MCP / daemon render or push the result
package advisor

import "time"

// Severity is the urgency hint each advisory carries. The TUI uses it
// for color, MCP for ordering, and the dedup window applies uniformly
// regardless of severity.
type Severity string

const (
	SeverityInfo Severity = "info"
	SeverityWarn Severity = "warn"
	SeverityHigh Severity = "high"
)

// Rule kind identifiers. v0.11.0 shipped five (token / session
// shape); v0.13.0 (ADR-017) adds KindGoalDrift.
const (
	KindTokenSpikeDay    = "token-spike-day"
	KindLongSession      = "long-session"
	KindLowCacheRatio    = "low-cache-ratio"
	KindSessionVolumeDay = "session-volume-day"
	KindTokenDailyCap    = "token-daily-cap"
	KindGoalDrift        = "goal-drift"
)

// AllKinds enumerates the rule kinds in declaration order so callers
// (CLI render, MCP listing) can iterate deterministically.
func AllKinds() []string {
	return []string{
		KindTokenSpikeDay,
		KindLongSession,
		KindLowCacheRatio,
		KindSessionVolumeDay,
		KindTokenDailyCap,
		KindGoalDrift,
	}
}

// Advisory is one piece of advice generated for the user. Persisted
// in the advisories table; also serialised as MCP / CLI output.
//
// ID is zero before insert. CreatedAt is the time the rule fired
// (not the time the row was written — they may differ in `--persist`
// flows where the CLI buffers).
type Advisory struct {
	ID        int64
	Kind      string
	Severity  Severity
	Message   string
	Evidence  []EvidenceItem
	CreatedAt time.Time
	Muted     bool
}

// EvidenceItem backs an advisory's claim. Two flavours:
//
//   - Type="metric" : Detail holds a metric value snippet
//                     (e.g., "오늘 토큰 500k, 7일 평균 200k").
//   - Type="chunk"  : Detail holds a short content preview, ChunkID
//                     points back to chunks.id for further drill-down.
type EvidenceItem struct {
	Type    string `json:"type"`
	Detail  string `json:"detail"`
	ChunkID int64  `json:"chunk_id,omitempty"`
}

// Thresholds are the rule knobs sourced from buddy config (per
// ADR-015 §Q2). Zero values fall back to DefaultThresholds().
type Thresholds struct {
	// Master toggle. When true, Run returns an empty slice.
	Disabled bool

	// TokenSpikeRatio: last-24h tokens / 7d daily-avg ratio that fires
	// KindTokenSpikeDay. Default 1.5.
	TokenSpikeRatio float64

	// LongSessionHours: a single active session whose
	// (last_active - started_at) ≥ this many hours fires KindLongSession.
	// Default 4h.
	LongSessionHours int

	// LowCachePct: last-24h cache-hit ratio percent below this fires
	// KindLowCacheRatio. Default 30 (below 30% → advisory).
	LowCachePct int

	// SessionVolumePerDay: sessions started in last 24h above this
	// fires KindSessionVolumeDay. Default 20.
	SessionVolumePerDay int

	// TokenDailyThreshold: last-24h total tokens above this fires
	// KindTokenDailyCap. Default 500_000.
	TokenDailyThreshold int64

	// GoalDriftDisabled bypasses ruleGoalDrift entirely (ADR-017).
	GoalDriftDisabled bool

	// GoalDriftThreshold: cosine similarity (0..1) below this between
	// goal_text and recent activity fires KindGoalDrift. Default 0.4.
	GoalDriftThreshold float64

	// GoalDriftSampleChunks: chunks-from-end window the drift score
	// averages over (ADR-017 Q2). Default 10.
	GoalDriftSampleChunks int

	// DedupWindow: same Kind doesn't re-fire inside this window.
	// Default 24h.
	DedupWindow time.Duration

	// PollInterval: daemon advisorMonitor cadence. Default 1h.
	PollInterval time.Duration
}

// DefaultThresholds returns spec-locked defaults (ADR-015 §Q2 + ADR-017).
func DefaultThresholds() Thresholds {
	return Thresholds{
		Disabled:              false,
		TokenSpikeRatio:       1.5,
		LongSessionHours:      4,
		LowCachePct:           30,
		SessionVolumePerDay:   20,
		TokenDailyThreshold:   500_000,
		GoalDriftDisabled:     false,
		GoalDriftThreshold:    0.4,
		GoalDriftSampleChunks: 10,
		DedupWindow:           24 * time.Hour,
		PollInterval:          1 * time.Hour,
	}
}

// WithDefaults fills zero fields from DefaultThresholds. Callers pass
// the user's partial config; this guarantees Run gets a complete set
// without forcing the config layer to set every field.
func (t Thresholds) WithDefaults() Thresholds {
	d := DefaultThresholds()
	if t.TokenSpikeRatio == 0 {
		t.TokenSpikeRatio = d.TokenSpikeRatio
	}
	if t.LongSessionHours == 0 {
		t.LongSessionHours = d.LongSessionHours
	}
	if t.LowCachePct == 0 {
		t.LowCachePct = d.LowCachePct
	}
	if t.SessionVolumePerDay == 0 {
		t.SessionVolumePerDay = d.SessionVolumePerDay
	}
	if t.TokenDailyThreshold == 0 {
		t.TokenDailyThreshold = d.TokenDailyThreshold
	}
	if t.GoalDriftThreshold == 0 {
		t.GoalDriftThreshold = d.GoalDriftThreshold
	}
	if t.GoalDriftSampleChunks == 0 {
		t.GoalDriftSampleChunks = d.GoalDriftSampleChunks
	}
	if t.DedupWindow == 0 {
		t.DedupWindow = d.DedupWindow
	}
	if t.PollInterval == 0 {
		t.PollInterval = d.PollInterval
	}
	return t
}
