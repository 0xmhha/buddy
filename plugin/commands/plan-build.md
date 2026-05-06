---
description: 구현 계획 단계 — 기술 설계를 받아 actor 별 task 분해 + 의존성 그래프 + 병렬 실행 계획 작성.
argument-hint: "<feature backlog 또는 technical design 경로>"
---

# /buddy:plan-build

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `plan-build`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
