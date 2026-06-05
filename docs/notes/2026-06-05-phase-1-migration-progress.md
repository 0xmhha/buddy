# Phase 1 마이그레이션 — 진행 상황 + 작업 계획 (2026-06-05)

> **목적**: ADR-020 (4-Layer Lazy-Load Architecture) 의 Phase 1 마이그레이션 진행 상황 + 잔여 작업 + 결정 보류 항목을 다음 세션 인계용으로 영속화. 본 문서 + `git log -5` + `docs/HANDOFF.md` + `docs/plugin-skills-authoring-guide.md` 만 읽으면 *컨텍스트 0* 으로 이어받을 수 있다.
>
> **선행 핸드오프**: [`2026-06-01-guide-audit-findings.md`](./2026-06-01-guide-audit-findings.md) (29건 가이드 결함 처리 baseline).
>
> **상위 ADR**: [`../superpowers/decisions/2026-06-02-lazy-load-architecture.md`](../superpowers/decisions/2026-06-02-lazy-load-architecture.md) (ADR-020).

---

## 1. 본 세션 작업 요약

### 1.1 docs/ 정리 작업 (commits 4건)

| commit | 내용 | files |
|--------|------|-------|
| `4440842` | 사용 후 검증 (dogfood) 가이드 3종 제거 + HANDOFF.md 갱신 (v0.13.0 baseline, 9/9 closed, ADR-018/019/020 lock-in 반영) | 7 |
| `77b205d` | 2026-05 시점 스킬 분석 snapshot 5종 archive 이동 (외부 스킬 분석, 인벤토리, engineering audit, best practice flow coverage, I/O contract 그래프) + cross-reference 정정 | 11 |
| `2837610` | 오래된 핸드오프 / cycle 결과 7건 archive 이동 (5/10 ~ 5/23) + 2026-05-11 cleanup 마무리 (4 active skill 의 missing skills inventory 참조를 inline 요약으로 흡수) | 15 |
| (origin push) | `fad6004..2837610` 10 commits 모두 origin/main 동기화 완료 (본 세션 3 + 이전 세션 7) | — |

### 1.2 Phase 1 마이그레이션 — 시범 진행

**Step 1: catalog 분할 = 다른 세션 완료** (2026-06-02 ~ 2026-06-03):
- `plugin/skills/router/references/catalog/` 디렉토리 신규 생성
- 9 phase 파일 + cross-cutting (계획) 작성
- Phase 1: `phase-1-discovery-impact.md` (13 stage skill 등재, 형식 = `Skill | Trigger | When to use | Not when`)

**Step 2: 시범 skill frontmatter 추가 (1건 / 13 건)**:
- 대상: `analyze-market-size` (Phase 1 / product / full-standalone — 가장 단순한 후보)
- 형식: ADR-020 운영 가이드 §0.3 표준 따름 (R1-R5 라인 안정성)
- frontmatter 라인 범위: **1-10**
- 결과: ✅ 추가 완료 (본 commit)

```yaml
---
name: analyze-market-size
description: |
  Use when 시장 규모를 정량 산출해야 할 때 (TAM/SAM/SOM ...)
  Trigger phrases: "시장 크기 분석", "TAM 산출", ...
when_to_use: |
  사업성 7차원 평가 (assess-business-viability) 내 시장 규모 차원 ...
disable-model-invocation: true
---
```

---

## 2. 결정 보류 사항 (사용자 결정 대기)

### 2.1 Catalog 형식 통합 결정 — 가장 시급

**상황**: `plugin-skills-authoring-guide.md` §0.5.2 표준 형식과 `catalog/phase-1-discovery-impact.md` (다른 세션이 신규 분할 시 작성한 형식) 가 다름.

| 형식 | 컬럼 | 근거 |
|------|------|------|
| **§0.5.2 표준 (가이드)** | `Skill \| Frontmatter \| 1줄 요약 \| Command` (4) | 가이드 §0.5.2 / §0.5.3 |
| **catalog/phase-1 현재 (다른 세션)** | `Skill \| Trigger \| When to use \| Not when (→ 대안)` (4) | 다른 세션 작성 |

**옵션**:

| 옵션 | 처리 | trade-off |
|------|------|----------|
| **(a-1)** | catalog/phase-1 analyze-market-size 1 행만 §0.5.2 형식 변환 (시범) + 나머지 12 행 현 형식 양립 → 점진 변환 | 시범 패턴 확립 우선, 13 skill frontmatter 작성하며 점진 |
| **(a-2)** | catalog/phase-1 전체 §0.5.2 표준 변환 + 12 skill frontmatter 미작성은 placeholder | 형식 일관성 즉시. 단 placeholder = 임시 |
| **(b)** | 6 컬럼 통합 (Skill \| Frontmatter \| Trigger \| When \| Not when \| Command) | 작업 부담 적음. SSoT 중복 (catalog + frontmatter 양쪽 trigger 정보) |

**권장 (본 세션)**: **(a-1)** — ADR-020 lazy-load 의도 (Layer 1 catalog = 위치 안내, Layer 2 frontmatter = dispatch 근거) 정합 + SSoT 단일화. 단 사용자 미결정 — **다음 세션 진입 시 첫 결정 사항**.

---

## 3. 잔여 작업 계획

### 3.1 Phase 1 마이그레이션 잔여 (Task #43)

**우선순위 순**:

1. **Catalog 형식 결정** (위 §2.1, 사용자 결정 대기)
2. **catalog/phase-1 analyze-market-size 행 변환** (시범 패턴 확립)
3. **`analyze-market-size` paired evaluation** — evaluate-skill 호출, 24항목 체크리스트 (F1-F5 + CE 3 + B1-B12 + P1-P4) + PC1-PC4 paired 케이스 점수 산출
4. **시범 결과 검토** → 패턴 확정
5. **나머지 12 skill frontmatter 작성 + catalog 행 변환** — 점진 또는 일괄 (사용자 결정)
6. **Phase 1 전체 paired evaluation** (13 skill)

**Phase 1 대상 skill 13건** (catalog 등재 순):
1. `assess-product-change` (Mode B)
2. `validate-idea`
3. `validate-advanced-edge-idea`
4. `assess-business-viability`
5. `analyze-market-size` ← **시범 진행, frontmatter 완료**
6. `map-customer-segments`
7. `map-jobs-to-be-done`
8. `conduct-customer-interview`
9. `analyze-competition-and-substitutes`
10. `decide-target-market`
11. `review-pricing-and-gtm`
12. `define-product-spec`
13. `write-hld`

### 3.2 Phase 2~9 + Cross-cutting 마이그레이션

Phase 1 패턴 확립 후 동일 절차로 점진:
- Phase 2 (Feature Definition) — 11 skill
- Phase 3 (Technical Design) — 33 skill
- Phase 4 (Implementation Planning) — 7 skill
- Phase 5 (Development) — 10 skill
- Phase 6 (Verification & Quality) — 19 skill
- Phase 7 (Release) — 14 skill
- Phase 8 (Operations & Iteration) — 21 skill (engineering 7 + product 10 + marketing 4)
- Phase 9 (Lifecycle Management) — 4 skill
- Cross-cutting Utilities — 11 skill (재분류 포함)

**합계**: 157 skill (Phase 1 시범 분 1 포함).

### 3.3 별도 작업 (Task list)

| Task | 내용 | trigger |
|------|------|---------|
| **#41** | two-tracks-charter 7 stale 항목 갱신 (수량 105/57→157/103, cli buddy 구성요소, baseline v0.13.0, 머신 경로, archive 위치, 진화 방향 진척, §8 완료 후속작업) | 즉시 진행 가능 (charter 변경 정책 §7 표준 절차) |
| **#42** | classification-matrix 갱신 (152→157 + catalog phase-1~9 신규 9 파일 분류 행 + §1 표 영문 명사 통일) | ADR-020 마이그레이션 완료 후 (catalog 구조 진화 중에는 stale 재발 위험) |
| **#43** | Phase 1 마이그레이션 (본 문서의 §3.1) | 사용자 주도, "하나씩 진행" 명시 |

### 3.4 v1.0.0 publish (별 세션 권장)

9 entry conditions 모두 충족 (commit `14d2f83` 2026-05-21 dogfood close). publish 절차:
1. `make set-version VERSION=1.0.0`
2. CHANGELOG 작성 (W7 5 surface ship [F2.A~F2.E] 강조)
3. release tag + binary publish
4. 후속: Phase 1 마이그레이션 (ADR-020) 점진

---

## 4. 알려진 위험 / 함정

### 4.1 catalog 형식 미결정 시 마이그레이션 stall

§2.1 결정 보류 → 12 skill frontmatter 작성 시 catalog 행 형식 미정 → 두 형식 혼재 위험. **다음 세션 진입 시 첫 결정 필수**.

### 4.2 다른 세션 동시 작업

buddy 디렉토리는 다른 세션 / daemon / background process 동시 작업 가능. commit 범위가 의도보다 넓어져도 정상 패턴 — 사용자 인지 작업이면 분할 회피. 메모리 [[feedback-concurrent-session-commit-bundling]] 참조.

### 4.3 약어 단독 사용 금지

프로젝트 내부 약어 (B-2, W4-5b, F2.C, BA-12, C1-C4 등) 단독 사용 금지. 약어 옆에 무엇을 의미하는지 풀어쓰기 병기. 메모리 [[feedback-abbreviation-expansion]] 참조.

### 4.4 다른 세션 작업 잔존 (router 영역)

본 세션 종료 시점에 다른 세션 작업 commit (`93d1751 refactor(router): fix routing precedence and guide entry-skill selection`) 이 origin/main 에 누적. 본 세션 작업과 호환 (lazy-load 마이그레이션 = router 갱신 일환).

---

## 5. 다음 세션 시작 권장 절차

### 5.1 즉시 (5 분)

1. `git pull` (buddy repo, branch: main)
2. `git log --oneline | head -10` — 최신 commit 확인
3. 본 문서 + `docs/HANDOFF.md` + `docs/plugin-skills-authoring-guide.md` §0 단원 읽기

### 5.2 사용자 결정 받기 (5 분)

§2.1 catalog 형식 통합 결정 — (a-1) / (a-2) / (b) 중 선택.

### 5.3 작업 진행

- 결정에 따라 catalog/phase-1 변환 (1 행 또는 전체)
- `analyze-market-size` paired evaluation 실행: `/buddy:evaluate-skill analyze-market-size`
- 결과 검토 → 시범 패턴 확정
- 나머지 12 skill 진행 방식 결정 (점진 vs 일괄)

---

## 6. 사용자 명시 운영 규칙 (보존)

- **Git commit attribution 금지**: Co-Authored-By / "Generated with Claude" 절대 금지
- **사용자가 한국어로 쓰면 한국어로 응답**
- **Uncommitted 변경 종료 시 사용자에게 commit 여부 먼저 확인**
- **Push 패턴**: 사용자 명시 승인 후만
- **Destructive git command 절대 금지** (reset --hard, push --force, checkout 등)
- **약어 풀이 의무** (메모리 [[feedback-abbreviation-expansion]])
- **다른 세션 commit bundling 정상** (메모리 [[feedback-concurrent-session-commit-bundling]])

---

## 7. 참조

- [`../HANDOFF.md`](../HANDOFF.md) — 세션 인계 (v0.13.0 baseline, 9/9 closed)
- [`../plugin-skills-authoring-guide.md`](../plugin-skills-authoring-guide.md) §0 — 4-Layer lazy-load 운영 안내 + frontmatter 표준 (R1-R5) + catalog 분할 정책
- [`../superpowers/decisions/2026-06-02-lazy-load-architecture.md`](../superpowers/decisions/2026-06-02-lazy-load-architecture.md) — ADR-020 결정 근거
- [`../plugin-skills-classification-matrix.md`](../plugin-skills-classification-matrix.md) — 152 skill 분류 (Task #42 갱신 예정)
- [`../../plugin/skills/router/references/catalog/phase-1-discovery-impact.md`](../../plugin/skills/router/references/catalog/phase-1-discovery-impact.md) — Phase 1 catalog (다른 세션 작성)
- [`../../plugin/skills/analyze-market-size/PROCEDURE.md`](../../plugin/skills/analyze-market-size/PROCEDURE.md) — 시범 frontmatter 추가 완료
