package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/0xmhha/buddy/internal/agent"
)

// Init kicks off the asynchronous loads the list mode renders against
// (the agent list itself, plus the notification banner if a fetcher is
// wired). tea.Batch lets both run in parallel so the first paint sees
// real data rather than two staggered re-renders.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadAgentsCmd(m.Store),
		loadNotifyCmd(m.NotifyFetcher),
	)
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

// loadHookStatsCmd fetches the hook-reliability snapshot for the given
// window and emits HookStatsLoadedMsg or HookStatsErrMsg. Defensive on
// nil fetcher (treated as "no data"). Window is forwarded verbatim so
// downstream queries.Run validation handles bad values.
func loadHookStatsCmd(fetcher HookStatsFetcher, window string) tea.Cmd {
	if fetcher == nil {
		return nil
	}
	return func() tea.Msg {
		res, err := fetcher(window)
		if err != nil {
			return HookStatsErrMsg{Err: err}
		}
		return HookStatsLoadedMsg{Window: window, Result: res}
	}
}

// loadUsageCmd fetches the Overview snapshot for the Usage pane.
// Mirrors loadHookStatsCmd's nil-tolerant + error-channelled pattern.
func loadUsageCmd(fetcher UsageFetcher) tea.Cmd {
	if fetcher == nil {
		return nil
	}
	return func() tea.Msg {
		res, err := fetcher()
		if err != nil {
			return UsageErrMsg{Err: err}
		}
		return UsageLoadedMsg{Result: res}
	}
}

// loadAdvisorCmd is the advisor companion. Dispatched alongside
// loadUsageCmd whenever the user enters / refreshes the Usage pane.
// Failure is non-fatal — the pane keeps rendering metric blocks while
// the advisor section shows the error.
func loadAdvisorCmd(fetcher AdvisorFetcher) tea.Cmd {
	if fetcher == nil {
		return nil
	}
	return func() tea.Msg {
		res, err := fetcher()
		if err != nil {
			return AdvisorErrMsg{Err: err}
		}
		return AdvisorLoadedMsg{Result: res}
	}
}

// loadNotifyCmd is the notify-banner companion. Dispatched on Init and
// on ModeList `r` so the banner reflects daemon-side dispatch state.
// nil fetcher → no banner.
func loadNotifyCmd(fetcher NotifyFetcher) tea.Cmd {
	if fetcher == nil {
		return nil
	}
	return func() tea.Msg {
		rows, err := fetcher()
		if err != nil {
			return NotifyErrMsg{Err: err}
		}
		return NotifyLoadedMsg{Rows: rows}
	}
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
//
// Also fetches the run row via GetRun so the reducer can detect
// completion (RunEnded) and stop scheduling polls. A GetRun error
// degrades to "still running" rather than failing the chunk: the user
// would rather see fresh log lines than have a transient DB hiccup
// abort their tail.
func loadLogChunkCmd(store AgentLister, runID, sinceID int64) tea.Cmd {
	return func() tea.Msg {
		lines, err := store.LogsSince(context.Background(), runID, sinceID)
		if err != nil {
			return LogTailErrMsg{Err: err}
		}
		ended := false
		if run, err := store.GetRun(context.Background(), runID); err == nil {
			ended = run.EndedAt != nil
		}
		return LogTailChunkMsg{Lines: lines, RunEnded: ended}
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
				Status:   a.Status,
			})
		}
		return SchedulerStatusLoadedMsg{Now: now, Entries: entries}
	}
}
