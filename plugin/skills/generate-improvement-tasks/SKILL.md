---
name: generate-improvement-tasks
description: This skill should be used when the user wants to "generate improvement tasks from analysis", "create backlog from experiment results", "convert insights to tasks", "turn postmortem action items into features", or has analysis results (A/B experiment, funnel, incident, customer feedback) and needs to convert them into actionable work items.
---

# generate-improvement-tasks

§8 분석 결과(A/B 실험, funnel 분석, postmortem, 고객 피드백)를 actionable improvement task로 변환한다. 각 task는 §2 `define-features` 재진입의 입력이 된다.

**§8 iterate-product stage skill.** 단독 호출도 가능 (dual-mode).

---

## 변환 절차

### 1. 입력 소스 식별

허용 입력:
- `analyze-ab-experiment` 결과 (Ship/Continue/No Result인 경우)
- `analyze-user-funnel` 결과 (drop-off 분석)
- `conduct-postmortem` action items
- 고객 지원 티켓 패턴
- `summarize-retro` 인사이트

### 2. Improvement Task 생성

각 인사이트를 improvement task로 변환한다:

```yaml
improvement_task:
  id: {source}-{N}
  source: {ab_experiment / funnel / postmortem / customer_feedback / retro}
  
  # §2 define-features 연결
  actor: {어느 actor의 어느 use case가 약한가}
  use_case: {개선이 필요한 use case}
  system_boundary: {frontend-spa / backend-service / external-saas}
  
  # 문제 정의
  observation: {데이터로 관찰된 사실}
  hypothesis: {개선 가설}
  
  # 우선순위
  rice_score:
    reach: {영향 사용자 수 / 일}
    impact: {1-10}
    confidence: {0-100%}
    effort: {person-days}
    score: {reach × impact × confidence / effort}
  
  # 다음 행동
  next_action:
    type: {a_b_test / feature_change / bug_fix / design_change}
    target_phase: {§2 define-features / §3 design-system / §5 build-feature}
    draft_spec:
      problem: {개선하려는 문제}
      proposed_solution: {제안 솔루션}
      success_metric: {개선 검증 지표}
```

### 3. 예시 변환

**입력**: funnel 분석 결과
```
user-actor: form-start → submit: 40% drop-off
관찰: 비밀번호 요구사항 에러가 가장 많은 탈락 지점
```

**출력**: improvement task
```yaml
improvement_task:
  id: funnel-001
  source: funnel
  actor: user-actor
  use_case: "이메일/비밀번호로 회원가입한다"
  system_boundary: frontend-spa
  observation: "form-start → submit 단계에서 40% drop-off. 비밀번호 에러 메시지 노출 후 38%가 이탈."
  hypothesis: "inline 실시간 비밀번호 강도 표시를 추가하면 form-submit conversion이 15% 향상된다."
  rice_score:
    reach: 400  # 일 400명 이탈
    impact: 7
    confidence: 70
    effort: 3  # person-days
    score: 653
  next_action:
    type: a_b_test
    target_phase: §5 build-feature
    draft_spec:
      problem: "비밀번호 요구사항이 제출 후에만 표시되어 UX 마찰 발생"
      proposed_solution: "입력 중 실시간 비밀번호 강도 indicator + 요구사항 체크리스트 표시"
      success_metric: "form-submit conversion rate +15% (현재 60% → 69%)"
```

### 4. 우선순위 정렬

생성된 task를 RICE score 내림차순으로 정렬한다.

```markdown
| 순위 | Task ID | Actor | 이슈 | RICE | 다음 행동 |
|------|---------|-------|------|------|---------|
| 1 | funnel-001 | user-actor | form drop-off | 653 | A/B test |
| 2 | funnel-002 | email-verifier | link click | 420 | design change |
| 3 | postmortem-001 | backend | retry logic | 180 | bug fix |
```

### 5. §2 재진입 판단

각 task의 규모에 따라 진입 단계를 결정한다:

| Task 유형 | 진입 단계 |
|---------|---------|
| 신규 use case 발견 | §2 `define-features` (처음부터) |
| 기존 use case 변경 | §2 `define-features` (stage 5 `define-feature-spec`부터) |
| UI/UX 변경만 | §5 `build-feature` |
| Bug fix | §5 `build-feature` → `diagnose-bug` |
| A/B test → Ship | §7 `ship-release` fast path |

---

## 출력 형식

```markdown
## Improvement Backlog — {날짜}

### 소스: {sources}

### Top 5 Improvement Tasks

{RICE score 내림차순 task 목록}

### §2 재진입 필요 task
{define-features 재진입이 필요한 task 목록}

### 즉시 실행 가능 task (§5 이후)
{build-feature 또는 bug fix로 바로 처리 가능한 task}
```

---

## 다음 단계

- RICE score 상위 → `define-features` (§2) 또는 `design-ab-experiment`
- Bug fix → `diagnose-bug` → `build-feature` (§5)
- `triage-work-items` — backlog 상태 머신 진입
