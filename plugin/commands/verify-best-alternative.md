---
description: AI 편향 방지 — 엔지니어링 결정의 첫 답 commit 직전 강제 다관점 검토. orthogonal N개 대안 발산 + rubric 비교로 "어떤 관점에서 봐도 최선" 검증.
argument-hint: "<결정 컨텍스트> [--count N]"
disable-model-invocation: true
---

# /buddy:verify-best-alternative

엔지니어링 결정의 *AI 편향 방지*를 위한 강제 다관점 검토. 모델이 첫 답에 commit하려는 경향과 학습 분포에 의한 편향을 차단하고, *어떤 관점에서 봐도 최선*인 설계·구현·알고리즘을 선택하도록 강제.

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `verify-best-alternative`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
