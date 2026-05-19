package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/agent"
	"github.com/0xmhha/buddy/internal/queries"
	"github.com/0xmhha/buddy/internal/usage"
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

	logsBySince map[int64][]agent.AgentLog // sinceID → canned response
	logsDefault []agent.AgentLog           // fallback when no sinceID match
	logsErr     error                      // forces error from LogsSince
	logsCalls   []logsCall                 // recorded (runID, sinceID) pairs

	updated   []updateCall // ids/specs that UpdateSpec was called with
	updateErr error        // canned error from UpdateSpec

	created   []createCall // ids/specs that Create was called with
	createErr error        // canned error from Create
}

type logsCall struct {
	RunID   int64
	SinceID int64
}

type updateCall struct {
	ID       string
	Name     string
	Schedule string
	YAML     string
}

type createCall struct {
	ID       string
	Name     string
	Schedule string
	YAML     string
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

// LogsSince returns logsBySince[sinceID] if present, otherwise logsDefault.
// logsErr forces an error path. Captures (runID, sinceID) for assertion.
func (f *fakeLister) LogsSince(_ context.Context, runID int64, sinceID int64) ([]agent.AgentLog, error) {
	f.logsCalls = append(f.logsCalls, logsCall{RunID: runID, SinceID: sinceID})
	if f.logsErr != nil {
		return nil, f.logsErr
	}
	if lines, ok := f.logsBySince[sinceID]; ok {
		return lines, nil
	}
	return f.logsDefault, nil
}

// UpdateSpec records the call into updated so tests can assert what was
// saved + optionally returns updateErr. Mirrors Delete's pattern.
func (f *fakeLister) UpdateSpec(_ context.Context, spec agent.AgentSpec, yaml string) error {
	f.updated = append(f.updated, updateCall{ID: spec.ID, Name: spec.Name, Schedule: spec.Schedule, YAML: yaml})
	return f.updateErr
}

// Create records the call into created and returns the canned agent
// (zero-value Agent with ID echoed from spec.ID by default) + optional
// createErr.
func (f *fakeLister) Create(_ context.Context, spec agent.AgentSpec, yaml string) (agent.Agent, error) {
	f.created = append(f.created, createCall{ID: spec.ID, Name: spec.Name, Schedule: spec.Schedule, YAML: yaml})
	if f.createErr != nil {
		return agent.Agent{}, f.createErr
	}
	return agent.Agent{ID: spec.ID, Name: spec.Name, Schedule: spec.Schedule}, nil
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

// ─── Live log tail (W3-2 follow-on #4 — P1-1) ──────────────────────────

// TestUpdate_TInDetailEntersLogTailAndFiresInitialLoad — `t` in detail
// mode switches to ModeLogTail, locks in the detail run's ID, and
// schedules the initial chunk fetch (sinceID=0).
func TestUpdate_TInDetailEntersLogTailAndFiresInitialLoad(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, DetailLoaded: true,
		Selected: "alpha", Detail: agent.AgentRun{ID: 42, AgentID: "alpha"}}
	next, cmd := m.Update(keyMsg("t"))
	mm := next.(Model)
	require.Equal(t, ModeLogTail, mm.Mode)
	require.Equal(t, int64(42), mm.LogTailRunID)
	require.Equal(t, int64(0), mm.LogTailLastID,
		"LogTailLastID resets to 0 on entry so the first poll fetches everything")
	require.False(t, mm.LogTailLoaded)
	require.Empty(t, mm.LogTailLines, "previous tail state must clear on entry")
	require.NotNil(t, cmd, "t in detail must schedule an initial log chunk fetch")
}

// TestUpdate_TInDetailNoRunIsNoOp — `t` in detail when there is no run
// (DetailErr=ErrNotFound, Detail zero) must not switch modes.
func TestUpdate_TInDetailNoRunIsNoOp(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, DetailLoaded: true,
		Selected: "alpha", DetailErr: agent.ErrNotFound}
	next, cmd := m.Update(keyMsg("t"))
	require.Equal(t, ModeDetail, next.(Model).Mode)
	require.Nil(t, cmd)
}

// TestUpdate_LogTailChunkAppendsAndAdvancesLastID — incoming
// LogTailChunkMsg appends to LogTailLines and advances LogTailLastID
// to the highest received ID. Marks LogTailLoaded=true.
func TestUpdate_LogTailChunkAppendsAndAdvancesLastID(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeLogTail, LogTailRunID: 42}
	chunk := []agent.AgentLog{
		{ID: 1, RunID: 42, Level: "info", Message: "first"},
		{ID: 2, RunID: 42, Level: "info", Message: "second"},
	}
	next, _ := m.Update(LogTailChunkMsg{Lines: chunk})
	mm := next.(Model)
	require.True(t, mm.LogTailLoaded)
	require.Len(t, mm.LogTailLines, 2)
	require.Equal(t, int64(2), mm.LogTailLastID)

	// Second chunk appends, advances watermark.
	more := []agent.AgentLog{
		{ID: 3, RunID: 42, Level: "info", Message: "third"},
	}
	next, _ = mm.Update(LogTailChunkMsg{Lines: more})
	mm = next.(Model)
	require.Len(t, mm.LogTailLines, 3)
	require.Equal(t, int64(3), mm.LogTailLastID)
}

// TestUpdate_LogTailChunkEmptyIsHarmless — an empty chunk must not move
// LastID or reset Loaded. Common case: poll fired but no new lines yet.
func TestUpdate_LogTailChunkEmptyIsHarmless(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeLogTail, LogTailRunID: 42,
		LogTailLastID: 7, LogTailLoaded: true,
		LogTailLines: []agent.AgentLog{{ID: 7, Message: "prior"}}}
	next, _ := m.Update(LogTailChunkMsg{Lines: nil})
	mm := next.(Model)
	require.True(t, mm.LogTailLoaded)
	require.Equal(t, int64(7), mm.LogTailLastID)
	require.Len(t, mm.LogTailLines, 1)
}

// TestUpdate_LogTailErrFoldsIntoState — the err msg records the error
// and flips LogTailLoaded=true so View knows the fetch resolved.
func TestUpdate_LogTailErrFoldsIntoState(t *testing.T) {
	t.Parallel()
	bang := errors.New("db locked")
	m := Model{Mode: ModeLogTail, LogTailRunID: 42}
	next, _ := m.Update(LogTailErrMsg{Err: bang})
	mm := next.(Model)
	require.True(t, mm.LogTailLoaded)
	require.ErrorIs(t, mm.LogTailErr, bang)
}

// TestUpdate_LogTailTickFiresLoadCmd — LogTailTickMsg in tail mode
// dispatches a load cmd with the current LogTailLastID as sinceID.
func TestUpdate_LogTailTickFiresLoadCmd(t *testing.T) {
	t.Parallel()
	store := &fakeLister{logsDefault: nil}
	m := Model{Store: store, Mode: ModeLogTail, LogTailRunID: 42,
		LogTailLastID: 5, LogTailLoaded: true}
	_, cmd := m.Update(LogTailTickMsg{})
	require.NotNil(t, cmd)
	// Execute the cmd; it must call LogsSince(42, 5).
	msg := cmd()
	_, ok := msg.(LogTailChunkMsg)
	require.True(t, ok, "expected LogTailChunkMsg, got %T", msg)
	require.Len(t, store.logsCalls, 1)
	require.Equal(t, int64(42), store.logsCalls[0].RunID)
	require.Equal(t, int64(5), store.logsCalls[0].SinceID,
		"tick must use the high-water mark as sinceID")
}

// TestUpdate_LogTailTickInOtherModeIsInert — a tick that arrives after
// the user pressed esc must NOT dispatch a load cmd (self-cancelling).
func TestUpdate_LogTailTickInOtherModeIsInert(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, LogTailRunID: 42,
		LogTailLastID: 5}
	_, cmd := m.Update(LogTailTickMsg{})
	require.Nil(t, cmd, "stale tick after mode change must not poll")
	require.Empty(t, store.logsCalls)
}

// TestUpdate_RInLogTailRefetches — `r` in tail mode fires an immediate
// load cmd using the current LastID.
func TestUpdate_RInLogTailRefetches(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeLogTail, LogTailRunID: 42,
		LogTailLastID: 9, LogTailLoaded: true}
	_, cmd := m.Update(keyMsg("r"))
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(LogTailChunkMsg)
	require.True(t, ok)
	require.Equal(t, int64(9), store.logsCalls[0].SinceID)
}

// TestUpdate_EscReturnsFromLogTailToDetail — esc returns to the detail
// pane (not the list) so the user keeps the same agent context.
func TestUpdate_EscReturnsFromLogTailToDetail(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeLogTail, LogTailRunID: 42,
		Selected: "alpha", DetailLoaded: true}
	next, _ := m.Update(keyMsg("esc"))
	require.Equal(t, ModeDetail, next.(Model).Mode)
}

// TestUpdate_HReturnsFromLogTailToDetail — vi-style 'h' alias.
func TestUpdate_HReturnsFromLogTailToDetail(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeLogTail, LogTailRunID: 42}
	next, _ := m.Update(keyMsg("h"))
	require.Equal(t, ModeDetail, next.(Model).Mode)
}

// TestUpdate_QuitWorksInLogTailMode — q quits from tail.
func TestUpdate_QuitWorksInLogTailMode(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Mode: ModeLogTail}
	_, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	_, ok := cmd().(tea.QuitMsg)
	require.True(t, ok)
}

// TestView_LogTailRendersLines — pane shows run id + each line's message,
// plus the back/refresh footer hint.
func TestView_LogTailRendersLines(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:          ModeLogTail,
		LogTailRunID:  42,
		LogTailLoaded: true,
		LogTailLines: []agent.AgentLog{
			{ID: 1, Level: "info", Message: "agent alpha starting"},
			{ID: 2, Level: "warn", Message: "step 1 retry 1"},
			{ID: 3, Level: "info", Message: "step 1 complete"},
		},
		Selected: "alpha",
	}
	out := m.View()
	require.Contains(t, out, "42", "run id must be visible in header")
	require.Contains(t, out, "alpha")
	require.Contains(t, out, "agent alpha starting")
	require.Contains(t, out, "step 1 complete")
	require.Contains(t, out, "esc", "back hint visible")
}

// TestView_LogTailEmptyCopy — Loaded with 0 lines shows a friendly hint.
func TestView_LogTailEmptyCopy(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeLogTail, LogTailRunID: 42, LogTailLoaded: true,
		Selected: "alpha"}
	out := m.View()
	require.Contains(t, out, "no log lines",
		"empty-state copy must explain the absence rather than render blank")
}

// TestView_LogTailLoadingPlaceholder — pre-fetch placeholder.
func TestView_LogTailLoadingPlaceholder(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeLogTail, LogTailRunID: 42, Selected: "alpha"}
	out := m.View()
	require.Contains(t, out, "loading")
}

// TestView_LogTailErrorState — whole-pane error renders the err string.
func TestView_LogTailErrorState(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeLogTail, LogTailRunID: 42, LogTailLoaded: true,
		Selected: "alpha", LogTailErr: errors.New("db locked")}
	out := m.View()
	require.Contains(t, out, "error")
	require.Contains(t, out, "db locked")
}

// ─── In-app edit (W3-2 follow-on #5 — P1-2) ────────────────────────────

// minimalEditableYAML mirrors internal/agent's minimalSpecYAML — kept
// local so the TUI test file doesn't have to reach into a sibling
// package's unexported constant.
const minimalEditableYAML = `
id: alpha
name: "Alpha agent"
chain:
  - command: status
`

// TestSaveEditedSpec_ValidYAMLDispatchesUpdate — the save cmd parses the
// content, sees the id matches the original, calls UpdateSpec, and emits
// AgentSpecUpdatedMsg. fakeLister.updated records what was passed.
func TestSaveEditedSpec_ValidYAMLDispatchesUpdate(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	cmd := saveEditedSpecCmd(store, "alpha", []byte(minimalEditableYAML))
	require.NotNil(t, cmd)
	msg := cmd()
	updated, ok := msg.(AgentSpecUpdatedMsg)
	require.True(t, ok, "expected AgentSpecUpdatedMsg, got %T", msg)
	require.Equal(t, "alpha", updated.ID)
	require.Len(t, store.updated, 1)
	require.Equal(t, "alpha", store.updated[0].ID)
	require.Equal(t, "Alpha agent", store.updated[0].Name)
}

// TestSaveEditedSpec_RenameIsRejected — the cmd refuses to call
// UpdateSpec when the parsed spec.ID differs from the original.
// Renames are out of scope for the edit flow (would orphan runs/logs).
func TestSaveEditedSpec_RenameIsRejected(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	renamed := []byte(`
id: alpha-renamed
name: "Alpha agent"
chain:
  - command: status
`)
	cmd := saveEditedSpecCmd(store, "alpha", renamed)
	msg := cmd()
	errMsg, ok := msg.(AgentSpecUpdateErrMsg)
	require.True(t, ok, "expected AgentSpecUpdateErrMsg, got %T", msg)
	require.Equal(t, "alpha", errMsg.ID, "ID in err must be the ORIGINAL (so the user can find it)")
	require.Contains(t, errMsg.Err.Error(), "rename")
	require.Empty(t, store.updated, "rename rejection must NOT call UpdateSpec")
}

// TestSaveEditedSpec_InvalidYAMLReturnsErr — ParseSpec failure short-
// circuits before UpdateSpec is called.
func TestSaveEditedSpec_InvalidYAMLReturnsErr(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	cmd := saveEditedSpecCmd(store, "alpha", []byte("not :: valid yaml ::"))
	msg := cmd()
	errMsg, ok := msg.(AgentSpecUpdateErrMsg)
	require.True(t, ok)
	require.Equal(t, "alpha", errMsg.ID)
	require.Error(t, errMsg.Err)
	require.Empty(t, store.updated)
}

// TestSaveEditedSpec_StoreErrPropagates — UpdateSpec failure surfaces as
// AgentSpecUpdateErrMsg with the wrapped error.
func TestSaveEditedSpec_StoreErrPropagates(t *testing.T) {
	t.Parallel()
	bang := errors.New("disk full")
	store := &fakeLister{updateErr: bang}
	cmd := saveEditedSpecCmd(store, "alpha", []byte(minimalEditableYAML))
	msg := cmd()
	errMsg, ok := msg.(AgentSpecUpdateErrMsg)
	require.True(t, ok)
	require.ErrorIs(t, errMsg.Err, bang)
}

// TestUpdate_EInDetailWithSpecDispatchesEditCmd — e in detail mode when
// the selected agent has spec_yaml dispatches a cmd. We can't verify
// the inner tea.ExecProcess without spinning a real editor; what we
// assert is the cmd is non-nil + EditErr is cleared (retry after a
// prior failed edit).
func TestUpdate_EInDetailWithSpecDispatchesEditCmd(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, DetailLoaded: true,
		Selected: "alpha",
		Agents: []agent.Agent{
			{ID: "alpha", SpecYAML: minimalEditableYAML, Name: "Alpha"},
		},
		EditErr: errors.New("prior error"),
	}
	next, cmd := m.Update(keyMsg("e"))
	mm := next.(Model)
	require.NotNil(t, cmd, "e must dispatch the edit cmd")
	require.Nil(t, mm.EditErr, "EditErr must clear when the user retries")
	require.Equal(t, ModeDetail, mm.Mode, "mode stays Detail; editor shell-out is opaque to the model")
}

// TestUpdate_EInDetailWithoutAgentInListIsNoOp — if Selected isn't found
// in m.Agents (rare — list shrank between detail entry and the e
// keypress) we don't dispatch the edit cmd.
func TestUpdate_EInDetailWithoutAgentInListIsNoOp(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, DetailLoaded: true,
		Selected: "alpha",
		Agents:   []agent.Agent{{ID: "beta", SpecYAML: "..."}}, // alpha gone
	}
	next, cmd := m.Update(keyMsg("e"))
	require.Nil(t, cmd)
	require.Equal(t, ModeDetail, next.(Model).Mode)
}

// TestUpdate_EditorExitedWithErrorRecordsErr — when EditorExitedMsg.Err
// is non-nil (editor crash, file IO fail, etc.), the model records it
// in EditErr and does NOT dispatch a save cmd.
func TestUpdate_EditorExitedWithErrorRecordsErr(t *testing.T) {
	t.Parallel()
	bang := errors.New("editor crashed")
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, Selected: "alpha"}
	next, cmd := m.Update(EditorExitedMsg{AgentID: "alpha", Err: bang})
	mm := next.(Model)
	require.ErrorIs(t, mm.EditErr, bang)
	require.Nil(t, cmd, "editor error must not chain a save cmd")
	require.Empty(t, store.updated)
}

// TestUpdate_EditorExitedWithContentDispatchesSaveCmd — successful editor
// exit (Content non-nil, Err nil) dispatches the save cmd.
func TestUpdate_EditorExitedWithContentDispatchesSaveCmd(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, Selected: "alpha"}
	next, cmd := m.Update(EditorExitedMsg{
		AgentID: "alpha",
		Content: []byte(minimalEditableYAML),
	})
	require.NotNil(t, cmd, "non-nil content must dispatch save cmd")
	// Execute the save cmd; assert AgentSpecUpdatedMsg comes out.
	msg := cmd()
	_, ok := msg.(AgentSpecUpdatedMsg)
	require.True(t, ok, "expected AgentSpecUpdatedMsg, got %T", msg)
	require.Len(t, store.updated, 1)
	_ = next
}

// TestUpdate_AgentSpecUpdatedMsgReloadsList — success returns to a
// fresh-list state (Loaded=false) and dispatches loadAgentsCmd so the
// edited spec's new name/schedule reflects in the list row.
func TestUpdate_AgentSpecUpdatedMsgReloadsList(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Mode: ModeDetail, Loaded: true,
		Selected: "alpha", EditErr: errors.New("prior")}
	next, cmd := m.Update(AgentSpecUpdatedMsg{ID: "alpha"})
	mm := next.(Model)
	require.Nil(t, mm.EditErr, "success must clear prior edit error")
	require.False(t, mm.Loaded, "Loaded resets so the reload placeholder shows")
	require.NotNil(t, cmd, "must schedule a list reload")
}

// TestUpdate_AgentSpecUpdateErrMsgRecordsErr — store err surfaces as
// EditErr; mode stays Detail; list is not touched.
func TestUpdate_AgentSpecUpdateErrMsgRecordsErr(t *testing.T) {
	t.Parallel()
	bang := errors.New("disk full")
	m := Model{Mode: ModeDetail, Loaded: true, Selected: "alpha",
		Agents: []agent.Agent{{ID: "alpha"}}}
	next, _ := m.Update(AgentSpecUpdateErrMsg{ID: "alpha", Err: bang})
	mm := next.(Model)
	require.ErrorIs(t, mm.EditErr, bang)
	require.Equal(t, ModeDetail, mm.Mode)
	require.True(t, mm.Loaded, "list state must NOT be reset on save failure")
	require.Len(t, mm.Agents, 1, "list rows must stay intact")
}

// TestView_DetailRendersEditErrorBanner — when EditErr is set, the detail
// pane shows an error banner with the message + a retry hint.
func TestView_DetailRendersEditErrorBanner(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:         ModeDetail,
		Loaded:       true,
		DetailLoaded: true,
		Agents:       []agent.Agent{{ID: "alpha"}},
		Selected:     "alpha",
		DetailErr:    agent.ErrNotFound, // empty-state body so we can focus on the banner
		EditErr:      errors.New("rename not allowed: spec id \"beta\" != original \"alpha\""),
	}
	out := m.View()
	require.Contains(t, out, "edit error", "banner label must be visible")
	require.Contains(t, out, "rename not allowed")
	require.Contains(t, out, "press e", "retry hint must guide the user")
}

// ─── Create form (W3-2 follow-on #6 — P3-1) ────────────────────────────

// TestSaveNewSpec_ValidYAMLDispatchesCreate — successful parse + Store
// .Create → AgentCreatedMsg with the new agent's ID.
func TestSaveNewSpec_ValidYAMLDispatchesCreate(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	cmd := saveNewSpecCmd(store, []byte(minimalEditableYAML))
	require.NotNil(t, cmd)
	msg := cmd()
	created, ok := msg.(AgentCreatedMsg)
	require.True(t, ok, "expected AgentCreatedMsg, got %T", msg)
	require.Equal(t, "alpha", created.ID)
	require.Len(t, store.created, 1)
	require.Equal(t, "alpha", store.created[0].ID)
}

// TestSaveNewSpec_InvalidYAMLReturnsErr — ParseSpec failure short-
// circuits before Store.Create is called.
func TestSaveNewSpec_InvalidYAMLReturnsErr(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	cmd := saveNewSpecCmd(store, []byte("not :: valid yaml ::"))
	msg := cmd()
	errMsg, ok := msg.(AgentCreateErrMsg)
	require.True(t, ok)
	require.Error(t, errMsg.Err)
	require.Empty(t, store.created)
}

// TestSaveNewSpec_StoreErrPropagates — duplicate-id (SQLite UNIQUE) or
// other Store.Create errors surface as AgentCreateErrMsg.
func TestSaveNewSpec_StoreErrPropagates(t *testing.T) {
	t.Parallel()
	bang := errors.New("UNIQUE constraint failed: agents.id")
	store := &fakeLister{createErr: bang}
	cmd := saveNewSpecCmd(store, []byte(minimalEditableYAML))
	msg := cmd()
	errMsg, ok := msg.(AgentCreateErrMsg)
	require.True(t, ok)
	require.ErrorIs(t, errMsg.Err, bang)
}

// TestUpdate_CInListDispatchesCreateCmd — c in list mode dispatches the
// create cmd (which launches the editor; we can't drive it but we can
// assert the cmd is non-nil and m.Err is cleared).
func TestUpdate_CInListDispatchesCreateCmd(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Loaded: true,
		Err: errors.New("prior list err")}
	next, cmd := m.Update(keyMsg("c"))
	mm := next.(Model)
	require.NotNil(t, cmd, "c must dispatch the create cmd")
	require.Nil(t, mm.Err, "Err must clear before going to the editor")
	require.Equal(t, ModeList, mm.Mode, "mode stays List; editor shell-out is opaque")
}

// TestUpdate_CInListWorksOnEmptyList — c does NOT depend on cursor; an
// empty list must still allow creation.
func TestUpdate_CInListWorksOnEmptyList(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Loaded: true}
	_, cmd := m.Update(keyMsg("c"))
	require.NotNil(t, cmd, "c must work on an empty list (otherwise users with no agents can't bootstrap)")
}

// TestUpdate_NewSpecEditorExitedWithErrorRecordsErr — editor IO / crash
// surfaces as m.Err with a `create:` wrapper, and does NOT dispatch a
// save cmd.
func TestUpdate_NewSpecEditorExitedWithErrorRecordsErr(t *testing.T) {
	t.Parallel()
	bang := errors.New("editor crashed")
	store := &fakeLister{}
	m := Model{Store: store, Loaded: true}
	next, cmd := m.Update(NewSpecEditorExitedMsg{Err: bang})
	mm := next.(Model)
	require.Nil(t, cmd, "editor error must not chain a save cmd")
	require.ErrorIs(t, mm.Err, bang)
	require.Contains(t, mm.Err.Error(), "create:", "err must be wrapped with create context")
	require.Empty(t, store.created)
}

// TestUpdate_NewSpecEditorExitedWithContentDispatchesSaveCmd — happy
// path: content non-nil + nil err → save cmd dispatched; running it
// emits AgentCreatedMsg.
func TestUpdate_NewSpecEditorExitedWithContentDispatchesSaveCmd(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Loaded: true}
	next, cmd := m.Update(NewSpecEditorExitedMsg{
		Content: []byte(minimalEditableYAML),
	})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(AgentCreatedMsg)
	require.True(t, ok)
	require.Len(t, store.created, 1)
	_ = next
}

// TestUpdate_AgentCreatedMsgReloadsList — success clears m.Err, flips
// Loaded=false, dispatches list reload.
func TestUpdate_AgentCreatedMsgReloadsList(t *testing.T) {
	t.Parallel()
	store := &fakeLister{}
	m := Model{Store: store, Loaded: true, Mode: ModeList,
		Err: errors.New("prior")}
	next, cmd := m.Update(AgentCreatedMsg{ID: "alpha"})
	mm := next.(Model)
	require.Nil(t, mm.Err, "success must clear prior list err")
	require.False(t, mm.Loaded)
	require.NotNil(t, cmd)
}

// TestUpdate_AgentCreateErrMsgRecordsErr — store err surfaces as
// m.Err with create wrapper; list is left alone.
func TestUpdate_AgentCreateErrMsgRecordsErr(t *testing.T) {
	t.Parallel()
	bang := errors.New("UNIQUE constraint failed")
	m := Model{Loaded: true, Mode: ModeList,
		Agents: []agent.Agent{{ID: "existing"}}}
	next, _ := m.Update(AgentCreateErrMsg{Err: bang})
	mm := next.(Model)
	require.ErrorIs(t, mm.Err, bang)
	require.Contains(t, mm.Err.Error(), "create:")
	require.Len(t, mm.Agents, 1, "list rows untouched on create failure")
}

// TestCreateStarterYAML_ParsesAsValidSpec — the starter template the
// user sees on first `c` must round-trip ParseSpec. If we typo the
// template, every fresh save would land in the parse-fail branch and
// users would think edit is broken.
func TestCreateStarterYAML_ParsesAsValidSpec(t *testing.T) {
	t.Parallel()
	spec, err := agent.ParseSpec([]byte(CreateStarterYAML))
	require.NoError(t, err, "starter template must parse as a valid AgentSpec")
	require.Equal(t, "new-agent", spec.ID)
	require.NotEmpty(t, spec.Chain, "starter template must contain at least one chain step")
}

// ─── Hook stats pane (A-3.2 W3-5 follow-on) ────────────────────────────

// fakeStatsFetcher returns canned rows + records the window arg.
func fakeStatsFetcher(rows []queries.Row, err error, captured *[]string) HookStatsFetcher {
	return func(window string) (queries.Result, error) {
		*captured = append(*captured, window)
		if err != nil {
			return queries.Result{}, err
		}
		return queries.Result{WindowLabel: window, Rows: rows}, nil
	}
}

// TestUpdate_HInListWithFetcherEntersHookStats — H with a wired fetcher
// switches to ModeHookStats, defaults the window to "1h", and
// dispatches the fetch cmd.
func TestUpdate_HInListWithFetcherEntersHookStats(t *testing.T) {
	t.Parallel()
	var captured []string
	m := Model{Store: &fakeLister{}, Loaded: true,
		HookStatsFetcher: fakeStatsFetcher(nil, nil, &captured)}
	next, cmd := m.Update(keyMsg("H"))
	mm := next.(Model)
	require.Equal(t, ModeHookStats, mm.Mode)
	require.Equal(t, "1h", mm.HookStatsWindow, "default window is 1h")
	require.False(t, mm.HookStatsLoaded)
	require.NotNil(t, cmd, "H must schedule a stats fetch when fetcher is wired")
	// Execute the cmd; assert HookStatsLoadedMsg comes out.
	msg := cmd()
	loaded, ok := msg.(HookStatsLoadedMsg)
	require.True(t, ok, "expected HookStatsLoadedMsg, got %T", msg)
	require.Equal(t, "1h", loaded.Window)
	require.Equal(t, []string{"1h"}, captured, "fetcher must be called once with the locked-in window")
}

// TestUpdate_HInListWithoutFetcherIsNoOp — without a fetcher, H stays
// in list mode and dispatches nothing.
func TestUpdate_HInListWithoutFetcherIsNoOp(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Loaded: true}
	next, cmd := m.Update(keyMsg("H"))
	require.Equal(t, ModeList, next.(Model).Mode, "no fetcher → no mode switch")
	require.Nil(t, cmd)
}

// TestUpdate_HookStatsLoadedFoldsIntoState — the loaded msg populates
// rows/window and flips Loaded=true.
func TestUpdate_HookStatsLoadedFoldsIntoState(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeHookStats}
	rows := []queries.Row{{HookName: "pre-commit", ToolName: "Bash", Count: 10, Failures: 1, P50Ms: 50, P95Ms: 200}}
	next, _ := m.Update(HookStatsLoadedMsg{Window: "1h",
		Result: queries.Result{WindowLabel: "1시간", Rows: rows}})
	mm := next.(Model)
	require.True(t, mm.HookStatsLoaded)
	require.Equal(t, "1h", mm.HookStatsWindow)
	require.Len(t, mm.HookStatsResult.Rows, 1)
	require.Nil(t, mm.HookStatsErr, "successful load clears any prior err")
}

// TestUpdate_HookStatsErrFoldsIntoState — err msg records the error
// and flips Loaded=true.
func TestUpdate_HookStatsErrFoldsIntoState(t *testing.T) {
	t.Parallel()
	bang := errors.New("db locked")
	m := Model{Mode: ModeHookStats}
	next, _ := m.Update(HookStatsErrMsg{Err: bang})
	mm := next.(Model)
	require.True(t, mm.HookStatsLoaded)
	require.ErrorIs(t, mm.HookStatsErr, bang)
}

// TestUpdate_EscFromHookStatsReturnsToList — esc/h return to list.
func TestUpdate_EscFromHookStatsReturnsToList(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeHookStats}
	next, _ := m.Update(keyMsg("esc"))
	require.Equal(t, ModeList, next.(Model).Mode)

	m2 := Model{Mode: ModeHookStats}
	next2, _ := m2.Update(keyMsg("h"))
	require.Equal(t, ModeList, next2.(Model).Mode)
}

// TestUpdate_QuitWorksInHookStatsMode — q quits.
func TestUpdate_QuitWorksInHookStatsMode(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeHookStats}
	_, cmd := m.Update(keyMsg("q"))
	require.NotNil(t, cmd)
	_, ok := cmd().(tea.QuitMsg)
	require.True(t, ok)
}

// TestUpdate_RInHookStatsRefetches — r in the pane refires the fetch
// cmd with the locked-in window.
func TestUpdate_RInHookStatsRefetches(t *testing.T) {
	t.Parallel()
	var captured []string
	m := Model{Mode: ModeHookStats, HookStatsWindow: "5m",
		HookStatsLoaded: true,
		HookStatsFetcher: fakeStatsFetcher(nil, nil, &captured)}
	next, cmd := m.Update(keyMsg("r"))
	mm := next.(Model)
	require.NotNil(t, cmd)
	require.False(t, mm.HookStatsLoaded)
	msg := cmd()
	require.IsType(t, HookStatsLoadedMsg{}, msg)
	require.Equal(t, []string{"5m"}, captured, "refresh must preserve the current window")
}

// TestView_HookStatsRendersRows — pane shows window label + each row's
// hook name + counts + p95.
func TestView_HookStatsRendersRows(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:            ModeHookStats,
		HookStatsLoaded: true,
		HookStatsWindow: "1h",
		HookStatsResult: queries.Result{
			WindowLabel: "1시간",
			Rows: []queries.Row{
				{HookName: "PreToolUse", ToolName: "Bash", Count: 42, Failures: 2, P50Ms: 80, P95Ms: 350},
				{HookName: "Stop", ToolName: "", Count: 7, Failures: 0, P50Ms: 5, P95Ms: 12},
			},
		},
	}
	out := m.View()
	require.Contains(t, out, "1h", "window label visible")
	require.Contains(t, out, "PreToolUse")
	require.Contains(t, out, "Bash")
	require.Contains(t, out, "Stop")
	require.Contains(t, out, "350", "p95 visible")
	require.Contains(t, out, "esc", "back hint visible")
}

// TestView_HookStatsEmptyCopy — Loaded with no rows shows friendly hint.
func TestView_HookStatsEmptyCopy(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeHookStats, HookStatsLoaded: true, HookStatsWindow: "1h"}
	out := m.View()
	require.Contains(t, out, "no hook events")
}

// TestView_HookStatsLoadingPlaceholder — pre-fetch placeholder.
func TestView_HookStatsLoadingPlaceholder(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeHookStats, HookStatsWindow: "1h"}
	out := m.View()
	require.Contains(t, out, "loading")
}

// TestView_HookStatsErrorState — error string surfaces.
func TestView_HookStatsErrorState(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeHookStats, HookStatsLoaded: true, HookStatsWindow: "1h",
		HookStatsErr: errors.New("db locked")}
	out := m.View()
	require.Contains(t, out, "error")
	require.Contains(t, out, "db locked")
}

// ─── Usage pane (W7-2 / ADR-013) ───────────────────────────────────────

func fakeUsageFetcher(ov usage.Overview, err error, calls *int) UsageFetcher {
	return func() (usage.Overview, error) {
		*calls++
		if err != nil {
			return usage.Overview{}, err
		}
		return ov, nil
	}
}

// TestUpdate_UInListWithFetcherEntersUsage — U switches to ModeUsage
// and dispatches the fetch cmd when a fetcher is wired.
func TestUpdate_UInListWithFetcherEntersUsage(t *testing.T) {
	t.Parallel()
	calls := 0
	ov := usage.Overview{Spend: usage.TokenSpend{InputTokens: 100}}
	m := Model{Store: &fakeLister{}, Loaded: true,
		UsageFetcher: fakeUsageFetcher(ov, nil, &calls)}
	next, cmd := m.Update(keyMsg("U"))
	mm := next.(Model)
	require.Equal(t, ModeUsage, mm.Mode)
	require.False(t, mm.UsageLoaded)
	require.NotNil(t, cmd, "U must schedule a usage fetch when fetcher is wired")
	msg := cmd()
	loaded, ok := msg.(UsageLoadedMsg)
	require.True(t, ok, "expected UsageLoadedMsg, got %T", msg)
	require.Equal(t, int64(100), loaded.Result.Spend.InputTokens)
	require.Equal(t, 1, calls)
}

// TestUpdate_UInListWithoutFetcherIsNoOp — no fetcher → no switch.
func TestUpdate_UInListWithoutFetcherIsNoOp(t *testing.T) {
	t.Parallel()
	m := Model{Store: &fakeLister{}, Loaded: true}
	next, cmd := m.Update(keyMsg("U"))
	require.Equal(t, ModeList, next.(Model).Mode)
	require.Nil(t, cmd)
}

// TestUpdate_UsageLoadedFoldsIntoState — loaded msg populates state.
func TestUpdate_UsageLoadedFoldsIntoState(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeUsage}
	ov := usage.Overview{Spend: usage.TokenSpend{OutputTokens: 42}}
	next, _ := m.Update(UsageLoadedMsg{Result: ov})
	mm := next.(Model)
	require.True(t, mm.UsageLoaded)
	require.Equal(t, int64(42), mm.UsageResult.Spend.OutputTokens)
	require.Nil(t, mm.UsageErr)
}

// TestUpdate_UsageErrFoldsIntoState — err msg records and flips Loaded.
func TestUpdate_UsageErrFoldsIntoState(t *testing.T) {
	t.Parallel()
	bang := errors.New("sessions table missing")
	m := Model{Mode: ModeUsage}
	next, _ := m.Update(UsageErrMsg{Err: bang})
	mm := next.(Model)
	require.True(t, mm.UsageLoaded)
	require.ErrorIs(t, mm.UsageErr, bang)
}

// TestUpdate_EscFromUsageReturnsToList — esc/h navigate back.
func TestUpdate_EscFromUsageReturnsToList(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeUsage}
	next, _ := m.Update(keyMsg("esc"))
	require.Equal(t, ModeList, next.(Model).Mode)

	m2 := Model{Mode: ModeUsage}
	next2, _ := m2.Update(keyMsg("h"))
	require.Equal(t, ModeList, next2.(Model).Mode)
}

// TestView_UsageRendersOverview — happy path render shows token + stats.
func TestView_UsageRendersOverview(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:        ModeUsage,
		UsageLoaded: true,
		UsageFetcher: func() (usage.Overview, error) { return usage.Overview{}, nil },
		UsageResult: usage.Overview{
			Spend: usage.TokenSpend{InputTokens: 123, OutputTokens: 456, CacheReadTokens: 789},
			Stats: usage.SessionStats{TotalSessions: 7, ActiveSessions: 2, EndedSessions: 5},
		},
	}
	out := m.View()
	require.Contains(t, out, "토큰 사용량")
	require.Contains(t, out, "세션 통계")
	require.Contains(t, out, "esc")
}

// TestView_UsageUnavailableWithoutFetcher — no fetcher renders friendly note.
func TestView_UsageUnavailableWithoutFetcher(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeUsage}
	out := m.View()
	require.Contains(t, out, "UsageFetcher 미설정")
}

// ─── Advisor section (W7-3b / ADR-015) ────────────────────────────────

func fakeAdvisorFetcher(advs []advisor.Advisory, err error, calls *int) AdvisorFetcher {
	return func() ([]advisor.Advisory, error) {
		*calls++
		return advs, err
	}
}

// TestUpdate_AdvisorLoadedFoldsIntoState — loaded msg populates results.
func TestUpdate_AdvisorLoadedFoldsIntoState(t *testing.T) {
	t.Parallel()
	m := Model{Mode: ModeUsage}
	adv := []advisor.Advisory{{Kind: advisor.KindTokenSpikeDay, Severity: advisor.SeverityWarn, Message: "spike"}}
	next, _ := m.Update(AdvisorLoadedMsg{Result: adv})
	mm := next.(Model)
	require.True(t, mm.AdvisorLoaded)
	require.Len(t, mm.AdvisorResult, 1)
	require.Nil(t, mm.AdvisorErr)
}

// TestUpdate_AdvisorErrFoldsIntoState — err msg records and flips Loaded.
func TestUpdate_AdvisorErrFoldsIntoState(t *testing.T) {
	t.Parallel()
	bang := errors.New("advisor down")
	m := Model{Mode: ModeUsage}
	next, _ := m.Update(AdvisorErrMsg{Err: bang})
	mm := next.(Model)
	require.True(t, mm.AdvisorLoaded)
	require.ErrorIs(t, mm.AdvisorErr, bang)
}

// TestUpdate_UInListAlsoDispatchesAdvisorWhenFetcherSet — pressing U
// fires both usage + advisor loaders when both fetchers are wired.
func TestUpdate_UInListAlsoDispatchesAdvisorWhenFetcherSet(t *testing.T) {
	t.Parallel()
	usageCalls, advisorCalls := 0, 0
	m := Model{Store: &fakeLister{}, Loaded: true,
		UsageFetcher:   fakeUsageFetcher(usage.Overview{}, nil, &usageCalls),
		AdvisorFetcher: fakeAdvisorFetcher(nil, nil, &advisorCalls),
	}
	next, cmd := m.Update(keyMsg("U"))
	require.Equal(t, ModeUsage, next.(Model).Mode)
	require.NotNil(t, cmd)
	// tea.Batch returns a Cmd that, when executed, fans out the children
	// as BatchMsg containing more Cmds. We can just execute it and
	// observe the resulting msg.
	msg := cmd()
	_, ok := msg.(tea.BatchMsg)
	require.True(t, ok, "U must produce a Batch cmd combining loaders")
}

// TestView_UsageRendersAdvisorSection — advisor advisories appear in
// the Usage pane render when fetcher is wired + result loaded.
func TestView_UsageRendersAdvisorSection(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:        ModeUsage,
		UsageLoaded: true,
		UsageFetcher: func() (usage.Overview, error) { return usage.Overview{}, nil },
		AdvisorFetcher: func() ([]advisor.Advisory, error) { return nil, nil },
		AdvisorLoaded: true,
		AdvisorResult: []advisor.Advisory{
			{Kind: advisor.KindTokenSpikeDay, Severity: advisor.SeverityWarn, Message: "오늘 토큰 2x"},
		},
	}
	out := m.View()
	require.Contains(t, out, "조언")
	require.Contains(t, out, "token-spike-day")
	require.Contains(t, out, "오늘 토큰 2x")
}

// TestView_UsageAdvisorSectionEmpty — no advisories yet renders soft note.
func TestView_UsageAdvisorSectionEmpty(t *testing.T) {
	t.Parallel()
	m := Model{
		Mode:        ModeUsage,
		UsageLoaded: true,
		UsageFetcher: func() (usage.Overview, error) { return usage.Overview{}, nil },
		AdvisorFetcher: func() ([]advisor.Advisory, error) { return nil, nil },
		AdvisorLoaded:  true,
	}
	out := m.View()
	require.Contains(t, out, "조언")
	require.Contains(t, out, "지금은 알릴 조언이 없어")
}
