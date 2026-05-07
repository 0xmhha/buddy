# decompose-feature-to-actor-tracks — feature → actor 별 implementation track 분해

§4 Implementation Plan 의 첫 단계. §3 design 산출물 (tech stack / data model / API contract) 과 §2 feature spec (actor / use case / system boundary) 을 입력으로, feature 를 actor 별 implementation track (frontend / backend / 3rd-party / data 등) 으로 분해한다. 본 skill 의 산출물은 후속 5 stage (decompose-track-to-tasks / map-task-dependencies / plan-parallel-execution / define-acceptance-test-plan / estimate-build-timeline) 의 입력이고, 궁극적으로 §5 build-feature 의 actor-별 worker 분배 입력이 된다. **Q8=(a) cascade 의 §4 진입점**.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 산출물에 다음이 발견되면 §5 로 돌아가 보강한다:

- **Track = system component 1:1 매핑** — track 은 implementation 책임 단위. system component 와 1:1 일 수도 있지만 다를 수 있음 (예: shared library track, ops track).
- **Cross-track contract 누락** — track 간 interface (API / event / shared schema) 명시 안 하면 후속 task DAG 의 cross-actor edge 가 무근거.
- **Independent vs sequential 판단 생략** — 모든 track 을 parallel safe 로 가정. integration risk 누수.
- **Track 의 책임이 카테고리 수준** — "frontend 책임은 UI" 같이 vague. concrete sub-system / data / 외부 의존 명시 필요.
- **Worker capability 무관심** — track 분배가 후속 plan-parallel-execution 의 입력이라 capability 매핑 가능한 단위로 분해.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

§3 design 의 system boundary / actor 매핑을 implementation 의 actor track 단위로 변환. track 은 implementation 책임의 atomic unit (1 worker 가 own 가능, cross-track 은 contract 로만 통신). 잘못된 분해는 후속 모든 §4 stage 에 cascade.

## 2. 사용 시점 (When to invoke)

- §3 design 산출물 받은 직후 (chain 권장)
- §2 feature spec 의 actor 변경 시 재분해
- cross-actor contract 변경으로 track 경계 재평가
- 신규 worker pool (예: ML engineer 추가) 도입 시 track 재배정

## 3. 입력 (Inputs)

### 필수
- §2 feature spec (actor list + per-actor use cases + system boundary)
- §3 system topology (어느 component 가 어느 actor 의 system boundary 에 속하는지)
- §3 tech stack ADR (각 component 의 stack 결정)
- §3 API contract (cross-actor interface)

### 선택
- 기존 codebase 구조 (brownfield 의 경우 track 경계 제약)
- 가용 worker pool composition (frontend / backend / DevOps 비율)
- 운영 history (이전 비슷 feature 의 track 분배 결과)

### 입력이 부족할 때 forcing question
- "이 feature 의 actor 가 frontend / backend / 3rd-party / data 외 또 있나? (예: ML model serving track, streaming pipeline track, ops automation track)"
- "cross-actor contract 가 sync API 만인가, async event 도 있나? event-driven 부분 있으면 track 분해가 달라진다."
- "이 track 의 worker 가 다른 track 도 동시에 own 가능한가, 전담인가? worker capability 매트릭스 입력에 영향."

## 4. 핵심 원칙 (Principles + Posture)

이 skill 의 운영 posture:

- **입장 취함, hedge 금지** — track 경계는 단호한 결정. "frontend 와 backend 사이 어딘가" 거부. 어느 쪽 track 에 속하는지 명시.
- **사용자 입력을 challenge** — 사용자가 "frontend, backend" 만 언급하면 "data track 은? ops track 은? 3rd-party 통합 별 track 분리는?" push.
- **Specificity 강제** — "shared utility track" 거부. 어떤 sub-system, 어떤 data, 어떤 외부 의존 명시.

도메인 원칙:

1. **Track = ownership atomic unit** — 1 worker / 1 small team 이 own 가능한 크기. 너무 크면 sub-track 분할.
2. **Cross-track contract = API 경계** — track 간 통신은 sync API / async event / shared DB / message queue 중 명시된 것만. ad-hoc shared state 금지.
3. **Independent first, sequential when forced** — track 간 의존성은 contract 가 정의된 시점부터 parallel-safe. contract 미정 시 sequential.
4. **Worker capability fit** — track 의 stack / 영역이 실제 보유 worker capability 와 매핑 가능한지 검증.
5. **Integration risk identification** — track 간 hand-off / merge 지점에서 발생 가능한 risk 미리 명시.

## 5. 단계 (Phases)

### Phase 1. Actor → track 매핑

§2 feature spec 의 actor 가 어느 track 에 속하는지:

| Actor | System boundary | Track | Sub-system |
|-------|-----------------|-------|------------|
| user (frontend) | SPA | frontend-spa-track | <pages, components, state> |
| auth-system (backend) | auth service | auth-backend-track | <session, JWT, password hash> |
| email-verifier (3rd-party) | external SaaS | email-verification-track | <SendGrid integration, webhook handler> |
| ... | ... | ... | ... |

actor 가 여러 track 에 분산 가능 (예: ops actor → monitoring-track + alerting-track).

### Phase 2. Track 책임 분해

각 track 의:
- **Sub-system** — 어떤 module / service / library 가 이 track 안에서 own
- **Data ownership** — 어떤 schema / collection / cache key 를 own
- **External dependency** — 어떤 외부 SaaS / package / service 에 의존
- **Worker capability** — 어떤 stack / 영역 capability 필요 (예: React + TS / Go + Postgres / SendGrid SDK)

### Phase 3. Cross-track contract 식별

track 간 interface:

| From | To | Contract type | Definition |
|------|-----|---------------|------------|
| frontend-spa-track | auth-backend-track | sync REST API | POST /signup, GET /me |
| auth-backend-track | email-verification-track | async event | UserSignedUp event → email handler |
| ... | ... | ... | ... |

각 contract 는 §3 의 API contract / event schema 와 매핑 (없으면 §3 보강 필요 → 그 단계로 회귀).

### Phase 4. Independent vs sequential 판단

각 track 쌍에 대해:
- **Parallel safe**: contract 가 §3 에서 정의됨 → 양 track 동시 개발 가능
- **Partial sequential**: contract 의 일부만 정의됨 → 정의된 부분 parallel, 미정의 부분 sequential
- **Strict sequential**: contract 미정 또는 한쪽이 다른 쪽의 schema 에 hard-depend → 한쪽 완료 후 다른쪽 시작

이 판단은 후속 `map-task-dependencies` 와 `plan-parallel-execution` 의 입력.

### Phase 5. Track output 정리

각 track 의 deliverable:
- 어떤 file / module 이 track 결과로 생기는지
- 어느 PR scope 가 track 단위인지
- track 완료 = 어떤 acceptance test 통과 인지 (후속 `define-acceptance-test-plan` 의 입력)

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## decompose-feature-to-actor-tracks Output — <feature name>

### Summary
<3 줄: track 수 / 가장 큰 cross-track risk / parallel 가능 batch 수>

### Track Table
| Track ID | Owner Actor(s) | System Boundary | Sub-systems | Data Ownership | External Deps | Worker Capability |
|----------|----------------|-----------------|-------------|----------------|---------------|-------------------|
| ... | ... | ... | ... | ... | ... | ... |

(at minimum 3 tracks for non-trivial feature)

### Cross-Track Contracts
| From | To | Type | Definition | Defined in §3? |
|------|-----|------|------------|----------------|
| ... | ... | sync REST / async event / shared DB / queue | ... | yes/no (no면 §3 보강 필요) |

### Independence Matrix
| Track A × Track B | Status | Reason |
|-------------------|--------|--------|
| frontend × auth-backend | parallel safe | API contract defined |
| auth-backend × email-verifier | partial sequential | event schema 일부 미정 |
| ... | ... | ... |

### Track Outputs
| Track | Deliverable | PR scope | Acceptance Hint |
|-------|-------------|----------|-----------------|
| ... | ... | ... | ... |

### Cascade to Next Skills
- **decompose-track-to-tasks**: 각 track 의 sub-system 을 ordered task list 로 분해
- **map-task-dependencies**: cross-track contract 가 cross-actor dependency edge 의 출처
- **plan-parallel-execution**: Independence Matrix 가 parallel batch 식별의 입력

### Next Step
<구체 action — 1줄: 예 "decompose-track-to-tasks 호출해 각 track 을 atomic task 로 분해">
```

## 7. Cross-phase cascade

- **§4 decompose-track-to-tasks**: track table 이 입력
- **§4 map-task-dependencies**: cross-track contract 가 cross-actor edge 의 출처
- **§4 plan-parallel-execution**: Independence Matrix 가 parallel batch 의 입력
- **§4 define-acceptance-test-plan**: track output (acceptance hint) 이 test plan 의 입력
- **§5 build-feature**: track 단위가 actor-track worker 배정 단위
- **§6 verify-quality**: track 별 acceptance 가 quality gate 의 일부

## 8. 다음 skill (next in stage flow)

- `decompose-track-to-tasks` — 본 skill 의 직접 후속
- 권장 chain: `decompose-feature-to-actor-tracks → decompose-track-to-tasks → map-task-dependencies → plan-parallel-execution → define-acceptance-test-plan → estimate-build-timeline`

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `decompose-track-to-tasks`** — 본 skill 은 track 단위 (큰 ownership unit). 다음 skill 은 task 단위 (atomic PR scope). 본 skill 이 먼저.
- **vs `map-feature-dependencies`** — map-feature-dependencies 는 §2 의 feature 간 선후 (다른 feature 와의 관계). 본 skill 은 §4 의 feature 내부 actor track 분해. 다른 layer.
- **vs `compose-feature-from-use-cases`** — compose-feature-from-use-cases 는 §2 의 use case → feature 합성. 본 skill 은 §4 의 feature → track 분해. 반대 방향.
- **vs `dispatch-parallel-agents`** — dispatch-parallel-agents 는 §5 build 의 worker 디스패치 메커니즘. 본 skill 은 §4 의 plan. 본 skill 이 먼저, dispatch-parallel-agents 가 plan 을 받아 실행.

## 10. 중요 규칙

- **Read-only on production** — code 수정 안 함. 분해와 plan 만 산출.
- **Track = atomic ownership unit** — 너무 크면 분할, 너무 작으면 합치기.
- **Cross-track contract 는 §3 에서 정의된 것만 사용** — 미정 contract 발견 시 §3 보강 강제.
- **Worker capability 매핑 가능성 검증** — track 이 실제 보유 worker 로 실행 가능한지.
- **Integration risk 명시 의무** — track 간 hand-off / merge 지점의 risk 항목.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 6 출력 섹션 (Summary / Track Table / Cross-Track Contracts / Independence Matrix / Track Outputs / Cascade / Next Step) 모두 채워짐
- [ ] Track Table 이 ≥ 3 row, 각 track 에 7 column 모두 채워짐
- [ ] Cross-Track Contracts 의 모든 contract 가 §3 에서 정의됨 ("no" 가 있으면 §3 보강 후 재실행)
- [ ] Independence Matrix 가 모든 track 쌍 평가
- [ ] Track Outputs 의 acceptance hint 가 §4 define-acceptance-test-plan 입력으로 작동 가능
- [ ] §4 posture 적용 — hedge 표현 없음, "shared utility" 류 카테고리 답변 없음
- [ ] §0 anti-pattern 들이 산출물에 등장하지 않음
- [ ] Next Step 에 구체 후속 skill 호출 라인 명시

하나라도 no 면 해당 phase 로 돌아가 보강 후 재검증.
