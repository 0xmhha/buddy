---
name: review-terms-policy-readiness
description: "상용 출시 전 ToS(Terms of Service), Privacy Policy, AUP(Acceptable Use Policy), Refund Policy, Cookie Policy, DPA(Data Processing Agreement), SLA(Service Level Agreement) 7종 문서의 준비 상태와 일관성을 검토. 트리거: '이용약관 검토' / '개인정보처리방침 readiness' / '환불 정책 봐줘' / 'ToS 작성' / '쿠키 정책' / 'DPA 필요해?' / 'SLA 정의'. 입력: 제품 spec, pricing, 데이터 처리 방식, 대상 시장. 출력: 7종 문서 readiness checklist + gap list + 작성 우선순위. 흐름: review-pricing-and-gtm/review-privacy-data-risk → review-terms-policy-readiness → define-product-spec/write-changelog."
type: skill
---

# Review Terms and Policy Readiness — 7종 문서 출시 준비도 점검

## 1. 목적

상용 출시 전 **법적 문서 7종**의 준비 상태와 상호 일관성을 검토한다. 이 스킬은 변호사 자문을 대체하지 않는다. **전문가 검토 필요한 쟁점을 빠뜨리지 않게 checklist + evidence package**를 만든다.

7종 문서:
1. **ToS (Terms of Service)**: 이용약관 — 계약 핵심
2. **Privacy Policy**: 개인정보처리방침 — GDPR / PIPA / CCPA 의무
3. **AUP (Acceptable Use Policy)**: 사용 제한 — 악용 방지
4. **Refund Policy**: 환불 정책 — 소비자보호법 영향
5. **Cookie Policy**: 쿠키 정책 — GDPR / ePrivacy
6. **DPA (Data Processing Agreement)**: 사업자 간 데이터 처리 합의 — B2B SaaS
7. **SLA (Service Level Agreement)**: uptime / 응답 / 환불 보증 — Pro/Enterprise tier

각 문서에 대해 readiness 5단계: missing / draft / lawyer-reviewed / published / signed-by-users.

## 2. 사용 시점 (When to invoke)

- 첫 상용 출시 전 (free → paid 전환 포함)
- 신규 시장 진입 (US → EU / EU → KR — 규제 다름)
- 신규 tier (Free → Enterprise — DPA 필요해짐)
- 결제 / 환불 / 구독 모델 변경
- 데이터 처리 방식 변경 (analytics 추가, third-party 통합)
- 인수합병 / fundraise due diligence
- 고객 / 파트너가 ToS / DPA 요청

## 3. 입력 (Inputs)

### 필수
- 제품 spec / 기능 목록
- pricing / refund 모델
- 데이터 처리 방식 (`review-privacy-data-risk` 출력)
- 대상 시장 (regulatory frame 결정)
- 결제 provider (Stripe / Toss / 카카오 등)

### 선택
- 기존 ToS / Privacy 문서
- 회사 법무팀 가이드
- 경쟁사 문서 (positioning, not copy)

### 입력 부족 시 forcing question
- "B2B 고객이 DPA 요구할 때 줄 수 있는 template 있어?"
- "환불 정책이 한국 소비자보호법(7일 청약철회) 충족해?"
- "SLA 위반 시 service credit 정책 명시? 그냥 '노력하겠다'면 약함."
- "ToS 변경 시 기존 사용자 통보 정책 있어? 30일 전 명시 필요한 지역 있음."

## 4. 핵심 원칙 (Principles)

1. **법무 검토는 이 스킬 후** — 이 스킬은 readiness check. 실제 출시 전 변호사.
2. **시장별 최소 준수 확인** — KR (전자상거래법, 소비자보호법, PIPA), EU (GDPR, ePrivacy, Consumer Rights Directive), US (FTC, state laws), JP (APPI), CN (PIPL, e-commerce law).
3. **문서 간 일관성** — ToS와 Privacy의 data 처리 명시 일치, Refund와 SLA의 service credit 일치.
4. **Plain language 우선** — 법률 용어 + plain language summary. EU Consumer Rights는 plain language 의무.
5. **Versioning + audit log** — 변경 이력 보존. 사용자 동의 시점의 버전 명시.
6. **Click-wrap > browse-wrap** — 명시 동의 우선. browse-wrap은 enforceability 약함.
7. **Refund 정책은 segment-specific** — B2C는 청약철회 7일. B2B는 prorated.
8. **DPA template 미리 준비** — B2B 고객 요구 시 즉시 제공. 협상 지연 방지.

## 5. 7종 문서 점검 항목

### 1) ToS (Terms of Service)
필수:
- 서비스 정의 / scope
- 계정 (등록, 자격, 책임)
- 사용자 의무 / 금지 행위 (AUP 참조)
- 지적재산권 (회사 IP, user content license)
- 결제 / 환불 (Refund Policy 참조)
- 약관 변경 (통보 방법, 효력 발생)
- limitation of liability / disclaimer
- 분쟁 해결 / 준거법 / 관할
- termination

### 2) Privacy Policy
필수 (GDPR Art. 13-14 기준):
- controller 정보 + DPO 연락처
- 처리 목적 + lawful basis
- 수집 데이터 카테고리
- recipient (third-party processor)
- retention period
- data subject rights (열람/삭제/이전/이의)
- cross-border 전송 (SCC / adequacy)
- automated decision-making (Art. 22)
- supervisory authority 신고권

### 3) AUP (Acceptable Use Policy)
- 금지 use case (illegal, harassment, malware, spam)
- AI 특화 (deepfake, fake content, manipulation)
- enforcement (warning → suspension → termination)
- abuse 신고 채널

### 4) Refund Policy
- KR 청약철회 7일 (전자상거래법 17조)
- EU Consumer Rights Directive (14일 cooling-off)
- B2B prorated 또는 no-refund (계약 명시)
- subscription cancel timing
- service credit (SLA 위반 시)

### 5) Cookie Policy
- 쿠키 분류 (necessary / preferences / analytics / marketing)
- 동의 메커니즘 (banner, granular)
- third-party cookie 명시
- 거부 / 변경 방법
- ePrivacy 의무 (EU)

### 6) DPA (Data Processing Agreement)
- controller / processor 역할 정의
- 처리 instruction
- security measures (Art. 32)
- sub-processor 통보 / 동의
- breach notification (controller 통보 SLA)
- data subject rights 협력
- audit 권리
- 종료 시 데이터 처리 (delete or return)
- SCC 첨부 (cross-border 시)

### 7) SLA (Service Level Agreement)
- uptime 목표 (99.9% / 99.95% / 99.99%)
- 측정 방법 (period, exclusion)
- service credit (위반 시)
- support response time (Critical / High / Medium / Low)
- reporting (status page, monthly report)
- exclusion (force majeure, scheduled maintenance)

## 6. 단계 (Phases)

### Phase 1. Document Inventory
7종 각각 현재 status: missing / draft / lawyer-reviewed / published / signed-by-users

### Phase 2. Market Mapping
대상 시장 → 적용 규제:
- KR → 전자상거래법 / 소비자보호법 / PIPA
- EU → GDPR / ePrivacy / CRD / DSA
- US → FTC / state (CCPA, NY SHIELD, etc.)
- JP → APPI
- CN → PIPL / e-commerce

### Phase 3. Per-Document Checklist
각 문서 필수 항목 충족 여부.

### Phase 4. Cross-Document Consistency
- ToS data 처리 ↔ Privacy 데이터 처리
- Refund ↔ SLA service credit
- AUP ↔ ToS termination clause
- Cookie ↔ Privacy

### Phase 5. Update / Notification Workflow
- 변경 통보 메커니즘 (email, in-app, banner)
- 사전 통보 기간 (30일 일반, 30일+ 특정 지역)
- 동의 갱신 필요 여부
- versioning + audit log

### Phase 6. Gap List + Priority
missing / inconsistent 항목 → 우선순위 (출시 전 vs 출시 후 보완)

## 7. 출력 템플릿 (Output Format)

```yaml
document_readiness:
  - doc: ToS
    status: draft
    coverage: 60%
    missing: ["limitation of liability", "AI use clause"]
    lawyer_reviewed: no
    target_publish: "<date>"
  - doc: Privacy Policy
    status: draft
    gdpr_compliance: partial
    missing: ["DPO contact", "Art. 22 automated decision"]
  - doc: AUP
    status: missing
    priority: high
  - doc: Refund Policy
    status: draft
    kr_consumer_law: not_yet
    eu_crd: not_yet
  - doc: Cookie Policy
    status: published
    granular_consent: yes
  - doc: DPA
    status: missing
    template_needed: yes
    target_b2b_tier: Team+
  - doc: SLA
    status: draft
    uptime_target: "99.9%"
    service_credit: not_specified

market_compliance:
  KR:
    e_commerce_law: partial
    consumer_protection_law: missing  # 청약철회 7일
    pipa: ready
  EU:
    gdpr: partial
    e_privacy: ready
    crd: missing
    dsa: not_applicable
  US:
    ccpa: ready
    coppa: not_applicable
    state_laws_checked: ["CA", "NY", "VA"]

cross_document_inconsistency:
  - issue: "ToS data retention 1 year vs Privacy 30 days"
    severity: high
  - issue: "AUP AI 금지 vs ToS 사용 권한"
    severity: medium

update_workflow:
  notification_method: email + in_app
  advance_notice_days: 30
  consent_renewal_required: yes  # major change
  versioning: semantic_versioning
  audit_log: yes

gap_list:
  pre_launch_blockers:
    - doc: AUP
      action: draft + lawyer review
      eta: "2 weeks"
    - doc: Refund Policy
      action: KR 청약철회 7일 명시
      eta: "1 week"
  post_launch_improvements:
    - doc: DPA
      action: B2B template 작성 (Team tier 출시 시)
    - doc: SLA
      action: Pro tier 출시 시 service credit 정책

verdict: ready | needs-remediation | blocked
```

## 8. 자매 스킬 (Sibling Skills)

- 앞 단계: `review-pricing-and-gtm` — `Skill` tool로 invoke (refund / SLA 가격 정렬)
- 페어: `review-privacy-data-risk` — Privacy Policy 데이터 처리 명시
- 페어: `review-license-and-ip-risk` — IP 조항 일관성
- 페어: `review-ai-safety-liability` — AI 사용 ToS / AUP 조항
- 다음 단계: `define-product-spec` — 약관 요구사항 PRD에 반영
- 다음 단계: `write-changelog` — 출시 ship gate

## 9. Anti-patterns

1. **"법무가 알아서 한다"** — 출시 임박해서 발견. 사전 readiness check 필수.
2. **경쟁사 ToS 복사** — 우리 제품과 다름. 자체 작성 + 변호사 검토.
3. **Plain language 무시** — 사용자 동의 의미 약화. EU CRD 위반 가능.
4. **Versioning 없이 변경** — 사용자 동의 시점 버전 모름. audit log 필수.
5. **DPA 없이 B2B 출시** — 고객 요구 시 협상 지연. 미리 template.
6. **Refund 정책 vague** — "사례별 결정"은 KR 7일 위반 가능. 명시.
7. **SLA를 "best effort"로** — Pro/Enterprise tier 신뢰 약화. service credit 명시.
8. **Cross-document 불일치** — ToS와 Privacy data 처리 다름. consistency check 강제.
