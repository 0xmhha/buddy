# Skill Matrix 3a — Group 2 + Group 4 × External 9 Candidates

> **Step 3a 산출**: [`docs/notes/2026-05-10-missing-skills-inventory.md`](./2026-05-10-missing-skills-inventory.md) 의 그룹 2 (신규 1: `review-legal-regulatory`) + 그룹 4 (4 영역: form-factor / design / 그로스 / 마케팅) × [`docs/notes/2026-05-10-external-skills-inventory.md`](./2026-05-10-external-skills-inventory.md) §7 의 우선 9 외부 후보 매트릭스.
>
> **사용자 발화 (2026-05-10)**: "이 수많은 skill 이 겹치는 부분이 너무 많고, 조금씩 아쉬운 부분이 존재" + "스킬 능력이 오히려 떨어질 수 있는 부분도 존재하니 무조건 가져올 것이 아니라, 우리의 'buddy' 프로젝트 스킬들과 호환 및 궁합이 좋은지 점검이 필요"
>
> 본 문서는 *호환성 점검 + 매핑 결정 + 그룹 4 명명 제안* 까지. **신규 발견**: marketingskills + designer-skills 가 *그룹 1* 영역까지 침투 — Step 3b (그룹 1 29 skill 매트릭스) 의 직접 입력으로 활용.

---

## 1. 검토 방법

### 1.1 검토 dimension (per 후보)

| dimension | 측정 |
|----------|-----|
| **목적 정합** | buddy missing-skill 의 책임 범위와 외부 후보의 use case 일치도 |
| **호환성** | Claude Code skill spec (frontmatter `description` / `argument-hint` / `disable-model-invocation` / SKILL.md 양식) 준수 |
| **명명 컨벤션** | buddy 의 *동사+명사* / *kebab-case* 컨벤션 정합 |
| **scope 정확도** | 단일 책임 (buddy PROCEDURE.md 양식) vs 광범위 (외부는 종종 다중 책임) |
| **언어 / 지역 종속** | 한국어 / 영어 / 지역 특화 vs buddy 의 multi-locale 지향 |

### 1.2 4 결정 옵션

| 옵션 | 내용 |
|------|------|
| **(a) 그대로 차용** | PROCEDURE.md 본문 그대로 + frontmatter 만 buddy 표준화 |
| **(b) 수정 차용** | 외부 자산을 입력으로 buddy scope 정합 / 명명 정합 / multi-locale 지원 추가 후 작성 |
| **(c) 참고만 + 신규 작성** | 외부 자산은 reference 만, buddy 는 처음부터 신규 |
| **(d) 무관 + 신규 작성** | 외부 자산 매칭 없음 / 차원 다름 — 신규 작성 |

---

## 2. 그룹 2 매핑 — `review-legal-regulatory` × 한국 법률 3

### 2.1 외부 후보 3

| 후보 | scope | 언어 / 지역 | 호환성 |
|------|------|-----------|------|
| `ai-professional-replacement-legal-exploration_skill` | AI 의 한국 전문직 (변리사 / 변호사 / 의사 등) 업무 대체 가능 여부 탐색 | **한국어 / 한국 시장** | Claude SKILL.md 양식 호환 |
| `korean-legal-guide_skill` | 대한민국 모든 법률 분야 (특허 / 저작권 / 노동 / 민사 / 형사 / 부동산 / 세금) 일반인 눈높이 답변 | **한국어 / 한국 시장** | 호환 |
| `patent-application-drafting_skill` | 한국 특허 출원 초안 작성 (KIPO 형식) | **한국어 / 한국 시장 / 특정 use case** | 호환 |

### 2.2 buddy `review-legal-regulatory` 의 scope (skill-map.md §9 정의 인용)

> "법률 검토 / 개인정보 / PII 처리 검토 / AI 서비스 책임 범위 / 저작권·라이선스 검토 / 약관·개인정보처리방침 readiness / 결제·환불·소비자보호 / 산업별 규제"
>
> 즉 *제품 빌딩 시 사업성 / 운영 영역 법률·규제 영향 검토* — checklist + evidence package 작성 (전문가 검토 의뢰 입력).

### 2.3 매핑 결과

| 외부 후보 | buddy scope 정합 | 차용 결정 |
|---------|--------------|---------|
| `ai-professional-replacement-legal-exploration_skill` | ❌ — *AI 가 인간 전문직 대체* 라는 *use case 종속*. buddy 의 *제품 빌딩 법률 검토* 와 차원 다름 | **(c) 참고만** |
| `korean-legal-guide_skill` | △ — 한국 시장 진출 시 *지역 특화 보충 reference* 가치. 단 buddy *일반 review-legal-regulatory* 의 main scope 아님 | **(c) 참고만** + *deferred* `consult-korean-legal-context` 후보 |
| `patent-application-drafting_skill` | △ — 특허 영역만. buddy 의 일반 법률 검토 후 *특허 출원 단계* 진입 시 plugin buddy 의 별도 stage 후보 | **(c) 참고만** + *deferred* `draft-patent-application` 후보 |

### 2.4 결정

`review-legal-regulatory` = **(d) 신규 작성** (외부 직접 차용 없음). 단 한국 시장 진출 시 추가 영역 skill 2개 *deferred*:
- `consult-korean-legal-context` (한국 법률 일반 영역, korean-legal-guide_skill 입력 활용)
- `draft-patent-application` (특허 영역, patent-application-drafting_skill 입력 활용)

→ 위 2 deferred 는 *그룹 4-extension* 으로 등록 (charter scope 외 *지역 특화* 영역).

---

## 3. 그룹 4 stage 11 마케팅 매핑 — `marketingskills` (1 후보 = 30+ skills)

### 3.1 marketingskills 의 30+ skill 분류

buddy charter scope 매핑 + 기존 buddy skill 충돌 표:

| marketingskills 항목 | buddy 매칭 영역 | 결과 |
|------------------|-------------|------|
| **경쟁 / 시장** | | |
| competitor-alternatives | 그룹 1 §1 `analyze-competition-and-substitutes` ⭐ | **그룹 1 매칭** (Step 3b 입력) |
| competitor-profiling | 그룹 1 §1 `analyze-competition-and-substitutes` ⭐ | **그룹 1 매칭** |
| customer-research | 그룹 1 §1 `conduct-customer-interview` ⭐ | **그룹 1 매칭** |
| pricing-strategy | 기존 `review-pricing-and-gtm` (구현됨) | *흡수 또는 PROCEDURE 갱신* |
| **AB / 분석 / 그로스** | | |
| ab-test-setup | 기존 `design-ab-experiment` (구현됨) | *흡수* |
| analytics-tracking | 그룹 1 §8 `analyze-user-cohort` 등 | **그룹 1 매칭** |
| churn-prevention | 그룹 1 §8 `analyze-actor-failure-rate` 영역 (cohort + churn) | **그룹 1 매칭** |
| onboarding-cro | **그룹 4 stage 10 그로스** ⭐ | 그룹 4 후보 |
| form-cro | **그룹 4 stage 10 그로스** ⭐ | 그룹 4 후보 |
| page-cro | **그룹 4 stage 10 그로스** | 그룹 4 후보 |
| paywall-upgrade-cro | **그룹 4 stage 10 그로스** | 그룹 4 후보 |
| popup-cro | **그룹 4 stage 10 그로스** | 그룹 4 후보 |
| **콘텐츠 / 메시지** | | |
| copywriting | **그룹 4 stage 11 마케팅** ⭐ | 그룹 4 후보 |
| copy-editing | 그룹 4 stage 11 (copywriting 의 보조) | 그룹 4 후보 (sub) |
| content-strategy | 그룹 4 stage 11 마케팅 | 그룹 4 후보 |
| email-sequence | 그룹 4 stage 11 (cold-email 과 묶음) | 그룹 4 후보 |
| cold-email | 그룹 4 stage 11 마케팅 | 그룹 4 후보 |
| **채널 / 전술** | | |
| ai-seo | 그룹 4 stage 11 (audit-seo 후보) | 그룹 4 후보 |
| aso-audit | 그룹 4 stage 11 (모바일 앱 store optimization) | 그룹 4 후보 (form-factor 의존) |
| paid-ads | 그룹 4 stage 11 마케팅 | 그룹 4 후보 |
| ad-creative | 그룹 4 stage 11 (ads 보조) | 그룹 4 후보 (sub) |
| co-marketing | 그룹 4 stage 11 마케팅 | 그룹 4 후보 |
| community-marketing | 그룹 4 stage 11 마케팅 | 그룹 4 후보 |
| directory-submissions | 그룹 4 stage 11 (구체적 launch tactic) | 그룹 4 후보 (sub) |
| **launch / 전략** | | |
| launch-strategy | 기존 `prepare-launch-checklist` (구현됨) | *흡수* |
| marketing-ideas | 그룹 4 stage 11 (메타 — ideation) | 그룹 4 후보 |
| marketing-psychology | 그룹 4 stage 11 (도메인 지식) | 그룹 4 후보 (reference) |
| **lead generation** | | |
| lead-magnets | 그룹 4 stage 11 마케팅 | 그룹 4 후보 |
| free-tool-strategy | 그룹 4 stage 11 마케팅 | 그룹 4 후보 |
| image | (보조 자산 — image generation) | (charter 외) |

### 3.2 분포 요약

| 영역 | marketingskills 항목 |
|------|------------------|
| 그룹 1 §1 (경쟁 / 고객) | 3 (competitor-alternatives, competitor-profiling, customer-research) |
| 그룹 1 §8 (분석) | 2 (analytics-tracking, churn-prevention) |
| 기존 skill 흡수 | 3 (pricing-strategy → review-pricing-and-gtm, ab-test-setup → design-ab-experiment, launch-strategy → prepare-launch-checklist) |
| **그룹 4 stage 10 그로스 (CRO 군)** | 5 (onboarding-cro, form-cro, page-cro, paywall-upgrade-cro, popup-cro) |
| **그룹 4 stage 11 마케팅** | 13 (copywriting / content-strategy / email-sequence / cold-email / ai-seo / aso-audit / paid-ads / co-marketing / community-marketing / lead-magnets / free-tool-strategy / marketing-ideas / 기타 sub) |
| 보조 / 도메인 reference | 4 (ad-creative, copy-editing, directory-submissions, marketing-psychology, image) |

### 3.3 호환성 점검 — marketingskills

| dimension | 평가 |
|----------|------|
| 목적 정합 | 🟢 — Claude Code agentskills 표준 따름 (README 명시) |
| 호환성 (Claude skill spec) | 🟢 — agentskills.io 호환 |
| 명명 컨벤션 | 🟡 — kebab-case OK, 단 buddy *동사+명사* 컨벤션과 일부 차이 (e.g. `pricing-strategy` 는 명사형) |
| scope 정확도 | 🟢 — 각 skill 단일 책임 |
| 언어 종속 | 🟢 — 영어 (buddy en/ko 카탈로그와 호환) |

→ marketingskills 는 **(b) 수정 차용** 가능 — *명명만 buddy 컨벤션 정합*, 본문은 그대로 활용.

### 3.4 결정 — 그룹 4 stage 10 + 11 통합 + 5 + 13 = 18 skill 후보

charter scope 10 (그로스 해킹) + 11 (마케팅 지원) 영역 겹침 — **통합 처리**.

권장 *5 핵심 skill* (외부 18 후보 중 응축):

| 권장 buddy skill 명 | 외부 marketingskills 입력 | 책임 |
|------------------|---------------------|------|
| `optimize-conversion-funnel` | onboarding-cro + form-cro + page-cro + paywall-upgrade-cro + popup-cro (5 통합) | 전환율 최적화 — funnel stage 별 CRO |
| `plan-growth-experiment` | (marketingskills 직접 후보 없음 — design-ab-experiment 와 묶음) | growth experiment 설계 |
| `draft-marketing-copy` | copywriting + copy-editing + ad-creative (3 통합) | 마케팅 카피 작성 / 편집 |
| `plan-marketing-channel` | paid-ads + co-marketing + community-marketing + directory-submissions + lead-magnets + free-tool-strategy (6 통합) | 마케팅 채널 전략 |
| `audit-seo-aso` | ai-seo + aso-audit (2 통합) | SEO + ASO 감사 |
| `automate-marketing-content` | email-sequence + cold-email + content-strategy (3 통합) | 마케팅 콘텐츠 자동화 |

→ **그룹 4 stage 10 + 11 통합 결과: 6 신규 skill** (외부 18 항목을 응축).

---

## 4. 그룹 4 stage 4 디자인 적용 매핑 — designer-skills + 5

### 4.1 외부 후보 6 호환성 sweep

| 후보 | spec | 호환성 | 차용 가치 |
|------|------|------|--------|
| `designer-skills` | Claude Code 87 skills + 27 commands + 8 plugins | 🟢 직접 호환 | **HIGH** ⭐ |
| `make-interfaces-feel-better` | Agent Skill (Claude Code) | 🟢 직접 호환 | MEDIUM (단일 skill) |
| `ui-design-brain` | Cursor 전용 | ❌ 비호환 | LOW (참고 reference 만) |
| `ui-ux-pro-max-skill` | skill.json (다른 표준) | 🟡 부분 | LOW |
| `Star-Office-UI` | Electron 제품 (skill 아님) | ❌ N/A | LOW (charter 외) |
| `better-icons` | MCP server | 🟡 그룹 3 영역 | (그룹 3 deferred 와 묶음) |

### 4.2 designer-skills 의 8 plugin 구조

| plugin | buddy 매칭 영역 |
|--------|-------------|
| `design-ops` | charter scope 4 디자인 적용 의 운영 영역 |
| `design-research` | 그룹 1 §1 (customer / market research) ⭐ |
| `design-systems` | charter scope 4 디자인 적용 ⭐ |
| `designer-toolkit` | charter scope 4 디자인 적용 (도구 영역) |
| `interaction-design` | charter scope 4 디자인 적용 (interaction layer) |
| `prototyping-testing` | charter scope 4 디자인 적용 (prototyping) + 그룹 1 §6 verify |

→ 6 plugin 중 4 plugin (design-systems / designer-toolkit / interaction-design / prototyping-testing) 이 *charter scope 4 디자인 적용* 직접 cover. design-research 는 *그룹 1 §1* 영역.

### 4.3 그룹 4 stage 4 명명 제안

designer-skills 의 87 skill 정밀 인벤토리는 *Step 3b 에서 진행*. Step 3a 시점 권장 *4 핵심 skill*:

| 권장 buddy skill 명 | 외부 designer-skills 입력 | 책임 |
|------------------|----------------------|------|
| `apply-design-system` | design-systems plugin 입력 | 기존 design system 채택 / 적용 |
| `audit-ui-quality` | make-interfaces-feel-better + designer-toolkit 일부 | UI 품질 / 디테일 검토 |
| `prototype-from-spec` | prototyping-testing plugin | 명세 → prototype 작성 |
| `design-interaction-pattern` | interaction-design plugin | interaction / motion 설계 |

→ **그룹 4 stage 4: 4 신규 skill**.

### 4.4 호환성 점검 — designer-skills

| dimension | 평가 |
|----------|------|
| 목적 정합 | 🟢 — Claude Code 87 skills + 27 commands |
| 호환성 | 🟢 — Claude Code 직접 호환 |
| 명명 컨벤션 | (Step 3b 정밀 검토 시 확인) |
| scope 정확도 | (Step 3b 정밀 검토 시 확인) |
| 언어 | 🟢 — 영어 |

→ designer-skills 는 **(b) 수정 차용** 가능. Step 3b 에서 87 skill 중 buddy 4 권장 skill 에 정합하는 항목 선별.

---

## 5. 그룹 4 stage 3 (form-factor) 매핑 — 직접 후보 부재

### 5.1 검토 결과

외부 9 후보 중 *form-factor 결정 (앱 vs web)* 직접 매칭 없음.

| 후보 | form-factor 영역 cover? |
|------|---------------------|
| 한국 법률 3 | ❌ |
| marketingskills | △ (aso-audit 가 *모바일 앱 가정* — 즉 form-factor *결정 후* 입력) |
| designer-skills | △ (design-systems 의 *responsive design* 등 — 단 form-factor *결정 자체* 가 아님) |
| 기타 | ❌ |

### 5.2 결정

`decide-form-factor-app-vs-web` = **(d) 무관 + 신규 작성**.

scope 정의 (Step 3a 신규 제안):
- input: PRD + business viability + target user device 패턴
- output: app (native iOS/Android, hybrid) vs web (SPA, MPA, PWA) 결정 + ADR
- buddy 의 §3 design-system 안 *stage 0* 으로 진입 권장 (define-tech-stack 의 *전제* 단계)

→ **그룹 4 stage 3: 1 신규 skill**.

---

## 6. 그룹 4 명명 / scope 종합 제안

### 6.1 4 신규 영역 → 11 신규 skill 제안

| stage | 신규 skill 수 | 명단 |
|-------|-----------|------|
| **3 form-factor 결정** | 1 | `decide-form-factor-app-vs-web` |
| **4 디자인 적용** | 4 | `apply-design-system`, `audit-ui-quality`, `prototype-from-spec`, `design-interaction-pattern` |
| **10 그로스 해킹 + 11 마케팅 (통합)** | 6 | `optimize-conversion-funnel`, `plan-growth-experiment`, `draft-marketing-copy`, `plan-marketing-channel`, `audit-seo-aso`, `automate-marketing-content` |
| **합계 그룹 4** | **11** | |

### 6.2 deferred 그룹 4-extension (지역 특화)

| 후보 | trigger |
|------|---------|
| `consult-korean-legal-context` | buddy 사용자가 한국 시장 진출 시 |
| `draft-patent-application` | 특허 영역 진입 시 |

---

## 7. per-skill 결정 매트릭스 (Step 3a)

### 7.1 그룹 2 (영역 겹침 결정 lock-in 됨, 이번 시점 재확인)

| skill | 결정 | 외부 입력 |
|-------|------|---------|
| `review-legal-regulatory` | (d) 신규 | 한국 법률 3 = 참고만 |

### 7.2 그룹 4 (이번 Step 3a 결정)

| skill | 결정 | 외부 입력 |
|-------|------|---------|
| `decide-form-factor-app-vs-web` | (d) 신규 | (외부 후보 없음) |
| `apply-design-system` | (b) 수정 차용 | designer-skills/design-systems plugin |
| `audit-ui-quality` | (b) 수정 차용 | make-interfaces-feel-better + designer-skills/designer-toolkit |
| `prototype-from-spec` | (b) 수정 차용 | designer-skills/prototyping-testing plugin |
| `design-interaction-pattern` | (b) 수정 차용 | designer-skills/interaction-design plugin |
| `optimize-conversion-funnel` | (b) 수정 차용 | marketingskills 5 CRO skill 통합 |
| `plan-growth-experiment` | (d) 신규 | (외부 직접 후보 없음) |
| `draft-marketing-copy` | (b) 수정 차용 | marketingskills copywriting + copy-editing + ad-creative |
| `plan-marketing-channel` | (b) 수정 차용 | marketingskills 6 channel skill 통합 |
| `audit-seo-aso` | (b) 수정 차용 | marketingskills ai-seo + aso-audit |
| `automate-marketing-content` | (b) 수정 차용 | marketingskills email-sequence + cold-email + content-strategy |

→ **그룹 4 11 skill 중 (b) 9 + (d) 2**. 외부 자산 활용도 높음.

---

## 8. 신규 발견 — Step 3b 입력

### 8.1 marketingskills 가 그룹 1 영역 침투

marketingskills 30+ skill 중 5 개가 *그룹 1* 영역:

| marketingskills | buddy 그룹 1 매칭 |
|------------|--------------|
| competitor-alternatives | §1 `analyze-competition-and-substitutes` (그룹 1 의 5 §1 customer/market) |
| competitor-profiling | §1 `analyze-competition-and-substitutes` |
| customer-research | §1 `conduct-customer-interview` |
| analytics-tracking | §8 데이터 분석 영역 |
| churn-prevention | §8 `analyze-actor-failure-rate` 영역 |

### 8.2 designer-skills 의 design-research plugin 이 그룹 1 영역 침투

design-research → 그룹 1 §1 (customer / market research) 보조

### 8.3 의미

Step 3b (그룹 1 29 매트릭스) 진입 시:
- marketingskills 의 *경쟁 / 고객 / 분석* 5 skill 이 그룹 1 §1 + §8 의 *직접 입력*
- designer-skills 의 design-research plugin 이 그룹 1 §1 의 *추가 입력*
- 즉 그룹 1 매트릭스가 *예상보다 풍부* — Step 3b 의 외부 자산 매칭 율 ↑

---

## 9. 호환성 / 궁합 종합 점검

| 후보 그룹 | 호환성 | 차용 정도 | 비고 |
|---------|------|--------|------|
| 한국 법률 3 | 🟡 SKILL.md 양식 호환, scope 차원 다름 | (c) 참고만 | 지역 특화 deferred 2 후보로 보존 |
| marketingskills | 🟢 agentskills.io 호환, 명명 일부 차이 | (b) 수정 차용 우세 | buddy *동사+명사* 컨벤션 적용 시 그대로 활용 |
| designer-skills | 🟢 Claude Code 87 skills 호환 | (b) 수정 차용 | 8 plugin 구조 → buddy 의 stage 단위 재구성 필요 |
| make-interfaces-feel-better | 🟢 Agent Skill 호환 | (b) 수정 차용 (sub) | 단일 skill, audit-ui-quality 의 입력 |
| ui-design-brain / ui-ux-pro-max-skill / Star-Office-UI / better-icons | ❌ / 🟡 | LOW / charter 외 | 미사용 또는 그룹 3 (better-icons MCP) |

### 9.1 명명 컨벤션 충돌

| 외부 명명 패턴 | buddy 표준 | 정정 필요 |
|------------|---------|--------|
| 명사형 (`pricing-strategy`, `marketing-psychology`) | 동사 + 명사 | 수정 차용 시 명명 변경 |
| 약어 사용 (CRO / SEO / ASO) | 풀 단어 + 약어 병기 | 본문 첫 사용 시 풀이 |
| skill 명에 `-skill` 접미사 (한국 법률 3) | 접미사 없음 | 수정 차용 시 제거 |

### 9.2 PROCEDURE.md 양식 정합성

| 외부 양식 | buddy 양식 | 변환 비용 |
|---------|---------|---------|
| 한국 법률 SKILL.md | PROCEDURE.md (Stage 흐름 / 실행 절차 / 산출물 형식 등 12 section) | **HIGH** — 단일 skill 양식 → 12 section 변환 |
| marketingskills (agentskills 양식) | PROCEDURE.md | MEDIUM — agentskills 가 buddy 의 sub-set |
| designer-skills (Claude Code 87 skill) | PROCEDURE.md | LOW — 양식 가까움 |

→ designer-skills 가 변환 비용 가장 낮음.

---

## 10. 다음 액션

### 10.1 Step 3a 종료 — commit

본 매트릭스 commit. 그룹 4 11 skill 명명 lock-in (단 사용자 confirm 필요한 항목 §10.2).

### 10.2 사용자 confirm 받을 항목 (Step 3a 결과)

| 결정 | 권장 |
|------|------|
| 그룹 4 11 skill 명명 — 권장 그대로? | **(권장)** 11 skill 명 모두 OK. 단 `optimize-conversion-funnel` / `plan-marketing-channel` 같은 통합 명이 *너무 광범위* 인지 사용자 의도 확인 |
| 그룹 4 stage 10 + 11 통합? (charter 는 *분리* 명시) | **(권장)** 통합 — 외부 자산도 통합 영역 (CRO + 콘텐츠 + 채널) |
| deferred 2 (한국 시장 진출) lock-in? | **(권장)** *deferred 후보* 로 본 매트릭스에 보존, 실제 작성은 trigger 발생 시 |

### 10.3 Step 3b 진입 — 그룹 1 29 skill 매트릭스

본 §8 의 신규 발견 (marketingskills + designer-skills 가 그룹 1 영역 침투) 을 입력으로 그룹 1 29 skill × 외부 자산 매트릭스 작성.

분량 추정: 200~300 줄.
