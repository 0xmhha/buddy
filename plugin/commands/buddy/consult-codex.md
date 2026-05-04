---
description: 외부 LLM CLI(codex 등)를 호출해 review / challenge / consult 3 modes로 second opinion 획득.
argument-hint: <mode> <질의 또는 diff 경로>
---

# /buddy:consult-codex

외부 LLM CLI(codex 등)를 호출해 review / challenge / consult 3 modes로 second opinion 획득.

이 command는 `consult-codex` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/consult-codex/SKILL.md`

## 실행 지시

`Skill` 도구로 `consult-codex` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
