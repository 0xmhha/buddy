package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/agent"
	"github.com/0xmhha/buddy/internal/db"
)

// newAgentCmd assembles the `buddy agent ...` subtree. Per cli-buddy-spec §6.2
// (locked in by ADR-005), v0.3 ships create / list / show / run / delete. tui
// / edit / log / schedule list land in W3-2 (TUI) and W3-3 follow-ons.
func newAgentCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "agent",
		Short: "Manage cli buddy automation agents",
		Long: "Manage automation agents that drive plugin buddy command chains.\n" +
			"Per cli-buddy-spec §3 + ADR-005, agents = static YAML spec + on-demand or\n" +
			"scheduled Run(). v0.3 ships on-demand only; scheduler arrives in W3-3 follow-on.",
	}
	c.AddCommand(
		newAgentCreateCmd(),
		newAgentListCmd(),
		newAgentShowCmd(),
		newAgentRunCmd(),
		newAgentLogCmd(),
		newAgentDeleteCmd(),
		newAgentSchedulerCmd(),
	)
	return c
}

// openAgentStore is the wire-up shared by every subcommand: open DB → run
// migrations → wrap in agent.Store. Caller closes via t.Cleanup-equivalent
// returned closer.
func openAgentStore(dbFlag string) (*agent.Store, func(), error) {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	return agent.NewStore(conn), func() { _ = conn.Close() }, nil
}

// ─── create ────────────────────────────────────────────────────────────────

func newAgentCreateCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "create <yaml-path>",
		Short: "Register an agent from a YAML spec file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			yamlBytes, err := readFileOrStdin(args[0])
			if err != nil {
				return err
			}
			spec, err := agent.ParseSpec(yamlBytes)
			if err != nil {
				return err
			}
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			created, err := store.Create(ctx, spec, string(yamlBytes))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s 등록 완료 (status=%s)\n", created.ID, created.Status)
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

// readFileOrStdin returns the YAML body. "-" means stdin so users can pipe
// `cat spec.yaml | buddy agent create -`.
func readFileOrStdin(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

// ─── list ──────────────────────────────────────────────────────────────────

func newAgentListCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "list",
		Short: "List all registered agents (most recently updated first)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			agents, err := store.List(cmd.Context())
			if err != nil {
				return err
			}
			if len(agents) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "등록된 agent 가 없어. `buddy agent create <spec.yaml>` 로 시작.")
				return nil
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-24s %-10s %-19s %s\n", "ID", "STATUS", "LAST RUN", "NAME")
			for _, a := range agents {
				last := "-"
				if a.LastRunAt != nil {
					last = a.LastRunAt.Format("2006-01-02 15:04:05")
				}
				fmt.Fprintf(out, "%-24s %-10s %-19s %s\n", a.ID, a.Status, last, a.Name)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

// ─── show ──────────────────────────────────────────────────────────────────

func newAgentShowCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "show <agent-id>",
		Short: "Show one agent's spec + status detail",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			a, err := store.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "ID:         %s\n", a.ID)
			fmt.Fprintf(out, "Name:       %s\n", a.Name)
			fmt.Fprintf(out, "Status:     %s\n", a.Status)
			fmt.Fprintf(out, "Schedule:   %s\n", emptyDash(a.Schedule))
			fmt.Fprintf(out, "Created:    %s\n", a.CreatedAt.Format("2006-01-02 15:04:05 MST"))
			fmt.Fprintf(out, "Updated:    %s\n", a.UpdatedAt.Format("2006-01-02 15:04:05 MST"))
			if a.LastRunAt != nil {
				fmt.Fprintf(out, "Last run:   %s\n", a.LastRunAt.Format("2006-01-02 15:04:05 MST"))
			}
			fmt.Fprintln(out, "Spec YAML:")
			fmt.Fprintln(out, indent(a.SpecYAML, "    "))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

// ─── log ───────────────────────────────────────────────────────────────────

func newAgentLogCmd() *cobra.Command {
	var (
		dbFlag string
		limit  int
	)
	c := &cobra.Command{
		Use:   "log <agent-id>",
		Short: "Tail logs from the agent's most recent run",
		Long: "Reads agent_logs for the latest run of <agent-id>, oldest first.\n" +
			"Surfaces the per-step info / warn / error lines the runtime writes,\n" +
			"including the v0.5.0+ self-check verdict line and the v0.6.0+\n" +
			"next-phase branch lines. Use --limit to cap rows when a run\n" +
			"produced many lines.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			run, err := store.LatestRun(ctx, args[0])
			if err != nil {
				if errors.Is(err, agent.ErrNotFound) {
					fmt.Fprintf(cmd.ErrOrStderr(),
						"agent %q has no runs yet (or does not exist). Try `buddy agent run %s`.\n",
						args[0], args[0])
					return err
				}
				return err
			}
			logs, err := store.Logs(ctx, run.ID, limit)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Run ID:     %d\n", run.ID)
			fmt.Fprintf(out, "Started:    %s\n", run.StartedAt.Format("2006-01-02 15:04:05 MST"))
			if run.EndedAt != nil {
				fmt.Fprintf(out, "Ended:      %s\n", run.EndedAt.Format("2006-01-02 15:04:05 MST"))
			}
			fmt.Fprintf(out, "Exit code:  %d\n", run.ExitCode)
			if run.Error != "" {
				fmt.Fprintf(out, "Error:      %s\n", run.Error)
			}
			fmt.Fprintln(out, "Log:")
			if len(logs) == 0 {
				fmt.Fprintln(out, "    (no log lines)")
				return nil
			}
			for _, l := range logs {
				fmt.Fprintf(out, "    %s  %-5s  %s\n",
					l.Ts.Format("2006-01-02 15:04:05"), l.Level, l.Message)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().IntVar(&limit, "limit", 0, "maximum log lines to print (0 = all)")
	return c
}

// ─── run ───────────────────────────────────────────────────────────────────

func newAgentRunCmd() *cobra.Command {
	var (
		dbFlag       string
		claudeBinary string
	)
	c := &cobra.Command{
		Use:   "run <agent-id>",
		Short: "Execute the agent's chain synchronously and stream the result",
		Long: "Spawns `claude` once per chain step (cli-buddy-spec §4.1 option (a)\n" +
			"lock-in via ADR-005). Output is the JSON-marshalled RunResult.\n" +
			"Errors that come from a missing claude binary include an install hint.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			a, err := store.Get(ctx, args[0])
			if err != nil {
				return err
			}
			exec := agent.NewSubprocessExecutor()
			if claudeBinary != "" {
				exec.ClaudeBinary = claudeBinary
			}
			rt := agent.NewRuntime(store, exec)
			res, runErr := rt.Run(ctx, a)
			if runErr != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "agent run encountered an error:", runErr)
			}
			// Always emit the captured result so users see partial progress on failure.
			if err := writeRunResultJSON(cmd.OutOrStdout(), res); err != nil {
				return err
			}
			if runErr != nil {
				return runErr
			}
			if res.ExitCode != 0 {
				return fmt.Errorf("agent %s finished with non-zero exit code %d", a.ID, res.ExitCode)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().StringVar(&claudeBinary, "claude-binary", "", "override claude CLI path (default: 'claude' on PATH)")
	return c
}

// writeRunResultJSON emits the same shape as agent_runs.result_json so the
// user can compare what the CLI printed with what was persisted.
func writeRunResultJSON(w io.Writer, res agent.RunResult) error {
	return writeJSONIndented(w, res)
}

// ─── scheduler ─────────────────────────────────────────────────────────────

func newAgentSchedulerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scheduler",
		Short: "Background scheduler for agents with a non-empty schedule field",
		Long: "Loads every agent whose spec.schedule is a valid cron expression and\n" +
			"ticks them in the foreground until interrupted. Use `buddy agent scheduler\n" +
			"start` to run; `status` to one-shot the current entry list.\n\n" +
			"v0.3 caveats:\n" +
			"  - sub-second cron (@every 100ms) does not work — robfig/cron rounds to\n" +
			"    whole seconds. Use @every 1m or finer-grained CLI tools.\n" +
			"  - the scheduler does not auto-reload when agents are added/removed —\n" +
			"    restart to pick up changes.",
	}
	c.AddCommand(newAgentSchedulerStartCmd(), newAgentSchedulerStatusCmd())
	return c
}

func newAgentSchedulerStartCmd() *cobra.Command {
	var (
		dbFlag       string
		claudeBinary string
	)
	c := &cobra.Command{
		Use:   "start",
		Short: "Run the scheduler in the foreground (blocks until Ctrl-C)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			exec := agent.NewSubprocessExecutor()
			if claudeBinary != "" {
				exec.ClaudeBinary = claudeBinary
			}
			rt := agent.NewRuntime(store, exec)
			sched := agent.NewScheduler(store, rt, agent.SchedulerOptions{Logger: cmd.ErrOrStderr()})

			loaded, skipped, err := sched.Load(ctx)
			if err != nil {
				return err
			}
			for _, s := range skipped {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"scheduler: skipping agent %q — invalid cron %q: %v\n",
					s.ID, s.Schedule, s.Err)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"scheduler: loaded %d agent(s) (%d skipped). Ctrl-C to stop.\n",
				loaded, len(skipped))

			// cobra contexts already honour SIGINT/SIGTERM when set up via
			// signal.NotifyContext in main(); the scheduler's Start blocks
			// until ctx is cancelled, then gracefully drains in-flight jobs.
			return sched.Start(ctx)
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().StringVar(&claudeBinary, "claude-binary", "", "override claude CLI path (default: 'claude' on PATH)")
	return c
}

func newAgentSchedulerStatusCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "status",
		Short: "Show scheduled agents and their next planned tick (one-shot)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			rt := agent.NewRuntime(store, agent.NewSubprocessExecutor())
			sched := agent.NewScheduler(store, rt, agent.SchedulerOptions{})
			loaded, skipped, err := sched.Load(ctx)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Loaded:   %d\n", loaded)
			fmt.Fprintf(out, "Skipped:  %d\n", len(skipped))
			for _, s := range skipped {
				fmt.Fprintf(out, "  - %s (%q): %v\n", s.ID, s.Schedule, s.Err)
			}
			if loaded > 0 {
				fmt.Fprintln(out, "Entries:")
				now := time.Now()
				for _, e := range sched.Entries() {
					// cron library populates entry.Next only after Start();
					// status is a one-shot, so compute the next fire time
					// manually from the schedule.
					next := e.Schedule.Next(now)
					fmt.Fprintf(out, "  - entry_id=%d next=%s\n", e.ID, next.Format("2006-01-02 15:04:05 MST"))
				}
			}
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

// ─── delete ────────────────────────────────────────────────────────────────

func newAgentDeleteCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "delete <agent-id>",
		Short: "Remove an agent and cascade-delete its runs + logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			if err := store.Delete(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s 삭제 완료\n", args[0])
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

// ─── small helpers ────────────────────────────────────────────────────────

func emptyDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func indent(s, pad string) string {
	if s == "" {
		return pad + "(empty)"
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		b.WriteString(pad)
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

