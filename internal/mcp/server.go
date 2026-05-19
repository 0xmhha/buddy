// Package mcp exposes buddy's internal capabilities as an MCP server.
// Tools are registered once in NewBuddyServer and served over stdio.
package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/0xmhha/buddy/internal/analytics"
	"github.com/0xmhha/buddy/internal/usage"
)

// NewBuddyServer creates an MCP server with all buddy tools registered.
// Callers connect it to a transport (e.g. mcp.StdioTransport{}) via Run.
func NewBuddyServer(opts Options) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "buddy",
		Version: "0.9.0",
	}, &mcp.ServerOptions{
		Instructions: "buddy — Claude Code hook harness control plane + analytics + AI-usage coaching surface. " +
			"Use these tools to inspect hook health, query hook statistics, manage the " +
			"local feature registry, read product analytics (analytics_query_* — backed by " +
			"BUDDY_ANALYTICS_BACKEND), and query AI-usage metrics over the local sessions " +
			"table (usage_query_* — populated by F2.A Session Monitor / ADR-012). " +
			"usage_query_* tools require Options.Usage to be wired (server callers do this " +
			"automatically when sessions table exists).",
	})

	addDoctorTool(s, opts)
	addStatsTool(s, opts)
	addFeatureTools(s, opts)
	addAnalyticsTools(s, opts)
	addUsageTools(s, opts)

	return s
}

// Options configure the buddy MCP server tools.
type Options struct {
	// DBPath is the path to buddy.db. Empty means the default (~/.buddy/buddy.db).
	DBPath string

	// Analytics is the adapter backing the analytics_query_* tools. When nil,
	// the tools register but each handler returns a friend-tone "backend not
	// configured" text body instead of querying real data.
	Analytics analytics.Adapter

	// Usage is the service backing the usage_query_* tools (W7-2 / ADR-013).
	// When nil, those tools register but report "sessions store not wired"
	// instead of querying. The CLI wires this from the same buddy.db it
	// already opens for the agent / feature stores.
	Usage *usage.Service
}
