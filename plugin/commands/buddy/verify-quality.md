---
description: §6 Quality — code complete → QA report + security/legal sign-off. actor별 통합 테스트, E2E, audit-security, measure-code-health, compliance review를 포함.
argument-hint: "<feature 이름 또는 테스트 대상>"
---

# /buddy:verify-quality

이 command는 `verify-quality` skill을 즉시 invoke한다.

- Skill: `plugin/skills/verify-quality/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `verify-quality`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
