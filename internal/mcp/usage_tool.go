package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/0xmhha/buddy/internal/usage"
)

// usage_tool.go wires the 5 `usage_query_*` MCP tools per ADR-013.
// Read-only over the sessions table. Each tool takes an optional ISO
// `since` (and `until` for time-bounded calls) plus tool-specific args.
//
// The advisor consumes these tools as LLM input to derive actionable
// Korean prose recommendations.

// ─── shared time-range arg (analytics_tool 와 형식 일치) ─────────────────

type usageRangeArg struct {
	Since string `json:"since,omitempty" jsonschema:"ISO-8601 start (inclusive). e.g. 2026-04-01T00:00:00Z. Empty = all time."`
	Until string `json:"until,omitempty" jsonschema:"ISO-8601 end (exclusive). e.g. 2026-05-01T00:00:00Z. Empty = now."`
}

func (a usageRangeArg) toWindow() (usage.TimeWindow, error) {
	var out usage.TimeWindow
	if a.Since != "" {
		t, err := time.Parse(time.RFC3339, a.Since)
		if err != nil {
			return out, fmt.Errorf("since: %w", err)
		}
		out.Since = t
	}
	if a.Until != "" {
		t, err := time.Parse(time.RFC3339, a.Until)
		if err != nil {
			return out, fmt.Errorf("until: %w", err)
		}
		out.Until = t
	}
	return out, nil
}

// ─── arg types per tool ───────────────────────────────────────────────

type usageTokenSpendArgs struct {
	Range usageRangeArg `json:"range,omitempty" jsonschema:"Time window. Empty = all-time."`
}

type usageSessionStatsArgs struct {
	Range usageRangeArg `json:"range,omitempty" jsonschema:"Time window. Empty = all-time."`
}

type usageTimeDistributionArgs struct {
	Range usageRangeArg `json:"range,omitempty" jsonschema:"Time window. Empty = all-time."`
}

type usageTopSessionsArgs struct {
	Range usageRangeArg `json:"range,omitempty" jsonschema:"Time window. Empty = all-time."`
	Limit int           `json:"limit,omitempty" jsonschema:"Max rows. Default 10."`
}

// usageTopSessionsResult wraps []TopSession so the MCP SDK has an
// object root for its output schema (arrays-at-root are rejected).
type usageTopSessionsResult struct {
	Sessions []usage.TopSession `json:"sessions"`
}

type usageOverviewArgs struct {
	Range usageRangeArg `json:"range,omitempty" jsonschema:"Time window. Empty = all-time."`
	Top   int           `json:"top,omitempty"   jsonschema:"Top sessions limit. Default 5."`
}

// ─── registration ─────────────────────────────────────────────────────

func addUsageTools(s *mcp.Server, opts Options) {
	svc := opts.Usage

	mcp.AddTool(s, &mcp.Tool{
		Name:        "usage_query_token_spend",
		Description: "Sum of token usage (input + output + cache_read + cache_create) over the given window. Reads from buddy's sessions table populated by the session monitor (ADR-012). Returns aggregate counts and the cache hit ratio.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args usageTokenSpendArgs) (*mcp.CallToolResult, usage.TokenSpend, error) {
		if svc == nil {
			return usageNotWired("usage_query_token_spend"), usage.TokenSpend{}, nil
		}
		w, err := args.Range.toWindow()
		if err != nil {
			return nil, usage.TokenSpend{}, err
		}
		res, err := svc.QueryTokenSpend(ctx, w)
		if err != nil {
			return nil, usage.TokenSpend{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "usage_query_session_stats",
		Description: "Per-window session counts (total / active / ended), duration percentiles (p50 / p90 / max over ended sessions), and the goal_text coverage ratio. Useful for the advisor to detect 'long sessions' or 'sessions without recorded goal'.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args usageSessionStatsArgs) (*mcp.CallToolResult, usage.SessionStats, error) {
		if svc == nil {
			return usageNotWired("usage_query_session_stats"), usage.SessionStats{}, nil
		}
		w, err := args.Range.toWindow()
		if err != nil {
			return nil, usage.SessionStats{}, err
		}
		res, err := svc.QuerySessionStats(ctx, w)
		if err != nil {
			return nil, usage.SessionStats{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "usage_query_time_distribution",
		Description: "Hour-of-day histogram of session starts (local time of the buddy host). Used to surface peak-hour patterns to the user.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args usageTimeDistributionArgs) (*mcp.CallToolResult, usage.TimeDistribution, error) {
		if svc == nil {
			return usageNotWired("usage_query_time_distribution"), usage.TimeDistribution{}, nil
		}
		w, err := args.Range.toWindow()
		if err != nil {
			return nil, usage.TimeDistribution{}, err
		}
		res, err := svc.QueryTimeDistribution(ctx, w)
		if err != nil {
			return nil, usage.TimeDistribution{}, err
		}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "usage_query_top_sessions",
		Description: "Top-N sessions in the window sorted by total tokens DESC. Each row includes id, started_at, goal_text, total_tokens. Useful for 'which sessions used the most tokens this week?' style queries.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args usageTopSessionsArgs) (*mcp.CallToolResult, usageTopSessionsResult, error) {
		if svc == nil {
			return usageNotWired("usage_query_top_sessions"), usageTopSessionsResult{}, nil
		}
		w, err := args.Range.toWindow()
		if err != nil {
			return nil, usageTopSessionsResult{}, err
		}
		rows, err := svc.QueryTopSessions(ctx, w, args.Limit)
		if err != nil {
			return nil, usageTopSessionsResult{}, err
		}
		res := usageTopSessionsResult{Sessions: rows}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "usage_query_overview",
		Description: "Combined snapshot: token_spend + session_stats + time_distribution + top sessions in a single call. The TUI Usage pane and `buddy usage overview` CLI use this. Cheaper than four separate tool calls for the advisor's one-shot summarisation.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args usageOverviewArgs) (*mcp.CallToolResult, usage.Overview, error) {
		if svc == nil {
			return usageNotWired("usage_query_overview"), usage.Overview{}, nil
		}
		w, err := args.Range.toWindow()
		if err != nil {
			return nil, usage.Overview{}, err
		}
		topN := args.Top
		if topN <= 0 {
			topN = 5
		}
		res, err := svc.QueryOverview(ctx, w, topN)
		if err != nil {
			return nil, usage.Overview{}, err
		}
		return jsonContent(res), res, nil
	})
}

// usageNotWired mirrors notConfiguredResponse but for the usage suite —
// when Options.Usage is nil the tool registers (so Claude can see it)
// but each handler returns a friend-tone note instead of querying.
// This path triggers when the server is constructed without `Sessions
// → Usage` wiring (e.g., an old caller path).
func usageNotWired(toolName string) *mcp.CallToolResult {
	body := fmt.Sprintf(
		"usage-mcp/%s 호출은 받았는데, sessions store 연결이 안 돼 있어.\n"+
			"`buddy mcp serve` 가 sessions 테이블이 있는 buddy.db 를 가리키는지 확인해줘.",
		toolName,
	)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: body}}}
}
