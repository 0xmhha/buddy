---
description: OWASP Top 10, secrets 노출, JWT, SQL injection 등 보안 취약점 점검.
argument-hint: "[--scope <path>] [--severity <level>]"
---

# /buddy:audit-security

CSO-mode 8-category 보안 감사 — secrets, supply chain, CI/CD, LLM/AI threats, OWASP Top 10, STRIDE.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `audit-security`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
