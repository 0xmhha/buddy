# Triage Customer Support Ticket — 분류 + recurring issue 패턴화 + product feedback loop

## 1. 목적

CS (customer support) 티켓을 *분류 + 우선순위 + recurring pattern 식별 + product feedback* 로 변환. 단순 *답변* 이 아니라 **product 개선 input** 으로.

`generate-improvement-tasks` (구현됨) + `analyze-customer-feedback-corpus` (§8) 와 cascade.

## 2. 사용 시점

- §8 iterate-product 의 *분기 CS 분석*
- 신규 release 후 1~2 주 — release 관련 ticket 급증 검토
- CS 팀 backlog 늘어남 — 자동 분류 / 자동 답변 후보 식별
- product 결정 시 *어느 영역이 사용자 마찰 큰가* 입력
- `analyze-customer-feedback-corpus` 의 raw input

## 3. 입력

### 필수
- ticket data (subject + body + timestamp + user)
- 분류 taxonomy (사전 정의 또는 본 skill 산출)
- CS 도구 (Zendesk / Intercom / Helpscout 등)

### 선택
- product roadmap (current sprint / 다음 release)
- 사용자 segment (`map-customer-segments`) — segment 별 분포

## 4. Stage 흐름

### Stage 1: 1차 분류 (자동 가능)

| 분류 | 비율 (예시) |
|------|---------|
| **How-to question** | 30% — 사용법 문의 |
| **Bug report** | 25% — 동작 안 함 |
| **Feature request** | 20% — 새 기능 / 변경 요청 |
| **Account / billing** | 15% — 결제 / 구독 / 계정 |
| **Praise / general** | 5% — 칭찬 / 일반 |
| **Other** | 5% |

→ ML / keyword 기반 자동 분류 가능 (Intercom Resolution Bot 등). 인간 review 필요한 경우만 확대.

### Stage 2: Severity / impact 측정

| severity | 정의 | SLA |
|---------|------|-----|
| P0 | 사용 불가 / 데이터 손실 | 1h |
| P1 | 핵심 기능 영향 | 4h |
| P2 | 일부 기능 영향 / workaround 있음 | 24h |
| P3 | minor / cosmetic | 7d |

→ SLA 자동 측정 + breach alert.

### Stage 3: Recurring pattern 식별

같은 issue 가 *반복 발생* 시:
- 1주 내 5+ 같은 ticket → *알려진 issue* 표시 + 자동 답변 후보
- 30일 내 20+ 같은 issue → *product fix* 후보
- 같은 segment / channel 에서 다수 → segment-specific issue

→ recurring issue 의 *runbook + auto-resolution* 작성.

### Stage 4: Product feedback loop

각 ticket 의 *product 입력* 추출:

| 추출 | 대상 |
|------|------|
| Bug → bug tracker | linked GitHub issue / Linear ticket |
| Feature request → product backlog | `score-feature-priority` 입력 |
| How-to → docs gap | docs improvement backlog |
| Account → account fix + retention | `analyze-user-cohort` 입력 |

### Stage 5: Knowledge base 운영

자주 발생하는 issue 의 *self-service* :
- FAQ 작성 (top-10 ticket → FAQ entry)
- in-app help (context-aware)
- chatbot training data

→ *ticket 감소 = 좋은 product*. KB 가 ticket 줄임.

### Stage 6: 분기 / 월 dashboard

| metric | 측정 |
|--------|-----|
| Ticket volume | 월 / 일 |
| Avg resolution time | 분류 별 |
| Recurring rate | 같은 issue 반복 비율 |
| Self-service rate | KB 조회 / ticket 비율 |
| CSAT | post-resolution survey |

## 5. 산출물 형식

```markdown
## CS Ticket Triage — {분기}

### 1차 분류 분포
| 분류 | volume | % |

### Severity / SLA breach
| severity | count | SLA breach |

### Recurring pattern
- {issue}: {N tickets / period} → {action}

### Product feedback
- bug → tracker: {N}
- feature request → backlog: {N}
- how-to → docs gap: {N}

### KB / self-service
- new FAQ: {N}
- self-service rate: {%}
```

## 6. 검증

- [ ] 5+ 분류 taxonomy?
- [ ] Severity 4 단계 (P0~P3) + SLA?
- [ ] Recurring pattern 식별 (1주 5+ / 30일 20+ threshold) ?
- [ ] Product feedback 4 영역 (bug / feature / docs / account) loop?
- [ ] KB / self-service 운영?
- [ ] 분기 dashboard metric 5+ ?

## 7. 다음 phase

- `analyze-customer-feedback-corpus` (§8) — 텍스트 corpus 토픽 모델링 입력
- `generate-improvement-tasks` (구현됨) — product fix 백로그
- `audit-error-budget` (§8) — P0/P1 ticket 이 SLO budget burn

## 8. 참조

- Intercom Customer Support Trends Report
- The Effortless Experience (Matthew Dixon) — CS quality
- Zendesk Triage automation patterns
