# Map Jobs To Be Done — JTBD 프레임워크 적용

## 1. 목적

*제품* 이 아니라 *고객의 job* 부터 출발한다. JTBD (Jobs To Be Done) 프레임워크는 "사용자가 *우리 제품을 고용해서* 어떤 *progress* 를 만들려 하는가?" 를 명시한다 (Clayton Christensen).

`map-customer-segments` 가 *누가 (who)* 라면 본 skill 은 *왜 (why) — 어떤 progress 를 위해* 다.

## 2. 사용 시점

- `map-customer-segments` 종료 후 PRD 작성 직전
- pivot 결정 시 — 같은 persona 의 *다른 job* 발견하면 pivot 후보
- 경쟁 정의 시 — 경쟁은 *같은 job 을 다른 방법* 으로 해결하는 것 (substitute 포함)
- 가격 결정 시 — *job 의 가치* 가 willingness-to-pay 의 ceiling
- feature 우선순위 — *job 진척에 직접 기여* 하는 feature 우선

## 3. 입력

### 필수
- `map-customer-segments` 산출 (early adopter persona)
- PRD 의 problem statement / value hypothesis
- 인터뷰 데이터 (있으면 — `conduct-customer-interview` 산출)

### 선택
- 경쟁사 / substitute 정보 (같은 job 의 다른 해결책)

## 4. Stage 흐름

### Stage 1: Functional / Emotional / Social Job 분리

JTBD 는 3 차원:

| 차원 | 정의 | 예시 (개인 가계부 앱) |
|------|-----|----------------|
| Functional | 객관적 task 완수 | "월별 지출을 카테고리별로 분류한다" |
| Emotional | 본인 감정 / 자기 인식 | "재정 통제하고 있다는 안심" |
| Social | 타인 인식 / 사회적 위치 | "가족에게 책임감 있는 모습" |

세 차원 모두 식별. *Functional 만* 가정하면 design 이 표면적임 (감정 / 사회 차원이 진짜 driver 인 경우 많음).

### Stage 2: Job map 작성

job 을 *시간순* 8 단계로 분해 (Anthony Ulwick 의 Universal Job Map):

1. Define — 목표 명시
2. Locate — 입력 / 자원 식별
3. Prepare — 환경 준비
4. Confirm — 준비 검증
5. Execute — 실행
6. Monitor — 진행 감시
7. Modify — 조정
8. Conclude — 종료

각 단계별:
- 현재 *어떻게* 해결하는가 (current solution)
- *불만족* 영역 (pain point)
- *가능 개선* 영역 (opportunity)

### Stage 3: Outcome statement

각 job 에 *측정 가능한 outcome* 명시:

> "Minimize the time it takes to {action} when {circumstance}"

예: "Minimize the time it takes to categorize expenses when reviewing monthly spending"

### Stage 4: Competitive job analysis

*같은 job 을 해결하는 다른 방법* 명시:

| 해결책 | 종류 | 사용자 평가 |
|------|----|----------|
| 우리 제품 | direct | 가설 |
| 경쟁사 A | direct | 시장 |
| 스프레드시트 | substitute | "충분히 동작" 가능성 |
| 안 함 (status quo) | non-consumption | 가장 강한 경쟁 |

→ *non-consumption* (사용자가 *아무것도 안 함*) 이 종종 가장 큰 경쟁자.

## 5. 산출물 형식

```markdown
## JTBD — {제품 이름} / {persona}

### Job statement
"When {circumstance}, I want to {motivation}, so I can {desired outcome}."

### 3 차원 분리
- Functional: {job}
- Emotional: {job}
- Social: {job}

### Job map (8 단계)
| 단계 | 현재 해결 | pain | opportunity |
|------|--------|-----|---------|

### Outcome statements (3~5)
- Minimize {time/effort/error} to {action} when {circumstance}
- Maximize {value} when {circumstance}

### Competitive analysis
| 해결책 | 종류 | 평가 |
```

## 6. 검증

- [ ] Functional + Emotional + Social 3 차원 모두 식별?
- [ ] Job map 8 단계 모두 채움 (또는 *해당 없음* 명시)?
- [ ] Outcome statement 3+ 작성? (측정 가능한 형태)
- [ ] *non-consumption* (status quo) 를 경쟁 해결책으로 포함?
- [ ] 경쟁사 + substitute 모두 검토?

## 7. 다음 phase

- `define-product-spec` 의 *core features* 우선순위 입력 — outcome 직접 기여 feature
- `analyze-competition-and-substitutes` 의 *substitute* 매트릭스 입력
- `review-pricing-and-gtm` 의 *job 가치 기준 pricing ceiling* 입력

## 8. 참조

- Competing Against Luck (Clayton Christensen) — JTBD 기초
- The Outcome-Driven Innovation (Anthony Ulwick) — Job map + outcome statement
- Strategyzer Value Proposition Canvas — JTBD 와 정합
