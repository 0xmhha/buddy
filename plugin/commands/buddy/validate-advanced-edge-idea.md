---
description: validate-idea 후속 — edge case, hidden assumption, second-order effect를 압박 인터뷰(grilling)로 박멸.
argument-hint: "<validate-idea 산출물 또는 가설>"
---

# /buddy:validate-advanced-edge-idea

validate-idea 후속 — edge case, hidden assumption, second-order effect를 압박 인터뷰(grilling)로 박멸.

이 command는 `validate-advanced-edge-idea` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/validate-advanced-edge-idea/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `validate-advanced-edge-idea`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
