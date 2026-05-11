# Analyze Feature Adoption — 신규 기능 활성화율 + power user 코호트

## 1. 목적

신규 release 된 feature 의 **adoption funnel** (awareness → trial → habit) 측정. *power user 코호트* 식별 + *abandonment 단계* 진단.

`analyze-user-cohort` (§8) 와 cascade — feature adoption 이 *user cohort* 의 sub-dimension.

## 2. 사용 시점

- §8 iterate-product 의 *feature release 사후 분석*
- A/B test 종료 후 — 어느 variant 가 adoption 높은지
- pivot 결정 시 — 핵심 feature 가 *adopt 되고 있나*
- onboarding 흐름 개선 — drop-off step 식별
- pricing tier 결정 — power user 와 free tier 분리

## 3. 입력

### 필수
- 분석 대상 feature 명 + release 시점
- analytics 데이터 (event tracking — feature 진입 / 사용 / 재사용)
- 기준 user base (active users 정의)

### 선택
- A/B test 분기 (variant 별 비교)
- 사용자 segment (`map-customer-segments` 산출)

## 4. Stage 흐름

### Stage 1: Awareness (feature 발견)

- 관련 page / button impression
- 관련 notification 노출 / 읽음
- 검색 / 메뉴에서 feature 도달율

→ awareness 율 = (impression unique users) / (active users)

### Stage 2: Trial (1회 사용)

- 첫 사용 conversion (impression → action)
- 사용 시간 (engagement 깊이)
- 첫 사용 *완료* 율 vs 도중 abandonment

→ trial 율 = (1회+ 사용 users) / (impression unique)

### Stage 3: Habit (반복 사용)

- D1 / D7 / D30 retention (feature 단위)
- weekly active feature use
- power user 정의 (e.g. weekly 5+ 사용)

→ habit 율 = (D30 retention) / (D1 retention)

### Stage 4: Power user 코호트 식별

| 차원 | 측정 |
|------|------|
| Demographics | segment / persona |
| Use depth | 평균 사용 빈도 / session 길이 |
| Use breadth | 다른 feature 동시 사용 |
| Tenure | 가입 후 기간 |
| Acquisition channel | 어디로부터 왔나 |

→ power user 의 *공통 패턴* 추출. acquisition / onboarding 에 활용.

### Stage 5: Abandonment 단계 진단

drop-off 큰 step 식별:
- impression → trial 낮음 → CTA 명확성 / value prop 문제
- trial → habit 낮음 → first-use experience 문제
- habit 후 churn → continuous value 부족

각 step 의 *fix hypothesis* + A/B test 후보.

## 5. 산출물 형식

```markdown
## Feature Adoption — {feature 이름}

### Funnel
| stage | 비율 | absolute |
| awareness | X% | N |
| trial | Y% | N |
| habit | Z% | N |

### Power user 코호트
| 차원 | 값 | sample |

### Abandonment 단계
- {step}: drop-off X%
- hypothesis: ...
- A/B 후보: ...
```

## 6. 검증

- [ ] 3 stage funnel (awareness / trial / habit) 모두 측정?
- [ ] Power user 정의 명시 (e.g. weekly 5+)?
- [ ] Power user 코호트 5 차원 분석?
- [ ] Abandonment step 의 fix hypothesis?
- [ ] 다음 A/B test 후보 명시?

## 7. 다음 phase

- `analyze-user-cohort` — feature adoption 이 user cohort 의 sub
- `design-ab-experiment` (구현됨) — abandonment fix 가설 검증
- `generate-improvement-tasks` (구현됨) — fix backlog 입력

## 8. 참조

- Hooked (Nir Eyal) — habit formation
- Lean Analytics (Croll + Yoskovitz) — funnel metric
- marketingskills/analytics-tracking (외부 reference, MIT — analytics 패턴)

---

## MCP integration (analytics-mcp v0.2.0+)

본 skill 의 *adoption funnel 데이터 수집* 단계에서 `analytics_query_funnel` MCP tool 을 호출해 직접 데이터를 가져올 수 있다 (buddy-mcp 서버 enable 시).

**호출 예 (Claude Code session 내부)**:

```json
{
  "tool": "analytics_query_funnel",
  "arguments": {
    "stages": ["feature_seen", "feature_used", "feature_used_d7"],
    "time_range": { "from": "2026-04-01T00:00:00Z", "to": "2026-05-01T00:00:00Z" },
    "segment": { "dimension": "plan_tier", "value": "pro" }
  }
}
```

**전제**: `BUDDY_ANALYTICS_BACKEND` env var (sql / mixpanel / amplitude / datadog / stripe / elasticsearch) 설정. v0.2.0 ships stubs — 모든 호출은 friend-tone "backend not configured" 응답. 본격 adapter 구현 (W4-2.2 ~ W4-2.3) 후 실 데이터.

**관련 spec**: `../../../docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` §4.1
