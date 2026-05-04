---
name: identify-actors
description: This skill should be used when the user wants to "identify actors", "find who uses the system", "map stakeholders", "identify system participants", or is starting feature definition and needs to enumerate all actors in the system.
---

# identify-actors

시스템에 참여하는 모든 actor를 열거하고 분류한다. §2 `define-features`의 첫 번째 stage.

feature는 **여러 actor의 use case 합성**이므로, actor 식별이 feature 정의의 선행 조건이다.

---

## Actor 분류 기준

| 분류 | 정의 | 예시 |
|------|------|------|
| **user** | 최종 사용자 (role별 세분화) | anonymous-visitor, authenticated-user, admin, super-admin |
| **system** | 내부 서비스 / 백엔드 컴포넌트 | auth-service, payment-service, notification-service |
| **3rd-party** | 외부 SaaS / API | SendGrid (email), Stripe (payment), GitHub OAuth |
| **external-tool** | 관리 / 운영 도구 | monitoring (Datadog), CI/CD (GitHub Actions), analytics |

---

## 식별 절차

### 1. PRD에서 Actor 후보 추출

PRD의 다음 섹션에서 actor를 추출한다:
- User Stories ("As a [user type], I want to...")
- Stakeholder 목록
- Integration 목록
- Non-functional requirements (모니터링, 감사 로그 등)

### 2. Actor 열거

```yaml
actors:
  user_actors:
    - id: anonymous-visitor
      description: 로그인 없이 접근하는 사용자
      auth: none
      
    - id: authenticated-user
      description: 이메일/비밀번호 또는 OAuth로 로그인한 사용자
      auth: jwt
      
    - id: admin
      description: 관리자 권한을 가진 사용자
      auth: jwt + role=admin
      
  system_actors:
    - id: auth-service
      description: 인증/인가 담당 내부 서비스
      tech: Go + PostgreSQL
      
    - id: notification-service
      description: 이메일/SMS/push 발송 서비스
      tech: Node.js + queue
      
  third_party_actors:
    - id: sendgrid
      description: 이메일 발송 SaaS
      type: email-provider
      sdk: sendgrid/sendgrid-go
      
    - id: stripe
      description: 결제 처리 SaaS
      type: payment-provider
      sdk: stripe/stripe-go
      
  external_tool_actors:
    - id: github-actions
      description: CI/CD 파이프라인
      type: cicd
      
    - id: datadog
      description: 모니터링 / 알림
      type: monitoring
```

### 3. Actor 검증 체크

- [ ] 각 user role이 실제로 다른 권한 / use case를 가지는가? (같으면 통합)
- [ ] 모든 외부 SaaS 의존성이 actor로 등록됐는가?
- [ ] 내부 서비스가 분리된 배포 단위로 존재하는가? (모놀리스면 하나의 system actor)
- [ ] 미래 actor가 아닌 현재 scope 내 actor만 포함했는가? (YAGNI)

---

## 출력 형식

```yaml
project: {프로젝트명}
actors_identified_at: {ISO 8601}

user_actors:
  - id: ...
    description: ...
    auth: ...

system_actors:
  - id: ...
    description: ...
    tech: ...

third_party_actors:
  - id: ...
    description: ...
    sdk: ...

external_tool_actors:
  - id: ...
    description: ...

total_actors: {N}
```

---

## 다음 단계

→ `map-actor-use-cases` — actor별 use case 식별
