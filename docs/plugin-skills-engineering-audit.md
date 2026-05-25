# Plugin Skills — Engineering Track Audit — 2026-05-25

> **목적**: [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) §1 의 **A. 엔지니어링 프로세스 (~30 source skill)** 카테고리를 buddy 현 자산과 cross-reference 한 단일 audit SSoT. *작업물에 직접 영향을 주는 engineering 스킬* (TDD / debugging / review / planning / verification / safety / handoff / DDD / 메타스킬 등) 의 13 영역 매핑 + 흡수 상태 + 잔여 작업 우선순위.
>
> **상위 문서**: [`plugin-skills-inventory.md`](./plugin-skills-inventory.md) — plugin 트랙 전체 152 skill inventory. 본 audit 은 *engineering 영역 deep-dive*.
>
> **선행 작업**: HIGH (review-engineering anti-rationalization, `830eb04`) + MID-1 (verification-discipline, `1523e65`) + MID-2 (loop-methods, `4754e84`) + MID-3 (publish-to-tracker, `5c9fba5`) + MID-4 (finish-development-branch + git-safety-rules, `54efc0f`) — 모두 closed.

---

## 0. 한 줄 요약

A 카테고리 **13 영역** 중 **9 영역 완료** + **1 영역 부분 완료** + **3 영역 잔여 (LOW 1-3)**. 잔여 3건 중 *작업물 직접 영향* 기준 **#11 DDD/도메인 (LOW 2)** 가 최우선. **#9 위생/안전 (LOW 1)** 은 *흡수 가치 낮음* — buddy 이미 동등 cover. **#10 컨텍스트 핸드오프 (LOW 3)** 는 *cli buddy 트랙* 으로 분리 권장.

> **상태 갱신 (2026-05-26)**: **#11 LOW 2 ✅ closed** — commit `cd99818` (`audit-ubiquitous-language` 신규 skill, inspired-by DDD theory). 흡수 분류 변경 = adopt-with-edits → **inspired-by** (mattpocock 본문 cross-machine 부재로 0 read). engineering-flow H1+H3 hole 동시 closed. **현 잔여**: #8 부분완료 종결 결정 + #9 LOW 1 종결 결정 (둘 다 사용자 명시 결정 대기) + #10 LOW 3 cli 트랙 이동 결정.

---

## 1. 13 영역 매핑 매트릭스

| # | 영역 | buddy 현 자산 | source 매핑 | 상태 | 흡수 분류 |
|---|------|-------------|-----------|------|----------|
| 1 | TDD | `build-with-tdd` | superpowers `test-driven-development` + mattpocock `tdd` (vertical slice) | ✅ 완료 | inspired-by |
| 2 | 디버깅 | `diagnose-bug` + `references/loop-methods.md` | superpowers `systematic-debugging` (4-phase) + mattpocock `diagnose` (10 methods) | ✅ 완료 (MID-2) | inspired-by |
| 3 | 플래닝 | `plan-build` + `publish-to-tracker` + `define-features` + `define-feature-spec` | superpowers `writing-plans` + mattpocock `to-prd` / `to-issues` | ✅ 완료 (MID-3) | adopt-with-edits + inspired-by |
| 4 | 코드 리뷰 | `review-engineering` (HIGH anti-rationalization) + `review-architecture` + `review-scope` + `review-design` + `review-devex` | superpowers `requesting-code-review` / `receiving-code-review` + mattpocock `review` / `qa` | ✅ 완료 (HIGH) | inspired-by |
| 5 | 에이전트 협업 | `dispatch-parallel-agents` + `pair-program-loop` | superpowers `subagent-driven-development` + `dispatching-parallel-agents` | ✅ 완료 | inspired-by |
| 6 | 완료 검증 | `router/references/verification-discipline.md` (cross-skill SSoT) | superpowers `verification-before-completion` | ✅ 완료 (MID-1) | inspired-by |
| 7 | 브랜치 종료 | `finish-development-branch` (sub-orchestrator) + `git-safety-rules.md` | superpowers `finishing-a-development-branch` (참조), buddy 자체 발명 | ✅ 완료 (MID-4) | reference-only |
| 8 | 사고 확장 | `verify-best-alternative` (engineering-only, ADR-018) + `decompose-blocker` + (`concretize-idea` + `validate-idea` cascade) | superpowers `brainstorming` (9-step) + mattpocock `grill-*` | 🟡 부분 | inspired-by + buddy 변형 |
| 9 | 위생/안전 | `guard-destructive-commands` + `compose-safety-mode` + `git-safety-rules.md` + `setup-quality-gates` + `freeze-edit-scope` | superpowers `using-git-worktrees` + mattpocock `git-guardrails` / `setup-pre-commit` | 🟡 부분 — LOW 1 미진행 | (결정 보류) |
| 10 | 컨텍스트 핸드오프 | `save-context` + `restore-context` + `docs/HANDOFF.md` | mattpocock `handoff` (15줄) + `caveman` (토큰 압축) | ❌ 미진행 — LOW 3 | (트랙 분리 권장) |
| 11 | DDD/도메인 | `identify-actors` + `define-features` (actor identification) + `compose-feature-from-use-cases` | mattpocock `ubiquitous-language` (93줄) + `improve-codebase-architecture` (71줄) | ❌ 미진행 — LOW 2 | (흡수 진행 권장) |
| 12 | 부트스트랩 | `router` (auto-loaded SKILL.md) | superpowers `using-superpowers` (117줄) | ✅ 완료 | inspired-by |
| 13 | 메타-스킬 작성 | `write-a-skill` (401줄 + references 3개 492줄) | superpowers `writing-skills` (655줄) + mattpocock `write-a-skill` (117줄) + Anthropic `skill-creator` (356줄) | ✅ 완료 | adopt-with-edits + inspired-by |

---

## 2. 부분 완료 / 미진행 영역 — 상세 분석

### §2.1 #8 사고 확장 — 부분 완료 (결정 후보 (b) 권장 — 종결)

**원본 매핑** ([`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) §A.1):
- superpowers `brainstorming` (164줄, 9-step) — *idea → design pipeline*
- mattpocock `grill-with-docs` (88줄) + `grill-me` (10줄) + `zoom-out` (7줄) — *adversarial questioning*

**buddy 현 자산 매핑**:
- `verify-best-alternative` (engineering-only AI 편향 방지, ADR-018) — *결정 단계* 다관점 검토
- `decompose-blocker` (335 → 383줄, language-independent) — *코드 작업 stuck* 분해
- `concretize-idea` §1 phase orchestrator — *idea → PRD* pipeline (9-phase 안)
- `validate-idea` + `validate-advanced-edge-idea` — YC 스타일 + 압박 인터뷰 (adversarial)

**격차 분석**:

| 원본 사용 패턴 | buddy 대응 | 격차 |
|--------------|----------|------|
| `brainstorming` ad-hoc 호출 (idea → design) | `concretize-idea` + `validate-idea` cascade (§1 phase 안) | buddy 는 *phase-gated* — ad-hoc 단독 호출 불가 |
| `grill-*` adversarial questioning | `validate-advanced-edge-idea` | 동일 §1 phase 안. ad-hoc 사용 불가 |
| coding-time 사고 확장 | `decompose-blocker` + `verify-best-alternative` | buddy-side 확장 (원본에 없음) |

**결정 후보**:
- (a) Gap 인정 — *§1 phase 진입 없이* ad-hoc 사고 확장 skill 신규 추가
- **(b) Gap 무시 — buddy 의 9-phase cascade 가 동일 기능 제공, ad-hoc 호출 불필요** ← 권장
- 권장 근거: buddy 아키텍처가 *의도적으로 phase-gated*. ad-hoc brainstorming 은 사용자 자유 발화로 충분. *9-phase 정체성* 훼손 방지.

**상태 마킹**: ✅ **종결 확정 (Gap 무시, 9-phase cascade cover) — 사용자 명시 2026-05-26, Path 1 진입 의지**. [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) §3 A.1 마킹 갱신.

---

### §2.2 #9 위생/안전 — LOW 1 미진행 (결정 후보 (b) 권장 — 흡수 가치 낮음)

**원본 매핑**:
- superpowers `using-git-worktrees` (215줄)
- mattpocock `git-guardrails-claude-code` (95줄) — runtime git hook 차단
- mattpocock `setup-pre-commit` (91줄) — husky / lint-staged / Prettier 자동화
- mattpocock `setup-matt-pocock-skills` (121줄) — skill bootstrap

**buddy 현 자산 매핑**:
- `dispatch-parallel-agents/PROCEDURE.md` — worktree 패턴 내장
- `guard-destructive-commands/PROCEDURE.md` — runtime hook 차단 (rm -rf / DROP TABLE / kubectl delete / force push 등 범용)
- `compose-safety-mode/PROCEDURE.md` — multi-hook compose
- `git-safety-rules.md` (commit `54efc0f`) — design-time SSoT
- `setup-quality-gates/PROCEDURE.md` — husky / lint-staged / Prettier / typecheck / unit test / secret scan / commitlint pre-commit/pre-push

**격차 분석**:

| 원본 항목 | buddy 대응 | 흡수 가치 |
|---------|----------|---------|
| `setup-pre-commit` (구체 husky 설정) | `setup-quality-gates` 동일 영역 + 이미 *더 광범위* | **낮음** — 구체 예시 reference-only 보강만 가능 |
| `git-guardrails-claude-code` (runtime hook) | `guard-destructive-commands` 동일 영역 + *더 범용* (git 외 rm/DROP/kubectl) | **낮음** — buddy 가 우월 |
| `using-git-worktrees` (worktree 패턴) | `dispatch-parallel-agents` 내장 | **0** — 이미 cover |
| `setup-matt-pocock-skills` (skill bootstrap) | `router` SKILL.md auto-loaded | **0** — buddy 의 single-SKILL.md 정책 (ADR-006) 으로 의도적 차단 |

**결정 후보**:
- (a) LOW 1 흡수 진행 — `setup-quality-gates` 본문에 mattpocock 의 구체 husky 설정 *reference-only* 보강
- **(b) LOW 1 종결 — buddy 가 이미 동등 또는 우월 cover, 흡수 가치 낮음** ← 권장
- 권장 근거: 4 항목 모두 buddy 가 *동등 또는 우월*. 흡수 시 *문서 비대화* > *실 가치*. *reference-only* 분류는 audit trail 용 한정.

**상태 마킹**: ✅ **종결 확정 (흡수 안 함) — 사용자 명시 2026-05-26, Path 1 진입 의지**. mattpocock `setup-pre-commit` / `git-guardrails` 4 항목 모두 buddy 가 동등 또는 우월 cover. 흡수 시 *문서 비대화 > 실 가치*. SKILLS_ANALYSIS §3 A.1 LOW 1 마킹 갱신.

---

### §2.3 #10 컨텍스트 핸드오프 — LOW 3 (cli buddy 트랙 분리 권장)

**원본 매핑**:
- mattpocock `handoff` (15줄) — 짧은 인계 양식
- mattpocock `caveman` (49줄, **75% 토큰 압축**)

**buddy 현 자산 매핑**:
- `save-context` / `restore-context` — *checkpoint* 패턴 (decisions / remaining work / git status 영속, cross-branch)
- `docs/HANDOFF.md` — 문서형 cross-session 인계
- `docs/handoff/2026-05-23-skills-consolidation-handoff.md` — cross-machine handoff (24329 bytes, ~460줄)

**격차 분석**:

| 원본 항목 | buddy 대응 | 트랙 |
|---------|----------|------|
| `handoff` (15줄 짧은 양식) | `save-context` + `HANDOFF.md` — 이미 풍부 | plugin |
| `caveman` (75% 토큰 압축) | **부재** — token monitor / compaction skill 없음 | **cli buddy** (`internal/usage/` token tracking 시너지) |

**결정 후보**:
- (a) plugin 트랙에 신규 skill `compress-context` 추가 (caveman 흡수)
- **(b) plugin 트랙 종결 — caveman 은 cli buddy 트랙의 token monitor 작업 시점에 검토** ← 권장
- (c) 양쪽 모두 신규 skill

**권장 근거**: caveman 의 *75% 토큰 압축* 은 *runtime token state* 에 직결. cli buddy 의 *AI-usage coaching* (Wave 7 W7-2 Usage Analysis, ADR-013) 과 자연 매핑. plugin 트랙의 *작업 절차* 와는 분리.

**상태 마킹**: 🔄 **cli buddy 트랙 이동 확정 — 사용자 명시 2026-05-26, Path 1 진입 의지**. caveman 의 75% 토큰 압축은 *runtime token state* 영역. cli buddy Wave 7 W7-2 (Usage Analysis, ADR-013) 의 trigger 발생 시 검토. plugin engineering audit 범주에서 *제외 확정*. SKILLS_ANALYSIS §3 A.1 LOW 3 마킹 갱신.

---

### §2.4 #11 DDD/도메인 — LOW 2 (결정 후보 (a) 권장 — **본 audit 최우선 진입점**)

**원본 매핑**:
- mattpocock `ubiquitous-language` (93줄) — DDD 도메인 어휘 일관성 검사
- mattpocock `improve-codebase-architecture` (71줄) — DDD 기반 architecture 개선

**buddy 현 자산 매핑**:
- `identify-actors` — actor 식별 (user/admin/system/3rd-party/external-tool 5 분류)
- `define-features` — actor identification + actor × use case
- `compose-feature-from-use-cases` — cross-actor use case 합성
- `map-actor-use-cases` — actor 별 use case 매핑
- `review-architecture` — 시스템 구조 검토

**격차 분석**:

| 차원 | mattpocock `ubiquitous-language` | buddy 현 자산 | 격차 |
|------|--------------------------------|-------------|------|
| 분석 단위 | 도메인 모델 + *어휘* | actor (5 분류) | buddy 의 actor 는 *역할* 단위 — *도메인 어휘* 와 직교 |
| 검사 대상 | 코드 안 *식별자 일관성* (변수/함수/클래스명 vs PRD 어휘) | (부재) | **buddy 명확히 부족** |
| 출력 | 어휘 mismatch 리스트 + 리팩터 제안 | (해당 출력 없음) | **부재** |
| Trigger 시점 | refactor 직전 / PR 리뷰 / 신규 feature 추가 시 | (해당 trigger skill 없음) | **부재** |

**작업물 직접 영향 평가** (사용자 발화 우선순위 기준):
- 도메인 어휘 일관성 부재 → *코드 식별자 drift* → *신규 contributor 인지 비용 ↑* → *bug 유발 (oldName vs newName confusion)*
- 사용자 발화: "*스킬의 내용들이 작업물에 직접적인 영향을 주는 것들이야*" — **DDD/도메인 audit 이 가장 직접적**

**결정 후보**:
- **(a) LOW 2 흡수 진행 — 신규 skill `audit-ubiquitous-language` 또는 `review-domain-vocabulary`** ← 권장
- (b) LOW 2 종결 — buddy actor 모델이 *대체*, 흡수 안 함
- (c) `define-features` 본문 보강 — actor identification 단계에 어휘 일관성 체크 추가

**권장 근거**: actor 모델과 도메인 어휘는 *직교 차원* — actor 가 어휘를 *대체* 못 함. buddy 의 *작업물 영향 가치* 측면에서 명확한 gap. 흡수 분류 *adopt-with-edits* — mattpocock 의 검사 패턴 채택, buddy actor 모델 어휘로 재진술.

**작업 범위 (가정)**:
- 신규 skill `audit-ubiquitous-language/PROCEDURE.md` (예상 200-300줄)
- §3 stage skill 또는 §6 quality stage 결정 필요 (어디 phase 의 자산인가)
- mattpocock 본문 ADR-003 *adopt-with-edits* 흡수 (verbatim 0건 유지)
- `define-features` cross-link 추가 (선택)
- catalog 등재 + command 동반 (`/buddy:audit-ubiquitous-language`)
- write-a-skill RED-GREEN-REFACTOR 5-라운드 검증

---

## 3. 우선순위 — 작업물 직접 영향 기준

| Tier | 영역 | 비용 | 작업물 영향 가치 | 권장 결정 |
|------|------|------|-----------------|----------|
| **Tier 1** | #11 DDD/도메인 LOW 2 | 중 (신규 skill 1건, ~1-2 일) | **High** — refactor 결정 / 신규 contributor 인지 / bug 회피에 직결 | **흡수 진행 (a)** |
| Tier 2 | #8 사고 확장 부분완료 | 0 (결정 종결) | 0 (이미 9-phase cover) | 종결 (b) |
| Tier 2 | #9 위생/안전 LOW 1 | 0 (결정 종결) 또는 저 (reference 보강) | 낮음-중간 (이미 cover) | 종결 (b) 권장 |
| Tier 3 | #10 컨텍스트 핸드오프 LOW 3 | (cli 트랙) | 중 (token compaction) | cli buddy 트랙 이동 |

---

## 4. 광의 engineering scope (참고)

본 audit 은 SKILLS_ANALYSIS A 카테고리 (13 영역 / ~30 source skill) 의 *좁은 정의*. *작업물 직접 영향* 의 광의 해석 시 buddy §3 / §4 / §5 / §6 / §7 stage 가 모두 포함 — 약 **80+ skill**.

### 4.1 광의 engineering 후속 audit 후보

| 영역 | 스킬 수 | 현 상태 audit 필요? |
|------|--------|-------------------|
| §3 Technical Design (cascade bridges + core) | 15+19 = 34 | description 품질 / dispatch routing 정확도 audit 후보 |
| §4 Implementation Plan | 7 | 의존 그래프 정합 audit 후보 (G6 parallel-agent follow-up 과 묶음) |
| §5 Development | 10 | TDD/디버깅/리팩터 cross-skill 정합 audit 완료 |
| §6 Quality | 19 | classify-qa-tiers / audit-* 19종 의 *trigger 중복* audit 후보 |
| §7 Release & Beta | 14 | setup-* 5종 의 *책임 경계* audit 후보 |

본 광의 audit 은 *후속 wave* — 본 문서 (좁은 정의) 완료 후 진입.

---

## 5. 다음 작업 권장 (Tier 1)

### §5.1 LOW 2 진입 — `audit-ubiquitous-language` 신규 skill 작성

**4-블록 설명** ([`handoff/2026-05-23-skills-consolidation-handoff.md`](./handoff/2026-05-23-skills-consolidation-handoff.md) §7.1 패턴):

1. **무엇을 수정?** — `plugin/skills/audit-ubiquitous-language/PROCEDURE.md` 신규 작성 + `plugin/commands/audit-ubiquitous-language.md` 신규 + catalog 등재 + `define-features` cross-link 1건
2. **배경 / 이유** — buddy 의 actor 모델은 *역할* 단위, mattpocock `ubiquitous-language` 는 *어휘* 단위. 직교 차원. 코드 식별자 ↔ PRD 어휘 drift 감지 skill 부재가 *명확한 gap*. 작업물 직접 영향 가치 큼.
3. **미설정 시 문제** — (a) 신규 contributor 가 *oldName* 보고 PRD 의 *newName* 인지 못 함 → 의도 추측 → 잘못된 변경. (b) refactor 직후 코드 50%만 newName, 나머지 50%는 oldName → mixed-vocabulary codebase. (c) review 단계에서 어휘 drift 발견되어도 *체계적 fix 절차 부재* → ad-hoc grep + 누락 위험. (d) PRD 어휘가 *코드와 동기 안 됨* → 신규 feature 정의 시 stale PRD reference. (e) cross-team 협업 시 *동일 개념을 다른 어휘로* 사용 → 의사소통 비용.
4. **흡수 분류** — adopt-with-edits (mattpocock 의 *검사 패턴* 채택 + buddy actor 모델 어휘로 재진술, verbatim 0건 유지)

**선행 확인 필요**:
- 스킬 위치 결정 — §3 Technical Design / §6 Quality / cross-cutting?
- 검사 trigger — refactor 직전 / PR 리뷰 / 신규 feature 정의 시 / on-demand?
- 출력 양식 — 어휘 mismatch 리스트 + 리팩터 제안 / health score / both?

**진행 도구**: `write-a-skill` skill 사용 (RED-GREEN-REFACTOR + subagent pressure test 강제). ADR-003 *adopt-with-edits* 분류 명시.

---

## 6. SKILLS_ANALYSIS.md 갱신 후보 (Tier 1 완료 후)

본 audit 의 결정이 적용되면 [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) §3 A.1 잔여 표를 다음과 같이 갱신:

| 영역 | 변경 |
|------|------|
| #8 사고 확장 | 🟡 부분 → ✅ 종결 (Gap 무시, 9-phase cover 인정) |
| #9 위생/안전 LOW 1 | ❌ 미진행 → ✅ 종결 (buddy covered, 흡수 안 함) |
| #11 DDD/도메인 LOW 2 | ❌ 미진행 → ✅ 완료 (audit-ubiquitous-language 신설) |
| #10 컨텍스트 핸드오프 LOW 3 | ❌ 미진행 → 🔄 cli buddy 트랙으로 이동 |

→ A 카테고리 13 영역 *전체 closed* — engineering audit 종결.

---

## 7. 참조

- [`SKILLS_ANALYSIS.md`](./SKILLS_ANALYSIS.md) — 17 source repo / 167 skill 통합 분석 (§A.1 13 영역 매핑, §3 진행 상태)
- [`plugin-skills-inventory.md`](./plugin-skills-inventory.md) — plugin 트랙 152 skill inventory (본 audit 의 상위)
- [`two-tracks-charter.md`](./two-tracks-charter.md) — plugin/cli 트랙 책임 경계
- [`handoff/2026-05-23-skills-consolidation-handoff.md`](./handoff/2026-05-23-skills-consolidation-handoff.md) — 4-블록 설명 + 승인 패턴, ADR-003 4분류, A-카테고리 4-layer 시스템
- [`plugin/skills/router/references/skill-catalog.md`](../plugin/skills/router/references/skill-catalog.md) — 9-phase × Priority 카탈로그
- [`plugin/skills/router/references/verification-discipline.md`](../plugin/skills/router/references/verification-discipline.md) — 완료 발화 Iron Law (MID-1)
- [`plugin/skills/router/references/git-safety-rules.md`](../plugin/skills/router/references/git-safety-rules.md) — git 안전 SSoT (MID-4)
- [`docs/superpowers/decisions/2026-05-21-verify-best-alternative-and-bias-prevention.md`](./superpowers/decisions/2026-05-21-verify-best-alternative-and-bias-prevention.md) — ADR-018 (#8 사고 확장 buddy 변형)
- mattpocock 본문 (외부 source repo): `mattpocock-skill/skills/engineering/ubiquitous-language/SKILL.md` (LOW 2 진입 시 직접 read)

---

## 8. 본 문서 변경 정책

본 audit 은 *snapshot* (2026-05-25 baseline). 다음 시점 갱신:

| trigger | 갱신 부분 |
|---------|----------|
| Tier 1 (LOW 2) skill 작성 완료 | §1 매트릭스 #11 ✅ + §2.4 종결 + §6 SKILLS_ANALYSIS 갱신 적용 |
| Tier 2 결정 (#8, #9 종결 확정) | §1 매트릭스 + §2.1 / §2.2 종결 + §6 |
| Tier 3 cli buddy 트랙 진입 시 | §1 매트릭스 #10 cli 이동 + §6 |
| 광의 engineering audit 신설 | §4 후속 audit 권장 → 실제 audit 문서 link |

본 문서가 stale 해지면 *본 문서부터 갱신*. SKILLS_ANALYSIS.md / plugin-skills-inventory.md / skill-catalog.md 와 충돌 시 — 본 audit 이 *engineering 영역 책임*, 다른 문서는 각자 책임.
