package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/tui"
)

// newTuiCmd wires `buddy tui` — the W3-2 terminal UI. v0.6.6 shipped the
// minimum-viable list. Follow-on cycles add: detail pane (Enter/l opens
// the latest-run summary), scheduler-preview pane (s shows when each
// scheduled agent would fire next). Create form and live log tail remain
// follow-on per cli-buddy-spec §9 W3-2.
//
// AltScreen is enabled so the previous shell content is preserved and
// restored on quit — the friend-tone "silent default" stays intact.
func newTuiCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "tui",
		Short: "Open the cli buddy terminal UI (agent list, detail, scheduler preview)",
		Long: "Launches the bubbletea-based TUI. List view: j/k or ↑/↓ to move,\n" +
			"g/G to jump to top/bottom, r to refresh, enter/l to open detail,\n" +
			"s to open the scheduler-preview pane, q (or Ctrl-C) to quit.\n" +
			"Detail / scheduler views: esc/h to return, r to refetch.\n" +
			"The scheduler pane is preview-only (does not run jobs) — it\n" +
			"shows when each scheduled agent would fire next based on its\n" +
			"cron expression. AltScreen is used so your prior shell content\n" +
			"is restored on exit.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, closer, err := openAgentStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			program := tea.NewProgram(
				tui.NewModel(store),
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
