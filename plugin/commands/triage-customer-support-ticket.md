---
description: ticket 5+ 분류 (how-to / bug / feature / billing / other) + severity 4 단계 (P0~P3) + recurring pattern (1주 5+ / 30일 20+ threshold) + 4 영역 product feedback loop + KB self-service.
argument-hint: "<CS 도구 export 또는 분기>"
disable-model-invocation: true
---

# /buddy:triage-customer-support-ticket

CS ticket → product 개선 input. analyze-customer-feedback-corpus 의 raw input + generate-improvement-tasks 백로그. self-service rate 가 ticket 감소 신호.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `triage-customer-support-ticket`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
