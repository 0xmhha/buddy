---
description: app (native iOS/Android/hybrid) vs web (SPA/PWA) vs desktop (Electron/Tauri) 결정. 7 차원 (user device fit / job context / native API / distribution / dev cost / 유지 / 시장 정합) 평가. ADR 작성.
argument-hint: "<제품 이름 또는 PRD 경로>"
disable-model-invocation: true
---

# /buddy:decide-form-factor-app-vs-web

§3 design-system 의 stage 0 — define-tech-stack 직전 발화. 다년 락인 — 잘못 정하면 재구현 비용 1~2 자릿수 증가. user device + job context + native API 3 핵심.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `decide-form-factor-app-vs-web`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
