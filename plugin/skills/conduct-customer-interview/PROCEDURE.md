# Conduct Customer Interview — 인터뷰 스크립트 + 결과 코딩

## 1. 목적

가설 검증의 *가장 신뢰할 수 있는 신호* 는 잠재 고객 인터뷰. 이 skill 은 **인터뷰 스크립트 작성 → 인터뷰 진행 → 결과 코딩 → 가설 업데이트** 의 표준 절차를 제공한다.

`assess-business-viability` 의 *willingness-to-pay 검증* + `map-customer-segments` 의 *persona 정합 검증* 의 핵심 입력.

## 2. 사용 시점

- `validate-idea` 통과 후 *문제 진짜로 존재하나?* 검증 단계
- `assess-business-viability` 의 WTP / GTM / 경쟁 차원 검증
- `map-customer-segments` 의 persona 가설 정합 검증
- `map-jobs-to-be-done` 의 job 가설 / outcome 검증
- pivot 결정 전 — 어느 가설이 깨졌는지 식별
- pricing 결정 전 — buyer 가 *얼마면 산다* 신호
- 경쟁사 등장 후 — *우리 제품 vs 경쟁* user 평가

## 3. 입력

### 필수
- 검증할 *가설 list* (최소 3, 최대 7) — 한 인터뷰에 너무 많으면 깊이 부족
- target persona (`map-customer-segments` 산출)
- 인터뷰 mode (1:1 깊이 / focus group / async survey)

### 선택
- 경쟁사 / substitute 정보 (비교 질문 입력)
- 기존 customer base (있으면 — re-engagement)

## 4. Stage 흐름

### Stage 1: 인터뷰 스크립트 작성

3 단계 구조:

| 구간 | 시간 | 목적 |
|-----|------|------|
| Warm-up | 5 min | rapport 형성, 자유 발화 유도 |
| Past behavior | 15 min | 과거 *실제 행동* 추출 — "the past predicts the future" (The Mom Test) |
| Future hypothesis | 5 min | 가설 검증 — *조건부* 질문 ("if X then would you ...?") |
| Wrap-up | 5 min | 추천인 / 다음 인터뷰 후보 |

**금지 질문 (The Mom Test 의 anti-pattern)**:
- "쓸 것 같아요?" — 의도 (intent) 묻기 — false positive 자주
- "X feature 있으면 좋겠어요?" — pitch 형 질문 — 친절 답변 유도
- "얼마면 사겠어요?" — willingness-to-pay 직접 — false signal

**권장 질문**:
- "마지막으로 {문제 발생} 한 게 언제였나요?" — past behavior
- "그때 어떻게 해결했나요?" — current solution
- "그게 얼마나 [시간 / 돈 / 짜증] 들었나요?" — pain magnitude
- "다른 사람도 같은 문제 겪나요?" — virality 신호

### Stage 2: 인터뷰 진행

- 1:1 30~45 min (focus group 60~90 min)
- 녹화 / 메모 (사용자 동의)
- 인터뷰 mode 별 *동기 다름* (free interview vs paid customer dev)
- gpt-researcher 같은 자동 도구는 *부분 보조* — 인간 인터뷰 대체 X (감정 / 사회 차원 캐치 어려움)

### Stage 3: 결과 코딩

raw transcript → 구조화:

| 컬럼 | 내용 |
|------|------|
| segment | persona 일치 여부 |
| pain | 명시된 pain (직접 인용) |
| current solution | 현재 해결책 |
| willingness-to-switch | 1~5 점 (5 = 즉시 전환 의향) |
| WTP signal | 가격 / 시간 가치 발화 |
| objection | 반대 / 우려 |
| quote | inline quote (verbatim) |

5~10 인터뷰 후 *패턴* 추출 — 3+ 인터뷰가 같은 pain 언급 시 신호 강함.

### Stage 4: 가설 업데이트

각 검증 가설에 *evidence* 라벨:

- ✅ Confirmed (3+ 인터뷰 정합)
- ⚠️ Partial (1~2 인터뷰만 정합 — 추가 검증)
- ❌ Refuted (반대 신호 dominant)
- 🆕 New hypothesis (인터뷰 중 surface — 별도 검증 필요)

## 5. 산출물 형식

```markdown
## Customer Interview Results — {제품 이름} (n=N)

### Methodology
- Persona: {early adopter}
- Mode: {1:1 / focus / async}
- N = {인터뷰 수}

### Coding table
| ID | segment | pain | solution | switch | WTP | objection |

### Hypothesis updates
- ✅ {가설 1}: ... (evidence: ...)
- ⚠️ {가설 2}: ... (추가 검증 필요)
- ❌ {가설 3}: ... (refuted)
- 🆕 {신규 가설}: ... (별도 검증)

### Next action
- 가설 update → assess-business-viability 입력
- pivot 후보 / scope narrow 결정
```

## 6. 검증

- [ ] Mom Test anti-pattern (intent / pitch / WTP-direct) 회피?
- [ ] Past behavior 가 future hypothesis 보다 더 많은 시간 할당?
- [ ] N ≥ 5 (sample 작은 경우 패턴 confidence 낮음)
- [ ] 코딩 table 모든 인터뷰 행 채움?
- [ ] 가설 update 4 라벨 (confirmed / partial / refuted / new) 적용?

## 7. 다음 phase

- `assess-business-viability` 의 *evidence 보강* 입력 (특히 WTP / GTM 차원)
- `map-customer-segments` 의 persona refinement (인터뷰 결과와 차이 시 segment 재정의)
- `map-jobs-to-be-done` 의 *job verbalization* 보강 — 인터뷰의 verbatim quote 활용

## 8. 참조

- The Mom Test (Rob Fitzpatrick) — anti-pattern + past-behavior
- Talking to Humans (Giff Constable) — interview methodology
- marketingskills/customer-research (외부 reference, MIT — 차용 시 attribution)
- designer-skills/design-research (외부 reference, MIT)
- gpt-researcher (외부 — 자동 보조 도구, 인간 인터뷰 대체 X)
