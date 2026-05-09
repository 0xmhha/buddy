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
> 작성일: 2026-05-05 / 최종 갱신: 2026-05-08 (v1.0.5 — Phase 1+2+3+4 Done 반영) / 상태: WORKING

---

## 트랙 상태 요약 (한 줄)

| 트랙 | 상태 | 마지막 release |
|------|------|---------------|
| **Plugin** (9-phase orchestrator + 97 procedures) | ACTIVE — Phase 1+2+3+4 Done, Q8=(a) cascade 완성 | v1.0.5 (2026-05-08) |
| **Go CLI** (hook reliability monitor) | PAUSED — dogfood feedback 대기 | v0.1.0 (2026-04-26) |
| **Housekeeping** | ad-hoc | — |

---

## A. Plugin 트랙

### A-1. 9-phase stage skill 채우기 (v1.0.5 기준)

> 출처: [`spec §4 Stage Skill Gap`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md#4-단계별-skill-군집화--gap-분석).
> Phase 1+2+3+4 (v1.0.2~v1.0.5) 완료 — Q8=(a) cascade 5-단계 chain (use case → system → actor track → build → test) 완성.
> 잔여는 Phase 7 deferred re-evaluation 의 Cluster A/B/C/D/E/F/G/H 분류 참조.

| Phase | 보유 | 누락 잔여 | Phase 7 Cluster |
|-------|------|-----------|-----------------|
| §1 concretize-idea | validate-idea, validate-advanced-edge-idea, assess-business-viability, review-pricing-and-gtm, define-product-spec | `analyze-competition-and-substitutes`, `map-customer-segments`, `map-jobs-to-be-done`, `analyze-market-size`, `conduct-customer-interview` (5) | Cluster E — conditional |
| §2 define-features | identify-actors + 9 stage skill | `estimate-feature-effort` (1 small) | low priority |
| §3 design-system | review-* (4), design-* (6 기존), consult-*, explore-design-variants, **define-tech-stack ✓ v1.0.2**, **design-data-model ✓ v1.0.2**, **design-api-contract ✓ v1.0.2**, **write-adr ✓ v1.0.2** | `map-use-cases-to-infra`, `derive-system-topology` (Cluster C — **HIGH cascade value**), `design-event-schema`, `design-auth-model`, `design-tenant-model` (Cluster B — **MEDIUM**), `design-observability`, `design-secret-management`, `design-i18n-strategy`, `design-accessibility-baseline` (Cluster B residual — conditional) (9) | Cluster B + C |
| §4 plan-build | **6 stage 모두 ✓ v1.0.3** (decompose-feature-to-actor-tracks / decompose-track-to-tasks / map-task-dependencies / plan-parallel-execution / define-acceptance-test-plan / estimate-build-timeline) | **0 잔여** | Done |
| §5 build-feature | build-with-tdd, iterate-fix-verify, freeze-edit-scope, dispatch-parallel-agents, diagnose-bug, consult-codex | `generate-from-api-contract`, `generate-tests-from-spec`, `pair-program-loop`, `refactor-with-rename-trace`, `update-docs-with-code` (5) | Cluster D — low priority (IDE/codegen 도구로 대체 가능) |
| §6 verify-quality | 11 기존 + **test-per-actor-use-case ✓ v1.0.5**, **test-cross-actor-flow ✓ v1.0.5** | `run-load-test`, `audit-accessibility`, `audit-cost-efficiency` (Cluster A — **HIGH immediate**), `audit-i18n-coverage`, `chaos-test`, `audit-test-coverage-meaningful` (3 conditional) (6) | Cluster A — high |
| §7 ship-release | 7 기존 + **setup-canary-deploy ✓ v1.0.4**, **setup-feature-flags ✓ v1.0.4**, **setup-rollback-runbook ✓ v1.0.4**, **run-uat ✓ v1.0.4**, **run-beta-program ✓ v1.0.4**, **prepare-launch-checklist ✓ v1.0.4**, **setup-incident-paging ✓ v1.0.4** | **0 잔여**. Q7=(b) §7.5 분리 = §7 안의 7-2/7-3 stage 로 흡수 결정 (별도 phase 미채택) | Done |
| §8 iterate-product | design-ab-experiment, analyze-ab-experiment, analyze-user-funnel, generate-improvement-tasks, handle-incident, conduct-postmortem, monitor-regressions, summarize-retro, save-context, restore-context, persist-learning-jsonl | `analyze-feature-adoption`, `analyze-user-cohort`, `analyze-actor-failure-rate`, `analyze-cost-anomaly`, `triage-customer-support-ticket`, `analyze-customer-feedback-corpus`, `audit-error-budget` (7) | Cluster F — production traffic 의존 |
| §9 manage-lifecycle | (orchestrator만) | `deprecate-feature`, `migrate-customers`, `archive-product`, `spin-off-feature` (4) | Cluster G — 1년+ deferred |

**총 잔여: 37 skill** (Phase 1+2+3+4 의 15 Done 차감). 자세한 cluster 분류 + immediate value 분석은 [`Phase 7 deferred re-evaluation`](./superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md).

**Phase 5 extension 후보 (immediate value, 8 skill):**
- Cluster A (§6 audit): `run-load-test`, `audit-accessibility`, `audit-cost-efficiency` — launch checklist 보강
- Cluster B (§3 design): `design-event-schema`, `design-auth-model`, `design-tenant-model` — SaaS 공통 패턴
- Cluster C (§3 cascade bridge): `map-use-cases-to-infra`, `derive-system-topology` — Q8=(a) cascade §2→§3 정합 layer

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

## 우선순위 권장 (v1.0.5 기준 갱신)

1. **Phase 5 extension Cluster C** — `map-use-cases-to-infra`, `derive-system-topology` (Q8=(a) cascade §2→§3 정합 — silent gap 채움)
2. **Phase 5 extension Cluster A** — `run-load-test`, `audit-accessibility`, `audit-cost-efficiency` (launch checklist Performance/a11y/Cost row 보강)
3. **Phase 5 extension Cluster B** — `design-event-schema`, `design-auth-model`, `design-tenant-model` (SaaS 공통 패턴)
4. **A-4 plugin dogfood** — 실 프로젝트 install → orchestrator 동작 검증
5. **A-2 buddy MCP feature.* tools** — Q4 결정 trigger 발생 시 (feature registry 작업)
6. **B-1 Go CLI dogfood feedback** — 사용자 페이스
7. **deferred (production traffic / 1년+ 운영 후)** — Cluster F (§8 데이터 7) + Cluster G (§9 4)

자세한 priority 분석 + critical path: [`Phase 7 deferred re-evaluation`](./superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md).

---

## 참조

- [`HANDOFF.md`](./HANDOFF.md) — 세션 인계 + 워크플로우 skill 분기
- [`roadmap.md`](./roadmap.md) — Go CLI 마일스톤 SSoT
- [`superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md) — plugin 9-phase 아키텍처 SSoT
- [`v0.1-spec.md`](./v0.1-spec.md) — Go CLI v0.1 spec (LOCKED)
- [`skill-map.md`](./skill-map.md) — 11-stage → 9-phase 매핑 (참조용)
- [`DOGFOOD.md`](../DOGFOOD.md) + [`dogfood-feedback-template.md`](./dogfood-feedback-template.md) — Go CLI dogfood
