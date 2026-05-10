# Missing Skills Inventory — 2026-05-10

> **목적**: plugin buddy 의 미구현 skill 종합 list 작성. [`docs/two-tracks-charter.md`](../two-tracks-charter.md) §6.1 의 진화 1 순위 작업 ("미구현 skill 보완 — `docs/tasks.md` §A-2 의 잔여 + `@docs/` 추가 검토 발견 항목") 의 *Step 1 산출물*.
>
> **Step 2 (외부 참조 4 경로 탐색)** + **Step 3 (매트릭스 작성)** + **Step 4 (per-skill 결정 plan)** 의 입력 자료로 사용.
>
> **위치**: `docs/notes/2026-05-10-missing-skills-inventory.md`

---

## 1. 입력 docs (6 종) + 검증 방법

### 1.1 검토 대상 6 docs

| # | docs | 검토 포인트 |
|---|------|----------|
| 1 | `docs/tasks.md` §A-2 | 잔여 29 skill (cluster A~G 분류) — baseline |
| 2 | `docs/skill-map.md` "추천 신규 buddy 스킬 목록" | 17 추천 skill (Matt skills 기반 + buddy 자체 추가) |
| 3 | `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` §4 + §10 | 단계별 Skill 군집화 + Gap, 잔여 35 generic 표현 |
| 4 | `docs/superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md` Cluster A~H | 8 cluster 분류 + immediate value 평가 |
| 5 | `docs/HANDOFF.md` plugin buddy 트랙 상태 표 | 진행 상태 + Open Questions |
| 6 | `docs/two-tracks-charter.md` §6.1 plugin buddy 진화 방향 | 사용자 의도 lock-in (12 stage scope) |

### 1.2 검증 방법

각 docs 에서 추출한 skill 명을 `plugin/skills/<name>/` 디렉토리 존재 여부로 검증.

| 출처 | 추출 | plugin/skills/ 미존재 (= 미구현) | plugin/skills/ 존재 (= stale) |
|------|------|--------------------------|------------------------|
| docs 1 (tasks.md A-2) | 29 | **29** | 0 |
| docs 2 (skill-map.md 17 추천) | 17 | **4** | 13 |
| docs 3 (spec §4 generic) | — (cluster 표현) | docs 1 과 정합 | — |
| docs 4 (phase7 8 cluster) | — (cluster 분류) | docs 1 과 대부분 정합. **Cluster H (MCP 2)** 만 별도 | — |
| docs 5 (HANDOFF) | — | docs 1 ~ 4 와 정합 | — |
| docs 6 (charter §6.1) | — | scope 12 stage × plugin buddy 9-phase 매핑 시 추가 *gap 후보* 발견 | — |

> plugin/skills/ 실측 = **106 디렉토리** (2026-05-10 기준). HANDOFF.md "105 procedures" 와 +1 차이 — `router/` 디렉토리 포함 여부 차이로 추정.

---

## 2. 미구현 skill 종합 — 4 그룹 분류

### 2.1 그룹 1 — 직접 미구현 29 (`tasks.md` §A-2)

명확한 *신규 작성 후보*. plugin/skills/ 실재 검증으로 모두 미구현 확인.

#### 2.1.1 §1 concretize-idea — Cluster E (5)

| skill | 1줄 용도 | charter scope 매핑 |
|-------|---------|------------------|
| `analyze-competition-and-substitutes` | 경쟁 / 대체재 분석 (positioning matrix + moat) | 2 사업성 분석 |
| `map-customer-segments` | 사용자 vs 구매자 세그먼트 분리 + persona | 2 사업성 분석 |
| `map-jobs-to-be-done` | JTBD 프레임워크 적용 | 1 idea 발굴 + 디벨롭 |
| `analyze-market-size` | TAM / SAM / SOM 산출 + bottom-up / top-down cross-check | 2 사업성 분석 |
| `conduct-customer-interview` | 인터뷰 스크립트 + 결과 코딩 | 1 idea 발굴 |

#### 2.1.2 §2 define-features (1)

| skill | 1줄 용도 | charter scope 매핑 |
|-------|---------|------------------|
| `estimate-feature-effort` | feature 단위 effort estimation (T-shirt + ideal-h) | 5 설계 |

#### 2.1.3 §3 design-system 부가 — Cluster B residual (4)

| skill | 1줄 용도 | charter scope 매핑 |
|-------|---------|------------------|
| `design-observability` | logs / metrics / traces 3-pillar 설계 + SLO / SLI | 5 설계 |
| `design-secret-management` | secret store + rotation + audit + leak detection | 5 설계 |
| `design-i18n-strategy` | locale 분기 + ICU MessageFormat + RTL + 번역 워크플로우 | 5 설계 |
| `design-accessibility-baseline` | WCAG 2.2 AA baseline + a11y annotation 컨벤션 | 4 디자인 + 5 설계 |

#### 2.1.4 §5 build-feature — Cluster D (5)

| skill | 1줄 용도 | charter scope 매핑 |
|-------|---------|------------------|
| `generate-from-api-contract` | OpenAPI / GraphQL → 클라이언트 / 서버 stub 자동 생성 | 6 빌드 |
| `generate-tests-from-spec` | spec → unit / integration test skeleton | 6 빌드 + 7 자동화 테스트 |
| `pair-program-loop` | red-green-refactor pair driving 컨텍스트 | 6 빌드 |
| `refactor-with-rename-trace` | LSP rename + 호출 그래프 추적 | 6 빌드 + 12 유지보수 |
| `update-docs-with-code` | 코드 변경 → README / ADR / changelog 동기화 | 12 유지보수 |

#### 2.1.5 §6 verify-quality 부가 — Cluster A residual (3)

| skill | 1줄 용도 | charter scope 매핑 |
|-------|---------|------------------|
| `audit-i18n-coverage` | locale 별 번역 누락 / fallback 누수 검출 | 7 자동화 테스트 |
| `chaos-test` | failure injection (network / pod kill / latency) | 7 자동화 테스트 |
| `audit-test-coverage-meaningful` | line coverage 가 아니라 mutation / behavior 검증 | 7 자동화 테스트 |

#### 2.1.6 §8 iterate-product — Cluster F (7)

| skill | 1줄 용도 | charter scope 매핑 |
|-------|---------|------------------|
| `analyze-feature-adoption` | 신규 기능 활성화율 + power user 코호트 | 9 A/B 테스트 + 10 그로스 해킹 |
| `analyze-user-cohort` | acquisition cohort retention curve | 10 그로스 해킹 |
| `analyze-actor-failure-rate` | actor 별 실패 패턴 + 회복 전략 | 12 유지보수 |
| `analyze-cost-anomaly` | 비용 spike 탐지 + root cause | 12 유지보수 |
| `triage-customer-support-ticket` | 티켓 분류 + recurring issue 패턴화 | 12 유지보수 |
| `analyze-customer-feedback-corpus` | NPS / 리뷰 텍스트 토픽 모델링 | 11 마케팅 지원 |
| `audit-error-budget` | SLO 잔여 burn rate + release 통제 | 12 유지보수 |

#### 2.1.7 §9 manage-lifecycle — Cluster G (4)

| skill | 1줄 용도 | charter scope 매핑 |
|-------|---------|------------------|
| `deprecate-feature` | deprecation timeline + sunset notice + telemetry | 12 유지보수 |
| `migrate-customers` | major change 시 고객 마이그레이션 plan | 12 유지보수 |
| `archive-product` | EOL 체크리스트 + data export + tombstone | 12 유지보수 |
| `spin-off-feature` | 기능 분리 후 별도 product / repo 로 전환 | 12 유지보수 |

> **그룹 1 합계: 29 skill**.

---

### 2.2 그룹 2 — skill-map.md 추가 후보 4 (영역 겹침 — 결정 lock-in 2026-05-10)

`skill-map.md` "추천 신규 buddy 스킬 목록" 의 17 중 13 은 이미 구현 (build-with-tdd, diagnose-bug, setup-quality-gates, define-feature-spec 등). 미구현 4 건의 *영역 겹침 결정* 은 사용자 confirm (2026-05-10) 으로 lock-in.

| skill | 1줄 용도 | 영역 겹침 | 결정 |
|-------|---------|----------|------|
| `define-product-context` | Matt skills `grill-with-docs` 기반 — 도메인 컨텍스트 / ADR 기반 운영 | `define-product-spec` (구현됨) | **통합** — define-product-spec 의 *internal sub-section* 으로 흡수. 별도 skill 미작성 |
| `write-prd` | Matt skills `to-prd` 기반 — PRD 작성 | `define-product-spec` (구현됨, 동일 책임) | **폐기** — define-product-spec 가 동일 책임 cover |
| `review-code-architecture` | Matt skills `improve-codebase-architecture` 기반 — deep module + interface depth | `review-engineering` (구현됨), `consult-codex` (구현됨) | **흡수** — review-engineering 의 stage 로 통합 |
| `review-legal-regulatory` | buddy 자체 추가 후보 — 법률 / 규제 영향 검토 | (없음) | **신규 작성** — Cluster E 와 묶어 §1 사업성 영역으로 |

> **그룹 2 결정 결과: 신규 작성 1 (review-legal-regulatory) + 기존 skill 보강 3 (define-product-spec / review-engineering 의 PROCEDURE 갱신)**.

---

### 2.3 그룹 3 — MCP 2 (Q4 보류 재평가 — 결정 lock-in 2026-05-10: C2 분할 진입)

phase7-deferred-reevaluation.md 의 Cluster H. 사용자 confirm (2026-05-10) 으로 *분할 진입* 결정.

| MCP server | 1줄 용도 | 결정 | 트랙 |
|-----------|---------|------|------|
| `feature-management-mcp` | feature.query / store / update / link_code / export_patch — feature registry MCP | **cli buddy 트랙으로 분리** — charter §5 의 `cmd/buddy-mcp/` "두 트랙 공유" 위치 활용. plugin buddy inventory 에서 제외 | cli buddy |
| `analytics-mcp` | funnel / AB / cohort 노출 | **§8 일부 구현 후 자연 진입** — Cluster F (§8 데이터 분석 7 skill) 일부 완료가 trigger. plugin buddy inventory 에 잔존 | plugin buddy (deferred) |

> **그룹 3 결정 결과: plugin buddy inventory 에 1 MCP (analytics-mcp, deferred) 잔존, 1 MCP (feature-management-mcp) 는 cli buddy 트랙으로 이동**.

---

### 2.4 그룹 4 — charter scope 12 stage × 9-phase 매핑 gap (Step 2 외부 탐색 후 결정 — 2026-05-10 lock-in)

**결정 (2026-05-10)**: 사용자 confirm — *Step 2 외부 4 경로 탐색 후 명명 / scope 결정*. 외부 자산이 form-factor / 그로스 / 마케팅 영역 단서 제공 가능.



charter §3 의 plugin buddy scope 12 stage 와 plugin buddy 의 9-phase orchestrator 를 매핑할 때 *cover 안 되는 영역* 식별. 신규 발견 후보.

| charter scope | 9-phase 매핑 | gap 추정 후보 | 우선순위 |
|--------------|------------|------------|--------|
| 1 idea 발굴 + 디벨롭 | §1 | 그룹 1 의 5 skill 으로 cover | 0 |
| 2 사업성 분석 | §1 | 그룹 1 의 5 + 그룹 2 의 1 (review-legal-regulatory) | 0 |
| 3 앱 vs web 결정 | §3 design-system 의 일부 | (없음? — 명시적 *form-factor 결정* skill 부재 가능. `decide-form-factor-app-vs-web` 후보) | **신규 발견** |
| 4 디자인 적용 | §3 design 부가 | 그룹 1 의 design-accessibility-baseline + (별도 *visual design* skill 부재 가능. `apply-design-system` 후보) | **신규 발견** |
| 5 설계 | §3 + §4 | 모두 cover | 0 |
| 6 빌드 | §5 | 그룹 1 의 5 build skill | 0 |
| 7 자동화 테스트 | §6 | 그룹 1 의 3 audit | 0 |
| 8 배포 | §7 | 모두 cover (Phase 3 v1.0.4) | 0 |
| 9 A/B 테스트 | §8 | 그룹 1 의 일부 (analyze-feature-adoption / -user-cohort) | 0 |
| 10 그로스 해킹 | §8 + (별도?) | (별도 *growth experiment* skill 부재 가능. `design-growth-experiment`, `optimize-conversion-funnel` 후보) | **신규 발견** |
| 11 마케팅 지원 | §8 일부 | 그룹 1 의 analyze-customer-feedback-corpus + (별도 *마케팅 채널 / 콘텐츠* skill 부재 가능. `plan-marketing-channel`, `automate-content-publishing` 후보) | **신규 발견** |
| 12 유지보수 | §8 + §9 | 그룹 1 의 다수 cover | 0 |

> **그룹 4 합계: 4 영역의 신규 후보** (skill 명은 *임시*. 사용자 confirm 후 정식 제안). 실제 작성 시 1~7 skill 사이 변동 가능.

---

## 3. 종합 매트릭스 (2026-05-10 결정 반영)

| 그룹 | 건수 | 결정 | plugin buddy 신규 작성 | 우선순위 |
|------|-----|------|---------------------|--------|
| **그룹 1** 직접 미구현 (A-2) | 29 | 신규 작성 또는 외부 참조 통합 | **29** | HIGH (charter 1 순위) |
| **그룹 2** skill-map 추가 (영역 겹침) | 4 | 통합 3 + 신규 1 | **1** (review-legal-regulatory) | MEDIUM (Cluster E 묶음) |
| **그룹 3** MCP — Q4 보류 재평가 | 2 | C2 분할 진입 | **0 + 1 deferred** (analytics-mcp §8 후) | DEFERRED |
| **그룹 4** charter scope gap | 4+ 영역 | Step 2 외부 탐색 후 결정 | **TBD** | MEDIUM-LOW |
| **합계 (확정)** | **30 신규 + 1 deferred + 3 통합 + 4 영역 TBD** | | | |

> *통합 3* (define-product-context / write-prd / review-code-architecture) 는 신규 skill 작성이 아니라 *기존 PROCEDURE 갱신* 이라 plugin/skills/ 디렉토리 카운트 영향 없음 (현재 106 → 신규 30 후 = 136).
> *cli buddy 트랙으로 이동* 1 (feature-management-mcp) 은 plugin buddy inventory 에서 제외됨.

### 3.1 charter scope cover 율

| stage | cover 상태 |
|-------|----------|
| 1 idea 발굴 | 그룹 1 으로 cover (5 신규 작성 시 100%) |
| 2 사업성 분석 | 그룹 1 + 그룹 2 review-legal-regulatory 로 cover |
| 3 앱 vs web 결정 | 🟡 gap 후보 (그룹 4) |
| 4 디자인 적용 | 🟡 partial — 그룹 1 design-accessibility-baseline 만 |
| 5 설계 | 🟢 cover (현재) |
| 6 빌드 | 그룹 1 5 신규 작성 시 cover |
| 7 자동화 테스트 | 그룹 1 3 신규 작성 시 cover |
| 8 배포 | 🟢 cover (현재) |
| 9 A/B 테스트 | 그룹 1 일부 신규 작성 시 cover |
| 10 그로스 해킹 | 🟡 gap 후보 (그룹 4) |
| 11 마케팅 지원 | 🟡 gap 후보 (그룹 4) |
| 12 유지보수 | 그룹 1 다수 신규 작성 시 cover |

→ 그룹 1 + 그룹 2 + 그룹 4 모두 진행 시 charter scope 12 stage 모두 cover.

---

## 4. 다음 Step 진행 지침

### 4.1 Step 2 — 외부 참조 4 경로 탐색 (다음 응답)

| 경로 | 탐색 목적 |
|------|--------|
| `/Users/kevin/work/github/aidax-dag/ai-cli/skill/<projects>/` | skill 영역 — 그룹 1 + 그룹 2 + 그룹 4 매칭 후보 |
| `/Users/kevin/work/github/aidax-dag/ai-cli/agent/<projects>/` | agent 영역 — plugin/agents/ 자산 영역 + cli buddy spec 작성 시 reference |
| `/Users/kevin/work/github/aidax-dag/ai-cli/harness/<projects>/` | harness pattern — orchestrator / dispatch / cascade 패턴 참고 |
| `/Users/kevin/work/github/aidax-dag/ai-cli/mcp/<projects>/` | MCP server 영역 — 그룹 3 (deferred) + plugin/mcp/ 자산 영역 |

### 4.2 Step 3 — 매트릭스 작성

본 inventory 의 39+ × Step 2 의 외부 자산 매트릭스. 각 cell:
- 외부 자산 *유사 / 같음 / 무관*
- 호환성 / 궁합 점검 결과 (PROCEDURE 양식 / dispatch 패턴 / cascade 정합 / 단어 일관성)
- 결정 옵션 (a) 그대로 차용 / (b) 수정 차용 / (c) 참고만 + 신규 작성 / (d) 무관 + 신규 작성

### 4.3 Step 4 — per-skill 결정 plan + 통합 작업 plan

각 미구현 skill 마다 결정 영구화. plan 위치 후보: `docs/superpowers/plans/<date>-skill-completion-plan.md`.

### 4.4 Step 5 — 실제 작성 + commit

skill 별 PR 단위. plan 따라 진행.

---

## 5. 사용자 confirm — 완료 (2026-05-10)

| 결정 | 선택 |
|------|------|
| D-A (그룹 4 처리) | Step 2 외부 4 경로 탐색 후 결정 (Recommended) |
| D-B (그룹 2 영역 겹침) | 통합 3 + 신규 1 (Recommended) |
| D-C (그룹 3 MCP 재평가) | C2 분할 진입 — analytics-mcp §8 후 / feature-management-mcp cli buddy 분리 (Recommended) |

원본 옵션 / 영향 분석은 git history 의 `404fd1e` commit message 참조.

## 5'. (historical) confirm 옵션 원본

### 5.1 그룹 4 (charter scope gap) 신규 후보 4 영역

본 inventory 가 식별한 4 신규 영역:

| 영역 | 임시 skill 후보 | charter stage | confirm 사항 |
|------|--------------|--------------|------------|
| 3 앱 vs web 결정 | `decide-form-factor-app-vs-web` | 3 | 명시적 form-factor 결정 stage 가 필요한가? §3 design-system 의 *비공식 sub-step* 으로 충분한가? |
| 4 디자인 적용 | `apply-design-system` | 4 | charter 의 "디자인 적용" 이 *기존 design-system 채택* 인지 / *신규 design 작성* 인지? |
| 10 그로스 해킹 | `design-growth-experiment`, `optimize-conversion-funnel` | 10 | 사용자가 *그로스 해킹* 으로 의미하는 작업 범위? §8 의 AB / funnel skill 로 충분한가 / 별도 stage 필요한가? |
| 11 마케팅 지원 | `plan-marketing-channel`, `automate-content-publishing` | 11 | "마케팅 지원" 의 구체 형태 — 컨텐츠 자동화 / 채널 분석 / 캠페인 운영 중 어디까지? |

### 5.2 그룹 2 (skill-map 추가 4) 영역 겹침 결정

| skill | 결정 옵션 |
|-------|---------|
| `define-product-context` | (a) `define-product-spec` 의 sub-section 으로 통합 / (b) 별도 skill 신규 |
| `write-prd` | (a) `define-product-spec` 의 alias / (b) rename / (c) 폐기 |
| `review-code-architecture` | (a) `review-engineering` 의 stage 로 흡수 / (b) 별도 skill 신규 |
| `review-legal-regulatory` | 신규 작성 결정 (자체 후보 — 영역 겹침 없음) |

### 5.3 그룹 3 (MCP 2) trigger 재평가

Q4=(c) 보류 결정 (2026-05-06 spec) 이 본 charter §3 의 plugin buddy 정의 (MCP 가 *4 자산 중 하나*) 와 정합 — *MCP 가 보류* 라는 결정은 charter 시점에 재평가 필요.

| 옵션 | 영향 |
|------|------|
| (a) 보류 유지 | charter 와 silent conflict — plugin buddy scope 의 MCP 영역 미진행 |
| (b) Cluster F (§8 데이터 분석) 일부 구현 후 진입 | phase7 plan 과 정합 |
| (c) 즉시 진입 | charter §6.1 1 순위 (skill 보완) 와 병행 |

---

## 6. 검증 (본 inventory 작성 시점)

| 항목 | 측정 |
|------|------|
| 입력 6 docs 검토 | 완료 |
| plugin/skills/ 실재 검증 — A-2 29 | 29/29 미구현 확인 (✗) |
| plugin/skills/ 실재 검증 — skill-map 17 | 13 stale + 4 미구현 |
| 미구현 종합 | 그룹 1 (29) + 그룹 2 (4) + 그룹 3 (2 MCP) + 그룹 4 (4+ 영역) = **39+ 후보** |
| charter scope 12 stage cover 율 | 5 stage 🟢 cover, 4 stage 그룹 1 작성 시 cover, 3 stage 🟡 gap (그룹 4) |
| 본 inventory 줄 수 | (작성 후 측정) |

---

## 7. 다음 액션

1. 본 inventory commit
2. 사용자 §5 confirm (그룹 4 신규 영역 4 + 그룹 2 영역 겹침 4 + 그룹 3 MCP trigger 재평가)
3. confirm 결과 반영 → Step 2 외부 4 경로 탐색 진입
4. Step 3 매트릭스 작성 → Step 4 plan → Step 5 실제 작성

본 inventory 는 *Step 1 산출물* 로 lock-in. 미래 cycle 에서 *진척 측정 baseline* 으로 활용.
