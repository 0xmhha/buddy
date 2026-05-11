---
description: low-fi (wireframe) → state diagram (5+ state — empty/loading/content/error/success) → high-fi (system 정합) → user testing 5명 (Nielsen 80% 발견율) → dev handoff (inspect + a11y annotation + asset).
argument-hint: "<feature spec 경로 또는 PRD 섹션>"
disable-model-invocation: true
---

# /buddy:prototype-from-spec

define-feature-spec → 시각 prototype. low-fi 우선 (high-fi 직진 회피) + state 5+ 모두 명시 + design system 정합 + dev handoff 완비. user testing 옵션.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `prototype-from-spec`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
