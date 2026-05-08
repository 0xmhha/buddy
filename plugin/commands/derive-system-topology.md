---
description: actor 그래프 + infra 매핑 → 시스템 토폴로지 자동 도출 — service map + data flow + trust boundary diagram (mermaid + JSON).
argument-hint: "<project name>"
---

# /buddy:derive-system-topology

actor 그래프 + infra 매핑 → 시스템 토폴로지 자동 도출 — service map + data flow + trust boundary diagram (mermaid + JSON).

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `derive-system-topology`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
