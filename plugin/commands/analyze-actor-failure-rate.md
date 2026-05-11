---
description: 4 actor (user/system/3rd-party/external) failure rate matrix + trust score (reliability/predictability/MTTR/blast) + 6 recovery 패턴 (retry/circuit/fallback/bulkhead/timeout/idempotent) + cascade.
argument-hint: "<system 또는 분기>"
disable-model-invocation: true
---

# /buddy:analyze-actor-failure-rate

agent-evaluation 의 input vs output trust scoring 변형 — actor 별 trust score. Release It! (Nygard) 6 recovery 패턴 적용 검증. audit-error-budget actor 차원 입력.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `analyze-actor-failure-rate`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
