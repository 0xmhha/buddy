---
description: 엔지니어링 작업 stuck 상태(언어·플랫폼 독립)에서 문제 분해 + 행동 후보 도출. 자동 trigger — AI 동일 문제 3회 시도 후 미해결 시 token escalation 차단 + 문제 세분화. 사용자 명시 호출도 가능.
argument-hint: "<stuck 상태 묘사 또는 시도 history> [도메인 컨텍스트]"
disable-model-invocation: true
---

# /buddy:decompose-blocker

엔지니어링 작업 stuck 상태에서 *문제 분해 + 행동 후보 도출* 사이클. **자동 trigger**: AI가 동일 문제·동일 코드 영역에서 *3회 시도 후에도 미해결*이면 자동 호출 (token escalation 차단). 사용자 명시 호출도 가능 ("어디부터 봐야 할지 모르겠어" 등).

절차: fact·추측·모름 3분류 → 분해 축 결정 (4D/5-Whys/fishbone/binary-search) → 모름의 지도 → 사용자 질문(≤3) → 가설 압축 → 비용-정보 매트릭스 → 다음 행동 선택. *직접 fix는 수행하지 않음* — 다음 스킬(`diagnose-bug` / `iterate-fix-verify` / `verify-best-alternative` 등)로 dispatch 준비까지. *언어·플랫폼 독립* — Python/Go/Rust/TS/Java 모두 동일.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `decompose-blocker`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
