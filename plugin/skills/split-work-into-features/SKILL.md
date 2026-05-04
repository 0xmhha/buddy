---
name: split-work-into-features
description: "PRD를 받아 vertical slice 기반 재사용 가능한 feature 단위로 분해. task가 아닌 feature(schema→API→UI→test 관통하는 응집 단위) 명세 작성. 트리거: 'feature로 쪼개줘' / 'vertical slice 분해' / 'PRD를 feature로' / '재사용 가능한 feature' / 'feature 분리' / '독립 단위로 쪼개' / 'FEAT-id 부여'. 입력: define-product-spec 출력 PRD + feature candidate 목록. 출력: feature_id 부여된 feature spec set + dependency graph. 흐름: define-product-spec → split-work-into-features → review-architecture/triage-work-items/build-with-tdd."
type: skill
---

# Split Work Into Features — Vertical Slice 분해

## 1. 목적

PRD 전체 계획을 **feature 단위로 분해**한다. feature는 **단순 task가 아니라**, 제품의 핵심 가치를 담고 향후 재사용 가능한 독립적 기능 단위.

핵심 차별: feature = (DB schema + API + UI + test)를 관통하는 vertical slice. horizontal layer로 분할 금지.

이 스킬은 PRD의 "feature_candidates"를 받아 다음 두 보장을 가진 feature spec set으로 변환:
1. 각 feature는 patch / cherry-pick으로 다른 제품에 이식 가능한 응집도
2. 각 feature는 acceptance criteria로 완료 판단 가능

## 2. 사용 시점 (When to invoke)

- `define-product-spec` PRD 확정 후 첫 분해
- 새 epic 추가로 신규 feature 후보 등장
- 기존 feature가 너무 커서 sub-feature로 재분해 필요
- agent가 구현 시작 전 명확한 unit 필요
- feature registry SaaS에 등록 전 표준 명세 필요
- pivot 후 PRD 변경에 따른 feature 재구성

## 3. 입력 (Inputs)

### 필수
- `define-product-spec` 출력 PRD
- feature_candidates 목록 (PRD §10)
- 기존 feature registry (있으면 — query-feature-registry로 조회)

### 선택
- 기술 스택 결정 (interface 형식 영향)
- 기존 코드베이스 구조 (의존성 매핑)
- 우선순위 stakeholder

### 입력 부족 시 forcing question
- "이 candidate는 user story 어느 것을 충족해? 1:1, N:1, 1:N 매핑?"
- "schema → API → UI 관통 가능해? 한 layer에만 머물면 vertical slice 아님."
- "다른 제품에 이식할 만한 응집도야 아니면 product-specific task야?"
- "이 feature가 너무 크지 않아? 2주+ 걸리면 sub-feature로 분해."

## 4. 핵심 원칙 (Principles)

1. **Vertical slice 강제** — DB schema + API + UI + test 관통. horizontal layer 분할 금지.
2. **Feature ≠ task** — 사용자 가치 단위. "DB 스키마 추가"는 task. "이메일 로그인"은 feature.
3. **재사용 가능 응집도** — 다른 제품에 patch / cherry-pick으로 이식 가능 수준.
4. **Acceptance criteria 자동 검증 우선** — manual QA 가능하지만 자동화 가능한 것 우선.
5. **Feature registry 먼저 조회** — `query-feature-registry`로 유사 feature 있는지. reuse > new.
6. **feature_id는 안정적** — 한 번 부여하면 변경 금지. 폐기 시 deprecated 마킹.
7. **너무 크면 분해, 너무 작으면 task로 강등** — 2주+ 작업은 sub-feature. 1일 미만은 task.
8. **Dependency graph 명시** — A가 B를 막는다는 정보. 병렬 가능 여부 판단.

## 5. 단계 (Phases)

### Phase 1. PRD 분석
1. user story 목록 추출
2. acceptance criteria 정리
3. functional / non-functional requirements 매핑
4. out-of-scope 명시 (이번 분해에서 제외)

### Phase 2. Reuse 검색
각 feature_candidate에 대해:
- `query-feature-registry`로 유사 feature 검색
- 결과 분류: `reuse` | `adapt` | `inspired-by` | `new`
- reuse 가능하면 patch + apply recipe 가져오기

### Phase 3. Vertical Slice 식별
candidate를 vertical slice로 변환:
- 사용자 진입점 (entry: route / button / API call)
- 처리 로직 (logic: validation / business rule / external call)
- 데이터 모델 (data: table / column / relation)
- 출력 (output: response / UI render / event emit)
- 검증 (test: unit / integration / e2e)

5개 layer 중 1-2개만 있으면 vertical slice 아님 → 다른 feature와 합치거나 task로.

### Phase 4. Feature Spec 작성
각 feature에 표준 명세 필드:
- `feature_id`: `<domain>.<action-target>` 또는 FEAT-<num>
- `name`, `summary`, `problem`
- `scope`, `out_of_scope`
- `reuse_intent`: 어떤 도메인에서 재사용 가능
- `interfaces`: API / UI / data / event
- `dependencies`: 선행 feature, library, 인프라
- `acceptance_criteria`: done 판단 기준
- `test_plan`: unit / integration / e2e
- `implementation_notes`: trade-off, migration
- `status`: draft → candidate → ready → in-progress → implemented → verified → reusable

### Phase 5. Dependency Graph
1. feature 간 dependency 명시
2. 순환 dependency 검출 → 분해 또는 추상화
3. 병렬 가능 group 식별
4. critical path 식별

### Phase 6. Priority & Sequencing
- P0: 출시 필수 (MVP 정의)
- P1: 1차 출시 후 1-2주 내
- P2: 추후
- P3: nice-to-have

### Phase 7. Hand-off
- `review-architecture`: 구조 리스크 검토
- `triage-work-items`: 상태 lifecycle 운영
- `build-with-tdd`: 구현 시작

## 6. 출력 템플릿 (Output Format)

```yaml
feature_map:
  total: 12
  by_priority: { P0: 5, P1: 4, P2: 3 }
  by_status: { draft: 12, ready: 0 }
  critical_path: [FEAT-001, FEAT-003, FEAT-007]

features:
  - feature_id: auth.email-password-login
    feature_num: FEAT-001
    name: Email Password Login
    summary: 사용자가 이메일과 비밀번호로 로그인
    problem: 사용자가 계정 기반 서비스에 다시 접근
    scope:
      - 로그인 form
      - 로그인 API
      - session 생성
    out_of_scope:
      - social login
      - MFA
    reuse_intent:
      - B2C SaaS
      - admin app
    interfaces:
      api: ["POST /auth/login"]
      ui: ["login form", "error toast"]
      data: ["users.email", "sessions.user_id"]
      events: ["user.logged_in"]
    dependencies:
      - feature: auth.user-table  # 선행 feature
      - lib: bcrypt
      - infra: postgres
    acceptance_criteria:
      - "올바른 email/password로 로그인 → session 생성"
      - "잘못된 password → 401 + generic error (계정 노출 금지)"
      - "rate limit 초과 → 429 + retry-after header"
    test_plan:
      unit: ["password verification"]
      integration: ["login API"]
      e2e: ["login form success/failure"]
    implementation_notes:
      - "password hash algorithm 교체 가능 (bcrypt → argon2 migration path)"
    status: draft
    priority: P0
    estimated_size: 3-day
    owner: tbd
    related_user_stories: [US-1, US-2]
    reuse_candidate_score: 0.85  # query-feature-registry 결과
    reuse_decision: new  # reuse | adapt | inspired-by | new

dependency_graph:
  - from: FEAT-001
    to: FEAT-003
    type: blocks
  - from: FEAT-002
    to: FEAT-005
    type: prerequisite

parallel_groups:
  - [FEAT-001, FEAT-002]  # 병렬 가능
  - [FEAT-004, FEAT-006]
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `define-product-spec` — `Skill` tool로 invoke
- 페어: `query-feature-registry` (있으면) — reuse 검색
- 다음 단계: `review-architecture` — 구조 리스크 검토
- 다음 단계: `triage-work-items` — 상태 lifecycle 운영
- 다음 단계: `build-with-tdd` — 구현 시작

## 8. Anti-patterns

1. **Horizontal layer 분할** — "FEAT-1: DB schema, FEAT-2: API, FEAT-3: UI" — 사용자 가치 미완결. vertical slice 위반.
2. **Task를 feature로** — "버튼 색상 변경"은 task. feature는 사용자 가치 단위.
3. **Acceptance criteria 없는 feature** — done 판단 불가. 자동 검증 가능 우선.
4. **Reuse 검색 skip** — 매번 처음부터. registry 먼저 조회 강제.
5. **Dependency graph 미작성** — 병렬 가능 누락 + 순환 dependency 발견 늦어짐.
6. **너무 큰 feature (epic)** — 2주+ 작업은 sub-feature. integration 위험 증가.
7. **너무 작은 feature** — 1일 미만은 task로. feature registry 오염.
8. **feature_id 변경** — 이름 변경, 폐기, merge 모두 deprecated 마킹 + new id. 기존 ID 보존.
