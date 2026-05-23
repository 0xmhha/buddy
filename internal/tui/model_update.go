package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/0xmhha/buddy/internal/agent"
)

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
		if msg.RunEnded {
			m.LogTailDone = true
		}
		// Stop scheduling polls once the run has finished. The
		// final chunk we just folded in carried EndedAt; any further
		// log lines would only land if the runtime were resurrected
		// (it doesn't). handleKey on `esc`/`h` returns the user to
		// the detail pane normally.
		if m.LogTailDone {
			return m, nil
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
		// chunk clears it. Done-runs still don't get polled.
		if m.LogTailDone {
			return m, nil
		}
		return m, tickLogTailCmd()

	case LogTailTickMsg:
		// Self-cancel if the user has navigated away. Without this guard
		// every esc/h would leak a goroutine until quit.
		if m.Mode != ModeLogTail {
			return m, nil
		}
		if m.LogTailDone {
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

	case HookStatsLoadedMsg:
		m.HookStatsResult = msg.Result
		m.HookStatsWindow = msg.Window
		m.HookStatsLoaded = true
		m.HookStatsErr = nil
		return m, nil

	case HookStatsErrMsg:
		m.HookStatsErr = msg.Err
		m.HookStatsLoaded = true
		return m, nil

	case UsageLoadedMsg:
		m.UsageResult = msg.Result
		m.UsageLoaded = true
		m.UsageErr = nil
		return m, nil

	case UsageErrMsg:
		m.UsageErr = msg.Err
		m.UsageLoaded = true
		return m, nil

	case AdvisorLoadedMsg:
		m.AdvisorResult = msg.Result
		m.AdvisorLoaded = true
		m.AdvisorErr = nil
		return m, nil

	case AdvisorErrMsg:
		m.AdvisorErr = msg.Err
		m.AdvisorLoaded = true
		return m, nil

	case NotifyLoadedMsg:
		m.NotifyRows = msg.Rows
		m.NotifyLoaded = true
		m.NotifyErr = nil
		return m, nil

	case NotifyErrMsg:
		m.NotifyErr = msg.Err
		m.NotifyLoaded = true
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
			m.LogTailDone = false
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

	if m.Mode == ModeHookStats {
		switch key {
		case "esc", "h":
			m.Mode = ModeList
			return m, nil
		case "r":
			m.HookStatsLoaded = false
			return m, loadHookStatsCmd(m.HookStatsFetcher, m.HookStatsWindow)
		}
		// Other keys (j/k/g/G/etc.) are intentionally inert — the pane
		// is a read-only snapshot.
		return m, nil
	}

	if m.Mode == ModeUsage {
		switch key {
		case "esc", "h":
			m.Mode = ModeList
			return m, nil
		case "r":
			m.UsageLoaded = false
			m.AdvisorLoaded = false
			return m, tea.Batch(
				loadUsageCmd(m.UsageFetcher),
				loadAdvisorCmd(m.AdvisorFetcher),
			)
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
		// wait for the scheduler's next poll. Refresh notify banner
		// too so a fresh daemon dispatch shows up immediately.
		m.Loaded = false
		m.NotifyLoaded = false
		return m, tea.Batch(
			loadAgentsCmd(m.Store),
			loadNotifyCmd(m.NotifyFetcher),
		)

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

	case "H":
		// Hook reliability stats pane — surfaces the daemon/aggregator
		// output inside the cli buddy TUI. Capital H so lowercase `h`
		// stays free for back-nav in other modes. No-op when no fetcher
		// is wired (e.g., a TUI invocation without DB-stats access).
		if m.HookStatsFetcher == nil {
			return m, nil
		}
		m.Mode = ModeHookStats
		if m.HookStatsWindow == "" {
			m.HookStatsWindow = "1h" // align with `buddy stats` default
		}
		m.HookStatsLoaded = false
		m.HookStatsErr = nil
		return m, loadHookStatsCmd(m.HookStatsFetcher, m.HookStatsWindow)

	case "U":
		// Usage pane. Capital U so lowercase u stays free for future
		// use. No-op when fetcher unset.
		if m.UsageFetcher == nil {
			return m, nil
		}
		m.Mode = ModeUsage
		m.UsageLoaded = false
		m.UsageErr = nil
		m.AdvisorLoaded = false
		m.AdvisorErr = nil
		return m, tea.Batch(
			loadUsageCmd(m.UsageFetcher),
			loadAdvisorCmd(m.AdvisorFetcher),
		)
	}
	return m, nil
}
