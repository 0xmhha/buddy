# define-acceptance-test-plan — actor 별 + cross-actor 완료 기준 + test infra

§4 의 acceptance plan stage. feature spec (per-actor acceptance) + task DAG (cross-actor edges) + tech stack (test framework 결정) 을 입력으로 verifiable test plan 을 작성한다. per-actor (unit / integration / contract) + cross-actor (E2E flow) 분류 + test infra (runner / fixture / mock 전략) 결정. §6 verify-quality 의 입력.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Feature specs + acceptance criteria | ✅ | artifact / knowledge | `define-feature-spec` 산출물 또는 사용자 설명 | "테스트 계획을 세울 feature와 수용 기준을 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Acceptance test plan (per-actor + cross-actor) | artifact | structured YAML | `generate-tests-from-spec`, `verify-quality` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다:

- **"happy path 만"** — error / edge case test 누락 시 production 에서 깨짐.
- **Test category 무분류** — unit / integration / contract / E2E 분류 안 하고 나열만 → coverage gap 보임.
- **Test infra 결정 안 함** — runner / fixture / mock 전략 미정 시 build-feature 단계에서 매번 즉흥 결정.
- **Cross-actor flow test 누락** — actor 별 unit 만 있고 actor 협업 시나리오 검증 없음 → integration 시점 chaos.
- **Acceptance gate 모호** — 어느 test 통과율이 build 완료 조건인지 명시 안 함.

§5 모든 phase 누락 없이 수행하라.

## 1. 목적

acceptance criteria 를 verifiable test plan 으로 변환. test 가 build-feature 의 done condition 이고 §6 verify-quality 의 input. 잘못된 test plan 은 production 결함 누수의 직접 원인.

## 2. 사용 시점 (When to invoke)

- task DAG + parallel plan 확정 후 (chain 권장)
- pre-build verification gate 정의 시점
- test infra 결정 또는 변경 시
- regression 발생 후 test plan 보강

## 3. 입력 (Inputs)

### 필수
- feature spec (per-actor acceptance criteria)
- task DAG (`map-task-dependencies` 산출물 — cross-actor edges 가 cross-actor flow 의 trigger)
- §3 tech stack ADR (test framework / runner 결정 영향)

### 선택
- 기존 test infra (CI runner, fixture management, mock strategy)
- coverage 목표 (line / branch / mutation %)
- 외부 contract test 요구 (3rd-party SaaS contract)

### 입력이 부족할 때 forcing question
- "이 actor 의 acceptance 가 manual 검증인가, automated 인가? automated 면 어느 framework?"
- "test fixture 가 real DB 인가, mock 인가? schema 변경 시 fixture sync 전략?"
- "cross-actor flow test 가 E2E (real 모든 component) 인가, integration (partial mock) 인가?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — test category 결정은 단호. "unit 또는 integration" 거부, 어느 layer 가 적합한지 명시.
- **사용자 입력을 challenge** — "test 만들면 됨" 발화에 "어느 layer? coverage 목표? mutation test 도?" push.
- **Specificity 강제** — "test it" 거부. assertion (어떤 input → 어떤 output / state change), test name, framework 명시.

도메인 원칙:

1. **Per-actor + cross-actor 분리** — 둘 다 필요. unit / integration / contract / E2E layer 명시.
2. **Test infra 는 plan 의 일부** — runner / fixture / mock 전략이 acceptance 의 reproducibility 보장.
3. **Acceptance gate 는 measurable** — "test 통과" 보다 "X% line coverage + 모든 critical-path test pass + Y mutation score" 같이 정량.
4. **Mock vs real trade-off 명시** — fast (mock) vs realistic (real DB). 각 결정에 trade-off 적용.
5. **3rd-party contract test 의무** — 외부 의존이 있으면 contract test 필수 (Pact 등).

## 5. 단계 (Phases)

### Phase 1. Per-actor test category 매핑

각 actor 별로 적합한 test layer:

| Actor | Layer | Framework | Scope |
|-------|-------|-----------|-------|
| User (frontend) | Component (unit) | jest + @testing-library/react | render, interaction, validation |
| User (frontend) | E2E | playwright / cypress | full signup flow in browser |
| Auth-system (backend) | Unit | go test | password hashing, token issuance |
| Auth-system (backend) | Integration | go test + testcontainers postgres | DB write + commit |
| Email-verifier (3rd-party) | Contract | Pact | webhook handler 가 SendGrid event schema 처리 |

### Phase 2. Cross-actor flow test

actor 협업 시나리오:

| Flow | Actors involved | Test framework | Scope |
|------|------------------|----------------|-------|
| signup → email verification → first login | frontend + auth-backend + email-verifier | playwright + test SendGrid sandbox | 전 flow real component |
| forgot-password reset flow | frontend + auth-backend + email-verifier | playwright | 전 flow |
| ... | ... | ... | ... |

cross-actor edges (DAG) 가 flow 의 trigger.

### Phase 3. Test infra 결정

| Infra component | Decision | Rationale |
|-----------------|----------|-----------|
| CI runner | GitHub Actions | tech stack ADR 의 결정과 일치 |
| Fixture management | testcontainers (real) for backend, MSW (mock) for frontend | speed vs realism trade-off |
| Mock strategy | record-replay for 3rd-party | API stability |
| Coverage tool | jest + go cover | tech stack 별 native |
| Mutation testing | (deferred) | initial scope 외 |

### Phase 4. Acceptance gate

build-feature 완료 조건:

| Gate | Threshold | Tool |
|------|-----------|------|
| Per-actor unit test | 100% pass | jest / go test |
| Cross-actor E2E | 모든 critical flow pass | playwright |
| Contract test | 모든 contract pass | Pact |
| Coverage (line) | ≥ 80% (critical paths 100%) | jest --coverage / go cover |
| Mutation test | (Phase 2 +) | (deferred) |

### Phase 5. §6 verify-quality cascade

본 plan 이 §6 의 입력 schema:
- §6 audit-security 와 결합: security test 는 별도 layer (penetration / fuzz)
- §6 measure-code-health 와 결합: coverage 는 health metric 의 일부
- §6 의 quality gate = 본 plan 의 acceptance gate + security audit pass

## 6. 산출물 형식 (Output format)

> structured 출력, prose 변환 금지.

```markdown
## define-acceptance-test-plan Output — <feature name>

### Summary
<3 줄: total test count / category 분포 / coverage 목표 / 가장 큰 test infra 결정>

### Per-Actor Test Plan
| Actor | Layer | Framework | Test Count (est.) | Scope |
|-------|-------|-----------|-------------------|-------|
| ... | ... | ... | ... | ... |

### Cross-Actor Flow Tests
| Flow | Actors | Framework | Trigger Edge | Scope |
|------|--------|-----------|--------------|-------|
| ... | ... | ... | ... | ... |

### Test Infrastructure
| Component | Decision | Rationale |
|-----------|----------|-----------|
| ... | ... | ... |

### Acceptance Gate
| Gate | Threshold | Tool |
|------|-----------|------|
| ... | ... | ... |

### Cascade to §6
- §6 audit-security: 본 plan 외 보안 test 추가 layer
- §6 measure-code-health: coverage 가 health metric 입력
- §6 quality gate = 본 acceptance gate + security pass

### Next Step
<구체 action — 1줄: 예 "estimate-build-timeline 호출 — test 작성 task 도 timeline 에 포함">
```

## 7. Cross-phase cascade

- **§4 estimate-build-timeline**: test 작성 task 도 build-feature task duration 에 포함
- **§5 build-feature**: TDD 의 RED phase 가 본 plan 의 test 일부 작성
- **§6 verify-quality**: acceptance gate 가 quality gate 의 일부
- **§7 ship-release**: pre-release smoke test 는 본 plan 의 critical flow subset

## 8. 다음 skill (next in stage flow)

- `estimate-build-timeline` — test 작성도 timeline 에 포함

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `classify-qa-tiers`** — classify-qa-tiers 는 §6 의 quality intensity (Quick / Standard / Exhaustive) 결정. 본 skill 은 §4 의 acceptance test plan. 본 skill 이 먼저, classify-qa-tiers 가 §6 에서 그 plan 을 강도 분류.
- **vs `audit-security`** — audit-security 는 §6 의 보안 audit. 본 skill 은 functional acceptance. 두 plan 이 quality gate 에서 결합.
- **vs `run-browser-qa`** — run-browser-qa 는 §6 의 actual browser QA 실행. 본 skill 은 §4 plan. 본 skill 이 먼저.
- **vs `monitor-regressions`** — monitor-regressions 는 §8 의 regression 감지. 본 skill 의 test 가 regression baseline.

## 10. 중요 규칙

- **Per-actor + cross-actor 둘 다** — 한쪽 누락 금지.
- **Test infra 결정 의무** — runner / fixture / mock 명시.
- **Acceptance gate 정량** — measurable threshold.
- **3rd-party contract test 의무** — 외부 의존 있으면 필수.
- **TDD 통합** — §5 build-with-tdd 와 결합 시 본 plan 의 test 가 RED phase 출발점.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 6 출력 섹션 (Summary / Per-Actor / Cross-Actor / Test Infra / Acceptance Gate / Cascade / Next Step) 모두 채워짐
- [ ] Per-Actor Test Plan 이 모든 actor 포함, layer 명시
- [ ] Cross-Actor Flow Tests 가 ≥ 1 critical flow + 모든 cross-actor edges 의 trigger 매핑
- [ ] Test Infrastructure 가 runner / fixture / mock / coverage tool 모두 결정
- [ ] Acceptance Gate 가 정량 threshold 명시
- [ ] §4 posture 적용
- [ ] §0 anti-pattern 들이 등장하지 않음

하나라도 no 면 해당 phase 로 돌아가 보강.
