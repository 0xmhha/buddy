# Plugin Skills Inventory — 2026-05-25

> **목적**: `plugin/` 트랙 (= plugin buddy, [`two-tracks-charter.md`](./two-tracks-charter.md) §2) 의 *현 자산 정합성 audit + 재정리 우선순위* 단일 SSoT. 본 문서는 *현 상태 inventory* 와 *발견된 gap* 만 담는다. 작업 진행은 본 문서를 baseline 으로 한다.
>
> **범위**: `plugin/skills/`, `plugin/commands/`, `plugin/mcp/`, `plugin/agents/`, `plugin/hooks/`, `plugin/rules/`, `plugin/.claude-plugin/`, `plugin/_archive/`.
>
> **트랙 외**: `cmd/` + `internal/` (= cli buddy 트랙) 은 본 문서 범위 아님 → 별도 inventory 가 필요한 시점에 `docs/cli-internal-inventory.md` 등으로 분리.
>
> **선행 문서**: [`two-tracks-charter.md`](./two-tracks-charter.md) (트랙 정의), [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) (17 source repo / 167 skill 통합 분석), [`plugin/skills/router/references/skill-catalog.md`](../plugin/skills/router/references/skill-catalog.md) (현 카탈로그 SSoT, ~277 줄).

---

## 0. 한 줄 요약

`plugin/skills/` **152 PROCEDURE.md + 1 router SKILL.md**. `plugin/commands/` **103** (skill-backed 100 + meta 3). 발견된 *정합성 gap* **7건** — 그 중 **doc-only fix 가능 (Tier 1) 3건** + **외부 source 흡수 결정 필요 (Tier 2-3) 4건**. `plugin/mcp/`, `plugin/agents/`, `plugin/hooks/`, `plugin/rules/` **모두 비어있음** (charter §2.4 의 4 자산 약속 중 *skill 1 자산만* 실 구현).

---

## 1. 현 자산 통계 (2026-05-25 baseline)

| 자산 | 수량 | 위치 | 비고 |
|------|------|------|------|
| Active skill (PROCEDURE.md) | **151** | `plugin/skills/<name>/PROCEDURE.md` | catalog 등재 150 + 미등재 1 (`status`) |
| Router orchestrator (SKILL.md) | **1** | `plugin/skills/router/SKILL.md` | auto-discovery 진입점, 다른 모든 스킬은 PROCEDURE.md 형식 강제 (ADR-006) |
| Slash command | **103** | `plugin/commands/<name>.md` | skill-backed 100 + meta 3 (`chain`, `parallel`, `run`) |
| Archived skill | **3** | `plugin/_archive/` | `route-intent`, `route-multi-platform`, `route-spec-to-code` (dispatch/command 모두 금지) |
| MCP server | **0** | `plugin/mcp/` (디렉토리 빈 상태) | charter §2.4 약속 — *미구현*. analytics_* / feature_* / advise / knowledge / notify / usage MCP tools 는 `internal/mcp/` (Go side) 구현, plugin 노출 layer 부재 |
| Agent | **0** | `plugin/agents/` | charter §2.4 약속 — *미구현* |
| Hook | **0** | `plugin/hooks/` | charter §2.4 약속 — *미구현* |
| Rules | **0** | `plugin/rules/` | charter §2.4 표 외 추가 — *미구현* |
| Plugin manifest | 1 | `plugin/.claude-plugin/plugin.json` | `version: "1.0.0"` (cli buddy binary version 과 별 SSoT — 아래 G4 참조) |

### 1.1 Command 동반 매트릭스

| 분류 | 수량 | 의미 |
|------|------|------|
| Skill + command 동반 | **100** | 사용자 명시 호출 (`/buddy:<name>`) + dispatch 양쪽 |
| Skill dispatch-only | **52** | description-based 자율 호출만 — 사용자 명시 호출 없음 |
| Meta command (no skill) | **3** | `chain` (cascade), `parallel` (multi-dispatch), `run` (single skill 강제 호출) |
| **합계** | **152 skill + 103 command** | catalog `__archive` 3건 제외 |

**Dispatch-only 52건** 의 *trigger description quality* 가 catalog routing 정확도의 직접 입력. 본 inventory 는 *수량만* 측정 — 각 description 의 LLM 매칭 적합성은 별도 audit 필요 (Tier 3 후보).

---

## 2. 9-Phase × Priority 분포

| Phase / 분류 | 스킬 수 | 카탈로그 §위치 |
|-------------|--------|---------------|
| **Priority 1** Phase Orchestrator | 9 | catalog §2 "Phase Orchestrators" |
| **Priority 2** Cross-phase Sub-orchestrator | 2 (`autoplan`, `finish-development-branch`) | catalog §2 "Cross-Phase Review" |
| §1 Idea & Business Validation | 13 | catalog §2 §1 |
| §2 Feature Definition & Backlog | 11 | catalog §2 §2 |
| §3 Technical Design — cascade bridges | 15 | catalog §2 §3 (first) |
| §3 Technical Design — core | 19 | catalog §2 §3 (second) |
| §4 Implementation Plan | 7 | catalog §2 §4 |
| §5 Development | 10 | catalog §2 §5 |
| §6 Quality | 19 | catalog §2 §6 |
| §7 Release & Beta | 14 | catalog §2 §7 |
| §8 Operate & Iterate | 23 | catalog §2 §8 (monitor-regressions 중복 -1) |
| §9 Lifecycle Management | 4 | catalog §2 §9 |
| Cross-cutting Utilities | 5 (apply-builder-ethos 중복 -1) | catalog §2 끝 |
| **Filesystem only (catalog 미등재)** | 1 (`status`) | — |
| **합계 (unique)** | **152** | 152 = filesystem PROCEDURE.md 수 |

> **주의**: 위 표는 *catalog 의 분류* 를 그대로 집계한 것. catalog 의 `[패턴 라이브러리]` 표기 (= Priority 5) 는 Phase 안에 섞여 있음 — 별도 컬럼 분리 시 약 12 스킬이 *Priority 5 pattern library*. 본 inventory 는 *그 분리는 catalog 갱신 시 함께 처리* 권장 (Tier 1 G2 와 묶음).

---

## 3. 발견된 정합성 Gap (7건)

### Tier 1 — Doc-only, 즉시 fix 가능 (예상 < 30 분)

> **상태 갱신 (2026-05-26)**: G1 / G2 / G3 모두 ✅ **closed**. G1 = status skill 을 catalog Cross-cutting Utilities 표에 등재 (사용자 명시 2026-05-26 Path 1 진입). G2/G3 = sibling commit `9831756` 의 catalog 중복 제거. H1 신규 skill `audit-ubiquitous-language` 추가 (commit `cd99818`). **Tier 1 전체 closed**. 다음 진입 = Tier 2 또는 신규 자산 wave (C1 MCP plugin 노출 layer).

#### G1. `status` 스킬이 catalog 미등재 — ✅ **closed (사용자 명시 2026-05-26, Path 1)**

- **위치**: `plugin/skills/status/PROCEDURE.md` 존재. `plugin/commands/status.md` 도 존재.
- **직전 상태**: catalog §2 어느 표에도 등장 안 함.
- **성격**: artifact-detection 기반 phase 추론 utility (`docs/actor-track-plan.yaml` / `docs/tech-spec.md` / `docs/prd.md` 등 탐지) — Phase 무관 cross-cutting utility.
- **종결 작업**: catalog `Cross-cutting Utilities` 표 알파벳 순 (`guide-setup-wizard` 와 `write-a-skill` 사이) 에 1 줄 등재 — `| status | command + dispatch | artifact 탐지 (...) 로 현재 lifecycle phase 추론 + 다음 권장 command 안내. phase 무관 호출 |`.

#### G2. `monitor-regressions` catalog 중복 등재

- **위치**: catalog line 166 (§6 Quality) + line 205 (§8 Operate & Iterate).
- **성격**: 동일 description, 두 phase 양쪽 등장. PROCEDURE.md 한 개만 존재.
- **권장 결정**: §8 operate 가 *production traffic* 기반 monitoring 의 자연 phase. §6 quality 에서 제거 + §8 유지. 또는 *Pattern Library* 표 신설 후 거기로 이동 (다른 패턴 라이브러리 스킬 11개 와 묶음).
- **선택지**:
  - (a) §6 라인 삭제만 (최소 변경)
  - (b) Pattern Library 표 신설 + 12 스킬 (`iterate-fix-verify`, `freeze-edit-scope`, `classify-qa-tiers`, `run-browser-qa`, `monitor-regressions`, `audit-live-devex`, `classify-review-risks`, `write-changelog`, `guard-destructive-commands`, `compose-safety-mode`, `persist-learning-jsonl`, `benchmark-llm-models`, `detect-install-type`, `guide-setup-wizard`) 재배치 — *재정리* 폭 확대.

#### G3. `apply-builder-ethos` catalog 중복 등재

- **위치**: catalog line 64 (§1 Idea & Business Validation) + line 237 (Cross-cutting Utilities).
- **성격**: 동일 description, 두 표 양쪽 등장.
- **권장 결정**: *AI collaboration project 의 ethos 주입* 은 phase 무관 cross-cutting 성격이 더 적합. §1 라인 삭제 + Cross-cutting 유지.

### Tier 2 — 외부 source 흡수 (1-2 skill 단위 결정)

#### G4. `plugin.json` version `"1.0.0"` vs binary release v0.13.0 mismatch

- **위치**: `plugin/.claude-plugin/plugin.json` 의 `version: "1.0.0"` vs `VERSION` 파일 (cli buddy binary) `0.13.0`.
- **성격**: ADR (`2026-05-11-plugin-version-reset.md`) 가 plugin version 을 *독립 SSoT* 로 결정. 즉 의도된 분리. 다만 **사용자 / 외부 reader 가 두 version 이 의미가 다르다는 것을 *어떻게 알 수 있는지*** 의 단서가 README / charter 어디에도 없음.
- **권장 결정**: `two-tracks-charter.md` §2.4 또는 §5 에 *version SSoT 가 트랙 별로 다름* 명시 1 줄 추가. 또는 plugin.json 에 `"$comment"` 필드로 ADR 참조 명시 (비표준이지만 가능).

#### G5. SKILLS_ANALYSIS.md A.1 잔여 흡수 — LOW 1-3

- **출처**: [`handoff/2026-05-23-skills-consolidation-handoff.md`](./handoff/2026-05-23-skills-consolidation-handoff.md) §6.1.
- **3 항목 (2026-05-26 갱신)**:
  - **LOW 1**: mattpocock `setup-pre-commit` / `git-guardrails` → engineering-audit §2.2 결정: **종결 권장** (buddy `compose-safety-mode` / `guard-destructive-commands` / `setup-quality-gates` / `git-safety-rules.md` 가 동등 또는 우월 cover). 사용자 명시 결정 대기.
  - **LOW 2**: mattpocock `ubiquitous-language` → ✅ **closed** (commit `cd99818`, `audit-ubiquitous-language` 신규 skill, inspired-by DDD theory). engineering-audit Tier 1 / engineering-flow H1 동시 closed.
  - **LOW 3**: mattpocock `caveman` (75% 토큰 압축) → engineering-audit §2.3 결정: **cli buddy 트랙으로 이동** (Wave 7 W7-2 Usage Analysis 시너지). plugin 트랙 종결.
- **남은 진입점**: LOW 1 종결 확정 (사용자 결정) — 본 cycle 의 doc-sync 에서 *상태 마킹 권장 사항* 으로 기록만, 흡수 작업 자체는 안 함.

#### G6. Parallel-agent 충돌 방지 follow-up — 4 skill 보강

- **출처**: handoff §6.1 신규 Follow-up.
- **4 skill 동시 보강**:
  - `decompose-track-to-tasks/PROCEDURE.md` — task yaml 에 `expected_touched_files: [path1, path2]` 필드 추가
  - `map-task-dependencies/PROCEDURE.md` — file-overlap edge 자동 추가
  - `plan-parallel-execution/PROCEDURE.md` — same-wave file-overlap=0 검증 + high-conflict zone (schema / migration / shared config) single-track 강제
  - `dispatch-parallel-agents/PROCEDURE.md` — worktree base periodic refresh
- **난이도**: 중-상 (4 skill 정합성).

### Tier 3 — 외부 source 흡수 (cluster 결정)

#### G7. SKILLS_ANALYSIS.md B-I 카테고리 — 73+ 스킬 통합 결정

- **출처**: [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) §3 의 체크리스트 (B-I 항목 모두 unchecked).
- **각 카테고리 의사결정 질문**:
  - **B. designer-skills 73 개** — 전부 유지 vs 핵심 ~30개 압축?
  - **C. 테마/브랜드 5종** — 어느 하나를 *프라이머리* 로 정할지? (ui-ux-pro-max 권장)
  - **D. 마케팅 SEO 4종** — `seo-audit` / `ai-seo` / `programmatic-seo` / `schema` 통합 vs 유지?
  - **E. obsidian** — 개인 vault (mattpocock) 와 일반 skills 둘 다 유지하나?
  - **F. doc suite (docx/pdf/pptx/xlsx)** — 변경 없음 — 확인만 (canonical Anthropic 유지보수, 자동화 스크립트 동반)
  - **G. Vercel 8종** — 얇은 2개 (`composition-patterns` 89줄, `web-design-guidelines` 39줄) 를 어디로 보낼지?
  - **H. 메타-스킬** — 단일화 vs 3-tier 유지? 현 buddy 는 `write-a-skill` (mattpocock 기반 401줄) 보유
  - **I. 도구** — `oh-my-agentic-score` (PyPI CLI) 는 별도 운영 인정?
- **결정 비용**: 각 cluster 별 *사용자 결정 1 회* + impl 1-7 일.

---

## 4. 4 자산 격차 — charter §2.4 vs 현 자산

| 자산 | charter §2.4 약속 | 현 상태 | gap |
|------|------------------|---------|-----|
| skill | 카탈로그 | ✅ 152 PROCEDURE.md | — |
| MCP | "Claude 가 외부 도구를 호출하는 표준 인터페이스" | ❌ `plugin/mcp/` 빈 디렉토리 | analytics_* / feature_* / advise / knowledge / notify / usage MCP tools 는 `internal/mcp/` (Go side) 구현, **plugin 노출 layer 부재** — Claude Code plugin install 만으로 MCP 도구 접근 불가능 |
| agent | "특정 책임을 가진 sub-agent" | ❌ `plugin/agents/` 빈 디렉토리 | Sub-agent 정의 0건 |
| hook | "Claude Code 이벤트 시점에 실행되는 자동 절차" | ❌ `plugin/hooks/` 빈 디렉토리 | M1-M6 hook reliability monitor 는 `internal/hookwrap/` (Go side, 사용자 `~/.claude/settings.json` 직접 편집), **plugin hook 정의 0건** |

**해석**: 현재 `plugin install buddy` 시 사용자는 *skill + command 만* 받는다. charter 의 4 자산 약속 중 *3 자산 (MCP/agent/hook) 미이행*. 이건 *재정리* 범주 외 — *신규 자산 구현* 작업으로 별도 wave 필요.

### 4.1 우선순위 후보

| 자산 | 진입 비용 | 예상 가치 |
|------|----------|----------|
| **MCP (plugin 노출 layer)** | 중 — `internal/mcp/` Go server 가 이미 동작, plugin 측 manifest + transport 정의만 | **High** — plugin install 한 번으로 analytics / feature / knowledge / notify / usage 도구 5+개 즉시 접근 가능 |
| **hook** | 중-상 — design-claude-hooks skill 이 이미 존재. 표준 hook 1-3 개 (예: PostToolUse 자동 ADR draft) 시작 | **Mid** — hook reliability monitor (cli buddy 트랙) 와 별 — plugin hook 은 *사용자 작업 흐름* 보조 |
| **agent** | 상 — sub-agent 정의 형식 + 책임 경계 + dispatch 패턴 결정 필요. ADR 선행 | **Mid** — code-reviewer / designer / docs-writer 등 |
| **rules** | 저 — 현재 charter §2.4 표 외. 미정의 자산 type | **Low** — 정체 불분명, 사용자 의도 확인 선행 |

---

## 5. Origin 분류 ([`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) A.1 cross-reference)

> ADR-003 의 4분류 (`verbatim` / `adopt-with-edits` / `inspired-by` / `reference-only`) 기준. **현재까지 verbatim 0건 유지** (정책 lock-in).

| Origin | 스킬 (예시) | 분류 | 비고 |
|--------|-----------|------|------|
| **buddy 자체 발명** | 9 phase orchestrator 9건 + `finish-development-branch` + `verify-best-alternative` + `decompose-blocker` + 대다수 §1-§9 stage 스킬 | reference-only | buddy 도메인 어휘 + 책임 경계 자체 정의 |
| **superpowers** | `build-with-tdd` (TDD), `diagnose-bug` (4-phase), `write-a-skill` (RED-GREEN-REFACTOR meta), `verification-discipline.md` SSoT, `decompose-blocker` (idea), `dispatch-parallel-agents` | inspired-by | superpowers 의 *철학* 채택, buddy 어휘 재진술 |
| **mattpocock** | `diagnose-bug/references/loop-methods.md` (10 methods), `publish-to-tracker` (to-prd/to-issues 변형), 잠재 LOW 1-3 흡수 후보 | adopt-with-edits / inspired-by | 형식 채택 + buddy 도메인 재진술 |
| **Anthropic 공식 (`skills/skills/`)** | — | (미흡수) | docx/pdf/pptx/xlsx canonical 유지, F 카테고리 — 흡수 결정 보류 |
| **designer-skills 73 개** | — | (미흡수) | B 카테고리 — 흡수 결정 보류 |
| **marketingskills 40 개** | `draft-marketing-copy`, `plan-marketing-channel`, `audit-seo-aso`, `automate-marketing-content` | inspired-by | 4 스킬 흡수 완료. 나머지 36개 — D 카테고리 흡수 결정 보류 |
| **vercel / supabase** | — | (미흡수) | G 카테고리 — 흡수 결정 보류 |
| **obsidian / notebooklm** | — | (미흡수) | E 카테고리 — 흡수 결정 보류 |

**해석**: 본 inventory 시점 기준 흡수율은 *전체 167 source 중 약 8-12 스킬* (mattpocock + superpowers + marketingskills 일부) — 약 5-7%. 나머지는 *buddy 자체 발명 또는 미흡수*. 흡수 결정의 비용/가치 분석은 [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) §2.

---

## 6. 재정리 우선순위 (실행 시퀀스)

### 6.1 Tier 1 — doc-only fix (한 commit, 예상 30분)

1. **G1 fix**: catalog Cross-cutting 표에 `status` 1 줄 추가
2. **G2 fix**: catalog §6 의 `monitor-regressions` 라인 삭제 (또는 Pattern Library 표 신설로 확장)
3. **G3 fix**: catalog §1 의 `apply-builder-ethos` 라인 삭제 (Cross-cutting 유지)
4. **검증**: `make test-routing` 10/10 pass + `scripts/test-router-wireup.sh` count 변동 없음 (PROCEDURE.md 수 152 유지)

### 6.2 Tier 2 — 1-2 skill 단위 외부 흡수

5. **G5 LOW 1**: mattpocock 자동화 흡수 결정 (handoff §7.1 의 4-블록 설명 + 승인 패턴)
6. **G5 LOW 2**: ubiquitous-language
7. **G5 LOW 3**: caveman
8. **G6**: parallel-agent follow-up 4 skill 보강 (중-상 난이도, *단독 cycle*)

### 6.3 Tier 3 — cluster 결정 (사용자 의사결정 선행)

9. **G7 B**: designer-skills 73개 결정 (압축 30 vs 전부 유지)
10. **G7 C-I**: 7 cluster 의사결정

### 6.4 별도 wave (재정리 외) — 신규 자산 구현

11. **§4 MCP plugin 노출 layer** — High priority, 진입 비용 중. `internal/mcp/` Go server 와 plugin manifest 연결.
12. **§4 hook 정의** — Mid priority, design-claude-hooks skill 활용.
13. **§4 agent 정의** — Mid priority, ADR 선행 필요.

---

## 7. Open Questions

| # | 질문 | trigger |
|---|------|---------|
| Q1 | catalog 의 `[패턴 라이브러리]` 표기를 별도 Priority 5 표로 분리할지? (G2 fix 와 동시에 처리 가능) | Tier 1 진입 시 결정 |
| Q2 | `plugin.json` version 의 SSoT 위치 명시 — `two-tracks-charter.md` §2.4 / §5 / 별도 ADR? | G4 fix 시점 |
| Q3 | MCP plugin 노출 layer 진입 시 `cmd/buddy-mcp/` 와의 책임 경계 — charter §5 "두 트랙 공유" 표기를 어떻게 lock-in? | §4 신규 자산 wave 진입 시 |
| Q4 | designer-skills 73개 의 *압축 후보 30개* 를 선정하는 기준 — 사용 빈도 / 카테고리 균형 / 코드 산출물 동반 비율? | G7 B 진입 시 |
| Q5 | dispatch-only 52건의 description 품질 audit — 어느 description 이 LLM routing 정확도 낮은지 측정 방법? (sample dispatch + grade) | Tier 3 후보 작업 |

---

## 8. 검증 / 적용 후 effects

본 inventory 시점 *기존 시스템 영향 0* (read-only). Tier 1-3 작업 *적용 후* 예상 변화:

| 작업 | 측정 가능 변화 |
|------|---------------|
| Tier 1 (G1-G3 fix) | catalog 유효 라인 153 → 152 (중복 -2 / `status` 추가 +1). dispatch routing 충돌 가능성 -1 (apply-builder-ethos 양쪽 후보 해소) |
| Tier 2 G5 LOW 1-3 흡수 | 152 → 153~155 skill (LOW 1 신규 + LOW 2-3 흡수 또는 기존 보강) |
| Tier 2 G6 follow-up | 152 skill 수 유지 (4 skill 보강만). 단 dispatch 시 file-overlap 충돌 발생률 측정 가능 (현 baseline 미수집) |
| Tier 3 G7 | cluster 별 차이 큼. 예: B (designer 압축 30) → 152 + 30 = 182. C (테마/브랜드 5종 → 1종 통합) → 152 + 1 = 153 |
| 신규 자산 wave (MCP) | `plugin install buddy` 만으로 5+ MCP tool 접근 가능 — charter §2.4 약속 50% 이행 (skill + MCP) |

---

## 9. 참조

- [`two-tracks-charter.md`](./two-tracks-charter.md) — plugin/cli 트랙 책임 경계 SSoT (§2 plugin buddy 정의)
- [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) — 17 source repo / 167 skill 통합 분석 + A.1 진행 상태
- [`handoff/2026-05-23-skills-consolidation-handoff.md`](./handoff/2026-05-23-skills-consolidation-handoff.md) — skill 통합 cross-machine handoff (G5/G6 출처)
- [`plugin/skills/router/references/skill-catalog.md`](../plugin/skills/router/references/skill-catalog.md) — 현 카탈로그 SSoT (~277 줄, 9-phase × Priority 표)
- [`plugin/skills/router/references/routing-rules.md`](../plugin/skills/router/references/routing-rules.md) — 라우팅 충돌 결정 SSoT
- [`plugin/skills/router/references/verification-discipline.md`](../plugin/skills/router/references/verification-discipline.md) — 완료 발화 Iron Law SSoT
- [`plugin/skills/router/references/git-safety-rules.md`](../plugin/skills/router/references/git-safety-rules.md) — 자동화 git 안전 SSoT
- [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](./superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md) — 9-phase 아키텍처 spec
- [`docs/superpowers/decisions/`](./superpowers/decisions/) — ADR 18건 (특히 ADR-006 PROCEDURE form, ADR-007 router no-cross-state, ADR-010 v1.0 scope, ADR-018 verify-best-alternative + bias)

---

## 10. 본 문서 변경 정책

본 inventory 는 *snapshot* (2026-05-25 baseline). 다음 시점에 갱신:

| trigger | 갱신 부분 |
|---------|----------|
| Tier 1 G1-G3 fix 적용 | §3 Tier 1 (✅ 마킹) + §1 카운트 (catalog 등재 150 → 152) |
| Tier 2 G5 LOW 1-3 흡수 | §3 Tier 2 + §1 카운트 + §5 Origin |
| 신규 자산 (MCP / agent / hook) 1건 이상 추가 | §1 카운트 + §4 4 자산 격차 + charter §2.4 동시 갱신 |
| catalog 구조 변경 (Pattern Library 표 신설 등) | §2 분포 표 |
| SKILLS_ANALYSIS.md 신규 항목 추가 | §3 Tier 3 + §5 Origin |

> 본 문서가 stale 해지면 **본 문서부터 갱신**. catalog / charter / SKILLS_ANALYSIS 와 충돌 시 — 본 문서가 *inventory 책임*, 다른 문서는 각자 책임 영역 SSoT (catalog = 라우팅, charter = 정체성, SKILLS_ANALYSIS = source repo 분석).
