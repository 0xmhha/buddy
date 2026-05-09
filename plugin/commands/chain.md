---
description: 여러 buddy skill 을 순서대로 chain 실행. 이전 step 의 산출물을 다음 step 의 입력으로 전달.
argument-hint: "<skill1>,<skill2>,...[ -- <shared args>]"
disable-model-invocation: true
---

# /buddy:chain

여러 buddy skill 을 콤마로 분리해 순차 실행한다. 직전 step 의 산출물(요약·결정·생성된 파일 경로 등)이 다음 step 의 입력 컨텍스트에 자동 포함된다. autoplan 처럼 chain 형태가 빈번한 워크플로우를 사용자가 ad-hoc 으로 조립할 때 사용.

예: `/buddy:chain validate-idea,assess-business-viability -- "내 SaaS 아이디어"`

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `chain`
- targets: `$ARGUMENTS` 에서 ` -- ` 앞 부분의 콤마 분리 skill 이름 리스트
- 사용자 인자: `$ARGUMENTS` 에서 ` -- ` 뒤 부분 (구분자 없으면 빈 문자열)

targets 가 비어 있거나 1 개뿐이면 사용자에게 chain 의 의도를 확인하고 (1 개면 `/buddy:run` 또는 전용 커맨드를 권장) 진행 여부를 묻는다.
