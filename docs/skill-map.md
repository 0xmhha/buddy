# Buddy 상용 제품 빌딩 스킬맵 초안

## 목적

`buddy` 프로젝트는 새로운 아이디어를 상용 수준의 제품으로 구체화하고, 구현, 테스트, 배포, 운영까지 지원하는 스킬셋을 모으는 것을 목표로 한다.

이 문서는 현재 `buddy` 경로의 스킬들을 제품 빌딩 흐름의 작은 단계로 분리하고, `/Users/wm-it-22-00661/Work/github/study/ai/01.study/docs/projects/mattpocock-skills`에서 가져오면 좋은 기능을 검토하기 위한 초안이다.

검토 대상:

- Buddy: `./ko`
- Matt skills: `/Users/wm-it-22-00661/Work/github/study/ai/01.study/docs/projects/mattpocock-skills`

## 전체 흐름

상용 제품 빌딩 흐름은 아래 단계로 나눌 수 있다.

1. Idea Discovery
2. Business Validation
3. Product Definition
4. Feature Definition & Orchestration
5. Architecture & Code Design
6. Development
7. Debugging
8. Testing & QA
9. Security, Legal, Compliance
10. Release & Deployment
11. Operations

각 단계는 더 작은 작업 단위로 쪼갤 수 있고, 각 작업 단위는 독립 스킬 또는 조합 스킬로 지원할 수 있다. PRD 이후의 실행 단위는 단순 task가 아니라 `feature`를 중심으로 관리한다. 여기서 feature는 구현 완료 후 cherry-pick 또는 patch로 다른 제품에 재사용할 수 있을 만큼 독립적인 기능 단위다.

## 1. Idea Discovery

목적: 아이디어가 실제 문제인지, 누가 그 문제를 겪는지, 해결할 가치가 있는지 확인한다.

현재 Buddy 스킬:

- `validate-idea`
- `review-scope`
- `critique-plan`
- `apply-builder-ethos`

Matt skills에서 추출할 가치:

- `grill-me`: 애매한 아이디어를 끝까지 질문해서 모호성을 제거하는 인터뷰 루프.
- `grill-with-docs`: 아이디어 검토 중 생긴 핵심 용어를 도메인 문서로 남기는 방식.

통합 아이디어:

- `validate-idea`에 "모든 결정 분기가 해소되기 전 구현하지 않는다"는 원칙을 추가한다.
- `review-scope`에 fuzzy language를 domain language로 정제하는 절차를 넣는다.
- 중요한 결정, 되돌리기 어려운 선택, 명시적 trade-off는 ADR 후보로 기록한다.

## 2. Business Validation

목적: 시장성, 수익성, 경쟁, 고객 획득, 가격, 유통 가능성을 검토한다.

현재 Buddy 스킬:

- `validate-idea`
- `review-scope`
- `critique-plan`

부족한 기능:

- 시장 규모 검토
- 고객 세그먼트와 구매자 구분
- pricing 및 willingness-to-pay 검토
- GTM/channel 검토
- 경쟁/대체재 분석
- unit economics 검토
- 법률/규제 영향이 사업성에 미치는 영향 검토

Matt skills에서 직접 가져올 기능은 많지 않다. 다만 `grill-me`의 강한 질문 구조를 응용해 별도 `assess-business-viability` 스킬을 만드는 것이 좋다.

추가 후보 스킬:

- `assess-business-viability`
- `map-customer-segments`
- `review-pricing-and-gtm`
- `analyze-competition-and-substitutes`

## 3. Product Definition

목적: 검증된 아이디어를 PRD, 유저스토리, 범위, out-of-scope, 성공 기준으로 고정하고, 이후 feature로 분해할 기준을 만든다.

현재 Buddy 스킬:

- `review-scope`
- `critique-plan`
- `autoplan`
- `route-spec-to-code`

Matt skills에서 추출할 가치:

- `to-prd`: 대화/계획을 PRD로 만드는 템플릿.
- `grill-with-docs`: 제품 용어와 도메인 지식 유지.
- `to-issues`: PRD를 실행 가능한 vertical slice로 분해.

통합 아이디어:

- `write-prd` 스킬을 만든다.
- PRD에는 Problem Statement, Solution, User Stories, Implementation Decisions, Testing Decisions, Out of Scope를 포함한다.
- PRD에는 성공 기준을 반드시 포함한다. 성공 기준은 사용자 행동, 비즈니스 지표, 품질 기준, 운영 기준으로 나누어 측정 가능하게 작성한다.
- PRD에는 feature 후보 목록을 포함한다. 각 후보는 재사용 가능한 기능 단위인지, 특정 제품에만 종속된 작업인지 구분한다.
- 구현 모듈, 인터페이스 변경, 테스트 대상까지 PRD에 기록한다.
- PRD는 가능하면 issue tracker와 feature registry에 발행해 이후 triage, feature orchestration, 구현 이력 추적으로 연결한다.

## 4. Feature Definition & Orchestration

목적: 큰 제품 계획을 human/agent가 수행 가능한 feature 단위로 분해하고, feature 명세, 재사용 여부, 구현 상태, 커밋 링크를 관리한다.

Feature 정의:

- feature는 사용자가 인식할 수 있거나 제품 동작을 독립적으로 개선하는 수행 가능한 기능 단위다.
- feature는 구현 완료 후 cherry-pick 또는 patch로 다른 제품에 재사용할 수 있을 만큼 코드 변경 범위가 응집되어야 한다.
- feature는 여러 product 구현에서 재사용 가능한 단위여야 하며, 제품별 copy 작업이 아니라 명세, 코드, 테스트, 커밋 이력을 함께 재활용할 수 있어야 한다.
- feature가 너무 크면 여러 feature로 분리하고, 너무 작아 재사용 가치가 없으면 task 또는 subtask로 낮춘다.

Feature 명세서 필수 포맷:

- `feature_id`: 안정적인 고유 ID.
- `name`: 기능명.
- `summary`: 한 문장 설명.
- `problem`: 해결하는 사용자/제품 문제.
- `scope`: 포함 범위.
- `out_of_scope`: 제외 범위.
- `reuse_intent`: 어떤 제품/도메인에서 재사용 가능한지.
- `interfaces`: API, UI, CLI, 이벤트, 데이터 모델 등 외부 접점.
- `dependencies`: 선행 feature, 라이브러리, 인프라, 권한.
- `acceptance_criteria`: 완료 판단 기준.
- `test_plan`: 단위/통합/E2E/수동 QA 기준.
- `implementation_notes`: 설계 결정, trade-off, migration 주의점.
- `code_links`: commit, PR, patch, cherry-pick source 링크.
- `status`: `draft`, `candidate`, `ready`, `in-progress`, `implemented`, `verified`, `deprecated`.
- `owners`: human/agent 책임자.
- `updated_at`: 마지막 갱신 시각.

현재 Buddy 스킬:

- `autoplan`
- `route-spec-to-code`
- `route-intent`
- `route-multi-platform`

Matt skills에서 추출할 가치:

- `to-issues`: vertical slice 기반 작업 분해.
- `triage`: issue lifecycle state machine.
- `setup-matt-pocock-skills`: issue tracker, triage label, domain docs 설정.

통합 아이디어:

- `define-feature-spec` 스킬을 만든다.
- `query-feature-registry` 스킬을 만든다. PRD와 설계 초안을 기준으로 기존 feature를 유사도 검색하고, 재사용/변형/신규 작성 여부를 판단한다.
- `split-work-into-features` 스킬을 만든다.
- 각 feature는 schema, API, UI, tests를 가능한 한 좁고 완결된 경로로 관통한다.
- 각 feature에는 acceptance criteria, blocked by, user stories covered, HITL/AFK 여부를 기록한다.
- feature 내부의 구현 작업은 task/subtask로 분해하되, 최상위 재사용 단위는 feature로 유지한다.
- `triage-work-items` 스킬을 만들어 `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix` 상태를 운영한다.

Feature MCP:

- `feature.query`: natural language, PRD fragment, API shape, domain tag를 입력받아 유사 feature를 검색한다.
- `feature.store`: 신규 feature 명세를 저장한다.
- `feature.update`: status, acceptance criteria, code_links, implementation_notes, test_plan을 갱신한다.
- `feature.link_code`: feature와 commit/PR/patch/cherry-pick source를 연결한다.
- `feature.similarity`: 기존 feature와 신규 후보의 재사용 가능성을 점수화하고 차이를 요약한다.
- `feature.export_patch`: 연결된 commit 또는 PR에서 재사용 가능한 patch/cherry-pick 후보를 만든다.

Feature 관리 SaaS 방향:

- GitHub가 repository, issue, PR, commit, release를 관리하듯이 feature SaaS는 product, feature, spec, implementation, patch, reuse graph를 관리한다.
- Product는 여러 feature를 참조하고, feature는 여러 product에서 재사용될 수 있다.
- Feature page는 README처럼 명세서를 보여주고, Issues처럼 논의와 상태를 관리하며, Pull Request처럼 코드 링크와 검증 결과를 연결한다.
- 검색은 키워드뿐 아니라 embedding 기반 유사도, interface shape, domain tag, dependency graph를 함께 사용한다.
- 재사용 가능한 feature에는 canonical implementation과 variant implementation을 구분한다.
- 각 feature는 commit history, patch artifact, test evidence, release adoption 정보를 가진다.
- 조직은 private/public feature registry를 운영할 수 있고, public feature는 GitHub marketplace와 유사하게 discovery, star, fork, adoption metric을 제공한다.
- agent는 계획 수립 단계에서 feature SaaS/MCP를 먼저 조회하고, 기존 구현이 충분히 유사하면 patch 또는 cherry-pick 기반으로 구현 단계를 단축한다.

상세 설계는 [Feature Management SaaS & MCP 구체화](./feature-management-saas-mcp.md)를 기준으로 별도 관리한다.

## 5. Architecture & Code Design

목적: 구현 전후로 구조적 리스크를 줄이고 유지보수 가능한 설계를 만든다.

현재 Buddy 스킬:

- `review-engineering`
- `measure-code-health`
- `classify-review-risks`
- `consult-codex`

Matt skills에서 추출할 가치:

- `improve-codebase-architecture`: deep module, interface depth, locality, leverage.
- `zoom-out`: 낯선 코드 영역의 시스템 맥락 설명.
- `tdd`의 interface-design/deep-modules 참고 문서.

통합 아이디어:

- `review-code-architecture` 스킬을 강화한다.
- `review-engineering`에 deep module 기준을 추가한다.
- "interface가 test surface다"는 원칙을 명시한다.
- shallow module, pass-through abstraction, locality 부족, interface leakage를 탐지한다.
- 도메인 용어에 없는 개념으로 모듈명을 짓는다면 `CONTEXT.md` 갱신 후보로 본다.

## 6. Development

목적: 기능을 실제 코드로 만들되 빠르고 정확한 피드백 루프를 유지한다.

현재 Buddy 스킬:

- `iterate-fix-verify`
- `review-engineering`
- `route-spec-to-code`
- `freeze-edit-scope`

Matt skills에서 추출할 가치:

- `tdd`: red-green-refactor 개발 루프.
- `diagnose`: bug reproduction 중심 수정 루프.

통합 아이디어:

- `build-with-tdd` 스킬을 만든다.
- tracer bullet을 먼저 만든다.
- 한 번에 하나의 behavior test만 작성한다.
- 현재 test를 통과시키는 만큼만 구현한다.
- refactor는 test가 통과한 뒤 수행한다.
- implementation detail보다 observable behavior를 테스트한다.

## 7. Debugging

목적: 증상에 반응해 임의 수정하지 않고, 재현 가능한 원인 분석을 통해 버그를 해결한다.

현재 Buddy 스킬:

- `iterate-fix-verify`
- `run-browser-qa`
- `monitor-regressions`

Matt skills에서 추출할 가치:

- `diagnose`: reproduce, hypothesize, instrument, fix, regression test로 이어지는 전용 디버깅 루프.

통합 아이디어:

- `diagnose-bug` 스킬을 만든다.
- feedback loop를 먼저 만든다.
- minimized repro를 만든다.
- 여러 hypothesis를 세우고, 각 hypothesis를 구분할 targeted instrumentation을 추가한다.
- fix 후 regression test를 추가한다.
- 원래 재현 경로도 다시 검증한다.

## 8. Testing & QA

목적: 구현된 기능이 실제 사용자 흐름과 운영 환경에서 동작하는지 확인한다.

현재 Buddy 스킬:

- `classify-qa-tiers`
- `run-browser-qa`
- `iterate-fix-verify`
- `monitor-regressions`

Matt skills에서 추출할 가치:

- `tdd`: behavior-first test 기준.
- `diagnose`: 버그 재현을 regression test로 바꾸는 기준.

통합 아이디어:

- Quick, Standard, Exhaustive QA tier와 TDD cycle을 연결한다.
- bug 발견 시 `minimized repro -> failing test -> fix -> re-verify`로 처리한다.
- UI QA는 browser automation과 accessibility tree interaction을 포함한다.
- 테스트는 data shape나 implementation detail이 아니라 외부 behavior 중심으로 설계한다.

## 9. Security, Legal, Compliance

목적: 출시 전 공격, 개인정보, 법률, 규제, AI 책임 리스크를 별도로 검토한다.

현재 Buddy 스킬:

- `audit-security`
- `guard-destructive-commands`
- `compose-safety-mode`
- `classify-review-risks`

Matt skills에서 추출할 가치:

- `git-guardrails-claude-code`: destructive git command hook 구현 레시피.

부족한 기능:

- 법률 검토
- 개인정보/PII 처리 검토
- AI 서비스의 책임 범위 검토
- 저작권/라이선스 검토
- 약관/개인정보처리방침 readiness 검토
- 결제/환불/소비자보호 검토
- 산업별 규제 검토

추가 후보 스킬:

- `review-legal-regulatory`
- `review-privacy-data-risk`
- `review-ai-safety-liability`
- `review-terms-policy-readiness`
- `review-license-and-ip-risk`

주의: 이 단계는 법률 자문을 대체하지 않는다. 목적은 전문가 검토가 필요한 쟁점을 빠뜨리지 않도록 checklist와 evidence package를 만드는 것이다.

## 10. Release & Deployment

목적: 배포 전 품질, 문서, changelog, 회귀, security gate를 확인한다.

현재 Buddy 스킬:

- `write-changelog`
- `sync-release-docs`
- `measure-code-health`
- `monitor-regressions`
- `audit-security`

Matt skills에서 추출할 가치:

- `setup-pre-commit`: pre-commit 품질 gate 구성.

통합 아이디어:

- `setup-quality-gates` 스킬을 만든다.
- Husky, lint-staged, Prettier, typecheck, test를 개발환경에 연결한다.
- 배포 전 typecheck, lint, test, health score를 확인한다.
- release docs drift와 changelog 누락을 확인한다.
- security audit과 regression check를 release gate로 연결한다.

## 11. Operations

목적: 출시 후 운영 이슈, 회귀, 컨텍스트 보존, 회고, 학습 축적을 지원한다.

현재 Buddy 스킬:

- `monitor-regressions`
- `summarize-retro`
- `save-context`
- `restore-context`
- `persist-learning-jsonl`
- `measure-code-health`

Matt skills에서 추출할 가치:

- `triage`: 운영 중 들어오는 issue state machine.
- `diagnose`: 운영 이슈 원인 분석.
- `grill-with-docs`: 운영 중 배운 도메인/결정 업데이트.

통합 아이디어:

- 운영 이슈는 `needs-triage`부터 시작한다.
- 반복되는 거절/범위 제외는 out-of-scope 문서 또는 ADR로 기록한다.
- 운영 중 배운 내용은 append-only learning/context로 저장한다.
- weekly/sprint retro는 health trend, regression trend, issue throughput과 연결한다.

## Matt Skills 우선 추출 후보

우선순위가 높은 순서:

1. `diagnose`
   - Buddy의 debugging 역량을 크게 보강한다.
   - `diagnose-bug`로 재구성 추천.

2. `tdd`
   - 개발 단계의 red-green-refactor discipline을 추가한다.
   - `build-with-tdd`로 재구성 추천.

3. `to-issues`
   - product plan을 실행 가능한 feature와 하위 task로 바꾼다.
   - `split-work-into-features`와 `define-feature-spec`로 재구성 추천.

4. `to-prd`
   - idea/review 결과를 공식 제품 문서로 고정한다.
   - `write-prd`로 재구성 추천.

5. `triage`
   - 운영/이슈 관리 state machine을 제공한다.
   - `triage-work-items`로 재구성 추천.

6. `grill-with-docs`
   - 도메인 언어, ADR, 장기 컨텍스트 관리에 중요하다.
   - `define-product-context` 또는 `maintain-domain-context`로 재구성 추천.

7. `improve-codebase-architecture`
   - 상용 제품의 장기 유지보수성을 보강한다.
   - `review-code-architecture`에 통합 추천.

8. `setup-pre-commit`
   - DevEx/quality gate 자동화에 바로 쓸 수 있다.
   - `setup-quality-gates`로 재구성 추천.

## 낮은 우선순위 또는 선택적 추출 후보

- `zoom-out`
  - 낯선 코드 영역을 이해하는 데 유용하다.
  - `review-engineering` 또는 `consult-codex`의 보조 패턴으로 흡수 가능.

- `caveman`
  - 커뮤니케이션 압축 모드.
  - 장기 작업 비용 절감에는 유용하지만 제품 빌딩 핵심 단계는 아니다.

- `write-a-skill`
  - Buddy가 자체 스킬을 계속 만들 계획이면 가치가 있다.
  - skill authoring guideline으로 별도 유지 가능.

- `git-guardrails-claude-code`
  - Buddy의 `guard-destructive-commands`와 중복된다.
  - Claude hook 구현 레시피만 추출하면 된다.

- `scaffold-exercises`
  - 교육/코스 제작용이라 Buddy의 상용 제품 빌딩 목적과는 거리가 있다.

- `migrate-to-shoehorn`
  - TypeScript 테스트 fixture 전용 마이그레이션이다.
  - 범용 Buddy 스킬셋에는 낮은 우선순위다.

## 추천 신규 Buddy 스킬 목록

Matt skills 기반:

- `define-product-context`
- `write-prd`
- `define-feature-spec`
- `query-feature-registry`
- `split-work-into-features`
- `triage-work-items`
- `build-with-tdd`
- `diagnose-bug`
- `review-code-architecture`
- `setup-quality-gates`

Buddy 목표상 추가 필요:

- `assess-business-viability`
- `review-pricing-and-gtm`
- `review-legal-regulatory`
- `review-privacy-data-risk`
- `review-ai-safety-liability`
- `review-license-and-ip-risk`
- `review-terms-policy-readiness`

## 추천 번들

### Idea To PRD

```text
validate-idea
-> review-scope
-> critique-plan
-> define-product-context
-> write-prd
```

### PRD To Build Plan

```text
write-prd
-> query-feature-registry
-> define-feature-spec
-> review-engineering
-> review-design
-> review-devex
-> split-work-into-features
-> triage-work-items
```

### Build Loop

```text
build-with-tdd
-> iterate-fix-verify
-> run-browser-qa
-> diagnose-bug
```

### Release Gate

```text
measure-code-health
-> audit-security
-> monitor-regressions
-> sync-release-docs
-> write-changelog
```

### Operations Loop

```text
monitor-regressions
-> triage-work-items
-> diagnose-bug
-> save-context
-> summarize-retro
```

## 검토 포인트

다음 검토에서는 아래를 결정하면 된다.

1. Matt skills를 원문에 가깝게 가져올지, Buddy 스타일의 새 skill로 재작성할지.
2. `legal/regulatory/business viability` 영역을 별도 phase로 강하게 둘지.
3. Buddy의 기존 `autoplan`이 어디까지 feature orchestration하고, 어디서 feature registry 또는 issue/task 발행 skill로 넘길지.
4. `CONTEXT.md`/ADR 기반 도메인 문서화를 Buddy의 핵심 운영 원칙으로 둘지.
5. 상용 제품 기준 release gate의 최소 필수 항목을 무엇으로 둘지.
6. feature MCP의 최소 API를 `query/store/update/link_code/export_patch`로 시작할지, SaaS UI의 GitHub 유사 모델까지 함께 설계할지.
