package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/usage"
)

// newUsageCmd wires `buddy usage ...`. Reads the sessions table (no
// writes) and renders the metrics in friend-tone Korean prose.
// Surfaces:
//   - today      → spend + counts for the last 24h
//   - trend      → spend per N days
//   - top        → top-N sessions by tokens
//   - overview   → all metrics in one snapshot (used by TUI Usage pane)
//   - distribution → hour-of-day usage histogram
func newUsageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "AI-usage analytics over the sessions table",
	}
	cmd.AddCommand(
		newUsageTodayCmd(),
		newUsageTrendCmd(),
		newUsageTopCmd(),
		newUsageOverviewCmd(),
		newUsageDistributionCmd(),
	)
	return cmd
}

func openUsageService(dbFlag string) (*usage.Service, func(), error) {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	return usage.NewService(conn), func() { _ = conn.Close() }, nil
}

// parseWindow turns --since flag into a TimeWindow. Empty = all-time.
func parseWindow(sinceStr string) (usage.TimeWindow, error) {
	if sinceStr == "" {
		return usage.TimeWindow{}, nil
	}
	dur, err := parseDurationDays(sinceStr)
	if err != nil {
		return usage.TimeWindow{}, fmt.Errorf("--since: %w", err)
	}
	return usage.TimeWindow{Since: time.Now().UTC().Add(-dur)}, nil
}

// parseDurationDays accepts Go duration formats plus `Nd` for days.
// Reuses session_cmd.go's parsePurgeBefore idiom but returns a Duration.
func parseDurationDays(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		body := strings.TrimSuffix(s, "d")
		if days, err := time.ParseDuration(body + "h"); err == nil {
			return days * 24, nil
		}
	}
	return time.ParseDuration(s)
}

// ─── today ─────────────────────────────────────────────────────────────

func newUsageTodayCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "today",
		Short: "Token spend + session count over the last 24h",
		RunE: func(cmd *cobra.Command, _ []string) error {
			svc, closer, err := openUsageService(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			w := usage.TimeWindow{Since: time.Now().UTC().Add(-24 * time.Hour)}
			return printSpendAndStats(cmd.Context(), svc, w, "지난 24시간")
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

// ─── trend ─────────────────────────────────────────────────────────────

func newUsageTrendCmd() *cobra.Command {
	var (
		dbFlag string
		days   int
	)
	c := &cobra.Command{
		Use:   "trend",
		Short: "Token spend + session count over the last --days days (default 7)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if days <= 0 {
				return errors.New("--days 는 양수여야 해")
			}
			svc, closer, err := openUsageService(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			w := usage.TimeWindow{Since: time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)}
			if err := printSpendAndStats(cmd.Context(), svc, w, fmt.Sprintf("지난 %d일", days)); err != nil {
				return err
			}
			daily, err := svc.QueryDailySpend(cmd.Context(), days)
			if err != nil {
				return err
			}
			if chart := renderDailyBarChart(daily); chart != "" {
				fmt.Println()
				fmt.Println(chart)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().IntVar(&days, "days", 7, "lookback window in days")
	return c
}

// ─── top ───────────────────────────────────────────────────────────────

func newUsageTopCmd() *cobra.Command {
	var (
		dbFlag   string
		limit    int
		sinceStr string
	)
	c := &cobra.Command{
		Use:   "top",
		Short: "Top sessions by total tokens (default limit 10)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w, err := parseWindow(sinceStr)
			if err != nil {
				return err
			}
			svc, closer, err := openUsageService(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			top, err := svc.QueryTopSessions(cmd.Context(), w, limit)
			if err != nil {
				return err
			}
			if len(top) == 0 {
				fmt.Println("표시할 세션이 없어. `buddy session list` 또는 `buddy daemon start` 로 모니터 활성.")
				return nil
			}
			fmt.Println(renderTopSessions(top))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().IntVar(&limit, "limit", 10, "max sessions to return")
	c.Flags().StringVar(&sinceStr, "since", "", "lookback window (e.g., 24h, 7d). empty = all time")
	return c
}

// ─── overview ──────────────────────────────────────────────────────────

func newUsageOverviewCmd() *cobra.Command {
	var (
		dbFlag   string
		sinceStr string
		topN     int
	)
	c := &cobra.Command{
		Use:   "overview",
		Short: "All 7 metric in one snapshot (used by TUI Usage pane)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w, err := parseWindow(sinceStr)
			if err != nil {
				return err
			}
			svc, closer, err := openUsageService(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			ov, err := svc.QueryOverview(cmd.Context(), w, topN)
			if err != nil {
				return err
			}
			fmt.Println(renderOverview(ov))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().StringVar(&sinceStr, "since", "", "lookback window (e.g., 24h, 7d). empty = all time")
	c.Flags().IntVar(&topN, "top", 5, "rows in the top sessions section")
	return c
}

// ─── distribution ──────────────────────────────────────────────────────

func newUsageDistributionCmd() *cobra.Command {
	var (
		dbFlag   string
		sinceStr string
	)
	c := &cobra.Command{
		Use:   "distribution",
		Short: "Hour-of-day usage histogram (local time)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w, err := parseWindow(sinceStr)
			if err != nil {
				return err
			}
			svc, closer, err := openUsageService(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			td, err := svc.QueryTimeDistribution(cmd.Context(), w)
			if err != nil {
				return err
			}
			fmt.Println(renderTimeDistribution(td))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().StringVar(&sinceStr, "since", "", "lookback window (e.g., 24h, 7d). empty = all time")
	return c
}

// ─── shared print helpers ─────────────────────────────────────────────

func printSpendAndStats(ctx context.Context, svc *usage.Service, w usage.TimeWindow, label string) error {
	spend, err := svc.QueryTokenSpend(ctx, w)
	if err != nil {
		return err
	}
	stats, err := svc.QuerySessionStats(ctx, w)
	if err != nil {
		return err
	}
	if stats.TotalSessions == 0 {
		fmt.Printf("%s 동안 관찰된 세션이 없어.\n", label)
		return nil
	}
	fmt.Printf("%s 사용 현황\n\n", label)
	fmt.Println(renderTokenSpend(spend))
	fmt.Println()
	fmt.Println(renderSessionStats(stats))
	return nil
}

// ─── render ───────────────────────────────────────────────────────────

func renderTokenSpend(s usage.TokenSpend) string {
	var b strings.Builder
	fmt.Fprintln(&b, "토큰 사용량")
	fmt.Fprintf(&b, "  total:        %s\n", humanInt(s.TotalTokens()))
	fmt.Fprintf(&b, "  input:        %s\n", humanInt(s.InputTokens))
	fmt.Fprintf(&b, "  output:       %s\n", humanInt(s.OutputTokens))
	fmt.Fprintf(&b, "  cache_read:   %s\n", humanInt(s.CacheReadTokens))
	fmt.Fprintf(&b, "  cache_create: %s\n", humanInt(s.CacheCreateTokens))
	fmt.Fprintf(&b, "  cache hit:    %.1f%%\n", s.CacheHitRatio()*100)
	return b.String()
}

func renderSessionStats(s usage.SessionStats) string {
	var b strings.Builder
	fmt.Fprintln(&b, "세션 통계")
	fmt.Fprintf(&b, "  total:        %d (active %d / ended %d)\n",
		s.TotalSessions, s.ActiveSessions, s.EndedSessions)
	if s.EndedSessions > 0 {
		fmt.Fprintf(&b, "  duration:     p50=%s p90=%s max=%s\n",
			shortDur(s.DurationP50), shortDur(s.DurationP90), shortDur(s.DurationMax))
	}
	fmt.Fprintf(&b, "  goal 기록률:  %.0f%%\n", s.GoalTextRatio*100)
	return b.String()
}

func renderTopSessions(top []usage.TopSession) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-12s %-20s %12s  %s\n", "ID", "STARTED", "TOKENS", "GOAL")
	for _, s := range top {
		short := s.ID
		if len(short) > 8 {
			short = short[:8]
		}
		goal := s.GoalText
		if len(goal) > 60 {
			goal = goal[:60] + "…"
		}
		fmt.Fprintf(&b, "%-12s %-20s %12s  %s\n",
			short, s.StartedAt.Local().Format("01-02 15:04"),
			humanInt(s.TotalTokens), goal)
	}
	return b.String()
}

// renderDailyBarChart renders the per-day token totals as a horizontal
// bar chart. Mirrors renderTimeDistribution's bar style — 30-cell width
// scaled to the peak day, friend-tone Korean header. Returns "" when
// the series is empty or every day is zero (no chart adds noise).
func renderDailyBarChart(daily []usage.DailySpend) string {
	if len(daily) == 0 {
		return ""
	}
	var peak int64
	for _, d := range daily {
		if t := d.TotalTokens(); t > peak {
			peak = t
		}
	}
	if peak == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "일별 토큰 추이 (peak %s)\n", humanInt(peak))
	for _, d := range daily {
		total := d.TotalTokens()
		barLen := int(float64(total) / float64(peak) * 30)
		bar := strings.Repeat("█", barLen)
		fmt.Fprintf(&b, "  %s  %12s  %s\n",
			d.Date.Local().Format("01-02"), humanInt(total), bar)
	}
	return b.String()
}

func renderTimeDistribution(td usage.TimeDistribution) string {
	var b strings.Builder
	fmt.Fprintf(&b, "시간대 분포 (local time, 총 %d 세션)\n", td.Total)
	if td.Total == 0 {
		return b.String()
	}
	var peak int64
	for _, c := range td.HourCounts {
		if c > peak {
			peak = c
		}
	}
	for h := 0; h < 24; h++ {
		bar := ""
		if peak > 0 {
			barLen := int(float64(td.HourCounts[h]) / float64(peak) * 30)
			bar = strings.Repeat("█", barLen)
		}
		fmt.Fprintf(&b, "  %02d시: %3d  %s\n", h, td.HourCounts[h], bar)
	}
	return b.String()
}

func renderOverview(ov usage.Overview) string {
	var b strings.Builder
	label := "all time"
	if !ov.Window.IsAllTime() {
		label = fmt.Sprintf("since %s", ov.Window.Since.Local().Format("2006-01-02 15:04"))
	}
	fmt.Fprintf(&b, "===== Usage Overview (%s) =====\n\n", label)
	fmt.Fprintln(&b, renderTokenSpend(ov.Spend))
	fmt.Fprintln(&b, renderSessionStats(ov.Stats))
	fmt.Fprintln(&b, renderTimeDistribution(ov.TimeDistribution))
	if len(ov.Top) > 0 {
		fmt.Fprintln(&b, "상위 세션")
		fmt.Fprintln(&b, renderTopSessions(ov.Top))
	}
	return b.String()
}

// humanInt formats large integers with thousands separators (en_US style).
// "1234567" → "1,234,567". Keeps render width predictable across locales.
func humanInt(n int64) string {
	if n < 0 {
		return "-" + humanInt(-n)
	}
	s := fmt.Sprintf("%d", n)
	var out strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteRune(r)
	}
	return out.String()
}

// shortDur renders durations as "12s" / "3m" / "1h23m" rather than Go's
// verbose default. Used in the duration percentile row.
func shortDur(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) - h*60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
