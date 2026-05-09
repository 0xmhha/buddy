---
description: 운영·개선 단계 — production 트래픽 데이터로 A/B 실험 분석, 인시던트 대응, funnel 분석, 개선 백로그 생성.
argument-hint: "<분석 대상 feature 또는 문제 설명>"
disable-model-invocation: true
---

# /buddy:iterate-product

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `iterate-product`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
