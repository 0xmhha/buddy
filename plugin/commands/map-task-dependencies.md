---
description: task DAG 작성 — internal (intra-track) + cross-actor (contract-based) edges + cycle 감지 + critical path + parallel-safe levels.
argument-hint: "<task list 입력 또는 feature 이름>"
---

# /buddy:map-task-dependencies

task DAG 작성 — internal (intra-track) + cross-actor (contract-based) edges + cycle 감지 + critical path + parallel-safe levels.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `map-task-dependencies`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
