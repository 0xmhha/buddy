# Buddy — 작업 인벤토리

> 두 트랙(Plugin + Go CLI) + 코드 housekeeping의 *pending* 작업을 한 자리에서 본다.
> 우선순위는 §끝의 [우선순위](#우선순위-권장)에서.
>
> **SSoT 분담:**
> - 이 문서: cross-track 작업 인벤토리 (실행 단위)
> - [`roadmap.md`](./roadmap.md): Go CLI 트랙의 마일스톤 SSoT (M5/M6/v0.2/v0.3/v1.0)
> - [`superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md): plugin 9-phase 아키텍처 SSoT
> - [`HANDOFF.md`](./HANDOFF.md): 세션 인계 가이드
>
> 작성일: 2026-05-05 / 상태: WORKING

---

## 트랙 상태 요약 (한 줄)

| 트랙 | 상태 | 마지막 release |
|------|------|---------------|
| **Plugin** (Claude Code plugin scaffold + 9-phase orchestrator) | ACTIVE | v1.0.0 (2026-05-04) |
| **Go CLI** (hook reliability monitor) | PAUSED | v0.1.0 (2026-04-26) |
| **Housekeeping** | ad-hoc | — |

---

## A. Plugin 트랙

### A-1. 9-phase stage skill 채우기 (Q3 순서: §8 → §1~§5 → 나머지)

> 출처: [`spec §4 Stage Skill Gap`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md#4-단계별-skill-군집화--gap-분석).
> Q3=c→b: §8 iterate-product 먼저 채우고 → §1~§5 → §6/§7/§9.

| Phase | 보유 | 누락 (작성 대상) |
|-------|------|-----------------|
| §1 concretize-idea | validate-idea, validate-advanced-edge-idea, assess-business-viability, review-pricing-and-gtm, define-product-spec | `analyze-competition-and-substitutes`, `map-customer-segments`, `map-jobs-to-be-done`, `analyze-market-size`, `conduct-customer-interview` |
| §2 define-features | identify-actors, map-actor-use-cases, map-use-case-to-system-boundary, compose-feature-from-use-cases, define-feature-spec, score-feature-priority, map-feature-dependencies, split-work-into-features, query-feature-registry, triage-work-items | `estimate-feature-effort` (story point / t-shirt sizing) |
| §3 design-system | review-architecture, review-engineering, review-design, review-devex, design-* (6), consult-codex, consult-design-system, explore-design-variants | `map-use-cases-to-infra`, `derive-system-topology`, `define-tech-stack`, `design-data-model`, `design-api-contract`, `design-event-schema`, `design-auth-model`, `design-observability`, `design-secret-management`, `design-tenant-model`, `design-i18n-strategy`, `design-accessibility-baseline`, `write-adr` |
| §4 plan-build | (orchestrator만) | `decompose-feature-to-actor-tracks`, `decompose-track-to-tasks`, `map-task-dependencies`, `plan-parallel-execution`, `define-acceptance-test-plan`, `estimate-build-timeline` |
| §5 build-feature | build-with-tdd, iterate-fix-verify, freeze-edit-scope, dispatch-parallel-agents, diagnose-bug, consult-codex | `generate-from-api-contract`, `generate-tests-from-spec`, `pair-program-loop`, `refactor-with-rename-trace`, `update-docs-with-code` |
| §6 verify-quality | classify-qa-tiers, run-browser-qa, monitor-regressions, audit-security, audit-live-devex, measure-code-health, classify-review-risks, review-{ai-safety,privacy,license,terms} | `test-per-actor-use-case`, `test-cross-actor-flow`, `run-load-test`, `audit-accessibility`, `audit-i18n-coverage`, `audit-cost-efficiency`, `chaos-test`, `audit-test-coverage-meaningful` |
| §7 ship-release | setup-quality-gates, auto-create-pr, automate-release-tagging, sync-release-docs, write-changelog, guard-destructive-commands, compose-safety-mode | `setup-canary-deploy`, `setup-feature-flags`, `setup-rollback-runbook`, `run-uat`, `run-beta-program`, `prepare-launch-checklist`, `setup-incident-paging`, **§7.5 Beta/UAT 분리 (Q7=b)** |
| §8 iterate-product | design-ab-experiment, analyze-ab-experiment, analyze-user-funnel, generate-improvement-tasks, handle-incident, conduct-postmortem, monitor-regressions, summarize-retro, save-context, restore-context, persist-learning-jsonl | `analyze-feature-adoption`, `analyze-user-cohort`, `analyze-actor-failure-rate`, `analyze-cost-anomaly`, `triage-customer-support-ticket`, `analyze-customer-feedback-corpus`, `audit-error-budget` |
| §9 manage-lifecycle | (orchestrator만) | `deprecate-feature`, `migrate-customers`, `archive-product`, `spin-off-feature` |

**총 누락 약 60+개.** 작업 단위는 phase별 PR로 묶는 게 자연스러워.

### A-2. buddy MCP server (M-1 진행 중)

| 항목 | 상태 |
|------|------|
| `cmd/buddy-mcp/` 추가 | DONE (28e9fc9) |
| doctor / stats / feature tools | DONE (28e9fc9) |
| `claude mcp add/remove` 통합 | DONE (a8aee51) |
| PIDFile 전달 fix | DONE (6f40901) |
| feature.query / store / update / link_code / export_patch (기획 스펙) | TODO 다음 단계 |

> Q4(MCP 우선순위)는 (c) 보류였지만 feature_cmd / feature registry 작업이 들어가면서 in-house 진행 중. 우선순위 재확인 필요.

### A-3. 카탈로그 정합성 — DONE (2026-05-05)

> 출처: [`spec §10 Step 2`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md#10-다음-단계).

- [x] SKILL_ROUTER.md §2/§3 multi-orchestrator 모델 갱신 — 6 카테고리 우선순위 표 + 9-phase 라우팅 표 + 7 케이스 라우팅 결정
- [x] SKILLS.md 카탈로그 재구성 — 9-phase 섹션 구조, phase orchestrator + cross-phase review + stage 분리. Trigger 컬럼이 이미 17 stage 커맨드를 `command + dispatch` 로 표기.
- [x] archive 3개(`route-intent`, `route-multi-platform`, `route-spec-to-code`) → `plugin/_archive/` 격리 (Q5=b)
- [x] `docs/skill-map.md` 11-stage → 9-phase 정렬 헤더 + 매핑 표 추가
- [x] **Command surface reconcile — 옵션 (a) spec 충실 방향 적용**:
  - `plugin/.claude-plugin/plugin.json` `commands` 배열을 15 → **27개**로 확장 (status + 9 phase + 17 stage = spec Q2=(b) 26 + status utility)
  - `SKILL_ROUTER.md §5` 를 5 sub-section (Status / Phase / Cross-phase review / Cross-cutting utility / Phase stage)으로 분리 + 27개 모두 등재
  - `jq` JSON 검증 통과, command 이름 list 27개 sorted 확인

### A-4. dogfood — plugin

- [ ] 실제 작은 프로젝트에 plugin install → phase orchestrator 동작 검증 ([`spec §10 Step 5`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md#10-다음-단계))
- 회수 결과 → A-1 우선순위 재정렬 입력

---

## B. Go CLI 트랙

### B-1. Dogfood feedback 회수

- [ ] 사용자가 며칠(3~7일) 사용 후 [`docs/dogfood-feedback-template.md`](./dogfood-feedback-template.md) 채워서 제출
- 영향: v0.2 dashboard UX (TUI vs web) + v0.3 task tracker 통합 여부 결정 입력

### B-2. v0.2 i18n sweep (M5 deferred)

| 항목 | 위치 |
|------|------|
| `config.ValidationError.Reason` persona catalog wiring | `internal/persona/`, `translateConfigError` |
| `queries.ErrInvalidLimit` / `ErrInvalidWindow` 카탈로그 이전 | `internal/queries/` |
| English locale 카탈로그 채우기 | `internal/persona/en.go` (현재 빈 map) |
| Subcommand `--config` 인지 locale 해석 | root `PersistentPreRunE` |

### B-3. v0.2 release polish (M6 deferred)

- [ ] macOS notarization (현재는 `xattr -d com.apple.quarantine` 안내)
- [ ] Third-party Actions SHA pinning + Dependabot
- [ ] 별도 `ci.yml` (PR/push 시 test 강제 — 현재 `release.yml`만 있음)
- [ ] `VERSION` 파일 / build-time embed (현재 5곳 분산, tag↔Makefile sanity check로 보호)

### B-4. v0.2 Control Plane (multi-session dashboard)

> 출처: [`roadmap.md §4`](./roadmap.md#4-v02--control-plane-멀티-세션-dashboard).

| Task | 비고 |
|------|------|
| T1 활성 세션 발견 (recon 패턴 차용) | `internal/sessions/` |
| T2 token usage parser (transcript JSONL) | `tokenUsage` 스키마 재사용 |
| T3 multi-session stats 확장 | `--all-sessions` or `buddy sessions list/show` |
| T4 dashboard UI | **TUI vs web — open question** (dogfood feedback 대기) |
| T5 cost estimate | `internal/pricing/` 단가 테이블 |
| T6 i18n full split (en/ko) | M5 T5 forward-pointer 회수 |

### B-5. v0.3 Orchestration (task DAG)

> 출처: [`roadmap.md §5`](./roadmap.md#5-v03--orchestration-task-dag-executor).

| Task | 비고 |
|------|------|
| T1 task DAG schema | SQLite migration |
| T2 `buddy task add/list/run/status` CLI | `internal/cli/task.go` |
| T3 wave 그룹화 (errgroup) | `maxParallelTasks` config |
| T4 retry policy (exponential backoff) | quality gate retry |
| T5 외부 task tracker 통합? | **Open question** (Linear/Jira/GitHub Issues vs 실행 엔진 only) |

### B-6. v1.0 — 통합

> 출처: [`roadmap.md §6`](./roadmap.md#6-v10--통합-agentsmd-auto-sync-plugin-model-mcp-server).

- T1 AGENTS.md auto-sync (`internal/agentsmd/`)
- T2 plugin model (외부 binary + MCP transport 재활용)
- T3 MCP server (`cmd/buddy-mcp/` — A-2와 통합 가능성)
- T4 cross-harness 시도? (Codex/OpenCode — 비범위 검토만)

---

## C. 코드 housekeeping

| 항목 | 위치 | 비고 |
|------|------|------|
| `cmd/buddy/main.go` 분할 | `cmd/buddy/main.go` | HANDOFF §11에서 685 lines, v0.2 새 명령 추가 전 정리 권장 |
| Module path drift 정리 | `go.mod` + 전체 import | `github.com/wm-it-22-00661/buddy` → `github.com/0xmhha/buddy` |
| gofmt drift | 전체 | 한 commit 정리 (현재는 clean) |

---

## 우선순위 권장

1. **A-1 §1~§5 stage skill** — Q3 순서대로 채우기 (가장 큰 작업, 60+ skill)
2. **A-4 plugin dogfood** — 실 프로젝트에 install → 마찰 회수 → A-1/A-2 우선순위 재정렬
3. **A-2 buddy MCP feature.* tools** — feature registry 작업과 짝 맞춤
4. **B-1 Go CLI dogfood feedback** — 사용자 페이스
5. 나머지는 위 4개 결과 입력 받고 결정

---

## 참조

- [`HANDOFF.md`](./HANDOFF.md) — 세션 인계 + 워크플로우 skill 분기
- [`roadmap.md`](./roadmap.md) — Go CLI 마일스톤 SSoT
- [`superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md) — plugin 9-phase 아키텍처 SSoT
- [`superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`](./superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md) — plugin scaffold 설계
- [`v0.1-spec.md`](./v0.1-spec.md) — Go CLI v0.1 spec (LOCKED)
- [`skill-map.md`](./skill-map.md) — 11-stage → 9-phase 매핑 (참조용)
- [`DOGFOOD.md`](../DOGFOOD.md) + [`dogfood-feedback-template.md`](./dogfood-feedback-template.md) — Go CLI dogfood
