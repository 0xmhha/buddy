---
description: sustained + soak + spike + stress 4 시나리오 실행 + breaking point 식별 + capacity headroom 측정. production launch 직전 SLA 근거 확보.
argument-hint: "<project name> v<version>"
---

# /buddy:run-load-test

sustained + soak + spike + stress 4 시나리오 실행 + breaking point 식별 + capacity headroom 측정.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `run-load-test`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
