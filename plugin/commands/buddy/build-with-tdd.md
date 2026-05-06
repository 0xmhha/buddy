---
description: Red-Green-Refactor TDD 루프로 신규 기능 구현 — test 먼저 → 실패 확인 → 최소 구현 → 통과 → 리팩터.
argument-hint: "<기능/스펙 설명>"
---

# /buddy:build-with-tdd

Red-Green-Refactor TDD 루프로 신규 기능 구현 — test 먼저 → 실패 확인 → 최소 구현 → 통과 → 리팩터.

이 command는 `build-with-tdd` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/build-with-tdd/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `build-with-tdd`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
