// Package tui hosts the bubbletea-based terminal UI for cli buddy. The
// surface started as a read-only agent list with j/k navigation and a
// q-to-quit binding, then accreted the detail view (Enter on a list row
// jumps to the latest-run summary), create form, live log tail, and
// scheduler status pane in subsequent releases.
//
// The Update method is intentionally a pure-function reducer so tests
// can drive it with synthetic tea.Msg values without spinning up a real
// terminal. The View method is best-effort — its rendered output is
// asserted only via substring matches.
package tui

import (
	"context"
	"time"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/agent"
	"github.com/0xmhha/buddy/internal/notify"
	"github.com/0xmhha/buddy/internal/queries"
	"github.com/0xmhha/buddy/internal/usage"
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
	GetRun(ctx context.Context, runID int64) (agent.AgentRun, error)
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
	ModeHookStats
	ModeUsage
)

// HookStatsFetcher is the read-only surface the TUI's hook-stats pane
// calls when the user presses `H` in list mode. Production wiring (see
// cmd/buddy/tui_cmd.go) closes over a DB path and delegates to
// internal/queries.Run; tests can inject a stub that returns canned
// rows without touching SQLite.
//
// nil is a legitimate value: the TUI treats a nil fetcher as "hook
// stats unavailable in this session" and shows a friend-tone copy
// instead of crashing on the H keypress.
type HookStatsFetcher func(window string) (queries.Result, error)

// UsageFetcher is the read-only surface for the TUI's Usage pane
// (ADR-013). Production wiring closes over a *usage.Service; tests
// can inject a stub that returns canned Overview without touching SQLite.
//
// nil = pane shows "unavailable" copy and stays in list mode (mirrors
// HookStatsFetcher's nil-tolerant policy).
type UsageFetcher func() (usage.Overview, error)

// AdvisorFetcher returns advisories to render under the Usage pane's
// metric blocks (ADR-015). Same nil-tolerant policy as
// UsageFetcher — nil renders an empty advisory section silently rather
// than erroring.
type AdvisorFetcher func() ([]advisor.Advisory, error)

// NotifyFetcher returns recent notification_log rows for the ModeList
// top banner (ADR-016). Daemon dispatched advisories show up
// here even without TUI navigation. nil = banner suppressed (silent
// install).
type NotifyFetcher func() ([]notify.LogRow, error)

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
	// Status mirrors agent.Status — only "running" gets a visible
	// marker; the rest render blank so non-running schedules don't add
	// glyph noise.
	Status agent.Status
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
	// LogTailDone latches once GetRun reports EndedAt != nil for the
	// tailed run. The reducer stops scheduling the next tea.Tick once
	// it is true, so a finished run doesn't keep the polling loop alive.
	LogTailDone bool

	// Edit-flow state. EditErr stashes the most recent edit failure
	// (editor crash, parse fail, rename rejected, store err); the detail
	// pane renders it as a banner with a retry hint. Cleared on the next
	// `e` keypress (about to retry) or on a successful save.
	EditErr error

	// Hook-stats pane state — only meaningful when Mode==ModeHookStats.
	// HookStatsFetcher is the injected read function (nil = pane shows
	// "unavailable" copy on the H keypress and stays in list mode).
	HookStatsFetcher HookStatsFetcher
	HookStatsWindow  string // "5m" | "1h" | "24h" (default "1h")
	HookStatsResult  queries.Result
	HookStatsErr     error
	HookStatsLoaded  bool

	// Usage-pane state — only meaningful when Mode==ModeUsage.
	// UsageFetcher is the injected read closure that wraps a
	// usage.Service. nil = pane shows "unavailable" copy and the U
	// keypress is a no-op (ADR-013).
	UsageFetcher UsageFetcher
	UsageResult  usage.Overview
	UsageErr     error
	UsageLoaded  bool

	// Advisor section state (ADR-015) — co-rendered inside the
	// Usage pane. Independent loaded flag so the metric blocks render
	// even when advisories are still in-flight.
	AdvisorFetcher    AdvisorFetcher
	AdvisorResult     []advisor.Advisory
	AdvisorErr        error
	AdvisorLoaded     bool

	// Notify banner state (ADR-016) — rendered at the top of
	// ModeList view. NotifyFetcher nil keeps the banner suppressed.
	NotifyFetcher NotifyFetcher
	NotifyRows    []notify.LogRow
	NotifyLoaded  bool
	NotifyErr     error
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
	LogTailChunkMsg struct {
		Lines []agent.AgentLog
		// RunEnded is true when the tailed run row has EndedAt != nil.
		// The reducer uses this to stop scheduling the next tea.Tick so
		// long-completed runs don't keep churning the polling loop.
		RunEnded bool
	}
	LogTailErrMsg  struct{ Err error }
	LogTailTickMsg struct{}
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
	HookStatsLoadedMsg  struct {
		Window string
		Result queries.Result
	}
	HookStatsErrMsg struct{ Err error }
	UsageLoadedMsg  struct{ Result usage.Overview }
	UsageErrMsg     struct{ Err error }
	AdvisorLoadedMsg struct{ Result []advisor.Advisory }
	AdvisorErrMsg    struct{ Err error }
	NotifyLoadedMsg  struct{ Rows []notify.LogRow }
	NotifyErrMsg     struct{ Err error }
)

// NewModel constructs an empty Model around an AgentLister. The Init
// command is what kicks off the first load — until that resolves, View
// renders a "loading…" placeholder.
func NewModel(store AgentLister) Model {
	return Model{Store: store}
}

