# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/0xmhha/buddy/compare/v0.6.1...HEAD
[0.6.1]: https://github.com/0xmhha/buddy/releases/tag/v0.6.1
[0.6.0]: https://github.com/0xmhha/buddy/releases/tag/v0.6.0
[0.5.0]: https://github.com/0xmhha/buddy/releases/tag/v0.5.0
[0.4.1]: https://github.com/0xmhha/buddy/releases/tag/v0.4.1
[0.4.0]: https://github.com/0xmhha/buddy/releases/tag/v0.4.0
[0.3.0]: https://github.com/0xmhha/buddy/releases/tag/v0.3.0
[0.2.0]: https://github.com/0xmhha/buddy/releases/tag/v0.2.0
[0.1.0]: https://github.com/0xmhha/buddy/releases/tag/v0.1.0
