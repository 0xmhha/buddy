---
description: §5 Development — implementation plan → working code per feature + tests. build-with-tdd, dispatch-parallel-agents, diagnose-bug, iterate-fix-verify를 actor track별로 실행.
argument-hint: "<feature 이름 또는 implementation plan 경로>"
---

# /buddy:build-feature

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `build-feature`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
