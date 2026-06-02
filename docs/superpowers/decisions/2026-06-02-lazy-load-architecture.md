# ADR-020 — Lazy-Load Architecture: 4-Layer skill loading

| 항목 | 값 |
|------|----|
| **Status** | Accepted |
| **Date** | 2026-06-02 |
| **Deciders** | project owner |
| **Tags** | architecture, lazy-load, token-cost, skill-conflict-prevention, frontmatter, catalog-split, lifecycle-isolation |
| **Related** | [ADR-001](./2026-05-09-buddy-commands-disable-model-invocation.md) (commands disable-model-invocation), [ADR-007](./2026-05-19-router-no-cross-invocation-state.md) (router stateless), [ADR-018](./2026-05-21-verify-best-alternative-and-bias-prevention.md) (write-a-skill dogfood), `docs/se-lifecycle-naming.md`, `plugin/skills/write-a-skill/references/naming-convention.md` |

---

## Context

buddy는 약 150개의 PROCEDURE.md 형태로 작성된 skill을 9-phase 라이프사이클 + Cross-cutting 영역으로 조직한다. 사용자가 진입점(`/buddy:start` 등)을 호출하면 router가 적절한 skill을 dispatch해 작업을 진행한다. 이 구조에서 다음 3가지 문제가 발생한다:

### 1. 토큰 비대화 (Token Cost)

Anthropic Claude Code 표준은 모든 `SKILL.md`의 frontmatter (`description` + `when_to_use`)를 **세션 시작 시 자동 listing context에 로드**한다. 150 skill × 평균 500자 = 약 19,500 토큰의 listing이 매 세션마다 상시 점유된다 (200K 컨텍스트의 ~9.75%). default listing budget (컨텍스트의 1%)을 초과하므로 일부 skill description이 자동으로 dropped되어 dispatch 정확도가 저하된다.

### 2. Lost-in-the-Middle

비대한 listing context 내부에서도 lost-in-the-middle 현상이 발생한다. 중간 위치의 skill description이 정확히 매칭되지 않거나 일부가 drop된다. 본문 차원에서도 한 PROCEDURE.md가 500줄을 초과하면 중간 절차가 약하게 처리된다.

### 3. 유사 패턴 Skill 동시 로드 시 충돌

다음과 같이 표면적으로 유사하나 결정 방식·전개 방식이 다른 skill 그룹이 존재한다:

- **신규 아이디어 브레인스토밍** (문제·기회 검증 단계): 현실 가능성을 따지지 않고 발산·분류·deep dive
- **버그 수정 방안 모색** (개발 단계): 각 수정안의 현실 가능성·장단점·의사결정 조건 적용

두 skill이 동시에 컨텍스트에 있으면 dispatch가 모호해지고, 사용자 의도와 다른 처리 방식이 적용될 수 있다. 단순한 description 차이로 분리되지 않으며, **각 작업 단계의 결정 기준 자체가 격리되어야** 충돌이 방지된다.

### 기존 buddy 패턴의 한계

현재 buddy는 router/SKILL.md 1개만 frontmatter를 가지고 나머지는 catalog (`skill-catalog.md`) entry로 dispatch description을 대신한다. 이는 토큰 효율은 높으나:

- Anthropic 공식 표준에서 벗어남 (호환성·외부 도구 지원 약함)
- catalog 단일 파일이 ~280줄로 비대해지면서 lazy-load 시점에도 lost-in-the-middle 위험 존재
- catalog entry 형식의 평가 기준이 비표준 (frontmatter 평가 F1-F5 적용 불가)

---

## Decision

**4-Layer Lazy-Load Architecture**를 채택한다.

### Layer 0 — Router Skill (자동 로드)

| 항목 | 내용 |
|------|------|
| 파일 | `plugin/skills/router/SKILL.md` |
| 로드 시점 | 세션 시작 시 자동 (Claude Code 표준 SKILL.md 스캔) |
| 로드 분량 | frontmatter ~180자 |
| 역할 | 진입 라우터. command 호출 시 어디로 dispatch할지 결정 |

### Layer 1 — Phase Catalog (Phase별 lazy-load)

| 항목 | 내용 |
|------|------|
| 파일 | `plugin/skills/router/references/catalog-<phase>.md` (예: `catalog-phase-1.md`, `catalog-cross-cutting.md`) |
| 분할 단위 | 9 phase + Cross-cutting = 10 파일 |
| 로드 시점 | 사용자 작업이 특정 phase에 해당함이 식별된 후 router가 명시적 Read |
| 로드 분량 | phase당 ~3,000자 (15 skill × 평균 200자 요약) |
| 역할 | 해당 phase의 skill 인덱스 + 각 skill의 frontmatter 위치 (파일 경로 + 라인 범위) 안내 |

### Layer 2 — PROCEDURE Frontmatter (Skill별 lazy-load)

| 항목 | 내용 |
|------|------|
| 파일 | `plugin/skills/<name>/PROCEDURE.md` 의 첫 N줄 (frontmatter 영역만) |
| 로드 시점 | Layer 1 catalog가 안내한 위치를 router가 명시적 Read (`offset`/`limit` 파라미터로 라인 범위 지정) |
| 로드 분량 | skill당 ~500-800자 (description + when_to_use 합산, 1,536자 cap) |
| 역할 | 해당 skill의 정확한 trigger·use case·persona·side-effect 정보로 최종 dispatch 결정 |
| 형식 | Anthropic 공식 frontmatter 규격 준수 (호환성 + 외부 도구 지원) |

### Layer 3 — PROCEDURE Body (Full lazy-load)

| 항목 | 내용 |
|------|------|
| 파일 | 선택된 skill의 `PROCEDURE.md` 본문 전체 |
| 로드 시점 | router가 최종 dispatch 후 본문 실행 시점 |
| 로드 분량 | skill당 평균 ~5,000-10,000자 (가이드 ≤ 500줄 권장) |
| 역할 | 실행 절차 본체 (Step 1-N, 출력 형식, 검증 체크리스트 등) |

### 토큰 예산 비교

| 시나리오 | 평소 로드 | 작업 시 최대 로드 | 비고 |
|---------|----------|----------------|------|
| 현재 buddy (router + catalog) | router 1줄 (~180자) | + catalog ~12K 토큰 (dispatch 모호 시) | 단 표준 호환성 약함 |
| 시나리오 A (모든 frontmatter 자동 listing) | ~19K 토큰 상시 점유 | 추가 본문 ~5K 토큰 | 표준 호환 + 토큰 비효율 + lost-in-middle |
| **본 결정 (4-Layer)** | router 1줄 (~180자) | + Layer 1 ~3K자 + Layer 2 ~1.5K자 + Layer 3 ~6K자 = **~10.7K자 (~2.7K 토큰)** | 표준 호환 + 토큰 효율 + phase 격리 |

본 결정의 작업 시 최대 로드는 1M 컨텍스트의 0.27% 수준. 시나리오 A의 1/7 효율.

---

## Consequences

### 긍정적 영향

1. **토큰 효율**: 평소 router frontmatter 1개 (~180자)만 로드. 작업 시점에만 phase-scoped catalog + 선택된 skill의 frontmatter + 본문 lazy-load.
2. **Phase 격리**: ideation 작업 시 engineering 단계의 skill frontmatter는 로드되지 않음. 유사 패턴 skill 간 충돌 구조적 방지.
3. **Anthropic 표준 호환**: PROCEDURE.md에 공식 frontmatter 추가로 외부 도구·평가 시스템·향후 표준화 변화 대응 가능.
4. **Lost-in-middle 완화**: 각 Layer가 작은 단위로 분할되어 컨텍스트 내 attention 집중도 향상.
5. **`evaluate-skill` F1-F5 적용 가능**: PROCEDURE.md frontmatter 자체 평가 가능해짐.

### 부정적 영향 / Trade-off

1. **마이그레이션 비용**: ~150 PROCEDURE.md 모두에 frontmatter 추가 + catalog 분할 작업 필요. Phase 1부터 점진 진행 계획.
2. **2-Tier catalog 메커니즘 복잡도**: Layer 1 (catalog 분할) + Layer 2 (frontmatter 라인 범위 지정 로드) 두 단계. router는 이를 일관 처리해야 함.
3. **catalog와 frontmatter의 SSoT 중복 가능성**: catalog 요약과 frontmatter description이 중복될 위험. catalog는 **요약 + 라인 위치 안내** 역할로 한정하여 SSoT 충돌 방지.
4. **Phase 경계 케이스 처리 부담**: 단계 전환 시 이전 phase 결과를 input으로 다음 phase 시작 (context 정리 또는 서브에이전트 분리). orchestration 정책 별도 정의 필요 (C4 작업).

### 영향 받는 시스템

- 모든 PROCEDURE.md (frontmatter 추가)
- `plugin/skills/router/SKILL.md` ("라우팅 결정 워크플로우" 4-Layer 로 갱신)
- `plugin/skills/router/references/skill-catalog.md` (10개 파일로 분할 또는 deprecate)
- `plugin/skills/router/references/routing-rules.md` (Layer 1 catalog 참조 갱신)
- `plugin/skills/evaluate-skill/PROCEDURE.md` (F1-F5 PROCEDURE.md에 적용, CE1-CE5 폐지/축소)
- `plugin/skills/write-a-skill/PROCEDURE.md` (frontmatter 추가를 신규 skill 작성 절차에 포함)
- `docs/plugin-skills-authoring-guide.md` (§0 신규 단원으로 4-Layer 아키텍처 명시)

---

## Alternatives Considered

### Alternative A: 모든 PROCEDURE에 frontmatter 추가 + Claude Code 자동 listing 활용

**거부 이유**: Listing context에 ~19K 토큰 상시 점유. default budget (컨텍스트의 1%) 9-10배 초과 → 자동 truncation → 일부 skill description dropped → dispatch 정확도 저하. 또한 비대한 listing 내부에서 lost-in-middle 발생.

### Alternative B: 현재 buddy 패턴 유지 (catalog만, frontmatter 없음)

**거부 이유**: Anthropic 공식 표준 호환성 약함. catalog 단일 파일 ~280줄 비대화. evaluate-skill의 F1-F5 항목을 PROCEDURE.md에 적용 불가. 미래 외부 도구·표준화 대응 어려움.

### Alternative C: 자주 호출되는 일부 skill만 frontmatter 추가 (hybrid)

**거부 이유**: "자주" 기준 결정 부담. 혼합 모델로 일관성 손상. evaluate-skill 케이스 분기 복잡도 증가.

### Alternative D: catalog 분할 없이 frontmatter만 추가

**거부 이유**: PROCEDURE.md frontmatter는 buddy 명명상 자동 listing 대상이 아니므로 (Claude Code는 `SKILL.md`만 자동 스캔) listing 부담은 0. 그러나 catalog 분할이 없으면 dispatch 시 ~280줄 catalog 전체를 매번 lazy-load하는 부담 + phase 격리 효과 없음.

---

## Verification

### 단계별 검증

| 단계 | 검증 방법 |
|------|---------|
| Layer 0 (router) | 세션 시작 시 router frontmatter만 listing context에 있는지 (다른 PROCEDURE는 listing 없음) `claude --debug` 또는 token 사용량 측정으로 확인 |
| Layer 1 (catalog 분할) | router/SKILL.md "라우팅 결정 워크플로우"가 phase 식별 후 해당 phase catalog만 Read하는지 dispatch 흐름 trace |
| Layer 2 (frontmatter 라인 범위 lazy-load) | router가 catalog로부터 받은 라인 범위(`offset`/`limit`)로 PROCEDURE 첫 N줄만 Read하는지 확인 |
| Layer 3 (본문 full lazy-load) | 선택된 skill의 본문만 추가 로드되는지 확인 |
| 전체 phase 격리 | ideation 작업 시 engineering 단계 catalog가 컨텍스트에 없는지 확인 |

### `evaluate-skill` paired evaluation 검증

Phase 1부터 마이그레이션 진행 후 `/buddy:evaluate-skill <name>` 호출 시:
- PROCEDURE.md frontmatter F1-F5 평가 (신규 적용)
- catalog entry 평가 (CE1-CE5 → 축소 또는 폐지 결정 필요, 별도 D2 작업)
- 본문 B1-B12 평가
- (적용 시) persona P1-P4 평가
- 통합 점수 80점 이상 권장 + 100점 우수 목표

### Phase 1 마이그레이션 종료 조건 (E1-E4)

- `catalog-phase-1.md` 생성 + Phase 1 모든 skill의 frontmatter 라인 위치 명시
- Phase 1 약 17개 skill 각각에 frontmatter 추가
- `evaluate-skill` paired evaluation 80점 이상
- Phase 1 → Phase 2 cascade 검증 (context 정리 또는 서브에이전트 분리 동작 확인)

---

## Trigger to Revisit

다음 중 하나라도 발생 시 본 ADR 재검토:

| Trigger | 재검토 사항 |
|---------|----------|
| Anthropic이 SKILL.md 자동 listing 메커니즘을 변경 (예: opt-in 방식, phase-aware listing 등) | Layer 0-1 메커니즘 재설계 |
| Claude Code의 `skillListingBudgetFraction` 정책 변경 | 토큰 예산 재계산 |
| buddy skill 수가 ~300+ 로 증가하여 phase-scoped catalog도 비대화 | Layer 1을 sub-phase 단위로 추가 분할 |
| catalog와 frontmatter의 SSoT 중복 운영 부담이 임계치 초과 | catalog 폐지 + frontmatter 단일 SSoT 검토 |
| Phase 경계 작업 (cascade)에서 빈번한 context 누락 발생 | Layer 2 lazy-load 시점·범위 조정 |

---

## References

### buddy 자체 문서
- `docs/se-lifecycle-naming.md` — SE Lifecycle 명사 SSoT (Phase 1-9 정의)
- `docs/plugin-skills-authoring-guide.md` — skill 작성·평가 기준 SSoT
- `plugin/skills/router/references/engineering-phases.md` — 각 phase 정체성·산출물·전이 규칙
- `plugin/skills/write-a-skill/references/naming-convention.md` — 스킬 명명 규칙
- `plugin/skills/router/references/skill-catalog.md` — 현재 단일 catalog (마이그레이션 대상)

### Anthropic 공식
- Claude Code Skills: https://code.claude.com/docs/en/skills.md
- Model configuration (`skillListingBudgetFraction` 등): https://code.claude.com/docs/en/model-config
- Prompt Engineering: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering

### 후속 작업 (본 ADR 기반 마이그레이션)

| 작업 | 산출물 |
|------|------|
| C2 | PROCEDURE.md frontmatter 표준 (첫 N줄 라인 안정성 보장 규칙) — authoring-guide §0 갱신 |
| C3 | catalog 분할 정책 — phase별 10 파일 + 각 catalog 형식 |
| C4 | phase orchestrator cascade 패턴 (context 정리 vs 서브에이전트 분리 결정 트리) |
| C5 | command 노출 정책 (현재 정책 유지 결정) |
| D1-D3 | `evaluate-skill` 정책 갱신 (F1-F5 PROCEDURE 적용, CE 항목 처리, 케이스 분모 단순화) |
| E1-E12 | Phase 1부터 점진 마이그레이션 (10 phase + cross-cutting) |

본 ADR이 정착되어 모든 skill이 4-Layer로 운영되면 활성 reference는 가이드·write-a-skill 등에 이관되며, 본 ADR은 "왜 이 아키텍처를 선택했는가"의 historical record로 유지된다.
