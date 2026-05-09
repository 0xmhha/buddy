---
description: 현재 작업 단계 확인 + 다음에 실행할 명령 안내. 어디서 시작할지 막막할 때 먼저 실행.
argument-hint: ""
disable-model-invocation: true
---

# /buddy:status

현재 작업 단계 확인 + 다음에 실행할 명령 안내. 어디서 시작할지 막막할 때 먼저 실행.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `status`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
