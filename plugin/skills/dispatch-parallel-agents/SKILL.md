---
name: dispatch-parallel-agents
description: "feature/task를 worktree로 격리해 Sonnet worker agent에 병렬 분배하고 결과를 aggregate. concurrency 제어, dependency 그래프 준수, 실패 격리, 결과 reconcile, 충돌 처리 표준화. 트리거: '병렬로 구현해' / '여러 feature 동시에' / 'worker 분배' / 'fan-out fan-in' / 'multi-agent dispatch' / 'agent에 분배해' / '병렬 작업'. 입력: feature spec set + dependency graph + concurrency limit + worker model. 출력: dispatch plan + per-worker briefs + aggregation policy + status board. 흐름: split-work-into-features/triage-work-items → dispatch-parallel-agents → auto-create-pr."
type: skill
---

# Dispatch Parallel Agents — Feature 병렬 분배 + Aggregate

## 1. 목적

feature/task set을 받아 **worktree 격리 + Sonnet worker agent 병렬 dispatch + 결과 reconcile**까지 관통한다.

이 스킬은 **Opus orchestrator + Sonnet worker** 모델의 핵심 backbone:
- **Opus**: dependency graph 분석, brief 작성, 결과 평가, 충돌 reconcile
- **Sonnet workers**: 병렬 코드 구현 (각 worktree 격리)

핵심 가치: **wall time을 N배 압축** (5 feature × 30분 sequential → 30분 parallel).

## 2. 사용 시점 (When to invoke)

- `split-work-into-features` 출력 feature set이 3개 이상이고 dependency 병렬 가능
- 여러 feature를 동시에 ship해야 하는 release 전 sprint
- 대규모 refactor에서 layer/module별 독립 분할 가능
- migration 작업이 file별 독립 실행 가능
- bug triage 후 ready-for-agent 항목 다수
- E2E test 시나리오 multiple 자동 작성

## 3. 입력 (Inputs)

### 필수
- feature spec set (`split-work-into-features` 출력) 또는 task list (`triage-work-items` ready-for-agent)
- dependency graph (병렬 가능 group 식별)
- concurrency limit (default: 3-5, system resource 따라)
- worker model 선택 (Sonnet 4.6 default)

### 선택
- worktree base directory (default: `~/buddy-worktrees/`)
- shared context (DESIGN.md, CONTEXT.md, ADR)
- API key / credential (worker가 외부 호출 시)
- timeout per worker (default: 30분)

### 입력 부족 시 forcing question
- "dependency 그래프 명확해? 순환 dependency 있으면 dispatch 못 함."
- "concurrency limit 몇으로 갈래? 동시 worker 수 = API rate limit + 머신 리소스."
- "worker가 외부 API 호출해? 결제 / 배포 / DB write이면 sandboxed mode 강제."
- "timeout 30분 충분해? E2E test 같은 건 60분+ 필요할 수 있음."

## 4. 핵심 원칙 (Principles)

1. **Worktree 격리** — 각 worker는 자기 worktree에서만 작업. 다른 worker 변경 안 보임. branch 분리.
2. **Brief 자기완결** — worker는 전체 context 모름. brief에 spec + 필요 file + acceptance criteria + verification 명령 모두 포함.
3. **Concurrency limit 강제** — API rate limit + 머신 CPU/RAM 보호. 동적 throttle.
4. **Dependency graph 준수** — A→B 순서면 A 완료 후 B dispatch. 순환은 reject.
5. **실패 격리** — 1개 worker 실패해도 다른 worker 계속. orchestrator가 실패 worker만 retry / escalate.
6. **결과 reconcile** — branch merge / rebase 시 충돌 처리는 orchestrator (Opus). worker는 자기 영역만.
7. **Idempotent retry** — 같은 brief로 재dispatch 시 동일 결과 (또는 명시 변형). worker side-effect dedup.
8. **Status board 실시간** — 각 worker pending / running / done / failed. user가 진행 상황 볼 수 있어야.

## 5. 단계 (Phases)

### Phase 1. Dispatch Plan 작성
1. feature/task set 입력
2. dependency graph 분석:
   - 순환 dependency 검출 → reject
   - parallel group 식별 (A, B 동시 가능, C는 A 후, D는 B 후)
   - critical path 식별
3. concurrency limit 적용 → wave 분할

### Phase 2. Worktree Setup
1. base branch에서 worker별 branch 생성
2. `git worktree add <path> <branch>` 각 worker
3. worker별 isolated working directory
4. shared context 복사 (DESIGN.md, CONTEXT.md)

### Phase 3. Per-Worker Brief 작성
brief 필수 항목:
- feature_id / task_id
- problem statement
- acceptance criteria (측정 가능)
- 영향 file list
- dependency (선행 worker 결과 필요 시)
- verification 명령 (`npm test`, `pytest tests/auth_test.py`)
- timeout
- escalation policy (실패 시 어떻게)

### Phase 4. Worker Dispatch
1. concurrency 그룹씩 fire
2. 각 worker는 Sonnet model + brief + worktree 받음
3. worker 작업 완료 시 결과 aggregate queue로 보고

### Phase 5. Status Tracking
실시간 status board:
- worker_id / branch / status / started_at / latest_action / failures
- user가 watch 가능
- failure threshold 초과 시 alert (예: 30% worker fail → STOP wave)

### Phase 6. Result Aggregation
worker 완료 후:
1. 검증: test pass? acceptance criteria 충족?
2. PASS → 다음 wave dispatch (dependency 따라)
3. FAIL → retry (1회) 또는 escalate (Opus reconcile)

### Phase 7. Branch Reconcile
모든 worker 완료 후:
1. main 또는 integration branch에 merge
2. 충돌 발생 시 Opus가 reconcile
3. linear history (rebase) vs merge commit 정책 결정
4. 통합 test 실행

### Phase 8. Cleanup
- worktree 제거 (`git worktree remove`)
- worker branch 정책 (delete merged / preserve for audit)
- status board 최종 보고

## 6. 출력 템플릿 (Output Format)

```yaml
dispatch_plan:
  total_features: 12
  parallel_groups:
    - wave: 1
      features: [FEAT-001, FEAT-002, FEAT-005]  # 병렬 가능
      depends_on: []
    - wave: 2
      features: [FEAT-003, FEAT-006]
      depends_on: [FEAT-001]
    - wave: 3
      features: [FEAT-007]
      depends_on: [FEAT-003, FEAT-005]
  concurrency_limit: 3
  worker_model: claude-sonnet-4-6
  estimated_wall_time_min: 90

worktree_setup:
  base_dir: ~/buddy-worktrees/
  base_branch: main
  worker_branches:
    - worker_id: w1
      branch: feature/auth-login
      worktree_path: ~/buddy-worktrees/w1-auth-login

worker_briefs:
  - worker_id: w1
    feature_id: FEAT-001
    spec_summary: "Email/password login with session"
    files_in_scope:
      - app/api/auth/login/route.ts
      - app/login/page.tsx
      - tests/auth-login.spec.ts
    acceptance_criteria:
      - "POST /auth/login with valid credentials returns 200 + session"
      - "Invalid password returns 401 + generic error"
    verification:
      - "pnpm test -- auth-login"
      - "pnpm typecheck"
    timeout_min: 30
    escalation: "if test fails after retry, escalate to Opus"

status_board:
  generated_at: "<timestamp>"
  workers:
    - worker_id: w1
      status: done
      branch: feature/auth-login
      duration_min: 22
      tests_passed: yes
      commits: 3
    - worker_id: w2
      status: running
      branch: feature/billing-integration
      latest_action: "writing test for refund flow"
      elapsed_min: 18
    - worker_id: w3
      status: failed
      branch: feature/admin-panel
      failure: "test failure on permission check"
      escalated_to: opus
  summary:
    total: 3
    done: 1
    running: 1
    failed: 1
    pending: 0

aggregation:
  successful_branches: [feature/auth-login]
  failed_branches: [feature/admin-panel]
  retries:
    - worker_id: w3
      attempt: 2
      result: still_failing
      action: escalate_to_human

reconcile:
  merge_strategy: rebase  # or merge_commit
  conflict_resolution: opus_orchestrator
  integration_branch: integration/sprint-23
  conflicts_resolved: 0
  conflicts_escalated: 0

cleanup:
  worktrees_removed: [w1, w2]
  worktrees_preserved: [w3]  # for debugging
  branches_deleted: [feature/auth-login]
  branches_kept: [feature/admin-panel]
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `split-work-into-features` 또는 `triage-work-items` (ready-for-agent) — `Skill` tool로 invoke
- 페어: `build-with-tdd` (worker 본 작업)
- 페어: `freeze-edit-scope` (worker 영역 lock)
- 다음 단계: `auto-create-pr` (각 worker branch에서 PR 생성)
- 후속: `iterate-fix-verify` (실패 worker fix loop)
- 후속: `automate-release-tagging` (모든 worker 통합 후 release)

## 8. Anti-patterns

1. **순환 dependency dispatch** — A↔B 양방향. dispatch 전 graph 검증 강제.
2. **Brief에 context 부족** — worker가 다른 file 추정 → 잘못된 변경. self-contained brief.
3. **Concurrency limit 무시** — 동시 20 worker → API rate limit + OOM. 강제 throttle.
4. **실패 1개로 wave 전체 stop** — 1개 worker fail이 다른 worker 결과 invalidate? 격리 원칙.
5. **Worker가 main branch 직접 commit** — race condition. worktree branch 강제.
6. **Reconcile을 worker가 함** — worker는 자기 영역만. 충돌 처리는 Opus orchestrator.
7. **Status 없이 dispatch** — user가 진행 상황 모름. real-time board 강제.
8. **Cleanup skip** — worktree 누적 → 디스크 가득. 완료 후 제거 강제.
