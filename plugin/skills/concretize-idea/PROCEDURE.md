# concretize-idea — 1단계 Idea & Business Validation Orchestrator

1단계 라이프사이클 단계의 진입점. idea/concept → PRD draft + business viability report 를 생성하는 멀티-stage 파이프라인.

**진입 조건**: idea 또는 concept만 존재. 코드베이스 미존재 또는 상용 빌딩 시작 전.
**산출물**: PRD draft, business viability report, market position summary.
**다음 phase**: PRD 확정 후 → `define-features` (2단계).

---

## Stage 흐름 (재구조화 2026-05-29 — customer/use case 위치 정정 + HLD 추가)

```
concretize-idea (1단계 phase orchestrator)
├── stage 1: validate-idea          (idea stress-test — 6 forcing questions)
├── stage 2: validate-advanced-edge-idea  (edge case / hidden assumption grilling)
├── stage 3: map-customer-segments  (사용자/구매자 분리 — 후속 단계의 입력)
├── stage 4: analyze-competition-and-substitutes  (경쟁/대체재 — target customer 기준)
├── stage 5: assess-business-viability  (TAM/SAM/SOM — customer 정의 기반)
├── stage 6: review-pricing-and-gtm  (pricing/GTM — customer의 WTP/채널 기반)
├── stage 7: define-product-spec    (PRD draft — actors + use cases(logical) 포함)
│       └── invokes identify-actors + map-actor-use-cases (logical)
├── stage 8: write-hld              (High Level Design — product 구성 + tech stack + use case → product mapping)
└── stage 9: autoplan               (PRD + HLD 통합 4-mode review)
        └── invokes review-scope / review-design / review-devex / review-engineering
```

**중요 변경 (2026-05-29)**:
- `map-customer-segments`를 stage 6 → stage 3으로 이동 (후속 3개 단계가 customer 정의를 입력으로 요구)
- `assess-business-viability` 위치를 stage 3 → stage 5로 이동 (customer + 경쟁 분석 후 사업성 평가)
- `write-hld` 신규 stage 8 추가 — review-design / review-devex / review-engineering의 검증 대상 산출물 생산
- stage 7 `define-product-spec` 내부에서 `identify-actors` + `map-actor-use-cases` 호출 (logical 수준)

---

## 실행 절차

### Phase 1 전제 확인

시작 전 사용자에게 확인:
1. "어떤 아이디어를 구체화하려 하나요? 한 문장으로 설명해 주세요."
2. "현재 보유한 자료(리서치, 경쟁사 분석, 고객 인터뷰 등)가 있으면 공유해 주세요."

### Stage 1-2: Idea Validation

`validate-idea` skill을 invoke해 6 forcing question으로 아이디어를 stress-test한다.
사용자 응답에서 모호성이 발견되면 `validate-advanced-edge-idea`로 escalate해 edge case와 hidden assumption을 그릴링한다.

**Stage 1 gate**: idea가 충분히 명확하지 않으면 stage 3으로 이동하지 않는다. 사용자와 함께 모호성을 해소한 뒤 진행.

### Stage 3: Customer Segmentation (먼저)

`map-customer-segments` skill을 invoke해 고객 세그먼트를 식별한다:
- Primary user vs Buyer 분리 (B2B의 경우 특히 중요)
- Early adopter 프로필 (demographics, pain intensity, current solution)
- Secondary segment 2-3개

**왜 먼저인가**: stage 4-6 (경쟁/사업성/pricing)이 모두 "고객이 누구"인지를 입력으로 요구한다.

### Stage 4: Competition Analysis

`analyze-competition-and-substitutes` skill을 invoke해 stage 3에서 정의된 target customer 기준으로 경쟁/대체재 매트릭스를 작성한다:
- Direct competitors, indirect competitors, substitutes 3분류
- 각 항목별 price, target user, key differentiator, market share(추정) 표

### Stage 5: Business Viability

`assess-business-viability` skill을 invoke해 TAM/SAM/SOM, willingness-to-pay, GTM, unit economics, 규제 7차원을 평가한다.

**Stage 5 gate**: 사업성 평가 결과 치명적 결함(willingness-to-pay 없음, regulatory block 등)이 발견되면 사용자에게 보고하고 계속 여부를 묻는다. **비상업 프로젝트(학습/해커톤/OSS/hobby)는 본 stage skip 가능** — 사용자가 명시.

### Stage 6: Pricing + GTM (상업 프로젝트만)

`review-pricing-and-gtm` skill을 invoke해 pricing model 설계와 GTM channel 전략을 평가한다.

비상업 프로젝트는 skip.

### Stage 7: PRD Generation (Actors + Use Cases 포함)

`define-product-spec` skill을 invoke해 공식 PRD를 생성한다. 내부에서 `identify-actors` + `map-actor-use-cases`(logical)를 호출하여 actors와 use cases를 PRD에 명시.

PRD 필수 포함 항목:
- Problem Statement + Solution
- Target user + Buyer (stage 3 결과 반영)
- **Actors** (identify-actors 산출)
- **Use Cases (logical, with data flow)** (map-actor-use-cases 산출 — actors_involved + interactions + service_provided)
- User Stories (use case의 narrative 버전)
- Core features (3-5개, MoSCoW 분류)
- Out of scope
- Success metrics (user behavior + business + quality + operations)
- Risk register (technical / business / legal)

### Stage 8: High Level Design (HLD)

`write-hld` skill을 invoke해 PRD를 받아 High Level Design 문서를 작성한다.

HLD 필수 포함 항목 (9 섹션):
1. Product Decomposition (frontend/app/backend/DB/infra/SDK)
2. Per-product Role
3. Per-product Tech Stack (language + framework + key library)
4. Inter-product Communication (REST/gRPC/IPC/event 등)
5. **Use Case → Product Mapping** (PRD의 use case 데이터 흐름을 product에 매핑)
6. External Integration
7. Deployment Topology
8. Licensing
9. Distribution Model

**왜 필요한가**: stage 9의 review-design / review-devex / review-engineering이 검증할 산출물.

### Stage 9: PRD + HLD Review

`autoplan` (cross-phase review sub-orchestrator)을 invoke해 PRD + HLD를 통합 4-mode review한다:
- `review-scope` — PRD의 problem/use cases/scope 검증
- `review-design` — HLD의 product 구성 + use case → product mapping 검증
- `review-devex` — HLD의 SDK/CLI/API surface가 있으면 그 surface 검증
- `review-engineering` — HLD의 per-product tech stack + communication patterns 검증

---

## 산출물 형식

```markdown
## 1단계 산출물 — {idea 이름}

### Idea Validation Summary
- Core hypothesis: ...
- Validated assumptions: ...
- Open questions: ...

### Business Viability
- Market: TAM $X / SAM $X / SOM $X (추정 근거 포함)
- Customer: [Primary user] vs [Buyer]
- WTP: $X/month 또는 $X/feature (evidence)
- GTM: [channel 1], [channel 2]
- Key risks: ...

### PRD Draft
[define-product-spec 산출물]

### autoplan Review Summary
[autoplan 4-mode review 결과]
```

---

## User Gate

이 orchestrator는 2개 gate에서 사용자 확인을 요구한다:
1. **Stage 1 gate**: idea 명확성 부족 시 — 계속 여부
2. **Stage 3 gate**: 사업성 치명 결함 발견 시 — 피벗 여부

Gate 없이 자동 진행하지 않는다.

---

## 다음 phase

PRD 확정 후:
- `/buddy:define-features` — 2단계 Feature Definition & Backlog (권장)
- `/buddy:autoplan` — PRD 재검토가 필요하면 standalone으로 추가 review

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md` §§1, §2.1
- autoplan 위치 설명: 동 문서 §2.1
