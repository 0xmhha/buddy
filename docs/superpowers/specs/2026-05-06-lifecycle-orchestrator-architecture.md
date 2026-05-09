# Buddy Plugin — Lifecycle Orchestrator Architecture

> **Date**: 2026-05-06
> **Scope**: 9-phase lifecycle orchestrator 모델 + 단일-router 라우팅 인프라 + stage skill 구현 현황. plugin 의 architecture / routing / command surface 의 현행 SSoT.
> **Related**: [`docs/skill-map.md`](../../skill-map.md), [`plugin/skills/router/SKILL.md`](../../../plugin/skills/router/SKILL.md), [`plugin/skills/router/references/routing-rules.md`](../../../plugin/skills/router/references/routing-rules.md), [`plugin/skills/router/references/skill-catalog.md`](../../../plugin/skills/router/references/skill-catalog.md).

---

## 0. 변경 요약 (2026-05-04 → 2026-05-06)

| 영역 | 2026-05-04 spec 가정 | 2026-05-06 현재 상태 |
|------|-------------------|------------------|
| **Skill discovery** | 78 개 SKILL.md 자동 로드 (≈28KB description 상시 컨텍스트 점유) | **router skill 1 개만** auto-discoverable. 나머지 78 procedure 는 `PROCEDURE.md` 로 rename되어 lazy-load. description 토큰 ≈99.44% 감축 (28,125 → 157 chars). |
| **Slash commands** | 17 → 12 (재정의 권장) → 26 (Q2=b 채택) | **30 commands** — 27 lifecycle commands + 3 신규 dispatch commands (`run` / `chain` / `parallel`). status command md 누락 보완. |
| **Routing files** | `plugin/SKILLS.md`, `plugin/SKILL_ROUTER.md` (top-level) | **`plugin/skills/router/references/`** 하위로 이동 — `skill-catalog.md`, `routing-rules.md`. router skill 의 lazy-load reference 로 통합. |
| **9-phase orchestrator 구현** | 미정 (신규 작성 필요) | **9/9 모두 구현** — `concretize-idea`, `define-features`, `design-system`, `plan-build`, `build-feature`, `verify-quality`, `ship-release`, `iterate-product`, `manage-lifecycle` 전부 PROCEDURE.md 존재. |
| **§4 gap skill (당초 67 개) 진행** | 모두 미구현 | **17 개 구현됨** (13 기존 + 4 Phase 1: define-tech-stack / design-data-model / design-api-contract / write-adr), 50 개 pending (§4 status 표 참조). |
| **MCP** | 7 개 설계 필요 (Q4=c "보류") | 보류 상태 유지. buddy MCP 핵심 controlplane (doctor / feature_* / stats) 만 동작 중. |
| **archive** | _archive 격리 (Q5=b) | 현 상태 유지. `route-intent`, `route-multi-platform`, `route-spec-to-code` 3개 격리됨. |

**라우팅 흐름 (현재)**:

```
/buddy:<name> "args"
    → plugin/commands/buddy/<name>.md (Skill 도구로 router 호출)
        → plugin/skills/router/SKILL.md (auto-loaded; mode/target 파싱)
            → ${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md (lazy Read)
                → procedure 본문 실행
```

라우팅 디테일 + path resolution fallback: [`plugin/skills/router/SKILL.md`](../../../plugin/skills/router/SKILL.md).

---

## 1. 배경 및 문제 정의

### 1.1 현재 상태 (갱신)

- buddy plugin 은 **78 개 skill** 을 보유 — 모두 `plugin/skills/<name>/PROCEDURE.md` 형태로 존재. 이 중 1 개(`router/SKILL.md`)만 auto-discoverable.
- **autoplan** 은 cross-phase review sub-orchestrator 로 위치 확정 (Q6=a). §1 / §3 / §4 phase 의 review stage 로 공유 사용.
- **30 개 slash commands** 노출 — 9 phase orchestrators + 17 stage / utility commands + 3 dispatch commands (run / chain / parallel) + status.

### 1.2 원래 문제 정의 (참고)

`autoplan` 을 라이프사이클 전체의 단일 orchestrator 로 가정한 모델은 폐기됨. 현재 multi-orchestrator 모델 채택 — 각 phase 가 자체 orchestrator 와 stage skill 군을 가짐.

### 1.3 목표 (달성 상태)

| 목표 | 상태 |
|------|-----|
| Multi-orchestrator 모델 채택 | [Done] 9 phase orchestrator 전부 구현 |
| Phase 자율성 + stage dual-mode | [Done] skill-catalog.md 에 priority 표 반영 |
| 상용 누락 영역 (UAT / 인시던트 / A/B / 코호트 / 피드백 / EOL) | [Partial] §8 의 design-ab-experiment / analyze-ab-experiment / handle-incident / conduct-postmortem 등 핵심 skill 구현. cohort / feedback corpus / cost anomaly 등 보강 필요 |
| 풀 사이클 단일 plugin 지원 | [Partial] §1~§9 orchestrator 모두 존재. stage 채움 진행 중 (§4 참조) |

---

## 2. 설계 원칙 (불변)

| 원칙 | 정의 | 영향 |
|------|------|------|
| **Phase autonomy** | 각 phase orchestrator 는 자체 결정 분기·stage·산출물·MCP 를 가짐. 다른 phase 의 stage 호출 금지. | orchestrator 간 의존성을 산출물 (artifact) 로 한정. |
| **Stage dual-mode** | Stage skill 은 ① orchestrator 안에서 호출 ② 사용자 명시 단독 호출 모두 가능. | User Sovereignty — 사용자가 stage 단독 호출 시 orchestrator 로 escalate 금지. |
| **Command = 시작 gate** | `/buddy:<name>` command 는 phase orchestrator 와 stage 진입점 + cross-cutting utility 만 노출. | 30 command 안에서도 orchestrator 9 + stage 17 + dispatch 3 + status 1 로 분류. |
| **MCP for cross-system** | Phase 간·외부 시스템 통합은 MCP 로 일원화. | buddy 자체 MCP 는 control plane (doctor / feature_*) 까지 구현. 외부 SaaS 어댑터는 보류. |
| **Commercial-first, not MVP** | "MVP 충분" 으로 빠질 만한 단계도 상용 기준에서는 분리. | §3 / §4 분리, §8 / §9 별도 phase 채택. |
| **Review-pipeline ≠ phase orchestrator** | `autoplan` 같은 review pipeline 은 cross-phase 공유 sub-orchestrator. | autoplan 을 §1 / §3 / §4 / §6 의 review stage 로 공유. |
| **Use case as primitive** | feature 는 여러 actor 의 use case 합성. actor / use case / system boundary 매핑이 §2 의 입력. | identify-actors → map-actor-use-cases → map-use-case-to-system-boundary → compose-feature-from-use-cases stage 모두 구현 완료. |
| **Single-router dispatch** *(2026-05-06 추가)* | 78 procedure 는 router skill 의 lazy-Read 로만 도달 가능. auto-discoverable 은 router 1 개. | description 토큰 99.44% 감축. 새 skill 추가 시 SKILL.md 가 아닌 PROCEDURE.md 로 작성. |

---

## 2.1 autoplan 의 위치 (확정)

**결론**: cross-phase review sub-orchestrator. §1 PRD draft / §3 ADR · API spec / §4 task plan / §6 release plan 어느 산출물에든 호출 가능. 단일 phase 종속 X.

`concretize-idea` (§1) 안의 stage 8 로 `autoplan` 을 invoke 하면 PRD draft → 4-mode review (review-scope / review-engineering / review-design / review-devex) 자동 통과. §3 / §4 도 동일 패턴.

---

## 3. 9-Phase 라이프사이클 모델 (구현 상태)

| Phase | Orchestrator | 진입 조건 | 산출물 | 상태 |
|-------|-------------|----------|--------|-----|
| §1 Idea & Business Validation | `concretize-idea` | idea/concept | PRD draft + business viability report | [Done] orchestrator 존재. stage 일부 보강 필요 |
| §2 Feature Definition & Backlog | `define-features` | PRD 확정 | Feature backlog (use case 분해 포함) | [Done] orchestrator + actor/use-case stage 5/5 존재 |
| §3 Technical Design | `design-system` | Feature backlog | Tech stack ADR, infra blueprint, API/data model | [Partial] orchestrator + 핵심 4 stage (define-tech-stack / design-data-model / design-api-contract / write-adr) 구현 (Phase 1 완료, v1.0.2). 부수 design-* (auth-model / observability / tenant-model 등) 5+ 개 pending — Phase 7 deferred |
| §4 Implementation Plan | `plan-build` | Technical design | Ordered task graph + parallelization plan | [Done] orchestrator + 6 stage (decompose-feature-to-actor-tracks / decompose-track-to-tasks / map-task-dependencies / plan-parallel-execution / define-acceptance-test-plan / estimate-build-timeline) 모두 구현 (Phase 2 완료, v1.0.3) |
| §5 Development | `build-feature` | Implementation plan | Working code + tests | [Done] orchestrator + TDD/parallel-agent stage 다수 존재 |
| §6 Quality | `verify-quality` | Code complete | QA report + security/legal sign-off | [Partial] orchestrator + 보안/리뷰 stage 존재 + use-case test 2 stage (test-per-actor-use-case / test-cross-actor-flow) 구현 (Phase 4 완료, v1.0.5). load / a11y / i18n / cost / chaos / mutation 6 보강 pending — Phase 7 deferred |
| §7 Release & Beta | `ship-release` | Quality gate pass | Tagged release, canary/UAT pass, GA | [Done] orchestrator + 14 stage (release prep 4 + safety nets 5 + UAT/beta 2 + GA 3) 모두 구현. Phase 3 완료 (v1.0.4) — canary / feature-flags / rollback / UAT / beta / launch-checklist / incident-paging 7 신규 stage 추가 |
| §8 Operate & Iterate | `iterate-product` | Production traffic | Experiment results, improvement backlog | [Partial] orchestrator + AB / funnel / incident / postmortem / improvement-tasks 존재. cohort / feedback-corpus / cost / SLO pending |
| §9 Lifecycle Management | `manage-lifecycle` | Feature/product 노후화 | Deprecation, migration, EOL | [Partial] orchestrator 존재. deprecate-feature / migrate-customers / archive-product / spin-off-feature pending |

### 3.1 §3 / §4 분리 근거 (불변)

§3 의사결정 권한자=Architect/Tech Lead, 시간 지평=다년, 락인 영향 큼. §4 권한자=Tech Lead/Eng Manager, 시간 지평=분기/스프린트, 변경 비용 낮음. → 상용은 분리해야 의사결정 누수 차단.

### 3.2 §8 / §9 commercial-only 성격 (불변)

§8 (Operate & Iterate) — A/B 실험·코호트·인시던트·고객 피드백 → 백로그. **현재 핵심 stage 구현 완료**. cohort / feedback corpus / cost anomaly / SLO 보강 필요.

§9 (Lifecycle Management) — feature deprecation, customer migration, product EOL. **stage 4개 모두 pending**. 1년 이상 운영 시 채움.

### 3.3 §0 (Discovery) 별도 phase 미채택 (불변)

시장조사·고객 인터뷰는 인간 영역. §1 안에 보조 stage skill 로 흡수 (`conduct-customer-interview`, `map-customer-segments`, `analyze-market-size` — 모두 pending).

### 3.4 Use Case 분해 — §2 의 핵심 (구현 완료)

Q8=(a) 채택 결과: §2 의 첫 단계는 actor 식별 → use case 매핑 → system boundary 매핑. 이후 feature 는 use case 합성의 *결과*.

**§2 stage 구현 상태**:

```
define-features (§2 phase orchestrator)
├── [Done] stage 1: identify-actors
├── [Done] stage 2: map-actor-use-cases
├── [Done] stage 3: map-use-case-to-system-boundary
├── [Done] stage 4: compose-feature-from-use-cases
├── [Done] stage 5: define-feature-spec
├── [Done] stage 6: query-feature-registry
├── [Done] stage 7: score-feature-priority
├── [Done] stage 8: map-feature-dependencies
├── [Done] stage 9: split-work-into-features
├── [Done] stage 10: triage-work-items
└── [Pending] stage 11: estimate-feature-effort (pending — story point/t-shirt sizing)
```

§2 는 1 개 stage 보강 외에는 완성. cross-phase cascade 활용 (use case 분해 결과가 §3 infra / §4 implementation / §6 test / §8 metric 의 입력 schema 가 됨) 은 §3~§8 stage 보강 진행에 따라 자연스럽게 활성화됨.

---

## 4. 단계별 Skill 군집화 + Gap 분석 (현황)

> [Done] = 구현됨 (PROCEDURE.md 존재)
> [Pending] = 미구현, 신규 작성 필요
> [Renamed] = 명칭 변경/통합되어 다른 skill 로 흡수됨

### §1 `concretize-idea`

**구현됨**: [Done] `validate-idea`, [Done] `validate-advanced-edge-idea`, [Done] `assess-business-viability`, [Done] `review-pricing-and-gtm`, [Done] `define-product-spec`, [Done] `apply-builder-ethos`, [Done] `autoplan` (review sub-orchestrator), [Done] `review-scope`, [Done] `critique-plan`.

**Pending (상용 필수)**:
- [Pending] `analyze-competition-and-substitutes` — 경쟁/대체재 매트릭스
- [Pending] `map-customer-segments` — 고객 세그먼트와 구매자 분리
- [Pending] `map-jobs-to-be-done` — JTBD 프레임 인터뷰
- [Pending] `analyze-market-size` — TAM/SAM/SOM 정량화
- [Pending] `conduct-customer-interview` — 인터뷰 스크립트 + 합성

### §2 `define-features` — Use Case Mapping & Feature Definition

**구현됨**: [Done] `identify-actors`, [Done] `map-actor-use-cases`, [Done] `map-use-case-to-system-boundary`, [Done] `compose-feature-from-use-cases`, [Done] `define-feature-spec`, [Done] `score-feature-priority`, [Done] `map-feature-dependencies`, [Done] `query-feature-registry`, [Done] `triage-work-items`, [Done] `split-work-into-features`.

**Pending**:
- [Pending] `estimate-feature-effort` — story point / t-shirt sizing

### §3 `design-system`

**구현됨 (Phase 1 + 기존)**: [Done] `define-tech-stack`, [Done] `design-data-model`, [Done] `design-api-contract`, [Done] `write-adr`, [Done] `review-architecture`, [Done] `review-engineering`, [Done] `review-design`, [Done] `review-devex`, [Done] `design-artifact-storage`, [Done] `design-billing-system`, [Done] `design-claude-hooks`, [Done] `design-deploy-strategy`, [Done] `design-embedding-search`, [Done] `design-mcp-server`, [Done] `consult-codex`, [Done] `consult-design-system`, [Done] `explore-design-variants`, [Done] `autoplan` (review).

**Pending (use case → infra 브릿지)**:
- [Pending] `map-use-cases-to-infra` — actor system boundary → 실제 infra component
- [Pending] `derive-system-topology` — actor 그래프 → 시스템 토폴로지

**Pending (상용 부수 design-*)**:
- [Pending] `design-event-schema`, [Pending] `design-auth-model`, [Pending] `design-observability`, [Pending] `design-secret-management`, [Pending] `design-tenant-model`, [Pending] `design-i18n-strategy`, [Pending] `design-accessibility-baseline`

### §4 `plan-build`

**구현됨**: [Done] orchestrator (`plan-build`), [Done] `autoplan` 호출 가능, [Done] `decompose-feature-to-actor-tracks`, [Done] `decompose-track-to-tasks`, [Done] `map-task-dependencies`, [Done] `plan-parallel-execution`, [Done] `define-acceptance-test-plan`, [Done] `estimate-build-timeline` (Phase 2 완료, v1.0.3).

**Pending**: 없음 — §4 핵심 stage 완성.

### §5 `build-feature`

**구현됨**: [Done] `build-with-tdd`, [Done] `iterate-fix-verify`, [Done] `freeze-edit-scope`, [Done] `dispatch-parallel-agents`, [Done] `diagnose-bug`, [Done] `consult-codex`.

**Pending**:
- [Pending] `generate-from-api-contract`, [Pending] `generate-tests-from-spec`, [Pending] `pair-program-loop`, [Pending] `refactor-with-rename-trace`, [Pending] `update-docs-with-code`

### §6 `verify-quality`

**구현됨**: [Done] `classify-qa-tiers`, [Done] `run-browser-qa`, [Done] `monitor-regressions`, [Done] `audit-security`, [Done] `audit-live-devex`, [Done] `measure-code-health`, [Done] `classify-review-risks`, [Done] `review-ai-safety-liability`, [Done] `review-privacy-data-risk`, [Done] `review-license-and-ip-risk`, [Done] `review-terms-policy-readiness`, [Done] `test-per-actor-use-case`, [Done] `test-cross-actor-flow` (Phase 4 완료, v1.0.5).

**Pending (상용 발표 전 필수)**:
- [Pending] `run-load-test`, [Pending] `audit-accessibility`, [Pending] `audit-i18n-coverage`, [Pending] `audit-cost-efficiency`, [Pending] `chaos-test`, [Pending] `audit-test-coverage-meaningful`

### §7 `ship-release`

**구현됨**: [Done] `setup-quality-gates`, [Done] `auto-create-pr`, [Done] `automate-release-tagging`, [Done] `sync-release-docs`, [Done] `write-changelog`, [Done] `guard-destructive-commands`, [Done] `compose-safety-mode`, [Done] `setup-canary-deploy`, [Done] `setup-feature-flags`, [Done] `setup-rollback-runbook`, [Done] `run-uat`, [Done] `run-beta-program`, [Done] `prepare-launch-checklist`, [Done] `setup-incident-paging` (Phase 3 완료, v1.0.4).

**Pending**: 없음 — §7 핵심 + 안전망 stage 완성. orchestrator 9 → 14 stage 확장.

### §8 `iterate-product` (사용자 명시 영역)

**구현됨**: [Done] `monitor-regressions`, [Done] `save-context`, [Done] `restore-context`, [Done] `summarize-retro`, [Done] `persist-learning-jsonl`, [Done] `design-ab-experiment`, [Done] `analyze-ab-experiment`, [Done] `analyze-user-funnel`, [Done] `generate-improvement-tasks`, [Done] `handle-incident`, [Done] `conduct-postmortem`.

**Pending**:
- [Pending] `analyze-feature-adoption`, [Pending] `analyze-user-cohort`, [Pending] `analyze-actor-failure-rate`, [Pending] `analyze-cost-anomaly`, [Pending] `triage-customer-support-ticket`, [Pending] `analyze-customer-feedback-corpus`, [Pending] `audit-error-budget`

### §9 `manage-lifecycle` (상용 장기운영)

**구현됨**: [Done] orchestrator only (`manage-lifecycle`).

**Pending**:
- [Pending] `deprecate-feature`, [Pending] `migrate-customers`, [Pending] `archive-product`, [Pending] `spin-off-feature`

### Cross-cutting

- [Done] `apply-builder-ethos`, [Done] `detect-install-type`, [Done] `guide-setup-wizard`, [Done] `benchmark-llm-models`
- [Renamed] archive 3개 (`route-intent`, `route-multi-platform`, `route-spec-to-code`) — `plugin/_archive/` 격리됨

---

## 5. MCP 요구사항 (보류 유지)

Q4=(c) 결정에 따라 MCP 작성 보류. 현재 buddy MCP control plane (doctor / feature_list / feature_get / feature_search / feature_upsert / feature_delete / stats) 는 동작.

| MCP | 우선순위 | 현재 상태 |
|-----|--------|----------|
| `feature-management-mcp` | 1순위 (in-house) | buddy MCP 의 feature_* 도구로 흡수됨 (부분) |
| `analytics-mcp` | 2순위 (in-house) | 미착수 |
| `monitoring-mcp` / `support-mcp` / `cost-mcp` / `billing-mcp` / `feature-flag-mcp` | 외부 SaaS 어댑터 | 미착수 |

---

## 6. Command Gate (현황)

### 6.1 Phase orchestrator commands (9, [Done] 모두 등록)

`/buddy:concretize-idea`, `/buddy:define-features`, `/buddy:design-system`, `/buddy:plan-build`, `/buddy:build-feature`, `/buddy:verify-quality`, `/buddy:ship-release`, `/buddy:iterate-product`, `/buddy:manage-lifecycle`.

### 6.2 Cross-cutting utility commands (3, [Done] 등록)

`/buddy:save-context`, `/buddy:restore-context`, `/buddy:consult-codex`.

### 6.3 Stage / Domain commands (15, [Done] 등록)

Q2=(b) "dual-full" 결정에 따라 실제로는 14 개 제거 대신 보존:
`/buddy:validate-idea`, `/buddy:validate-advanced-edge-idea`, `/buddy:assess-business-viability`, `/buddy:define-product-spec`, `/buddy:autoplan`, `/buddy:explore-design-variants`, `/buddy:dispatch-parallel-agents`, `/buddy:build-with-tdd`, `/buddy:diagnose-bug`, `/buddy:audit-security`, `/buddy:measure-code-health`, `/buddy:auto-create-pr`, `/buddy:setup-quality-gates`, `/buddy:summarize-retro`, `/buddy:status`.

### 6.4 Dispatch commands (3, 신규 — 2026-05-06)

`/buddy:run`, `/buddy:chain`, `/buddy:parallel` — router 의 single / chain / parallel mode 직접 진입. 전용 command 가 없는 stage skill 호출 + 다중 skill 합성 지원.

→ **총 30 commands** (9 phase + 3 utility + 15 stage/domain + 3 dispatch). 모두 plugin.json `"skill": "router"` 통일.

### 6.5 SKILL_ROUTER 도메인 우선순위 표

`plugin/skills/router/references/routing-rules.md` §2 에서 관리. 현 우선순위:

1. Phase orchestrator
2. Cross-phase review sub-orchestrator (`autoplan`)
3. Stage skill (dual-mode)
4. Domain skill
5. Pattern library (직접 dispatch 금지)
6. Archive

---

## 7. 작업량 견적 (잔여)

| 항목 | 잔여 수량 | 비고 |
|------|---------|------|
| Stage skill 신규 작성 | **35** | Phase 1+2+3+4 완료 후 잔여 — §1 (5), §2 (1 small), §3 부가 (9), §5 부가 (5), §6 부가 (6 — load/a11y/i18n/cost/chaos/mutation), §8 (7), §9 (4) — 모두 Phase 7 deferred 또는 별도 commercial trigger |
| In-house MCP | 2 | feature-management-mcp 부분 구현됨, analytics-mcp 미착수 |
| 외부 SaaS MCP 어댑터 | 5 | monitoring/support/cost/billing/feature-flag |
| Command 정렬 | [Done] 완료 | 49 commands (v1.0.5 기준) 모두 router 통한 dispatch — Phase 1 (+4) + Phase 2 (+6) + Phase 3 (+7) + Phase 4 (+2) |
| 라우팅 인프라 | [Done] 완료 | router/SKILL.md, references/, dispatch contract |
| 문서 | [Done] 부분 완료 | skill-catalog / routing-rules / spec / plan / README / CHANGELOG 갱신 완료 (v1.0.5 까지) |

---

## 8. 결정 포인트 — 결과 (Q1~Q8)

| Q | 채택 | 결과 |
|---|------|------|
| Q1: 9-phase 모델 | (a) 9-phase 그대로 | [Done] 9 orchestrator 모두 구현 |
| Q2: Command 재정의 | (b) dual full | [Done] 30 commands (9 phase + 3 utility + 15 stage + 3 dispatch) |
| Q3: skill 작성 우선순위 | (c) §8 → (b) §1~§5 | [Partial] §3 / §4 / §6 use-case test / §7 safety net 모두 핵심 완료 (Phase 1-4). §1 / §3 부가 / §5 부가 / §6 보강 / §8 보강 / §9 Phase 7 deferred |
| Q4: MCP 우선순위 | (c) MCP 보류 | [Done] 보류 유지 — buddy MCP control plane 만 동작 |
| Q5: archive 처리 | (b) `_archive/` 격리 | [Done] 3개 격리됨 |
| Q6: autoplan 위치 | (a) cross-phase review sub-orchestrator | [Done] 적용됨 — §1 stage 8 / §3 review / §4 review 로 호출 |
| Q7: phase 추가 | (b) §7.5 Beta/UAT 분리 | [Done] §7 안의 7-2 / 7-3 단계로 흡수 — run-uat + run-beta-program 신규 (Phase 3, v1.0.4). 별도 phase 분리 미채택 (현재 stage 분류로 충분 판단) |
| Q8: Use case 분해 | (a) §2 첫 단계 강제 | [Done] §2 5 stage + §4 actor track 분해 6 stage + §6 use-case test 2 stage 완성 — Q8=(a) cascade 5-단계 chain (use case → system → actor track → build → test) 완전 활성화 (Phase 4, v1.0.5) |

---

## 9. 참조

- 11-stage 상용 제품 빌딩 flow → [`docs/skill-map.md`](../../skill-map.md)
- 현재 skill router → [`plugin/skills/router/SKILL.md`](../../../plugin/skills/router/SKILL.md)
- routing rules → [`plugin/skills/router/references/routing-rules.md`](../../../plugin/skills/router/references/routing-rules.md)
- skill catalog → [`plugin/skills/router/references/skill-catalog.md`](../../../plugin/skills/router/references/skill-catalog.md)

> 라우팅 리팩터 (78 SKILL.md → 1 router + 78 PROCEDURE.md, description 토큰 99.44% 감축) 는 git 히스토리 `b2627b1..b2f0ce1` (2026-05-06) 에 기록됨.

---

## 10. 다음 단계 (잔여)

1. **§3 design 핵심 stage 작성** — `define-tech-stack`, `design-data-model`, `design-api-contract`, `write-adr` 우선 (락인 영향 큼).
2. **§4 plan-build stage 작성** — feature → actor track 분해 + task DAG. 6개 모두 영향 큼.
3. **§7 ship-release 안전망 stage** — canary / feature-flags / rollback / UAT — 상용 배포 직전 필수.
4. **§8 데이터 분석 보강** — cohort / feedback corpus / cost / SLO — 핵심 funnel/AB 구현 후 다음 우선.
5. **§9 lifecycle stage 작성** — 1년 이후 운영 시 채움. 우선순위는 후순위.
6. **§6 use-case 기반 테스트 stage** — `test-per-actor-use-case`, `test-cross-actor-flow` — Q8=(a) cascade 활성화에 직접 기여.
7. **MCP 단계** (Q4=c 보류 해제 시) — feature-management-mcp 완성 → analytics-mcp 신규.
8. **Live smoke test** — 라우팅 리팩터 후 실제 Claude Code 세션에서 30 commands 동작 검증 (현재는 static wire-up 만 검증됨, Phase 6.2 verification log 참조).
9. **Description drift cleanup** — plugin.json ↔ command md frontmatter 의 26 개 description 차이 정렬 (별도 cleanup task).

## 명령

- 잔여 작업은 우선순위에 따라 별도 plan 으로 쪼개어 단계 실행할 것.
- 본 spec 의 status 표는 stage skill 추가/삭제 시 갱신할 것 — `plugin/skills/router/references/skill-catalog.md` 가 1차 SSoT, 본 spec 은 cross-phase 진척도 SSoT.
