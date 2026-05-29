---
description: buddy 진입 라우터 — 자연어로 의도 표현하면 자연스러운 질문으로 프로젝트 경로/작업 유형/상업성을 수집한 뒤 적절한 작업 흐름으로 안내. 입력 없으면 처음부터 질문 시작.
argument-hint: "[의도 자유 발화 — 예: '아이디어가 있어서 구체화하고 싶어' / '이 버그 좀 봐줘' / 또는 비워두고 질문 받기]"
disable-model-invocation: true
---

# /buddy:start

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `start`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
