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
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/0xmhha/buddy/internal/agent"
)

// AgentLister is the Store surface the TUI needs. Narrowing the
// dependency to a small interface keeps Model tests decoupled from
// SQLite and makes it easy to inject canned data in unit tests.
//
// Despite the historical name, the interface now covers a non-trivial
// in-app mutation (Delete). Renaming to "AgentStore" would be cleaner
// but breaks every external caller — left as-is per cli-buddy-spec's
// "additive only until v1.0" stance.
//
// LatestRun mirrors the Store method of the same name. Callers must
// receive agent.ErrNotFound when the agent has never run — the View
// renders that case as the friend-tone "no runs yet" copy rather than a
// generic error.
//
// Delete mirrors Store.Delete: it removes the agent (FK cascade drops
// runs + logs). Returns agent.ErrNotFound when the ID isn't there.
type AgentLister interface {
	List(ctx context.Context) ([]agent.Agent, error)
	LatestRun(ctx context.Context, agentID string) (agent.AgentRun, error)
	Delete(ctx context.Context, agentID string) error
	LogsSince(ctx context.Context, runID int64, sinceLogID int64) ([]agent.AgentLog, error)
	UpdateSpec(ctx context.Context, spec agent.AgentSpec, yaml string) error
	Create(ctx context.Context, spec agent.AgentSpec, yaml string) (agent.Agent, error)
}

// Mode is the top-level view state: list (default), detail, or the
// scheduler-preview pane.
type Mode int

const (
	ModeList Mode = iota
	ModeDetail
	ModeScheduler
	ModeDeleteConfirm
	ModeLogTail
)

// logTailPollInterval is the cadence at which the log-tail pane polls
// Store.LogsSince for new lines. Keep it modest — 1s gives sub-second
// "feels live" perception without hammering the DB. Exposed as a var
// so future tests can monkey-patch it without exposing a Model field.
var logTailPollInterval = time.Second

// SchedulerPreviewEntry is one row in the scheduler-status pane. It is
// computed by walking the agent list and asking agent.PreviewSchedule to
// turn each non-empty cron string into a next-fire time. Per-row Err
// surfaces parse failures inline so a single typo'd spec doesn't blank
// the whole pane.
type SchedulerPreviewEntry struct {
	AgentID  string
	Schedule string
	Next     time.Time
	Err      error
}

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

	// Scheduler-pane state — only meaningful when Mode==ModeScheduler.
	SchedulerNow     time.Time
	SchedulerEntries []SchedulerPreviewEntry
	SchedulerErr     error
	SchedulerLoaded  bool

	// Delete-confirm state — only meaningful when Mode==ModeDeleteConfirm.
	// PendingDeleteID is the agent ID the user is about to delete; it is
	// captured when `d` is pressed so a list refresh that lands while the
	// dialog is open does not retarget the deletion to a different row.
	PendingDeleteID string

	// Log-tail state — only meaningful when Mode==ModeLogTail.
	//
	// LogTailRunID is the agent_runs.id row we're tailing (captured on
	// entry from the detail pane). LogTailLastID is the high-water mark
	// of agent_logs.id we've already accumulated; the next poll passes
	// it as sinceID so the DB returns only new lines.
	LogTailRunID  int64
	LogTailLastID int64
	LogTailLines  []agent.AgentLog
	LogTailErr    error
	LogTailLoaded bool

	// Edit-flow state. EditErr stashes the most recent edit failure
	// (editor crash, parse fail, rename rejected, store err); the detail
	// pane renders it as a banner with a retry hint. Cleared on the next
	// `e` keypress (about to retry) or on a successful save.
	EditErr error
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
	AgentsLoadedMsg          struct{ Agents []agent.Agent }
	ErrMsg                   struct{ Err error }
	AgentDetailLoadedMsg     struct{ Run agent.AgentRun }
	AgentDetailErrMsg        struct{ Err error }
	SchedulerStatusLoadedMsg struct {
		Now     time.Time
		Entries []SchedulerPreviewEntry
	}
	SchedulerStatusErrMsg struct{ Err error }
	AgentDeletedMsg       struct{ ID string }
	AgentDeleteErrMsg     struct {
		ID  string
		Err error
	}
	LogTailChunkMsg struct{ Lines []agent.AgentLog }
	LogTailErrMsg   struct{ Err error }
	LogTailTickMsg  struct{}
	EditorExitedMsg struct {
		AgentID string
		Content []byte
		Err     error
	}
	AgentSpecUpdatedMsg   struct{ ID string }
	AgentSpecUpdateErrMsg struct {
		ID  string
		Err error
	}
	NewSpecEditorExitedMsg struct {
		Content []byte
		Err     error
	}
	AgentCreatedMsg     struct{ ID string }
	AgentCreateErrMsg   struct{ Err error }
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

// beginEditCmd writes the agent's current spec_yaml to a temp file,
// launches $EDITOR (fall back to $VISUAL then `vi`) on it via
// tea.ExecProcess (which suspends bubbletea's AltScreen for the
// duration), then reads the file back. The callback emits
// EditorExitedMsg with either Content (success) or Err. The temp file
// is best-effort removed before the callback returns.
//
// The shell-out is intentionally NOT covered by unit tests — they
// would have to launch a real editor. Manual dogfood + the cmd's
// downstream save flow (validateEditedSpec / Store.UpdateSpec) is
// where the verification happens.
func beginEditCmd(agentID, currentSpec string) tea.Cmd {
	f, err := os.CreateTemp("", "buddy-edit-*.yaml")
	if err != nil {
		return func() tea.Msg {
			return EditorExitedMsg{AgentID: agentID, Err: fmt.Errorf("temp file: %w", err)}
		}
	}
	if _, err := f.WriteString(currentSpec); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return func() tea.Msg {
			return EditorExitedMsg{AgentID: agentID, Err: fmt.Errorf("write temp: %w", err)}
		}
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return func() tea.Msg {
			return EditorExitedMsg{AgentID: agentID, Err: fmt.Errorf("close temp: %w", err)}
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, f.Name())
	return tea.ExecProcess(c, func(execErr error) tea.Msg {
		defer os.Remove(f.Name())
		if execErr != nil {
			return EditorExitedMsg{AgentID: agentID, Err: fmt.Errorf("editor: %w", execErr)}
		}
		content, readErr := os.ReadFile(f.Name())
		if readErr != nil {
			return EditorExitedMsg{AgentID: agentID, Err: fmt.Errorf("read back: %w", readErr)}
		}
		return EditorExitedMsg{AgentID: agentID, Content: content}
	})
}

// CreateStarterYAML is the initial buffer the user gets when they press
// `c` in list mode. Kept exported so tests + tui_cmd can reference the
// same source-of-truth. Comments at the top guide first-time users; the
// minimal `chain:` row keeps ParseSpec happy on an immediate save.
const CreateStarterYAML = `# Edit this spec to create a new agent. Save and exit your editor to
# submit. The id must be unique within the buddy.db you launched the
# TUI against.

id: new-agent
name: "New agent"

# Optional: a cron expression (e.g. "0 * * * *" or "@daily") to enable
# scheduled runs. Leave empty for on-demand only.
schedule: ""

chain:
  - command: status
    args: ""
`

// beginCreateCmd is the create-agent twin of beginEditCmd: write
// CreateStarterYAML to a temp file, launch $EDITOR, emit
// NewSpecEditorExitedMsg on return. Distinct message type so the
// reducer doesn't have to branch on an "is-create" flag inside the
// shared EditorExitedMsg handler.
func beginCreateCmd() tea.Cmd {
	f, err := os.CreateTemp("", "buddy-create-*.yaml")
	if err != nil {
		return func() tea.Msg {
			return NewSpecEditorExitedMsg{Err: fmt.Errorf("temp file: %w", err)}
		}
	}
	if _, err := f.WriteString(CreateStarterYAML); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return func() tea.Msg {
			return NewSpecEditorExitedMsg{Err: fmt.Errorf("write temp: %w", err)}
		}
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return func() tea.Msg {
			return NewSpecEditorExitedMsg{Err: fmt.Errorf("close temp: %w", err)}
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, f.Name())
	return tea.ExecProcess(c, func(execErr error) tea.Msg {
		defer os.Remove(f.Name())
		if execErr != nil {
			return NewSpecEditorExitedMsg{Err: fmt.Errorf("editor: %w", execErr)}
		}
		content, readErr := os.ReadFile(f.Name())
		if readErr != nil {
			return NewSpecEditorExitedMsg{Err: fmt.Errorf("read back: %w", readErr)}
		}
		return NewSpecEditorExitedMsg{Content: content}
	})
}

// saveNewSpecCmd is the create-agent twin of saveEditedSpecCmd. Parses
// yaml + calls Store.Create. Unlike the edit path there's no rename
// guard (the user is naming the new agent for the first time); the
// id-uniqueness check is the SQLite UNIQUE constraint on agents.id,
// which Store.Create surfaces as a wrapped error.
//
// Refusal path for an unchanged starter is intentional: if the user
// exits the editor without changing the template, the spec id stays
// "new-agent" and the SECOND `c` invocation will hit the UNIQUE
// constraint — the resulting "create: UNIQUE constraint failed"
// message tells the user to pick a real id rather than us guessing.
func saveNewSpecCmd(store AgentLister, yaml []byte) tea.Cmd {
	return func() tea.Msg {
		spec, err := agent.ParseSpec(yaml)
		if err != nil {
			return AgentCreateErrMsg{Err: fmt.Errorf("parse: %w", err)}
		}
		created, err := store.Create(context.Background(), spec, string(yaml))
		if err != nil {
			return AgentCreateErrMsg{Err: err}
		}
		return AgentCreatedMsg{ID: created.ID}
	}
}

// saveEditedSpecCmd validates yaml against the original ID, calls
// Store.UpdateSpec, and emits AgentSpecUpdatedMsg / AgentSpecUpdateErrMsg.
// originalID is the locked-in id from when the user pressed `e`; the
// spec in yaml is rejected if its id has changed (renames are out of
// scope for the edit flow — they would orphan runs/logs).
func saveEditedSpecCmd(store AgentLister, originalID string, yaml []byte) tea.Cmd {
	return func() tea.Msg {
		spec, err := agent.ParseSpec(yaml)
		if err != nil {
			return AgentSpecUpdateErrMsg{ID: originalID, Err: fmt.Errorf("parse: %w", err)}
		}
		if spec.ID != originalID {
			return AgentSpecUpdateErrMsg{
				ID: originalID,
				Err: fmt.Errorf("rename not allowed: spec id %q != original %q",
					spec.ID, originalID),
			}
		}
		if err := store.UpdateSpec(context.Background(), spec, string(yaml)); err != nil {
			return AgentSpecUpdateErrMsg{ID: originalID, Err: err}
		}
		return AgentSpecUpdatedMsg{ID: originalID}
	}
}

// loadLogChunkCmd fetches new log lines for runID with id strictly
// greater than sinceID and emits LogTailChunkMsg (or LogTailErrMsg).
// Sorted oldest-first by id (matches Store.LogsSince contract).
func loadLogChunkCmd(store AgentLister, runID, sinceID int64) tea.Cmd {
	return func() tea.Msg {
		lines, err := store.LogsSince(context.Background(), runID, sinceID)
		if err != nil {
			return LogTailErrMsg{Err: err}
		}
		return LogTailChunkMsg{Lines: lines}
	}
}

// tickLogTailCmd schedules the next LogTailTickMsg via tea.Tick. The
// reducer self-cancels stale ticks (mode-checked) so we never have to
// teardown the timer.
func tickLogTailCmd() tea.Cmd {
	return tea.Tick(logTailPollInterval, func(time.Time) tea.Msg {
		return LogTailTickMsg{}
	})
}

// deleteAgentCmd dispatches Store.Delete on a background goroutine and
// emits AgentDeletedMsg / AgentDeleteErrMsg. Carries the agent ID in
// both messages so the reducer can log it in error feedback without
// reaching back into mutable state.
func deleteAgentCmd(store AgentLister, agentID string) tea.Cmd {
	return func() tea.Msg {
		if err := store.Delete(context.Background(), agentID); err != nil {
			return AgentDeleteErrMsg{ID: agentID, Err: err}
		}
		return AgentDeletedMsg{ID: agentID}
	}
}

// loadSchedulerStatusCmd computes the scheduler-preview snapshot by
// listing agents and turning each non-empty cron string into a next-fire
// time via agent.PreviewSchedule. Now() is captured once per fetch so the
// pane has a stable reference clock between user-driven refreshes.
//
// Agents without a schedule are filtered out — the pane is "what's
// scheduled", on-demand agents already show up in the list pane.
func loadSchedulerStatusCmd(store AgentLister) tea.Cmd {
	return func() tea.Msg {
		agents, err := store.List(context.Background())
		if err != nil {
			return SchedulerStatusErrMsg{Err: err}
		}
		now := time.Now()
		entries := make([]SchedulerPreviewEntry, 0, len(agents))
		for _, a := range agents {
			if a.Schedule == "" {
				continue
			}
			p := agent.PreviewSchedule(a.Schedule, now)
			entries = append(entries, SchedulerPreviewEntry{
				AgentID:  a.ID,
				Schedule: a.Schedule,
				Next:     p.Next,
				Err:      p.Err,
			})
		}
		return SchedulerStatusLoadedMsg{Now: now, Entries: entries}
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

	case SchedulerStatusLoadedMsg:
		m.SchedulerNow = msg.Now
		m.SchedulerEntries = msg.Entries
		m.SchedulerLoaded = true
		m.SchedulerErr = nil
		return m, nil

	case SchedulerStatusErrMsg:
		m.SchedulerErr = msg.Err
		m.SchedulerLoaded = true
		return m, nil

	case AgentDeletedMsg:
		// Delete succeeded — return to the list and refresh so the deleted
		// row disappears. Friend-tone: success is silent (no toast banner),
		// the disappearance of the row is itself the confirmation.
		m.Mode = ModeList
		m.PendingDeleteID = ""
		m.Loaded = false
		m.Err = nil
		return m, loadAgentsCmd(m.Store)

	case AgentDeleteErrMsg:
		// Delete failed — return to list mode and surface the error via
		// m.Err so the list-pane error state renders it. The list rows are
		// left untouched (no optimistic removal).
		m.Mode = ModeList
		m.PendingDeleteID = ""
		m.Err = fmt.Errorf("delete agent %q: %w", msg.ID, msg.Err)
		return m, nil

	case LogTailChunkMsg:
		m.LogTailLoaded = true
		m.LogTailErr = nil
		if len(msg.Lines) > 0 {
			m.LogTailLines = append(m.LogTailLines, msg.Lines...)
			// Lines are oldest-first; the last one is the new high-water.
			m.LogTailLastID = msg.Lines[len(msg.Lines)-1].ID
		}
		// Chain the next tick so polling continues. handleKey on `esc`/`h`
		// just changes Mode; the next tick will see the mode change and
		// self-cancel without dispatching.
		return m, tickLogTailCmd()

	case LogTailErrMsg:
		m.LogTailErr = msg.Err
		m.LogTailLoaded = true
		// Keep polling on error too — transient DB locks shouldn't freeze
		// the pane. The error stays visible until the next successful
		// chunk clears it.
		return m, tickLogTailCmd()

	case LogTailTickMsg:
		// Self-cancel if the user has navigated away. Without this guard
		// every esc/h would leak a goroutine until quit.
		if m.Mode != ModeLogTail {
			return m, nil
		}
		return m, loadLogChunkCmd(m.Store, m.LogTailRunID, m.LogTailLastID)

	case EditorExitedMsg:
		if msg.Err != nil {
			m.EditErr = msg.Err
			return m, nil
		}
		// Editor exited cleanly with content. Dispatch the save cmd; its
		// AgentSpecUpdatedMsg / AgentSpecUpdateErrMsg drives the rest.
		return m, saveEditedSpecCmd(m.Store, msg.AgentID, msg.Content)

	case AgentSpecUpdatedMsg:
		// Spec saved. Clear any prior edit error and reload the list so
		// the edited name/schedule shows up in the row immediately. Stay
		// in detail mode — the user's context is preserved.
		m.EditErr = nil
		m.Loaded = false
		return m, loadAgentsCmd(m.Store)

	case AgentSpecUpdateErrMsg:
		// Save failed. Stash the error so the detail banner renders it;
		// list state is left intact (no optimistic mutation).
		m.EditErr = msg.Err
		return m, nil

	case NewSpecEditorExitedMsg:
		if msg.Err != nil {
			// Surface in m.Err so the list pane renders the error state.
			m.Err = fmt.Errorf("create: %w", msg.Err)
			return m, nil
		}
		return m, saveNewSpecCmd(m.Store, msg.Content)

	case AgentCreatedMsg:
		// Spec created. Friend-tone: silent success. Reload the list so
		// the new row appears at the top (List orders by updated_at DESC).
		m.Err = nil
		m.Loaded = false
		return m, loadAgentsCmd(m.Store)

	case AgentCreateErrMsg:
		// Surface via m.Err — same convention as delete-failure / create-
		// editor IO failure. Next list refresh clears it.
		m.Err = fmt.Errorf("create: %w", msg.Err)
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
		case "t":
			// Tail logs of the currently-shown run. No-op if no run
			// (DetailErr=ErrNotFound or Detail zero).
			if m.Detail.ID == 0 {
				return m, nil
			}
			m.Mode = ModeLogTail
			m.LogTailRunID = m.Detail.ID
			m.LogTailLastID = 0
			m.LogTailLines = nil
			m.LogTailErr = nil
			m.LogTailLoaded = false
			return m, loadLogChunkCmd(m.Store, m.LogTailRunID, 0)
		case "e":
			// Edit the spec via $EDITOR shell-out. Need the current
			// spec_yaml from m.Agents (list pane fetched it). No-op if
			// Selected isn't in the current list (rare — list may have
			// shrunk between detail entry and the e keypress).
			var specYAML string
			for _, a := range m.Agents {
				if a.ID == m.Selected {
					specYAML = a.SpecYAML
					break
				}
			}
			if specYAML == "" {
				return m, nil
			}
			m.EditErr = nil // clear stale banner on retry
			return m, beginEditCmd(m.Selected, specYAML)
		}
		// In detail mode every other key (j/k/g/G/etc.) is intentionally
		// inert — the detail pane is read-only.
		return m, nil
	}

	if m.Mode == ModeLogTail {
		switch key {
		case "esc", "h":
			m.Mode = ModeDetail
			return m, nil
		case "r":
			return m, loadLogChunkCmd(m.Store, m.LogTailRunID, m.LogTailLastID)
		}
		// Every other key is inert — the pane is a passive viewer.
		return m, nil
	}

	if m.Mode == ModeScheduler {
		switch key {
		case "esc", "h":
			m.Mode = ModeList
			return m, nil
		case "r":
			m.SchedulerLoaded = false
			return m, loadSchedulerStatusCmd(m.Store)
		}
		return m, nil
	}

	if m.Mode == ModeDeleteConfirm {
		switch key {
		case "y", "Y":
			if m.PendingDeleteID == "" {
				// Defensive: nothing to delete — treat as cancel.
				m.Mode = ModeList
				return m, nil
			}
			return m, deleteAgentCmd(m.Store, m.PendingDeleteID)
		case "n", "N", "esc":
			m.Mode = ModeList
			m.PendingDeleteID = ""
			return m, nil
		}
		// In confirm mode every other key (including nav keys) is
		// intentionally inert — the user must explicitly answer y or n
		// (or esc / quit).
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

	case "s":
		m.Mode = ModeScheduler
		m.SchedulerLoaded = false
		m.SchedulerErr = nil
		return m, loadSchedulerStatusCmd(m.Store)

	case "d":
		if len(m.Agents) == 0 {
			return m, nil
		}
		m.Mode = ModeDeleteConfirm
		m.PendingDeleteID = m.Agents[m.Cursor].ID
		return m, nil

	case "c":
		// Create a new agent via $EDITOR shell-out on a starter YAML.
		// No cursor dependency — works on an empty list too.
		m.Err = nil // clear any prior list-pane error before going to editor
		return m, beginCreateCmd()
	}
	return m, nil
}

// View renders the current Model state. Styles are lipgloss-driven but
// kept restrained: cli buddy's persona is "친구 — silent default", and a
// neon dashboard works against that.
func (m Model) View() string {
	switch m.Mode {
	case ModeDetail:
		return m.renderDetail()
	case ModeScheduler:
		return m.renderScheduler()
	case ModeDeleteConfirm:
		return m.renderDeleteConfirm()
	case ModeLogTail:
		return m.renderLogTail()
	default:
		return m.renderList()
	}
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
		if e.Err != nil {
			b.WriteString(fmt.Sprintf("  %-30s %-20s %s\n",
				e.AgentID,
				e.Schedule,
				errorStyle.Render("invalid: "+e.Err.Error())))
			continue
		}
		b.WriteString(fmt.Sprintf("  %-30s %-20s next %s\n",
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
	b.WriteString("\n")
	b.WriteString(footerHintLogTail())
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
	return footerStyle.Render("j/k or ↑/↓ move · g/G top/bottom · enter/l detail · s scheduler · c create · d delete · r refresh · q quit")
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
