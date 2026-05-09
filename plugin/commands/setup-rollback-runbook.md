---
description: rollback decision tree (언제 rollback / 언제 forward fix) + 실행 절차 + verification — incident response 의 핵심 도구.
argument-hint: "<project name> v<version>"
disable-model-invocation: true
---

# /buddy:setup-rollback-runbook

rollback decision tree (언제 rollback / 언제 forward fix) + 실행 절차 + verification — incident response 의 핵심 도구.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `setup-rollback-runbook`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
