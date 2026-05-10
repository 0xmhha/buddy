# Dogfood Validation Scenarios — Skill Completion Cycle 종료 후 검증

> **Date**: 2026-05-10
> **Status**: 사용자 액션 대기 (claude plugin 재install + skill 호출 테스트)
> **Predecessor**: Skill Completion Cycle 종료 (commit `900944f`), SSoT refresh (`87529e2`), ADR-003 (`3c3c4bb`)

본 문서는 *Skill Completion Cycle 100% 완료* (44 skill / 통합 2) 후 **plugin buddy 가 실제로 동작하는지** 검증하는 시나리오. 사용자 액션이 필요 (claude plugin re-install + Claude Code session 안에서 skill 호출).

---

## 1. 검증 목적

| 영역 | 측정 |
|------|------|
| **Discovery** | 신규 skill 의 `/buddy:<name>` 가 자동완성에 등장 |
| **Description** | command frontmatter 의 description 이 baseline context 에서 *제외* (ADR-001 정합) |
| **Routing** | router skill 이 자연어 → 신규 skill dispatch 가능 |
| **Cascade** | skill A → skill B 호출 흐름 정합 (예: decide-target-market → review-legal-regulatory) |
| **PROCEDURE 정합** | skill 본문의 12 section 양식 일관 |
| **외부 자산 attribution** | NOTICE / README / ADR 의 attribution 명시 도달성 |

---

## 2. 사용자 액션 1 — plugin 재install

### 2.1 절차

```bash
# 1. 기존 buddy plugin 의 버전 확인
claude plugin list | grep buddy

# 2. 재install (marketplace 갱신 후 install)
claude plugin marketplace add 0xmhha/buddy
claude plugin install buddy@buddy

# 3. /reload-plugins (Claude Code 안에서)
# /reload-plugins
```

### 2.2 검증 항목

- [ ] `claude plugin list` → `buddy 1.0.8 enabled` (또는 신규 version 발견 시)
- [ ] `~/.claude/plugins/marketplaces/buddy/plugin/skills/` 안에 *148 디렉토리* 존재
- [ ] `~/.claude/plugins/marketplaces/buddy/plugin/commands/` 안에 *99 .md 파일* 존재

> **참고**: buddy repo 가 *push 안 된 상태* 면 marketplace fetch 시 *기존 v1.0.8* 만 보임. 본 검증은 *push 후* 의미 있음.

---

## 3. 사용자 액션 2 — skill discovery 검증

### 3.1 신규 commands 자동완성 테스트

Claude Code session 안에서 `/buddy:` 입력 → 자동완성 list 검증:

```
/buddy:de    → 자동완성에 다음 신규 모두 표시 되어야:
  - decide-form-factor-app-vs-web
  - decide-target-market
  - deprecate-feature
  - design-accessibility-baseline
  - design-i18n-strategy
  - design-interaction-pattern
  - design-observability
  - design-secret-management

/buddy:an    → 자동완성:
  - analyze-actor-failure-rate
  - analyze-competition-and-substitutes
  - analyze-cost-anomaly
  - analyze-customer-feedback-corpus
  - analyze-feature-adoption
  - analyze-market-size
  - analyze-user-cohort
```

### 3.2 검증 항목

- [ ] 신규 44 commands 모두 자동완성 등장
- [ ] frontmatter `disable-model-invocation: true` 활성 (description 이 baseline 에 안 등장)
- [ ] `/reload-plugins` 후 system-reminder 의 available-skills 에 `buddy:router` 만 (44 신규 description 미등장)

### 3.3 Canned business scenarios (forcing-question dialogue 입력용)

본 doc 의 §4-§5 single-skill / cascade 검증에서 *진짜 founder 응답* 을 모사할 때 사용 가능한 canned idea 모음. mock 한계 (anti-sycophancy 진단의 valid 가치 zero) 는 [`docs/notes/2026-05-10-dogfood-result-cycle-1.md`](./2026-05-10-dogfood-result-cycle-1.md) §3 참조 — 본 표는 *PROCEDURE 본문 + 분기 동작* 검증에 한해 활용.

| ID | Idea statement (1 문장) | Target user | Status quo | 1차 wedge |
|----|-------------------------|-------------|-----------|-----------|
| S1 | 한국 거주 외국인 전문직(엔지니어/디자이너) 대상 LLM 1:1 한국어 튜터링 + 회사 도메인 컨텍스트 학습 SaaS | 외국인 시니어 엔지니어 / PM | Papago 번역 + 동료 손빌림 + 사내 위키 | 이메일·미팅 한국어 교정 chrome extension |
| S2 | 한국 SMB 회계 자동화 — 영수증/은행거래/세금계산서 OCR + 회계처리 매핑 | 5~50인 SMB 대표 + 외주 회계사 | 엑셀 + 손입력 + 분기말 회계사 정정 | 영수증 OCR → 분개 자동 매핑 |
| S3 | 글로벌 SaaS 의 i18n release 자동화 — 코드 변경 → 번역 누락 / fallback 누수 lint + PR 차단 | i18n 책임 SWE / DevRel | 번역 누락 prod buggy + 분기 audit | i18n lint CI hook (Stage 1: en/ko 만) |

> S1/S2 는 region-bound (한국 cluster) — `decide-target-market` 의 Korea cluster trigger 검증에 사용.
> S3 은 region-agnostic — 글로벌 default cluster 검증에 사용.

---

## 4. 사용자 액션 3 — 단일 skill 호출 테스트

### 4.1 테스트 1: `decide-target-market` (Batch 1, foundation)

```
/buddy:decide-target-market "내 SaaS 의 target market 결정"
```

**기대 결과**:
- router → decide-target-market PROCEDURE.md load
- Stage 1 (market matrix) ~ Stage 4 (ADR) 순서 진행
- 산출 형식 §5 따름 (matrix + decision + cluster trigger + ADR reference)
- self-check §6 6 항목 자가 검증

### 4.2 테스트 2: `review-legal-regulatory` (Batch 2, region cluster trigger)

```
/buddy:review-legal-regulatory "내 SaaS 글로벌 default 가정"
```

**기대 결과**:
- 7 sub-domain (privacy / IP / AI / 약관 / 결제 / 산업 / audit) 검토
- 3 기존 sub-skill (`review-privacy-data-risk` / `review-license-and-ip-risk` / `review-ai-safety-liability`) cascade
- region cluster trigger = "글로벌 default" 명시 (Korea cluster 활성화 X)
- 약관 / 결제 inline 검토 표시
- evidence package 산출

### 4.3 테스트 3: `audit-test-coverage-meaningful` (Batch 5, agent-evaluation 차용)

```
/buddy:audit-test-coverage-meaningful "test suite trust score"
```

**기대 결과**:
- 4 차원 trust score (line 0.2 + mutation 0.4 + behavior 0.2 + edge case 0.2)
- agent-evaluation OMAS v2 pattern reference (NOTICE attribution 정합)
- mutation testing 도구 (Stryker / mutmut / go-mutesting) 추천
- survived mutant 분석 → `generate-tests-from-spec` cascade

### 4.4 테스트 4: `optimize-conversion-funnel` (Batch 6b, marketingskills 5 통합)

```
/buddy:optimize-conversion-funnel "AARRR funnel 분기 review"
```

**기대 결과**:
- AARRR 5 단계 (acquisition / activation / retention / revenue / referral)
- 5 CRO sub-domain (onboarding / form / page / paywall / popup) 통합 적용
- biggest-drop bottleneck 식별 + A/B test pipeline
- marketingskills attribution reference

### 4.5 테스트 5: `archive-product` (Batch 7, lifecycle)

```
/buddy:archive-product "product EOL plan"
```

**기대 결과**:
- 6~12 month timeline (단계 별)
- migrate-customers cascade
- data export 3 format (GDPR Article 20)
- service shutdown 6 phase
- legal compliance 5 영역
- knowledge preservation → persist-learning-jsonl

---

## 5. 사용자 액션 4 — cascade 검증

### 5.1 cascade 시나리오: §1 idea → §3 design

```
1. /buddy:concretize-idea "<idea>"
   → 산출: PRD draft

2. /buddy:decide-target-market (Batch 1, 신규)
   → 산출: target market 결정 + region cluster trigger

3. /buddy:review-legal-regulatory (Batch 2, 신규)
   → 산출: legal review + cluster trigger 결과 (글로벌 / Korea / etc)

4. /buddy:design-system "<scope>"
   → 산출: tech stack + data model + API contract + ADR
```

### 5.2 검증 항목

- [ ] 각 stage 산출이 다음 stage 입력으로 *명확하게* 전달
- [ ] decide-target-market 의 region cluster trigger 가 review-legal-regulatory 에 정합
- [ ] 자연어 dispatch (router 가 *어느 skill* 호출할지 결정) 정확

---

## 6. 사용자 액션 5 — attribution 도달성

### 6.1 NOTICE 검증

`~/.claude/plugins/marketplaces/buddy/plugin/NOTICE` (또는 buddy repo 의 NOTICE) 안에 다음 entries:

- [ ] mattpocock/skills (Matt Pocock, MIT)
- [ ] gstack (Garry Tan, MIT)
- [ ] marketingskills (Corey Haines, MIT)
- [ ] designer-skills (MC Dean, MIT)
- [ ] gpt-researcher (referenced)
- [ ] agent-evaluation (Kevin + Claude, MIT)
- [ ] humanizer (Siqi Chen, MIT)
- [ ] superpowers (TBD — ADR-003 따라 작성 예정)

### 6.2 README 검증

- [ ] Two tracks 섹션 (plugin buddy / cli buddy 정의)
- [ ] Features 표 (148 skills / 99 commands)
- [ ] Acknowledgments 섹션 (외부 차용 source)

---

## 7. 검증 결과 보고 양식

사용자가 위 시나리오 진행 후 발견된 이슈 보고:

```markdown
## Dogfood Validation Result — 2026-05-XX

### Discovery
- 자동완성 등장 신규 commands: {N/44}
- 누락: ...

### Routing
- router → 신규 skill dispatch 정확도: ...
- 자연어 dispatch 부정확 case: ...

### PROCEDURE 정합
- 12 section 양식 누락 skill: ...
- self-check (§6) 동작 안 한 skill: ...

### Cascade
- {skill A → B} cascade 정합: ...
- 부정합 case: ...

### Attribution 도달성
- NOTICE entries 도달성: pass / fail
- README Acknowledgments: pass / fail

### Follow-up
- bug fix 후보: ...
- PROCEDURE 보강 후보: ...
```

→ 발견된 issue 의 *severity 분류* (critical / major / minor) + *fix 우선순위*. 본 결과가 *다음 cycle 의 input*.

---

## 8. 다음 액션

1. 사용자가 `git push origin main` 실행 (origin 동기화 필요)
2. 사용자 Claude Code session 에서 `claude plugin install buddy@buddy` (또는 `/reload-plugins`)
3. §3-§6 의 시나리오 5 진행
4. §7 양식으로 결과 정리 → 다음 응답에서 review
5. 발견된 issue 의 fix cycle 시작 (별도 작업)

본 검증은 *Skill Completion Cycle 의 quality gate* . 통과 시 cycle 공식 종료. 미통과 시 *fix iteration* 진입.
