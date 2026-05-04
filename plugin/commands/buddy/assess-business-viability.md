---
description: 아이디어가 사업으로 성립하는지 7차원(TAM/SAM/SOM, 고객, WTP, GTM, 경쟁, unit economics, 규제)으로 평가.
argument-hint: "<아이디어 또는 PRD 경로>"
---

# /buddy:assess-business-viability

아이디어가 사업으로 성립하는지 7차원(TAM/SAM/SOM, 고객, WTP, GTM, 경쟁, unit economics, 규제)으로 평가.

이 command는 `assess-business-viability` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/assess-business-viability/SKILL.md`

## 실행 지시

`Skill` 도구로 `assess-business-viability` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
