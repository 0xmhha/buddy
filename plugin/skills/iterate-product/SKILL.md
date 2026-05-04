---
name: iterate-product
description: This skill should be used when the user wants to "run an A/B test", "analyze user behavior", "analyze funnels", "improve the product based on data", "handle an incident", "conduct a postmortem", "analyze user cohorts", "triage customer feedback", or is operating a product in production and wants to iterate. Orchestrates §8 Operate & Iterate phase — the primary loop for data-driven product improvement.
---

# iterate-product — §8 Operate & Iterate Orchestrator

§8 라이프사이클 단계의 진입점. Production traffic → experiment results + improvement backlog → §2 loop.

**진입 조건**: Production traffic 존재 (§7 GA release 완료).
**산출물**: Experiment results, actor별 funnel analysis, improvement backlog (→ §2 재진입).
**다음 phase**: 분석 결과 기반 개선 → `define-features` (§2) 재진입.

> **Q3=(c) 우선순위**: §8은 현재 buddy의 최대 약점 영역. 본 orchestrator를 통해 A/B 실험, funnel 분석, 인시던트 대응, 고객 피드백 처리를 지원한다.

---

## Stage 흐름

```
iterate-product (§8 phase orchestrator)
├── [Observation loop]
│   ├── stage 1: design-ab-experiment      (가설/표본/대조군 설계)
│   ├── stage 2: analyze-ab-experiment     (통계 유의성 → 결정)
│   ├── stage 3: analyze-user-funnel       (actor별 전환/이탈 분석)
│   ├── stage 4: monitor-regressions       (regression 감지)
│   └── stage 5: summarize-retro          (evidence-based retrospective)
├── [Incident loop]
│   ├── stage 6: handle-incident           (인시던트 대응 런북)
│   └── stage 7: conduct-postmortem        (비난 없는 포스트모템)
├── [Feedback loop]
│   ├── stage 8: [triage-customer-support-ticket]  🆕
│   └── stage 9: [analyze-customer-feedback-corpus]  🆕
├── [Learning loop]
│   ├── stage 10: generate-improvement-tasks  (분석 → 백로그 자동 생성)
│   ├── stage 11: persist-learning-jsonl    (학습 데이터 저장)
│   └── stage 12: save-context / restore-context  (context 보존)
└── → define-features (§2 재진입) — 개선 backlog 처리
```

---

## 실행 절차

### Observation Loop

**Stage 1: A/B 실험 설계**

`design-ab-experiment` skill을 invoke해 실험을 설계한다:
- 가설: "X를 Y로 바꾸면 Z가 N% 개선된다"
- 표본 크기: 통계적 유의성을 위한 최소 사용자 수 (95% confidence, 80% power)
- 대조군/실험군 분배 (50/50 또는 10/90 canary)
- 측정 지표: primary metric + guardrail metrics
- 실험 기간: 최소 1-2주 (주 효과 사이클 포함)

**Stage 2: A/B 분석**

실험 완료 후 `analyze-ab-experiment` skill을 invoke한다:
- 통계 유의성 검정 (t-test / chi-square)
- p-value < 0.05 && practical significance 모두 충족해야 결정
- Ship / Revert / Continue 3가지 결정

**Stage 3: Funnel 분석**

`analyze-user-funnel` skill을 invoke해 actor별 전환/이탈을 분석한다:

§2 use case 분해 결과를 기반으로 actor별 funnel을 추적한다:
```
signup-email-password funnel:
  user-actor: form-open → form-submit → validation-pass [conversion: X%]
  backend-actor: API-call → DB-write → JWT-issued [success-rate: Y%]
  email-verifier: email-sent → link-clicked → webhook-processed [deliverability: Z%]
```

어느 actor 단계에서 drop-off가 발생하는지 식별 → 개선 우선순위 결정.

**Stage 4: Regression 모니터링**

`monitor-regressions` skill로 production 지표의 regression을 지속 감지한다.

**Stage 5: Retrospective**

`summarize-retro` skill로 git history 기반 evidence-based retrospective를 생성한다:
- work types, hotspots, focus score, AI collaboration 분석

---

### Incident Loop

**Stage 6: 인시던트 대응**

`handle-incident` skill을 invoke해 런북에 따라 대응한다:
1. 영향 범위 확인 (어느 actor가 영향받는가?)
2. 즉각 완화 (feature flag off / rollback / rate limit)
3. 근본 원인 조사 (`diagnose-bug` invoke)
4. Fix 배포 (`ship-release` §7.3 fast path)
5. 고객 커뮤니케이션

**Stage 7: 포스트모템**

`conduct-postmortem` skill을 invoke해 비난 없는 포스트모템을 진행한다:
- 타임라인 재구성
- 근본 원인 (5 Whys)
- 재발 방지 action items → `generate-improvement-tasks` 입력

---

### Feedback Loop

고객 피드백을 처리한다 (skill 미존재 시 orchestrator가 직접 수행):

**Stage 8: 고객 지원 티켓 트리아지**
- 티켓 분류: bug / feature request / question / complaint
- 심각도: P0~P3
- actor 연결: 어느 actor use case에 관련된 이슈인가?

**Stage 9: 피드백 코퍼스 분석**
- 반복 패턴 식별 (top 5 pain points)
- 감정 분석 (sentiment)
- actor별 pain point 집계

---

### Learning Loop

**Stage 10: 개선 작업 생성**

`generate-improvement-tasks` skill을 invoke해 분석 결과를 backlog로 변환한다:

```yaml
improvement_task:
  source: A/B experiment / funnel analysis / postmortem / customer feedback
  actor: {어느 actor의 어느 use case가 약한가}
  hypothesis: {개선 가설}
  priority: RICE score
  next_action: define-features (§2) → {feature spec draft}
```

**Stage 11-12: Context 보존**

`persist-learning-jsonl`로 실험 결과와 인사이트를 append-only로 저장한다.
`save-context`로 다음 세션이 이어받을 checkpoint를 생성한다.

---

## §2 재진입 트리거

다음 조건에서 `define-features` (§2)로 재진입한다:
- 개선 backlog의 priority ≥ RICE 기준 threshold
- 인시던트 재발 방지를 위한 구조적 변경 필요
- 고객 피드백에서 신규 use case 식별
- A/B 실험 결과로 새로운 feature 방향 확정

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §4 §8
- §8 신규 필요 skill 목록: 동 문서 §4 §8 Gap
