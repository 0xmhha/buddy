# Buddy Plugin — Skill Catalog

> Plugin이 제공하는 skill 전체 카탈로그. 각 skill의 이름·트리거·용도·1줄 description.
> Claude는 이 문서만 보고도 대부분의 skill 라우팅 결정을 내릴 수 있어야 한다.
> 라우팅 결정이 모호할 때만 [`SKILL_ROUTER.md`](./SKILL_ROUTER.md)를 참조한다 (lazy-load — 토큰 절약).

---

## 1. 트리거 메커니즘

Buddy plugin의 skill은 세 경로로 활성화된다.

| 경로 | 형식 | 사용 시점 |
|------|------|----------|
| **사용자 명시 트리거 (command)** | `/buddy:<command-name> [args]` | 사용자가 의도적으로 호출 |
| **Plugin auto-trigger (hook)** | `~/.claude/settings.json`의 `hooks` 항목 | 특정 이벤트(PreToolUse 등)에 자동 |
| **Skill description-based dispatch** | `plugin/skills/<name>/SKILL.md` frontmatter `description` 매칭 | Claude가 상황 판단으로 자율 호출 |

---

## 2. Skill 목록 — 9-Phase 라이프사이클 구조

> **라우팅 우선순위**: Phase orchestrator > Stage skill > Domain skill > Pattern library > Archive
> 어느 라이프사이클 단계에 있는지 먼저 판단하고, 그 phase의 skill만 후보로 둔다.
> 라우팅 충돌이 있을 때만 [`SKILL_ROUTER.md`](./SKILL_ROUTER.md) §3 참조.

### Phase Orchestrators (Priority 1 — 라이프사이클 단계 진입점)

| Skill name | Command | When to use (1줄) |
|------------|---------|------------------|
| `concretize-idea` | `/buddy:concretize-idea` | idea/concept → PRD + business viability. idea만 있을 때 시작 |
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

### §1 Stage Skills — Idea & Business Validation

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `validate-idea` | command + dispatch | YC 스타일 아이디어 검증 인터뷰 — 6 forcing question으로 product idea를 stress-test |
| `validate-advanced-edge-idea` | command + dispatch | validate-idea 통과 후 edge case, hidden assumption, second-order effect를 압박 인터뷰(grilling)로 박멸 |
| `assess-business-viability` | command + dispatch | 아이디어가 사업으로 성립하는지 7차원(TAM/SAM/SOM, 고객-구매자, willingness-to-pay, GTM, 경쟁, unit economics, 규제)으로 평가 |
| `review-pricing-and-gtm` | dispatch | pricing model 설계와 GTM(Go-To-Market) channel 전략 평가 |
| `define-product-spec` | command + dispatch | 아이디어 검증과 사업성 검증 결과를 공식 PRD(Product Requirements Document)로 고정 |
| `apply-builder-ethos` | dispatch | Boil the Lake, Search Before Building, User Sovereignty 3 원칙을 주입해 AI collaboration project에 적용 |

### §2 Stage Skills — Feature Definition & Backlog

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `identify-actors` | dispatch | 시스템 참여 actor 열거 — user/admin/system/3rd-party/external-tool 분류 |
| `map-actor-use-cases` | dispatch | actor별 use case 식별 — UML use case 다이어그램 등가 |
| `map-use-case-to-system-boundary` | dispatch | 각 actor의 use case가 어느 시스템 경계(frontend/backend/external SaaS 등)에서 실행되는지 매핑 |
| `compose-feature-from-use-cases` | dispatch | cross-actor use case를 묶어 feature 정의 — feature = 여러 actor use case의 합성 |
| `define-feature-spec` | dispatch | feature의 완전한 명세서 작성 — actor/use case/system boundary/acceptance criteria/test plan 포함 |
| `score-feature-priority` | dispatch | RICE/ICE/MoSCoW로 feature 우선순위 결정 |
| `map-feature-dependencies` | dispatch | feature 간 선후 의존성 DAG 작성 — critical path + 병렬 실행 그룹 식별 |
| `split-work-into-features` | dispatch | PRD를 받아 vertical slice 기반 재사용 가능한 feature 단위로 분해 |
| `query-feature-registry` | dispatch | PRD 또는 feature candidate를 받아 feature-management-saas-mcp registry에서 유사 feature 검색해 reuse / adapt / inspire |
| `triage-work-items` | dispatch | 이슈/feature/task 같은 work item의 우선순위 결정과 lifecycle state machine 운영 |

### §3 Stage Skills — Technical Design

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `review-architecture` | dispatch | 시스템 구조적 무결성, 모듈 결합도, 추상화 깊이(deep module), interface depth, locality, leverage를 상위 레벨에서 검토 |
| `review-engineering` | dispatch | Engineering manager 페르소나로 implementation plan의 아키텍처·data flow·edge case·테스트 coverage·performance 리뷰 |
| `review-scope` | dispatch | Creator 페르소나로 plan의 scope를 형성·결정하는 early-stage 리뷰 |
| `review-design` | dispatch | Designer-mode plan review — 각 design dimension을 0-10으로 score하고 reverse-path technique으로 10점 만들 path를 명시 |
| `review-devex` | dispatch | 3 modes(EXPANSION/POLISH/TRIAGE)로 developer-facing product DX plan review — persona, competitor, friction map |
| `design-artifact-storage` | dispatch | patch / git_bundle / template / package 같은 immutable artifact의 저장·검증·배포 설계 |
| `design-billing-system` | dispatch | SaaS 결제 시스템 설계 — Stripe/Toss + point wallet + subscription tier + usage metering + invoice + dunning |
| `design-claude-hooks` | dispatch | Claude Code plugin/.claude scope의 PreToolUse, PostToolUse, Stop, SessionStart hook 표준 설계 |
| `design-deploy-strategy` | dispatch | production 배포 전략 설계 — canary / blue-green / rolling / recreate 선택, env 분리, secret 관리, IaC |
| `design-embedding-search` | dispatch | BM25 + vector embedding + metadata filter + reranking 결합한 hybrid search 설계 |
| `design-mcp-server` | dispatch | MCP(Model Context Protocol) server 설계 |
| `consult-codex` | command + dispatch | 독립 컨텍스트의 외부 LLM CLI(codex 등)를 호출해 review/challenge/consult 3 modes로 second opinion을 얻음 |
| `consult-design-system` | dispatch | research → synthesize → output pipeline으로 complete design system 생성 |
| `explore-design-variants` | command + dispatch | N variants를 parallel 생성하고 structured feedback으로 iterate |
| `critique-plan` | dispatch | Implementation plan에 대한 strategic critique (CEO/founder 페르소나) |

### §4 Stage Skills — Implementation Plan

> 현재 단계별 stage skill은 `plan-build` orchestrator가 직접 수행. 신규 stage skill은 추후 추가.

### §5 Stage Skills — Development

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `build-with-tdd` | command + dispatch | 신규 기능을 red-green-refactor TDD 루프(test 먼저 → 실패 확인 → 최소 구현 → 통과 → 리팩터)로 구현 |
| `iterate-fix-verify` | dispatch | [패턴 라이브러리] finding 하나씩 fix → atomic commit → re-verify 반복 repair loop |
| `freeze-edit-scope` | dispatch | [패턴 라이브러리] session 동안 Edit/Write를 single directory로 lock (Read/Grep/Glob은 열어 둠) |
| `dispatch-parallel-agents` | command + dispatch | feature/task를 worktree로 격리해 Sonnet worker agent에 병렬 분배하고 결과를 aggregate |
| `diagnose-bug` | command + dispatch | 버그를 증상 반응이 아닌 재현 가능한 원인 분석으로 해결 |

### §6 Stage Skills — Quality

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `classify-qa-tiers` | dispatch | [패턴 라이브러리] QA intensity를 Quick/Standard/Exhaustive 3 tiers로 분류 + fix→commit→re-verify loop |
| `run-browser-qa` | dispatch | [패턴 라이브러리] browser automation QA 패턴 — snapshot diff, form testing, responsive check, dialog, accessibility |
| `monitor-regressions` | dispatch | [패턴 라이브러리] delta-based threshold + transient tolerance + per-page isolation으로 monitoring + regression detect |
| `audit-security` | command + dispatch | CSO-mode security audit을 수행한다 |
| `audit-live-devex` | dispatch | [패턴 라이브러리] 빌드/배포된 live developer product를 실제로 따라 하며 TTHW timing, evidence, literal doc-following으로 DX audit |
| `measure-code-health` | command + dispatch | project tool을 auto-detect해 typecheck/lint/test/deadcode/shell 결과를 0-10 weighted composite health dashboard로 |
| `classify-review-risks` | dispatch | [패턴 라이브러리] structural code review에서 반복적으로 놓치는 risk 11 category 분류 (SQL safety, LLM trust boundary 등) |
| `review-ai-safety-liability` | dispatch | AI 기반 기능의 책임 범위, 할루시네이션 리스크, 자동 의사결정 영향, content provenance, model output safeguards 검토 |
| `review-privacy-data-risk` | dispatch | PII/personal/sensitive data lifecycle을 GDPR/PIPL/PIPA/HIPAA/COPPA 등 규제 frame으로 검토 |
| `review-license-and-ip-risk` | dispatch | 의존성/asset/AI 생성 코드의 라이선스 호환성, IP 출처, 상업 사용 가능성 검토 + risk register & remediation |
| `review-terms-policy-readiness` | dispatch | 상용 출시 전 ToS / Privacy Policy / AUP / Refund Policy / Cookie Policy / DPA 준비도 검토 |

### §7 Stage Skills — Release & Beta

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `setup-quality-gates` | command + dispatch | 개발 환경에 husky + lint-staged + Prettier + typecheck + unit test + secret scan + commitlint를 pre-commit/pre-push에 |
| `auto-create-pr` | command + dispatch | feature/task 완료 후 commit → branch push → PR 생성을 자동화 |
| `automate-release-tagging` | dispatch | merged PR set으로부터 semver auto-decision (breaking → MAJOR, feat → MINOR, fix → PATCH), git tag 생성, release note 발행 |
| `sync-release-docs` | dispatch | code change diff를 기준으로 affected docs를 audit하고 auto-update 또는 ask를 결정한다 |
| `write-changelog` | dispatch | [패턴 라이브러리] version bump + CHANGELOG release-summary format + voice rules + user-facing change summary |
| `guard-destructive-commands` | dispatch | [패턴 라이브러리] rm -rf, DROP TABLE, force push 등 destructive bash command 전 risk taxonomy + safe exception |
| `compose-safety-mode` | dispatch | [패턴 라이브러리 / META] guard-destructive-commands + freeze-edit-scope 같은 multiple safety hooks를 max safety mode로 합성 |

### §8 Stage Skills — Operate & Iterate

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `design-ab-experiment` | dispatch | A/B 실험 설계 — 가설/표본 크기/대조군/측정 지표/실험 기간을 통계적으로 설계 |
| `analyze-ab-experiment` | dispatch | 완료된 A/B 실험 결과 분석 — 통계 유의성 + 실용적 유의성으로 Ship/Revert/Continue 결정 |
| `analyze-user-funnel` | dispatch | actor별 funnel 전환/이탈 분석 — §2 use case 기반으로 어느 단계에서 drop-off 발생하는지 식별 |
| `generate-improvement-tasks` | dispatch | A/B 실험·funnel·postmortem·고객 피드백 분석 결과를 RICE 기반 improvement task로 변환 → §2 재진입 준비 |
| `handle-incident` | dispatch | 프로덕션 인시던트 대응 런북 — 심각도 분류 → 완화 → 근본 원인 → fix 배포 → 고객 커뮤니케이션 |
| `conduct-postmortem` | dispatch | 인시던트 종료 후 비난 없는 포스트모템 — 타임라인 재구성 + 5 Whys + 재발 방지 action items |
| `monitor-regressions` | dispatch | [패턴 라이브러리] delta-based threshold + transient tolerance + per-page isolation으로 monitoring + regression detect |
| `summarize-retro` | command + dispatch | git history를 evidence-based weekly retrospective로 변환 — work types, hotspots, focus score, AI collaboration |
| `save-context` | command + dispatch | decisions, remaining work, git status를 checkpoint로 저장해 future session이 branch가 달라도 이어받게 한다 |
| `restore-context` | command + dispatch | context-save가 저장한 most recent work checkpoint를 cross-branch로 load한다 |
| `persist-learning-jsonl` | dispatch | [패턴 라이브러리] JSONL append-only learning store data model + 누적/조회 패턴 (pattern/pitfall/preference taxonomy) |

### §9 Stage Skills — Lifecycle Management

> 현재 단계별 stage skill은 `manage-lifecycle` orchestrator가 직접 수행. 신규 stage skill은 추후 추가.

### Cross-cutting Utilities (Phase 소속 없음)

| Skill name | Trigger | When to use (1줄) |
|------------|---------|------------------|
| `apply-builder-ethos` | dispatch | Boil the Lake, Search Before Building, User Sovereignty 3 원칙을 주입해 AI collaboration project에 적용 |
| `benchmark-llm-models` | dispatch | [패턴 라이브러리] multi-provider LLM benchmark 패턴 (Claude/GPT/Gemini) — dry-run auth verify, provider select, comparison |
| `detect-install-type` | dispatch | [패턴 라이브러리] tool install type(global-git/local-git/vendored/package-manager/dev-symlink) detect + upgrade path |
| `guide-setup-wizard` | dispatch | [패턴 라이브러리] auto-detect → picker → verify pattern으로 credential/config setup flow 설계 |

---

## 3. 추가 / 수정 규칙

새 skill을 카탈로그에 등재할 때:

1. `plugin/skills/<name>/SKILL.md` 생성 (frontmatter `name`, `description` 필수).
2. 위 §2 표에 한 줄 추가 — `name` / phase 섹션 / `trigger` / `when to use(1줄)`.
3. **라우팅이 다른 skill과 겹치거나 우선순위가 필요한 경우에만** `SKILL_ROUTER.md`에 항목 추가.
4. command 트리거를 추가하면 `plugin/commands/buddy/<name>.md` 도 함께 등재.
5. 속하는 phase를 명시 (`SKILL_ROUTER.md` §4 표 업데이트).

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

- 라우팅 결정이 모호하거나 skill 간 충돌이 있을 때 → [`SKILL_ROUTER.md`](./SKILL_ROUTER.md)
- 9-phase 라이프사이클 아키텍처 설계 → [`docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md`](../docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md)
- Plugin manifest → [`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json)
- Archive 스킬 → [`_archive/`](./_archive/)
