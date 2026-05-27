# Optimize Conversion Funnel — CRO (Conversion Rate Optimization) 통합

## 1. 목적

acquisition → activation → retention → revenue → referral 의 **Pirate Metrics (AARRR) funnel** 각 단계의 conversion rate 측정 + bottleneck 식별 + A/B test driven 개선.

`marketingskills` 의 5 CRO sub-skill (onboarding / form / page / paywall-upgrade / popup CRO) 을 buddy 1 skill 로 통합.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Funnel 데이터 | ✅ | artifact | analytics 플랫폼 | "최적화할 funnel과 데이터를 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| CRO 추천 (biggest-drop bottleneck + A/B pipeline) | artifact | structured report | `design-ab-experiment` |

## 2. 사용 시점

- §8 iterate-product 의 *분기 funnel review*
- 신규 acquisition channel 도입 후 — channel 별 conversion 비교
- pricing tier 변경 후 — paywall conversion 변화
- onboarding 흐름 개선 — D1 retention 영향
- A/B test pipeline 우선순위 결정

## 3. 입력

### 필수
- analytics 데이터 (event tracking — 단계 별 진입 / 이탈)
- AARRR funnel 단계 정의
- baseline conversion rate (이전 분기)

### 선택
- segment 별 funnel (`map-customer-segments` 산출)
- A/B test 결과 (`analyze-ab-experiment` 구현됨 — cascade in)

## 4. Stage 흐름

### Stage 1: AARRR funnel 정의

| 단계 | 정의 | 측정 |
|------|------|------|
| **Acquisition** | 첫 방문 / 가입 | unique signups |
| **Activation** | "aha moment" 도달 | first value action 완료율 |
| **Retention** | 반복 사용 | D7 / D30 retention |
| **Revenue** | 결제 / upgrade | free → paid conversion |
| **Referral** | 추천 / 공유 | viral coefficient |

### Stage 2: CRO sub-domain 5 영역

각 단계별 *CRO 도구*:

| sub-domain | 적용 단계 | 핵심 lever |
|---------|---------|---------|
| **Onboarding CRO** | Activation | first-use experience / aha moment 빠른 도달 |
| **Form CRO** | Acquisition / Revenue | field 수 / validation timing / progressive disclosure |
| **Page CRO** | 모든 단계 | landing page / 가격 page / feature page CTA |
| **Paywall / Upgrade CRO** | Revenue | tier 표시 / 가격 visibility / 가치 강조 |
| **Popup CRO** | Activation / Retention | timing / 빈도 / dismissal cost |

→ marketingskills 의 5 sub-skill 본 skill 로 통합 — 일관 funnel 관점.

### Stage 3: Bottleneck 식별

각 단계의 *drop-off* 측정 + bottleneck 우선순위:

```
Acquisition  100   (baseline)
Activation    40   (60% drop) ← biggest bottleneck
Retention     20   (50% drop)
Revenue        4   (80% drop)
Referral       1   (75% drop)
```

→ *biggest drop* 부터 fix. 다른 단계 fix 가 *상위 단계 영향* 도 검토.

### Stage 4: A/B test pipeline

각 bottleneck 의 *fix hypothesis* + A/B test:

| 단계 | hypothesis | test |
|------|---------|------|
| Activation drop | onboarding 너무 길음 | tutorial step 5 → 3 |
| Form drop | field 너무 많음 | optional field 제거 |
| Paywall drop | 가치 보이지 않음 | feature 비교 표 추가 |

→ `design-ab-experiment` (구현됨) cascade.

### Stage 5: Continuous monitoring

분기별 funnel snapshot + drift 감지:
- 각 단계 conversion rate 의 ±10% 변화 alert
- 새 release / channel 도입 후 *funnel 영향* 측정
- segment 별 funnel 비교 (어느 segment 가 drop dominant)

## 5. 산출물 형식

```markdown
## Conversion Funnel — {분기}

### AARRR funnel
| 단계 | conversion | drop-off |

### Bottleneck 우선순위
1. Activation (60% drop) → onboarding CRO
2. Revenue (80% drop) → paywall CRO

### A/B test pipeline
| bottleneck | hypothesis | test 설계 | 우선순위 |

### Drift monitoring
- alert 정책: ...
```

## 6. 검증

- [ ] AARRR 5 단계 모두 conversion 측정?
- [ ] 5 CRO sub-domain 적용 영역 매핑?
- [ ] Biggest drop bottleneck 우선순위?
- [ ] A/B test hypothesis 명시?
- [ ] Drift monitoring 알림 정책?

## 7. 다음 phase

- `design-ab-experiment` (구현됨) — fix hypothesis 검증
- `analyze-ab-experiment` (구현됨) — Ship/Revert 결정
- `analyze-feature-adoption` 와 cross-reference

## 8. 참조

- AARRR Pirate Metrics (Dave McClure) — funnel framework
- Hooked (Nir Eyal) — habit / activation
- marketingskills 5 CRO sub-skills (외부 reference, MIT — onboarding-cro / form-cro / page-cro / paywall-upgrade-cro / popup-cro)

---

## MCP integration (analytics-mcp v0.2.0+)

본 skill 의 *AARRR funnel 5 단계별 conversion 측정* 에서 `analytics_query_funnel` MCP tool 호출 (analyze-feature-adoption 과 동일 tool 공유).

**호출 예 — Activation 단계 변환율 측정**:

```json
{
  "tool": "analytics_query_funnel",
  "arguments": {
    "stages": ["acquisition", "activation_event_1", "activation_event_2"],
    "time_range": { "from": "2026-04-01T00:00:00Z", "to": "2026-05-01T00:00:00Z" },
    "segment": { "dimension": "channel", "value": "organic" }
  }
}
```

5 단계 (acquisition / activation / retention / revenue / referral) 별로 호출해 biggest-drop bottleneck 식별 → A/B test pipeline 입력.

**전제**: `BUDDY_ANALYTICS_BACKEND` env var 설정. v0.2.0 stubs — 본격 adapter 는 W4-2.2 ~ W4-2.3.

**관련 spec**: `../../../docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` §4.1
