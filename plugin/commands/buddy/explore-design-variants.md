---
description: 여러 설계 안을 병렬로 생성하고 구조화된 피드백으로 반복 개선.
argument-hint: "<디자인 컨텍스트> [--count N]"
---

# /buddy:explore-design-variants

design variant N개를 parallel 생성하고 structured feedback으로 iterate.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `explore-design-variants`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
