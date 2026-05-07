---
description: Architecture Decision Record 작성 — context/decision/consequences/alternatives 표준 양식으로 의사결정 영속화 + supersede 체인 + Index 갱신.
argument-hint: "<title 또는 결정 요약>"
---

# /buddy:write-adr

Architecture Decision Record 작성 — context/decision/consequences/alternatives 표준 양식으로 의사결정 영속화 + supersede 체인 + Index 갱신.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `write-adr`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
