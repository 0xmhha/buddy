# Buddy Plugin — Lifecycle Orchestrator Architecture (Draft)

> **Status**: Draft for review
> **Date**: 2026-05-04
> **Scope**: 상용 제품 빌딩 풀 라이프사이클을 위한 buddy plugin skill/command/MCP 재구조화 제안.
> **Replaces**: 2026-04-24 plugin scaffold 의 단일 orchestrator (`autoplan`) 가정.
> **Related**: [`docs/skill-map.md`](../../skill-map.md), [`plugin/SKILL_ROUTER.md`](../../../plugin/SKILL_ROUTER.md), [`plugin/SKILLS.md`](../../../plugin/SKILLS.md).

---

## 1. 배경 및 문제 정의

### 1.1 현재 상태

- buddy plugin 은 `/Users/wm-it-22-00661/Work/github/stable-net/study/docs/projects/buddy/extracts/ko/plugin/skills/` 의 58 skill 을 import 한 상태.
- 직전 SKILL_ROUTER.md 변경에서 `autoplan` 을 단일 **Priority 1 Orchestrator** 로 정의하고, `validate-idea` / `review-scope` / `critique-plan` 을 그 안의 **dual-mode stage skill** 로 위치시킴.
- 17 skill 이 `/buddy:<name>` command 로 노출됨.

### 1.2 문제

`autoplan` 을 라이프사이클 전체의 단일 orchestrator 로 가정한 것은 잘못된 모델이다.

| 항목 | 단일 orchestrator 가정의 문제 |
|------|------------------------------|
| `autoplan` 의 실제 정의 | description 은 "multi-stage 리뷰 파이프라인" — review-scope/review-engineering/review-design/review-devex 4-mode plan **review** 도구이지 라이프사이클 orchestrator 가 아님. |
| 라이프사이클 단계의 비균질성 | 각 단계는 결정 분기·산출물·호출하는 agent/MCP 가 다름. 동일 review 패턴을 모든 단계에 강제하면 단계 특수성 소실. |
| 상용 제품 누락 영역 | 단일 orchestrator 모델은 idea→PRD 흐름만 가정. launch 이후 (UAT, 인시던트 대응, A/B 실험, 코호트 분석, 고객 피드백 → 백로그, lifecycle 종료) 영역이 구조적으로 누락. |

### 1.3 목표

- **Multi-orchestrator 모델** 채택 — 라이프사이클 단계마다 **자체 orchestrator** 를 가짐.
- 각 orchestrator 가 자체 stage skill / agent / MCP 호출 흐름을 가지고, 단계 특수성을 보존.
- 상용 제품 (NOT MVP) 기준으로 누락 단계·skill·MCP 식별.
- buddy 단독으로 **아이디어 → 사업성 → 설계 → 개발 → 테스트/보안 → 배포 → A/B 분석 → 신규 작업 생성** 풀 사이클 지원.

---

## 2. 설계 원칙

| 원칙 | 정의 | 영향 |
|------|------|------|
| **Phase autonomy** | 각 phase orchestrator 는 자체 결정 분기·stage·산출물·MCP 를 가짐. 다른 phase 의 stage 호출 금지. | orchestrator 간 의존성을 산출물 (artifact) 로 한정 — 예: §1 산출물 = PRD, §2 입력 = PRD. |
| **Stage dual-mode** | Stage skill 은 ① orchestrator 안에서 호출 ② 사용자 명시 단독 호출 모두 가능. | User Sovereignty — 사용자가 stage 단독 호출 시 orchestrator 로 escalate 금지. |
| **Command = 시작 gate** | `/buddy:<name>` command 는 phase orchestrator 와 cross-cutting utility 만 노출. stage 는 dispatch 전용. | Command 카탈로그 = 라이프사이클 진입점 표. 사용자 의사결정 부담 최소화. |
| **MCP for cross-system** | Phase 간·외부 시스템 통합은 MCP 로 일원화. orchestrator 본문이 MCP 호출만 알면 됨. | 외부 SaaS (Stripe/Sentry/Datadog 등) 변경에 buddy skill 본문이 영향받지 않음. |
| **Commercial-first, not MVP** | "MVP 충분" 으로 빠질 만한 단계도 상용 기준에서는 분리. (예: §3 / §4 분리, §8 / §9 별도 phase.) | 작업량은 늘지만 의사결정 누수 차단. |
| **Review-pipeline ≠ phase orchestrator** | `autoplan` 같은 review pipeline 은 phase orchestrator 가 아니라 **cross-phase 공유 sub-orchestrator**. plan/PRD/design 산출물이 생긴 *어떤 phase 든* 호출 가능. | `autoplan` 을 §1 / §3 / §4 의 review stage 로 공유 사용. 단계마다 별도 review skill 만들 필요 없음. |
| **Use case as primitive** | feature 는 **여러 actor 의 use case 합성** 으로 정의. actor / use case / system boundary 매핑이 §2 의 입력이고, §3(infra) / §4(병렬 trace) / §6(actor 별 통합 테스트) / §8(actor 별 metric) 으로 cascade. | feature 정의 전 use case 분해가 강제 필수. 누락 시 infra·테스트·metric 정의가 모두 부정확해짐. |

---

## 2.1 autoplan 의 위치 — 별도 Q1 답변

**질문 (검토 의견)**: `concretize-idea` 가 `autoplan` 으로 처리 가능한 스킬 아닌가?

**답**: **아니다. 둘은 다른 층위.** 그러나 `concretize-idea` 안에서 `autoplan` 이 review stage 로 invoke 됨.

| 항목 | `autoplan` | `concretize-idea` (제안) |
|------|------------|--------------------------|
| 입력 | **이미 존재하는** rough plan (design doc / RFC / spec) | idea / concept (plan 미존재) |
| 작업 방향 | review-only — 4 review skill 순차 chain | generation → review → synthesis |
| 내부 호출 skill | `review-scope` / `review-engineering` / `review-design` / `review-devex` (4개 모두 review-type) | `validate-idea` (gen) / `assess-business-viability` (gen) / `analyze-competition...` (gen) / `define-product-spec` (gen-final) **+ autoplan(review)** |
| 산출물 | annotated plan + decision audit trail | PRD draft + business viability report + market position |
| 의사결정 boundary | 6 원칙 (완성도 / 범위 / pragmatic / DRY / explicit / action 편향) | 단계별 다름 (idea 검증 결정 vs business 결정 vs PRD content 결정) |
| User gate 수 | 2개 (Phase 1 전제 확인 + User Challenge) | 다수 (idea / business / scope / PRD 단계마다) |

**올바른 관계**:

```
concretize-idea (§1 phase orchestrator)
├── stage 1: validate-idea          (gen)
├── stage 2: validate-advanced-edge-idea (gen)
├── stage 3: assess-business-viability (gen)
├── stage 4: analyze-competition... (gen) ← 신규
├── stage 5: review-pricing-and-gtm (review of business)
├── stage 6: map-customer-segments  (gen) ← 신규
├── stage 7: define-product-spec    (gen-final → PRD draft)
└── stage 8: autoplan                (review of PRD draft)  ← review-pipeline 호출
        └── invokes review-scope / review-engineering / review-design / review-devex
```

→ **autoplan 을 §1 의 한 stage 로 invoke** 하면 PRD 가 4-mode review 를 자동으로 통과. autoplan 은 §3 (technical design 산출물 review) 와 §4 (implementation plan review) 에서도 동일하게 호출 가능.

**autoplan 의 위치 결정 (Q6 갱신)**:
- ❌ phase orchestrator (틀림)
- ✓ **cross-phase review sub-orchestrator** — §1 / §3 / §4 의 review stage 로 shared 사용
- 어느 phase 에서도 호출 가능하지만 각 phase 의 산출물 (PRD / ADR / task plan) 이 생긴 시점에만 의미 있음

---

## 3. 9-Phase 라이프사이클 모델

> 사용자가 제시한 6 단계 (아이디어 구체화 / feature 정의 / 기술스택+인프라+코드설계 / 개발 / 테스트+보안 / 배포) 를 분석하여 §3 분할 + §8 / §9 추가.

| Phase | Orchestrator (신규) | 사용자 6단계 매핑 | 진입 조건 | 산출물 |
|-------|---------------------|------------------|----------|--------|
| §1 Idea & Business Validation | `concretize-idea` | "아이디어 구체화" | idea/concept 만 존재 | PRD draft + business viability report |
| §2 Feature Definition & Backlog | `define-features` | "feature 정의 및 구체화" | PRD 확정 | Feature backlog (each with spec + dependencies) |
| §3 Technical Design | `design-system` | "기술 스택, 인프라, 코드 설계" 의 절반 | Feature backlog 확정 | Tech stack ADR, infra blueprint, API/data model |
| §4 Implementation Plan | `plan-build` | "기술 스택, 인프라, 코드 설계" 의 나머지 (구현 계획) | Technical design 확정 | Ordered task graph with dependencies + parallelization plan |
| §5 Development | `build-feature` | "개발" | Implementation plan 확정 | Working code per feature + 단위/통합 테스트 |
| §6 Quality (Test + Security + Compliance) | `verify-quality` | "테스트 및 보안 취약점" | Code complete | QA report + security/legal sign-off |
| §7 Release & Beta | `ship-release` | "배포" | Quality gate pass | Tagged release, canary/UAT pass, GA |
| §8 Operate & Iterate | `iterate-product` | (사용자 명시: A/B 테스트, 분석, 신규 작업 생성) | Production traffic | Experiment results, improvement backlog → §2 loop |
| §9 Lifecycle Management | `manage-lifecycle` | (상용 장기운영 추가) | Feature/product 노후화 | Deprecation notice, migration playbook, EOL |

### 3.1 §3 / §4 분리 근거

| 항목 | §3 (Technical Design) | §4 (Implementation Plan) |
|------|----------------------|-------------------------|
| 의사결정 권한자 | Architect / Tech Lead | Tech Lead / Eng Manager |
| 시간 지평 | 다년 (락인 영향) | 분기/스프린트 |
| 결정 단위 | 언어·프레임워크·DB·tenancy 모델 | task 분해·의존성·병렬화 |
| 변경 비용 | 매우 높음 (마이그레이션) | 낮음 (재계획 가능) |

→ MVP 라면 합쳐도 되지만, 상용은 분리해야 의사결정 누수 차단.

### 3.2 §8 / §9 의 commercial-only 성격

- **§8 Operate & Iterate** — A/B 실험·코호트 분석·인시던트 대응·고객 피드백 → 백로그. 사용자가 명시한 "a/b 테스트 기반 개선 및 보완점 분석과 신규 작업 생성" 영역. 현재 buddy 에 거의 비어 있음.
- **§9 Lifecycle Management** — feature deprecation, customer migration, product EOL. 1년 이상 운영하면 불가피하게 등장. 우선순위는 후순위 가능.

### 3.3 §0 (Discovery) 를 별도 phase 로 두지 않은 이유

시장조사·고객 인터뷰는 buddy 가 자동화하기 어려운 인간 활동 영역. §1 안에 보조 stage skill (`conduct-customer-interview`, `map-customer-segments` 등) 로 흡수.

### 3.4 Use Case 분해 — §2 의 핵심 원칙 (Q2 답변)

**문제 인식**: 직전 draft 는 §2 를 단순 "feature backlog 생성" 으로 정의했으나, 상용 제품에서 feature 는 단일 entity 가 아니라 **여러 actor 의 use case 합성** 이다.

**예시 — `sign-up-with-email-password`**:

| Actor | Use Case (actor 시점) | 시스템 경계 | 산출물 |
|-------|----------------------|------------|--------|
| **User** (frontend client) | "이메일/비밀번호로 회원가입한다" | Frontend SPA / mobile app | signup form, validation feedback, success redirect |
| **System** (backend auth server) | "credentials 를 검증·저장하고 session 발급" | Backend auth service + DB | password hashing, user record creation, session/JWT 발급 |
| **3rd-party** (email verifier) | "이메일 소유 검증 (deliverability + double opt-in)" | External SaaS (SendGrid / SES + verification webhook) | verification email 발송, click confirmation 처리 |

→ feature 1개 = use case 3개의 합성. 각 use case 는 **다른 actor 가 다른 시스템 경계에서 제공**. use case 분해 없이 feature 만 정의하면:

- ❌ infra 구성 시 어느 시스템이 무엇을 담당해야 하는지 불명확 (§3 Technical Design 입력 부족)
- ❌ 병렬 implementation track 분배 불가 (§4 Implementation Plan 입력 부족)
- ❌ actor 별 통합 테스트 누락 (§6 Quality 입력 부족)
- ❌ actor 별 funnel metric 측정 불가 (§8 Iterate 입력 부족)

**원칙**: §2 의 첫 작업은 actor 식별 → use case 매핑 → system boundary 매핑. feature 는 use case 합성의 *결과* 이지 입력 아님.

**§2 stage 흐름 (수정)**:

```
define-features (§2 phase orchestrator)
├── stage 1: identify-actors            ← 신규: 시스템 참여 actor 열거 (user/admin/system/3rd-party 분류)
├── stage 2: map-actor-use-cases        ← 신규: actor 별 use case 식별 (UML use case 다이어그램 등가)
├── stage 3: map-use-case-to-system-boundary ← 신규: 각 actor 의 use case 가 어느 시스템 경계에 속하는지
├── stage 4: compose-feature-from-use-cases  ← 신규: cross-actor use case 묶음 → feature
├── stage 5: define-feature-spec        ← 신규: feature_id, scope, **actor list**, **per-actor use cases**, **system boundary**, acceptance, test_plan
├── stage 6: query-feature-registry     ✓ 기존: 유사 feature 검색 (재사용 / fork / new)
├── stage 7: score-feature-priority     ← 신규: RICE/ICE/MoSCoW
├── stage 8: map-feature-dependencies   ← 신규: feature 간 선후 그래프
├── stage 9: split-work-into-features   ✓ 기존: 큰 feature 분리
└── stage 10: triage-work-items         ✓ 기존: 백로그 상태 머신
```

**Cross-phase cascade**:

| Phase | Use case 분해 결과의 영향 |
|-------|--------------------------|
| §3 Technical Design | 각 actor 의 system boundary → infra 토폴로지. frontend / backend / 3rd-party 어댑터 의 책임 경계가 use case 단위로 결정. |
| §4 Implementation Plan | actor 별 implementation track = 병렬 worker 분배 단위. frontend 팀 / backend 팀 / integration 팀 별 task list. |
| §6 Quality | actor 단위 통합 테스트 (user-actor E2E / system-actor unit / 3rd-party contract test). |
| §8 Iterate | actor 별 funnel metric (frontend conversion / backend success rate / 3rd-party deliverability). 어디서 drop-off 가 발생하는지 actor 차원으로 추적. |

**Feature spec format 갱신** (skill-map.md §4 의 포맷에 추가):

```yaml
feature_id: signup-email-password
name: Email/Password Sign-Up
summary: 사용자가 이메일/비밀번호로 회원가입
problem: ...
actors:                              # ← 신규
  - id: user
    use_cases:                       # ← 신규
      - 이메일/비밀번호 입력 + 제출
      - 검증 실패 메시지 확인
      - 회원가입 성공 후 redirect
    system_boundary: frontend-spa
  - id: auth-system
    use_cases:
      - credentials 형식 검증
      - 비밀번호 hash 저장
      - session/JWT 발급
    system_boundary: backend-auth-service
  - id: email-verifier
    use_cases:
      - verification 이메일 발송
      - click confirmation webhook 처리
    system_boundary: external-saas (SendGrid/SES)
scope: ...
out_of_scope: ...
acceptance_criteria:                 # actor 별로 분리
  user: ...
  auth-system: ...
  email-verifier: ...
test_plan:                           # actor 별 + cross-actor
  per_actor: ...
  integration: ...
implementation_notes: ...
code_links: ...
status: draft
owners: ...
updated_at: ...
```

→ 이 포맷은 §3 (infra), §4 (implementation), §6 (test), §8 (metric) 의 입력 schema 가 됨.

---

## 4. 단계별 Skill 군집화 + Gap 분석

> ✓ = buddy 보유 (`plugin/skills/`)
> ⚠ = 보유하지만 phase orchestrator 안의 stage 로 재배치 필요
> 🆕 = 누락 (신규 작성 필요)

### §1 `concretize-idea`

**Stage skills (보유, generation 패턴):**
- ✓ `validate-idea`, `validate-advanced-edge-idea` — idea 검증 인터뷰
- ✓ `assess-business-viability`, `review-pricing-and-gtm` — 사업성
- ✓ `define-product-spec` — PRD 작성 (generation-final)
- ✓ `apply-builder-ethos` (cross-cutting, ambient 적용)

**Stage skills (보유, review 패턴 — sub-orchestrator):**
- ✓ `autoplan` — PRD draft 가 생기면 호출 (4-mode review). §1 에서는 PRD 직전 review stage. §3 / §4 에서도 동일하게 호출 가능.
- ✓ `review-scope`, `critique-plan` — autoplan 안에서 호출되거나 standalone

**Gap (상용 필수):**
- 🆕 `analyze-competition-and-substitutes` — 경쟁/대체재 매트릭스
- 🆕 `map-customer-segments` — 고객 세그먼트와 구매자 분리
- 🆕 `map-jobs-to-be-done` — JTBD 프레임 인터뷰
- 🆕 `analyze-market-size` — TAM/SAM/SOM 정량화
- 🆕 `conduct-customer-interview` — 인터뷰 스크립트 + 합성

> autoplan 의 위치 상세는 §2.1 참조.

### §2 `define-features` — Use Case Mapping & Feature Definition

> 직전 draft 가 단순 "feature backlog 생성" 으로 정의했던 것을 §3.4 (Use Case 분해 원칙) 기준으로 재정의. actor / use case / system boundary 매핑이 feature 정의의 *입력*.

**Stage skills (보유, feature 합성 후 단계):**
- ✓ `split-work-into-features`, `query-feature-registry`, `triage-work-items`

**Gap (use case 분해 — 신규, §2 의 첫 단계):**
- 🆕 `identify-actors` — 시스템 참여 actor 열거 (user / admin / system / 3rd-party / external-tool 분류)
- 🆕 `map-actor-use-cases` — actor 별 use case 식별 (UML use case 다이어그램 등가)
- 🆕 `map-use-case-to-system-boundary` — 각 use case 의 시스템 경계 (frontend / backend / external SaaS / DB / queue)
- 🆕 `compose-feature-from-use-cases` — cross-actor use case 합성 → feature

**Gap (feature 정의 — 신규, use case 분해 후):**
- 🆕 `define-feature-spec` — feature_id / scope / **actor list** / **per-actor use cases** / **system boundary** / acceptance / test_plan (skill-map.md §4 포맷 + use case 확장; §3.4 참조)
- 🆕 `score-feature-priority` — RICE/ICE/MoSCoW
- 🆕 `map-feature-dependencies` — feature 간 선후 그래프
- 🆕 `estimate-feature-effort` — story point / t-shirt sizing

**Stage 흐름 (구체):** §3.4 의 stage 1~10 다이어그램 참조.

**산출물 → 다음 phase 입력**: feature spec 의 actor / use case / system boundary 가 §3 (infra), §4 (implementation), §6 (test), §8 (metric) 의 schema 가 됨.

### §3 `design-system`

> 입력: §2 의 feature spec (actor list + per-actor use cases + system boundary 포함). §3 의 첫 작업은 use case → 실제 infra component 매핑.

**Stage skills (보유):**
- ✓ `review-architecture`, `review-engineering`, `review-design`, `review-devex`
- ✓ `design-artifact-storage`, `design-billing-system`, `design-claude-hooks`, `design-deploy-strategy`, `design-embedding-search`, `design-mcp-server`
- ✓ `consult-codex`, `consult-design-system`, `explore-design-variants`
- ✓ `autoplan` (review sub-orchestrator) — design 산출물(ADR / API spec / data model) draft 후 호출하여 4-mode review

**Gap (use case → infra 브릿지 — 신규, §3 의 첫 단계):**
- 🆕 `map-use-cases-to-infra` — §2 actor system boundary → 실제 infra component (예: frontend SPA = Next.js + CDN, auth-system = Go service + Postgres, email-verifier = SendGrid SDK)
- 🆕 `derive-system-topology` — actor 그래프 + use case 흐름 → 시스템 토폴로지 다이어그램 (sync/async, request/response 관계)

**Gap (상용 큰 누락):**
- 🆕 `define-tech-stack` — 언어/프레임워크/DB 선택 (락인 영향 평가)
- 🆕 `design-data-model` — 스키마/마이그레이션/인덱싱
- 🆕 `design-api-contract` — REST/GraphQL/RPC 계약 (actor 간 경계 = API 경계)
- 🆕 `design-event-schema` — Kafka/queue/webhook 계약
- 🆕 `design-auth-model` — RBAC/ABAC, 멀티테넌트 격리
- 🆕 `design-observability` — 로깅/메트릭/트레이싱 표준
- 🆕 `design-secret-management` — KMS/Vault 통합
- 🆕 `design-tenant-model` — multi-tenancy 전략 (상용 SaaS 핵심)
- 🆕 `design-i18n-strategy` — 다국어/로케일
- 🆕 `design-accessibility-baseline` — WCAG 2.1 AA 기준선
- 🆕 `write-adr` — Architecture Decision Record

### §4 `plan-build`

> 입력: §2 feature spec + §3 system topology. §4 의 핵심: actor 별 implementation track 분리 (병렬 worker 분배 단위).

**Stage skills (보유):** 거의 없음 (큰 gap), `autoplan` 호출 가능 (task plan review)

**Gap:**
- 🆕 `decompose-feature-to-actor-tracks` — feature → actor 별 task track (frontend / backend / 3rd-party 통합 별 독립 track)
- 🆕 `decompose-track-to-tasks` — actor track → ordered task list
- 🆕 `map-task-dependencies` — task DAG (actor 내부 의존성 + actor 간 contract 의존성)
- 🆕 `plan-parallel-execution` — actor track 별 병렬 worker 분배 (use case 분해 결과로 자연스럽게 도출)
- 🆕 `define-acceptance-test-plan` — actor 별 + cross-actor 완료 기준
- 🆕 `estimate-build-timeline` — 의존성 + 병렬도 → 일정 합성

### §5 `build-feature`

**Stage skills (보유):**
- ✓ `build-with-tdd`, `iterate-fix-verify`, `freeze-edit-scope`
- ✓ `dispatch-parallel-agents`, `diagnose-bug`
- ✓ `consult-codex` (cross-cutting)

**Gap:**
- 🆕 `generate-from-api-contract` — codegen (OpenAPI → server stub)
- 🆕 `generate-tests-from-spec` — feature spec → test 스캐폴딩
- 🆕 `pair-program-loop` — 실시간 co-coding 루프
- 🆕 `refactor-with-rename-trace` — 이름 변경 안전 전파
- 🆕 `update-docs-with-code` — doc-as-code 동기화

### §6 `verify-quality`

> 입력: §2 feature spec (per-actor use cases + system boundary). actor 별 통합 테스트 + cross-actor E2E 가 분리 설계됨.

**Stage skills (보유):**
- ✓ `classify-qa-tiers`, `run-browser-qa`, `monitor-regressions`
- ✓ `audit-security`, `audit-live-devex`, `measure-code-health`, `classify-review-risks`
- ✓ `review-ai-safety-liability`, `review-privacy-data-risk`, `review-license-and-ip-risk`, `review-terms-policy-readiness`

**Gap (use case 기반 — 신규):**
- 🆕 `test-per-actor-use-case` — actor 시점에서 use case 통합 테스트 (frontend E2E / backend unit·integration / 3rd-party contract)
- 🆕 `test-cross-actor-flow` — actor 간 협업 흐름 테스트 (e.g., signup → email verification → first login full flow)

**Gap (상용 발표 전 필수):**
- 🆕 `run-load-test` — 성능/스케일 베이스라인
- 🆕 `audit-accessibility` — a11y 자동 점검 (use case 의 user-actor 단계에 한정)
- 🆕 `audit-i18n-coverage` — 번역 커버리지/locale 깨짐
- 🆕 `audit-cost-efficiency` — LLM/인프라 단가 분석
- 🆕 `chaos-test` — 장애 주입 테스트 (3rd-party actor 실패 시나리오 포함)
- 🆕 `audit-test-coverage-meaningful` — mutation testing

### §7 `ship-release`

**Stage skills (보유):**
- ✓ `setup-quality-gates`, `auto-create-pr`, `automate-release-tagging`, `sync-release-docs`, `write-changelog`
- ✓ `guard-destructive-commands`, `compose-safety-mode`

**Gap (상용 배포 안전망):**
- 🆕 `setup-canary-deploy` — canary/blue-green 자동화
- 🆕 `setup-feature-flags` — 점진 배포 인프라
- 🆕 `setup-rollback-runbook`
- 🆕 `run-uat` — User Acceptance Testing 오케스트레이션
- 🆕 `run-beta-program` — 클로즈드 베타 + 피드백 수집
- 🆕 `prepare-launch-checklist` — go-live readiness gate
- 🆕 `setup-incident-paging` — PagerDuty/Opsgenie 연동

### §8 `iterate-product` (사용자 명시 영역, 큰 gap)

> 입력: production traffic + §2 feature spec (per-actor use cases). funnel / 전환 / 실패율 metric 을 actor 별로 측정 — 어느 actor 단계에서 drop-off 가 발생하는지 추적 가능.

**Stage skills (보유):**
- ✓ `monitor-regressions`, `save-context`, `restore-context`, `summarize-retro`, `persist-learning-jsonl`

**Gap (사용자 요구의 핵심):**
- 🆕 `design-ab-experiment` — 가설/표본 크기/대조군
- 🆕 `analyze-ab-experiment` — 통계 유의성 → 결정
- 🆕 `analyze-user-funnel` — actor 단계별 전환/이탈 분석 (e.g., signup feature: form-submit → backend-success → email-verified 의 단계별 conversion)
- 🆕 `analyze-feature-adoption` — 기능 도입률 → 개선 신호
- 🆕 `analyze-user-cohort` — 리텐션/세그먼트
- 🆕 `analyze-actor-failure-rate` — actor 별 실패율 분리 (frontend validation 실패 / backend 5xx / 3rd-party timeout) — 개선 우선순위 결정
- 🆕 `generate-improvement-tasks` — 분석 → 백로그 자동 생성 (어느 actor 의 어느 use case 가 약한가 → §2 로 재진입)
- 🆕 `handle-incident` — 인시던트 대응 런북
- 🆕 `conduct-postmortem` — 비난 없는 포스트모템
- 🆕 `analyze-cost-anomaly` — 비용 스파이크 탐지
- 🆕 `triage-customer-support-ticket` — 고객 피드백 → 트리아지
- 🆕 `analyze-customer-feedback-corpus` — NLP 로 review/티켓 합성
- 🆕 `audit-error-budget` — SLO/SLI 추적

### §9 `manage-lifecycle` (상용 장기운영)

**Stage skills (보유):** 없음

**Gap:**
- 🆕 `deprecate-feature` — 점진적 sunset
- 🆕 `migrate-customers` — 강제 마이그레이션 플레이북
- 🆕 `archive-product` — EOL 흐름
- 🆕 `spin-off-feature` — 별도 제품으로 분리

### Cross-cutting (phase 소속 없음)

- ✓ `apply-builder-ethos` — 모든 phase 에서 ambient 적용
- ✓ `detect-install-type`, `guide-setup-wizard`, `benchmark-llm-models` — 패턴 라이브러리, 다른 skill 내부에서 호출
- ⚠ archive 3개 (`route-intent`, `route-multi-platform`, `route-spec-to-code`) — orchestrator 모델 채택 시 reference 가치 재평가 필요

---

## 5. MCP 요구사항

상용 제품 운영에 buddy 단독으로는 부족. orchestrator 들이 외부 시스템과 통합하려면:

| MCP | 사용 phase | 역할 | 우선순위 근거 |
|-----|-----------|------|---------------|
| `feature-management-mcp` | §2, §3, §8 | feature.query / store / update / link_code / export_patch (skill-map.md §4 의 feature MCP) | 3개 phase 에서 사용, buddy 단독 가치 압도적 → **1순위** |
| `analytics-mcp` | §8 | A/B 실험 데이터, 사용자 funnel, 코호트 | §8 핵심 기능 → 2순위 |
| `monitoring-mcp` | §7, §8 | Datadog/Sentry/CloudWatch | 외부 SaaS 의존, 통합만 |
| `support-mcp` | §8 | Zendesk/Intercom 티켓 ingestion | 외부 SaaS 의존 |
| `cost-mcp` | §6, §8 | 클라우드/LLM 단가 | 회계 시스템 의존 |
| `billing-mcp` | §7, §8 | Stripe/Toss subscription | 외부 SaaS 의존 |
| `feature-flag-mcp` | §7 | LaunchDarkly/Unleash | 외부 SaaS 의존 |

→ **buddy in-house 개발 권장**: `feature-management-mcp` (1순위), `analytics-mcp` (2순위).
→ **외부 SaaS 의존 — 통합 어댑터만**: 나머지 5개.

---

## 6. Command Gate 재정의

### 6.1 Phase orchestrator commands (9, 신규 — 시작 gate)

| Command | Phase |
|---------|-------|
| `/buddy:concretize-idea` | §1 |
| `/buddy:define-features` | §2 |
| `/buddy:design-system` | §3 |
| `/buddy:plan-build` | §4 |
| `/buddy:build-feature` | §5 |
| `/buddy:verify-quality` | §6 |
| `/buddy:ship-release` | §7 |
| `/buddy:iterate-product` | §8 |
| `/buddy:manage-lifecycle` | §9 |

### 6.2 Cross-cutting utility commands (3, 유지)

| Command | 사유 |
|---------|------|
| `/buddy:save-context` | 어느 phase 에서든 호출 |
| `/buddy:restore-context` | 동일 |
| `/buddy:consult-codex` | 외부 second opinion (도구 성격, phase 종속 X) |

### 6.3 제거 권장 (현재 17개 중 14개)

→ phase orchestrator 안의 stage 로 재배치, dispatch 만 유지:

| 제거 command | 흡수 phase |
|--------------|-----------|
| `/buddy:validate-idea`, `/buddy:validate-advanced-edge-idea`, `/buddy:assess-business-viability`, `/buddy:define-product-spec` | §1 |
| `/buddy:autoplan`, `/buddy:explore-design-variants` | §3 (plan review stage) |
| `/buddy:dispatch-parallel-agents`, `/buddy:build-with-tdd`, `/buddy:diagnose-bug` | §5 |
| `/buddy:audit-security`, `/buddy:measure-code-health` | §6 |
| `/buddy:auto-create-pr`, `/buddy:setup-quality-gates` | §7 |
| `/buddy:summarize-retro` | §8 |

→ **최종 12 commands** (9 phase + 3 utility).

### 6.4 SKILL_ROUTER §2 도메인 우선순위 표 변경

| Priority | Category | 대표 skill | Rationale |
|----------|----------|-----------|-----------|
| 1 | **Phase orchestrator** | `concretize-idea`, `define-features`, `design-system`, `plan-build`, `build-feature`, `verify-quality`, `ship-release`, `iterate-product`, `manage-lifecycle` | 라이프사이클 시작 gate. Tie-breaker: 진입 조건 매칭 (idea만 있으면 §1, code 존재하면 §3+, production traffic 있으면 §8). |
| 2 | **Stage skill (dual-mode)** | 각 phase 안의 stage skill | 사용자 명시 호출 시 standalone, orchestrator 안에서는 단계로 호출. |
| 3 | **Cross-cutting utility** | `save-context`, `restore-context`, `consult-codex`, `apply-builder-ethos` | phase 종속 없음. 어디서든 호출 가능. |
| 4 | **Pattern library** | `[패턴 라이브러리]` 태그 | 다른 skill 내부에서 ambient 적용. |
| 5 | **Archive** | `route-intent`, `route-multi-platform`, `route-spec-to-code` | dispatch / command 모두 금지. |

---

## 7. 작업량 견적

| 항목 | 신규 작성 수량 | 비고 |
|------|--------------|------|
| Phase orchestrator skill | 9 | 각각 stage 호출 흐름 + 결정 분기 + 산출물 정의 |
| Phase 별 누락 stage skill | 약 60+ | §3 (11개), §4 (5개), §6 (6개), §7 (7개), §8 (12개), §9 (4개) 등 |
| 신규 in-house MCP | 2 | `feature-management-mcp`, `analytics-mcp` |
| 외부 SaaS MCP 어댑터 | 5 | monitoring/support/cost/billing/feature-flag |
| Command 재정의 | 12 (9 신규 + 3 유지) | 14 개 제거 |
| 문서 업데이트 | SKILLS.md 카탈로그 재구성 / SKILL_ROUTER.md §2 변경 / skill-map.md 9-phase 모델 반영 |

→ **현재의 약 2배 분량의 신규 skill 작성 + MCP 7개 설계**.

---

## 8. 결정 포인트 (검토 후 확정 필요)

### Q1. 9-phase 모델 범위
- (a) **9-phase 그대로** — 전부 채택
- (b) **§9 lifecycle 후순위** — 우선 §1~§8 만, §9 는 1년 후
- (c) **§3 / §4 합치기** — 상용도 design + plan 통합 가능하면
"A": Q1=(a)

### Q2. Command 재정의 범위
- (a) **12 commands** (9 phase + 3 utility) — 14 개 제거, 권장
- (b) **17 + 9 = 26 commands** — phase orchestrator 추가, 기존 stage command 도 유지 (dual full)
- (c) **9 commands** only — utility 도 제거, orchestrator 만
"A": Q2=(b)

### Q3. 신규 skill 작성 우선순위
- (a) **모든 phase 동시 진행** — 작업량 큼
- (b) **§1~§5 먼저** — idea → 개발까지 우선 완성, §6~§9 후속
- (c) **§8 (iterate) 먼저** — buddy 최대 약점 보강 우선
- (d) **현재 보유 skill 의 phase 재배치만** 먼저, 신규 skill 은 후속
"A": Q3=(c)->(b) 단계.

### Q4. MCP 우선순위
- (a) **`feature-management-mcp` 1순위** — §2/§3/§8 에 모두 쓰임, in-house 가치 압도적 (권장)
- (b) **`analytics-mcp` 1순위** — §8 핵심
- (c) **MCP 작성 보류** — skill 만 먼저 정리, MCP 는 별도 트랙
"A": Q4=(c)

### Q5. archive 3개 처리
- (a) **그대로 유지** — reference 가치
- (b) **`plugin/_archive/` 디렉토리로 격리** — 메인 카탈로그에서 제외
- (c) **삭제** — orchestrator 모델 확정 후 reference 가치 소멸
"A": Q5=(b)

### Q6. autoplan 의 위치 (§2.1 갱신 반영)
- (a) **Cross-phase review sub-orchestrator** (권장) — §1 PRD draft / §3 ADR · API spec / §4 task plan / §6 release plan 의 어느 산출물에든 호출 가능. 단일 phase 종속 X.
- (b) §3 단일 phase 의 stage 로만 한정 — 다른 phase 는 별도 review skill 필요 (중복)
- (c) `concretize-idea` 와 합쳐 §1 단일 orchestrator 로 (생성 + 리뷰 한 skill, scope 비대화 위험)
"A": Q6=(a)

### Q7. 현재 사용자가 명시하지 않은 phase 추가 여부
- (a) **§0 Discovery 추가** — 시장조사/고객 인터뷰 별도 phase
- (b) **§7.5 Beta/UAT 분리** — §7 안에서 분리
- (c) **추가 없음** — 9-phase 로 충분
"A": Q7=(b)

### Q8. Use case 분해 (§3.4) 채택 범위
- (a) **§2 첫 단계로 강제** + actor / use case / system boundary 가 feature spec 의 필수 필드 (권장) — §3.4 의 cascade 가 §3 / §4 / §6 / §8 에 모두 반영
- (b) **§2 의 권장 단계** — 강제하지 않고 복잡한 feature 에서만 사용
- (c) **별도 phase §2.5 Use Case Mapping** 분리 — 9-phase → 10-phase 로 확장
- (d) **현재 단순 feature backlog 유지** — use case 분해는 미래 확장으로 연기
"A": Q8=(a)

---

## 9. 참조

- 11-stage 상용 제품 빌딩 flow (이 문서의 9-phase 의 source) → [`docs/skill-map.md`](../../skill-map.md)
- Feature Management SaaS / MCP 상세 (skill-map §4) → [`extracts/ko/feature-management-saas-mcp.md`](https://github.com/0xmhha/study/blob/main/docs/projects/buddy/extracts/ko/feature-management-saas-mcp.md) (외부 레포)
- 현재 plugin scaffold spec → [`docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`](./2026-04-24-buddy-plugin-architecture-design.md)
- 현재 skill router → [`plugin/SKILL_ROUTER.md`](../../../plugin/SKILL_ROUTER.md)
- 현재 skill catalog → [`plugin/SKILLS.md`](../../../plugin/SKILLS.md)

---

## 10. 다음 단계

1. **이 문서 검토** — Q1~Q7 결정.
2. **결정 반영 PR**:
   - SKILL_ROUTER.md §2 / §3 multi-orchestrator 모델로 갱신
   - SKILLS.md 카탈로그 재구성
   - 14 command 제거 + 9 phase command + 3 utility 추가
   - skill-map.md 의 11-stage 를 9-phase 로 정렬
3. **신규 skill 작성** (Q3 우선순위 따라 순차).
4. **MCP 설계** (Q4 우선순위 따라).
5. **dogfood**: 실제 작은 프로젝트 하나에 buddy 사용해 phase orchestrator 동작 검증.

## 명령
- 결정 반영 PR은 Q1~Q8의 응답을 반영하여 작업을 갱신할 것.
