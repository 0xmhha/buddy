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
	"github.com/0xmhha/buddy/internal/db"
	buddymcp "github.com/0xmhha/buddy/internal/mcp"
	"github.com/0xmhha/buddy/internal/usage"
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

	// Wire the F2.B Usage service against the same buddy.db. The MCP
	// server is read-only here — usage_query_* tools live or die with
	// the sessions table existing in this DB. If Open fails (no DB
	// yet) the tools register and return the friend-tone "not wired"
	// hint instead of failing process startup.
	if svc, err := configureUsage(opts.DBPath); err != nil {
		log.Printf("buddy-mcp: usage tools disabled: %v", err)
	} else {
		opts.Usage = svc
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

// configureUsage opens buddy.db (sessions table substrate) and returns
// a usage.Service. Per ADR-013 the service is stateless and read-only,
// so opening the same DB the rest of the CLI uses is safe.
func configureUsage(dbPath string) (*usage.Service, error) {
	conn, err := db.Open(db.Options{Path: dbPath})
	if err != nil {
		return nil, err
	}
	return usage.NewService(conn), nil
}
