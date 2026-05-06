---
description: feature/task 완료 후 commit → branch push → PR 생성 자동화.
argument-hint: "[<PR 제목 또는 비고>]"
---

# /buddy:auto-create-pr

feature/task 완료 후 commit → branch push → PR 생성 자동화.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `auto-create-pr`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
