# §4 Iteration Planning (구현 계획) — Stage Skills

- **Orchestrator (entry)**: `plan-build`
- **DoR → DoD**: software design (SDD) + feature specs → Iteration Plan = task DAG + actor-track plan
- 정체성 원본: `engineering-phases.md` §2 Phase 4 | 명사 원본: `se-lifecycle-naming.md` §1

## Stage skills

| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `decompose-feature-to-actor-tracks` | command + dispatch | feature → actor별 implementation track 분해 (frontend/backend/3rd-party/data) | track 내부 atomic task 분해 → `decompose-track-to-tasks` |
| `decompose-track-to-tasks` | command + dispatch | actor track → ordered atomic task list (1 PR scope, verifiable) | track 자체 미분해 → `decompose-feature-to-actor-tracks` 먼저 |
| `map-task-dependencies` | command + dispatch | task DAG — internal + cross-actor edges, cycle 감지, critical path | feature 단위 의존성(§2) → `map-feature-dependencies` |
| `plan-parallel-execution` | command + dispatch | worker batch + sync points — capability fit + bottleneck, AI agent 통합 | 실제 worker 분배 실행(§5) → `dispatch-parallel-agents` |
| `define-acceptance-test-plan` | command + dispatch | per-actor + cross-actor test plan + infra 결정 + acceptance gate | 실제 test 코드 생성(§5) → `generate-tests-from-spec` |
| `estimate-build-timeline` | command + dispatch | critical path 기반 calendar timeline + 4-point + risk buffer | feature 단위 공수(§2) → `estimate-feature-effort` |
| `publish-to-tracker` | command + dispatch | §2 feature spec / §4 task plan → 외부 tracker(GitHub/Linear/Jira) 발행 | 내부 계획 단계 (외부 발행 불필요 시) |

## Disambiguation (노드 내)
- `decompose-feature-to-actor-tracks`(feature→track) vs `decompose-track-to-tasks`(track→task): 분해 깊이.
- `plan-parallel-execution`(§4 계획) vs `dispatch-parallel-agents`(§5 실행): 계획 vs 실제 worker 분배.
- `define-acceptance-test-plan`(§4 계획) vs `generate-tests-from-spec`(§5 생성): test plan vs test skeleton.

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3. DoR(SDD) 부재 → §3 `design-system` 선행.
