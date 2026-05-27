---
description: 기존 프로덕트 변경 평가 — 버그, 기능, 개선 무관하게 영향 평가 + scope 분류 + 다음 단계 안내.
argument-hint: "<변경 요청 설명 (버그 리포트, 기능 요청, 개선 사항 등)>"
disable-model-invocation: true
---

# /buddy:assess-product-change

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `assess-product-change`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
