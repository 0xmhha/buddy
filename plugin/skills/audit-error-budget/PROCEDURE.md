# Audit Error Budget — SLO burn rate + release gate

## 1. 목적

`design-observability` (§3) 에서 정의된 SLO 의 **error budget consumption** 측정 + *release gate* 적용 + *burn rate alert*. SLO 가 *명시* 만 됐으면 의미 없음 — *측정 + decision-driver* 가 본 skill.

`analyze-actor-failure-rate` + `chaos-test` + `analyze-cost-anomaly` 의 incident burn 통합.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| SLO 정의 | ✅ | artifact / knowledge | `design-observability` 산출물 또는 사용자 정의 | "SLO 지표와 목표를 알려주세요." |
| 운영 데이터 | ✅ | artifact | 모니터링 시스템 | (자동 수집) |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Error budget burn rate report (multi-window + release gate) | artifact | structured report | `ship-release` (gate) |

## 2. 사용 시점

- §8 iterate-product 의 *분기 SLO review*
- release 직전 — error budget 잔여로 *release 결정*
- incident 사후 — budget burn 측정
- chaos-test 결과 분석 — controlled burn
- on-call review — alert 정합 검증

## 3. 입력

### 필수
- `design-observability` 산출 — SLO / SLI 정의
- production metrics (latency / error rate / availability)
- incident log (`conduct-postmortem` 산출)

### 선택
- 사용자 SLA contractual obligation
- `analyze-actor-failure-rate` 산출 — actor 별 burn 분포

## 4. Stage 흐름

### Stage 1: Error budget 계산

```
SLO = 99.9% / month
error budget = 100% - 99.9% = 0.1%
month minutes = 30 * 24 * 60 = 43,200
allowed downtime = 43,200 * 0.001 = 43.2 min / month
```

### Stage 2: Burn rate 측정

| window | burn rate alert |
|--------|------------|
| 1 hour | > 14.4× → 6 hour 안 budget 소진 |
| 6 hour | > 6× → 24 hour 안 소진 |
| 1 day | > 3× → 4 day 안 소진 |
| 3 day | > 1× → 30 day 안 소진 (slow burn) |

→ multi-window alert (Google SRE Book 패턴).

### Stage 3: Release gate

| budget 잔여 | release 정책 |
|----------|----------|
| > 50% | 정상 release |
| 25~50% | release 가능, 단 *high-risk feature* 유보 |
| < 25% | release 동결 — bug fix 만 |
| 0% (소진) | *모든 release 동결* — reliability 우선 |

→ *budget 으로 release velocity 협의* — incident 후 *페널티* + reliability work 의 *명시 budget*.

### Stage 4: Burn 출처 분석

burn 발생 시 *어디서* :
- incident (P0/P1) — single incident 가 큰 burn
- chronic degradation — 여러 작은 incident 누적
- chaos experiment — controlled burn (intentional)
- 3rd-party fail — 외부 의존
- deployment regression — release 직후 spike

→ `analyze-actor-failure-rate` (§8) cross-reference.

### Stage 5: Budget 정책 review

분기 review:
- SLO target 가 *현실적* 인가? (너무 빡빡 / 너무 헐거움)
- 사용자 *체감* SLA 와 정합?
- 비용 / reliability trade-off 적절?

→ SLO 변경은 *큰 결정* — `write-adr` (§3) 으로 영속화.

### Stage 6: Multi-window alert 운영

```
fast burn (1h × 14.4×) → page on-call
medium burn (6h × 6×) → Slack warning
slow burn (3d × 1×) → weekly review
```

→ alert fatigue 회피 — fast burn 만 page.

## 5. 산출물 형식

```markdown
## Error Budget Audit — {분기}

### SLO 정의
| SLI | target | budget |

### Burn 측정
| window | actual rate | alert level |

### Release gate 결정
- 잔여: X%
- 정책: 정상 / 제한 / 동결

### Burn 출처
| 출처 | budget burn % | mitigation |

### SLO review
- target 적정성: ...
- 변경 결정: ADR-{N}
```

## 6. 검증

- [ ] SLO error budget 계산 (allowed downtime / month)?
- [ ] Multi-window burn rate alert (1h / 6h / 1d / 3d)?
- [ ] Release gate 4 단계 정책 (>50 / 25-50 / <25 / 0%)?
- [ ] Burn 출처 분류 5+ 영역?
- [ ] 분기 SLO review + ADR 영속화?
- [ ] Alert fatigue 회피 (fast burn 만 page)?

## 7. 다음 phase

- `analyze-actor-failure-rate` (§8) — actor 별 burn 분포
- `conduct-postmortem` (구현됨) — incident → budget 영향 측정
- `prepare-launch-checklist` (§7, 구현됨) — release gate 입력
- `chaos-test` (§6) — controlled burn 의 *학습 가치*

## 8. 참조

- Google SRE Book Ch.3-4 + Ch.21 (Multi-Window Multi-Burn-Rate Alerts)
- The Site Reliability Workbook
- buddy `design-observability` (§3) — SLO 정의 입력

---

## MCP integration (analytics-mcp v0.2.0+)

본 skill 의 *SLO burn rate + release gate decision* 에서 `analytics_query_slo_burn` MCP tool 호출.

**호출 예**:

```json
{
  "tool": "analytics_query_slo_burn",
  "arguments": {
    "sli": "p99_latency_ms",
    "slo_target": 250,
    "time_window": "1d"
  }
}
```

기대 응답: current_value + budget_remaining_pct + burn_rate (multi-window) + alert_level (info / warning / critical) + release_gate_decision (ship / limit / freeze / rollback).

**전제**: `BUDDY_ANALYTICS_BACKEND` env var 설정 (datadog 권장 — SLO 영역 강함). v0.2.0 stubs — 본격 adapter 는 W4-2.2 ~ W4-2.3.

**관련 spec**: `../../../docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` §4.6
