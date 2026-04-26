package main

// purge_cmd.go owns the `buddy purge --before <date> [--apply]` subcommand.
// Lives next to main.go (rather than inside it) so main.go doesn't keep
// growing — see HANDOFF.md §11 watch list.
//
// Design contract:
//   - DEFAULT is dry-run. Apply requires the explicit --apply flag.
//   - SUCCESS prints a one-line friend-tone summary to stderr (count of
//     deletion candidates / actually-deleted rows). Always mentions that
//     hook_outbox is intentionally untouched, so the user knows they don't
//     need a separate flag for it.
//   - FAILURE flows through friendError so the user gets the friend-tone
//     message verbatim and the process exits 1.
//   - The actual SQL / cutoff logic is delegated to internal/purge — this
//     file is pure CLI plumbing + i18n (English from internal/purge → friend
//     tone Korean here).

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/purge"
)

func newPurgeCmd() *cobra.Command {
	var (
		dbFlag     string
		beforeFlag string
		applyFlag  bool
	)
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "오래된 hook_events·hook_stats 정리 (outbox는 안 건드림)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if beforeFlag == "" {
				return newFriendError(
					"buddy: --before 가 필요해 (예: --before 30d, --before 2026-01-01).")
			}
			cutoff, err := purge.ParseBefore(beforeFlag, time.Now().UTC())
			if err != nil {
				return newFriendError(fmt.Sprintf(
					"buddy: --before 형식이 이상해 (%v). 예: 30d, 2026-01-01, 2026-01-01T00:00:00Z",
					err))
			}
			// Existence-check before writable open so a missing --db path
			// doesn't silently create an empty DB (and worse, run migrations
			// on it). db.Open in writable mode would MkdirAll + lazy-create;
			// purge against "no DB" is a user error, not a bootstrap action.
			// Mirrors the read-only stat-check in db.Open for the T8 sentinel.
			pathToCheck := dbFlag
			if pathToCheck == "" {
				if p, perr := db.DefaultPath(); perr == nil {
					pathToCheck = p
				}
			}
			if pathToCheck != "" {
				if _, statErr := os.Stat(pathToCheck); statErr != nil {
					if errors.Is(statErr, fs.ErrNotExist) {
						return dbMissingFriendError(dbFlag)
					}
					return newFriendError(fmt.Sprintf("buddy: DB를 못 열었어 (%v).", statErr))
				}
			}
			conn, err := db.Open(db.Options{Path: dbFlag})
			if err != nil {
				if errors.Is(err, db.ErrDBMissing) {
					return dbMissingFriendError(dbFlag)
				}
				return newFriendError(fmt.Sprintf("buddy: DB를 못 열었어 (%v).", err))
			}
			defer conn.Close()

			res, err := purge.Run(conn, purge.Options{
				BeforeMillis: cutoff.UnixMilli(),
				DryRun:       !applyFlag,
			})
			if err != nil {
				return newFriendError(fmt.Sprintf("buddy: purge 실패 (%v).", err))
			}

			// Use cmd.ErrOrStderr() (defaults to os.Stderr at runtime; tests
			// swap it via cmd.SetErr). All friend-tone summary lines go to
			// stderr — stdout stays clean for future tooling pipes.
			out := cmd.ErrOrStderr()
			if !applyFlag {
				fmt.Fprintf(out,
					"buddy: dry-run. %d개 hook_events, %d개 hook_stats 가 삭제 대상이야.\n"+
						"buddy: 진짜 지우려면 --apply 추가해줘. (outbox는 안 건드려.)\n",
					res.Events, res.Stats)
			} else {
				fmt.Fprintf(out,
					"buddy: %d개 hook_events, %d개 hook_stats 삭제했어. (outbox는 그대로.)\n",
					res.Events, res.Stats)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&beforeFlag, "before", "",
		"삭제 기준 시점. 예: 30d, 2026-01-01, 2026-01-01T00:00:00Z")
	cmd.Flags().BoolVar(&applyFlag, "apply", false,
		"실제로 삭제 (기본은 dry-run; outbox는 어떤 모드에서도 안 건드림)")
	return cmd
}
