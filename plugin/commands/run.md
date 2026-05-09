---
description: 임의의 단일 buddy skill 을 직접 invoke. 첫 번째 인자가 target skill 이름, 나머지는 그 skill 의 인자.
argument-hint: "<target-skill-name> [skill arguments...]"
disable-model-invocation: true
---

# /buddy:run

`/buddy:run <target> <args...>` 형태로 buddy skill 카탈로그의 임의 skill 을 직접 호출한다. 전용 슬래시 커맨드(`/buddy:concretize-idea` 등)가 없는 stage / pattern skill 을 명시적으로 실행할 때 사용한다.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `$ARGUMENTS` 의 첫 번째 토큰 (whitespace 분리)
- 사용자 인자: `$ARGUMENTS` 의 나머지 (첫 토큰 이후)

target PROCEDURE 가 비어 있거나, 해당 이름의 PROCEDURE.md 가 존재하지 않으면 즉시 사용자에게 사용법을 안내하고 중단한다 — 임의 추론 금지.
