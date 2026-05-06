---
description: design variant N개를 parallel 생성하고 structured feedback으로 iterate.
argument-hint: "<디자인 컨텍스트> [--count N]"
---

# /buddy:explore-design-variants

design variant N개를 parallel 생성하고 structured feedback으로 iterate.

이 command는 `explore-design-variants` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/explore-design-variants/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `explore-design-variants`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
