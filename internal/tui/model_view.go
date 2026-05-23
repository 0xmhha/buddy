package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/0xmhha/buddy/internal/agent"
	"github.com/0xmhha/buddy/internal/notify"
)

// View renders the current Model state. Styles are lipgloss-driven but
// kept restrained: cli buddy's persona is "친구 — silent default", and a
// neon dashboard works against that.
func (m Model) View() string {
	var s string
	switch m.Mode {
	case ModeDetail:
		s = m.renderDetail()
	case ModeScheduler:
		s = m.renderScheduler()
	case ModeDeleteConfirm:
		s = m.renderDeleteConfirm()
	case ModeLogTail:
		s = m.renderLogTail()
	case ModeHookStats:
		s = m.renderHookStats()
	case ModeUsage:
		s = m.renderUsage()
	default:
		s = m.renderList()
	}
	return m.constrainWidth(s)
}

// constrainWidth caps every visible line in s at m.Width display columns,
// appending an ellipsis when a line is truncated. ANSI-aware via
// lipgloss.Width. No-op when m.Width <= 0 (no WindowSizeMsg seen yet, or
// test fixture) — that branch matches bubbletea's pre-resize behavior
// where View() may be called before the terminal size is reported.
//
// Width-only: vertical overflow is left to the terminal's own scroll
// region. Dogfood reported horizontal layout breakage on resize;
// per-line truncation is the smallest fix that restores readability
// without rewriting each render function's column math.
func (m Model) constrainWidth(s string) string {
	if m.Width <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) <= m.Width {
			continue
		}
		runes := []rune(line)
		for len(runes) > 0 && lipgloss.Width(string(runes)+"…") > m.Width {
			runes = runes[:len(runes)-1]
		}
		lines[i] = string(runes) + "…"
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderList() string {
	var b strings.Builder

	// Notify banner — top-of-screen advisory teaser. Suppressed when
	// no fetcher wired, no rows loaded, or rows are all dedup/severity
	// skips. Shows up to 3 most recent "sent" outcomes so a fresh
	// daemon dispatch surfaces immediately.
	if m.NotifyFetcher != nil && m.NotifyLoaded && m.NotifyErr == nil {
		shown := 0
		for _, r := range m.NotifyRows {
			if r.Outcome != notify.OutcomeSent {
				continue
			}
			if shown == 0 {
				b.WriteString(headerStyle.Render("buddy 알림"))
				b.WriteString("\n")
			}
			glyph := "·"
			switch r.Severity {
			case notify.SeverityHigh:
				glyph = "⚠"
			case notify.SeverityWarn:
				glyph = "!"
			case notify.SeverityInfo:
				glyph = "i"
			}
			b.WriteString(fmt.Sprintf("  %s [%s] %s — %s\n",
				glyph, r.Channel, r.Kind,
				r.SentAt.Local().Format("01-02 15:04")))
			shown++
			if shown >= 3 {
				break
			}
		}
		if shown > 0 {
			b.WriteString("\n")
		}
	}

	header := headerStyle.Render("buddy agent list")
	b.WriteString(header)
	b.WriteString("\n\n")

	if !m.Loaded {
		b.WriteString(dimStyle.Render("loading agents…"))
		b.WriteString("\n")
		b.WriteString(footerHintList())
		return b.String()
	}

	if m.Err != nil {
		b.WriteString(errorStyle.Render("error: " + m.Err.Error()))
		b.WriteString("\n")
		b.WriteString(footerHintList())
		return b.String()
	}

	if len(m.Agents) == 0 {
		b.WriteString(dimStyle.Render("(no agents yet — run `buddy agent create <spec.yaml>` in another shell)"))
		b.WriteString("\n")
		b.WriteString(footerHintList())
		return b.String()
	}

	for i, a := range m.Agents {
		line := renderAgentRow(a, i == m.Cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(footerHintList())
	return b.String()
}

func (m Model) renderDetail() string {
	var b strings.Builder

	// Resolve the agent record from the list (we keep the full list in
	// memory anyway; refetching would mean a second SQL round-trip just
	// for the header).
	var a agent.Agent
	for _, candidate := range m.Agents {
		if candidate.ID == m.Selected {
			a = candidate
			break
		}
	}

	title := "buddy agent — " + a.ID
	if a.Name != "" && a.Name != a.ID {
		title += "  (" + a.Name + ")"
	}
	b.WriteString(headerStyle.Render(title))
	b.WriteString("\n\n")

	schedule := a.Schedule
	if schedule == "" {
		schedule = "(on-demand)"
	}
	b.WriteString(fmt.Sprintf("  schedule: %s\n", schedule))
	b.WriteString(fmt.Sprintf("  status:   %s\n", a.Status))
	if !a.CreatedAt.IsZero() {
		b.WriteString(fmt.Sprintf("  created:  %s\n", a.CreatedAt.UTC().Format(time.RFC3339)))
	}
	if !a.UpdatedAt.IsZero() {
		b.WriteString(fmt.Sprintf("  updated:  %s\n", a.UpdatedAt.UTC().Format(time.RFC3339)))
	}
	b.WriteString("\n")

	switch {
	case !m.DetailLoaded:
		b.WriteString(dimStyle.Render("  loading latest run…"))
		b.WriteString("\n")

	case m.DetailErr != nil && errors.Is(m.DetailErr, agent.ErrNotFound):
		b.WriteString(dimStyle.Render("  (no runs yet — try `buddy agent run " + a.ID + "`)"))
		b.WriteString("\n")

	case m.DetailErr != nil:
		b.WriteString(errorStyle.Render("  error: " + m.DetailErr.Error()))
		b.WriteString("\n")

	default:
		b.WriteString(fmt.Sprintf("  latest run #%d\n", m.Detail.ID))
		b.WriteString(fmt.Sprintf("    started:   %s\n",
			m.Detail.StartedAt.UTC().Format(time.RFC3339)))
		if m.Detail.EndedAt != nil {
			ended := m.Detail.EndedAt.UTC()
			b.WriteString(fmt.Sprintf("    ended:     %s (duration %s)\n",
				ended.Format(time.RFC3339),
				ended.Sub(m.Detail.StartedAt).Round(time.Second)))
		} else {
			b.WriteString("    ended:     (in-flight)\n")
		}
		b.WriteString(fmt.Sprintf("    exit code: %d\n", m.Detail.ExitCode))
		if m.Detail.Error != "" {
			b.WriteString(fmt.Sprintf("    error:     %s\n", m.Detail.Error))
		}
	}

	if m.EditErr != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("  edit error: " + m.EditErr.Error()))
		b.WriteString("\n  ")
		b.WriteString(dimStyle.Render("(press e to try again)"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footerHintDetail())
	return b.String()
}

func (m Model) renderScheduler() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("buddy scheduler — preview"))
	b.WriteString("\n\n")

	if !m.SchedulerLoaded {
		b.WriteString(dimStyle.Render("  loading scheduler preview…"))
		b.WriteString("\n\n")
		b.WriteString(footerHintScheduler())
		return b.String()
	}

	if m.SchedulerErr != nil {
		b.WriteString(errorStyle.Render("  error: " + m.SchedulerErr.Error()))
		b.WriteString("\n\n")
		b.WriteString(footerHintScheduler())
		return b.String()
	}

	b.WriteString(dimStyle.Render(fmt.Sprintf(
		"  reference now: %s   (preview only — does not run jobs)\n\n",
		m.SchedulerNow.UTC().Format(time.RFC3339))))

	if len(m.SchedulerEntries) == 0 {
		b.WriteString(dimStyle.Render("  (no scheduled agents — every agent in the list is on-demand)"))
		b.WriteString("\n\n")
		b.WriteString(footerHintScheduler())
		return b.String()
	}

	for _, e := range m.SchedulerEntries {
		// A one-glyph "currently running" marker. Only running
		// agents get ⏵; everyone else stays blank so the scheduler pane
		// doesn't gain noise for the common idle case.
		marker := " "
		if e.Status == agent.StatusRunning {
			marker = "⏵"
		}
		if e.Err != nil {
			b.WriteString(fmt.Sprintf("  %s %-30s %-20s %s\n",
				marker,
				e.AgentID,
				e.Schedule,
				errorStyle.Render("invalid: "+e.Err.Error())))
			continue
		}
		b.WriteString(fmt.Sprintf("  %s %-30s %-20s next %s\n",
			marker,
			e.AgentID,
			e.Schedule,
			e.Next.UTC().Format(time.RFC3339)))
	}
	b.WriteString("\n")
	b.WriteString(footerHintScheduler())
	return b.String()
}

func (m Model) renderLogTail() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf(
		"buddy log tail — agent %s — run #%d", m.Selected, m.LogTailRunID)))
	b.WriteString("\n\n")

	if !m.LogTailLoaded {
		b.WriteString(dimStyle.Render("  loading log lines…"))
		b.WriteString("\n\n")
		b.WriteString(footerHintLogTail())
		return b.String()
	}

	if m.LogTailErr != nil {
		b.WriteString(errorStyle.Render("  error: " + m.LogTailErr.Error()))
		b.WriteString("\n  ")
		b.WriteString(dimStyle.Render("(polling continues — last successful chunk preserved above)"))
		b.WriteString("\n\n")
	}

	if len(m.LogTailLines) == 0 {
		b.WriteString(dimStyle.Render("  (no log lines yet — the run may not have produced any output)"))
		b.WriteString("\n\n")
		b.WriteString(footerHintLogTail())
		return b.String()
	}

	for _, l := range m.LogTailLines {
		b.WriteString(fmt.Sprintf("  %s  %-5s  %s\n",
			l.Ts.UTC().Format("15:04:05"),
			l.Level,
			l.Message))
	}
	if m.LogTailDone {
		// Surface "run done, polling stopped" so the user knows
		// the silence is intentional rather than a stuck tail.
		b.WriteString(dimStyle.Render("  (run finished — polling stopped)"))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(footerHintLogTail())
	return b.String()
}

func (m Model) renderHookStats() string {
	var b strings.Builder

	window := m.HookStatsWindow
	if window == "" {
		window = "1h"
	}
	b.WriteString(headerStyle.Render("buddy hook stats — window " + window))
	b.WriteString("\n\n")

	if !m.HookStatsLoaded {
		b.WriteString(dimStyle.Render("  loading hook stats…"))
		b.WriteString("\n\n")
		b.WriteString(footerHintHookStats())
		return b.String()
	}

	if m.HookStatsErr != nil {
		b.WriteString(errorStyle.Render("  error: " + m.HookStatsErr.Error()))
		b.WriteString("\n\n")
		b.WriteString(footerHintHookStats())
		return b.String()
	}

	if len(m.HookStatsResult.Rows) == 0 {
		b.WriteString(dimStyle.Render("  (no hook events in this window — daemon may be idle or DB empty)"))
		b.WriteString("\n\n")
		b.WriteString(footerHintHookStats())
		return b.String()
	}

	// Column header — same column ordering as `buddy stats` CLI so users
	// who switch between the two surfaces see the same shape.
	b.WriteString(fmt.Sprintf("  %-24s %-12s %8s %8s %8s %8s\n",
		"hook", "tool", "count", "fail", "p50ms", "p95ms"))
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %s\n",
		strings.Repeat("─", 72))))
	for _, row := range m.HookStatsResult.Rows {
		tool := row.ToolName
		if tool == "" {
			tool = "-"
		}
		b.WriteString(fmt.Sprintf("  %-24s %-12s %8d %8d %8d %8d\n",
			row.HookName, tool, row.Count, row.Failures, row.P50Ms, row.P95Ms))
	}
	b.WriteString("\n")
	b.WriteString(footerHintHookStats())
	return b.String()
}

func (m Model) renderUsage() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("buddy usage — AI-usage analytics"))
	b.WriteString("\n\n")

	if m.UsageFetcher == nil {
		b.WriteString(dimStyle.Render("  Usage 데이터를 가져올 수 없어 (UsageFetcher 미설정)."))
		b.WriteString("\n  ")
		b.WriteString(dimStyle.Render("buddy.db 에 sessions 테이블이 있는지 확인해줘."))
		b.WriteString("\n\n")
		b.WriteString(footerHintUsage())
		return b.String()
	}

	if !m.UsageLoaded {
		b.WriteString(dimStyle.Render("  loading usage overview…"))
		b.WriteString("\n\n")
		b.WriteString(footerHintUsage())
		return b.String()
	}

	if m.UsageErr != nil {
		b.WriteString(errorStyle.Render("  error: " + m.UsageErr.Error()))
		b.WriteString("\n\n")
		b.WriteString(footerHintUsage())
		return b.String()
	}

	ov := m.UsageResult

	label := "all time"
	if !ov.Window.IsAllTime() {
		label = "since " + ov.Window.Since.Local().Format("2006-01-02 15:04")
	}
	b.WriteString(dimStyle.Render(fmt.Sprintf("  window: %s\n\n", label)))

	// Token spend
	b.WriteString("  토큰 사용량\n")
	b.WriteString(fmt.Sprintf("    total:        %d\n", ov.Spend.TotalTokens()))
	b.WriteString(fmt.Sprintf("    input:        %d   output: %d\n",
		ov.Spend.InputTokens, ov.Spend.OutputTokens))
	b.WriteString(fmt.Sprintf("    cache_read:   %d   cache_create: %d\n",
		ov.Spend.CacheReadTokens, ov.Spend.CacheCreateTokens))
	b.WriteString(fmt.Sprintf("    cache hit:    %.1f%%\n\n",
		ov.Spend.CacheHitRatio()*100))

	// Session stats
	b.WriteString("  세션 통계\n")
	b.WriteString(fmt.Sprintf("    total: %d   active: %d   ended: %d\n",
		ov.Stats.TotalSessions, ov.Stats.ActiveSessions, ov.Stats.EndedSessions))
	if ov.Stats.EndedSessions > 0 {
		b.WriteString(fmt.Sprintf("    duration: p50=%s  p90=%s  max=%s\n",
			ov.Stats.DurationP50, ov.Stats.DurationP90, ov.Stats.DurationMax))
	}
	b.WriteString(fmt.Sprintf("    goal 기록률: %.0f%%\n\n", ov.Stats.GoalTextRatio*100))

	// Time distribution (compact — top 5 peak hours only in TUI)
	if ov.TimeDistribution.Total > 0 {
		b.WriteString("  피크 시간대 (local)\n")
		type hourCount struct {
			Hour  int
			Count int64
		}
		peaks := make([]hourCount, 0, 24)
		for h := 0; h < 24; h++ {
			if ov.TimeDistribution.HourCounts[h] > 0 {
				peaks = append(peaks, hourCount{Hour: h, Count: ov.TimeDistribution.HourCounts[h]})
			}
		}
		// Simple in-place sort: small N (≤24) — bubble sort fine.
		for i := 0; i < len(peaks); i++ {
			for j := i + 1; j < len(peaks); j++ {
				if peaks[j].Count > peaks[i].Count {
					peaks[i], peaks[j] = peaks[j], peaks[i]
				}
			}
		}
		max := 5
		if len(peaks) < max {
			max = len(peaks)
		}
		for i := 0; i < max; i++ {
			b.WriteString(fmt.Sprintf("    %02d시: %d 세션\n", peaks[i].Hour, peaks[i].Count))
		}
		b.WriteString("\n")
	}

	// Top sessions
	if len(ov.Top) > 0 {
		b.WriteString("  상위 세션 (토큰 기준)\n")
		for _, ts := range ov.Top {
			short := ts.ID
			if len(short) > 8 {
				short = short[:8]
			}
			goal := ts.GoalText
			if len(goal) > 50 {
				goal = goal[:50] + "…"
			}
			b.WriteString(fmt.Sprintf("    %-10s %12d  %s\n", short, ts.TotalTokens, goal))
		}
		b.WriteString("\n")
	}

	// Advisor section — co-rendered when fetcher is wired. Skipping
	// entirely when no fetcher keeps the pane clean for installs that
	// haven't enabled the advisor.
	if m.AdvisorFetcher != nil {
		b.WriteString("  조언\n")
		switch {
		case !m.AdvisorLoaded:
			b.WriteString(dimStyle.Render("    loading advisories…"))
			b.WriteString("\n\n")
		case m.AdvisorErr != nil:
			b.WriteString(errorStyle.Render("    error: " + m.AdvisorErr.Error()))
			b.WriteString("\n\n")
		case len(m.AdvisorResult) == 0:
			b.WriteString(dimStyle.Render("    (지금은 알릴 조언이 없어)"))
			b.WriteString("\n\n")
		default:
			for _, a := range m.AdvisorResult {
				b.WriteString(fmt.Sprintf("    [%s · %s] %s\n", a.Kind, a.Severity, a.Message))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(footerHintUsage())
	return b.String()
}

// renderDeleteConfirm draws the modal-style confirmation prompt. The list
// rows are NOT shown — the user's full attention is on the destructive
// choice. ID is rendered explicitly so a redraw / refresh from another
// shell can't trick the user into deleting the wrong row.
func (m Model) renderDeleteConfirm() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("buddy agent — delete?"))
	b.WriteString("\n\n")

	id := m.PendingDeleteID
	if id == "" {
		id = "(no target — press n to return to the list)"
	}
	b.WriteString(fmt.Sprintf("  About to delete agent: %s\n", id))
	b.WriteString(dimStyle.Render("  This also drops the agent's runs and logs (FK cascade)."))
	b.WriteString("\n  ")
	b.WriteString(dimStyle.Render("This cannot be undone."))
	b.WriteString("\n\n")

	b.WriteString(errorStyle.Render("  press y to confirm · N / esc to cancel (default: cancel)"))
	b.WriteString("\n\n")
	b.WriteString(footerHintConfirm())
	return b.String()
}

// renderAgentRow formats one agent in the list. Caller passes whether the
// row is currently selected; that toggles the highlight.
func renderAgentRow(a agent.Agent, selected bool) string {
	prefix := "  "
	style := rowStyle
	if selected {
		prefix = "▸ "
		style = selectedRowStyle
	}
	schedule := a.Schedule
	if schedule == "" {
		schedule = "(on-demand)"
	}
	return style.Render(fmt.Sprintf("%s%-30s %-20s %s", prefix, a.ID, schedule, a.Status))
}

func footerHintList() string {
	return footerStyle.Render("j/k or ↑/↓ move · g/G top/bottom · enter/l detail · s scheduler · c create · d delete · H hook stats · U usage · r refresh · q quit")
}

func footerHintHookStats() string {
	return footerStyle.Render("esc/h back · r refresh · q quit")
}

func footerHintUsage() string {
	return footerStyle.Render("esc/h back · r refresh · q quit")
}

func footerHintDetail() string {
	return footerStyle.Render("esc/h back · r refresh · t tail logs · e edit spec · q quit")
}

func footerHintLogTail() string {
	return footerStyle.Render("esc/h back · r refresh now · q quit · (auto-refresh ~1s)")
}

func footerHintScheduler() string {
	return footerStyle.Render("esc/h back · r refresh · q quit")
}

func footerHintConfirm() string {
	return footerStyle.Render("y confirm · n/esc cancel · q quit")
}

// ─── styles ────────────────────────────────────────────────────────────

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Padding(0, 1)

	rowStyle = lipgloss.NewStyle()

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("11")).
				Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
)
