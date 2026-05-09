# Buddy — Session Handoff: Skill Context Bloat 조사 + 잔여 작업

> **목적:** 다른 세션에서 작업을 이어받을 때, 현재 v1.0.8 상태와 *조사 의뢰된 핵심 이슈*(buddy plugin skill load → context 폭증) + 잔여 작업 우선순위를 5분 안에 파악하기 위한 단일 문서.

**작성일:** 2026-05-08
**현재 plugin version:** v1.0.8 (origin/main `fe5363f`)
**현재 session 종료 상태:** working tree clean, push 완료

---

## 1. 직전 세션의 마지막 컨텍스트 (한 줄)

`/compact` 직후 사용자가 다음 작업을 다른 세션에서 이어받기 위한 핸드오프 문서 작성 요청. 직전 작업은 v1.0.8 retest doc push 완료 (`fe5363f`).

---

## 2. 다음 세션이 *가장 먼저* 다뤄야 할 의뢰 (NEW — 미착수)

### 2.1 의뢰 내용 (skill-creator 트리거 경유)

> `'buddy' 프로젝트에 의해 추가된 skill 이 모두 load 되고, prompt 사용시마다 모두가 포함되어 context 를 대량으로 사용하게 만드는 문제가 있어. indexing 을 통해 개선 방안을 검토해.`

### 2.2 사실 정리 (현재 알려진 것)

- `plugin/skills/` 산하 SKILL.md 파일 = **1개** (router 만 frontmatter 보유)
- 그 외 105개 절차는 `PROCEDURE.md` (frontmatter 없음 — Claude Code skill discovery 대상이 아님)
- `plugin/commands/` = **57개** slash command md (각 평균 ~560 byte, 합계 ~32KB)
- 명목상 설계는: "단일 router skill + 절차 lazy load" (single-router dispatch)

### 2.3 그러나 관측되는 현상

이 세션의 system-reminder 블록에는 **19개의 buddy:* skill 본문**이 "EARLIER invoked" 라벨로 누적 노출됨:
- `buddy:router`, `buddy:status`, `buddy:concretize-idea`, `buddy:define-tech-stack`, `buddy:design-data-model`, `buddy:design-api-contract`, `buddy:write-adr`, `buddy:decompose-feature-to-actor-tracks`, `buddy:decompose-track-to-tasks`, `buddy:map-task-dependencies`, `buddy:plan-parallel-execution`, `buddy:define-acceptance-test-plan`, `buddy:estimate-build-timeline`, `buddy:run-uat`, `buddy:test-per-actor-use-case`, `buddy:map-use-cases-to-infra`, `buddy:audit-security`, `buddy:chain`, `buddy:parallel`

각 항목이 commands/*.md 본문 ~600 byte + 자체 description 등을 포함 → 누적 시 **수십 KB context overhead**.

### 2.4 가설 (확신도 라벨)

| 가설 | 확신도 |
|------|--------|
| Claude Code 가 plugin install 시 commands/*.md 의 frontmatter `description` 을 **모든** 명령어에 대해 메타데이터 인덱스에 자동 등록 | High — 의뢰 내용과 system-reminder 양상이 일치 |
| 한 번 invoke 한 command 의 본문은 같은 세션 내내 system-reminder 로 반복 노출 (cumulative) | High — 이 세션 자체가 evidence |
| `description` 이 한국어 + 길게 작성되어 token 비용 증폭 | Mid — 57개 × 평균 80~120 byte description = 수 KB 고정 비용 |
| Skill index 를 별도 reference 파일로 분리해 commands/*.md 를 thin invocation stub 으로만 두면 누적 비용 감소 | Mid — 근거: progressive disclosure 원칙 (skill-creator references) |

### 2.5 검토할 개선 방향 (의견)

다음 세션에서 우선 시도할 후보 — *순서는 비용 대비 효과 추정 기준*:

1. **Description token diet (Quick win, Low risk)**
   - 57개 `commands/*.md` 의 frontmatter description 을 *반드시 필요한 trigger keyword* 만 남기고 단축
   - 현재: `"아이디어 구체화 단계 — 검증, 사업성 평가, PRD 작성까지 한 번에."` (40~60 token)
   - 후보: `"§1 Idea: PRD + viability."` (10 token 이하)
   - **expected gain:** 57 × 30~50 token = 1500~3000 token 절감 (세션 시작 고정 비용)

2. **commands/*.md 본문 슬림화 (Medium effort)**
   - 현재 본문에 매번 동일한 router 호출 보일러플레이트 (`mode: single` / `target PROCEDURE: <name>`) 반복
   - 후보: 본문은 1줄 (`Run /buddy:router with target=<name>`), 나머지는 router skill body 가 알아서 처리
   - 이미 일부 단순 — 추가 절감 여지는 ~100~200 byte/file × 57 = 6~10 KB

3. **Skill index 분할 (Architectural)**
   - 현재 `plugin/skills/router/references/skill-catalog.md` 가 실재 — lazy load 대상
   - 그러나 일부 commands 가 이를 우회해 직접 dispatch contract 를 본문에 inline → 인라인 표 제거 검토
   - 문제: dispatch contract 가 인라인이어야 router 가 frontmatter 없이도 라우팅 결정 가능. 절충점 필요.

4. **누적 invocation 재사용 (실현성 ?)**
   - 동일 세션에서 같은 command 를 두 번째 invoke 시 system-reminder 가 다시 본문 전체를 노출하는지 확인
   - Claude Code 측 caching 동작 검증 필요 — repo 외부 의존

5. **Plugin manifest 레벨 가이드**
   - `plugin.json` 에 `"contextHints"` 같은 필드로 "only load on explicit /buddy:* invocation" 명시 가능한지 Claude Code 문서 재확인

### 2.6 다음 세션 권장 첫 액션

```
- [ ] commands/*.md 57개 description 길이 audit (실제 token 측정)
- [ ] description 단축 PR 시안 (1~2개 sample 로 효과 측정 후 일괄 적용)
- [ ] Claude Code 의 plugin discovery / context inclusion 동작 문서 재확인 (WebFetch 로 docs.anthropic.com plugin spec)
- [ ] 위 검증 후 design 결정 → ADR 로 영속화
```

**열려있는 질문:**
- Claude Code 가 정말 `description` 만 인덱싱하나, 아니면 commands/*.md 본문 전체를 항상 로드하나? (사용자 환경에서 측정 필요)
- system-reminder 누적은 Claude Code 의 conversation history 정책인가, plugin manifest 의 영향인가?

---

## 3. 현재 v1.0.8 상태 (직전 세션 산출물)

### 3.1 Release 진행 (5 minor releases this session)

| Version | Date | Phase | Skills added |
|---------|------|-------|--------------|
| v1.0.4 | 2026-05-08 | Phase 3 (§7 Release Safety Nets) | run-uat, run-beta-program, setup-canary-deploy, setup-feature-flags, setup-rollback-runbook, prepare-launch-checklist, setup-incident-paging (7) |
| v1.0.5 | 2026-05-08 | Phase 4 (§6 Use-case Test) | test-per-actor-use-case, test-cross-actor-flow (2) |
| v1.0.6 | 2026-05-08 | Phase 5 ext Cluster C (§3 cascade bridge) | map-use-cases-to-infra, derive-system-topology (2) |
| v1.0.7 | 2026-05-08 | Phase 5 ext Cluster A (§6 launch readiness) | run-load-test, audit-accessibility, audit-cost-efficiency (3) |
| v1.0.8 | 2026-05-08 | Phase 5 ext Cluster B (§3 SaaS pattern) | design-event-schema, design-auth-model, design-tenant-model (3) |

**누적:** 78 baseline → **105 PROCEDUREs**, 30 → **57 slash commands**.

### 3.2 Retest 결과

- v1.0.4 retest: 17/17 PASS (`docs/notes/2026-05-08-live-retest-v1.0.4.md`)
- v1.0.5 retest: 19/19 PASS (Phase 4)
- v1.0.8 retest: 27/27 PASS (Phase 5 ext Cluster A+B+C 8 skill commercial-grade verified)

### 3.3 Q8=(a) cascade 5-단계 chain 완성

`use case → system → actor track → build → test` — silent gap (§2→§3) 해소.

권장 chain 패턴:
```bash
# §3 cascade
/buddy:chain define-tech-stack,map-use-cases-to-infra,derive-system-topology,design-data-model,design-api-contract,write-adr -- "<project>"

# SaaS pattern
/buddy:chain design-event-schema,design-auth-model,design-tenant-model,write-adr -- "<project>"

# Launch readiness
/buddy:chain run-load-test,audit-accessibility,audit-cost-efficiency,prepare-launch-checklist -- "<project> v<version>"
```

---

## 4. 잔여 작업 리스트 (우선순위 순)

### 4.1 Top Priority — NEW (이번 세션 의뢰)

| # | 작업 | 예상 effort | 비고 |
|---|------|-------------|------|
| **N-1** | **Skill context bloat 조사 + indexing 개선** | 1~2 세션 | §2 참조. 첫 액션은 description audit |

### 4.2 Plugin 트랙 (잔여 37 skill, A-1 표 출처)

> 출처: `docs/tasks.md` §A-1, `docs/superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md`

| Phase | 누락 잔여 | Cluster | 즉시 가치 |
|-------|-----------|---------|----------|
| §1 concretize-idea | 5 (analyze-competition-and-substitutes 등) | E (conditional) | Low |
| §2 define-features | 1 (estimate-feature-effort) | low priority | Low |
| §3 design-system | 4 (design-observability, design-secret-management, design-i18n-strategy, design-accessibility-baseline) | B residual | Mid (조건부) |
| §5 build-feature | 5 (generate-from-api-contract 등) | D | Low (IDE 대체 가능) |
| §6 verify-quality | 3 (audit-i18n-coverage, chaos-test, audit-test-coverage-meaningful) | A residual | Mid |
| §8 iterate-product | 7 (analyze-feature-adoption 등) | F (production traffic 필요) | Deferred |
| §9 manage-lifecycle | 4 (deprecate-feature 등) | G (1년+ deferred) | Deferred |

### 4.3 Plugin dogfood (HIGH value, 사용자 페이스 의존)

| # | 작업 | 비고 |
|---|------|------|
| **A-4-1** | 실 SaaS 프로젝트에 plugin install → 9-phase orchestrator 동작 검증 | 우선순위 1순위 (피드백이 4.2 우선순위 재정렬 입력) |
| A-4-2 | 회수된 피드백으로 A-1 잔여 37 skill 재정렬 | A-4-1 의존 |

### 4.4 buddy MCP server (Q4 결정 trigger 필요)

| 항목 | 상태 |
|------|------|
| `cmd/buddy-mcp/` baseline | DONE (28e9fc9) |
| doctor / stats / feature tools | DONE (28e9fc9) |
| `claude mcp add/remove` 통합 | DONE (a8aee51) |
| feature.query / store / update / link_code / export_patch | TODO (Q4=(c) 보류 상태였으나 in-house 진행 중) |

### 4.5 Go CLI 트랙 (PAUSED)

- B-1: dogfood feedback 회수 (사용자 3~7일 사용 후 template 작성)
- B-2: i18n sweep (M5 deferred — config.ValidationError, queries.Err* 카탈로그 이전)
- B-3: release polish (notarization, SHA pinning, ci.yml 분리)
- B-4~B-6: v0.2 Control Plane / v0.3 Orchestration / v1.0 통합 — outline only

### 4.6 코드 housekeeping

| 항목 | 위치 | 비고 |
|------|------|------|
| `cmd/buddy/main.go` 685 lines 분할 | `cmd/buddy/main.go` | v0.2 새 명령 전 정리 권장 |
| Module path drift | `go.mod` + 전체 import | `wm-it-22-00661/buddy` → `0xmhha/buddy` |

---

## 5. 우선순위 권장 (최종)

1. **N-1 — Skill context bloat 조사** (이번 세션 의뢰, 즉시 착수)
2. **A-4 — Plugin dogfood** (실 프로젝트 install — 모든 후속 결정의 입력)
3. Phase 5 deferred residual (Cluster A residual 3 / B residual 4 — dogfood 후 우선순위 재평가)
4. A-2 — buddy MCP feature.* tools (Q4 결정 trigger 발생 시)
5. B-1 — Go CLI dogfood feedback (사용자 페이스)
6. Cluster F (§8 7) + G (§9 4) — production traffic / 1년+ 운영 후

---

## 6. 다음 세션 시작 시 읽어야 할 문서 (3분 catch-up)

1. **이 문서** — 현재 상태 + 의뢰 + 우선순위
2. `docs/HANDOFF.md` — 트랙 상태 + Plugin v1.0.5 시점 SSoT (※ v1.0.8 으로 갱신 필요 — 이 세션 미반영)
3. `docs/tasks.md` §A-1 표 — 잔여 37 skill cluster 분류
4. `docs/notes/2026-05-08-live-retest-v1.0.4.md` — v1.0.4~v1.0.8 retest 누적 결과 (27/27)

**선택적:**
- `docs/superpowers/plans/2026-05-08-phase7-deferred-reevaluation.md` — Cluster A~H 분석
- `CHANGELOG.md` — v1.0.4~v1.0.8 entry

---

## 7. Repo 상태 (다음 세션 즉시 확인)

```
branch: main (sync with origin)
HEAD: fe5363f docs(notes): record v1.0.8 Phase 5 ext retest results
working tree: clean
plugin version: 1.0.8
total PROCEDUREs: 105
total slash commands: 57
```

**Sync 명령 (다음 세션 첫 액션 — repo root 기준):**
```bash
# (cwd = buddy repo root, e.g. via `cd $(git rev-parse --show-toplevel)`)
git fetch origin && git status
git log --oneline -5
```

---

## 8. Pending docs drift (다음 세션이 처리 가능)

이 세션에서 5 release 를 거치며 일부 SSoT 문서가 v1.0.5 시점으로 stale:

- [ ] `docs/HANDOFF.md` — Plugin 진행 상태 표 v1.0.5 → v1.0.8 갱신 필요 (Phase 5 ext A+B+C Done 반영)
- [ ] `docs/tasks.md` — A-1 잔여 표 갱신 (8 skill 추가 Done — 잔여 37 → 29)
- [ ] `README.md` — 49 commands → 57 commands, 97 skills → 105 skills

(이 세션은 retest doc 만 commit, 위 SSoT 갱신은 의도적으로 deferred — context 절약)

---

## 9. 참고 — Fact-based summary

<Fact-based Answer>
- **Fact (확신도 None = 사실):**
  - plugin v1.0.8 published, origin/main 까지 push 완료 (`fe5363f`)
  - 105 PROCEDUREs, 57 slash commands, working tree clean
  - 5 minor releases this session (1.0.4~1.0.8), 27/27 retest PASS
  - 직전 세션 마지막 user 메시지: "다음 작업을 다른세션에서 진행하려고 하는데, 현재까지 context 와 남은 작업리스트를 정리해서 문서로 정리해줘"
  - skill-creator 경유 의뢰: skill load → context bloat 개선 검토

- **Your Opinion:**
  - **High prediction:** 다음 세션의 첫 작업은 N-1 (skill context bloat 조사) — 이번 세션 사용자 의뢰가 명시
  - **Mid prediction:** description token diet 가 가장 빠른 quick win — 1500~3000 token 절감 추정
  - **Low prediction:** Claude Code 의 plugin manifest spec 에 `contextHints` 같은 hint 가 있을 가능성 — 문서 재확인 필요
  - **None.**
</Fact-based Answer>

---

> **다음 세션 첫 발화 권장 템플릿:**
> "직전 세션의 핸드오프 문서 `docs/notes/2026-05-08-handoff-context-bloat-investigation.md` 읽고 §2 의 N-1 작업부터 시작해줘."
