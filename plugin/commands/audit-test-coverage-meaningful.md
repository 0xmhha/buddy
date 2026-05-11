---
description: line coverage (input metric, weight 0.2) + mutation score (output metric, weight 0.4, target 80%+) + behavior assertion 비율 (weight 0.2) + edge case (weight 0.2) → trust score 종합. Stryker / mutmut / go-mutesting.
argument-hint: "<test suite 경로 또는 module>"
disable-model-invocation: true
---

# /buddy:audit-test-coverage-meaningful

line coverage = false confidence. agent-evaluation (OMAS v2) 의 input vs output trust scoring 패턴 차용. mutation testing + behavior assertion + edge case 종합 trust score. survived mutant → generate-tests-from-spec 의 새 skeleton 입력.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `audit-test-coverage-meaningful`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
