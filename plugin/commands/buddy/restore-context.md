---
description: context-save가 저장한 most recent work checkpoint를 cross-branch로 load.
argument-hint: "[<체크포인트 라벨 또는 ID>]"
---

# /buddy:restore-context

context-save가 저장한 most recent work checkpoint를 cross-branch로 load.

이 command는 `restore-context` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/restore-context/SKILL.md`

## 실행 지시

`Skill` 도구로 `restore-context` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
