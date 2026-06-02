# §5 Construction (개발) — Stage Skills

- **Orchestrator (entry)**: `build-feature`
- **DoR → DoD**: iteration plan (task DAG) → working code + developer tests (green)
- 정체성 원본: `engineering-phases.md` §2 Phase 5 | 명사 원본: `se-lifecycle-naming.md` §1

## Stage skills

| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `build-with-tdd` | command + dispatch | 신규 기능을 red-green-refactor TDD 루프로 구현 | 기존 버그 원인 추적 → `diagnose-bug` / 순수 rename → `refactor-with-rename-trace` |
| `diagnose-bug` | command + dispatch | 버그의 재현·근본 원인을 규명 | 원인이 이미 명확, fix 반복 적용만 → `iterate-fix-verify` |
| `iterate-fix-verify` | dispatch [패턴 라이브러리] | review/QA finding 목록을 하나씩 fix→commit→re-verify | 단일 버그 원인 미상 → `diagnose-bug` / 신규 기능 → `build-with-tdd` |
| `refactor-with-rename-trace` | command + dispatch | 동작 불변 rename/이동 (LSP rename + grep 누락 검증) | 동작 변경 포함 → `build-with-tdd` |
| `dispatch-parallel-agents` | command + dispatch | 독립 task 다수를 worktree 격리 병렬 worker로 분배 | 단일·순차 작업이면 불필요 (직접 구현) |
| `pair-program-loop` | command + dispatch | driver/navigator 역할 분리가 필요한 고난도 구현 | 단독 구현으로 충분 → `build-with-tdd` |
| `generate-from-api-contract` | command + dispatch | API contract → 클라이언트 SDK + 서버 stub + type 자동 생성 | contract 미확정 → §3 `design-api-contract` 선행 |
| `generate-tests-from-spec` | command + dispatch | acceptance criteria + test plan → test skeleton 자동 생성 | 구현 후 의미 coverage 평가(§6) → `audit-test-coverage-meaningful` |
| `freeze-edit-scope` | dispatch [패턴 라이브러리] | 세션 중 Edit/Write를 단일 디렉토리로 lock (ambient) | 직접 진입점 아님 — 다른 스킬 내부에서 적용 |
| `update-docs-with-code` | command + dispatch | 코드 변경 → README/ADR/CHANGELOG/HANDOFF/catalog 5영역 동기화 | 출시 시점 문서 동기화(§7) → `sync-release-docs` |

## Disambiguation (노드 내)
- `build-with-tdd`(새것 구현) vs `diagnose-bug`(깨진 것 원인 규명) vs `refactor-with-rename-trace`(동작 불변 정리).
- `diagnose-bug`(원인 규명) vs `iterate-fix-verify`(원인 알고 다건 수리 반복).

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3 케이스 D. DoR(iteration plan) 부재 → §4 `plan-build` 선행 (Mode B small이면 work item만으로 직접 진입 — SKILL.md 워크플로우 step 3).
- 코드 작업 stuck(동일 문제 3회) → `decompose-blocker` (cross-cutting) 자동 trigger.
