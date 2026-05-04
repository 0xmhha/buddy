# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/0xmhha/buddy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/0xmhha/buddy/releases/tag/v0.1.0
