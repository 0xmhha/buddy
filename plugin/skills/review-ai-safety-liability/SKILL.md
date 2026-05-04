---
name: review-ai-safety-liability
description: "AI 기반 기능의 책임 범위, 할루시네이션 리스크, 자동 의사결정의 영향, content provenance, model output safeguards를 검토. high-risk 도메인(의료/법률/재무/안전)에서 사용자 피해 가능성 + 회사 liability 평가. 트리거: 'AI 책임 범위' / 'hallucination 리스크' / '자동 의사결정 검토' / 'AI safety 봐줘' / 'LLM 출력 책임' / 'AI 윤리' / 'content provenance'. 입력: AI 기능 spec, 사용 도메인, autonomous level, content type. 출력: liability matrix + safeguard list + disclosure plan + incident response. 흐름: validate-advanced-edge-idea → review-ai-safety-liability → define-product-spec/audit-security."
type: skill
---

# Review AI Safety and Liability — AI 책임 6차원 검증

## 1. 목적

AI 기반 기능(LLM 응답, 자동 의사결정, content 생성, agent action)의 **사용자 피해 가능성 + 회사 liability**를 6차원으로 평가한다.

핵심 질문:
- AI가 잘못된 답을 했을 때 누가 책임?
- AI가 자동으로 결정한 것이 사용자에게 significant effect 주면 GDPR Art. 22 적용?
- 생성된 content의 출처(provenance)는 표시되나?
- vulnerable user (minor / patient / debtor) 보호 메커니즘?
- model이 deepfake / fake news / harassment 도구로 악용되면?
- AI 의존 사용자가 직접 능력 잃어도 책임 어디?

이 스킬을 통과한 AI 기능은 **liability matrix + disclosure + safeguard + incident response 4세트**를 가진다.

## 2. 사용 시점 (When to invoke)

- LLM API / agent / 자동 의사결정 기능 출시 전
- high-risk 도메인 (의료 / 법률 / 재무 / 안전 / 미성년) 진입
- 자동 content 생성 (text / image / voice / video)
- automated moderation / scoring / filtering
- recommendation engine (significant effect 가능)
- AI agent autonomous action (사용자 대신 결정)
- model 변경 시 재검증 (output 분포 변화)

## 3. 입력 (Inputs)

### 필수
- AI 기능 spec (input → model → output)
- 사용 도메인 (health / finance / law / general / creative)
- autonomy level (suggestion / decision / action)
- content type (text / image / voice / video / decision)
- target user (general / professional / vulnerable)

### 선택
- 사용 model (Claude / GPT / Gemini / open-source)
- model card / safety evaluation
- 기존 incident report

### 입력 부족 시 forcing question
- "이 AI가 잘못된 의료 / 법률 / 재무 조언 했을 때 책임이 누구야? 회사? 사용자? model 회사?"
- "사용자가 이 AI 결정에 따라 행동하면 회사가 신뢰를 보장해? 그냥 'AI는 도구일 뿐'이야?"
- "minor / 환자 / 정신질환자 / 외국어 사용자 같은 vulnerable group 보호 어떻게?"
- "model output이 잘못되면 사용자에게 immediate harm 가능해 (medication 오안내 등)?"

## 4. 핵심 원칙 (Principles)

1. **AI는 책임 회피 도구 아님** — "AI가 그랬다"로 회사 면책 안 됨. autonomous level 따라 회사 책임.
2. **High-risk 도메인은 human-in-the-loop** — 의료 / 법률 / 재무 / 안전은 AI suggestion + human decision.
3. **Disclosure 의무** — AI 생성 content 표시. EU AI Act, 미국 일부 주, 한국 가이드라인.
4. **Hallucination은 dataset 문제 아님** — generative model 본질. user 보호 mechanism 필수.
5. **Vulnerable user는 별도 가드** — minor: parental consent. patient: provider 검증. 외국어: 정확도 fallback.
6. **Content provenance** — 생성 출처 표시 (C2PA, watermark, metadata).
7. **Automated decision은 GDPR Art. 22** — significant effect 시 명시 동의 + human review 권리 + logic 설명.
8. **Model 변경은 재평가** — output 분포 변화. 안전 평가 재실행 강제.

## 5. 6 평가 차원 (Risk Dimensions)

### Dimension 1. Hallucination Risk
LLM 출력의 사실 오류 가능성.

질문:
- 도메인별 정확도 측정?
- factual claim에 source citation?
- "I don't know" fallback?
- confidence score 노출?

mitigation:
- retrieval-augmented generation (RAG)
- citation 강제
- confidence threshold (낮으면 fallback)
- human review for high-stakes

### Dimension 2. Autonomous Decision Impact
AI 결정이 사용자에 미치는 영향.

레벨:
- **Suggestion**: 사용자가 결정 (recommendation, autocomplete)
- **Filter**: AI가 보여줄 것 결정 (content moderation, search ranking)
- **Decision**: AI가 자동 결정 (loan approval, hiring screen, fraud)
- **Action**: AI가 실행 (auto-trade, auto-pilot, agent execute)

decision / action 레벨은 GDPR Art. 22 적용 가능 — 명시 동의 + human review.

### Dimension 3. Content Provenance & Misuse
생성 content의 출처 + 악용 가능성.

질문:
- AI 생성 표시? (watermark, metadata, UI label)
- deepfake / 음성 합성 가능?
- harassment / fake news 도구로 사용 가능?
- C2PA / IPTC 같은 표준 적용?

### Dimension 4. Vulnerable User Protection
보호 대상 + 메커니즘.

대상:
- minor (under 13/16/18 by region)
- patient / health condition
- 외국어 사용자 (정확도 저하)
- 정신건강 (suicide / self-harm 콘텐츠)
- debtor / financial vulnerable
- elderly

메커니즘:
- age gate
- domain-specific filter
- escalation to human (crisis)
- fallback to professional referral

### Dimension 5. Liability Allocation
책임 분배.

actors:
- 회사 (서비스 제공자)
- 사용자
- model 회사 (Anthropic / OpenAI / Google)
- third-party (data, tool)

질문:
- AI 잘못된 답변으로 사용자 피해 시 누가 배상?
- ToS에서 limitation of liability 명시?
- AI 사용 제한 / 면책 사항 사용자 동의?

### Dimension 6. Sustained AI Dependence
사용자가 AI에 의존하면서 직접 능력 잃을 가능성.

질문:
- "AI가 없으면 못함" 상태 만들지 않나?
- learning / skill development 저해?
- AI offline / 오류 시 fallback?

## 6. 단계 (Phases)

### Phase 1. AI Function Mapping
input → model → output 전체 mapping.
- 어떤 input
- 어떤 model (version, provider)
- 어떤 output (text / structured / action)

### Phase 2. Risk Dimension Scoring
6 차원 각각 0-1.0 risk score.

### Phase 3. Mitigation Design
각 risk에 대해 mitigation:
- technical (RAG, threshold, filter)
- UX (disclosure, citation, confidence)
- process (human review, escalation)
- legal (ToS, consent)

### Phase 4. Disclosure Plan
사용자 통지:
- AI 사용 표시 (UI label, watermark)
- 한계 명시 ("not medical advice")
- data 사용 (training, retention)
- 권리 (opt-out, human review request)

### Phase 5. Incident Response Plan
- detection (user report, monitoring)
- triage (severity, scope)
- mitigation (rollback, hotfix, model swap)
- communication (affected users, regulators)
- post-mortem + learning

### Phase 6. Liability Matrix
각 시나리오에 대해 책임 분배:
| 시나리오 | 회사 | 사용자 | Model 회사 |
|---|---|---|---|
| hallucination → 의료 오판 | 50% (human-in-loop 부재) | 30% (의사 검증 안 함) | 20% (model error) |

## 7. 출력 템플릿 (Output Format)

```yaml
ai_function_spec:
  feature: "<...>"
  input: ["<...>"]
  model: { provider: anthropic, name: claude-opus-4-7, version: 2026-04 }
  output_type: text | image | structured | action
  autonomy_level: suggestion | filter | decision | action

risk_scores:
  hallucination: 0.0-1.0
  autonomous_decision: 0.0-1.0
  content_provenance: 0.0-1.0
  vulnerable_user: 0.0-1.0
  liability_clarity: 0.0-1.0
  sustained_dependence: 0.0-1.0

mitigations:
  - dimension: hallucination
    technical: ["RAG", "citation enforcement"]
    ux: ["confidence indicator", "I don't know fallback"]
    process: ["human review for medical claims"]
    legal: ["disclaimer in ToS"]

disclosure_plan:
  ai_label: "Generated by AI"  # UI 표시
  watermark: c2pa_signed | none
  limitations_text: ["not medical advice", "verify before action"]
  data_usage_disclosure: "input not used for training"
  opt_out: yes | no
  human_review_right: yes  # GDPR Art. 22

vulnerable_user_safeguards:
  minor:
    age_gate: yes
    parental_consent: yes
    safe_filter: enforced
  health:
    medical_disclaimer: yes
    crisis_escalation: 988 / KR helpline
    professional_referral: yes
  financial:
    disclaimer: "not financial advice"
    professional_referral: yes

provenance:
  standard: C2PA | IPTC | none
  watermark_visible: yes | no
  metadata_signed: yes | no

liability_matrix:
  - scenario: "hallucination → user takes wrong action"
    company_share: 60
    user_share: 30
    model_provider_share: 10
    legal_basis: "ToS Section X + GDPR Art. 22"
  - scenario: "deepfake misuse"
    company_share: 40
    user_share: 60
    legal_basis: "AUP violation"

incident_response:
  detection_sources: ["user report", "automated monitoring"]
  severity_levels:
    critical: "harm to user"
    high: "regulatory violation"
    medium: "service degradation"
  rollback_capability: yes
  user_notification_threshold: "critical"
  authority_notification: "AI Act Art. 26"

verdict: clear | needs-remediation | blocked
```

## 8. 자매 스킬 (Sibling Skills)

- 앞 단계: `validate-advanced-edge-idea` — `Skill` tool로 invoke (ethical blind spot 차원)
- 페어: `review-privacy-data-risk` — AI training data, output 보유
- 페어: `review-license-and-ip-risk` — 학습 data IP, output copyright
- 페어: `audit-security` — adversarial input, prompt injection
- 다음 단계: `define-product-spec` — AI safeguard requirements를 PRD에

## 9. Anti-patterns

1. **"AI는 도구일 뿐"** — 책임 회피 안 됨. autonomous level 따라 회사 책임.
2. **High-risk 도메인 자동 결정** — 의료 / 재무 / 법률을 AI가 결정. human-in-the-loop 강제.
3. **AI 표시 없이 출력** — EU AI Act 위반 가능. watermark / label 강제.
4. **Hallucination을 "model 발전으로 해결"로** — generative 본질. mitigation 강제.
5. **Vulnerable user 일반 처리** — minor / patient / financial vulnerable 별도 가드.
6. **Liability ToS에 다 떠넘기기** — 소비자보호법 위반. fair allocation.
7. **Model 변경 후 재평가 skip** — output 분포 변화. 안전 평가 재실행.
8. **Incident response 사전 미설계** — 발생 후 처음 만들면 너무 늦음.
