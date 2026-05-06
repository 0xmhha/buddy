---
description: project tool을 auto-detect해 typecheck/lint/test/deadcode/shell 결과를 0-10 weighted composite dashboard로.
argument-hint: [--baseline]
---

# /buddy:measure-code-health

project tool을 auto-detect해 typecheck/lint/test/deadcode/shell 결과를 0-10 weighted composite dashboard로.

이 command는 `measure-code-health` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/measure-code-health/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `measure-code-health`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
