# ADR-003 — `superpowers` external project attribution policy

> **Status**: Accepted
> **Date**: 2026-05-10
> **Deciders**: buddy maintainer
> **Tags**: attribution, naming-collision, charter, skill-catalog
> **Related**:
> - [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §6.3 호환성 우선 원칙
> - skill-matrix-3b note (`docs/notes/2026-05-10-skill-matrix-3b-group-1.md`) §4.1 — removed in v0.2.0 doc cleanup; *외부 superpowers 발견* 핵심 내용은 §1 Context 에 인용
> - [`docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md`](./2026-05-09-buddy-commands-disable-model-invocation.md) (ADR-001)
> - [`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`](./2026-05-10-roadmap-charter-gap.md) (ADR-002)

---

## 1. Context

Step 3b 매트릭스 작성 (Group 1 × external assets) 중 *외부 자산 추가 탐색* 에서 다음 발견:

> `/Users/kevin/work/github/aidax-dag/ai-cli/skill/superpowers/` 의 README:
>
> > "Superpowers is a complete software development methodology for your coding agents, built on top of a set of composable skills and some initial instructions that make sure your agent uses them."

이 외부 자산은 buddy repo 의 다음 자산과 *명명 + 패턴 모두 충돌*:

| 항목 | buddy 내부 | external superpowers |
|------|---------|------------------|
| 명명 | `docs/superpowers/` 디렉토리 | 프로젝트 이름 자체 |
| 핵심 패턴 | router skill + PROCEDURE.md + dispatch | composable skills + initial instructions |
| 본문 형식 | YAML frontmatter + 12 section | (검증 필요 — README 만 read) |
| 자산 분류 | spec / decision / plan 영속화 | software development methodology |

### 1.1 추정 inspiration 출처

buddy 의 `docs/superpowers/` 디렉토리 + *router + composable skills* 패턴은 *external superpowers project 영감* 가능성 높음:

- 명명 동일 (`superpowers`)
- "composable skills" 개념 정합 (buddy 의 *skill catalog* 와 일치)
- "initial instructions that make sure your agent uses them" = buddy 의 *router skill* 의 동일 책임

### 1.2 attribution 의 시급성

charter §6.3 의 *호환성 우선 원칙* 인용:
> "이 수많은 skill 이 겹치는 부분이 너무 많고, 조금씩 아쉬운 부분이 존재"
> "스킬 능력이 오히려 떨어질 수 있는 부분도 존재하니 무조건 가져올 것이 아니라, 우리의 'buddy' 프로젝트 스킬들과 호환 및 궁합이 좋은지 점검이 필요"

본 ADR 은 *체계적 attribution + 명명 충돌 해소 정책* 을 정한다.

---

## 2. Decision

### 2.1 NOTICE 에 superpowers attribution 추가

`NOTICE` 파일에 다음 attribution block 추가 (license 검증 후 — README 만 본 상태에서 추정):

```
---

superpowers (external)
Copyright (c) 2026 (TBD — verify upstream)
Source: <upstream URL TBD — likely GitHub>

Inspiration acknowledgment for:
- buddy's docs/superpowers/ directory naming
- composable-skill + router-instruction pattern adopted in
  plugin/skills/router/ + plugin/skills/<name>/PROCEDURE.md

No verbatim code or text adopted from upstream superpowers.
buddy's implementation is independently authored.
```

> **Note**: actual attribution block 작성은 *upstream license + author 검증 후*. 본 ADR 은 *작성 정책* lock-in.

### 2.2 명명 충돌 해소 — 현 상태 유지

| 옵션 | 영향 |
|------|------|
| (a) buddy `docs/superpowers/` rename (예: `docs/governance/`) | 19+ commit 의 reference 모두 변경 — 큰 변경 |
| (b) **현 상태 유지 + ADR-003 attribution** (Recommended) | 명명 동일하나 *attribution 으로 명확화*. 변경 0 |
| (c) external superpowers 와 *동기화 시도* | 외부 자산 변경 추적 부담 |

**채택**: (b) — buddy `docs/superpowers/` 의 *책임 (governance / spec / decision / plan 영속화)* 가 external superpowers 의 *책임 (software development methodology)* 와 *유사하지만 다른 차원*. attribution 으로 충분.

### 2.3 같은 패턴의 외부 자산 발견 시 정책 확장

향후 *추가 외부 자산 발견 시*:

1. NOTICE 에 attribution 즉시 추가 (LICENSE 호환 검증 후)
2. README 의 Acknowledgments 섹션 갱신
3. 영향 큰 경우 별도 ADR (본 ADR 패턴 reuse)
4. *명명 충돌* 시 본 ADR §2.2 의 3 옵션 중 결정

### 2.4 Inspiration vs adoption 의 명시

모든 PROCEDURE.md / NOTICE entry 는 다음 분류 명시:

| 분류 | 의미 |
|------|------|
| **Verbatim adoption** | 본문 / 코드 그대로 차용 (현재 0 — 모든 buddy 자산 독립 작성) |
| **Adopt with edits** | 양식 / pattern 차용 후 수정 (예: marketingskills 18 항목 → buddy 6 skill 통합) |
| **Reference only** | 이름 / 개념만 mention, 본문 독립 (예: superpowers 명명 + composable skill pattern) |
| **Inspired by** | 디자인 결정 영향, 직접 차용 X (예: agent-evaluation 의 OMAS v2 pattern) |

---

## 3. Alternatives considered (rejected)

### 3.1 Step 3b 시점 즉시 처리 (rejected)

**Approach**: Skill Completion Cycle 진행 중 ADR 작성.

**Rejected because**: 사용자 D-H 결정 (H1 — 별도 cycle) 과 충돌. main cycle (skill 보완) 우선.

### 3.2 Buddy `docs/superpowers/` rename (rejected)

**Approach**: `docs/governance/` 또는 다른 이름으로 변경.

**Rejected because**:
- 19+ commit 의 cross-reference 모두 영향 (HANDOFF / tasks / charter / spec / decision / plan)
- 명명 자체가 *buddy 의 정체성* 일부 — rename 시 git history grep 으로 도달성 손상
- attribution 만으로 의미 정합 — rename 의 *정량 가치* 낮음

### 3.3 Attribution skip (rejected)

**Approach**: 외부 자산 명명 충돌 무시.

**Rejected because**:
- charter §6.3 *호환성 우선 원칙* 위반
- 미래 *외부 사용자 / 본인 / 다른 세션* 이 명명 출처 모름
- legal / 윤리 차원 attribution 가치 손실

---

## 4. Consequences

### 4.1 Positive

- 외부 자산 inspiration 출처 *명시 영속화* — git history + NOTICE + ADR 3 layer
- 미래 발견 외부 자산의 *attribution 정책 reuse* 가능 (§2.3)
- charter §6.3 호환성 우선 원칙 정합
- *명명 충돌* 자체는 ADR 로 해소 (rename 비용 0)

### 4.2 Negative

- NOTICE 갱신 작업 *upstream license 검증* 의존 (TBD 항목)
- *명명 동일* 이라 미래 사용자 / 본인이 *직접 비교* 시 혼란 가능 — ADR 가 그 혼란 해소 역할

### 4.3 Neutral

- buddy 기존 코드 / 본문 *변경 0*. 본 ADR + NOTICE 갱신만.

---

## 5. Verification

| 항목 | 측정 |
|------|------|
| ADR 작성 | 본 파일 등록 |
| NOTICE 갱신 | superpowers attribution block 추가 (별도 commit, upstream 검증 후) |
| 4 분류 (verbatim / adopt-with-edits / reference / inspired) 적용 | 모든 PROCEDURE.md 및 NOTICE entry |
| 명명 충돌 silent 회피 | 본 ADR 영구화로 해소 |

---

## 6. Trigger to revisit

- upstream superpowers 가 *commercial license* 로 변경 시
- buddy 가 *외부 사용자 fork* 받아 superpowers 명명 충돌 보고 시
- 새 external 자산 발견 + 명명 충돌 (본 ADR §2.3 정책 reuse)

---

## 7. References

- [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §6.3 — 호환성 우선 원칙
- skill-matrix-3b note §4.1 신규 발견 — file removed in v0.2.0 cleanup, see git history
- [`docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md`](./2026-05-09-buddy-commands-disable-model-invocation.md) (ADR-001) — N-1 closure
- [`docs/superpowers/decisions/2026-05-10-roadmap-charter-gap.md`](./2026-05-10-roadmap-charter-gap.md) (ADR-002) — roadmap charter gap
- 외부 superpowers (skill 경로 inventory): `aidax-dag/ai-cli/skill/superpowers/`
