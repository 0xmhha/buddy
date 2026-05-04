---
name: define-feature-spec
description: This skill should be used when the user wants to "write a feature spec", "define feature requirements", "create feature specification", "document a feature", or has a composed feature and needs to write a complete specification including actors, use cases, acceptance criteria, and test plans.
---

# define-feature-spec

feature의 완전한 명세서를 작성한다. §2 `define-features`의 다섯 번째 stage.

§3 (infra), §4 (implementation), §6 (test), §8 (metric)의 입력 schema가 되는 핵심 artifact.

---

## Feature Spec 포맷 (Q8=(a) 확장)

```yaml
feature_id: {kebab-case-id}
name: {feature 이름}
summary: {한 문장}
problem: {해결하는 사용자/제품 문제}

# Q8=(a): actor / use case / system_boundary 필수 필드
actors:
  - id: {actor_id}
    use_cases:
      - {use case 제목}
      - {use case 제목}
    system_boundary: {frontend-spa / backend-{service} / external-saas / ...}

scope: |
  {포함 범위 — 구체적으로}

out_of_scope: |
  {제외 범위 — 특히 "MVP 이후"라고 할 만한 것}

acceptance_criteria:
  {actor_id}: |
    {actor 관점 완료 기준 — Given/When/Then 형식 권장}

test_plan:
  per_actor:
    {actor_id}: |
      {테스트 유형 + 구체 시나리오}
  integration: |
    {cross-actor 흐름 테스트 — E2E 시나리오}

implementation_notes: |
  {설계 결정 / trade-off / migration 주의점}

# 구현 후 채워지는 필드
code_links: []
status: draft  # draft / candidate / ready / in-progress / implemented / verified / deprecated
owners: []
updated_at: {ISO 8601}
```

---

## 작성 예시

```yaml
feature_id: signup-email-password
name: "Email/Password Sign-Up"
summary: "사용자가 이메일/비밀번호로 회원가입하고 이메일 인증을 완료한다"
problem: |
  현재 신규 사용자는 계정을 만들 방법이 없어 서비스를 사용할 수 없다.
  
actors:
  - id: authenticated-user
    use_cases:
      - "이메일/비밀번호 폼을 작성하고 제출한다"
      - "실시간 validation feedback을 확인한다"
      - "회원가입 성공 후 dashboard로 redirect된다"
    system_boundary: frontend-spa
    
  - id: auth-service
    use_cases:
      - "이메일 형식과 비밀번호 요구사항을 검증한다"
      - "비밀번호를 bcrypt로 해시 저장한다"
      - "JWT를 발급한다"
    system_boundary: backend-auth-service
    
  - id: sendgrid
    use_cases:
      - "이메일 인증 메시지를 발송한다"
      - "인증 링크 클릭 webhook을 처리한다"
    system_boundary: external-saas

scope: |
  - 이메일/비밀번호 회원가입 폼
  - 이메일 형식 + 비밀번호 strength 실시간 검증
  - bcrypt 비밀번호 해시 저장
  - JWT 발급 (exp: 24h)
  - SendGrid를 통한 이메일 인증
  - 인증 완료 후 계정 활성화

out_of_scope: |
  - OAuth (Google/GitHub) 로그인 — 별도 feature
  - 2FA — 별도 feature
  - 비밀번호 재설정 — 별도 feature

acceptance_criteria:
  authenticated-user: |
    Given: 회원가입 페이지에 있다
    When: 유효한 이메일 + 강도 충족 비밀번호를 입력하고 제출한다
    Then: "가입 확인 이메일을 확인하세요" 메시지가 표시된다
    
    Given: 가입 확인 이메일을 받았다
    When: 인증 링크를 클릭한다
    Then: 계정이 활성화되고 dashboard로 redirect된다
    
  auth-service: |
    Given: POST /auth/signup 요청
    When: 유효한 email + password
    Then: 201 Created + { token: JWT, user_id }
    
    Given: POST /auth/signup 요청
    When: 이미 사용 중인 email
    Then: 409 Conflict + { error: "email_already_exists" }
    
  sendgrid: |
    Given: 신규 user record 생성
    When: auth-service → sendgrid API 호출
    Then: 5초 내 verification email 발송됨 (delivery rate ≥ 95%)

test_plan:
  per_actor:
    authenticated-user: |
      E2E (Playwright):
        - signup → email receive → link click → dashboard 전체 흐름
        - validation error 상태 (invalid email, weak password)
        - 이미 사용 중인 email 에러
        
    auth-service: |
      Unit:
        - bcrypt hash 검증 (known input → known hash)
        - JWT payload 구조 + expiry 검증
      Integration:
        - POST /auth/signup 성공/실패 케이스 (real DB)
        - webhook handler 처리 검증
        
    sendgrid: |
      Contract test:
        - sendgrid mock server로 API 호출 schema 검증
        - webhook payload schema 검증
        
  integration: |
    Full flow E2E:
      signup form submit → backend 201 → email sent → link click → 
      webhook 200 → account verified → login success

implementation_notes: |
  - bcrypt cost factor: 12 (성능/보안 trade-off)
  - JWT 24h expiry — refresh token은 이 feature scope 밖
  - SendGrid API key는 환경변수로 관리 (SENDGRID_API_KEY)
  - webhook endpoint는 /webhooks/sendgrid — HMAC 서명 검증 필수

code_links: []
status: draft
owners: []
updated_at: 2026-05-04T00:00:00Z
```

---

## 작성 체크리스트

- [ ] 모든 actor가 use_cases와 system_boundary를 가지는가?
- [ ] acceptance_criteria가 actor별로 Given/When/Then 형식인가?
- [ ] test_plan이 actor별 + cross-actor integration을 모두 포함하는가?
- [ ] out_of_scope가 명확히 정의됐는가?
- [ ] implementation_notes에 security 관련 결정이 포함됐는가?

---

## 다음 단계

→ `query-feature-registry` — 유사 feature 검색
→ `score-feature-priority` — RICE 우선순위
