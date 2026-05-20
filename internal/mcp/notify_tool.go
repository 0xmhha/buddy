package mcp

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/notify"
)

// notify_tool.go wires the W7-5 / ADR-016 notify MCP surface. Two
// tools:
//
//   - notify_status : read notification_log (visibility + dedup debug)
//   - notify_test   : dispatch a synthetic notification through one
//                     channel (smoke test before going live)
//
// Dispatch through the daemon is the production path; these tools are
// for inspection and provisioning.

type notifyStatusArgs struct {
	Channel string `json:"channel,omitempty" jsonschema:"Filter to a single channel: desktop | webhook | tui-banner | shell."`
	Since   string `json:"since,omitempty"   jsonschema:"Lookback duration (Go syntax e.g. 24h). Empty = all time."`
	Limit   int    `json:"limit,omitempty"   jsonschema:"Max rows. Default 50."`
}

type notifyStatusResult struct {
	Rows []notify.LogRow `json:"rows"`
}

type notifyTestArgs struct {
	Channel string `json:"channel" jsonschema:"Channel to fire the test through: desktop | webhook | tui-banner | shell."`
}

type notifyTestResult struct {
	Sent    bool   `json:"sent"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

// NotifyOptions wires the notify MCP tools. Store holds the log;
// Dispatcher carries the configured channels for `notify_test`.
type NotifyOptions struct {
	Store      *notify.Store
	Dispatcher *notify.Dispatcher
}

func addNotifyTool(s *mcp.Server, opts Options) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "notify_status",
		Description: "Read recent notification_log rows. Useful for debugging 'why didn't I get pinged' or auditing the dispatcher's per-channel dedup decisions. Returns rows sorted by sent_at DESC.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args notifyStatusArgs) (*mcp.CallToolResult, notifyStatusResult, error) {
		if opts.Notify.Store == nil {
			return notifyNotWired(), notifyStatusResult{}, nil
		}
		lo := notify.ListOptions{Channel: args.Channel, Limit: args.Limit}
		if lo.Limit <= 0 {
			lo.Limit = 50
		}
		if args.Since != "" {
			if d, err := time.ParseDuration(args.Since); err == nil {
				lo.Since = time.Now().UTC().Add(-d)
			}
		}
		rows, err := opts.Notify.Store.List(ctx, lo)
		if err != nil {
			return nil, notifyStatusResult{}, err
		}
		res := notifyStatusResult{Rows: rows}
		return jsonContent(res), res, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "notify_test",
		Description: "Dispatch a synthetic notification through one channel (smoke test). The fake advisory has Severity=warn so per-channel severity_min floors of info|warn allow it through; channels at high block it. Result.sent=false means dispatcher filtered it (check notify_status for the outcome row).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args notifyTestArgs) (*mcp.CallToolResult, notifyTestResult, error) {
		if opts.Notify.Dispatcher == nil {
			return notifyNotWired(), notifyTestResult{Sent: false, Channel: args.Channel,
				Message: "notify dispatcher not wired"}, nil
		}
		fake := advisor.Advisory{
			ID: -1, Kind: "buddy-test", Severity: advisor.SeverityWarn,
			Message: "테스트 알림 — MCP 채널 동작 확인용", CreatedAt: time.Now().UTC(),
		}
		sent := opts.Notify.Dispatcher.Dispatch(ctx, []advisor.Advisory{fake})
		res := notifyTestResult{
			Sent:    sent[args.Channel] > 0,
			Channel: args.Channel,
			Message: "테스트 dispatch 완료. notify_status 로 결과 확인.",
		}
		return jsonContent(res), res, nil
	})
}

func notifyNotWired() *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{
		Text: "notify-mcp: notify.Store / Dispatcher 가 연결 안 돼 있어. " +
			"buddy.db 와 config 가 노출된 환경에서 `buddy mcp serve` 실행 중인지 확인해줘.",
	}}}
}
