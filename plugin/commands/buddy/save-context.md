---
description: 현재 git state, decisions, remaining work를 checkpoint로 저장 — branch가 달라도 future session이 이어받게.
argument-hint: "[<체크포인트 라벨>]"
---

# /buddy:save-context

현재 git state, decisions, remaining work를 checkpoint로 저장 — branch가 달라도 future session이 이어받게.

이 command는 `save-context` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/save-context/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `save-context`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
