---
description: gesture / motion / feedback / state transition 4 영역. 200~300ms ease-out 기본 + 100ms 안 immediate response + mobile gesture vocabulary 표준 + prefers-reduced-motion 대응.
argument-hint: "<제품 이름 또는 component>"
disable-model-invocation: true
---

# /buddy:design-interaction-pattern

UI 동적 layer 설계 — apply-design-system 의 default interaction 보강. 6 state 일관 transition + immediate feedback 100ms + mobile gesture 표준. design-accessibility-baseline 의 prefers-reduced-motion 정합.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `design-interaction-pattern`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
