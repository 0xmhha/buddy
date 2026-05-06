---
description: §1 Idea & Business Validation — idea/concept → PRD draft + business viability report. validate-idea, assess-business-viability, define-product-spec, autoplan(review)을 순차 실행.
argument-hint: "<아이디어 설명 또는 컨셉>"
---

# /buddy:concretize-idea

이 command는 `concretize-idea` skill을 즉시 invoke한다.

- Skill: `plugin/skills/concretize-idea/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `concretize-idea`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
