---
name: design-billing-system
description: "SaaS 결제 시스템 설계 — Stripe/Toss + point wallet + subscription tier + usage metering + invoice + dunning + revenue share + tax. 결제 보안, idempotency, webhook 정합성, refund 정책 표준화. 트리거: '결제 시스템 설계' / 'Stripe 통합' / 'point wallet 만들자' / 'subscription 모델' / '환불 정책' / 'usage metering' / 'revenue share 정산' / 'dunning flow'. 입력: pricing tier, 시장(KR/EU/US), 결제 provider, multi-tenant 요구. 출력: billing module spec + payment flow + webhook handler + reconciliation + tax. 흐름: review-pricing-and-gtm → design-billing-system → build-with-tdd."
type: skill
---

# Design Billing System — 결제 / 구독 / 정산 9 모듈 설계

## 1. 목적

SaaS의 **결제 backbone**을 설계한다. payment provider 통합부터 multi-tenant point wallet, usage metering, dunning, refund, revenue share, tax까지 9 모듈을 표준화.

이 스킬은 `feature-management-saas-mcp.md` Billing & Licensing Module을 직접 구현 가능한 수준으로 변환한다. point + subscription 결합 모델 (`feature-management-saas-mcp.md` "Point와 Subscription 결제 모델" 섹션) 우선 지원.

## 2. 사용 시점 (When to invoke)

- 첫 paid tier 출시 (free → paid 전환)
- subscription 추가 (one-time → recurring)
- multi-tenant SaaS의 organization billing 분리
- usage-based / hybrid 가격 모델 도입
- point wallet 시스템 구축 (paid artifact download 등)
- multi-region 결제 (KR + US + EU)
- B2B Enterprise contract billing
- reseller / affiliate revenue share

## 3. 입력 (Inputs)

### 필수
- pricing tier 구조 (`review-pricing-and-gtm` 출력)
- 결제 provider (Stripe / Toss / 카카오페이 / Adyen / Paddle)
- 시장 (KR / EU / US / JP / 기타)
- multi-tenant 요구 (organization / team / individual)
- usage metering 단위 (있으면)

### 선택
- 기존 결제 시스템 (migration 시)
- tax / VAT 정책
- accounting / ERP 통합 요구

### 입력 부족 시 forcing question
- "KR + EU 동시 출시면 VAT MOSS / 부가세 처리 다름. 어떤 provider 쓸 거야?"
- "환불 정책이 청약철회 7일이야 prorated refund야?"
- "annual prepay할 때 revenue recognition 어떻게? deferred revenue ledger?"
- "point wallet에 expiration 정책? 미사용 1년 후 소멸?"

## 4. 핵심 원칙 (Principles)

1. **결제 정보는 provider에 보관** — PCI-DSS 회피. token만 보관.
2. **Idempotency key 모든 mutating call** — webhook retry, double-charge 방지.
3. **Webhook 정합성 검증** — signature verify + idempotency. 중복 처리 차단.
4. **Reconciliation 자동화** — 매일 provider ledger vs DB 비교. drift detection.
5. **Refund는 audit trail** — 누가 언제 왜 refund. 분쟁 대응.
6. **Tax는 jurisdiction별** — VAT (EU), GST (KR/JP), sales tax (US state).
7. **Dunning는 grace period** — 결제 실패 즉시 cut off 금지. retry 3회 + 7일 grace.
8. **Point wallet은 ledger 기반** — append-only transaction log. 잔액은 ledger sum.

## 5. 9 모듈 설계

### Module 1. Customer & Subscription
- customer entity (user / organization / team)
- subscription entity (tier, period, start/end, auto-renew)
- seat management (B2B)
- 상태: active / trialing / past_due / canceled / unpaid

### Module 2. Pricing Catalog
- product / plan 정의 (provider sync)
- tier 가격 (monthly / annual)
- 통화 (KRW / USD / EUR / JPY)
- coupon / discount 코드
- enterprise contract (custom price)

### Module 3. Payment Method
- provider token 보관 (Stripe customer ID, payment method ID)
- 다중 결제수단 (card primary / fallback)
- 만료 알림 (60일 / 30일 / 7일 전)
- 3DS / SCA 처리 (EU)

### Module 4. Subscription Lifecycle
- 가입 (subscribe + first payment)
- 갱신 (auto-renew + payment)
- upgrade / downgrade (prorated)
- cancel (period end / immediate)
- pause (temporary)
- reactivate

### Module 5. Usage Metering & Quota
- usage event 수집 (event-based)
- aggregation (per period)
- quota check (real-time / batch)
- overage handling (block / soft / paid)
- reset (period boundary)

### Module 6. Point Wallet (paid artifact 결제)
- ledger entries (credit / debit / expire / refund)
- balance = ledger sum
- 충전 (top-up + payment)
- subscription monthly credit
- expiration policy
- reservation (download 전 hold)
- confirmation (download 성공 시 차감)

### Module 7. Invoice & Receipt
- invoice generation (period 마감)
- line items (subscription + usage + tax)
- PDF / 전자세금계산서 (KR)
- delivery (email / portal)
- past invoice 조회

### Module 8. Dunning & Retry
- payment failed → 3 retry (1 / 3 / 7일 후)
- email 알림 (each retry + final)
- 7일 grace period (downgrade / suspend)
- recovery (payment update → resume)

### Module 9. Refund & Dispute
- refund 정책 (자동 / 수동 / no-refund)
- partial refund (prorated)
- dispute / chargeback handling
- audit trail

## 6. 단계 (Phases)

### Phase 1. Provider Selection
- Stripe (글로벌, US/EU 강함)
- Toss (KR 세금계산서)
- 카카오페이 (KR retail)
- Paddle (Merchant of Record, tax 처리)
- Adyen (enterprise)

### Phase 2. Data Model
customer / subscription / payment_method / invoice / usage_event / point_ledger / refund

### Phase 3. Webhook Handler
- event 수신 (signature verify)
- idempotency check
- DB update
- side effect (email, slack)

### Phase 4. Reconciliation
- 매일 provider API → DB 비교
- diff report
- alert on drift

### Phase 5. Multi-Tenant Scoping
- organization → owner_user
- subscription → organization
- billing portal → admin role

### Phase 6. Tax & Compliance
- VAT MOSS (EU)
- 세금계산서 (KR — Toss, 우후세무)
- sales tax (US — TaxJar, Avalara)
- invoice 보관 5-10년 (jurisdiction)

## 7. 출력 템플릿 (Output Format)

```yaml
billing_system:
  provider:
    primary: stripe
    secondary: toss  # KR
    merchant_of_record: yes  # Paddle 사용 시

data_model:
  customer:
    id, owner_user_id, organization_id, stripe_customer_id, country, vat_id, created_at
  subscription:
    id, customer_id, plan_id, status, period_start, period_end, auto_renew, seat_count
  payment_method:
    id, customer_id, provider_token, brand, last4, exp_month, exp_year, is_default
  invoice:
    id, customer_id, subscription_id, period_start, period_end, amount, tax, currency, status, pdf_url
  usage_event:
    id, customer_id, event_type, quantity, occurred_at
  point_ledger:
    id, wallet_id, type (credit/debit/expire/refund), amount, reference, balance_after, created_at
  refund:
    id, invoice_id, amount, reason, refunded_by, refunded_at

webhook_handler:
  endpoint: "/webhooks/stripe"
  signature_verification: required
  idempotency: by event.id
  events_handled:
    - invoice.paid
    - invoice.payment_failed
    - customer.subscription.updated
    - customer.subscription.deleted
    - payment_method.attached

subscription_lifecycle:
  trial:
    duration: "14 days"
    require_payment_method: yes  # avoid card-required friction trade-off
  upgrade:
    proration: yes
    immediate_charge: yes
  downgrade:
    proration: credit_to_next_invoice
    effective_date: period_end
  cancel:
    default: period_end
    option: immediate
    save_offer: "1 month free"

usage_metering:
  events:
    - name: api_call
      aggregation: count
      quota_period: month
    - name: storage_gb
      aggregation: max
      quota_period: month
  overage_policy: paid  # block | soft | paid
  reset: period_boundary

point_wallet:
  monthly_credit:
    pro: 1000
    team: 5000
  expiration: "12 months from credit"
  reservation_ttl: "15 minutes"
  refund_to_wallet: yes

dunning:
  retry_schedule: [1, 3, 7]  # days
  grace_period_days: 7
  notification_channels: [email, in_app]
  post_grace_action: suspend  # downgrade | suspend | terminate

refund_policy:
  kr_consumer_law:
    cooling_off: 7  # days
    full_refund: yes
  eu_crd:
    cooling_off: 14
  b2b: prorated
  point_wallet: refundable_to_method

tax:
  kr:
    type: VAT
    rate: 10
    invoice_format: 전자세금계산서
    provider: Toss / 우후세무
  eu:
    type: VAT_MOSS
    rate: per_country
    provider: Stripe Tax | Paddle
  us:
    type: sales_tax
    provider: TaxJar | Avalara

reconciliation:
  frequency: daily
  source: provider_api
  target: db
  diff_alert: slack + email
  manual_review_threshold: ">$100 drift"

multi_tenant:
  scope: organization
  billing_admin_role: org_admin
  seat_management:
    add_seat: prorated_charge
    remove_seat: credit_next_invoice
```

## 8. 자매 스킬 (Sibling Skills)

- 앞 단계: `review-pricing-and-gtm` — `Skill` tool로 invoke (tier 결정 후)
- 페어: `audit-security` — payment data 보호, PCI-DSS scope 회피
- 페어: `review-privacy-data-risk` — 결제 PII 처리
- 페어: `review-terms-policy-readiness` — Refund Policy 정렬
- 다음 단계: `setup-quality-gates` — webhook idempotency 자동 테스트
- 다음 단계: `build-with-tdd` — billing TDD (race condition / refund 시나리오)

## 9. Anti-patterns

1. **결제 정보 직접 보관** — PCI-DSS 위반. token만.
2. **Webhook signature 검증 누락** — webhook spoofing 가능. 강제.
3. **Idempotency 없이 retry** — double charge / double credit. key 강제.
4. **Reconciliation 수동** — drift 발견 늦음. daily 자동.
5. **Dunning 즉시 cut off** — 카드 만료 / 일시 오류로 churn. 3 retry + 7일 grace.
6. **Tax 단일 정책** — 시장별 다름. VAT / GST / sales tax 분기.
7. **Point ledger snapshot 기반 잔액** — race condition. ledger sum.
8. **Refund audit log 없음** — 분쟁 시 증거 없음. who / when / why 기록.
