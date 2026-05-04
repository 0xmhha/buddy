package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/wm-it-22-00661/buddy/internal/diagnose"
)

type doctorArgs struct{}

type doctorResult struct {
	Healthy bool     `json:"healthy"`
	Issues  []string `json:"issues"`
}

func addDoctorTool(s *mcp.Server, opts Options) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "doctor",
		Description: "Run buddy health check. Returns healthy=true when DB, daemon, and hook metrics are all green. When healthy=false, issues lists the problems found.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ doctorArgs) (*mcp.CallToolResult, doctorResult, error) {
		rep, err := diagnose.Check(diagnose.Options{DBPath: opts.DBPath})
		if err != nil {
			return nil, doctorResult{}, err
		}

		out := doctorResult{Healthy: rep.Healthy}
		for _, d := range rep.Issues {
			out.Issues = append(out.Issues, fmt.Sprintf("[%s] %s", d.Kind, d.Message))
		}

		var sb strings.Builder
		if rep.Healthy {
			sb.WriteString("All checks passed.")
		} else {
			sb.WriteString(fmt.Sprintf("%d issue(s) found:\n", len(rep.Issues)))
			for _, line := range out.Issues {
				sb.WriteString("  • ")
				sb.WriteString(line)
				sb.WriteByte('\n')
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: sb.String()},
			},
		}, out, nil
	})
}
