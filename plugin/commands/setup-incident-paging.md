---
description: on-call rotation + escalation policy + alert wiring + runbook 인덱스 — production incident first response 구조.
argument-hint: "<project name>"
---

# /buddy:setup-incident-paging

on-call rotation + escalation policy + alert wiring + runbook 인덱스 — production incident first response 구조.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `setup-incident-paging`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
