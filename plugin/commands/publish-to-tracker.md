---
description: 내부 plan/spec 산출물을 외부 issue tracker (GitHub/Linear/Jira) 로 발행. 모드 — `prd` (§2 feature spec → PRD 1건), `issues` (§4 task plan → tracer-bullet vertical slice N건). dispatch-parallel-agents 의 grabbable surface 생성.
argument-hint: "<--mode=prd|issues> <spec/plan 경로> [--tracker=github|linear|jira] [--dry-run] [--hitl-only|--afk-only]"
disable-model-invocation: true
---

# /buddy:publish-to-tracker

내부 plan/spec → 외부 issue tracker 발행. PRD 모드 (§2 feature spec) 또는 issues 모드 (§4 task plan vertical slice). 의존성 순서 발행 + HITL/AFK 라벨 + ready-for-agent surface 생성. tracker abstraction (gh/Linear/Jira). dry-run 우선. 시크릿은 env 변수 (`GITHUB_TOKEN` / `LINEAR_API_KEY` / `JIRA_API_TOKEN`) 만.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `publish-to-tracker`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
