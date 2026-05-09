---
description: typecheck, lint, 테스트, 데드코드, shell 자동 감지 → 0-10 가중 점수 대시보드.
argument-hint: [--baseline]
disable-model-invocation: true
---

# /buddy:measure-code-health

project tool을 auto-detect해 typecheck/lint/test/deadcode/shell 결과를 0-10 weighted composite dashboard로.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `measure-code-health`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
