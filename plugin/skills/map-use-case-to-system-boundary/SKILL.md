---
name: map-use-case-to-system-boundary
description: This skill should be used when the user wants to "map use cases to systems", "assign system boundaries", "define which system handles what", "map responsibilities to services", or has use cases identified and needs to assign each to a specific system component.
---

# map-use-case-to-system-boundary

각 use case가 어느 시스템 경계에서 실행되는지 매핑한다. §2 `define-features`의 세 번째 stage.

이 매핑이 §3 `design-system`의 infra topology 설계 입력이 된다.

---

## System Boundary 분류

| 경계 유형 | 정의 | 예시 |
|---------|------|------|
| `frontend-spa` | 브라우저에서 실행되는 SPA | Next.js, React, Vue |
| `frontend-mobile` | 모바일 앱 | React Native, Flutter |
| `backend-{service}` | 서버사이드 서비스 | backend-auth-service, backend-api |
| `database-{type}` | 데이터 저장소 | database-postgresql, database-redis |
| `external-saas` | 외부 SaaS API | SendGrid, Stripe, Twilio |
| `queue-{type}` | 메시지 큐 | queue-kafka, queue-sqs |
| `cdn` | CDN / 정적 자산 | Vercel, CloudFront |
| `cicd` | CI/CD 파이프라인 | GitHub Actions |

---

## 매핑 절차

### 1. `map-actor-use-cases` 결과 입력

actor_use_cases 목록을 입력으로 받는다.

### 2. Use Case → System Boundary 매핑

```yaml
use_case_boundaries:
  - use_case_id: uv-001  # anonymous-visitor: 회원가입 페이지 접근
    system_boundary: frontend-spa
    tech_stack: Next.js (App Router)
    data_access: none
    external_calls: []
    
  - use_case_id: au-001  # authenticated-user: 이메일/비밀번호로 회원가입
    system_boundary: frontend-spa
    tech_stack: Next.js
    data_access: none
    external_calls:
      - target: backend-auth-service
        method: POST /auth/signup
        
  - use_case_id: as-001  # auth-service: 검증
    system_boundary: backend-auth-service
    tech_stack: Go (net/http)
    data_access: database-postgresql (users table)
    external_calls: []
    
  - use_case_id: as-002  # auth-service: 해시 저장
    system_boundary: backend-auth-service
    tech_stack: Go (bcrypt)
    data_access: database-postgresql (WRITE users)
    external_calls:
      - target: notification-service
        method: event: user.created
        
  - use_case_id: sg-001  # sendgrid: 이메일 발송
    system_boundary: external-saas
    tech_stack: SendGrid API v3
    data_access: none
    external_calls: []
    
  - use_case_id: sg-002  # sendgrid: webhook 처리
    system_boundary: backend-auth-service  # webhook handler는 backend에 있음
    tech_stack: Go (webhook handler)
    data_access: database-postgresql (UPDATE users SET verified=true)
    external_calls: []
```

### 3. System Boundary 검증

- [ ] 각 use case가 정확히 하나의 system_boundary에 매핑됐는가?
- [ ] 동일 actor의 use case들이 적절한 system boundary에 분산됐는가?
- [ ] External SaaS call이 모두 명시됐는가?
- [ ] Database access (READ/WRITE) 방향이 명시됐는가?

### 4. Boundary 간 의존성 도식

```
frontend-spa
  → backend-auth-service (POST /auth/signup)
  → backend-auth-service (POST /auth/login)

backend-auth-service
  → database-postgresql (READ/WRITE users)
  → external-saas/sendgrid (email 발송)
  ← external-saas/sendgrid (webhook 수신)

external-saas/sendgrid
  → backend-auth-service (POST /webhooks/sendgrid)
```

---

## 출력 형식

```yaml
use_case_boundaries:
  - use_case_id: {id}
    system_boundary: {boundary 유형}
    tech_stack: {구체 기술}
    data_access: {none / READ {table} / WRITE {table}}
    external_calls:
      - target: {system_boundary}
        method: {HTTP method + path / event name}

boundary_dependency_graph:
  - from: {system_boundary}
    to: {system_boundary}
    protocol: {HTTP / event / webhook}
    use_cases: [{use_case_ids}]
```

---

## 다음 단계

→ `compose-feature-from-use-cases` — use case → feature 합성
