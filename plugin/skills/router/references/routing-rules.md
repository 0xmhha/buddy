# Buddy Plugin — Skill Router

> **Lazy-load 문서.** 평상시 컨텍스트에 자동 포함되지 않는다.
> [`skill-catalog.md`](./skill-catalog.md)의 description만 보고 skill 라우팅이 결정되는 경우에는 이 문서를 읽지 않는다 — 토큰을 아낀다.
>
> **이 문서를 읽어야 할 때 (4가지):**
> 1. `skill-catalog.md`에서 후보 skill이 2개 이상이고 우선순위가 명확하지 않을 때
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

### 케이스 A1: §1 Mode A — idea/concept (Greenfield) → `concretize-idea`

- 조건: idea/concept만 존재, **코드베이스 미존재** (start Step 2.2에서 경로 '없음' 또는 path-missing 확인)
- 후보: `concretize-idea` vs `validate-idea` / `assess-business-viability` 단독
- 선택: **`concretize-idea`** — 이유: §1 안의 stage가 9단계 (idea → validation → customer → competition → business → pricing → PRD → HLD → review) 라 단독 호출 시 결정 분기를 빠뜨림.
- 다음 단계: PRD 확정 후 `define-features` (§2) → `design-system` (§3)으로 이관.

### 케이스 A2: §1 Mode B — 기존 프로덕트 변경 → `assess-product-change`

- 조건: **기존 코드베이스 존재** (start Step 2.2에서 경로 EXISTS 확인). 변경 유형 무관 (버그/기능/리팩토링/의존성 갱신)
- 후보: `assess-product-change` 단일 (Mode B orchestrator)
- 선택: **`assess-product-change`** — 영향 평가 + scope 분류(small/medium/large) + 다음 phase routing
- 다음 단계 (scope별 routing):
  - Small → §5 `build-feature` 직접 진입 (버그 수정, 설정 변경)
  - Medium → §3 `design-system` (설계 검토 후 구현)
  - Large → §2 `define-features` (feature 정의부터)
  - Defer/Reject → backlog 기록, cycle 종료

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
| §1 Idea & Business Validation (Mode A) | `concretize-idea` | idea/concept만 존재 (greenfield) | `validate-idea`, `validate-advanced-edge-idea`, `assess-business-viability`, `review-pricing-and-gtm`, `define-product-spec` | `apply-builder-ethos`, `autoplan`(review) |
| §1 Problem/Change Assessment (Mode B) | `assess-product-change` | 기존 프로덕트에 변경 필요 | — (scope 평가 후 §2/§3/§5로 routing) | — |
| §2 Feature Definition & Backlog | `define-features` | PRD 확정 | `identify-actors`, `map-actor-use-cases`, `map-use-case-to-system-boundary`, `compose-feature-from-use-cases`, `define-feature-spec`, `score-feature-priority`, `map-feature-dependencies`, `split-work-into-features`, `query-feature-registry`, `triage-work-items` | — |
| §3 Technical Design | `design-system` | Feature backlog 확정 | `review-architecture`, `review-engineering`, `design-artifact-storage`, `design-billing-system`, `design-claude-hooks`, `design-deploy-strategy`, `design-embedding-search`, `design-mcp-server`, `consult-codex`, `consult-design-system`, `verify-best-alternative` | `autoplan`(review) |
| §4 Implementation Plan | `plan-build` | Technical design 확정 | — | `autoplan`(review) |
| §5 Development | `build-feature` | Implementation plan 확정 | `build-with-tdd`, `iterate-fix-verify`, `freeze-edit-scope`, `dispatch-parallel-agents`, `diagnose-bug`, `consult-codex` | — |
| §6 Quality | `verify-quality` | Code complete | `classify-qa-tiers`, `run-browser-qa`, `monitor-regressions`, `audit-security`, `audit-live-devex`, `measure-code-health`, `classify-review-risks`, `review-ai-safety-liability`, `review-privacy-data-risk`, `review-license-and-ip-risk`, `review-terms-policy-readiness` | — |
| §7 Release & Beta | `ship-release` | Quality gate pass | `setup-quality-gates`, `auto-create-pr`, `automate-release-tagging`, `sync-release-docs`, `write-changelog`, `guard-destructive-commands`, `compose-safety-mode` | — |
| §8 Operate & Iterate | `iterate-product` | Production traffic | `design-ab-experiment`, `analyze-ab-experiment`, `analyze-user-funnel`, `generate-improvement-tasks`, `handle-incident`, `conduct-postmortem`, `monitor-regressions`, `summarize-retro` | — |
| §9 Lifecycle Management | `manage-lifecycle` | Feature/product 노후화 | — | — |

---

## 5. 노출된 커맨드 목록 (plugin.json commands)

> 사용자가 `/buddy:<name>`으로 직접 호출할 수 있는 29개 커맨드.
> 9개 단계 진입점 + 1개 다각도 리뷰 + 5개 공통 도구 + 13개 단계별 세부 작업 + 1개 상태 확인 = 29.
> 패턴 라이브러리와 보관 스킬은 manifest 에 노출하지 않는다.

> **9-phase 라이프사이클 단계 약칭** (이하 표에서 사용):
> 1) 아이디어 구체화 / 2) Feature 정의 / 3) 기술 설계 / 4) 구현 계획 / 5) 개발 / 6) 품질 검증 / 7) 릴리즈 / 8) 운영·개선 / 9) 수명주기 관리.

### 5.1 상태 확인 (1)

| 커맨드 | 단계 | 용도 |
|--------|------|------|
| `/buddy:status` | 공통 | 현재 작업 단계 확인 + 다음에 실행할 명령 안내 |

### 5.2 단계 진입점 (9)

> 각 라이프사이클 단계의 시작 게이트. 진입 조건이 맞으면 해당 단계의 모든 작업을 자동 진행.

| 커맨드 | 단계 | 용도 |
|--------|------|------|
| `/buddy:start` | 1. 진입 라우터 | 신규 아이디어 → `concretize-idea` / 기존 프로덕트 변경 → `assess-product-change` 자동 dispatch |
| `/buddy:define-features` | 2. Feature 정의 | PRD → actor / use case → feature backlog |
| `/buddy:design-system` | 3. 기술 설계 | 기술 스택 / API 계약 / infra / 데이터 모델 |
| `/buddy:plan-build` | 4. 구현 계획 | actor 별 task 분해 + 의존성 그래프 |
| `/buddy:build-feature` | 5. 개발 | TDD 루프 + 병렬 worker agent |
| `/buddy:verify-quality` | 6. 품질 검증 | 테스트 + 보안 + 컴플라이언스 |
| `/buddy:ship-release` | 7. 릴리즈 | PR + 태깅 + canary + UAT |
| `/buddy:iterate-product` | 8. 운영·개선 | A/B 분석 + 인시던트 + funnel |
| `/buddy:manage-lifecycle` | 9. 수명주기 관리 | deprecation + 마이그레이션 + EOL |

### 5.3 다각도 리뷰 (1)

> 어느 단계든 plan / PRD / 설계 산출물이 생기면 호출 가능. 단일 단계 종속 없음.

| 커맨드 | 단계 | 용도 |
|--------|------|------|
| `/buddy:autoplan` | 공통 (리뷰) | 산출물을 scope / design / engineering / DX 4개 관점으로 자동 리뷰 |

### 5.4 공통 도구 (5)

> 단계 종속 없음. 어디서든 호출 가능.

| 커맨드 | 단계 | 용도 |
|--------|------|------|
| `/buddy:consult-codex` | 공통 | 외부 LLM (codex 등) 으로 second opinion |
| `/buddy:save-context` | 공통 | 체크포인트 저장 (브랜치 무관 이어받기) |
| `/buddy:restore-context` | 공통 | 체크포인트 복원 |
| `/buddy:write-a-skill` | 공통 / 메타 | 신규 buddy 스킬 작성 + catalog 등재 + 차용 4분류 정책 적용 + RED-GREEN-REFACTOR subagent pressure test (한 사이클) |
| `/buddy:decompose-blocker` | 공통 | 코드 작업 중 stuck 상태에서 문제 분해 + 비용-정보 매트릭스 기반 행동 후보 도출 (다음 스킬로 dispatch 준비) |

### 5.5 단계별 세부 작업 (13)

> 단계 진입점 안에서 자동 호출되거나, 사용자가 단독 호출 가능 (dual-mode).

| 커맨드 | 단계 | 용도 |
|--------|------|------|
| `/buddy:validate-idea` | 1. 아이디어 구체화 | YC 스타일 검증 인터뷰 |
| `/buddy:validate-advanced-edge-idea` | 1. 아이디어 구체화 | 엣지 케이스 / 숨은 가정 박멸 |
| `/buddy:assess-business-viability` | 1. 아이디어 구체화 | 사업성 7차원 평가 |
| `/buddy:define-product-spec` | 1. 아이디어 구체화 | PRD 고정 |
| `/buddy:verify-best-alternative` | 3. 기술 설계 | AI 편향 방지 강제 다관점 검토 |
| `/buddy:build-with-tdd` | 5. 개발 | TDD 루프 단독 실행 |
| `/buddy:diagnose-bug` | 5. 개발 | 버그 재현 → 원인 → fix |
| `/buddy:dispatch-parallel-agents` | 5. 개발 | worktree 격리 + worker 분배 |
| `/buddy:audit-security` | 6. 품질 검증 | OWASP / secrets / JWT 점검 |
| `/buddy:measure-code-health` | 6. 품질 검증 | 0-10 가중 점수 대시보드 |
| `/buddy:auto-create-pr` | 7. 릴리즈 | PR 자동 생성 |
| `/buddy:setup-quality-gates` | 7. 릴리즈 | pre-commit / pre-push 게이트 설치 |
| `/buddy:summarize-retro` | 8. 운영·개선 | git history → 주간 회고 |

> 단계 2 / 4 / 9 의 세부 작업 커맨드는 현재 0개 — 단계 진입점 안의 기존 stage skill 만 활성. 신규 작업은 [`docs/archive/tasks.md`](../../../../docs/archive/tasks.md) A-1 참조 (잔여 작업 SSoT 는 [`docs/BACKLOG.md`](../../../../docs/BACKLOG.md)).

**규칙**: plugin.json `commands` 에 새 항목을 추가하려면 §2 의 도메인 우선순위 표에서 1~4 등급에 속해야 하고, 이 §5 의 적절한 sub-section 에 먼저 등재해야 한다. 패턴 라이브러리와 보관 스킬은 영구 비공개.

---

## 6. 새 skill 추가 시 router 검토 체크리스트

새 skill을 `skill-catalog.md`에 등재한 직후 다음을 확인 — 위반하면 이 문서에 항목 추가:

- [ ] 동일 command 이름이 이미 등재되어 있지 않다 (`/buddy:<name>` 충돌 X)
- [ ] Description이 다른 skill의 description과 의미상 90% 이상 겹치지 않는다
- [ ] 같은 hook event(PreToolUse 등)에 매핑된 skill이 이미 있을 때, 실행 순서가 명시되었다
- [ ] 사용자 발화 패턴 1~2개로 이 skill이 dispatch되는지 mental sim 통과
- [ ] Phase 소속이 명확히 정의되었다 (§4 Phase 표에 항목 추가)

---

## 7. 참조

- Skill 카탈로그 본문 → [`skill-catalog.md`](./skill-catalog.md)
- Plugin manifest → [`.claude-plugin/plugin.json`](../../../.claude-plugin/plugin.json)
- 9-phase 라이프사이클 아키텍처 설계 → [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../../../../docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md)
- 11-stage 상용 제품 빌딩 flow (참조용) → [`docs/archive/skill-map.md`](../../../../docs/archive/skill-map.md)
- Plugin scaffold spec → [`docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`](../../../../docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md)
- Archive 스킬 → [`plugin/_archive/`](../../../_archive/)
