// Command buddy-mcp runs the buddy MCP server over stdio.
// Claude Code connects to it via the "buddy" entry in mcpServers.
//
// Environment:
//   BUDDY_DB                — path to buddy.db (empty = default).
//   BUDDY_ANALYTICS_BACKEND — when set, enables the analytics_query_* tools
//                             against the matching adapter. v0.2.0 ships
//                             "sql" only (SQLite reference); other values
//                             register the tools as stubs.
//   BUDDY_ANALYTICS_DSN     — required when BUDDY_ANALYTICS_BACKEND=sql.
//                             Path to a SQLite file (e.g. ~/.buddy/analytics.db).
//                             Schema is migrated on startup if missing.
package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	_ "modernc.org/sqlite"

	"github.com/0xmhha/buddy/internal/analytics"
	buddymcp "github.com/0xmhha/buddy/internal/mcp"
)

func main() {
	opts := buddymcp.Options{
		DBPath: os.Getenv("BUDDY_DB"),
	}

	if adapter, err := configureAnalytics(); err != nil {
		// Surface the misconfiguration on stderr and exit — silently dropping
		// to the stub path would mask "I thought I configured this" mistakes.
		log.Fatalf("buddy-mcp: analytics: %v", err)
	} else {
		opts.Analytics = adapter
	}

	s := buddymcp.NewBuddyServer(opts)
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("buddy-mcp: %v", err)
	}
}

// configureAnalytics inspects BUDDY_ANALYTICS_BACKEND / BUDDY_ANALYTICS_DSN
// and returns the appropriate adapter. Returns (nil, nil) when the user did
// not opt in — that is the documented stub mode, not an error.
func configureAnalytics() (analytics.Adapter, error) {
	backend := os.Getenv("BUDDY_ANALYTICS_BACKEND")
	if backend == "" {
		return nil, nil
	}
	if backend != "sql" {
		// Recognised but not implemented in v0.2.0 — let the tool registration
		// path fall back to the stub response.
		return nil, nil
	}
	dsn := os.Getenv("BUDDY_ANALYTICS_DSN")
	if dsn == "" {
		log.Println("buddy-mcp: BUDDY_ANALYTICS_BACKEND=sql but BUDDY_ANALYTICS_DSN is empty — analytics tools will return stub responses")
		return nil, nil
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := analytics.Migrate(context.Background(), db); err != nil {
		return nil, err
	}
	return analytics.NewSQLAdapter(db), nil
}
