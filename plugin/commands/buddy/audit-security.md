---
description: CSO-mode 8-category 보안 감사 — secrets, supply chain, CI/CD, LLM/AI threats, OWASP Top 10, STRIDE.
argument-hint: [--scope <path>] [--severity <level>]
---

# /buddy:audit-security

CSO-mode 8-category 보안 감사 — secrets, supply chain, CI/CD, LLM/AI threats, OWASP Top 10, STRIDE.

이 command는 `audit-security` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/audit-security/SKILL.md`

## 실행 지시

`Skill` 도구로 `audit-security` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
