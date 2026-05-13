package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/agent"
)

// fakeLister is the test-side AgentLister: returns canned agents or a
// canned error. The TUI Update path never spawns a real subprocess so
// these are enough.
type fakeLister struct {
	agents []agent.Agent
	err    error
}

func (f *fakeLister) List(_ context.Context) ([]agent.Agent, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.agents, nil
}

// keyMsg builds a tea.KeyMsg for the given rune/key — bubbletea exposes
// the same construction as tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
// but using KeyType for j/k feels brittle; the public string-equivalent
// constructor mirrors what a real keypress produces.
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestUpdate_QuitOnQ(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	_, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd, "q must return a tea.Cmd (tea.Quit)")
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok, "expected tea.QuitMsg, got %T", msg)
}

func TestUpdate_QuitOnCtrlC(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	_, cmd := m.Update(keyMsg("ctrl+c"))
	require.NotNil(t, cmd)
	_, ok := cmd().(tea.QuitMsg)
	require.True(t, ok)
}

func TestUpdate_AgentsLoadedFoldsIntoState(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	require.False(t, m.Loaded)
	require.Empty(t, m.Agents)

	loaded := AgentsLoadedMsg{Agents: []agent.Agent{
		{ID: "a"},
		{ID: "b"},
		{ID: "c"},
	}}
	next, cmd := m.Update(loaded)
	require.Nil(t, cmd, "agents-loaded message must not chain a follow-up command")
	mm := next.(Model)
	require.True(t, mm.Loaded)
	require.Len(t, mm.Agents, 3)
	require.Equal(t, 0, mm.Cursor, "cursor stays at 0 on first load")
}

func TestUpdate_AgentsLoadedClampsCursorWhenListShrinks(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	// First load: 3 agents. Cursor moves to the bottom.
	next, _ := m.Update(AgentsLoadedMsg{Agents: []agent.Agent{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}})
	m = next.(Model)
	next, _ = m.Update(keyMsg("G")) // jump to end
	m = next.(Model)
	require.Equal(t, 2, m.Cursor)

	// Second load: only 1 agent left. Cursor must clamp to the new end.
	next, _ = m.Update(AgentsLoadedMsg{Agents: []agent.Agent{{ID: "a"}}})
	m = next.(Model)
	require.Equal(t, 0, m.Cursor,
		"cursor must clamp when the agent list shrinks below it")
}

func TestUpdate_ErrMsgRecordsErrorAndMarksLoaded(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	bang := errors.New("db locked")
	next, _ := m.Update(ErrMsg{Err: bang})
	mm := next.(Model)
	require.True(t, mm.Loaded)
	require.ErrorIs(t, mm.Err, bang)
}

func TestUpdate_NavigationRespectsBounds(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	next, _ := m.Update(AgentsLoadedMsg{Agents: []agent.Agent{
		{ID: "a"}, {ID: "b"}, {ID: "c"},
	}})
	m = next.(Model)
	require.Equal(t, 0, m.Cursor)

	// k at top is a no-op.
	next, _ = m.Update(keyMsg("k"))
	m = next.(Model)
	require.Equal(t, 0, m.Cursor)

	// j j j — third j stops at the last row.
	for i := 0; i < 5; i++ {
		next, _ = m.Update(keyMsg("j"))
		m = next.(Model)
	}
	require.Equal(t, 2, m.Cursor, "cursor must not run past len(agents)-1")

	// k twice — back to row 0.
	next, _ = m.Update(keyMsg("k"))
	m = next.(Model)
	next, _ = m.Update(keyMsg("k"))
	m = next.(Model)
	require.Equal(t, 0, m.Cursor)

	// Arrow keys behave identically.
	next, _ = m.Update(keyMsg("down"))
	m = next.(Model)
	require.Equal(t, 1, m.Cursor)
	next, _ = m.Update(keyMsg("up"))
	m = next.(Model)
	require.Equal(t, 0, m.Cursor)
}

func TestUpdate_GAndShiftGJumpToEnds(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	next, _ := m.Update(AgentsLoadedMsg{Agents: []agent.Agent{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"},
	}})
	m = next.(Model)

	next, _ = m.Update(keyMsg("G")) // bottom
	m = next.(Model)
	require.Equal(t, 3, m.Cursor)

	next, _ = m.Update(keyMsg("g")) // top
	m = next.(Model)
	require.Equal(t, 0, m.Cursor)
}

func TestUpdate_RTriggersRefresh(t *testing.T) {
	t.Parallel()
	m := Model{Loaded: true, Store: &fakeLister{}}
	next, cmd := m.Update(keyMsg("r"))
	require.NotNil(t, cmd, "r must schedule a reload command")
	mm := next.(Model)
	require.False(t, mm.Loaded, "Loaded flips back to false until the reload resolves")
}

func TestUpdate_WindowSizeTracked(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	mm := next.(Model)
	require.Equal(t, 120, mm.Width)
	require.Equal(t, 40, mm.Height)
}

// ─── View smoke ────────────────────────────────────────────────────────

// TestView_LoadingPlaceholderBeforeFirstLoad ensures the initial render
// shows a "loading" hint rather than an empty pane.
func TestView_LoadingPlaceholderBeforeFirstLoad(t *testing.T) {
	t.Parallel()
	m := NewModel(&fakeLister{})
	out := m.View()
	require.Contains(t, out, "loading")
	require.Contains(t, out, "q quit", "footer hint must always be visible")
}

// TestView_EmptyStateAfterLoad covers the empty-list path explicitly so
// the friend-tone "(no agents yet — ...)" copy survives refactors.
func TestView_EmptyStateAfterLoad(t *testing.T) {
	t.Parallel()
	m := Model{Loaded: true}
	out := m.View()
	require.Contains(t, out, "no agents yet")
	require.Contains(t, out, "buddy agent create")
}

// TestView_RendersAgentRowsWithCursor covers the happy path: agent IDs
// and schedules appear, and the selected row has the cursor marker.
func TestView_RendersAgentRowsWithCursor(t *testing.T) {
	t.Parallel()
	m := Model{
		Loaded: true,
		Agents: []agent.Agent{
			{ID: "alpha", Schedule: "@daily", Status: agent.StatusIdle},
			{ID: "beta", Schedule: "", Status: agent.StatusIdle},
		},
		Cursor: 1,
	}
	out := m.View()
	require.Contains(t, out, "alpha")
	require.Contains(t, out, "@daily")
	require.Contains(t, out, "beta")
	require.Contains(t, out, "(on-demand)",
		"agents with empty schedule should show (on-demand) in the schedule column")
	// The cursor marker "▸" should appear on the second line (cursor=1).
	// We can't assert positional layout precisely after lipgloss styling,
	// but the marker must show up somewhere.
	require.Contains(t, out, "▸")
}

func TestView_ErrorState(t *testing.T) {
	t.Parallel()
	m := Model{Loaded: true, Err: errors.New("db locked")}
	out := m.View()
	require.True(t,
		strings.Contains(out, "error") && strings.Contains(out, "db locked"),
		"error message must be visible in View output, got:\n%s", out)
}
