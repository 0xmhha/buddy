---
description: acquisition cohort retention curve (D1/D7/D30/D90) + LTV/CAC 3:1+ + 3 churn 분류 (voluntary/involuntary/implicit) + 통계적 유의 검증.
argument-hint: "<분기 또는 cohort 차원>"
disable-model-invocation: true
---

# /buddy:analyze-user-cohort

cohort 2~3 차원 cross + retention curve + LTV / CAC + churn 분류. analyze-feature-adoption 와 cross-reference. analyze-actor-failure-rate 의 actor 차원 분리.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `analyze-user-cohort`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
