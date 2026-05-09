---
description: feature 또는 task 를 worktree 로 격리해 병렬 worker agent 에 분배하고 결과 집계.
argument-hint: "<feature 목록 또는 plan 경로>"
disable-model-invocation: true
---

# /buddy:dispatch-parallel-agents

feature/task를 worktree로 격리해 Sonnet worker agent에 병렬 분배하고 결과를 aggregate.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `dispatch-parallel-agents`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
