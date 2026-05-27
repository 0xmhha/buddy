# publish-to-tracker — 내부 plan/spec → 외부 issue tracker 발행

## 1. 목적

buddy 내부의 *feature spec* (§2 `define-features` 산출물) 또는 *task plan* (§4 `plan-build` 산출물) 을 외부 issue tracker (GitHub Issues / Linear / Jira) 로 발행. *수동 동기화* 제거.

두 모드:
- **`prd`**: §2 feature spec → PRD 1건 발행 (stakeholder communication, ready-for-agent 큰 그림)
- **`issues`**: §4 task plan → tracer-bullet vertical-slice issue N건 발행 (dispatch-parallel-agents 의 grabbable surface)

> 흡수 출처: mattpocock-skill `to-prd` + `to-issues` 의 PRD template + tracer-bullet vertical-slice rule + HITL/AFK 구분을 *adopt-with-edits* (ADR-003 §2.1). multi-tracker 추상화는 *inspired-by* — buddy 자체 발명 (mattpocock 은 단일 tracker 가정).


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Task plan 또는 PRD | ✅ | artifact | Phase 4 산출물 또는 Phase 1 PRD | 먼저 구현 계획을 수립하세요 |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| 외부 tracker issues (GitHub/Linear/Jira) | artifact | issue links | 프로젝트 관리, `dispatch-parallel-agents` |

## 2. 사용 시점

### `prd` 모드
- §2 `define-features` 완료 후 stakeholder 공유 필요 시
- 비기술 stakeholder (PM / 마케팅 / legal) 가 feature 의 *비기술 view* 필요
- AI agent 가 큰 그림을 grab 해 *sub-issue 발행* 의 parent reference 로 활용

### `issues` 모드
- §4 `plan-build` 완료 후 (autoplan review 통과 권장)
- `dispatch-parallel-agents` 호출 직전 — agent 가 *외부 tracker 의 grabbable issue* 를 가져갈 surface 가 필요
- 팀 collaboration 시점 — buddy 외부 인력 (다른 팀 / contractor) 이 작업 참여
- 부분 발행: 일부 vertical slice 만 외부 위탁 (HITL slice) + 나머지는 buddy 내부 처리 (AFK slice)

### 사용 부적합
- 1인 개발 + buddy 단독 사용: 외부 tracker 불필요
- spec 미확정: §2 / §4 산출물이 *임시 draft* 면 발행 후 수정 비용 ↑

## 3. 입력

### 공통
- `--mode`: `prd` 또는 `issues`
- `--tracker`: `github` (기본) / `linear` / `jira` — env `BUDDY_TRACKER` 로 default override
- `--dry-run`: 발행 없이 *발행 예정 본문* 만 출력 (검토 용도)

### `prd` 모드
- 필수: §2 feature spec 경로 (`docs/features/<name>.yaml` 또는 `define-features` 산출 yaml)
- 선택: parent issue (epic / theme — 있으면 link)
- 선택: `--label` 추가 라벨 (default: `prd`, `ready-for-agent`)

### `issues` 모드
- 필수: §4 task plan 경로 (`docs/plans/<feature>.yaml` 또는 `plan-build` 산출 yaml — vertical slice 분해 결과)
- 필수: parent PRD issue ID (prd 모드로 먼저 발행한 결과)
- 선택: `--start-from=<slice-id>` 일부만 발행. *주의*: 해당 slice 의 모든 `blocked_by` 가 *이미 tracker 에 발행된 상태* 여야 함. 미발행 blocker 발견 시 STOP — 사용자에게 "blocker 먼저 발행" 안내 후 종료. dangling reference 차단.
- 선택: `--hitl-only` / `--afk-only` 한 타입만 발행

### 환경 변수
- `BUDDY_TRACKER`: 기본 tracker
- `GITHUB_TOKEN` (github) / `LINEAR_API_KEY` (linear) / `JIRA_API_TOKEN` + `JIRA_BASE_URL` (jira)
- `BUDDY_TRACKER_PROJECT`: GitHub repo (`owner/repo`) / Linear team / Jira project key

> *시크릿 처리*: env 변수 또는 secret manager 만. 코드 / 로그 / 에러 메시지 노출 금지 (`design-secret-management` 와 정합).

## 4. 핵심 원칙

1. **의존성 순서 발행** — blocker issue 먼저 publish 후 *실제 issue ID* 를 후속 issue 의 `Blocked by` 에 채움. 역순 발행 시 dangling reference.
2. **HITL ↔ AFK 라벨 정확성** — HITL (Human-In-The-Loop) 은 *architectural decision / design review* 필요 슬라이스. AFK (Away-From-Keyboard) 는 *agent 가 자율 완료 가능*. 라벨 오분류 시 dispatch-parallel-agents 가 잘못된 agent 분배.
3. **`ready-for-agent` 라벨** — AFK slice 발행 시 자동 부여. 외부 AI agent (Claude Code agent / Devin 등) 가 grab 할 수 있는 *self-contained* 형식 보증.
4. **vertical slice 강제** — 각 slice 는 *모든 layer* (schema → API → UI → test) 를 *얇게 관통*. horizontal layer-only slice 금지.
5. **file 경로 / 코드 snippet 금지** (예외: prototype 산출 — state machine, reducer, schema, type shape) — 코드는 빠르게 stale, prose 는 살아남음.
6. **tracker 추상화 의무** — GitHub-only / Linear-only 코드 작성 금지. 모든 발행은 *tracker adapter interface* 통과 (§7 참조).
7. **parent reference 양방향** — `prd` 발행 후 PRD issue ID 를 *모든 child issue* 에 명시. PRD issue 본문에는 child issue ID 목록 *역참조* 갱신 (모두 발행 완료 후 1회).
8. **dry-run 우선** — 첫 실행은 `--dry-run` 권장. 다중 issue 일괄 발행 후 rollback 어려움.
9. **anti-rationalization 게이트 정합** — *발행 발화 직전* [`router/references/verification-discipline.md`](../router/references/verification-discipline.md) Iron Law 적용 — N개 issue 발행했다 발화는 *gh / linear / jira API 호출 출력 확인 후* 에만.

## 5. 실행 절차

### 5.1 `prd` 모드

#### Step 1: 입력 검증
- feature spec yaml 로드 + 필수 필드 (actor / use case / system boundary) 확인
- tracker auth 확인 (gh auth status / linear API ping / jira API /myself)

#### Step 2: PRD 본문 생성

다음 7 섹션 (mattpocock template adopt-with-edits, buddy 어휘 재진술):

```markdown
## Problem Statement
<§2 feature spec 의 "actor 가 처한 문제" 를 사용자 시점에서 진술>

## Solution
<sojution 1줄 + 어떤 actor 의 어떤 use case 가 어떻게 해결되는지>

## User Stories
<long numbered list: "As an <actor>, I want <feature>, so that <benefit>">
1. ...
2. ...

## Implementation Decisions
- 영향 받는 system boundary (§2 결과 인용)
- 도입할 deep module 후보 (작은 interface + 큰 functionality)
- §3 design 산출물에 인용된 ADR (link)
- API contract / schema change 핵심 항목

> 파일 경로 / 코드 snippet 금지. 예외: prototype 산출 — state machine / reducer / schema / type shape.

## Testing Decisions
- per-actor test layer (`test-per-actor-use-case` 매핑)
- cross-actor integration (`test-cross-actor-flow` 매핑)
- prior art (codebase 의 유사 test pattern 인용)

## Out of Scope
- §2 backlog 의 *deferred / explicitly-not-this-feature* 항목

## Further Notes
- §3 design ADR 결정 근거 요약
- 알려진 risk (assess-business-viability 산출 인용)
```

#### Step 3: tracker 발행

```bash
# github 예시
gh issue create \
  --title "PRD: <feature-name>" \
  --body-file /tmp/prd-<feature>.md \
  --label "prd,ready-for-agent" \
  --repo $BUDDY_TRACKER_PROJECT
```

발행 결과 issue ID (e.g. `#247`) 를 *상태 파일* 에 저장 — `issues` 모드 가 parent 로 참조.

### 5.2 `issues` 모드

#### Step 1: 입력 검증
- task plan yaml 로드 + vertical-slice 분해 결과 + dependency edges 확인
- parent PRD issue ID 필수 — 누락 시 `prd` 모드 선행 호출 권유

#### Step 2: 슬라이스 분류

각 vertical slice 에 대해:

| 필드 | 값 | 결정 기준 |
|------|----|---------|
| `type` | HITL / AFK | architectural decision / design review 필요? → HITL. 그 외 → AFK |
| `blocked_by` | [issue-id, ...] | task DAG (§4 `map-task-dependencies`) 의 edge 반영 |
| `acceptance` | checklist | §4 `define-acceptance-test-plan` 의 per-slice criteria |
| `labels` | `vertical-slice` + `<actor-track>` + (AFK 면 `ready-for-agent`) | dispatch-parallel-agents 가 actor 별 grab |

#### Step 3: 발행 순서 결정

topological sort (`map-task-dependencies` 의 critical path 활용):
- depth 0 (blocker 없음) 먼저 publish
- 발행 결과 issue ID 를 *후속 slice 의 `Blocked by` 필드* 에 binding
- 모두 발행 완료 후 *parent PRD issue 본문* 에 child issue ID 목록 1회 갱신

#### Step 4: Issue 본문 생성 (slice 당)

```markdown
## Parent

#<PRD issue ID>

## What to build

<concise vertical-slice 설명: schema → API → UI → test 의 *얇은 end-to-end*>

> 파일 경로 / 코드 snippet 금지. 예외: prototype 산출 (state machine / reducer / schema / type shape).

## Acceptance criteria

- [ ] <criterion 1 — define-acceptance-test-plan 산출>
- [ ] <criterion 2>
- [ ] <criterion 3>

## Blocked by

- #<blocker issue ID> ("Blocker title")
또는
- None — can start immediately

## Slice metadata

- Actor track: <frontend / backend / 3rd-party>
- Type: <HITL / AFK>
- Estimated hours: <N>
```

#### Step 5: 일괄 발행 + parent 역참조 갱신

```bash
# github 예시 — depth 0 슬라이스부터
gh issue create --title "<slice title>" --body-file /tmp/slice-001.md \
  --label "vertical-slice,backend,ready-for-agent" \
  --repo $BUDDY_TRACKER_PROJECT

# 발행 결과 issue ID 캡처 → 후속 slice 의 Blocked by 에 binding

# 모두 발행 완료 후 parent PRD 갱신
gh issue edit <PRD-issue-ID> --body "<원래 본문>\n\n## Child issues\n- #<id1>\n- #<id2>\n..."
```

> 발행 실패 시 부분 rollback: 이미 발행된 issue 를 `gh issue close --reason 'not planned'` (delete 권한 없음). 그래서 *dry-run 우선* 강조 (§4 원칙 8).

## 6. Tracker 추상화

각 adapter 는 다음 4 메서드 구현 의무:

| 메서드 | 입력 | 출력 |
|--------|------|------|
| `auth_check()` | env | success / error |
| `issue_create(title, body, labels)` | str / str / list[str] | issue ID + URL |
| `issue_edit(id, body)` | str / str | success |
| `issue_link(parent_id, child_id)` | str / str | success (tracker 가 native link 지원 시) / no-op |

내장 adapter:

| Tracker | CLI / API | auth |
|---------|----------|------|
| GitHub | `gh` CLI | `gh auth login` 또는 `GITHUB_TOKEN` |
| Linear | REST API | `LINEAR_API_KEY` |
| Jira | REST API | `JIRA_API_TOKEN` + `JIRA_BASE_URL` |

추가 tracker (Asana / ClickUp / GitLab Issues) 는 같은 interface 로 확장.

## 7. 산출물 (호출자 반환)

```yaml
publish_result:
  mode: prd | issues
  tracker: github | linear | jira
  prd:
    id: "#247"
    url: "https://github.com/owner/repo/issues/247"
  issues:
    - slice_id: "backend-1"
      issue_id: "#248"
      url: "..."
      labels: ["vertical-slice", "backend", "ready-for-agent"]
      blocked_by: []
    - slice_id: "frontend-2"
      issue_id: "#249"
      url: "..."
      labels: ["vertical-slice", "frontend"]
      blocked_by: ["#248"]
  parent_back_reference_updated: true
  total_published: N
  dry_run: false
```

> *완료 발화 의무*: 본 yaml 의 `total_published` / `parent_back_reference_updated` 값은 *실제 tracker API 응답 확인 후* 에만 채움 (verification-discipline Iron Law).

## 8. 검증

- [ ] tracker auth 확인 후 발행 시작
- [ ] dry-run 결과 사용자 confirm 후 실제 발행 (대량 발행 시)
- [ ] 의존성 순서 준수 (blocker issue ID 가 `Blocked by` 에 실제 채워짐)
- [ ] HITL / AFK 라벨 정확
- [ ] AFK slice 에 `ready-for-agent` 라벨 부여
- [ ] parent PRD 의 child issue 역참조 갱신 (issues 모드 완료 시)
- [ ] 시크릿 노출 없음 (`design-secret-management` 검증)
- [ ] 발행 발화 (`N개 issue 발행 완료`) 직전 Iron Law 적용 — API 응답 확인 후 발화

## 9. 다음 phase

- `prd` 모드 발행 후 → `issues` 모드 호출 (§4 plan-build 완료 시)
- `issues` 모드 발행 후 → [`dispatch-parallel-agents`](../dispatch-parallel-agents/PROCEDURE.md) 호출 (외부 tracker 의 issue 가 agent 의 grabbable surface)
- §5 `build-feature` 본 단계 진입 — agent 가 issue 를 grab 하여 actor track 별 구현

## 10. Anti-patterns

1. **역순 발행** — 후속 slice 를 먼저 publish → `Blocked by` 가 dangling reference
2. **HITL/AFK 오분류** — architectural decision 슬라이스를 AFK 라벨링 → agent 가 자체 결정 시도 → 잘못된 architecture lock-in
3. **`ready-for-agent` 남발** — HITL slice 에도 부여 → agent 가 grab 후 잘못된 결정
4. **file 경로 / 코드 snippet 본문 포함** — 코드 stale 시 issue 도 stale
5. **수동 tracker API 호출** — adapter 우회, 다른 tracker 로 이전 시 재작성
6. **dry-run 생략** — 대량 발행 후 부분 rollback 어려움
7. **시크릿 코드 / 로그 노출** — `GITHUB_TOKEN` 등 출력
8. **parent back-reference 누락** — child issue 만 발행, parent PRD 에 child 목록 갱신 안 함 → audit trail 단절
9. **Iron Law 위반 발화** — API 호출 미실행 상태에서 "N개 발행 완료" 발화
10. **tracker 추상화 단일 lock-in** — 모든 발행 코드를 `gh` CLI hardcode → Linear 이전 시 재작성

## 11. 참조

- 흡수 출처: mattpocock-skill `engineering/to-prd/SKILL.md` + `engineering/to-issues/SKILL.md` (adopt-with-edits, ADR-003 §2.1)
- 정합: [`define-features/PROCEDURE.md`](../define-features/PROCEDURE.md) (§2 입력 시점), [`plan-build/PROCEDURE.md`](../plan-build/PROCEDURE.md) (§4 입력 시점), [`dispatch-parallel-agents/PROCEDURE.md`](../dispatch-parallel-agents/PROCEDURE.md) (발행 후 cascade), [`router/references/verification-discipline.md`](../router/references/verification-discipline.md) (완료 발화 Iron Law)
- ADR-003 §2.1 (adopt-with-edits 정책), §2.4 (verbatim 0건 유지)
- buddy 9-phase 라이프사이클 spec: `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`
