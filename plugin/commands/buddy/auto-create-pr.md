---
description: feature/task 완료 후 commit → branch push → PR 생성 자동화.
argument-hint: [<PR 제목 또는 비고>]
---

# /buddy:auto-create-pr

feature/task 완료 후 commit → branch push → PR 생성 자동화.

이 command는 `auto-create-pr` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/auto-create-pr/SKILL.md`

## 실행 지시

`Skill` 도구로 `auto-create-pr` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
