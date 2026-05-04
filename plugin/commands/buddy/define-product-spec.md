---
description: 검증 결과를 공식 PRD(Product Requirements Document)로 고정.
argument-hint: <검증 산출물 또는 컨텍스트>
---

# /buddy:define-product-spec

검증 결과를 공식 PRD(Product Requirements Document)로 고정.

이 command는 `define-product-spec` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/define-product-spec/SKILL.md`

## 실행 지시

`Skill` 도구로 `define-product-spec` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
