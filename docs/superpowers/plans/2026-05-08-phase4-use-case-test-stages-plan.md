# Phase 4 — §6 Use-case 테스트 2 Stage Skill 작성 Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement task-by-task.

**Goal:** §6 verify-quality 의 use-case 기반 테스트 layer 를 채워 Q8=(a) cascade 의 마지막 puzzle piece 를 닫는다. §2 use case → §3 system boundary → §4 actor track → §5 build → **§6 actor-별 + cross-actor test** 의 5-단계 cascade 가 본 phase 로 완전 활성화.

**Parent plan:** [`2026-05-06-stage-buildout-plan.md`](./2026-05-06-stage-buildout-plan.md) Phase 4.
**SSoT:** [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../specs/2026-05-06-lifecycle-orchestrator-architecture.md) §6 + §10 항목 6.
**Predecessor plans:** Phase 1 / Phase 2 / Phase 3 — PROCEDURE template (§0~§11) reuse.

## Scope (2 skills)

| Skill name | 1줄 용도 | 의존 (입력) |
|-----------|---------|------------|
| `test-per-actor-use-case` | actor 의 use case 단위 통합 테스트 — frontend (E2E browser), backend (API integration), 3rd-party (contract) | §2 feature spec actor + use case, §4 acceptance test plan |
| `test-cross-actor-flow` | actor 협업 시나리오 E2E 테스트 — cross-actor edge 가 있는 flow 의 full-stack 검증 | §4 task DAG cross-actor edges, acceptance test plan cross-actor flow |

**왜 이 2개:** verify-quality orchestrator 가 이미 stage 2 (test-per-actor-use-case) + stage 3 (test-cross-actor-flow) 로 reference 중인데 PROCEDURE.md 부재 → orchestrator 가 actual skill 호출 못 하고 inline 설명만 가짐. 본 phase 가 그 빈 stage 채워 Q8=(a) cascade 닫기.

**Use case cascade 활성화 (Q8=a 완성):**
- §2 use case (actor → use case → acceptance criteria)
- §3 system boundary (actor → component)
- §4 actor track (decompose-feature-to-actor-tracks → tasks → DAG)
- §5 build (actor 별 worker dispatch → code + tests via TDD)
- **§6 test-per-actor-use-case + test-cross-actor-flow** ← 본 phase

5-단계 chain 의 §6 가 본 phase 로 활성화.

## Out of scope

- §6 의 보강 audit (load test, a11y, i18n, cost, chaos, mutation) — Phase 7 deferred
- §6 의 review-* (AI safety / privacy / license / terms) — 이미 [Done], 본 phase 무관
- 별도 testing framework 도입 — `define-acceptance-test-plan` 산출 (Vitest / Playwright / Pact / Schemathesis / k6) 그대로 재사용

## 진행 원칙

- **순차 작성** — Task 4.1 (per-actor) → Task 4.2 (cross-actor). 후자가 전자의 acceptance result 를 참조.
- **각 skill = 단일 commit**
- **PROCEDURE template 재사용** — Phase 1-3 의 §0 STOP / §4 posture / §6 explicit output / §11 verification gate 4 enhancement 그대로
- **Side-effecting 처리 명시** — 본 skill 들은 actual test 실행 산출 (test report) — read-only on production code, but write test file / fixture / report

---

## PROCEDURE template

Phase 1 plan 의 [common template](2026-05-07-phase1-design-stages-plan.md#proceduremd-공통-템플릿) 그대로 사용. 모든 §6 use-case test skill 은 frontmatter 없이 §0 ~ §11 의 12 sections.

---

## Tasks

### Task 4.1: `test-per-actor-use-case` — actor 단위 통합 테스트

**Files:**
- Create: `plugin/skills/test-per-actor-use-case/PROCEDURE.md`
- Create: `plugin/commands/test-per-actor-use-case.md`
- Modify: `plugin/skills/router/references/skill-catalog.md` (§6 row 추가)
- Modify: `plugin/skills/verify-quality/PROCEDURE.md` (stage 2 의 inline 설명을 본 skill 호출로 redirect)
- Modify: `scripts/test-router-wireup.sh` (count 95 → 96)

**Skill 정의:**
- description: "actor 의 use case 단위 통합 테스트 — frontend (Playwright E2E), backend (Vitest+testcontainers integration), 3rd-party (Pact contract). per-actor coverage gap 0 maintain."
- when-to-use: §6 verify-quality 의 stage 2 / 신규 actor 추가 후 / use case 변경 시 / actor 별 coverage gap 발견 후 보강
- inputs: §2 feature spec (actor + use case + acceptance criteria), §4 define-acceptance-test-plan 산출물 (per-actor section), §5 build 의 code complete state
- outputs: actor × use case × test status 매트릭스, test report (pass/fail/skipped + coverage), gap report (use case 가 test 없는 row), test infra 검증 (testcontainers / MSW / Pact mock 모두 동작)
- 충돌 방지: vs `run-browser-qa` (그것은 UI 자체 QA, 본 skill 은 use case 단위 통합) / vs `test-cross-actor-flow` (cross-actor 는 다음 stage) / vs `define-acceptance-test-plan` (그것은 plan, 본 skill 은 plan 의 actual 실행)

**§5 phase 골자:**
1. **Per-actor coverage matrix** — actor × use case × layer (unit/integration/contract/E2E) 매핑, gap 0 verify
2. **Layer 별 test 실행** — actor 별로 적합한 framework 호출 (Vitest unit + integration / Playwright E2E / Pact contract)
3. **Test infra 검증** — testcontainers / MSW / Pact / SES simulator / elasticmq / miniredis 모두 active 검증
4. **Coverage measurement** — line / branch + critical path 100% verify
5. **Gap report + acceptance gate** — use case 별 pass/fail/skipped, gap 발견 시 §4 acceptance test plan 으로 회귀 trigger

**Acceptance:** 표준 (PROCEDURE/command 신규 + catalog/orchestrator 갱신 + CI 95→96)

**Commit:** `feat(skill): add test-per-actor-use-case §6 stage skill`

---

### Task 4.2: `test-cross-actor-flow` — cross-actor 협업 시나리오 E2E

**Files:**
- Create: `plugin/skills/test-cross-actor-flow/PROCEDURE.md`
- Create: `plugin/commands/test-cross-actor-flow.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/verify-quality/PROCEDURE.md` (stage 3 redirect)
- Modify: `scripts/test-router-wireup.sh` (count 96 → 97)

**Skill 정의:**
- description: "cross-actor flow E2E test — actor 협업 시나리오 (signup → email → verify → login → me 같은 multi-actor chain) full-stack 검증. cross-actor edge 가 있는 flow 의 integration coverage."
- when-to-use: §6 verify-quality stage 3 / cross-actor contract 변경 후 / new flow 추가 시
- inputs: §4 task DAG cross-actor edges, define-acceptance-test-plan 의 cross-actor flow section, test-per-actor-use-case 산출 (per-actor green 이 prerequisite)
- outputs: flow × actor list × test status, integration evidence (Loom / video / log + screenshot), cross-actor edge coverage matrix, contract drift detection
- 충돌 방지: vs `test-per-actor-use-case` (그것은 actor 내부, 본 skill 은 actor 간) / vs `run-browser-qa` (그것은 UI 자체, 본 skill 은 multi-system flow) / vs `monitor-regressions` (post-launch regression detection)

**§5 phase 골자:**
1. **Flow inventory** — define-acceptance-test-plan 의 cross-actor flow 모두 enumerate (signup happy / forgot pw / GDPR / bounce / lockout / RLS / rate-limit / idempotency 등)
2. **Flow 별 actor chain 매핑** — 각 flow 가 거치는 actor list + cross-actor edge sequence (DAG cross-actor edges)
3. **Test 실행** — Playwright + LocalStack + SES simulator + miniredis + testcontainers 전부 동시 active, per-flow scenario step 실행
4. **Cross-actor edge coverage** — task DAG 의 cross-actor edge 가 모두 어떤 flow 의 일부로 cover 되는지 verify
5. **Contract drift detection** — Pact provider verify + Schemathesis fuzz + oasdiff 종합, drift 발견 시 §3 design-api-contract 회귀

**Acceptance:** 표준

**Commit:** `feat(skill): add test-cross-actor-flow §6 stage skill`

---

### Task 4.3: `verify-quality` orchestrator wiring

**Files:**
- Modify: `plugin/skills/verify-quality/PROCEDURE.md` (stage 2 + stage 3 의 inline 설명을 skill 호출로 redirect, [Done] marker 추가)

**Background:** orchestrator 가 이미 stage 2 + stage 3 로 reference 했지만 inline 설명만 있고 actual skill 호출 라인 없음. Phase 4 후 본 skill 호출 라인 추가.

**Acceptance:** stage 2 + stage 3 모두 [Done] 표기 + invoke line 명시.

**Commit:** `feat(skill): wire test-per-actor + test-cross-actor stages into verify-quality orchestrator`

---

### Task 4.4: v1.0.5 release

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json` (1.0.4 → 1.0.5)
- Modify: `.claude-plugin/marketplace.json` (95 → 97 procedures, 47 → 49 commands)
- Modify: `CHANGELOG.md` (1.0.5 entry)
- Modify: `scripts/test-router-wireup.sh` (이미 4.1+4.2 에서 갱신됨, final verify)

**Background:** Phase 4 의 2 skill 추가로 PROCEDURE.md 95 → 97, command md 47 → 49.

**Acceptance:**
- `find plugin/skills -name PROCEDURE.md | wc -l` = 97
- `find plugin/commands -name "*.md" | wc -l` ≥ 49
- CI script 통과
- `claude plugin validate` 통과

**Commit:** `release: v1.0.5 (Phase 4 §6 use-case test 2 stage skills)`

---

## Verification (Phase 4 종료 시)

- [ ] 2 신규 skill 모두 PROCEDURE.md 존재 (frontmatter 없음, §0~§11 12 sections)
- [ ] 2 신규 command md 가 router single-mode dispatch 패턴 일치
- [ ] skill-catalog.md §6 section 에 2 row 추가
- [ ] verify-quality orchestrator stage 2 + stage 3 모두 [Done] + invoke line
- [ ] spec §3 §6 status 가 [Partial] → [Partial-still-some-pending] (load/a11y/cost/chaos 잔여) 명시
- [ ] CI script PROCEDURE count 97 통과
- [ ] plugin validate 통과

## Rollback

각 task = 단일 commit, `git revert` 안전 롤백. release commit 만 별도 (version bump + CHANGELOG + count). rollback 순서: release revert → orchestrator wiring revert → 2 skill commits revert (역순).

## 다음 plan (Phase 4 완료 후)

Q8=(a) cascade 가 본 phase 로 완료. 다음 priority candidate:
- (B) v1.0.5 통합 documentation update (spec / skill-map / README sync)
- (C) Phase 7 (deferred) 재평가 — §1 / §3 부가 / §5 부가 / §6 부가 / MCP
- Phase 5 (§8 데이터 분석 보강 7 skill) — production traffic 발생 후 가치 발휘
- Phase 6 (§9 Lifecycle 4 skill) — 1 년+ 운영 시 가치
