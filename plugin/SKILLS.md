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

세 경로 모두 동일한 skill 본문(`plugin/skills/<name>/SKILL.md`)을 실행한다.

---

## 2. Skill 목록

> 각 항목은 다른 세션에서 점증 추가 중. 본 카탈로그는 **항상 참조 가능**해야 하므로 한 항목당 한 줄을 넘지 않게 유지한다.
>
> 한 줄을 넘는 라우팅 정보(상황별 우선순위, 충돌 해소, 의사결정 트리)는 `SKILL_ROUTER.md`로 분리한다.

| Skill name | Path | Trigger 경로 | When to use (1줄) |
|------------|------|------------|------------------|
| `apply-builder-ethos` | `plugin/skills/apply-builder-ethos/SKILL.md` | dispatch | Boil the Lake, Search Before Building, User Sovereignty 3 원칙을 주입해 AI collaboration project에 적용 |
| `assess-business-viability` | `plugin/skills/assess-business-viability/SKILL.md` | dispatch | 아이디어가 사업으로 성립하는지 7차원(TAM/SAM/SOM, 고객-구매자, willingness-to-pay, GTM, 경쟁, unit economics, 규제)으로 평가 |
| `audit-live-devex` | `plugin/skills/audit-live-devex/SKILL.md` | dispatch | [패턴 라이브러리] 빌드/배포된 live developer product를 실제로 따라 하며 TTHW timing, evidence, literal doc-following으로 DX audit |
| `audit-security` | `plugin/skills/audit-security/SKILL.md` | dispatch | CSO-mode security audit을 수행한다 |
| `auto-create-pr` | `plugin/skills/auto-create-pr/SKILL.md` | dispatch | feature/task 완료 후 commit → branch push → PR 생성을 자동화 |
| `automate-release-tagging` | `plugin/skills/automate-release-tagging/SKILL.md` | dispatch | merged PR set으로부터 semver auto-decision (breaking → MAJOR, feat → MINOR, fix → PATCH), git tag 생성, release note 발행 |
| `autoplan` | `plugin/skills/autoplan/SKILL.md` | dispatch | 자동 multi-stage 리뷰 파이프라인 |
| `benchmark-llm-models` | `plugin/skills/benchmark-llm-models/SKILL.md` | dispatch | [패턴 라이브러리] multi-provider LLM benchmark 패턴 (Claude/GPT/Gemini) — dry-run auth verify, provider select, comparison |
| `build-with-tdd` | `plugin/skills/build-with-tdd/SKILL.md` | dispatch | 신규 기능을 red-green-refactor TDD 루프(test 먼저 → 실패 확인 → 최소 구현 → 통과 → 리팩터)로 구현 |
| `classify-qa-tiers` | `plugin/skills/classify-qa-tiers/SKILL.md` | dispatch | [패턴 라이브러리] QA intensity를 Quick/Standard/Exhaustive 3 tiers로 분류 + fix→commit→re-verify loop |
| `classify-review-risks` | `plugin/skills/classify-review-risks/SKILL.md` | dispatch | [패턴 라이브러리] structural code review에서 반복적으로 놓치는 risk 11 category 분류 (SQL safety, LLM trust boundary 등) |
| `compose-safety-mode` | `plugin/skills/compose-safety-mode/SKILL.md` | dispatch | [패턴 라이브러리 / META] guard-destructive-commands + freeze-edit-scope 같은 multiple safety hooks를 max safety mode로 합성 |
| `consult-codex` | `plugin/skills/consult-codex/SKILL.md` | dispatch | 독립 컨텍스트의 외부 LLM CLI(codex 등)를 호출해 review/challenge/consult 3 modes로 second opinion을 얻음 |
| `consult-design-system` | `plugin/skills/consult-design-system/SKILL.md` | dispatch | research → synthesize → output pipeline으로 complete design system 생성 |
| `critique-plan` | `plugin/skills/critique-plan/SKILL.md` | dispatch | Implementation plan에 대한 strategic critique (CEO/founder 페르소나) |
| `define-product-spec` | `plugin/skills/define-product-spec/SKILL.md` | dispatch | 아이디어 검증과 사업성 검증 결과를 공식 PRD(Product Requirements Document)로 고정 |
| `design-artifact-storage` | `plugin/skills/design-artifact-storage/SKILL.md` | dispatch | patch / git_bundle / template / package 같은 immutable artifact의 저장·검증·배포 설계 |
| `design-billing-system` | `plugin/skills/design-billing-system/SKILL.md` | dispatch | SaaS 결제 시스템 설계 — Stripe/Toss + point wallet + subscription tier + usage metering + invoice + dunning |
| `design-claude-hooks` | `plugin/skills/design-claude-hooks/SKILL.md` | dispatch | Claude Code plugin/.claude scope의 PreToolUse, PostToolUse, Stop, SessionStart hook 표준 설계 |
| `design-deploy-strategy` | `plugin/skills/design-deploy-strategy/SKILL.md` | dispatch | production 배포 전략 설계 — canary / blue-green / rolling / recreate 선택, env 분리, secret 관리, IaC |
| `design-embedding-search` | `plugin/skills/design-embedding-search/SKILL.md` | dispatch | BM25 + vector embedding + metadata filter + reranking 결합한 hybrid search 설계 |
| `design-mcp-server` | `plugin/skills/design-mcp-server/SKILL.md` | dispatch | MCP(Model Context Protocol) server 설계 |
| `detect-install-type` | `plugin/skills/detect-install-type/SKILL.md` | dispatch | [패턴 라이브러리] tool install type(global-git/local-git/vendored/package-manager/dev-symlink) detect + upgrade path |
| `diagnose-bug` | `plugin/skills/diagnose-bug/SKILL.md` | dispatch | 버그를 증상 반응이 아닌 재현 가능한 원인 분석으로 해결 |
| `dispatch-parallel-agents` | `plugin/skills/dispatch-parallel-agents/SKILL.md` | dispatch | feature/task를 worktree로 격리해 Sonnet worker agent에 병렬 분배하고 결과를 aggregate |
| `explore-design-variants` | `plugin/skills/explore-design-variants/SKILL.md` | dispatch | N variants를 parallel 생성하고 structured feedback으로 iterate |
| `freeze-edit-scope` | `plugin/skills/freeze-edit-scope/SKILL.md` | dispatch | [패턴 라이브러리] session 동안 Edit/Write를 single directory로 lock (Read/Grep/Glob은 열어 둠) |
| `guard-destructive-commands` | `plugin/skills/guard-destructive-commands/SKILL.md` | dispatch | [패턴 라이브러리] rm -rf, DROP TABLE, force push 등 destructive bash command 전 risk taxonomy + safe exception |
| `guide-setup-wizard` | `plugin/skills/guide-setup-wizard/SKILL.md` | dispatch | [패턴 라이브러리] auto-detect → picker → verify pattern으로 credential/config setup flow 설계 |
| `iterate-fix-verify` | `plugin/skills/iterate-fix-verify/SKILL.md` | dispatch | [패턴 라이브러리] finding 하나씩 fix → atomic commit → re-verify 반복 repair loop |
| `measure-code-health` | `plugin/skills/measure-code-health/SKILL.md` | dispatch | project tool을 auto-detect해 typecheck/lint/test/deadcode/shell 결과를 0-10 weighted composite health dashboard로 |
| `monitor-regressions` | `plugin/skills/monitor-regressions/SKILL.md` | dispatch | [패턴 라이브러리] delta-based threshold + transient tolerance + per-page isolation으로 monitoring + regression detect |
| `persist-learning-jsonl` | `plugin/skills/persist-learning-jsonl/SKILL.md` | dispatch | [패턴 라이브러리] JSONL append-only learning store data model + 누적/조회 패턴 (pattern/pitfall/preference taxonomy) |
| `query-feature-registry` | `plugin/skills/query-feature-registry/SKILL.md` | dispatch | PRD 또는 feature candidate를 받아 feature-management-saas-mcp registry에서 유사 feature 검색해 reuse / adapt / inspire |
| `restore-context` | `plugin/skills/restore-context/SKILL.md` | dispatch | context-save가 저장한 most recent work checkpoint를 cross-branch로 load한다 |
| `review-ai-safety-liability` | `plugin/skills/review-ai-safety-liability/SKILL.md` | dispatch | AI 기반 기능의 책임 범위, 할루시네이션 리스크, 자동 의사결정 영향, content provenance, model output safeguards 검토 |
| `review-architecture` | `plugin/skills/review-architecture/SKILL.md` | dispatch | 시스템 구조적 무결성, 모듈 결합도, 추상화 깊이(deep module), interface depth, locality, leverage를 상위 레벨에서 검토 |
| `review-design` | `plugin/skills/review-design/SKILL.md` | dispatch | Designer-mode plan review — 각 design dimension을 0-10으로 score하고 reverse-path technique으로 10점 만들 path를 명시 |
| `review-devex` | `plugin/skills/review-devex/SKILL.md` | dispatch | 3 modes(EXPANSION/POLISH/TRIAGE)로 developer-facing product DX plan review — persona, competitor, friction map |
| `review-engineering` | `plugin/skills/review-engineering/SKILL.md` | dispatch | Engineering manager 페르소나로 implementation plan의 아키텍처·data flow·edge case·테스트 coverage·performance 리뷰 |
| `review-license-and-ip-risk` | `plugin/skills/review-license-and-ip-risk/SKILL.md` | dispatch | 의존성/asset/AI 생성 코드의 라이선스 호환성, IP 출처, 상업 사용 가능성 검토 + risk register & remediation |
| `review-pricing-and-gtm` | `plugin/skills/review-pricing-and-gtm/SKILL.md` | dispatch | pricing model 설계와 GTM(Go-To-Market) channel 전략 평가 |
| `review-privacy-data-risk` | `plugin/skills/review-privacy-data-risk/SKILL.md` | dispatch | PII/personal/sensitive data lifecycle을 GDPR/PIPL/PIPA/HIPAA/COPPA 등 규제 frame으로 검토 |
| `review-scope` | `plugin/skills/review-scope/SKILL.md` | dispatch | Creator 페르소나로 plan의 scope를 형성·결정하는 early-stage 리뷰 |
| `review-terms-policy-readiness` | `plugin/skills/review-terms-policy-readiness/SKILL.md` | dispatch | 상용 출시 전 ToS / Privacy Policy / AUP / Refund Policy / Cookie Policy / DPA 준비도 검토 |
| `route-intent` | `plugin/skills/route-intent/SKILL.md` | archive | [ARCHIVE / 참조 전용] LLM이 직접 invoke 금지 |
| `route-multi-platform` | `plugin/skills/route-multi-platform/SKILL.md` | archive | [ARCHIVE / 참조 전용] LLM이 직접 invoke 금지 |
| `route-spec-to-code` | `plugin/skills/route-spec-to-code/SKILL.md` | archive | [ARCHIVE / 참조 전용] LLM이 직접 invoke 금지 |
| `run-browser-qa` | `plugin/skills/run-browser-qa/SKILL.md` | dispatch | [패턴 라이브러리] browser automation QA 패턴 — snapshot diff, form testing, responsive check, dialog, accessibility |
| `save-context` | `plugin/skills/save-context/SKILL.md` | dispatch | decisions, remaining work, git status를 checkpoint로 저장해 future session이 branch가 달라도 이어받게 한다 |
| `setup-quality-gates` | `plugin/skills/setup-quality-gates/SKILL.md` | dispatch | 개발 환경에 husky + lint-staged + Prettier + typecheck + unit test + secret scan + commitlint를 pre-commit/pre-push에 |
| `split-work-into-features` | `plugin/skills/split-work-into-features/SKILL.md` | dispatch | PRD를 받아 vertical slice 기반 재사용 가능한 feature 단위로 분해 |
| `summarize-retro` | `plugin/skills/summarize-retro/SKILL.md` | dispatch | git history를 evidence-based weekly retrospective로 변환 — work types, hotspots, focus score, AI collaboration |
| `sync-release-docs` | `plugin/skills/sync-release-docs/SKILL.md` | dispatch | code change diff를 기준으로 affected docs를 audit하고 auto-update 또는 ask를 결정한다 |
| `triage-work-items` | `plugin/skills/triage-work-items/SKILL.md` | dispatch | 이슈/feature/task 같은 work item의 우선순위 결정과 lifecycle state machine 운영 |
| `validate-advanced-edge-idea` | `plugin/skills/validate-advanced-edge-idea/SKILL.md` | dispatch | validate-idea 통과 후 edge case, hidden assumption, second-order effect를 압박 인터뷰(grilling)로 박멸 |
| `validate-idea` | `plugin/skills/validate-idea/SKILL.md` | dispatch | YC 스타일 아이디어 검증 인터뷰 — 6 forcing question으로 product idea를 stress-test |
| `write-changelog` | `plugin/skills/write-changelog/SKILL.md` | dispatch | [패턴 라이브러리] version bump + CHANGELOG release-summary format + voice rules + user-facing change summary |

---

## 3. 추가 / 수정 규칙

새 skill을 카탈로그에 등재할 때:

1. `plugin/skills/<name>/SKILL.md` 생성 (frontmatter `name`, `description` 필수).
2. 위 §2 표에 한 줄 추가 — `name` / `path` / `trigger` / `when to use(1줄)`.
3. **라우팅이 다른 skill과 겹치거나 우선순위가 필요한 경우에만** `SKILL_ROUTER.md`에 항목 추가.
4. command 트리거를 추가하면 `plugin/commands/buddy/<name>.md` 도 함께 등재.

---

## 4. 참조

- 라우팅 결정이 모호하거나 skill 간 충돌이 있을 때 → [`SKILL_ROUTER.md`](./SKILL_ROUTER.md)
- Plugin scaffold 전체 구조 → [`docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`](../docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md)
- Plugin manifest → [`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json)
