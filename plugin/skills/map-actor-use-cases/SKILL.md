---
name: map-actor-use-cases
description: This skill should be used when the user wants to "map use cases", "define use cases per actor", "create use case diagram", "identify what each actor does", or has actors identified and needs to map what each actor does in the system.
---

# map-actor-use-cases

actor별 use case를 식별한다. §2 `define-features`의 두 번째 stage.

UML use case diagram 등가 작업. actor 시점에서 시스템과의 상호작용을 동사+목적어 형태로 나열한다.

---

## Use Case 작성 원칙

1. **actor 시점**: "{actor}가 {동사} + {목적어}" 형태로 작성
2. **독립성**: 각 use case는 actor 단독으로 식별 가능한 상호작용
3. **측정 가능**: use case는 §6 QA에서 테스트 가능해야 함
4. **현재 scope**: §8의 iteration에서 발견된 use case는 §2 재진입 시 추가

---

## 매핑 절차

### 1. `identify-actors` 결과 입력

`identify-actors`의 출력(actor 목록)을 입력으로 받는다.

### 2. Actor별 Use Case 열거

각 actor에 대해 시스템과의 상호작용을 나열한다.

```yaml
actor_use_cases:
  - actor_id: anonymous-visitor
    use_cases:
      - id: uv-001
        title: "회원가입 페이지에 접근한다"
        trigger: "landing page CTA 클릭"
        outcome: "회원가입 폼 표시"
        
      - id: uv-002
        title: "로그인 페이지에 접근한다"
        trigger: "login 링크 클릭"
        outcome: "로그인 폼 표시"
        
  - actor_id: authenticated-user
    use_cases:
      - id: au-001
        title: "이메일/비밀번호로 회원가입한다"
        trigger: "회원가입 폼 제출"
        outcome: "계정 생성 + verification email 수신"
        
      - id: au-002
        title: "이메일/비밀번호로 로그인한다"
        trigger: "로그인 폼 제출"
        outcome: "JWT 발급 + dashboard redirect"
        
      - id: au-003
        title: "비밀번호를 재설정한다"
        trigger: "forgot password 링크 클릭"
        outcome: "reset email 수신 + 새 비밀번호 설정"
        
  - actor_id: auth-service
    use_cases:
      - id: as-001
        title: "이메일 형식과 비밀번호 요구사항을 검증한다"
        trigger: "POST /auth/signup"
        outcome: "validation result + 400/200"
        
      - id: as-002
        title: "비밀번호를 해시 저장한다"
        trigger: "validation 통과 후"
        outcome: "bcrypt hash → DB users table"
        
      - id: as-003
        title: "JWT를 발급한다"
        trigger: "user record 생성 후"
        outcome: "signed JWT (exp: 24h)"
        
  - actor_id: sendgrid
    use_cases:
      - id: sg-001
        title: "이메일 인증 메시지를 발송한다"
        trigger: "auth-service → sendgrid API call"
        outcome: "verification email delivered"
        
      - id: sg-002
        title: "인증 링크 클릭을 webhook으로 처리한다"
        trigger: "사용자 링크 클릭"
        outcome: "POST /webhooks/sendgrid → account verified"
```

### 3. Use Case 검증 체크

- [ ] 각 use case가 단일 actor의 단일 상호작용인가?
- [ ] Trigger와 outcome이 명확한가?
- [ ] §6 QA에서 테스트 가능한가?
- [ ] actor 간 중복이 없는가? (use case는 한 actor 소속)

---

## 출력 형식

```yaml
actor_use_cases:
  - actor_id: {id}
    actor_type: {user/system/3rd-party/external-tool}
    use_cases:
      - id: {actor_prefix}-{N}
        title: "{동사 + 목적어}"
        trigger: "{무엇이 이 use case를 시작하는가}"
        outcome: "{성공 시 결과}"
        
total_use_cases: {N}
```

---

## 다음 단계

→ `map-use-case-to-system-boundary` — use case별 시스템 경계 매핑
