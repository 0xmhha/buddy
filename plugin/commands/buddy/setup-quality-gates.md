---
description: husky + lint-staged + Prettier + typecheck + unit test + secret scan + commitlint를 pre-commit/pre-push에 게이트 설정.
argument-hint: [--strict]
---

# /buddy:setup-quality-gates

husky + lint-staged + Prettier + typecheck + unit test + secret scan + commitlint를 pre-commit/pre-push에 게이트 설정.

이 command는 `setup-quality-gates` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/setup-quality-gates/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `setup-quality-gates`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
