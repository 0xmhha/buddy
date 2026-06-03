# Buddy Plugin — Skill Catalog (Index)

> Plugin이 제공하는 skill 카탈로그의 **인덱스**. orchestrator·cross-phase·cross-cutting skill + 노드별 shard 맵.
> **노드별 stage skill 목록은 [`catalog/phase-N-*.md`](./catalog/) shard로 분리**되어 있다 (lazy-load — 노드가 정해진 뒤 해당 shard만 로드 → 토큰 절약).
> Claude는 이 인덱스 + 결정된 노드의 shard 1개로 대부분의 라우팅 결정을 내린다. 모호할 때만 [`routing-rules.md`](./routing-rules.md) §3 참조.

---

## 1. 트리거 메커니즘

Buddy plugin의 skill은 세 경로로 활성화된다.

| 경로 | 형식 | 사용 시점 |
|------|------|----------|
| **사용자 명시 트리거 (command)** | `/buddy:<command-name> [args]` | 사용자가 의도적으로 호출 |
| **Plugin auto-trigger (hook)** | `~/.claude/settings.json`의 `hooks` 항목 | 특정 이벤트(PreToolUse 등)에 자동 |
| **Router 내부 dispatch** | shard의 entry (각 행) — router가 사용자 발화·상황과 매칭하여 dispatch | start/evaluate-skill가 진입 후 후속 노드로 진행할 때 |

> **중요 — frontmatter 부재**: buddy의 PROCEDURE.md는 frontmatter 없음 (router 경유 lazy-load — `docs/plugin-skills-authoring-guide.md` §0.1 SSoT). 본 카탈로그(인덱스 + shard)의 entry가 PROCEDURE.md frontmatter description의 역할을 대신한다.

---

## 2. 노드 shard 맵 (stage skill 목록 위치)

> 라우팅 우선순위: Phase orchestrator > Stage skill > Domain skill > Pattern library > Archive (상세 SSoT: [`routing-rules.md`](./routing-rules.md) §2).
> 노드(§N)가 정해지면 아래 shard 1개만 lazy-load한다. 각 shard는 노드 컨텍스트(DoR→DoD) + stage skill 표(`When to use` + `Not when` anti-trigger) + disambiguation을 담는다.

| 노드 | 영문 표준 용어 | stage skill shard |
|------|--------------|------------------|
| §1 | Discovery / Impact Analysis | [`catalog/phase-1-discovery-impact.md`](./catalog/phase-1-discovery-impact.md) |
| §2 | Requirements Specification | [`catalog/phase-2-requirements.md`](./catalog/phase-2-requirements.md) |
| §3 | Software Design | [`catalog/phase-3-design.md`](./catalog/phase-3-design.md) |
| §4 | Iteration Planning | [`catalog/phase-4-planning.md`](./catalog/phase-4-planning.md) |
| §5 | Construction | [`catalog/phase-5-construction.md`](./catalog/phase-5-construction.md) |
| §6 | Verification & Validation | [`catalog/phase-6-verification.md`](./catalog/phase-6-verification.md) |
| §7 | Release & Deployment | [`catalog/phase-7-release.md`](./catalog/phase-7-release.md) |
| §8 | Operation & Maintenance | [`catalog/phase-8-operation.md`](./catalog/phase-8-operation.md) |
| §9 | Retirement / Decommissioning | [`catalog/phase-9-retirement.md`](./catalog/phase-9-retirement.md) |

> 노드 정체성(DoR/DoD·전이 규칙) 원본 → [`engineering-phases.md`](./engineering-phases.md). 노드 명사 원본 → [`se-lifecycle-naming.md`](./se-lifecycle-naming.md).

### Phase Orchestrators (Priority 1 — 라이프사이클 노드 진입점)

| Skill name | Command | When to use (1줄) |
|------------|---------|------------------|
| `concretize-idea` | (직접 호출 불가, `/buddy:start` 경유) | idea/concept → PRD + HLD + review. 신규 프로덕트(greenfield)일 때 — `/buddy:start`로 진입 → 라우팅 |
| `assess-product-change` | (직접 호출 불가, `/buddy:start` 경유) | 기존 프로덕트 변경 → 영향 평가 + scope 분류 + 다음 phase routing — `/buddy:start`로 진입 → 라우팅 |
| `define-features` | `/buddy:define-features` | PRD → feature backlog (actor/use case/system boundary 포함) |
| `design-system` | `/buddy:design-system` | feature backlog → tech stack ADR + infra + API + data model |
| `plan-build` | `/buddy:plan-build` | technical design → actor별 task graph + parallel execution plan |
| `build-feature` | `/buddy:build-feature` | implementation plan → working code + tests (TDD + parallel agents) |
| `verify-quality` | `/buddy:verify-quality` | code complete → QA report + security + compliance sign-off |
| `ship-release` | `/buddy:ship-release` | quality gate pass → tagged release + UAT + GA |
| `iterate-product` | `/buddy:iterate-product` | production traffic → A/B 실험 + funnel 분석 + improvement backlog |
| `manage-lifecycle` | `/buddy:manage-lifecycle` | feature/product 노후화 → deprecation + migration + EOL |

### Cross-Phase Review Sub-Orchestrator (Priority 2)

| Skill name | Command | When to use (1줄) |
|------------|---------|------------------|
| `autoplan` | `/buddy:autoplan` | 기존 plan/PRD/ADR/task plan을 4-mode review (review-scope/engineering/design/devex 순차) |
| `finish-development-branch` | `/buddy:finish-development-branch` | §5 build-feature 후 PR 생성까지 5-stage sub-orchestrator (pre-flight sync + quality-gate + changelog + docs-sync + PR + mergeable verify). git 안전 정책 적용. Iron Law mergeable=CLEAN 검증. |

### Cross-cutting Utilities (노드 소속 없음 — 어느 노드에서든 호출)

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `apply-builder-ethos` | dispatch | Boil the Lake, Search Before Building, User Sovereignty 3 원칙 주입 |
| `benchmark-llm-models` | dispatch [패턴 라이브러리] | multi-provider LLM benchmark (Claude/GPT/Gemini) — auth verify, select, comparison |
| `consult-codex` | command + dispatch | 외부 LLM CLI(codex 등)로 review/challenge/consult 3 모드 second opinion. 어느 노드에서든 결정 검토에 호출 (특히 §3 설계·§5 구현) |
| `decompose-blocker` | command + dispatch | [엔지니어링·언어독립] 코드 작업 stuck 시 문제 분해 + 비용-정보 매트릭스. **자동 trigger: 동일 문제 3회 시도 후 미해결** 또는 사용자 명시. *fix 수행 X*, 다음 스킬로 dispatch 준비 |
| `detect-install-type` | dispatch [패턴 라이브러리] | tool install type(global-git/local-git/vendored/package-manager/dev-symlink) detect + upgrade path |
| `guide-setup-wizard` | dispatch [패턴 라이브러리] | auto-detect → picker → verify로 credential/config setup flow 설계 |
| `start` | command + dispatch | 사용자 의도 기반 진입 라우터 — 자연어 발화 + 3 질문(경로/유형/상업성) → 경로 유무로 dispatch (있음→assess-product-change, 없음→concretize-idea). command 이름 모를 때 첫 entry |
| `evaluate-skill` | command + dispatch | PROCEDURE.md/SKILL.md/command를 24-항목 체크리스트(authoring-guide §4)로 평가 — 가중 점수 + actionable 개선. paired 자동 확장 |
| `status` | command + dispatch | artifact 탐지(docs/prd.md / docs/feature-spec/ 등)로 현재 lifecycle 노드 추론 + 다음 권장 command 안내. 노드 무관 |
| `write-a-skill` | command + dispatch | [META] 신규 buddy 스킬을 PROCEDURE.md + catalog 등재 + 차용 4분류 + RED-GREEN-REFACTOR pressure test로 작성 |
| `save-context` | command + dispatch | decisions·remaining work·git status를 checkpoint로 저장 (cross-branch 이어받기). 노드 무관 |
| `restore-context` | command + dispatch | save-context의 most recent checkpoint를 cross-branch로 load. 노드 무관 |
| `persist-learning-jsonl` | dispatch [패턴 라이브러리] | JSONL append-only learning store + 누적/조회 패턴. 노드 무관 |
| `review-legal-regulatory` | command + dispatch | region-agnostic 법률/규제 frame (privacy/IP/AI/약관/결제/산업/audit) + region cluster trigger. §1/§7 등 복수 노드 |

---

## 3. 추가 / 수정 규칙

새 skill을 카탈로그에 등재할 때:

1. `plugin/skills/<name>/PROCEDURE.md` 생성 (frontmatter 미사용 — dispatch description은 shard 표의 "When to use" + "Not when" 컬럼에 작성).
2. **stage skill** → 해당 노드 shard [`catalog/phase-N-*.md`](./catalog/)의 표에 한 줄 추가 (`name` / `trigger` / `When to use` / `Not when` anti-trigger). **orchestrator·cross-phase·cross-cutting** → 본 인덱스 §2 표에 추가.
3. 중복 검사 grep 대상은 **인덱스 + shard 전체**: `grep -riE "<keyword>" plugin/skills/router/references/skill-catalog.md plugin/skills/router/references/catalog/`. 1개라도 hit이면 차별점 명시 + `Not when` 으로 경계.
4. **라우팅이 다른 skill과 겹치거나 우선순위가 필요한 경우에만** `routing-rules.md`에 항목 추가 (노드 *내* 모호는 shard의 `Not when`/disambiguation으로, 노드 *간* 모호만 routing-rules).
5. command 트리거 추가 시 `plugin/commands/<name>.md` 도 등재 (subdir 금지 — 단일 파일, `test-router-wireup.sh` Check 5 enforcement). 노드 소속은 `engineering-phases.md` + 해당 shard에 반영.

---

## 4. Archive

> `plugin/_archive/`로 격리. dispatch / command 모두 금지. 참조 전용.

| Skill name | 이유 |
|------------|------|
| `route-intent` | multi-orchestrator 모델 채택으로 역할 소멸 |
| `route-multi-platform` | multi-orchestrator 모델 채택으로 역할 소멸 |
| `route-spec-to-code` | `build-feature` (§5 orchestrator)로 기능 흡수 |

---

## 5. 참조

- 노드별 stage skill 목록 → [`catalog/`](./catalog/) shard 9개
- 라우팅 결정이 모호하거나 skill 간 충돌이 있을 때 → [`routing-rules.md`](./routing-rules.md)
- **완료 발화 직전 evidence 게이트 (cross-skill SSoT)** → [`verification-discipline.md`](./verification-discipline.md). build-with-tdd / iterate-fix-verify / verify-quality / build-feature / review-engineering / diagnose-bug / agent dispatch 후 모두 lazy-load.
- **자동화 git 안전 원칙 (cross-skill SSoT)** → [`git-safety-rules.md`](./git-safety-rules.md). force / rewrite-pushed / 자동 복구 시도 금지.
- 노드 정체성·산출물·전이 규칙 → [`engineering-phases.md`](./engineering-phases.md) | 노드 명사 → [`se-lifecycle-naming.md`](./se-lifecycle-naming.md)
- 9-phase 라이프사이클 아키텍처 설계 → [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../../../../docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md)
- Plugin manifest → [`.claude-plugin/plugin.json`](../../../.claude-plugin/plugin.json)
- Archive 스킬 → [`_archive/`](../../../_archive/)
