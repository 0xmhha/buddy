---
description: 검증 결과를 공식 PRD (Product Requirements Document) 로 고정.
argument-hint: "<검증 산출물 또는 컨텍스트>"
disable-model-invocation: true
---

# /buddy:define-product-spec

검증 결과를 공식 PRD(Product Requirements Document)로 고정.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `define-product-spec`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
