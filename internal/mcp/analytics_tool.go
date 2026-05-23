package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/0xmhha/buddy/internal/analytics"
)

// analytics_tool.go wires the 7 analytics_query_* MCP tools to a
// pluggable analytics.Adapter. Friend-tone stub behaviour is the
// fallback: if Options.Analytics is nil (no adapter configured), every
// handler returns guidance text. With an adapter, the handler delegates
// to it and renders the typed result.
//
// Args carry MCP-side jsonschema annotations; results reuse analytics
// package types directly.

// ─── shared MCP-side args ─────────────────────────────────────────────────

type timeRangeArg struct {
	From string `json:"from" jsonschema:"ISO-8601 start (inclusive). e.g. 2026-04-01T00:00:00Z"`
	To   string `json:"to"   jsonschema:"ISO-8601 end (exclusive). e.g. 2026-05-01T00:00:00Z"`
}

func (a timeRangeArg) toAnalytics() (analytics.TimeRange, error) {
	var out analytics.TimeRange
	if a.From != "" {
		t, err := time.Parse(time.RFC3339, a.From)
		if err != nil {
			return out, fmt.Errorf("time_range.from: %w", err)
		}
		out.From = t
	}
	if a.To != "" {
		t, err := time.Parse(time.RFC3339, a.To)
		if err != nil {
			return out, fmt.Errorf("time_range.to: %w", err)
		}
		out.To = t
	}
	return out, nil
}

type segmentArg struct {
	Dimension string `json:"dimension" jsonschema:"Segmentation dimension (e.g. plan_tier, country)."`
	Value     string `json:"value"     jsonschema:"Value within the dimension to filter on."`
}

// ─── argument types per tool ──────────────────────────────────────────────

type funnelArgs struct {
	Product   string       `json:"product,omitempty" jsonschema:"Product identifier when more than one product shares this backend."`
	Stages    []string     `json:"stages"            jsonschema:"Ordered funnel stages, e.g. [signup, first_action, retention_d7]."`
	TimeRange timeRangeArg `json:"time_range"        jsonschema:"Reporting window."`
	Segment   *segmentArg  `json:"segment,omitempty" jsonschema:"Optional segmentation filter."`
}

type cohortArgs struct {
	CohortDimension string   `json:"cohort_dimension"      jsonschema:"weekly | monthly | quarterly"`
	RetentionMetric string   `json:"retention_metric"      jsonschema:"active | revenue | feature_use"`
	Segments        []string `json:"segments,omitempty"    jsonschema:"Optional segment filters."`
}

type abExperimentArgs struct {
	ExperimentID string `json:"experiment_id" jsonschema:"Experiment identifier in the backend."`
}

type actorFailureArgs struct {
	Actor     string       `json:"actor"      jsonschema:"user | system | 3rd-party | external-tool"`
	TimeRange timeRangeArg `json:"time_range" jsonschema:"Reporting window."`
}

type costArgs struct {
	TimeRange        timeRangeArg `json:"time_range"                  jsonschema:"Reporting window."`
	DrillDown        string       `json:"drill_down,omitempty"        jsonschema:"service | component | region | account | tag"`
	AnomalyDetection bool         `json:"anomaly_detection,omitempty" jsonschema:"When true, return spike rows in addition to totals."`
}

type sloBurnArgs struct {
	SLI        string  `json:"sli"         jsonschema:"SLI identifier (e.g. p99_latency_ms)."`
	SLOTarget  float64 `json:"slo_target"  jsonschema:"SLO target value (units match the SLI)."`
	TimeWindow string  `json:"time_window" jsonschema:"1h | 6h | 1d | 3d | 30d"`
}

type feedbackCorpusArgs struct {
	Source            string       `json:"source,omitempty"             jsonschema:"cs_ticket | nps_comment | app_review | interview"`
	TimeRange         timeRangeArg `json:"time_range"                   jsonschema:"Reporting window."`
	TopicModeling     bool         `json:"topic_modeling,omitempty"     jsonschema:"When true, run topic modelling on items."`
	SentimentAnalysis bool         `json:"sentiment_analysis,omitempty" jsonschema:"When true, score each topic with positive/negative/neutral."`
}

// ─── env-var helpers (kept so the stub path stays observable) ─────────────

func resolveAnalyticsBackend() (string, error) {
	backend := os.Getenv("BUDDY_ANALYTICS_BACKEND")
	if backend == "" {
		return "", errBackendNotConfigured
	}
	switch backend {
	case "sql", "mixpanel", "amplitude", "datadog", "stripe", "elasticsearch":
		return backend, errBackendStubOnly
	default:
		return "", fmt.Errorf("unknown analytics backend %q (set BUDDY_ANALYTICS_BACKEND to one of: sql, mixpanel, amplitude, datadog, stripe, elasticsearch)", backend)
	}
}

var (
	errBackendNotConfigured = errors.New(
		"analytics backend not configured — set BUDDY_ANALYTICS_BACKEND to one of: sql, mixpanel, amplitude, datadog, stripe, elasticsearch",
	)
	errBackendStubOnly = errors.New(
		"analytics backend recognised but the adapter is not implemented yet (analytics-mcp ships stubs only — see docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md §8 for the adapter roadmap)",
	)
)

func notConfiguredResponse(toolName string, err error) *mcp.CallToolResult {
	body := fmt.Sprintf(
		"analytics-mcp/%s 호출은 받았는데, 데이터 백엔드 연결이 아직 안 돼 있어.\n\n%s\n\n"+
			"임시 처리: 본 tool 의 응답은 빈 값. 본격 데이터 query 는 adapter 구현 후.\n"+
			"요청한 시각: %s",
		toolName, err.Error(), time.Now().UTC().Format(time.RFC3339),
	)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: body}}}
}

// jsonContent renders a typed result as a single TextContent JSON blob so
// Claude can either parse it back or quote it verbatim. We pretty-print so
// session transcripts stay readable.
func jsonContent(v any) *mcp.CallToolResult {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{
			Text: fmt.Sprintf("analytics-mcp: marshal result: %v", err),
		}}}
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(buf)}}}
}

// ─── registration ─────────────────────────────────────────────────────────

func addAnalyticsTools(s *mcp.Server, opts Options) {
	adapter := opts.Analytics

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_funnel",
		Description: "Funnel stage counts + conversion/drop-off rates for analyze-feature-adoption and optimize-conversion-funnel skills. Reads from BUDDY_ANALYTICS_BACKEND (sql/mixpanel/amplitude/datadog/stripe/elasticsearch). With BUDDY_ANALYTICS_BACKEND=sql + BUDDY_ANALYTICS_DSN, queries the local SQL reference adapter; otherwise returns guidance text.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args funnelArgs) (*mcp.CallToolResult, analytics.FunnelResult, error) {
		if adapter == nil {
			_, err := resolveAnalyticsBackend()
			return notConfiguredResponse("analytics_query_funnel", err), analytics.FunnelResult{}, nil
		}
		tr, err := args.TimeRange.toAnalytics()
		if err != nil {
			return nil, analytics.FunnelResult{}, err
		}
		var seg *analytics.Segment
		if args.Segment != nil && args.Segment.Dimension != "" {
			seg = &analytics.Segment{Dimension: args.Segment.Dimension, Value: args.Segment.Value}
		}
		res, err := adapter.QueryFunnel(ctx, analytics.FunnelQuery{
			Product: args.Product, Stages: args.Stages, TimeRange: tr, Segment: seg,
		})
		if err != nil {
			return nil, analytics.FunnelResult{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_cohort",
		Description: "Acquisition cohort retention curves (D1/D7/D30/D90/D180) for analyze-user-cohort skill. Optional LTV/CAC when revenue + attribution data exist. SQL reference adapter buckets users by their first 'signup' event.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args cohortArgs) (*mcp.CallToolResult, analytics.CohortResult, error) {
		if adapter == nil {
			_, err := resolveAnalyticsBackend()
			return notConfiguredResponse("analytics_query_cohort", err), analytics.CohortResult{}, nil
		}
		res, err := adapter.QueryCohort(ctx, analytics.CohortQuery{
			CohortDimension: analytics.CohortDimension(args.CohortDimension),
			RetentionMetric: analytics.RetentionMetric(args.RetentionMetric),
			Segments:        args.Segments,
		})
		if err != nil {
			return nil, analytics.CohortResult{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_ab_experiment",
		Description: "A/B test results (variants + significance + guardrails + recommendation) for analyze-ab-experiment skill. SQL reference adapter computes Welch's t-test on the experiment's primary metric.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args abExperimentArgs) (*mcp.CallToolResult, analytics.ABExperimentResult, error) {
		if adapter == nil {
			_, err := resolveAnalyticsBackend()
			return notConfiguredResponse("analytics_query_ab_experiment", err), analytics.ABExperimentResult{}, nil
		}
		res, err := adapter.QueryABExperiment(ctx, analytics.ABExperimentQuery{ExperimentID: args.ExperimentID})
		if errors.Is(err, analytics.ErrNotFound) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("analytics-mcp: experiment %q not found in backend", args.ExperimentID),
			}}}, analytics.ABExperimentResult{}, nil
		}
		if err != nil {
			return nil, analytics.ABExperimentResult{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_actor_failure",
		Description: "Per-actor failure rate + MTTR + trust score (Release-It Nygard 4-dim) for analyze-actor-failure-rate skill. SQL reference adapter reads the failures table.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args actorFailureArgs) (*mcp.CallToolResult, analytics.ActorFailureResult, error) {
		if adapter == nil {
			_, err := resolveAnalyticsBackend()
			return notConfiguredResponse("analytics_query_actor_failure", err), analytics.ActorFailureResult{}, nil
		}
		tr, err := args.TimeRange.toAnalytics()
		if err != nil {
			return nil, analytics.ActorFailureResult{}, err
		}
		res, err := adapter.QueryActorFailure(ctx, analytics.ActorFailureQuery{
			Actor: analytics.ActorType(args.Actor), TimeRange: tr,
		})
		if err != nil {
			return nil, analytics.ActorFailureResult{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_cost",
		Description: "Cost timeline + drill-down + anomaly detection for analyze-cost-anomaly and audit-cost-efficiency skills. SQL reference adapter sums cost_records grouped by service/component/region/account/tag, with z-score anomaly detection.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args costArgs) (*mcp.CallToolResult, analytics.CostResult, error) {
		if adapter == nil {
			_, err := resolveAnalyticsBackend()
			return notConfiguredResponse("analytics_query_cost", err), analytics.CostResult{}, nil
		}
		tr, err := args.TimeRange.toAnalytics()
		if err != nil {
			return nil, analytics.CostResult{}, err
		}
		res, err := adapter.QueryCost(ctx, analytics.CostQuery{
			TimeRange: tr, DrillDown: analytics.CostDrillDown(args.DrillDown), AnomalyDetection: args.AnomalyDetection,
		})
		if err != nil {
			return nil, analytics.CostResult{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_slo_burn",
		Description: "SLO burn rate (multi-window) + release-gate decision for audit-error-budget skill. SQL reference adapter implements the Google SRE multi-window thresholds (burn_rate ≥14 → rollback, ≥6 → freeze, ≥3 → limit, ≥1 → ship+warning).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args sloBurnArgs) (*mcp.CallToolResult, analytics.SLOBurnResult, error) {
		if adapter == nil {
			_, err := resolveAnalyticsBackend()
			return notConfiguredResponse("analytics_query_slo_burn", err), analytics.SLOBurnResult{}, nil
		}
		res, err := adapter.QuerySLOBurn(ctx, analytics.SLOBurnQuery{
			SLI: args.SLI, SLOTarget: args.SLOTarget, TimeWindow: args.TimeWindow,
		})
		if errors.Is(err, analytics.ErrNotFound) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("analytics-mcp: no observations for SLI %q in the requested window", args.SLI),
			}}}, analytics.SLOBurnResult{}, nil
		}
		if err != nil {
			return nil, analytics.SLOBurnResult{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "analytics_query_feedback_corpus",
		Description: "Text corpus search + topic modelling + sentiment + NPS band split for analyze-customer-feedback-corpus and triage-customer-support-ticket skills. SQL reference adapter aggregates by source + optionally clusters by pre-classified topic/sentiment columns.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args feedbackCorpusArgs) (*mcp.CallToolResult, analytics.FeedbackResult, error) {
		if adapter == nil {
			_, err := resolveAnalyticsBackend()
			return notConfiguredResponse("analytics_query_feedback_corpus", err), analytics.FeedbackResult{}, nil
		}
		tr, err := args.TimeRange.toAnalytics()
		if err != nil {
			return nil, analytics.FeedbackResult{}, err
		}
		res, err := adapter.QueryFeedbackCorpus(ctx, analytics.FeedbackQuery{
			Source: analytics.FeedbackSource(args.Source), TimeRange: tr,
			TopicModeling: args.TopicModeling, SentimentAnalysis: args.SentimentAnalysis,
		})
		if err != nil {
			return nil, analytics.FeedbackResult{}, err
		}
		return jsonContent(res), res, nil
	})
}
