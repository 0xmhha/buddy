---
description: save-context 가 저장한 가장 최근 작업 체크포인트를 복원. 브랜치 무관.
argument-hint: "[<체크포인트 라벨 또는 ID>]"
---

# /buddy:restore-context

context-save가 저장한 most recent work checkpoint를 cross-branch로 load.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `restore-context`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
