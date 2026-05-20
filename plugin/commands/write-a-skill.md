---
description: 신규 buddy 스킬을 PROCEDURE.md + skill-catalog 등재 + 차용 4분류 정책 적용까지 한 사이클로 작성. RED-GREEN-REFACTOR subagent pressure test 강제.
argument-hint: "<skill-name> [lifecycle-stage §1~§9|Cross-cutting] (skill-type / description / 외부 자산 등은 Step 1 forcing question으로 수집)"
disable-model-invocation: true
---

# /buddy:write-a-skill

신규 buddy 스킬을 4 layer(PROCEDURE.md + skill-catalog + routing-rules + NOTICE attribution)로 영속화한다. description-driven dispatch, progressive disclosure, *차용 4분류 정책*(외부 스킬을 verbatim/adopt-with-edits/reference-only/inspired-by 중 하나로 분류 — 자세한 정의는 PROCEDURE.md 용어 안내), subagent pressure test를 강제한다.

## 실행 지시

`Skill` 도구로 `router` skill을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `write-a-skill`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
