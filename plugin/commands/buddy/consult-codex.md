---
description: 외부 LLM CLI(codex 등)를 호출해 review / challenge / consult 3 modes로 second opinion 획득.
argument-hint: "<mode> <질의 또는 diff 경로>"
---

# /buddy:consult-codex

외부 LLM CLI(codex 등)를 호출해 review / challenge / consult 3 modes로 second opinion 획득.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `consult-codex`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
