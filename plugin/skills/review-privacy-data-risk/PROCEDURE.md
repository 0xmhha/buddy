---
name: review-privacy-data-risk
description: "PII / personal data / sensitive data의 수집·저장·처리·전송·삭제 lifecycle을 GDPR/PIPL/PIPA/HIPAA/COPPA 등 규제 frame으로 검토. data flow map + DPIA + breach response 산출. 트리거: '개인정보 검토' / 'PII 처리 봐줘' / 'GDPR 영향 평가' / 'data retention 정책' / '데이터 삭제 권리' / 'DPIA 필요해?' / 'cross-border data 전송'. 입력: 데이터 모델, data flow, 사용자 지역, 도메인(health/finance/minor). 출력: data inventory + DPIA + breach response + retention policy. 흐름: assess-business-viability → review-privacy-data-risk → define-product-spec/audit-security."
type: skill
---

# Review Privacy and Data Risk — 데이터 lifecycle 6단계 검증

## 1. 목적

`audit-security`가 **공격으로부터 보호**한다면, 이 스킬은 **사용자 데이터의 권리**를 본다. 두 영역은 독립 — 보안 통과해도 GDPR 위반 시 매출 4% 또는 €20M 과징금.

검토 대상: PII (personal identifiable information), sensitive data (health, finance, biometric, minor), behavioral data (cookie, analytics).

핵심 산출물 4개:
1. **Data inventory + flow map** — 어떤 데이터가 어디서 들어와 어디로 가고 어디 저장되나
2. **DPIA (Data Protection Impact Assessment)** — high-risk 처리 시 필수
3. **Retention & deletion policy** — 보유 기간 + 자동 삭제 + 권리 행사
4. **Breach response plan** — 인지 → 통보 → 보고 (GDPR 72시간)

## 2. 사용 시점 (When to invoke)

- `assess-business-viability` 후 EU/CN/KR 사용자 대상 출시 전
- 신규 데이터 필드 수집 결정 시 (특히 sensitive)
- 외부 service / SDK / analytics 통합 (data processor 추가)
- cross-border data 전송 결정 (US → EU / KR → US)
- GDPR / PIPL / PIPA 감사 대응
- data breach 발생 시
- data subject request (열람 / 삭제 / 이전) 대응 정책 수립

## 3. 입력 (Inputs)

### 필수
- 데이터 모델 (table / collection / field 목록)
- data flow (수집 → 처리 → 저장 → 삭제)
- 사용자 지역 (regulatory frame 결정)
- 도메인 특성 (health / finance / minor / general)
- 외부 data processor 목록 (analytics, CDN, payment, ML)

### 선택
- 기존 privacy policy
- DPO (Data Protection Officer) 정보
- 회사 data classification 정책

### 입력 부족 시 forcing question
- "EU 사용자 받아? 1명이라도 있으면 GDPR 적용."
- "어떤 third-party SDK 써? GA / Mixpanel / Sentry — 모두 data processor."
- "회원 탈퇴 시 데이터 며칠 안에 삭제? 자동? 수동?"
- "데이터 백업에서도 삭제돼? backup retention은 별도 정책 필요."

## 4. 핵심 원칙 (Principles)

1. **Minimum necessary 원칙** — 수집은 필요한 만큼만. "나중에 쓸지 몰라" 금지.
2. **Purpose limitation** — 수집 목적 외 사용 금지. 새 목적엔 새 동의.
3. **Lawful basis 명시** — consent / contract / legal obligation / vital interest / public task / legitimate interest 중 하나.
4. **Cross-border 전송 시 SCC / DPA** — EU → 비EU는 SCC (Standard Contractual Clauses) 또는 adequacy decision.
5. **Data subject rights 자동화** — 열람 / 정정 / 삭제 / 이전 / 처리 정지 — 1개월 내 응답.
6. **Backup도 동일 정책** — 운영 DB만 삭제 + backup 7년 보관 = 위반. backup retention 별도 명시.
7. **Breach 72시간 통보 (GDPR)** — high-risk면 사용자도 직접 통보.
8. **Sensitive data는 explicit consent** — health / biometric / political / sexual / religious / minor.

## 5. 단계 (Phases)

### Phase 1. Data Inventory
1. 모든 데이터 필드 목록
2. 분류: PII / sensitive / behavioral / non-personal
3. 출처: 사용자 입력 / 자동 수집 / 추론 / third-party
4. 저장 위치: DB / file / cache / log / backup / analytics

### Phase 2. Data Flow Map
입력 → 처리 → 저장 → 전송 → 삭제 5단계.
각 단계에:
- actor (시스템 / 인간 / 외부)
- purpose
- legal basis
- retention
- protection (encryption at rest, encryption in transit, access control)

cross-border 전송 별도 표시 (red flag).

### Phase 3. Regulatory Mapping
적용 규제 식별:
- **GDPR** (EU 거주자): consent, DSR, DPIA, DPO, breach notification
- **PIPL** (China): consent, cross-border transfer, sensitive data
- **PIPA** (Korea): collection consent, retention, third-party 제공 동의
- **HIPAA** (US health): PHI, BAA, security/privacy/breach rules
- **COPPA** (US <13세): parental consent, data minimization
- **CCPA/CPRA** (California): right to know/delete/opt-out
- **LGPD** (Brazil): controller/processor, lawful basis

### Phase 4. DPIA (high-risk 시)
GDPR Art. 35 high-risk 시 DPIA 필수:
- automated decision-making (significant effect)
- large-scale sensitive data
- public area systematic monitoring
- vulnerable subjects (children, employees)

DPIA 항목:
- processing description
- necessity / proportionality
- risks to data subjects
- mitigations
- DPO consultation

### Phase 5. Retention & Deletion Policy
각 데이터 카테고리:
- 보유 기간 (active / archive / backup)
- trigger (회원 탈퇴 / 7년 / 법적 의무)
- 삭제 방법 (soft delete + 30일 후 hard delete 일반)
- backup 정책 (full delete vs anonymize)
- audit log 보존 (별도 — 보안 목적)

### Phase 6. Data Subject Rights (DSR) Workflow
- access (열람): 1개월 내, 무료
- rectification (정정)
- erasure (삭제, "right to be forgotten")
- portability (machine-readable export)
- restriction (처리 정지)
- objection
- automated decision objection

각 rights에 대해:
- 자동화 워크플로우 (admin panel / API)
- 검증 (본인 확인)
- 응답 기간 (1개월 → 2개월 연장 가능)

### Phase 7. Breach Response Plan
인지 → 분류 → 통보 → 보고:
- 인지 후 72시간 내 supervisory authority 통보 (GDPR)
- high-risk면 사용자 직접 통보 (지체 없이)
- breach log 보존 (감사 대비)
- post-mortem + remediation

## 6. 출력 템플릿 (Output Format)

```yaml
data_inventory:
  - field: "users.email"
    classification: PII
    source: user_input
    storage: ["postgres.users", "elasticsearch", "backup.s3"]
    legal_basis: contract
    retention: "until account deletion + 30d soft delete"
  - field: "users.health_condition"
    classification: sensitive
    source: user_input
    legal_basis: explicit_consent
    retention: "...  + special protection"
  - field: "events.click_history"
    classification: behavioral
    source: auto_collected
    legal_basis: legitimate_interest

data_flow_map:
  - stage: collection
    source: signup form
    fields: [email, name, country]
    actor: user
    encryption_in_transit: TLS 1.3
  - stage: processing
    purpose: account_creation
    actor: backend
  - stage: storage
    location: postgres (eu-west-1)
    encryption_at_rest: AES-256
  - stage: transfer
    destination: Sentry (US)
    legal_basis: SCC + DPA signed
  - stage: deletion
    trigger: account_deleted
    method: soft_delete + 30d hard_delete

regulatory_applicability:
  gdpr: yes  # EU residents
  pipl: yes  # China users
  pipa: yes  # Korea users
  ccpa: yes  # California
  hipaa: no
  coppa: no  # no minors

dpia:
  required: yes | no
  reason: "<...>"
  risk_level: low | medium | high
  mitigations: ["..."]

retention_policy:
  - category: account_data
    active: "until deletion"
    archive: "0 days"
    backup: "30 days"
    legal_hold_exception: "tax records 5 years"

dsr_workflow:
  access:
    self_serve: yes
    response_sla: "30 days"
  erasure:
    self_serve: yes  # cancel account flow
    backup_treatment: "anonymize after 30d"
  portability:
    format: JSON
    delivery: download_link

cross_border_transfers:
  - destination: US
    service: Sentry
    legal_basis: SCC
    dpa_signed: yes
  - destination: Singapore
    service: AWS
    legal_basis: adequacy_decision  # PIPA recognized

breach_response:
  detection_sla: "1 hour"
  authority_notification_sla: "72 hours"
  user_notification_threshold: "high-risk"
  drill_frequency: "quarterly"

risks:
  - severity: critical
    description: "<...>"
    remediation: "<...>"

verdict: clear | needs-remediation | blocked
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `assess-business-viability` — `Skill` tool로 invoke
- 페어: `audit-security` — 보안 / 권한 / encryption 별도 검토
- 페어: `review-license-and-ip-risk` — 데이터 라이선스 분리
- 페어: `review-terms-policy-readiness` — privacy policy 작성
- 다음 단계: `define-product-spec` — privacy 요구사항 PRD에 반영

## 8. Anti-patterns

1. **"우리 보안 강해서 GDPR 안 봐도 됨"** — 보안 ≠ privacy. 별도 frame.
2. **"한 명도 EU 없는데"** — 한 명이라도 있으면 적용. IP 기반 정확히 모름.
3. **GA / Mixpanel을 "그냥 분석"** — data processor임. DPA 서명 + cookie 동의.
4. **Soft delete만 + backup 영구 보존** — 위반. backup retention 별도 정책.
5. **"동의 받았으니 영원히 사용"** — purpose limitation. 새 목적 = 새 동의.
6. **DSR을 수동 처리 (자동화 없음)** — 1개월 SLA 못 맞춤. 자동 워크플로우.
7. **Breach 인지 후 자체 조사 우선** — 72시간 안에 통보. 조사 평행.
8. **Sensitive data를 일반 PII로** — explicit consent + special protection 누락.
