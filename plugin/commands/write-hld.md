---
description: PRD 완료 후 호출하는 High Level Design 작성 — product decomposition + tech stack + communication + use case mapping. Phase 1 Mode A의 PRD 직후 단계.
argument-hint: "<PRD 경로 또는 PRD 내용 요약>"
disable-model-invocation: true
---

# /buddy:write-hld

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `write-hld`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
