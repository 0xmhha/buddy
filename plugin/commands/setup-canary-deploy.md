---
description: canary deploy 단계 비율 + metric gate + auto-promote / rollback 정책 — staged rollout 으로 blast radius 제한.
argument-hint: "<feature name> v<version>"
disable-model-invocation: true
---

# /buddy:setup-canary-deploy

canary deploy 단계 비율 + metric gate + auto-promote / rollback 정책 — staged rollout 으로 blast radius 제한.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `setup-canary-deploy`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
