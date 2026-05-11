# Analyze User Cohort — acquisition cohort retention curve + segment 분석

## 1. 목적

사용자를 *acquisition 시점 / channel / segment* 별로 cohort 묶어 **retention curve + LTV + churn** 측정. 평균만으로 가려진 *cohort 별 실제 패턴* 노출.

`analyze-feature-adoption` 의 power user 분석과 cascade — feature 차원 + cohort 차원 cross.

## 2. 사용 시점

- §8 iterate-product 의 *분기 retention 분석*
- 신규 acquisition channel 도입 후 — channel 별 quality 비교
- pricing tier 변경 후 — tier 별 retention 변화
- product 변경 (큰 release) 후 — pre / post cohort 비교
- churn rate spike 시 — 어느 cohort 가 영향인가

## 3. 입력

### 필수
- user 가입 데이터 (timestamp + acquisition channel + segment)
- activity 데이터 (각 user 의 후속 사용 timestamp)
- cohort 정의 (week / month / quarter)

### 선택
- pricing tier / segment / persona 정보
- 외부 channel attribution data

## 4. Stage 흐름

### Stage 1: Cohort 정의

| 차원 | 옵션 |
|------|-----|
| Time bucket | weekly / monthly / quarterly |
| Acquisition channel | organic / paid / referral / community |
| Segment | persona 별 |
| Pricing tier | free / paid / enterprise |
| Geographic | locale 별 |

→ 한 분석에 *2~3 차원 cross*. 너무 많으면 sample 작아짐.

### Stage 2: Retention curve

각 cohort 의 D1 / D7 / D30 / D90 retention:

```
       D1    D7    D30   D90   D180
Jan   100%  60%   45%   35%   30%
Feb   100%  65%   50%   40%   35%   ← 개선
Mar   100%  55%   38%   28%   22%   ← 악화
```

→ *flat curve* (long retention) = 가치 강 / *sharp drop* = 첫 경험 문제.

### Stage 3: LTV (Lifetime Value) 계산

| metric | 계산 |
|--------|-----|
| Average revenue per cohort | sum(revenue) / cohort size |
| Cumulative revenue curve | 시간별 누적 |
| LTV | retention curve × ARPU 적분 |
| CAC payback | CAC / monthly ARPU |

→ LTV / CAC 비율 *3:1 이상* 권장 (sustainable unit economics).

### Stage 4: Churn 분석

| 분류 | 정의 |
|------|------|
| Voluntary churn | 사용자가 명시적 cancel |
| Involuntary churn | 결제 실패 / 카드 만료 |
| Implicit churn | 사용 0, cancel 안 함 |

각 분류 별 *원인 + fix*:
- voluntary → exit survey / cancel flow 분석
- involuntary → dunning (결제 재시도) / 카드 갱신 알림
- implicit → re-engagement campaign

→ marketingskills/churn-prevention 패턴 차용.

### Stage 5: Cohort 비교 / segmentation

cohort 간 *통계적 유의 차이* 검증:
- t-test (평균 비교)
- chi-square (분류 변수)
- cohort plot (visualize)

발견 패턴:
- "paid acquisition cohort 의 LTV 가 organic 대비 30% 낮음"
- "Q3 cohort 의 D30 retention 이 Q2 대비 10% 높음 (왜?)"

## 5. 산출물 형식

```markdown
## User Cohort Analysis — {분기}

### Cohort 정의
- bucket: monthly
- segments: free / paid / channel

### Retention curve
| cohort | D1 | D7 | D30 | D90 | D180 |

### LTV / CAC
| cohort | LTV | CAC | 비율 | payback |

### Churn 분류
| 분류 | rate | top cause | fix |

### 발견 패턴
- ...
```

## 6. 검증

- [ ] Cohort 2~3 차원 cross 정의?
- [ ] Retention curve D1/D7/D30/D90 측정?
- [ ] LTV / CAC 비율 3:1+ 검증?
- [ ] 3 churn 분류 (voluntary / involuntary / implicit) 모두 분석?
- [ ] Cohort 비교 통계적 유의 검증?

## 7. 다음 phase

- `analyze-feature-adoption` 와 cross-reference
- `analyze-actor-failure-rate` — churn 의 *actor 차원* 분석
- `generate-improvement-tasks` — fix 우선순위 입력

## 8. 참조

- Cohort Analysis (Greg Linden / Lean Analytics)
- marketingskills/churn-prevention + analytics-tracking (외부 reference, MIT)
- Hooked (Nir Eyal) — habit / retention 원칙

---

## MCP integration (analytics-mcp v0.2.0+)

본 skill 의 *cohort retention curve 데이터 수집* 단계에서 `analytics_query_cohort` MCP tool 호출.

**호출 예**:

```json
{
  "tool": "analytics_query_cohort",
  "arguments": {
    "cohort_dimension": "weekly",
    "retention_metric": "active",
    "segments": ["plan_tier:pro", "country:KR"]
  }
}
```

**전제**: `BUDDY_ANALYTICS_BACKEND` env var 설정. v0.2.0 stubs — 본격 adapter 는 W4-2.2 ~ W4-2.3.

**관련 spec**: `../../../docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` §4.2
