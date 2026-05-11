package mcp

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// analytics_tool.go implements Phase W4-2.1 of the analytics-mcp spec
// (docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md):
// register 7 MCP tools (funnel/cohort/ab_experiment/actor_failure/cost/
// slo_burn/feedback_corpus) so the §8 Cluster F skills (analyze-feature-adoption,
// analyze-user-cohort, analyze-actor-failure-rate, analyze-cost-anomaly,
// triage-customer-support-ticket, analyze-customer-feedback-corpus,
// audit-error-budget) can call them directly.
//
// v0.2.0 ships stubs only: every handler reports "backend not configured"
// in friend-tone Korean text unless BUDDY_ANALYTICS_BACKEND is set. Real
// adapters (custom SQL / Mixpanel / Amplitude / Datadog) land in W4-2.2+.
//
// Args/result shapes mirror spec §4 verbatim so future adapter work is
// purely handler-internal — no signature churn.

// ─── shared input types ─────────────────────────────────────────────────────

type timeRange struct {
	From string `json:"from" jsonschema:"ISO-8601 start (inclusive). e.g. 2026-04-01T00:00:00Z"`
	To   string `json:"to"   jsonschema:"ISO-8601 end (exclusive). e.g. 2026-05-01T00:00:00Z"`
}

type segment struct {
	Dimension string `json:"dimension" jsonschema:"Segmentation dimension (e.g. plan_tier, country)."`
	Value     string `json:"value"     jsonschema:"Value within the dimension to filter on."`
}

// ─── analytics_query_funnel — spec §4.1 ────────────────────────────────────

type funnelArgs struct {
	Product   string    `json:"product,omitempty"  jsonschema:"Product identifier when more than one product shares this backend."`
	Stages    []string  `json:"stages"             jsonschema:"Ordered funnel stages, e.g. [signup, first_action, retention_d7]."`
	TimeRange timeRange `json:"time_range"         jsonschema:"Reporting window."`
	Segment   *segment  `json:"segment,omitempty"  jsonschema:"Optional segmentation filter."`
}

type funnelStage struct {
	Stage                 string  `json:"stage"`
	Count                 int64   `json:"count"`
	ConversionFromPrior   float64 `json:"conversion_from_prior"`
	DropOffFromPrior      float64 `json:"drop_off_from_prior"`
}

type funnelResult struct {
	Funnel        []funnelStage `json:"funnel"`
	TotalUsers    int64         `json:"total_users"`
	TimeRangeUsed timeRange     `json:"time_range_used"`
}

// ─── analytics_query_cohort — spec §4.2 ────────────────────────────────────

type cohortArgs struct {
	CohortDimension string   `json:"cohort_dimension" jsonschema:"weekly | monthly | quarterly"`
	RetentionMetric string   `json:"retention_metric" jsonschema:"active | revenue | feature_use"`
	Segments        []string `json:"segments,omitempty" jsonschema:"Optional segment filters."`
}

type cohortRow struct {
	CohortID  string             `json:"cohort_id"`
	Size      int64              `json:"size"`
	Retention map[string]float64 `json:"retention"` // D1, D7, D30, D90, D180
	LTV       *float64           `json:"ltv,omitempty"`
	CAC       *float64           `json:"cac,omitempty"`
}

type cohortResult struct {
	Cohorts []cohortRow `json:"cohorts"`
}

// ─── analytics_query_ab_experiment — spec §4.3 ─────────────────────────────

type abExperimentArgs struct {
	ExperimentID string `json:"experiment_id" jsonschema:"Experiment identifier in the backend."`
}

type abVariant struct {
	Variant            string     `json:"variant"`
	SampleSize         int64      `json:"sample_size"`
	PrimaryMetricValue float64    `json:"primary_metric_value"`
	ConfidenceInterval [2]float64 `json:"confidence_interval"`
}

type abSignificance struct {
	PValue                    float64 `json:"p_value"`
	Confidence                float64 `json:"confidence"`
	StatisticallySignificant  bool    `json:"statistically_significant"`
}

type abGuardrail struct {
	Metric string  `json:"metric"`
	Change float64 `json:"change"`
	Alert  bool    `json:"alert"`
}

type abExperimentResult struct {
	Experiment struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		StartedAt string  `json:"started_at"`
		EndedAt   *string `json:"ended_at,omitempty"`
	} `json:"experiment"`
	Variants         []abVariant    `json:"variants"`
	Significance     abSignificance `json:"significance"`
	GuardrailMetrics []abGuardrail  `json:"guardrail_metrics,omitempty"`
	Recommendation   string         `json:"recommendation" jsonschema:"ship | revert | continue | inconclusive"`
}

// ─── analytics_query_actor_failure — spec §4.4 ─────────────────────────────

type actorFailureArgs struct {
	Actor     string    `json:"actor"      jsonschema:"user | system | 3rd-party | external-tool"`
	TimeRange timeRange `json:"time_range" jsonschema:"Reporting window."`
}

type actorFailureResult struct {
	Actor                    string   `json:"actor"`
	FailureRate              float64  `json:"failure_rate"`
	PredictabilityVariance   float64  `json:"predictability_variance"`
	MTTRMinutes              float64  `json:"mttr_minutes"`
	BlastRadius              int64    `json:"blast_radius"`
	TrustScore               float64  `json:"trust_score"`
	RecoveryPatternsApplied  []string `json:"recovery_patterns_applied"`
}

// ─── analytics_query_cost — spec §4.5 ──────────────────────────────────────

type costArgs struct {
	TimeRange        timeRange `json:"time_range"               jsonschema:"Reporting window."`
	DrillDown        string    `json:"drill_down,omitempty"     jsonschema:"service | component | region | account | tag"`
	AnomalyDetection bool      `json:"anomaly_detection,omitempty" jsonschema:"When true, return spike rows in addition to totals."`
}

type costAnomaly struct {
	Dimension string  `json:"dimension"`
	SpikeAt   string  `json:"spike_at"`
	CostCents int64   `json:"cost_cents"`
	ZScore    float64 `json:"z_score"`
}

type costResult struct {
	TotalCostCents int64            `json:"total_cost_cents"`
	ByDimension    map[string]int64 `json:"by_dimension"`
	Anomalies      []costAnomaly    `json:"anomalies,omitempty"`
}

// ─── analytics_query_slo_burn — spec §4.6 ──────────────────────────────────

type sloBurnArgs struct {
	SLI        string  `json:"sli"         jsonschema:"SLI identifier (e.g. p99_latency_ms)."`
	SLOTarget  float64 `json:"slo_target"  jsonschema:"SLO target value (units match the SLI)."`
	TimeWindow string  `json:"time_window" jsonschema:"1h | 6h | 1d | 3d | 30d"`
}

type sloBurnResult struct {
	SLI                  string  `json:"sli"`
	SLOTarget            float64 `json:"slo_target"`
	TimeWindow           string  `json:"time_window"`
	CurrentValue         float64 `json:"current_value"`
	BudgetRemainingPct   float64 `json:"budget_remaining_pct"`
	BurnRate             float64 `json:"burn_rate"`
	AlertLevel           string  `json:"alert_level"           jsonschema:"info | warning | critical"`
	ReleaseGateDecision  string  `json:"release_gate_decision" jsonschema:"ship | limit | freeze | rollback"`
}

// ─── analytics_query_feedback_corpus — spec §4.7 ───────────────────────────

type feedbackCorpusArgs struct {
	Source            string    `json:"source,omitempty"             jsonschema:"cs_ticket | nps_comment | app_review | interview"`
	TimeRange         timeRange `json:"time_range"                   jsonschema:"Reporting window."`
	TopicModeling     bool      `json:"topic_modeling,omitempty"     jsonschema:"When true, run topic modelling on items."`
	SentimentAnalysis bool      `json:"sentiment_analysis,omitempty" jsonschema:"When true, score each topic with positive/negative/neutral."`
}

type feedbackTopic struct {
	Topic           string         `json:"topic"`
	ItemCount       int64          `json:"item_count"`
	Sentiment       map[string]int64 `json:"sentiment,omitempty"`
	VerbatimQuotes  []string       `json:"verbatim_quotes"`
}

type npsBand struct {
	Count     int64    `json:"count"`
	TopTopics []string `json:"top_topics"`
}

type feedbackCorpusResult struct {
	TotalItems  int64                  `json:"total_items"`
	Topics      []feedbackTopic        `json:"topics,omitempty"`
	NPSSegments map[string]npsBand     `json:"nps_segments,omitempty"`
}

// ─── adapter resolution ────────────────────────────────────────────────────

// resolveAnalyticsBackend reads the BUDDY_ANALYTICS_BACKEND environment
// variable. Until W4-2.2 lands a real adapter, every supported value returns
// notConfiguredErr — the handlers funnel it into a friend-tone text response.
func resolveAnalyticsBackend() (string, error) {
	backend := os.Getenv("BUDDY_ANALYTICS_BACKEND")
	if backend == "" {
		return "", errBackendNotConfigured
	}
	switch backend {
	case "sql", "mixpanel", "amplitude", "datadog", "stripe", "elasticsearch":
		// Recognised but no adapter ships in v0.2.0.
		return backend, errBackendStubOnly
	default:
		return "", fmt.Errorf("unknown analytics backend %q (set BUDDY_ANALYTICS_BACKEND to one of: sql, mixpanel, amplitude, datadog, stripe, elasticsearch)", backend)
	}
}

// Sentinel errors used by all 7 stub handlers. Real adapters (W4-2.2+)
// keep these as the *unconfigured* signal and add their own typed errors.
var (
	errBackendNotConfigured = fmt.Errorf(
		"analytics backend not configured — set BUDDY_ANALYTICS_BACKEND to one of: sql, mixpanel, amplitude, datadog, stripe, elasticsearch",
	)
	errBackendStubOnly = fmt.Errorf(
		"analytics backend recognised but the adapter is not implemented yet (analytics-mcp v0.2.0 ships stubs only — see docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md §8 phase W4-2.2)",
	)
)

// notConfiguredResponse renders a friend-tone text body for the MCP client.
// Returning text (not just error) lets Claude surface the guidance verbatim
// to the user rather than swallowing it as a transport error.
func notConfiguredResponse(toolName string, err error) *mcp.CallToolResult {
	body := fmt.Sprintf(
		"analytics-mcp/%s 호출은 받았는데, 데이터 백엔드 연결이 아직 안 돼 있어.\n\n%s\n\n"+
			"임시 처리: 본 tool 의 응답은 빈 값. 본격 데이터 query 는 adapter 구현 (W4-2.2~W4-2.3) 후.\n"+
			"요청한 시각: %s",
		toolName, err.Error(), time.Now().UTC().Format(time.RFC3339),
	)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: body}},
	}
}

// ─── registration ──────────────────────────────────────────────────────────

func addAnalyticsTools(s *mcp.Server, _ Options) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_funnel",
		Description: "Funnel stage counts + conversion/drop-off rates for analyze-feature-adoption and optimize-conversion-funnel skills. Reads from BUDDY_ANALYTICS_BACKEND (sql/mixpanel/amplitude/datadog/stripe/elasticsearch). Stubbed in v0.2.0 — real adapter lands in spec §8 phase W4-2.2.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ funnelArgs) (*mcp.CallToolResult, funnelResult, error) {
		_, err := resolveAnalyticsBackend()
		return notConfiguredResponse("analytics_query_funnel", err), funnelResult{}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_cohort",
		Description: "Acquisition cohort retention curves (D1/D7/D30/D90/D180) for analyze-user-cohort skill. Optional LTV/CAC when revenue + attribution data exist. Stubbed in v0.2.0.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ cohortArgs) (*mcp.CallToolResult, cohortResult, error) {
		_, err := resolveAnalyticsBackend()
		return notConfiguredResponse("analytics_query_cohort", err), cohortResult{}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_ab_experiment",
		Description: "A/B test results (variants + significance + guardrails + recommendation) for analyze-ab-experiment skill. Stubbed in v0.2.0.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ abExperimentArgs) (*mcp.CallToolResult, abExperimentResult, error) {
		_, err := resolveAnalyticsBackend()
		return notConfiguredResponse("analytics_query_ab_experiment", err), abExperimentResult{}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_actor_failure",
		Description: "Per-actor failure rate + MTTR + trust score (Release-It Nygard 4-dim) for analyze-actor-failure-rate skill. Stubbed in v0.2.0.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ actorFailureArgs) (*mcp.CallToolResult, actorFailureResult, error) {
		_, err := resolveAnalyticsBackend()
		return notConfiguredResponse("analytics_query_actor_failure", err), actorFailureResult{}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_cost",
		Description: "Cost timeline + drill-down + anomaly detection for analyze-cost-anomaly and audit-cost-efficiency skills. Stubbed in v0.2.0.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ costArgs) (*mcp.CallToolResult, costResult, error) {
		_, err := resolveAnalyticsBackend()
		return notConfiguredResponse("analytics_query_cost", err), costResult{}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_slo_burn",
		Description: "SLO burn rate (multi-window) + release-gate decision for audit-error-budget skill. Stubbed in v0.2.0.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ sloBurnArgs) (*mcp.CallToolResult, sloBurnResult, error) {
		_, err := resolveAnalyticsBackend()
		return notConfiguredResponse("analytics_query_slo_burn", err), sloBurnResult{}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_feedback_corpus",
		Description: "Text corpus search + topic modelling + sentiment + NPS band split for analyze-customer-feedback-corpus and triage-customer-support-ticket skills. Stubbed in v0.2.0.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ feedbackCorpusArgs) (*mcp.CallToolResult, feedbackCorpusResult, error) {
		_, err := resolveAnalyticsBackend()
		return notConfiguredResponse("analytics_query_feedback_corpus", err), feedbackCorpusResult{}, nil
	})
}
