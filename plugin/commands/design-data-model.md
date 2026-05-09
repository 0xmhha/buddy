---
description: 데이터 모델 설계 — 스키마, 마이그레이션 전략, 인덱싱, 정규화 결정. production 운영 변경 비용을 사전 평가.
argument-hint: "<entity 초안 또는 sub-domain 이름>"
disable-model-invocation: true
---

# /buddy:design-data-model

데이터 모델 설계 — 스키마, 마이그레이션 전략, 인덱싱, 정규화 결정. production 운영 변경 비용을 사전 평가.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `design-data-model`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
