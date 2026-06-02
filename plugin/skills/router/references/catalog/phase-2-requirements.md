# §2 Requirements Specification (기능 정의) — Stage Skills

- **Orchestrator (entry)**: `define-features`
- **DoR → DoD**: PRD / validated work item → SRS = feature backlog + actor·use case map
- 정체성 원본: `engineering-phases.md` §2 Phase 2 | 명사 원본: `se-lifecycle-naming.md` §1

## Stage skills

| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `identify-actors` | dispatch | 시스템 참여 actor 열거 — user/admin/system/3rd-party/external-tool 분류 | use case까지 필요 → `map-actor-use-cases` |
| `map-actor-use-cases` | dispatch | actor별 use case 식별 (UML use case 다이어그램 등가) | actor 미정 → `identify-actors` 먼저 |
| `map-use-case-to-system-boundary` | dispatch | 각 use case가 어느 경계(frontend/backend/external SaaS)에서 실행되는지 매핑 | infra component 매핑(§3) → `map-use-cases-to-infra` |
| `compose-feature-from-use-cases` | dispatch | cross-actor use case를 묶어 feature 정의 | 단일 feature 완전 명세 → `define-feature-spec` |
| `define-feature-spec` | dispatch | feature 완전 명세서 — actor/use case/boundary/acceptance/test plan | feature 후보 합성 전 → `compose-feature-from-use-cases` |
| `score-feature-priority` | dispatch | RICE/ICE/MoSCoW 우선순위 결정 | effort 추정 → `estimate-feature-effort` |
| `estimate-feature-effort` | command + dispatch | T-shirt sizing + ideal-h × multiplier + 4-point uncertainty | 가치/우선순위 → `score-feature-priority` |
| `map-feature-dependencies` | dispatch | feature 간 선후 의존성 DAG + critical path + 병렬 그룹 | task 단위 의존성(§4) → `map-task-dependencies` |
| `split-work-into-features` | dispatch | PRD를 vertical slice 기반 재사용 feature 단위로 분해 | 이미 feature 확정, 명세만 → `define-feature-spec` |
| `query-feature-registry` | dispatch | PRD/feature 후보를 registry에서 검색해 reuse/adapt/inspire | 신규 정의(재사용 후보 없음) → `compose-feature-from-use-cases` |
| `triage-work-items` | dispatch | 이슈/feature/task work item 우선순위 + lifecycle state machine | feature 가치 점수만 → `score-feature-priority` |

## Disambiguation (노드 내)
- `score-feature-priority`(가치·우선순위) vs `estimate-feature-effort`(공수·불확실성): RICE의 분자 vs 분모.
- `map-feature-dependencies`(§2 feature 단위) vs `map-task-dependencies`(§4 task 단위): 추상화 레벨 차이.

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3. DoR(PRD) 부재 → §1 `concretize-idea`/`define-product-spec` 선행.
