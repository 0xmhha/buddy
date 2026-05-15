package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/tui"
)

// newTuiCmd wires `buddy tui` — the W3-2 terminal UI. v0.6.6 shipped the
// minimum-viable list. The current cycle adds the detail pane (Enter or
// l on a row → latest-run summary; Esc or h returns). Create form, live
// log tail, and scheduler status pane remain follow-on per
// cli-buddy-spec §9 W3-2.
//
// AltScreen is enabled so the previous shell content is preserved and
// restored on quit — the friend-tone "silent default" stays intact.
func newTuiCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "tui",
		Short: "Open the cli buddy terminal UI (agent list + detail view)",
		Long: "Launches the bubbletea-based TUI. List view: j/k or ↑/↓ to move,\n" +
			"g/G to jump to top/bottom, r to refresh, enter/l to open detail,\n" +
			"q (or Ctrl-C) to quit. Detail view: esc/h to return, r to\n" +
			"refresh the latest-run summary. AltScreen is used so your prior\n" +
			"shell content is restored on exit.",
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
