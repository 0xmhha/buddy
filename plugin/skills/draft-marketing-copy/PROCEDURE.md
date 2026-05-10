# Draft Marketing Copy — copywriting + editing + ad creative 통합

## 1. 목적

마케팅 콘텐츠의 *카피 작성 + 편집 + ad creative variant* 통합. **landing page / email / ad / social / vs page** 5 영역 + buyer voice 정합 + A/B test 친화 양식.

`map-customer-segments` (§1) 의 buyer voice + `analyze-competition-and-substitutes` (§1) 의 positioning + `analyze-customer-feedback-corpus` (§8) 의 verbatim quote 입력.

## 2. 사용 시점

- 신규 feature launch 직전 — landing page / email / ad
- 마케팅 캠페인 시작 전 — ad creative variant 작성
- A/B test 후 — 잘 동작한 카피 *변형 확장*
- 사용자 verbatim quote 발견 후 — 카피로 변환
- pricing tier 변경 — paywall / upgrade 카피

## 3. 입력

### 필수
- target audience (`map-customer-segments` 의 buyer / user)
- positioning statement (`analyze-competition-and-substitutes` 산출)
- 카피 유형 (landing / email / ad / social / vs page)

### 선택
- 사용자 verbatim quote (`analyze-customer-feedback-corpus` 산출 — 가장 강력한 input)
- 경쟁사 카피 (비교 baseline)
- brand voice 가이드 (있으면)

## 4. Stage 흐름

### Stage 1: 카피 유형별 양식

| 유형 | 길이 | 핵심 |
|------|-----|------|
| Landing page | hero + 3-5 sub-section | hero = 한 문장 value prop |
| Email subject | < 50 char | curiosity / specificity / urgency |
| Email body | 100~300 word | 1 CTA |
| Ad headline | 30~60 char | hook + value |
| Ad body | 50~150 char | benefit 명시 |
| Social post | platform 별 (Twitter 280 / LinkedIn 1300+) | platform native voice |
| vs page | 비교 표 + verbatim quote | "objective" tone |

### Stage 2: Headline 생성 — 5+ variant

ad / hero / email subject 는 *동일 메시지 5+ 변형*:

| 변형 종류 | 예 (가계부 앱) |
|---------|------------|
| Question | "왜 매달 돈이 어디 갔는지 모르나요?" |
| Number | "한 번 입력에 평균 3.7초" |
| Direct benefit | "매달 30분 안에 가계부 정리" |
| Social proof | "10만 명이 사용하는 가계부" |
| Pain-first | "가계부 적기 귀찮으셨죠?" |

→ A/B test 입력. winning variant 발견 후 *왜* 분석.

### Stage 3: Body 작성 — PAS / AIDA framework

| framework | 단계 |
|---------|------|
| **PAS** (Problem-Agitate-Solution) | 1. 문제 명시 / 2. 결과 강조 / 3. 해결 제시 |
| **AIDA** (Attention-Interest-Desire-Action) | 1. 주목 / 2. 흥미 / 3. 욕구 / 4. 행동 |
| **FAB** (Feature-Advantage-Benefit) | 1. feature / 2. 장점 / 3. 사용자 이익 |

→ 카피 유형 + 사용자 단계 (cold / warm / hot) 별 framework 선택.

### Stage 4: Editing pass

draft 후 *editing* :

| 검토 | action |
|------|------|
| Length cut | 30% 줄이기 (보통 너무 김) |
| Verb 강화 | "be" 동사 → action verb |
| Jargon 제거 | 도메인 외 사용자에게 통하는가 |
| CTA specificity | "click here" → "Start free trial" |
| Proof 추가 | 숫자 / 사례 / quote |
| Anti-cliche | "world-class" / "revolutionary" 등 진부 표현 회피 |

→ marketingskills/copy-editing 패턴 차용.

### Stage 5: Ad creative variation

ad 의 visual + copy 조합:
- 같은 copy × 다른 visual
- 같은 visual × 다른 copy
- 형식 (image / video / carousel)

→ marketingskills/ad-creative 패턴.

### Stage 6: Voice consistency

모든 카피의 *brand voice* 정합:

| dimension | 옵션 |
|---------|-----|
| Formality | casual / professional / luxurious |
| Energy | calm / energetic / urgent |
| Persona | mentor / friend / expert |

→ voice guide 가 *모든 카피 작성자* (in-house + outsourced) 공유.

## 5. 산출물 형식

```markdown
## Marketing Copy — {캠페인 / launch}

### 카피 유형
| 유형 | 양식 | 길이 |

### Headline 5+ variant
| variant | 예 | A/B test 후보 |

### Body
- framework: PAS / AIDA / FAB
- draft: ...

### Editing checklist
- length cut / verb / jargon / CTA / proof / cliche

### Ad creative
| variant | visual | copy |

### Voice guide
- formality / energy / persona
```

## 6. 검증

- [ ] 카피 유형 별 양식 적용?
- [ ] Headline 5+ variant?
- [ ] Body framework (PAS / AIDA / FAB) 명시?
- [ ] Editing 6 checklist 모두 적용?
- [ ] Voice consistency (formality / energy / persona) ?
- [ ] Verbatim quote 활용 (있을 시) ?

## 7. 다음 phase

- `design-ab-experiment` (구현됨) — copy variant A/B test
- `plan-marketing-channel` — 어디 게재할지
- `automate-marketing-content` — 콘텐츠 스케줄

## 8. 참조

- Made to Stick (Heath brothers) — sticky message principle
- Words That Work (Frank Luntz) — political copy 원칙
- marketingskills/copywriting + copy-editing + ad-creative (외부 reference, MIT)
- humanizer (외부 reference, MIT — Siqi Chen, AI text 자연화 inverse 검증)
