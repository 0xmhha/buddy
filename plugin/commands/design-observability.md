---
description: 3 pillar (logs / metrics / traces) 도구 / cost / retention + SLO/SLI + alert top-3 + PII redaction. observability 비용 인프라의 5~15% 가이드.
argument-hint: "<제품 이름 또는 system topology 경로>"
disable-model-invocation: true
---

# /buddy:design-observability

design-system 안 production observability 사전 설계. logs / metrics / traces 3 pillar + Google SRE SLO/SLI + alert fatigue 회피 (top-3) + secret/PII redaction.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `design-observability`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
