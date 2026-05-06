---
name: router
description: "Use when a buddy command requests dispatch to a target PROCEDURE. Reads `${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md` and executes its instructions."
type: skill
---

# Buddy Router

Buddy plugin 내부 라우터. 모든 `/buddy:*` slash command가 이 skill을 호출하며, command가 전달한 `target PROCEDURE` 이름으로 해당 절차 파일을 읽어 그대로 수행한다.

## Dispatch contract

각 buddy command md 의 본문은 router 호출 시 다음을 명시한다:
- `mode:` `single` (기본) / `chain` / `parallel`
- `target PROCEDURE:` (single 모드) — 1 개 skill 이름
- `targets:` (chain·parallel 모드) — 콤마 분리 skill 이름 리스트
- 사용자 인자: `$ARGUMENTS` (command 가 받은 원본 인자 그대로)

router 는 이 4 개 필드를 command md 의 invocation block 에서 읽고, 그 외 frontmatter 는 참조하지 않는다.

## How to dispatch (single)

입력으로 다음을 받는다:

- `target PROCEDURE: <name>` — 실행할 절차의 이름 (Buddy skill 디렉토리 이름과 동일)
- 사용자 인자 — command가 그대로 전달한 `$ARGUMENTS` 본문

수행 절차:

1. `Read` 도구로 `${CLAUDE_PLUGIN_ROOT}/skills/<name>/PROCEDURE.md` 파일을 로드한다. (`${CLAUDE_PLUGIN_ROOT}` 는 buddy plugin 설치 경로 — hardcode 금지)
2. 로드한 PROCEDURE 본문을 그 자체로 실행 지시문으로 취급한다 — 사용자 인자를 PROCEDURE의 입력으로 사용한다.
3. PROCEDURE가 요구하는 모든 단계를 누락 없이 수행한다.
4. PROCEDURE 파일이 존재하지 않으면 즉시 에러 메시지를 반환하고 임의로 추론하지 않는다.

## How to dispatch (chain mode)

Trigger: command md 가 `mode: chain` 과 `targets: name1, name2, name3` (또는 동등한 콤마 구분 형식)을 전달한 경우.

수행 절차:

1. `targets` 문자열을 콤마(`,`)로 분리해 순서 보존 리스트로 만든다. 각 이름의 앞뒤 공백을 trim 한다. 빈 토큰은 제거한다.
2. 리스트의 각 target 에 대해 **순서대로** 다음을 수행한다:
   1. `Read` 도구로 `${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md` 를 로드한다.
   2. 로드한 PROCEDURE 본문을 실행한다. 입력 컨텍스트로는 (a) 사용자 인자 원본 + (b) 직전까지 실행된 모든 step 의 산출물을 함께 제공한다.
   3. 해당 step 이 만든 산출물(요약·결정·생성 파일 경로 등)을 현재 대화 컨텍스트에 보존해, 다음 target 이 이를 입력으로 참조 가능하게 한다.
      - 만약 PROCEDURE 가 명시 산출물을 정의하지 않으면(side-effecting skill: `auto-create-pr`, `ship-release` 등), 실행 결과 한 줄 요약(성공 / 실패 + 변경된 외부 상태 식별자: PR URL, git tag, 배포 ID 등)을 산출물로 간주한다.
3. 어느 target 의 PROCEDURE.md 가 존재하지 않으면 즉시 중단하고 어떤 target 이 실패했는지 명확히 보고한다 — 임의 추론·skip 금지.
4. 마지막 target 까지 완료되면, 각 step 이 무엇을 산출했는지 한 줄씩 요약한 최종 리포트를 출력한다 (step 순서 유지).

## How to dispatch (parallel mode)

Trigger: command md 가 `mode: parallel` 과 `targets: name1, name2, name3` 을 전달한 경우.

수행 절차:

1. `targets` 문자열을 콤마(`,`)로 분리해 리스트로 만든다. 각 이름의 앞뒤 공백을 trim 한다. 빈 토큰은 제거한다.
2. 각 target 마다 `Agent` 도구로 fresh subagent 를 하나씩 디스패치한다 (`subagent_type: general-purpose`). 각 subagent 에게 다음을 전달한다:
   - 지시: `${CLAUDE_PLUGIN_ROOT}/skills/<their-target>/PROCEDURE.md` 를 `Read` 로 로드해 본문 절차를 그대로 수행할 것.
   - 공유 사용자 인자(원본 `$ARGUMENTS`).
   - 제약: 다른 target 의 작업물·파일을 수정하지 말 것. 자기 결과만 보고로 반환할 것.
3. 모든 subagent 가 완료될 때까지 대기한다. 각 subagent 는 자기 PROCEDURE 의 실행 리포트를 반환한다.
4. 결과 집계: target 별로 그룹핑해 결과를 제시한다. 그 후 cross-target 관찰(상호 모순, 공통 finding, 시너지)이 있으면 별도 단락으로 합성한다.
5. 어느 subagent 가 PROCEDURE 부재로 실패하면 그 target 만 실패로 표기하고, 나머지 결과는 그대로 보고한다 — 전체 중단 금지.

## Skill index

Buddy 는 9-phase 라이프사이클로 78 개 skill 을 조직한다. 각 phase 는 진입점 orchestrator 와 그 phase 안의 stage skill 집합을 가진다.

| Phase | Orchestrator (entry) | 1줄 설명 |
|-------|----------------------|---------|
| §1 Idea & Business Validation | `concretize-idea` | idea/concept → PRD + 사업성 검증 |
| §2 Feature Definition & Backlog | `define-features` | PRD → actor / use case / system boundary → feature backlog |
| §3 Technical Design | `design-system` | feature backlog → tech stack ADR + infra + API + data model |
| §4 Implementation Plan | `plan-build` | technical design → actor 별 task graph + 병렬 실행 plan |
| §5 Development | `build-feature` | implementation plan → working code + tests (TDD + parallel agents) |
| §6 Quality | `verify-quality` | code complete → QA + security + compliance sign-off |
| §7 Release & Beta | `ship-release` | quality gate pass → tagged release + UAT + GA |
| §8 Operate & Iterate | `iterate-product` | production traffic → A/B + funnel + improvement backlog |
| §9 Lifecycle Management | `manage-lifecycle` | feature/product 노후화 → deprecation + migration + EOL |

Cross-phase 보조:

- `autoplan` — 어느 phase 의 산출물(plan/PRD/ADR/task plan)에든 호출 가능한 4-mode review (review-scope → review-engineering → review-design → review-devex 순차).
- `consult-codex`, `save-context`, `restore-context` — phase 종속 없는 공통 도구.

라우팅 결정 워크플로우:

1. 사용자 발화·command 의 진입 조건이 어느 phase 에 속하는지 위 표에서 식별한다.
2. 그 phase 의 orchestrator 를 기본 dispatch target 으로 둔다 — 사용자가 stage 단독을 명시하지 않은 한 orchestrator 우선.
3. lazy-load 트리거:
   - 사용자 발화 또는 command name 이 위 9-phase 표의 entry-point skill 1개와 정확히 매칭되면 그 skill 을 dispatch — 인라인 표만으로 충분.
   - 정확 매칭이 없거나, 사용자가 stage skill 명을 직접 언급하거나, command 가 9-phase orchestrator 가 아닌 stage·domain·pattern skill 을 target 으로 지정하면 → `Read ${CLAUDE_PLUGIN_ROOT}/skills/router/references/skill-catalog.md` 를 호출해 전체 카탈로그 확인.
   - 라우팅이 2개 이상의 skill 사이에서 모호하면 → `Read ${CLAUDE_PLUGIN_ROOT}/skills/router/references/routing-rules.md` 의 §3 케이스별 결정 참조.
