---
name: design-system
description: This skill should be used when the user wants to "design the system", "choose tech stack", "design infrastructure", "design API", "design data model", "create ADR", "design architecture", or has a feature backlog ready and needs technical design. Orchestrates §3 Technical Design phase.
---

# design-system — §3 Technical Design Orchestrator

§3 라이프사이클 단계의 진입점. Feature backlog (actor / use case / system boundary 포함) → Tech stack ADR + infra blueprint + API/data model.

**진입 조건**: §2 feature backlog 확정 (actor + use case + system boundary 포함).
**산출물**: Tech stack ADR, infra topology diagram, API contract, data model schema.
**다음 phase**: Technical design 확정 후 → `plan-build` (§4).

---

## Stage 흐름

```
design-system (§3 phase orchestrator)
├── stage 1: map-use-cases-to-infra     (use case → infra component 브릿지)
├── stage 2: derive-system-topology     (actor 그래프 + use case → 시스템 토폴로지)
├── stage 3: define-tech-stack          (언어/프레임워크/DB 선택 — 락인 영향 평가)
├── stage 4: design-api-contract        (REST/GraphQL/RPC 계약 — actor 간 경계 = API 경계)
├── stage 5: design-data-model          (스키마/마이그레이션/인덱싱)
├── stage 6: design-auth-model          (RBAC/ABAC, 멀티테넌트 격리)
├── stage 7: design-observability       (로깅/메트릭/트레이싱 표준)
├── stage 8: [design-deploy-strategy]   (배포 전략 — canary/blue-green/rolling)
├── stage 9: write-adr                  (Architecture Decision Record)
└── stage 10: autoplan                  (technical design 산출물 4-mode review)
```

> 🆕 표시 stage는 신규 작성 필요. 현재는 orchestrator가 직접 수행.

---

## 실행 절차

### Stage 1: Use Case → Infra 브릿지

§2 feature spec의 actor system boundary를 실제 infra component로 매핑한다.

예시:
```
frontend-spa → Next.js + Vercel CDN
backend-auth-service → Go service + PostgreSQL
external-saas (SendGrid) → SendGrid SDK + webhook handler
```

`design-mcp-server` skill을 invoke해 MCP 연동이 필요한 external SaaS를 식별한다.

### Stage 2: 시스템 토폴로지

actor 그래프와 use case 흐름을 기반으로 시스템 토폴로지를 도식화한다:
- sync / async 통신 구분
- request/response vs event-driven 관계
- 데이터 흐름 방향

`explore-design-variants` skill을 invoke해 토폴로지 후보를 2-3개 생성하고 trade-off를 비교한다.

### Stage 3: Tech Stack 선택

언어 / 프레임워크 / DB 선택에서 **락인 영향**을 반드시 평가한다.

선택 기준 매트릭스:
| 항목 | 옵션 A | 옵션 B | 락인 비용 | 추천 |
|------|--------|--------|---------|------|
| Backend | Go | Node.js | 낮음 | ... |
| DB | PostgreSQL | MongoDB | 중간 | ... |

`consult-codex` skill로 second opinion을 얻는다.

### Stage 4: API Contract

actor 간 경계 = API 경계 원칙으로 API를 설계한다.

필수 포함:
- Endpoint 목록 (method, path, request/response schema)
- Authentication method
- Error response format
- Rate limiting 정책
- Versioning 전략

`design-embedding-search` skill을 invoke해 검색 API가 필요하면 embedding + hybrid search를 설계한다.

### Stage 5: Data Model

스키마 / 마이그레이션 / 인덱싱을 설계한다.

```sql
-- 예시 포맷
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  ...
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index 설계 근거 포함
CREATE INDEX idx_users_email ON users(email); -- login lookup
```

### Stage 6: Auth Model

RBAC / ABAC / 멀티테넌트 격리 전략을 결정한다.
`design-billing-system` skill을 invoke해 billing이 필요하면 subscription tier와 연결한다.

### Stage 7: Observability

로깅 / 메트릭 / 트레이싱 표준을 정의한다:
- Log format (structured JSON with trace_id, user_id, request_id)
- Metric namespacing 규칙
- Tracing sampling rate
- Alert threshold 기준

### Stage 8: 배포 전략

`design-deploy-strategy` skill을 invoke해 canary / blue-green / rolling / recreate 중 선택한다.
`design-artifact-storage` skill을 invoke해 artifact 저장/배포 설계를 한다.

### Stage 9: ADR 작성

각 주요 결정에 대해 Architecture Decision Record를 작성한다.

ADR 포맷:
```markdown
# ADR-{N}: {결정 제목}

**Status**: Accepted
**Date**: {ISO date}
**Deciders**: {이름}

## Context
{결정이 필요한 배경}

## Decision
{내린 결정}

## Rationale
{이유 + 대안 비교}

## Consequences
{영향: 긍정 / 부정}
```

### Stage 10: autoplan Review

`autoplan` (cross-phase review sub-orchestrator)을 invoke해 technical design 산출물을 4-mode review한다.
- `review-architecture` — 구조적 무결성
- `review-engineering` — 구현 가능성
- `review-design` — 디자인 차원
- `review-devex` — developer-facing이면 DX 검토

`consult-design-system` skill로 UI 설계가 필요하면 design system을 생성한다.

---

## 다음 phase

- `/buddy:plan-build` — §4 Implementation Plan (권장)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §§3, §3.1
- Use case → infra 브릿지 설계: 동 문서 §4 §3
