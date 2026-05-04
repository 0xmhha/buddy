---
description: YC 스타일 아이디어 검증 인터뷰 — 6 forcing question으로 product idea를 stress-test.
argument-hint: "<아이디어 1-3 paragraph>"
---

# /buddy:validate-idea

YC 스타일 아이디어 검증 인터뷰 — 6 forcing question으로 product idea를 stress-test.

이 command는 `validate-idea` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/validate-idea/SKILL.md`

## 실행 지시

`Skill` 도구로 `validate-idea` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
