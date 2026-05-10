# Map Customer Segments — 사용자 vs 구매자 분리 + persona

## 1. 목적

PRD 의 "target user" 가 종종 *과도하게 단순화* (단일 persona) 된다. 이 skill 은 **사용자 (user) vs 구매자 (buyer)** 를 명시 분리하고, early adopter persona 와 secondary segment 를 식별한다.

B2B 에서는 *user ≠ buyer* 가 거의 항상 (예: dev 가 user, eng manager 가 buyer). B2C 에서도 부모 vs 자녀, 본인 vs 선물 받는 사람 등 분리.

## 2. 사용 시점

- `validate-idea` 통과 후 PRD 작성 (`define-product-spec`) 직전
- 가격 / GTM 채널 결정 전 — buyer 가 user 와 다른 경우 채널 분리
- pricing 검토 시 — willingness-to-pay 는 *buyer* 의 신호
- 마케팅 메시지 분리 — user 향 vs buyer 향 메시지 다름
- onboarding 흐름 설계 — buyer signs up, user uses

## 3. 입력

### 필수
- PRD 또는 idea 의 target user 가설
- 시장 segmentation 가설 (`analyze-market-size` 산출 활용 가능)

### 선택
- 인터뷰 데이터 (`conduct-customer-interview` 산출 활용)
- 경쟁사 customer base 정보

## 4. Stage 흐름

### Stage 1: User vs Buyer 분리

질문 매트릭스:

| 질문 | 답변 → segment |
|------|------------|
| 누가 *직접 사용* 하나? | user |
| 누가 *돈을 내나*? | buyer |
| 둘이 같은 사람인가? | B2C 개인 도구 — 가능 |
| 같지 않으면 어떻게 다른가? | role / 권한 / 동기 명시 |

→ 분리 결과: user persona + buyer persona (서로 다를 때).

### Stage 2: Early adopter persona

*첫 진입 사용자* 프로필. innovator / early adopter (Geoffrey Moore — Crossing the Chasm) 의 5 차원:

| 차원 | 측정 |
|------|------|
| demographics | 직군 / 회사 규모 / 지역 / age |
| pain intensity | 1~5 점 (5 = 매일 손해) |
| current solution | manual / spreadsheet / 경쟁사 |
| switching cost | 낮음 / 중 / 높음 |
| reachability | 어디에서 만나나 (slack / linkedin / conference) |

### Stage 3: Secondary segment 식별

primary segment 외 2~3 개 후보 segment. 각 segment 의:
- 시장 크기 (sub-SAM)
- 진입 시점 (year 1 / 2 / 3)
- value proposition 변형 (primary 와 메시지 다른 부분)

### Stage 4: Anti-persona 명시

*명시적으로 타겟 아닌* persona — scope 를 좁히는 효과.

## 5. 산출물 형식

```markdown
## Customer Segments — {제품 이름}

### User vs Buyer
- **User**: {persona}
- **Buyer**: {persona — same / different}
- 같지 않으면 분리 메시지 / 채널 / 가격 검토 필요

### Early Adopter Persona
| 차원 | 값 |
|------|-----|

### Secondary Segments (2~3)
| segment | sub-SAM | 진입 시점 | 메시지 변형 |
|---------|--------|---------|---------|

### Anti-persona
- 타겟 아닌 persona: ... — 이유: ...
```

## 6. 검증

- [ ] User 와 Buyer 가 *명시적으로 분리* 되는가? (B2B / 가족 / 선물 케이스)
- [ ] Early adopter 5 차원 모두 채워졌는가?
- [ ] Secondary segment 2~3 식별?
- [ ] Anti-persona 1+ 명시?
- [ ] 인터뷰 데이터 또는 surrogate (경쟁사 customer 정보) 로 *evidence* 보강?

## 7. 다음 phase

- `define-product-spec` 의 target user 섹션 입력
- `review-pricing-and-gtm` 의 *buyer 향 가격* + *user 향 channel* 입력
- `analyze-competition-and-substitutes` 의 *segment 별 경쟁사* 입력

## 8. 참조

- Strategyzer Value Proposition Canvas — segment 정의
- Crossing the Chasm (Geoffrey Moore) — early adopter
- marketingskills/customer-research (외부 reference, MIT)
