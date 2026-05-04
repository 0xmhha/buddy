// Package mcp exposes buddy's internal capabilities as an MCP server.
// Tools are registered once in NewBuddyServer and served over stdio.
package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewBuddyServer creates an MCP server with all buddy tools registered.
// Callers connect it to a transport (e.g. mcp.StdioTransport{}) via Run.
func NewBuddyServer(opts Options) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "buddy",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		Instructions: "buddy — Claude Code hook harness control plane. " +
			"Use these tools to inspect health, query hook statistics, " +
			"and manage the local feature registry.",
	})

	addDoctorTool(s, opts)
	addStatsTool(s, opts)
	addFeatureTools(s, opts)

	return s
}

// Options configure the buddy MCP server tools.
type Options struct {
	// DBPath is the path to buddy.db. Empty means the default (~/.buddy/buddy.db).
	DBPath string
}
