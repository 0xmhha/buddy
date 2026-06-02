# §1 Discovery / Impact Analysis (문제·기회 검증 / 변경 영향 분석) — Stage Skills

- **Orchestrator (entry)**: `concretize-idea` (Mode A, 신규) / `assess-product-change` (Mode B, 기존 제품)
- **DoR → DoD**: idea·concept → 검증된 PRD + HLD (Mode A) · change request + 기존 codebase → 영향 평가 + scope 분류 (Mode B)
- 정체성 원본: `engineering-phases.md` §2 Phase 1 | 명사 원본: `se-lifecycle-naming.md` §1

## Stage skills

| Skill | Trigger | When to use | Not when (→ 대안) |
|---|---|---|---|
| `assess-product-change` | dispatch-only (via `/buddy:start`) | 기존 프로덕트 변경 평가 — 버그/기능/개선 무관 change trigger 접수 → 영향 평가 → scope 분류(small/medium/large) → 다음 phase routing. "있는 것을 바꾼다" | 코드베이스 없는 신규 아이디어 → `concretize-idea`(Mode A) |
| `validate-idea` | command + dispatch | YC 스타일 6 forcing question으로 product idea를 stress-test | 이미 검증됨·사업 성립 여부 판단 → `assess-business-viability` |
| `validate-advanced-edge-idea` | command + dispatch | validate-idea 통과 후 edge case·hidden assumption·second-order effect를 grilling으로 박멸 | 1차 검증 전 → `validate-idea` 먼저 |
| `assess-business-viability` | command + dispatch | 아이디어가 사업으로 성립하는지 7차원(TAM/SAM/SOM·고객-구매자·WTP·GTM·경쟁·unit economics·규제) 평가 | 시장 규모만 → `analyze-market-size` / 고객 정의만 → `map-customer-segments` |
| `analyze-market-size` | command + dispatch | TAM/SAM/SOM 산출 — top-down + bottom-up cross-check + ±50% sensitivity | 사업성 전반 → `assess-business-viability` |
| `map-customer-segments` | command + dispatch | 사용자 vs 구매자 분리 + early adopter 5차원 persona + anti-persona | job 중심 분석 → `map-jobs-to-be-done` |
| `map-jobs-to-be-done` | command + dispatch | JTBD — functional/emotional/social job + Ulwick job map 8단계 + outcome statement | segment·persona 분리 → `map-customer-segments` |
| `conduct-customer-interview` | command + dispatch | 인터뷰 스크립트 + Mom Test anti-pattern 회피 + 결과 코딩 + 가설 4-라벨 update | 정량 시장 추정 → `analyze-market-size` |
| `analyze-competition-and-substitutes` | command + dispatch | 경쟁/대체재 4분류 × 4차원 매트릭스 + positioning(Moore) + moat(7 Powers) | pricing·진입 채널 → `review-pricing-and-gtm` |
| `decide-target-market` | command + dispatch | target market 결정(글로벌/단일/다지역) + region cluster trigger | 경쟁 구도 분석 → `analyze-competition-and-substitutes` |
| `review-pricing-and-gtm` | dispatch | pricing model + GTM(Go-To-Market) channel 전략 평가 (상업 프로젝트) | 비상업 프로젝트면 skip |
| `define-product-spec` | command + dispatch | 검증·사업성 결과를 공식 PRD(Product Requirements Document)로 고정 — actors + use cases(logical) | 기술 구성·스택 결정(HLD) → `write-hld` |
| `write-hld` | command + dispatch | PRD 완료 후 High Level Design — product 분해 + tech stack + use case→product mapping | PRD 미작성 → `define-product-spec` 먼저 / 상세 컴포넌트 설계 → §3 |

## Disambiguation (노드 내)
- `validate-idea`(아이디어가 말이 되나) vs `assess-business-viability`(사업으로 돈이 되나): 검증 단계 차이.
- `define-product-spec`(무엇을 만드나 — logical) vs `write-hld`(어떻게 구성하나 — 기술 구조).

## Escalation
- 노드 내 2개+ 모호 → `../routing-rules.md` §3. Mode A/B 분기 자체가 모호 → `routing-rules.md` §3 케이스 A1/A2.
- Mode A는 `concretize-idea`가 위 stage를 9단계로 순서 호출 — 단독 stage 명시가 없으면 orchestrator 우선.
