# Plan Marketing Channel — paid / co-marketing / community / directory / lead-magnet 통합

## 1. 목적

마케팅 channel 의 *분포 + ROI 측정 + 우선순위* 계획. **paid ads / co-marketing / community / directory / lead magnet / free tool** 6 channel 통합 + channel-fit 평가.

`marketingskills` 의 6 channel sub-skill 통합.

## 2. 사용 시점

- §8 iterate-product 의 *분기 marketing planning*
- 신규 channel 진입 결정
- channel ROI drift 시 — 어느 channel 줄이고 늘릴지
- product launch 전 — channel mix 결정
- 경쟁사가 새 channel 진입 후 — defensive 검토

## 3. 입력

### 필수
- `map-customer-segments` 의 buyer / user reachability
- `analyze-user-cohort` (§8) 의 channel 별 LTV / CAC
- 마케팅 budget

### 선택
- 경쟁사 channel 분포 (visible 부분)
- brand 단계 (early / growth / mature)

## 4. Stage 흐름

### Stage 1: 6 channel 분류

| channel | 정의 | 적합 단계 | 측정 |
|---------|-----|--------|------|
| **Paid ads** | Google / Meta / LinkedIn / TikTok | growth+ | CAC + ROI |
| **Co-marketing** | partnership content / webinar / bundle | mature+ | reach + co-conversion |
| **Community** | Reddit / Discord / Slack / forum | early+ | mention + sentiment |
| **Directory** | Product Hunt / G2 / SaaS directory | early | listing + review |
| **Lead magnet** | ebook / template / calculator | growth+ | leads + qualified |
| **Free tool** | mini-tool / freemium feature | early-growth | usage + conversion |

### Stage 2: Channel-fit 평가

각 channel 의 *우리 제품 적합도* :

| 차원 | 측정 |
|------|-----|
| Audience match | buyer / user 가 해당 channel 에 *있나* |
| Voice fit | 우리 brand voice 가 channel native 와 정합 |
| Effort | 진입 비용 (시간 + 돈) |
| Defensibility | competitor 가 쉽게 따라할 수 있나 |

→ 4 차원 점수 (1~5). 합계 12+ channel 부터 진입.

### Stage 3: ROI 측정 — channel 별 LTV / CAC

`analyze-user-cohort` (§8) 의 channel attribution:

| channel | CAC | LTV | LTV/CAC | payback |
|---------|-----|-----|---------|--------|
| Paid Google | $50 | $200 | 4× | 6 month |
| Community | $5 | $150 | 30× | 1 month |
| Directory | $0 | $100 | ∞ | 0 |
| ... | ... | ... | ... | ... |

→ LTV/CAC < 3 인 channel 은 *재검토* (efficiency 낮음).

### Stage 4: Channel mix 결정

| brand 단계 | 권장 mix |
|---------|--------|
| Early (PMF 검증) | Community 60% + Directory 30% + 기타 10% |
| Growth | Paid 40% + Community 20% + Lead-magnet 20% + 기타 |
| Mature | Paid 50% + Co-marketing 30% + 기타 |

→ 단계 별 mix 다름. *전 단계 mix* 그대로 두면 비효율.

### Stage 5: Channel 별 sub-strategy

| channel | sub-strategy |
|---------|----------|
| Paid ads | keyword / audience targeting / bidding strategy |
| Co-marketing | partner 선정 / 형식 (webinar / blog / bundle) / 분배 |
| Community | 어느 community / engage 빈도 / authentic vs promotion 균형 |
| Directory | listing 작성 / review 유도 / pricing tier 표기 |
| Lead magnet | 어떤 자료 / gating 강도 / nurture sequence |
| Free tool | 기능 범위 / paid 전환 path |

### Stage 6: Drift monitoring

분기 별 channel mix 변화 + ROI drift:
- mix 비율 변화 (의도 vs 실제)
- channel ROI 추세 (개선 / 악화)
- saturation 신호 (CAC 상승)

## 5. 산출물 형식

```markdown
## Marketing Channel Plan — {분기}

### Channel-fit 평가
| channel | audience | voice | effort | defensibility | 합 |

### ROI (channel × LTV/CAC)
| channel | CAC | LTV | 비율 |

### Channel mix (현재 vs 권장)
| channel | 현재 % | 권장 % |

### Sub-strategy per channel
- ...

### Drift monitoring
- 분기 review schedule: ...
```

## 6. 검증

- [ ] 6 channel 모두 channel-fit 평가?
- [ ] ROI (LTV/CAC) channel 별 측정?
- [ ] Brand 단계 별 mix 적용?
- [ ] Sub-strategy 활성 channel 만 작성 (saturation 회피)?
- [ ] Drift monitoring 분기 review?
- [ ] LTV/CAC < 3 channel *재검토* 결정?

## 7. 다음 phase

- `draft-marketing-copy` — 채널 별 카피
- `automate-marketing-content` — 콘텐츠 스케줄
- `analyze-user-cohort` (§8) — channel attribution 정밀 측정

## 8. 참조

- Traction (Gabriel Weinberg) — 19 channel 분류
- marketingskills 6 channel sub-skills (외부 reference, MIT — paid-ads / co-marketing / community-marketing / directory-submissions / lead-magnets / free-tool-strategy)
