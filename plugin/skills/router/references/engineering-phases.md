# Engineering Phases — Artifact-Based Definition

> **목적**: 9-phase 라이프사이클의 각 단계를 **산출물(artifact) 기준**으로 정의한다.
> 각 phase는 "무슨 산출물을 입력받고, 무슨 산출물을 생산하며, 무슨 의사결정을 내리는가"로 정체성이 결정된다.
>
> **용도**:
> - 개별 스킬의 Input/Output Contract 작성 시 기준 문서
> - 스킬 간 cascade 연결의 근거 (A의 output = B의 required input)
> - `status` 스킬의 artifact 탐지 로직 근거
>
> **관계**:
> - [`skill-catalog.md`](./skill-catalog.md) — 9-phase별 스킬 목록 (what)
> - [`routing-rules.md`](./routing-rules.md) — 스킬 간 충돌 결정 (which)
> - 본 문서 — phase 정체성 + 산출물 계약 + 전이 규칙 (why & when)

---

## §1. 정의 원칙

### 왜 Artifact-based인가

소프트웨어 공학의 phase 정의 접근은 세 갈래가 있다:

| 접근 | 정의 기준 | 장점 | 단점 |
|------|----------|------|------|
| Activity-based | 뭘 하는가 | 직관적 | 순차적으로 보이지만 실제로는 아님 |
| **Artifact-based** | **뭘 만드는가** | **진입/종료 조건 명확, Input/Output Contract와 자연 정합** | phase 경계에서 artifact 정의 합의 필요 |
| Decision-based | 뭘 결정하는가 | 비가역성 기반 리스크 관리 | 결정이 phase를 넘나드는 경우 분류 모호 |

buddy는 **Artifact-based를 주축**으로 채택한다.

- 스킬의 Input/Output Contract가 artifact 단위이므로, phase 정의도 동일 단위로 통일
- cascade 연결이 "A의 output artifact → B의 required input artifact"로 기계적 도출 가능
- 진입 조건 판단이 "이 artifact가 존재하는가?"로 통일 (ADR-007 stateless 원칙 호환)

각 phase의 Activity(뭘 하는가)와 Decision(뭘 결정하는가)도 참조 정보로 함께 기술하되, **정체성의 근거는 artifact**이다.

### Phase 간 관계

phase 간 기본 흐름은 순차적이지만, 조건에 따라 backtrack(이전 phase 복귀) 또는 skip(건너뛰기)이 발생한다. 이 전이 규칙은 §3에서 정의한다.

```
§1 → §2 → §3 → §4 → §5 → §6 → §7 → §8 ⇄ §2
                                              ↓
                                             §9
```

---

## §2. Phase 정의

### Phase 1 — Problem/Opportunity Identification & Validation

| 항목 | 내용 |
|------|------|
| **정체성** | 문제 또는 기회의 존재를 확인하고, 해결/실행할 가치가 있는지 검증한다 |
| **핵심 질문** | "무엇이 문제(또는 기회)이고, 해결할 가치가 있는가?" |

Phase 1은 **프로덕트 존재 여부**에 따라 2가지 mode로 동작한다. 어떤 mode든 종료 시 동일한 output 형태 — **검증된 작업 항목(validated work item) + 다음 phase 결정** — 을 생산한다.

#### Mode A — Greenfield (신규 프로덕트)

기존 프로덕트가 없고, 아이디어 단계에서 시작하는 경우.

| 항목 | 내용 |
|------|------|
| **Orchestrator** | `concretize-idea` |
| **연상** | 막연한 아이디어를 concrete(구체적)하게 굳힌다 |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Problem statement / idea sketch | ✅ | 사용자가 제공하는 비정형 입력 (대화, 메모, 한 줄 아이디어) |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| PRD (Product Requirements Document) | `docs/prd.md` | Phase 2 `define-features` |
| Business viability report | PRD 내 섹션 또는 별도 문서 | Phase 1 내부 결정 근거 |
| Market/competitor analysis | PRD 내 섹션 | Phase 1 내부 결정 근거 |
| Customer segment map | PRD 내 섹션 | Phase 2 `identify-actors` |

**종료 → 다음**: Phase 2 `define-features` (전체 흐름 진입).

#### Mode B — Existing Product (기존 프로덕트 변경)

이미 운영 중인 프로덕트에 대한 모든 변경 작업 — 버그 수정, 신규 기능, 성능 개선, 기술 부채 정리, 의존성 갱신 등 유형 무관. 핵심은 **변경의 scope(범위)를 평가**하여 다음 phase를 결정하는 것.

| 항목 | 내용 |
|------|------|
| **Orchestrator** | `assess-product-change` *(신규)* |
| **연상** | 기존 프로덕트에 대한 변경(change)을 평가(assess)한다 |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Change trigger | ✅ | 버그 리포트, 기능 요청, 성능 메트릭, 기술 부채 신호 등 — 유형 무관 |
| Existing codebase / product context | ✅ | 현재 아키텍처, 코드, 사용자 기반, 기술 스택 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Validated work item | structured (문제/기회 설명 + 재현/근거 + 수용 기준) | 다음 phase 진입 스킬 |
| Impact assessment | 영향 범위 + severity/priority + 기존 시스템 호환성 | 다음 phase 결정 근거 |
| Scope classification + routing decision | small / medium / large + 다음 phase 번호 | Phase 전이 |

**종료 → scope별 routing:**

| Scope | 다음 경로 | 예시 |
|-------|----------|------|
| **Small** | → Phase 5 직접 진입 | 버그 수정, 설정 변경, 작은 UI 수정 |
| **Medium** | → Phase 3 (설계 검토 후 구현) | 새 API endpoint, 컴포넌트 리팩토링, 스키마 변경 |
| **Large** | → Phase 2 (feature 정의부터) | 신규 기능, 대규모 재설계, 아키텍처 변경 |
| **Defer/Reject** | → backlog 기록, 현재 cycle 종료 | 우선순위 낮음, ROI 부족 |

---

### Phase 2 — Feature Definition & Backlog

| 항목 | 내용 |
|------|------|
| **정체성** | PRD를 actor/use case 분해를 거쳐 구현 가능한 feature 단위로 변환한다 |
| **핵심 질문** | "무엇을 만드는가?" |
| **Orchestrator** | `define-features` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| PRD (Mode A 경유) 또는 Validated work item (Mode B, scope=large) | ✅ | Phase 1 산출물 — greenfield는 PRD, 기존 프로덕트 대규모 변경은 validated work item |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Actor list | structured (PRD 내 또는 별도) | Phase 2 내부, Phase 3 `design-api-contract` |
| Use case map (actor별) | structured | Phase 2 내부, Phase 4 `decompose-feature-to-actor-tracks` |
| System boundary map | structured | Phase 3 `derive-system-topology` |
| Feature specs | `docs/feature-spec/` | Phase 3 `design-system`, Phase 4 `plan-build` |
| Feature backlog (priority + estimate) | structured | Phase 4 `plan-build` |

**종료 조건**: Feature backlog이 priority-ordered 상태로 존재하고, 각 feature에 actor/use case/acceptance criteria가 정의된 상태.

---

### Phase 3 — Technical Design (Architecture)

| 항목 | 내용 |
|------|------|
| **정체성** | feature를 구현하기 위한 기술적 구조를 결정한다 |
| **핵심 질문** | "어떻게 만드는가?" |
| **Orchestrator** | `design-system` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Feature specs / backlog | ✅ | Phase 2 산출물 (정상 흐름) |
| Validated work item (Mode B, scope=medium) | ✅ | Phase 1 `assess-product-change`에서 직접 진입 시 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Tech stack ADR | `docs/decisions/` | Phase 5 구현 기준 |
| API contracts | `docs/api-contract/` 또는 OpenAPI spec | Phase 5 `generate-from-api-contract` |
| Data model (schema + migration plan) | `docs/data-model/` | Phase 5 구현, Phase 6 테스트 |
| Architecture Decision Records | `docs/decisions/` | 전 phase 참조 |
| Design system / interaction patterns | 별도 문서 | Phase 5 UI 구현 |
| Deploy strategy | ADR 또는 별도 문서 | Phase 7 `ship-release` |
| Observability / secret / auth strategy | 각 별도 문서 | Phase 5 구현, Phase 8 운영 |

**종료 조건**: 핵심 기술 결정(tech stack, API, data model)이 ADR로 기록되고, 리뷰(autoplan 4-mode)를 통과한 상태.

**참고**: Phase 3은 가장 많은 stage skill을 보유한 phase. 각 design-* 스킬이 독립적으로도 호출 가능(standalone-with-context)하지만, orchestrator 경유 시 결정 간 일관성 보장.

---

### Phase 4 — Implementation Planning

| 항목 | 내용 |
|------|------|
| **정체성** | 기술 설계를 실행 가능한 task 단위로 분해하고 순서를 정한다 |
| **핵심 질문** | "어떤 순서로 만드는가?" |
| **Orchestrator** | `plan-build` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Technical design docs (ADR, API contract, data model) | ✅ | Phase 3 산출물 |
| Feature specs | ✅ | Phase 2 산출물 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Actor-track decomposition | `docs/actor-track-plan.yaml` | Phase 5 `build-feature` |
| Task DAG (dependency graph) | structured | Phase 5 `dispatch-parallel-agents` |
| Parallel execution plan | structured | Phase 5 병렬 개발 |
| Acceptance test plan | structured | Phase 5 `generate-tests-from-spec`, Phase 6 검증 기준 |
| Build timeline (estimate) | structured | 프로젝트 관리 |

**종료 조건**: task DAG가 존재하고, critical path가 식별되며, acceptance test plan이 정의된 상태.

---

### Phase 5 — Development (Implementation)

| 항목 | 내용 |
|------|------|
| **정체성** | 계획된 task를 코드로 구현한다 |
| **핵심 질문** | "만든다" |
| **Orchestrator** | `build-feature` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Actor-track plan / Task DAG | ✅ | Phase 4 산출물 (정상 흐름) |
| Technical design docs | ✅ | Phase 3 산출물 (구현 기준) |
| Validated work item (Mode B, scope=small) | ✅ | Phase 1 `assess-product-change`에서 직접 진입 시 — plan/design 없이 work item만으로 구현 |
| Acceptance test plan | 선택 | Phase 4 산출물 (TDD 시 활용) |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Working code | source files + commit history | Phase 6 검증 대상 |
| Tests (unit / integration / contract) | test files | Phase 6 `verify-quality` |
| Updated docs (코드 변경 동기화) | docs/ 갱신 | Phase 7 release docs |

**종료 조건**: 모든 task가 완료되고, 테스트가 통과하며, 코드가 commit된 상태.

**특이사항**: Phase 5의 스킬 다수(build-with-tdd, diagnose-bug, iterate-fix-verify)는 full-standalone 등급 — 별도 orchestrator 없이 독립 실행이 자연스러운 영역.

---

### Phase 6 — Verification (Quality)

| 항목 | 내용 |
|------|------|
| **정체성** | 구현된 코드가 요구사항을 만족하고 상용 품질 기준을 통과하는지 검증한다 |
| **핵심 질문** | "제대로 작동하는가?" |
| **Orchestrator** | `verify-quality` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Working code + tests | ✅ | Phase 5 산출물 |
| Acceptance criteria / test plan | ✅ | Phase 4 산출물 또는 feature spec |
| API contracts / data model | 선택 | Phase 3 산출물 (contract test 기준) |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| QA report (per-actor + cross-actor) | structured | Phase 7 launch checklist 입력 |
| Security audit report | structured | Phase 7 launch checklist 입력 |
| Compliance sign-off (privacy, license, terms) | structured | Phase 7 launch checklist 입력 |
| Code health score | 0-10 composite | Phase 7 release 판단 |
| Coverage report (line + mutation + behavior) | structured | Phase 5 backtrack 시 보강 기준 |

**종료 조건**: 모든 quality gate(test, security, compliance, code health)가 통과한 상태.

**Backtrack trigger**: quality gate 실패 → Phase 5로 복귀 (fix and re-verify).

---

### Phase 7 — Release

| 항목 | 내용 |
|------|------|
| **정체성** | 검증된 코드를 사용자에게 전달한다 |
| **핵심 질문** | "내보낸다" |
| **Orchestrator** | `ship-release` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| QA report + sign-offs | ✅ | Phase 6 산출물 |
| Working code (quality gate 통과) | ✅ | Phase 5→6 통과 산출물 |
| Changelog draft | 선택 | Phase 5 `update-docs-with-code` 산출물 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Tagged release (semver) | git tag + release notes | 사용자, Phase 8 운영 기준 |
| Deployed artifact | 배포된 binary / container / package | Phase 8 모니터링 대상 |
| Launch checklist pass | structured checklist | 감사 증적 |
| Updated docs (CHANGELOG, README, ADR) | docs/ 갱신 | 사용자, 다음 cycle |

**종료 조건**: release tag가 존재하고, 배포가 완료되며, launch checklist 전 항목이 통과한 상태.

**Backtrack trigger**: UAT 실패 → Phase 5 또는 Phase 6로 복귀.

---

### Phase 8 — Operations (Operate & Iterate)

| 항목 | 내용 |
|------|------|
| **정체성** | 운영 중인 시스템을 관찰하고, 데이터 기반으로 개선한다 |
| **핵심 질문** | "잘 돌아가고 있는가?" |
| **Orchestrator** | `iterate-product` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Production system (deployed, traffic 발생 중) | ✅ | Phase 7 산출물 |
| Monitoring data (metrics, logs, traces) | ✅ | 운영 인프라 산출물 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Experiment results (A/B, funnel) | structured report | Phase 2 재진입 (improvement → feature) |
| Incident reports + postmortem | structured | Phase 5 hotfix, Phase 3 design 재검토 |
| Improvement backlog | structured task list | Phase 2 `define-features` 재진입 |
| Operational metrics (SLO, error budget, cost) | dashboard / report | Phase 8 자체 loop |

**종료 조건**: 자연적 종료 없음 — 지속적 loop. Phase 2 재진입(개선 기능) 또는 Phase 9 진입(폐기 결정) 시 해당 cycle 종료.

---

### Phase 9 — Lifecycle Management

| 항목 | 내용 |
|------|------|
| **정체성** | 노후화된 feature/product의 수명을 결정하고 정리한다 |
| **핵심 질문** | "유지할 것인가, 끝낼 것인가?" |
| **Orchestrator** | `manage-lifecycle` |

**Input Artifacts:**

| Artifact | Required | 설명 |
|----------|----------|------|
| Usage/adoption data | ✅ | Phase 8 산출물 |
| Business decision (유지/폐기) | ✅ | 사용자 의사결정 |

**Output Artifacts:**

| Artifact | 형식 | 소비자 |
|----------|------|--------|
| Deprecation plan + timeline | structured | 사용자 공지, Phase 5 cleanup 코드 |
| Migration plan | structured | 사용자, 연관 시스템 |
| EOL documentation | structured | 감사 증적 |
| Knowledge preservation | docs/ archive | 다음 프로젝트 참고 |

**종료 조건**: feature/product가 sunset 되고, 사용자 migration이 완료되며, 코드 cleanup이 끝난 상태.

---

### Cross-cutting (Phase 소속 없음)

어느 phase에서든 호출 가능한 스킬. phase 정체성이 아닌 **적용 맥락**으로 정의된다.

| 범주 | 스킬 예시 | 적용 시점 |
|------|----------|----------|
| 편향 방지 | `verify-best-alternative` | 의사결정 직전 (주로 Phase 3) |
| 문맥 보존 | `save-context`, `restore-context` | 세션 전환 시 |
| 메타-스킬 | `write-a-skill`, `status` | 스킬 개발, 현재 위치 파악 |
| 안전장치 | `guard-destructive-commands`, `compose-safety-mode`, `freeze-edit-scope` | 위험 명령 실행 시 |
| 패턴 라이브러리 | `classify-qa-tiers`, `classify-review-risks`, `monitor-regressions` | 다른 스킬 내부에서 ambient 적용 |
| Blocker 분해 | `decompose-blocker` | 작업 stuck 상태 (주로 Phase 5) |
| 어휘 일관성 | `audit-ubiquitous-language` | 리팩토링 전, PR 리뷰, 신규 feature 정의 시 |

---

## §3. Phase 전이 규칙

### 정상 흐름 (forward)

```
§1 → §2 → §3 → §4 → §5 → §6 → §7 → §8 → §2 (개선 cycle)
                                              → §9 (폐기 결정 시)
```

### Backtrack (이전 phase 복귀)

| 현재 Phase | 복귀 대상 | 조건 |
|-----------|----------|------|
| §5 Development | §3 Design | 구현 중 설계 모호/누락 발견 |
| §6 Verification | §5 Development | quality gate 실패 (fix 필요) |
| §7 Release | §5 or §6 | UAT 실패 |
| §8 Operations | §5 Development | hotfix 필요 (인시던트) |
| §8 Operations | §3 Design | 구조적 문제 발견 (design 재검토) |

### Skip (phase 건너뛰기)

**Greenfield (Mode A) skip:**

| 상황 | 경로 | 근거 |
|------|------|------|
| Prototype / POC | §1 → §2 → §3 → **§5** (§4 skip) | 형식적 task 분해 불필요 |
| Design-only 작업 | §1 → §2 → §3 → 종료 (§4-§9 skip) | 구현 없이 설계 문서만 산출 |

**Existing Product (Mode B) scope-based routing:**

Mode B에서는 skip이 아니라 `assess-product-change`의 **scope 판단 결과에 따른 routing**으로 처리된다:

| Scope | 경로 | 예시 |
|-------|------|------|
| Small | §1 → **§5** | 버그 수정, 설정 변경, 작은 UI 수정 |
| Medium | §1 → **§3** → §5 → ... | 새 API endpoint, 스키마 변경, 컴포넌트 리팩토링 |
| Large | §1 → **§2** → §3 → §4 → §5 → ... | 신규 기능, 대규모 재설계 |

**공통 skip:**

| 상황 | 경로 | 근거 |
|------|------|------|
| Hotfix (긴급) | **§5** 직접 진입 (§1 포함 전부 skip) | 문제가 이미 확인되고 즉시 수정이 필요한 경우에만 |

### 전이 판단 기준

phase 전이는 **output artifact의 존재 + 품질**로 판단한다:

1. 현재 phase의 output artifact가 존재하는가?
2. 해당 artifact가 다음 phase의 required input을 만족하는가?
3. 만족하면 forward. 불만족이면 현재 phase 계속 또는 backtrack.

이 판단은 각 스킬의 Input/Output Contract에 의해 자동으로 발생한다 — 스킬 진입 시 required input이 없으면 사용자에게 질의하거나 선행 스킬을 제안한다.

---

## §4. Input/Output Contract 표준 형식

각 스킬의 PROCEDURE.md에 아래 형식의 섹션을 포함한다.

### Input Requirements

```markdown
## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| <name> | ✅ | artifact | <Phase N 산출물 / 특정 스킬 output> | "질의문" |
| <name> | ✅ | knowledge | 사용자 도메인 지식 | "질의문" |
| <name> | 선택 | artifact | <source> | 기본 동작 설명 |
```

**Type 분류:**

| Type | 정의 | 예시 |
|------|------|------|
| `artifact` | 파일 또는 문서로 존재하는 산출물 | PRD, API contract, test report |
| `knowledge` | 사용자의 도메인 지식, 정형화되지 않은 정보 | "주요 read/write 패턴은?", "target 사용자는?" |
| `decision` | 선행 의사결정 결과 | "tech stack 결정", "deploy 전략 결정" |

**미제공 시 처리 원칙:**

1. `Required` + `artifact` type → 선행 스킬 제안 ("먼저 `/buddy:<producing-skill>` 을 실행하세요")
2. `Required` + `knowledge` type → 사용자에게 질의 ("질의문" 컬럼의 질문을 사용)
3. `Required` + `decision` type → 결정을 내릴 수 있는 스킬 제안
4. `선택` → 없어도 진행, 산출물 품질이 낮아질 수 있음을 안내

### Output Contract

```markdown
## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| <name> | artifact | <structured YAML / prose / file path> | <소비 스킬 목록> |
| <name> | decision | <ADR / inline record> | <소비 스킬 목록> |
```

**Consumers 작성 규칙:**

- 직접 소비하는 스킬만 기재 (transitive dependency 제외)
- Phase orchestrator는 소비자로 기재하지 않음 (orchestrator는 stage를 호출할 뿐, artifact를 직접 소비하지 않음)

### Standalone 등급 (자동 도출)

Input Requirements 표에서 자동 결정:

| 등급 | 조건 | 설명 |
|------|------|------|
| **full-standalone** | required input이 모두 `knowledge` type | 사용자 응답만으로 진행 가능 |
| **standalone-with-context** | required input에 `artifact` type이 있지만, 미제공 시 사용자 질의로 대체 가능 | artifact 없이도 사용자가 구두로 정보 제공하면 진행 |
| **orchestrator-preferred** | required input의 `artifact` type이 복수이고, 선행 phase 전체 산출물에 의존 | 독립 실행 가능하나 orchestrator 경유가 품질 보장 |

standalone 등급은 PROCEDURE.md에 별도 기재하지 않는다 — Input Requirements 표에서 읽는 사람이 자연히 판단할 수 있다.

---

## §5. 본 문서의 범위와 한계

### 범위

- 9-phase 정체성 정의 (artifact 기준)
- Input/Output Contract 표준 형식 정의
- Phase 전이 규칙 (forward / backtrack / skip)

### 범위 외

- 개별 스킬의 실제 Input/Output Contract 내용 → 각 PROCEDURE.md에 기재
- 스킬 간 라우팅 충돌 결정 → [`routing-rules.md`](./routing-rules.md)
- 스킬 목록 및 description → [`skill-catalog.md`](./skill-catalog.md)
- SE 이론 baseline 상세 → [`docs/plugin-skills-engineering-flow.md`](../../../../docs/plugin-skills-engineering-flow.md)

### 변경 trigger

| trigger | 갱신 부분 |
|---------|----------|
| 새 phase 추가/분리/병합 | §2 전반 + §3 전이 규칙 |
| Input/Output Contract 형식 변경 | §4 |
| 새 backtrack/skip 패턴 발견 | §3 |
| 개별 스킬의 Input/Output Contract 작성 완료 후 phase 산출물 보정 필요 시 | §2 해당 phase |
