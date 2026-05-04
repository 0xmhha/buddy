---
name: query-feature-registry
description: "PRD 또는 feature candidate를 받아 feature-management-saas-mcp registry에서 유사 feature를 검색해 reuse / adapt / inspired-by / new 판단. semantic + interface + stack + license + security gate. 트리거: '비슷한 feature 있어?' / 'reuse 가능해?' / 'feature registry 검색' / '유사 구현 찾아줘' / 'patch 가져와' / 'feature 재사용' / 'registry query'. 입력: feature candidate (problem, interfaces, stack, constraints) + target product context. 출력: matched features + reuse_decision + patch artifact + apply recipe. 흐름: split-work-into-features → query-feature-registry → build-with-tdd."
type: skill
---

# Query Feature Registry — Reuse 검색 + 판단 루프

## 1. 목적

`split-work-into-features` 출력 feature candidate를 **feature registry SaaS / MCP에 조회**해 유사 feature가 있으면 reuse / adapt / inspired-by / new 판단을 내린다.

핵심 가치: **0부터 만들지 말고 검색 먼저**. autoplan의 "Search Before Building" ethos를 feature 단위로 강제.

이 스킬은 `feature-management-saas-mcp.md`의 `feature.query` MCP tool과 직결된다. 동일 입력 형식 + 동일 결과 등급(0.85+ reuse / 0.70-0.84 adapt / 0.50-0.69 inspired-by / <0.50 new) 사용.

## 2. 사용 시점 (When to invoke)

- `split-work-into-features` feature candidate set 생성 직후
- 신규 epic 추가 시 (재사용 가능 feature 우선 탐색)
- "비슷한 거 만든 적 있나?" 질문 시
- 외부 feature catalog (GitHub, npm, marketplace) 탐색 전 internal 검색
- agent가 구현 시작 전 patch가 있으면 시간 단축
- 기술 스택 변경 시 (variant 검색)

## 3. 입력 (Inputs)

### 필수
- feature candidate 1개+
  - problem statement (한 문장)
  - desired flow (단계별)
  - interfaces (API / UI / data shape)
  - target stack (frontend / backend / database)
- intended use (commercial / internal / OSS)
- constraints (license, security, customization)

### 선택
- 기존 feature_id (특정 검색)
- 도메인 tag (auth / billing / analytics)
- 경쟁사 / vendor 후보

### 입력 부족 시 forcing question
- "이 feature의 핵심 user flow가 뭐야? 막연한 description은 검색 품질 떨어져."
- "target stack이 뭐야? 같은 feature도 stack 따라 patch 다름."
- "intended use가 commercial이면 paid / commercial license patch는 결제 필요."
- "modification 허용 가정해? 안 하면 license 후보 좁아짐."

## 4. 핵심 원칙 (Principles)

1. **Search Before Building** — autoplan ethos. 검색 0건이라도 검색은 했다는 audit trail.
2. **Score만으로 판단 금지** — license / security / visibility gate 통과해야 reuse 추천.
3. **Slug는 보조** — slug 일치는 일부 신호. semantic / flow / interface 유사도가 핵심.
4. **Stack compatibility는 hard filter** — Next.js stack에 Rails 구현 → 의미 약함.
5. **Patch applicability 별도 평가** — feature 유사해도 patch conflict 높으면 cost 큼.
6. **Adoption evidence 가중** — 다른 product 7개 사용 + 0 issue >> 0 product 사용.
7. **Customization compatibility** — 우리가 변경해야 할 부분이 license / 구조상 가능한지.
8. **Verdict는 4 카테고리** — reuse / adapt / inspired-by / new. ambiguity 강제 해소.

## 5. 단계 (Phases)

### Phase 1. Query Construction
feature candidate → MCP query 형식:
- query string (problem + intent)
- domain_tags
- desired_flow (구체 단계)
- target_stack
- interfaces (API path, UI surface, data shape)
- constraints (license, security severity)
- limit (5 default)

### Phase 2. Registry Search
`feature.query` MCP tool 또는 직접 vector + BM25 검색:
- semantic similarity (problem / summary / scope)
- interface similarity (API / UI / data)
- flow similarity (user flow / state transition)
- stack compatibility
- license / security gate

각 매치에 detailed score + mismatch reasons.

### Phase 3. Gate Filtering
검색 결과를 hard gate로 필터:
- license incompatible → exclude
- security scan failed → exclude
- visibility / entitlement 없음 → exclude
- patent grant 필요한데 없음 → exclude

### Phase 4. Detailed Comparison
top 후보 (3-5개)에 대해 비교:
- feature intent
- user flow
- interface shape
- implementation stack
- package dependency
- customization 가능성
- license / commercial policy
- security posture
- patch applicability
- adoption evidence

### Phase 5. Reuse Decision
4 카테고리 verdict:
- **reuse** (0.85+): acceptance criteria + interface 거의 동일, patch cost 낮음
- **adapt** (0.70-0.84): 핵심 동작 같으나 stack / schema / UI 차이 — variant 작성
- **inspired-by** (0.50-0.69): 컨셉 유사하나 구현 재사용 어려움 — 참고만
- **new** (<0.50): 유사 feature 없음 — 새로 작성

verdict는 점수 + gate 모두 통과해야.

### Phase 6. Patch Artifact 검증 (reuse / adapt 시)
- `patch.verify` MCP tool: hash / signature / security scan / license
- conflict risk 평가 (target repo vs source)
- prerequisites checklist
- apply recipe + verification commands

### Phase 7. Adoption Recording
- adoption record 등록 (`feature.record_adoption`)
- product_id + feature_id + implementation_id + status
- adaptation notes

## 6. 출력 템플릿 (Output Format)

```yaml
query_id: "<uuid>"
query_summary: "<feature candidate 요약>"
target_stack: { frontend: "Next.js", backend: "Supabase" }
intended_use: commercial

raw_matches:
  - feature_id: feat_123
    canonical_key: "u_7f3a/shopmate/FEAT-42"
    project_unique_slug: "auth.email-password-login"
    name: "Email Password Login"
    publisher: "shopmate"
    version: "1.1.0"
    similarity:
      overall: 0.91
      semantic: 0.94
      flow: 0.88
      interface: 0.90
      stack: 0.96
      license: 1.0
      security: 1.0
      patch_applicability: 0.82
    gates:
      license_compatible: yes
      security_passed: yes
      visibility_accessible: yes
    mismatch_reasons: []

filtered_matches:
  - feature_id: feat_123
    score: 0.91
    verdict: reuse

verdict_summary:
  reuse: ["feat_123"]
  adapt: []
  inspired_by: ["feat_456"]
  new: []
  recommended_action: "feat_123 reuse — patch download + apply recipe 실행"

patch_verification:
  feature_id: feat_123
  patch_artifact_id: patch_123
  hash:
    algorithm: sha256
    value: "..."
    verified: yes
  signature:
    status: valid
    signed_by: "author_123"
  security:
    status: passed
    high_severity_count: 0
  license:
    license: MIT
    commercial_use: allowed
  conflict_risk: low
  prerequisites:
    - "Next.js app router"
    - "Supabase project"
    - "users / sessions table"
  apply_recipe:
    steps:
      - "install packages: @supabase/supabase-js"
      - "apply patch"
      - "run migration"
      - "run tests"
    verification:
      - "npm test"
      - "npm run e2e -- auth-login"

adoption_record:
  product_id: prod_my_app
  feature_id: feat_123
  implementation_id: impl_123
  status: applying
  adaptation_notes: "UI copy + redirect path 변경 예정"

audit_trail:
  query_at: "<timestamp>"
  searched_repositories: 1
  queries_run: 3  # original + 2 expansions
  total_candidates_seen: 47
  filtered_after_gates: 5
  verdict_assigned: 4
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `split-work-into-features` — `Skill` tool로 invoke
- 페어: `apply-builder-ethos` — Search Before Building 원칙 강제
- 페어: `review-license-and-ip-risk` — patch license 검증
- 페어: `audit-security` — patch security scan 검증
- 다음 단계: `build-with-tdd` (verdict가 new) 또는 patch apply (reuse / adapt)
- 후속: `triage-work-items` — adopted feature lifecycle 등록

## 8. Anti-patterns

1. **Search 안 하고 직진** — autoplan ethos 위반. 0건이라도 audit trail.
2. **Slug 일치만으로 reuse** — 같은 slug 다른 project = 다른 feature. detailed comparison 강제.
3. **Score만으로 판단** — license / security gate 통과 필수.
4. **License compatible 가정** — MIT 표시만 보고 commercial OK? attribution 의무 있음.
5. **Patch verify 없이 download** — hash / signature / scan 통과 필수.
6. **Conflict risk 무시** — 유사도 높아도 patch conflict 많으면 새로 작성이 빠를 수 있음.
7. **Adoption record 누락** — 다른 product가 사용했는지 알 수 없음. registry 가치 약화.
8. **Customization 가능 가정** — license가 modification 금지면 reuse 불가.
