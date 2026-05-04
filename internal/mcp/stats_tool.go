package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/queries"
)

type statsArgs struct {
	Window string `json:"window" jsonschema:"Aggregation window: 5m, 1h, or 24h. Default: 1h"`
	Hook   string `json:"hook"   jsonschema:"Filter by hook name (case-insensitive). Optional."`
}

type statsResult struct {
	Window string      `json:"window"`
	Rows   []statsRow  `json:"rows"`
}

type statsRow struct {
	HookName   string `json:"hook_name"`
	ToolName   string `json:"tool_name,omitempty"`
	Count      int64  `json:"count"`
	Failures   int64  `json:"failures"`
	FailRatePct int   `json:"fail_rate_pct"`
	P50Ms      int64  `json:"p50_ms"`
	P95Ms      int64  `json:"p95_ms"`
}

func addStatsTool(s *mcp.Server, opts Options) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "stats",
		Description: "Query hook execution statistics from the local buddy DB. Returns counts, failure rates, and latency percentiles per hook (and optionally per tool).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args statsArgs) (*mcp.CallToolResult, statsResult, error) {
		window := args.Window
		if window == "" {
			window = "1h"
		}
		res, err := queries.Run(queries.Options{
			DBPath:     opts.DBPath,
			Window:     window,
			HookFilter: args.Hook,
		})
		if err != nil {
			if errors.Is(err, db.ErrDBMissing) {
				return nil, statsResult{}, fmt.Errorf("buddy DB not found — run 'buddy install' first")
			}
			return nil, statsResult{}, err
		}

		out := statsResult{Window: window}
		for _, r := range res.Rows {
			var failRate int
			if r.Count > 0 {
				failRate = int(r.Failures * 100 / r.Count)
			}
			out.Rows = append(out.Rows, statsRow{
				HookName:    r.HookName,
				ToolName:    r.ToolName,
				Count:       r.Count,
				Failures:    r.Failures,
				FailRatePct: failRate,
				P50Ms:       r.P50Ms,
				P95Ms:       r.P95Ms,
			})
		}

		var sb strings.Builder
		if len(out.Rows) == 0 {
			sb.WriteString("No stats yet for window " + window + ".")
		} else {
			sb.WriteString(fmt.Sprintf("Hook stats (window=%s):\n", window))
			for _, r := range out.Rows {
				tool := r.ToolName
				if tool == "" {
					tool = "-"
				}
				sb.WriteString(fmt.Sprintf("  %-20s %-12s count=%-6d fail=%d%% p50=%dms p95=%dms\n",
					r.HookName, tool, r.Count, r.FailRatePct, r.P50Ms, r.P95Ms))
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, out, nil
	})
}
