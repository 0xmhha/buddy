---
name: assess-business-viability
description: "아이디어가 사업으로 성립하는지 7차원(TAM/SAM/SOM, 고객-구매자, willingness-to-pay, GTM, 경쟁, unit economics, 규제)으로 평가하고 go/pivot/no-go 결정. 트리거: '이거 사업 되겠어?' / 'TAM 분석' / '사업성 검토' / '이 가격에 살 사람 있을까?' / 'go/no-go 결정' / 'unit economics 봐줘' / '비즈니스 모델 검증'. 입력: validate-idea 통과 후 아이디어. 출력: viability score + verdict + critical assumptions. 흐름: validate-idea → assess-business-viability → define-product-spec."
type: skill
---

# Assess Business Viability — 사업성 7차원 검증

## 1. 목적

`validate-idea`가 "이 문제 진짜 존재해?"를 검증한다면, 이 스킬은 **"이걸로 사업이 되나?"**를 검증한다.

가설을 측정 가능한 형태로 변환하고, 7개 차원에서 score를 매겨 **go / pivot / no-go** 결정을 강제한다. 막연한 "괜찮을 것 같다"가 통과하지 못하도록 actual signal — 인터뷰, letter-of-intent, conversion data — 만 인정한다.

## 2. 사용 시점 (When to invoke)

- `validate-idea` 통과 후 SaaS / B2B / B2C 출시 결정 직전
- 가격 모델 설계 전 willingness-to-pay 검증
- 투자 자료(deck) 작성 전 TAM 검증
- pivot 결정 시점 — 어느 차원이 깨졌는지 식별
- 경쟁사 등장으로 defensibility 재검증
- 운영 1-3개월 후 unit economics 재계산
- 신규 시장 확장(geo expansion) 전 SAM/SOM 재산정

## 3. 입력 (Inputs)

### 필수
- `validate-idea` 통과 산출물 (problem statement, target user, value hypothesis)
- 현재 가격 모델 가설 (subscription / usage / freemium / one-time / tiered)
- 직접 경쟁사 또는 substitute 인지 여부

### 선택
- 고객 인터뷰 N개 transcript
- landing page 또는 letter-of-intent 결과
- 경쟁사 공개 ARR / 사용자 수
- 산업 표준 churn / CAC 데이터

### 입력이 부족할 때 forcing question
- "TAM의 단위가 뭐야? 사용자 수 / 매출 / 거래 횟수?"
- "buyer는 user와 같아? 다르면 buyer의 동기가 뭐야?"
- "$X에 산다고 말한 게 인터뷰 기반이야 landing 기반이야 survey 기반이야?"
- "CAC를 paid 채널로 모델링했어 organic으로?"
- "5년 후 우리만의 우위는 무엇이야? '더 잘 만든다'는 답이면 다시 생각해봐."

## 4. 핵심 원칙 (Principles)

1. **가설은 측정 가능한 형태로** — "괜찮을 듯", "수요 있을 듯" 금지. 숫자·기간·검증 방법 명시.
2. **TAM은 bottom-up** — "X% 마켓 점유" 가설 금지. comparable companies + customer count로 누적.
3. **구매자 ≠ 사용자** — B2B에서 두드러지지만 B2C에서도 부모/자녀, 회사/직원 분리. 누가 지갑을 여는지 명시.
4. **willingness-to-pay는 actual signal** — survey 응답이 아니라 결제 의사 인터뷰, letter of intent, landing page conversion.
5. **CAC/LTV는 paid 채널 가정 명시** — organic-only 모델은 스케일 가정과 모순. paid 비용 포함.
6. **unit economics negative면 즉시 stop** — 스케일이 적자를 키운다. payback > LTV이면 거래할수록 손해.
7. **defensibility는 first-mover ≠ moat** — network effect / switching cost / brand / IP / scale economy 중 어느 카테고리?
8. **규제 산업이면 규제 리스크가 사업 모델 자체를 무효화** — `review-license-and-ip-risk`와 페어 호출 강제.

## 5. 단계 (Phases)

### Phase 1. Market Sizing (Bottom-Up)
1. ICP(ideal customer profile) 정의 — segment, geography, size, role
2. ICP에 해당하는 잠재 고객 수 (LinkedIn search, public registry, industry report)
3. 고객당 평균 거래 금액 (comparable company ARPU)
4. TAM = 고객 수 × ARPU
5. SAM = TAM 중 진입 가능 (지역/언어/규제)
6. SOM = SAM 중 5년 점유 가능 (실현 가능한 최대)

### Phase 2. Customer & Buyer Mapping
1. user persona — 누가 매일 쓰는가
2. buyer persona — 누가 결제 결정
3. influencer — 결정에 영향을 주는 사람
4. blocker — 거부할 수 있는 사람 (보안팀, 법무팀, IT)
5. buyer motivation — KPI 연결, ROI 회수 시간

### Phase 3. Willingness-to-Pay & Pricing Model
1. pricing model: subscription | usage | one-time | freemium | tiered | usage-based-with-floor
2. anchor price 후보 3개 (low / mid / high)
3. validation method 선택: interview | landing | LOI | smoke-test
4. validation 실행 + 결과 기록 (n명 중 m명이 결제 의사)
5. price elasticity 가설 (10% 가격 인상 시 churn?)

### Phase 4. GTM & Channel Feasibility
1. primary channel: PLG | sales-led | community | partnership | content | paid-ad
2. CAC 가설 per channel
3. conversion funnel: 인지 → 가입 → activation → paid
4. 각 단계 conversion rate 가설 (industry benchmark + 우리 가정)
5. payback period: paid 후 몇 개월에 회수

### Phase 5. Competition & Substitutes
1. direct 경쟁사 — 같은 문제를 같은 방식으로 해결
2. substitute — 다른 방식으로 같은 문제 해결 (Excel, manual, status quo)
3. defensibility category: network effect | switching cost | brand | IP | scale economy
4. 5년 후 우리만의 우위는 무엇인가
5. status quo도 경쟁자 — "현재 어떻게 하고 있나?" 질문

### Phase 6. Unit Economics Simulation
1. CAC = (마케팅 + 영업 비용) / 신규 paid customer 수
2. LTV = ARPU × gross margin × (1 / monthly churn rate)
3. LTV : CAC 목표 3:1 이상
4. payback period 12개월 이하
5. magic number (영업 효율, ARR 증가 / S&M 비용) 0.7 이상

### Phase 7. Aggregate Score & Verdict
각 차원 0-1.0 점수. 가중 평균.

verdict 기준:
- **0.7+**: go
- **0.4-0.7**: pivot — 어느 차원이 약한지 명시 + 재검증 계획 + 30일 deadline
- **<0.4**: no-go — sunk cost 인정. fund/effort를 다른 곳으로.

## 6. 출력 템플릿 (Output Format)

```yaml
viability_score: 0.0-1.0
verdict: go | pivot | no-go
verdict_reasoning: "<2-3문장 핵심 근거>"

critical_assumptions:
  - assumption: "<측정 가능한 가설>"
    test_required: "<검증 방법>"
    test_status: untested | testing | passed | failed
    deadline: "<날짜>"

market:
  tam: "$<숫자>"
  sam: "$<숫자>"
  som: "$<숫자>"
  approach: bottom-up | top-down  # top-down은 reject
  comparable_companies: ["<comp1>", "<comp2>"]

customer:
  user_segment: "<persona>"
  buyer_segment: "<persona>"
  buyer_motivation: "<KPI 연결>"
  influencers: ["<...>"]
  blockers: ["<...>"]

pricing:
  model: subscription | usage | one-time | freemium | tiered
  anchor_low: "$X/mo"
  anchor_mid: "$Y/mo"
  anchor_high: "$Z/mo"
  validation_method: interview | landing | LOI | smoke-test
  validation_result: "<n개 중 m개 결제 의사>"

unit_economics:
  cac: "$..."
  ltv: "$..."
  ltv_cac_ratio: "X:1"
  payback_months: "..."
  gross_margin: "X%"
  churn_monthly: "X%"
  magic_number: "..."

gtm:
  primary_channel: PLG | sales-led | community | partnership | content | paid-ad
  conversion_funnel:
    awareness_to_signup: "X%"
    signup_to_activation: "X%"
    activation_to_paid: "X%"

competition:
  direct: ["<comp>"]
  substitutes: ["<sub>"]
  defensibility: "<category + 설명>"

risks:
  - severity: critical | high | medium
    category: regulatory | business_model | execution | market_timing
    description: "..."
    mitigation: "..."
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `validate-idea` — `Skill` tool로 invoke (문제 검증 후 사업성 검증)
- 페어: `review-license-and-ip-risk` — 규제 산업이면 강제 호출 (규제가 사업 모델 무효화 가능)
- 페어: `validate-advanced-edge-idea` — pivot 결정 시 hidden assumption 재검증
- 다음 단계: `define-product-spec` — go verdict 후 PRD 작성으로 진행

## 8. Anti-patterns

1. **TAM top-down 사용** — "헬스케어는 $4T이고 우리는 0.001%만 잡으면 $40M" — 무의미. bottom-up으로 다시.
2. **Survey 응답을 willingness-to-pay로 사용** — "$50까지 낼 수 있냐"에 73% yes — 결제와 무관. actual signal 필요.
3. **CAC를 organic-only로 계산** — "처음엔 organic으로" — 스케일 단계엔 paid 필수. paid 가정 포함.
4. **Defensibility를 "더 잘 만든다"로 정의** — 경쟁사도 잘 만들 수 있음. structural moat 카테고리 명시.
5. **가설 검증 deadline 없이 오픈** — "언젠가 인터뷰" — 가설은 30일 deadline 강제.
6. **Competition을 "우리는 다르다"로 회피** — substitute(Excel/manual)도 경쟁. status quo가 가장 강한 경쟁자.
7. **Unit economics를 fundraise 후로 미룸** — VC도 본다. seed 단계라도 1년치 모델 필수.
8. **Pivot 결정을 "조금만 더 해보고"로 미룸** — 0.4 미만이면 sunk cost. 다른 idea로.
