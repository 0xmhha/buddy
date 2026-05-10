# Plan Growth Experiment — ICE / RICE 우선순위 + 가설 → 실험 → 학습

## 1. 목적

growth team 의 *실험 backlog 운영* — idea → ICE/RICE score → 실험 설계 → 결과 → 학습 영속화 의 표준 cycle. **가설-기반 growth** (Sean Ellis — Hacking Growth) 적용.

`design-ab-experiment` (구현됨) 와 cascade — 본 skill 은 *backlog + 우선순위*, ab-experiment 는 *실험 설계 detail*.

## 2. 사용 시점

- §8 iterate-product 의 growth team 분기 / 월 sprint 시작
- product-market fit 확보 후 — *증폭* 단계
- acquisition channel 후보 검증
- retention / referral 개선 idea 평가
- pricing 변경 후 *retest* 실험

## 3. 입력

### 필수
- growth idea backlog (자유 형식)
- AARRR funnel 분석 (`optimize-conversion-funnel` 산출)
- 현재 baseline metric

### 선택
- 경쟁사 growth 사례 (`analyze-competition-and-substitutes` 산출)
- 사용자 인터뷰 / feedback corpus 의 idea source

## 4. Stage 흐름

### Stage 1: Idea capture

자유 형식 idea (slack / notion / spreadsheet) → *구조화 backlog* :

| 양식 | 내용 |
|------|------|
| Hypothesis | "If we do X, then Y will happen because Z" |
| Target metric | 영향 받을 metric (e.g. activation rate) |
| Source | 어디서 왔나 (feedback / 경쟁 / 인터뷰) |
| Owner | 추진 책임자 |

### Stage 2: ICE / RICE score

| framework | 산식 | 적용 |
|---------|-----|------|
| **ICE** | (Impact × Confidence × Ease) / 3 | 빠른 평가 (1~10 점) |
| **RICE** | (Reach × Impact × Confidence) / Effort | 더 정밀 (실 수치) |

ICE 빠른 분류 후 *상위 후보* 만 RICE 로 정밀 평가.

### Stage 3: 실험 설계 (ab-experiment cascade)

상위 우선순위 idea → `design-ab-experiment` 호출:
- statistical power (sample size 계산)
- 대조군 / 실험군
- 측정 metric (primary + guardrail)
- 실험 기간

### Stage 4: 실험 종료 후 학습 영속화

| 결과 | action |
|------|------|
| Win (통계 유의 + 실용 유의) | ship + 백로그 다음 후보 |
| Loss | revert + *왜 hypothesis 틀렸나* 분석 |
| Inconclusive | 더 큰 sample 재실행 또는 hypothesis 수정 |

→ 모든 결과 → `persist-learning-jsonl` (구현됨) 영속화. 다음 cycle 의 *경험 baseline*.

### Stage 5: Sprint cadence

| period | activity |
|--------|---------|
| Weekly | idea capture + ICE 분류 |
| Bi-weekly | RICE 정밀 평가 + sprint planning |
| Monthly | 실험 결과 retro + learning publish |
| Quarterly | growth strategy review + framework 갱신 |

### Stage 6: Anti-pattern 회피

| anti-pattern | 회피 |
|------------|-----|
| Idea inflation (모든 idea 평가) | ICE 1차 cutoff (e.g. 5점 미만 즉시 폐기) |
| Implementation 우선 | hypothesis + measurement 미정 시 실험 X |
| Vanity metric | guardrail metric 동시 측정 (Reach 만 측정 X) |
| Cherry-picking | 모든 결과 (win + loss) 영속화 |
| Loss 무시 | loss 의 *왜* 분석 — 가장 큰 학습 source |

## 5. 산출물 형식

```markdown
## Growth Experiment Backlog — {분기}

### Idea backlog
| hypothesis | target metric | ICE | RICE | priority |

### Sprint plan (next 2 weeks)
1. {experiment}: cascade to design-ab-experiment

### Result log
| experiment | result | learning |

### Quarterly retro
- win rate: X%
- biggest learning: ...
```

## 6. 검증

- [ ] Idea backlog 구조화 (hypothesis + metric + source + owner)?
- [ ] ICE 1차 + RICE 2차 평가?
- [ ] 실험 설계는 design-ab-experiment 호출?
- [ ] 모든 결과 (win + loss + inconclusive) 영속화?
- [ ] Sprint cadence (weekly / bi-weekly / monthly / quarterly)?
- [ ] 5 anti-pattern 회피 명시?

## 7. 다음 phase

- `design-ab-experiment` (구현됨) — 실험 설계 detail
- `analyze-ab-experiment` (구현됨) — Ship/Revert 결정
- `persist-learning-jsonl` (구현됨) — learning 영속화
- `summarize-retro` (구현됨) — 분기 retro

## 8. 참조

- Hacking Growth (Sean Ellis + Morgan Brown) — growth experiment framework
- ICE / RICE prioritization (Sean Ellis / Intercom)
- Lean Analytics (Croll + Yoskovitz) — vanity metric 회피
