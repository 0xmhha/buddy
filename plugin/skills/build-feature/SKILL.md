---
name: build-feature
description: This skill should be used when the user wants to "build a feature", "implement this", "start coding", "develop the feature", "write the code", or has an implementation plan ready and needs to execute it. Orchestrates §5 Development phase — manages TDD loops, parallel agent dispatch, and actor-track execution.
---

# build-feature — §5 Development Orchestrator

§5 라이프사이클 단계의 진입점. Implementation plan (actor별 task track) → Working code per feature + unit/integration tests.

**진입 조건**: §4 implementation plan 확정 (task DAG + parallel execution plan).
**산출물**: Working code per feature, unit tests, integration tests, PR-ready branch.
**다음 phase**: Code complete → `verify-quality` (§6).

---

## Stage 흐름

```
build-feature (§5 phase orchestrator)
├── stage 1: setup-development-environment  (quality gates + hooks 설정)
├── stage 2: dispatch-parallel-agents       (actor track별 병렬 agent 분배)
│   ├── track A: build-with-tdd (frontend)
│   ├── track B: build-with-tdd (backend)
│   └── track C: build-with-tdd (3rd-party integration)
├── stage 3: iterate-fix-verify             (finding→fix→atomic commit 루프)
├── stage 4: monitor-regressions            (delta-based regression 감지)
└── stage 5: diagnose-bug                   (버그 발견 시 재현→원인→fix)
```

---

## 실행 절차

### Stage 1: 개발 환경 설정

`setup-quality-gates` skill을 invoke해 개발 환경에 quality gates를 설정한다:
- husky + lint-staged (pre-commit)
- Prettier + typecheck + unit test (pre-push)
- secret scan + commitlint

`design-claude-hooks` skill을 invoke해 Claude Code hook이 필요하면 PreToolUse / PostToolUse / Stop / SessionStart hook을 설계한다.

### Stage 2: 병렬 Actor Track 실행

§4 parallel execution plan을 기반으로 `dispatch-parallel-agents` skill을 invoke해 actor track별로 worktree를 격리하고 병렬 agent를 분배한다.

각 agent는 `build-with-tdd` skill을 invoke해 TDD 루프로 구현한다:
```
red: failing test 먼저 작성
green: test를 통과하는 최소 구현
refactor: test 통과 상태 유지하며 개선
```

**Synchronization point**: actor 간 contract 의존성이 있는 지점에서 다른 track의 완료를 기다린다.

### Stage 3: Fix-Verify 루프

각 finding에 대해 `iterate-fix-verify` skill을 invoke한다:
- finding 하나씩 fix
- atomic commit
- re-verify (모든 test 통과 확인)

`freeze-edit-scope` skill을 invoke해 특정 디렉토리 외부 수정을 방지한다.

### Stage 4: Regression 감지

`monitor-regressions` skill을 invoke해 새 코드가 기존 동작을 깨지 않는지 delta-based로 감지한다.

### Stage 5: Bug 대응

버그 발견 시 `diagnose-bug` skill을 invoke한다:
1. Minimized repro 먼저 만든다
2. 여러 hypothesis를 세운다
3. Targeted instrumentation으로 hypothesis를 구분한다
4. Fix 후 regression test를 추가한다
5. 원래 재현 경로도 재검증한다

`consult-codex` skill로 second opinion이 필요하면 외부 LLM에 challenge한다.

---

## MCP / Agent / Hook 지원

이 phase에서 buddy가 지원하는 자동화:

| 도구 | 역할 | Skill |
|------|------|-------|
| **Hook** | PreToolUse: Edit/Write 전 scope 확인 | `freeze-edit-scope`, `guard-destructive-commands` |
| **Hook** | PostToolUse: test 자동 실행 | `design-claude-hooks` |
| **Agent** | actor track별 병렬 구현 | `dispatch-parallel-agents` |
| **MCP** | feature registry 조회/업데이트 | `query-feature-registry` |

---

## Commit Policy

### Commit 트리거 조건 (모두 충족 시 즉시 commit)

구현 작업 중 다음 조건이 **모두 통과**하면 commit을 실행한다. 조건 중 하나라도 실패하면 fix 후 재확인.

```
1. unit test    : all pass (0 failures, 0 errors)
2. lint         : 0 warnings or errors
3. typecheck    : 0 type errors (해당 언어의 정적 분석 포함)
4. build        : 빌드 성공 (artifact 생성 확인)
```

조건 확인 명령 (언어별):
```bash
# Go
go test ./... && go vet ./... && go build ./...

# Node/TypeScript
npm test && npm run lint && npx tsc --noEmit && npm run build

# Python
pytest && ruff check . && mypy .
```

### Commit 단위 원칙

- **atomic commit**: 하나의 논리적 변경만. test + implementation을 함께 commit.
- **task 단위**: `plan-build`의 task_id 하나 = commit 하나가 기본.
- **WIP commit 금지**: 위 4가지 조건 미통과 상태로 commit 금지.
- **이유 없는 `--no-verify` 금지**: pre-commit hook 우회는 명시적 사유가 있을 때만.

### Commit Message 형식 (Conventional Commits)

```
<type>(<scope>): <subject>

<body> (선택)
```

type:
- `feat`: 새 기능
- `fix`: 버그 수정
- `refactor`: 기능 변경 없는 코드 개선
- `test`: 테스트 추가/수정
- `docs`: 문서 변경
- `chore`: 빌드/설정/의존성 변경
- `perf`: 성능 개선

예시:
```
feat(auth): add email/password signup endpoint

- POST /auth/signup validates email format and password strength
- bcrypt hash stored in users table (cost=12)
- JWT issued with 24h expiry
```

### §5 내 Commit 빈도

TDD 루프 기준:
```
red   → (failing test commit은 선택 — 팀 정책에 따라)
green → commit (test + minimal implementation)
refactor → commit (if non-trivial)
```

actor track별로 최소 task 1개 완료 시 1 commit.

---

## 완료 기준

- [ ] 모든 actor track의 task 완료
- [ ] unit test 통과 (TDD 루프 green)
- [ ] integration test 통과 (cross-actor 흐름)
- [ ] lint / typecheck / build 모두 통과
- [ ] regression 없음
- [ ] 모든 변경사항 commit 완료 (uncommitted 없음)
- [ ] PR-ready branch 준비

---

## 다음 phase

- `/buddy:verify-quality` — §6 Quality (권장)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §4 §5
