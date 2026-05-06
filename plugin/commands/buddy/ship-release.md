---
description: §7 Release & Beta — quality gate pass → tagged release + canary/UAT + GA. changelog, doc sync, PR, UAT, release tagging을 포함.
argument-hint: "<릴리즈 버전 또는 릴리즈 설명>"
---

# /buddy:ship-release

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `ship-release`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
