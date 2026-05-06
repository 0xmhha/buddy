---
name: router
description: "Use when a buddy command requests dispatch to a target PROCEDURE. Reads `${CLAUDE_PLUGIN_ROOT}/skills/<target>/PROCEDURE.md` and executes its instructions."
type: skill
---

# Buddy Router

Buddy plugin 내부 라우터. 모든 `/buddy:*` slash command가 이 skill을 호출하며, command가 전달한 `target PROCEDURE` 이름으로 해당 절차 파일을 읽어 그대로 수행한다.

## How to dispatch (single)

입력으로 다음을 받는다:

- `target PROCEDURE: <name>` — 실행할 절차의 이름 (Buddy skill 디렉토리 이름과 동일)
- 사용자 인자 — command가 그대로 전달한 `$ARGUMENTS` 본문

수행 절차:

1. `Read` 도구로 `${CLAUDE_PLUGIN_ROOT}/skills/<name>/PROCEDURE.md` 파일을 로드한다. (`${CLAUDE_PLUGIN_ROOT}` 는 buddy plugin 설치 경로 — hardcode 금지)
2. 로드한 PROCEDURE 본문을 그 자체로 실행 지시문으로 취급한다 — 사용자 인자를 PROCEDURE의 입력으로 사용한다.
3. PROCEDURE가 요구하는 모든 단계를 누락 없이 수행한다.
4. PROCEDURE 파일이 존재하지 않으면 즉시 에러 메시지를 반환하고 임의로 추론하지 않는다.

## Skill index

<!-- TODO(Task 2.1): Buddy 전체 PROCEDURE 매핑을 여기서 채운다. 현재는 단일-target dispatch만 지원. -->
