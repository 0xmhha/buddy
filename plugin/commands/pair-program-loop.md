---
description: driver / navigator 역할 분리 + 15min swap + AI agent pair (driver 또는 navigator). TDD red-green-refactor 사이클 정합 + anti-pattern (역할 흐림 / swap X / distraction) 회피.
argument-hint: "<task 이름 또는 acceptance criteria>"
disable-model-invocation: true
---

# /buddy:pair-program-loop

전통 pair programming + AI agent pair 변형. driver = 코드 작성, navigator = 검토 / 다음 step. 사용자 통제 유지 (agent 가 driver+navigator 동시 X). build-with-tdd cascade.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `pair-program-loop`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
