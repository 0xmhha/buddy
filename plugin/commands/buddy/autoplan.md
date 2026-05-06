---
description: 기존 plan / PRD / 설계 산출물을 받아 scope, design, engineering, developer experience 4개 관점으로 자동 리뷰.
argument-hint: "<plan 경로 또는 설명>"
---

# /buddy:autoplan

multi-stage review pipeline 자동 실행 — CEO/design/eng/DX 리뷰를 순차로 돌려 plan을 finalize.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `autoplan`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
