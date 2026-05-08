# test-per-actor-use-case — actor 단위 통합 테스트 실행

§6 verify-quality phase 의 stage 2. §2 feature spec 의 actor × use case 매트릭스를 기준으로 actor 단위 통합 test 를 실행. **frontend (Playwright E2E), backend (Vitest + testcontainers integration), 3rd-party (Pact contract)** 가 actor 별 적합 layer. coverage gap = use case 가 test 없는 row → 0 maintain. Q8=(a) cascade 의 §6 entry. 산출물은 actor × use case × test status 매트릭스 + gap report + acceptance gate decision.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Layer mismatch** — frontend actor 에 backend unit test 만 / backend 에 E2E browser test 만. actor × layer fit 검증 필수.
- **Use case ↔ test 매핑 모호** — test 는 있지만 어느 use case 를 cover 하는지 불명. acceptance criteria → test name 1:1 매핑 강제.
- **Gap silent skip** — test 없는 use case 발견 시 "나중에" 로 skip → coverage 가짜. gap 발견 시 §4 acceptance test plan 회귀 trigger 의무.
- **Test infra dependency 무시** — testcontainers / MSW / Pact mock / SES simulator / elasticmq / miniredis 중 하나라도 inactive 시 false positive (mock-only test pass) → 모든 infra 동시 active 검증.
- **Coverage measurement 부재** — test 통과만 보고 coverage % 측정 안 함 → critical path coverage gap 노출 안 됨.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

actor 별 use case 가 test 로 verify 되는지 확인. coverage gap 없이 acceptance gate 통과 = §6 quality 의 절반 (cross-actor 가 나머지 절반).

## 2. 사용 시점 (When to invoke)

- §6 verify-quality stage 2 (default)
- 신규 actor 추가 후 (use case × test 매트릭스 재실행)
- use case 변경 / acceptance criteria 변경 후
- `define-acceptance-test-plan` 갱신 후 plan ↔ actual test sync 검증
- regression 발생 후 actor 별 coverage gap 식별

## 3. 입력 (Inputs)

### 필수
- §2 feature spec — actor list + per-actor use case + acceptance criteria
- §4 `define-acceptance-test-plan` 산출물 — per-actor test plan (framework / scope / count est.)
- §5 build code complete — actual test file 존재
- test infra (CI runner, fixture management, mock strategy) ready

### 선택
- coverage 목표 (default: line ≥ 80%, critical path 100%)
- 이전 release 의 per-actor test report (regression 비교)
- mutation testing 결과 (Stryker 가 critical actor 에 적용된 경우)

### 입력이 부족할 때 forcing question
- "per-actor acceptance criteria 가 명시됐나? define-acceptance-test-plan 의 per-actor section 비어있으면 본 skill 실행 불가."
- "test infra 가 모두 active 인가? testcontainers + MSW + Pact + LocalStack + miniredis 중 하나라도 down 시 false positive risk."
- "coverage 목표가 정량 명시됐나? 'high coverage' 거부 — line / branch / mutation 명확."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — pass/fail/gap 단호. "거의 cover" 거부.
- **사용자 입력을 challenge** — "test 다 통과" 발화에 "어느 use case? coverage %? gap 어디?" push.
- **Specificity 강제** — vague "OK" 거부, "Vitest unit 50/50, integration 12/12, Playwright 8/8 + line 87% / critical path 100%".
- **Gap-first bias** — 통과 row 보다 gap row 우선 보고.

도메인 원칙:

1. **Actor × layer fit** — frontend → component+E2E, backend → unit+integration, 3rd-party → contract. layer mismatch 거부.
2. **Use case → test 1:1 매핑** — 모든 acceptance criteria 가 named test 와 매핑. orphan acceptance 0.
3. **Coverage 정량** — line / branch / critical path 분리 측정.
4. **Test infra 동시 active 검증** — partial active = false positive risk.
5. **Gap → §4 회귀 trigger** — gap 발견 시 acceptance test plan 보강 task 자동 생성.

## 5. 단계 (Phases)

### Phase 1. Per-actor coverage matrix

actor × use case × layer 매트릭스 작성:

| Actor | Use case | Layer (적합) | Test file | Test count |
|-------|----------|--------------|-----------|------------|
| User (frontend) | signup | E2E (Playwright) + Component (Vitest) | `apps/web/tests/signup.spec.ts` + `apps/web/components/SignupForm.test.tsx` | 12 |
| User (frontend) | login | E2E + Component | `tests/login.spec.ts` + `LoginForm.test.tsx` | 14 |
| Backend (auth-api) | POST /signup | Integration (Vitest + testcontainers) + Unit (Vitest) | `services/auth/tests/signup.int.test.ts` + `signup.unit.test.ts` | 18 |
| Backend (auth-api) | POST /login + lockout | Integration + Unit + Property | `login.int.test.ts` + `login.unit.test.ts` + `login.property.test.ts` | 22 |
| 3rd-party (SES) | send verification email | Contract (Pact) + Integration (LocalStack SES) | `tests/ses.contract.test.ts` + `ses.int.test.ts` | 6 |
| 3rd-party (SES bounce) | bounce webhook | Contract + Integration | `tests/ses-bounce.contract.test.ts` | 4 |
| ... | ... | ... | ... | ... |

orphan check: acceptance criteria 가 test 와 매핑 안 된 row = gap.

### Phase 2. Layer 별 test 실행

actor 별 적합 framework 호출:

```bash
# frontend (User actor)
cd apps/web && pnpm test:component    # Vitest unit
cd apps/web && pnpm test:e2e          # Playwright

# backend (Auth-api actor)
cd services/auth && pnpm test:unit               # Vitest unit
cd services/auth && pnpm test:integration        # Vitest + testcontainers (Postgres + Redis + elasticmq)
cd services/auth && pnpm test:property            # fast-check property test

# 3rd-party (SES, SQS)
cd services/email && pnpm test:contract          # Pact provider verify
cd services/email && pnpm test:integration       # LocalStack SES + bounce simulator
```

### Phase 3. Test infra 검증

infra ready check:

| Infra | Verification | Status |
|-------|--------------|--------|
| Postgres testcontainer | `docker ps | grep postgres-test` + connection check | active / inactive |
| Redis miniredis | port 6380 listening | active |
| SQS elasticmq | port 9324 listening | active |
| SES LocalStack | LocalStack health endpoint | active |
| Pact broker | broker URL ping | active |
| MSW (frontend mock) | service worker registered | active |
| Playwright browser | `playwright install --check` | active |

partial inactive 시 → 본 skill 중단 + infra 복구 후 재실행.

### Phase 4. Coverage measurement

```bash
# overall
vitest run --coverage   # line / branch / function / statement
playwright test --reporter=html  # E2E coverage via instrumented build (optional)

# critical path
vitest run --coverage --include "**/auth-crypto/**" --include "**/jwt/**" --include "**/audit-log/**"
# → 100% 강제

# pgTAP (DB layer)
psql -d auth_test -f tests/db/coverage.sql
```

reports:
- `coverage/lcov-report/index.html` — line / branch / function
- critical path: line 100%, branch ≥ 95%
- overall: line ≥ 80%, branch ≥ 70%

### Phase 5. Gap report + acceptance gate

| Use case | Test mapping | Coverage | Gap |
|----------|--------------|----------|-----|
| signup happy | mapped (12 test) | line 92% | none |
| signup with marketing opt-in | **gap** (no test) | n/a | **GAP** — §4 회귀 |
| ... | ... | ... | ... |

acceptance gate:
- pass: orphan acceptance 0, line coverage ≥ 80%, critical path 100%, branch ≥ 70%
- fail: 1+ gap or coverage 미달 → §4 acceptance test plan 보강 task 생성

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## test-per-actor-use-case Output — <feature name> v<version>

### Summary
<3 줄: actor count / total use case / total test / coverage % / gap count / decision>

### Coverage Matrix
| Actor | Use case | Layer | Test file | Test count | Status | Coverage | Gap |
|-------|----------|-------|-----------|------------|--------|----------|-----|
| ... | ... | ... | ... | ... | pass/fail/skipped | line% / critical-path% | yes/no |

### Test Infra Status
| Infra | Status | Note |
|-------|--------|------|
| Postgres testcontainer | active | container id abc123 |
| Redis miniredis | active | port 6380 |
| ... | ... | ... |

### Coverage Summary
| Layer | Line | Branch | Critical path |
|-------|------|--------|---------------|
| Frontend (Vitest+Playwright) | 87% | 76% | 100% |
| Backend (Vitest) | 92% | 84% | 100% |
| 3rd-party (Pact) | 100% (contract) | n/a | n/a |
| Overall | 89% | 79% | 100% |

### Gap Report
| Use case | Reason | Action |
|----------|--------|--------|
| signup with marketing opt-in | no test mapped | §4 acceptance plan 의 per-actor section 보강 task 생성 |
| ... | ... | ... |

### Acceptance Gate
| Field | Value |
|-------|-------|
| Decision | pass / fail |
| Rationale | orphan 0 + coverage ≥ 80% + critical 100% / 또는 gap 발견 |
| Next action (if fail) | §4 회귀 — define-acceptance-test-plan 보강 |

### Cascade
- **§6 test-cross-actor-flow**: per-actor pass = cross-actor 진입 prerequisite
- **§6 measure-code-health**: coverage % 가 health metric
- **§7 prepare-launch-checklist**: per-actor test pass row 입력
- **§4 define-acceptance-test-plan** (gap 발견 시): plan 보강 회귀

### Next Step
<구체 action — 1줄: 예 "test-cross-actor-flow 호출, gap 0 confirmed">
```

## 7. Cross-phase cascade

- **§4 define-acceptance-test-plan**: gap 발견 시 plan 보강 회귀
- **§6 test-cross-actor-flow**: 본 skill pass 가 cross-actor 진입 prerequisite
- **§6 measure-code-health**: coverage % 가 health metric 입력
- **§7 prepare-launch-checklist**: per-actor row green 의 입력
- **§8 monitor-regressions**: per-actor test 가 regression baseline

## 8. 다음 skill (next in stage flow)

- `test-cross-actor-flow` — 본 skill pass 후 cross-actor flow E2E
- `measure-code-health` — coverage 가 health metric

권장 chain:
```
/buddy:chain test-per-actor-use-case,test-cross-actor-flow,measure-code-health -- "<feature> v<version>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `run-browser-qa`** (§6) — run-browser-qa 는 UI 자체 (visual regression / a11y / form / responsive), 본 skill 은 use case 단위 통합 (browser 사용하지만 use case 검증 목적).
- **vs `test-cross-actor-flow`** (§6) — 본 skill 은 single actor scope, 다음 stage 는 multi-actor flow.
- **vs `define-acceptance-test-plan`** (§4) — 그것은 plan, 본 skill 은 plan 의 actual 실행 + gap detection.
- **vs `monitor-regressions`** (§8) — 그것은 production 의 long-term regression, 본 skill 은 pre-release per-actor coverage.
- **vs `classify-qa-tiers`** (§6) — classify 는 intensity 결정 (Quick/Standard/Exhaustive), 본 skill 은 tier 안의 per-actor execution.

## 10. 중요 규칙

- **Actor × layer fit 의무** — layer mismatch 거부.
- **Use case → test 1:1 매핑** — orphan acceptance 0.
- **Coverage 정량** — line / branch / critical 분리.
- **Test infra 동시 active 검증** — partial 거부.
- **Gap → §4 회귀** — silent skip 금지.
- **Read-only on production code** — test file / fixture / report 만 산출, prod code 수정 안 함.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 6 출력 섹션 (Summary / Coverage Matrix / Infra / Coverage Summary / Gap / Acceptance Gate / Cascade) 모두 채워짐
- [ ] Coverage Matrix 의 모든 actor × use case row 에 layer + test file + count + status + coverage + gap 명시
- [ ] Test Infra Status 6+ row, partial active = inactive 거부
- [ ] Coverage Summary line / branch / critical path 분리, critical 100% 충족
- [ ] Gap Report 의 모든 row 에 reason + action (§4 회귀 task) 명시
- [ ] Acceptance Gate decision 단일 (pass/fail) + rationale
- [ ] §4 posture 적용 — gap-first, hedge 없음
- [ ] §0 anti-pattern 부재 — layer mismatch / 매핑 모호 / silent skip / infra 무시 / coverage 없음 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
