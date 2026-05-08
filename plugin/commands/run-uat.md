---
description: UAT scenario 실행 + go/no-go 판단 — designated stakeholder 가 critical flow 를 verify, evidence + sign-off 수집.
argument-hint: "<feature name> v<version>"
---

# /buddy:run-uat

UAT scenario 실행 + go/no-go 판단 — designated stakeholder 가 critical flow 를 verify, evidence + sign-off 수집.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `run-uat`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
