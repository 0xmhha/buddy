# Buddy Plugin — Skill Router

> **Lazy-load 문서.** 평상시 컨텍스트에 자동 포함되지 않는다.
> [`SKILLS.md`](./SKILLS.md)의 description만 보고 skill 라우팅이 결정되는 경우에는 이 문서를 읽지 않는다 — 토큰을 아낀다.
>
> **이 문서를 읽어야 할 때 (4가지):**
> 1. `SKILLS.md`에서 후보 skill이 2개 이상이고 우선순위가 명확하지 않을 때
> 2. 동일 trigger 경로(command / hook / dispatch)에 여러 skill이 매핑되어 있을 때
> 3. Skill 호출 순서·체이닝이 필요한 워크플로우일 때
> 4. 새 skill을 추가하면서 기존 skill과의 라우팅 충돌을 검토할 때

---

## 1. 라우팅 우선순위 (general)

여러 skill이 후보일 때 적용 순서:

1. **사용자 명시 트리거 (command)** — `/buddy:<name>`은 항상 최우선. dispatch 후보를 무시.
2. **Hook auto-trigger** — Claude Code hook이 fire한 skill은 사용자 의도와 동급.
3. **Description-based dispatch** — 사용자 발화·상황과 description 매칭이 가장 강한 skill.
4. **Tie-breaker** — 2개 이상이 동등하면 §2 도메인 우선순위 표를 따른다.

---

## 2. 도메인 우선순위 표

> Skill 카테고리 간 충돌 시 어떤 카테고리가 우선하는지 정의.
> 동일 카테고리 내 충돌은 §3에서 케이스별 처리.

| Priority | Category | 대표 skill | Rationale |
|----------|----------|-----------|-----------|
| 1 | **Phase orchestrator** | `concretize-idea`, `define-features`, `design-system`, `plan-build`, `build-feature`, `verify-quality`, `ship-release`, `iterate-product`, `manage-lifecycle` | 라이프사이클 단계 시작 gate. 각 phase는 독립 진입점·결정 분기·산출물을 가진다. Tie-breaker: 진입 조건 매칭 (idea만 있으면 §1, PRD 있으면 §2+, 코드 있으면 §3+, production traffic 있으면 §8). |
| 2 | **Cross-phase review sub-orchestrator** | `autoplan` | 어느 phase의 산출물(PRD / ADR / task plan)에든 호출 가능한 4-mode review pipeline. Phase orchestrator 안의 review stage로 공유 사용. 사용자 명시 호출 시 standalone 동작. |
| 3 | **Stage skill (dual-mode)** | 각 phase 안의 stage skill | orchestrator 안에서 단계로도 호출되고, 사용자 명시 호출 시 standalone으로도 동작. 사용자가 stage 단독 명시 호출하면 orchestrator로 escalate 금지 (User Sovereignty). |
| 4 | **Domain skill** | `build-with-tdd`, `diagnose-bug`, `run-browser-qa`, `auto-create-pr`, … | 특정 단계의 ritual workflow. 진입점이 아니라 진행 중 호출. |
| 5 | **Pattern library** | `audit-live-devex`, `classify-qa-tiers`, `freeze-edit-scope`, `apply-builder-ethos`, `guard-destructive-commands`, `compose-safety-mode`, `detect-install-type`, `save-context`, `restore-context`, `persist-learning-jsonl`, `classify-review-risks`, `monitor-regressions` | 다른 skill 내부에서 ambient 적용. **plugin.json commands에 등재 금지. 직접 dispatch 금지.** |
| 6 | **Archive** | `route-intent`, `route-multi-platform`, `route-spec-to-code` | `plugin/_archive/` 격리. dispatch / command 모두 금지. |

---

## 3. 알려진 라우팅 충돌 / 케이스별 결정

> Skill 카탈로그가 자라면서 발견된 구체 충돌 사례. 각 사례는 *조건 → 선택* 형태.

### 케이스 A: idea/concept 단계 → §1 concretize-idea

- 조건: idea/concept만 존재, 코드베이스 미존재 또는 상용 빌딩 시작 전
- 후보: `concretize-idea` vs `validate-idea` / `assess-business-viability` 단독
- 선택: **`concretize-idea`** — 이유: §1 안의 stage가 8단계 (idea → business validation → competition → customer segments → PRD → review) 라 단독 호출 시 결정 분기를 빠뜨림.
- 다음 단계: PRD 확정 후 `define-features` (§2) → `design-system` (§3)으로 이관.

### 케이스 B: PRD 확정, feature 정의 필요 → §2 define-features

- 조건: PRD(또는 idea의 구체 spec)가 존재, feature backlog가 미정의
- 선택: **`define-features`** — 이유: use case 분해 → actor 식별 → system boundary → feature 합성 순서 보장.
- Q8=(a): actor / use case / system boundary 매핑이 §2 첫 단계로 강제됨.

### 케이스 C: 코드베이스 존재, 기술 설계 필요 → §3 design-system

- 조건: feature backlog 확정, infra/tech stack/API 설계 필요
- 선택: **`design-system`** — 이유: use case → infra 브릿지가 §3 첫 단계.
- 보조: `autoplan`을 technical design 산출물 review stage로 호출.

### 케이스 D: 코드 작성 단계 → §5 build-feature

- 조건: implementation plan 확정, 실제 구현 시작
- 선택: **`build-feature`** — 이유: actor track 별 병렬 개발, TDD 루프, agent dispatch 포함.
- 개선 이슈 발생 시: `verify-quality` (§6) → `iterate-fix-verify` → `build-feature` 재진입.

### 케이스 E: production 운영 중 → §8 iterate-product

- 조건: production traffic 존재, A/B 실험·분석·인시던트 대응
- 선택: **`iterate-product`** — 이유: §8은 §1과 달리 가설 검증·metric 기반 의사결정이 중심.
- 다음 루프: 분석 결과 → `generate-improvement-tasks` → §2 `define-features` 재진입.

### 케이스 F: stage skill 단독 호출 (사용자 명시)

- 조건: 사용자가 `/buddy:validate-idea`처럼 단일 stage만 명시, 또는 발화 매칭이 단일 stage에 강하게 일치
- 선택: **stage 단독** — 이유: User Sovereignty. AI는 사용자 의도를 묶을지 분리할지 결정 권한 없음.
- 행동: stage 실행 후 "이 결과를 `concretize-idea` 등 phase orchestrator의 다음 단계로 이어갈까요?"라고 *제안*하되, 실행은 사용자 승인 후.

### 케이스 G: autoplan 단독 호출

- 조건: 사용자가 `/buddy:autoplan`을 직접 호출하거나 plan/PRD/design 산출물이 이미 존재
- 선택: **`autoplan` standalone** — 이유: cross-phase review sub-orchestrator. 산출물이 존재하는 어느 phase에서든 유효.
- 입력: 이미 존재하는 rough plan / PRD / ADR / task plan.

---

## 4. 9-Phase 라이프사이클 라우팅 표

> 사용자 발화나 상황이 어느 phase에 속하는지 먼저 식별하고, 그 phase의 skill만 후보로 둔다.

| Phase | Orchestrator | 진입 조건 | Stage skills (보유) | Cross-cutting |
|-------|-------------|---------|---------------------|---------------|
| §1 Idea & Business Validation | `concretize-idea` | idea/concept만 존재 | `validate-idea`, `validate-advanced-edge-idea`, `assess-business-viability`, `review-pricing-and-gtm`, `define-product-spec` | `apply-builder-ethos`, `autoplan`(review) |
| §2 Feature Definition & Backlog | `define-features` | PRD 확정 | `identify-actors`, `map-actor-use-cases`, `map-use-case-to-system-boundary`, `compose-feature-from-use-cases`, `define-feature-spec`, `score-feature-priority`, `map-feature-dependencies`, `split-work-into-features`, `query-feature-registry`, `triage-work-items` | — |
| §3 Technical Design | `design-system` | Feature backlog 확정 | `review-architecture`, `review-engineering`, `design-artifact-storage`, `design-billing-system`, `design-claude-hooks`, `design-deploy-strategy`, `design-embedding-search`, `design-mcp-server`, `consult-codex`, `consult-design-system`, `explore-design-variants` | `autoplan`(review) |
| §4 Implementation Plan | `plan-build` | Technical design 확정 | — | `autoplan`(review) |
| §5 Development | `build-feature` | Implementation plan 확정 | `build-with-tdd`, `iterate-fix-verify`, `freeze-edit-scope`, `dispatch-parallel-agents`, `diagnose-bug`, `consult-codex` | — |
| §6 Quality | `verify-quality` | Code complete | `classify-qa-tiers`, `run-browser-qa`, `monitor-regressions`, `audit-security`, `audit-live-devex`, `measure-code-health`, `classify-review-risks`, `review-ai-safety-liability`, `review-privacy-data-risk`, `review-license-and-ip-risk`, `review-terms-policy-readiness` | — |
| §7 Release & Beta | `ship-release` | Quality gate pass | `setup-quality-gates`, `auto-create-pr`, `automate-release-tagging`, `sync-release-docs`, `write-changelog`, `guard-destructive-commands`, `compose-safety-mode` | — |
| §8 Operate & Iterate | `iterate-product` | Production traffic | `design-ab-experiment`, `analyze-ab-experiment`, `analyze-user-funnel`, `generate-improvement-tasks`, `handle-incident`, `conduct-postmortem`, `monitor-regressions`, `save-context`, `restore-context`, `summarize-retro`, `persist-learning-jsonl` | — |
| §9 Lifecycle Management | `manage-lifecycle` | Feature/product 노후화 | — | — |

---

## 5. 노출된 커맨드 목록 (plugin.json commands)

> 사용자가 `/buddy:<name>`으로 직접 호출할 수 있는 15개 커맨드.
> 이 목록 외의 skill은 orchestrator 내부에서만 invoke된다.

| 커맨드 | Priority | 용도 |
|--------|----------|------|
| `/buddy:status` | — | 현재 phase 확인 + 다음 커맨드 안내 |
| `/buddy:concretize-idea` | P1 §1 | Idea → PRD |
| `/buddy:define-features` | P1 §2 | Feature backlog |
| `/buddy:design-system` | P1 §3 | Technical design |
| `/buddy:plan-build` | P1 §4 | Implementation plan |
| `/buddy:build-feature` | P1 §5 | Development |
| `/buddy:verify-quality` | P1 §6 | Quality gate |
| `/buddy:ship-release` | P1 §7 | Release |
| `/buddy:iterate-product` | P1 §8 | Operate & iterate |
| `/buddy:manage-lifecycle` | P1 §9 | Lifecycle |
| `/buddy:autoplan` | P2 | Multi-phase plan review |
| `/buddy:diagnose-bug` | P4 | Bug investigation |
| `/buddy:audit-security` | P4 | Security review |
| `/buddy:auto-create-pr` | P4 | PR creation |
| `/buddy:build-with-tdd` | P4 | TDD standalone |

**Rule**: plugin.json commands에 새 항목을 추가하려면 Priority ≤ P4이고
이 테이블에 먼저 등재해야 한다. P5 (Pattern library) / P6 (Archive)는 영구 제외.

---

## 6. 새 skill 추가 시 router 검토 체크리스트

새 skill을 `SKILLS.md`에 등재한 직후 다음을 확인 — 위반하면 이 문서에 항목 추가:

- [ ] 동일 command 이름이 이미 등재되어 있지 않다 (`/buddy:<name>` 충돌 X)
- [ ] Description이 다른 skill의 description과 의미상 90% 이상 겹치지 않는다
- [ ] 같은 hook event(PreToolUse 등)에 매핑된 skill이 이미 있을 때, 실행 순서가 명시되었다
- [ ] 사용자 발화 패턴 1~2개로 이 skill이 dispatch되는지 mental sim 통과
- [ ] Phase 소속이 명확히 정의되었다 (§4 Phase 표에 항목 추가)

---

## 7. 참조

- Skill 카탈로그 본문 → [`SKILLS.md`](./SKILLS.md)
- Plugin manifest → [`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json)
- 9-phase 라이프사이클 아키텍처 설계 → [`docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md`](../docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md)
- 11-stage 상용 제품 빌딩 flow (참조용) → [`docs/skill-map.md`](../docs/skill-map.md)
- Plugin scaffold spec → [`docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`](../docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md)
- Archive 스킬 → [`plugin/_archive/`](./_archive/)
