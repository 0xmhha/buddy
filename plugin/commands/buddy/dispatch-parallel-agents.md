---
description: feature/task를 worktree로 격리해 Sonnet worker agent에 병렬 분배하고 결과를 aggregate.
argument-hint: "<feature 목록 또는 plan 경로>"
---

# /buddy:dispatch-parallel-agents

feature/task를 worktree로 격리해 Sonnet worker agent에 병렬 분배하고 결과를 aggregate.

이 command는 `dispatch-parallel-agents` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/dispatch-parallel-agents/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `dispatch-parallel-agents`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
