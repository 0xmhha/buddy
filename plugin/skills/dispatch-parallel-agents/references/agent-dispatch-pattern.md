# Agent Dispatch 패턴 — actor-track별 병렬 실행

§4 plan-build가 생성한 actor-track plan을 기반으로 Claude Code Agent tool을 병렬 분배하는 완전 절차.

---

## 1. 입력 포맷 (§4 산출물)

`docs/actor-track-plan.yaml` 예시:

```yaml
feature: signup-email-password
base_branch: main
tracks:
  - actor_id: frontend
    system_boundary: frontend-spa
    tasks:
      - id: FE-1
        description: 회원가입 폼 컴포넌트 구현 (이메일/비밀번호 입력 + 검증)
        dependencies: []
      - id: FE-2
        description: 회원가입 성공/실패 상태 처리 + redirect
        dependencies: [FE-1]
    contracts_needed:
      - from: backend
        artifact: api-contract.yaml
        endpoint: POST /auth/signup

  - actor_id: backend
    system_boundary: backend-auth-service
    tasks:
      - id: BE-1
        description: POST /auth/signup 엔드포인트 구현
        dependencies: []
      - id: BE-2
        description: bcrypt 해싱 + DB 저장
        dependencies: [BE-1]
    contracts_produced:
      - artifact: api-contract.yaml
        endpoint: POST /auth/signup

  - actor_id: integration
    system_boundary: external-saas
    tasks:
      - id: INT-1
        description: SendGrid 이메일 검증 웹훅 연동
        dependencies: [BE-1]
```

---

## 2. system_boundary → subagent_type 매핑

| system_boundary | 권장 subagent_type | 대안 |
|----------------|-------------------|------|
| `frontend-spa` | `senior-frontend` | `feature-dev:code-architect` |
| `frontend-mobile` | `senior-frontend` | `feature-dev:code-architect` |
| `backend-*` | `senior-backend` | `feature-dev:code-architect` |
| `backend-auth-service` | `senior-backend` | `senior-secops` (인증 코드) |
| `external-saas` | `senior-devops` | `senior-backend` |
| `data-pipeline` | `senior-data-engineer` | `senior-backend` |
| `infra-*` | `senior-devops` | — |
| (기본) | `feature-dev:code-architect` | — |

---

## 3. Worktree 생성

```bash
# actor-track-plan.yaml에서 actor 목록 추출
ACTORS=$(yq e '.tracks[].actor_id' docs/actor-track-plan.yaml | tr '\n' ' ')

# worktree 생성
bash plugin/skills/dispatch-parallel-agents/scripts/create-worktrees.sh main $ACTORS
# 결과:
#   ../myrepo-actor-frontend  (branch: actor/frontend)
#   ../myrepo-actor-backend   (branch: actor/backend)
#   ../myrepo-actor-integration (branch: actor/integration)
```

---

## 4. Agent tool 병렬 호출 형식

Claude Code Agent tool 호출 시 각 actor-track의 컨텍스트를 완전히 포함:

```
Agent(
  subagent_type: "senior-backend",
  isolation: "worktree",    // 이미 worktree 생성했으면 생략
  description: "actor/backend track 구현 — signup 엔드포인트",
  prompt: """
  ## 작업 컨텍스트
  Repository: /path/to/myrepo-actor-backend
  Branch: actor/backend
  Feature: signup-email-password
  Actor: backend (system_boundary: backend-auth-service)

  ## 구현할 Tasks
  - BE-1: POST /auth/signup 엔드포인트 구현
  - BE-2: bcrypt 해싱 + DB 저장 (BE-1 의존)

  ## Contract 산출물 (완료 후 생성 필요)
  - docs/api-contract.yaml: POST /auth/signup 명세
    (frontend actor가 이 파일을 기다림)

  ## 완료 기준
  - [ ] unit test 통과 (go test ./... -count=1)
  - [ ] lint 통과 (golangci-lint run)
  - [ ] docs/api-contract.yaml 생성
  - [ ] actor/backend 브랜치에 commit

  ## Commit Policy (build-feature 참조)
  - Conventional Commits: feat(auth): add signup endpoint
  - 4-gate 통과 후만 commit (test + lint + typecheck + build)
  - WIP commit 금지
  """
)
```

---

## 5. 동기화 포인트 (Synchronization Point)

contract 의존성이 있는 지점에서 다음 track이 시작되기 전 대기.

```
동기화 포인트 선언 형식:
SYNC_POINT: {
  waiting_actor: "frontend",
  blocking_actor: "backend",
  artifact: "docs/api-contract.yaml",
  check_command: "test -f docs/api-contract.yaml"
}
```

실제 대기 방법 (Claude가 직접 수행):
1. backend agent가 `docs/api-contract.yaml` 생성 완료 알림
2. frontend agent dispatch 시작
3. frontend agent prompt에 contract 파일 경로 포함

자세한 내용: `references/synchronization.md`

---

## 6. 완료 후 병합

```bash
# 모든 actor-track 완료 후
bash plugin/skills/dispatch-parallel-agents/scripts/cleanup-worktrees.sh main frontend backend integration
```

병합 충돌이 발생하면:
1. 충돌 파일 수동 해결
2. `git add <파일>` + `git commit`
3. cleanup-worktrees.sh 재실행 (--force 옵션)
