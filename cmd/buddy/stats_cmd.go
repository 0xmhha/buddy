package main

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/persona"
	"github.com/0xmhha/buddy/internal/queries"
)

// newStatsCmd wires the read-only hook_stats report. Output goes to stdout
// (it's the user-facing report, not log noise). Exit is always 0 unless we
// fail to open the DB or the user passes a bad --window — both surface as
// friendError values that main() prints and exits 1 on. See m4-plan §Task 3.
func newStatsCmd() *cobra.Command {
	var (
		dbFlag     string
		windowFlag string
		byToolFlag bool
		hookFlag   string
	)
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "최근 hook 통계 (read-only, daemon 의존 없음)",
		RunE: func(_ *cobra.Command, _ []string) error {
			res, err := queries.Run(queries.Options{
				DBPath:     dbFlag,
				Window:     windowFlag,
				ByTool:     byToolFlag,
				HookFilter: hookFlag,
			})
			if err != nil {
				if errors.Is(err, queries.ErrInvalidWindow) {
					return newFriendError("buddy: " + err.Error())
				}
				if errors.Is(err, db.ErrDBMissing) {
					return dbMissingFriendError(dbFlag)
				}
				// Any other failure is a DB-side problem (open or query). Match
				// doctor's wording so the two read-only commands feel consistent.
				return newFriendError(persona.M(persona.KeyDBReadFailed, err))
			}
			res.Render(os.Stdout)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&windowFlag, "window", "1h", "집계 윈도우: 5m | 1h | 24h")
	cmd.Flags().BoolVar(&byToolFlag, "by-tool", false, "tool 단위로 분리해서 보기")
	cmd.Flags().StringVar(&hookFlag, "hook", "", "특정 hook 이름으로 필터")
	return cmd
}
