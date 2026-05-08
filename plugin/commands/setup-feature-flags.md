---
description: feature flag system 설계 + kill switch + targeting rule + flag lifecycle (cleanup) 정책.
argument-hint: "<project name>"
---

# /buddy:setup-feature-flags

feature flag system 설계 + kill switch + targeting rule + flag lifecycle (cleanup) 정책.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `setup-feature-flags`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
