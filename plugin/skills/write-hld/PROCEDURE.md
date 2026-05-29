# write-hld — High Level Design 문서 작성

PRD에 정의된 use case를 만족시킬 수 있는 **product 구성(physical structure)** 을 설계한다. Phase 1 Mode A의 마지막 직전 단계로, `define-product-spec`(PRD) 완료 후 호출되어 `autoplan`(validate-spec)의 review-design / review-devex / review-engineering 검증 대상을 생산한다.

**진입 조건**: PRD 완료 (`define-product-spec` 산출물 존재).
**산출물**: HLD 문서 — product decomposition + tech stack + communication patterns + use case → product mapping.
**다음 phase**: `autoplan`(validate-spec)으로 PRD + HLD 통합 리뷰.

PRD가 "what & why"(logical)라면 HLD는 "how at high level"(physical). Phase 3 detailed design은 HLD가 정한 제약 안에서 상세화한다.

## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| PRD (problem, actors, use cases, requirements) | ✅ | artifact | `define-product-spec` 산출물 | 먼저 `/buddy:define-product-spec` 을 실행하세요 |
| 기술 선호도 / 제약 | 선택 | knowledge | 사용자 도메인 지식 | "선호하는 기술 스택이나 제약(예: 회사 표준 stack)이 있나요?" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| HLD 문서 (9 섹션) | artifact | `docs/hld.md` 또는 PRD 내 §HLD 섹션 | `autoplan` (validate-spec), Phase 3 `design-system` (detailed expansion 입력) |
| Use Case → Product Mapping | artifact | structured (HLD §5) | Phase 2 `compose-feature-from-use-cases`, `map-use-case-to-system-boundary` |

---

## HLD 9 섹션

### §1. Product Decomposition

프로덕트를 구성하는 deployable / distributable 단위를 열거한다. "product"는 다음을 모두 포함:

| 유형 | 예시 |
|------|------|
| frontend-web | Next.js SPA, Vue dashboard |
| frontend-mobile | iOS app, Android app, React Native app |
| backend-service | Go API server, Python worker, Node.js BFF |
| database | PostgreSQL, Redis, MongoDB |
| infra | nginx reverse proxy, k8s cluster, CDN |
| sdk | TypeScript client SDK, Go client SDK |
| cli | command-line tool |
| library | published package (npm, pip, cargo) |
| extension | browser extension, IDE extension |

각 product는 deployable / publishable 단위. 단일 monolith라면 product 1개.

```yaml
products:
  - id: web-app
    type: frontend-web
    summary: "사용자가 접근하는 메인 웹 인터페이스"

  - id: api-server
    type: backend-service
    summary: "비즈니스 로직 + 데이터 처리 API"

  - id: postgres-main
    type: database
    summary: "사용자/주문/콘텐츠 마스터 데이터"

  - id: client-sdk-ts
    type: sdk
    summary: "외부 개발자가 API를 호출하기 위한 TypeScript SDK"
```

### §2. Per-product Role

각 product가 책임지는 영역과 책임 경계를 명시한다.

```yaml
products:
  - id: web-app
    responsibilities:
      - "사용자 인증 UI"
      - "데이터 입력 폼 + 검증"
      - "결과 시각화"
    not_responsible_for:
      - "비즈니스 룰 (api-server에 위임)"
      - "데이터 영속성"

  - id: api-server
    responsibilities:
      - "비즈니스 룰 실행"
      - "데이터 CRUD"
      - "외부 SaaS 연동 (결제, 이메일)"
    not_responsible_for:
      - "UI 렌더링"
      - "파일 저장 (object storage에 위임)"
```

### §3. Per-product Tech Stack

각 product의 high-level 기술 선택. **detailed library version / config는 Phase 3**.

```yaml
products:
  - id: web-app
    language: TypeScript
    framework: Next.js 14 (App Router)
    key_libraries: [react-query, tailwind, shadcn/ui]
    rationale: "SSR + React 생태계 + 사용자 경험"

  - id: api-server
    language: Go 1.22+
    framework: net/http + chi router
    key_libraries: [sqlx, zerolog, validator/v10]
    rationale: "고성능 + 단순한 binary 배포 + 팀 숙련도"

  - id: postgres-main
    version: PostgreSQL 16
    extensions: [pgvector (검색)]
    rationale: "관계형 + JSONB + 안정성"

  - id: client-sdk-ts
    language: TypeScript
    framework: 없음 (zero-dependency)
    build: tsup → dual ESM/CJS
    rationale: "외부 의존성 최소화로 통합 마찰 감소"
```

### §4. Inter-product Communication

product 간 통신 패턴을 명시한다. 통신 채널과 데이터 형식을 모두 포함.

| 패턴 | 예시 |
|------|------|
| REST | HTTP/JSON, OpenAPI 정의 |
| GraphQL | HTTP/JSON, schema 정의 |
| gRPC | HTTP/2, protobuf |
| WebSocket | 양방향 실시간 |
| Server-Sent Events | 단방향 server → client |
| Message Queue | Kafka, RabbitMQ, SQS |
| Pub/Sub | Redis pub/sub, NATS |
| IPC | Unix socket, named pipe (같은 머신) |
| Shared DB | 두 service가 동일 DB read/write (caution: coupling) |
| File system | 공유 파일 (S3, NFS) |

```yaml
communications:
  - from: web-app
    to: api-server
    pattern: REST (HTTP/JSON)
    auth: JWT bearer token
    notes: "OpenAPI spec은 Phase 3 design-api-contract에서 정의"

  - from: api-server
    to: postgres-main
    pattern: SQL (TCP, libpq)
    pooling: pgxpool (max 25 conn)

  - from: api-server
    to: stripe (external)
    pattern: REST (HTTPS/JSON)
    webhook_back: api-server /webhooks/stripe

  - from: api-server
    to: notification-worker
    pattern: message queue (SQS)
    format: JSON event
    notes: "비동기 처리 — 이메일 발송 등"
```

### §5. Use Case → Product Mapping

**HLD의 핵심 섹션**. PRD의 각 use case를 product 흐름에 매핑한다. 데이터 흐름을 step 단위로 명시.

```yaml
use_case_mapping:
  - use_case_id: "이메일 로그인"
    steps:
      - step: 1
        product: web-app
        action: "이메일/비밀번호 폼 표시, 클라이언트 측 validation"
        data_out: "credentials (email, password)"

      - step: 2
        product: api-server
        action: "POST /auth/login — credentials 검증"
        data_in: "credentials (REST/JSON)"
        data_out: "user_id, JWT, refresh_token"
        depends_on: postgres-main

      - step: 3
        product: postgres-main
        action: "SELECT user WHERE email=? + bcrypt 비교"
        data_in: "SQL query"
        data_out: "user row"

      - step: 4
        product: web-app
        action: "JWT를 localStorage에 저장 + dashboard redirect"
        data_in: "JWT (REST/JSON)"

  - use_case_id: "결제 진행"
    steps:
      - step: 1
        product: web-app
        action: "Stripe Elements 폼 표시"
        data_out: "(no data — Stripe iframe direct)"

      - step: 2
        product: stripe (external)
        action: "카드 정보 tokenize"
        data_in: "card details"
        data_out: "payment_token"

      - step: 3
        product: web-app
        action: "payment_token을 api-server로 전달"
        data_out: "payment_token + order_id"

      - step: 4
        product: api-server
        action: "POST /orders/{id}/pay — Stripe charge API 호출"
        data_in: "payment_token + order_id"
        data_out: "order updated (paid)"
        depends_on: [postgres-main, stripe, notification-worker]

      - step: 5
        product: notification-worker
        action: "주문 확인 이메일 발송 (비동기)"
        data_in: "order paid event"
```

매핑이 끝나면 **각 use case가 어느 product들을 거치는지 명확해진다**. Phase 2의 `map-use-case-to-system-boundary`는 이 매핑을 detailed boundary 분류로 확장.

### §6. External Integration

3rd-party SaaS, 외부 API, 결제 게이트웨이, 인증 provider 등.

```yaml
external_integrations:
  - service: Stripe
    purpose: 결제 처리
    used_by: api-server
    risk: vendor lock-in (mitigation: Payment Service interface 분리)

  - service: SendGrid
    purpose: 트랜잭션 이메일 발송
    used_by: notification-worker
    risk: 발송 실패 → retry queue (DLQ)

  - service: Sentry
    purpose: 에러 추적
    used_by: [web-app, api-server, notification-worker]
    risk: 데이터 유출 (mitigation: PII scrubbing)
```

### §7. Deployment Topology

product가 어디에 배포되는지 high-level 구조.

```
[CDN: Cloudflare]
       ↓
[Vercel: web-app]
       ↓ REST
[AWS ECS: api-server x3]
       ↓ SQL              ↓ SQS
[RDS Postgres]      [SQS Queue]
                          ↓
                  [Lambda: notification-worker]
```

배포 환경 + 스케일링 단위 + 가용성 목표 명시. **detailed IaC는 Phase 3**.

### §8. Licensing

```yaml
licensing:
  product_overall: Apache-2.0
  per_product:
    - id: web-app
      license: Apache-2.0
    - id: api-server
      license: Apache-2.0
    - id: client-sdk-ts
      license: MIT
      rationale: "외부 사용자 마찰 최소화"
  dependencies_check: "Phase 6 review-license-and-ip-risk에서 호환성 검증"
```

### §9. Distribution Model

```yaml
distribution:
  - product: web-app
    model: hosted SaaS (사용자가 URL로 접속)
    deployment: continuous deployment (main branch → vercel)

  - product: api-server
    model: hosted SaaS (직접 노출 안 함, web-app + sdk 경유)
    deployment: blue-green via ECS

  - product: client-sdk-ts
    model: published npm package
    versioning: semver
    release: GitHub Actions on git tag

  - product: cli (있다면)
    model: GitHub Releases binary + brew formula
```

---

## 실행 절차

### Step 1. PRD 입력 확인

`define-product-spec` 산출물에서 다음을 추출:
- Problem statement
- Target user & buyer
- Actors
- Use cases (logical, with data flow)
- Functional / Non-functional requirements

PRD가 없으면 사용자에게 먼저 PRD 작성을 안내.

### Step 2. Product Decomposition (§1)

actor + use case를 기반으로 어떤 product가 필요한지 식별:
- frontend가 필요한 actor가 있는가? → frontend-web / frontend-mobile
- 데이터 영속성이 필요한가? → database
- 외부 노출 API가 필요한가? → backend-service
- 외부 개발자 사용 surface가 있는가? → sdk / cli / library

각 product에 안정적 id 부여 (kebab-case).

### Step 3. Per-product Role + Tech Stack (§2, §3)

각 product의 책임과 high-level tech stack 결정. **detailed config는 명시 X — Phase 3에서**.

### Step 4. Inter-product Communication (§4)

product 간 통신 매트릭스 작성:
- 누가 누구에게 무엇을 보내는가
- 어떤 프로토콜 / 형식
- 동기 vs 비동기

### Step 5. Use Case → Product Mapping (§5) — 핵심

PRD의 각 use case에 대해:
- step 단위로 어느 product가 어떤 action을 하는지
- 각 step의 data in/out
- 외부 의존성 표시

이 매핑이 review-design의 검증 대상.

### Step 6. External Integration + Deployment + Licensing + Distribution (§6~§9)

상용 / 비상업 프로젝트별 progressive:
- 비상업/학습 프로젝트: §1~§5 + §8(라이센스) 만
- 상용 프로젝트: 전체 9 섹션

### Step 7. 검증 체크리스트

- [ ] 모든 product가 안정적 id + role 정의됨
- [ ] PRD의 모든 use case가 §5 mapping에 포함됨
- [ ] 모든 inter-product communication에 프로토콜 명시됨
- [ ] 외부 의존성이 §6에 열거됨
- [ ] 각 product의 license 결정됨
- [ ] (상용) deployment topology 도식 존재

---

## 출력 형식

`docs/hld.md` (또는 PRD 내 HLD 섹션) 으로 저장. 9 섹션 모두 포함. 미완 섹션은 `<TBD>` 표시 후 진행.

---

## 다음 단계

→ `autoplan` (validate-spec) — PRD + HLD 통합 리뷰 (review-scope, review-design, review-devex, review-engineering)
→ Phase 2 `define-features` — use case → feature 변환
→ Phase 3 `design-system` — HLD 제약 안에서 detailed design
