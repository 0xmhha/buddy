---
description: feature/task 완료 후 commit → branch push → PR 생성 자동화.
argument-hint: "[<PR 제목 또는 비고>]"
---

# /buddy:auto-create-pr

feature/task 완료 후 commit → branch push → PR 생성 자동화.

이 command는 `auto-create-pr` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/auto-create-pr/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `auto-create-pr`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
