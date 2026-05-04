---
name: concretize-idea
description: This skill should be used when the user wants to "develop an idea", "build something new", "validate a startup idea", "turn an idea into a product", "concretize an idea", or starts with "I want to build X". Orchestrates §1 Idea & Business Validation phase — runs idea validation, business viability, competition analysis, customer segmentation, and PRD generation in sequence.
---

# concretize-idea — §1 Idea & Business Validation Orchestrator

§1 라이프사이클 단계의 진입점. idea/concept → PRD draft + business viability report 를 생성하는 멀티-stage 파이프라인.

**진입 조건**: idea 또는 concept만 존재. 코드베이스 미존재 또는 상용 빌딩 시작 전.
**산출물**: PRD draft, business viability report, market position summary.
**다음 phase**: PRD 확정 후 → `define-features` (§2).

---

## Stage 흐름

```
concretize-idea (§1 phase orchestrator)
├── stage 1: validate-idea          (idea stress-test — 6 forcing questions)
├── stage 2: validate-advanced-edge-idea  (edge case / hidden assumption grilling)
├── stage 3: assess-business-viability  (7차원 사업성 평가)
├── stage 4: [analyze-competition-and-substitutes]  🆕 경쟁/대체재 매트릭스
├── stage 5: review-pricing-and-gtm  (pricing model + GTM channel 평가)
├── stage 6: [map-customer-segments]  🆕 고객 세그먼트 + 구매자 분리
├── stage 7: define-product-spec    (PRD draft 생성)
└── stage 8: autoplan               (PRD 4-mode review — cross-phase sub-orchestrator)
        └── invokes review-scope / review-engineering / review-design / review-devex
```

> 🆕 = 신규 작성 필요 skill. 현재는 해당 단계를 orchestrator가 직접 수행.

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

### Stage 3-5: Business Validation

`assess-business-viability` skill을 invoke해 TAM/SAM/SOM, 고객-구매자 분리, willingness-to-pay, GTM, 경쟁, unit economics, 규제 7차원을 평가한다.

경쟁 분석이 필요하면 경쟁/대체재 매트릭스를 직접 작성한다 (`analyze-competition-and-substitutes` skill 미존재 시 orchestrator가 수행):
- Direct competitors, indirect competitors, substitutes 3분류
- 각 항목별 price, target user, key differentiator, market share(추정) 표

`review-pricing-and-gtm` skill을 invoke해 pricing model 설계와 GTM channel 전략을 평가한다.

**Stage 3 gate**: 사업성 평가 결과 치명적 결함(willingness-to-pay 없음, regulatory block 등)이 발견되면 사용자에게 보고하고 계속 여부를 묻는다.

### Stage 6: Customer Segmentation

고객 세그먼트를 식별한다 (`map-customer-segments` skill 미존재 시 orchestrator가 수행):
- Primary user vs Buyer 분리 (B2B의 경우 특히 중요)
- Early adopter 프로필 (demographics, pain intensity, current solution)
- Secondary segment 2-3개

### Stage 7: PRD Generation

`define-product-spec` skill을 invoke해 공식 PRD를 생성한다.

PRD 필수 포함 항목:
- Problem Statement + Solution
- Target user + Buyer (stage 6 결과 반영)
- Core features (3-5개, MoSCoW 분류)
- Out of scope
- Success metrics (user behavior + business + quality + operations)
- Risk register (technical / business / legal)

### Stage 8: PRD Review

`autoplan` (cross-phase review sub-orchestrator)을 invoke해 PRD draft를 4-mode review한다:
- `review-scope` — 범위 형성 / 결정
- `review-engineering` — 기술 구현 가능성
- `review-design` — 디자인 차원 (0-10 score)
- `review-devex` — developer-facing product이면 DX 검토

---

## 산출물 형식

```markdown
## §1 산출물 — {idea 이름}

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
- `/buddy:define-features` — §2 Feature Definition & Backlog (권장)
- `/buddy:autoplan` — PRD 재검토가 필요하면 standalone으로 추가 review

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §§1, §2.1
- autoplan 위치 설명: 동 문서 §2.1
