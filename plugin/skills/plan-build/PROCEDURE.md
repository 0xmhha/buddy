# plan-build — 4단계 Implementation Plan Orchestrator

4단계 라이프사이클 단계의 진입점. 2단계 feature spec + 3단계 system topology → ordered task graph with dependencies + parallelization plan.

**진입 조건**: 3단계 technical design 확정 (ADR + API contract + data model).
**산출물**: Actor별 ordered task list + dependency DAG + parallel execution plan + build timeline.
**다음 phase**: Implementation plan 확정 후 → `build-feature` (5단계).

---

## 3단계/4단계 분리 근거

| 항목 | 3단계 Technical Design | 4단계 Implementation Plan |
|------|---------------------|----------------------|
| 의사결정 권한자 | Architect / Tech Lead | Tech Lead / Eng Manager |
| 시간 지평 | 다년 (락인 영향) | 분기/스프린트 |
| 결정 단위 | 언어/프레임워크/DB/tenancy | task 분해/의존성/병렬화 |
| 변경 비용 | 매우 높음 | 낮음 (재계획 가능) |

---

## Stage 흐름

```
plan-build (4단계 phase orchestrator)
├── stage 1: decompose-feature-to-actor-tracks  (feature → actor별 task track)
├── stage 2: decompose-track-to-tasks           (actor track → ordered task list)
├── stage 3: map-task-dependencies              (task DAG — actor 내부 + actor 간 contract)
├── stage 4: plan-parallel-execution            (actor track별 병렬 worker 분배)
├── stage 5: define-acceptance-test-plan        (actor별 + cross-actor 완료 기준)
├── stage 6: estimate-build-timeline            (의존성 + 병렬도 → 일정 합성)
└── stage 7: autoplan                           (task plan 4-mode review)
```

> 모든 stage는 신규 작성 필요. 현재 orchestrator가 직접 수행.

---

## 실행 절차

### Stage 1: Feature → Actor Track 분해

`decompose-feature-to-actor-tracks` skill 을 invoke 한다 — §3 design 산출물 (tech stack / data model / API contract) + §2 feature spec 의 actor / use case / system boundary 를 입력으로 actor 별 implementation track 으로 분해. cross-track contract 식별 + Independence Matrix + Track Outputs 로 후속 stage 의 입력 schema 생성.

호출 형태:
- 단독: `/buddy:decompose-feature-to-actor-tracks "<feature>"`
- chain (권장): Stage 1~6 일괄 — 본 phase 끝의 권장 chain 패턴 참조

### Stage 2: Actor Track → Task List

`decompose-track-to-tasks` skill 을 invoke 한다 — 각 track 을 atomic task (single PR scope) 로 분해. naming convention + acceptance criteria + diff size + internal dependency edge 강제. 산출물은 `map-task-dependencies` 의 입력.

호출 형태: `/buddy:decompose-track-to-tasks "<track table 또는 feature>"`

Task 필수 속성:
```yaml
task_id: {actor-track}-{N}
title: {동사 + 명사}
actor_track: {frontend / backend / 3rd-party}
estimated_hours: {N}
dependencies: [{task_id, ...}]
acceptance: {완료 판단 기준}
```

### Stage 3: Task Dependency DAG

`map-task-dependencies` skill 을 invoke 한다 — internal (intra-track) + cross-actor (contract-based) edges 통합 + cycle 감지 + critical path 계산 + parallel-safe levels 식별.

호출 형태: `/buddy:map-task-dependencies "<task list 또는 feature>"`

이전 stage 와의 차이:

actor 내부 의존성과 actor 간 contract 의존성을 DAG로 표현한다.

```
frontend-1 (signup form) → frontend-2 (validation UI) ← backend-1 (API spec)
backend-1 (credentials endpoint) → backend-2 (JWT endpoint)
3rdparty-1 (email template) → backend-3 (webhook handler)
```

Critical path를 식별한다: 전체 feature의 완료를 block하는 task 체인.

### Stage 4: 병렬 실행 계획

`plan-parallel-execution` skill 을 invoke 한다 — DAG (Stage 3) + worker capability matrix → batch schedule + sync points + bottleneck mitigation. AI agent (`dispatch-parallel-agents`) 와 인간 worker 혼합 plan.

호출 형태: `/buddy:plan-parallel-execution "<DAG 또는 feature>"`

산출물 예시 (간이 형태 — 실제는 batch schedule + worker capability matrix + sync points 표):

```yaml
parallel_tracks:
  - track: frontend
    worker: agent-1
    tasks: [frontend-1, frontend-2, frontend-3]
  - track: backend
    worker: agent-2
    tasks: [backend-1, backend-2, backend-3]
synchronization_points:
  - after: [backend-1]
    before: [frontend-2]
```

### Stage 5: Acceptance Test Plan

`define-acceptance-test-plan` skill 을 invoke 한다 — per-actor (unit/integration/contract) + cross-actor (E2E flow) test plan + test infra 결정 + acceptance gate. §6 verify-quality 의 입력.

호출 형태: `/buddy:define-acceptance-test-plan "<feature 또는 task DAG>"`

산출물 예시 (간이):

```yaml
per_actor:
  frontend: "signup form submit → success redirect 동작"
  backend: "POST /auth/signup → 201 Created + JWT 반환"
  3rd-party: "verification email 수신 → click → 200 OK webhook"
integration:
  cross_actor: "full signup flow: form submit → email verification → first login"
```

### Stage 6: 빌드 타임라인

`estimate-build-timeline` skill 을 invoke 한다 — task DAG + batch schedule + per-task duration → calendar timeline 합성. confidence interval (best/expected/p90/worst) + risk buffer + holiday/availability 반영.

호출 형태: `/buddy:estimate-build-timeline "<DAG / batch / start date>"`

산출물 예시 (간이):

```
Critical path: 25 real-h (frontend track 의 main flow)
p50 (expected): 5 calendar days
p90 (commit 권장): 7 calendar days
Worst: 9 days (known risk: 3rd-party API uncertainty)
```

### Stage 7: autoplan Review

`autoplan`을 invoke해 task plan을 4-mode review한다.

---

## 다음 phase

- `/buddy:build-feature` — 5단계 Development (권장)
- `/buddy:dispatch-parallel-agents` — 병렬 agent 분배를 즉시 시작할 때

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` §§4, §3.1
