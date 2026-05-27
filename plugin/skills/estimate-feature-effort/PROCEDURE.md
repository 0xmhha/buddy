# Estimate Feature Effort — T-shirt sizing + ideal-h + uncertainty range

## 1. 목적

`define-feature-spec` 산출의 각 feature 에 *작업 effort* 를 추정한다. **T-shirt sizing** (XS/S/M/L/XL) 으로 *상대적 크기* 를 잡고, **ideal-h** (방해 없는 집중 시간) 로 *절대 시간* 변환, **uncertainty range** (best / expected / p90 / worst) 로 추정 신뢰도를 명시.

`map-feature-dependencies` + `score-feature-priority` 의 입력. `estimate-build-timeline` (§4) 의 *task 단위 estimation* 보다 *상위 layer* — feature 단위.

## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Feature spec 또는 feature 설명 | ✅ | artifact / knowledge | `define-feature-spec` 산출물 또는 사용자 설명 | "effort를 추정할 feature를 설명해 주세요." |
| 기술 스택 정보 | 선택 | knowledge | 사용자 도메인 지식 또는 `define-tech-stack` 산출물 | 없으면 범용 추정 (기술 보정 없음) |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Feature effort estimate (T-shirt + ideal-h + uncertainty range) | artifact | structured YAML | `score-feature-priority`, `map-feature-dependencies`, `estimate-build-timeline` |

## 2. 사용 시점

- `define-feature-spec` 산출 후 `score-feature-priority` 직전 — RICE 의 effort 항목 입력
- `map-feature-dependencies` 후 critical path 식별 — 큰 feature 가 path 위에 있는지 확인
- 분기 / 스프린트 plan 시 — feature backlog 우선순위 결정
- pivot 결정 시점 — 큰 feature 가 *진짜 가치* 대비 비싼지 검증
- 외주 / 채용 결정 시 — 인력 규모 추정 입력

## 3. 입력

### 필수
- `define-feature-spec` 의 feature list — 각 feature 의 acceptance criteria + technical scope
- 팀 capability (인원 / 스택 친숙도)
- 기존 유사 feature 의 *실제 소요 시간* (있으면 — historical baseline)

### 선택
- `score-feature-priority` 의 가설 RICE score (effort 컬럼 보강용)
- 의존 / blocking feature 정보

## 4. Stage 흐름

### Stage 1: T-shirt sizing

각 feature 에 5 단계 size:

| size | ideal-h 환산 (가이드) | 특징 |
|------|----------------|------|
| XS | < 4h | 단일 file 변경 / config flip |
| S | 4~16h | 단일 component / 1 endpoint |
| M | 16~40h | 1~2 component + integration |
| L | 40~120h | cross-component + DB schema |
| XL | > 120h | 새 sub-system / 다중 actor |

**규칙**: XL 발견 시 *분해 강제* — `split-work-into-features` skill cascade 권장.

### Stage 2: ideal-h 변환 + multiplier

T-shirt → ideal-h 의 *raw* 변환은 위 가이드. 단 *현실 시간* 으로는 multiplier 적용:

| multiplier | 사유 |
|----------|------|
| 1.0× | 신규 작업, 명확한 spec |
| 1.5× | 기존 코드 *수정* (이해 + 변경 + 검증) |
| 2.0× | unfamiliar 영역 (팀 처음) |
| 2.5× | 외부 의존 (3rd-party API / 규제 / 다른 팀) |

→ *현실 시간* = ideal-h × multiplier.

### Stage 3: Uncertainty range — 4 point estimate

각 feature 에 4 점 추정 (PERT 변형):

| point | 정의 |
|-------|------|
| Best (p20) | 모든 게 잘 풀릴 때 |
| Expected (p50) | 평균 / 일반적 |
| P90 (commit) | 90% 확률로 끝나는 시점 |
| Worst | 알려진 최악 risk 발생 시 |

Uncertainty 폭 (= worst - best) 이 expected 의 *3 배 이상* 이면 *분해 강제*. 추정 정확도 부족 = 작업 정의 부족.

### Stage 4: Critical path 후보 식별

`map-feature-dependencies` 산출과 cross-reference. *큰 feature 가 critical path 에 있으면* 분해 / 병렬화 / 외부 의존 줄이기 검토.

### Stage 5: Calibration (선택)

이전 분기 / 스프린트 의 *추정 vs 실제* 비교. 팀 multiplier 보정. 본 skill 의 산출이 *진척에 따라 정확도 향상*.

## 5. 산출물 형식

```markdown
## Feature Effort Estimation — {제품 / 분기}

### 추정 표
| Feature | size | ideal-h | multiplier | best / p50 / p90 / worst |
|---------|------|---------|----------|----------------------|

### XL 발견 (분해 권장)
- {feature}: 사유 — split-work-into-features 호출 권장

### Uncertainty 큰 feature (분해 권장)
- {feature}: best=X, worst=Y (3× 이상)

### Critical path 위 큰 feature
- {feature} (size=L, p90=120h): mitigation 후보

### Calibration log (이전 분기 vs 실제)
- 평균 multiplier drift: {%}
```

## 6. 검증

- [ ] 모든 feature 에 size / ideal-h / 4-point 추정?
- [ ] XL 발견 시 분해 권장 명시?
- [ ] Uncertainty 폭 *3× 초과* feature 분해 권장 명시?
- [ ] Multiplier 적용 (raw ideal-h 가 아닌 *현실 시간* 산출)?
- [ ] 이전 calibration 데이터 활용 (있을 시)?

## 7. 다음 phase

- `score-feature-priority` 의 effort 컬럼 입력
- `map-feature-dependencies` 의 critical path 분석 입력
- `estimate-build-timeline` (§4) 의 *task 단위 estimation* 으로 cascade — *feature → task* 분해 후 보다 정밀

## 8. 참조

- Agile Estimating and Planning (Mike Cohn) — T-shirt sizing + planning poker
- Software Estimation: Demystifying the Black Art (Steve McConnell) — uncertainty range + PERT
- 기존 buddy `estimate-build-timeline` (§4 stage) — task 단위 추정과의 layer 분리
