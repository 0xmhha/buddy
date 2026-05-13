package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/tui"
)

// newTuiCmd wires `buddy tui` — the W3-2 minimum-viable terminal UI. v0.6.6
// ships a read-only agent list view (j/k or arrow-key navigation, q to
// quit, r to refresh). The detail / create / live-log views remain
// follow-on per cli-buddy-spec §9 W3-2.
//
// AltScreen is enabled so the previous shell content is preserved and
// restored on quit — the friend-tone "silent default" stays intact.
func newTuiCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "tui",
		Short: "Open the cli buddy terminal UI (agent list, navigation, refresh)",
		Long: "Launches the bubbletea-based TUI. v0.6.6 ships a read-only\n" +
			"agent list; create/edit/log views land in follow-on cycles.\n\n" +
			"Keys: j/k or ↑/↓ to move, g/G to jump to top/bottom, r to\n" +
			"refresh, q (or Ctrl-C) to quit. AltScreen is used so your\n" +
			"prior shell content is restored on exit.",
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
