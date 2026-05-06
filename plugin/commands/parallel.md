---
description: 여러 buddy skill 을 동시에 병렬 실행. 각 target 마다 fresh subagent 디스패치, 결과 집계.
argument-hint: "<skill1>,<skill2>,...[ -- <shared args>]"
---

# /buddy:parallel

여러 buddy skill 을 콤마로 분리해 병렬 실행한다. 각 target 마다 fresh subagent (Agent 도구, `subagent_type: general-purpose`) 가 디스패치되어 자신의 PROCEDURE.md 를 Read 하고 실행한다. 결과는 target 별로 그룹핑되어 집계된다. review-engineering / review-design / review-scope 같이 독립적으로 평가 가능한 multi-perspective 리뷰에 적합.

예: `/buddy:parallel review-engineering,review-design,review-scope -- "이 PR"`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `parallel`
- targets: `$ARGUMENTS` 에서 ` -- ` 앞 부분의 콤마 분리 skill 이름 리스트
- 사용자 인자: `$ARGUMENTS` 에서 ` -- ` 뒤 부분 (구분자 없으면 빈 문자열)

targets 가 비어 있거나 1 개뿐이면 사용자에게 parallel 의도를 확인하고 (1 개면 `/buddy:run` 또는 전용 커맨드 권장) 진행 여부를 묻는다.
