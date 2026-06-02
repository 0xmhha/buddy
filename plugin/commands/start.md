---
description: |
  Use when the user is unsure which buddy command to run, or starts a new task without specifying intent. Trigger phrases: "어디서부터 시작", "어떻게 시작해야 해", "buddy로 뭐 할까", "처음 써봐", "start over", "where to start", "begin a new project", "이 버그 좀", "이 프로젝트", "아이디어가 있어". Collects 3 inputs (project path / change type / commercial flag) via plain-language questions, then dispatches to concretize-idea (greenfield — no path) or assess-product-change (existing product — path provided). If no $ARGUMENTS, starts with the path question.
argument-hint: "[의도 자유 발화 — 예: '아이디어가 있어서 구체화하고 싶어' / '이 버그 좀 봐줘' / 비워두면 질문부터 시작]"
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
