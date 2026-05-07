---
description: actor track → ordered task list (atomic units, single PR scope, verifiable acceptance, sized for 1 worker × short period).
argument-hint: "<track table 입력 또는 feature 이름>"
---

# /buddy:decompose-track-to-tasks

actor track → ordered task list (atomic units, single PR scope, verifiable acceptance, sized for 1 worker × short period).

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `decompose-track-to-tasks`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
