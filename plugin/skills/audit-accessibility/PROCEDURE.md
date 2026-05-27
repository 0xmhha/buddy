# audit-accessibility — WCAG 2.1 AA + axe + Lighthouse 통합 a11y 감사

§6 verify-quality phase 의 stage. **prepare-launch-checklist Engineering / Legal axis 의 evidence**. public 제품 의 a11y compliance (WCAG 2.1 AA) 가 의무에 가까움 (ADA / EAA / EU Accessibility Act, 산업별 차이). axe (rule-based) + Lighthouse (perceived) + manual screen reader pass 통합. 산출물은 violation matrix + fix priority + WCAG conformance level + acceptance gate.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| UI (배포 또는 로컬) | ✅ | artifact | 배포된 UI 또는 로컬 dev 서버 | "감사할 UI의 URL을 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Accessibility audit report (WCAG 2.1 AA + axe) | artifact | structured report | `iterate-fix-verify` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Lighthouse 점수만 보기** — 점수 95+ 라도 critical violation 존재 가능. axe rule-based + manual SR pass 동시.
- **Automated 만 (manual SR pass 없음)** — automated tool 은 30-50% violation 만 감지. NVDA / VoiceOver / TalkBack 실제 navigation 의무.
- **Color contrast 만** — WCAG AA 의 4 principle (Perceivable / Operable / Understandable / Robust) 4 level — color contrast 는 1 success criterion.
- **Mobile / responsive 누락** — desktop 만 audit. mobile breakpoint 의 touch target 44px / orientation / pinch zoom 별도.
- **Focus order 무검증** — keyboard navigation 만으로 끝, focus order logical 검증 안 함.
- **Aria 무남용** — `aria-*` 추가가 native HTML 보다 우선 → Robust principle 위반.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

a11y compliance 정량 측정 + violation 식별 + WCAG conformance level 보고. legal exposure (ADA lawsuit / EAA fine / EU AA enforcement 2025+) 회피.

## 2. 사용 시점 (When to invoke)

- launch 직전 (필수, prepare-launch-checklist Engineering / Legal axis)
- 신규 UI page / component 추가 후
- design system 변경 후 (color token / typography 변경)
- compliance audit 의무 (정부 / 교육 / healthcare)
- legal complaint / customer report 후 보강

## 3. 입력 (Inputs)

### 필수
- staging URL (live UI)
- §3 design 의 design system / DESIGN.md (색상 / typography / motion)
- target conformance level (A / AA / AAA — 보통 AA)
- target locale (ko / en / etc., RTL 여부)

### 선택
- 이전 audit 의 violation count (regression 비교)
- legal jurisdiction (US ADA / EU EAA / KR 장애인차별금지법)
- assistive tech 분포 (NVDA % / JAWS % / VoiceOver % — Stack Overflow Survey 보조)

### 입력이 부족할 때 forcing question
- "target conformance level 결정됐나? AA 가 industry standard, AAA 는 정부/공공 한정."
- "manual screen reader pass 가능한 인력? 인력 0 시 외부 vendor (Deque / TPGi) consulting."
- "RTL locale (Arabic / Hebrew) 지원? 미지원 시 본 audit scope 명시 (LTR only)."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — pass / fail / conformance level 단호. "거의 통과" 거부.
- **사용자 입력을 challenge** — "a11y OK" 발화에 "Lighthouse score? axe critical count? SR pass?" push.
- **Specificity 강제** — vague "good a11y" 거부, "axe 0 critical / 2 serious / Lighthouse 96 / WCAG 2.1 AA conformance, NVDA + VoiceOver pass on 5 page".
- **Assumed-malicious legal threat bias** — 가정: legal complaint 가능. 매 violation 의 ADA settlement cost ≈ $20-50k.

도메인 원칙:

1. **WCAG 2.1 AA 4 principle 강제** — Perceivable / Operable / Understandable / Robust 모두 cover.
2. **Tool stack 3 layer** — axe (rule), Lighthouse (perceived), manual SR (실 user).
3. **Mobile + desktop 둘 다** — responsive breakpoint 별 audit.
4. **Critical / Serious / Moderate / Minor 분류** — Critical/Serious 는 launch blocker, Moderate/Minor 는 post-launch.
5. **WCAG conformance level 명시** — A / AA / AAA 중 어느 level 통과.

## 5. 단계 (Phases)

### Phase 1. Tool stack + scope

| Tool | Layer | 자동화 | Coverage |
|------|-------|--------|----------|
| **axe-core / axe-playwright** | rule-based | 자동 | ~30-40% violation 감지, low false-positive |
| **Lighthouse CI** | perceived UX | 자동 | a11y score 0-100, summary view |
| **WAVE (browser ext)** | manual review | 반자동 | visual + structural |
| **NVDA (Windows)** | manual SR | 수동 | 실제 SR navigation |
| **VoiceOver (macOS / iOS)** | manual SR | 수동 | macOS + mobile Safari |
| **TalkBack (Android)** | manual SR | 수동 | mobile Chrome |
| **Playwright keyboard navigation** | manual | 자동 | Tab order + focus visible |
| **pa11y-ci** | rule-based + perceived | 자동 | CI integration |

scope:
- 5 page (signup / login / verify / reset / me) × 2 breakpoint (mobile 375px / desktop 1280px) = 10 page-state
- locale: en (primary) + ko (secondary)
- assistive tech: NVDA + VoiceOver (TalkBack 보조 — mobile TBD)

### Phase 2. WCAG 2.1 AA 4 principle audit

#### Perceivable (정보 + UI 가 user 인지 가능)

| Success Criterion | Test | Status |
|-------------------|------|--------|
| 1.1.1 Non-text content (alt text) | axe rule `image-alt` | pass / fail |
| 1.3.1 Info and relationships (semantic HTML) | axe `aria-roles` + manual review | pass / fail |
| 1.4.3 Contrast (minimum 4.5:1 normal / 3:1 large) | axe `color-contrast` + Lighthouse | pass / fail |
| 1.4.10 Reflow (320px no horizontal scroll) | manual responsive | pass / fail |
| 1.4.11 Non-text contrast (UI components 3:1) | manual review | pass / fail |
| 1.4.12 Text spacing (override stylesheet) | manual override test | pass / fail |

#### Operable (UI navigable + 사용 가능)

| Success Criterion | Test | Status |
|-------------------|------|--------|
| 2.1.1 Keyboard | Playwright keyboard navigation | pass / fail |
| 2.1.2 No keyboard trap | manual Tab navigation | pass / fail |
| 2.4.3 Focus order | manual visual + Playwright | pass / fail |
| 2.4.7 Focus visible | manual + axe `focus-visible` | pass / fail |
| 2.5.5 Target size (44×44px touch) | mobile breakpoint manual | pass / fail |
| 2.5.8 Target size minimum (24×24, WCAG 2.2) | mobile breakpoint | pass / fail |

#### Understandable (UI + content predictable)

| Success Criterion | Test | Status |
|-------------------|------|--------|
| 3.1.1 Language of page (`lang` attr) | axe rule | pass / fail |
| 3.2.1 On focus (no unexpected change) | manual focus test | pass / fail |
| 3.3.1 Error identification | manual form error | pass / fail |
| 3.3.3 Error suggestion | manual form review | pass / fail |
| 3.3.4 Error prevention (legal/financial) | manual confirm flow | pass / fail |

#### Robust (assistive tech 호환)

| Success Criterion | Test | Status |
|-------------------|------|--------|
| 4.1.1 Parsing (valid HTML) | W3C validator + axe | pass / fail |
| 4.1.2 Name / role / value (ARIA) | axe `aria-*` rules | pass / fail |
| 4.1.3 Status messages (live region) | manual `aria-live` review | pass / fail |

### Phase 3. Manual screen reader pass

scenario × SR matrix:

| Scenario | NVDA (Win) | VoiceOver (macOS) | TalkBack (Android) |
|----------|------------|---------------------|----------------------|
| signup form fill + submit | pass / fail | pass | pass |
| login + error announce | pass | pass | pass |
| verify-email landing | pass | pass | pass |
| password reset flow | pass | pass | pass |
| /me profile edit | pass | pass | pass |

각 SR 의 announce 정확도 (label / role / state / error) + navigation order 확인.

### Phase 4. Violation 분류 + priority

| Severity | Definition | Launch impact | Fix SLA |
|----------|------------|---------------|---------|
| **Critical** | WCAG A level 위반 OR critical user task 차단 (예: form submit 불가능 with SR) | **launch blocker** | 24-72h |
| **Serious** | WCAG AA level 위반 OR major UX degradation | **launch blocker (or conditional with hotfix ETA)** | 1 week |
| **Moderate** | WCAG AA level 부분 위반 (예: focus visible weak) | post-launch backlog | next sprint |
| **Minor** | WCAG AAA level OR cosmetic | post-launch backlog | when convenient |

본 SaaS auth example violation report:

| Page | Violation | Severity | WCAG | Tool |
|------|-----------|----------|------|------|
| /signup | password input aria-describedby missing | **moderate** | 1.3.1 | axe |
| /reset-password/confirm | i18n token round-trip → focus lost | **moderate** | 2.4.3 | manual |
| /me | delete-me modal focus trap broken | **serious** | 2.1.2 | NVDA |
| (all pages) | Lighthouse score 92-95 (target 95) | **moderate** | overall | Lighthouse |

### Phase 5. WCAG conformance level + acceptance gate

WCAG conformance:
- **A**: 25/25 A criteria pass — Yes/No
- **AA**: 38/38 AA criteria pass (cumulative A + 13 AA) — Yes/No
- **AAA**: 61/61 AAA criteria pass (target only for high-stakes — government, healthcare)

본 example: A pass 25/25, AA pass 36/38 (2 fail: 1 moderate + 1 serious) → **AA partial conformance (94.7%)**.

acceptance gate:
- **pass**: A 100% + AA 100%, Lighthouse ≥ 95, axe critical+serious 0, manual SR 5/5 scenario pass
- **conditional pass**: 1-2 serious with hotfix ETA ≤ 1 week + workaround documented + monitoring
- **fail**: critical violation OR AA conformance < 95% OR launch-blocking SR scenario fail

## 6. 산출물 형식 (Output format)

```markdown
## audit-accessibility Output — <project name> v<version>

### Summary
<3 줄: tool stack / page count / WCAG level / violation count by severity / decision>

### Audit Scope
| Field | Value |
|-------|-------|
| Pages | 5 (signup / login / verify / reset / me) |
| Breakpoints | 2 (mobile 375px / desktop 1280px) |
| Locales | en + ko |
| Assistive tech | NVDA + VoiceOver |
| Target conformance | WCAG 2.1 AA |

### WCAG 2.1 4 Principle Results
| Principle | A | AA | AAA |
|-----------|---|-----|------|
| 1. Perceivable | 9/9 ✓ | 6/7 (1 fail) | 4/8 |
| 2. Operable | 8/8 ✓ | 5/6 (1 fail) | 3/7 |
| 3. Understandable | 5/5 ✓ | 5/5 ✓ | 3/6 |
| 4. Robust | 3/3 ✓ | 3/3 ✓ | 0/0 |
| **Total** | **25/25 ✓** | **19/21 (94%)** | **10/21** |

### Manual Screen Reader Results
| Scenario | NVDA | VoiceOver | Status |
|----------|------|-----------|--------|
| signup | pass | pass | green |
| login + error | pass | pass | green |
| verify-email | pass | pass | green |
| password reset | pass | pass | green |
| /me + delete-me | **fail** (focus trap) | fail | **serious — launch blocker** |

### Violation Report
| # | Page | Description | Severity | WCAG | Tool | Fix |
|---|------|-------------|----------|------|------|-----|
| V-001 | /me | delete-me modal focus trap broken (Esc 비활성) | serious | 2.1.2 | NVDA | implement focus trap with `react-focus-lock`, ETA 24h |
| V-002 | /signup | password aria-describedby 누락 | moderate | 1.3.1 | axe | post-launch backlog |
| V-003 | /reset-password/confirm | i18n switch 시 focus lost | moderate | 2.4.3 | manual | post-launch backlog (UAT-BUG-003 dup) |
| V-004 | (all) | Lighthouse score 92-95 (target 95) | moderate | overall | Lighthouse | tracking, target 95 enforce in CI |

### Acceptance Gate
| Field | Value |
|-------|-------|
| Decision | conditional pass |
| Rationale | A 25/25 + AA 19/21 (94%) — 1 serious (V-001 focus trap) with hotfix ETA 24h. moderate 3 → post-launch backlog. |
| Conformance level | WCAG 2.1 AA (94%) → AA full after V-001 fix |
| Hotfix ETA | V-001 within 24h before GA |

### Cascade
- **§7 prepare-launch-checklist**: Engineering (a11y) + Legal axis row 입력
- **§7 setup-incident-paging**: V-001 fix 후 alert routing 검증
- **§8 monitor-regressions**: Lighthouse score baseline = production regression detection

### Next Step
<구체 action — 1줄: 예 "V-001 focus trap fix 24h 내 + audit-cost-efficiency 진입">
```

## 7. Cross-phase cascade

- **§7 prepare-launch-checklist**: Engineering + Legal axis evidence
- **§8 monitor-regressions**: Lighthouse score baseline
- **§3 design-accessibility-baseline** (deferred): 본 skill 의 violation pattern → design baseline 보강

## 8. 다음 skill (next in stage flow)

- `audit-cost-efficiency` (§6 Cluster A) — economics 정량
- `prepare-launch-checklist` (§7) — 통합

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `run-browser-qa`** (§6) — 그것은 일반 UI QA (visual regression / form), 본 skill 은 a11y 전문. 보완.
- **vs `audit-live-devex`** (§6) — 그것은 developer-facing product 의 DX, 본 skill 은 end-user a11y. 다른 audience.
- **vs `review-terms-policy-readiness`** (§6) — 그것은 ToS / privacy policy, 본 skill 은 product UI a11y. legal 다른 영역.
- **vs `design-accessibility-baseline`** (deferred) — 그것은 design 단계 baseline (색 / focus / motion), 본 skill 은 implementation audit. 본 skill 의 violation 이 baseline 보강 input.

## 10. 중요 규칙

- **WCAG 4 principle 의무** — single principle audit 거부.
- **3 layer tool stack 의무** — axe + Lighthouse + manual SR.
- **Mobile + desktop 둘 다** — desktop only 거부.
- **Severity 4 분류** — vague 거부.
- **Conformance level 명시** — A / AA / AAA 중 어느 level.
- **Read-only on production** — staging 만 audit.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 7 출력 섹션 모두 채워짐
- [ ] WCAG 4 principle 모두 audit + cumulative result
- [ ] Manual SR scenario 모두 pass / fail 명시
- [ ] Violation Report severity 4 분류 + WCAG SC + tool + fix ETA
- [ ] Conformance level 명시 (A / AA / AAA + %)
- [ ] Acceptance Gate decision 단일 + rationale
- [ ] §4 posture — 정량 단호
- [ ] §0 anti-pattern 부재 — Lighthouse 점수만 / automated 만 / color only / mobile 누락 / focus order / aria 남용 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
