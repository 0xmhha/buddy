---
description: deprecation timeline (영향 별 1~12 month) + sunset notice 5 layer (email/banner/reminders/final/sunset day) + 4 차원 telemetry + 5 영역 migration path + post-sunset cleanup.
argument-hint: "<feature 이름>"
disable-model-invocation: true
---

# /buddy:deprecate-feature

§9 lifecycle 기능 sunset. trust 손상 위험 영역 — 명시 + 충분 기간 + alternative 제공이 핵심. refactor-with-rename-trace 의 deprecation alias chain 정합.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `deprecate-feature`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
