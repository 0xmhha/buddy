---
description: actor use case 단위 통합 테스트 — frontend E2E + backend integration + 3rd-party contract.
argument-hint: "<feature name> v<version>"
---

# /buddy:test-per-actor-use-case

actor 의 use case 단위 통합 테스트 — frontend (Playwright E2E), backend (Vitest+testcontainers), 3rd-party (Pact). per-actor coverage gap 0 maintain.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `test-per-actor-use-case`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
