---
description: 코드 식별자 / PRD 어휘 / 도메인 어휘 3자 일관성 audit + drift mismatch 리스트 + remediation 제안.
argument-hint: "[--scope <path>] [--prd <path>] [--bounded-context <name>]"
disable-model-invocation: true
---

# /buddy:audit-ubiquitous-language

DDD ubiquitous language audit — 코드 식별자 (변수/함수/클래스/주석), PRD 어휘 (`docs/prd.md`, ADR, design doc), 도메인 비즈니스 어휘 (인터뷰/Slack archive) *3 source* cross-reference 검사. drift 발견 시 mismatch pair + severity + remediation 제안 출력. 실 rename 실행은 `refactor-with-rename-trace` 위임.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `audit-ubiquitous-language`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
