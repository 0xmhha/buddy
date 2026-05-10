# Analyze Competition and Substitutes — 경쟁 / 대체재 매트릭스 + positioning + moat

## 1. 목적

PRD 의 *경쟁* 차원을 4 분류 (direct / indirect / substitute / non-consumption) 로 명시 분석한다. 각 경쟁 entity 의 *price × target × differentiator × market share* 4 차원을 매트릭스로 비교하고, **positioning statement** + **moat 후보** 를 산출한다.

`assess-business-viability` 의 *경쟁 차원* 입력. `map-jobs-to-be-done` 의 *competitive job analysis* 와 cascade.

## 2. 사용 시점

- `validate-idea` 통과 후 사업성 평가 (`assess-business-viability`) 직전 또는 안에서
- 마케팅 카피 (`draft-marketing-copy`) 작성 전 — positioning 입력
- 경쟁사 신규 등장 시 — 매트릭스 갱신
- pivot 결정 시점 — 경쟁 영역 재정의
- 가격 결정 시 — 경쟁 가격 분포 입력
- 투자 자료 작성 — moat 명시 필요

## 3. 입력

### 필수
- PRD 또는 idea (problem statement + value hypothesis)
- 후보 경쟁사 URL list (있으면 — 또는 *발견 자체* 가 본 skill 의 부산물)
- `map-customer-segments` 산출 (early adopter persona — 같은 customer 두고 경쟁)
- `map-jobs-to-be-done` 산출 (같은 job 다른 해결책)

### 선택
- 경쟁사 매출 / 사용자 수 / 펀딩 (top-down cross-check)
- 사용자 발화 — 어떤 경쟁사 자주 언급되나

## 4. Stage 흐름

### Stage 1: 4 분류

| 분류 | 정의 | 예시 (가계부 앱) |
|------|-----|--------------|
| **Direct competitor** | 같은 customer + 같은 job + 같은 형태 | YNAB, Mint |
| **Indirect competitor** | 같은 customer + 같은 job + 다른 형태 | Excel template, 가계부 책 |
| **Substitute** | 다른 customer / job / 형태 — 같은 outcome | 은행 앱의 자동 분류 기능 |
| **Non-consumption (status quo)** | *아무것도 안 함* — 가장 강한 경쟁자 | 사용자가 가계부 안 적음 |

각 분류에 *최소 2~3 entity* 식별. *non-consumption* 명시 필수.

### Stage 2: 매트릭스 — 4 차원

| Entity | Price | Target | Differentiator | Share (추정) |
|--------|-------|--------|----------------|----------|
| Direct A | ... | ... | ... | ... |
| Direct B | ... | ... | ... | ... |
| Indirect | ... | ... | ... | ... |
| Substitute | ... | ... | ... | ... |
| Status quo | $0 | (전체) | "그냥 안 함" | 가장 큼 |

데이터 source: 자동 web scraping (외부 도구 — tavily-mcp / marketingskills/competitor-profiling 패턴 reference) + 사용자 인터뷰 발화.

### Stage 3: Positioning statement

다음 양식 채움 (Geoffrey Moore 의 elevator pitch):

> "For {target customer} who {pain / job}, our {product category} is the {key differentiator} that {primary benefit}. Unlike {direct competitor} or {substitute}, we {unique value}."

### Stage 4: Moat 후보 식별

지속 가능한 경쟁 우위 5 후보 (Hamilton Helmer — 7 Powers):

| moat 종류 | 우리 가능성 | 근거 |
|---------|---------|-----|
| Counter-positioning | ... | 기존 경쟁사가 따라하면 cannibalize |
| Switching cost | ... | 사용자 lock-in |
| Network effect | ... | 사용자 늘면 가치 증가 |
| Scale economy | ... | 비용 우위 |
| Branded moat | ... | 신뢰 / 정체성 |

→ 최소 1 moat 가 *실제 가능* 해야 (전부 *불가* 면 사업성 위험 신호).

### Stage 5: comparison page 작성 (선택, dispatch out)

Direct competitor 가 명확하면 *vs page* 작성 → marketingskills/competitor-alternatives 패턴 활용 (4 format: singular / plural / you-vs-competitor / competitor-vs-competitor).

→ 본 skill 산출 후 `draft-marketing-copy` skill cascade 가능.

## 5. 산출물 형식

```markdown
## Competition Analysis — {제품 이름}

### 4 분류
| 분류 | entity 1 | entity 2 |
|------|---------|---------|

### 매트릭스 (price × target × differentiator × share)
| Entity | Price | Target | Differentiator | Share |
|--------|-------|--------|----------------|-------|

### Positioning statement
"For ..., our ..."

### Moat 후보
| moat 종류 | 가능성 | 근거 |

### Comparison page (선택)
- vs {direct competitor}: {link / draft}
```

## 6. 검증

- [ ] 4 분류 모두 식별 (특히 *non-consumption* 누락 금지)?
- [ ] 매트릭스에 ≥4 entity (각 분류 1+)?
- [ ] Positioning statement 가 specific (target / differentiator / unlike-부분 명시)?
- [ ] Moat ≥1 *실제 가능* 후보?
- [ ] 데이터 source 명시 (URL / 인터뷰 / 통계)?

## 7. 다음 phase

- `assess-business-viability` *경쟁 차원* 입력
- `define-product-spec` 의 *value proposition* refinement
- `draft-marketing-copy` 의 *vs page* / *positioning copy* 입력
- `review-pricing-and-gtm` 의 *경쟁 가격 분포* 입력

## 8. 참조

- Crossing the Chasm (Geoffrey Moore) — positioning statement 양식
- 7 Powers (Hamilton Helmer) — moat 분류
- marketingskills/competitor-alternatives + competitor-profiling (외부 reference, MIT) — vs page format + URL → profile 패턴
- tavily-mcp (검색 도구)
