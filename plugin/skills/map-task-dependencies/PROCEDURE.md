# map-task-dependencies — task DAG (internal + cross-actor edges) + critical path

`decompose-track-to-tasks` 의 산출물 (per-track task list + internal dependency edges) 과 `decompose-feature-to-actor-tracks` 의 Cross-Track Contracts 를 통합해 전체 task DAG 를 작성한다. cycle 감지, critical path 계산, parallel-safe 그룹 식별까지 강제. `plan-parallel-execution` 과 `estimate-build-timeline` 의 입력.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다:

- **Cycle 무검증** — DAG validation 생략 시 후속 schedule 이 deadlock.
- **Cross-actor edge 누락** — Cross-Track Contracts 의 contract dependency 가 task 단위 edge 로 변환 안 됨 → integration 시점에 schema mismatch.
- **Critical path 계산 생략** — 일정 추정의 필수 입력. 누락 시 estimate-build-timeline 무근거.
- **Parallel-safe 그룹 식별 안 함** — plan-parallel-execution 의 batch 결정 입력 누락.
- **Edge type 미분류** — strict (hard dependency) vs soft (preference) 구분 없으면 schedule 유연성 사라짐.

§5 모든 phase 누락 없이 수행하라.

## 1. 목적

per-track task list 를 단일 DAG 로 통합. internal (intra-track) + cross-actor (contract-based) edges 모두 명시. cycle 감지 + critical path 추출 + parallel-safe 그룹 식별까지 자동화.

## 2. 사용 시점 (When to invoke)

- `decompose-track-to-tasks` 산출물 받은 직후 (chain 권장)
- task 추가 / 삭제로 dependency 재계산 필요
- cycle 의심 또는 schedule slip 발생 시 root cause 분석
- new contract 추가로 cross-actor edge 변경 시

## 3. 입력 (Inputs)

### 필수
- per-track task list (`decompose-track-to-tasks` 산출물)
- internal dependency edges (같은 track 안 task 선후)
- Cross-Track Contracts (`decompose-feature-to-actor-tracks` 산출물)

### 선택
- 기존 codebase 의 historical pattern (어느 task 가 늘 후행되는지)
- 외부 dependency (3rd-party contract 가 ready 되는 시점)

### 입력이 부족할 때 forcing question
- "이 contract dependency 가 strict 인가, soft 인가? strict = 한쪽 완성 후 다른쪽 시작. soft = mock 으로 parallel 가능."
- "cross-actor edge 가 어느 task 에 attach? 'auth-backend 끝나면 frontend 시작' 이 아니라 'auth-backend T-2.3 (token issuance) 끝나면 frontend T-1.4 (login submit) 시작'."
- "circular dependency 의심 — A.task1 ↔ B.task2 발견 시 mock / interface 도입으로 해결?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — edge type 은 strict 또는 soft 중 단호한 분류. "관련 있을 것" 거부.
- **사용자 입력을 challenge** — "이 task 가 저 task 에 의존" 발화에 "어느 task 의 어느 산출물? interface 가 §3 에 정의됐나?" push.
- **Specificity 강제** — "frontend 가 backend 기다림" 거부. "frontend T-1.4 가 backend T-2.3 의 token issue 결과 schema 의존" 명시.

도메인 원칙:

1. **DAG must be acyclic** — cycle 감지 시 mock / interface 로 끊기.
2. **Edge type 분류 의무** — strict (hard) vs soft (mock-able) 구분.
3. **Critical path 는 schedule 의 lower bound** — 일정 추정의 출발점.
4. **Parallel-safe = same level + no cross-edge** — DAG 의 topological level 기반.
5. **Contract dependency = §3 에서 정의된 것만** — ad-hoc 의존 금지.

## 5. 단계 (Phases)

### Phase 1. Internal edges 수집

각 track 의 task list 에서 internal dependency:

| From task | To task | Type | Reason |
|-----------|---------|------|--------|
| T-1.1 | T-1.3 | strict | T-1.3 이 T-1.1 의 component 사용 |
| T-1.2 | T-1.3 | strict | T-1.3 이 T-1.2 의 hook 사용 |
| ... | ... | ... | ... |

### Phase 2. Cross-actor edges 수집

Cross-Track Contracts 를 task 단위 edge 로 변환:

| Contract | From track / task | To track / task | Type | Reason |
|----------|-------------------|------------------|------|--------|
| POST /signup | auth-backend / T-2.1 | frontend-spa / T-1.2 | strict | hook 이 endpoint schema 사용 |
| UserSignedUp event | auth-backend / T-2.4 | email-verification / T-3.1 | soft | event consumer 는 mock 으로 parallel 가능 |
| ... | ... | ... | ... | ... |

### Phase 3. Cycle 감지

DAG validation:

```
all tasks = nodes
all edges = internal + cross-actor

DFS / topological sort 시도 → cycle 발견 시 어느 edge 인지 명시
```

cycle 발견 시:
- soft edge 라면 mock interface 도입으로 끊기 권장
- strict 라면 task 분해 재설계 (한쪽 task 를 더 분할해 dependency 반대 방향 제거)

### Phase 4. Critical path 계산

DAG 의 longest dependency chain (LoC est 또는 duration 가중):

```
critical path: T-2.1 → T-2.3 → T-1.4 → T-1.5 → T-3.1
total duration: <합계>
slack 에 있는 tasks: <list>
```

### Phase 5. Parallel-safe 그룹 식별

topological levels 별 parallel-safe batch:

| Level | Tasks (parallel safe) | Blocked by |
|-------|----------------------|-----------|
| 0 | T-1.1, T-2.1, T-3.0 (independent starts) | (none) |
| 1 | T-1.2, T-2.2 | level 0 의 일부 |
| ... | ... | ... |

## 6. 산출물 형식 (Output format)

> structured 출력, prose 변환 금지.

```markdown
## map-task-dependencies Output — <feature name>

### Summary
<3 줄: total nodes / total edges / critical path duration / parallel batch 수>

### DAG (mermaid 또는 adjacency list)
\`\`\`mermaid
graph LR
  T-1.1 --> T-1.3
  T-1.2 --> T-1.3
  T-2.1 --> T-1.2
  ...
\`\`\`

### Internal Edges
| From | To | Type | Reason |
|------|-----|------|--------|
| ... | ... | ... | ... |

### Cross-Actor Edges
| Contract | From | To | Type | Reason |
|----------|------|-----|------|--------|
| ... | ... | ... | ... | ... |

### Cycle Check
| Status | Detail |
|--------|--------|
| pass / fail | <cycle 발견 시 어느 edge — fail 시 §5 Phase 3 으로 회귀> |

### Critical Path
- Path: <T-x → T-y → ... → T-z>
- Duration: <합계 LoC 또는 시간>
- Slack tasks: <list>

### Parallel-Safe Levels
| Level | Tasks |
|-------|-------|
| 0 | ... |
| 1 | ... |
| ... | ... |

### Cascade to Next Skills
- **plan-parallel-execution**: parallel-safe levels 가 batch 식별 입력
- **estimate-build-timeline**: critical path 가 timeline 계산 시작점

### Next Step
<구체 action — 1줄: 예 "plan-parallel-execution 호출해 worker 별 batch 배정">
```

## 7. Cross-phase cascade

- **§4 plan-parallel-execution**: parallel-safe levels + critical path 가 worker batch 입력
- **§4 estimate-build-timeline**: critical path duration 이 timeline lower bound
- **§4 define-acceptance-test-plan**: cross-actor edges 가 cross-actor flow test 의 trigger
- **§5 build-feature**: DAG 가 worker dispatch 순서
- **§7 ship-release**: critical path 가 deploy gate 의 schedule 입력

## 8. 다음 skill (next in stage flow)

- `plan-parallel-execution` — 본 skill 의 직접 후속

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `decompose-track-to-tasks`** — 그것은 task 분해, 본 skill 은 task 간 dependency. 본 skill 이 후속.
- **vs `map-feature-dependencies`** — map-feature-dependencies 는 §2 의 feature 간 DAG. 본 skill 은 §4 의 task 간 DAG. 다른 layer.
- **vs `plan-parallel-execution`** — 본 skill 은 DAG 구조, 다음 skill 은 그 위의 worker schedule. 본 skill 이 먼저.

## 10. 중요 규칙

- **Acyclic 보장** — cycle 발견 시 분해 재설계.
- **Edge type 분류 의무** — strict / soft.
- **Cross-actor edge 는 §3 contract 기반만** — ad-hoc 의존 금지.
- **Critical path 명시 의무** — 후속 일정 입력.
- **Parallel-safe levels 명시 의무** — 후속 batch 입력.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 7 출력 섹션 (Summary / DAG / Internal Edges / Cross-Actor Edges / Cycle Check / Critical Path / Parallel-Safe Levels / Cascade / Next Step) 모두 채워짐
- [ ] DAG 가 mermaid 또는 adjacency list 로 명시
- [ ] Cycle Check 가 pass — fail 시 산출물 invalid
- [ ] Cross-Actor Edges 가 모두 §3 contract 와 매핑됨 (없는 dependency 발견 시 §3 보강 강제)
- [ ] Critical Path 가 명시 (start node + end node + duration)
- [ ] Parallel-Safe Levels 가 ≥ 2 levels (단일 level 만이면 task 간 의존 분석 불충분)
- [ ] §4 posture 적용 — vague edge 없음
- [ ] §0 anti-pattern 들이 등장하지 않음

하나라도 no 면 해당 phase 로 돌아가 보강.
