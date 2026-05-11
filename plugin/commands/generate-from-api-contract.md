---
description: OpenAPI / GraphQL / gRPC contract → 클라이언트 SDK + 서버 stub + type 자동 생성. CI 자동 동기화 + AUTO-GENERATED 헤더 강제 + wrapper layer 분리.
argument-hint: "<contract file 경로 또는 spec 이름>"
disable-model-invocation: true
---

# /buddy:generate-from-api-contract

design-api-contract 산출 → generator (openapi-generator / GraphQL Code Generator / protoc) → 클라이언트 + 서버 stub. type drift 차단. CI 자동 동기화.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `generate-from-api-contract`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
