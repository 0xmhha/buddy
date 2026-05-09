---
description: 개발 단계 — 구현 계획을 받아 TDD 루프 + 병렬 worker agent 로 코드와 테스트 완성.
argument-hint: "<feature 이름 또는 implementation plan 경로>"
disable-model-invocation: true
---

# /buddy:build-feature

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `build-feature`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
