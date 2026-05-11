---
description: TAM / SAM / SOM 산출 — top-down + bottom-up cross-check + sensitivity ±50%.
argument-hint: "<제품 이름 또는 PRD 경로>"
disable-model-invocation: true
---

# /buddy:analyze-market-size

PRD / 사업성 가설 → TAM / SAM / SOM. top-down (Statista / Gartner) 와 bottom-up (reachable customer × ARPC) 동시 적용 + cross-check + ±50% sensitivity. assess-business-viability 의 시장 차원 입력.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `analyze-market-size`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
