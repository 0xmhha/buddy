# plan-parallel-execution — worker batch + sync points

`map-task-dependencies` 의 산출물 (DAG + critical path + parallel-safe levels) 과 가용 worker pool (인간 + AI agent) 을 받아 worker 별 batch 배정 + 동기화 지점 (hand-off / merge gate / integration test) 을 plan 한다. Phase 1 의 `dispatch-parallel-agents` 와 결합 가능. 본 skill 의 산출물은 §5 build-feature 의 worker dispatch 입력.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Task DAG | ✅ | artifact | `map-task-dependencies` 산출물 | 먼저 `/buddy:map-task-dependencies` 를 실행하세요 |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Parallel execution plan (worker batch + sync points) | artifact | structured YAML | `dispatch-parallel-agents`, `build-feature` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다:

- **Worker capability 무시 배정** — frontend task 를 backend-only worker 에 배정. capability matrix 검증 필수.
- **동기화 지점 누락** — cross-track hand-off / merge gate 명시 안 하면 integration 시점에 chaos.
- **Critical path 무시** — load balancing 만 보고 critical path 의 worker 부족 시 일정 무너짐.
- **Bottleneck 식별 안 함** — single worker 에 strict-sequential task 가 몰리면 병목.
- **Batch 크기 polynomial growth** — parallel-safe level 의 task 수가 worker 수보다 압도적이면 실제로 parallel 안 됨.

§5 모든 phase 누락 없이 수행하라.

## 1. 목적

DAG 와 worker pool 을 schedule 로 변환. 각 worker 가 어느 task 를 어느 batch 에서 실행하는지, hand-off 지점이 어디인지 명시. parallel agent dispatch 의 input plan.

## 2. 사용 시점 (When to invoke)

- `map-task-dependencies` 산출물 받은 직후 (chain 권장)
- worker pool 변경 (augment / reduce / capability shift) 시 재배정
- bottleneck 발견으로 plan 재조정 필요 시
- pre-sprint kickoff / parallel agent dispatch 직전

## 3. 입력 (Inputs)

### 필수
- DAG + critical path + parallel-safe levels (`map-task-dependencies` 산출물)
- 가용 worker pool (인간 worker 수 + capability + AI agent 가능 수)
- worker capability matrix (각 worker 가 어느 stack / actor / task type 가능)

### 선택
- worker 일정 (휴가, 다른 sprint 와 겹침)
- AI agent 비용 / quota constraint
- 동기화 지점의 acceptable delay (hand-off lag tolerance)

### 입력이 부족할 때 forcing question
- "AI agent 를 worker 로 카운트할 때 review / hand-off 시간 포함했나?"
- "worker capability 가 strict 인가, soft 인가? 'frontend 만 가능' 인가, 'frontend 우선 + backend 가능' 인가?"
- "critical path 의 worker 가 1명뿐이면 병목 — 이 가정 받아들이는가, capability cross-train 으로 해결?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — worker 배정은 단호. "task X 는 worker A 또는 B" 거부, primary + backup 둘 다 명시.
- **사용자 입력을 challenge** — "5 worker 로 충분" 발화에 "critical path 의 worker 부족 검증했나? bottleneck 어디?" push.
- **Specificity 강제** — "1 sprint" 거부. "level 0 batch: 3 worker × 5 task = 1.5 day", "level 1 batch: 4 worker × 6 task = 2 day".

도메인 원칙:

1. **Capability fit 우선, load balance 차순** — 적합한 worker 부재 시 추가 capability 필요.
2. **Critical path 가 schedule lower bound** — 그 길이보다 짧은 일정 불가능.
3. **Sync point 는 explicit gate** — hand-off, merge, integration test 시점이 schedule 의 milestone.
4. **Bottleneck 식별 의무** — single worker / single-threaded 발견 시 mitigation 명시.
5. **Worker overload 방지** — 1 worker 가 동시 다수 task own 금지 (context switch cost 큼).

## 5. 단계 (Phases)

### Phase 1. Worker capability matrix

각 worker 의 capability:

| Worker | Stack capability | Actor expertise | AI / human | Availability |
|--------|------------------|-----------------|------------|--------------|
| W-1 (human) | TS / React / Next.js | frontend (primary), backend (soft) | human | full sprint |
| W-2 (human) | Go / Postgres | backend (primary), data (primary) | human | partial (50%) |
| W-3 (AI agent) | TS / Go (any with prompt) | flexible | AI (general-purpose subagent) | unlimited (modulo cost) |
| ... | ... | ... | ... | ... |

### Phase 2. 배정 알고리즘

- **Capability fit** 1순위: task 의 stack / actor 와 worker capability 매핑
- **Critical path 우선** 2순위: critical path task 는 high-availability worker 에 배정
- **Load balancing** 3순위: 다른 worker 들에 task 분산
- **Backup worker** 명시: primary 의 unavailability 시 picking up 가능 worker

### Phase 3. 동기화 지점 식별

- **Hand-off point**: cross-track edge 가 task 사이에 있는 시점 (예: backend T-2.3 완료 후 frontend T-1.4 시작)
- **Merge gate**: PR merge 가 다른 task 의 시작 조건
- **Integration test**: cross-actor flow 검증 시점 (define-acceptance-test-plan 의 cross-actor flow test 와 매핑)

각 sync point 의 acceptable delay 명시.

### Phase 4. Batch schedule

parallel-safe levels 별 batch:

| Batch | Tasks | Workers assigned | Est. duration |
|-------|-------|------------------|---------------|
| 0 (start) | T-1.1, T-2.1, T-3.0 | W-1 (T-1.1), W-2 (T-2.1), W-3 (T-3.0) | 1 day |
| 1 | T-1.2, T-2.2 | W-1 (T-1.2), W-2 (T-2.2) | 1.5 day |
| sync 0→1 | hand-off T-2.3 → T-1.4 | (gate, not work) | within 0.5 day |
| 2 | ... | ... | ... |

### Phase 5. Bottleneck 분석 + Mitigation

- **단일 worker bottleneck**: critical path 가 W-2 만 처리 가능 → cross-train 또는 W-2 augment
- **Capability gap**: 어떤 task 가 가용 worker capability 와 매핑 안 되는 발견 → 추가 worker / external 합류
- **AI agent 한계**: AI 가 처리 못하는 task type (예: 회의 / stakeholder communication) 명시

## 6. 산출물 형식 (Output format)

> structured 출력, prose 변환 금지.

```markdown
## plan-parallel-execution Output — <feature name>

### Summary
<3 줄: total worker / total batch / critical path batch 위치 / 가장 큰 bottleneck>

### Worker Capability Matrix
| Worker | Stack | Actor | Type | Availability |
|--------|-------|-------|------|--------------|
| ... | ... | ... | ... | ... |

### Batch Schedule
| Batch | Tasks | Workers | Est. Duration | Notes |
|-------|-------|---------|---------------|-------|
| ... | ... | ... | ... | ... |

### Sync Points
| Type | Between | Acceptable Delay | Validation |
|------|---------|------------------|------------|
| hand-off | T-2.3 → T-1.4 | 0.5 day | schema check |
| merge gate | T-X PR merge before T-Y starts | same day | CI green |
| integration test | after batch N | 1 day | cross-actor flow test pass |

### Worker Assignment Map
| Worker | Tasks (in order) | Total est. duration |
|--------|------------------|----------------------|
| W-1 | T-1.1, T-1.2, T-1.4 | 3 day |
| W-2 | T-2.1, T-2.2, T-2.3 | 4 day |
| ... | ... | ... |

### Bottlenecks + Mitigation
| # | Bottleneck | Severity | Mitigation |
|---|-----------|----------|-----------|
| 1 | W-2 on critical path 4 day | high | cross-train W-3 on backend OR add backup worker |
| ... | ... | ... | ... |

### Cascade to Next Skills
- **estimate-build-timeline**: batch duration + sync delay 가 timeline 합성 입력
- **§5 dispatch-parallel-agents**: AI worker 의 task list (Worker Assignment Map 의 W-3 등) 가 dispatch 입력

### Next Step
<구체 action — 1줄: 예 "estimate-build-timeline 호출해 calendar 일정 합성">
```

## 7. Cross-phase cascade

- **§4 estimate-build-timeline**: batch duration + sync delay 가 timeline 합성 핵심 입력
- **§4 define-acceptance-test-plan**: integration test sync point 가 test schedule 입력
- **§5 build-feature**: Worker Assignment Map 이 worker dispatch 의 plan
- **§5 dispatch-parallel-agents**: AI worker (W-3 등) 의 task assignment 가 dispatch 호출 schema
- **§7 ship-release**: critical path 의 last batch 가 release readiness 시점

## 8. 다음 skill (next in stage flow)

- `define-acceptance-test-plan` — sync point 의 integration test 정의 (병행 가능)
- `estimate-build-timeline` — batch duration → calendar 변환

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `map-task-dependencies`** — 그것은 DAG 구조, 본 skill 은 그 위의 worker schedule. 본 skill 이 후속.
- **vs `dispatch-parallel-agents`** — dispatch-parallel-agents 는 §5 build 의 실제 dispatch 메커니즘. 본 skill 은 §4 의 plan. 본 skill 이 먼저, dispatch-parallel-agents 가 plan 받아 실행.
- **vs `estimate-build-timeline`** — estimate-build-timeline 은 calendar 변환 (날짜 단위), 본 skill 은 batch + worker plan. 본 skill 이 먼저.

## 10. 중요 규칙

- **Capability fit 우선** — load balancing 은 차순.
- **Critical path 는 lower bound** — 그보다 짧은 schedule 불가.
- **Sync point explicit gate** — hand-off / merge / integration 모두 명시.
- **Bottleneck 명시 의무** — mitigation 포함.
- **Worker overload 방지** — 1 worker × 동시 1 active task 권장.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 6 출력 섹션 (Summary / Worker Capability / Batch Schedule / Sync Points / Worker Assignment / Bottlenecks / Cascade / Next Step) 모두 채워짐
- [ ] Worker Capability Matrix 가 모든 worker 의 stack + actor + availability 명시
- [ ] Batch Schedule 의 batch 수 ≥ map-task-dependencies 의 parallel-safe levels 수
- [ ] Sync Points 가 cross-track edge 마다 hand-off 또는 merge gate 로 매핑
- [ ] Worker Assignment Map 의 모든 task 가 capability fit 한 worker 에 배정
- [ ] Bottlenecks 식별 + mitigation 명시
- [ ] Critical path 가 batch schedule 의 longest worker chain 과 일치
- [ ] §4 posture 적용
- [ ] §0 anti-pattern 들이 등장하지 않음

하나라도 no 면 해당 phase 로 돌아가 보강.
