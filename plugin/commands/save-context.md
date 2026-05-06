---
description: 현재 git 상태, 결정 사항, 남은 작업을 체크포인트로 저장. 브랜치가 달라져도 이어받을 수 있게.
argument-hint: "[<체크포인트 라벨>]"
---

# /buddy:save-context

현재 git state, decisions, remaining work를 checkpoint로 저장 — branch가 달라도 future session이 이어받게.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `save-context`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
