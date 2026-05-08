# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/0xmhha/buddy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/0xmhha/buddy/releases/tag/v0.1.0
