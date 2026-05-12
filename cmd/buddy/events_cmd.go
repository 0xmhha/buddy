package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/persona"
	"github.com/0xmhha/buddy/internal/queries"
)

// newEventsCmd wires the read-only hook_events tail. Output is structured
// (one line per event) on stdout — debug surface, not friend tone. With
// --follow the command installs a signal-aware context so Ctrl-C / SIGTERM
// cleanly stops the polling loop and the command writes friend-tone start /
// end markers to stderr (the only friendly touch on this command). See
// m4-plan §Task 4.
func newEventsCmd() *cobra.Command {
	var (
		dbFlag     string
		hookFlag   string
		limitFlag  int
		followFlag bool
	)
	cmd := &cobra.Command{
		Use:   "events",
		Short: "최근 hook_events tail (read-only, --follow 으로 실시간)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := queries.EventsOptions{
				DBPath:     dbFlag,
				HookFilter: hookFlag,
				Limit:      limitFlag,
			}
			if followFlag {
				ctx, stop := signal.NotifyContext(cmd.Context(),
					syscall.SIGINT, syscall.SIGTERM)
				defer stop()
				if err := queries.Follow(ctx, opts, os.Stdout); err != nil {
					if errors.Is(err, queries.ErrInvalidLimit) {
						return newFriendError("buddy: " + err.Error())
					}
					if errors.Is(err, db.ErrDBMissing) {
						return dbMissingFriendError(dbFlag)
					}
					return newFriendError(persona.M(persona.KeyEventsFollowFailed, err))
				}
				return nil
			}
			res, err := queries.RunEvents(opts)
			if err != nil {
				if errors.Is(err, queries.ErrInvalidLimit) {
					return newFriendError("buddy: " + err.Error())
				}
				if errors.Is(err, db.ErrDBMissing) {
					return dbMissingFriendError(dbFlag)
				}
				return newFriendError(persona.M(persona.KeyDBReadFailed, err))
			}
			res.RenderLines(os.Stdout)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&hookFlag, "hook", "", "특정 hook 이름으로 필터 (대소문자 무시)")
	cmd.Flags().IntVar(&limitFlag, "limit", 20, "표시할 최근 event 개수 (기본 20)")
	cmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "새 event를 1초 간격으로 따라가기")
	return cmd
}
