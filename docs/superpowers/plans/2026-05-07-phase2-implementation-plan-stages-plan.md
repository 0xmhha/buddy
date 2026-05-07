# Phase 2 — §4 Implementation Plan 6 Stage Skill 작성 Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement task-by-task.

**Goal:** §4 Implementation Plan phase 의 6 stage skill 을 작성해 `plan-build` orchestrator 가 단순 진입점에서 actor-track 기반 병렬 실행 plan 까지 produce 하도록 만든다. Phase 1 의 §3 산출물 (tech stack, data model, API contract, ADR) 을 입력으로 받아 §5 build-feature 의 actor-별 task 입력으로 cascade.

**Parent plan:** [`2026-05-06-stage-buildout-plan.md`](./2026-05-06-stage-buildout-plan.md) Phase 2.
**SSoT:** [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../specs/2026-05-06-lifecycle-orchestrator-architecture.md) §3 + §4 §4.
**Predecessor plan:** [`2026-05-07-phase1-design-stages-plan.md`](./2026-05-07-phase1-design-stages-plan.md) — PROCEDURE.md template established here, reused.

## Scope (6 skills)

| Skill name | 1줄 용도 | 의존 |
|-----------|---------|------|
| `decompose-feature-to-actor-tracks` | feature → actor 별 implementation track (frontend / backend / 3rd-party) 분해 | §2 feature spec, §3 system topology |
| `decompose-track-to-tasks` | 각 actor track → ordered task list (atomic units) | track 분해 결과 |
| `map-task-dependencies` | task 간 선후 그래프 (actor 내부 + cross-actor contract) | task list |
| `plan-parallel-execution` | actor track 별 병렬 worker 분배 + 동기화 지점 | task DAG |
| `define-acceptance-test-plan` | actor 별 + cross-actor 완료 기준 + test infra 매핑 | feature spec + task DAG |
| `estimate-build-timeline` | 의존성 + 병렬도 → 일정 합성 (critical path) | task DAG + parallel plan |

**왜 이 6개:** §4 의 commercial 가치를 잠재 → 실제 작동으로 전환. plan-build orchestrator 가 현재 stage 가 거의 비어 있어 §5 build-feature 가 actor-track 분배 입력 없이 진입한다. 본 phase 가 §3 → §5 의 missing link.

**Use case cascade 활성화 (Q8=a):**
- §2 use case 분해 → §3 system boundary → **§4 actor-track 분배** → §5 actor-별 implementation
이 4-단계 chain 의 §4 가 본 phase 에서 채워진다.

## Out of scope

- §4 의 미정의 보조 skill (예: capacity planning, sprint sizing) — 별도 plan
- agile / waterfall methodology 결정 — orchestrator 의 입력으로 가정
- 외부 PM 도구 (Jira / Linear) 통합 — 별도 MCP 영역

## 진행 원칙 (Phase 1 와 동일)

- **순차 작성** — Task 2.1 (decompose-feature-to-actor-tracks) 가 후속 5 skill 의 입력 schema 를 정의. 첫 skill 의 quality 가 chain 전체에 cascade.
- **각 skill = 단일 commit** (per task)
- **PROCEDURE template 재사용** — Phase 1 의 §0 STOP / §4 posture / §6 explicit output / §11 verification gate 4 enhancement 를 그대로 적용.

---

## PROCEDURE template

Phase 1 plan 의 [common template](2026-05-07-phase1-design-stages-plan.md#proceduremd-공통-템플릿) 그대로 사용. 모든 §4 skill 은 frontmatter 없이 §0 ~ §11 의 12 sections 를 가진다.

---

## Tasks

### Task 2.1: `decompose-feature-to-actor-tracks` (template skill)

**Files:**
- Create: `plugin/skills/decompose-feature-to-actor-tracks/PROCEDURE.md`
- Create: `plugin/commands/decompose-feature-to-actor-tracks.md`
- Modify: `plugin/skills/router/references/skill-catalog.md` (§4 또는 신규 §4 stage skill section 행 추가)
- Modify: `plugin/skills/plan-build/PROCEDURE.md` (stage 흐름에 매핑)

**Skill 정의:**
- description: "feature 를 actor 별 implementation track 으로 분해 (frontend / backend / 3rd-party / data) — Q8=(a) cascade 의 §4 진입점. 각 track 의 system boundary + 책임 + interface 명시."
- when-to-use: §3 design 산출물 받은 직후 / actor track 별 worker 배정 전 / cross-actor contract 변경으로 재분해 필요 시
- inputs: feature spec (actor list + per-actor use cases + system boundary), §3 system topology, tech stack, API contract
- outputs: actor track table (track id, owner actor, system boundary, interface contracts), 각 track 의 책임 sub-system, dependency edges (track 간 contract 위치)
- 충돌 방지: vs `decompose-track-to-tasks` (본 skill = track 단위, 다음 skill = task 단위) / vs `map-feature-dependencies` (그것은 feature 간, 본 skill 은 feature 내부 actor 간)

**PROCEDURE 본문 골자 (§5 Phases):**
1. **Actor → track 매핑** — 각 actor 의 system boundary 가 어느 track 에 속하는지 (예: User actor → frontend-spa-track + auth-service-track)
2. **Track 책임 분해** — 각 track 의 sub-system / 데이터 / 외부 의존
3. **Cross-track contract 식별** — track 간 interface (API / event / shared DB / message queue)
4. **Independent vs sequential 판단** — track 간 의존성 강도 (parallel safe / partial / strict-sequential)
5. **Track output 정리** — 각 track 의 deliverable (실행 결과)

**Acceptance:** 표준 (PROCEDURE/command 신규 + catalog/orchestrator 갱신 + CI 82→83)

**Commit:** `feat(skill): add decompose-feature-to-actor-tracks §4 stage skill`

### Task 2.2: `decompose-track-to-tasks`

**Skill 정의:**
- description: "actor track 을 ordered task list 로 분해 — 각 task 는 atomic unit (single PR scope), input/output 명시, 검증 가능 acceptance"
- when-to-use: track 분해 결과 받은 직후 / 큰 task 가 multi-PR 로 깨질 위험 발견 시 / sprint 계획 입력
- inputs: track table (Task 2.1 산출물)
- outputs: 각 track 의 ordered task list (id, name, scope, expected diff size, acceptance criteria, dependencies within track)

**PROCEDURE 골자:**
1. **Atomic unit 정의** — task 가 1 PR 로 커버되는지 검증 (LoC 추정, 영향 file 수)
2. **Acceptance criteria 강제** — 각 task 가 verify 가능한 acceptance 명시
3. **Internal dependency 추출** — track 내부 task 간 선후
4. **Diff size 추정** — small / medium / large (각 임계 LoC 기준)
5. **Subdivide too-large tasks** — large 가 발견되면 분할

**Acceptance:** PROCEDURE/command 신규 + CI 83→84

**Commit:** `feat(skill): add decompose-track-to-tasks §4 stage skill`

### Task 2.3: `map-task-dependencies`

**Skill 정의:**
- description: "task DAG 작성 — actor 내부 dependency + cross-actor contract dependency 를 명시. critical path 와 parallel-safe 그룹 식별."
- when-to-use: track + task 분해 후 / parallel 실행 plan 작성 직전 / dependency cycle 감지
- inputs: per-track task list (Task 2.2 산출물), Phase 1 의 API contract (cross-actor dependency 출처)
- outputs: DAG (graphviz / mermaid / adjacency list), critical path, parallel-safe task 그룹

**PROCEDURE 골자:**
1. **Internal edge 수집** — 각 track 내부 task 의 direct dependency
2. **Cross-actor edge 수집** — API contract / event schema 가 정의한 contract 의존
3. **Cycle 감지** — DAG validation, cycle 있으면 분해 재검토
4. **Critical path 계산** — longest dependency chain
5. **Parallel-safe 그룹** — 같은 level 의 independent tasks

**Acceptance:** CI 84→85

**Commit:** `feat(skill): add map-task-dependencies §4 stage skill`

### Task 2.4: `plan-parallel-execution`

**Skill 정의:**
- description: "actor track 별 worker (인간 또는 AI agent) 분배 + 동기화 지점 명시 — Phase 1 의 dispatch-parallel-agents 와 결합 가능한 plan."
- when-to-use: task DAG 확정 후 / sprint kickoff / parallel agent dispatch 전
- inputs: task DAG (Task 2.3 산출물), available worker pool (인간 + agent 수), worker capability (어느 stack / actor 처리 가능)
- outputs: worker → track 배정, 동기화 지점 (hand-off / merge gate), parallel batch schedule

**PROCEDURE 골자:**
1. **Worker capability matrix** — 각 worker 가 어느 actor / stack / task type 가능한지
2. **배정 알고리즘** — capability fit + load balancing + critical path 우선
3. **동기화 지점 식별** — cross-track hand-off, merge gate, integration test 지점
4. **Batch schedule** — parallel-safe 그룹별 batch, batch 간 hand-off
5. **Bottleneck 분석** — 단일 worker / single-threaded task 식별

**Acceptance:** CI 85→86

**Commit:** `feat(skill): add plan-parallel-execution §4 stage skill`

### Task 2.5: `define-acceptance-test-plan`

**Skill 정의:**
- description: "actor 별 + cross-actor 완료 기준 + test infra 매핑 — feature spec 의 acceptance criteria 를 actor track 별 verifiable test plan 으로 변환."
- when-to-use: task DAG + parallel plan 확정 후 / pre-build verification gate 정의 / test infra 결정 지점
- inputs: feature spec (per-actor acceptance criteria), task DAG, tech stack (test framework 결정)
- outputs: per-actor test plan (unit / integration / contract / E2E), cross-actor flow test, test infra 결정 (CI runner / fixture / mock 전략)

**PROCEDURE 골자:**
1. **Per-actor test category 매핑** — frontend → component + E2E, backend → unit + integration, 3rd-party → contract test
2. **Cross-actor flow test** — actor 협업 시나리오 (예: signup full flow)
3. **Test infra 결정** — runner, fixture management, mock vs real DB
4. **Acceptance gate** — 어느 test 통과율이 build-feature 완료 조건
5. **§6 verify-quality cascade** — 본 plan 이 §6 의 입력 schema

**Acceptance:** CI 86→87

**Commit:** `feat(skill): add define-acceptance-test-plan §4 stage skill`

### Task 2.6: `estimate-build-timeline`

**Skill 정의:**
- description: "의존성 + 병렬도 → 일정 합성 — task DAG 와 worker 배정 plan 으로 critical path 기반 timeline 추정. confidence interval + risk buffer 포함."
- when-to-use: parallel plan 확정 후 / sprint planning / external commitment 직전
- inputs: task DAG, parallel execution plan (worker batch), per-task duration estimate (T-shirt sizing 또는 historical)
- outputs: critical path timeline, parallel batch schedule, total duration with confidence interval, risk buffer 권장

**PROCEDURE 골자:**
1. **Per-task duration 입력** — t-shirt (S/M/L/XL) 또는 historical pattern 매핑
2. **Critical path 계산** — DAG 의 longest path × parallel-aware schedule
3. **Confidence interval** — best-case / worst-case (1.5x ~ 2x worst pattern)
4. **Risk buffer** — known unknown 비례 buffer 추가
5. **Calendar 매핑** — 영업일 / holidays / worker availability 반영

**Acceptance:** CI 87→88

**Commit:** `feat(skill): add estimate-build-timeline §4 stage skill`

### Task 2.7: plan-build orchestrator stage 흐름 갱신

**Files:**
- Modify: `plugin/skills/plan-build/PROCEDURE.md`

**작업:**
- plan-build PROCEDURE 의 stage 흐름에 6 신규 skill 매핑 추가
- 권장 chain 패턴 명시:
  ```
  /buddy:chain decompose-feature-to-actor-tracks,decompose-track-to-tasks,map-task-dependencies,plan-parallel-execution,define-acceptance-test-plan,estimate-build-timeline -- "<feature>"
  ```
- §3 design-system → §4 plan-build 의 input cascade 명시 (어느 skill 의 어느 산출물이 어디로)

**Commit:** `feat(skill): wire 6 new §4 stages into plan-build orchestrator`

### Task 2.8: v1.0.3 release

**Files:**
- Modify: `scripts/test-router-wireup.sh` (PROCEDURE count 82 → 88)
- Modify: `plugin/.claude-plugin/plugin.json` (1.0.2 → 1.0.3)
- Modify: `.claude-plugin/marketplace.json` (sync)
- Modify: `CHANGELOG.md` (1.0.3 entry)

**Commit:** `release: v1.0.3 (§4 plan-build 6 stage skills)`

---

## Verification (Phase 2 종료 시점)

자동:
- `bash scripts/test-router-wireup.sh` 통과
- `claude plugin validate plugin/` 통과
- PROCEDURE.md count = 88
- 6 신규 skill 모두 frontmatter 없이 작성됨

수동 (Phase 1 + 2 누적 live retest — 본 phase 종료 후 (c) 옵션 진행 시점):
- `claude plugin marketplace update buddy && claude plugin update buddy@buddy` (1.0.3 fetch)
- 6 신규 슬래시 + Phase 1 의 4 신규 슬래시 모두 호출 검증
- 큰 chain test: `/buddy:chain decompose-feature-to-actor-tracks,decompose-track-to-tasks,map-task-dependencies -- "<test feature>"`

문서:
- spec `2026-05-06-lifecycle-orchestrator-architecture.md` §3 의 §4 row 갱신 (Partial 유지하되 6 stage 구현 표기) — 별도 commit

---

## Rollback

각 Task 가 독립 commit 이라 phase 단위 / task 단위 `git revert` 가능. 6 skill 모두 신규 파일이고 기존 파일 수정은 plan-build/PROCEDURE.md 와 catalog/CI/CHANGELOG 만이라 영향 범위 좁음.

---

## 다음 plan (Phase 3)

Phase 2 통과 후:
- Phase 3 (`§7 Release safety nets 7 skills`) plan 작성
- 또는 (c) live retest 우선 — Phase 1+2 누적 검증 후 Phase 3 진입

본 plan 의 §4 PROCEDURE template 가 Phase 3 의 작성 비용을 더 줄여줄 것 (Phase 1 정립 → Phase 2 재사용 → Phase 3 안정화).
