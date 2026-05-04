// Command buddy-mcp runs the buddy MCP server over stdio.
// Claude Code connects to it via the "buddy" entry in mcpServers.
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	buddymcp "github.com/wm-it-22-00661/buddy/internal/mcp"
)

func main() {
	opts := buddymcp.Options{
		DBPath: os.Getenv("BUDDY_DB"),
	}
	s := buddymcp.NewBuddyServer(opts)
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("buddy-mcp: %v", err)
	}
}
