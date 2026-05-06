---
description: §2 Feature Definition & Backlog — PRD → feature backlog (actor/use case/system boundary 포함). identify-actors, map-actor-use-cases, compose-feature-from-use-cases, define-feature-spec을 순차 실행.
argument-hint: "<PRD 경로 또는 제품 설명>"
---

# /buddy:define-features

이 command는 `define-features` skill을 즉시 invoke한다.

- Skill: `plugin/skills/define-features/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `define-features`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
