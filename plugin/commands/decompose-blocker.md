---
description: 코드 작업 중 stuck 상태(어디서 봐야 할지 모름·다음 행동 안 보임)에서 문제를 기계적으로 분해하고 비용-정보 매트릭스 기반 행동 후보 도출.
argument-hint: "<stuck 상태 묘사> [도메인 컨텍스트]"
disable-model-invocation: true
---

# /buddy:decompose-blocker

코드 작업 중 stuck 상태에서 *문제 분해 + 행동 후보 도출* 사이클. fact·추측·모름 3분류 → 분해 축 결정 → 모름의 지도 → 사용자 질문(≤3) → 가설 압축 → 비용-정보 매트릭스 → 다음 행동 선택. *직접 fix는 수행하지 않음* — 다음 스킬(diagnose-bug / iterate-fix-verify / verify-best-alternative 등)로 dispatch 준비까지.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `decompose-blocker`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
