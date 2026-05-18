# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added — TUI hook-stats pane (A-3.2 W3-5 follow-on, cli buddy ↔ hook monitor integration)

The W3-5 cycle-handoff §6.1 "hook reliability monitor → cli buddy sub-feature" item lands as a new TUI mode: pressing `H` in list view opens an inline pane that calls `internal/queries.Run` and renders the same count / failures / p50 / p95 per-hook snapshot the `buddy stats` CLI already produces. The capital `H` keeps lowercase `h` free for back-navigation in other modes.

This is the structural integration the cycle-handoff was pointing at — the v0.1.0 daemon/aggregator was already shipping (`buddy daemon/stats/events` CLI surface), but it lived "next to" the W3-3 agent management work rather than feeling like part of the same product. Surfacing it inside `buddy tui` makes the unified control-plane story real: the user opens one binary, sees agents in the list, hits `s` for the scheduler preview, `H` for the hook reliability monitor — all in the same AltScreen session.

Notably this commit does NOT touch the existing `buddy daemon` / `buddy stats` / `buddy events` CLI surface. Those keep working verbatim for users who scripted around them; the integration is *additive*. No code from `internal/daemon` or `internal/aggregator` was moved or restructured — the integration happens at the consumer (TUI) layer, not the producer (daemon) layer.

What ships:

- New `tui.HookStatsFetcher` function type: `func(window string) (queries.Result, error)`. Optional injection on `tui.Model.HookStatsFetcher`. nil → pressing `H` is a no-op (the TUI stays in list mode); production wiring (`cmd/buddy/tui_cmd.go`) closes over the `--db` flag and delegates to `queries.Run`.
- New `Mode` value `ModeHookStats` + state fields: `HookStatsWindow string` (default `"1h"`), `HookStatsResult queries.Result`, `HookStatsErr`, `HookStatsLoaded`.
- New reducer messages: `HookStatsLoadedMsg{Window, Result}`, `HookStatsErrMsg{Err}`. `loadHookStatsCmd(fetcher, window)` runs the fetch off the reducer.
- Key bindings (additive — every existing binding unchanged):
  - `H` (list mode) — open the hook-stats pane. Defaults `HookStatsWindow` to `"1h"` (matches the `buddy stats` CLI default). No-op when `HookStatsFetcher` is nil.
  - `esc` / `h` (hook-stats mode) — return to list.
  - `r` (hook-stats mode) — refetch with the locked-in window.
  - `q` / `Ctrl-C` — quit (every mode).
- Render layout: bold header `buddy hook stats — window <W>`, then a `hook · tool · count · fail · p50ms · p95ms` table (same column ordering as the `buddy stats` CLI, so users who switch between the two surfaces see the same shape). Per-row tool name is `-` when blank (mirrors the CLI's "hook-level aggregate" rendering for hooks like `Stop` that have no tool axis).
- Friend-tone empty / loading / error copy: `loading hook stats…` until the fetch resolves; `(no hook events in this window — daemon may be idle or DB empty)` for the empty case; `error: <message>` for fetch failure.

Test coverage (`internal/tui/model_test.go` — 10 new race-clean tests):

- Update reducer: `H` with a wired fetcher enters ModeHookStats + defaults window to "1h" + dispatches fetch cmd; `H` without a fetcher is a no-op; `HookStatsLoadedMsg` folds rows/window/Loaded; `HookStatsErrMsg` records err + flips Loaded; `esc` / `h` return to list; `q` quits; `r` refires with the locked-in window.
- View smoke: pane renders header + rows (HookName / ToolName / Count / P95Ms / esc-back-hint); empty-state shows `no hook events` copy; loading placeholder visible; error string visible.

`docs/cli-buddy-spec.md` §9 W3-5 row updated with the 2026-05-18 hook-stats-pane sub-item. W3-5 main.go split (v0.6.3) was already Done; with this commit the *integration* sub-item is also Done. W3-5 row is now fully closed.

## [0.7.1] — 2026-05-18

Patch release bundling the four post-v0.7.0 commits. Three close out the W3-2 follow-on backlog; one closes the deferred W3-4 chain-control sub-items. Strictly additive — every existing spec keeps its v0.7.0 semantics verbatim.

What's in this release (newest commit first):

- `59ba94b` — **W3-4 chain control** (per-step `continue_on_fail` + chain-level `auto_cascade`). New spec fields, runtime queue-based loop, `StepResult.CascadeDepth`. Cascade is skip-on-failure; default depth cap is 5. Closes the W3-4 deferred items from v0.6.x.
- `29fb89e` — **W3-2 create form** (the final named W3-2 follow-on). `c` in list mode opens `$EDITOR` on a starter YAML; `ParseSpec` + `Store.Create` persists. All 6 named W3-2 follow-on items are now shipped.
- `4ae1e0f` — **W3-2 in-app spec edit**. `e` in detail mode shells out to `$EDITOR` on the agent's spec_yaml; `Store.UpdateSpec` persists (status / created_at / last_run_at preserved). Renames rejected to avoid orphaning runs/logs.
- `b1b9b7c` — **W3-2 live log tail**. `t` in detail mode opens an inline pane that polls `Store.LogsSince(runID, sinceLogID)` every ~1s via `tea.Tick`. Self-cancels on mode change (no goroutine leak).

Counts and gates:

- 5 version sources (Makefile, plugin.json, marketplace.json, server.go, main.go) all on `0.7.1` (`make verify-versions` passes).
- `go build / go vet / go test -race -count=1 -timeout=180s ./...` — 23 packages green.
- `internal/tui` package at 85 race-clean tests (34 new since v0.7.0 across create / edit / log tail).
- `internal/agent` package added `Store.LogsSince` + `Store.UpdateSpec` + 10 new runtime tests (continue_on_fail 4 + auto_cascade 6).
- `cli-buddy-spec.md` §9 W3-2 and W3-4 both read **Done** with this release. The remaining cascade items are W3-5 partial (hook reliability monitor cli buddy integration deferred) and one branch-aware-selection sub-item of W3-4 (deferred pending dogfood signal).

What stays open after v0.7.1:

- **W3-5 hook reliability monitor → cli buddy sub-feature integration** (the remaining track-identity cleanup from cycle-handoff §6.1).
- **Branch-aware `auto_cascade` selection** (currently picks `Skills[0]`; conditional `Branches` only logged).
- **Scheduler pane live "currently running" indicator** (W3-2 follow-on-of-follow-on; requires coupling the TUI to a running `Scheduler` instance).
- **Log tail scrollback + auto-stop on run-end** (W3-2 follow-on-of-follow-on; current pane polls indefinitely and renders everything accumulated).
- **Plugin v1.0.0 entry conditions B-2 (production dogfood) / B-3 (PROCEDURE B6 unification) / B-4 (router smart-skip session state)** remain user-paced or trigger-bound.

### Added — W3-4 chain control: per-step `continue_on_fail` + chain-level `auto_cascade`

Two W3-4 cascade items shipped together. They were paired because both touch the runtime's chain-loop semantics, and shipping them in one commit keeps the spec field additions visible side-by-side.

**Per-step `continue_on_fail`** (P2-1): adds a per-`ChainStep` boolean that lets the chain continue past a step that exhausts its retry budget without success. Default `false` preserves the v0.6.x fail-fast behaviour verbatim.

- New spec field `chain[].continue_on_fail` (bool, default `false`). The failed step is still recorded in `RunResult.Steps` with its non-zero exit code; what changes is that the chain does NOT short-circuit. A `warn`-level log line is emitted at the runtime: `step[N] <command> failed (exit=X) — continue_on_fail=true, chain continues`.
- Final `RunResult.ExitCode` is the LAST non-zero step exit when any step failed (so a single failing cleanup step at the end still surfaces failure at the run level); 0 only when every step (including continue_on_fail ones) succeeded. Agent row transitions to `failed` whenever the run-level exit code is non-zero.
- Strictly backward-compatible: every existing spec gets default `false` and behaves exactly as before.

**Chain-level `auto_cascade`** (P2-2): adds an opt-in `auto_cascade: {}` block on the spec that turns on §next-phase-driven chain extension. After every *successful* step, the runtime looks at `ParsedOutput.NextPhase.Skills` and, if non-empty, appends the first listed skill as a new chain step. Cascading is bounded by `max_depth` (default 5 — wide enough for the §1→§9 happy-path orchestrator chain plus one tier of slack).

- New spec field `auto_cascade` (optional `AutoCascadeConfig` struct). A bare `auto_cascade: {}` enables with default depth; `auto_cascade: { max_depth: 10 }` raises the cap. Omitting the field preserves v0.6.x sequential semantics.
- New `agent.DefaultCascadeMaxDepth = 5` constant.
- New `StepResult.CascadeDepth int` field: 0 for original chain steps, N+1 for steps appended via §next-phase from a step at depth N. Lets `buddy agent log` / TUI distinguish "user wrote this step" from "the runtime inferred it".
- Runtime loop refactored from `for i, step := range spec.Chain` to a queue-based `for len(queue) > 0`. Original chain seeds the queue at depth 0; cascade `append`s to the same queue at `depth+1`. The queue can grow during iteration (Go-safe — we're indexing by element, not a fixed slice header).
- Cascade selection rule (minimum-viable): `pickCascadeTarget(NextPhase)` returns `NextPhase.Skills[0]`. PROCEDURE-side conditional branches (`Branches`) are surfaced in the run log but NOT followed automatically — they need runtime context (e.g., target-market env vars) the runtime doesn't have. A follow-on can add branch-aware selection once dogfood signal arrives.
- Cascade is strictly *skip-on-failure*: a step that exits non-zero or errors does NOT contribute a cascaded successor, even with `continue_on_fail: true`. This avoids "broken §next-phase cascades down a fault path".
- A new info-level log line on every cascade: `step[N] <command> auto-cascade → <skill> (depth K)`.

Interaction with `continue_on_fail`: orthogonal. A failed step with `continue_on_fail: true` keeps the *original* chain alive but does NOT cascade. A succeeded step with `auto_cascade` enabled queues its §next-phase successor regardless of whether any earlier step had `continue_on_fail` set.

Test coverage (`internal/agent/runtime_test.go` — 10 new race-clean tests):

- `continue_on_fail`: failed step advances chain to step 2; run-level exit is the last non-zero (cleanup-fails-at-end case); default `false` still short-circuits (regression guard); all-success path still returns exit 0.
- `auto_cascade`: parsed §next-phase appends a new step (CascadeDepth=1); MaxDepth caps the chain at depth 0..MaxDepth inclusive; disabled `auto_cascade` does NOT cascade (regression guard); failed step does NOT cascade even with continue_on_fail; absent §next-phase = no cascade (terminal-phase case); default cap = `DefaultCascadeMaxDepth` constant.

`docs/cli-buddy-spec.md` §9 W3-4 row updated: now reads `Done 2026-05-17` (was `partial Done 2026-05-12 — parser ship in v0.5.0 + conditional branches ship in v0.6.0, retry/fail 의미 변경 + auto-cascade deferred`). The deferred sub-items are now shipped.

### Added — W3-2 TUI create form (the last named W3-2 follow-on)

The sixth — and final — W3-2 follow-on item: pressing `c` in list mode opens `$EDITOR` (fall back to `$VISUAL`, then `vi`) on a starter YAML template (`new-agent` placeholder + commented hints). On exit, the TUI parses the saved content with `agent.ParseSpec` and persists via `Store.Create`. Success silently reloads the list; failure (parse error, duplicate id from SQLite UNIQUE, IO problem) surfaces in the existing list-pane error state.

Design note — instead of a bubbletea form state machine (textinput / textarea components) the create flow reuses the same `tea.ExecProcess` shell-out pattern that landed in the in-app edit cycle. Trade-off: users get full YAML editor power (syntax-highlighted vim, multi-cursor goodness, paste from clipboard) at near-zero TUI-side cost; the cost is the AltScreen flicker as the editor takes over. Given cli buddy's "친구 — silent default" persona and the YAML-first spec format, the shell-out feels more honest than a half-baked in-TUI form.

What ships:

- New `agent.AgentSpec`-emitting helper not required (the existing `ParseSpec` covers the validation surface).
- `AgentLister` interface widened with `Create(ctx, spec, yaml) (Agent, error)` — the existing `*agent.Store.Create` already matches the signature.
- New `tui.CreateStarterYAML` package constant — the starter template the user gets on first `c`. Exported so tests can round-trip it through `ParseSpec` (the test gates against typos in the template that would make every fresh save land in the parse-fail branch).
- New reducer messages: `NewSpecEditorExitedMsg{Content, Err}` (from the editor shell-out), `AgentCreatedMsg{ID}` (save success), `AgentCreateErrMsg{Err}` (parse fail / Store.Create err / IO). Distinct from the edit cycle's `EditorExitedMsg` so the reducer doesn't have to branch on an "is-create" flag.
- New cmds:
  - `beginCreateCmd()` — writes `CreateStarterYAML` to `os.CreateTemp` (`buddy-create-*.yaml`), launches editor via `tea.ExecProcess`, emits `NewSpecEditorExitedMsg` with read-back content. Best-effort temp file cleanup in the callback.
  - `saveNewSpecCmd(store, yaml)` — `ParseSpec` + `Store.Create` + emits success/err msg. Unlike the edit path there's no rename guard (the user is naming the new agent for the first time); the id-uniqueness check is the SQLite UNIQUE constraint on `agents.id`, which `Store.Create` surfaces as a wrapped error.
- Key bindings (additive — list/detail bindings unchanged):
  - `c` (list mode) — open the create flow. No cursor dependency — works on an empty list too (otherwise users with no agents can't bootstrap). Clears any prior list-pane error before going to the editor.
- Failure modes handled explicitly: temp-file create / write / close fails → `NewSpecEditorExitedMsg.Err` with wrapped error → surfaced as `m.Err = "create: <wrapped>"`; editor crash → same; read-back fail → same; ParseSpec fail → `AgentCreateErrMsg`; UNIQUE-constraint duplicate-id from `Store.Create` → `AgentCreateErrMsg` (the user sees the SQLite message and re-edits with a different id).
- Render: re-uses the list-pane error state for create failures (same convention as delete failures). Footer hint adds `c create`.

Test coverage (`internal/tui/model_test.go` — 9 new race-clean tests, 85 total in the package):

- Save cmd: valid YAML dispatches `Store.Create` + emits `AgentCreatedMsg`; invalid YAML short-circuits with `AgentCreateErrMsg`; store err propagates with `errors.Is` chain preserved.
- Update reducer: `c` in list dispatches the create cmd + clears prior list err; `c` on an empty list still dispatches (bootstrap path); `NewSpecEditorExitedMsg{Err}` records err with `create:` wrapper and does NOT chain a save cmd; `NewSpecEditorExitedMsg{Content}` dispatches save cmd; `AgentCreatedMsg` clears err + resets Loaded + dispatches list reload; `AgentCreateErrMsg` records err with `create:` wrapper and leaves list intact.
- Starter template integrity: `agent.ParseSpec(CreateStarterYAML)` round-trips to a valid spec with at least one chain step.

`docs/cli-buddy-spec.md` §9 W3-2 row updated: with this commit, the row reads "minimum-viable + 6/6 follow-on Done (detail, scheduler-preview, in-app delete, log tail, in-app edit, create)". W3-2 is no longer the limiting line item on the cli-buddy-spec cascade.

What stays open from W3-2 follow-on-of-follow-on (deferred): scheduler pane live "currently running" indicator (requires coupling the TUI to a running `Scheduler` instance); log tail scrollback + auto-stop on run-end.

### Added — W3-2 TUI in-app spec edit (follow-on of v0.7.0)

The fifth W3-2 follow-on item: pressing `e` in the detail pane writes the agent's current `spec_yaml` to a temp file, suspends the TUI (AltScreen), and shells out to `$EDITOR` (fall back to `$VISUAL`, then `vi`). On exit, the TUI reads the file back, runs `agent.ParseSpec` for validation, rejects ID renames (would orphan runs/logs), and dispatches `Store.UpdateSpec` for a persisted save. Friend-tone: success silently reloads the list so the new name/schedule renders in the row; failure shows an `edit error: ...` banner in the detail pane with a `press e to try again` hint.

What ships:

- New `Store.UpdateSpec(ctx, spec, yaml) error` in `internal/agent/store.go`. Swaps `name` / `schedule` / `spec_yaml` in one UPDATE and bumps `updated_at`. `status`, `created_at`, `last_run_at` are preserved verbatim so an editorial change does NOT reset run history or in-progress state. Returns `ErrNotFound` when `spec.ID` matches no row (e.g., the agent was deleted between the user pressing `e` and the save round-trip).
- `AgentLister` interface widened to include `UpdateSpec` (the existing `*agent.Store` already implements it).
- New `EditErr error` field on `tui.Model`. Cleared on each `e` keypress (about to retry) and on `AgentSpecUpdatedMsg`; persists across other state changes so the banner survives a list reload.
- New reducer messages: `EditorExitedMsg{AgentID, Content, Err}` (from the `tea.ExecProcess` callback), `AgentSpecUpdatedMsg{ID}` (save success), `AgentSpecUpdateErrMsg{ID, Err}` (parse-fail / rename-rejected / store-err).
- New cmds:
  - `beginEditCmd(agentID, currentSpec)` — writes `currentSpec` to `os.CreateTemp` (`buddy-edit-*.yaml`), launches `$EDITOR` via `tea.ExecProcess` (suspends/resumes AltScreen automatically), and emits `EditorExitedMsg` with the read-back content. Temp file is best-effort removed in the callback regardless of outcome.
  - `saveEditedSpecCmd(store, originalID, yaml)` — runs `agent.ParseSpec`, rejects renames (`spec.ID != originalID`), calls `Store.UpdateSpec`, emits the success/err msg.
- Key bindings (additive — list/detail/scheduler/log-tail bindings unchanged):
  - `e` (detail mode) — open the spec editor. No-op if the selected agent isn't in `m.Agents` (rare — list shrank between detail entry and the e keypress).
- Failure modes handled explicitly: temp-file create / write / close fails → `EditorExitedMsg.Err` with wrapped error; editor crashes / non-zero exit → same; read-back fails → same; ParseSpec fail → `AgentSpecUpdateErrMsg` with wrapped error; ID rename → `AgentSpecUpdateErrMsg` with `rename not allowed: spec id "X" != original "Y"`; `Store.UpdateSpec` err → same.
- Render: when `EditErr != nil`, the detail pane appends an error-styled `edit error: <message>` banner and a dim `(press e to try again)` retry hint below the latest-run section. Footer hint adds `e edit spec`.

Test coverage (`internal/tui/model_test.go` — 11 new race-clean tests, 76 total in the package; `internal/agent/store_test.go` — 3 new race-clean tests for `UpdateSpec`):

- Store: `UpdateSpec` replaces name/schedule/spec_yaml and preserves status/created_at; unknown ID returns `ErrNotFound`; `last_run_at` is preserved across the edit so the list keeps its "X minutes ago" cue.
- Save cmd unit tests: valid YAML dispatches Update + emits `AgentSpecUpdatedMsg`; rename rejected without calling `UpdateSpec`; invalid YAML returns `AgentSpecUpdateErrMsg` short-circuiting before the store; store err propagates as `AgentSpecUpdateErrMsg` with `errors.Is` chain preserved.
- Update reducer: `e` in detail with a known agent dispatches the edit cmd + clears prior `EditErr`; `e` with `Selected` missing from `m.Agents` is a no-op; `EditorExitedMsg{Err}` records the err and does NOT chain a save cmd; `EditorExitedMsg{Content}` dispatches the save cmd; `AgentSpecUpdatedMsg` clears `EditErr` + resets `Loaded` + dispatches list reload; `AgentSpecUpdateErrMsg` records err + leaves list state intact.
- View smoke: `EditErr != nil` renders the `edit error: ...` banner + `press e` retry hint.

The `tea.ExecProcess` shell-out is intentionally NOT covered by unit tests — they would have to launch a real editor. Manual dogfood is where it gets verified.

What stays open from W3-2 follow-on (after this):

- **Create form** (interactive spec builder) — HIGH cost, deserves a dedicated cycle. The last W3-2 follow-on still open.
- **Scheduler pane: live "currently running" indicator** (requires coupling the TUI to a running `Scheduler` instance).
- **Log tail: scrollback + auto-stop on run-end** (right now the pane shows all accumulated lines and polls forever).

`docs/cli-buddy-spec.md` §9 W3-2 row updated with the 2026-05-17 in-app-edit follow-on entry. With this, W3-2 has shipped 5 of its 6 named follow-on items (only create form remains).

### Added — W3-2 TUI live log tail (follow-on of v0.7.0)

The fourth W3-2 follow-on item: pressing `t` in the detail pane opens a live log-tail view of the run shown above. The pane polls `agent_logs` every ~1s for new lines (incremental — only rows with `id` strictly greater than the high-water mark are fetched), accumulates them oldest-first, and renders `HH:MM:SS  <level>  <message>` per row. Pair with the v0.6.4 streaming-log infrastructure: as the runtime appends per-line, the TUI sees each line within a tick.

What ships:

- New `Store.LogsSince(ctx, runID, sinceLogID) ([]AgentLog, error)` in `internal/agent/store.go`. SQL: `WHERE run_id = ? AND id > ? ORDER BY id ASC`. Sort by `id` (monotonic SQLite rowid alias) rather than `ts` so rapid line bursts on systems with low-precision clocks still arrive in order.
- `AgentLister` interface widened to include `LogsSince` (the existing `*agent.Store` already implements it).
- New `Mode` value `ModeLogTail` + state fields on `tui.Model`: `LogTailRunID` (the run we're tailing — captured on entry), `LogTailLastID` (high-water mark passed as `sinceID` on the next poll), `LogTailLines []agent.AgentLog`, `LogTailErr`, `LogTailLoaded`.
- New reducer messages: `LogTailChunkMsg{Lines}`, `LogTailErrMsg{Err}`, `LogTailTickMsg{}`. `loadLogChunkCmd` runs the fetch off the reducer; `tickLogTailCmd` wraps `tea.Tick(logTailPollInterval, …)`. The poll interval is a package-level `var` so future tests can monkey-patch it without exposing a Model field.
- Self-cancelling polling: every `LogTailChunkMsg` / `LogTailErrMsg` chains the next `LogTailTickMsg`. When the user navigates away (`esc` / `h` → ModeDetail, or quit), the next tick checks `m.Mode != ModeLogTail` and drops the load cmd — no goroutine leak, no explicit timer teardown.
- Key bindings (additive — list/detail/scheduler bindings unchanged):
  - `t` (detail mode) — open the log-tail pane for the run currently shown. No-op when `Detail.ID == 0` (e.g., `ErrNotFound` "no runs yet" state).
  - `esc` / `h` (tail mode) — return to the detail pane (not the list) so the user keeps the same agent context.
  - `r` (tail mode) — immediate manual refresh (does not reset `LastID`; just fires the load with the current watermark).
  - `q` / `Ctrl-C` — quit (every mode).
- Friend-tone copy: `loading log lines…` until the first chunk lands; `(no log lines yet — the run may not have produced any output)` for the empty-state path; `error: <message>` followed by `(polling continues — last successful chunk preserved above)` for the error path (polling does NOT freeze on transient DB errors — the existing accumulated lines stay visible).
- Render layout: bold header `buddy log tail — agent <id> — run #<N>`, then per-line `<HH:MM:SS> <level>  <message>`. Footer: `esc/h back · r refresh now · q quit · (auto-refresh ~1s)`.

Test coverage (`internal/tui/model_test.go` — 14 new race-clean tests, 65 total in the package; `internal/agent/store_test.go` — 3 new race-clean tests for `LogsSince`):

- Store: `LogsSince(_, _, 0)` returns everything in id ASC; sinceID is strictly less-than; empty result is `nil err + nil slice`; results are scoped by run_id (no cross-run leakage even when other-run ids fall above sinceID).
- Update reducer: `t` in detail switches to tail + locks `LogTailRunID = Detail.ID` + fires initial load (sinceID=0); `t` in detail when `Detail.ID == 0` is a no-op; `LogTailChunkMsg` appends and advances `LastID` to the last received ID; empty chunk is harmless (no `LastID` reset, no `Lines` mutation); `LogTailErrMsg` records err + flips `Loaded`; `LogTailTickMsg` dispatches a load cmd with the current `LastID` as sinceID; tick after navigation away is a no-op (self-cancellation); `r` fires immediate load; `esc` / `h` return to detail mode preserving context; `q` quits.
- View smoke: header includes run id + agent id; each line's message rendered; `Loaded=true` + 0 lines shows `no log lines` hint; pre-load shows `loading`; error state shows `error: <msg>`.

What stays open from W3-2 follow-on (after this):

- **Create form** (interactive spec builder) — HIGH cost, deserves a dedicated cycle.
- **In-app edit** (Spec YAML editor; needs new `Store.UpdateSpec` API).
- **Scheduler pane: live "currently running" indicator** (requires coupling the TUI to a running `Scheduler` instance).
- **Log tail: scrollback + auto-stop on run-end** (right now the pane shows all accumulated lines and polls forever; scrollback nav and "(run ended — polling stopped)" UX are follow-ons).

`docs/cli-buddy-spec.md` §9 W3-2 row updated with the 2026-05-17 live-log-tail follow-on entry.

## [0.7.0] — 2026-05-17

Bundle release of the four W3-2 follow-on items shipped since v0.6.6, plus the F5 log-retention command and the W3-6 reference webtoon agent example. Strictly additive — no breaking changes to specs, CLI flags, MCP tools, or stored DB rows.

What's in this release (newest commit first):

- `5aa89b3` — **W3-2 in-app delete with confirm**. `d` on a list row → modal y/N confirm → `Store.Delete` (FK cascade drops runs + logs). Default is cancel; nav keys inert in confirm; failure surfaces in the list error pane without optimistic removal.
- `9044aeb` — **W3-2 scheduler-preview pane**. `s` opens an inline pane showing when each scheduled agent would fire next. Preview-only (does not start cron / fire jobs); decoupled from any running `Scheduler` instance via the new `agent.PreviewSchedule(schedule, now) SchedulePreview` helper. Per-row parse errors surface inline.
- `67e571e` — **W3-2 detail view**. `enter`/`l` on a list row opens an inline pane with agent metadata + latest-run summary (started/ended/duration/exit code). `Store.LatestRun` was already implemented from v0.6.2; the `AgentLister` interface widened to expose it. Friend-tone "no runs yet" copy on `agent.ErrNotFound`.
- `b83a13d` — **W3-2 minimum-viable TUI** (originally landed pre-v0.7.0 cut but released here). Read-only agent list, vi/arrow navigation, refresh, AltScreen lifecycle.
- `8995577` — **F5 log retention** (`buddy agent purge --before <dur> [--apply]`). Dry-run by default; in-flight runs never deleted. FK cascade drops `agent_logs` rows transactionally.
- `ccc2170` — **W3-6 reference webtoon agent example** (`examples/webtoon-agent/spec.yaml` + README). Exercises Tier 1.4 backoff / 1.5 streaming / 1.6 scheduler refresh / 1.8 webhook + W3-3 chain. `ParseSpec` regression test gates the example.

Counts and gates:

- 5 version sources (Makefile, plugin.json, marketplace.json, server.go, main.go) all on `0.7.0` (`make verify-versions` passes).
- `go build ./...` clean; `go vet ./...` clean.
- `go test -race -count=1 -timeout=180s ./...` — 23 packages pass. `internal/tui` is at 51 tests (38 new since v0.6.6).
- `internal/agent` adds `PreviewSchedule` helper (+4 race-clean tests).
- `cli-buddy-spec.md §9` W3-2 row updated with three 2026-05-17 follow-on entries; W3-3/W3-4/W3-6 unchanged.

What stays open after v0.7.0:

- **W3-2 follow-on**: create form (interactive spec builder), live log tail (per-line `agent_logs` streaming), in-app edit.
- **Live "currently running" indicator on the scheduler pane** (requires sharing state with a running `Scheduler` instance).
- **Plugin v1.0.0 entry condition #2**: production dogfood (still user-paced).
- **B6 follow-up**: 43 PROCEDURE deviations (`make test-skill-form` still report-only).

### Added — W3-2 TUI in-app delete with confirm (follow-on of v0.6.6 minimum-viable)

The third W3-2 follow-on: pressing `d` on a list row opens a modal-style confirmation pane (`y` confirm, `N` / `esc` cancel — default is cancel). On confirm, the TUI calls `Store.Delete` for the locked-in ID; FK cascade drops the agent's runs + logs in the same transaction. Friend-tone success is silent — the row simply disappears on the post-delete reload.

What ships:

- `AgentLister` interface widened to include `Delete(ctx, id) error` (the existing `*agent.Store.Delete` already satisfies it). The interface docstring now flags that, despite the historical "Lister" name, the surface covers a destructive mutation too.
- New `Mode` value `ModeDeleteConfirm` + `PendingDeleteID` field on `tui.Model`. The ID is captured when `d` is pressed so a list refresh that lands while the dialog is open does not retarget the deletion to a different row.
- New reducer messages: `AgentDeletedMsg{ID}` and `AgentDeleteErrMsg{ID, Err}`. The delete cmd carries the ID in both messages so the reducer can produce error feedback (`delete agent "alpha": <wrapped>`) without reaching back into mutable state.
- Key bindings (additive — list/detail/scheduler bindings unchanged):
  - `d` (list mode) — open the confirm pane. No-op on an empty list. Cursor row's ID becomes `PendingDeleteID`.
  - `y` / `Y` (confirm mode) — fire the delete cmd.
  - `n` / `N` / `esc` (confirm mode) — cancel, return to list, clear `PendingDeleteID`.
  - `q` / `Ctrl-C` (every mode) — quit.
  - In confirm mode every other key (including nav keys) is intentionally inert — the user must explicitly answer y or n.
- Post-delete flow: on success, mode flips to `ModeList`, `Loaded` resets to `false`, and `loadAgentsCmd` is dispatched so the deleted row disappears on the next reducer tick. On failure, mode flips to `ModeList` and `m.Err` is set to `delete agent "<id>": <wrapped>` — the existing list-pane error state surfaces it. The list rows are left untouched (no optimistic removal).
- Render layout: bold header (`buddy agent — delete?`), the target ID called out explicitly (so a redraw can't trick the user into deleting the wrong row), a dim secondary line warning that runs + logs cascade with the delete, and an error-styled `press y to confirm · N / esc to cancel (default: cancel)` prompt. Footer hint `y confirm · n/esc cancel · q quit`.

Test coverage (`internal/tui/model_test.go` — 12 new race-clean tests, 51 total in the package):

- Update reducer: `d` enters confirm mode + captures cursor ID, `d` on empty list is a no-op, `y` schedules a delete cmd (success → `AgentDeletedMsg`, error → `AgentDeleteErrMsg`), `n` / `esc` cancel and clear `PendingDeleteID` + do NOT call Delete, `q` quits from confirm mode, `AgentDeletedMsg` returns to list + flips `Loaded=false` + dispatches reload cmd, `AgentDeleteErrMsg` returns to list and surfaces err via `m.Err` (and does not optimistically drop the row), nav keys (`j/k/g/G`) inert in confirm mode.
- View smoke: confirm pane renders the target agent ID + the `y/N` hint convention, list view surfaces a wrapped delete-error string in its existing error state.

What stays open from W3-2 follow-on:

- **Create form** (interactive spec builder vs. `buddy agent create` shell-out) — HIGH cost, deserves a dedicated cycle
- **Live log tail** (per-line streaming view of an in-flight run; pairs with v0.6.4 `agent_logs` streaming) — MED cost, polling vs `tea.Tick` channel decision deferred

`docs/cli-buddy-spec.md` §9 W3-2 row note updated with the 2026-05-17 in-app delete follow-on entry.

### Added — W3-2 TUI scheduler-preview pane (follow-on of v0.6.6 minimum-viable)

The second W3-2 follow-on item: pressing `s` in the list view opens an inline scheduler-preview pane that shows when each scheduled agent would fire next based on its cron expression. The pane is **preview-only** — it does not start cron, share state with a running `buddy agent scheduler` process, or fire any jobs. It exists so the user can sanity-check the schedule strings in their YAML specs without leaving the TUI.

What ships:

- New `agent.PreviewSchedule(schedule string, now time.Time) SchedulePreview` helper in `internal/agent/schedule_preview.go`. Uses the same `cron.ParseStandard` parser the scheduler uses (5-field cron + descriptors like `@daily` / `@every 30s`; second-precision intentionally off). `now` is passed in explicitly so callers stay deterministic and the helper is unit-testable. Parse failures are captured in `SchedulePreview.Err` rather than returned as a separate error — the TUI surfaces them inline on the offending row.
- New `Mode` value `ModeScheduler` on `tui.Model` + state fields (`SchedulerNow`, `SchedulerEntries []SchedulerPreviewEntry`, `SchedulerErr`, `SchedulerLoaded`). Two new reducer messages (`SchedulerStatusLoadedMsg{Now, Entries}` / `SchedulerStatusErrMsg{Err}`). Fetch runs via `loadSchedulerStatusCmd` — same `tea.Cmd` pattern as `loadAgentsCmd` / `loadDetailCmd`.
- Key bindings (additive — list/detail bindings unchanged):
  - `s` (list mode) — open the scheduler-preview pane. Filters agents with non-empty `schedule` only (on-demand agents already show up in the list).
  - `esc` / `h` (scheduler mode) — return to the list.
  - `r` (scheduler mode) — refetch the preview (re-captures `Now`, useful for "what's the next fire after I just edited a spec via another shell").
  - `q` / `Ctrl-C` — quit (works in every mode).
- Render layout: a `reference now: <RFC3339>` line at the top (so users know what clock the `next` times were computed against), then one row per scheduled agent showing `ID  schedule  next <RFC3339>`. Per-row parse failures render `<ID>  <bad-schedule>  invalid: <parser msg>` in the error style without wiping the rest of the pane.
- Friend-tone empty / loading / error copy: `loading scheduler preview…` until the fetch resolves; `(no scheduled agents — every agent in the list is on-demand)` for the empty case; `error: <message>` for a List() failure.

Test coverage (`internal/tui/model_test.go` — 12 new race-clean tests, 39 total in the package; plus 4 helper tests in `internal/agent/schedule_preview_test.go`):

- `PreviewSchedule`: 5-field cron rounds to next minute, `@daily` rolls to next 00:00, empty schedule is on-demand (no err), invalid string preserves the offending text + non-nil Err.
- Update reducer: `s` switches to scheduler + fires preview cmd, `SchedulerStatusLoadedMsg` folds state + clears any prior err, `SchedulerStatusErrMsg` records err + flips Loaded, `esc` / `h` return to list preserving cursor, `q` quits from scheduler mode, `r` refetches + flips Loaded back to false, `s` in detail mode is intentionally inert (no mode hijack).
- View smoke: pane renders agent IDs / schedules / esc hint, empty state shows `(no scheduled agents — …)`, parse-fail rows render inline without breaking the pane, loading placeholder visible before fetch resolves, `error: db locked` visible in the whole-pane error state.

What stays open from W3-2 follow-on (at scheduler-pane ship time):

- **Create form** (interactive spec builder vs. `buddy agent create` shell-out)
- **Live log tail** (per-line streaming view of an in-flight run; pairs with v0.6.4 `agent_logs` streaming)
- **In-app delete / edit** — *delete shipped above (2026-05-17); in-app edit still open*
- **Live "currently running" indicator on the scheduler pane** (requires sharing state with a running `Scheduler` instance; this preview iteration intentionally avoided that coupling)

`docs/cli-buddy-spec.md` §9 W3-2 row note updated with the 2026-05-17 scheduler-pane follow-on entry.

### Added — W3-2 TUI detail view (follow-on of v0.6.6 minimum-viable)

The v0.6.6 minimum-viable TUI shipped a read-only agent list and explicitly deferred the detail / create / log views. This change closes the first of those — pressing `enter` (or `l`) on a list row opens an inline detail pane showing the agent's metadata plus a summary of its most recent run (id, started/ended, duration, exit code, error). `esc` (or `h`) returns to the list with the cursor preserved.

What ships:

- `AgentLister` interface widened from a single-method (`List`) to two methods (`List` + `LatestRun`). The existing `*agent.Store` already implements `LatestRun` from v0.6.2, so the production wiring is transparent — only the test-side `fakeLister` had to grow a stubbed implementation.
- New `Model` fields: `Mode` (`ModeList` / `ModeDetail`), `Selected` (agent ID locked in for the pane), `Detail` (`agent.AgentRun`), `DetailErr`, `DetailLoaded`. Reducer test fixtures inspect these directly without going through a helper.
- New reducer messages: `AgentDetailLoadedMsg{Run}` and `AgentDetailErrMsg{Err}`. The detail fetch runs as a `tea.Cmd` (`loadDetailCmd`) so the reducer stays pure — same pattern v0.6.6 used for `loadAgentsCmd`.
- Key bindings (additive — list-mode shortcuts are unchanged):
  - `enter` / `l` / `→` (list mode) — open the detail pane for the cursor row. No-op on an empty list.
  - `esc` / `h` (detail mode) — back to the list, cursor + list state preserved.
  - `r` (detail mode) — refetch `LatestRun` for the locked-in agent (useful while a run is in-flight).
  - `q` / `Ctrl-C` (both modes) — quit.
  - `j` / `k` / `g` / `G` are scoped to list mode; in detail mode they are intentionally inert (the pane is read-only).
- Detail render shows: agent ID + optional `(name)` in the header, schedule (`(on-demand)` when empty), status, created/updated timestamps (RFC3339 UTC). For the latest run: `started`, `ended` (or `(in-flight)`), `duration`, `exit code`, and `error` when present.
- Friend-tone empty / loading / error copy: `loading latest run…` until the fetch resolves; `(no runs yet — try \`buddy agent run <id>\`)` when `LatestRun` returns `ErrNotFound`; `error: <message>` for any other error. Footer hint switches per-mode (`enter/l detail · …` in list, `esc/h back · r refresh · q quit` in detail).

Test coverage (`internal/tui/model_test.go` — 14 new race-clean tests, 27 total in the package):

- Update reducer: `enter` switches to detail + fires `LatestRun` cmd, `enter` on empty list is a no-op, `l` aliases `enter`, `AgentDetailLoadedMsg` folds into state, `AgentDetailErrMsg` records error + flips `DetailLoaded`, `esc` returns to list preserving cursor, `h` aliases `esc`, `q` quits from detail, list nav keys (`j/k/g/G`) are inert in detail mode, `r` in detail mode refetches `LatestRun`.
- View smoke: detail pane renders agent fields (ID / name / schedule / status / exit code label / esc footer hint), `ErrNotFound` produces the "no runs yet" copy + `buddy agent run` hint, generic errors produce `error: <message>`, `DetailLoaded=false` renders the loading placeholder.

What stays open from W3-2 follow-on (at detail-view ship time):

- **Create form** (interactive spec builder vs. `buddy agent create` shell-out)
- **Scheduler status pane** (`buddy agent scheduler status` inline) — *shipped above as scheduler-preview pane, 2026-05-17*
- **Live log tail** (per-line streaming view of an in-flight run; pairs with v0.6.4 `agent_logs` streaming)
- **In-app delete / edit** (currently the user shells out to `buddy agent delete`)

`docs/cli-buddy-spec.md` §9 W3-2 row note updated: minimum-viable + detail view shipped 2026-05-15.

### Added — W3-2 TUI minimum-viable (cli-buddy-spec §9)

`buddy tui` is the new entry point for the long-deferred W3-2 terminal UI. v0.6.6 ships the *minimum-viable* subset — a read-only agent list with vi-style navigation — using `charmbracelet/bubbletea` + `charmbracelet/lipgloss`. Detail view, create form, scheduler status, and live log tail land in W3-2 follow-on cycles, matching the partial-Done pattern used for W3-3 / W3-4.

What ships:

- New `internal/tui/` package: `Model` / `Update` / `View` with a pure-function reducer. The `AgentLister` interface narrows the Store dependency to a single `List(ctx)` method so reducer tests inject canned data without touching SQLite.
- New `buddy tui [--db <path>]` subcommand wires `tea.NewProgram(tui.NewModel(store), WithAltScreen, WithContext)`. AltScreen preserves the user's prior shell content; `q` or `Ctrl-C` exits cleanly and restores the terminal.
- Key bindings:
  - `j` / `↓` — move cursor down (clamped at last row)
  - `k` / `↑` — move cursor up (clamped at row 0)
  - `g` / `Home` — jump to first
  - `G` / `End` — jump to last
  - `r` — refresh (re-reads the store; useful when `buddy agent create` runs in another shell and the scheduler refresh has not yet ticked)
  - `q` / `Ctrl-C` — quit
- Empty / loading / error states each have a friend-tone copy: `loading agents…`, `(no agents yet — run \`buddy agent create <spec.yaml>\` in another shell)`, `error: <message>`. Footer hint (`j/k or ↑/↓ move · g/G top/bottom · r refresh · q quit`) is always visible.
- Cursor clamps when the list shrinks across reloads — deleting an agent while the cursor sat on it doesn't leave the cursor pointing past the end.

Test coverage (`internal/tui/model_test.go` — 13 race-clean tests):

- Update reducer: q quits, Ctrl-C quits, AgentsLoadedMsg folds into state, cursor clamps on list shrink, ErrMsg records error + marks Loaded, navigation respects bounds (j past end stays at end, k past top stays at 0), g/G jump to ends, r flips Loaded to false and returns a reload Cmd, WindowSizeMsg tracks width/height.
- View smoke: loading placeholder visible before first load, empty-state copy visible, agent rows render with ID + schedule (or `(on-demand)`) + status, cursor marker `▸` present on the selected row, error string visible in the error state.

`AltScreen` + `tea.QuitMsg` lifecycle is exercised through bubbletea's own goroutine model — the production path is not unit-tested, but bubbletea's contract is well-established and the reducer / View tests cover the Model surface.

What stays open (W3-2 follow-on — at v0.6.6 ship time):

- **Detail view** (`buddy agent show <id>` equivalent inline) — *shipped above, 2026-05-15*
- **Create form** (interactive spec builder vs. `buddy agent create` shell-out)
- **Scheduler status pane** (`buddy agent scheduler status` inline)
- **Live log tail** (per-line streaming view of an in-flight run; pairs with v0.6.4 `agent_logs` streaming)
- **In-app delete / edit** (currently the user shells out to `buddy agent delete`)

New direct deps:

- `github.com/charmbracelet/bubbletea` (MIT)
- `github.com/charmbracelet/lipgloss` (MIT)

(Plus transitive `charmbracelet/x/*`, `mattn/go-runewidth`, `muesli/termenv`, etc. — all MIT or BSD-licensed.)

`docs/cli-buddy-spec.md` §9 W3-2 row marked "partial Done 2026-05-13 — minimum-viable subset". This makes W3-2 the last `cli-buddy-spec` line item to enter Done state (every other W3-x cell now reads ✅ Done or partial Done).

### Added — agent log retention (verify-quality F5 follow-on)

The v0.6.4 verify-quality audit flagged F5 as a *medium* open finding: every line the executor emits becomes one `INSERT` into `agent_logs`, and there's no retention path — long-running steps over many days accumulate row counts without bound. This change adds a manual retention command. Auto-purge / batching remain follow-ons pending real-dogfood signal.

- **`buddy agent purge --before <duration> [--apply]`** — new subcommand. `--before` accepts the same shapes as `buddy purge` (relative `30d`, date `2026-04-01`, RFC 3339). Dry-run by default (prints the count of runs that *would* be deleted); pass `--apply` to actually perform the delete. In-flight runs (`ended_at IS NULL`) are never deleted regardless of cutoff.
- **`Store.CountRunsBefore(ctx, threshold)`** — preview helper. Counts `agent_runs` rows finished before `threshold`; in-flight runs excluded.
- **`Store.PurgeRunsBefore(ctx, threshold)`** — delete helper. Drops the matching `agent_runs`; the existing `agent_logs.run_id ... ON DELETE CASCADE` constraint (migration v4) drops the matching log lines in the same transaction.

Test coverage (3 new race-clean store tests):

- `CountRunsBefore` excludes in-flight runs (`ended_at IS NULL`) and respects strict less-than against the cutoff.
- `PurgeRunsBefore` deletes only the qualifying runs, preserves recent + in-flight runs, and the FK cascade actually drops the log rows transactionally (`SELECT COUNT(*) FROM agent_logs` drops by the expected amount).
- `PurgeRunsBefore` on an empty DB returns `(0, nil)` rather than an error.

CLI smoke verified against an empty DB (dry-run + apply both produce friendly zero-count messages).

What stays open from F5:

- **Per-line batching** — `AppendLog` still does one `INSERT` per line. Acceptable for moderate output volumes; the right batching design needs measurement against a real workload (Tier 6.1 dogfood would surface it). Tracked as a follow-on rather than guessed at here.
- **Auto-purge** — no scheduled cleanup yet. Users with `cron` available can wrap the manual command (`0 4 * * * buddy agent purge --before 30d --apply`). A built-in periodic purge is a future cycle once the right default cadence + threshold becomes clear from dogfood.

### Added — W3-6 reference webtoon agent example

`examples/webtoon-agent/` is the canonical end-to-end agent spec from `cli-buddy-spec.md` §2.2 — a daily-scheduled chain (`concretize-idea → write-prd → design-system → build-feature`) that publishes its `RunResult` to a downstream webtoon API via webhook. With v0.6.3 (exponential backoff) + v0.6.4 (streaming logs) + v0.6.5 (scheduler live refresh) + v0.6.6 (webhook output) shipped, the example exercises every cli buddy capability in one spec.

- `examples/webtoon-agent/spec.yaml` — 4-step chain, daily cron (`0 3 * * *`), exponential retry (1s base, 2m cap, 5 attempts), webhook POST to `https://webtoon-by-ai.example.com/api/v1/episodes` with `Authorization` + `X-Source` + `X-Agent-Version` headers and a 90s timeout. Inline comments call out which v0.6.x feature each section depends on.
- `examples/webtoon-agent/README.md` — usage walkthrough: customise the spec (template the Authorization header out-of-process), register via `buddy agent create`, start `buddy agent scheduler start --refresh 1m`, observe progress via `buddy agent log webtoon-publish`, modify-without-restart via the scheduler refresh, expected webhook payload shape, troubleshooting matrix.
- `internal/agent/runtime_test.go` gains `TestParseSpec_ReferenceWebtoonAgentValidates` — reads `examples/webtoon-agent/spec.yaml` from disk, runs it through `ParseSpec`, and asserts every notable field (schedule, chain length, exponential backoff strategy + 2-minute cap, webhook type + URL scheme + 90s timeout + Authorization header). Functions as a regression gate so a future schema change can't silently break the shipped example.

`docs/cli-buddy-spec.md` §9 W3-6 row marked Done.

## [0.6.6] — 2026-05-12

Ships the cli buddy `[Unreleased]` work accumulated after v0.6.5: Tier 1.8 webhook output target. With v0.6.4 streaming logs + v0.6.5 scheduler live refresh already in place, this closes the last dependency of the cli-buddy-spec §2.2 reference webtoon agent (W3-6) — agents can now self-publish to downstream services without a wrapper script.

### Added — webhook output target (Tier 1.8)

`OutputTarget` gains a third type, `webhook`, alongside the existing `stdout` and `file`. Agents whose YAML declares `output: { type: webhook, url: ... }` now have their `RunResult` JSON POST'ed (or PUT'ed / PATCH'ed) to the configured URL after the chain finishes. This is the dependency the cli-buddy-spec §2.2 reference webtoon agent (W3-6) was waiting on — agents can now self-publish to downstream services without a wrapper script.

Schema additions on `OutputTarget`:

| field | meaning | default |
|---|---|---|
| `type: webhook` | new output mode | — |
| `url` | POST/PUT/PATCH target (required) | — |
| `method` | HTTP verb | `POST` |
| `headers: { K: V, ... }` | extra HTTP headers (`Content-Type` auto unless overridden) | `{}` |
| `timeout: 30s` | per-request HTTP timeout | `30s` |

Example spec:

```yaml
id: webtoon-publish
name: "Daily webtoon publish"
schedule: "@daily"
chain:
  - command: build-feature
    args: "today's strip"
output:
  type: webhook
  url: https://webtoon-by-ai.example.com/api/v1/episodes
  method: POST
  headers:
    Authorization: "Bearer ${WEBTOON_API_TOKEN}"   # literal — template before agent create
    X-Source: buddy-agent
  timeout: 60s
```

Implementation:

- `runtime.go writeOutput` dispatches `case "webhook"` to a new `postWebhook(target, result)` helper.
- `postWebhook` serialises `RunResult` to JSON via `json.Marshal`, builds an `http.Request` with the configured method / headers / context-timeout, and POSTs it. Non-2xx responses surface as an `agent: webhook %s %s returned %s: %s` error (with up to 512 bytes of the response body included for diagnosability — server-side JSON error envelopes show up in the agent_logs warn line).
- `spec.go` rejects `type: webhook` without a URL at `ParseSpec` time, and rejects URL schemes other than `http://` / `https://` (no `ftp:`, `file:`, `javascript:` slipping into the HTTP client).
- Header values are written *literally* — secret expansion (e.g. `${ENV}`) is intentionally out of scope. Specs that need an Authorization secret should template the YAML before `buddy agent create` rather than commit the secret to source.

Tests (`internal/agent/runtime_test.go` + `spec` extension, 5 new race-clean):

- happy path against `httptest.Server`: default POST method, default Content-Type=application/json, custom Authorization header round-trips, request body parses back into the runtime's `RunResult`, server hit exactly once per run.
- spec can override method (`PUT`) and `Content-Type` (`application/vnd.buddy+json`).
- 500 response surfaces as a warn-level `agent_logs` line on the run; step itself still succeeds (output dispatch failure does not flip step exit code).
- `ParseSpec` rejects `type: webhook` without a URL.
- `ParseSpec` rejects URLs with disallowed schemes.

v0.3 contract preserved: webhook dispatch happens *after* the chain finishes; `ExitCode` still drives step success / failure. A failed webhook never converts a green run into a red one — it shows up in logs for diagnosis.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.6.5` → `0.6.6`.
- `internal/mcp/server.go` MCP server `Version` `0.6.5` → `0.6.6`.
- `cmd/buddy/main.go` `var version` `0.6.5` → `0.6.6`.
- `Makefile` `RELEASE_VERSION` `0.6.5` → `0.6.6`.
- `README.md` install snippet + sample output bumped to `0.6.6`.

### Versioning policy note

Tier 1.8 adds a new YAML schema option (`output.type: webhook` + four new sibling fields) — strictly additive. Existing agent specs with `output.type: stdout` or `output.type: file` are completely unaffected; running them produces the same bytes as v0.6.5. Per the v0.6.3 / v0.6.4 / v0.6.5 precedent of treating backward-compatible additions as patch, shipping as `v0.6.5 → v0.6.6`. ADR-004 §2.3 minor-bump trigger would activate if a future change *removed* a webhook escape hatch or changed the default header behavior.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches and upgrades to 0.6.6. Skill catalog + command surface unchanged from 0.6.x (148 skills / 99 commands).
- **cli binary users**: pull v0.6.6 if you want to declare `output.type: webhook` in agent specs. v0.6.5 binaries continue to work; DB stays compatible.
- **Existing agent specs**: no edit required. The new `OutputTarget` fields (`URL`, `Method`, `Headers`, `Timeout`) are opt-in.
- **Webhook target operators**: expect a JSON-encoded `RunResult` body via POST (or the configured method), Content-Type `application/json` unless the spec overrides it. Bodies up to ~64 KiB are normal; very long Stdout/Stderr captures from streaming runs can produce larger payloads.

## [0.6.5] — 2026-05-12

Ships the cli buddy `[Unreleased]` work accumulated after v0.6.4: scheduler live refresh (Tier 1.6 — backward-compatible, opt-out via `--no-refresh`). Closes the last open `cli-buddy-spec.md` §9 W3-3 follow-on item.

### Added — scheduler live refresh (Tier 1.6)

Before v0.6.4, `buddy agent scheduler start` loaded the agent set once at startup and never re-read the store. Adding / deleting an agent (or editing its `schedule`) while the scheduler was running had no effect until the user killed and restarted the scheduler. `cli-buddy-spec.md` §9 W3-3 explicitly listed this as a follow-on. This change closes it.

How it works:

- New `SchedulerOptions.RefreshInterval` (default 1m; tests pass 20ms) and `SchedulerOptions.RefreshDisabled` (escape hatch for the v0.6.4 load-once behavior).
- `Scheduler.tracked map[agentID]trackedEntry` records which cron entry each registered agent owns plus the schedule string it was registered with. Guarded by `trackedMu`.
- `Scheduler.refreshOnce(ctx)` is the diff engine: it lists every agent, registers the new ones, drops the deleted ones, and re-registers the ones whose `schedule` field changed (`cron.Remove(old) + AddFunc(new)`). Idempotent — calling it twice in a row with no DB changes produces an empty diff. `Load` is now a thin wrapper over `refreshOnce` so the initial-load path and the polling path share their logic.
- `Scheduler.Start` launches a `pollLoop` goroutine when `refreshInterval > 0`. The loop tickers at the configured cadence, calls `refreshOnce`, logs the `+N / -N / ~N` diff (added / removed / updated) only when anything actually moved, and exits cleanly when `ctx` is cancelled. Transient DB read failures log an error and continue rather than freezing the loop.
- `cmd/buddy/agent_cmd.go` exposes the cadence on the CLI: `buddy agent scheduler start --refresh 30s` (custom), `buddy agent scheduler start --no-refresh` (disabled).

What this enables in practice:

```bash
# shell A
$ buddy agent scheduler start --refresh 10s
scheduler: started (location=Asia/Seoul, entries=2, refresh=10s)

# shell B (any time)
$ buddy agent create new-spec.yaml
created agent "new-spec"

# shell A picks it up within ~10s:
# scheduler: agent "new-spec" scheduled (cron="@every 1h" entry_id=3)
# scheduler: refresh poll applied (+1 / -0 / ~0)
```

Tests (`internal/agent/scheduler_test.go` — 4 new race-clean tests):

- New agent created post-Start surfaces in `Entries()` within the deadline; log line records the registration.
- Deleted agent disappears from `Entries()` on the next poll; log line records the unschedule.
- Schedule change (simulated via direct DB `UPDATE agents SET schedule = …`) re-registers the agent under a fresh cron entry ID so stale ticks don't keep firing.
- `RefreshDisabled = true` keeps `Entries()` empty even after a post-Start `Create` — opt-out works and startup log shows `refresh=disabled`.

v0.3 contract preserved: the cron schedules themselves stay minute-precision, in-flight overlap is still dropped by the per-agent atomic flag, and on-demand agents (no `schedule` field) are still invisible to the cron tick.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.6.4` → `0.6.5`.
- `internal/mcp/server.go` MCP server `Version` `0.6.4` → `0.6.5`.
- `cmd/buddy/main.go` `var version` `0.6.4` → `0.6.5`.
- `Makefile` `RELEASE_VERSION` `0.6.4` → `0.6.5`.
- `README.md` install snippet + sample output bumped to `0.6.5`.

### Versioning policy note

Tier 1.6 changes the *default* scheduler behavior (live refresh is now on, polling every 60s). By ADR-004 §2.3 a default change leans toward minor — but the new behavior only *adds* automatic catch-up of DB-side agent edits, and `--no-refresh` (or `RefreshDisabled: true`) restores v0.6.4 behavior byte-identically. No existing agent spec runs differently. Shipping as patch (`v0.6.4 → v0.6.5`) at user direction, consistent with the v0.6.3 / v0.6.4 precedent of treating opt-out-restorable additions as patch. The next user-visible change that *removes* an escape hatch will reset to the §2.3 default.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches and upgrades to 0.6.5. Skill catalog + command surface unchanged from 0.6.x (148 skills / 99 commands).
- **cli binary users**: pull v0.6.5 if you run `buddy agent scheduler start` and want it to pick up agents created/deleted by another shell automatically. v0.6.4 binaries continue to work; DB stays compatible.
- **Existing `buddy agent scheduler start` users**: behavior change at the default level — the scheduler now polls the store every 60s and reconciles its entry set with the DB. Pass `--no-refresh` to keep the v0.6.4 load-once behavior.
- **CI / scripted callers**: the new flags are additive (`--refresh <duration>`, `--no-refresh`); existing invocations without these flags continue to work — they pick up the 60s default poll.

## [0.6.4] — 2026-05-12

Bundles the cli buddy `[Unreleased]` work that accumulated after v0.6.3: Tier 1.5 streaming log capture (new behavior surface, strictly additive) plus the verify-quality-driven cleanup of the streaming-path internals (F1–F4, pure refactor). One feature + one quality follow-up that together exercise the full *build → verify → fix → re-verify → ship* loop the buddy plugin describes.

### Added — streaming log capture for agent step runs (Tier 1.5)

Before v0.6.4, `SubprocessExecutor.Run` collected the child process's stdout / stderr into a single `bytes.Buffer` and returned only after the step finished. For a step that takes minutes (e.g. a Claude Code subprocess walking a 12-stage PROCEDURE) the user saw nothing in `buddy agent log <id>` until the very end. This change adds line-by-line streaming so each line surfaces in `agent_logs` *as it is emitted*.

Implementation:

- New `agent.LogSink` callback type — `func(stream, line string)`. `stream` is the literal `"stdout"` or `"stderr"`.
- `Executor.Run` signature extended with `sink LogSink` argument. **Backward-compatible at the runtime level** — `sink == nil` falls back to the v0.6.3 bulk-buffer behavior byte-identically.
- `SubprocessExecutor.Run` uses `StdoutPipe` + `StderrPipe` + `bufio.Scanner` in two goroutines (one per stream, sync.WaitGroup'd before `cmd.Wait()`) when `sink != nil`. Per-line `sink(stream, line)` calls happen synchronously with the scanner read. Scanner buffer caps at 1 MiB so very large JSON-formatted PROCEDURE outputs don't split mid-record.
- `MockExecutor.Run` simulates streaming by splitting the canned `Stdout` / `Stderr` on `\n` and emitting one sink call per line — making Runtime-level tests possible without spawning an actual subprocess.
- `Runtime.runOneStep` constructs a sink that forwards every line to `Store.AppendLog` with level `info` for stdout and `warn` for stderr. Each entry is formatted `step[N] <cmd> <stream>: <line>` so `buddy agent log <id>` shows them inline with the existing `attempt=N`, `self-check=…`, `next-phase branch: …` lines.

Tests (`internal/agent/executor_test.go` — new file, 8 tests; `runtime_test.go` — 2 new tests):

- `MockExecutor` nil sink keeps v0.6.3 behavior; non-nil sink emits per-line in stdout-then-stderr order; empty streams yield zero sink calls; trailing newline doesn't double-emit.
- `splitLines` table test covers `""`, `"a"`, `"a\n"`, `"a\nb"`, `"a\nb\n"`, `"\n"`.
- `SubprocessExecutor` real-subprocess tests against `/bin/sh -c 'printf …'`: sink receives each printed line in order, captured strings round-trip the full content, fast path (`sink == nil`) bulk-captures identically, non-zero exit codes still translate to `(code, nil)` rather than an error.
- `Runtime`: a mock step with multi-line stdout + stderr produces one `agent_logs` entry per line (3 stdout + 1 stderr), levels are `info` and `warn` respectively, and `StepResult.Stdout` / `StepResult.Stderr` retain the full captured strings unchanged.

v0.3 contract preserved: streaming is metadata. Step success / failure still tracks `ExitCode` from the executor; no behavior change in the retry loop or in how subsequent steps are gated.

Bumps the `cmd/buddy/agent log <agent-id>` command from "shows the final tail of a finished run" to "shows every line as the run progresses" — most useful when paired with the exponential backoff added in v0.6.3, since longer waits between retries make mid-progress visibility more valuable.

### Refactored — `SubprocessExecutor` cleanup (verify-quality F1–F4 follow-up)

`/buddy:verify-quality` on the v0.6.3 → Tier 1.5 streaming commit flagged four minor cleanups in `internal/agent/executor.go`. None block the quality gate, but cleaning them up keeps the streaming path readable for the next contributor. All four are behavior-preserving (zero diff in test outcomes — every streaming test stays byte-identical green).

- **F1** — Removed the `defer r.Close()` from `streamLines`. `exec.Cmd.Wait()` closes both pipes automatically once the child exits, and the caller (`Run`) calls `Wait` after this goroutine joins. The explicit close was redundant and would confuse a future reader. Added a doc comment in `streamLines` explaining why the pipe close stays implicit.
- **F2** — Closed `stdoutPipe` / `stderrPipe` explicitly in the spawn-error paths (`StderrPipe()` fail / `cmd.Start()` fail). Previously a spawn failure would leak the acquired pipe FDs until GC. Process-spawn failure is rare in practice but the leak path was real.
- **F3** — Dropped the two unused `*bytes.Buffer` parameters from `translateExitCode`. The buffers were carried over from an earlier draft; the function only ever read `runErr`. Signature is now `translateExitCode(error) (int, error)`.
- **F4** — Simplified the streaming-path return: dropped the `stdoutStr, stderrStr := stdout.String(), stderr.String()` intermediate, the `_ = exitErr` no-op, and a stale comment about `translateExitCode` returning four values. The path now reads `exitCode, exitErr := translateExitCode(cmd.Wait()); return stdout.String(), stderr.String(), exitCode, exitErr` — a direct mirror of the fast path's shape.

Net diff: `internal/agent/executor.go` loses ~10 lines of noise, gains ~6 lines of error-path cleanup + doc — readability higher, FD-leak surface smaller, behavior unchanged.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.6.3` → `0.6.4`.
- `internal/mcp/server.go` MCP server `Version` `0.6.3` → `0.6.4`.
- `cmd/buddy/main.go` `var version` `0.6.3` → `0.6.4`.
- `Makefile` `RELEASE_VERSION` `0.6.3` → `0.6.4`.
- `README.md` install snippet + sample output bumped to `0.6.4`.

### Versioning policy note

Tier 1.5 (streaming log capture) introduces a new behavior surface: the `Executor` interface gains a `sink LogSink` parameter, `Runtime` writes per-line streaming entries into `agent_logs`, and `buddy agent log <agent-id>` now shows mid-progress rather than only the post-step summary. By ADR-004 §2.3 this is borderline minor — but the surface is *strictly additive at the user level*:

- Existing agent specs run identically (the runtime always passes a non-nil sink internally, but the sink only *adds* log rows; nothing the executor returned before is gone or shaped differently).
- External Executor implementers do see a breaking signature change, but no external implementers exist today (the buddy plugin is the only consumer).
- `StepResult.Stdout` / `StepResult.Stderr` retain the full captured strings byte-identically.

Shipping as a patch (`v0.6.3 → v0.6.4`) at user direction, consistent with the v0.6.3 precedent of treating backward-compatible additions as patch. The F1–F4 cleanup contributes zero to the SemVer decision (internal refactor only).

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches and upgrades to 0.6.4. Skill catalog + command surface unchanged from 0.6.x (148 skills / 99 commands).
- **cli binary users**: pull v0.6.4 if you want mid-progress visibility in `buddy agent log <agent-id>` for long-running steps. v0.6.3 binaries continue to work; the on-disk DB stays compatible.
- **External Executor implementers**: the `Executor.Run` signature gains a trailing `sink LogSink` argument. Implementations need a one-line update (pass `nil` to keep v0.6.3 behavior, or use the sink to surface progress). No such implementers exist in the wild yet — internal-only change in practice.
- **`agent_logs` consumers**: expect new info / warn-level lines of the form `step[N] <cmd> stdout: <line>` and `step[N] <cmd> stderr: <line>` interleaved with the existing attempt / self-check / next-phase entries. Parsers that match exact line prefixes may need to recognise the new shapes; consumers using SQL filters on `level` or substring searches keep working unchanged.

## [0.6.3] — 2026-05-12

Bundles the cli buddy `[Unreleased]` work that accumulated after v0.6.2: exponential retry backoff (Tier 1.4 — new YAML field, backward-compatible) plus the `cmd/buddy/main.go` decomposition refactor (W3-5 retrofit). No release noise besides version bumps and notes.

### Added — exponential backoff for agent step retries (Tier 1.4)

Agent specs can now ask the runtime to grow the wait between retries instead of using a fixed delay. Useful for steps that hammer a flaky upstream (rate limit, transient 5xx) — the next attempt waits 2×, 4×, 8× the base delay until a cap kicks in.

YAML schema additions (backward-compatible — existing specs continue to fixed-delay):

```yaml
retry:
  max_attempts: 5
  backoff_delay: 500ms          # base delay (existing field)
  backoff_strategy: exponential # NEW: "fixed" (default) | "exponential"
  backoff_max: 8s               # NEW: cap when strategy=exponential, 0 = uncapped
```

Implementation:

- `RetryPolicy.BackoffStrategy` (`""` defaults to `"fixed"` — v0.6.x specs need no edit) + `RetryPolicy.BackoffMax` (`0` = uncapped).
- `agent.BackoffStrategyFixed` / `agent.BackoffStrategyExponential` constants exposed so callers can refer to them by name.
- `computeBackoff(retry, attemptJustFailed)` pure helper:
  - fixed → `BackoffDelay`
  - exponential → `BackoffDelay * 2^(attemptJustFailed - 1)`, capped at `BackoffMax` when set
  - exponent clamped to 30 internally so a misconfigured `MaxAttempts=50` cannot overflow `time.Duration`'s int64 range
- Runtime emits a new `step[N] <cmd> backoff <duration> before attempt <N+1>` info log line per retry, so `buddy agent log <id>` shows when and for how long the runtime is waiting.
- Spec validation rejects unknown strategies (e.g. `gaussian-random`) and inverted configs (`backoff_max < backoff_delay`) with friendly errors.

Tests:

- `internal/agent/backoff_test.go` — 12 race-clean unit tests covering nil policy, zero base, fixed/empty strategy (returns base), exponential doubling, cap kicks in at the right step, uncapped path, large-attempt overflow safety, unknown-strategy graceful fallback, and 4 spec-parsing happy/error paths.

v0.3 contract preserved: ExitCode still drives step success / failure; backoff only affects *when* the next attempt runs. Existing specs with `backoff_strategy` omitted behave byte-identically to v0.6.2.

### Changed — split `cmd/buddy/main.go` (W3-5 retrofit)

The `cmd/buddy/main.go` file had grown to 688 lines hosting eight unrelated sub-feature wirings (events / stats / doctor / install / uninstall / daemon-tree / hookwrap / boilerplate). `coding-style.md` recommends ≤400 lines per file and warns that mixing domains in one file is a single-responsibility violation regardless of size. The next subcommand addition would have pushed `main.go` past 800.

Decomposition into sibling `<feature>_cmd.go` files matches the pre-existing `config_cmd.go` / `feature_cmd.go` / `mcp_cmd.go` / `purge_cmd.go` pattern. No behavior change — pure refactor; all 20 packages stay race-clean under `go test -race -count=1 -timeout=120s ./...`.

| New file | Lines | Functions moved |
|---|---|---|
| `events_cmd.go` | 72 | `newEventsCmd` |
| `stats_cmd.go` | 55 | `newStatsCmd` |
| `doctor_cmd.go` | 52 | `newDoctorCmd` |
| `install_cmd.go` | 120 | `newInstallCmd`, `newUninstallCmd`, `translateInstallError` |
| `daemon_cmd.go` | 221 | `newDaemonCmd` + 4 subs (`run` / `start` / `stop` / `status`) + `resolvePIDFile`, `defaultPIDFromDB`, `spawnDetached` |
| `hookwrap_cmd.go` | 85 | `newHookWrapCmd`, `parseTags` |

`agent.go` → `agent_cmd.go` rename for naming consistency (`git mv` preserves history).

`main.go` after the split is 147 lines — boilerplate (`main()`, `newRootCmd()`, version helpers, error helpers) only. Import list trimmed accordingly (`context`, `errors`, `fmt`, `os`, `cobra`, `config`, `db`, `persona` — was 9 stdlib + 9 internal).

Resolves `docs/HANDOFF.md` §11 "Immediate small things" item: *`cmd/buddy/main.go` 분할 (685 lines)*.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.6.2` → `0.6.3`.
- `internal/mcp/server.go` MCP server `Version` `0.6.2` → `0.6.3`.
- `cmd/buddy/main.go` `var version` `0.6.2` → `0.6.3`.
- `Makefile` `RELEASE_VERSION` `0.6.2` → `0.6.3`.
- `README.md` install snippet + sample output bumped to `0.6.3`.

### Versioning policy note

ADR-004 §2.3 treats a new YAML schema field (Tier 1.4's `backoff_strategy` + `backoff_max`) as a *user-visible surface addition* that would normally pull this release into a minor (`0.7.0`). Shipping as a patch (`v0.6.2 → v0.6.3`) at user direction because the addition is *strictly backward-compatible*: every existing spec (no `backoff_strategy` field, or `backoff_strategy: fixed`) gets byte-identical retry timing to v0.6.2. The new fields are opt-in. Future user-visible additions that change defaults or remove backward-compat will reset to the §2.3 default.

The bundled W3-5 main.go split is internal refactor only (file layout, zero behavior change) so contributes nothing to the SemVer decision.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches and upgrades to 0.6.3. Skill catalog + command surface unchanged from 0.6.x (148 skills / 99 commands).
- **cli binary users**: optional bump. v0.6.2 binaries continue to work. Pull v0.6.3 if you want the exponential backoff schema in your agent specs, or for the matching version string.
- **Existing agent specs**: no edit required — `backoff_strategy` omitted is treated as `"fixed"` and the retry timing matches v0.6.2 exactly.
- **Plugin / cli contributors**: `cmd/buddy/main.go` is now 147 lines; new subcommands belong in their own `<feature>_cmd.go` sibling (events / stats / doctor / install / daemon / hookwrap / agent already follow this pattern).

## [0.6.2] — 2026-05-12

Bundles two release-pipeline hardening rounds (go-version + 5-version-sources drift checks) with a long-awaited `buddy agent log` viewer and a push/PR CI gate. All four entries derive from the v0.6.1 root-cause analysis surfacing how thin the test/sanity coverage on `main` had been before tag time.

### Added — release-pipeline drift sanity checks

The first v0.6.1 tag failed CI because `go.mod`'s `go` directive (1.25.0, bumped in `becbcf1` on 2026-05-11) had been silently out of sync with `.github/workflows/release.yml`'s `go-version: '1.22'` across four prior releases. `setup-go@v5`'s `GOTOOLCHAIN=auto` default had been auto-downloading the 1.25 toolchain on every CI run, hiding the drift. `setup-go@v6` exports `GOTOOLCHAIN=local` and removes that fallback, so the next tag after the action bump was the first one to break. v0.6.2 hardens against recurrence on both that axis and a sibling axis (the 5 version sources).

- **`make verify-go-version`** — new Makefile target. Reads `go.mod`'s `go` directive (e.g. `1.25.0`) and `release.yml`'s `go-version` input (e.g. `1.25`), compares major.minor, and exits non-zero with an explicit `::error::` line + remediation hint on mismatch. Dry-run on a simulated drift (`go-version: '1.22'`) produces the exact message future contributors will see.
- **`make verify-versions`** — new Makefile target. Reads `RELEASE_VERSION` + four other version sources (`plugin/.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`, `internal/mcp/server.go` Version, `cmd/buddy/main.go` var version) and reports any disagreeing source by name. Half-bumped releases that would publish e.g. a plugin manifest at `0.6.2` with a cli binary self-reporting `0.6.1` now fail CI before cross-compile.
- **`.github/workflows/release.yml`** runs both checks between `Set up Go` and the existing `Verify tag matches Makefile RELEASE_VERSION` step. Any future drift on either axis fails fast rather than silently regressing.
- **`README.md` + `docs/HANDOFF.md`** Stack / requirement lines now point at `go.mod` as the source of truth alongside the explicit minimum, mirroring the workflow check so doc updates stay grouped.

The pattern (drift accumulates → external dependency change surfaces it as a hard failure) is generic. The tag↔Makefile axis was already covered (`Verify tag matches Makefile RELEASE_VERSION`); v0.6.2 adds the go.mod↔workflow and Makefile↔(plugin.json+marketplace.json+server.go+main.go) axes. Other axes (e.g. README install snippet ↔ Makefile) remain candidates for follow-on hardening if a real drift surfaces.

### Added — `buddy agent log <agent-id>` subcommand

The v0.5.0 / v0.6.0 cycles added three new log line classes to `agent_logs` (per-step `self-check=<verdict> (M/T passed)`, `next-phase candidates: <list>`, and `next-phase branch: "<cond>" → <skills>`), but the CLI had no way to surface them. Users had to `sqlite3 buddy.db 'SELECT * FROM agent_logs WHERE run_id = ?'` to read them. v0.6.2 ships a friendly viewer.

- **`buddy agent log <agent-id> [--limit N] [--db path]`** — resolves the agent's most recent run, prints `Run ID`, `Started` / `Ended` / `Exit code` / `Error` header, then each log line in `YYYY-MM-DD HH:MM:SS  <level>  <message>` form (oldest first). Empty-state friendly: an agent that exists but has never run yields `agent <id> has no runs yet. Try \`buddy agent run <id>\``; an unknown agent surfaces the same `agent.ErrNotFound` path.
- **`Store.LatestRun(ctx, agentID)`** — new method backing the subcommand. Returns the highest-id row from `agent_runs` keyed by `agent_id`, or `ErrNotFound` when there is none.
- **`internal/agent/store_test.go`** — 3 new race-clean tests covering the empty-state, multi-run-picks-highest-id, and unknown-agent paths.

### Added — push/PR CI workflow (`ci.yml`)

Before v0.6.2 the only CI signal on `main` came from `release.yml`, which fires only on tag pushes. Anything that broke between tags surfaced as a failed release run rather than a failed PR — exactly the gap that let the v0.6.1 build break ship in the first place. `ci.yml` closes it.

- **`.github/workflows/ci.yml`** — runs on every push to `main` and every pull request targeting `main`. Steps: `actions/checkout@v6`, `actions/setup-go@v6` (go-version 1.25), `make verify-go-version`, `make verify-versions`, `go vet ./...`, `go test -race -count=1 -timeout=120s ./...`. Same drift checks as the release workflow so a contributor pushing a half-bumped version sees the failure in their PR, not at tag time.
- Consumes no `github.event.*` user-controlled input (explicit security note inline in the workflow file), so the GitHub Actions workflow-injection class of issues does not apply.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.6.1` → `0.6.2`.
- `internal/mcp/server.go` MCP server `Version` `0.6.1` → `0.6.2`.
- `cmd/buddy/main.go` `var version` `0.6.1` → `0.6.2`.
- `Makefile` `RELEASE_VERSION` `0.6.1` → `0.6.2`.
- `README.md` install snippet + sample output bumped to `0.6.2`.

### Versioning policy note

The new `buddy agent log` subcommand is a *user-visible surface addition*; ADR-004 §2.3 would normally pull this into a minor (`0.7.0`). Bundled as a patch in v0.6.2 at user direction — the surrounding entries (drift checks, CI gate) are all internal hardening + the subcommand wraps existing storage with no schema or behavior change. Future user-visible additions of similar scope will reset to the §2.3 default.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches and upgrades to 0.6.2. Skill catalog + command surface unchanged from 0.6.x (148 skills / 99 commands).
- **cli binary users**: optional bump. v0.6.1 binaries continue to work; pull v0.6.2 only if you want `buddy agent log <id>` and the matching version string.
- **Plugin / cli contributors / fork maintainers**: `ci.yml` now runs on every PR — local `make verify-versions verify-go-version test` before pushing avoids round-trips through the PR queue.
- **Existing buddy DBs**: no schema migration. `Store.LatestRun` reads the existing `agent_runs` table.

## [0.6.1] — 2026-05-12

### Fixed — GitHub Actions Node 20 deprecation

The v0.6.0 release workflow surfaced the Node.js 20 deprecation warning on its run page: `actions/checkout@v4`, `actions/setup-go@v5`, and `softprops/action-gh-release@v2` are all Node 20 actions. GitHub forces Node 24 as the default on 2026-06-02 and removes the Node 20 runner on 2026-09-16; without action bumps the next release after that date could surface broken runs.

Fix: bump all three actions to their Node 24-compatible majors. Functionality unchanged, but the workflow is now warning-free on current runners and forward-compatible past the 2026-06-02 default switch.

### Changed (workflow-only)

- `.github/workflows/release.yml`:
  - `actions/checkout@v4` → `@v6` (latest v6.0.2 — Node 24 runtime).
  - `actions/setup-go@v5` → `@v6` (latest v6.4.0 — Node 24 runtime).
  - `softprops/action-gh-release@v2` → `@v3` (latest v3.0.0 — release notes explicitly state "moves the action runtime from Node 20 to Node 24").
  - `go-version: '1.22'` → `'1.25'` to match `go.mod`'s `go 1.25.0` directive. setup-go@v6 introduces a breaking change — it exports `GOTOOLCHAIN=local` by default, so the runner no longer auto-downloads a newer toolchain when the requested `go-version` is older than `go.mod` requires. The first v0.6.1 build attempt failed on this; the workflow now pins the matching Go line explicitly.

Floating-major tags (`@v6`, `@v6`, `@v3`) follow GitHub Actions convention; minor patch releases auto-apply without further workflow edits.

### Changed (doc sync)

- `README.md` requirement line: `Go 1.22+` → `Go 1.25+`.
- `docs/HANDOFF.md` Stack line: `Go 1.22+` → `Go 1.25+`.

  `go.mod` was bumped to 1.25.0 earlier in the cycle without these doc lines being touched; the v0.6.1 build failure surfaced the drift.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.6.0` → `0.6.1`.
- `internal/mcp/server.go` MCP server `Version` `0.6.0` → `0.6.1`.
- `cmd/buddy/main.go` `var version` `0.6.0` → `0.6.1`.
- `Makefile` `RELEASE_VERSION` `0.6.0` → `0.6.1`.
- `README.md` install snippet + sample output bumped to `0.6.1`.

### Versioning policy note

Workflow-only patch, no behavior change visible to plugin / cli consumers. Per ADR-004 §2.3 (config / infra patch with no scope change → patch), this is `0.6.0 → 0.6.1`, not a minor. The tag-triggered publish doubles as a live verification of the bumped workflow.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches and upgrades to 0.6.1. No surface change from 0.6.0.
- **cli binary users**: optional bump — v0.6.0 binaries continue to work. Pull v0.6.1 only if you want the version-string to match the latest release page.
- **Plugin / cli contributors**: anyone forking the workflow should pull the same three action bumps; the v0.6.0 file's pins are deprecated.

## [0.6.0] — 2026-05-12

### Added — cli buddy W3-4 follow-on: conditional next-phase branches

- `NextPhase.Branches []NextPhaseBranch` — captures conditional cascade
  rules of the form `- <condition> → \`skill-a\` [+ \`skill-b\` ...]` so
  callers can pick the right target based on the run's environment /
  inputs instead of fanning out to the union `Skills` list. Each branch
  records the trimmed condition prose (≤30 runes, e.g. "글로벌",
  "Korea", "USA / EU / 기타") and the backtick-extracted RHS skills.
  Branches with no skill on the RHS ("→ (template 작성 필요)") keep the
  condition with an empty skill slice — still informative.
- Detection rejects backtick-leading bullets (`- \`skill\` — entity → API
  resource 매핑` stays a sequential candidate, not a branch) and prose
  bullets where the LHS exceeds 30 runes, so description-style arrows
  don't get misclassified.
- ASCII `->` and Unicode `→` both match.
- Runtime emits one `next-phase branch: "<cond>" → <skills>` log line per
  branch alongside the existing `next-phase candidates` union line, so
  `agent_logs` shows the cascade structure grep-line-by-line.
- `internal/agent/parser_test.go` — 7 race-clean unit tests covering
  single conditional, multi-conditional with descriptions and no-skill
  fallback, ASCII arrow form, sequential-only stays Branches-empty,
  long-LHS prose rejection, backtick-leading rejection, integration
  with §self-check populated.
- `internal/agent/runtime_test.go` — `TestRuntime_Run_ParsesConditionalBranches`
  end-to-end: mock returns Korea/global/USA branch syntax; asserts
  `StepResult.Parsed.NextPhase.Branches` populated and three branch log
  lines written.

v0.3 contract unchanged: parsed branches are metadata. Auto-cascade
into a chosen branch's skills remains deferred — the cascade engine
needs a branch-selection policy (env-var? CLI flag? interactive
prompt?) that hasn't been designed yet.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.5.0` → `0.6.0`.
- `internal/mcp/server.go` MCP server `Version` `0.5.0` → `0.6.0`.
- `cmd/buddy/main.go` `var version` `0.5.0` → `0.6.0`.
- `Makefile` `RELEASE_VERSION` `0.5.0` → `0.6.0`.
- `README.md` install snippet + sample output bumped to `0.6.0`.

### Versioning policy note

Conditional branch extraction is a *behavior-visible* addition — new `NextPhase.Branches` field surfaces through `agent_runs.result_json`, and the runtime emits one new `agent_logs` line per branch alongside the existing union-candidates line. Per ADR-004 §2.3 (new functionality under v0.x → minor bump), this is `0.5.0 → 0.6.0`, not a patch. Consistent with the v0.5.0 release rationale.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches metadata and upgrades to 0.6.0. Skill catalog + command surface unchanged from 0.5.x (148 skills / 99 commands). The branch parser runs for every `buddy agent run` invocation but only adds metadata — existing agent specs and PROCEDUREs need no edit.
- **cli binary users**: tag `v0.6.0` triggers the GitHub Actions release workflow → cross-compile matrix (4 buddy + 4 buddy-mcp) + `SHA256SUMS` published automatically. Install per README §"Release binary" using `VERSION=0.6.0`.
- **Agent JSON consumers**: `NextPhase` now has an optional `branches` field (`[{condition, skills}]`). Code that pins to the exact shape must accept the new field; `skills` (union) stays byte-identical to v0.5.0, so callers that only read `skills` need no change.
- **Agent log parsers**: one new info-level line per detected branch (`next-phase branch: "<cond>" → <skills>` or `→ (no skill)`). Parsers that match on exact line prefixes may need to ignore the new prefix. The existing `next-phase candidates: <list>` line is preserved unchanged.

## [0.5.0] — 2026-05-12

### Added — cli buddy W3-4 partial: PROCEDURE output parser

- `internal/agent/parser.go` — `ParseClaudeOutput(stdout)` returns
  `ParsedOutput{SelfCheck, NextPhase}`. Pattern-matches the §self-check
  section (Form A `## 6. 검증 (self-check)`, Form C `## 11. Verification
  gate`, English `## N. Verification`) and counts `- [x]` vs `- [ ]`
  checkboxes to compute a verdict (`pass` / `fail` / `pending` /
  `unknown`). Pattern-matches the §next-phase section (`## N. 다음
  phase` / `## N. 다음 skill` / English `Next phase` / `Next skill`)
  and extracts kebab-case skill identifiers from backticks
  (deduplicated, with sentence-like and path-shaped strings filtered
  out).
- `internal/agent/parser_test.go` — 11 race-clean unit tests covering:
  all-passed pass verdict, mixed fail, all-unchecked pending, no
  section unknown, Form C `Verification gate` header, case-insensitive
  capital X, next-phase skill extraction + filtering + dedupe + empty
  section + integration with both sections populated + empty input
  harmless.
- `StepResult.Parsed` field — surfaces verdict + per-item detail + next
  phase candidates in `agent_runs.result_json`. The MCP-side / TUI
  consumers (and the future cascade engine) read this rather than
  re-parsing stdout.
- Runtime self-check log line — `agent_logs` now records
  `step[N] <cmd> self-check=pass (M/T passed)` alongside the existing
  ok / warn / error lines, plus `next-phase candidates: <list>` when
  the parser found cascade hints.

v0.3 contract: parsed metadata only. Step success / failure still
tracks `ExitCode` from the executor. Auto-retry on `self_check=fail`
and auto-cascade to `next_phase.skills` are W3-4 follow-on — they
need real-world Claude dogfood signal first because the LLM may
consistently echo `- [ ]` without genuinely completing the check.

### Changed

- `cli-buddy-spec.md` §9 — W3-4 row marked "MED-HIGH (partial Done
  2026-05-11 — parser ship, retry/fail 의미 변경 deferred)" with explicit
  ship summary + deferred follow-on list.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.4.1` → `0.5.0`.
- `internal/mcp/server.go` MCP server `Version` `0.4.1` → `0.5.0`.
- `cmd/buddy/main.go` `var version` `0.4.1` → `0.5.0`.
- `Makefile` `RELEASE_VERSION` `0.4.1` → `0.5.0`.
- `README.md` install snippet + sample output bumped to `0.5.0`.

### Versioning policy note

W3-4 parser ships *behavior-visible* additions — new `internal/agent/parser.go` package member, new `StepResult.Parsed` field surfacing through `agent_runs.result_json`, two new `agent_logs` lines per step. Per ADR-004 §2.3 (new functionality under v0.x → minor bump), this is `0.4.1 → 0.5.0`, not a patch.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches metadata and upgrades to 0.5.0. Skill catalog + command surface unchanged from 0.4.x (148 skills / 99 commands). The parser runs for every `buddy agent run` invocation but only adds metadata — existing agent specs need no change.
- **cli binary users**: tag `v0.5.0` triggers the GitHub Actions release workflow → cross-compile matrix (4 buddy + 4 buddy-mcp) + `SHA256SUMS` published automatically. Install per README §"Release binary" using `VERSION=0.5.0`.
- **agent run JSON schema**: `StepResult` now has a `parsed` field (`{self_check, next_phase}`). Consumers that pin to the exact `StepResult` shape must accept the new field; everything else stays byte-identical.
- **agent_logs**: two new info-level lines per successful step (`self-check=<verdict> (M/T passed)` and `next-phase candidates: <list>`). Log parsers that match on exact line formats may need to ignore the new prefixes.

## [0.4.1] — 2026-05-11

### Fixed — release workflow includes buddy-mcp binaries

v0.4.0's `.github/workflows/release.yml` had `files: dist/buddy_*` which uploads only the cli binary; `buddy-mcp_*` (MCP server binary) was built but never attached. The plugin's `mcpServers` entry expects `buddy-mcp` on PATH, so users who installed via release binary had a working `buddy` CLI but no MCP server.

Fix: widen the pattern to include `dist/buddy-mcp_*` alongside `dist/buddy_*` + `dist/SHA256SUMS`. v0.4.1 tag-triggered release publishes **8 binaries** (4 buddy + 4 buddy-mcp) plus the checksum file.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.4.0` → `0.4.1`.
- `internal/mcp/server.go` MCP server `Version` `0.4.0` → `0.4.1`.
- `cmd/buddy/main.go` `var version` `0.4.0` → `0.4.1`.
- `Makefile` `RELEASE_VERSION` `0.4.0` → `0.4.1`.
- `README.md` install snippet bumped to `0.4.1`.

### Migration notes

- This is a patch release; no functionality changes beyond the workflow fix. Existing v0.4.0 users may stay on v0.4.0 if they only need the cli binary; pull v0.4.1 if you need `buddy-mcp` from the release page rather than from source build.

## [0.4.0] — 2026-05-11

### Versioning policy note

본 release 는 *cli buddy track 의 W3-3 ship + plugin 의 0.3.0 이후 누적 doc/infra 변경* 을 한 minor bump 로 묶음. ADR-004 §2.4 (revised) 의 *공통 `vX.Y.Z` namespace* 정합 — Go CLI binary 의 self-report `version` 도 본 release 부터 plugin 과 같이 `0.4.0` 으로 align. 향후 두 트랙은 *artifact 별 의미* 가 release title 로 disambiguate, internal version 은 namespace 동기.

### Added — cli buddy W3-3 agent runtime (minimum-viable subset)

Ships the agent runtime engine + on-demand CLI (`buddy agent ...`) following the
ADR-005 lock-in. v0.3.x ships the minimum-viable subset; background scheduler,
exponential backoff, streaming logs, and PROCEDURE §6 self-check parse are
W3-3 / W3-4 follow-ons.

- **SQLite migration v4** (`internal/db/migrations.go`) — `agents` (static
  definition: id/name/spec_yaml/schedule/status/created_at/updated_at/last_run_at),
  `agent_runs` (one row per Run invocation, FK cascade), `agent_logs`
  (per-run log lines). Idempotent.
- **`internal/agent/` package**:
  - `types.go` — Agent / AgentSpec / ChainStep / RetryPolicy / OutputTarget /
    AgentRun / AgentLog / StepResult.
  - `spec.go` — YAML parsing + validation + `NormalizeCommand` (strips
    `/buddy:` / `/` prefix so users can paste slash form).
  - `store.go` — agent CRUD + StartRun/FinishRun/AppendLog/Logs persistence.
    ErrNotFound sentinel matches the analytics package convention.
  - `executor.go` — `Executor` interface + `SubprocessExecutor` (spawns
    `claude` per ADR-005 §4.1 option (a)) + `MockExecutor` (test double).
    `ErrClaudeMissing` sentinel guides users to install Claude Code.
  - `runtime.go` — `Runtime.Run(agent)` walks `spec.Chain` step-by-step
    via executor, persists each attempt + result, handles per-step retry
    (capped `MaxAttempts` + optional fixed `BackoffDelay`), writes
    file/stdout output, transitions status idle → running → done|failed.
  - `runtime_test.go` + helpers — 9 race-clean integration tests covering
    spec parsing, store CRUD, happy-path persistence, non-zero exit short-circuit,
    retry exhaustion, executor error path, file output target.
- **`cmd/buddy/agent.go`** — `buddy agent create | list | show | run | delete`
  5 subcommands sharing the same `--db` flag and using
  `agent.NewSubprocessExecutor()` for `run`. `--claude-binary` override flag
  supported. End-to-end smoke verified.
- **Spec status**: `docs/cli-buddy-spec.md` §9 W3-3 row marked "HIGH (partial
  Done 2026-05-11 — minimum-viable subset)" with explicit ship summary and
  deferred follow-on list.

### Added — cli buddy W3-3 follow-on: background scheduler

- `internal/agent/scheduler.go` — `Scheduler` that loads agents with non-empty
  `spec.schedule` (cron expression), ticks them via `robfig/cron/v3`, dispatches
  `Runtime.Run` per tick, and guards against overlap via a per-agent atomic
  in-flight flag (concurrent ticks for the same agent are dropped, not piled).
  Sequential dispatch within one Scheduler instance — no parallel agent runs in
  v0.3 (W3-3 follow-on follow-on).
- `internal/agent/scheduler_test.go` — 5 race-clean tests covering: load (only
  scheduled agents register), invalid-cron reporting (skipped list, others
  still load), end-to-end Start with `@every 1s` cron firing ≥1 tick in a 2.5s
  window, overlap guard via direct `makeJob` unit test (sub-second timing
  reliability), and immediate-cancel shutdown sanity.
- `cmd/buddy/agent.go` — `buddy agent scheduler {start,status}` subcommands.
  `start` is a blocking foreground loop honouring SIGINT via cobra's
  context-driven cancel. `status` is one-shot — manually computes next fire
  time so users see a real timestamp before they run `start`. Both share the
  same `--db` / `--claude-binary` flags as `buddy agent run`.
- New direct dep: `github.com/robfig/cron/v3 v3.0.1` (MIT). Promoted to
  go.mod's direct require block by `go mod tidy`.

### Caveats / known limitations (W3-3 scheduler v0.3)

- **Sub-second cron is not supported.** robfig/cron's `ConstantDelaySchedule.Next`
  rounds to whole seconds; `@every 100ms` does not fire reliably. Use `@every 1m`
  or coarser. Documented in the `agent scheduler` long help.
- **No live refresh.** The scheduler reads agents once at startup. Adding /
  deleting agents while it runs has no effect until restart. Live refresh is a
  W3-3 follow-on follow-on.
- **No exponential backoff.** Retry honours the `MaxAttempts` + `BackoffDelay`
  fields with fixed delay only. Exponential remains W3-3 follow-on.

### Changed

- `internal/db/db_test.go` — schema_version assertion bumped 3 → 4; table-existence
  test gains agents / agent_runs / agent_logs.
- `cmd/buddy/agent.go` — `newAgentCmd` now wires in `newAgentSchedulerCmd`
  alongside the existing 5 subcommands.
- `docs/cli-buddy-spec.md` §9 W3-3 row — background scheduler moved out of
  "deferred follow-on" into the ✅ ship summary.

### Changed (release-only)

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.3.0` → **`0.4.0`** (minor bump per ADR-004 §2.3: cli buddy track major new functionality = minor, not patch).
- `internal/mcp/server.go` MCP server `Version` `0.3.0` → `0.4.0`.
- `cmd/buddy/main.go` `var version` `0.1.0` → **`0.4.0`** (cli binary self-report joins shared namespace per ADR-004 §2.4 revised).
- `Makefile` `RELEASE_VERSION` `0.1.0` → `0.4.0` — tag-triggered release workflow now matches and will auto-publish the cross-compile binary matrix for the first time since the v0.1.0 cli release (2026-04-26).
- `README.md` install snippet + sample output bumped to `0.4.0`.

### Migration notes

- **plugin users**: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches metadata and upgrades to 0.4.0. Skill catalog + command surface unchanged from 0.3.0 (148 skills / 99 commands). Analytics tools still opt-in via `BUDDY_ANALYTICS_BACKEND`.
- **cli binary users**: tag `v0.4.0` triggers the GitHub Actions release workflow → cross-compile matrix (4 binaries) + `SHA256SUMS` published automatically. Install per README §"Release binary" using `VERSION=0.4.0`. The earlier `v0.1.0` binary release stays available for compatibility.
- **Existing buddy DBs**: schema v4 migration runs automatically on next open (adds `agents` / `agent_runs` / `agent_logs` tables). Idempotent across reopens.
- **New CLI subtree**: `buddy agent {create,list,show,run,delete,scheduler}` is opt-in. Users who don't touch the agent subtree get the same surface as 0.3.0.

## [0.3.0] — 2026-05-11

### Added — analytics-mcp Phase W4-2.1 ~ W4-2.6 (full v1 ship modulo standalone tag)

- `internal/mcp/analytics_tool.go` — 7 MCP tools (`analytics_query_funnel` / `analytics_query_cohort` / `analytics_query_ab_experiment` / `analytics_query_actor_failure` / `analytics_query_cost` / `analytics_query_slo_burn` / `analytics_query_feedback_corpus`) registered in `cmd/buddy-mcp/`. When `Options.Analytics` is set, each handler queries the adapter and returns a JSON-marshalled typed result; when nil, falls back to the friend-tone "backend not configured" stub.
- **`internal/analytics/` new package** (W4-2.2 + W4-2.3) — Adapter interface + SQLite reference implementation:
  - `types.go` — input/output structs matching spec §4 verbatim (TimeRange / Segment / FunnelQuery / CohortQuery / ABExperimentQuery / ActorFailureQuery / CostQuery / SLOBurnQuery / FeedbackQuery + result types).
  - `schema.go` — 7-table SQLite DDL (`events` / `ab_experiments` / `ab_assignments` / `ab_metric_observations` / `failures` / `cost_records` / `slo_observations` / `feedback_items`) + idempotent `Migrate(ctx, *sql.DB)`.
  - `adapter.go` — `Adapter` interface (7 methods) + `ErrNotFound` sentinel.
  - `sql.go` — `SQLAdapter` with full implementations: loose funnel + retention curves + Welch's t-test on A/B experiments + Nygard 4-dim trust score + z-score cost anomaly detection + Google SRE multi-window SLO burn thresholds + NPS band split with topic clustering. Dependency-free `normalCDF` via `math.Erf`.
- `cmd/buddy-mcp/main.go` — reads `BUDDY_ANALYTICS_BACKEND` + `BUDDY_ANALYTICS_DSN` env vars; when `BUDDY_ANALYTICS_BACKEND=sql` + DSN set, opens SQLite, migrates schema, wires `analytics.NewSQLAdapter(db)` into `Options.Analytics`. Documented at top of file.
- `BUDDY_ANALYTICS_BACKEND` env var (recognised: `sql` / `mixpanel` / `amplitude` / `datadog` / `stripe` / `elasticsearch`). v0.2 ships SQL adapter; other values still register the tools as stubs.
- `internal/mcp/analytics_tool_test.go` — registration tests, stub behaviour test, backend resolution sentinel tests, **adapter-wired integration test** (`TestAnalyticsTools_AdapterWiredReturnsJSON`) proving JSON result body when `Options.Analytics` is non-nil.
- `internal/analytics/sql_test.go` — 9 race-clean integration tests against in-memory SQLite (one per query method plus not-found edge cases for A/B + SLO). Full repo `go test -race -count=1 ./...` clean across 19 packages.
- `## MCP integration (analytics-mcp v0.2.0+)` section appended to 10 skill PROCEDUREs (primary 7 + secondary 3: optimize-conversion-funnel / audit-cost-efficiency / triage-customer-support-ticket). Each section documents the matching `analytics_query_*` tool invocation example + `BUDDY_ANALYTICS_BACKEND` env var prerequisite + cross-reference to spec §4.
- `docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` Status flipped Draft → **Accepted (v0.2.0 — phase W4-2.1 stub published)**. §8 phase table updated: W4-2.1 / W4-2.2 / W4-2.3 / W4-2.4 / W4-2.5 / W4-2.6 all Done. W4-2.7 standalone `analytics-mcp-v0.1.0` tag deferred (analytics-mcp bundled inside plugin v0.x.x for now).

Deferred (per spec §8): W4-2.7 standalone `analytics-mcp-v0.1.0` tag — analytics-mcp ships inside plugin v0.x.x, separate tag is a future packaging decision (ADR-004 §2.2 condition 1).

### Changed

- `internal/mcp/server.go` — MCP server `Version` bumped `0.2.0` → `0.3.0` and `Instructions` mention the analytics surface.
- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `0.2.0` → **`0.3.0`** (minor bump per ADR-004 §2.3: new MCP surface + Adapter layer = minor, not patch).

### Migration notes

- Existing v0.2.0 installs: `claude plugin marketplace add 0xmhha/buddy && claude plugin install buddy@buddy` re-fetches metadata and upgrades to 0.3.0. Skill catalog + command surface unchanged from v0.2.0 (148 skills / 99 commands).
- Analytics tools are *opt-in*: nothing changes for users who don't set `BUDDY_ANALYTICS_BACKEND`. The 7 `analytics_query_*` tools continue to register and return the friend-tone stub message when no adapter is configured.
- To enable the SQL adapter:
  ```bash
  export BUDDY_ANALYTICS_BACKEND=sql
  export BUDDY_ANALYTICS_DSN=~/.buddy/analytics.db
  ```
  Then connect `cmd/buddy-mcp` from Claude Code's `mcpServers` config with the env block above. Schema is migrated on first connection (idempotent).

## [0.2.0] — 2026-05-11

### Versioning policy reset (BREAKING — version scheme only)

이전 v1.0.0 ~ v1.1.1 entry 5개는 사전적으로 `1.x` 로 진행됐으나, **plugin track 의 v1.0.0 milestone 정의는 cli buddy 통합 + production-proven 이후** (`docs/two-tracks-charter.md` + `docs/cli-buddy-spec.md`). 그 전까지 모든 plugin release 의 major 는 `0` 으로 유지. 새 ADR-004 (`docs/superpowers/decisions/2026-05-11-plugin-version-reset.md`) 가 결정 근거.

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version `1.1.1` → **`0.2.0`** (이전 v1.x.x history 는 본 CHANGELOG 의 [1.0.0]~[1.1.1] entry 로 historical 보존 — 마켓플레이스 fetch 시 본 entry 들은 superseded 로 간주, git tag `v1.0.x`~`v1.1.x` 가 publish 되지 않은 상태라 외부 영향 0).

### Doc cleanup — 12 superseded notes/plans 제거

Skill Completion Cycle / N-1 closure / Phase 5 ext / v0.2 outline (cli-buddy-spec 으로 superseded) 의 *planning + audit 산출물* 12개 제거. 결정 + 검증 결과는 ADR-001 / ADR-002 / ADR-003 / tasks.md / HANDOFF.md 로 흡수.

제거된 파일 (git history 보존):
- `docs/notes/2026-05-08-live-retest-v1.0.4.md`
- `docs/notes/2026-05-09-dogfood-meta-buddy-v02.md`
- `docs/notes/2026-05-09-handoff-N1-closure.md`
- `docs/notes/2026-05-10-external-skills-inventory.md`
- `docs/notes/2026-05-10-missing-skills-inventory.md`
- `docs/notes/2026-05-10-skill-matrix-3a-groups-2-4.md`
- `docs/notes/2026-05-10-skill-matrix-3b-group-1.md`
- `docs/notes/audits/2026-05-09-skill-context-bloat-audit.md` (+ `docs/notes/audits/` dir)
- `docs/superpowers/plans/2026-05-06-stage-buildout-plan.md`
- `docs/superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md`
- `docs/superpowers/plans/2026-05-09-v02-control-plane-plan.md`
- `docs/superpowers/plans/2026-05-10-skill-completion-plan.md` (+ `docs/superpowers/plans/` dir)

### Added — Quality infrastructure (B6 도입)

- `plugin/skills/.template/PROCEDURE.md` — 12-section 표준 skeleton template (신규 skill 작성 시 reference)
- `bin/lint-skill-procedure.sh` — 148 skill PROCEDURE 양식 inconsistency lint script (section 명칭 / 필수 §11 self-check / `## 다음 phase` 명시)
- `docs/superpowers/decisions/2026-05-11-plugin-version-reset.md` — ADR-004 (version reset rationale)

### Migration notes

- `claude plugin marketplace add 0xmhha/buddy` 후 `claude plugin install buddy@buddy` → `0.2.0` install. v1.x.x 가 *어떤 사용자에게도 publish 되지 않은 상태* 라 downgrade migration 부담 0.
- Skill catalog + command surface 모두 v1.1.1 시점과 동일 (148 skills / 99 commands). content 변경 없음 — content는 0.3.0 이후 cycle 에 진입.

## [1.1.1] — 2026-05-10 *(superseded — see [0.2.0] versioning policy reset)*

### Fixed — Cycle 1 dogfood validation findings

- `concretize-idea` PROCEDURE: stage 4 / stage 6 의 stale `[bracket]` notation 제거 + "skill 미존재 시 orchestrator가 수행" fallback prose 제거. 두 skill 모두 v1.1.0 에서 작성 완료된 상태이나 본문이 갱신되지 않아 *cascade flow 가 둘 갈래로 분기 가능* 하던 위험 해소. (Issue B3 / B4 / B5)
- `validate-idea` PROCEDURE: Q1 정확 phrasing 의 어색한 동사 형태 다듬기 ("내일 사라지면 진짜로 화날" → "내일 사라지면 진짜로 화내는 사람"). (Issue B2)

### Added

- `docs/notes/2026-05-10-dogfood-result-cycle-1.md` — Cycle 1 dogfood validation 결과 (`/buddy:concretize-idea` entry 검증, 7 issue 식별, quality gate minimum-viable subset 통과)
- `docs/notes/2026-05-10-dogfood-validation-scenarios.md` §3.3 — Canned business scenarios 표 (S1 한국어 SaaS / S2 SMB 회계 / S3 i18n release) 로 Cycle 2 single-skill / cascade 테스트의 입력 set 정의. (Issue B1)

### Changed

- `plugin/.claude-plugin/plugin.json` + `.claude-plugin/marketplace.json` version 1.1.0 → 1.1.1 (patch — content fix only, no API/scope change)

## [1.1.0] — 2026-05-10

### Added — Skill Completion Cycle 100% (44 신규 + 2 통합)

**Group 1 — 직접 미구현 29 skill (commits `18a79b6` ~ `900944f`):**

- §1 customer/market 5 (Cluster E): `/buddy:analyze-market-size`, `/buddy:map-customer-segments`, `/buddy:map-jobs-to-be-done`, `/buddy:conduct-customer-interview`, `/buddy:analyze-competition-and-substitutes`
- §2 effort 1: `/buddy:estimate-feature-effort`
- §3 design 부가 4 (Cluster B residual): `/buddy:design-observability`, `/buddy:design-secret-management`, `/buddy:design-i18n-strategy`, `/buddy:design-accessibility-baseline`
- §5 build 부가 5 (Cluster D): `/buddy:generate-from-api-contract`, `/buddy:generate-tests-from-spec`, `/buddy:pair-program-loop`, `/buddy:refactor-with-rename-trace`, `/buddy:update-docs-with-code`
- §6 verify 부가 3 (Cluster A residual): `/buddy:audit-i18n-coverage`, `/buddy:chaos-test`, `/buddy:audit-test-coverage-meaningful`
- §8 data 7 (Cluster F): `/buddy:analyze-feature-adoption`, `/buddy:analyze-user-cohort`, `/buddy:analyze-actor-failure-rate`, `/buddy:analyze-cost-anomaly`, `/buddy:triage-customer-support-ticket`, `/buddy:analyze-customer-feedback-corpus`, `/buddy:audit-error-budget`
- §9 lifecycle 4 (Cluster G): `/buddy:deprecate-feature`, `/buddy:migrate-customers`, `/buddy:archive-product`, `/buddy:spin-off-feature`

**Group 2 — 신규 1 (legal Layer 2):**

- `/buddy:review-legal-regulatory` — region-agnostic 법률 / 규제 검토 frame (7 sub-domain) + region cluster trigger

**Group 4 — charter scope gap 12 skill:**

- Layer 1: `/buddy:decide-target-market` — region 결정 (글로벌 / 단일 / 다지역)
- stage 3: `/buddy:decide-form-factor-app-vs-web`
- stage 4: `/buddy:apply-design-system`, `/buddy:audit-ui-quality`, `/buddy:prototype-from-spec`, `/buddy:design-interaction-pattern`
- stage 10+11 통합: `/buddy:optimize-conversion-funnel`, `/buddy:plan-growth-experiment`, `/buddy:draft-marketing-copy`, `/buddy:plan-marketing-channel`, `/buddy:audit-seo-aso`, `/buddy:automate-marketing-content`

### Changed — Group 2 통합 PROCEDURE 갱신 (Batch 7)

- `define-product-spec` PROCEDURE 갱신 — `define-product-context` + `write-prd` (Matt Pocock `to-prd`) 흡수
- `review-engineering` PROCEDURE 갱신 — `review-code-architecture` (Ousterhout deep module + interface depth + locality + leverage) 흡수
- `plugin/.claude-plugin/plugin.json` version 1.0.8 → 1.1.0
- `.claude-plugin/marketplace.json` version + description 갱신 (148 procedures / 99 commands)

### Architecture

- charter §3 plugin buddy scope 12 stage **100% cover** 도달
- 외부 자산 활용도: (b) 수정 차용 17 + (d) 신규 24 + (c) 참고만 3 (Korea cluster deferred)
- Counts: skills 106 → **148** (+42), commands 57 → **99** (+42)

### Decisions (ADR)

- **ADR-002** (`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`) — roadmap v0.2/v0.3/v1.0 outline × charter cli buddy 진짜 목적 gap 매핑
- **ADR-003** (`docs/superpowers/decisions/2026-05-10-superpowers-attribution.md`) — external `superpowers` project attribution 정책

### Documents

- `docs/two-tracks-charter.md` — plugin buddy / cli buddy 정체성 + 책임 경계 lock-in
- `docs/response-format-guide.md` — 논문 흐름 응답 양식 reference
- (v0.2.0 cleanup 에서 제거) Step 1~4 산출 5 docs (`docs/notes/2026-05-10-missing-skills-inventory.md` + `docs/notes/2026-05-10-external-skills-inventory.md` + skill-matrix 2건 + `docs/superpowers/plans/2026-05-10-skill-completion-plan.md`) — Skill Completion Cycle 100% 후 deferred-cluster 의사결정은 `docs/tasks.md` §A-2 / HANDOFF.md 로 이관, git history 보존
- `docs/notes/2026-05-10-dogfood-validation-scenarios.md` — quality gate 검증 시나리오

### Deferred (trigger 발화 시 활성)

- Korea cluster 3 (`consult-korea-legal-context`, `draft-korea-patent-application`, `audit-korea-cii-vulnerability`) — D-F F1 (target market = Korea trigger)
- `analytics-mcp` — D-C C2 (§8 cluster F 일부 구현 후 trigger 가능 — 현재 도달)
- `feature-management-mcp` — D-C C2 (cli buddy 트랙 분리)

### Migration notes

- `/plugin` 결과 1.0.8 그대로 표시되면 `claude plugin marketplace add 0xmhha/buddy` (re-fetch) → `claude plugin install buddy@buddy` 로 1.1.0 install. 또는 `/reload-plugins`.
- 기존 57 commands 모두 그대로 동작. 42 신규 commands 추가만 — breaking change 없음.

## [1.0.8] — 2026-05-08

### Added

- **§3 SaaS pattern 3 design skills** — Phase 5 extension Cluster B, common SaaS design layer:
  - `/buddy:design-event-schema` — async event schema-first design (producer/consumer contract + versioning + DLQ + idempotency). design-api-contract 의 sync-only gap 보강.
  - `/buddy:design-auth-model` — OAuth2 / JWT / SAML / SSO / RBAC 다층 결정 — 5 axis (authn / session / authz / federation / MFA) integrated design.
  - `/buddy:design-tenant-model` — multi-tenant 격리 전략 (shared RLS vs schema-per vs DB-per) + 3 layer defense in depth + onboarding/offboarding cost + compliance scope.
- **권장 chain 패턴** — `/buddy:chain design-event-schema,design-auth-model,design-tenant-model,write-adr -- "<project>"` 로 SaaS pattern 일괄.

### Changed

- **`design-system` orchestrator** stage 5a (event-schema) + stage 6 (auth-model) + stage 6a (tenant-model) 추가 — bracket pending → [Done].
- **`scripts/test-router-wireup.sh`** PROCEDURE.md count invariant 102 → 105.
- **`marketplace.json` description** "102 procedures / 54 commands" → "105 procedures / 57 commands".

### Migration notes

- 기존 54 commands 변경 없음. 3 신규 commands 추가 — 총 57 commands.
- Phase 5 extension Cluster A (v1.0.7) + B (v1.0.8) + C (v1.0.6) 8 skill 모두 완성 — Phase 7 deferred re-evaluation 의 immediate-value 후보 8 모두 commercial-grade implementation.

## [1.0.7] — 2026-05-08

### Added

- **§6 Launch readiness 3 audit skills** — Phase 5 extension Cluster A, prepare-launch-checklist evidence source:
  - `/buddy:run-load-test` — sustained + soak + spike + stress 4 시나리오 + breaking point + capacity headroom + cost projection. SLA commitment 근거.
  - `/buddy:audit-accessibility` — WCAG 2.1 AA + axe + Lighthouse + manual screen reader (NVDA/VoiceOver) — ADA / EAA / KR 장애인차별금지법 compliance + 4 principle audit.
  - `/buddy:audit-cost-efficiency` — Infracost + per-component breakdown + unit economics ($/MAU) + waste detection (5 category) + savings recommendation (RI / Savings Plan / right-sizing).
- **권장 chain 패턴** — `/buddy:chain run-load-test,audit-accessibility,audit-cost-efficiency,prepare-launch-checklist -- "<project> v<version>"` 로 launch readiness 일괄.

### Changed

- **`verify-quality` orchestrator** stage 3a/3b/3c (load / a11y / cost) 추가, launch readiness layer 명시.
- **`scripts/test-router-wireup.sh`** PROCEDURE.md count invariant 99 → 102.
- **`marketplace.json` description** "99 procedures / 51 commands" → "102 procedures / 54 commands".

### Migration notes

- 기존 51 commands 변경 없음. 3 신규 commands 추가 — 총 54 commands.
- prepare-launch-checklist 의 yellow row (Performance / a11y / Cost) 가 본 release 의 audit skill 산출로 evidence-based green 가능.

## [1.0.6] — 2026-05-08

### Added

- **§3 Cascade bridge 2 stage skills** — Phase 5 extension Cluster C, Q8=(a) cascade §2→§3 transition layer:
  - `/buddy:map-use-cases-to-infra` — actor × use case × infra bidirectional matrix + cross-actor shared ownership + compliance scope (encryption / RLS / audit retention / GDPR). silent gap 채움 — 이 layer 없으면 §3 design 이 actor model 과 disconnect.
  - `/buddy:derive-system-topology` — actor 그래프 + infra 매핑 → 시스템 토폴로지 자동 도출 (mermaid + JSON). 7 edge type (sync/async/db-W/db-R/cache/admin/observability) + 4 trust boundary layer + violation check.
- **권장 chain 패턴** — `/buddy:chain define-tech-stack,map-use-cases-to-infra,derive-system-topology,design-data-model,design-api-contract,write-adr -- "<project>"` 로 §3 cascade 일괄.

### Changed

- **`design-system` orchestrator** stage 1 (use case → infra) + stage 2 (topology) 의 inline 설명을 본 skill 호출로 redirect, [Done] marker 추가.
- **`scripts/test-router-wireup.sh`** PROCEDURE.md count invariant 97 → 99.
- **`marketplace.json` description** "97 procedures / 49 commands" → "99 procedures / 51 commands".

### Migration notes

- 기존 49 commands 변경 없음. 2 신규 commands 추가 — 총 51 commands.
- Q8=(a) cascade §2→§3 transition 의 silent gap 채워짐 — 후속 design-data-model / design-api-contract / decompose-feature-to-actor-tracks 가 명시 mapping layer 위에서 작동.

## [1.0.5] — 2026-05-08

### Added

- **§6 Use-case test 2 stage skills** — Phase 4 of stage-buildout-plan, completes Q8=(a) cascade:
  - `/buddy:test-per-actor-use-case` — actor 단위 통합 테스트 (frontend Playwright E2E + Vitest component, backend Vitest+testcontainers integration, 3rd-party Pact contract). per-actor coverage gap 0 maintain. layer × actor × use case 매트릭스 + gap report → §4 회귀 trigger.
  - `/buddy:test-cross-actor-flow` — multi-actor flow E2E (signup → email → verify → login → me chain). real component chain (Playwright + LocalStack + SES simulator + miniredis + testcontainers + Pact broker 동시 active). cross-actor edge coverage 100% + contract drift detection (Pact + Schemathesis + oasdiff).
- **권장 chain 패턴** — `/buddy:chain test-per-actor-use-case,test-cross-actor-flow,measure-code-health -- "<feature>"` 로 §6 use-case test layer 일괄.

### Changed

- **`verify-quality` orchestrator** stage 2 + stage 3 의 inline 설명을 본 skill 호출로 redirect, [Done] marker 추가.
- **`scripts/test-router-wireup.sh`** PROCEDURE.md count invariant 95 → 97.
- **`marketplace.json` description** "95 procedures / 47 commands" → "97 procedures / 49 commands".

### Migration notes

- 기존 47 commands 변경 없음. 2 신규 commands 추가 — 총 49 commands.
- Q8=(a) cascade 완성: §2 use case → §3 system → §4 actor track → §5 build → **§6 actor-별 + cross-actor test** 의 5-단계 chain 이 본 release 로 닫힘.
- §6 verify-quality orchestrator 는 backward-compat — 기존 호출 패턴 동작, stage 2/3 가 inline 설명에서 actual skill invoke 로 upgrade.

## [1.0.4] — 2026-05-08

### Added

- **§7 Release & Beta 7 safety net stage skills** — Phase 3 of stage-buildout-plan:
  - `/buddy:run-uat` — UAT scenario 실행 + go/no-go 판단 (designated stakeholder + evidence + sign-off)
  - `/buddy:run-beta-program` — 클로즈드 5-20 cohort + structured 피드백 + GA gating + post-beta cleanup
  - `/buddy:setup-canary-deploy` — canary 단계 (≥3) + dwell time + metric gate + auto-promote/rollback + platform 별 implementation
  - `/buddy:setup-feature-flags` — flag system 결정 + 4 taxonomy + kill switch + targeting + 90d cleanup SLA + governance
  - `/buddy:setup-rollback-runbook` — decision tree + platform 별 step-by-step + schema migration safety + verification + post-mortem trigger
  - `/buddy:prepare-launch-checklist` — 6 axis × 17+ row cross-functional readiness gate + go/conditional/no-go 권고
  - `/buddy:setup-incident-paging` — on-call rotation + severity 4 분류 + escalation policy + alert routing matrix + runbook index + drill cadence
- **권장 chain 패턴** — `/buddy:chain setup-feature-flags,setup-canary-deploy,setup-rollback-runbook,setup-incident-paging,prepare-launch-checklist -- "<project>"` 로 §7-2 (Pre-Launch Safety Nets) 일괄 합성

### Changed

- **`ship-release` orchestrator** stage 흐름 9 → 14 stage 로 확장 (Phase 3 신규 7 + 기존 7), 7-2 단계 (Pre-Launch Safety Nets) 신설, 7-3 (Beta/UAT) 의 bracketed pending 해소
- **`scripts/test-router-wireup.sh`** PROCEDURE.md count invariant 88 → 95
- **`marketplace.json` description** "78 procedures / 30 commands" → "95 procedures / 47 commands"

### Migration notes

- 기존 40 commands 변경 없음. 7 신규 commands 추가 — 총 47 commands.
- §7 ship-release orchestrator 의 호출 패턴은 backward-compat. 기존 9 stage chain 도 동작하며, 신규 5 safety net stage 는 production launch 시 옵션으로 추가 호출.
- 각 신규 skill 은 read-only on production (plan / runbook / checklist / decision 산출). 실제 deploy / paging / flag toggle 자동화는 §5 build-feature 의 별도 task 로 처리.

## [1.0.3] — 2026-05-07

### Added

- **§4 Implementation Plan 6 stage skills** — Phase 2 of stage-buildout-plan:
  - `/buddy:decompose-feature-to-actor-tracks` — feature → actor 별 implementation track 분해 + cross-track contracts + Independence Matrix
  - `/buddy:decompose-track-to-tasks` — actor track → atomic task list (1 PR scope, verifiable acceptance, diff size 추정)
  - `/buddy:map-task-dependencies` — task DAG (internal + cross-actor edges) + cycle 감지 + critical path + parallel-safe levels
  - `/buddy:plan-parallel-execution` — worker batch + sync points + bottleneck mitigation (AI agent + human worker mix)
  - `/buddy:define-acceptance-test-plan` — per-actor (unit/integration/contract) + cross-actor (E2E) test plan + test infra + acceptance gate
  - `/buddy:estimate-build-timeline` — critical path 기반 calendar timeline + CI (best/expected/p90/worst) + risk buffer
- **권장 chain 패턴** — `/buddy:chain decompose-feature-to-actor-tracks,decompose-track-to-tasks,map-task-dependencies,plan-parallel-execution,define-acceptance-test-plan,estimate-build-timeline -- "<feature>"` 로 §4 일괄 합성

### Changed

- **`plan-build` orchestrator** stage 흐름에 6 신규 skill 매핑 + chain 패턴 + autoplan-extended chain 추가
- **`scripts/test-router-wireup.sh`** PROCEDURE.md count invariant 82 → 88

### Migration notes

- 기존 34 commands 변경 없음. 6 신규 commands 추가 — 총 40 commands.
- Q8=(a) cascade 의 §4 채움 완료 — §2 use case → §3 system boundary → §4 actor track → §5 actor 별 implementation 의 4-단계 chain 의 §4 가 본 release 로 활성화.

## [1.0.2] — 2026-05-07

### Added

- **§3 Technical Design 핵심 4 stage skill** — Phase 1 of stage-buildout-plan:
  - `/buddy:define-tech-stack` — 언어 / 프레임워크 / DB / runtime / hosting 8+ 차원을 alternatives 비교 + 5년 lock-in 정량 평가로 evidence-based 결정
  - `/buddy:design-data-model` — entity 매핑 + read/write 패턴 분류 + normalization 결정 + index 전략 + zero-downtime migration plan
  - `/buddy:design-api-contract` — REST/GraphQL/RPC/Webhook style 결정 + actor → operation 매핑 + schema-first + error taxonomy + versioning 정책 + contract test 전략
  - `/buddy:write-adr` — 표준 7 섹션 ADR (Title/Status/Context/Decision/Consequences positive+negative+neutral/Alternatives/References) + supersede 체인 + Index 갱신
- **권장 chain 패턴** — `/buddy:chain define-tech-stack,design-data-model,design-api-contract,write-adr -- "<feature>"` 로 §3 일괄 처리

### Changed

- **PROCEDURE.md 공통 template 강화** — reference repo 학습 (skill/superpowers, harness/everything-claude-code, claude-opus-4.7 system prompt) 적용:
  - §0 STOP gate (anti-slop, superpowers AGENTS.md 패턴)
  - §4 engineering posture (입장 / specificity / challenge — review-engineering 패턴)
  - §6 explicit output schema (Opus 4.7 prose default 보정)
  - §11 verification gate (verification-before-completion 패턴)
- **`design-system` orchestrator** stage 흐름에 4 신규 skill 매핑 + 권장 chain 패턴 추가
- **`scripts/test-router-wireup.sh`** PROCEDURE.md count invariant 78 → 82

### Migration notes

- 기존 30개 slash commands 변경 없음. 4 신규 commands (`define-tech-stack`, `design-data-model`, `design-api-contract`, `write-adr`) 추가 — 총 34 commands.
- §3 미구현 stage 5개 (`map-use-cases-to-infra`, `derive-system-topology`, `design-auth-model`, `design-observability`, `design-deploy-strategy`) 는 별도 plan 으로 후속 Phase.

## [1.0.1] — 2026-05-07

### Architecture: Single-Router Skill Dispatch

플러그인의 78개 skill body 파일이 단일 auto-loaded `router` skill을 통해 lazy-load되도록 재구성됩니다. 세션마다 항상 로드되던 skill metadata가 ~28KB → ~200 chars로 축소되어, turn당 약 7K 토큰을 사용자 작업에 회수합니다.

### Changed

- **Skill loading goes through a single router** — `plugin/skills/router/SKILL.md` 가 유일한 auto-loaded entry point. 이전 78개 skill 의 body 파일은 `SKILL.md` → `PROCEDURE.md` 로 rename 되었고 YAML frontmatter 도 제거되어 더 이상 자동 발견되지 않습니다 (rename 만으로는 발견이 멈추지 않았기 때문).
- **Skill catalog 위치 이동** — `plugin/SKILLS.md` → `plugin/skills/router/references/skill-catalog.md`, `plugin/SKILL_ROUTER.md` → `plugin/skills/router/references/routing-rules.md`.
- **Slash command 경로 평탄화** — `plugin/commands/buddy/<name>.md` → `plugin/commands/<name>.md`. 이전 nested 구조는 슬래시를 `/buddy:buddy:<name>` 형태로 노출시켜 `/buddy:<name>` 호출이 안 됐습니다.
- **Command md description 정렬** — 26개 command md frontmatter description 을 plugin.json 의 plain-language register 와 byte-identical 정렬.
- **Parallel mode dispatch 패턴 변경** — fresh subagent 의 권한 boundary 제약 때문에 router (parent) 가 모든 PROCEDURE.md 를 읽고 본문을 subagent prompt 에 embed 하도록 수정.

### Added

- **3개 신규 dispatch commands** — 기존 27개 lifecycle commands 는 변경 없이 유지되고, 다음이 추가됨:
  - `/buddy:run <skill> [args]` — 카탈로그의 임의 skill 을 직접 호출 (전용 command 가 없는 skill 의 escape hatch)
  - `/buddy:chain skill1,skill2,... -- args` — 순차 실행, 직전 단계의 출력이 다음 단계로 흐름
  - `/buddy:parallel skill1,skill2,... -- args` — Agent dispatch 기반 병렬 실행, 결과 집계
- **`/buddy:status` command md 추가** — 이전엔 plugin.json 에 등록되어 있었으나 md 파일이 누락되어 있던 gap 보완.
- **Router CI smoke test** — `scripts/test-router-wireup.sh` + Makefile target `test-routing`. 10개 invariant 검증.
- **README slash command 카탈로그** — 30개 명령 모두 phase orchestrator / stage skill / cross-phase tool / dispatch composition 4개 그룹으로 분류해 표로 정리.

### Fixed

- **plugin.json `commands` 필드 제거** — Claude Code 의 plugin schema 가 거부하는 필드. 슬래시는 `plugin/commands/*.md` 자동 발견이라 manifest 에 선언 불필요. 이전 commands 배열 때문에 plugin install 이 실패하던 문제 해결.
- **`${CLAUDE_PLUGIN_ROOT}` path resolution fallback** — runtime substitution 이 안 되는 환경 대비 Bash 기반 install root 발견 절차 추가.

### Migration notes

- 사용자 조치 불필요. 기존 27개 slash commands (`/buddy:status`, `/buddy:concretize-idea` 등) 는 변경 없이 동작.
- 플러그인 기여자: 새 skill 은 `plugin/skills/<name>/PROCEDURE.md` 로 작성 (frontmatter 없이), catalog 와 routing rules 는 `plugin/skills/router/references/` 하위에서 갱신.

## [1.0.0] - 2026-05-04

### Architecture: 9-Phase Multi-Orchestrator Model

Buddy plugin이 단일 orchestrator(`autoplan`) 가정에서 **9-phase multi-orchestrator 모델**로 전환됩니다.
각 라이프사이클 단계가 독립적인 phase orchestrator를 가지며, `autoplan`은 cross-phase review sub-orchestrator로 재배치됩니다.

### Added

**Phase Orchestrator Skills (9개 신규)**
- `concretize-idea` — §1 Idea & Business Validation (idea → PRD + business viability)
- `define-features` — §2 Feature Definition & Backlog (PRD → actor/use case/system boundary 기반 feature backlog)
- `design-system` — §3 Technical Design (feature backlog → tech stack ADR + infra + API + data model)
- `plan-build` — §4 Implementation Plan (technical design → actor별 task DAG + parallel execution plan)
- `build-feature` — §5 Development (implementation plan → working code + tests)
- `verify-quality` — §6 Quality (code complete → QA + security + compliance sign-off)
- `ship-release` — §7 Release & Beta (quality gate pass → tagged release + UAT + GA)
- `iterate-product` — §8 Operate & Iterate (production traffic → A/B 실험 + funnel + improvement backlog)
- `manage-lifecycle` — §9 Lifecycle Management (feature/product 노후화 → deprecation + migration + EOL)

**§8 Stage Skills (6개 신규 — Q3 우선순위)**
- `design-ab-experiment` — 통계적으로 유효한 A/B 실험 설계 (가설/표본/대조군/지표/기간)
- `analyze-ab-experiment` — 실험 결과 분석 (통계 유의성 + 실용 유의성 → Ship/Revert/Continue)
- `analyze-user-funnel` — §2 use case 기반 actor별 funnel 전환/이탈 분석
- `generate-improvement-tasks` — 분석 결과 → RICE 기반 improvement backlog (§2 재진입 준비)
- `handle-incident` — 프로덕션 인시던트 대응 런북 (심각도 → 완화 → 근본 원인 → fix → 커뮤니케이션)
- `conduct-postmortem` — 비난 없는 포스트모템 (타임라인 + 5 Whys + action items)

**§2 Stage Skills (7개 신규 — Q8=(a) Use Case 분해)**
- `identify-actors` — 시스템 참여 actor 열거 (user/system/3rd-party/external-tool 분류)
- `map-actor-use-cases` — actor별 use case 식별 (UML use case 다이어그램 등가)
- `map-use-case-to-system-boundary` — use case → 시스템 경계 매핑 (frontend/backend/external SaaS)
- `compose-feature-from-use-cases` — cross-actor use case → feature 합성
- `define-feature-spec` — feature 완전 명세서 (actor/use case/system boundary/acceptance/test plan 포함)
- `score-feature-priority` — RICE/ICE/MoSCoW 우선순위 결정
- `map-feature-dependencies` — feature 간 선후 의존성 DAG + critical path + 병렬 그룹

**Phase Orchestrator Commands (9개 신규 — Q2=(b))**
- `/buddy:concretize-idea`, `/buddy:define-features`, `/buddy:design-system`, `/buddy:plan-build`
- `/buddy:build-feature`, `/buddy:verify-quality`, `/buddy:ship-release`
- `/buddy:iterate-product`, `/buddy:manage-lifecycle`
- 기존 17개 commands 유지 — 총 26개 commands

### Changed

**SKILL_ROUTER.md** — 9-phase multi-orchestrator 모델로 완전 재작성
- Priority 1: 9개 phase orchestrator (기존 `autoplan` 단일 orchestrator → 교체)
- Priority 2: `autoplan` (cross-phase review sub-orchestrator)
- 11-stage 라우팅 표 → 9-phase 라우팅 표로 교체
- 케이스 A~G 업데이트 (신규 orchestrator 기반)

**SKILLS.md** — 9-phase 라이프사이클 구조로 재구성
- Phase별 섹션으로 재분류 (기존 알파벳 순 → phase 소속 기준)
- 신규 22개 skill 등재
- archive 섹션 추가

**plugin.json** — version 0.1.0-dev → 1.0.0

### Moved

- `plugin/skills/route-intent/` → `plugin/_archive/route-intent/` (Q5=(b))
- `plugin/skills/route-multi-platform/` → `plugin/_archive/route-multi-platform/`
- `plugin/skills/route-spec-to-code/` → `plugin/_archive/route-spec-to-code/`

### Architecture Decisions

- **Q1=(a)**: 9-phase 모델 전체 채택 (§9 lifecycle 포함)
- **Q2=(b)**: 26 commands (9 phase orchestrator + 17 기존 stage commands 유지)
- **Q3=(c)→(b)**: §8 먼저 → §1~§5 순서로 신규 stage skill 작성
- **Q4=(c)**: MCP 작성 보류 — skill 정리 우선
- **Q5=(b)**: archive 3개 `plugin/_archive/`로 격리
- **Q6=(a)**: `autoplan` = cross-phase review sub-orchestrator
- **Q7=(b)**: §7.5 Beta/UAT를 §7 내부 sub-phase로 분리
- **Q8=(a)**: use case 분해를 §2 첫 단계로 강제 + actor/use case/system boundary를 feature spec 필수 필드로

## [0.1.0] - 2026-04-26

First public release of Buddy — a friend-tone CLI that observes Claude Code
hooks, records normalized events to a local SQLite store, and surfaces health,
performance, and recent activity through read-only commands.

### Added

- **Hook wrapping**: `buddy hook-wrap <hook-name> [-- <command...>]` wraps a
  Claude Code hook. Seven invariants are preserved end-to-end: silent stdout
  on success, streaming stdout/stderr passthrough, exit-code passthrough,
  signal-safe child handling, deadline enforcement, structured error reporting,
  and atomic outbox writes. Each invocation records latency, exit code, and
  event metadata to a SQLite outbox.
- **Background daemon**: `buddy daemon run|start|stop|status` drains the
  outbox into normalized `hook_events` and a rolling `hook_stats` aggregate.
  Implementation is a single-goroutine poll loop with a PID file guarded by
  `flock` and a graceful SIGTERM shutdown path. Optional `cli-wrapper`
  supervision is wired through `buddy install --with-cliwrap`.
- **Lifecycle commands**: `buddy install` and `buddy uninstall` wrap and
  unwrap Claude Code's `~/.claude/settings.json` hook entries. `--with-cliwrap`
  generates a `cliwrap.yaml` for daemon supervision. Re-installs are
  idempotent, and the original settings file is preserved as a write-once
  `.buddy.bak` backup.
- **Health diagnostics**: `buddy doctor` produces a one-shot, read-only
  health snapshot — outbox backlog, slow hooks (p95 over threshold),
  failure-rate spikes, and daemon liveness — in friend-tone Korean.
- **Statistics**: `buddy stats --window 5m|1h|24h` summarises hook
  performance from the rolling aggregate. `--by-tool` splits the output by
  tool, and `--hook` filters case-insensitively.
- **Event tail**: `buddy events [--limit] [--hook] [--follow]` prints raw
  events as a debug surface. `--follow` polls every second and surfaces
  start and end markers in friend-tone.
- **User config**: `~/.buddy/config.json` exposes eight knobs —
  `hookTimeoutMs`, `hookSlowMs`, `hookFailRatePct`, `outboxBacklog`,
  `notifyChannel`, `pollInterval`, `batchSize`, `personaLocale`. Fields are
  pointer-typed so absent values are distinguishable from explicit zeros.
- **Config CLI**: `buddy config show [--json]`, `buddy config get <field>`,
  `buddy config set <field> <value>`, and `buddy config unset <field>`. The
  `set`/`unset` commands are silent on success. `buddy doctor` and
  `buddy daemon` consume the config; precedence is explicit flag > config
  file > spec default.
- **Retention purge**: `buddy purge --before <date> [--apply]` deletes old
  `hook_events` and `hook_stats` rows. `<date>` accepts a relative duration
  (`30d`), a date (`2026-01-01`), or an RFC 3339 timestamp
  (`2026-01-01T00:00:00Z`). The default mode is dry-run; `--apply` performs
  the delete inside a single transaction. The `hook_outbox` table is never
  touched, preserving the synchronous-write WAL invariant.
- **Message catalog**: `internal/persona/` consolidates roughly fifty
  user-facing Korean strings behind typed `Key` constants. Lookup is
  locale-keyed with an `en → ko` fallback (the English map is intentionally
  empty in v0.1; see Deferred).
- **Versioned binary**: `buddy --version` reports
  `buddy 0.1.0 (sha=<short>, built=<rfc3339>)` with values injected at link
  time via Makefile ldflags.
- **Cross-compile matrix**: `make release-binaries` produces
  `dist/buddy_<version>_<os>_<arch>` for `linux/amd64`, `linux/arm64`,
  `darwin/amd64`, and `darwin/arm64`, alongside `dist/SHA256SUMS`. CGO is
  disabled — the SQLite driver is `modernc.org/sqlite`, which is pure Go.
- **Tag-triggered release workflow**: `.github/workflows/release.yml` fires
  on `v*` tag pushes, runs `make release-binaries`, and uploads the binaries
  and checksums to a GitHub Release.

### Changed

- **`buddy install` is now self-contained**: it pre-creates `~/.buddy/` and
  runs schema migrations, so `buddy doctor` works immediately after install.
  Previously, fresh installs surfaced raw `no such table: hook_outbox`
  errors on the first health check.
- **`buddy daemon start` reports the real PID**: the command previously
  printed `pid -1` because the PID was read after `os.Process.Release()`
  had zeroed it.
- **`buddy uninstall` stops a running daemon by default**: the PID file is
  consulted, and SIGTERM is sent if the daemon is live. Use `--keep-daemon`
  to opt out.
- **Friendly error for missing DB**: read-only commands (`doctor`, `stats`,
  `events`, `purge`) now print
  `DB가 아직 없어 (path). 먼저 'buddy install' 했는지 확인해줘.` instead of
  SQLite's cryptic `out of memory (14)` (parent directory missing) or
  `no such table` (database file missing).

### Fixed

- **DB lock contention race**: `db.Open` now passes
  `_pragma=busy_timeout(5000)` so concurrent opens (writer plus reader) wait
  for the WAL header lock instead of failing with `database is locked`.
- **Daemon SIGTERM race**: `signal.NotifyContext` is now installed before
  the PID file is published, closing a window in which a caller polling for
  the PID file could deliver SIGTERM into Go's default handler.
- **Daemon test flake**: the two races above made
  `internal/daemon/TestRun_DrainsOutboxThenStopsOnContextCancel` flaky under
  the race detector.

### Deferred (tracked for v0.2 i18n sweep)

- Routing `config.ValidationError.Reason` strings through the persona
  catalog (catalog keys are already declared and have fallback tests).
- Migrating `queries.ErrInvalidLimit` and `queries.ErrInvalidWindow` (whose
  Korean strings are currently embedded in the error sentinels) into the
  catalog.
- Populating the English locale map.
- Subcommand-flag-aware locale resolution (the persistent pre-run currently
  reads only `~/.buddy/config.json`).
- AGENTS.md, the plugin model, and an MCP server (v1.0+ scope).

[Unreleased]: https://github.com/0xmhha/buddy/compare/v0.6.6...HEAD
[0.6.6]: https://github.com/0xmhha/buddy/releases/tag/v0.6.6
[0.6.5]: https://github.com/0xmhha/buddy/releases/tag/v0.6.5
[0.6.4]: https://github.com/0xmhha/buddy/releases/tag/v0.6.4
[0.6.3]: https://github.com/0xmhha/buddy/releases/tag/v0.6.3
[0.6.2]: https://github.com/0xmhha/buddy/releases/tag/v0.6.2
[0.6.1]: https://github.com/0xmhha/buddy/releases/tag/v0.6.1
[0.6.0]: https://github.com/0xmhha/buddy/releases/tag/v0.6.0
[0.5.0]: https://github.com/0xmhha/buddy/releases/tag/v0.5.0
[0.4.1]: https://github.com/0xmhha/buddy/releases/tag/v0.4.1
[0.4.0]: https://github.com/0xmhha/buddy/releases/tag/v0.4.0
[0.3.0]: https://github.com/0xmhha/buddy/releases/tag/v0.3.0
[0.2.0]: https://github.com/0xmhha/buddy/releases/tag/v0.2.0
[0.1.0]: https://github.com/0xmhha/buddy/releases/tag/v0.1.0
