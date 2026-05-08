# Phase 7 (Deferred) 재평가 — v1.0.5 시점

**Goal:** Phase 1-4 완료 후 (v1.0.5) deferred 영역 31 skill 의 immediate value 재평가. 어느 cluster 가 (a) 즉시 구현 가치 / (b) 조건부 구현 (특정 trigger 발생 시) / (c) 장기 deferred 인지 분류.

**SSoT:** [`2026-05-06-stage-buildout-plan.md`](./2026-05-06-stage-buildout-plan.md) Phase 5/6/7 + [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../specs/2026-05-06-lifecycle-orchestrator-architecture.md) §10.

## 현재 상태 (v1.0.5)

| Phase | Status | 잔여 skill | 근거 |
|-------|--------|-----------|------|
| 1 (§3 핵심 4) | [Done] v1.0.2 | 0 | tech stack / data model / API contract / ADR — 락인 영향 큰 결정 cover |
| 2 (§4 6) | [Done] v1.0.3 | 0 | feature → actor track → task DAG → batch → test plan → timeline 완성 |
| 3 (§7 7) | [Done] v1.0.4 | 0 | safety nets — canary / flags / rollback / UAT / beta / launch / paging |
| 4 (§6 use-case test 2) | [Done] v1.0.5 | 0 | Q8=(a) cascade 완성: use case → system → actor track → build → test |
| 5 (§8 데이터 7) | [Deferred] | 7 | analyze-feature-adoption / cohort / actor-failure-rate / cost-anomaly / customer-support / feedback-corpus / error-budget |
| 6 (§9 Lifecycle 4) | [Deferred] | 4 | deprecate / migrate / archive / spin-off |
| 7 (잔여) | [Deferred] | ~31 | §1 customer 5 + §3 부가 11 + §5 부가 5 + §6 부가 6 + §3 use case→infra 2 + MCP 2 |

**총 잔여**: ~42 skill (v1.0.5 기준 plugin 의 97 skill 의 30%).

## Cluster 별 immediate value 분석

### Cluster A: §6 부가 audit 6 — **HIGH immediate value**

| Skill | 가치 | Trigger / 즉시성 |
|-------|------|------------------|
| `run-load-test` | 상용 launch 직전 SLO 검증 필수 — k6 nightly 만으론 부족 (sustained load + chaos 조합 필요) | **즉시 가치** — Phase 3 (§7) 의 prepare-launch-checklist row 의 Performance category 보강 |
| `audit-accessibility` | a11y compliance (WCAG 2.1 AA) — public 제품 launch 직전 의무에 가까움 | **즉시 가치** — launch checklist Engineering 또는 Legal axis 보강 |
| `audit-cost-efficiency` | Infracost 보강, $/MAU 단위 economics 정량 | **즉시 가치** — launch checklist Cost & Business axis 보강 |
| `audit-i18n-coverage` | multi-region launch 시 의무 | 조건부 — single-region 한정 launch 면 deferred 가능 |
| `chaos-test` | resiliency 검증 — high availability SLO 약속 시 의무 | 조건부 — 99.9% 약속 한정 |
| `audit-test-coverage-meaningful` | Stryker mutation testing — coverage % 의 가짜 양성 잡음 | 중기 가치 — initial coverage 가 정착된 후 |

**Cluster A 권고**: 3 high-priority skill (load / a11y / cost) → **Phase 5 extension** 으로 즉시 진입 가능. 3 conditional → trigger 발생 시 (multi-region / 99.9% commit / mature coverage).

### Cluster B: §3 부가 design 핵심 3 — **MEDIUM immediate value**

| Skill | 가치 | 즉시성 |
|-------|------|--------|
| `design-event-schema` | async event flow (SQS / Kafka / webhook) 의 schema-first design | **즉시 가치** — design-api-contract 의 sync-only gap 채움. SaaS auth example 의 EmailEnqueued / EmailBounced 등이 이미 생산 수요 |
| `design-auth-model` | OAuth2 / JWT / SAML / SSO / RBAC 의 다층 결정 | **즉시 가치** — auth-heavy 제품에서 design-tech-stack 보다 deep |
| `design-tenant-model` | multi-tenant 결정 (shared row / schema / DB per tenant) | **즉시 가치** — multi-tenant SaaS 의 공통 결정 |

| Skill | 가치 | 즉시성 |
|-------|------|--------|
| `design-observability` | OTel + 로그 + 메트릭 + 트레이스 + 알람 통합 plan | 중기 — define-tech-stack 에 일부 cover, 보강은 production ops 후 |
| `design-secret-management` | secret rotation / vault / DR | 중기 — 위와 유사 |
| `design-i18n-strategy` | 본격 i18n (DB level, URL, 콘텐츠) | 조건부 |
| `design-accessibility-baseline` | a11y design 표준 (color / focus / motion) | 조건부 |

**Cluster B 권고**: 3 high-priority (event-schema / auth-model / tenant-model) → **Phase 5 extension** 또는 ad-hoc 작성. 4 medium/conditional → 조건부 trigger.

### Cluster C: §3 use case → infra 브릿지 2 — **HIGH cascade value**

| Skill | 가치 | 즉시성 |
|-------|------|--------|
| `map-use-cases-to-infra` | actor system boundary → 실제 infra component 매핑 — Q8=(a) cascade 의 §2 → §3 transition 의 explicit step | **즉시 가치** — Q8 cascade 의 missing layer 채움 |
| `derive-system-topology` | actor 그래프 → 시스템 토폴로지 자동 도출 | **즉시 가치** — design-system orchestrator 의 sub-step 강화 |

**Cluster C 권고**: 2 skill 모두 **Phase 5 extension** 진입 — Q8=(a) cascade 의 정합성 강화.

### Cluster D: §5 부가 build 5 — **LOW-to-MEDIUM immediate value**

| Skill | 가치 | 즉시성 |
|-------|------|--------|
| `generate-from-api-contract` | OpenAPI → handler / model / client codegen | 중기 — 첫 API 작성 후 코드량 줄이는 가치, 그러나 codegen 도구 표준 (orval / openapi-zod-client) 으로 대체 가능 |
| `generate-tests-from-spec` | Schemathesis 와 일부 중복, 보강 가치 있으나 한정 | 낮음 |
| `pair-program-loop` | Claude × Claude 또는 Claude × human pair pattern | 중기 |
| `refactor-with-rename-trace` | rename refactor 의 cross-file trace | 낮음 — IDE 도구로 충분 |
| `update-docs-with-code` | code change → doc 자동 sync | 중기 — sync-release-docs (이미 Done) 와 일부 중복 |

**Cluster D 권고**: 전체 **deferred** 유지. ROI 가 codegen 도구 / IDE 와 비교해 marginal.

### Cluster E: §1 customer/market 5 — **CONDITIONAL value**

| Skill | 가치 | 즉시성 |
|-------|------|--------|
| `analyze-competition-and-substitutes` | 경쟁분석 — pre-PRD | 조건부 — 사용자가 idea 단계 (PRD 없는) 일 때만 가치 |
| `map-customer-segments` | segment 정의 | 조건부 |
| `map-jobs-to-be-done` | JTBD framework | 조건부 |
| `analyze-market-size` | TAM/SAM/SOM | 조건부 |
| `conduct-customer-interview` | 고객 인터뷰 진행 plan | 조건부 |

**Cluster E 권고**: **deferred** 유지. concretize-idea 에 보강 가능하지만 buddy 사용자가 idea 부터 시작하는 비율 추정 후 결정.

### Cluster F: §8 데이터 분석 보강 7 — **CONDITIONAL on production traffic**

| Skill | 가치 | 즉시성 |
|-------|------|--------|
| `analyze-feature-adoption` | feature 별 adoption rate 추적 | production 후 |
| `analyze-user-cohort` | cohort retention | production 후 |
| `analyze-actor-failure-rate` | actor 별 failure rate | production 후 |
| `analyze-cost-anomaly` | cost spike 탐지 | production 후 |
| `triage-customer-support-ticket` | ticket triage 자동화 | production 후 |
| `analyze-customer-feedback-corpus` | feedback 분석 | production 후 (beta corpus 는 run-beta-program 이 cover) |
| `audit-error-budget` | SLO error budget tracking | production + SLO 정의 후 |

**Cluster F 권고**: **deferred** 유지. production traffic + 6+ 개월 데이터 누적 후 가치 발휘.

### Cluster G: §9 Lifecycle 4 — **LONG-term deferred**

| Skill | 가치 | 즉시성 |
|-------|------|--------|
| `deprecate-feature` | feature deprecation runbook | 1년+ |
| `migrate-customers` | customer migration plan | 1년+ |
| `archive-product` | product EOL plan | 2년+ |
| `spin-off-feature` | feature → 독립 product spin-off | 2년+ |

**Cluster G 권고**: **deferred** 유지. 1년+ 운영 maturity 의무.

### Cluster H: MCP 2 — **Q4=(c) 보류 유지 결정**

| MCP | 가치 | 즉시성 |
|-----|------|--------|
| `feature-management-mcp` | feature registry 통합 — buddy MCP 의 feature_* 가 부분 cover | 부분 done, 보강 가치 |
| `analytics-mcp` | 분석 통합 | Cluster F 와 함께 production 후 |

**Cluster H 권고**: **deferred** 유지. Q4=(c) 결정 변경 없음.

---

## 종합 권고: Phase 5 extension candidate (8 skill)

다음 8 skill 을 **Phase 5 extension** (즉시 진입 가능) 후보로 식별:

| Cluster | Skill | 예상 우선순위 |
|---------|-------|----------------|
| A (§6 audit) | `run-load-test` | 1 |
| A (§6 audit) | `audit-accessibility` | 2 |
| A (§6 audit) | `audit-cost-efficiency` | 3 |
| B (§3 design) | `design-event-schema` | 4 |
| B (§3 design) | `design-auth-model` | 5 |
| B (§3 design) | `design-tenant-model` | 6 |
| C (§3 bridge) | `map-use-cases-to-infra` | 7 |
| C (§3 bridge) | `derive-system-topology` | 8 |

**Phase 5 extension 의 Out of scope (deferred 유지):**
- Cluster D (§5 부가 build 5) — IDE/codegen 도구로 대체 가능
- Cluster E (§1 customer/market 5) — 사용자 segment 측정 후 결정
- Cluster F (§8 production analytics 7) — production traffic + 6+ 개월 데이터 후
- Cluster G (§9 Lifecycle 4) — 1년+ 운영 maturity 후
- Cluster H (MCP 2) — Q4=(c) 결정 변경 시

---

## 결정 옵션 (사용자 선택)

### Option A: Phase 5 extension 즉시 진입 (8 skill)

**장점**: launch-grade 제품 빌딩에 가장 직접적인 가치. Q8 cascade 의 §3 layer 강화 (Cluster C). prepare-launch-checklist 의 row 보강 (Cluster A).
**단점**: ~8 추가 commit + Phase 5 plan 작성. v1.0.5 → v1.0.6 추가 1 release.

### Option B: 선택적 진입 (Cluster A 만, 3 skill)

**장점**: launch checklist 의 가장 부족한 row (load / a11y / cost) 만 채움. 시간 적게.
**단점**: §3 cascade 강화 (Cluster B + C) 보류.

### Option C: deferred 유지 (현 상태)

**장점**: v1.0.5 가 cohesive 한 milestone. live retest 결과 + push 가 우선.
**단점**: 즉시 가치 있는 8 skill 미실현.

### Option D: Phase 5 plan 작성만 + 구현은 trigger 발생 시

**장점**: 식별 + 우선순위만 영속화. 즉시 시간 소비 없음.
**단점**: code 산출 없음.

---

## 권고

**Option C (deferred 유지)** + **(D) Live retest 결과 docs + push** 를 우선 마무리.

**근거:**
1. v1.0.5 가 Q8=(a) cascade 완성 + Phase 1-4 정합 milestone — 여기서 한 번 stabilize 하는 것이 전체 코히어런시 좋음.
2. live retest 14/14 (Phase 1+2+3+4) 가 완료된 산출물이라 documentation + push 가 자연스러운 closure.
3. Phase 5 extension 진입 시 또 1+ release cycle (commits + retest + docs sync) 필요 — overhead 감안 시 다음 session 에 진입이 합리.
4. 본 재평가 doc 자체가 "deferred 정확히 무엇이 deferred 인지" 의 영속 record — 차후 trigger 발생 시 즉시 plan 작성 가능 base.

사용자 결정에 따라 Option A/B 진입 가능.
