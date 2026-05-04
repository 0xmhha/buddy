---
name: compose-feature-from-use-cases
description: This skill should be used when the user wants to "compose features from use cases", "group use cases into features", "create feature definitions from actors", or has actor use cases mapped to system boundaries and needs to synthesize them into coherent features.
---

# compose-feature-from-use-cases

cross-actor use case를 묶어 feature를 정의한다. §2 `define-features`의 네 번째 stage.

feature = 여러 actor의 use case 합성. 이 합성이 §3 infra 설계와 §4 구현 track 분리의 기반이 된다.

---

## Feature 합성 원칙

1. **Cross-actor**: feature는 최소 2개 actor use case의 합성 (단일 actor 전용이면 task로 낮춤)
2. **독립성**: feature 내부 use case들이 하나의 사용자 가치를 제공
3. **재사용성**: feature는 cherry-pick / patch로 다른 제품에 재사용 가능
4. **원자성**: feature 내 use case들이 함께 완료되어야 가치 발생

---

## 합성 절차

### 1. `map-use-case-to-system-boundary` 결과 입력

use_case_boundaries 목록을 입력으로 받는다.

### 2. Use Case 그룹화 기준

use case를 다음 기준으로 그룹화한다:
- **동일 사용자 목적**: 사용자가 달성하려는 goal이 동일
- **의존성**: use case A가 완료되어야 use case B가 가능
- **원자적 가치**: 묶음 전체가 완료되어야 가치 발생

예시:
```
"이메일/비밀번호 회원가입" 목적으로 묶이는 use cases:
  - au-001: 사용자가 폼 제출 (frontend-spa)
  - as-001: auth-service 검증 (backend-auth-service)
  - as-002: 비밀번호 해시 + DB 저장 (backend-auth-service)
  - as-003: JWT 발급 (backend-auth-service)
  - sg-001: verification email 발송 (external-saas)
  - sg-002: webhook 처리 (backend-auth-service)

→ feature: signup-email-password
```

### 3. Feature 정의

각 그룹을 feature로 정의한다:

```yaml
features:
  - feature_id: signup-email-password
    name: "Email/Password Sign-Up"
    summary: "사용자가 이메일/비밀번호로 회원가입하고 이메일 인증을 완료한다"
    
    composed_from:
      - use_case_id: au-001
        actor: authenticated-user
        system_boundary: frontend-spa
        
      - use_case_id: as-001
        actor: auth-service
        system_boundary: backend-auth-service
        
      - use_case_id: as-002
        actor: auth-service
        system_boundary: backend-auth-service
        
      - use_case_id: as-003
        actor: auth-service
        system_boundary: backend-auth-service
        
      - use_case_id: sg-001
        actor: sendgrid
        system_boundary: external-saas
        
      - use_case_id: sg-002
        actor: sendgrid
        system_boundary: backend-auth-service
        
    implementation_tracks:
      - track: frontend
        actor: authenticated-user
        system_boundary: frontend-spa
        use_cases: [au-001]
        
      - track: backend
        actor: auth-service
        system_boundary: backend-auth-service
        use_cases: [as-001, as-002, as-003, sg-002]
        
      - track: integration
        actor: sendgrid
        system_boundary: external-saas
        use_cases: [sg-001]
```

### 4. Feature 분리 판단

feature가 너무 크면 분리한다:

```
signup-email-password (원본)
  → signup-form-and-validation (frontend-spa만)
  → signup-account-creation (backend-auth-service만)
  → email-verification-flow (external-saas + webhook)
```

분리 기준: 한 feature가 3+ implementation track을 가지면 분리 고려.

### 5. Feature 검증 체크

- [ ] 각 feature가 독립적으로 배포/재사용 가능한가?
- [ ] Feature 내 모든 use case가 동일 사용자 목적을 향하는가?
- [ ] Implementation track이 명확히 분리됐는가?
- [ ] Feature 간 중복 use case가 없는가?

---

## 출력 형식

```yaml
features:
  - feature_id: {kebab-case}
    name: {feature 이름}
    summary: {한 문장}
    composed_from:
      - use_case_id: {id}
        actor: {actor_id}
        system_boundary: {boundary}
    implementation_tracks:
      - track: {frontend/backend/integration}
        actor: {actor_id}
        system_boundary: {boundary}
        use_cases: [{use_case_ids}]

total_features: {N}
```

---

## 다음 단계

→ `define-feature-spec` — feature spec 전체 포맷 작성 (actor / use case / acceptance / test_plan 포함)
