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

	deleted   []string // ids that Delete was called with, in order
	deleteErr error    // canned error from Delete (nil = success)
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

// Delete records the call into deleted so tests can assert what was
// removed, and optionally returns deleteErr. The fake does not mutate
// f.agents — reducer tests post AgentsLoadedMsg explicitly when they
// need to simulate the post-delete reload.
func (f *fakeLister) Delete(_ context.Context, agentID string) error {
	f.deleted = append(f.deleted, agentID)
	return f.deleteErr
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

// ─── Scheduler status pane (W3-2 follow-on #2) ─────────────────────────

// TestUpdate_SSwitchesToScheduler — pressing `s` in list mode opens the
// scheduler pane, marks SchedulerLoaded=false, and fires the preview cmd.
func TestUpdate_SSwitchesToScheduler(t *testing.T) {
	t.Parallel()
	loader := &fakeLister{}
	m := Model{Store: loader, Loaded: true, Agents: []agent.Agent{
		{ID: "alpha", Schedule: "@daily"},
	}}
	next, cmd := m.Update(keyMsg("s"))
	mm := next.(Model)
	require.Equal(t, ModeScheduler, mm.Mode)
	require.False(t, mm.SchedulerLoaded, "SchedulerLoaded resets until the preview resolves")
	require.NotNil(t, cmd, "s must schedule a preview fetch")
}

// TestUpdate_SchedulerStatusLoadedFoldsIntoState — the Loaded message
// populates entries + Now and flips SchedulerLoaded.
func TestUpdate_SchedulerStatusLoadedFoldsIntoState(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 15, 12, 34, 17, 0, time.UTC)
	entries := []SchedulerPreviewEntry{
		{AgentID: "alpha", Schedule: "@daily", Next: now.Add(time.Hour)},
		{AgentID: "beta", Schedule: "broken", Err: errors.New("parse fail")},
	}
	m := Model{Mode: ModeScheduler}
	next, cmd := m.Update(SchedulerStatusLoadedMsg{Now: now, Entries: entries})
	mm := next.(Model)
	require.Nil(t, cmd)
	require.True(t, mm.SchedulerLoaded)
	require.Equal(t, now, mm.SchedulerNow)
	require.Len(t, mm.SchedulerEntries, 2)
	require.Nil(t, mm.SchedulerErr, "successful load clears any prior SchedulerErr")
}

// TestUpdate_SchedulerStatusErrFoldsIntoState — SchedulerStatusErrMsg
// records the error + flips SchedulerLoaded=true so View knows the fetch
// resolved (with an error).
func TestUpdate_SchedulerStatusErrFoldsIntoState(t *testing.T) {
	t.Parallel()
	bang := errors.New("db locked")
	m := Model{Mode: ModeScheduler}
	next, _ := m.Update(SchedulerStatusErrMsg{Err: bang})
	mm := next.(Model)
	require.True(t, mm.SchedulerLoaded)
	require.ErrorIs(t, mm.SchedulerErr, bang)
}

// TestUpdate_EscFromSchedulerReturnsToList — Esc in scheduler mode
// returns to the list, list cursor preserved.
func TestUpdate_EscFromSchedulerReturnsToList(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeScheduler, Loaded: true,
		Agents: []agent.Agent{{ID: "a"}, {ID: "b"}}, Cursor: 1}
	next, _ := m.Update(keyMsg("esc"))
	mm := next.(Model)
	require.Equal(t, ModeList, mm.Mode)
	require.Equal(t, 1, mm.Cursor)
}

// TestUpdate_HFromSchedulerReturnsToList — vi-style 'h' alias.
func TestUpdate_HFromSchedulerReturnsToList(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeScheduler}
	next, _ := m.Update(keyMsg("h"))
	require.Equal(t, ModeList, next.(Model).Mode)
}

// TestUpdate_QuitWorksInSchedulerMode — q must still quit.
func TestUpdate_QuitWorksInSchedulerMode(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeScheduler}
	_, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	_, ok := cmd().(tea.QuitMsg)
	require.True(t, ok)
}

// TestUpdate_RInSchedulerRefetchesPreview — r in scheduler mode re-fires
// the preview cmd and flips SchedulerLoaded back to false.
func TestUpdate_RInSchedulerRefetchesPreview(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeScheduler, SchedulerLoaded: true}
	next, cmd := m.Update(keyMsg("r"))
	mm := next.(Model)
	require.NotNil(t, cmd, "r in scheduler mode must schedule a preview refresh")
	require.False(t, mm.SchedulerLoaded)
	require.Equal(t, ModeScheduler, mm.Mode)
}

// TestUpdate_SInDetailIsNoOp — `s` is a list-mode shortcut; it must not
// hijack the detail pane.
func TestUpdate_SInDetailIsNoOp(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDetail, Selected: "alpha"}
	next, cmd := m.Update(keyMsg("s"))
	require.Equal(t, ModeDetail, next.(Model).Mode, "s must not leak into detail mode")
	require.Nil(t, cmd)
}

// TestView_SchedulerRendersEntries — the pane shows each entry's agent ID,
// schedule, and the next-fire time (formatted RFC3339). The Now header is
// also visible so users know the reference clock.
func TestView_SchedulerRendersEntries(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 15, 12, 34, 17, 0, time.UTC)
	m := Model{
		Mode:            ModeScheduler,
		SchedulerLoaded: true,
		SchedulerNow:    now,
		SchedulerEntries: []SchedulerPreviewEntry{
			{AgentID: "alpha", Schedule: "@daily", Next: now.Add(time.Hour)},
			{AgentID: "beta", Schedule: "*/5 * * * *", Next: now.Add(3 * time.Minute)},
		},
	}
	out := m.View()
	require.Contains(t, out, "alpha")
	require.Contains(t, out, "@daily")
	require.Contains(t, out, "beta")
	require.Contains(t, out, "*/5 * * * *")
	require.Contains(t, out, "esc", "footer hint must show how to return")
}

// TestView_SchedulerEmptyState — when there are no scheduled agents the
// pane shows a friend-tone copy instead of a blank list.
func TestView_SchedulerEmptyState(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:            ModeScheduler,
		SchedulerLoaded: true,
		SchedulerNow:    time.Now(),
		SchedulerEntries: nil,
	}
	out := m.View()
	require.Contains(t, out, "no scheduled agents")
}

// TestView_SchedulerInvalidEntryShowsErrInline — a row whose Schedule
// failed to parse renders inline (the offending agent + a hint), the
// rest of the pane stays usable.
func TestView_SchedulerInvalidEntryShowsErrInline(t *testing.T) {
	t.Parallel()
	now := time.Now()
	m := Model{
		Mode:            ModeScheduler,
		SchedulerLoaded: true,
		SchedulerNow:    now,
		SchedulerEntries: []SchedulerPreviewEntry{
			{AgentID: "alpha", Schedule: "@daily", Next: now.Add(time.Hour)},
			{AgentID: "broken", Schedule: "not-a-cron", Err: errors.New("expected exactly 5 fields")},
		},
	}
	out := m.View()
	require.Contains(t, out, "alpha")
	require.Contains(t, out, "broken")
	require.Contains(t, out, "invalid",
		"per-row parse-fail must surface in the pane without wiping good rows")
}

// TestView_SchedulerLoadingPlaceholder — before the preview cmd resolves
// the pane shows "loading…".
func TestView_SchedulerLoadingPlaceholder(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeScheduler, SchedulerLoaded: false}
	out := m.View()
	require.Contains(t, out, "loading")
}

// TestView_SchedulerErrorState — a whole-pane error (e.g. the AgentLister
// List() call failed) renders the error string + the back hint.
func TestView_SchedulerErrorState(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:            ModeScheduler,
		SchedulerLoaded: true,
		SchedulerErr:    errors.New("db locked"),
	}
	out := m.View()
	require.Contains(t, out, "error")
	require.Contains(t, out, "db locked")
	require.Contains(t, out, "esc")
}

// ─── In-app delete confirm (W3-2 follow-on #3) ─────────────────────────

// TestUpdate_DEntersConfirmMode — pressing `d` on a populated list
// transitions to ModeDeleteConfirm and locks PendingDeleteID to the
// cursor row's ID.
func TestUpdate_DEntersConfirmMode(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Loaded: true, Agents: []agent.Agent{
		{ID: "alpha"}, {ID: "beta"},
	}, Cursor: 1}
	next, cmd := m.Update(keyMsg("d"))
	mm := next.(Model)
	require.Equal(t, ModeDeleteConfirm, mm.Mode)
	require.Equal(t, "beta", mm.PendingDeleteID)
	require.Nil(t, cmd, "d alone must not fire delete — confirmation required")
}

// TestUpdate_DOnEmptyListIsNoOp — d with no agents must not switch modes.
func TestUpdate_DOnEmptyListIsNoOp(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Loaded: true}
	next, cmd := m.Update(keyMsg("d"))
	require.Equal(t, ModeList, next.(Model).Mode)
	require.Nil(t, cmd)
}

// TestUpdate_YInConfirmFiresDeleteCmd — y in confirm mode dispatches the
// delete cmd. The cmd, when executed, calls Delete on the store and
// produces AgentDeletedMsg (success path).
func TestUpdate_YInConfirmFiresDeleteCmd(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDeleteConfirm, PendingDeleteID: "beta",
		Agents: []agent.Agent{{ID: "alpha"}, {ID: "beta"}}, Cursor: 1}
	next, cmd := m.Update(keyMsg("y"))
	require.NotNil(t, cmd, "y must schedule a delete command")
	msg := cmd()
	deleted, ok := msg.(AgentDeletedMsg)
	require.True(t, ok, "expected AgentDeletedMsg, got %T", msg)
	require.Equal(t, "beta", deleted.ID)
	require.Equal(t, []string{"beta"}, store.deleted)
	// Mode is allowed to stay in ModeDeleteConfirm until the message lands;
	// what matters is that the cmd was dispatched.
	_ = next
}

// TestUpdate_YInConfirmFiresDeleteCmd_Error — y + Delete() returning err
// produces AgentDeleteErrMsg.
func TestUpdate_YInConfirmFiresDeleteCmd_Error(t *testing.T) {
	t.Parallel()
	bang := errors.New("disk full")
	store := &fakeLister{deleteErr: bang}
	m := Model{Store: store, Mode: ModeDeleteConfirm, PendingDeleteID: "alpha",
		Agents: []agent.Agent{{ID: "alpha"}}, Cursor: 0}
	_, cmd := m.Update(keyMsg("y"))
	require.NotNil(t, cmd)
	msg := cmd()
	errMsg, ok := msg.(AgentDeleteErrMsg)
	require.True(t, ok, "expected AgentDeleteErrMsg, got %T", msg)
	require.Equal(t, "alpha", errMsg.ID)
	require.ErrorIs(t, errMsg.Err, bang)
}

// TestUpdate_NInConfirmReturnsToList — n cancels and returns to list mode.
func TestUpdate_NInConfirmReturnsToList(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDeleteConfirm, PendingDeleteID: "beta",
		Agents: []agent.Agent{{ID: "alpha"}, {ID: "beta"}}, Cursor: 1}
	next, cmd := m.Update(keyMsg("n"))
	mm := next.(Model)
	require.Equal(t, ModeList, mm.Mode)
	require.Empty(t, mm.PendingDeleteID, "PendingDeleteID must clear on cancel")
	require.Empty(t, store.deleted, "cancel must not call Delete")
	require.Nil(t, cmd)
}

// TestUpdate_EscInConfirmReturnsToList — esc is also a cancel.
func TestUpdate_EscInConfirmReturnsToList(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDeleteConfirm, PendingDeleteID: "alpha"}
	next, _ := m.Update(keyMsg("esc"))
	require.Equal(t, ModeList, next.(Model).Mode)
	require.Empty(t, store.deleted)
}

// TestUpdate_QuitWorksInConfirmMode — q / ctrl+c still quits.
func TestUpdate_QuitWorksInConfirmMode(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDeleteConfirm, PendingDeleteID: "alpha"}
	_, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	_, ok := cmd().(tea.QuitMsg)
	require.True(t, ok)
}

// TestUpdate_AgentDeletedMsgRefreshesList — AgentDeletedMsg returns the
// user to list mode + schedules a reload (so the deleted row vanishes).
func TestUpdate_AgentDeletedMsgRefreshesList(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDeleteConfirm,
		PendingDeleteID: "alpha"}
	next, cmd := m.Update(AgentDeletedMsg{ID: "alpha"})
	mm := next.(Model)
	require.Equal(t, ModeList, mm.Mode)
	require.Empty(t, mm.PendingDeleteID, "PendingDeleteID must clear after success")
	require.False(t, mm.Loaded, "Loaded resets so the reload placeholder shows")
	require.NotNil(t, cmd, "delete success must schedule a list reload")
}

// TestUpdate_AgentDeleteErrMsgRecordsErrInListMode — failure returns to
// list mode and records the error so View can surface it.
func TestUpdate_AgentDeleteErrMsgRecordsErrInListMode(t *testing.T) {
	t.Parallel()
	bang := errors.New("disk full")
	m := Model{Store: &fakeLister{}, Mode: ModeDeleteConfirm,
		PendingDeleteID: "alpha", Loaded: true,
		Agents: []agent.Agent{{ID: "alpha"}}}
	next, _ := m.Update(AgentDeleteErrMsg{ID: "alpha", Err: bang})
	mm := next.(Model)
	require.Equal(t, ModeList, mm.Mode)
	require.Empty(t, mm.PendingDeleteID)
	require.ErrorIs(t, mm.Err, bang, "delete error must surface as the list Err so the user sees it")
	require.Len(t, mm.Agents, 1, "delete failure must not drop the row optimistically")
}

// TestUpdate_NavKeysInertInConfirmMode — j/k/g/G do not move the cursor
// while a confirm dialog is open.
func TestUpdate_NavKeysInertInConfirmMode(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeDeleteConfirm, Loaded: true,
		Agents: []agent.Agent{{ID: "a"}, {ID: "b"}, {ID: "c"}}, Cursor: 1,
		PendingDeleteID: "b"}
	for _, k := range []string{"j", "k", "G", "g"} {
		next, _ := m.Update(keyMsg(k))
		require.Equal(t, 1, next.(Model).Cursor,
			"%s must not move cursor in confirm mode", k)
		require.Equal(t, ModeDeleteConfirm, next.(Model).Mode)
	}
}

// TestView_ConfirmModeRendersAgentIDAndYNHint — the confirm pane shows
// the target agent ID and the y/N prompt copy.
func TestView_ConfirmModeRendersAgentIDAndYNHint(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:            ModeDeleteConfirm,
		Loaded:          true,
		Agents:          []agent.Agent{{ID: "alpha"}, {ID: "beta"}},
		Cursor:          1,
		PendingDeleteID: "beta",
	}
	out := m.View()
	require.Contains(t, out, "beta", "target agent ID must be visible")
	require.Contains(t, out, "delete", "the action word must be visible")
	require.True(t,
		strings.Contains(out, "y") && strings.Contains(out, "N"),
		"y/N hint must be visible: %s", out)
}

// TestView_DeleteErrorShowsInListAfterFailure — once an AgentDeleteErrMsg
// folds in, the list view's error state surfaces the error string.
func TestView_DeleteErrorShowsInListAfterFailure(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeList, Loaded: true,
		Err: errors.New("delete agent \"alpha\": disk full")}
	out := m.View()
	require.Contains(t, out, "error")
	require.Contains(t, out, "disk full")
}
