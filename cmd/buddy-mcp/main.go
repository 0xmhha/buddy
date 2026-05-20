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

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/analytics"
	"github.com/0xmhha/buddy/internal/config"
	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/knowledge"
	buddymcp "github.com/0xmhha/buddy/internal/mcp"
	"github.com/0xmhha/buddy/internal/notify"
	"github.com/0xmhha/buddy/internal/sessions"
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

	// Wire the F2.C Phase 1 knowledge store + embedder (W7-3a /
	// ADR-014). Store opens against the same buddy.db; embedder is
	// always set so the MCP path tries vector first, falling back to
	// BM25 only when the Python venv is missing — handled inside
	// knowledge.PythonEmbedder.Embed via ErrEmbedderUnavailable.
	if kopt, err := configureKnowledge(opts.DBPath); err != nil {
		log.Printf("buddy-mcp: knowledge tools disabled: %v", err)
	} else {
		opts.Knowledge = kopt
	}

	// Wire the F2.C Phase 2 advisor (W7-3b / ADR-015). Builds an
	// Evaluator over the same buddy.db connection sources. Falls back
	// silently when DB open fails — usage_advise will report "not
	// wired" rather than crashing the MCP server boot.
	if aopt, err := configureAdvisor(opts.DBPath); err != nil {
		log.Printf("buddy-mcp: advisor tool disabled: %v", err)
	} else {
		opts.Advisor = aopt
	}

	// Wire the W7-5 notify tools. Same defensive pattern: failure on
	// DB open or config load disables only the notify path.
	if nopt, err := configureNotify(opts.DBPath); err != nil {
		log.Printf("buddy-mcp: notify tools disabled: %v", err)
	} else {
		opts.Notify = nopt
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

// configureKnowledge opens buddy.db and returns the W7-3a Knowledge
// options bundle. Always wires the Python embedder — its Embed method
// returns ErrEmbedderUnavailable if the script / venv isn't ready, and
// the knowledge_query handler degrades to BM25-only on that error.
func configureKnowledge(dbPath string) (buddymcp.KnowledgeOptions, error) {
	conn, err := db.Open(db.Options{Path: dbPath})
	if err != nil {
		return buddymcp.KnowledgeOptions{}, err
	}
	return buddymcp.KnowledgeOptions{
		Store:    knowledge.NewStore(conn),
		Embedder: knowledge.NewPythonEmbedder(),
	}, nil
}

// configureAdvisor opens buddy.db, loads the user config, and wires a
// fully-stocked Evaluator. Same defensive shape as configureKnowledge.
func configureAdvisor(dbPath string) (buddymcp.AdvisorOptions, error) {
	conn, err := db.Open(db.Options{Path: dbPath})
	if err != nil {
		return buddymcp.AdvisorOptions{}, err
	}
	t := advisor.DefaultThresholds()
	if cfgPath, err := config.DefaultPath(); err == nil {
		if cfg, err := config.Load(cfgPath); err == nil {
			eff := cfg.Effective()
			t = advisor.Thresholds{
				Disabled:            eff.AdvisorDisabled,
				TokenSpikeRatio:     eff.AdvisorTokenSpikeRatio,
				LongSessionHours:    eff.AdvisorLongSessionHours,
				LowCachePct:         eff.AdvisorLowCachePct,
				SessionVolumePerDay: eff.AdvisorSessionVolumePerDay,
				TokenDailyThreshold: eff.AdvisorTokenDailyThreshold,
				DedupWindow:         eff.AdvisorDedupWindow,
				PollInterval:        eff.AdvisorPollInterval,
			}.WithDefaults()
		}
	}
	store := advisor.NewStore(conn)
	runner := &advisor.Evaluator{
		Thresholds: t,
		Usage:      usage.NewService(conn),
		Sessions:   sessions.NewStore(conn),
		Knowledge:  knowledge.NewStore(conn),
		Embedder:   knowledge.NewPythonEmbedder(),
		Advisories: store,
	}
	return buddymcp.AdvisorOptions{Runner: runner, Store: store}, nil
}

// configureNotify opens buddy.db + loads config + builds a Dispatcher
// pre-populated with every enabled channel. Reuses the CLI's
// wireNotifyChannels indirectly by replicating the field projection
// inline (the cmd/buddy package isn't importable from cmd/buddy-mcp).
func configureNotify(dbPath string) (buddymcp.NotifyOptions, error) {
	conn, err := db.Open(db.Options{Path: dbPath})
	if err != nil {
		return buddymcp.NotifyOptions{}, err
	}
	store := notify.NewStore(conn)
	disp := notify.NewDispatcher(store)

	eff := config.Defaults()
	if cfgPath, err := config.DefaultPath(); err == nil {
		if cfg, err := config.Load(cfgPath); err == nil {
			eff = cfg.Effective()
		}
	}
	if eff.NotifyDesktopEnabled {
		disp.AddChannel(notify.NewDesktopChannel(), notify.ChannelConfig{
			Enabled: true, SeverityMin: notify.Severity(eff.NotifyDesktopSeverityMin),
			DedupWindow: eff.NotifyDesktopDedup,
		})
	}
	if eff.NotifyTUIBannerEnabled {
		disp.AddChannel(notify.NewTUIBannerChannel(), notify.ChannelConfig{
			Enabled: true, SeverityMin: notify.Severity(eff.NotifyTUIBannerSeverityMin),
		})
	}
	if eff.NotifyShellPromptEnabled {
		disp.AddChannel(notify.NewShellPromptChannel(), notify.ChannelConfig{
			Enabled: true, SeverityMin: notify.Severity(eff.NotifyShellPromptSeverityMin),
		})
	}
	for _, w := range eff.NotifyWebhooks {
		disp.AddChannel(notify.NewWebhookChannel(notify.WebhookConfig{
			URL: w.URL, Method: w.Method, Headers: w.Headers, Timeout: w.Timeout,
		}), notify.ChannelConfig{
			Enabled: true, SeverityMin: notify.Severity(w.SeverityMin),
			DedupWindow: w.DedupWindow,
		})
	}
	return buddymcp.NotifyOptions{Store: store, Dispatcher: disp}, nil
}
