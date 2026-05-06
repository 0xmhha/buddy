---
description: §4 Implementation Plan — technical design → ordered task graph + actor별 parallel execution plan. feature를 actor track으로 분해하고 task DAG와 build timeline을 생성.
argument-hint: "<feature backlog 또는 technical design 경로>"
---

# /buddy:plan-build

이 command는 `plan-build` skill을 즉시 invoke한다.

- Skill: `plugin/skills/plan-build/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `plan-build`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
