---
description: 검증 결과를 공식 PRD(Product Requirements Document)로 고정.
argument-hint: "<검증 산출물 또는 컨텍스트>"
---

# /buddy:define-product-spec

검증 결과를 공식 PRD(Product Requirements Document)로 고정.

이 command는 `define-product-spec` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/define-product-spec/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `define-product-spec`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
