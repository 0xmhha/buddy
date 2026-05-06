# Stage Skill Buildout — Buddy 9-Phase 채움 Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement Phase 0 task-by-task. Phases 1+ 는 각자 별도 plan 으로 분할되어 순차 실행됨.

**Goal:** 라우팅 인프라 완료(2026-05-06) 후 잔여 stage skill 54 개와 인프라 정리 항목을 우선순위 따라 점진적으로 채워, buddy 가 9-phase 라이프사이클 풀 사이클을 실제로 지원하도록 한다.

**SSoT:** [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../specs/2026-05-06-lifecycle-orchestrator-architecture.md) §10 의 미해결 작업.

---

## Strategy

| Phase | 주제 | 범위 | 형태 |
|-------|------|------|------|
| **Phase 0** | Foundation stabilization | description drift 정리 + CI smoke test 자동화 | 본 plan 에서 detailed task 로 실행 |
| **Phase 1** | §3 Technical Design 핵심 stage | `define-tech-stack`, `design-data-model`, `design-api-contract`, `write-adr` (4 skills) | 별도 plan |
| **Phase 2** | §4 Implementation Plan stage | `decompose-feature-to-actor-tracks`, `decompose-track-to-tasks`, `map-task-dependencies`, `plan-parallel-execution`, `define-acceptance-test-plan`, `estimate-build-timeline` (6 skills) | 별도 plan |
| **Phase 3** | §7 Release safety nets | `setup-canary-deploy`, `setup-feature-flags`, `setup-rollback-runbook`, `run-uat`, `run-beta-program`, `prepare-launch-checklist`, `setup-incident-paging` (7 skills) | 별도 plan |
| **Phase 4** | §6 Use-case 테스트 stage | `test-per-actor-use-case`, `test-cross-actor-flow` (2 skills) | 별도 plan |
| **Phase 5** | §8 데이터 분석 보강 | `analyze-feature-adoption`, `analyze-user-cohort`, `analyze-actor-failure-rate`, `analyze-cost-anomaly`, `triage-customer-support-ticket`, `analyze-customer-feedback-corpus`, `audit-error-budget` (7 skills) | 별도 plan |
| **Phase 6** | §9 Lifecycle stage | `deprecate-feature`, `migrate-customers`, `archive-product`, `spin-off-feature` (4 skills) | 별도 plan |
| **Phase 7** *(deferred)* | §1 customer/market + §3 부가 design + §5 부가 build + §6 부가 audit + MCP | ~28 skills + 2 MCPs | 1년 또는 단기 commercial pivot 시 trigger |

**진행 원칙:**
- Phase 0 종료 후 Phase 1 plan 작성 → 실행 → Phase 2 plan 작성 → … 의 순서.
- 각 Phase 가 끝날 때 spec §3 (9-Phase 라이프사이클 모델) 표의 해당 phase 상태가 🟡 → ✅ 로 바뀌어야 함.
- 동일 stage skill 의 PROCEDURE.md 작성 패턴은 Phase 1 의 첫 skill 에서 reference template 이 생성되어 후속 phase 가 재사용.

---

## Phase 0: Foundation Stabilization

라우팅 인프라는 정상 동작하지만 다음 두 영역에 정리가 필요. Phase 1 의 stage skill 작성 전에 기반을 안정화한다.

### Task 0.1: plugin.json ↔ command md description drift 정리

**Files:**
- Modify: `plugin/commands/buddy/*.md` × 26 (run/chain/parallel/status 제외 — 이미 일치)

**Background:** `plugin/.claude-plugin/plugin.json` 의 27 commands 중 26 개에서 description 이 `plugin/commands/buddy/<name>.md` frontmatter 의 description 과 다름. plugin.json 은 commit `197725a feat(plugin): expand command surface to 27 with plain-language descriptions` 에서 plain-language 로 재작성됐지만, command md 는 동기화되지 않음. 사용자가 `/buddy:` 슬래시 자동완성 + 명령 도움말 두 surface 에서 모순된 텍스트를 보게 됨.

**Decision:** plain-language 측(plugin.json) 을 canonical 로 하고 command md frontmatter 를 정렬한다. (commit 197725a 의 design intent 가 plain-language 였음.)

**Steps:**
- [ ] Step 1: `plugin.json` 의 26 개 entry 각각의 `(name, description)` 추출
- [ ] Step 2: 각 command md 의 frontmatter `description:` 라인을 plugin.json 값으로 교체. argument-hint, body, H1 모두 보존.
- [ ] Step 3: 검증 — 모든 26개 파일에서 `awk` 로 frontmatter description 추출 → plugin.json 동일 entry 와 byte-identical 확인
- [ ] Step 4: commit `chore(commands): align command md descriptions with plugin.json plain-language register`

**Acceptance:**
- `for cmd in $(jq -r '.commands[].name' plugin/.claude-plugin/plugin.json); do md_desc=$(awk '/^---$/{c++; next} c==1 && /^description:/{sub(/^description: */, ""); print; exit}' plugin/commands/buddy/$cmd.md 2>/dev/null); json_desc=$(jq -r ".commands[] | select(.name == \"$cmd\") | .description" plugin/.claude-plugin/plugin.json); [ "$md_desc" = "$json_desc" ] && echo "OK $cmd" || echo "DRIFT $cmd"; done` → 모든 OK
- 변경 외 영향 없음 (run/chain/parallel/status 는 이미 일치이므로 무수정)

### Task 0.2: Router wire-up CI smoke test 추가

**Files:**
- Create: `bin/test-router-wireup.sh` (또는 기존 bin/ 컨벤션에 맞는 경로)
- Modify: 기존 CI 설정 (`Makefile` 또는 `.github/workflows/*.yaml` 기존 파일)

**Background:** Phase 6.2 verification 은 정적 wire-up 만 검증했고 live `/buddy:*` 호출은 사용자가 reload 후 직접 실행해야 한다. 회귀 방지용 자동 검증 스크립트 필요.

**Steps:**
- [ ] Step 1: bash script 작성. 검증 항목:
    - `find plugin/skills -name SKILL.md` 결과가 정확히 1 (`plugin/skills/router/SKILL.md`)
    - `find plugin/skills -name PROCEDURE.md` 결과가 정확히 78
    - `jq '[.commands[] | select(.skill != "router")] | length'` plugin/.claude-plugin/plugin.json 결과가 0
    - 모든 plugin.json command 에 대응하는 `plugin/commands/buddy/<name>.md` 존재
    - 모든 plugin/commands/buddy/*.md (run/chain/parallel 제외) 에 대해 단일-mode dispatch 의 target PROCEDURE 가 실제로 PROCEDURE.md 로 존재
    - chain.md / parallel.md / run.md 의 target 추출 spec 이 router/SKILL.md 의 dispatch contract 와 일치 (정규식 또는 키 검색)
- [ ] Step 2: 스크립트가 모든 검증 통과 시 exit 0, 실패 시 명확한 에러 메시지 + exit 1
- [ ] Step 3: 기존 CI hook 에 추가 (Makefile target `test-routing` 또는 GitHub Actions step)
- [ ] Step 4: README 또는 CONTRIBUTING 에 `make test-routing` 사용법 1 줄 추가
- [ ] Step 5: commit `feat(ci): add router wire-up smoke test for skill-routing regression`

**Acceptance:**
- 새 PR 이 SKILL.md 추가 / PROCEDURE.md 삭제 / plugin.json 의 router-skip 등 reverts 를 시도하면 CI 가 실패
- 정상 상태에서 `bin/test-router-wireup.sh` 실행 시 0 exit + 모든 검증 통과 메시지

### Task 0.3: Live smoke test (사용자 실행)

> **이 task 는 subagent 가 자동 실행할 수 없음.** Phase 0 의 마지막 task 로 사용자가 수동 수행한다.

**Steps:**
- [ ] Step 1: Claude Code 세션을 새로 띄우거나 `/reload-plugins` 실행
- [ ] Step 2: 5 개 representative command 를 실제 호출:
    - `/buddy:status`
    - `/buddy:concretize-idea "test idea"`
    - `/buddy:audit-security`
    - `/buddy:chain validate-idea, assess-business-viability -- "test idea"`
    - `/buddy:parallel review-engineering, review-design -- "this branch"`
- [ ] Step 3: 각 호출에서 다음 검증:
    - router 가 fire 하는가
    - target PROCEDURE.md 가 실제로 read 되는가 (router 본문 절차에 따른 first-step 시도 → fallback 결과)
    - chain / parallel 모드의 인자 분해가 의도대로 작동하는가
- [ ] Step 4: 결과를 `docs/notes/2026-05-NN-live-smoke-test.md` (또는 사용자 선호 위치) 에 기록 — 실패 시 `${CLAUDE_PLUGIN_ROOT}` resolution 이 어느 환경에서 어떻게 실패했는지 포함
- [ ] Step 5: 실패 발견 시 router/SKILL.md 의 path resolution 섹션 보강 + commit

**Acceptance:**
- 5 개 command 모두 의도대로 작동, 또는 실패 사례에 대한 fix commit 으로 자기 보정

---

## Phase 1+ Outline (별도 plan 으로 분할 예정)

각 phase 는 자체 plan 파일을 가진다. 본 plan 은 phase 0 종료 후 다음 plan 작성을 위한 가이드 역할.

### Phase 1: §3 Technical Design 핵심 stage (4 skills)

대상: `define-tech-stack`, `design-data-model`, `design-api-contract`, `write-adr`

**왜 이 우선순위:** 락인 영향 가장 큼. 잘못된 결정 비용이 다른 phase 보다 1~2 자릿수 큼. design-system orchestrator 가 이미 존재하므로 stage 만 채우면 즉시 활용 가능.

**Plan 작성 시 포함할 항목:**
- 각 skill 의 `description` (라우팅 매칭에 사용) — 80~150 자
- PROCEDURE.md body 구조: when-to-use → input → steps → output format → cross-phase cascade (이 skill 산출물이 어느 phase 의 입력이 되는지)
- 4 skill 의 공통 패턴을 reference template 으로 정리 — Phase 2+ 가 재사용
- routing-rules.md / skill-catalog.md 갱신 (4 entry 추가)

**Acceptance criteria template:**
- `plugin/skills/<name>/PROCEDURE.md` 존재, frontmatter (name + description) 유효
- `design-system` orchestrator 의 stage 흐름 표에 새 skill 매핑됨
- skill-catalog.md §3 표에 row 추가
- `bin/test-router-wireup.sh` 통과

### Phase 2: §4 Implementation Plan stage (6 skills)

대상: `decompose-feature-to-actor-tracks`, `decompose-track-to-tasks`, `map-task-dependencies`, `plan-parallel-execution`, `define-acceptance-test-plan`, `estimate-build-timeline`

**왜 이 우선순위:** §3 산출물(actor / system boundary / API contract) 이 §4 의 입력. Phase 1 완료 후 자연스럽게 §4 가 활용 가능. plan-build orchestrator 안에 stage 가 거의 비어있어 §4 의 commercial 가치가 현재 잠재 상태.

### Phase 3: §7 Release safety nets (7 skills)

대상: canary, feature-flags, rollback, UAT, beta, launch-checklist, incident-paging

**왜 이 우선순위:** 상용 배포 직전 필수. 사고 발생 비용이 가장 높은 영역. ship-release orchestrator 의 현재 state 는 PR 자동화 + tagging 까지만 — 실제 production 안전망 없음.

### Phase 4: §6 Use-case 테스트 stage (2 skills)

대상: `test-per-actor-use-case`, `test-cross-actor-flow`

**왜 이 우선순위:** Q8=(a) cascade 활성화의 마지막 puzzle piece. §2 use case 분해 → §3 system boundary → §4 actor track → §6 actor 별 테스트로 흐름이 닫힘.

### Phase 5: §8 데이터 분석 보강 (7 skills)

대상: cohort, feedback corpus, cost anomaly, SLO, etc.

**왜 후순위:** 핵심 §8 stage (AB / funnel / incident / postmortem) 는 이미 구현됨. 이번 보강은 dataset / SaaS 통합 의존이 큼.

### Phase 6: §9 Lifecycle stage (4 skills)

**왜 후순위:** 1년 이상 운영 시 발생하는 deprecation / EOL 영역. 단기 가치 없음.

### Phase 7 (deferred): 잔여 ~28 skills + MCP

- §1 customer/market 분석 5개 (`analyze-competition-and-substitutes`, `map-customer-segments`, `map-jobs-to-be-done`, `analyze-market-size`, `conduct-customer-interview`)
- §3 부가 design 11개 (`design-tenant-model`, `design-i18n-strategy`, etc.)
- §5 부가 build 5개 (`generate-from-api-contract`, `pair-program-loop`, etc.)
- §6 부가 audit 6개 (load test, a11y, i18n, cost, chaos, mutation)
- §3 use case → infra 브릿지 2개 (`map-use-cases-to-infra`, `derive-system-topology`)
- MCP 단계: feature-management-mcp 완성, analytics-mcp 신규 — Q4=(c) 보류 해제 시 trigger

---

## Out of Scope (이 plan 시리즈 전체)

- 외부 SaaS MCP 어댑터 (monitoring/support/cost/billing/feature-flag) — 외부 의존, 별도 트랙
- Buddy plugin distribution / install 메커니즘 변경
- 신규 phase 추가 (§7.5 Beta/UAT 분리 같은 구조 변경 — Q7=(b) 결정이지만 현재는 §7 안의 stage 로 흡수)

---

## Verification

각 phase 종료 시점에 다음 갱신 확인:
- spec §3 의 phase 상태 표 (🟡 → ✅) 갱신
- skill-catalog.md / routing-rules.md 의 해당 phase row 갱신
- `bin/test-router-wireup.sh` 통과 (Phase 0 후 활성)
- Phase 끝마다 git tag 또는 milestone commit 으로 진척 marker

본 plan 은 Phase 0 완료 시점에 status 표 (Phase column 별 ✅/⏳) 추가하여 갱신한다.

---

## Rollback

Phase 0 의 두 task (description drift, CI script) 는 모두 독립 commit 이라 단일 `git revert` 로 안전 롤백. Phase 1+ 는 각 phase plan 안에서 자체 rollback 정책 정의.
