---
description: AI 편향 방지 — *엔지니어링 결정* 한정 강제 다관점 검토 (아키텍처·데이터모델·알고리즘·API·인증·스택·코드네이밍·prompt). 첫 답 commit 직전 orthogonal N개 대안 발산 + rubric 비교. 그래픽 디자인·브랜드·마케팅·사업기획은 scope 밖 (별도 스킬).
argument-hint: "<엔지니어링 결정 컨텍스트> [--count N]"
disable-model-invocation: true
---

# /buddy:verify-best-alternative

**엔지니어링 결정** (아키텍처·데이터모델·알고리즘·API·인증·스택·코드네이밍·prompt engineering)의 *AI 편향 방지*를 위한 강제 다관점 검토. 요구사항과 환경 제약 하에서 *베스트 선택지로 작업이 진행*되도록, 모델이 첫 답에 commit하려는 경향과 학습 분포에 의한 편향을 차단. *어떤 관점에서 봐도 최선*인 설계·구현·알고리즘을 선택하도록 강제.

**Scope 제한**: 엔지니어링 결정만. 그래픽 디자인·브랜드 네이밍·마케팅 카피·사업 기획은 별도 스킬(미래).

## 실행 지시

`Skill` 도구로 `router` skill 을 호출하라. 다음 컨텍스트를 전달한다:

- mode: `single`
- target PROCEDURE: `verify-best-alternative`
- 사용자 인자:
    ```
    $ARGUMENTS
    ```
