---
name: review-pricing-and-gtm
description: "pricing model 설계와 GTM(Go-To-Market) channel 전략을 평가. tier 구조, willingness-to-pay 검증, channel-CAC fit, conversion funnel, expansion revenue, churn 방어 메커니즘 검토. 트리거: 'pricing tier 짜자' / '가격 모델 검토' / 'GTM 전략 봐줘' / 'channel 어디로 갈까' / 'conversion funnel 평가' / 'pricing 검증' / '확장 매출 설계'. 입력: assess-business-viability 출력 + 가격 가설 + 경쟁사 가격. 출력: pricing tier set + GTM playbook + funnel target + churn defense. 흐름: assess-business-viability → review-pricing-and-gtm → define-product-spec."
type: skill
---

# Review Pricing and GTM — 가격 모델과 출시 전략 검증

## 1. 목적

`assess-business-viability`가 willingness-to-pay 가설만 검증한다면, 이 스킬은 그 가설을 **상용 출시 가능한 pricing tier 구조 + GTM playbook**으로 변환한다.

핵심 질문 5개에 답:
1. tier 구조는 segment / job-to-be-done과 정렬되었나?
2. anchor 가격은 actual signal로 검증되었나?
3. primary channel은 CAC와 fit하나?
4. conversion funnel은 industry benchmark 대비 합리적인가?
5. expansion revenue (upsell / cross-sell) 메커니즘이 명시되었나?

이 스킬을 통과한 pricing은 **가격 인상 시 churn rate 변화**를 시뮬레이션 가능한 구조를 가진다.

## 2. 사용 시점 (When to invoke)

- `assess-business-viability` go verdict 후 출시 가격 확정
- 기존 pricing 인상 / 인하 결정 전
- 신규 tier 추가 시 (Pro → Team → Enterprise)
- 신규 channel 진입 (PLG → sales-led 전환)
- churn rate 상승 시 pricing-product fit 재검증
- 경쟁사 가격 변경 대응
- usage-based → subscription 전환 같은 모델 변경

## 3. 입력 (Inputs)

### 필수
- `assess-business-viability` 출력 (validation_result, anchor prices, CAC/LTV)
- 직접 경쟁사 가격 (3-5개)
- 현재 사용자 / paid customer 수
- 채널별 acquisition 비용 (있으면)

### 선택
- churn rate (월/연)
- expansion revenue 데이터 (NRR / GRR)
- conversion funnel 실측 데이터

### 입력 부족 시 forcing question
- "tier 3개 모두 anchor price를 actual signal로 검증했어?"
- "primary channel CAC가 pricing의 첫 달 매출보다 작아? payback 가정?"
- "free tier가 paid tier conversion을 막지 않아? freemium trap 가능성?"
- "churn 1% 차이가 LTV에 얼마 영향? 시뮬 해봤어?"

## 4. 핵심 원칙 (Principles)

1. **Tier는 segment / JTBD 기반** — 임의 feature gating 금지. 누가 어떤 job 위해 어느 tier 사는지 명확.
2. **Anchor 가격은 actual signal** — competitor median 모방 금지. 인터뷰 / LOI / landing conversion으로 검증.
3. **Primary channel은 CAC fit** — PLG에 enterprise 가격 / sales-led에 자가가입 — 미스매치.
4. **Free tier는 conversion 가설** — "그냥 lead generation"은 약함. free → paid 전환 트리거 명시.
5. **Expansion > Acquisition** — NRR 110%+이면 acquisition 약해도 성장. expansion 메커니즘 (upsell tier, seat, usage) 강제.
6. **Annual discount는 cash flow / churn 무기** — 일반적 20% off. annual prepay 시 churn 30-50% 감소.
7. **Pricing page는 funnel** — landing → 가격 페이지 → signup → activation → paid. 각 단계 conversion 추적.
8. **가격 변경은 grandfathering 정책** — 기존 사용자 유지. announcement 28일+ 전.

## 5. 단계 (Phases)

### Phase 1. Tier Architecture
1. 사용자 segment 식별 (개인 / 팀 / 기업)
2. 각 segment의 JTBD (job-to-be-done)
3. tier별 value metric (seat, usage, feature, support)
4. 일반 패턴: free / starter / pro / team / enterprise
5. 각 tier 가격 anchor 후보

### Phase 2. Willingness-to-Pay Validation
각 tier에 대해 actual signal:
- 인터뷰 (n명 중 m명 결제 의사)
- LOI (signed)
- landing page conversion (visit → trial → paid)
- competitor benchmark (positioning, not copy)

가설 깨지면 tier 재구성 또는 가격 조정.

### Phase 3. Channel Strategy
primary channel 후보:
- **PLG (product-led growth)**: low-friction signup, time-to-value 짧게, viral hook
- **Sales-led**: enterprise, complex deal, high ACV
- **Community**: developer / creator 대상, content / OSS / event
- **Partnership**: integration / channel partner / reseller
- **Content**: SEO / blog / podcast — 시간 오래 걸림
- **Paid Ad**: B2C, growth 가속, churn 위험

각 channel에 대해 CAC 가설 + payback period.

### Phase 4. Funnel Targets
conversion 단계별 목표:
- visitor → signup: ~2-5% (B2C SaaS), ~10-20% (B2B SaaS landing)
- signup → activation: ~30-50%
- activation → paid: ~5-15% (freemium), ~30-50% (trial-to-paid)
- paid → expansion: NRR 100-130%

industry benchmark 대비 차이 식별.

### Phase 5. Expansion Revenue
NRR (Net Revenue Retention) 메커니즘:
- usage 증가 → 자동 tier upgrade
- seat 추가 → linear expansion
- add-on 모듈 (feature, support, integration)
- annual prepay → cash flow + churn 보호

목표: NRR 110%+ (SaaS world-class)

### Phase 6. Churn Defense
- onboarding optimization (activation rate)
- product engagement signal (low → outreach)
- annual contract migration
- usage 감소 trigger (90일+ inactive)
- cancel flow (downgrade option, save offer)

### Phase 7. Pricing Page Design
- tier 비교표 (3-5 column)
- "Most popular" highlighting (Pro tier)
- annual / monthly toggle
- usage calculator (usage-based 시)
- enterprise CTA ("Contact Sales")
- FAQ (refund, cancel, pricing change)

## 6. 출력 템플릿 (Output Format)

```yaml
pricing_architecture:
  tiers:
    - name: Free
      target_segment: "<해커, 개인 dev>"
      jtbd: "<...>"
      price_monthly: "$0"
      value_metric: "limited usage"
      conversion_to_paid_target: "5-10%"
      features: ["basic 기능 N"]
      gates: ["<paid 전환 트리거>"]
    - name: Pro
      target_segment: "<개인 paid 사용자>"
      jtbd: "<...>"
      price_monthly: "$X"
      price_annual: "$Y/yr (X% off)"
      value_metric: "seat or usage"
      validation: "interview 12명 중 8명 결제 의사"
    - name: Team
      target_segment: "<5-50명 팀>"
      ...
    - name: Enterprise
      pricing: "Contact Sales"
      target_acv: "$X-Y"

channel_strategy:
  primary: PLG | sales-led | community | partnership | content | paid-ad
  primary_reason: "<...>"
  cac_target: "$X"
  payback_target: "<12 months"
  secondary: ["<channel>"]

funnel_targets:
  visitor_to_signup: "X%"
  signup_to_activation: "X%"
  activation_to_paid: "X%"
  paid_to_expansion: "NRR X%"

expansion_revenue:
  mechanisms:
    - type: seat-expansion
      target_nrr_contribution: "X%"
    - type: usage-tier-upgrade
      target_nrr_contribution: "X%"
    - type: add-on
      products: ["<...>"]

churn_defense:
  onboarding:
    activation_target: "X% within 7 days"
  engagement_signal:
    inactive_threshold: "90 days"
    intervention: "in-app message + email"
  annual_migration:
    discount: "20% off"
    target_annual_share: "X%"
  cancel_flow:
    downgrade_option: yes
    save_offer: "1-month free"

pricing_page_outline:
  tier_columns: 3-5
  toggle: "monthly | annual"
  enterprise_cta: "Contact Sales"
  faq_topics: ["refund", "cancel", "pricing change", "seat add"]

risks:
  - risk: "<freemium trap>"
    severity: high
    mitigation: "<...>"
  - risk: "<channel-CAC mismatch>"
    severity: medium
    mitigation: "<...>"
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `assess-business-viability` — `Skill` tool로 invoke
- 페어: `review-license-and-ip-risk` — paid tier 출시 전 라이선스 점검
- 페어: `review-terms-policy-readiness` — 결제 약관 / 환불 정책
- 다음 단계: `define-product-spec` — pricing 결정을 PRD에 반영

## 8. Anti-patterns

1. **Competitor median 모방** — Notion이 $8이라고 우리도 $8 — segment / JTBD 무관 가격. 약함.
2. **Tier feature gating 임의** — 어느 segment가 어느 tier 사는지 모름. JTBD 분석 강제.
3. **Free tier가 conversion trap** — 너무 generous하면 paid로 안 옴. usage / seat / feature limit 명시.
4. **Channel-CAC mismatch** — PLG에 enterprise 가격 / sales-led에 self-serve. fit 검증.
5. **Expansion 메커니즘 없음** — NRR 100% 미만이면 churn에 항상 진다. expansion 강제.
6. **Pricing page 정보 과잉** — 30 row 비교표. "Most popular" 1개로 결정 유도.
7. **가격 변경 통보 없이** — grandfathering 안 하면 churn 폭발. 28일+ 전 announcement.
8. **Annual discount 없음** — cash flow + churn 보호 무기. 일반 20% off.
