# design-system — 3단계 Technical Design Orchestrator

3단계 라이프사이클 단계의 진입점. Feature backlog (actor / use case / system boundary 포함) → Tech stack ADR + infra blueprint + API/data model.

**진입 조건**: 2단계 feature backlog 확정 (actor + use case + system boundary 포함).
**산출물**: Tech stack ADR, infra topology diagram, API contract, data model schema.
**다음 phase**: Technical design 확정 후 → `plan-build` (4단계).

---

## Stage 흐름

```
design-system (3단계 phase orchestrator)
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

> 브라켓(`[name]`)으로 표시된 stage 는 신규 작성 필요. 현재는 orchestrator 가 직접 수행.

---

## 실행 절차

### Stage 1: Use Case → Infra 브릿지

2단계 feature spec의 actor system boundary를 실제 infra component로 매핑한다.

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

`define-tech-stack` skill 을 invoke 한다 — 언어 / 프레임워크 / DB / cache / queue / hosting / observability / CI 8+ 차원을 alternatives 비교 + cross-차원 호환성 매트릭스 + 5년 lock-in 정량 평가로 evidence-based 결정. 산출물은 Decisions Table + Risk Register + ADR draft handoff.

호출 형태:
- 단독: `/buddy:define-tech-stack "<feature backlog 또는 워크로드>"`
- chain (권장): `/buddy:chain define-tech-stack,write-adr -- "<feature>"` — 결정 즉시 ADR 영속화

본 stage 의 결정은 Stage 4 (API Contract) 의 backend framework 제약, Stage 5 (Data Model) 의 DB 제약, Stage 8 (배포 전략) 의 hosting/runtime 제약으로 cascade 된다.

`consult-codex` skill 로 second opinion 을 추가로 얻을 수 있다 (high-stakes 결정 시).

### Stage 4: API Contract

`design-api-contract` skill 을 invoke 한다 — actor 간 경계 = API 경계 원칙으로 REST/GraphQL/RPC 선택 + actor 매핑 + schema + error taxonomy + versioning + contract test 전략까지 contract-first 강제.

호출 형태:
- 단독: `/buddy:design-api-contract "<endpoint set>"`
- chain (권장): `/buddy:chain design-api-contract,write-adr -- "<endpoint set>"`

검색 API 가 필요하면 본 stage 후 `design-embedding-search` 를 invoke 해 embedding + hybrid search 를 설계.

본 stage 의 결정은 §4 plan-build 의 actor track 분배, §5 build-feature 의 contract-first codegen, §6 verify-quality 의 contract test, §7 ship-release 의 breaking change gate 로 cascade.

### Stage 5: Data Model

`design-data-model` skill 을 invoke 한다 — entity 매핑 + read/write 패턴 분류 + normalization 결정 + index 전략 + zero-downtime migration plan 까지 production 변경 비용을 사전 평가.

호출 형태:
- 단독: `/buddy:design-data-model "<entity 또는 sub-domain>"`
- chain (권장): `/buddy:chain design-data-model,write-adr -- "<entity>"`

본 stage 의 결정은 Stage 4 (API Contract) 의 resource 매핑, §5 build-feature 의 ORM model 구현, §7 ship-release 의 migration 절차로 cascade 된다.

산출물 예시 (DDL + index 근거 포함):

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

`write-adr` skill 을 invoke 한다 — 각 주요 결정 (Stage 3 tech stack, Stage 4 API contract, Stage 5 data model, Stage 6 auth model 등) 을 표준 7 섹션 (Status/Context/Decision/Consequences/Alternatives/References) 양식으로 영속화 + supersede 체인 + Index 갱신.

호출 형태:
- 결정 skill 후 chain (권장): `/buddy:chain define-tech-stack,write-adr -- "<feature>"` — 결정 즉시 영속화
- 단독: `/buddy:write-adr "<title 또는 결정 요약>"` — ad-hoc / retrospective ADR

본 stage 의 산출물은 `docs/adr/NNNN-<slug>.md` + `docs/adr/README.md` Index 갱신. 미래의 maintainer / 신규 팀원이 "왜 이 결정을 했는지" 를 ADR 1 파일로 재구성 가능하게 보장.

### Stage 10: autoplan Review

`autoplan` (cross-phase review sub-orchestrator)을 invoke해 technical design 산출물을 4-mode review한다.
- `review-architecture` — 구조적 무결성
- `review-engineering` — 구현 가능성
- `review-design` — 디자인 차원
- `review-devex` — developer-facing이면 DX 검토

`consult-design-system` skill로 UI 설계가 필요하면 design system을 생성한다.

---

## 다음 phase

- `/buddy:plan-build` — 4단계 Implementation Plan (권장)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` §§3, §3.1
- Use case → infra 브릿지 설계: 동 문서 §4 §3
