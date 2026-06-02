# Plugin Skills — Engineering Best Practice Flow Coverage — 2026-05-25

> **목적**: 소프트웨어 공학의 *이론적 베스트 프랙티스 flow* 를 baseline 으로 잡고, buddy 의 광의 engineering 관련 스킬 (약 106개) 를 1:1 매핑하여 **누락 hole** + **책임 충돌** + **cascade 단절** 을 식별. *engineering 작업물에 직접 영향* 을 주는 모든 영역의 coverage audit SSoT.
>
> **상위 문서**: [`plugin-skills-inventory.md`](./plugin-skills-inventory.md) (152 skill 전체 inventory).
>
> **자매 문서**: [`plugin-skills-engineering-audit.md`](./plugin-skills-engineering-audit.md) ([`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) A 카테고리 *좁은 13 영역* 결정 기록 — 본 문서와 직교).
>
> **선행 SSoT**: [`plugin/skills/router/references/skill-catalog.md`](../plugin/skills/router/references/skill-catalog.md) (9-phase × Priority 카탈로그).

---

## §0. 한 줄 요약

이론 10 phase × 약 76 fine-grained step 중 **buddy cover ~63 step (83%)** + **부분 cover 8 step (10%)** + **명확한 hole 5 step (7%)**. 책임 충돌 검토 결과 **명확한 중복 2건** (`monitor-regressions` 양쪽 등재 / `apply-builder-ethos` 양쪽 등재 — inventory G2/G3) + **잠재 책임 모호 4건**. Intra-phase cascade hint 가 PROCEDURE.md 본문에 *명시된 비율 추정 ~30%* — 나머지 ~70% 는 router 의 description-based dispatch 의존.

> **상태 갱신 (2026-05-26)**:
> - **H1 + H3 hole ✅ closed** — commit `cd99818` (`audit-ubiquitous-language` 신규 skill). Phase 2.4 ubiquitous language + Cross-cutting C.10 동시 cover.
> - **C1 + C2 중복 ✅ closed** — 본 cycle commit (`monitor-regressions` §6 라인 제거, `apply-builder-ethos` §1 라인 제거). catalog unique 153 → 152.
> - **현 cover 갱신**: hole 5 → 3 (H1/H3 closed, H2 잔여 cli 트랙, H4/H5 잠재 유지). Phase 2 cover 7/10 → 8/10. Cross-cutting cover 7/10 → 8/10.
> - **Tier 1 (engineering-flow §7) ✅ 모두 closed** (H1 + C1 + C2). 다음 진입점 = Tier 2 (P2/P3 DDD 차원 보강 / M1-M4 책임 매트릭스 / cascade Next 표준화).

---

## §1. Methodology

### §1.1 이론 baseline source (인용 분류 — ADR-003 verbatim 0건 유지)

| Source | 기여 영역 |
|--------|---------|
| ISO/IEC 12207 (Software Life Cycle Processes) | 10 phase 골격 (Discovery → Lifecycle) |
| IEEE Std 1471 (Architecture Description) | Architecture phase decomposition |
| PMBOK (PMI 7판) | Planning phase + risk |
| Clean Code (Martin) | Implementation phase 의 naming / function / refactoring |
| Pragmatic Programmer (Hunt, Thomas) | Cross-cutting (DRY / orthogonality / tracer bullet) |
| Continuous Delivery (Humble, Farley) | Release phase + quality gate |
| Domain-Driven Design (Evans) | Analysis phase 의 ubiquitous language / bounded context |
| Refactoring (Fowler) | Implementation phase 의 refactoring catalog |
| Test-Driven Development (Beck) | Implementation/Verification 의 TDD cycle |
| Site Reliability Engineering (Google) | Operations phase + SLO/error budget |
| Accelerate (Forsgren, Humble, Kim) | Operations 의 4 DORA metric + continuous improvement |
| Building Microservices (Newman) | Architecture 의 decomposition + boundary |
| Working Effectively with Legacy Code (Feathers) | Lifecycle phase 의 deprecation / refactor |

본 audit 은 *통합 frame* — 단일 source 권위 추종 아님.

### §1.2 매핑 단위

- **이론 step** = 베스트 프랙티스의 *최소 actionable activity* (예: "API contract design" — 한 commit 또는 한 PR 의 작업 단위)
- **buddy skill** = `plugin/skills/<name>/PROCEDURE.md` 한 개
- **매핑 분류**:
  - `✅ cover` — buddy skill 1+ 개가 *step 의 90%+ 책임* 수행
  - `🟡 부분 cover` — *50-89% 책임* 수행, *명시되지 않은 책임* 잔여
  - `❌ hole` — *50% 미만* 또는 *직접 매핑 skill 없음*
  - `⚠️ 책임 모호` — 매핑 skill 있으나 *다른 skill 과 책임 경계 불명*

### §1.3 Cascade 정의

- **자연 cascade** — phase orchestrator 의 9-phase chain (§1 concretize-idea → §2 define-features → ... → §9 manage-lifecycle)
- **명시 cascade** — PROCEDURE.md 본문에 *next skill* 또는 *prerequisite* 명시
- **암묵 cascade** — description-based dispatch 로 자율 호출 (router 가 판단)
- **단절** — step k 완료 후 step k+1 진입 *명시 hint 부재* + *router 가 dispatch 못 함* (현 사용자 발화로만 진입)

### §1.4 범위

광의 engineering = `plugin/skills/router/references/skill-catalog.md` 의:
- §2 (일부 11) + §3 (33) + §4 (7) + §5 (10) + §6 (19) + §7 (14) + §8 (engineering-side ~5) + Cross-cutting Utilities (5) + Cross-Phase Sub-Orchestrator (2) = **~106 skill**

제외: §1 (idea/business validation, *engineering 외*) + §8 의 *product analytics* 부분 (analyze-ab / funnel / cohort / cost / feedback / adoption — 사업 측면) + §9 (lifecycle — 사업 결정 동반).

단 §9 의 *engineering 측 작업* (deprecate-feature 의 *code 측 cleanup*) 은 cross-cutting 으로 포함 검토.

---

## §2. 이론적 SE Best Practice Flow

### §2.1 Phase 1 — Discovery & Requirements

| Step | 활동 |
|------|------|
| 1.1 | Problem identification (문제 정의, *what is the actual job*) |
| 1.2 | Stakeholder analysis (이해관계자 / 의사결정자 식별) |
| 1.3 | User research (실 사용자 인터뷰 + 행동 관찰) |
| 1.4 | Functional requirements gathering (기능 요구사항 수집) |
| 1.5 | Non-functional requirements (성능 / 보안 / 확장성 / a11y / i18n) |
| 1.6 | Constraints identification (기술 / 비즈니스 / 규제 / 시간 제약) |
| 1.7 | Acceptance criteria definition (성공 기준 정량 / Done 정의) |
| 1.8 | PRD / spec 작성 (요구사항의 문서화) |

### §2.2 Phase 2 — Analysis & Modeling

| Step | 활동 |
|------|------|
| 2.1 | Actor identification (시스템 참여자 분류) |
| 2.2 | Use case identification (actor 별 활동) |
| 2.3 | Domain modeling (entity / VO / aggregate) |
| 2.4 | Ubiquitous language definition (DDD 도메인 어휘 일관성) |
| 2.5 | Bounded context delineation (모듈 경계) |
| 2.6 | System boundary mapping (frontend / backend / external) |
| 2.7 | Use case → feature composition |
| 2.8 | Feature dependency mapping |
| 2.9 | Feature prioritization (RICE / ICE / MoSCoW) |
| 2.10 | Effort estimation (T-shirt / story point) |

### §2.3 Phase 3 — Architecture & Design

| Step | 활동 |
|------|------|
| 3.1 | Architecture style selection (microservices / monolith / serverless / event-driven) |
| 3.2 | Tech stack selection (language / framework / DB / hosting) |
| 3.3 | System topology derivation (service / data flow / trust boundary) |
| 3.4 | API contract design (REST / GraphQL / gRPC / event) |
| 3.5 | Data model design (schema / index / migration) |
| 3.6 | Event schema design (async / pub-sub / DLQ) |
| 3.7 | Auth / authz model (OAuth2 / RBAC / multi-tenant) |
| 3.8 | Multi-tenant model (RLS / schema-per / DB-per) |
| 3.9 | Observability strategy (logs / metrics / traces / SLO) |
| 3.10 | Secret management (rotation / audit / leak detection) |
| 3.11 | i18n strategy (locale / fallback / RTL) |
| 3.12 | Accessibility baseline (WCAG / a11y annotation) |
| 3.13 | Form factor decision (app / web / hybrid / desktop) |
| 3.14 | Design system application (token / component / pattern) |
| 3.15 | Interaction pattern design (gesture / motion / feedback) |
| 3.16 | Deploy strategy (canary / blue-green / rolling) |
| 3.17 | Artifact storage design (immutable / versioned) |
| 3.18 | Embedding / search design (BM25 + vector + rerank) |
| 3.19 | MCP server design (Model Context Protocol) |
| 3.20 | Architecture Decision Record (ADR 작성) |
| 3.21 | Anti-bias gate (다관점 검토 / 첫 답 commit 차단) |
| 3.22 | Architecture review (구조 무결성 검토) |
| 3.23 | Engineering plan review (data flow / edge case / coverage) |
| 3.24 | Design review (visual / UX dimension) |
| 3.25 | DX review (developer-facing product) |
| 3.26 | Scope review (creator persona, early stage) |
| 3.27 | External LLM consultation (codex review / challenge) |

### §2.4 Phase 4 — Implementation Planning

| Step | 활동 |
|------|------|
| 4.1 | Feature → actor track decomposition |
| 4.2 | Actor track → atomic task decomposition |
| 4.3 | Task dependency graph (DAG / cycle 감지) |
| 4.4 | Parallel execution plan (worker batch / sync point) |
| 4.5 | Acceptance test plan (per-actor + cross-actor) |
| 4.6 | Build timeline (CI / risk buffer / holiday) |
| 4.7 | External tracker publication (GitHub / Linear / Jira issues) |
| 4.8 | Strategic critique (CEO/founder persona) |
| 4.9 | Auto-review pipeline (scope/eng/design/devex 4-mode) |

### §2.5 Phase 5 — Implementation

| Step | 활동 |
|------|------|
| 5.1 | TDD cycle (red → green → refactor) |
| 5.2 | Pair programming (driver / navigator swap) |
| 5.3 | Parallel agent dispatch (worktree 격리 + aggregate) |
| 5.4 | Iterative fix → atomic commit → re-verify |
| 5.5 | Edit scope freeze (single directory lock) |
| 5.6 | Bug diagnosis (4-phase systematic, 10-method feedback loop) |
| 5.7 | Stuck state decomposition (3회 시도 후 자동 trigger) |
| 5.8 | Refactor with rename trace (LSP + grep + test baseline) |
| 5.9 | API contract → SDK / stub generation |
| 5.10 | Acceptance spec → test skeleton generation |
| 5.11 | Code change → docs sync (5 영역 매트릭스) |

### §2.6 Phase 6 — Verification

| Step | 활동 |
|------|------|
| 6.1 | QA tier classification (Quick / Standard / Exhaustive) |
| 6.2 | Per-actor use case test (unit / integration / contract) |
| 6.3 | Cross-actor flow test (E2E multi-actor chain) |
| 6.4 | Load test (sustained / soak / spike / stress) |
| 6.5 | Chaos test (failure injection + hypothesis-driven) |
| 6.6 | Browser QA (snapshot diff / form / responsive) |
| 6.7 | Test coverage analysis (line + mutation + behavior) |
| 6.8 | Code health measurement (composite 0-10 dashboard) |
| 6.9 | Security audit (CSO-mode + OWASP) |
| 6.10 | Accessibility audit (WCAG 2.1/2.2 + axe + manual) |
| 6.11 | i18n coverage audit (locale fallback rate) |
| 6.12 | Cost efficiency audit (per-component breakdown) |
| 6.13 | Live devex audit (TTHW + literal doc-following) |
| 6.14 | Regression monitoring (delta-based threshold) |
| 6.15 | Review risk classification (11 category) |
| 6.16 | AI safety / liability review (hallucination / autonomy) |
| 6.17 | Privacy data review (GDPR / PIPA / PIPL / HIPAA) |
| 6.18 | License / IP risk review (dep + asset + AI-gen code) |
| 6.19 | Terms / policy readiness (ToS / Privacy / Refund / DPA) |
| 6.20 | **Verification before completion** (Iron Law — no fresh evidence = no claim) |

### §2.7 Phase 7 — Release

| Step | 활동 |
|------|------|
| 7.1 | Quality gate setup (pre-commit / pre-push hook) |
| 7.2 | Feature flag system (kill switch / targeting / lifecycle) |
| 7.3 | Canary deploy setup (staged % + metric gate + auto-promote/rollback) |
| 7.4 | Rollback runbook (decision tree + verification) |
| 7.5 | Release tagging (semver auto-decision) |
| 7.6 | Changelog generation (release-summary voice rules) |
| 7.7 | UAT execution (stakeholder sign-off + evidence) |
| 7.8 | Beta program (closed cohort 5-20 + structured 피드백) |
| 7.9 | Launch checklist (17+ cross-functional gates) |
| 7.10 | Incident paging (on-call rotation + escalation + runbook index) |
| 7.11 | Release docs sync (changelog / README / ADR) |
| 7.12 | Auto PR creation (commit → branch push → PR) |
| 7.13 | Branch finalization (5-stage PR-ready sub-orchestrator) |
| 7.14 | Destructive command guard (rm -rf / DROP / force push) |
| 7.15 | Multi-safety mode composition |

### §2.8 Phase 8 — Operations (engineering 측면)

| Step | 활동 |
|------|------|
| 8.1 | Incident response (severity 분류 → 완화 → fix → 통신) |
| 8.2 | Postmortem (blameless, 5 Whys, action items) |
| 8.3 | Error budget audit (SLO burn rate multi-window) |
| 8.4 | Cost anomaly analysis (spike detection + 5-step recovery) |
| 8.5 | Actor failure rate analysis (trust score + 6 recovery patterns) |
| 8.6 | Continuous monitoring (regression / SLO / cost) |

### §2.9 Cross-cutting (any phase)

| Step | 활동 |
|------|------|
| C.1 | Anti-bias gate at decisions (engineering only, 3+ orthogonal) |
| C.2 | Anti-rationalization gate at reviews (5 rules + self-attestation) |
| C.3 | Iron Law at completion claims (fresh evidence required) |
| C.4 | Git safety (force / rewrite-pushed / auto-recovery 금지) |
| C.5 | Context save / restore (cross-branch checkpoint) |
| C.6 | Pattern documentation (skill 신규 작성 5-round verify) |
| C.7 | ADR 작성 (decision archive) |
| C.8 | Context handoff (cross-session / cross-machine) |
| C.9 | Context compaction (long-running session 토큰 압축) |
| C.10 | Ubiquitous language audit (코드 식별자 ↔ PRD 어휘 drift) |

---

## §3. 1:1 매핑 매트릭스 — buddy 광의 engineering skill

### §3.1 Phase 1 — Discovery & Requirements

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 1.1 problem identification | `concretize-idea` (§1 orchestrator) | ✅ |
| 1.2 stakeholder analysis | `identify-actors` + `map-customer-segments` (§1) | ✅ |
| 1.3 user research | `conduct-customer-interview` (§1) | ✅ |
| 1.4 functional requirements | `define-features` (§2 orchestrator) + `define-feature-spec` | ✅ |
| 1.5 non-functional requirements | `design-observability` + `design-secret-management` + `design-i18n-strategy` + `design-accessibility-baseline` (§3 cascade bridges) | ✅ |
| 1.6 constraints | `assess-business-viability` (§1) + `decide-form-factor-app-vs-web` (§3) | 🟡 부분 — *기술 제약* 만 cover, *규제 / 시간* 은 분산 |
| 1.7 acceptance criteria | `define-acceptance-test-plan` (§4) | ✅ |
| 1.8 PRD 작성 | `define-product-spec` (§1) | ✅ |

**Phase 1 cover**: 7/8 (88%) + 부분 1 (12%) + hole 0

### §3.2 Phase 2 — Analysis & Modeling

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 2.1 actor identification | `identify-actors` | ✅ |
| 2.2 use case identification | `map-actor-use-cases` | ✅ |
| 2.3 domain modeling | `design-data-model` (§3) | 🟡 부분 — *entity* cover, *VO/aggregate/bounded context* 명시 없음 |
| **2.4 ubiquitous language** | — | ❌ **hole** (engineering-audit §2.4 Tier 1 후보) |
| 2.5 bounded context | `derive-system-topology` + `map-use-cases-to-infra` | 🟡 부분 — *system boundary* cover, *DDD bounded context* 어휘 부재 |
| 2.6 system boundary | `map-use-case-to-system-boundary` + `derive-system-topology` | ✅ |
| 2.7 use case → feature | `compose-feature-from-use-cases` | ✅ |
| 2.8 feature dependency | `map-feature-dependencies` | ✅ |
| 2.9 feature prioritization | `score-feature-priority` | ✅ |
| 2.10 effort estimation | `estimate-feature-effort` | ✅ |

**Phase 2 cover**: 7/10 (70%) + 부분 2 (20%) + hole 1 (10%)

### §3.3 Phase 3 — Architecture & Design

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 3.1 architecture style | `design-system` (§3 orchestrator) | 🟡 부분 — orchestrator 가 *진입* 만, *style 선택 결정* 명시 skill 없음 |
| 3.2 tech stack | `define-tech-stack` | ✅ |
| 3.3 system topology | `derive-system-topology` | ✅ |
| 3.4 API contract | `design-api-contract` | ✅ |
| 3.5 data model | `design-data-model` | ✅ |
| 3.6 event schema | `design-event-schema` | ✅ |
| 3.7 auth/authz | `design-auth-model` | ✅ |
| 3.8 multi-tenant | `design-tenant-model` | ✅ |
| 3.9 observability | `design-observability` | ✅ |
| 3.10 secret management | `design-secret-management` | ✅ |
| 3.11 i18n strategy | `design-i18n-strategy` | ✅ |
| 3.12 a11y baseline | `design-accessibility-baseline` | ✅ |
| 3.13 form factor | `decide-form-factor-app-vs-web` | ✅ |
| 3.14 design system | `apply-design-system` + `consult-design-system` | ✅ |
| 3.15 interaction pattern | `design-interaction-pattern` | ✅ |
| 3.16 deploy strategy | `design-deploy-strategy` | ✅ |
| 3.17 artifact storage | `design-artifact-storage` | ✅ |
| 3.18 embedding / search | `design-embedding-search` | ✅ |
| 3.19 MCP server | `design-mcp-server` | ✅ |
| 3.20 ADR | `write-adr` | ✅ |
| 3.21 anti-bias gate | `verify-best-alternative` (engineering-only, ADR-018) | ✅ |
| 3.22 architecture review | `review-architecture` | ✅ |
| 3.23 engineering plan review | `review-engineering` + anti-rationalization 5 규칙 (HIGH commit `830eb04`) | ✅ |
| 3.24 design review | `review-design` | ✅ |
| 3.25 DX review | `review-devex` | ✅ |
| 3.26 scope review | `review-scope` | ✅ |
| 3.27 external LLM consult | `consult-codex` | ✅ |
| (추가) | `design-billing-system` | ✅ — payment / billing infrastructure (3.7/3.8 보강) |
| (추가) | `design-claude-hooks` | ✅ — plugin 트랙 특화 |
| (추가) | `audit-ui-quality` | ✅ — design 측 audit |
| (추가) | `prototype-from-spec` | ✅ — design exploration |
| (추가) | `critique-plan` | ✅ — strategic critique |

**Phase 3 cover**: 26/27 (96%) + 부분 1 (4%) + hole 0. 가장 강한 영역.

### §3.4 Phase 4 — Implementation Planning

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 4.1 feature → actor track | `decompose-feature-to-actor-tracks` | ✅ |
| 4.2 track → atomic task | `decompose-track-to-tasks` | ✅ |
| 4.3 task dependency graph | `map-task-dependencies` | ✅ |
| 4.4 parallel execution plan | `plan-parallel-execution` | ✅ |
| 4.5 acceptance test plan | `define-acceptance-test-plan` | ✅ |
| 4.6 build timeline | `estimate-build-timeline` | ✅ |
| 4.7 tracker publication | `publish-to-tracker` (MID-3) | ✅ |
| 4.8 strategic critique | `critique-plan` | ✅ |
| 4.9 auto-review pipeline | `autoplan` (sub-orchestrator) | ✅ |

**Phase 4 cover**: 9/9 (100%)

### §3.5 Phase 5 — Implementation

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 5.1 TDD cycle | `build-with-tdd` | ✅ |
| 5.2 pair programming | `pair-program-loop` | ✅ |
| 5.3 parallel agent dispatch | `dispatch-parallel-agents` | ✅ |
| 5.4 iterative fix → commit → verify | `iterate-fix-verify` | ✅ |
| 5.5 edit scope freeze | `freeze-edit-scope` | ✅ |
| 5.6 bug diagnosis | `diagnose-bug` + `loop-methods.md` (MID-2) | ✅ |
| 5.7 stuck state decomposition | `decompose-blocker` | ✅ |
| 5.8 refactor with rename | `refactor-with-rename-trace` | ✅ |
| 5.9 contract → SDK / stub | `generate-from-api-contract` | ✅ |
| 5.10 spec → test skeleton | `generate-tests-from-spec` | ✅ |
| 5.11 code → docs sync | `update-docs-with-code` | ✅ |

**Phase 5 cover**: 11/11 (100%)

### §3.6 Phase 6 — Verification

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 6.1 QA tier classification | `classify-qa-tiers` | ✅ |
| 6.2 per-actor test | `test-per-actor-use-case` | ✅ |
| 6.3 cross-actor flow test | `test-cross-actor-flow` | ✅ |
| 6.4 load test | `run-load-test` | ✅ |
| 6.5 chaos test | `chaos-test` | ✅ |
| 6.6 browser QA | `run-browser-qa` | ✅ |
| 6.7 test coverage | `audit-test-coverage-meaningful` | ✅ |
| 6.8 code health | `measure-code-health` | ✅ |
| 6.9 security audit | `audit-security` | ✅ |
| 6.10 a11y audit | `audit-accessibility` | ✅ |
| 6.11 i18n coverage | `audit-i18n-coverage` | ✅ |
| 6.12 cost audit | `audit-cost-efficiency` | ✅ |
| 6.13 live devex audit | `audit-live-devex` | ✅ |
| 6.14 regression monitoring | `monitor-regressions` | ✅ |
| 6.15 review risk classification | `classify-review-risks` | ✅ |
| 6.16 AI safety review | `review-ai-safety-liability` | ✅ |
| 6.17 privacy review | `review-privacy-data-risk` | ✅ |
| 6.18 license / IP review | `review-license-and-ip-risk` | ✅ |
| 6.19 ToS / policy readiness | `review-terms-policy-readiness` | ✅ |
| 6.20 verification before completion | `verification-discipline.md` (MID-1, cross-skill SSoT) | ✅ |
| (추가) | `verify-quality` (§6 orchestrator) | ✅ — phase orchestrator 가 19 sub-skill 조율 |

**Phase 6 cover**: 21/21 (100%) — 가장 광범위.

### §3.7 Phase 7 — Release

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 7.1 quality gate setup | `setup-quality-gates` | ✅ |
| 7.2 feature flag | `setup-feature-flags` | ✅ |
| 7.3 canary deploy | `setup-canary-deploy` | ✅ |
| 7.4 rollback runbook | `setup-rollback-runbook` | ✅ |
| 7.5 release tagging | `automate-release-tagging` | ✅ |
| 7.6 changelog | `write-changelog` | ✅ |
| 7.7 UAT | `run-uat` | ✅ |
| 7.8 beta program | `run-beta-program` | ✅ |
| 7.9 launch checklist | `prepare-launch-checklist` | ✅ |
| 7.10 incident paging | `setup-incident-paging` | ✅ |
| 7.11 release docs sync | `sync-release-docs` | ✅ |
| 7.12 auto PR | `auto-create-pr` | ✅ |
| 7.13 branch finalization | `finish-development-branch` (MID-4) | ✅ |
| 7.14 destructive guard | `guard-destructive-commands` + `git-safety-rules.md` | ✅ |
| 7.15 multi-safety | `compose-safety-mode` | ✅ |
| (추가) | `ship-release` (§7 orchestrator) | ✅ |

**Phase 7 cover**: 15/15 (100%)

### §3.8 Phase 8 — Operations (engineering 측)

| Step | buddy skill | 매핑 |
|------|-------------|------|
| 8.1 incident response | `handle-incident` | ✅ |
| 8.2 postmortem | `conduct-postmortem` | ✅ |
| 8.3 error budget | `audit-error-budget` | ✅ |
| 8.4 cost anomaly | `analyze-cost-anomaly` | ✅ |
| 8.5 actor failure rate | `analyze-actor-failure-rate` | ✅ |
| 8.6 continuous monitoring | `monitor-regressions` (재등장, §6 와 동일) | ✅ |

**Phase 8 (engineering 측) cover**: 6/6 (100%)

### §3.9 Cross-cutting

| Step | buddy skill | 매핑 |
|------|-------------|------|
| C.1 anti-bias at decisions | `verify-best-alternative` (ADR-018) | ✅ |
| C.2 anti-rationalization at reviews | `review-engineering` 5 규칙 (HIGH) | ✅ |
| C.3 Iron Law at completion | `router/references/verification-discipline.md` (MID-1) | ✅ |
| C.4 git safety | `router/references/git-safety-rules.md` (MID-4) | ✅ |
| C.5 context save / restore | `save-context` + `restore-context` | ✅ |
| C.6 pattern documentation | `write-a-skill` | ✅ |
| C.7 ADR 작성 | `write-adr` | ✅ |
| C.8 context handoff (문서) | `docs/HANDOFF.md` + `docs/handoff/` | 🟡 부분 — *skill 형식* 부재, 문서만 |
| **C.9 context compaction** | — | ❌ **hole** (caveman 흡수 후보 — engineering-audit Tier 3) |
| **C.10 ubiquitous language audit** | — | ❌ **hole** (engineering-audit Tier 1 후보) |

**Cross-cutting cover**: 7/10 (70%) + 부분 1 (10%) + hole 2 (20%)

---

## §4. 갭 식별 — buddy 미보유 / 부분 cover step

### §4.1 명확한 hole (5건)

| ID | Step | 영향 | 권장 |
|----|------|------|------|
| **H1** | 2.4 ubiquitous language | 코드 식별자 ↔ PRD 어휘 drift = bug / 인지비용 ↑ | `audit-ubiquitous-language` 신규 skill (mattpocock adopt-with-edits) |
| **H2** | C.9 context compaction | long-running session 토큰 bloat → 작업 중단 / re-bootstrap | `compress-context` 신규 skill (caveman 흡수) — 단 cli buddy 트랙 검토 |
| **H3** | C.10 ubiquitous language audit (cross-cutting 차원) | H1 과 동일, cross-phase 호출 가능 차원 | H1 과 동일 skill 로 cover |
| H4 | (잠재) 도메인 어휘 evolution tracking | 시간 경과 시 어휘 의미 drift | 후속 wave (post-H1) |
| H5 | (잠재) Knowledge persistence audit | learning JSONL 누적 시 의미 일관성 | `persist-learning-jsonl` 보강 |

### §4.2 부분 cover (5건)

| ID | Step | 부족한 부분 | 권장 |
|----|------|----------|------|
| **P1** | 1.6 constraints identification | *기술 제약* 만 cover, *규제 / 시간 / 비즈니스* 제약 분산 | 신규 skill `enumerate-constraints` 또는 `assess-business-viability` 본문 보강 |
| **P2** | 2.3 domain modeling | *entity* cover, *VO / aggregate / bounded context* 명시 없음 | `design-data-model` 본문에 DDD 차원 추가 |
| **P3** | 2.5 bounded context | *system boundary* cover, *DDD bounded context* 어휘 부재 | `derive-system-topology` 본문 보강 또는 H1 skill 과 묶음 |
| **P4** | 3.1 architecture style | *style 선택 결정* 명시 skill 없음, `design-system` orchestrator 진입만 | 신규 skill `select-architecture-style` 또는 `define-tech-stack` 본문 보강 |
| **P5** | C.8 context handoff | *skill 형식* 부재, 문서만 (HANDOFF.md / handoff/) | 신규 skill `produce-handoff-doc` 또는 `save-context` 본문 보강 |

### §4.3 우선순위 — 작업물 직접 영향 기준

| Tier | 갭 | 비용 | 가치 |
|------|----|------|------|
| **Tier 1** | H1 + H3 (ubiquitous language) — 1 skill 로 cover | 중 (신규 skill 1건) | **High** — refactor / bug 회피 / contributor 인지 직결 |
| Tier 2 | P2 + P3 (DDD bounded context / VO/aggregate) | 저 (기존 skill 본문 보강) | Mid-High — H1 과 시너지 |
| Tier 2 | P1 (constraints) | 저 | Mid — 1 phase 입력 정합 |
| Tier 2 | P4 (architecture style 결정) | 중 | Mid — design phase 진입 명확화 |
| Tier 3 | H2 (context compaction) | 중-상 | Mid — long-running session 효율 |
| Tier 3 | P5 (handoff skill 형식) | 저 | Low — 문서 형식 우월 |

---

## §5. 충돌 / 책임 모호

### §5.1 명확한 중복 (catalog 등재 중복, inventory G2/G3 와 동일)

| ID | Skill | 위치 | 해소 |
|----|-------|------|------|
| C1 | `monitor-regressions` | catalog §6 + §8 등재, PROCEDURE.md 1개 | §8 유지, §6 제거 또는 Pattern Library 분리 (inventory Tier 1) |
| C2 | `apply-builder-ethos` | catalog §1 + Cross-cutting 등재, PROCEDURE.md 1개 | Cross-cutting 유지, §1 제거 (inventory Tier 1) |

### §5.2 잠재 책임 모호 (4건)

| ID | 영역 | skill 후보 | 모호 지점 |
|----|------|----------|----------|
| M1 | 도메인 모델 / boundary | `design-data-model` / `derive-system-topology` / `map-use-cases-to-infra` / `compose-feature-from-use-cases` | DDD bounded context 결정 시 *어느 skill 진입* — 4 skill 책임 경계 불명. H1 신설 시 더 모호 |
| M2 | 리뷰 stage | `review-engineering` / `review-architecture` / `review-scope` / `review-design` / `review-devex` / `critique-plan` | 6 skill — *어떤 review 가 언제* 명시 부족. autoplan 4-mode 가 *부분 정합* 하지만 critique-plan 은 별도 |
| M3 | Security | `audit-security` / `review-ai-safety-liability` / `review-privacy-data-risk` / `review-license-and-ip-risk` / `design-secret-management` | 5 skill 영역 overlap — *어느 시점 무엇* 명확 |
| M4 | Test | `define-acceptance-test-plan` / `generate-tests-from-spec` / `test-per-actor-use-case` / `test-cross-actor-flow` / `build-with-tdd` / `audit-test-coverage-meaningful` | 6 skill — *plan vs gen vs execute vs audit* 책임 경계 |

### §5.3 권장 해소

- 각 영역 (M1-M4) 별로 `routing-rules.md` §3 에 *책임 매트릭스* 1 페이지 추가
- 또는 각 skill PROCEDURE.md 의 *Step 0 (선행 확인)* 에 *나는 무엇 / 다른 skill 은 무엇* 명시

---

## §6. Cascade 정합 — Command 맥락 지점

### §6.1 자연 cascade (9-phase 자동 chain)

`router/SKILL.md` 의 9-phase orchestrator 가 *phase 단위 cascade* 보장:

```
§1 concretize-idea
  → §2 define-features
    → §3 design-system
      → §4 plan-build
        → §5 build-feature
          → §6 verify-quality
            → §7 ship-release
              → §8 iterate-product
                → §9 manage-lifecycle
```

**문제**: 위 cascade 는 *phase 진입* 만 자연. *intra-phase stage skill 간 cascade* 는 부분적 / 암묵.

### §6.2 명시 cascade (PROCEDURE.md 본문 hint)

예: `decompose-blocker` 의 PROCEDURE.md 마지막 *"이 skill 의 출력 → 다음 skill candidate 4 개"* 명시. 이 패턴이 *완전 적용* 된 skill 비율 추정 **~30%**.

### §6.3 단절 지점 (예시)

| Skill 종료 | 다음 단계 *명시 hint* 부재 | 사용자 진입 시 마찰 |
|----------|------------------------|-------------------|
| `design-api-contract` 완료 | `design-data-model` / `design-event-schema` / `generate-from-api-contract` 중 어느 것이 *다음*? | 사용자가 *수동 결정* |
| `define-feature-spec` 완료 | `score-feature-priority` / `estimate-feature-effort` / `map-feature-dependencies` / `define-acceptance-test-plan` 어느 것? | 사용자가 *수동 결정* |
| `diagnose-bug` 완료 (Phase 3 root cause 발견) | `iterate-fix-verify` / `refactor-with-rename-trace` / `decompose-blocker` 어느 것? | 부분 명시 (loop-methods.md MID-2) |
| `review-engineering` 완료 (anti-rationalization 5 규칙 적용 후 findings 출력) | `iterate-fix-verify` / `build-with-tdd` (regression test) / `verify-quality` 어느 것? | 단절 |
| `verify-quality` 완료 (QA report sign-off) | `ship-release` 자연 cascade, 단 *fail 시* `iterate-fix-verify` 진입 hint 부재 | 부분 단절 |
| `setup-feature-flags` 완료 | `setup-canary-deploy` / `prepare-launch-checklist` 어느 것? | 단절 |

### §6.4 권장 cascade 보강

1. **각 stage skill 의 PROCEDURE.md 마지막에 "Next" 섹션 강제** — write-a-skill 의 표준 항목으로 추가
2. **router/references/routing-rules.md 에 *stage cascade 표* 추가** — phase 단위 외 stage skill 간 자연 순서 명시
3. **autoplan 의 4-mode review (scope/eng/design/devex) 사용 시 *findings 출력 → fix skill cascade* hint 명시**

---

## §7. 우선순위 — 갭 / 충돌 / cascade 보강

| Tier | 작업 | 비용 | 가치 | 의존 |
|------|------|------|------|------|
| **Tier 1** | H1 `audit-ubiquitous-language` 신규 skill | 중 (1-2 일) | **High** — 작업물 직접 영향 | engineering-audit Tier 1 (LOW 2) 와 동일 |
| **Tier 1** | C1/C2 catalog 중복 제거 (inventory G2/G3) | 저 (~30 분) | Mid — 라우팅 정확도 | — |
| Tier 2 | P2 + P3 보강 — `design-data-model` / `derive-system-topology` 본문에 DDD 차원 | 저-중 | Mid-High | H1 후 시너지 |
| Tier 2 | M1-M4 책임 매트릭스 — `routing-rules.md` §3 보강 | 중 (~1 일) | Mid — 라우팅 충돌 해소 | — |
| Tier 2 | §6.4 cascade "Next" 섹션 표준화 + 30% → 80% 적용 | 상 (수 일, 다 skill 본문 편집) | **High** — 사용자 진입 마찰 해소 | write-a-skill 표준 갱신 필요 |
| Tier 3 | P1 / P4 / P5 보강 | 중 | Mid | — |
| Tier 3 | H2 `compress-context` (cli buddy 트랙 검토 후) | 중-상 | Mid | cli buddy Wave 7 W7-2 trigger |
| Tier 3 | 광의 audit 후속 wave — description 품질 + dispatch 정확도 측정 | 상 | Mid | — |

---

## §8. 본 문서의 한계 / 가정

### §8.1 명시 가정

| # | 가정 | 잠재 반론 |
|---|------|----------|
| G1 | 이론 baseline 13 source 통합이 *공평* | 통합 시 *PMBOK vs Continuous Delivery* 의 phase 정의 충돌 — 본 문서는 *임의 정합* |
| G2 | 광의 engineering = §2-§8 의 ~106 skill | 사용자 의도가 더 좁을 수 있음 (예: §3-§6 만) |
| G3 | 각 buddy skill 의 *cover 평가* 가 catalog description (1 줄) 만으로 정확 | 실제 PROCEDURE.md 본문 read 시 cover 평가 변동 가능 |
| G4 | "cascade hint 명시 ~30%" 추정 | 측정 없이 추정 — 실 read 시 차이 가능 |
| G5 | H1 (ubiquitous language) 가 가장 High value gap | H2-H5 / P1-P5 와 *상대 가치* 측정 없음 |

### §8.2 미수행 작업 (후속 wave 후보)

- 각 buddy skill PROCEDURE.md 본문 직접 read + cover 재평가 (현재는 catalog description 기반)
- ~106 skill 의 *실 dispatch routing 정확도* 측정 (sample 호출 + grade)
- 이론 baseline 13 source 의 *내부 충돌* 명시 (예: Beck TDD vs PMBOK waterfall)
- *Cross-cutting 5-Tier* (skill 보강 적용 시 expected impact 정량)

---

## §9. References

### 9.1 buddy 자체 SSoT

- [`plugin-skills-inventory.md`](./plugin-skills-inventory.md) — 152 skill 전체 inventory
- [`plugin-skills-engineering-audit.md`](./plugin-skills-engineering-audit.md) — SKILLS_ANALYSIS A.1 13 영역 좁은 audit
- [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) — 17 source repo / 167 skill 분석
- [`two-tracks-charter.md`](./two-tracks-charter.md) — plugin / cli 트랙 정체성
- [`plugin/skills/router/references/skill-catalog.md`](../plugin/skills/router/references/skill-catalog.md) — 9-phase × Priority catalog
- [`plugin/skills/router/references/routing-rules.md`](../plugin/skills/router/references/routing-rules.md) — 라우팅 충돌 결정
- [`plugin/skills/router/references/verification-discipline.md`](../plugin/skills/router/references/verification-discipline.md) — Iron Law SSoT (MID-1)
- [`plugin/skills/router/references/git-safety-rules.md`](../plugin/skills/router/references/git-safety-rules.md) — git 안전 SSoT (MID-4)
- [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md) — 9-phase architecture spec

### 9.2 이론 baseline (외부, 참조 — 본문에서 verbatim 0건 유지)

- ISO/IEC 12207 (Software Life Cycle Processes)
- IEEE Std 1471 (Architecture Description)
- PMBOK 7판 (Project Management Body of Knowledge, PMI)
- Clean Code (Robert C. Martin)
- Pragmatic Programmer (Andrew Hunt, David Thomas)
- Continuous Delivery (Jez Humble, David Farley)
- Domain-Driven Design (Eric Evans)
- Refactoring (Martin Fowler)
- Test-Driven Development by Example (Kent Beck)
- Site Reliability Engineering (Google)
- Accelerate (Nicole Forsgren, Jez Humble, Gene Kim)
- Building Microservices (Sam Newman)
- Working Effectively with Legacy Code (Michael Feathers)

---

## §10. 본 문서 변경 정책

본 mapping 은 *snapshot* (2026-05-25 baseline). 다음 시점 갱신:

| trigger | 갱신 부분 |
|---------|----------|
| H1 (ubiquitous-language) skill 신설 | §3.2 / §4.1 / §7 / §5.2 M1 책임 매트릭스 |
| C1/C2 catalog 중복 제거 (inventory Tier 1) | §5.1 |
| P2/P3 본문 보강 (DDD 차원 추가) | §4.2 / §3.2 |
| Cascade "Next" 섹션 표준화 적용 | §6 전반 |
| 광의 audit 후속 wave (PROCEDURE 본문 read 기반 재평가) | §3 매핑 매트릭스 / §8.1 G3-G4 |
| 이론 baseline 추가 source 채택 | §1.1 / §2 |

본 문서가 stale 해지면 **본 문서부터 갱신**. inventory / engineering-audit / catalog 와 충돌 시 — 본 mapping 이 *광의 SE coverage 책임*, 다른 문서는 각자 책임 영역 SSoT.
