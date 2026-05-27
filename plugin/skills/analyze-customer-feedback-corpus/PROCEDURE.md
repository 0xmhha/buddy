# Analyze Customer Feedback Corpus — NPS / 리뷰 / 텍스트 토픽 모델링

## 1. 목적

CS ticket / NPS comment / app review / 인터뷰 transcript 같은 *비구조 텍스트* 를 *구조화 insight* 로. **토픽 모델링 + sentiment + verbatim quote 추출 + product input** 통합.

`triage-customer-support-ticket` (§8) 산출 + `conduct-customer-interview` (§1) transcript 입력.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Feedback corpus (CS/NPS/review/interview 텍스트) | ✅ | artifact | CS 시스템 또는 사용자 제공 | "분석할 피드백 데이터를 제공해 주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Topic analysis (토픽 모델링 + sentiment + verbatim quotes) | artifact | structured report | `generate-improvement-tasks` |

## 2. 사용 시점

- §8 iterate-product 의 *분기 voice-of-customer 분석*
- 신규 release 후 — *어떤 반응* 패턴 식별
- NPS 측정 후 — promoter / detractor 의 *이유* 분석
- App store rating drop 시 — review corpus 분석
- pivot 결정 시 — 사용자 *진짜 가치 인식*

## 3. 입력

### 필수
- 텍스트 corpus (CS / NPS / review / interview transcript)
- 시간 dimension (분기 / 월 별)
- 분석 도구 (LLM 또는 NLP — topic modeling)

### 선택
- 사용자 segment (segment 별 의견 분포)
- 경쟁사 review (비교 baseline)

## 4. Stage 흐름

### Stage 1: Corpus 수집 + 정규화

| source | 양식 |
|--------|-----|
| CS ticket | subject + body + 답변 |
| NPS comment | score (0~10) + comment |
| App review | rating + body + reviewer |
| Interview transcript | semi-structured |
| 외부 (Twitter / Reddit / 커뮤니티) | post / comment |

각 source 의 *내부 noise* 제거:
- humanizer 활용 (외부 reference, MIT) — AI text 자연화 inverse 패턴 (자연스러운 텍스트 패턴 검증)
- spam / duplicate 제거
- PII redaction

### Stage 2: 토픽 모델링

| 방법 | 도구 |
|------|-----|
| Manual coding | 5~10 카테고리 사전 정의 + 인간 분류 |
| LLM-assisted | Claude / GPT 가 *분류 제안* + 인간 review |
| LDA (Latent Dirichlet Allocation) | gensim / sklearn — topic 자동 추출 |
| BERTopic | sentence embedding + clustering |

→ 본격 분석 전 *5~10 sample* 으로 LLM 분류 검증. 잘못된 carrier-of-meaning 발견 시 prompt 정정.

### Stage 3: Sentiment 분석

각 mention 의 *감정* :
- Positive / Negative / Neutral / Mixed
- intensity (1~5)
- sarcasm 검출 (어려움 — LLM 보조)

토픽 × sentiment cross:
```
              Positive  Negative  Neutral
Onboarding    20%       45%       35%   ← 부정 dominant
Pricing       30%       40%       30%
Performance   60%       15%       25%   ← 긍정 dominant
```

→ *부정 dominant 토픽* = product 우선순위.

### Stage 4: Verbatim quote 추출

각 토픽의 *대표 quote* :
- 가장 *명확한* 발화
- 가장 *자주 등장* 하는 표현
- *가장 강한 감정* 의 발화

→ stakeholder 발표 / product spec 의 *evidence* 로 활용. *추상화된 평균* 보다 강력.

### Stage 5: NPS Verbatim 분석 — promoter / detractor 차이

NPS 의 *왜* 차원:

| segment | 평균 score | 자주 언급 토픽 | verbatim |
|---------|---------|------------|---------|
| Promoter (9-10) | 9.3 | "{ease of use}" / "{customer support}" | "..." |
| Passive (7-8) | 7.5 | "missing {feature}" / "occasional bug" | "..." |
| Detractor (0-6) | 4.2 | "{onboarding confusion}" / "{pricing}" | "..." |

→ promoter 의 가치 *강화* + detractor 의 friction *제거*.

### Stage 6: Product feedback loop

각 토픽 → product action:
- Bug → bug tracker
- Feature request → `score-feature-priority` 입력
- UX confusion → `audit-ui-quality` 입력
- Onboarding friction → `analyze-feature-adoption` cross

## 5. 산출물 형식

```markdown
## Customer Feedback Corpus — {분기}

### Source 분포
| source | volume | weight |

### 토픽 × Sentiment 매트릭스
| 토픽 | Pos% | Neg% | Neu% | priority |

### 부정 dominant 토픽
1. {topic}: {neg %}
   - quote: "..."
   - action: ...

### NPS Verbatim
| segment | score | top topics | quote |

### Product action 매핑
- bug tracker: {N}
- feature backlog: {N}
- UX gap: {N}
```

## 6. 검증

- [ ] Corpus 다중 source (CS / NPS / review / interview) 통합?
- [ ] 토픽 모델링 (manual + LLM 또는 LDA) ?
- [ ] Sentiment × topic cross 매트릭스?
- [ ] Verbatim quote 토픽 별 추출?
- [ ] NPS promoter / passive / detractor 분리?
- [ ] Product feedback loop 4 영역 매핑?

## 7. 다음 phase

- `score-feature-priority` 의 user-driven 입력
- `audit-ui-quality` (그룹 4) 의 friction 식별
- `generate-improvement-tasks` (구현됨) — product backlog
- `summarize-retro` (구현됨) — 분기 voice-of-customer 요약

## 8. 참조

- Net Promoter Score (Fred Reichheld) — promoter / detractor 원칙
- The Mom Test (Rob Fitzpatrick) — verbatim 발화 신뢰성
- marketingskills/customer-research (외부 reference, MIT)
- humanizer (외부 reference, MIT — Siqi Chen, AI text 자연화 inverse 패턴)

---

## MCP integration (analytics-mcp v0.2.0+)

본 skill 의 *feedback corpus search + topic / sentiment 분석* 에서 `analytics_query_feedback_corpus` MCP tool 호출.

**호출 예**:

```json
{
  "tool": "analytics_query_feedback_corpus",
  "arguments": {
    "source": "nps_comment",
    "time_range": { "from": "2026-04-01T00:00:00Z", "to": "2026-05-01T00:00:00Z" },
    "topic_modeling": true,
    "sentiment_analysis": true
  }
}
```

기대 응답: total_items + topics (with verbatim_quotes) + nps_segments (promoter / passive / detractor band 별 top_topics).

**전제**: `BUDDY_ANALYTICS_BACKEND` env var 설정 (elasticsearch 권장 — text corpus search 강함). v0.2.0 stubs — 본격 adapter 는 W4-2.2 ~ W4-2.3.

**관련 spec**: `../../../docs/superpowers/specs/2026-05-10-analytics-mcp-spec.md` §4.7
