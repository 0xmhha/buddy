# design-system — 3단계 Technical Design Orchestrator

3단계 라이프사이클 단계의 진입점. Feature backlog (actor / use case / system boundary 포함) → Tech stack ADR + infra blueprint + API/data model.

**진입 조건**: 2단계 feature backlog 확정 (actor + use case + system boundary 포함).
**산출물**: Tech stack ADR, infra topology diagram, API contract, data model schema.
**다음 phase**: Technical design 확정 후 → `plan-build` (4단계).

---

## Stage 흐름

```
design-system (3단계 phase orchestrator)
├── stage 1: map-use-cases-to-infra     [Done] use case → infra component bidirectional matrix + cross-actor shared ownership + compliance scope
├── stage 2: derive-system-topology     [Done] actor 그래프 + infra 매핑 → 시스템 토폴로지 (mermaid + JSON) + trust boundary
├── stage 3: define-tech-stack          [Done] 언어/프레임워크/DB 선택 — 락인 영향 평가
├── stage 4: design-api-contract        [Done] REST/GraphQL/RPC 계약 — actor 간 경계 = API 경계
├── stage 5: design-data-model          [Done] 스키마/마이그레이션/인덱싱
├── stage 5a: design-event-schema       [Done] async event schema-first — design-api-contract sync gap 보강
├── stage 6: design-auth-model          [Done] 5 axis (authn/session/authz/federation/MFA) — RBAC + JWT + SAML + WebAuthn
├── stage 6a: design-tenant-model       [Done] multi-tenant 격리 — shared (RLS) vs schema-per vs DB-per + 3 layer defense
├── stage 7: [design-observability]     (로깅/메트릭/트레이싱 표준) — partially covered by define-tech-stack OTel decision + Phase 5 ext candidate
├── stage 8: [design-deploy-strategy]   (배포 전략 — canary/blue-green/rolling) — partially covered by setup-canary-deploy (§7)
├── stage 9: write-adr                  [Done] Architecture Decision Record
└── stage 10: autoplan                  [Done] technical design 산출물 4-mode review
```

> [Done] = v1.0.6 기준 구현 완료. 브라켓(`[name]`)은 미구현 — Phase 5 extension Cluster B 후보 또는 orchestrator 가 임시로 직접 수행.

## 권장 호출 패턴 (Phase 1 핵심 4 skill chain)

대부분의 §3 작업은 다음 chain 으로 cover 가능:

```bash
# 단일 feature 의 §3 결정 + 영속화 일괄
/buddy:chain define-tech-stack,design-data-model,design-api-contract,write-adr -- "<feature 요약>"
```

각 step 의 산출물이 다음 step 입력으로 흘러:
- **Stack 결정** (`define-tech-stack`) → DB / framework 선택을 후속 step 의 제약으로 전달
- **Data model** (`design-data-model`) → entity / schema / migration plan 을 API resource 매핑 입력으로 전달
- **API contract** (`design-api-contract`) → resource / operation / error / versioning 을 ADR 입력으로 전달
- **ADR** (`write-adr`) → 위 3 결정을 표준 양식으로 영속화 (`docs/adr/NNNN-*.md`)

각 결정마다 별도 ADR 가 필요하면 4개 사이에 write-adr 를 끼워 넣는 형태:

```bash
/buddy:chain define-tech-stack,write-adr,design-data-model,write-adr,design-api-contract,write-adr -- "<feature>"
```

---

## 실행 절차

### Stage 1: Use Case → Infra 매핑

`map-use-cases-to-infra` skill 을 invoke 한다 — §2 feature spec 의 actor × use case × system boundary 를 §3 의 actual infra component 로 매핑. 산출물: actor × infra bidirectional matrix + cross-actor shared ownership + compliance scope (encryption / RLS / audit retention / GDPR). 후속 stage 의 입력 schema.

호출 형태:
- 단독: `/buddy:map-use-cases-to-infra "<project name>"`
- chain (권장): `/buddy:chain define-tech-stack,map-use-cases-to-infra,derive-system-topology,design-data-model,design-api-contract,write-adr -- "<project>"`

본 stage 가 Q8=(a) cascade 의 §2 → §3 transition 을 명시 layer 로 채워 silent gap 차단. 산출물의 cross-actor shared infra ownership 이 §4 plan-build 의 cross-track contract 의 source.

`design-mcp-server` skill을 invoke해 MCP 연동이 필요한 external SaaS 를 추가 식별 (해당 시).

### Stage 2: 시스템 토폴로지 자동 도출

`derive-system-topology` skill 을 invoke 한다 — Stage 1 의 actor × infra matrix 를 입력으로 시스템 토폴로지 자동 도출. 산출물: service map + data flow + trust boundary diagram (mermaid + JSON 둘 다). edge type 7 분류 (sync API / async event / DB W / DB R / cache / admin / observability), trust boundary 4 layer (public / VPC public / VPC private / DB subnet), boundary violation check.

호출 형태:
- 단독: `/buddy:derive-system-topology "<project>"`
- chain: 위 권장 chain 의 일부

본 stage 의 산출물은 (a) human-readable mermaid (시스템 한 화면 이해), (b) machine-readable JSON (후속 design-* skill 이 edge type filter 로 자기 영역 추출). 후속 design-data-model 은 db-write/db-read edge, design-api-contract 는 sync-api edge 만 필터링.

`verify-best-alternative` skill 을 invoke해 토폴로지 후보 2-3 개 생성 + trade-off 비교 (high-stakes 의 경우 — AI 편향 방지 의무).

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
