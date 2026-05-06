---
description: multi-stage review pipeline 자동 실행 — CEO/design/eng/DX 리뷰를 순차로 돌려 plan을 finalize.
argument-hint: "<plan 경로 또는 설명>"
---

# /buddy:autoplan

multi-stage review pipeline 자동 실행 — CEO/design/eng/DX 리뷰를 순차로 돌려 plan을 finalize.

이 command는 `autoplan` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/autoplan/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `autoplan`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
