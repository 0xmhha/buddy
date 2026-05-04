---
name: verify-quality
description: This skill should be used when the user wants to "verify quality", "run QA", "audit security", "check code health", "review compliance", "run tests", "check before release", or has code ready and needs quality verification before release. Orchestrates §6 Quality (Test + Security + Compliance) phase.
---

# verify-quality — §6 Quality Orchestrator

§6 라이프사이클 단계의 진입점. Code complete → QA report + security/legal sign-off.

**진입 조건**: §5 code complete (모든 actor track task 완료).
**산출물**: QA report, security audit result, legal/compliance sign-off, performance baseline.
**다음 phase**: Quality gate pass → `ship-release` (§7).

---

## Stage 흐름

```
verify-quality (§6 phase orchestrator)
├── stage 1: classify-qa-tiers         (QA intensity 분류 — Quick/Standard/Exhaustive)
├── stage 2: test-per-actor-use-case   (actor별 통합 테스트)
├── stage 3: test-cross-actor-flow     (cross-actor E2E 테스트)
├── stage 4: run-browser-qa            (UI/UX + accessibility browser 테스트)
├── stage 5: audit-security            (보안 감사 — CSO 모드)
├── stage 6: measure-code-health       (code health dashboard — 0-10)
├── stage 7: monitor-regressions       (regression baseline 확인)
├── stage 8: review-ai-safety-liability    (AI 기능 책임 검토)
├── stage 9: review-privacy-data-risk      (PII/privacy 검토)
├── stage 10: review-license-and-ip-risk   (라이선스/IP 검토)
└── stage 11: review-terms-policy-readiness (ToS/Privacy Policy 준비도)
```

---

## 실행 절차

### Stage 1: QA Tier 결정

`classify-qa-tiers` skill을 invoke해 QA intensity를 결정한다:
- **Quick**: 소규모 bugfix. unit test + smoke test.
- **Standard**: 일반 feature. full unit + integration + browser.
- **Exhaustive**: 결제/인증/데이터 처리 등 critical path. + security + load + accessibility.

### Stage 2: Actor별 통합 테스트 (Use Case 기반)

§2 feature spec의 per-actor use cases를 기준으로 actor별 통합 테스트를 설계하고 실행한다:

- **user-actor** (frontend): E2E browser test (`run-browser-qa` invoke)
- **system-actor** (backend): unit + integration test (API endpoint 기준)
- **3rd-party-actor**: contract test (실제 SaaS 응답 schema 검증)

### Stage 3: Cross-Actor 흐름 테스트

§2 feature spec의 integration test_plan 기준으로 cross-actor 흐름을 테스트한다.

예시: `signup-email-password` full flow
```
form submit → backend validation → JWT issuance → email sending → click confirmation → login success
```

`run-browser-qa` skill을 invoke해 browser automation으로 전체 흐름을 검증한다.

### Stage 4: Browser QA

`run-browser-qa` skill로 UI QA를 실행한다:
- snapshot diff (시각적 regression)
- form testing (validation feedback, error states)
- responsive check (모바일/태블릿/데스크톱)
- accessibility tree interaction

### Stage 5: 보안 감사

`audit-security` skill을 invoke해 CSO-mode security audit을 수행한다.
`classify-review-risks` skill로 반복적으로 놓치는 risk 11 category를 분류한다.
`audit-live-devex` skill로 developer-facing product이면 DX audit을 수행한다.

### Stage 6: Code Health

`measure-code-health` skill을 invoke해 typecheck/lint/test/deadcode/shell 결과를 0-10 composite health dashboard로 생성한다.

목표: health score ≥ 7.0. 7.0 미만이면 release block.

### Stage 7: Regression 확인

`monitor-regressions` skill로 이전 baseline 대비 regression이 없는지 확인한다.

### Stage 8-11: Compliance Review

프로젝트 특성에 따라 선택적으로 실행:

- AI 기능 포함 → `review-ai-safety-liability`
- PII/개인정보 처리 → `review-privacy-data-risk`
- 오픈소스 의존성 → `review-license-and-ip-risk`
- 상용 출시 → `review-terms-policy-readiness`

---

## Quality Gate 판단

| 항목 | 통과 기준 | Block 조건 |
|------|---------|-----------|
| Unit + Integration test | 100% pass | 1개라도 실패 |
| E2E (cross-actor) | 100% pass | critical path 실패 |
| Security audit | No critical/high | Critical 존재 |
| Code health | ≥ 7.0 | < 7.0 |
| Accessibility | No critical a11y | WCAG 2.1 AA 위반 |
| Compliance | Reviewed + signed off | Review 미실시 |

---

## 다음 phase

- `/buddy:ship-release` — §7 Release (quality gate 통과 시)
- `/buddy:build-feature` — §5 재진입 (quality gate 실패 시)

---

## 참조

- Architecture spec: `docs/superpowers/specs/2026-05-04-lifecycle-orchestrator-architecture.md` §4 §6
