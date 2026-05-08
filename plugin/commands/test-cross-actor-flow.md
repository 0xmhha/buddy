---
description: cross-actor flow E2E test — multi-actor chain (signup→email→verify→login 등) full-stack 검증. cross-actor edge coverage + contract drift detection.
argument-hint: "<feature name> v<version>"
---

# /buddy:test-cross-actor-flow

cross-actor flow E2E test — actor 협업 시나리오 (signup → email → verify → login → me) full-stack 검증. cross-actor edge coverage + contract drift detection.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `test-cross-actor-flow`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
