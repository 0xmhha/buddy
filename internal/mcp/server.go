// Package mcp exposes buddy's internal capabilities as an MCP server.
// Tools are registered once in NewBuddyServer and served over stdio.
package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/0xmhha/buddy/internal/analytics"
)

// NewBuddyServer creates an MCP server with all buddy tools registered.
// Callers connect it to a transport (e.g. mcp.StdioTransport{}) via Run.
func NewBuddyServer(opts Options) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "buddy",
		Version: "0.6.3",
	}, &mcp.ServerOptions{
		Instructions: "buddy — Claude Code hook harness control plane + analytics surface. " +
			"Use these tools to inspect hook health, query hook statistics, manage the " +
			"local feature registry, and (analytics_query_*) read funnel / cohort / A-B / " +
			"cost / SLO / feedback data from production backends. Analytics tools require " +
			"BUDDY_ANALYTICS_BACKEND to be set; v0.2.0 ships stubs and reports the missing " +
			"adapter in friend-tone text rather than a transport error.",
	})

	addDoctorTool(s, opts)
	addStatsTool(s, opts)
	addFeatureTools(s, opts)
	addAnalyticsTools(s, opts)

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
}
