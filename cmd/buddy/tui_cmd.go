package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/knowledge"
	"github.com/0xmhha/buddy/internal/queries"
	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/tui"
	"github.com/0xmhha/buddy/internal/usage"
)

// newTuiCmd wires `buddy tui` — the W3-2 terminal UI. v0.6.6 shipped the
// minimum-viable list. Follow-on cycles add: detail pane (Enter/l opens
// the latest-run summary), scheduler-preview pane (s shows when each
// scheduled agent would fire next), in-app delete with confirmation
// (d → y/N prompt → FK-cascading delete), live log-tail pane (t from
// detail view → ~1s polling of `agent_logs` for the run shown), in-app
// spec edit (e from detail view → $EDITOR shell-out → ParseSpec
// validation → Store.UpdateSpec), create form (c from list view →
// $EDITOR shell-out on a starter YAML → Store.Create), and a hook-
// reliability stats pane (H from list view → queries.Run snapshot of
// the v0.1.0 daemon/aggregator output, A-3.2 W3-5 follow-on integrating
// the hook monitor as a cli buddy sub-feature). All 6 named W3-2
// follow-on items per cli-buddy-spec §9 are now shipped.
//
// AltScreen is enabled so the previous shell content is preserved and
// restored on quit — the friend-tone "silent default" stays intact.
// During an `e` edit or `c` create the AltScreen is briefly suspended
// so $EDITOR can take over the terminal; on exit the TUI redraws.
func newTuiCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "tui",
		Short: "Open the cli buddy terminal UI (list, detail, scheduler, delete, log tail, edit, create, hook stats)",
		Long: "Launches the bubbletea-based TUI. List view: j/k or ↑/↓ to move,\n" +
			"g/G to jump to top/bottom, r to refresh, enter/l to open detail,\n" +
			"s to open the scheduler-preview pane, c to create a new agent\n" +
			"(opens $EDITOR on a starter YAML template), d to delete the\n" +
			"cursor row (y/N confirm required — also drops the agent's runs\n" +
			"+ logs via FK cascade), H to open the hook-reliability stats\n" +
			"pane (count / failures / p50 / p95 per hook for the last hour),\n" +
			"q (or Ctrl-C) to quit.\n" +
			"Detail view: esc/h to return, r to refetch the latest run,\n" +
			"t to tail that run's logs (~1s auto-refresh; esc/h back),\n" +
			"e to edit the agent spec via $EDITOR (fall back to $VISUAL,\n" +
			"then vi). Renames are rejected — the spec id must match.\n" +
			"Scheduler view: esc/h to return, r to refetch. The scheduler\n" +
			"pane is preview-only (does not run jobs) — it shows when each\n" +
			"scheduled agent would fire next based on its cron expression.\n" +
			"Hook stats view: esc/h to return, r to refetch.\n" +
			"AltScreen is used so your prior shell content is restored on exit.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			model := tui.NewModel(store)
			// Hook-stats fetcher: closes over dbFlag and delegates to
			// queries.Run, which opens its own read-only connection.
			// Wired here (rather than inside tui.NewModel) so the model
			// stays pure-Go and decoupled from the queries package.
			model.HookStatsFetcher = func(window string) (queries.Result, error) {
				return queries.Run(queries.Options{
					DBPath: dbFlag,
					Window: window,
				})
			}

			// Usage fetcher (W7-2 / ADR-013). Opens its own connection
			// per call; closes immediately so a stale TUI doesn't leak.
			// Best-effort: failures yield UsageErrMsg via the loader.
			model.UsageFetcher = func() (usage.Overview, error) {
				conn, err := db.Open(db.Options{Path: dbFlag})
				if err != nil {
					return usage.Overview{}, fmt.Errorf("open db: %w", err)
				}
				defer conn.Close()
				return usage.NewService(conn).QueryOverview(context.Background(), usage.TimeWindow{}, 5)
			}

			// Advisor fetcher (W7-3b / ADR-015). Same fresh-conn pattern.
			// Run() (not Persist) so the TUI never silently writes —
			// users explicitly opt in via `buddy advise --persist` or
			// the daemon's advisorMonitor.
			model.AdvisorFetcher = func() ([]advisor.Advisory, error) {
				conn, err := db.Open(db.Options{Path: dbFlag})
				if err != nil {
					return nil, fmt.Errorf("open db: %w", err)
				}
				defer conn.Close()
				runner := &advisor.Evaluator{
					Thresholds: advisor.DefaultThresholds(),
					Usage:      usage.NewService(conn),
					Sessions:   sessions.NewStore(conn),
					Knowledge:  knowledge.NewStore(conn),
					Embedder:   knowledge.NewPythonEmbedder(),
					Advisories: advisor.NewStore(conn),
				}
				return runner.Run(context.Background())
			}

			program := tea.NewProgram(
				model,
				tea.WithAltScreen(),
				tea.WithContext(cmd.Context()),
			)
			if _, err := program.Run(); err != nil {
				return fmt.Errorf("tui: %w", err)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}
