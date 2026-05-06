---
description: context-save가 저장한 most recent work checkpoint를 cross-branch로 load.
argument-hint: "[<체크포인트 라벨 또는 ID>]"
---

# /buddy:restore-context

context-save가 저장한 most recent work checkpoint를 cross-branch로 load.

이 command는 `restore-context` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/restore-context/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `restore-context`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
