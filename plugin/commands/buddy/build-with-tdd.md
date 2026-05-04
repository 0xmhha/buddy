---
description: Red-Green-Refactor TDD 루프로 신규 기능 구현 — test 먼저 → 실패 확인 → 최소 구현 → 통과 → 리팩터.
argument-hint: <기능/스펙 설명>
---

# /buddy:build-with-tdd

Red-Green-Refactor TDD 루프로 신규 기능 구현 — test 먼저 → 실패 확인 → 최소 구현 → 통과 → 리팩터.

이 command는 `build-with-tdd` skill을 즉시 invoke한다. 본 skill의 전체 절차·트리거·출력 포맷은 다음을 따른다:

- Skill: `plugin/skills/build-with-tdd/SKILL.md`

## 실행 지시

`Skill` 도구로 `build-with-tdd` skill을 호출하라. 사용자가 전달한 인자는 그대로 skill에 입력으로 넘긴다.

```
$ARGUMENTS
```

추가 컨텍스트가 필요하면 skill 본문이 요구하는 입력 항목을 사용자에게 물어본 뒤 진행한다.
