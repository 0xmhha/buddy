---
description: multi-tenant 격리 전략 — RLS / schema-per / DB-per 3 모델 + 3 layer defense + compliance scope.
argument-hint: "<project name>"
---

# /buddy:design-tenant-model

multi-tenant 격리 전략 — shared (RLS) vs schema-per vs DB-per 3 모델 + 3 layer defense + onboarding cost + compliance scope.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `design-tenant-model`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
