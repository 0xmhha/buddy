---
description: 버그 재현 → 원인 분석 → fix → regression test 까지. 버그 발견 시 단독 실행.
argument-hint: "<버그 증상 또는 재현 단계>"
---

# /buddy:diagnose-bug

증상 반응이 아닌 재현 가능한 원인 분석으로 버그 추적.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `diagnose-bug`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
