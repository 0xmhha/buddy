---
name: analyze-user-funnel
description: This skill should be used when the user wants to "analyze the funnel", "find where users drop off", "analyze conversion rates", "track user flow", "measure actor-level metrics", or needs to understand where users are failing in a product flow.
---

# analyze-user-funnel

§2 use case 분해 결과를 기반으로 actor별 전환/이탈을 분석한다. 어느 actor 단계에서 drop-off가 발생하는지 추적해 개선 우선순위를 결정한다.

**§8 iterate-product stage skill.** 단독 호출도 가능 (dual-mode).

---

## Funnel 분석 절차

### 1. §2 Use Case 기반 Funnel 정의

§2 feature spec의 actor / use case를 기반으로 funnel 단계를 정의한다.

예시 (`signup-email-password` feature):
```yaml
funnel:
  name: signup-email-password
  stages:
    - id: page-load
      actor: user-actor
      event: signup_page_view
      description: 회원가입 페이지 진입
    - id: form-start
      actor: user-actor
      event: signup_form_interact
      description: 폼 첫 입력
    - id: form-submit
      actor: user-actor
      event: signup_form_submit
      description: 제출 버튼 클릭
    - id: backend-success
      actor: backend-actor
      event: signup_api_200
      description: API 성공 응답
    - id: email-sent
      actor: email-verifier-actor
      event: verification_email_sent
      description: 인증 이메일 발송
    - id: email-clicked
      actor: email-verifier-actor
      event: verification_link_clicked
      description: 인증 링크 클릭
    - id: first-login
      actor: user-actor
      event: first_login_success
      description: 최초 로그인 성공
```

### 2. 데이터 수집 쿼리

각 funnel stage의 conversion을 측정하기 위한 데이터 쿼리 패턴:

```sql
-- 단계별 사용자 수
SELECT
  stage,
  COUNT(DISTINCT user_id) as users,
  COUNT(DISTINCT user_id) * 100.0 / LAG(COUNT(DISTINCT user_id)) OVER (ORDER BY stage_order) as conversion_rate
FROM funnel_events
WHERE date BETWEEN '{start}' AND '{end}'
GROUP BY stage, stage_order
ORDER BY stage_order;
```

### 3. Actor별 분석

actor 단위로 drop-off를 집계한다:

```
user-actor funnel:
  page-load → form-start:  85% (15% 이탈 — 진입 의지 없음)
  form-start → form-submit: 60% (40% 이탈 — UX 마찰)
  email-clicked → first-login: 70% (30% 이탈 — email 클릭 후 이탈)

backend-actor performance:
  signup_api success rate: 99.2% (0.8% 실패 — DB / validation error)
  p50 latency: 120ms, p99: 890ms

email-verifier-actor:
  email delivery rate: 94% (6% bounce)
  link click rate (중 delivered): 62%
```

### 4. Drop-off 분류

각 drop-off를 원인별로 분류한다:

| 단계 | Drop-off | 원인 분류 | Actor |
|------|---------|---------|-------|
| form-start → submit | 40% | UX 마찰 (폼 복잡도) | user-actor |
| backend: 0.8% 실패 | 0.8% | 기술 결함 | backend-actor |
| email delivery 6% bounce | 6% | 3rd-party 제한 | email-verifier-actor |
| link click rate 62% | 38% | 이메일 내용 / 타이밍 | email-verifier-actor |

### 5. 개선 우선순위 결정

```
Impact = drop-off_rate × stage_users × business_value_per_conversion

예시:
  form-start → submit: 40% × 1000명/일 × $5 = $2,000/일 개선 가능
  email link click: 38% × 940명/일 × $5 = $1,786/일 개선 가능
  backend failure: 0.8% × 560명/일 × $5 = $22/일 개선 가능
```

우선순위: form UX 개선 > email 최적화 > backend 안정성

---

## Funnel 리포트 포맷

```markdown
## Funnel 분석 — {feature_name}

**기간**: {start} ~ {end}
**총 진입**: {N}명/일

### Actor별 Conversion

#### user-actor
| 단계 | 사용자 | Conversion |
|------|--------|-----------|
| page-load | 1,000 | — |
| form-start | 850 | 85% |
| form-submit | 510 | 60% |
| first-login | 357 | 70% |

**전체 user-actor conversion**: 35.7%
**최대 drop-off**: form-start → submit (40%)

#### backend-actor
- API success rate: 99.2%
- p50 latency: 120ms, p99: 890ms

#### email-verifier-actor
- Delivery rate: 94%
- Link click rate: 62%
- 전체 completion: 58.3%

### 개선 기회

| 우선순위 | Actor | 이슈 | 예상 Impact |
|---------|-------|------|-----------|
| 1 | user-actor | form UX 마찰 | $2,000/일 |
| 2 | email-verifier | link click 저조 | $1,786/일 |
| 3 | backend | 0.8% failure | $22/일 |

### 권장 다음 단계
→ generate-improvement-tasks (상위 2개 이슈 → feature spec 생성)
```

---

## 다음 단계

- `generate-improvement-tasks` — 분석 결과 → improvement backlog 생성
- `design-ab-experiment` — 개선 가설 검증
- `define-features` (§2) — 구조적 변경이 필요한 경우 재진입
