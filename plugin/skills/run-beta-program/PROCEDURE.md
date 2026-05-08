# run-beta-program — 클로즈드 beta cohort 운영 + 피드백 + GA gating

§7 Release & Beta phase 의 stage. UAT pass 후 GA 직전, **5-20 명 early adopter cohort** 가 1-4 주 동안 production-like 환경에서 실 사용 → structured feedback 수집 → critical issue triage → GA go/no-go 결정. UAT (designated stakeholder) 와 GA (real production traffic) 사이의 마지막 안전망. 산출물은 cohort table + feedback corpus + GA gating decision + post-beta cleanup plan.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Cohort selection bias** — internal employee 만 / 친한 customer 만 → real-world coverage gap. industry / size / sophistication 다양성 강제.
- **Consent / NDA 없이 beta 진입** — 데이터 처리 권한 / 비공개 의무 명시 없이 invite 시 legal exposure. NDA + consent form 의무.
- **Feedback channel 단일** — Slack DM 만 또는 form 만 → 정량 vs 정성 어느 쪽 누락. 동기 (interview, NPS) + 비동기 (form, Slack) + 자동 (telemetry opt-in) 3 channel 필수.
- **GA gating 정량 기준 부재** — "feedback 좋았으니 GA" 식 → 정치적 결정. NPS / participation / critical issue 0 같은 정량 기준.
- **Beta 기간 무한정 연장** — 1-4 주 명확한 window. 연장 시 명시적 decision + rationale.
- **Feedback corpus 분석 없이 GA** — raw feedback 를 triage 안 하고 그대로 archive → improvement task 로 전환 안 됨, learning 손실.
- **Beta cleanup plan 없음** — beta 끝나도 cohort 가 production 으로 자동 transition 인지, 별도 termination 인지 결정 없음 → cohort confusion + churn risk.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

real users 의 production-like 사용 환경에서 UAT 가 못 잡은 도메인 / UX / scale issue 를 조기 검출. GA blast radius 가 cohort size 로 limited 된 안전망. structured feedback corpus 가 §8 iterate-product 의 baseline.

## 2. 사용 시점 (When to invoke)

- UAT pass 후 GA 직전 (권장)
- 신규 critical feature / breaking change major version 출시 시 (필수)
- 산업 규제 영역 (의료/금융 — beta 가 audit 산출물의 일부)
- 신규 product launch (beta 가 differentiation factor — early access 마케팅)
- 기존 product 의 major API change (3rd-party integration partner cohort 필요)

## 3. 입력 (Inputs)

### 필수
- UAT 산출물 (UAT pass / conditional go decision)
- staging or pre-production environment (production-like data + feature flag toggleable)
- cohort selection criteria (industry / size / risk tolerance / NDA status)
- beta 기간 결정 (1 주 / 2 주 / 4 주 옵션)
- feedback channel 인프라 (form / Slack workspace / interview slot)

### 선택
- 이전 beta 의 NPS / participation baseline
- 산업 규제 의무 (consent form 양식, retention policy)
- marketing collateral (early access 안내, embargo 정책)
- pricing / billing 모델 (beta 무료 / discount / paid)

### 입력이 부족할 때 forcing question
- "cohort size 가 5 미만이면 representative 부족, 20 초과면 운영 부담 — 적정 5-20 사이 어디?"
- "beta 기간 1 주는 너무 짧음 (사용 패턴 형성 시간 부족), 4 주 초과는 GA 지연 비용 큼 — 1-4 사이 어디 + rationale?"
- "consent + NDA 양식이 legal review 통과됐나? 없으면 cohort 리쿠르팅 전 차단."
- "GA gating 의 정량 기준이 명시됐나? '좋은 reception' 만으론 부족 — NPS 임계값 / participation rate / critical bug 수 정의."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — cohort selection / 기간 / GA gating 모두 단호한 결정. "그때 봐서" 거부.
- **사용자 입력을 challenge** — "5 명 정도 invite 할게" 발화에 "industry / size 다양성 검증? consent / NDA 준비?" push.
- **Specificity 강제** — "feedback 좋았어" 거부, NPS score / structured form completion rate / interview themes count.
- **Cohort fit > size** — 5 명 well-fit > 20 명 random. selection criteria 먼저, size 차순.

도메인 원칙:

1. **Cohort diversity 의무** — industry × size × sophistication 3 axis 에 분포. single-segment cohort 거부.
2. **Consent + NDA 의무** — 모든 cohort 멤버가 consent (data processing / feedback usage / retention) + NDA (embargo / non-disclosure) 서명.
3. **Multi-channel feedback** — 동기 + 비동기 + 자동 telemetry 3 channel 모두 활성. single-channel = bias.
4. **Structured triage 의무** — feedback 분류 (bug / feature request / UX / business model) + severity. raw archive 거부.
5. **GA gating 정량** — NPS ≥ X, participation ≥ Y%, critical bug ≤ Z, completion rate ≥ W%.
6. **Beta cleanup 명시** — beta 종료 후 cohort 의 transition (production 으로 graduate, churn allowed, 또는 별도 enterprise tier).

## 5. 단계 (Phases)

### Phase 1. Cohort selection + recruitment

selection criteria:

| Axis | Options | Min / Max |
|------|---------|-----------|
| Industry | SaaS / e-commerce / fintech / healthcare / education / non-profit | min 3 industries |
| Company size | startup (<50) / SMB (50-500) / enterprise (>500) | min 2 sizes |
| Sophistication | early adopter / mainstream / late majority | min 2 levels |
| Geography | NA / EU / APAC / LATAM | min 2 regions (또는 명시적 single-region 결정) |

cohort table:

| Cohort ID | Company | Industry | Size | Region | Contact | Consent / NDA | Onboarded | Status |
|-----------|---------|----------|------|--------|---------|---------------|-----------|--------|
| BC-001 | Acme Inc | SaaS | SMB | NA | jane@acme | signed 2026-05-NN | yes | active |
| BC-002 | Beta Co | fintech | startup | EU | tom@beta | signed | yes | active |
| ... | ... | ... | ... | ... | ... | ... | ... | ... |

recruitment 채널: existing customer outreach / waitlist / partner referral / cold outreach (rare). consent + NDA 양식 = legal-approved template (외부 reference, 본 skill 은 양식 location 만 명시).

### Phase 2. Beta period 설계

| Period option | Pros | Cons | Use case |
|---------------|------|------|----------|
| **1 week** | fast iteration, GA 지연 최소 | 사용 패턴 형성 부족, deep feedback 어려움 | minor feature / patch |
| **2 weeks** | balance, 권장 default | medium commit | 일반 feature launch |
| **4 weeks** | deep feedback, edge case 노출 | GA 지연 비용, beta fatigue | critical / 산업 규제 / first product launch |

milestone schedule (2 weeks 예시):
- **week 0**: onboarding (consent + NDA, environment access, first-call) — 2-3 day
- **week 1**: usage week 1 — daily check-in (Slack), 1 user interview / 5 cohort
- **week 2**: usage week 2 + structured feedback (form), 1 NPS survey, 1 group call
- **week 2 end**: feedback aggregation + triage + GA decision (2 day window)

### Phase 3. Feedback channel 정의

3 axis 모두 활성 의무:

| Channel type | Tool | Frequency | Purpose |
|--------------|------|-----------|---------|
| 동기 (interview) | Zoom / Google Meet, 30-45 min | week 1 / week 2 group call | deep qualitative — usability, mental model, pain point |
| 동기 (NPS / CSAT) | Typeform / Google Form, 1 일 응답 | end of week 2 | quantitative summary |
| 비동기 (structured form) | Notion / Airtable / Linear, daily fill | every day during beta | bug / feature request structured submit |
| 비동기 (Slack workspace) | shared Slack channel, on-demand | always-on | quick clarification, peer discussion |
| 자동 (telemetry) | OpenTelemetry / product analytics, opt-in | continuous | usage patterns, feature adoption rate, error rate |

opt-in / opt-out: telemetry 는 cohort consent 시 explicit opt-in. NPS 는 opt-out 가능 (단 GA gating 의 NPS pass criteria 영향).

### Phase 4. Structured triage

cohort feedback 을 daily / weekly 로 분류:

| Category | Definition | Severity scale | GA impact |
|----------|------------|----------------|-----------|
| **Bug** | unexpected behavior vs documented spec | blocker / major / minor | blocker = GA block, major = conditional |
| **Feature request** | functionality not in current scope | strong / moderate / nice-to-have | post-launch backlog |
| **UX issue** | flow / discoverability / friction | high / medium / low friction | high → patch within sprint, medium → backlog |
| **Business model feedback** | pricing / packaging / positioning | strategic / tactical | strategic → exec review, tactical → marketing |
| **Performance / scale** | latency / throughput / error rate | violates SLO / degraded / acceptable | SLO violation = blocker |

triage 의사결정자: PM (또는 release manager) + engineer lead. cadence: daily 15 min standup during beta.

### Phase 5. GA gating decision + post-beta cleanup

GA decision criteria:

| Decision | Pass criteria |
|----------|---------------|
| **go** | NPS ≥ +20, participation ≥ 60%, critical bug 0, completion rate ≥ 70%, cohort sign-off ≥ 80% |
| **conditional go** | NPS ≥ 0, critical bug ≤ 1 with hotfix ETA ≤ 48h, marketing message adjusted for known issues |
| **no-go (extend beta)** | NPS < 0, critical bug ≥ 2, completion rate < 50% — extend 1-2 weeks, fix issues, re-evaluate |
| **no-go (defer)** | systemic issue (architecture / business model) → back to §4 plan-build 또는 §3 design-system |

post-beta cleanup plan:

| Cleanup item | Decision |
|--------------|----------|
| Cohort transition | (a) graduate to production tier (default) / (b) churn allowed (free trial expired) / (c) special enterprise track |
| Feedback corpus retention | 1 year (compliance), structured archive in Notion / Linear |
| Beta-only feature flags | sunset within 1 sprint (cleanup task created) |
| NDA expiration | 6 months post-beta (allow customer testimonial, still no specific feature names) |
| Cohort acknowledgment | thank-you email + early access future betas opt-in offer |

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지. side-effecting (cohort recruit / consent collection / interview schedule) 은 외부 도구 위임, 본 skill 은 plan + 결과 표.

```markdown
## run-beta-program Output — <feature name> v<version>

### Summary
<3 줄: cohort N companies / period / NPS / decision>

### Cohort Table
| Cohort ID | Company | Industry | Size | Region | Consent/NDA | Onboarded | Status |
|-----------|---------|----------|------|--------|-------------|-----------|--------|
| BC-001 | ... | ... | ... | ... | ... | ... | ... |
| ... | ... | ... | ... | ... | ... | ... | ... |

### Diversity Validation
| Axis | Distribution | Pass criteria | Status |
|------|--------------|---------------|--------|
| Industry | SaaS 4, fintech 2, e-commerce 1 | ≥ 3 industries | OK |
| Size | startup 3, SMB 3, enterprise 1 | ≥ 2 sizes | OK |
| Region | NA 4, EU 3 | ≥ 2 regions | OK (note: APAC 미포함, GA marketing 시 region-specific outreach 필요) |

### Beta Schedule
| Phase | Period | Activities |
|-------|--------|------------|
| Onboarding | week 0 (3 days) | consent / NDA / environment / first-call |
| Usage week 1 | days 4-10 | daily Slack check-in, 1-on-1 interview × 5 |
| Usage week 2 | days 11-17 | structured form, NPS survey, group call |
| Feedback aggregation | days 18-19 | triage, GA decision |

### Feedback Corpus Summary
| Channel | Submissions | Action items |
|---------|-------------|--------------|
| Interview (week 1) | 5 sessions × ~40 min | 12 themes identified, 3 critical |
| Daily form | 47 entries (avg 6.7 / day) | 18 bugs / 12 feature requests / 8 UX / 9 other |
| Slack messages | 234 messages, 89 threads | quick clarifications, no formal action |
| NPS survey | 6/7 responded (86%), avg +27 | passes ≥ +20 threshold |
| Telemetry | continuous, opt-in 7/7 | feature adoption: signup 100%, login 100%, /me 71%, password-reset 14% (low — design opportunity) |

### Triage Output
| Category | Count | Critical | Major | Minor / nice |
|----------|-------|----------|-------|--------------|
| Bug | 18 | 0 | 2 | 16 |
| Feature request | 12 | n/a | 4 strong | 8 |
| UX issue | 8 | 1 high friction | 2 medium | 5 low |
| Business model | 4 | n/a | 2 strategic (pricing) | 2 tactical |
| Performance | 5 | 0 SLO violation | 1 degraded (login p99 480ms vs 500 budget) | 4 acceptable |

### GA Decision
| Field | Value |
|-------|-------|
| Decision | **conditional go** |
| Rationale | NPS +27 (pass), 0 blocker, 2 major bug + 1 high-friction UX → hotfix plan within 48h |
| Conditions | (1) hotfix bug-BC-014 (verification email delivery delay >5 min) — owner: backend lead, ETA: 36h. (2) UX hotfix on /reset-password copy — owner: design lead, ETA: 24h. |
| Marketing message | "early access launch" with known limitations footer (link to status page) |
| Post-beta cohort transition | graduate to production tier with 3-month complimentary support |

### Post-Beta Cleanup Plan
- Beta feature flag `beta_signup_v2` → sunset within next sprint (commit task assigned)
- NDA expiration: 2026-11-NN (6 mo from beta end)
- Feedback corpus archived in Notion: https://notion.so/beta-saas-auth-v1.0.0
- Cohort thank-you email batch: 2026-MM-NN
- Future beta opt-in form sent to all cohorts

### Cascade
- **§7 prepare-launch-checklist**: beta decision + cleanup plan 이 launch checklist 의 row
- **§8 analyze-customer-feedback-corpus** (해당 시): post-launch 에 beta feedback corpus 가 baseline
- **§8 generate-improvement-tasks**: triage output 의 strong / medium feedback 이 §2 backlog 재진입 input
- **§8 handle-incident** (해당 시): conditional go 의 hotfix 가 스케줄 이탈 시 incident 화

### Next Step
<구체 action — 1줄: 예 "hotfix bug-BC-014 + UX copy 변경 → 48h 내 GA, marketing kickoff 동시 진행">
```

## 7. Cross-phase cascade

- **§7 prepare-launch-checklist**: beta decision + cleanup 가 readiness gate row
- **§7 setup-feature-flags**: beta-only flag (예: `beta_signup_v2`) sunset task 가 setup-feature-flags 의 cleanup policy 와 정합
- **§8 analyze-customer-feedback-corpus**: beta feedback corpus 가 production corpus 의 baseline
- **§8 generate-improvement-tasks**: triage 의 medium / strong feedback 이 backlog 재진입
- **§8 design-ab-experiment**: beta 가 implicit A/B (control = non-beta production) — 향후 정식 A/B 의 hypothesis 출처

## 8. 다음 skill (next in stage flow)

- `prepare-launch-checklist` — beta 결과를 readiness gate 로 통합
- `setup-feature-flags` (이미 호출됐으면) — beta flag sunset cleanup task
- (re-execute 시) `run-beta-program` 자체 재호출 — extend beta 시 동일 cohort + 보강 scenario

권장 chain (Phase 3 §7 cascade 후반):
```
/buddy:chain run-beta-program,prepare-launch-checklist -- "<feature name> v<version>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `run-uat`** (§7) — UAT = designated stakeholder (PM / admin / 1-3 user, internal-leaning), beta = real users (5-20 cohort). UAT 가 먼저, beta 가 후속. UAT 는 binary go/no-go, beta 는 spectrum (+ post-cleanup).
- **vs `prepare-launch-checklist`** (§7) — checklist 는 17+ row의 final readiness gate, 본 skill 은 그 중 한 input (beta status row).
- **vs `analyze-customer-feedback-corpus`** (§8) — 그것은 production corpus (continuous), 본 skill 은 closed beta period (bounded). beta corpus 가 production corpus 의 baseline.
- **vs `design-ab-experiment`** (§8) — 그것은 통계적 A/B 설계 (formal hypothesis test), 본 skill 은 qualitative + quantitative beta (less formal, pre-launch).
- **vs `triage-customer-support-ticket`** (§8 — pending) — 본 skill 의 triage 는 beta period 한정, support ticket 은 production-scale.

## 10. 중요 규칙

- **Cohort diversity 의무** — single-segment cohort 거부.
- **Consent + NDA 의무** — 미서명 cohort 의 데이터 사용 / feedback 인용 금지.
- **Multi-channel feedback 의무** — 3 axis (sync / async / telemetry) 모두 활성.
- **Structured triage 의무** — raw feedback archive 거부.
- **GA gating 정량** — NPS / participation / critical bug 임계 명시.
- **Beta cleanup plan 의무** — flag sunset, NDA expiration, cohort transition 모두 명시.
- **Read-only on production** — beta 는 staging or feature-flagged production, 별도 prod 데이터 수정 안 함.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행 (cohort selection → period 설계 → channel 정의 → triage → GA gating + cleanup)
- [ ] §6 의 9 출력 섹션 (Summary / Cohort / Diversity / Schedule / Feedback Corpus / Triage / GA Decision / Cleanup / Cascade) 모두 채워짐
- [ ] Cohort Table 5-20 row, 각 row 에 consent/NDA status + onboarded + status 명시
- [ ] Diversity Validation 3 axis (industry / size / region) 모두 pass criteria 통과 (또는 미통과 시 GA 영향 명시)
- [ ] Beta Schedule 의 4 phase (onboarding / week 1 / week 2 / aggregation) 모두 정의, 1-4 주 window 내
- [ ] Feedback Corpus Summary 의 3 channel (sync / async / telemetry) 모두 활성, 정량 metric (count / NPS / rate)
- [ ] Triage Output 의 5 category 모두 분류, severity 명시
- [ ] GA Decision 의 정량 pass criteria (NPS / participation / critical / completion / sign-off) 모두 평가, conditional 시 condition + ETA + owner 명시
- [ ] Post-Beta Cleanup Plan 5 항목 (flag sunset / NDA / archive / cohort transition / future beta) 모두 명시
- [ ] §4 posture — go/no-go 단호, "feedback 좋았음" 류 hedge 없음
- [ ] §0 anti-pattern 부재 — selection bias / consent / channel diversity / triage / gating / cleanup 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
