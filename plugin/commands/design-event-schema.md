---
description: async event schema-first 설계 — producer/consumer contract + versioning + DLQ + idempotency. design-api-contract 의 sync-only gap 보강.
argument-hint: "<project name>"
---

# /buddy:design-event-schema

async event schema-first 설계 — producer/consumer contract + versioning + DLQ + idempotency.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `design-event-schema`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
