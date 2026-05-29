---
description: 지정한 skill의 PROCEDURE.md를 21-항목 체크리스트로 평가하고 가중치 점수 + 항목별 actionable 개선 제안 출력. authoring-guide §4 기준.
argument-hint: "<skill-name 또는 PROCEDURE.md 경로>"
disable-model-invocation: true
---

# /buddy:evaluate-skill

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `evaluate-skill`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
