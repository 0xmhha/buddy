# Engineering Decision Gate Mapping — verify-best-alternative invoke 시점 정의

**Status**: Accepted (2026-05-21)
**Authors**: mhha (plugin track)
**Related**: ADR-018 (verify-best-alternative rename + forced gate + decompose-blocker), SKILLS_ANALYSIS § A
**Purpose**: buddy 9-phase lifecycle 전반에서 `verify-best-alternative` 가 *언제·어디서·어떻게* 호출되어야 하는지의 단일 SSoT. ADR-018 §2 forced gate의 *실효성 확보*를 위해 결정 Step 본문에 추가될 instruction의 근거가 됨.

---

## 1. Scope — engineering decision vs others

`verify-best-alternative` 는 *엔지니어링 결정 한정* (ADR-018 + 2026-05-21 본문 rewrite로 확정). 따라서 본 매핑은 *엔지니어링 결정성 스킬*만 대상으로 한다.

### 1.1 Engineering decision (in scope)

요구사항·환경 제약 하에서 *기술적 alternatives* 중 베스트 선택지를 결정하는 모든 시점. 구체적으로:

- 아키텍처 / 시스템 분해
- 스택 / 언어 / framework / DB / hosting 선택
- 데이터 모델 / normalization / index 전략
- API shape / protocol 결정
- 인증 / 인가 / federation 모델
- 격리 / multi-tenancy 전략
- 이벤트 스키마 / DLQ / idempotency 결정
- 시크릿 관리 / rotation 전략
- 알고리즘 / 자료구조 선택
- 코드 네이밍 (함수·타입·변수·모듈)
- prompt engineering (AI 기능 구현)
- 관찰성 (logs / metrics / traces) 도구 선택
- 배포 전략 (canary / blue-green / rolling)
- 작업 분해·dependency·병렬 실행 plan 결정

### 1.2 Out of scope (별도 스킬, 미래 작업)

다음은 본 매핑·본 forced gate의 대상이 *아님*. 미래에 별도의 *non-engineering review skill* 신설 시 그쪽에서 다룬다:

- 그래픽 디자인 (typography·color·layout·grid)
- 브랜드 / 제품 네이밍 (마케팅 가치 중심)
- microcopy / UI text / 에러 메시지 voice
- 마케팅 카피 / GTM 채널 / 광고 카피
- 사업 기획 옵션 (TAM·SAM·SOM 외 strategy)
- 결제 모델·tier·pricing 옵션 (비즈니스 측면)
- a11y / WCAG baseline (UX 영역)
- 인터랙션 패턴 / gesture / motion (UX 영역)

§3에 있는 `apply-design-system`, `design-interaction-pattern`, `design-accessibility-baseline` 은 *UX/디자인 영역* → 본 forced gate **대상 아님**.

---

## 2. Phase-by-phase decision skill 매핑

각 phase의 결정성 스킬을 *engineering scope 적합도*로 분류.

### §1 Idea & Business Validation — N/A

비즈니스/idea 결정 영역. verify-best-alternative scope 밖. (필요 시 별도 *business review skill* 신설)

| 스킬 | 결정 종류 | gate 대상 |
|------|----------|----------|
| validate-idea / validate-advanced-edge-idea / assess-business-viability / map-customer-segments / analyze-competition / decide-target-market / review-pricing-and-gtm / define-product-spec / review-legal-regulatory | 비즈니스·idea·legal | ❌ scope 밖 |

### §2 Feature Definition & Backlog — Limited

feature priority / effort estimation 결정은 *비즈니스 가치 + 엔지니어링 비용*의 hybrid. 순수 엔지니어링 결정 아님. verify-best-alternative 의무화 대상 아님.

| 스킬 | 결정 종류 | gate 대상 |
|------|----------|----------|
| score-feature-priority / estimate-feature-effort / split-work-into-features / triage-work-items | priority·effort | ⚠️ optional (호출 가능하나 의무 아님) |
| identify-actors / map-actor-use-cases / compose-feature-from-use-cases / define-feature-spec | factual mapping | ❌ scope 밖 (사실 매핑) |

### §3 Technical Design — Primary scope

본 forced gate의 *주 영역*. 모든 엔지니어링 design 결정이 여기 집중됨.

#### 3-A. 이미 wire됨 (Phase 1, 2026-05-21)

ADR-018 §2 적용 완료. §11 Verification gate에 체크박스 추가됨:

| 스킬 | 결정 Step | wire 상태 |
|------|----------|----------|
| `design-system` | Stage 2 토폴로지 후보 생성 | ✅ Stage 2 본문에서 invoke 명시 |
| `define-tech-stack` | Phase 4. 선택 + 결정 근거 | ✅ §11 체크박스 |
| `design-api-contract` | Phase 1. API style 선택 | ✅ §11 체크박스 |
| `design-data-model` | Phase 3. Normalization 결정 | ✅ §11 체크박스 |
| `design-event-schema` | Phase 2. Schema format + registry | ✅ §11 체크박스 |
| `design-auth-model` | Phase 의 mechanism 결정 | ✅ §11 체크박스 |
| `design-tenant-model` | Phase 의 model 결정 (shared/schema/DB) | ✅ §11 체크박스 |
| `design-secret-management` | Stage 1 Secret store 결정 | ✅ §6 체크박스 |

**보강 필요** (A2-2 작업): 위 8개 모두 *§11 체크박스*만 있고, *결정 Step 본문*에 invoke instruction 없음. 결정 후 체크박스만 체크하는 회피 가능. 본문에 *결정 Step 시작 부분*에 invoke 강제 instruction 추가 필요.

#### 3-B. 추가 wire 대상 (Phase 2, 권장)

§3 design 스킬 중 *엔지니어링 결정 + 미wire*인 항목:

| 스킬 | 결정 종류 | wire 우선순위 |
|------|----------|-------------|
| `design-observability` | 3 pillar 도구 선택 (Datadog vs Prometheus vs OTel) | HIGH (lock-in 크고 비용 다양) |
| `design-deploy-strategy` | canary / blue-green / rolling / recreate | HIGH (운영 영향 큼) |
| `derive-system-topology` | service / data flow / trust boundary 결정 | MID (design-system 안에서 호출되긴 함) |
| `map-use-cases-to-infra` | actor × use case → infra component 매핑 | MID (매핑이라 결정 폭 작음) |
| `design-mcp-server` | MCP server 설계 결정 | MID (지정된 protocol이라 선택지 좁음) |
| `design-embedding-search` | BM25+vector hybrid 결정 | MID (각 component 선택) |
| `design-billing-system` | 결제 시스템 설계 결정 *(엔지니어링 측면만)* | MID (결제 비즈니스 측면은 scope 밖, *기술 구현*만) |
| `design-artifact-storage` | immutable artifact 저장·배포 설계 | LOW (선택지 비교적 명확) |
| `design-claude-hooks` | hook 표준 설계 | LOW (spec 따라야 함) |
| `design-i18n-strategy` | locale / fallback 전략 결정 | LOW (대부분 best practice 따름) |
| `decide-form-factor-app-vs-web` | app vs web vs hybrid vs desktop | HIGH (5년 lock-in) |

#### 3-C. UX/디자인 영역 (gate 대상 아님)

| 스킬 | 이유 |
|------|------|
| `apply-design-system` | 그래픽 디자인 라이브러리 선택 (shadcn/MUI/HIG) — UX 영역 |
| `design-interaction-pattern` | gesture / motion / feedback — UX 영역 |
| `design-accessibility-baseline` | WCAG / a11y — UX 영역 |

#### 3-D. Review 페르소나 스킬 (호출 *내부*에서 가능, gate 아님)

| 스킬 | 역할 | verify-best-alternative와 관계 |
|------|------|-----------------------------|
| `review-architecture` / `review-engineering` / `review-scope` | 페르소나 리뷰 (post-decision) | verify-best-alternative는 *pre-decision*. 두 스킬 서로 *상보적*, gate 아님 |
| `review-design` / `review-devex` | designer / devex 페르소나 | UX 영역, gate 아님 |
| `critique-plan` | strategic critique | post-plan critique, gate 아님 |

### §4 Implementation Plan — Limited gate

implementation plan은 *기술적 결정*이지만 §3 design 결정의 *집행*에 가까움. 그러나 *decomposition / dependency / parallel execution* 자체에 trade-off 있음.

| 스킬 | 결정 종류 | wire 우선순위 |
|------|----------|-------------|
| `decompose-feature-to-actor-tracks` | actor 분해 전략 (frontend/backend/3rd-party/data) | MID |
| `decompose-track-to-tasks` | task 분해 입자도 | MID |
| `map-task-dependencies` | DAG 구성 | LOW (factual mapping) |
| `plan-parallel-execution` | batch + sync points | HIGH (capability fit · critical path 최적화) |
| `define-acceptance-test-plan` | test 전략 결정 | MID |
| `estimate-build-timeline` | timeline 추정 (확률적) | LOW (추정이라 alternatives 적음) |

### §5 Development — Excluded (per ADR-018)

build-with-tdd / refactor-with-rename-trace / iterate-fix-verify 등은 *결정 집행*이지 *결정* 아님. ADR-018 §2에서 명시적으로 excluded.

### §6 Quality — Limited

| 스킬 | 결정 종류 | wire 우선순위 |
|------|----------|-------------|
| `classify-qa-tiers` | QA intensity (Quick/Standard/Exhaustive) | MID (3 tier 미리 정의돼 선택지 좁음) |
| `classify-review-risks` | review risk 11 category 분류 | LOW (factual classification) |
| `audit-test-coverage-meaningful` | coverage trust score 계산 | LOW (factual measurement) |

### §7~§9 — Limited

대부분 *집행* 또는 *measurement*. 결정성 스킬:

| 스킬 | 결정 종류 | wire 우선순위 |
|------|----------|-------------|
| §7 `setup-canary-deploy` | canary 단계·gate metric 결정 | MID (§3 design-deploy-strategy에서 이미 큰 결정) |
| §7 `setup-feature-flags` | flag system 설계 | MID |
| §7 `setup-rollback-runbook` | rollback decision tree | MID |
| §8 `design-ab-experiment` | 실험 설계 (가설·표본 크기·기간) | MID |
| §8 `plan-growth-experiment` | ICE/RICE 가설 우선순위 | LOW (priority frame) |

---

## 3. Forced invocation instruction template

§11 Verification gate 체크박스는 *post-decision* — 회피 가능. 본문 결정 Step에 *pre-decision* instruction 추가 필요.

### 3.1 본문 instruction 표준 형식

각 결정 Step의 *시작 부분*에 다음 1~2 문장 삽입:

```markdown
> **AI 편향 차단 게이트** — 본 Step에서 [결정 종류] 후보를 *3개 이상 발산*시킨 뒤 평가한다. 발산은 본 Step에서 직접 또는 `verify-best-alternative` 호출로 처리. *첫 답으로 commit 금지*. 본문에서 명시적으로 3개 이상의 *서로 orthogonal한* 후보가 나란히 비교되지 않으면 §11 검증 게이트 통과 금지.
```

### 3.2 §11 Verification gate 보강

기존 체크박스(*post-decision 확인*) 유지 + 다음 anti-rationalization 줄 추가:

```markdown
- [ ] `verify-best-alternative` 1회 이상 호출 완료 — AI 편향 방지 의무. *체크박스만 체크하고 실제 발산 안 한 회피 금지* (산출물에 3개 이상 후보의 rubric 비교 표 존재로 증명)
```

### 3.3 §3 Phase 1, 3-A wire 8개 보강 작업 (A2-2 task)

기존 wire된 8개 스킬의 결정 Step 본문에 §3.1 instruction 추가:

| 스킬 | 본문 추가 위치 |
|------|-------------|
| `define-tech-stack` | Phase 4 시작 (line ~128) |
| `design-api-contract` | Phase 1 시작 (line ~70) |
| `design-data-model` | Normalization decision Phase (Phase 3 부근) |
| `design-event-schema` | Phase 2 시작 (line ~90) |
| `design-auth-model` | mechanism 결정 Phase |
| `design-tenant-model` | model 결정 Phase |
| `design-secret-management` | Stage 1 (line ~29) |
| `design-system` | Stage 3 Tech Stack 선택 부분에 추가 reinforcement |

---

## 4. Out-of-scope skill 신설 backlog (미래 작업)

본 매핑이 *out of scope*로 분류한 영역의 별도 review skill 신설 필요 시점에 다음 후보:

| 영역 | 후보 스킬 이름 (잠정) | 목적 |
|------|---------------------|------|
| 그래픽 디자인 | `verify-best-visual-direction` (TBD) | 시각 디자인 옵션 (typography·color·layout) 다관점 검토 |
| 브랜드 / 제품 네이밍 | `verify-best-brand-naming` (TBD) | 제품·feature 마케팅 네이밍 후보 비교 |
| 마케팅 카피 / 채널 | `verify-best-marketing-direction` (TBD) | 카피·채널·crew 옵션 비교 |
| 사업 전략 옵션 | `verify-best-business-direction` (TBD) | 사업 모델·tier·pricing 옵션 비교 |
| UX / 인터랙션 | `verify-best-ux-pattern` (TBD) | 인터랙션·gesture·motion 옵션 비교 |

각 신설 시 동일 메커니즘(N variants·rubric·비교 framework)을 *각 도메인의 rubric*에 맞춰 적용. 별 ADR + spec 필요.

---

## 5. Verification

- §3 3-A의 8개 스킬 *본문 instruction 추가* 완료 시 forced gate 실효 확보 (A2-2 task)
- §3 3-B 추가 wire 진행 시 본 매핑의 우선순위 표 참조
- 향후 새 §3 design 스킬 추가 시 본 매핑 §2.3-A·3-B에 등재 + write-a-skill PROCEDURE의 forced-gate 가이드 적용
- §1·§2·§5 등 *out of scope* 영역의 결정성 스킬에서 verify-best-alternative 호출 시도 → router가 catalog description 보고 *redirect 권유* (예: "본 결정은 비즈니스 영역. business review skill은 미래 작업 — 현재는 critique-plan 또는 consult-codex 사용")

---

## 6. References

- ADR-018 — verify-best-alternative rename + forced gate 정책
- `plugin/skills/verify-best-alternative/PROCEDURE.md` — 본 스킬 본문 (엔지니어링 한정)
- `plugin/skills/router/references/skill-catalog.md` — 결정성 스킬 카탈로그
- SKILLS_ANALYSIS.md § A — 잔여 정리 사항 (외부 분석 문서)
- ADR-003 — 외부 자산 attribution 정책 (관련 없음, 참조만)
