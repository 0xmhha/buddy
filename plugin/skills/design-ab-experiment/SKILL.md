---
name: design-ab-experiment
description: This skill should be used when the user wants to "design an A/B test", "set up an experiment", "test a hypothesis", "split test", "run a controlled experiment", or needs to design a statistically valid experiment for a product change.
---

# design-ab-experiment

A/B 실험을 통계적으로 유효하게 설계한다. 가설 → 표본 크기 → 대조군/실험군 → 측정 지표 → 실험 기간을 순서대로 정의한다.

**§8 iterate-product stage skill.** 단독 호출도 가능 (dual-mode).

---

## 실험 설계 절차

### 1. 가설 정의

가설 형식 (측정 가능해야 함):
```
"[변수]를 [변경]하면 [지표]가 [방향]으로 [크기] 변화한다."
예: "회원가입 버튼 색을 파란색에서 초록색으로 바꾸면 CTR이 10% 향상된다."
```

- **Null hypothesis (H₀)**: 변경이 지표에 영향 없음
- **Alternative hypothesis (H₁)**: 변경이 지표에 [방향] 영향
- **Primary metric**: 실험 성공을 판단하는 단일 지표 (one per experiment)
- **Guardrail metrics**: 개선되어도 해쳐서는 안 되는 지표 (예: latency, error rate)

### 2. 표본 크기 계산

```
필요 입력:
- Baseline conversion rate (p₀): 현재 primary metric 값
- Minimum detectable effect (MDE): 의미 있는 최소 변화량 (예: +5%)
- Significance level (α): 0.05 (95% confidence)
- Statistical power (1-β): 0.80 (80% power)

표본 크기 공식 (두 비율 비교):
n = (Z_α/2 + Z_β)² × (p₀(1-p₀) + p₁(1-p₁)) / (p₀ - p₁)²

여기서 p₁ = p₀ × (1 + MDE)
Z_0.05/2 = 1.96, Z_0.80 = 0.84
```

실용적 계산표 (MDE 5%, baseline 20%):
| α | Power | 표본/그룹 |
|---|-------|---------|
| 0.05 | 0.80 | ~2,150 |
| 0.05 | 0.90 | ~2,900 |
| 0.01 | 0.80 | ~2,950 |

### 3. 분배 전략

- **50/50 split**: 표준. 동등한 exposure, 최소 실험 기간.
- **10/90 canary**: 리스크 최소화. 표본 확보에 더 긴 시간 필요.
- **Multi-arm**: 3+ 변형 비교 시. Bonferroni correction 필요.

**randomization unit** 선택:
- user_id (user-level): 일관된 경험. 권장.
- session_id: 신규/비로그인 포함. user contamination 위험.
- request_id: 페이지별. SUTVA 위반 가능.

### 4. 측정 지표 정의

```yaml
experiment_id: {kebab-case-id}
hypothesis: "{가설 문장}"
primary_metric:
  name: {지표명}
  type: {conversion_rate / mean / ratio}
  baseline: {현재 값}
  mde: {최소 감지 효과}
guardrail_metrics:
  - name: {지표명}
    direction: not-decrease
secondary_metrics:
  - name: {지표명}
    purpose: {관찰 목적}
sample_size:
  per_group: {N}
  total: {2N 또는 그 이상}
allocation:
  control: {%}
  treatment: {%}
duration_estimate: {일수}
actor_segment: {어느 actor use case에 해당하는가}
```

### 5. 실험 기간

최소 실험 기간:
- 주 효과 사이클 1회 이상 포함 (주중/주말 포함 최소 7일)
- 표본 크기 달성 예상 시간 고려
- Novelty effect 소멸 시간 (신기능 → 7-14일 후 정상화)

권장: **2주** (주 효과 2사이클 + novelty effect 소멸).

---

## 실험 체크리스트

- [ ] 가설이 측정 가능하고 falsifiable한가?
- [ ] Primary metric이 단 하나인가?
- [ ] Guardrail metric을 정의했는가?
- [ ] 표본 크기가 충분한가? (계산 완료)
- [ ] Randomization unit이 결정됐는가?
- [ ] 최소 7일 이상 실험 기간인가?
- [ ] Feature flag로 rollback 가능한가?
- [ ] §8 `iterate-product`의 actor funnel에 연결됐는가?

---

## 다음 단계

실험 완료 후 → `analyze-ab-experiment` skill
