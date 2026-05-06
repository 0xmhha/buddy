---
description: git history를 evidence-based weekly retrospective로 변환 — work types, hotspots, focus score, AI collaboration.
argument-hint: "[<기간: 예 7d, 14d>]"
---

# /buddy:summarize-retro

git history를 evidence-based weekly retrospective로 변환 — work types, hotspots, focus score, AI collaboration.

이 command는 `summarize-retro` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/summarize-retro/SKILL.md`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `summarize-retro`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
