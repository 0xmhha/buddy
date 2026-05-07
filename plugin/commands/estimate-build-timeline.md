---
description: critical path 기반 일정 합성 — confidence interval (best/expected/p90/worst) + risk buffer + holiday/availability 반영.
argument-hint: "<task DAG / batch schedule / start date>"
---

# /buddy:estimate-build-timeline

critical path 기반 일정 합성 — confidence interval (best/expected/p90/worst) + risk buffer + holiday/availability 반영.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `estimate-build-timeline`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
