// Package tui hosts the bubbletea-based terminal UI for cli buddy. v0.6.6
// ships the W3-2 minimum-viable subset: a read-only agent list with
// j/k navigation and q-to-quit. Detail views, create form, and live
// log tail are W3-2 follow-on cycles tracked in cli-buddy-spec.md §9.
//
// The Update method is intentionally a pure-function reducer so tests
// can drive it with synthetic tea.Msg values without spinning up a real
// terminal. The View method is best-effort — its rendered output is
// asserted only via substring matches.
package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/0xmhha/buddy/internal/agent"
)

// AgentLister is the read-only Store surface the TUI needs. Narrowing the
// dependency to a one-method interface keeps Model tests decoupled from
// SQLite and makes it easy to inject canned agents in unit tests.
type AgentLister interface {
	List(ctx context.Context) ([]agent.Agent, error)
}

// Model is the bubbletea model. Fields are exported so reducer tests can
// inspect state directly without going through a helper.
type Model struct {
	Store        AgentLister
	Agents       []agent.Agent
	Cursor       int
	Width        int
	Height       int
	Err          error
	Loaded       bool // false until the first agentsLoadedMsg / errMsg lands
}

// Msg variants the reducer consumes. Kept exported so tests can post them
// directly to Update.
type (
	AgentsLoadedMsg struct{ Agents []agent.Agent }
	ErrMsg          struct{ Err error }
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
// and emits AgentsLoadedMsg / ErrMsg back into the reducer. Exposed for
// tests via Update's return value (so tests can also call this directly
// for refresh scenarios).
func loadAgentsCmd(store AgentLister) tea.Cmd {
	return func() tea.Msg {
		agents, err := store.List(context.Background())
		if err != nil {
			return ErrMsg{Err: err}
		}
		return AgentsLoadedMsg{Agents: agents}
	}
}

// Update is the pure-function reducer. It maps (state, msg) → (state', cmd).
//   - tea.KeyMsg handles q / ctrl+c / j / k / r (refresh).
//   - AgentsLoadedMsg / ErrMsg fold loader results into state.
//   - tea.WindowSizeMsg tracks terminal geometry for View rendering.
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

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
		}
	}
	return m, nil
}

// View renders the current Model state. Styles are lipgloss-driven but
// kept restrained: cli buddy's persona is "친구 — silent default", and a
// neon dashboard works against that.
func (m Model) View() string {
	var b strings.Builder

	header := headerStyle.Render("buddy agent list")
	b.WriteString(header)
	b.WriteString("\n\n")

	if !m.Loaded {
		b.WriteString(dimStyle.Render("loading agents…"))
		b.WriteString("\n")
		b.WriteString(footerHint())
		return b.String()
	}

	if m.Err != nil {
		b.WriteString(errorStyle.Render("error: " + m.Err.Error()))
		b.WriteString("\n")
		b.WriteString(footerHint())
		return b.String()
	}

	if len(m.Agents) == 0 {
		b.WriteString(dimStyle.Render("(no agents yet — run `buddy agent create <spec.yaml>` in another shell)"))
		b.WriteString("\n")
		b.WriteString(footerHint())
		return b.String()
	}

	for i, a := range m.Agents {
		line := renderAgentRow(a, i == m.Cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(footerHint())
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

func footerHint() string {
	return footerStyle.Render("j/k or ↑/↓ move · g/G top/bottom · r refresh · q quit")
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
