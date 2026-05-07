# Phase 1 — §3 Technical Design 핵심 4 Stage Skill 작성 Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement task-by-task.

**Goal:** §3 Technical Design phase 의 가장 락인-영향 큰 4 stage skill 을 작성해 `design-system` orchestrator 가 단순 진입점에서 실제 multi-stage 파이프라인으로 작동하도록 만든다.

**Parent plan:** [`2026-05-06-stage-buildout-plan.md`](./2026-05-06-stage-buildout-plan.md) Phase 1.
**SSoT:** [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../specs/2026-05-06-lifecycle-orchestrator-architecture.md) §3 + §4 §3.

## Scope (4 skills)

| Skill name | 1줄 용도 | 락인 영향 |
|-----------|---------|----------|
| `define-tech-stack` | 언어 / 프레임워크 / DB / runtime 선택 | 매우 높음 (마이그레이션 비용 1~2 자릿수 큼) |
| `design-data-model` | 스키마 / 마이그레이션 전략 / 인덱싱 | 높음 (production 운영 시 변경 비용 큼) |
| `design-api-contract` | REST/GraphQL/RPC contract — actor 간 경계 = API 경계 | 높음 (clients 영향, versioning 부담) |
| `write-adr` | Architecture Decision Record 표준 양식으로 결정 보존 | 무 (메타) — 위 3개 결정의 근거 영속화 |

**왜 이 4개 우선:** §3 의 미구현 stage 중 의사결정 권한자가 Architect/Tech Lead 이고 시간 지평이 다년인 stage. 잘못된 결정의 비용이 부수 design-* 보다 압도적으로 큼. design-system orchestrator 가 이미 존재하므로 stage 만 채우면 즉시 활용 가능.

## Out of scope

- §3 의 부수 design-* (`design-tenant-model`, `design-i18n-strategy`, `design-accessibility-baseline` 등 11개) — 별도 plan
- §3 의 use case → infra 브릿지 (`map-use-cases-to-infra`, `derive-system-topology`) — Phase 1' 또는 별도 plan
- design-system orchestrator 자체 재작성 (기존 PROCEDURE.md 유지, stage 매핑만 추가)
- MCP / 외부 SaaS 통합

## 진행 원칙

- **순차 작성** (parallel 아님): 첫 skill (`define-tech-stack`) 작성 시 PROCEDURE.md template 정착 → 후속 3 skill 이 그 template 재사용. parallel 로 가면 4개가 다른 형식으로 갈 위험.
- **각 skill = 단일 commit** (per task): `feat(skill): add <name> stage skill` 메시지. Phase 1 plan 의 모든 task 완료 후 version bump (1.0.1 → 1.0.2) + release commit 으로 마무리.
- **Acceptance criteria 통일**: 각 skill 이 동일 기준 통과해야 함 (§ Acceptance template 참조).

---

## PROCEDURE.md 공통 템플릿

> **Reference 학습 적용** (2026-05-07): 4 reference repo (`skill/superpowers`, `skill/awesome-claude-skills`, `harness/everything-claude-code`, `system-prompt/.../claude-opus-4.7.md`) 분석 결과 다음 4 enhancement 가 template 에 반영됨:
> - **A. Anti-slop opener (§0)** — superpowers/AGENTS.md "94% PR 거절률" 패턴. skill 시작에서 명시적 STOP gate.
> - **B. Explicit output structure (§6)** — Opus 4.7 default = prose minimal formatting. skill 이 explicit 하게 table/structured 요구해야 모델이 구조화 출력함.
> - **C. Engineering posture (§4)** — review-engineering 류의 "input 자체를 challenge, 입장 취함, hedge 금지" posture 명시. AI sycophancy 방지.
> - **D. Verification gate (§11)** — verification-before-completion 패턴. 완료 선언 전 self-check checklist.

새 stage skill 의 `plugin/skills/<name>/PROCEDURE.md` 는 **frontmatter 없이** 다음 구조:

```markdown
# <skill-name> — <한 줄 부제>

## 0. STOP — 시작 전 읽기

이 skill 은 <가장 흔한 실패 모드> 를 방지하기 위한 절차다. 다음 anti-pattern 이 발견되면 즉시 중단:
- <anti-pattern 1: 예 "사용자 입력을 그대로 stack 결정으로 채택">
- <anti-pattern 2: 예 "alternatives 검토 없이 단일 후보만 평가">
- <anti-pattern 3: 예 "lock-in cost 정량 평가 생략">

§5 의 모든 단계를 누락 없이 수행하라. skip 시 산출물의 신뢰도가 무너진다.

## 1. 목적

<문제 정의 + 이 skill 이 해결하는 의사결정 누수 1~3 단락>

## 2. 사용 시점 (When to invoke)

- <트리거 조건 1>
- <트리거 조건 2>

## 3. 입력 (Inputs)

### 필수
- <required input 1>

### 선택
- <optional input 1>

### 입력이 부족할 때 forcing question
- "<질문 1>"

## 4. 핵심 원칙 (Principles + Posture)

이 skill 의 운영 posture:
- **입장 취함, hedge 금지** — 모든 결정에 추천 + 근거 + 변경 조건 명시. "둘 다 가능" 같은 fence-sitting 금지.
- **사용자 입력을 challenge** — 모호하거나 evidence 부족하면 forcing question 으로 push back.
- **Specificity 강제** — "fast", "scalable" 같은 카테고리 답변 거부, 숫자·조건·검증 방법 요구.

도메인 원칙:
1. <원칙 1>
2. <원칙 2>

## 5. 단계 (Phases)

### Phase 1. <단계명>
1. <step>

### Phase 2. <단계명>
1. <step>

## 6. 산출물 형식 (Output format)

> **Note**: Opus 4.7 / Sonnet 4.6 의 default 는 prose 출력. 다음 구조를 **명시적으로 요구**해야 모델이 structured 출력함.

다음 형식으로 출력하라 (요약 / prose 변환 금지, 모든 섹션 채우기 강제):

\`\`\`markdown
## <skill-name> Output

### Summary
<핵심 3 줄 요약>

### Decisions Table
| Dimension | Selected | Alternatives Considered | Rationale | Lock-in Cost |
|-----------|----------|-------------------------|-----------|--------------|
| ... | ... | ... | ... | ... |

### Risk Register
| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| ... | ... | ... | ... |

### Next Step
<구체 action — 1줄>
\`\`\`

## 7. Cross-phase cascade

이 skill 의 산출물이 다음 phase 의 어떤 input 으로 흘러가는지:
- §<N>: <어떤 입력으로>

## 8. 다음 skill (next in stage flow)

- <후속 skill name> — <왜 다음이어야>

## 9. 다른 skill 과의 경계 (충돌 방지)

- vs `<adjacent skill>`: <어떻게 다른지 1줄>

## 10. 중요 규칙

- <invariant 1>

## 11. Verification gate — 완료 선언 전 self-check

다음 체크가 모두 yes 여야 절차 완료 보고:
- [ ] §5 의 모든 phase 가 누락 없이 실행됨
- [ ] §6 의 output 섹션이 모두 채워졌고 prose 로 변환되지 않음
- [ ] §4 의 posture (입장 / specificity / challenge) 가 적용됨 — hedge 표현 없음
- [ ] §0 의 anti-pattern 들이 산출물에 등장하지 않음
- [ ] (skill-specific 검증 항목 추가)

하나라도 no 면 해당 단계로 돌아가 보강 후 재검증. user 에게 incomplete 산출물을 "충분하다" 고 보고하지 말 것.
```

각 section 의 **§9 (충돌 방지)** 가 핵심 — router 가 fuzzy 매칭 시 이 skill 이 다른 skill 보다 우선해야 할 케이스를 명시. **§0 (STOP) + §11 (Verification)** 는 reference 학습 기반 신규 — anti-slop 강제.

## Command md template (slash command 노출)

각 skill 마다 `plugin/commands/<name>.md` 도 함께 작성 (기존 stage skill 의 dual-mode 컨벤션 따름):

```markdown
---
description: <skill 의 1줄 plain-language 설명>
argument-hint: "<input 형태 hint>"
---

# /buddy:<name>

<description 동일 또는 약간 부연>

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `<name>`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
```

---

## Tasks

### Task 1.1: `define-tech-stack` (template skill)

**Files:**
- Create: `plugin/skills/define-tech-stack/PROCEDURE.md`
- Create: `plugin/commands/define-tech-stack.md`
- Modify: `plugin/skills/router/references/skill-catalog.md` (§3 표 행 추가)
- Modify: `plugin/skills/design-system/PROCEDURE.md` (stage 흐름에 매핑)

**Skill 정의:**
- description: "기술 스택 결정 — 언어 / 프레임워크 / DB / runtime / hosting 선택. 락인 영향 큰 의사결정을 evidence-based 로 기록"
- when-to-use: PRD 확정 후 첫 §3 진입 / 신규 service 추가 / 기술 부채 누적으로 stack 재평가
- inputs: feature backlog, team 보유 expertise, 운영 환경 제약, budget, latency/throughput SLO
- outputs: tech stack ADR + rationale matrix (대안 vs 선택지) + 락인 risk register
- 충돌 방지: vs `consult-design-system` (UI design system 결정, 다른 영역) / vs `design-mcp-server` (specific MCP 설계, narrower)

**PROCEDURE 본문 골자 (§5 Phases):**
1. **Stack 차원 분해** — language / framework / DB / cache / queue / hosting / observability / CI 7+ 차원
2. **각 차원별 후보 3개 + scoring** — fit, maturity, hiring pool, cost, lock-in 5축
3. **Cross-차원 호환성 매트릭스** — 어느 조합이 충돌 / 시너지
4. **선택 + ADR 작성** (별도 `write-adr` 호출)
5. **5년 lock-in risk** — migration cost 정량 예상

**Acceptance:**
- PROCEDURE.md 존재, frontmatter 없음, body 200~400 lines
- skill-catalog.md §3 Stage Skills 표에 row 추가됨
- design-system/PROCEDURE.md 의 stage 흐름에 `define-tech-stack` 매핑됨
- command md 가 router single-mode dispatch 호출
- `bash scripts/test-router-wireup.sh` 11/11 통과 (PROCEDURE 카운트 78 → 79, command 카운트 30 → 31)
- CI script 의 expected counts 갱신 commit 동반

**Commit:** `feat(skill): add define-tech-stack §3 stage skill`

### Task 1.2: `design-data-model`

**Files:**
- Create: `plugin/skills/design-data-model/PROCEDURE.md`
- Create: `plugin/commands/design-data-model.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/design-system/PROCEDURE.md`

**Skill 정의:**
- description: "데이터 모델 설계 — 스키마, 마이그레이션 전략, 인덱싱, 정규화/denormalization 결정. production 운영 변경 비용을 사전 평가"
- when-to-use: tech stack 확정 후 / 신규 entity 추가 / read pattern 변화로 인덱스/스키마 재설계
- inputs: tech stack ADR, entity-relationship 초안, expected read/write 패턴, scale (rows/sec, GB)
- outputs: schema DDL draft, migration plan, index 전략, ER diagram, hot path query plan
- 충돌 방지: vs `define-tech-stack` (DB 선택은 거기, 스키마는 여기) / vs `design-api-contract` (API resource 모양은 거기)

**PROCEDURE 본문 골자:**
1. **Entity 식별 + 관계 매핑**
2. **Read/write 패턴 분류** (OLTP vs analytical, hot path vs cold)
3. **Normalization decision** (3NF vs denormalized for read perf)
4. **Index 전략** (어느 column, partial index, covering index)
5. **Migration 호환성** (zero-downtime / forward-compatible 강제)
6. **ADR 작성** (별도 `write-adr` 호출)

**Acceptance:** Task 1.1 동일 패턴, 단 PROCEDURE/command count 79→80, 31→32

**Commit:** `feat(skill): add design-data-model §3 stage skill`

### Task 1.3: `design-api-contract`

**Files:** (Task 1.1 동일 구조 — 4 files)

**Skill 정의:**
- description: "API 계약 설계 — REST/GraphQL/RPC 선택 + endpoint/schema 정의. actor 간 경계 = API 경계 원칙으로 contract-first 강제"
- when-to-use: tech stack + data model 확정 후 / 외부 통합 추가 / breaking change 감지
- inputs: actor list (use case 매핑), data model schema, expected client (web/mobile/3rd-party), versioning 정책
- outputs: OpenAPI / GraphQL SDL / proto file, error taxonomy, versioning 가이드
- 충돌 방지: vs `design-event-schema` (async event, 별도 skill — 부재시 본 skill 이 broader 다룸) / vs `design-data-model` (DB 모양 vs API 모양 분리)

**PROCEDURE 본문 골자:**
1. **Style 선택** — REST / GraphQL / gRPC / hybrid (use case 별 적합도 매트릭스)
2. **Resource / operation 매핑** — actor use case → endpoint
3. **Schema 정의** — request/response, error format
4. **Versioning + deprecation 정책**
5. **Contract test 전략** — consumer-driven contract 또는 schema validation
6. **ADR 작성**

**Acceptance:** PROCEDURE/command count 80→81, 32→33

**Commit:** `feat(skill): add design-api-contract §3 stage skill`

### Task 1.4: `write-adr`

**Files:** (Task 1.1 동일 구조 — 4 files)

**Skill 정의:**
- description: "Architecture Decision Record 작성 — context / decision / consequences 표준 양식으로 의사결정 보존. tech-stack / data-model / api-contract 등 의사결정의 근거 영속화"
- when-to-use: 위 3 skill 이 의사결정 산출 시 마무리 단계 / 기존 결정 변경 (supersede) 시 / 외부 결정 inheritance 시
- inputs: 의사결정 context, 검토한 alternatives, 선택지, 결정자, 영향 범위
- outputs: ADR markdown (`docs/adr/NNNN-<slug>.md`), index update, supersede 체인 갱신
- 충돌 방지: vs `define-tech-stack`/`design-data-model`/`design-api-contract` (그들은 의사결정 process, 본 skill 은 보존 format) — chain 으로 활용 권장

**PROCEDURE 본문 골자:**
1. **ADR id 결정** (next sequential)
2. **표준 섹션 작성** — Title / Status (proposed/accepted/deprecated/superseded) / Context / Decision / Consequences (positive + negative + neutral) / Alternatives Considered
3. **Index update** — `docs/adr/README.md` 또는 `INDEX.md` 에 row 추가
4. **Supersede 체인** — 기존 ADR 대체 시 양방향 link
5. **Validation** — Status, Context, Decision, Consequences 필수 섹션 존재 확인

**Acceptance:** PROCEDURE/command count 81→82, 33→34

**Commit:** `feat(skill): add write-adr §3 stage skill (ADR boilerplate)`

### Task 1.5: design-system orchestrator stage 흐름 갱신

**Files:**
- Modify: `plugin/skills/design-system/PROCEDURE.md`

**작업:**
- 현재 design-system PROCEDURE 의 stage 흐름 섹션에 신규 4 skill 매핑 추가
- 기존 review-architecture / explore-design-variants / consult-codex 같은 보유 skill 과의 순서 정의
- 권장 chain 패턴: `define-tech-stack → design-data-model → design-api-contract` (의존 순서) + 각 단계 끝에 `write-adr` 호출

**Acceptance:**
- design-system PROCEDURE 에 새 4 skill 이 stage 표 또는 흐름 다이어그램으로 명시
- chain 권장 syntax 예시 포함 (`/buddy:chain define-tech-stack,design-data-model,design-api-contract -- "<feature>"`)

**Commit:** `feat(skill): wire 4 new §3 stages into design-system orchestrator flow`

### Task 1.6: CI script invariant 갱신 + version release

**Files:**
- Modify: `scripts/test-router-wireup.sh` (PROCEDURE/command count expected values)
- Modify: `plugin/.claude-plugin/plugin.json` (version 1.0.1 → 1.0.2)
- Modify: `.claude-plugin/marketplace.json` (version sync)
- Modify: `CHANGELOG.md` (1.0.2 entry)

**작업:**
- `procedure_count` 검증을 78 → 82 로 갱신
- `command_md_count` 검증 lower bound 가 30 이상이라 변경 불필요하지만 comment 갱신 권장
- Version bump + CHANGELOG entry 작성:
  - Added: 4 §3 stage skills (define-tech-stack, design-data-model, design-api-contract, write-adr)
  - Changed: design-system orchestrator stage 흐름에 신규 4 skill 매핑

**Acceptance:**
- `bash scripts/test-router-wireup.sh` 통과 (10/10 또는 추가 invariant 시 11/11)
- `claude plugin validate plugin/` 통과
- CHANGELOG 1.0.2 entry 작성

**Commit:** `release: v1.0.2 (§3 design stages)`

---

## Verification (Phase 1 종료 시점)

자동 (CI):
- `bash scripts/test-router-wireup.sh` 통과
- `claude plugin validate plugin/` 통과
- PROCEDURE.md count = 82
- 4 신규 skill 파일이 모두 frontmatter 없이 작성됨

수동 (live test, 사용자 수행):
- `claude plugin marketplace update buddy && claude plugin update buddy@buddy` (1.0.2 fetch)
- 4 신규 슬래시 호출:
  - `/buddy:define-tech-stack "<test feature>"`
  - `/buddy:design-data-model "<test entity>"`
  - `/buddy:design-api-contract "<test endpoint>"`
  - `/buddy:write-adr "<test decision>"`
- chain test: `/buddy:chain define-tech-stack,design-data-model,design-api-contract -- "test feature"`

각 호출이 router → PROCEDURE Read → 절차 시작 흐름 완성하는지 확인.

문서 검증:
- spec `2026-05-06-lifecycle-orchestrator-architecture.md` §3 의 status 표를 🟡 → ✅ 로 갱신 (별도 commit)

---

## Rollback

각 Task 가 독립 commit 이라 phase 단위 / task 단위 `git revert` 가능. 4 skill 모두 신규 파일이고 기존 파일 수정은 design-system/PROCEDURE.md 와 catalog/CI/CHANGELOG 만이라 영향 범위 좁음.

---

## 다음 plan (Phase 2)

Phase 1 통과 후:
- Phase 2 (`§4 Implementation Plan stage 6 skills`) plan 작성
- 대상: `decompose-feature-to-actor-tracks`, `decompose-track-to-tasks`, `map-task-dependencies`, `plan-parallel-execution`, `define-acceptance-test-plan`, `estimate-build-timeline`

본 Phase 1 의 PROCEDURE.md template 가 Phase 2 의 작성 비용을 크게 줄여줄 것 (예상).

---

## Reference repos 학습 — 별도 plan 후보 (Phase 1 외)

본 Phase 1 plan 에는 PROCEDURE.md template 강화로 4 enhancement (anti-slop, output structure, posture, verification gate) 만 반영. 더 큰 reference 학습은 별도 plan 으로 분리:

| Enhancement | Source | 영향 범위 | Plan 우선순위 |
|------------|--------|----------|--------------|
| **plugin/agents/ 채우기** | `harness/everything-claude-code/agents/` (10+ specialized agents — architect, code-explorer, code-reviewer 등) | buddy 전체 | High — `plugin/agents/` 가 비어있음. agent 도구가 buddy 내부에서도 활용 가능 |
| **Context modes (dev/research/review)** | `harness/everything-claude-code/contexts/` | buddy 전체 | Mid — phase orchestrator 가 context mode 와 결합되면 더 정밀한 동작 |
| **AGENTS.md anti-slop guide** | `skill/superpowers/AGENTS.md` | buddy contributor 가이드 | Mid — buddy 자체에 contribute 하는 AI agent 들의 품질 강제 |
| **Process skills missing in buddy** | `skill/superpowers/skills/` (writing-plans, verification-before-completion 등 14 process skill 비교) | buddy stage skill catalog | Low — 일부는 buddy 가 다른 형태로 보유, 직접 매핑은 검증 필요 |

이 항목들은 Phase 1 완료 후 별도 plan 으로 평가·작성.
