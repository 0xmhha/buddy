---
description: failure injection (network / pod / CPU / dependency / DB / time) + 4 원칙 (steady state / 실세계 event / production-like / abort) + blast radius 5 단계 + hypothesis-driven 실험 + game day.
argument-hint: "<experiment 이름 또는 hypothesis>"
disable-model-invocation: true
---

# /buddy:chaos-test

Chaos Engineering (Casey Rosenthal) 4 원칙. run-load-test 와 책임 분리 — load 는 부하 측정, chaos 는 실패 회복 측정. design-observability SLO 정합 + audit-error-budget burn 입력 + conduct-postmortem 학습 영속화.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `chaos-test`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
