---
description: buddy 진입 라우터 — 자연어로 의도 표현하면 적절한 orchestrator로 자동 라우팅. 입력 없으면 메뉴 제시.
argument-hint: "[의도 자유 발화 — 예: '아이디어가 있어서 구체화하고 싶어' / '이 버그 좀 봐줘' / 또는 비워두면 메뉴]"
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
