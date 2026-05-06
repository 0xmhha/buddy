---
description: 증상 반응이 아닌 재현 가능한 원인 분석으로 버그 추적.
argument-hint: "<버그 증상 또는 재현 단계>"
---

# /buddy:diagnose-bug

증상 반응이 아닌 재현 가능한 원인 분석으로 버그 추적.

이 command는 `diagnose-bug` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/diagnose-bug/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `diagnose-bug`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
