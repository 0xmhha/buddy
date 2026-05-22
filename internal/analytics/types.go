// Package analytics is the adapter layer that backs analytics-mcp tools.
// It owns the schema, the query interface, and the SQLite-flavoured SQL
// reference implementation. Mixpanel / Amplitude / Datadog adapters land
// under the same Adapter interface in later cycles (spec §8).
//
// The package is internal/-only; the MCP wrapper (internal/mcp) is the
// public surface that translates JSON-RPC arguments into these typed
// requests and back.
package analytics

import "time"

// ─── shared inputs ─────────────────────────────────────────────────────────

// TimeRange is a half-open [From, To) window. Both timestamps are UTC.
// Empty values mean "no bound on that side".
type TimeRange struct {
	From time.Time
	To   time.Time
}

// Segment is an optional dimension/value pair used to filter results.
type Segment struct {
	Dimension string
	Value     string
}

// ─── 4.1 funnel ────────────────────────────────────────────────────────────

// FunnelQuery asks "of N stages, how many users entered each in order?".
// Stages must be event_name values present in the events table.
type FunnelQuery struct {
	Product   string // optional namespace when more than one product shares the DB
	Stages    []string
	TimeRange TimeRange
	Segment   *Segment
}

// FunnelStage is one row of the funnel result.
type FunnelStage struct {
	Stage               string  `json:"stage"`
	Count               int64   `json:"count"`
	ConversionFromPrior float64 `json:"conversion_from_prior"`
	DropOffFromPrior    float64 `json:"drop_off_from_prior"`
}

// FunnelResult is the FunnelQuery output.
type FunnelResult struct {
	Funnel        []FunnelStage `json:"funnel"`
	TotalUsers    int64         `json:"total_users"`
	TimeRangeUsed TimeRange     `json:"time_range_used"`
}

// ─── 4.2 cohort ────────────────────────────────────────────────────────────

// CohortDimension is the cohort bucketing strategy.
type CohortDimension string

const (
	CohortWeekly    CohortDimension = "weekly"
	CohortMonthly   CohortDimension = "monthly"
	CohortQuarterly CohortDimension = "quarterly"
)

// RetentionMetric is what counts as "retained" in a cohort.
type RetentionMetric string

const (
	RetentionActive     RetentionMetric = "active"
	RetentionRevenue    RetentionMetric = "revenue"
	RetentionFeatureUse RetentionMetric = "feature_use"
)

// CohortQuery asks for retention curves grouped by acquisition cohort.
type CohortQuery struct {
	CohortDimension CohortDimension
	RetentionMetric RetentionMetric
	Segments        []string // freeform "dim:value" tags applied to signup events
}

// CohortRow is one acquisition cohort's retention curve.
type CohortRow struct {
	CohortID  string             `json:"cohort_id"`
	Size      int64              `json:"size"`
	Retention map[string]float64 `json:"retention"` // keys: D1/D7/D30/D90/D180
	LTV       *float64           `json:"ltv,omitempty"`
	CAC       *float64           `json:"cac,omitempty"`
}

// CohortResult is the CohortQuery output.
type CohortResult struct {
	Cohorts []CohortRow `json:"cohorts"`
}

// ─── 4.3 A/B experiment ────────────────────────────────────────────────────

// ABExperimentQuery names a single experiment to summarise.
type ABExperimentQuery struct {
	ExperimentID string
}

// ABVariant is one variant's measurement.
type ABVariant struct {
	Variant            string     `json:"variant"`
	SampleSize         int64      `json:"sample_size"`
	PrimaryMetricValue float64    `json:"primary_metric_value"`
	ConfidenceInterval [2]float64 `json:"confidence_interval"`
}

// ABSignificance is the hypothesis-test verdict on the primary metric.
type ABSignificance struct {
	PValue                   float64 `json:"p_value"`
	Confidence               float64 `json:"confidence"`
	StatisticallySignificant bool    `json:"statistically_significant"`
}

// ABGuardrail captures one metric outside the primary that must not regress.
type ABGuardrail struct {
	Metric string  `json:"metric"`
	Change float64 `json:"change"`
	Alert  bool    `json:"alert"`
}

// ABExperimentResult is the ABExperimentQuery output.
type ABExperimentResult struct {
	ExperimentID     string         `json:"experiment_id"`
	Name             string         `json:"name"`
	StartedAt        time.Time      `json:"started_at"`
	EndedAt          *time.Time     `json:"ended_at,omitempty"`
	Variants         []ABVariant    `json:"variants"`
	Significance     ABSignificance `json:"significance"`
	GuardrailMetrics []ABGuardrail  `json:"guardrail_metrics,omitempty"`
	Recommendation   string         `json:"recommendation"` // ship | revert | continue | inconclusive
}

// ─── 4.4 actor failure ─────────────────────────────────────────────────────

// ActorType matches the §2 use-case model: user / system / 3rd-party / external-tool.
type ActorType string

// ActorFailureQuery asks for failure stats for one actor type.
type ActorFailureQuery struct {
	Actor     ActorType
	TimeRange TimeRange
}

// ActorFailureResult composites failure_rate, MTTR, blast radius, and
// the Release-It! Nygard 4-dimensional trust score.
type ActorFailureResult struct {
	Actor                   string   `json:"actor"`
	FailureRate             float64  `json:"failure_rate"`
	PredictabilityVariance  float64  `json:"predictability_variance"`
	MTTRMinutes             float64  `json:"mttr_minutes"`
	BlastRadius             int64    `json:"blast_radius"`
	TrustScore              float64  `json:"trust_score"`
	RecoveryPatternsApplied []string `json:"recovery_patterns_applied"`
}

// ─── 4.5 cost ──────────────────────────────────────────────────────────────

// CostDrillDown picks the dimension used to bucket cost totals.
type CostDrillDown string

const (
	CostByService   CostDrillDown = "service"
	CostByComponent CostDrillDown = "component"
	CostByRegion    CostDrillDown = "region"
	CostByAccount   CostDrillDown = "account"
	CostByTag       CostDrillDown = "tag"
)

// CostQuery asks for total cost plus an optional drill-down and anomaly list.
type CostQuery struct {
	TimeRange        TimeRange
	DrillDown        CostDrillDown // empty means "no drill-down"
	AnomalyDetection bool
}

// CostAnomaly is one spike row.
type CostAnomaly struct {
	Dimension string    `json:"dimension"`
	SpikeAt   time.Time `json:"spike_at"`
	CostCents int64     `json:"cost_cents"`
	ZScore    float64   `json:"z_score"`
}

// CostResult is the CostQuery output.
type CostResult struct {
	TotalCostCents int64            `json:"total_cost_cents"`
	ByDimension    map[string]int64 `json:"by_dimension"`
	Anomalies      []CostAnomaly    `json:"anomalies,omitempty"`
}

// ─── 4.6 SLO burn ──────────────────────────────────────────────────────────

// SLOBurnQuery asks for current SLI value, budget remaining, and a
// release-gate decision under a given time_window.
type SLOBurnQuery struct {
	SLI        string
	SLOTarget  float64
	TimeWindow string // 1h | 6h | 1d | 3d | 30d
}

// SLOBurnResult is the SLOBurnQuery output.
type SLOBurnResult struct {
	SLI                 string  `json:"sli"`
	SLOTarget           float64 `json:"slo_target"`
	TimeWindow          string  `json:"time_window"`
	CurrentValue        float64 `json:"current_value"`
	BudgetRemainingPct  float64 `json:"budget_remaining_pct"`
	BurnRate            float64 `json:"burn_rate"`
	AlertLevel          string  `json:"alert_level"`           // info | warning | critical
	ReleaseGateDecision string  `json:"release_gate_decision"` // ship | limit | freeze | rollback
}

// ─── 4.7 feedback corpus ───────────────────────────────────────────────────

// FeedbackSource matches the spec source enum.
type FeedbackSource string

const (
	FeedbackCSTicket   FeedbackSource = "cs_ticket"
	FeedbackNPSComment FeedbackSource = "nps_comment"
	FeedbackAppReview  FeedbackSource = "app_review"
	FeedbackInterview  FeedbackSource = "interview"
)

// FeedbackQuery asks for corpus aggregates, optional topic clustering, and
// optional sentiment band.
type FeedbackQuery struct {
	Source            FeedbackSource // empty means "all sources"
	TimeRange         TimeRange
	TopicModeling     bool
	SentimentAnalysis bool
}

// FeedbackTopic is one cluster.
type FeedbackTopic struct {
	Topic          string           `json:"topic"`
	ItemCount      int64            `json:"item_count"`
	Sentiment      map[string]int64 `json:"sentiment,omitempty"` // positive/negative/neutral counts
	VerbatimQuotes []string         `json:"verbatim_quotes"`
}

// NPSBand is one of the three NPS segments.
type NPSBand struct {
	Count     int64    `json:"count"`
	TopTopics []string `json:"top_topics"`
}

// FeedbackResult is the FeedbackQuery output.
type FeedbackResult struct {
	TotalItems  int64              `json:"total_items"`
	Topics      []FeedbackTopic    `json:"topics,omitempty"`
	NPSSegments map[string]NPSBand `json:"nps_segments,omitempty"` // keys: promoter/passive/detractor
}
