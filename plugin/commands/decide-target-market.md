---
description: target market 결정 — 글로벌 / 단일 지역 / 다지역. region cluster trigger 결정.
argument-hint: "<제품 이름 또는 PRD 경로>"
disable-model-invocation: true
---

# /buddy:decide-target-market

assess-business-viability 통과 후 *어느 시장에 먼저 진출하나?* 를 결정. 글로벌 default / 단일 지역 / 다지역 sequencing 중 하나 채택. region cluster (Korea / USA / EU 등) 활성화 trigger.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `decide-target-market`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
