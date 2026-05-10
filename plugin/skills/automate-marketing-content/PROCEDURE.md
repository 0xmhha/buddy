# Automate Marketing Content — email sequence + cold email + content schedule

## 1. 목적

마케팅 콘텐츠의 *작성 → 스케줄 → 발송 → 측정 → 학습* cycle 자동화. **email sequence + cold email + content calendar** 통합.

`draft-marketing-copy` 가 *콘텐츠 작성* 이라면 본 skill 은 *delivery + automation*.

## 2. 사용 시점

- §8 iterate-product 의 *마케팅 운영 cadence*
- 신규 lead nurture sequence 설계
- cold email outreach 캠페인 시작
- content calendar 분기 / 월 plan
- email engagement metric drift 시 (open / click rate 하락)

## 3. 입력

### 필수
- `map-customer-segments` — segment 별 nurture 분기
- `draft-marketing-copy` 산출 — 콘텐츠 자산
- email tool (SendGrid / Postmark / Mailgun / Resend / native)

### 선택
- 기존 sequence 데이터 (open / click / reply rate)
- compliance 영역 (CAN-SPAM / GDPR / KISA)

## 4. Stage 흐름

### Stage 1: Email sequence 설계

| 단계 | timing | 목적 |
|------|------|------|
| Welcome | 즉시 | onboarding + 가치 강조 |
| Day 1 | 1 day | first-value action 유도 |
| Day 3 | 3 day | tip / feature 도입 |
| Day 7 | 7 day | use case 사례 |
| Day 14 | 14 day | advanced feature |
| Day 30 | 30 day | upgrade 제안 (free → paid) |

→ 각 email 의 *single CTA* + *segment 별 변형*.

### Stage 2: Cold email outreach

cold = *미가입 사용자* 대상. compliance 강함.

| 항목 | 가이드 |
|------|------|
| 주제줄 | < 50 char / personalized |
| 본문 | < 100 word / specific value / 단일 CTA |
| Personalization | 회사 / 역할 / 최근 trigger 언급 |
| Sender | 개인 이름 (no-reply X) |
| Compliance | 수신 거부 link 명시 / 회사 정보 |
| Cadence | 첫 발송 + 3일 후 follow-up + 7일 후 final |

→ open rate 30%+ / click 10%+ / reply 3%+ 가 baseline. cold 의 success metric 은 *response* 우선.

### Stage 3: Content calendar

| 빈도 | 콘텐츠 종류 |
|------|----------|
| Daily | social post (Twitter / LinkedIn) |
| Weekly | blog post / newsletter |
| Bi-weekly | long-form (case study / tutorial) |
| Monthly | webinar / podcast |
| Quarterly | major report / launch |

calendar 도구: Notion / Trello / Buffer / Hypefury / 자체.

### Stage 4: Automation

| 도구 종류 | 적용 |
|--------|------|
| Email automation | Mailchimp / ConvertKit / customer.io / native |
| Social scheduling | Buffer / Hootsuite / Hypefury |
| CRM trigger | HubSpot / Salesforce — event 기반 발송 |
| Workflow | n8n / Zapier — multi-step automation |
| AI 보조 | LLM 으로 draft 자동 + 인간 review |

→ automation 의 *실패 fallback* 명시 (예: send fail → 다음 발송 cycle 자동 retry).

### Stage 5: Measurement + 학습

| metric | 측정 |
|--------|-----|
| Delivery rate | bounce / spam 회피 |
| Open rate | subject 강도 |
| Click rate | body + CTA 강도 |
| Reply rate (cold) | personalization quality |
| Conversion rate | full funnel 정합 |
| Unsubscribe rate | < 0.5% baseline |

각 sequence / email 의 결과 → `analyze-customer-feedback-corpus` 의 *implicit signal* 입력.

### Stage 6: Compliance

| 지역 | 법령 |
|------|-----|
| USA | CAN-SPAM Act |
| EU | GDPR (consent-based) |
| Korea | 정보통신망법 / KISA 가이드 |
| Canada | CASL |

→ all email 에 *수신 거부 link* + *회사 정보* + *consent record* (GDPR / KISA).

## 5. 산출물 형식

```markdown
## Marketing Content Automation — {분기 / 캠페인}

### Email sequence
| 단계 | timing | CTA | segment |

### Cold email cadence
- 첫 발송 / follow-up 3d / final 7d
- compliance: ...

### Content calendar
| frequency | type | platform |

### Automation tools
- email / social / CRM / workflow

### Metric baseline
| metric | target | actual |

### Compliance check
- region: USA / EU / KR
```

## 6. 검증

- [ ] Email sequence 6+ 단계 timing?
- [ ] Cold email compliance (수신 거부 + 회사 정보)?
- [ ] Content calendar 5 frequency 영역?
- [ ] Automation 도구 5+ 영역 (email / social / CRM / workflow / AI)?
- [ ] 6 metric 모두 측정 + baseline?
- [ ] Region 별 compliance 적용?

## 7. 다음 phase

- `analyze-customer-feedback-corpus` (§8) — implicit signal (email engagement)
- `analyze-user-cohort` (§8) — channel attribution
- `optimize-conversion-funnel` — email → conversion cascade

## 8. 참조

- Influence (Robert Cialdini) — 6 persuasion principle
- The Mom Test (Fitzpatrick) — cold outreach 의 진실 발견
- marketingskills 3 sub-skills (외부 reference, MIT — email-sequence / cold-email / content-strategy)
