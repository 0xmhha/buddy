package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/sessions"
)

// newSessionCmd wires `buddy session ...` per ADR-012. Surfaces the
// W7-1 Session Monitor read API to the shell. Mirrors `buddy agent` 의
// list/show/purge shape.
func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Claude Code session observability (W7-1 / ADR-012)",
	}
	cmd.AddCommand(
		newSessionListCmd(),
		newSessionShowCmd(),
		newSessionPurgeCmd(),
		newSessionRegisterCmd(),
	)
	return cmd
}

func openSessionStore(dbFlag string) (*sessions.Store, func(), error) {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	return sessions.NewStore(conn), func() { _ = conn.Close() }, nil
}

// ─── list ──────────────────────────────────────────────────────────────

func newSessionListCmd() *cobra.Command {
	var (
		dbFlag   string
		all      bool
		sinceStr string
		refresh  bool
	)
	c := &cobra.Command{
		Use:   "list",
		Short: "List observed Claude Code sessions (default: active only)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			store, closer, err := openSessionStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			if refresh {
				// On-demand fsLister scan; daemon may also be doing this
				// in the background.
				l := sessions.NewFSLister(store)
				if _, err := l.List(ctx); err != nil {
					return fmt.Errorf("refresh: %w", err)
				}
			}

			opts := sessions.ListOptions{IncludeEnded: all}
			if sinceStr != "" {
				dur, err := time.ParseDuration(sinceStr)
				if err != nil {
					return fmt.Errorf("--since: %w", err)
				}
				opts.Since = time.Now().Add(-dur)
			}
			ss, err := store.List(ctx, opts)
			if err != nil {
				return err
			}
			if len(ss) == 0 {
				fmt.Println("관찰 중인 세션이 없어. `buddy session list --refresh` 또는 `buddy daemon start` 로 모니터 활성.")
				return nil
			}
			fmt.Println(renderSessionTable(ss))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().BoolVar(&all, "all", false, "include ended sessions")
	c.Flags().StringVar(&sinceStr, "since", "", "filter to sessions with last_active within duration (e.g., 1h, 24h)")
	c.Flags().BoolVar(&refresh, "refresh", false, "run on-demand fsLister scan before listing")
	return c
}

// ─── show ──────────────────────────────────────────────────────────────

func newSessionShowCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one session's full detail",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			store, closer, err := openSessionStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			sess, err := store.Get(ctx, args[0])
			if err != nil {
				if errors.Is(err, sessions.ErrNotFound) {
					return fmt.Errorf("세션 %q 를 못 찾아", args[0])
				}
				return err
			}
			fmt.Println(renderSessionDetail(sess))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

// ─── purge ─────────────────────────────────────────────────────────────

func newSessionPurgeCmd() *cobra.Command {
	var (
		dbFlag    string
		beforeStr string
		apply     bool
	)
	c := &cobra.Command{
		Use:   "purge",
		Short: "Purge ended sessions older than --before (dry-run by default)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if beforeStr == "" {
				return errors.New("--before 가 필요해 (e.g., 30d, 168h, 2026-01-01)")
			}
			cutoff, err := parsePurgeBefore(beforeStr)
			if err != nil {
				return err
			}
			store, closer, err := openSessionStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			if !apply {
				n, err := store.CountBefore(ctx, cutoff)
				if err != nil {
					return err
				}
				fmt.Printf("Dry-run: %d 개의 종료된 세션이 삭제 대상이야. 실행하려면 --apply.\n", n)
				return nil
			}
			n, err := store.PurgeBefore(ctx, cutoff)
			if err != nil {
				return err
			}
			fmt.Printf("%d 개 종료 세션 삭제 완료.\n", n)
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().StringVar(&beforeStr, "before", "", "duration (30d / 168h) or date (2026-01-01) or RFC3339")
	c.Flags().BoolVar(&apply, "apply", false, "actually delete (default: dry-run preview)")
	return c
}

// parsePurgeBefore accepts the same shapes as `buddy agent purge --before`.
// Supports relative durations including `Nd` (days), RFC 3339 timestamps,
// or YYYY-MM-DD dates.
func parsePurgeBefore(s string) (time.Time, error) {
	// Try Nd shorthand (Go time.ParseDuration doesn't support days).
	if strings.HasSuffix(s, "d") {
		body := strings.TrimSuffix(s, "d")
		if days, err := time.ParseDuration(body + "h"); err == nil {
			return time.Now().Add(-days * 24), nil
		}
	}
	if dur, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-dur), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("--before: 알 수 없는 형식 %q (30d / 168h / 2026-01-01 / RFC3339)", s)
}

// ─── register (hook entry; hidden) ────────────────────────────────────

func newSessionRegisterCmd() *cobra.Command {
	var (
		dbFlag         string
		idFlag         string
		transcriptFlag string
		pidFlag        int
	)
	c := &cobra.Command{
		Use:    "register",
		Short:  "Register a session (called by SessionStart hook; not for user)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if idFlag == "" {
				return errors.New("--id 필수")
			}
			store, closer, err := openSessionStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			now := time.Now().UTC()
			return store.Upsert(ctx, sessions.Session{
				ID:             idFlag,
				PID:            pidFlag,
				TranscriptPath: transcriptFlag,
				StartedAt:      now,
				LastActive:     now,
				Metadata:       "{}",
			})
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "")
	c.Flags().StringVar(&idFlag, "id", "", "")
	c.Flags().StringVar(&transcriptFlag, "transcript-path", "", "")
	c.Flags().IntVar(&pidFlag, "pid", 0, "")
	return c
}

// ─── render ────────────────────────────────────────────────────────────

func renderSessionTable(ss []sessions.Session) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-12s %-10s %-20s %-20s %12s %12s\n",
		"ID", "STATUS", "STARTED", "LAST ACTIVE", "INPUT TOK", "OUTPUT TOK")
	for _, s := range ss {
		status := "active"
		if s.EndedAt != nil {
			status = "ended"
		}
		shortID := s.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		fmt.Fprintf(&b, "%-12s %-10s %-20s %-20s %12d %12d\n",
			shortID, status,
			s.StartedAt.Format("01-02 15:04"),
			s.LastActive.Format("01-02 15:04"),
			s.Usage.InputTokens, s.Usage.OutputTokens)
	}
	return b.String()
}

func renderSessionDetail(s sessions.Session) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ID:              %s\n", s.ID)
	fmt.Fprintf(&b, "PID:             %d\n", s.PID)
	fmt.Fprintf(&b, "Transcript:      %s\n", s.TranscriptPath)
	fmt.Fprintf(&b, "Started:         %s\n", s.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Last active:     %s\n", s.LastActive.Format(time.RFC3339))
	if s.EndedAt != nil {
		fmt.Fprintf(&b, "Ended at:        %s\n", s.EndedAt.Format(time.RFC3339))
	} else {
		fmt.Fprintln(&b, "Ended at:        (active)")
	}
	fmt.Fprintf(&b, "Tokens:          input=%d output=%d cache_read=%d cache_create=%d\n",
		s.Usage.InputTokens, s.Usage.OutputTokens,
		s.Usage.CacheReadTokens, s.Usage.CacheCreateTokens)
	fmt.Fprintf(&b, "Last offset:     %d\n", s.LastOffset)
	if s.GoalText != "" {
		fmt.Fprintf(&b, "Goal:            %s\n", s.GoalText)
	}
	if s.Metadata != "" && s.Metadata != "{}" {
		fmt.Fprintf(&b, "Metadata:        %s\n", s.Metadata)
	}
	return b.String()
}
