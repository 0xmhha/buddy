# Engineering Phases — Artifact-Based Definition

> **목적**: 9-phase 라이프사이클의 각 단계를 **산출물(artifact) 기준**으로 정의한다.
> 각 phase는 "무슨 산출물을 입력받고, 무슨 산출물을 생산하며, 무슨 의사결정을 내리는가"로 정체성이 결정된다.
>
> **용도**:
> - 개별 스킬의 Input/Output Contract 작성 시 기준 문서
> - 스킬 간 cascade 연결의 근거 (A의 output = B의 required input)
> - `status` 스킬의 artifact 탐지 로직 근거
>
> **관계**:
> - [`skill-catalog.md`](./skill-catalog.md) — 9-phase별 스킬 목록 (what)
> - [`routing-rules.md`](./routing-rules.md) — 스킬 간 충돌 결정 (which)
> - 본 문서 — phase 정체성 + 산출물 계약 + 전이 규칙 (why & when)

---

## 목차 (Table of Contents) — B-M4 부분 참조 효율 보강

| 섹션 / 헤더 | 내용 |
|------|------|
| `## §1. 정의 원칙` | Artifact-based 채택 근거 + Phase 간 흐름 다이어그램 |
| `## §2. Phase 정의` | 각 Phase의 정체성·핵심 질문·orchestrator·Input/Output·소속 스킬 |
| └ `### Phase 1 — ...` (Mode A, Mode B 2 sub-block) | concretize-idea + Stage 매핑 / assess-product-change + scope·routing |
| └ `### Phase 2 — Requirements Specification` | define-features + actor/use case |
| └ `### Phase 3 — Software Design` | design-system + 9 카테고리 (3-A~3-I) |
| └ `### Phase 4 — Iteration Planning` | plan-build + Task DAG |
| └ `### Phase 5 — Construction` | build-feature + developer-authored tests **(Phase 6과 책임 경계)** |
| └ `### Phase 6 — Verification & Validation` | verify-quality + 상용 quality bar **(Phase 5와 책임 경계)** |
| └ `### Phase 7 — Release & Deployment` | ship-release + launch readiness **(UAT 위치 + Phase 6 경계)** |
| └ `### Phase 8 — Operation & Maintenance` | iterate-product + Engineering/Product/Marketing |
| └ `### Phase 9 — Retirement / Decommissioning` | manage-lifecycle + deprecation **(Mode B와 분류 정책)** |
| └ `### Cross-cutting (Phase 소속 없음)` | phase 무관 스킬 (decompose-blocker, status 등) |
| `## §3. Phase 전이 규칙` | 정상 흐름 / Backtrack / Skip vs Routing / 전이 판단 기준 |
| `## §4. Input/Output Contract 표준 형식` | Input Requirements 표 / Output Contract 표 / Type 분류 / 미제공 시 처리 |
| `## §5. 본 문서의 범위와 한계` | 범위 / 범위 외 / 변경 trigger |

> **부분 참조 가이드**: LLM이 본 문서를 부분 참조할 때, 위 TOC로 필요한 헤더 텍스트를 먼저 식별하고 해당 헤더로 jump하라. 본 문서는 800+ 줄이지만 phase별 sub-section은 60-100줄 단위로 독립적이므로 필요한 phase만 로드 가능.

---

## §1. 정의 원칙

### 왜 Artifact-based인가

소프트웨어 공학의 phase 정의 접근은 세 갈래가 있다:

| 접근 | 정의 기준 | 장점 | 단점 |
|------|----------|------|------|
| Activity-based | 뭘 하는가 | 직관적 | 순차적으로 보이지만 실제로는 아님 |
| **Artifact-based** | **뭘 만드는가** | **진입/종료 조건 명확, Input/Output Contract와 자연 정합** | phase 경계에서 artifact 정의 합의 필요 |
| Decision-based | 뭘 결정하는가 | 비가역성 기반 리스크 관리 | 결정이 phase를 넘나드는 경우 분류 모호 |

buddy는 **Artifact-based를 주축**으로 채택한다.

- 스킬의 Input/Output Contract가 artifact 단위이므로, phase 정의도 동일 단위로 통일
- cascade 연결이 "A의 output artifact → B의 required input artifact"로 기계적 도출 가능
- 진입 조건 판단이 "이 artifact가 존재하는가?"로 통일 (ADR-007 stateless 원칙 호환)

각 phase의 Activity(뭘 하는가)와 Decision(뭘 결정하는가)도 참조 정보로 함께 기술하되, **정체성의 근거는 artifact**이다.

### 라이프사이클은 Artifact 의존성 그래프(DAG)다

phase는 **시간순 단계가 아니라 산출물 의존성 그래프의 노드**다. 노드는 자신이 생산하는 산출물(output)로 명명되며, 엣지 `A ──► B`는 "A의 output = B의 required input"을 뜻한다. 각 노드의 **Input Artifacts 표 = Definition of Ready(DoR)**, **Output Artifacts 표 = Definition of Done(DoD)** 다 (§2).

- **진입점**은 사용자가 고르는 게 아니라 "지금 존재하는 artifact의 frontier"로 계산된다 (`status` 스킬이 artifact 탐지로 자동 추론).
- **prerequisite(DoR) 미충족** 시 그 input을 생산하는 upstream 노드로 자동 선행한다 (§3 backtrack의 재정의 — 예외가 아니라 정상 빌드 규칙).
- **사이클**은 trigger(인시던트·실험 결과·변경 요청)가 새 source를 주입하면 downstream 노드만 증분 재평가되는 것이다 (Make/Bazel의 stale-rebuild 의미론).
- "Phase 1~9" 번호는 그래프를 사람이 읽기 쉽게 묶은 **클러스터 라벨**이며 강제 실행 순서가 아니다. 아래 표기는 *위상 정렬(topological order)의 한 예시*일 뿐이다.

노드 표준 용어(SE/Agile) + DoR/DoD 매핑 SSoT는 [`se-lifecycle-naming.md §1`](./se-lifecycle-naming.md)을 따른다.

```
trigger: idea | change request | incident | metric
   │ (어떤 source + 어떤 artifact가 이미 존재하나 → 진입 노드 결정)
   ▼
Discovery ─(PRD)─► Requirements Spec ─(SRS)─► Software Design ─(SDD)─► Iteration Planning
   ─(task DAG)─► Construction ─(code+tests)─► V&V ─(evidence)─► Release & Deployment
   ─(deployed)─► Operation & Maintenance ─(improvement)─► [Requirements Spec로 역류]
                                          └─(EOL 결정)─► Retirement
   ※ ─(…)─ = "왼쪽 노드의 DoD = 오른쪽 노드의 DoR"
```

---

## §2. Phase 정의

### Phase 1 — Discovery / Impact Analysis (문제·기회 검증 / 변경 영향 분석)

| 항목 | 내용 |
|------|------|
| **정체성** | 문제 또는 기회의 존재를 확인하고, 해결/실행할 가치가 있는지 검증한다 |
| **핵심 질문** | "무엇이 문제(또는 기회)이고, 해결할 가치가 있는가?" |

Phase 1은 **프로덕트 존재 여부**에 따라 2가지 mode로 동작한다. 어떤 mode든 종료 시 동일한 output 형태 — **검증된 작업 항목(validated work item) + 다음 phase 결정** — 을 생산한다.

#### Mode A — Discovery / Inception (Greenfield, 신규 프로덕트)

기존 프로덕트가 없고, 아이디어 단계에서 시작하는 경우.

| 항목 | 내용 |
|------|------|
| **Orchestrator** | `concretize-idea` |
| **연상** | 막연한 아이디어를 concrete(구체적)하게 굳힌다 |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Problem statement / idea sketch | ✅ | 사용자가 제공하는 비정형 입력 (대화, 메모, 한 줄 아이디어) |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| PRD (Product Requirements Document) | `docs/prd.md` — actors + use cases(logical) 포함 | Phase 2 `define-features`, Phase 1 `write-hld` |
| HLD (High Level Design) | `docs/hld.md` — 9 섹션 (product 구성 + tech stack + use case → product mapping) | Phase 1 `autoplan`(validate-spec), Phase 3 `design-system` |
| Business viability report | PRD 내 섹션 또는 별도 문서 | Phase 1 내부 결정 근거 |
| Market/competitor analysis | PRD 내 섹션 | Phase 1 내부 결정 근거 |
| Customer segment map | PRD 내 섹션 | Phase 1 내부 (Stage 4-6 입력). Phase 2 actor 정의의 입력으로도 사용. **이관 결정 근거**: 2026-05-29 `concretize-idea` 재구조화 — 경쟁/사업성/pricing(Stage 4-6)이 customer 정의를 입력으로 요구하므로 Phase 1 내부 Stage 3에 배치 (`plugin/skills/concretize-idea/PROCEDURE.md` "중요 변경 (2026-05-29)" 노트 참조) |
| PRD + HLD review report (autoplan) | structured findings | Phase 2 `define-features`로 forward 또는 PRD/HLD 수정 |

**종료 → 다음**: Phase 2 `define-features` (전체 흐름 진입). PRD + HLD가 review 통과 후.

**Mode A 소속 스킬 + concretize-idea Stage 매핑:**

| Skill | concretize-idea Stage | 역할 |
|-------|---------------------|------|
| `validate-idea` | Stage 1 | 아이디어 stress-test (6 forcing questions) |
| `validate-advanced-edge-idea` | Stage 2 | edge case / hidden assumption grilling |
| `map-customer-segments` | Stage 3 | 사용자 vs 구매자 분리 + persona — 후속 stage 입력 |
| `analyze-competition-and-substitutes` | Stage 4 | 경쟁/대체재 매트릭스 |
| `assess-business-viability` | Stage 5 | 7차원 사업성 평가 (gate: 치명적 결함 시 사용자 확인) |
| `review-pricing-and-gtm` | Stage 6 | pricing + GTM channel 평가 (상업 프로젝트만, 비상업은 skip) |
| `define-product-spec` | Stage 7 | PRD 작성 (내부에서 `identify-actors` + `map-actor-use-cases` 호출) |
| `write-hld` | Stage 8 | High Level Design — product 구성 + tech stack + use case → product mapping |
| (autoplan) | Stage 9 | PRD + HLD 통합 4-mode review (review-scope / design / devex / engineering) |

**Mode A supporting skills (concretize-idea 직접 stage 매핑 없음 — 옵션/심층 분석용):**

| Skill | 호출 패턴 | 역할 |
|-------|---------|------|
| `analyze-market-size` | `assess-business-viability` 내부 또는 사용자 직접 호출 | TAM/SAM/SOM 산출 — Stage 5의 입력 보강 |
| `map-jobs-to-be-done` | `map-customer-segments` 내부 또는 사용자 직접 호출 | JTBD 프레임워크 — Stage 3 customer 정의 심화 |
| `conduct-customer-interview` | 사용자 직접 호출 (Stage 3-5 사이 데이터 수집 시) | 고객 인터뷰 + Mom Test |
| `decide-target-market` | 사용자 직접 호출 또는 `assess-business-viability` 후 region 결정 시 | target market 결정 + region trigger |

> 매핑 SSoT: `plugin/skills/concretize-idea/PROCEDURE.md` "Stage 흐름" 섹션. 본 표는 그것의 reverse index.

#### Mode B — Impact Analysis (Existing Product, 기존 프로덕트 변경)

이미 운영 중인 프로덕트에 대한 모든 변경 작업 — 버그 수정, 신규 기능, 성능 개선, 기술 부채 정리, 의존성 갱신 등 유형 무관. 핵심은 **변경의 scope(범위)를 평가**하여 다음 phase를 결정하는 것.

| 항목 | 내용 |
|------|------|
| **Orchestrator** | `assess-product-change` *(신규)* |
| **연상** | 기존 프로덕트에 대한 변경(change)을 평가(assess)한다 |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Change trigger | ✅ | 버그 리포트, 기능 요청, 성능 메트릭, 기술 부채 신호 등 — 유형 무관 |
| Existing codebase / product context | ✅ | 현재 아키텍처, 코드, 사용자 기반, 기술 스택 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Validated work item | structured (문제/기회 설명 + 재현/근거 + 수용 기준) | 다음 phase 진입 스킬 |
| Impact assessment | 영향 범위 + severity/priority + 기존 시스템 호환성 | 다음 phase 결정 근거 |
| Scope classification | decision — **3종**: small / medium / large | Routing decision의 입력 |
| Routing decision | decision — **4종 옵션**: §2 / §3 / §5 / defer (scope 3종 + Defer/Reject) | Phase 전이 |

**종료 → scope별 routing:**

| Scope | 다음 경로 | 분류 기준 (`assess-product-change` Step 4) | 예시 |
|-------|----------|----------------------------------------|------|
| **Small** | → Phase 5 직접 진입 | 1-3 파일 변경, non-breaking, 기존 테스트 유지, 설계 변경 없음 | 버그 수정, 설정 변경, 작은 UI 수정 |
| **Medium** | → Phase 3 (설계 검토 후 구현) | 4-15 파일 변경, 또는 새 모듈/API 추가, 또는 스키마 변경, 설계 검토 필요 | 새 API endpoint, 컴포넌트 리팩토링, 스키마 변경 |
| **Large** | → Phase 2 (feature 정의부터) | 15+ 파일 변경, 또는 아키텍처 변경, 또는 새 actor/use case 도입, feature 정의 필요 | 신규 기능, 대규모 재설계, 아키텍처 변경 |
| **Defer/Reject** | → backlog 기록, 현재 cycle 종료 | scope 분류 자체가 아닌 routing 결정의 4번째 옵션 (우선순위/ROI 평가 결과) | 우선순위 낮음, ROI 부족 |

> 상세 기준은 `plugin/skills/assess-product-change/PROCEDURE.md` Step 4 (Scope 분류) 참조.

---

### Phase 2 — Requirements Specification (기능 정의)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Requirements Specification (SWEBOK · IEEE 830/29148) — Output 산출물 **SRS**가 이름에 박혀 있음 |
| **정체성** | PRD를 actor/use case 분해를 거쳐 구현 가능한 feature 단위로 변환한다 |
| **핵심 질문** | "무엇을 만드는가?" |
| **Orchestrator** | `define-features` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| PRD (Mode A 경유) 또는 Validated work item (Mode B, scope=large) | ✅ | Phase 1 산출물 — greenfield는 PRD, 기존 프로덕트 대규모 변경은 validated work item |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Actor list | structured (PRD 내 또는 별도) | Phase 2 내부, Phase 3 `design-api-contract` |
| Use case map (actor별) | structured | Phase 2 내부, Phase 4 `decompose-feature-to-actor-tracks` |
| System boundary map | structured | Phase 3 `derive-system-topology` |
| Feature specs | `docs/feature-spec/` | Phase 3 `design-system`, Phase 4 `plan-build` |
| Feature backlog (priority + estimate) | structured | Phase 4 `plan-build` |

**종료 조건**: Feature backlog이 priority-ordered 상태로 존재하고, 각 feature에 actor/use case/acceptance criteria가 정의된 상태.

**소속 스킬:**

| Skill | 역할 |
|-------|------|
| `identify-actors` | 시스템 참여 actor 열거 (user/system/3rd-party/external-tool) |
| `map-actor-use-cases` | actor별 use case 식별 |
| `map-use-case-to-system-boundary` | use case → 시스템 경계 매핑 |
| `compose-feature-from-use-cases` | cross-actor use case 합성 → feature 정의 |
| `define-feature-spec` | feature 완전 명세서 작성 |
| `score-feature-priority` | RICE/ICE/MoSCoW 우선순위 |
| `estimate-feature-effort` | T-shirt sizing + ideal-h + uncertainty |
| `map-feature-dependencies` | feature 간 의존성 DAG |
| `split-work-into-features` | PRD → vertical slice 분해 |
| `query-feature-registry` | 기존 feature registry 검색 (reuse 판단) |
| `triage-work-items` | work item 분류 + lifecycle state machine |

---

### Phase 3 — Software Design (기술 설계)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Software Design (SWEBOK · IEEE 1016) — Output 산출물 **SDD**가 이름에 박혀 있음 |
| **정체성** | feature를 구현하기 위한 기술적 구조를 결정한다 |
| **핵심 질문** | "어떻게 만드는가?" |
| **Orchestrator** | `design-system` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Feature specs / backlog | ✅ | Phase 2 산출물 (정상 흐름) |
| Validated work item (Mode B, scope=medium) | ✅ | Phase 1 `assess-product-change`에서 직접 진입 시 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Tech stack ADR | `docs/decisions/` | Phase 5 구현 기준 |
| API contracts | `docs/api-contract/` 또는 OpenAPI spec | Phase 5 `generate-from-api-contract` |
| Data model (schema + migration plan) | `docs/data-model/` | Phase 5 구현, Phase 6 테스트 |
| Architecture Decision Records | `docs/decisions/` | 전 phase 참조 |
| Design system / interaction patterns | 별도 문서 | Phase 5 UI 구현 |
| Deploy strategy | ADR 또는 별도 문서 | Phase 7 `ship-release` |
| Observability / secret / auth strategy | 각 별도 문서 | Phase 5 구현, Phase 8 운영 |

**종료 조건**: 핵심 기술 결정(tech stack, API, data model)이 ADR로 기록되고, 리뷰(autoplan 4-mode)를 통과한 상태.

**호출 패턴**: 각 design-* 스킬은 독립 호출 가능(standalone-with-context). orchestrator(`design-system`) 경유 시 결정 간 일관성 보장 (예: tech stack 결정과 data model 설계가 모순되지 않도록).

**소속 스킬** (Phase 3은 design 스킬이 가장 많은 phase. 카테고리별 그룹화로 정리):

**3-A. Architecture Foundation (5)** — 핵심 기술 결정

| Skill | 역할 |
|-------|------|
| `define-tech-stack` | 기술 스택 결정 (8차원 + 5년 lock-in 평가) |
| `derive-system-topology` | 시스템 토폴로지 도출 (service/data flow/trust boundary) |
| `decide-form-factor-app-vs-web` | 앱 vs 웹 폼팩터 결정 |
| `map-use-cases-to-infra` | use case → infra 매핑 |
| `write-adr` | ADR 작성 |

**3-B. Data & Contract Design (4)** — 인터페이스/스키마

| Skill | 역할 |
|-------|------|
| `design-data-model` | 데이터 모델 설계 (entity + read/write 패턴 + migration) |
| `design-api-contract` | API 계약 설계 (REST/GraphQL/gRPC + schema + error) |
| `design-event-schema` | 이벤트 스키마 설계 (async/pub-sub/DLQ) |
| `design-embedding-search` | 하이브리드 검색 설계 (BM25 + vector + rerank) |

**3-C. Security & Tenancy (3)** — 보안/멀티테넌시/시크릿

| Skill | 역할 |
|-------|------|
| `design-auth-model` | 인증/인가 모델 설계 (OAuth2/RBAC/multi-tenant) |
| `design-tenant-model` | 멀티테넌트 모델 설계 (RLS/schema-per/DB-per) |
| `design-secret-management` | 시크릿 관리 전략 (rotation/audit/leak detection) |

**3-D. Operations Strategy (3)** — 관측성/배포/저장

| Skill | 역할 |
|-------|------|
| `design-observability` | 관측성 전략 (logs/metrics/traces/SLO) |
| `design-deploy-strategy` | 배포 전략 (canary/blue-green/rolling) |
| `design-artifact-storage` | artifact 저장/검증/배포 설계 |

**3-E. Quality Baseline (2)** — i18n/접근성

| Skill | 역할 |
|-------|------|
| `design-i18n-strategy` | i18n 전략 (locale/fallback/RTL) |
| `design-accessibility-baseline` | 접근성 기준선 (WCAG/a11y) |

**3-F. UX/UI Design (4)** — 디자인 시스템·인터랙션·프로토타입

| Skill | 역할 |
|-------|------|
| `apply-design-system` | 디자인 시스템 적용 (token/component/pattern) |
| `consult-design-system` | 디자인 시스템 생성/참조 |
| `design-interaction-pattern` | 인터랙션 패턴 설계 (gesture/motion/feedback) |
| `prototype-from-spec` | spec → 프로토타입 |

**3-G. Domain-Specific Design (3)** — 결제·MCP·Claude hook

| Skill | 역할 |
|-------|------|
| `design-billing-system` | 결제 시스템 설계 (Stripe/Toss + subscription) |
| `design-mcp-server` | MCP 서버 설계 |
| `design-claude-hooks` | Claude Code hook 설계 |

**3-H. Design Reviews (6)** — autoplan 4-mode + 기타

| Skill | 역할 |
|-------|------|
| `review-architecture` | 아키텍처 구조 무결성 검토 |
| `review-engineering` | implementation plan 리뷰 |
| `review-scope` | scope 형성/결정 리뷰 |
| `review-design` | 디자인 차원 0-10 score 리뷰 |
| `review-devex` | DX plan 리뷰 |
| `audit-ui-quality` | UI 품질 감사 |

**3-I. Decision Support (3)** — 검토 도구·외부 의견

| Skill | 역할 |
|-------|------|
| `consult-codex` | 외부 LLM second opinion |
| `verify-best-alternative` | 엔지니어링 결정의 다관점 검토 (편향 방지) |
| `critique-plan` | 전략적 plan critique (CEO/founder 페르소나) |

---

### Phase 4 — Iteration Planning (구현 계획)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Iteration Planning / Work Breakdown (Scrum · PMBOK) — Output 산출물 **Iteration Plan(task DAG)** |
| **정체성** | 기술 설계를 실행 가능한 task 단위로 분해하고 순서를 정한다 |
| **핵심 질문** | "어떤 순서로 만드는가?" |
| **Orchestrator** | `plan-build` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Technical design docs (ADR, API contract, data model) | ✅ | Phase 3 산출물 |
| Feature specs | ✅ | Phase 2 산출물 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Actor-track decomposition | `docs/actor-track-plan.yaml` | Phase 5 `build-feature` |
| Task DAG (dependency graph) | structured | Phase 5 `dispatch-parallel-agents` |
| Parallel execution plan | structured | Phase 5 병렬 개발 |
| Acceptance test plan | structured | Phase 5 `generate-tests-from-spec`, Phase 6 검증 기준 |
| Build timeline (estimate) | structured | 프로젝트 관리 |

**종료 조건**: task DAG가 존재하고, critical path가 식별되며, acceptance test plan이 정의된 상태.

**소속 스킬:**

| Skill | 역할 |
|-------|------|
| `decompose-feature-to-actor-tracks` | feature → actor별 implementation track 분해 |
| `decompose-track-to-tasks` | actor track → atomic task list |
| `map-task-dependencies` | task DAG (내부 + cross-actor 의존성) |
| `plan-parallel-execution` | worker batch + sync point 계획 |
| `define-acceptance-test-plan` | per-actor + cross-actor test plan |
| `estimate-build-timeline` | critical path 기반 calendar timeline |
| `publish-to-tracker` | task plan → 외부 issue tracker 발행 |

---

### Phase 5 — Construction (개발)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Construction / Implementation (SWEBOK Software Construction) — Output 산출물 **working code + developer tests** |
| **정체성** | 계획된 task를 **코드 + 자체 테스트(developer-authored)**로 구현하고, 해당 자체 테스트가 통과하는 상태까지 만든다 |
| **핵심 질문** | "구현했고, 자체 테스트는 통과하는가?" |
| **Orchestrator** | `build-feature` |

**Phase 6과의 책임 경계** (B-L8 명확화, 2026-06-01):

| Phase | 책임 테스트 종류 | 평가 기준 |
|-------|---------------|---------|
| **Phase 5** | 개발자가 코드와 함께 작성한 unit / integration / contract test | green/red (작성된 테스트가 통과하는가) |
| **Phase 6** | 상용 quality gate — coverage(line/mutation/behavior), security audit, a11y, i18n, compliance, code health, load/chaos, cross-actor E2E | quality bar 통과 여부 (단순 green이 아닌 의미 있는 coverage·취약점 제로 등) |

요약: Phase 5는 "내가 만든 테스트가 green"이면 종료. Phase 6은 "이 코드가 **상용 출시 가능 수준의 품질**인가"를 별도 차원에서 평가.

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Actor-track plan / Task DAG | ✅ | Phase 4 산출물 (정상 흐름) |
| Technical design docs | ✅ | Phase 3 산출물 (구현 기준) |
| Validated work item (Mode B, scope=small) | ✅ | Phase 1 `assess-product-change`에서 직접 진입 시 — plan/design 없이 work item만으로 구현 |
| Acceptance test plan | 선택 | Phase 4 산출물 (TDD 시 활용) |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Working code | source files + commit history | Phase 6 검증 대상 |
| Developer-authored tests (unit / integration / contract) | test files | Phase 6 `audit-test-coverage-meaningful` (의미 있는 coverage 평가의 입력) |
| Updated docs (코드 변경 동기화) | docs/ 갱신 | Phase 7 release docs |

**종료 조건**: 모든 task가 완료되고, **개발자가 작성한 자체 테스트(unit/integration/contract)가 green**이며, 코드가 commit된 상태. (상용 quality bar 통과 여부는 Phase 6의 책임 — 본 phase 종료 조건에 포함되지 않음.)

**특이사항**: Phase 5의 스킬 다수(build-with-tdd, diagnose-bug, iterate-fix-verify)는 full-standalone 등급 — 별도 orchestrator 없이 독립 실행이 자연스러운 영역.

**소속 스킬:**

| Skill | 역할 |
|-------|------|
| `build-with-tdd` | red-green-refactor TDD 루프 |
| `diagnose-bug` | 버그 재현 → 원인 분석 → fix |
| `iterate-fix-verify` | finding별 fix → atomic commit → re-verify 반복 |
| `pair-program-loop` | driver/navigator 역할 분리 + 15min swap |
| `refactor-with-rename-trace` | LSP rename + 호출 그래프 검증 |
| `dispatch-parallel-agents` | worktree 격리 병렬 agent 분배 |
| `generate-from-api-contract` | API contract → SDK/stub 자동 생성 |
| `generate-tests-from-spec` | acceptance criteria → test skeleton |
| `freeze-edit-scope` | 단일 디렉토리 edit lock |
| `update-docs-with-code` | 코드 변경 → 5영역 문서 동기화 |

---

### Phase 6 — Verification & Validation (검증·확인)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Verification & Validation, V&V (IEEE 1012) — Output 산출물 **V&V evidence**. Verification("올바르게 만들었나") + Validation("올바른 것을 만들었나") 둘 다 포함 |
| **정체성** | Phase 5 산출물(코드 + 자체 테스트)이 **상용 출시 가능한 quality bar**(meaningful coverage, security, a11y, i18n, compliance, code health, performance)에 도달했는지 별도 차원에서 평가한다 |
| **핵심 질문** | "단순히 작동하는 것을 넘어 출시할 수 있는 품질인가?" |
| **Orchestrator** | `verify-quality` |

**Phase 5와의 책임 경계** (B-L8 명확화, 2026-06-01): Phase 5는 "내가 작성한 테스트가 green"으로 충분. Phase 6은 그 위에 **별도 quality 차원**을 평가 — 작성된 테스트의 의미·취약점·접근성·법규 등.

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Working code + developer-authored tests | ✅ | Phase 5 산출물 (자체 테스트 green 상태로 인계) |
| Acceptance criteria / test plan | ✅ | Phase 4 산출물 또는 feature spec |
| API contracts / data model | 선택 | Phase 3 산출물 (contract test 기준) |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| QA report (per-actor + cross-actor) | structured | Phase 7 launch checklist 입력 |
| Security audit report | structured | Phase 7 launch checklist 입력 |
| Compliance sign-off (privacy, license, terms) | structured | Phase 7 launch checklist 입력 |
| Code health score | 0-10 composite | Phase 7 release 판단 |
| Coverage report (line + mutation + behavior) | structured | Phase 5 backtrack 시 보강 기준 |

**종료 조건**: 모든 quality gate(test, security, compliance, code health)가 통과한 상태.

**Backtrack trigger**: quality gate 실패 → Phase 5로 복귀 (fix and re-verify).

**소속 스킬:**

| Skill | 역할 |
|-------|------|
| `classify-qa-tiers` | QA intensity 3-tier 분류 (Quick/Standard/Exhaustive) |
| `test-per-actor-use-case` | actor별 use case 단위 통합 테스트 |
| `test-cross-actor-flow` | cross-actor E2E flow 검증 |
| `run-load-test` | sustained/soak/spike/stress 4 시나리오 |
| `chaos-test` | failure injection + hypothesis-driven 검증 |
| `run-browser-qa` | 브라우저 자동화 QA (snapshot/form/responsive) |
| `audit-test-coverage-meaningful` | line + mutation + behavior coverage |
| `measure-code-health` | composite 0-10 code health dashboard |
| `audit-security` | CSO-mode 보안 감사 |
| `audit-accessibility` | WCAG 2.1 AA a11y 감사 |
| `audit-i18n-coverage` | locale별 번역 커버리지 |
| `audit-cost-efficiency` | per-component 비용 분석 |
| `audit-live-devex` | 실제 DX audit (TTHW timing) |
| `audit-ubiquitous-language` | 코드/PRD/도메인 어휘 일관성 감사 |
| `classify-review-risks` | 11 category 코드 리뷰 리스크 분류 |
| `review-ai-safety-liability` | AI 기능 책임/안전 검토 |
| `review-privacy-data-risk` | GDPR/PIPA 등 개인정보 검토 |
| `review-license-and-ip-risk` | 라이선스/IP 호환성 검토 |
| `review-terms-policy-readiness` | 약관/정책 준비도 검토 |

---

### Phase 7 — Release & Deployment (출시·배포)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Release & Deployment (DevOps · ITIL) — Output 산출물 **release package + deployed artifact** |
| **정체성** | 검증된 코드를 사용자에게 전달한다 (패키징/태깅/배포/공지/launch readiness 확인 포함). **UAT는 본 phase의 launch readiness 일환**으로 분류 (Phase 6 quality bar와 별개로, 실사용자 acceptance 관점 검증) |
| **핵심 질문** | "사용자에게 안전하게 전달할 준비가 되었고, 전달했는가?" |
| **Orchestrator** | `ship-release` |

**Phase 6 vs Phase 7 UAT 경계** (B-L9 명확화 2026-06-01):

| Phase | 검증 관점 | 책임 스킬 |
|-------|---------|---------|
| Phase 6 (Verification) | 코드의 **상용 품질 bar** (coverage / security / a11y / compliance / code health) | `verify-quality`, `audit-*`, `review-*` |
| Phase 7 (Release) | **실사용자 acceptance**: UAT, beta program, canary, rollback runbook | `run-uat`, `run-beta-program`, `setup-canary-deploy` |

즉 Phase 6은 "코드 자체"의 품질, Phase 7은 "사용자가 만났을 때"의 품질을 본다. UAT를 Phase 7로 두는 이유: 실사용자 시나리오는 deploy/canary 환경과 결합되어 평가되므로 launch readiness의 일부.

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| QA report + sign-offs | ✅ | Phase 6 산출물 |
| Working code (quality gate 통과) | ✅ | Phase 5→6 통과 산출물 |
| Changelog draft | 선택 | Phase 5 `update-docs-with-code` 산출물 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Tagged release (semver) | git tag + release notes. 버전 결정 정책: breaking change 시 major, 신규 feature 시 minor, 버그 fix/내부 개선 시 patch (semver 2.0.0 준수, `automate-release-tagging`이 변경 셋 분석 후 자동 제안) | 사용자, Phase 8 운영 기준 |
| Deployed artifact | 배포된 binary / container / package | Phase 8 모니터링 대상 |
| Launch checklist pass | structured checklist | 감사 증적 |
| Updated docs (CHANGELOG, README, ADR) | docs/ 갱신 | 사용자, 다음 cycle |

**종료 조건**: release tag가 존재하고, 배포가 완료되며, launch checklist 전 항목이 통과한 상태.

**Backtrack trigger**: UAT 실패 → Phase 5 또는 Phase 6로 복귀.

**소속 스킬:**

| Skill | 역할 |
|-------|------|
| `setup-quality-gates` | pre-commit/pre-push hook 구성 |
| `auto-create-pr` | commit → branch → PR 자동화 |
| `automate-release-tagging` | semver auto-decision + git tag |
| `sync-release-docs` | diff 기반 문서 auto-update |
| `write-changelog` | CHANGELOG release-summary |
| `guard-destructive-commands` | 위험 명령 실행 전 경고 |
| `compose-safety-mode` | 복수 safety hook 합성 |
| `run-uat` | UAT 시나리오 실행 + go/no-go |
| `run-beta-program` | 클로즈드 beta 코호트 운영 |
| `setup-canary-deploy` | canary 단계 + metric gate |
| `setup-feature-flags` | feature flag 시스템 설계 |
| `setup-rollback-runbook` | rollback decision tree + 실행 절차 |
| `prepare-launch-checklist` | 17+ 항목 cross-functional launch gate |
| `setup-incident-paging` | on-call + escalation + alert 구조 |

---

### Phase 8 — Operation & Maintenance (운영·개선)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Operation & Maintenance (ISO/IEC 12207 · 14764) — Output 산출물 **ops metrics + improvement backlog**. Lean의 Build-Measure-Learn 루프를 포함 |
| **정체성** | 운영 중인 시스템을 관찰하고, 데이터 기반으로 개선한다 |
| **핵심 질문** | "잘 돌아가고 있는가?" |
| **Orchestrator** | `iterate-product` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Production system (deployed, traffic 발생 중) | ✅ | Phase 7 산출물 |
| Monitoring data (metrics, logs, traces) | ✅ | 운영 인프라 산출물 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Experiment results (A/B, funnel) | structured report | Phase 2 재진입 (improvement → feature) |
| Incident reports + postmortem | structured | Phase 5 hotfix, Phase 3 design 재검토 |
| Improvement backlog | structured task list | Phase 2 `define-features` 재진입 |
| Operational metrics (SLO, error budget, cost) | dashboard / report | Phase 8 자체 loop |

**종료 조건**: 자연적 종료 없음 — 지속적 loop. Phase 2 재진입(개선 기능) 또는 Phase 9 진입(폐기 결정) 시 해당 cycle 종료.

**소속 스킬 (Engineering):**

| Skill | 역할 |
|-------|------|
| `handle-incident` | 인시던트 대응 (severity 분류 → 완화 → fix) |
| `conduct-postmortem` | 비난 없는 포스트모템 (5 Whys + action items) |
| `monitor-regressions` | delta-based regression 감지 |
| `analyze-actor-failure-rate` | actor failure matrix + trust score |
| `analyze-cost-anomaly` | cloud/SaaS spike 감지 + recovery |
| `audit-error-budget` | SLO burn rate multi-window 감사 |
| `summarize-retro` | git history → evidence-based 주간 회고 |
| `generate-improvement-tasks` | 분석 결과 → improvement task 변환 (§2 재진입 bridge) |

**소속 스킬 (Product/Analytics):**

| Skill | 역할 |
|-------|------|
| `design-ab-experiment` | A/B 실험 설계 |
| `analyze-ab-experiment` | A/B 실험 결과 분석 |
| `analyze-user-funnel` | funnel 전환/이탈 분석 |
| `analyze-feature-adoption` | feature adoption funnel |
| `analyze-user-cohort` | cohort retention + LTV/CAC |
| `triage-customer-support-ticket` | 지원 티켓 분류 |
| `analyze-customer-feedback-corpus` | 피드백 토픽 모델링 |
| `optimize-conversion-funnel` | AARRR funnel CRO |
| `plan-growth-experiment` | growth 실험 sprint |

**소속 스킬 (Marketing):**

| Skill | 역할 |
|-------|------|
| `draft-marketing-copy` | 마케팅 카피 작성 |
| `plan-marketing-channel` | 마케팅 채널 전략 |
| `audit-seo-aso` | SEO/ASO 감사 |
| `automate-marketing-content` | 마케팅 콘텐츠 자동화 |

---

### Phase 9 — Retirement / Decommissioning (폐기·종료)

| 항목 | 내용 |
|------|------|
| **표준 용어** | Retirement / Decommissioning (ISO/IEC/IEEE 12207 Disposal process) — Output 산출물 **deprecation·migration·EOL plan**. 구 명칭 `Lifecycle Management`는 상위 lifecycle 개념과 충돌하여 폐기 |
| **정체성** | 노후화된 feature/product의 수명을 결정하고 정리한다 |
| **핵심 질문** | "유지할 것인가, 끝낼 것인가?" |
| **Orchestrator** | `manage-lifecycle` |

**Phase 9 vs Mode B 분류 정책** (B-N11 명확화 2026-06-01): deprecation 작업은 외부 사용자 영향 유무로 분류한다.

| 작업 유형 | 처리 phase | 근거 |
|---------|----------|------|
| **외부 사용자 영향 있음** (API consumer, end user에게 마이그레이션 안내 필요) | Phase 9 `deprecate-feature` | sunset 소통·timeline·migration plan이 필요 — Phase 9의 정체성에 부합 |
| **내부만 영향** (dead code 제거, 사용 안 하는 helper 삭제, 내부 helper 통합) | Phase 1 Mode B (scope에 따라 small/medium) | 외부 announce 불필요 — 일반 변경 작업으로 처리. `assess-product-change`가 scope 평가 후 routing |

**판단 기준 예시**:
- 공개 API endpoint 삭제 → Phase 9 (API consumer 알림 + 대체 endpoint 안내)
- 내부 helper 함수 인라인화 → Mode B small (외부 영향 X, 내부 정리)
- DB 컬럼 deprecation (API 응답에서 사라짐) → Phase 9 (consumer 영향)
- 미사용 내부 util 모듈 삭제 → Mode B small

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Usage/adoption data | ✅ | Phase 8 산출물 |
| Business decision (유지/폐기) | ✅ | 사용자 의사결정 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Deprecation plan + timeline | structured | 사용자 공지, Phase 5 cleanup 코드 |
| Migration plan | structured | 사용자, 연관 시스템 |
| EOL documentation | structured | 감사 증적 |
| Knowledge preservation | docs/ archive | 다음 프로젝트 참고 |

**종료 조건**: feature/product가 sunset 되고, 사용자 migration이 완료되며, 코드 cleanup이 끝난 상태.

**소속 스킬:**

| Skill | 역할 |
|-------|------|
| `deprecate-feature` | feature sunset (timeline + notice + telemetry) |
| `migrate-customers` | 대규모 고객 마이그레이션 (Strangler Fig) |
| `archive-product` | product EOL (data export + tombstone + legal) |
| `spin-off-feature` | 기능 → 별도 product/repo 분리 |

---

### Cross-cutting (Phase 소속 없음)

어느 phase에서든 호출 가능한 스킬. phase 정체성이 아닌 **적용 맥락**으로 정의된다.

**소속 스킬:**

| Skill | 범주 | 적용 시점 |
|-------|------|----------|
| `decompose-blocker` | Blocker 분해 | 작업 stuck 상태 (자동 trigger: 3회 시도 후 미해결) |
| `status` | 현재 위치 파악 | artifact 탐지로 현재 phase 추론 |
| `write-a-skill` | 메타-스킬 | 신규 스킬 작성 |
| `apply-builder-ethos` | 철학 주입 | Boil the Lake / Search Before Building / User Sovereignty |
| `benchmark-llm-models` | LLM 도구 | multi-provider LLM 성능 비교 |
| `detect-install-type` | 설치 감지 | tool install type 자동 감지 |
| `guide-setup-wizard` | 설정 가이드 | credential/config setup flow |
| `save-context` | 문맥 보존 | 세션 전환 시 상태 저장 |
| `restore-context` | 문맥 복원 | 저장된 checkpoint 로드 |
| `persist-learning-jsonl` | 학습 저장 | JSONL append-only learning store |
| `review-legal-regulatory` | 법률/규제 검토 | §1/§7 등 복수 phase에서 호출 |

---

## §3. Phase 전이 규칙

> **DAG 관점 (2026-06-02)**: 아래 "정상 흐름 / backtrack / skip / routing"은 모두 artifact 의존성 그래프의 **단일 규칙** — *"노드의 DoR(required input)이 충족되면 실행, 미충족이면 그 input을 생산하는 upstream 노드로 선행"* — 의 표현이다. backtrack = DoR 미충족 시 upstream 선행, skip = 이미 충족된 노드 통과, routing = 충족된 frontier에서 다음 노드 선택. "정상 흐름"은 위상 정렬(topological order)의 한 예시일 뿐 강제 순서가 아니다.

### 정상 흐름 (forward)

```
§1 → §2 → §3 → §4 → §5 → §6 → §7 → §8 → §2 (개선 cycle)
                                              → §9 (폐기 결정 시)
```

### Backtrack (이전 phase 복귀)

| 현재 Phase | 복귀 대상 | 조건 |
|-----------|----------|------|
| §5 Development | §3 Design | 구현 중 설계 모호/누락 발견 |
| §6 Verification | §5 Development | quality gate 실패 (fix 필요) |
| §7 Release | §5 or §6 | UAT 실패 |
| §8 Operations | §5 Development | hotfix 필요 (인시던트) |
| §8 Operations | §3 Design | 구조적 문제 발견 (design 재검토) |

### 분기 — Skip vs Routing

phase 정상 흐름을 벗어나는 경로는 두 종류로 명확히 구분한다 (용어 정정 2026-06-01):

- **Skip**: 특정 phase를 **건너뛰는** 경우. 그 phase에 진입하지 않음.
- **Routing**: phase 내부 orchestrator의 판단에 따라 **다음 phase를 선택**하는 경우. 해당 phase에는 정상 진입함.

#### Skip (Mode A — Greenfield)

| 상황 | 경로 | 건너뛰는 phase | 근거 |
|------|------|---------------|------|
| Prototype / POC | §1 → §2 → §3 → §5 | §4 | 형식적 task 분해 불필요 |
| Design-only 작업 | §1 → §2 → §3 → 종료 | §4-§9 | 구현 없이 설계 문서만 산출 |

#### Routing (Mode B — Existing Product, scope-based)

`assess-product-change`(Phase 1 Mode B orchestrator)의 scope 판단 결과에 따라 다음 phase를 선택. **§1은 실행되며, 건너뛰지 않음**.

| Scope | 경로 | §1 다음 phase | 분류 기준 (`assess-product-change` Step 4) | 예시 |
|-------|------|--------------|----------------------------------------|------|
| Small | §1 → §5 | §5 | 1-3 파일, non-breaking, 기존 테스트 유지, 설계 변경 없음 | 버그 수정, 설정 변경, 작은 UI 수정 |
| Medium | §1 → §3 → §5 → ... | §3 | 4-15 파일, 또는 새 모듈/API 추가, 또는 스키마 변경, 설계 검토 필요 | 새 API endpoint, 스키마 변경, 컴포넌트 리팩토링 |
| Large | §1 → §2 → §3 → §4 → §5 → ... | §2 | 15+ 파일, 또는 아키텍처 변경, 또는 새 actor/use case 도입, feature 정의 필요 | 신규 기능, 대규모 재설계 |
| Defer/Reject | §1 종료 | (없음) | scope 분류와 무관 — ROI/우선순위 평가 결과 | 우선순위 낮음, ROI 부족 — backlog 기록 |

**Medium-Large 경계 케이스** (B-N10 명확화 2026-06-01): 다음 중 **하나라도 해당하면 Large**로 분류 (보수적 판단):
- 새 actor 또는 use case 등장 → Phase 2 feature 정의 필요
- 아키텍처 결정이 ADR을 요구할 수준 → Phase 3 design 진입만으론 부족
- cross-team 협업 필요 → Phase 2 backlog 단위 분해 필요

반대로 단순한 파일 수만으로는 Large로 분류하지 않는다 (예: 자동 생성 코드 30 파일 변경은 Medium 유지). 판단 책임은 `assess-product-change` orchestrator + 사용자 확인.

#### Skip — 공통 (긴급)

| 상황 | 경로 | 건너뛰는 phase | 근거 |
|------|------|---------------|------|
| Hotfix (긴급) | §5 직접 진입 | §1-§4 전부 | **발동 기준: 24시간 내 fix가 필요한 작업** (사용자/이해관계자의 timeline 요구). severity와 무관하게 시간 압박만으로 결정 — production incident, 외부 demo 직전 발견된 결함, 법적 deadline 등. **Mode B small과의 차이**: Mode B small도 버그 수정이지만 normal cycle(§1 assess → §5)을 따름. hotfix는 §1 자체를 건너뛰어 즉시 §5 진입, **사후 24시간 내 incident report/postmortem으로 §1을 보상**해야 함 |

### 전이 판단 기준

phase 전이는 **output artifact의 존재 + 품질**로 판단한다:

1. 현재 phase의 output artifact가 존재하는가?
2. 해당 artifact가 다음 phase의 required input을 만족하는가?
3. 만족하면 forward. 불만족이면 현재 phase 계속 또는 backtrack.

이 판단은 각 스킬의 Input/Output Contract에 의해 자동으로 발생한다 — 스킬 진입 시 required input이 없으면 사용자에게 질의하거나 선행 스킬을 제안한다.

---

## §4. Input/Output Contract 표준 형식

각 스킬의 PROCEDURE.md에 아래 형식의 섹션을 포함한다.

### Input Requirements

```markdown
## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| <name> | ✅ | artifact | <Phase N 산출물 / 특정 스킬 output> | "질의문" |
| <name> | ✅ | knowledge | 사용자 도메인 지식 | "질의문" |
| <name> | 선택 | artifact | <source> | 기본 동작 설명 |
```

**Type 분류:**

| Type | 정의 | 예시 |
|------|------|------|
| `artifact` | 파일 또는 문서로 존재하는 산출물 | PRD, API contract, test report |
| `knowledge` | 사용자의 도메인 지식, 정형화되지 않은 정보 | "주요 read/write 패턴은?", "target 사용자는?" |
| `decision` | 선행 의사결정 결과 | "tech stack 결정", "deploy 전략 결정" |

**미제공 시 처리 원칙:**

1. `Required` + `artifact` type → 선행 스킬 제안 ("먼저 `/buddy:<producing-skill>` 을 실행하세요")
2. `Required` + `knowledge` type → 사용자에게 질의 ("질의문" 컬럼의 질문을 사용)
3. `Required` + `decision` type → 결정을 내릴 수 있는 스킬 제안
4. `선택` → 없어도 진행, 산출물 품질이 낮아질 수 있음을 안내

### Output Contract

```markdown
## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| <name> | artifact | <structured YAML / prose / file path> | <소비 스킬 목록> |
| <name> | decision | <ADR / inline record> | <소비 스킬 목록> |
```

**Consumers 작성 규칙:**

- 직접 소비하는 스킬만 기재 (transitive dependency 제외)
- Phase orchestrator는 소비자로 기재하지 않음 (orchestrator는 stage를 호출할 뿐, artifact를 직접 소비하지 않음)

### Standalone 등급 (자동 도출)

Input Requirements 표에서 자동 결정:

| 등급 | 조건 | 설명 |
|------|------|------|
| **full-standalone** | required input이 모두 `knowledge` type | 사용자 응답만으로 진행 가능 |
| **standalone-with-context** | required input에 `artifact` type이 있지만, 미제공 시 사용자 질의로 대체 가능 | artifact 없이도 사용자가 구두로 정보 제공하면 진행 |
| **orchestrator-preferred** | required input의 `artifact` type이 복수이고, 선행 phase 전체 산출물에 의존 | 독립 실행 가능하나 orchestrator 경유가 품질 보장 |

standalone 등급은 PROCEDURE.md에 별도 기재하지 않는다 — Input Requirements 표에서 읽는 사람이 자연히 판단할 수 있다.

---

## §5. 본 문서의 범위와 한계

### 범위

- 9-phase 정체성 정의 (artifact 기준)
- Input/Output Contract 표준 형식 정의
- Phase 전이 규칙 (forward / backtrack / skip)

### 범위 외

- 개별 스킬의 실제 Input/Output Contract 내용 → 각 PROCEDURE.md에 기재
- 스킬 간 라우팅 충돌 결정 → [`routing-rules.md`](./routing-rules.md)
- 스킬 목록 및 description → [`skill-catalog.md`](./skill-catalog.md)
- SE 이론 baseline 상세 → [`docs/plugin-skills-engineering-flow.md`](../../../../docs/plugin-skills-engineering-flow.md)

### 변경 trigger

| trigger | 갱신 부분 |
|---------|----------|
| 새 phase 추가/분리/병합 | §2 전반 + §3 전이 규칙 |
| Input/Output Contract 형식 변경 | §4 |
| 새 backtrack/skip 패턴 발견 | §3 |
| 개별 스킬의 Input/Output Contract 작성 완료 후 phase 산출물 보정 필요 시 | §2 해당 phase |
