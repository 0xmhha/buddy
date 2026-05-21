# ADR-002 — roadmap.md v0.2 / v0.3 / v1.0 outline × two-tracks-charter cli buddy gap

> **Status**: Accepted
> **Date**: 2026-05-10
> **Deciders**: buddy maintainer
> **Tags**: roadmap, charter, cli-buddy, naming, scope-alignment
> **Related**:
> - [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §3 (cli buddy 정의) + §6.2 (cli buddy 진화 방향)
> - [`docs/archive/roadmap.md`](../../archive/roadmap.md) §4 (v0.2) / §5 (v0.3) / §6 (v1.0)
> - [`docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md`](./2026-05-09-buddy-commands-disable-model-invocation.md) (ADR-001)

---

## 1. Context

`docs/roadmap.md` 의 §4 (v0.2 Control Plane), §5 (v0.3 Orchestration), §6 (v1.0 통합) 은 v0.1 시점 (2026-04-23) 에 *Go CLI 트랙 — hook reliability monitor 의 후속 마일스톤* 으로 작성됐다. 그 시점에서 cli buddy 의 진짜 목적은 명시되지 않았다.

2026-05-10 에 사용자 발화 [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §3 으로 cli buddy 의 진짜 목적이 *자동화 agent 관리 — 생성 / 실행 / 종료 / 설정 변경 + TUI 화면* 으로 lock-in 됐다. 동시에 charter §3.6 은 v0.1.0 의 hook reliability monitor 가 cli buddy 의 한 sub-feature 라고 명시했다.

이 charter lock-in 으로 roadmap 의 v0.2 / v0.3 / v1.0 outline 이 *charter 의 cli buddy 정의와 부분 정합 / 부분 충돌* 상태가 됐다. 본 ADR 은 그 gap 을 매핑하고 처리 방향을 결정한다.

### 1.1 charter §3 cli buddy 핵심 인용

> "cli buddy 는 어떤 용도냐면, 'plugin buddy' 를 활용하여 자동화 agent 로 동작할수 있도록 지원하는 툴이다. 자동화된 agent 를 관리하고, 실행 및 종료 시키고, 설정을 변경하는등을 지원하는 툴이다. cli buddy 는 tui 로 화면을 지원하면서, 여러 자동화된 agent 를 설정하고 관리하는 툴"
>
> 핵심 4 책임: agent **생성 / 실행 / 종료 / 설정 변경**.

### 1.2 charter §6.2 cli buddy 진화 순서 (불변)

| 우선순위 | 작업 | 트리거 |
|---------|------|--------|
| 1 | 진짜 목적 spec 작성 — `docs/cli-buddy-spec.md` (가칭) | plugin buddy 1번 (미구현 skill 보완) 종료 후 |
| 2 | TUI / agent runtime / plugin buddy 내재화 layer 설계 | 1번 종료 후 |
| 3 | 기존 v0.1.0 (hook reliability monitor) 을 cli buddy sub-feature 로 재배치 또는 별개 유지 결정 | 1번 진행 중 |
| 4 | reference implementation (예: 웹툰 agent) | 2번 종료 후 |

→ 즉 *roadmap outline 의 actual rewrite* 는 charter §6.2 의 *우선순위 1번 작업 시점* 까지 deferred. 본 ADR 시점에서는 *마킹 + gap 영구화* 만.

---

## 2. Gap 매핑 — roadmap 15 항목 × charter cli buddy 정합

### 2.1 v0.2 — Control Plane (T1 ~ T6)

| # | 항목 | 정합 부분 | 불일치 부분 | 처리 패턴 |
|---|------|---------|----------|----------|
| T1 | 활성 세션 발견 (`internal/sessions/`) | session = Claude Code 인스턴스 단위. agent 가 세션을 *주도* 할 때 그 세션을 *관찰* 하는 layer 로 차용 가능 | session 자체는 *Claude Code 가 만드는 단위*. cli buddy 의 agent (= 사용자가 *생성* 한 자동화 단위) 와 *주체 / 책임* 다름 | **P2** 의미 재해석 — session 은 *agent 안의 한 실행 trace* 로 강등 |
| T2 | Token usage parser | agent 의 비용 추적에 그대로 활용 | 변경 없음 | **P1** 정합 |
| T3 | Multi-session stats | 여러 *agent 의 session* 동시 보기로 변환 | 현재 outline 은 "사용자가 띄운 세션 N 개" 가정 | **P2** agent ⊃ session 으로 재해석 |
| T4 | Dashboard UI (TUI vs web) | TUI 채택은 charter §3.1 와 직접 정합 | dashboard = *모니터링 only* / charter cli buddy = *control* 책임 (생성 / 종료 / 설정) | **P2** TUI OK, *dashboard* → *agent manager* 확장 |
| T5 | Cost estimate | 정합 (agent 비용 추적) | — | **P1** 정합 |
| T6 | i18n full split | 정합 (어느 트랙에서도 필요) | — | **P1** 정합 |

### 2.2 v0.3 — Orchestration (T1 ~ T5)

| # | 항목 | 정합 부분 | 불일치 부분 | 처리 패턴 |
|---|------|---------|----------|----------|
| T1 | task DAG schema | agent 의 작업 분해 backbone | task 가 *agent 안의 sub-action* 인지 / *agent 자체* 인지 모호 | **P2** agent layer 가 task DAG 위에 추가 — 스키마 보존 |
| T2 | `buddy task add/list/run/status` CLI | task 단위 명령 | agent 단위 명령 (`buddy agent create/run/stop`) 과 *별개* 라 단어 충돌 위험 | **P3** naming 재정비 — *task* 는 agent 안의 sub-unit, *agent* 는 상위 단위 |
| T3 | Wave 병렬 실행 | agent 안의 병렬 task 실행 | — | **P1** 정합 |
| T4 | Retry policy | agent 안정성 (failure → retry) 의 핵심 | — | **P1** 정합 |
| T5 | 외부 task tracker 통합 | agent 가 외부 tracker 와 연동 가능 | 우선순위 낮음 | **P1** 정합 (후순위) |

### 2.3 v1.0 — 통합 (T1 ~ T4)

| # | 항목 | 정합 부분 | 불일치 부분 | 처리 패턴 |
|---|------|---------|----------|----------|
| T1 | AGENTS.md auto-sync | "AGENTS.md" 는 사용자 프로젝트 capability 기록 | "AGENTS.md 의 agent (capability metadata)" ≠ "cli buddy 의 agent (자동화 실행체)" — *단어 충돌 명백* | **P3** 단어 분리 — AGENTS.md 는 *plugin buddy* 의 capability 기록 영역, cli buddy 의 agent 는 별개 단위 |
| T2 | plugin model | "사용자 정의 hook health checker 등록" — 외부 plugin 로드 | charter §3.6 의 *plugin buddy 내재화* 는 *cli buddy 가 plugin buddy 를 호출하는 인터페이스 layer*. 다른 차원 | **P3** 두 의미 분리 — *외부 plugin* (현재 outline) vs *plugin buddy 내재화* (charter) |
| T3 | MCP server | 정합 — agent 가 MCP 통해 외부 도구 호출 | — | **P1** 정합 |
| T4 | cross-harness (Codex / OpenCode) | charter 무관 (다른 트랙 영역) | — | **별도** — charter 외연 |

### 2.4 종합 — 15 항목 분포

| 패턴 | 개수 | 항목 |
|------|-----|------|
| P1 그대로 차용 | 7 | v0.2 T2 / T5 / T6, v0.3 T3 / T4 / T5, v1.0 T3 |
| P2 의미 재해석 / 상위 layer 추가 | 4 | v0.2 T1 / T3 / T4, v0.3 T1 |
| P3 단어 충돌 / 의미 분리 | 3 | v0.3 T2, v1.0 T1, v1.0 T2 |
| 별도 (charter 외연) | 1 | v1.0 T4 |

→ **47% (7/15) 는 무수정 차용 가능**, **27% (4/15) 는 의미 재해석 필요**, **20% (3/15) 는 단어 충돌 해소 필요**.

---

## 3. Decision

### 3.1 처리 방법 — (iv) 하이브리드

본 ADR 시점에서는:

1. `docs/roadmap.md` §4 / §5 / §6 *각 섹션 헤더 직후* 에 *재평가 필요* 마킹 추가 (3 건). silent conflict 차단.
2. 본 ADR (gap 매핑 표 + 처리 패턴 분류) 작성 — 영구화. 미래 cli buddy spec 작성 시점에 *입력 자료* 로 활용.

### 3.2 Actual rewrite 시점

charter §6.2 의 *우선순위 1번* (`docs/cli-buddy-spec.md` 신규 작성) 시점. 그 시점에:

- P1 7 항목 → cli-buddy-spec.md 의 sub-task 로 그대로 흡수
- P2 4 항목 → cli-buddy-spec.md 안에서 의미 재해석 후 sub-task 로 흡수
- P3 3 항목 → cli-buddy-spec.md 작성 시 *naming 결정* 의 입력. 결정 후 outline 에서 제거 또는 rename
- 별도 1 항목 (cross-harness) → *cli buddy 의 outscope*. roadmap 에서 분리 또는 그대로 보존

### 3.3 Trigger to revisit

다음 중 하나 발생 시 본 ADR 재평가:

- `docs/cli-buddy-spec.md` 신규 작성 시작 (= rewrite trigger)
- charter §3 의 cli buddy 정의 변경 (단어 / 책임 경계 변경)
- 외부 사용자가 buddy fork 후 *현재 roadmap outline* 을 그대로 적용해 마찰 발생 (= silent conflict 가 실제 발화)

---

## 4. Alternatives considered (rejected)

### 4.1 (i) 재정의 (rewrite now)

**Approach**: v0.2 / v0.3 / v1.0 outline 을 cli buddy 의 agent 관리 기준으로 즉시 다시 작성.

**Rejected because**:
- charter §6.2 *우선순위 1번* (plugin buddy 정리) 과 충돌. plugin buddy 작업 중인 시점에 cli buddy spec 작성은 우선순위 위반.
- 작업량 큼 — 15 항목 모두 재작성 + 의존 관계 재정의.
- v0.1 시점 사고 (P1 7 항목의 정합 부분) 이 *지금 시점에서 옳다고 검증* 안 된 채 폐기 위험.

### 4.2 (ii) 폐기 (discard now)

**Approach**: roadmap.md 의 §4 / §5 / §6 archive 로 옮김. cli-buddy-spec.md 는 *처음부터 charter 기준* 으로 작성.

**Rejected because**:
- v0.1 시점의 *부분 정합 사고* 손실. 특히 P2 / P3 의 9 항목 (의미 재해석 / 단어 충돌) 은 *cli-buddy-spec.md 작성 시 trade-off 학습 자료* 로 가치.
- archive 로 옮긴 outline 은 *접근성 낮음*. 미래 spec 작성자 (본인 또는 다른 세션) 가 outline 을 reference 로 활용 어려움.

### 4.3 (iii) 그대로 두기 (keep as-is)

**Approach**: outline 변경 0. cli buddy spec 작성 시 결정.

**Rejected because**:
- silent conflict 위험. 미래 사용자 / 본인 / 다른 세션이 roadmap.md 를 *그대로 truth* 로 받아들임.
- charter 와 roadmap 두 SSoT 간 정합성 점검 부재 → 두 문서 간 drift 누적.

---

## 5. Consequences

### 5.1 Positive

- silent conflict 차단 — roadmap.md 진입 시 *charter 정합 재평가 필요* 명시 노출
- v0.1 시점 사고 보존 — outline 항목들이 *historical 사고* 로 read 되며, P1 7 항목 / P2 4 항목 / P3 3 항목 분류가 cli buddy spec 작성 시 입력으로 활용
- charter §6.2 진화 순서 와 정합 — plugin buddy 정리 우선, cli buddy spec 후순위
- 최소 변경 원칙 정합 — outline 본문 무수정, 마킹 3 건 + ADR 1 건만 추가
- Trigger to revisit 명시로 미래 ADR 재평가 시점 명확

### 5.2 Negative

- outline 의 일부 표현 (예: v1.0 T1 AGENTS.md auto-sync) 이 *진짜 의미와 다른* 채로 남음. 마킹으로 명시되지만 *각 항목 단위로는* 무수정
- ADR-002 가 *gap 표 보존용* — 미래 cli buddy spec 작성 시 본 ADR 을 *반드시 read* 해야 actual rewrite 시 일관성 유지. 따라서 spec 작성 procedure 에 본 ADR reference 의무화 필요 (별도 PROCEDURE.md 갱신)

### 5.3 Neutral

- v0.1.0 ship 자산 (M1 ~ M5 의 hook reliability monitor) 은 본 ADR 영향 없음. v0.2 / v0.3 / v1.0 *미래 항목* 만 영향

---

## 6. Verification (적용 후 측정)

| 항목 | 측정 |
|------|------|
| roadmap.md §4 / §5 / §6 *재평가 필요* 마킹 등재 | 3 건 (각 섹션 헤더 직후) |
| 본 ADR 줄 수 | 추정 200~250 줄 |
| ADR 의 처리 패턴 분류 정확도 | 4 패턴 (P1 / P2 / P3 / 별도) × 15 항목 = 60 셀 매핑 표 |
| Trigger to revisit 명시 | 3 trigger (cli-buddy-spec.md 작성 / charter §3 변경 / 외부 fork silent conflict) |

---

## 7. References

- [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §3 / §6.2
- [`docs/archive/roadmap.md`](../../archive/roadmap.md) §4 / §5 / §6
- [`docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md`](./2026-05-09-buddy-commands-disable-model-invocation.md) — ADR-001 (commands frontmatter 컨벤션)

본 ADR 의 결정은 *charter §6.2 의 우선순위 1번 작업 시점* 에 재평가된다. 그때까지 본 ADR 은 *gap 영구화* 로 read.
