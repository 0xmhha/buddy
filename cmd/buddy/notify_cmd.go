package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/config"
	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/notify"
)

// newNotifyCmd wires `buddy notify ...`. Subcommands + flags:
//
//   - test    --channel <name>    : fire a synthetic notification
//   - status  [--since DUR]       : recent log rows
//   - --prompt                    : single-line shell PS1 output
func newNotifyCmd() *cobra.Command {
	var prompt bool
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Notification delivery for advisor advisories",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !prompt {
				return cmd.Help()
			}
			return runPromptShell(cmd.Context(), cmd)
		},
	}
	cmd.Flags().BoolVar(&prompt, "prompt", false, "emit single-line summary for shell PS1 integration")

	// Shared sub-cmd flag — db. Each sub-cmd declares its own to avoid
	// the root --prompt mode capturing the wrong arg.
	cmd.AddCommand(newNotifyTestCmd(), newNotifyStatusCmd())
	return cmd
}

// ─── test ──────────────────────────────────────────────────────────────

func newNotifyTestCmd() *cobra.Command {
	var (
		dbFlag   string
		cfgFlag  string
		channel  string
	)
	c := &cobra.Command{
		Use:   "test",
		Short: "Dispatch a synthetic notification through one channel",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if channel == "" {
				return errors.New("--channel 필요해 (desktop|webhook|tui-banner|shell)")
			}
			disp, closer, err := buildDispatcher(dbFlag, cfgFlag, channel)
			if err != nil {
				return err
			}
			defer closer()

			fake := advisor.Advisory{
				ID: -1, Kind: "buddy-test", Severity: advisor.SeverityWarn,
				Message: "테스트 알림 — 채널 동작 확인용", CreatedAt: time.Now().UTC(),
			}
			sent := disp.Dispatch(ctx, []notify.Notifiable{fake})
			if sent[channel] == 0 {
				return fmt.Errorf("channel=%q 가 dispatch 안 됐어. config 확인해줘 (severity_min / dedup window 등)", channel)
			}
			fmt.Printf("buddy: %s 채널로 테스트 알림 보냈어.\n", channel)
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db")
	c.Flags().StringVar(&cfgFlag, "config", "", "path to config.json")
	c.Flags().StringVar(&channel, "channel", "", "desktop|webhook|tui-banner|shell")
	return c
}

// ─── status ────────────────────────────────────────────────────────────

func newNotifyStatusCmd() *cobra.Command {
	var (
		dbFlag   string
		sinceStr string
		channel  string
		limit    int
	)
	c := &cobra.Command{
		Use:   "status",
		Short: "List recent notification log rows",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			conn, err := db.Open(db.Options{Path: dbFlag})
			if err != nil {
				return err
			}
			defer conn.Close()
			store := notify.NewStore(conn)

			opts := notify.ListOptions{Channel: channel, Limit: limit}
			if sinceStr != "" {
				dur, err := parseDurationDays(sinceStr)
				if err != nil {
					return fmt.Errorf("--since: %w", err)
				}
				opts.Since = time.Now().UTC().Add(-dur)
			}
			rows, err := store.List(ctx, opts)
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				fmt.Println("buddy: 기록된 알림이 없어. daemon 가동 + advisor 생성 후 확인해줘.")
				return nil
			}
			fmt.Println(renderNotifyLog(rows))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db")
	c.Flags().StringVar(&sinceStr, "since", "", "lookback duration (e.g. 24h, 7d)")
	c.Flags().StringVar(&channel, "channel", "", "filter by channel name")
	c.Flags().IntVar(&limit, "limit", 50, "max rows to return")
	return c
}

// ─── --prompt ──────────────────────────────────────────────────────────

func runPromptShell(ctx context.Context, cmd *cobra.Command) error {
	// Single-line summary for PS1 hooks. Reads the most recent
	// unmuted advisory (any channel that sent in the last hour);
	// silent when nothing fresh.
	conn, err := db.Open(db.Options{})
	if err != nil {
		fmt.Println("")
		return nil // PS1 must not error-out the user's shell
	}
	defer conn.Close()
	store := notify.NewStore(conn)
	rows, err := store.List(ctx, notify.ListOptions{
		Since: time.Now().UTC().Add(-1 * time.Hour),
		Limit: 1,
	})
	if err != nil || len(rows) == 0 {
		fmt.Println("")
		return nil
	}
	r := rows[0]
	glyph := "·"
	switch r.Severity {
	case notify.SeverityHigh:
		glyph = "⚠"
	case notify.SeverityWarn:
		glyph = "!"
	case notify.SeverityInfo:
		glyph = "i"
	}
	fmt.Printf("%s buddy %s", glyph, r.Kind)
	return nil
}

// ─── helpers ───────────────────────────────────────────────────────────

func buildDispatcher(dbFlag, cfgFlag, channel string) (*notify.Dispatcher, func(), error) {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	closer := func() { _ = conn.Close() }
	store := notify.NewStore(conn)
	disp := notify.NewDispatcher(store)

	eff := config.Defaults()
	if cfg, err := loadConfigForNotify(cfgFlag); err == nil {
		eff = cfg.Effective()
	}
	wireNotifyChannels(disp, eff, channel)
	return disp, closer, nil
}

// wireNotifyChannels registers only the requested channel (for `test`)
// or all configured channels (when channel is empty).
func wireNotifyChannels(disp *notify.Dispatcher, eff config.Effective, only string) {
	want := func(name string) bool { return only == "" || only == name }

	if want(notify.ChannelDesktop) && eff.NotifyDesktopEnabled {
		disp.AddChannel(notify.NewDesktopChannel(), notify.ChannelConfig{
			Enabled: true, SeverityMin: notify.Severity(eff.NotifyDesktopSeverityMin),
			DedupWindow: eff.NotifyDesktopDedup,
		})
	}
	if want(notify.ChannelTUIBanner) && eff.NotifyTUIBannerEnabled {
		disp.AddChannel(notify.NewTUIBannerChannel(), notify.ChannelConfig{
			Enabled: true, SeverityMin: notify.Severity(eff.NotifyTUIBannerSeverityMin),
			DedupWindow: 0,
		})
	}
	if want(notify.ChannelShell) && eff.NotifyShellPromptEnabled {
		disp.AddChannel(notify.NewShellPromptChannel(), notify.ChannelConfig{
			Enabled: true, SeverityMin: notify.Severity(eff.NotifyShellPromptSeverityMin),
			DedupWindow: 0,
		})
	}
	if want(notify.ChannelWebhook) {
		for _, w := range eff.NotifyWebhooks {
			disp.AddChannel(notify.NewWebhookChannel(notify.WebhookConfig{
				URL: w.URL, Method: w.Method, Headers: w.Headers,
				Timeout: w.Timeout,
			}), notify.ChannelConfig{
				Enabled: true, SeverityMin: notify.Severity(w.SeverityMin),
				DedupWindow: w.DedupWindow,
			})
		}
	}
}

func loadConfigForNotify(path string) (config.Config, error) {
	if path == "" {
		p, err := config.DefaultPath()
		if err != nil {
			return config.Config{}, err
		}
		path = p
	}
	return config.Load(path)
}

func renderNotifyLog(rows []notify.LogRow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-6s %-12s %-22s %-6s %-19s %-18s %s\n",
		"ID", "CHANNEL", "KIND", "SEV", "SENT_AT", "OUTCOME", "DETAIL")
	for _, r := range rows {
		detail := r.Detail
		if len(detail) > 50 {
			detail = detail[:50] + "…"
		}
		fmt.Fprintf(&b, "%-6d %-12s %-22s %-6s %-19s %-18s %s\n",
			r.ID, r.Channel, r.Kind, r.Severity,
			r.SentAt.Local().Format("2006-01-02 15:04"),
			r.Outcome, detail)
	}
	return b.String()
}
