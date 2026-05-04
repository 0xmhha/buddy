---
name: triage-work-items
description: "이슈/feature/task 같은 work item의 우선순위 결정과 lifecycle state machine 운영. 이슈는 needs-triage→ready-for-agent/human/wontfix→in-progress→resolved 순환. feature는 draft→candidate→ready→in-progress→implemented→verified→reusable→deprecated. 트리거: '어디까지 됐지?' / '이거 누가 봐?' / 'ready인지 봐줘' / 'triage 해줘' / 'wontfix 결정' / 'feature 상태 업데이트' / 'needs-info 처리'. 입력: bug report, monitor-regressions 출력, feature 후보, code health 위반. 출력: triage board + action plan + state 변경 ledger. 흐름: monitor-regressions/measure-code-health/split-work-into-features → triage-work-items → diagnose-bug/build-with-tdd."
type: skill
---

# Triage Work Items — Lifecycle State Machine 운영

## 1. 목적

운영 중 들어온 버그 리포트, `measure-code-health` 위반, `monitor-regressions` 알림, PRD 파생 작업을 **체계적으로 분류**하고 lifecycle state machine으로 관리한다.

이 스킬은 **2개의 state machine**을 동시 운영:
- **Issue lifecycle**: 들어온 work item이 어떤 상태인지
- **Feature lifecycle**: feature가 spec → 구현 → 검증 → 재사용까지 어느 단계인지

핵심 가치: 모든 work item이 명시적 state를 가져 "잊혀진 이슈" 0건. agent / human 분기 명확.

## 2. 사용 시점 (When to invoke)

- 신규 bug report 도착
- `monitor-regressions`에서 console err / page fail 알림
- `measure-code-health` 점수 하락으로 위반 항목 등록
- `split-work-into-features` 출력 feature spec set 상태 초기화
- 매일 / 매주 standup 전 status 업데이트
- 외부 stakeholder가 "이거 어떻게 됐어?" 질문 시
- agent가 작업 중 "이건 human이 해야겠다" 판단

## 3. 입력 (Inputs)

### 필수
- work item 인스턴스 (bug / feature / task)
- 현재 known state (있으면)
- impact / severity 정보

### 선택
- reporter 정보
- 재현 가능 여부 (`diagnose-bug` 호출 가능)
- 관련 feature_id (`split-work-into-features` 매핑)
- code health 위반 카테고리

### 입력 부족 시 forcing question
- "재현 단계가 명확해? 명확하지 않으면 needs-info."
- "이거 agent가 해결할 수 있는 작업이야 human 판단 필요해?"
- "wontfix면 사유 기록했어? 추후 같은 이슈 재유입 방지."
- "feature 상태 업데이트면 evidence 있어? code link / test result?"

## 4. 핵심 원칙 (Principles)

1. **모든 work item은 명시적 state** — "어딘가 backlog"는 missing state. needs-triage부터 시작.
2. **Agent vs human 명시 분기** — ready-for-agent / ready-for-human. 모호하면 human.
3. **wontfix는 사유 + 대안 기록** — 사유 없으면 같은 이슈 재유입. 대안 있으면 알림.
4. **needs-info 무한 정체 차단** — 7일 응답 없으면 자동 close (정책에 따라).
5. **Feature lifecycle은 evidence 기반** — implemented는 code link. verified는 test result. reusable은 patch artifact.
6. **State 변경은 ledger로** — 누가 언제 왜. audit trail.
7. **Escalation은 agent의 책임** — agent가 "안 되겠다" 판단 시 즉시 ready-for-human 전환.
8. **Priority는 impact + reach + urgency** — Critical은 핵심 마비. nice-to-have는 Low.

## 5. 단계 (Phases)

### Phase 1. Issue Intake
신규 work item:
- needs-triage state 초기화
- reporter / source / timestamp 기록
- preliminary severity 추정

### Phase 2. Triage Decision
각 needs-triage 항목 분류:
- 정보 부족 → `needs-info` + 무엇이 필요한지 명시
- agent 단독 해결 가능 → `ready-for-agent`
- human 판단 필요 → `ready-for-human`
- 수정하지 않음 → `wontfix` + 사유

### Phase 3. Priority Assignment
각 항목에 priority:
- **Critical**: 제품 핵심 기능 마비, 데이터 손실, 보안 사고, 결제 실패
- **High**: 주요 기능 장애, 심각한 성능 저하 (>50% degradation)
- **Medium**: 보조 기능 장애, 우회로 존재
- **Low**: UI 오타, 사소한 레이아웃, nice-to-have

priority + state 조합으로 action plan 산출.

### Phase 4. Feature Lifecycle Update (feature 항목인 경우)
feature는 8 state machine:

```
draft         → 아이디어 수준
  ↓
candidate     → PRD에서 추출, feature 가능성 있음
  ↓
ready         → 명세 / acceptance criteria / dependency 충분
  ↓
in-progress   → 구현 중 (build-with-tdd / iterate-fix-verify)
  ↓
implemented   → code link 연결 (commit / PR)
  ↓
verified      → test evidence 연결 (test pass / coverage)
  ↓
reusable      → patch / cherry-pick으로 재사용 검증됨
  ↓
deprecated    → 더 이상 권장 안 함 (대체 feature 명시)
```

각 state 전환은 evidence 필수:
- `ready`: spec field 모두 채움 + dependency 식별
- `in-progress`: owner 배정 + branch 생성
- `implemented`: code_links (commit_sha / PR url)
- `verified`: test_evidence (test result / coverage %)
- `reusable`: patch_artifact (hash / signature / apply guide)

### Phase 5. Action Plan
오늘 / 이번 주 처리할 항목:
- ready-for-agent + Critical/High → agent 즉시 처리
- ready-for-human + Critical → escalate
- needs-info 7일+ → reporter 재요청 또는 close
- in-progress 14일+ → blocker 확인

### Phase 6. Audit Trail
state 변경 ledger에 기록:
- timestamp
- 변경자 (agent / human ID)
- 이전 state → 새 state
- 이유

## 6. 출력 템플릿 (Output Format)

```yaml
triage_board:
  generated_at: "<timestamp>"
  total_items: 47
  by_state:
    needs-triage: 8
    needs-info: 3
    ready-for-agent: 12
    ready-for-human: 6
    in-progress: 9
    wontfix: 4
    resolved: 5
  by_priority:
    critical: 2
    high: 9
    medium: 21
    low: 15

action_plan:
  immediate:
    - id: BUG-042
      title: "결제 API 500 error"
      priority: critical
      state: ready-for-human
      reason: "billing 코드 + production data — human 판단"
      assigned_to: "@alice"
    - id: FEAT-007
      title: "feature.query API 구현"
      priority: high
      state: ready-for-agent
      reason: "spec ready + test plan 존재"
      next_skill: build-with-tdd

  this_week:
    - id: TASK-091
      ...

stale:
  - id: BUG-013
    state: needs-info
    days_stale: 9
    action: "reporter 재요청 또는 close"

state_transitions:
  - timestamp: "<...>"
    item_id: BUG-042
    from: needs-triage
    to: ready-for-human
    actor: agent
    reason: "재현 가능 + billing 영역 — human 판단"

feature_lifecycle:
  - feature_id: auth.email-password-login
    feature_num: FEAT-001
    state: implemented
    last_transition: "<...>"
    evidence:
      code_links: ["github.com/org/app/pull/42"]
      test_evidence: null  # → verified로 가려면 필요
      patch_artifact: null
    next_required: "test result + coverage report"

deprecated:
  - feature_id: auth.legacy-cookie-session
    deprecated_at: "<date>"
    replacement: auth.jwt-session
    sunset_date: "<date>"
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `monitor-regressions` (이슈 유입), `measure-code-health` (위반 등록), `split-work-into-features` (feature 후보 등록)
- 페어: `diagnose-bug` — bug 항목 처리 시
- 다음 단계: `build-with-tdd` (ready-for-agent → in-progress), `iterate-fix-verify`
- 운영 동반: `summarize-retro` (주간 처리량 요약)

## 8. Anti-patterns

1. **needs-triage 무한 정체** — 매일 triage. 7일+ 항목 자동 escalate.
2. **wontfix 사유 미기록** — 같은 이슈 재유입. 사유 + 대안 강제.
3. **agent / human 분기 모호** — "그냥 backlog". ready-for-agent / ready-for-human 명시 분기.
4. **State 우회** — implemented 직접 verified로 점프 (test 없이). state machine 강제.
5. **Priority 인플레이션** — 모든 게 critical. Critical은 4 카테고리 (마비/데이터/보안/결제)로 제한.
6. **Stale 항목 방치** — needs-info 7일+, in-progress 14일+ 알림. blocker 확인.
7. **Feature lifecycle evidence 생략** — implemented만 마킹하고 code_links 없음. evidence 강제.
8. **Audit trail 없이 state 변경** — 누가 언제 왜? ledger 강제.
