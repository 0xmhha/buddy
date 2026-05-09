---
description: actor 별 + cross-actor 완료 기준 + test infra 결정 — feature spec → verifiable test plan (unit / integration / contract / E2E).
argument-hint: "<feature spec 또는 task DAG 입력>"
disable-model-invocation: true
---

# /buddy:define-acceptance-test-plan

actor 별 + cross-actor 완료 기준 + test infra 결정 — feature spec → verifiable test plan (unit / integration / contract / E2E).

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `define-acceptance-test-plan`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
