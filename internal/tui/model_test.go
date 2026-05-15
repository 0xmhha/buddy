package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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
	// run is returned by LatestRun, keyed by agent ID. If runErr is set
	// (per ID), LatestRun returns that error instead. A missing key falls
	// back to runDefault / runDefaultErr — handy for "every agent has the
	// same canned run" tests.
	run           map[string]agent.AgentRun
	runErr        map[string]error
	runDefault    agent.AgentRun
	runDefaultErr error
}

func (f *fakeLister) List(_ context.Context) ([]agent.Agent, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.agents, nil
}

func (f *fakeLister) LatestRun(_ context.Context, agentID string) (agent.AgentRun, error) {
	if err, ok := f.runErr[agentID]; ok {
		return agent.AgentRun{}, err
	}
	if r, ok := f.run[agentID]; ok {
		return r, nil
	}
	if f.runDefaultErr != nil {
		return agent.AgentRun{}, f.runDefaultErr
	}
	return f.runDefault, nil
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
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
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

// ─── Detail view (W3-2 follow-on) ──────────────────────────────────────

// TestUpdate_EnterSwitchesToDetailAndFiresLoad — Enter on a populated list
// transitions Mode to Detail, marks DetailLoaded=false, and returns the
// LatestRun fetch command so the reducer stays pure (no I/O on the path).
func TestUpdate_EnterSwitchesToDetailAndFiresLoad(t *testing.T) {
	t.Parallel()
	loader := &fakeLister{}
	m := Model{Store: loader, Loaded: true, Agents: []agent.Agent{
		{ID: "alpha"}, {ID: "beta"},
	}, Cursor: 1}

	next, cmd := m.Update(keyMsg("enter"))
	mm := next.(Model)
	require.Equal(t, ModeDetail, mm.Mode, "enter must switch to detail mode")
	require.False(t, mm.DetailLoaded, "DetailLoaded must reset until LatestRun resolves")
	require.Equal(t, "beta", mm.SelectedID(), "SelectedID must echo the cursor row's ID")
	require.NotNil(t, cmd, "enter must schedule a LatestRun fetch")
}

// TestUpdate_EnterOnEmptyListIsNoOp — Enter when the list is empty must not
// switch modes (there is nothing to show).
func TestUpdate_EnterOnEmptyListIsNoOp(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Loaded: true}
	next, cmd := m.Update(keyMsg("enter"))
	mm := next.(Model)
	require.Equal(t, ModeList, mm.Mode, "enter on empty list must stay in list mode")
	require.Nil(t, cmd, "enter on empty list must not fire a fetch")
}

// TestUpdate_LSwitchesToDetailLikeEnter — vi-style 'l' is an alias for Enter
// so right-hand-side keyboard users have the same affordance.
func TestUpdate_LSwitchesToDetailLikeEnter(t *testing.T) {
	t.Parallel()
	loader := &fakeLister{}
	m := Model{Store: loader, Loaded: true, Agents: []agent.Agent{{ID: "alpha"}}}
	next, cmd := m.Update(keyMsg("l"))
	mm := next.(Model)
	require.Equal(t, ModeDetail, mm.Mode)
	require.NotNil(t, cmd)
}

// TestUpdate_DetailLoadedFoldsIntoState — the AgentDetailLoadedMsg payload
// lands on the model and DetailLoaded flips to true.
func TestUpdate_DetailLoadedFoldsIntoState(t *testing.T) {
	t.Parallel()
	started := time.Date(2026, 5, 13, 18, 10, 0, 0, time.UTC)
	ended := started.Add(82 * time.Second)
	run := agent.AgentRun{
		ID: 42, AgentID: "alpha",
		StartedAt: started, EndedAt: &ended,
		ExitCode: 0,
	}
	m := Model{Mode: ModeDetail, Selected: "alpha"}
	next, cmd := m.Update(AgentDetailLoadedMsg{Run: run})
	mm := next.(Model)
	require.Nil(t, cmd)
	require.True(t, mm.DetailLoaded)
	require.Equal(t, int64(42), mm.Detail.ID)
	require.Equal(t, 0, mm.Detail.ExitCode)
	require.Nil(t, mm.DetailErr, "successful load clears any prior DetailErr")
}

// TestUpdate_DetailErrFoldsIntoState — AgentDetailErrMsg records the error,
// flips DetailLoaded to true (so the View knows the fetch resolved), and
// leaves Detail itself unchanged.
func TestUpdate_DetailErrFoldsIntoState(t *testing.T) {
	t.Parallel()
	bang := errors.New("db locked")
	m := Model{Mode: ModeDetail, Selected: "alpha"}
	next, _ := m.Update(AgentDetailErrMsg{Err: bang})
	mm := next.(Model)
	require.True(t, mm.DetailLoaded)
	require.ErrorIs(t, mm.DetailErr, bang)
}

// TestUpdate_EscReturnsToList — Esc in detail mode goes back to the list,
// preserving cursor + list state.
func TestUpdate_EscReturnsToList(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDetail, Loaded: true,
		Agents:   []agent.Agent{{ID: "a"}, {ID: "b"}},
		Cursor:   1,
		Selected: "b",
	}
	next, cmd := m.Update(keyMsg("esc"))
	mm := next.(Model)
	require.Equal(t, ModeList, mm.Mode)
	require.Equal(t, 1, mm.Cursor, "Esc must preserve list cursor")
	require.Len(t, mm.Agents, 2, "Esc must not touch the list")
	require.Nil(t, cmd)
}

// TestUpdate_HReturnsToListLikeEsc — vi-style 'h' is an alias for Esc.
func TestUpdate_HReturnsToListLikeEsc(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDetail, Loaded: true,
		Agents: []agent.Agent{{ID: "a"}}, Selected: "a"}
	next, _ := m.Update(keyMsg("h"))
	require.Equal(t, ModeList, next.(Model).Mode)
}

// TestUpdate_QuitWorksInDetailMode — q / ctrl+c must still quit when the
// user is in the detail pane.
func TestUpdate_QuitWorksInDetailMode(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDetail}
	_, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	_, ok := cmd().(tea.QuitMsg)
	require.True(t, ok, "q in detail mode must produce QuitMsg")
}

// TestUpdate_ListNavKeysIgnoredInDetail — j/k/g/G must not move the list
// cursor while the user is reading the detail pane (they're scoped to
// list mode).
func TestUpdate_ListNavKeysIgnoredInDetail(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDetail, Loaded: true,
		Agents: []agent.Agent{{ID: "a"}, {ID: "b"}, {ID: "c"}}, Cursor: 0}
	for _, k := range []string{"j", "G", "k", "g"} {
		next, _ := m.Update(keyMsg(k))
		require.Equal(t, 0, next.(Model).Cursor,
			"%s must not move cursor in detail mode", k)
		require.Equal(t, ModeDetail, next.(Model).Mode)
	}
}

// TestUpdate_RInDetailRefetchesDetail — r refresh in detail mode re-fires
// the LatestRun cmd (so the user sees the in-progress run's latest exit).
func TestUpdate_RInDetailRefetchesDetail(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDetail, Selected: "alpha",
		DetailLoaded: true}
	next, cmd := m.Update(keyMsg("r"))
	mm := next.(Model)
	require.NotNil(t, cmd, "r in detail mode must schedule a LatestRun fetch")
	require.False(t, mm.DetailLoaded, "DetailLoaded flips back to false until fetch resolves")
	require.Equal(t, ModeDetail, mm.Mode)
}

// TestView_DetailRendersAgentFields — the detail pane shows the selected
// agent's ID, schedule, status, and the latest-run summary fields. We
// assert via substring so lipgloss padding doesn't break the test.
func TestView_DetailRendersAgentFields(t *testing.T) {
	t.Parallel()
	started := time.Date(2026, 5, 13, 18, 10, 0, 0, time.UTC)
	ended := started.Add(82 * time.Second)
	m := Model{
		Mode:         ModeDetail,
		Loaded:       true,
		DetailLoaded: true,
		Agents: []agent.Agent{{
			ID: "alpha", Name: "webtoon pipeline",
			Schedule: "@daily", Status: agent.StatusIdle,
		}},
		Cursor:   0,
		Selected: "alpha",
		Detail: agent.AgentRun{
			ID: 42, AgentID: "alpha",
			StartedAt: started, EndedAt: &ended,
			ExitCode: 0,
		},
	}
	out := m.View()
	require.Contains(t, out, "alpha")
	require.Contains(t, out, "@daily")
	require.Contains(t, out, "webtoon pipeline")
	require.Contains(t, out, "exit", "exit code label must be visible")
	require.Contains(t, out, "esc", "footer hint must show how to return")
}

// TestView_DetailEmptyRunCopy — ErrNotFound means "agent exists but has
// never run yet". Surface the friend-tone copy, not a stack trace.
func TestView_DetailEmptyRunCopy(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:         ModeDetail,
		Loaded:       true,
		DetailLoaded: true,
		Agents:       []agent.Agent{{ID: "alpha"}},
		Cursor:       0,
		Selected:     "alpha",
		DetailErr:    agent.ErrNotFound,
	}
	out := m.View()
	require.Contains(t, out, "no runs yet")
	require.Contains(t, out, "buddy agent run", "hint must point to how to trigger a run")
}

// TestView_DetailGenericErrorCopy — any non-ErrNotFound detail error must
// render the error message (mirrors the list-mode error state, not the
// empty-state copy).
func TestView_DetailGenericErrorCopy(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:         ModeDetail,
		Loaded:       true,
		DetailLoaded: true,
		Agents:       []agent.Agent{{ID: "alpha"}},
		Cursor:       0,
		Selected:     "alpha",
		DetailErr:    errors.New("disk full"),
	}
	out := m.View()
	require.Contains(t, out, "error")
	require.Contains(t, out, "disk full")
}

// TestView_DetailLoadingPlaceholder — before the detail fetch resolves,
// View must show "loading…" rather than an empty pane.
func TestView_DetailLoadingPlaceholder(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:         ModeDetail,
		Loaded:       true,
		DetailLoaded: false,
		Agents:       []agent.Agent{{ID: "alpha"}},
		Cursor:       0,
		Selected:     "alpha",
	}
	out := m.View()
	require.Contains(t, out, "loading")
}
