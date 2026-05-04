---
description: multi-stage review pipeline 자동 실행 — CEO/design/eng/DX 리뷰를 순차로 돌려 plan을 finalize.
argument-hint: "<plan 경로 또는 설명>"
---

# /buddy:autoplan

multi-stage review pipeline 자동 실행 — CEO/design/eng/DX 리뷰를 순차로 돌려 plan을 finalize.

이 command는 `autoplan` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/autoplan/SKILL.md`

## 실행 지시

`Skill` 도구로 `autoplan` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
