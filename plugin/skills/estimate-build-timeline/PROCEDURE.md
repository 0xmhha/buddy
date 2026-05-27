# estimate-build-timeline — critical path 기반 일정 + confidence interval + risk buffer

§4 의 마지막 stage. task DAG (`map-task-dependencies`) + worker batch (`plan-parallel-execution`) + per-task duration 추정 (LoC est 또는 historical) 을 받아 critical path 기반 calendar timeline 을 합성한다. confidence interval (best / expected / worst) + risk buffer + holiday / availability 반영. external commitment / sprint planning 의 input.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Task DAG + effort estimates | ✅ | artifact | `map-task-dependencies` + `estimate-feature-effort` 산출물 | 먼저 task 분해와 effort 추정을 실행하세요 |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Calendar timeline (critical path + risk buffer) | artifact | structured YAML | 프로젝트 관리 |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다:

- **Single-point estimate** — "5 day" 만 출력. CI 없으면 stakeholder 가 worst-case 무방비.
- **Ideal-day 가정** — meeting / interruption / context switch / debug 무시. real productivity 70% 정도 적용.
- **Risk buffer 0** — known unknown 비례 buffer 없으면 첫 risk 발생에서 일정 무너짐.
- **Holiday / weekend 무반영** — calendar 기준이 아닌 person-day 기준만 → external commitment 빗나감.
- **Critical path 만 보고 cumulative parallel slack 무시** — 실제 일정은 critical path + sync delay + integration test buffer 합계.

§5 모든 phase 누락 없이 수행하라.

## 1. 목적

batch schedule 을 calendar timeline 으로 변환. external commitment 가능한 정확도 (best / expected / worst CI) + risk buffer 정량화. 일정 slip 의 most common cause 를 plan 단계에서 노출.

## 2. 사용 시점 (When to invoke)

- `plan-parallel-execution` 산출물 받은 직후 (chain 권장 — Phase 2 의 종착점)
- sprint planning / external stakeholder commitment 직전
- mid-sprint 일정 점검 / re-baseline
- risk 발생 후 timeline 재계산

## 3. 입력 (Inputs)

### 필수
- task DAG + critical path (`map-task-dependencies`)
- worker batch + sync points (`plan-parallel-execution`)
- per-task duration estimate (LoC est 또는 historical)
- worker availability (full-time / partial / vacation)
- start date

### 선택
- holiday calendar (region 별)
- historical productivity ratio (실제 / ideal — default 0.7)
- known risks + likelihood (이미 plan-parallel-execution Bottlenecks 식별 시)

### 입력이 부족할 때 forcing question
- "duration 추정 근거? historical 비슷 task 인가, t-shirt sizing (S=1d, M=3d, L=7d) 인가? 둘 다 아니면 estimate 신뢰도 낮음."
- "productivity ratio 가 0.7 (default) 인가, 다른 값? meeting 비율, debug 비율 영향."
- "risk buffer 비율 — 알려진 unknown 정도 (0%: 친숙 영역, 30%+: 신규 영역)."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — timeline 은 단호한 commit + CI. "1~3주" 거부, "best 5d, expected 8d, worst 13d, p90 commit 12d" 명시.
- **사용자 입력을 challenge** — "5 day 면 되겠지" 발화에 "critical path 가 어느 정도? sync delay 포함했나? holiday 며칠?" push.
- **Specificity 강제** — calendar date 단위, person-day 와 elapsed-day 구분.

도메인 원칙:

1. **Critical path = lower bound** — 그보다 짧은 timeline 불가능.
2. **Productivity ratio 적용** — ideal-day → real-day 변환. default 0.7.
3. **Risk buffer 명시** — known unknown 비례 (15~30%).
4. **Calendar 변환** — holiday / weekend / availability 반영.
5. **CI 표기** — best / expected / worst 또는 p50 / p90 / p99.

## 5. 단계 (Phases)

### Phase 1. Per-task duration estimate (ideal-day)

각 task 의 duration:

| Task ID | LoC est | T-shirt | Ideal duration (h) | Source |
|---------|---------|---------|---------------------|--------|
| T-1.1 | 80 | S | 4h | LoC ratio (20 LoC/h) |
| T-2.1 | 200 | M | 12h | T-shirt mapping |
| ... | ... | ... | ... | ... |

LoC ratio default: 20 LoC/h (high uncertainty, depends on stack/complexity).

### Phase 2. Productivity ratio + real-day 변환

```
real-duration = ideal-duration / productivity-ratio
default: 0.7 → real = ideal / 0.7 = ideal × 1.43
```

각 task 의 real-h 계산.

### Phase 3. Critical path duration + parallel slack

- critical path 의 real-h 합산
- parallel batch 별 max(worker real-h) 가 batch 의 real duration
- batch 간 sync delay 추가

### Phase 4. Calendar 변환

```
person-day = real-h / 8 (h per workday)
elapsed-day = person-day × (1 + weekend/holiday ratio)
calendar duration = elapsed-day from start-date, 영업일 + holiday 반영
```

worker availability (partial / vacation) 도 차감.

### Phase 5. Risk buffer + CI

- **Best case**: 모든 가정 성립 → critical path 그대로
- **Expected**: critical path × (1 + buffer 15%)
- **Worst case**: critical path × (1 + buffer 30~50%) — known risks 반영
- **p50**: expected
- **p90**: expected + (worst - expected) × 0.7
- **p99**: worst

stakeholder commit 권장: p90 (실제 slip 가능성 10% 이하).

## 6. 산출물 형식 (Output format)

> structured 출력, prose 변환 금지.

```markdown
## estimate-build-timeline Output — <feature name>

### Summary
<3 줄: critical path duration / p50 / p90 commit / 가장 큰 risk>

### Per-Task Duration
| Task ID | LoC est | T-shirt | Ideal h | Real h | Source |
|---------|---------|---------|---------|--------|--------|
| ... | ... | ... | ... | ... | ... |

### Critical Path Timeline
| Step | Task | Real h | Cumulative |
|------|------|--------|-----------|
| 1 | T-2.1 | 17 | 17 |
| 2 | T-2.3 | 8 | 25 |
| ... | ... | ... | ... |

Total critical path: <real h>

### Calendar Translation
| Metric | Value |
|--------|-------|
| Critical path real-h | ... |
| Person-day (8h) | ... |
| Productivity ratio | 0.7 |
| Weekend/holiday ratio | ... |
| Worker availability adj. | ... |
| Elapsed-day | ... |
| Start date | ... |
| Best-case end date | ... |
| Expected end date | ... |
| Worst-case end date | ... |

### Risk Buffer
| Buffer | % | Reason |
|--------|---|--------|
| Known unknown | 15% | average for familiar territory |
| Risk-specific | +N% | <plan-parallel-execution Bottlenecks 의 risk> |
| Total buffer | ... | ... |

### Confidence Interval
| Percentile | Date |
|-----------|------|
| Best (p10) | ... |
| Expected (p50) | ... |
| p90 (commit 권장) | ... |
| Worst (p99) | ... |

### Cascade
- §7 ship-release: p90 가 release readiness target date
- external commitment: stakeholder 에게 p90 또는 worst commit, expected 는 internal
- §5 build-feature: per-task real-h 가 worker 에게 visible 한 deadline

### Next Step
<구체 action — 1줄: 예 "stakeholder 에게 p90 commit + expected internal — sprint kickoff 진행">
```

## 7. Cross-phase cascade

- **§5 build-feature**: per-task real-h 가 worker visibility
- **§7 ship-release**: p90 / worst 가 release date commit
- **§8 iterate-product**: 실제 vs estimate gap 이 historical productivity ratio 보정 입력

## 8. 다음 skill (next in stage flow)

본 skill 은 §4 의 종착점. 후속:
- `autoplan` — §4 plan 전체 4-mode review
- §5 진입 — `build-feature` orchestrator

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `plan-parallel-execution`** — 그것은 batch + worker, 본 skill 은 calendar 변환. 본 skill 이 후속.
- **vs `score-feature-priority`** — score-feature-priority 는 §2 의 feature 선택. 본 skill 은 §4 의 선택된 feature 의 일정. 다른 layer.
- **vs `summarize-retro`** — summarize-retro 는 §8 의 retrospective. 본 skill 은 forward-looking estimate.

## 10. 중요 규칙

- **CI 표기 의무** — single point 거부.
- **Productivity ratio 적용 의무** — ideal vs real 구분.
- **Calendar 변환 의무** — person-day 가 아닌 elapsed-day commit.
- **Risk buffer 명시 의무** — 0% 거부.
- **Stakeholder commit 은 p90 권장** — internal 은 p50.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 7 출력 섹션 (Summary / Per-Task / Critical Path / Calendar / Risk Buffer / CI / Cascade / Next Step) 모두 채워짐
- [ ] Per-Task Duration 이 모든 task 포함, source 명시
- [ ] Critical Path Timeline 의 cumulative h 가 critical path 와 일치
- [ ] Calendar Translation 이 person-day → elapsed-day → calendar 변환 모두 표기
- [ ] Risk Buffer 가 0 보다 큼 (default 15% 미만 시 근거 필요)
- [ ] Confidence Interval 의 best / expected / p90 / worst 4 값 모두 명시
- [ ] §4 posture 적용
- [ ] §0 anti-pattern 들이 등장하지 않음

하나라도 no 면 해당 phase 로 돌아가 보강.
