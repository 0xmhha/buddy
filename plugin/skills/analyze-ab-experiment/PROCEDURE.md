# analyze-ab-experiment

완료된 A/B 실험 결과를 분석하고 Ship / Revert / Continue 결정을 내린다. 통계적 유의성 + 실용적 유의성을 모두 검토한다.

**8단계 iterate-product stage skill.** 단독 호출도 가능 (dual-mode).


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 실험 데이터 (control/treatment 결과) | ✅ | artifact | 실험 플랫폼 데이터 | "실험 결과 데이터를 제공해 주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| 실험 분석 결과 (Ship/Revert/Continue 결정) | decision | structured report | `generate-improvement-tasks` |

---

## 분석 절차

### 1. 데이터 수집 체크

분석 전 확인:
- [ ] 계획된 표본 크기 달성 (early stopping 금지)
- [ ] 실험 기간 충족 (최소 7일)
- [ ] 데이터 품질 (샘플 불균형 없음, 로깅 정상)
- [ ] AA test 통과 여부 (randomization 정상)

### 2. 통계 검정

**비율 비교 (conversion rate):**
```
z = (p_treatment - p_control) / sqrt(p_pooled × (1-p_pooled) × (1/n_t + 1/n_c))
p-value = 2 × (1 - Φ(|z|))  # 양측 검정

결정:
  p-value < α(0.05): reject H₀ (통계적으로 유의)
  p-value ≥ α(0.05): fail to reject H₀
```

**연속값 비교 (mean):**
```
t = (μ_treatment - μ_control) / sqrt(s²_t/n_t + s²_c/n_c)
degrees of freedom: Welch's approximation
```

**신뢰구간:**
```
95% CI = (p_t - p_c) ± 1.96 × SE
```

### 3. 실용적 유의성 확인

통계적 유의성 ≠ 실용적 유의성. 둘 다 충족해야 Ship 가능.

```
실용적 유의성 확인:
- 관측된 effect size ≥ MDE (최소 감지 효과)
- CI 하한도 MDE보다 큰가? (확신 있는 개선)
- Guardrail metric 악화 없음
- 비용/구현 대비 gain이 충분한가?
```

### 4. 세그먼트 분석

전체 결과 외에 actor별 세그먼트를 분석한다:
```
actor: user-actor
  conversion (form → submit): control 20% vs treatment 24% (+4pp)
  p-value: 0.03 ✓

actor: backend-actor
  success_rate: control 99.1% vs treatment 99.3% (+0.2pp)
  p-value: 0.4 (not significant)

actor: email-verifier
  deliverability: control 92% vs treatment 91% (-1pp)
  p-value: 0.08 (guardrail: not degraded significantly)
```

### 5. 결정 프레임워크

```
Ship (배포):
  조건: p-value < 0.05 AND effect ≥ MDE AND guardrail 악화 없음
  행동: feature flag 100% 전환 → ship-release 7-3 단계 fast path

Revert (롤백):
  조건: guardrail metric 유의미하게 악화 OR primary metric 악화 확인
  행동: feature flag off → conduct-postmortem → 재설계

Continue (연장):
  조건: 방향은 맞지만 표본 부족 OR 효과 불확실
  행동: 실험 기간 연장 (단, α inflation 주의 — sequential testing 고려)

No Result (결론 없음):
  조건: p-value ≥ 0.05 AND CI에 MDE 포함
  행동: 가설 재검토 → design-ab-experiment 재설계
```

---

## 결과 리포트 포맷

```markdown
## A/B 실험 결과 — {experiment_id}

**기간**: {start} ~ {end} ({N}일)
**표본**: Control {N_c}명 / Treatment {N_t}명

### Primary Metric: {metric_name}
| 그룹 | 값 | 95% CI |
|------|-----|--------|
| Control | {p_c} | ± {E_c} |
| Treatment | {p_t} | ± {E_t} |

- 효과: {p_t - p_c} ({%} 상대 변화)
- p-value: {값}
- 통계 유의성: {✓ / ✗}
- 실용 유의성: {✓ / ✗} (MDE {mde}% 대비)

### Guardrail Metrics
{각 guardrail metric 요약}

### Actor 세그먼트
{actor별 세그먼트 결과}

### 결정
**{Ship / Revert / Continue / No Result}**
이유: {결정 근거}

### 다음 단계
{→ generate-improvement-tasks 또는 재설계}
```

---

## 다음 단계

- Ship → `ship-release` 7-3 단계 fast path
- Revert → `conduct-postmortem`
- 결과 기반 개선 → `generate-improvement-tasks`

---

## MCP integration (analytics-mcp v0.2.0+)

본 skill 의 *A/B 결과 fetch* 단계에서 `analytics_query_ab_experiment` MCP tool 호출.

**호출 예**:

```json
{
  "tool": "analytics_query_ab_experiment",
  "arguments": {
    "experiment_id": "exp_2026_q2_signup_flow"
  }
}
```

기대 응답: variants + sample_size + primary metric + 95% CI + significance + guardrail + recommendation ("ship" | "revert" | "continue" | "inconclusive").

**전제**: `BUDDY_ANALYTICS_BACKEND` env var 설정. v0.2.0 stubs — 본격 adapter 는 W4-2.2 ~ W4-2.3.

**관련 spec**: `../../../docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` §4.3
