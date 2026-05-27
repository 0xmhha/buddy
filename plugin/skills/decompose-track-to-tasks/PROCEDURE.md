# decompose-track-to-tasks — actor track → ordered task list (atomic units)

`decompose-feature-to-actor-tracks` 의 산출물 (Track Table) 을 받아 각 track 을 ordered task list 로 분해한다. 각 task 는 atomic unit (single PR scope, verify 가능 acceptance, sized for 1 worker × short period). 너무 큰 task 는 sub-divide, 너무 작은 task 는 합치기. 본 skill 의 산출물은 `map-task-dependencies` 의 입력.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Actor track | ✅ | artifact / knowledge | `decompose-feature-to-actor-tracks` 산출물 또는 사용자 설명 | "분해할 actor track을 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Ordered task list (atomic, 1 PR scope, verifiable) | artifact | structured YAML | `map-task-dependencies` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 산출물에 발견되면 §5 로 돌아가 보강:

- **Task = vague verb phrase** — "Implement signup" 같이 scope 모호. concrete acceptance + diff size 추정 + verify 방법 필수.
- **Multi-PR task** — 1 task 가 multiple PR 로 깨질 크기. 발견 시 sub-task 로 분할.
- **Acceptance criteria 누락** — task 완료를 verify 못 하면 build-feature 에서 incomplete state 누수.
- **Internal dependency 누락** — 같은 track 내 task 간 선후가 무근거 → map-task-dependencies 가 invalid.
- **Diff size 무추정** — 일정 / capacity planning 입력 누락. estimate-build-timeline 정확도 무너짐.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

track 단위의 큰 ownership 을 implementation 가능한 atomic task 로 분해. 각 task 는 1 PR scope 안에 들어가야 후속 dispatch (병렬 execute) 가 가능. 잘못된 task 분해는 critical path 추정 오차로 cascade.

## 2. 사용 시점 (When to invoke)

- `decompose-feature-to-actor-tracks` 산출물 받은 직후 (chain 권장)
- track 의 sub-system 추가 / 변경 시 해당 track 의 task 재분해
- sprint 입력 작성 시 (1~2 weeks horizon)
- task 가 PR scope 초과 발견 시 retroactive 분해

## 3. 입력 (Inputs)

### 필수
- Track Table (`decompose-feature-to-actor-tracks` 산출물)
- 각 track 의 Sub-systems / Data ownership / External dependencies
- Track output 의 deliverable (acceptance hint)

### 선택
- diff size threshold (default: small ≤ 100 LoC, medium ≤ 500, large > 500 = 분할 강제)
- task naming convention (예: `verb + noun + scope`)
- 기존 codebase 의 file structure (file-level task scope 가능)

### 입력이 부족할 때 forcing question
- "이 track 의 deliverable 이 단일 endpoint / page / module 인가, 다수인가? 각각 별 task?"
- "이 task 의 acceptance 가 manual 검증인가, automated test 인가? automated 면 어느 test 가 통과해야 함?"
- "task 의 diff size 추정 근거? 비슷한 과거 작업의 LoC 또는 file 수 참고?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — task 분해는 단호. "이 정도면 task 1개" 가 아니라 "small / medium / large 분류 → large 면 분할" 명시.
- **사용자 입력을 challenge** — "auth flow 구현" 같은 vague task 거부. "endpoint 추가? schema 변경? token 발급 logic? 각각 다른 task" push.
- **Specificity 강제** — diff size / file 영향 / acceptance 의 measurable 형태.

도메인 원칙:

1. **Atomic = 1 PR scope** — review 가능 크기 (≤ 500 LoC 권장).
2. **Acceptance criteria 는 verifiable** — manual test step 또는 automated assertion. 모호한 "looks good" 거부.
3. **Internal dependency 는 explicit edge** — 같은 track 안 task 간 선후 명시.
4. **Naming convention 일관** — verb + noun + scope ("add signup endpoint to auth-service").
5. **Diff size 추정 의무** — small / medium / large 분류 + LoC 또는 file 수 추정.

## 5. 단계 (Phases)

### Phase 1. Track 별 sub-system 분해

각 track 의 sub-system / module / file 단위로 분해:

| Track | Sub-system | Files / modules affected (예상) |
|-------|-----------|--------------------------------|
| frontend-spa-track | Signup form | components/SignupForm.tsx, hooks/useSignup.ts, pages/signup.tsx |
| auth-backend-track | Credentials endpoint | services/auth/handler.go, services/auth/password.go |
| ... | ... | ... |

### Phase 2. Sub-system → task list

각 sub-system 을 atomic task 로 분해. naming convention: `<verb> <noun> <scope>`.

| Task ID | Track | Name | Description | Diff size | LoC est. |
|---------|-------|------|-------------|-----------|----------|
| T-1.1 | frontend-spa-track | Add SignupForm component with validation | ... | small | ~80 |
| T-1.2 | frontend-spa-track | Add useSignup hook for state + submit | ... | small | ~60 |
| T-1.3 | frontend-spa-track | Add /signup page with redirect on success | ... | small | ~40 |
| T-2.1 | auth-backend-track | Add POST /signup endpoint with bcrypt hash | ... | medium | ~200 |
| ... | ... | ... | ... | ... | ... |

### Phase 3. Acceptance criteria 강제

각 task 에 verifiable acceptance:

| Task ID | Acceptance Criteria |
|---------|---------------------|
| T-1.1 | Component renders email + password inputs; client-side validation rejects invalid email; submit button disabled when invalid; jest unit test passes |
| T-1.2 | hook submits form data via fetch; loading / error / success states; mock test for each state |
| ... | ... |

### Phase 4. Internal dependency 추출

같은 track 안 task 간 선후:

| Task | Depends on (internal) | Reason |
|------|----------------------|--------|
| T-1.3 | T-1.1, T-1.2 | page 가 component + hook 사용 |
| T-2.1 (자체) | (none) | atomic |
| ... | ... | ... |

### Phase 5. Subdivide too-large tasks

Phase 2 의 large 분류 task 를 sub-task 로 분할. 분할 후 다시 phase 2-4 거쳐 atomic 인지 검증.

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## decompose-track-to-tasks Output — <feature name>

### Summary
<3 줄: total task 수 / 가장 큰 track / 가장 큰 single task LoC est>

### Per-Track Task Lists
#### Track: <track-id>
| Task ID | Name | Sub-system | Diff size | LoC est | Internal Dependencies |
|---------|------|-----------|-----------|---------|----------------------|
| ... | ... | ... | ... | ... | ... |

(repeat per track)

### Acceptance Criteria Table
| Task ID | Acceptance |
|---------|------------|
| ... | ... |

### Subdivisions
<원래 large 였다 분할된 task 의 before / after — 분할 결과 atomic 검증>

### Cascade to Next Skills
- **map-task-dependencies**: per-track task list + internal edges + cross-track contract (Phase 1 output) 통합해 DAG 작성
- **estimate-build-timeline**: diff size + LoC est 가 task duration 추정의 입력

### Next Step
<구체 action — 1줄: 예 "map-task-dependencies 호출해 cross-track edges 추가">
```

## 7. Cross-phase cascade

- **§4 map-task-dependencies**: per-track task + internal dependency 를 DAG 의 node + intra-track edge 로 사용
- **§4 plan-parallel-execution**: parallel-safe task batch 식별 시 task list 사용
- **§4 estimate-build-timeline**: diff size / LoC est 가 task duration 의 입력
- **§4 define-acceptance-test-plan**: acceptance criteria 가 test plan 의 atomic 단위
- **§5 build-feature**: task 가 worker 배정 단위
- **§7 ship-release**: task 가 PR 단위, atomic commit / merge gate 의 단위

## 8. 다음 skill (next in stage flow)

- `map-task-dependencies` — 본 skill 의 직접 후속

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `decompose-feature-to-actor-tracks`** — 그것은 track 단위, 본 skill 은 task 단위. 본 skill 이 후속.
- **vs `triage-work-items`** — triage-work-items 는 §2 feature backlog 의 우선순위. 본 skill 은 §4 의 task 분해. 다른 layer.
- **vs `split-work-into-features`** — split-work-into-features 는 §2 의 PRD → feature 분해 (broader). 본 skill 은 feature → task 분해 (narrower).
- **vs `score-feature-priority`** — score-feature-priority 는 feature 단위 priority. task 단위 priority 는 다음 skill (`map-task-dependencies` + critical path).

## 10. 중요 규칙

- **Atomic 보장** — 1 PR scope 초과 시 분할.
- **Acceptance verifiable** — manual step 또는 automated test 명시.
- **Internal dependency edge 누락 금지** — 후속 DAG 의 input 무결성.
- **Diff size 추정 의무** — estimate-build-timeline 입력.
- **Naming convention 일관** — verb + noun + scope.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 5 출력 섹션 (Summary / Per-Track Task Lists / Acceptance Criteria / Subdivisions / Cascade / Next Step) 모두 채워짐
- [ ] 모든 task 가 atomic (large 분류 없음 또는 분할됨)
- [ ] 모든 task 에 acceptance criteria + diff size + LoC est 명시
- [ ] Internal dependency edge 가 같은 track 내 모든 의존 task 쌍에 명시
- [ ] §4 posture 적용 — vague task name 없음
- [ ] §0 anti-pattern 들이 산출물에 등장하지 않음

하나라도 no 면 해당 phase 로 돌아가 보강.
