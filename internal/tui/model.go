// Package tui hosts the bubbletea-based terminal UI for cli buddy. v0.6.6
// shipped the W3-2 minimum-viable subset: a read-only agent list with
// j/k navigation and q-to-quit. The detail view (Enter on a list row →
// the latest-run summary) is the first W3-2 follow-on item, tracked in
// cli-buddy-spec.md §9. Create form, live log tail, and scheduler status
// pane remain follow-on cycles.
//
// The Update method is intentionally a pure-function reducer so tests
// can drive it with synthetic tea.Msg values without spinning up a real
// terminal. The View method is best-effort — its rendered output is
// asserted only via substring matches.
package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/0xmhha/buddy/internal/agent"
)

// AgentLister is the read-only Store surface the TUI needs. Narrowing the
// dependency to a two-method interface keeps Model tests decoupled from
// SQLite and makes it easy to inject canned agents + runs in unit tests.
//
// LatestRun mirrors the Store method of the same name. Callers must
// receive agent.ErrNotFound when the agent has never run — the View
// renders that case as the friend-tone "no runs yet" copy rather than a
// generic error.
type AgentLister interface {
	List(ctx context.Context) ([]agent.Agent, error)
	LatestRun(ctx context.Context, agentID string) (agent.AgentRun, error)
}

// Mode is the top-level view state: list (default) vs detail.
type Mode int

const (
	ModeList Mode = iota
	ModeDetail
)

// Model is the bubbletea model. Fields are exported so reducer tests can
// inspect state directly without going through a helper.
type Model struct {
	Store  AgentLister
	Agents []agent.Agent
	Cursor int
	Width  int
	Height int
	Err    error
	Loaded bool // false until the first agentsLoadedMsg / errMsg lands

	// Detail-view state — only meaningful when Mode==ModeDetail.
	Mode         Mode
	Selected     string // agent ID the detail pane is rendering
	Detail       agent.AgentRun
	DetailErr    error
	DetailLoaded bool // false until the first AgentDetailLoadedMsg / AgentDetailErrMsg lands
}

// SelectedID returns the agent ID currently focused for detail rendering.
// In list mode this returns the cursor row's ID (or "" on an empty list);
// in detail mode it returns the locked-in Selected field.
func (m Model) SelectedID() string {
	if m.Mode == ModeDetail {
		return m.Selected
	}
	if len(m.Agents) == 0 {
		return ""
	}
	return m.Agents[m.Cursor].ID
}

// Msg variants the reducer consumes. Kept exported so tests can post them
// directly to Update.
type (
	AgentsLoadedMsg       struct{ Agents []agent.Agent }
	ErrMsg                struct{ Err error }
	AgentDetailLoadedMsg  struct{ Run agent.AgentRun }
	AgentDetailErrMsg     struct{ Err error }
)

// NewModel constructs an empty Model around an AgentLister. The Init
// command is what kicks off the first load — until that resolves, View
// renders a "loading…" placeholder.
func NewModel(store AgentLister) Model {
	return Model{Store: store}
}

// Init is bubbletea's startup hook. It fires the initial agent load.
func (m Model) Init() tea.Cmd {
	return loadAgentsCmd(m.Store)
}

// loadAgentsCmd is the tea.Cmd that calls Store.List in the background
// and emits AgentsLoadedMsg / ErrMsg back into the reducer.
func loadAgentsCmd(store AgentLister) tea.Cmd {
	return func() tea.Msg {
		agents, err := store.List(context.Background())
		if err != nil {
			return ErrMsg{Err: err}
		}
		return AgentsLoadedMsg{Agents: agents}
	}
}

// loadDetailCmd fetches the LatestRun for a single agent. We surface
// ErrNotFound on AgentDetailErrMsg too — the View distinguishes the two
// (no-runs-yet vs real error) by checking errors.Is on the recorded
// DetailErr.
func loadDetailCmd(store AgentLister, agentID string) tea.Cmd {
	return func() tea.Msg {
		run, err := store.LatestRun(context.Background(), agentID)
		if err != nil {
			return AgentDetailErrMsg{Err: err}
		}
		return AgentDetailLoadedMsg{Run: run}
	}
}

// Update is the pure-function reducer. It maps (state, msg) → (state', cmd).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case AgentsLoadedMsg:
		m.Agents = msg.Agents
		m.Loaded = true
		m.Err = nil
		if m.Cursor >= len(m.Agents) {
			m.Cursor = max(0, len(m.Agents)-1)
		}
		return m, nil

	case ErrMsg:
		m.Err = msg.Err
		m.Loaded = true
		return m, nil

	case AgentDetailLoadedMsg:
		m.Detail = msg.Run
		m.DetailLoaded = true
		m.DetailErr = nil
		return m, nil

	case AgentDetailErrMsg:
		m.DetailErr = msg.Err
		m.DetailLoaded = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey dispatches keystrokes per-mode. Quit keys work in every mode;
// navigation keys are scoped to list mode; back/return keys are scoped to
// detail mode. r refreshes whichever pane is active.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit, no matter the mode.
	if key == "ctrl+c" || key == "q" {
		return m, tea.Quit
	}

	if m.Mode == ModeDetail {
		switch key {
		case "esc", "h":
			m.Mode = ModeList
			return m, nil
		case "r":
			if m.Selected == "" {
				return m, nil
			}
			m.DetailLoaded = false
			return m, loadDetailCmd(m.Store, m.Selected)
		}
		// In detail mode every other key (j/k/g/G/etc.) is intentionally
		// inert — the detail pane is read-only.
		return m, nil
	}

	// ── list mode ────────────────────────────────────────────────────
	switch key {
	case "j", "down":
		if m.Cursor < len(m.Agents)-1 {
			m.Cursor++
		}
		return m, nil

	case "k", "up":
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil

	case "g", "home":
		m.Cursor = 0
		return m, nil

	case "G", "end":
		if len(m.Agents) > 0 {
			m.Cursor = len(m.Agents) - 1
		}
		return m, nil

	case "r":
		// Manual refresh — handy when the user has just `buddy agent
		// create`d something from another shell and doesn't want to
		// wait for the scheduler's next poll.
		m.Loaded = false
		return m, loadAgentsCmd(m.Store)

	case "enter", "l", "right":
		if len(m.Agents) == 0 {
			return m, nil
		}
		m.Mode = ModeDetail
		m.Selected = m.Agents[m.Cursor].ID
		m.DetailLoaded = false
		m.DetailErr = nil
		m.Detail = agent.AgentRun{}
		return m, loadDetailCmd(m.Store, m.Selected)
	}
	return m, nil
}

// View renders the current Model state. Styles are lipgloss-driven but
// kept restrained: cli buddy's persona is "친구 — silent default", and a
// neon dashboard works against that.
func (m Model) View() string {
	if m.Mode == ModeDetail {
		return m.renderDetail()
	}
	return m.renderList()
}

func (m Model) renderList() string {
	var b strings.Builder

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

	b.WriteString("\n")
	b.WriteString(footerHintDetail())
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
	return footerStyle.Render("j/k or ↑/↓ move · g/G top/bottom · enter/l detail · r refresh · q quit")
}

func footerHintDetail() string {
	return footerStyle.Render("esc/h back · r refresh · q quit")
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
