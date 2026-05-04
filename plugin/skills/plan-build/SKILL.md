---
name: plan-build
description: This skill should be used when the user wants to "plan the implementation", "create a task plan", "decompose features into tasks", "plan parallel development", "create sprint plan", or has technical design ready and needs to create an ordered implementation plan. Orchestrates §4 Implementation Plan phase.
---

# plan-build — §4 Implementation Plan Orchestrator

§4 라이프사이클 단계의 진입점. §2 feature spec + §3 system topology → ordered task graph with dependencies + parallelization plan.

**진입 조건**: §3 technical design 확정 (ADR + API contract + data model).
**산출물**: Actor별 ordered task list + dependency DAG + parallel execution plan + build timeline.
**다음 phase**: Implementation plan 확정 후 → `build-feature` (§5).

---

## §3/§4 분리 근거

| 항목 | §3 Technical Design | §4 Implementation Plan |
|------|---------------------|----------------------|
| 의사결정 권한자 | Architect / Tech Lead | Tech Lead / Eng Manager |
| 시간 지평 | 다년 (락인 영향) | 분기/스프린트 |
| 결정 단위 | 언어/프레임워크/DB/tenancy | task 분해/의존성/병렬화 |
| 변경 비용 | 매우 높음 | 낮음 (재계획 가능) |

---

## Stage 흐름

```
plan-build (§4 phase orchestrator)
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

§2 feature spec의 actor list를 기반으로 각 feature를 actor별 implementation track으로 분리한다.

예시 (`signup-email-password` feature):
```
Track A: frontend-actor (Next.js)
  - signup form component
  - validation feedback UI
  - success redirect flow

Track B: backend-actor (Go auth-service)
  - credentials validation endpoint
  - password hashing (bcrypt)
  - JWT issuance

Track C: 3rd-party-actor (SendGrid)
  - verification email template
  - webhook handler for click confirmation
```

### Stage 2: Actor Track → Task List

각 actor track을 ordered task list로 분해한다.

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

actor 내부 의존성과 actor 간 contract 의존성을 DAG로 표현한다.

```
frontend-1 (signup form) → frontend-2 (validation UI) ← backend-1 (API spec)
backend-1 (credentials endpoint) → backend-2 (JWT endpoint)
3rdparty-1 (email template) → backend-3 (webhook handler)
```

Critical path를 식별한다: 전체 feature의 완료를 block하는 task 체인.

### Stage 4: 병렬 실행 계획

actor track 간 독립성을 활용해 병렬 worker 분배 계획을 작성한다.

`dispatch-parallel-agents` skill과 연동 가능한 형식으로 작성:
```yaml
parallel_tracks:
  - track: frontend
    worker: agent-1
    tasks: [frontend-1, frontend-2, frontend-3]
  - track: backend
    worker: agent-2
    tasks: [backend-1, backend-2, backend-3]
  - track: 3rd-party-integration
    worker: agent-3
    tasks: [3rdparty-1, backend-3]
synchronization_points:
  - after: [backend-1]
    before: [frontend-2]  # API contract 확정 후 frontend 구현 가능
```

### Stage 5: Acceptance Test Plan

actor별 + cross-actor 완료 기준을 정의한다.

```yaml
per_actor:
  frontend: "signup form submit → success redirect 동작"
  backend: "POST /auth/signup → 201 Created + JWT 반환"
  3rd-party: "verification email 수신 → click → 200 OK webhook"
integration:
  cross_actor: "full signup flow: form submit → email verification → first login"
```

### Stage 6: 빌드 타임라인

의존성과 병렬도를 고려해 일정을 합성한다.

```
Week 1: Track A (frontend) + Track B (backend) 병렬 시작
Week 2: Track C (3rd-party) — backend API 확정 후 가능
Week 3: Integration testing (cross-actor)
```

### Stage 7: autoplan Review

`autoplan`을 invoke해 task plan을 4-mode review한다.

---

## 다음 phase

- `/buddy:build-feature` — §5 Development (권장)
- `/buddy:dispatch-parallel-agents` — 병렬 agent 분배를 즉시 시작할 때

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §§4, §3.1
