package mcp

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/0xmhha/buddy/internal/advisor"
)

// advise_tool.go wires the `usage_advise` MCP tool. Pulls live usage
// metric + retrieval, returns the structured Advisory[] payload. The
// notification layer consumes this for desktop / webhook / banner /
// shell dispatch.

type adviseArgs struct {
	Persist bool   `json:"persist,omitempty" jsonschema:"When true, write fresh advisories to the advisories table (subject to dedup window). Default: false."`
	Since   string `json:"since,omitempty"   jsonschema:"With --history, lookback window in Go duration syntax (e.g. 24h, 168h). Empty = all-time."`
	History bool   `json:"history,omitempty" jsonschema:"When true, return persisted advisories instead of running rules. Default: false."`
}

type adviseResult struct {
	Advisories []advisor.Advisory `json:"advisories"`
}

// AdvisorOptions configures the usage_advise tool. The two fields
// mirror the substrate split: Runner is the rule path (preview /
// persist); Store is the read-back path (history).
type AdvisorOptions struct {
	Runner *advisor.Evaluator
	Store  *advisor.Store
}

func addAdvisorTool(s *mcp.Server, opts Options) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "usage_advise",
		Description: "Generate friend-tone Korean advisories from local usage + knowledge data. With history=true, returns persisted advisories from the advisories table instead of running rules. Each advisory has Kind / Severity / Message / Evidence[]. The 5 built-in rule kinds: token-spike-day, long-session, low-cache-ratio, session-volume-day, token-daily-cap. Threshold tuning lives in `buddy config` (advisor.* keys).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args adviseArgs) (*mcp.CallToolResult, adviseResult, error) {
		if opts.Advisor.Runner == nil {
			return advisorNotWired(), adviseResult{}, nil
		}
		if args.History {
			if opts.Advisor.Store == nil {
				return advisorNotWired(), adviseResult{}, nil
			}
			lo := advisor.ListOptions{IncludeMuted: true}
			if args.Since != "" {
				if d, err := time.ParseDuration(args.Since); err == nil {
					lo.Since = time.Now().UTC().Add(-d)
				}
			}
			rows, err := opts.Advisor.Store.List(ctx, lo)
			if err != nil {
				return nil, adviseResult{}, err
			}
			res := adviseResult{Advisories: rows}
			return jsonContent(res), res, nil
		}

		var advs []advisor.Advisory
		var err error
		if args.Persist {
			advs, err = opts.Advisor.Runner.Persist(ctx)
		} else {
			advs, err = opts.Advisor.Runner.Run(ctx)
		}
		if err != nil {
			return nil, adviseResult{}, err
		}
		res := adviseResult{Advisories: advs}
		return jsonContent(res), res, nil
	})
}

func advisorNotWired() *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{
		Text: "advisor-mcp: advisor.Evaluator 가 연결 안 돼 있어. " +
			"`buddy mcp serve` 가 sessions + chunks + advisories 테이블이 있는 " +
			"buddy.db 를 가리키는지 확인해줘.",
	}}}
}
