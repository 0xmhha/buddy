# N-1 Audit — buddy plugin skill context bloat

> **Date:** 2026-05-09
> **Plugin version:** v1.0.8 (HEAD `2e94e04`)
> **Trigger:** handoff §2 (`docs/notes/2026-05-08-handoff-context-bloat-investigation.md`)
> **Goal:** 57개 `plugin/commands/*.md` 의 frontmatter description / body 비용을 측정하고 indexing 개선 quick win 을 식별.

## 1. Summary

| 지표 | 값 |
|------|-----|
| Commands 총 개수 | **57** |
| Description chars 합 | **4,887** (avg 85.7, min 35, max 176) |
| Description UTF-8 bytes 합 | **6,803** |
| `argument-hint` chars 합 | **1,867** (avg 32.8, max 64) |
| commands/*.md 전체 파일 bytes | **32,128** (avg ~564/file) |
| Body boilerplate 균질도 | **57 / 57** (100%) — 모두 동일한 router invocation pattern |

**Per-prompt 추정 baseline cost (Korean+English 혼합, 1 token ≈ 2.5 chars):**

| 시나리오 | Frontmatter only | Full body |
|---------|-----------------|-----------|
| 절대 chars | ~8,500 | ~32,128 |
| 추정 tokens | **~3,400** | **~10,700** |

→ 직전 핸드오프의 "1,500~3,000 token 절감" 추정은 **description-only 단축 가정**에서만 성립. body 도 baseline 에 포함된다면 절감 여지는 약 **5배**.

## 2. Distribution

| Description 길이 | 파일 수 | 비율 |
|----------------|--------|------|
| < 50 chars | 5 | 8.8% |
| 50–79 chars | 23 | 40.4% |
| 80–109 chars | 14 | 24.6% |
| 110–139 chars | 12 | 21.1% |
| ≥ 140 chars | **3** | 5.3% |

상위 3건이 전체 description 길이 합의 **9.6%** (467/4887) 를 차지 — Pareto outlier.

## 3. Top 10 longest descriptions (단축 우선 후보)

| # | Command | Chars | Description |
|---|---------|------:|-------------|
| 1 | map-use-cases-to-infra | **176** | actor system boundary → infra component 매핑 — Q8=(a) cascade 의 §2→§3 transition layer. 산출물은 actor × infra bidirectional matrix + cross-actor shared ownership + compliance scope. |
| 2 | audit-cost-efficiency | 147 | Infracost monthly + per-component breakdown + unit economics ($/MAU) + waste detection + savings recommendation (RI / Savings Plan / right-sizing). |
| 3 | test-cross-actor-flow | 144 | cross-actor flow E2E test — multi-actor chain (signup→email→verify→login 등) full-stack 검증. cross-actor edge coverage + contract drift detection. |
| 4 | design-event-schema | 130 | async event schema-first 설계 — producer/consumer contract + versioning + DLQ + idempotency. design-api-contract 의 sync-only gap 보강. |
| 5 | design-tenant-model | 128 | multi-tenant 격리 전략 — shared (RLS) vs schema-per vs DB-per 3 모델 trade-off + 3 layer defense + onboarding cost + compliance scope. |
| 6 | map-task-dependencies | 124 | task DAG 작성 — internal (intra-track) + cross-actor (contract-based) edges + cycle 감지 + critical path + parallel-safe levels. |
| 7 | prepare-launch-checklist | 123 | launch readiness 17+ 항목 gate (6 axis - engineering/security/ops/product/legal/cost) — GA 직전 cross-functional final check. |
| 8 | decompose-track-to-tasks | 122 | actor track → ordered task list (atomic units, single PR scope, verifiable acceptance, sized for 1 worker × short period). |
| 9 | test-per-actor-use-case | 121 | actor 의 use case 단위 통합 테스트 — frontend(E2E), backend(integration), 3rd-party(contract). per-actor coverage gap 0 maintain. |
| 10 | write-adr | 120 | Architecture Decision Record 작성 — context/decision/consequences/alternatives 표준 양식으로 의사결정 영속화 + supersede 체인 + Index 갱신. |

## 4. Findings

### 4.1 Description = "lifted PROCEDURE summary" anti-pattern

상위 10개를 보면 description 이 **router 가 dispatch 결정에 필요한 trigger keyword** 가 아니라 **PROCEDURE 산출물 요약문** 으로 작성됨. 예:

- 현재: `actor system boundary → infra component 매핑 — Q8=(a) cascade 의 §2→§3 transition layer. 산출물은 actor × infra bidirectional matrix + cross-actor shared ownership + compliance scope.` (176 chars)
- 의도: 사용자/router 가 "이 명령이 무엇인지" 1초 안에 식별
- 실제 효과: PROCEDURE 본문에서 fetch 해야 할 산출물 명세를 description 에 inline 해 매 prompt 마다 비용 부담

### 4.2 Body 100% 보일러플레이트 균질

57/57 파일이 동일한 패턴:

```markdown
---
description: <…>
argument-hint: <…>
---

# /buddy:<name>

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `<name>`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
```

→ Body 자체는 일괄 변환 가능 (sed 1회).

### 4.3 Frontmatter 가 Claude Code 에 어떻게 들어가는가 (open question)

이번 audit 으로는 측정 불가 — Claude Code 의 plugin discovery semantics:

1. (가설 A) frontmatter 의 `description` + `argument-hint` 만 인덱스 → 매 prompt 비용 ≈ **3,400 token**
2. (가설 B) 파일 전체가 항상 로드 → 매 prompt 비용 ≈ **10,700 token**
3. (가설 C) 이름·description 만 인덱스, body 는 invocation 시 lazy load → 매 prompt 비용 ≈ **2,000 token** (description only)

**검증 방법 (다음 액션):**
- WebFetch 로 Claude Code plugin spec 문서 확인
- 또는 controlled experiment: 빈 plugin vs 57-command plugin 의 같은 prompt 의 system token 차이 측정

## 5. Quick win 제안 (effort × 절감 견적)

### 5.1 Quick Win A — Top 10 description 단축 (Low effort, Low risk)

상위 10개를 router trigger 에 필요한 최소 keyword 로 단축:

| Command | Before (chars) | After 후보 (chars) | 절감 |
|---------|--------------:|------------------:|----:|
| map-use-cases-to-infra | 176 | `§3 cascade: actor → infra mapping (matrix + ownership + compliance)` (62) | -114 |
| audit-cost-efficiency | 147 | `§6 audit: $/MAU + waste + RI/Savings Plan recommendations` (54) | -93 |
| test-cross-actor-flow | 144 | `§6 test: multi-actor E2E chain + contract drift detection` (54) | -90 |
| design-event-schema | 130 | `§3 design: async event schema (producer/consumer + DLQ + idempotency)` (66) | -64 |
| design-tenant-model | 128 | `§3 design: multi-tenant isolation (RLS / schema / DB) + defense layers` (68) | -60 |
| map-task-dependencies | 124 | `§4 plan: task DAG + cycle detection + critical path + parallel levels` (66) | -58 |
| prepare-launch-checklist | 123 | `§7 release: GA gate — 17 items × 6 axes (eng/sec/ops/prod/legal/cost)` (66) | -57 |
| decompose-track-to-tasks | 122 | `§4 plan: actor track → atomic ordered tasks (1 PR / 1 worker scope)` (62) | -60 |
| test-per-actor-use-case | 121 | `§6 test: per-actor use case (E2E + integration + contract)` (56) | -65 |
| write-adr | 120 | `§3 design: ADR (context/decision/consequences) + supersede chain` (62) | -58 |
| **Total** | **1,335** | **~616** | **-719 chars (~290 tokens)** |

### 5.2 Quick Win B — 전체 description 일괄 단축 + 구조화 (Medium effort)

Phase prefix (`§N <category>: ...`) 컨벤션을 57개 모두 적용:

- 평균 85.7 → ~50 chars (음의 outlier 5개 제외)
- 절감 추정: (85.7-50) × 52 ≈ -1,860 chars ≈ **~750 token / prompt**

### 5.3 Quick Win C — Body slim (Medium effort, Architectural)

57/57 동일 보일러플레이트 → 단일 형식으로 정리:

- 현재 평균 body: ~445 bytes/file × 57 = ~25,300 bytes
- 후보 body: 1줄 (`/buddy:router single <name> -- $ARGUMENTS` 같은 thin invocation)
- 절감: ~24,000 bytes ≈ **~8,000 token / prompt** (단, **본문이 매 prompt 마다 로드되는 경우에만 의미**)

→ Quick win C 의 ROI 는 §4.3 가설 검증에 종속.

## 6. Decision tree (다음 액션)

```
가설 검증 (§4.3)
├─ A or C 입증 (description-only 인덱싱)
│   → Quick Win A or B 만 실행 (Low risk, ~290~750 token 절감)
│   → C 는 cosmetic 개선 수준
└─ B 입증 (body 도 baseline)
    → Quick Win A + B + C 모두 실행 (~9,000+ token 절감)
    → ADR 로 buddy plugin 의 "thin command stub" 컨벤션 영속화
```

## 7. Recommended next steps

1. **WebFetch + 실측** — Claude Code plugin spec 확인. baseline 가설 결정.
2. **Quick Win A 시안 PR** — 상위 10개 description 단축. 효과 측정 baseline 확보.
3. **결정 영속화** — gap 결정 후 ADR (`docs/superpowers/decisions/`) 작성.
4. **Future** — Quick Win B/C 적용은 가설 검증 결과에 따라 분기.

## 8. Fact-based summary

<Fact-based Answer>
- **Fact:**
  - 57 commands, description 합 4,887 chars (UTF-8 6,803 bytes)
  - argument-hint 합 1,867 chars
  - commands/*.md 전체 파일 합 32,128 bytes
  - 57/57 파일이 동일 router invocation 보일러플레이트 사용
  - description ≥140 chars 인 outlier 3건이 전체의 9.6% 를 차지

- **Your Opinion:**
  - **High prediction:** Body slim (Quick Win C) 의 절감 효과는 가설 §4.3 결정에 5배 의존 — 측정 우선 필요
  - **Mid prediction:** Top 10 description 단축은 ~290 token 절감, 1시간 effort 의 deterministic quick win
  - **Low prediction:** Claude Code 가 frontmatter 를 fully indexed 형태가 아니라 hash-only 또는 prefix-only 로 처리할 가능성 — 그 경우 description 길이가 의미 없어짐
  - **None.**
</Fact-based Answer>
