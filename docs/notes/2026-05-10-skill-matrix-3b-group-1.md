# Skill Matrix 3b — Group 1 (29 skills) × External Candidates

> **Step 3b 산출**: [`docs/notes/2026-05-10-missing-skills-inventory.md`](./2026-05-10-missing-skills-inventory.md) 의 그룹 1 (직접 미구현 29 skill) × 외부 자산 (Step 2 inventory + 추가 탐색) 매트릭스. Step 3a 의 신규 발견 (marketingskills 5 + designer-skills/design-research 가 그룹 1 영역 침투) 을 입력으로 활용.
>
> Step 3a 와 동일한 *4 결정 옵션* 사용: (a) 그대로 차용 / (b) 수정 차용 / (c) 참고만 + 신규 / (d) 무관 + 신규.

---

## 1. 검토 방법

[`2026-05-10-skill-matrix-3a-groups-2-4.md`](./2026-05-10-skill-matrix-3a-groups-2-4.md) §1 과 동일 (목적 정합 / 호환성 / 명명 / scope / 언어).

추가로 본 Step 3b 에서는 **외부 추가 탐색** — Step 2 inventory 에서 *명명 추정만* 한 프로젝트 중 README 첫 줄 검토:

| 신규 식별 후보 | 출처 | 영역 |
|------------|------|------|
| `superpowers` | skill 경로 | software development methodology (buddy docs/superpowers/ 와 명명 충돌, *주요 inspiration 출처 추정*) |
| `KESE-KIT` | skill 경로 | 한국 보안 (CII 취약점 분석평가) — 그룹 4-extension 후보 |
| `humanizer` | skill 경로 | AI 텍스트 자연화 — `draft-marketing-copy` 보조 |
| `gpt-researcher` | agent 경로 | research agent — §1 보조 |
| `autonomous-coding-agents` | agent 경로 | multi-agent orchestration — §5 보조 |
| `agent-evaluation` | agent 경로 | agent 평가 — §6 / §8 보조 |
| `memory-bank` | harness 경로 | Claude Code 대화 → knowledge graph |
| `tavily-mcp` | mcp 경로 | 검색 MCP — §1 customer research 보조 |
| `arxiv-mcp-server` | mcp 경로 | 학술 검색 MCP — §1 / §6 보조 |

---

## 2. cluster 별 매핑

### 2.1 §1 customer / market (5 skill, Cluster E)

| missing-skill | 외부 후보 | 결정 |
|--------------|--------|------|
| `analyze-competition-and-substitutes` | marketingskills/competitor-alternatives + competitor-profiling (직접 매칭 ⭐) | **(b) 수정 차용** |
| `map-customer-segments` | marketingskills/customer-research (보조) | **(d) 신규** + customer-research 참고 |
| `map-jobs-to-be-done` | (외부 직접 후보 없음 — JTBD 프레임워크) | **(d) 신규** |
| `analyze-market-size` | tavily-mcp (검색 보조) + arxiv-mcp-server (학술 보조) | **(d) 신규** + MCP tool 사용 |
| `conduct-customer-interview` | marketingskills/customer-research + designer-skills/design-research + gpt-researcher 보조 | **(b) 수정 차용** |

**§1 분포**: (b) 2 + (d) 3.

### 2.2 §2 feature effort (1 skill)

| missing-skill | 외부 후보 | 결정 |
|--------------|--------|------|
| `estimate-feature-effort` | (외부 직접 후보 없음 — T-shirt sizing / ideal-h estimation 일반 영역) | **(d) 신규** |

**§2 분포**: (d) 1.

### 2.3 §3 design 부가 (4 skill, Cluster B residual)

| missing-skill | 외부 후보 | 결정 |
|--------------|--------|------|
| `design-observability` | (외부 직접 후보 없음 — SLO/SLI 영역 일반) | **(d) 신규** |
| `design-secret-management` | (외부 직접 후보 없음 — secret rotation / leak 일반) | **(d) 신규** |
| `design-i18n-strategy` | (외부 직접 후보 없음 — ICU MessageFormat / RTL 일반) | **(d) 신규** |
| `design-accessibility-baseline` | designer-skills/interaction-design 의 일부 (a11y 패턴) | **(b) 수정 차용 (일부)** |

**§3 부가 분포**: (b) 1 + (d) 3.

### 2.4 §5 build 부가 (5 skill, Cluster D)

| missing-skill | 외부 후보 | 결정 |
|--------------|--------|------|
| `generate-from-api-contract` | autonomous-coding-agents (코드 생성 multi-agent) 보조 | **(d) 신규** + 참고 |
| `generate-tests-from-spec` | autonomous-coding-agents 의 QA agent 보조 | **(d) 신규** + 참고 |
| `pair-program-loop` | (외부 직접 후보 없음 — TDD red-green-refactor 일반) | **(d) 신규** |
| `refactor-with-rename-trace` | memory-bank (대화 → knowledge graph) 보조 | **(d) 신규** + 참고 |
| `update-docs-with-code` | (외부 직접 후보 없음 — README/ADR/changelog 동기화 일반) | **(d) 신규** |

**§5 분포**: (d) 5. — Cluster D 가 *IDE 대체 가능* 영역이라 외부 매칭 약함 (사용자 의도와 정합).

### 2.5 §6 verify 부가 (3 skill, Cluster A residual)

| missing-skill | 외부 후보 | 결정 |
|--------------|--------|------|
| `audit-i18n-coverage` | (외부 직접 후보 없음) | **(d) 신규** |
| `chaos-test` | (외부 직접 후보 없음 — chaos engineering 일반) | **(d) 신규** |
| `audit-test-coverage-meaningful` | agent-evaluation (agent 평가 패턴 — mutation / behavior 검증과 영역 가까움) | **(b) 수정 차용 (일부)** |

**§6 분포**: (b) 1 + (d) 2.

### 2.6 §8 data 분석 (7 skill, Cluster F)

| missing-skill | 외부 후보 | 결정 |
|--------------|--------|------|
| `analyze-feature-adoption` | marketingskills/analytics-tracking (직접 매칭 ⭐) | **(b) 수정 차용** |
| `analyze-user-cohort` | marketingskills/churn-prevention + analytics-tracking | **(b) 수정 차용** |
| `analyze-actor-failure-rate` | agent-evaluation (agent 실패 평가 일부) | **(b) 수정 차용 (일부)** |
| `analyze-cost-anomaly` | (외부 직접 후보 없음 — cost anomaly detection 일반) | **(d) 신규** |
| `triage-customer-support-ticket` | (외부 직접 후보 없음 — ticket 분류 일반) | **(d) 신규** |
| `analyze-customer-feedback-corpus` | marketingskills/customer-research 일부 + humanizer (텍스트 처리 보조) | **(b) 수정 차용** |
| `audit-error-budget` | (외부 직접 후보 없음 — SLO burn rate 일반) | **(d) 신규** |

**§8 분포**: (b) 4 + (d) 3.

### 2.7 §9 lifecycle (4 skill, Cluster G)

| missing-skill | 외부 후보 | 결정 |
|--------------|--------|------|
| `deprecate-feature` | (외부 직접 후보 없음 — sunset notice + telemetry 일반) | **(d) 신규** |
| `migrate-customers` | (외부 직접 후보 없음) | **(d) 신규** |
| `archive-product` | (외부 직접 후보 없음 — EOL checklist 일반) | **(d) 신규** |
| `spin-off-feature` | (외부 직접 후보 없음) | **(d) 신규** |

**§9 분포**: (d) 4. — Cluster G 가 *1 년+ deferred* 영역이라 외부 자산 자체가 적음 (성숙도 부족).

---

## 3. per-skill 결정 종합 매트릭스 (29 skill)

| # | cluster | skill | 결정 | 외부 입력 |
|---|--------|-------|------|---------|
| 1 | §1 E | `analyze-competition-and-substitutes` | (b) | marketingskills/competitor-* (2) |
| 2 | §1 E | `map-customer-segments` | (d) | marketingskills/customer-research (참고) |
| 3 | §1 E | `map-jobs-to-be-done` | (d) | — |
| 4 | §1 E | `analyze-market-size` | (d) | tavily-mcp + arxiv-mcp-server (도구) |
| 5 | §1 E | `conduct-customer-interview` | (b) | marketingskills/customer-research + designer-skills/design-research |
| 6 | §2 | `estimate-feature-effort` | (d) | — |
| 7 | §3 B | `design-observability` | (d) | — |
| 8 | §3 B | `design-secret-management` | (d) | — |
| 9 | §3 B | `design-i18n-strategy` | (d) | — |
| 10 | §3 B | `design-accessibility-baseline` | (b) | designer-skills/interaction-design |
| 11 | §5 D | `generate-from-api-contract` | (d) | autonomous-coding-agents (참고) |
| 12 | §5 D | `generate-tests-from-spec` | (d) | autonomous-coding-agents/QA (참고) |
| 13 | §5 D | `pair-program-loop` | (d) | — |
| 14 | §5 D | `refactor-with-rename-trace` | (d) | memory-bank (참고) |
| 15 | §5 D | `update-docs-with-code` | (d) | — |
| 16 | §6 A | `audit-i18n-coverage` | (d) | — |
| 17 | §6 A | `chaos-test` | (d) | — |
| 18 | §6 A | `audit-test-coverage-meaningful` | (b) | agent-evaluation |
| 19 | §8 F | `analyze-feature-adoption` | (b) | marketingskills/analytics-tracking |
| 20 | §8 F | `analyze-user-cohort` | (b) | marketingskills/churn-prevention + analytics-tracking |
| 21 | §8 F | `analyze-actor-failure-rate` | (b) | agent-evaluation |
| 22 | §8 F | `analyze-cost-anomaly` | (d) | — |
| 23 | §8 F | `triage-customer-support-ticket` | (d) | — |
| 24 | §8 F | `analyze-customer-feedback-corpus` | (b) | marketingskills/customer-research + humanizer |
| 25 | §8 F | `audit-error-budget` | (d) | — |
| 26 | §9 G | `deprecate-feature` | (d) | — |
| 27 | §9 G | `migrate-customers` | (d) | — |
| 28 | §9 G | `archive-product` | (d) | — |
| 29 | §9 G | `spin-off-feature` | (d) | — |

### 3.1 결정 분포 요약

| 결정 | 건수 | % |
|------|-----|---|
| (a) 그대로 차용 | 0 | 0% |
| (b) 수정 차용 | 8 | 28% |
| (c) 참고만 | 0 | 0% |
| (d) 신규 작성 | 21 | 72% |
| **합계** | **29** | **100%** |

→ 그룹 1 의 **72% 가 신규 작성**. 외부 매칭 28% (8 skill) — 모두 (b) 수정 차용.

### 3.2 cluster 별 외부 매칭 율

| cluster | (b) | (d) | 매칭 율 |
|---------|-----|-----|------|
| §1 E (5) | 2 | 3 | 40% |
| §2 (1) | 0 | 1 | 0% |
| §3 B (4) | 1 | 3 | 25% |
| §5 D (5) | 0 | 5 | 0% |
| §6 A (3) | 1 | 2 | 33% |
| §8 F (7) | 4 | 3 | **57%** |
| §9 G (4) | 0 | 4 | 0% |

→ **§8 F 데이터 분석** 이 외부 매칭 가장 풍부 (57%). marketingskills 가 그룹 1 영역 침투 신규 발견 효과 (Step 3a §8.1 의 prediction 적중).
→ **§5 D 빌드 / §9 G lifecycle** 외부 매칭 0% — IDE 대체 가능 / 1 년+ deferred 영역이라 자연.

---

## 4. 신규 발견 (외부 추가 탐색 결과)

### 4.1 `superpowers` (skill 경로)

README 첫 줄: *"Superpowers is a complete software development methodology for your coding agents, built on top of a set of composable skills and some initial instructions that make sure your agent uses them."*

→ **buddy 의 `docs/superpowers/` 와 명명 충돌 + 추정 inspiration 출처**.

| 시사점 |
|------|
| buddy 의 docs/superpowers/ 디렉토리 명명이 본 외부 자산 *영향 받았을 가능성 높음* |
| *composable skills + initial instructions* 패턴 = buddy 의 *router skill + PROCEDURE.md + dispatch* 패턴과 유사 |
| Step 3a / 3b 의 매트릭스 작업 자체가 superpowers 의 *methodology* 입력으로 활용 가능 |
| **별도 별도 ADR 후보**: superpowers 가 buddy 의 inspiration 출처라면 *명시적 attribution* (README 의 NOTICE 와 동일 패턴) 추가 검토 |

### 4.2 `KESE-KIT` (skill 경로)

README 첫 줄: *"주요정보통신기반시설(CII) 취약점 분석평가를 위한 Claude Code 스킬 플러그인"*

→ **그룹 4-extension 후보 추가**:
- 한국 보안 인증 영역 (CII 취약점 분석)
- *trigger*: 사용자가 한국 시장 진출 + 보안 인증 영역 진입 시
- 후보 명: `audit-cii-vulnerability` (한국 CII 특화) 또는 `consult-korean-security-compliance`

→ 본 발견 결과 **그룹 4-extension 후보 3 으로 확장** (기존 2 + 신규 1):
- `consult-korean-legal-context`
- `draft-patent-application`
- **`audit-cii-vulnerability` (신규 발견 — Step 3b)**

### 4.3 `humanizer` (skill 경로)

README 첫 줄: *"AI 텍스트 자연화 — Claude Code + OpenCode 호환"*

→ 그룹 4 의 `draft-marketing-copy` 의 *후처리 단계* 또는 sub-skill 후보. 단 *Cluster D 영역 (text editing tool)* 에 가까워 IDE 대체 가능. **현재 권장**: `draft-marketing-copy` 안에 *humanize 옵션* 으로 흡수.

---

## 5. 호환성 / 궁합 점검

### 5.1 외부 자산 호환성

| 외부 자산 | spec 호환 | 명명 | scope |
|---------|--------|------|------|
| marketingskills | 🟢 agentskills.io | 🟡 명사형 일부 | 단일 책임 |
| designer-skills | 🟢 Claude Code | (Step 3b 정밀 검토 시) | 단일 책임 |
| agent-evaluation | (검증 안 됨 — README 정보 부족) | (검증 필요) | (검증 필요) |
| autonomous-coding-agents | (검증 안 됨) | (검증 필요) | multi-agent orchestration (광범위) |
| memory-bank | (검증 안 됨) | (검증 필요) | knowledge graph |
| gpt-researcher | (검증 안 됨) | (검증 필요) | research agent |
| tavily-mcp | 🟢 MCP | (MCP server 명명 컨벤션) | 도구 (검색) |
| arxiv-mcp-server | 🟢 MCP | (동일) | 도구 (학술 검색) |
| superpowers | (검증 필요 — buddy 자산과 명명 충돌) | 🔴 충돌 | methodology (광범위) |
| KESE-KIT | 🟢 Claude Code | (한국 특화) | 한국 보안 |
| humanizer | 🟢 Claude Code + OpenCode | (검증 필요) | 단일 책임 |

### 5.2 (b) 수정 차용 시 작업 비용 추정

| skill | 작업 비용 |
|-------|--------|
| `analyze-competition-and-substitutes` | LOW — marketingskills 2 항목 통합 (PROCEDURE 12 section 양식 변환만) |
| `conduct-customer-interview` | LOW — marketingskills/customer-research 본문 + buddy stage 양식 |
| `design-accessibility-baseline` | MEDIUM — designer-skills/interaction-design 의 a11y *부분만* 추출 + WCAG 2.2 AA 가이드 보강 |
| `audit-test-coverage-meaningful` | MEDIUM — agent-evaluation 패턴이 *agent 평가* 영역, *test coverage* 로 변환 필요 |
| `analyze-feature-adoption` | LOW — marketingskills/analytics-tracking 직접 |
| `analyze-user-cohort` | LOW — marketingskills/churn-prevention + analytics-tracking 통합 |
| `analyze-actor-failure-rate` | MEDIUM — agent-evaluation 의 *agent 실패* → buddy 의 *actor 실패 패턴* 변환 |
| `analyze-customer-feedback-corpus` | MEDIUM — marketingskills/customer-research + 텍스트 처리 (humanizer 보조) 통합 |

→ (b) 8 건 중 LOW 4 + MEDIUM 4. (b) 평균 비용 LOW-MEDIUM.

### 5.3 (d) 21 건 신규 작성 비용 추정

| 영역 | (d) 건수 | 평균 비용 |
|------|------|--------|
| §1 customer 신규 (3) | 3 | MEDIUM (customer / market 도메인 지식) |
| §2 effort (1) | 1 | LOW (T-shirt sizing 표준) |
| §3 부가 (3) | 3 | MEDIUM-HIGH (observability / secret / i18n 도메인 깊이) |
| §5 build (5) | 5 | LOW (코드 작업 표준 패턴) |
| §6 verify (2) | 2 | MEDIUM (i18n-coverage / chaos 도메인) |
| §8 data (3) | 3 | MEDIUM-HIGH (cost anomaly / ticket triage / error budget 도메인) |
| §9 lifecycle (4) | 4 | MEDIUM (deprecation / migration 표준) |

→ (d) 21 건 평균 비용 MEDIUM. 일부 §3 부가 / §8 data 가 HIGH (도메인 깊이 필요).

---

## 6. Step 3 종합 — 이번 cycle 의 *외부 자산 차용 비율*

| 그룹 | (a) | (b) | (c) | (d) | 합계 |
|------|-----|-----|-----|-----|------|
| 그룹 1 (29) | 0 | 8 | 0 | 21 | 29 |
| 그룹 2 신규 (1) | 0 | 0 | 0 | 1 | 1 |
| 그룹 4 (11) | 0 | 9 | 0 | 2 | 11 |
| **합계 (41)** | **0** | **17** | **0** | **24** | **41** |
| % | 0% | **41%** | 0% | **59%** | 100% |

→ **41 신규 작성 중 41% (17 skill) 외부 자산 차용**. 59% (24 skill) 신규 작성. 외부 자산 활용도 *적정* — 무조건 차용 X (사용자 의도 정합).

### 6.1 Step 3 종료 — Step 4 plan 작성 진입 가능

본 매트릭스 commit 후 **Step 4 (`docs/superpowers/plans/<date>-skill-completion-plan.md` 작성)** 진입 가능.

Step 4 는:
- 41 skill 의 작성 순서 (cluster 별 / cascade 정합 / 의존 순서)
- 작성 batch 분할 (한 batch 에 5~10 skill)
- batch 별 acceptance criteria
- 외부 자산 (b) 17 건의 차용 절차 (license / attribution / 변환 패턴)
- 신규 (d) 24 건의 도메인 지식 입력 source

분량 추정: 250~350 줄.

---

## 7. 다음 액션

### 7.1 본 매트릭스 commit
### 7.2 사용자 confirm 받을 1 항목

| 결정 | 권장 |
|------|------|
| **D-G** 그룹 4-extension 의 *KESE-KIT* (한국 CII 보안) 추가 deferred 후보로 lock-in? | **(권장)** — 그룹 4-extension 3 으로 확장 (한국 시장 진출 + 보안 인증 영역) |
| **D-H** *superpowers* 외부 자산이 buddy 의 inspiration 출처일 가능성 — *명시적 attribution* 추가 검토 trigger 적용? | **(권장)** — README 의 NOTICE 패턴에 추가 (별도 commit, 본 cycle 외) |

### 7.3 Step 4 진입 — plan 작성

D-G / D-H confirm 후 **Step 4 (skill-completion-plan.md)** 진입.

Step 4 산출 = *실제 skill 작성 진입 plan*. Step 5 (skill 작성) 의 직접 입력.
