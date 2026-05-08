# prepare-launch-checklist — launch readiness gate (17+ 항목)

§7 Release & Beta phase 의 stage. **GA 직전 final cross-functional readiness check** — Engineering / Security & Compliance / Operations / Product / Legal & Comms / Cost & Business 6 axis 의 17+ 항목을 evidence + owner + status 로 통합. §6 quality gate (자동) + §7 의 다른 6 stage 산출물을 input 으로 받아 launch go/no-go 권고. 산출물은 launch checklist + blocker 리스트 + 권고 + sign-off 절차.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Engineering 항목만 cover** — security / legal / marketing 누락 시 GA 후 compliance / customer / press incident.
- **Evidence 없이 "OK" 만 표시** — 각 row 의 status 가 evidence link 없이 "green" 만이면 검증 불가.
- **Owner 부재** — "OK" 인데 누가 sign-off 했는지 모름 → 사고 시 책임 단절.
- **Blocker 의 ETA 없음** — red row 만 표시하고 fix ETA 없으면 GA 일정 영향 추정 불가.
- **Pass criteria 정량 없음** — "대부분 green" 같이 모호 → 정치적 결정.
- **17 항목 미만** — 6 axis × 평균 3 row = 17+ baseline. 그보다 적으면 axis 누락 의심.
- **Static checklist** — 이전 release 의 checklist 복사 후 update 안 함 → stale state.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

§6 quality gate (자동) 와 §7 의 다른 6 stage 결과를 통합해 cross-functional GA readiness 의 single-source-of-truth. 사고 시 후행 audit 산출물의 일부.

## 2. 사용 시점 (When to invoke)

- GA 직전 (default for new release)
- 신규 product launch / region expansion / breaking change major version 시 (필수)
- 산업 규제 영역 (의료/금융 — checklist 가 audit 산출물)
- post-incident review 의 "checklist gap" finding 후 보강
- regular cadence (분기 GA — checklist template 보강)

## 3. 입력 (Inputs)

### 필수
- §6 verify-quality 결과 (test pass / coverage / security audit / code health)
- §7 의 다른 6 skill 산출물:
  - `setup-canary-deploy` plan
  - `setup-feature-flags` system + kill switch inventory
  - `setup-rollback-runbook`
  - `setup-incident-paging` rotation + escalation
  - `run-uat` decision
  - `run-beta-program` decision (해당 시)
- legal review status (ToS / Privacy / DPA / regulatory filings)
- marketing readiness (announcement, support channel)

### 선택
- 이전 release 의 checklist 결과 (regression 비교)
- 산업별 추가 항목 (의료 HIPAA / 금융 SOC 2 / EU GDPR)
- regional launch 별 별도 row (timezone / language / payment)

### 입력이 부족할 때 forcing question
- "§7 의 다른 6 skill 이 모두 invoked 됐나? 누락된 skill 의 row 는 기본 'red — not yet performed'."
- "legal review 가 GA gate 인가, post-launch (terms 수정 가능) 인가? 산업별로 다름."
- "marketing announcement timing 이 GA 와 동기인가, async 인가? async 면 messaging 일관성 책임자?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — go/no-go 권고 단호. "거의 ready" 거부, "green ≥ 90% AND red 0 → go".
- **사용자 입력을 challenge** — "OK" 발화에 "evidence 어디? owner 이름? 시간?" push.
- **Specificity 강제** — vague "테스트 통과" 거부, "Vitest unit 215/215 + Playwright 8/8 + Pact 5/5 + k6 p99 488ms (SLO 500ms 통과)".
- **Cross-functional bias** — engineering 외 5 axis (security / ops / product / legal / cost) 모두 cover.

도메인 원칙:

1. **6 axis 의무** — Engineering / Security & Compliance / Operations / Product / Legal & Comms / Cost & Business.
2. **각 row 4 attribute** — status (green/yellow/red) + owner + evidence + (red/yellow 시) ETA + mitigation.
3. **Pass criteria 정량** — green ≥ 90%, red 0, yellow ≤ 10% with documented dispositions.
4. **Cross-skill cascade** — §7 의 다른 6 skill 산출물이 row input. 미실행 skill 의 row 는 default red.
5. **17+ row** — 6 axis × 평균 3 row 이상. 그 이하면 axis 누락.
6. **Owner 명시 의무** — 모든 row 에 named owner (이름 + role), 익명 sign-off 거부.

## 5. 단계 (Phases)

### Phase 1. Category 정의 (6 axis)

| Axis | Sub-rows (typical) |
|------|---------------------|
| **Engineering** (5) | unit/integration test / contract test / performance / security audit / observability |
| **Security & Compliance** (4) | pen-test / privacy audit / data residency / supply chain (SBOM) |
| **Operations** (4) | incident paging / rollback runbook / canary plan / feature flag governance |
| **Product** (3) | UAT / beta / pricing & packaging |
| **Legal & Comms** (4) | ToS / Privacy / DPA / marketing announcement |
| **Cost & Business** (3) | cost estimate (Infracost) / unit economics / SLA commitments |

총 23 baseline row. 산업별 추가 (HIPAA: BAA, GDPR: DPIA, SOC 2: audit log retention).

### Phase 2. Per-row status 추출

| Row | Status | Evidence | Owner | ETA / mitigation |
|-----|--------|----------|-------|------------------|
| Unit test pass | green | Vitest 215/215 in CI run #N | tech lead | n/a |
| ... | ... | ... | ... | ... |

automated extraction (CI status, k6 result) 와 manual sign-off (legal, marketing) 분리. automated 는 link 가 CI build URL, manual 은 link 가 sign-off doc / email thread.

### Phase 3. Cross-skill cascade integration

본 skill 이 §7 의 다른 6 skill 산출물 통합:

| Source skill | Maps to checklist row(s) |
|--------------|--------------------------|
| `setup-canary-deploy` | "canary plan exists" + "metric gate ≥ 4 metric" + "rollback policy auto" |
| `setup-feature-flags` | "flag system selected" + "kill switch inventory ≥ 2" + "cleanup SLA defined" |
| `setup-rollback-runbook` | "rollback runbook exists" + "schema migration safe" + "RTO < 30 min" |
| `setup-incident-paging` | "on-call rotation defined" + "escalation policy" + "alert routing complete" |
| `run-uat` | "UAT decision = go or conditional go" |
| `run-beta-program` | "beta decision = go" (해당 시 — small release 는 skip 가능) |

§7 skill 누락 시 해당 row 는 default red ("not yet performed"). 산출물 status = red 시 동일.

### Phase 4. Blocker triage

red row 의 처리:

| Severity | Definition | Resolution |
|----------|------------|------------|
| **Hard blocker** | regulatory / data loss / 사용자 기본 기능 차단 | GA block, fix 후 re-run checklist |
| **Soft blocker** | UX 저하 / 마이너 SLO 위반 / 알려진 workaround | conditional go with hotfix ETA + monitoring |
| **Warning (yellow)** | 미완 but post-launch 가능 / 시간 한정 acceptable | go with documented disposition |

each blocker 의 ETA + accountable owner + monitoring plan 필수.

### Phase 5. Go/no-go 권고

| Decision | Pass criteria |
|----------|---------------|
| **go** | green ≥ 90%, red (hard blocker) 0, yellow ≤ 10% with disposition |
| **conditional go** | red (soft blocker) ≤ 1 with hotfix ETA ≤ 48h + accountable owner + monitoring |
| **no-go** | red (hard blocker) ≥ 1 OR green < 80% |

권고 article: "Recommend [go / conditional go / no-go] based on [N green / M red / K yellow]." + 권고 의 owner (release manager). final decision 은 leadership (VP+).

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## prepare-launch-checklist Output — <feature name> v<version>

### Summary
<3 줄: total row / green:yellow:red 분포 / decision 권고 / largest blocker>

### Checklist (6 Axis)

#### Engineering (N=5)
| Row | Status | Evidence | Owner | ETA / mitigation |
|-----|--------|----------|-------|------------------|
| Unit test pass | green | Vitest 215/215 (CI #N) + go cover (n/a) | W1 | n/a |
| Integration test pass | green | Vitest+testcontainers 50/50 (CI #N) | W1 | n/a |
| Contract test pass | green | Pact provider verify 5/5, oasdiff 0 breaking | W2 | n/a |
| Performance (k6 SLO) | green | login p99 488ms vs 500ms SLO, error 0.05% | W4 | n/a |
| Observability ready | green | OTel SDK init + 4 dashboards + 5 alarms | W4 | n/a |

#### Security & Compliance (N=4)
| Row | Status | Evidence | Owner | ETA / mitigation |
|-----|--------|----------|-------|------------------|
| Security audit (SAST/DAST) | green | OWASP ZAP 0 high, Semgrep 0 high, npm audit 0 high | W2 | n/a |
| Pen-test (gated, post-launch v1) | yellow | scheduled 2 weeks post-GA | security lead | post-launch acceptable for v1 |
| Privacy audit (GDPR) | green | DPIA signed, data flow diagram approved | privacy officer | n/a |
| Supply chain (SBOM) | green | CycloneDX SBOM generated, no CVE high | W2 | n/a |

#### Operations (N=4)
| Row | Status | Evidence | Owner | ETA / mitigation |
|-----|--------|----------|-------|------------------|
| Incident paging | green | rotation defined (4 eng), escalation 5/10/20 min, runbook indexed | platform lead | n/a |
| Rollback runbook | green | docs/runbooks/rollback-saas-auth.md, RTO 5 min app + 30 min schema | W4 | n/a |
| Canary plan | green | 5 stage (1/5/25/50/100%) + metric gate + auto-rollback | W4 | n/a |
| Feature flag governance | green | LaunchDarkly Pro setup, 2 ops kill switch, 90d cleanup SLA | platform lead | n/a |

#### Product (N=3)
| Row | Status | Evidence | Owner | ETA / mitigation |
|-----|--------|----------|-------|------------------|
| UAT decision | green | UAT 13/13 scenario pass, 0 blocker (run-uat output) | PM | n/a |
| Beta decision | green | NPS +27, 0 blocker, 7 cohort sign-off (run-beta-program output) | PM | n/a |
| Pricing & packaging | green | Free / Pro / Enterprise tier defined, billing integrated | product lead | n/a |

#### Legal & Comms (N=4)
| Row | Status | Evidence | Owner | ETA / mitigation |
|-----|--------|----------|-------|------------------|
| Terms of Service | green | reviewed by legal, signed PDF | legal | n/a |
| Privacy Policy | green | reviewed, GDPR/CCPA compliant | legal | n/a |
| DPA template | green | enterprise DPA template ready | legal | n/a |
| Marketing announcement | yellow | blog post draft pending CMO review | marketing | 24h before launch |

#### Cost & Business (N=3)
| Row | Status | Evidence | Owner | ETA / mitigation |
|-----|--------|----------|-------|------------------|
| Cost estimate (Infracost) | green | dev $480/mo, staging $1,420/mo, prod estimate $2,100/mo (within budget) | platform lead | n/a |
| Unit economics | green | $/MAU 추정 $0.04 vs revenue $X/MAU → margin healthy | product lead | n/a |
| SLA commitments | green | 99.9% availability + p99 500ms login + support response < 4h business / 1h enterprise | success lead + product | n/a |

### Distribution
| Status | Count | % |
|--------|-------|---|
| green | 21 | 91% |
| yellow | 2 | 9% |
| red | 0 | 0% |
| Total | 23 | 100% |

### Blocker Triage (red rows only)
| Row | Severity | Mitigation | Owner | ETA |
|-----|----------|------------|-------|-----|
| (none) | n/a | n/a | n/a | n/a |

### Yellow Dispositions
| Row | Disposition |
|-----|-------------|
| Pen-test (post-launch v1) | Accepted as gated post-launch (2 weeks). v1 has internal SAST/DAST pass. v2 launches gated on pen-test pass. |
| Marketing announcement | Acceptable yellow — final CMO review 24h before launch is standard process. Not a launch blocker. |

### Decision Recommendation
| Field | Value |
|-------|-------|
| Recommendation | **go** |
| Rationale | green 21/23 (91%, ≥ 90% threshold pass) + red 0 + yellow 2 with documented dispositions |
| Final approver (leadership) | VP Engineering + VP Product (joint sign-off) |
| Launch window (proposed) | 2026-XX-XX HH:MM UTC, business hours preferred for fast support response |

### Cascade
- **§7 ship-release final stage**: prepare-launch-checklist 의 go 권고 → automate-release-tagging + GA deploy
- **§7 setup-incident-paging**: 본 skill 의 ops 항목과 정합 (paging row 가 본 checklist 의 입력)
- **§8 monitor-regressions**: 본 checklist 의 SLO row 가 §8 의 baseline
- **§8 conduct-postmortem** (incident 시): 본 checklist 산출물이 post-mortem 의 baseline (어느 row 가 prevent 했어야 하는지)

### Next Step
<구체 action — 1줄: 예 "VP joint sign-off 받은 후 launch window 에 GA deploy 시작 — automate-release-tagging + canary stage 1 시작">
```

## 7. Cross-phase cascade

- **§7 ship-release final**: go 권고가 GA deploy 의 trigger
- **§7 setup-incident-paging**: ops row 가 paging plan 정합
- **§8 monitor-regressions**: SLO row 가 baseline
- **§8 conduct-postmortem**: incident 시 본 checklist 가 prevention baseline

## 8. 다음 skill (next in stage flow)

- `automate-release-tagging` — go 권고 후 GA tagging
- (해당 시) post-launch monitoring 진입

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `setup-quality-gates`** (§7) — quality gates = automated CI gates (개발 환경 commit-time), 본 skill = cross-functional human sign-off (release-time). 두 layer 모두 release 의 일부.
- **vs `run-uat`** (§7) — UAT 결과가 본 checklist 의 한 row (Product / UAT decision). UAT 가 input.
- **vs `run-beta-program`** (§7) — beta 결과도 한 row (Product / Beta decision).
- **vs `automate-release-tagging`** (§7) — release tagging 은 본 skill go 권고 후 실행. 본 skill 이 prerequisite.
- **vs orchestrator inline checklist** (existing in ship-release/PROCEDURE.md) — orchestrator 의 7-항목 high-level → 본 skill 의 17+ 세분화. orchestrator 의 inline 은 본 skill 호출로 대체.

## 10. 중요 규칙

- **6 axis 의무** — single-axis (engineering only) checklist 거부.
- **17+ row 의무** — axis 당 평균 3 row 이상.
- **각 row 4 attribute** — status / owner / evidence / (red/yellow 시) ETA + mitigation.
- **Pass criteria 정량** — green ≥ 90%, red 0, yellow ≤ 10%.
- **Cross-skill cascade** — §7 의 다른 6 skill 산출물 통합, 미실행 skill 의 row default red.
- **Owner named** — 익명 거부, 이름 + role.
- **Recommendation 단일** — "go" / "conditional go" / "no-go" 셋 중 하나, hedge 거부.
- **Final approver leadership** — VP+ joint sign-off, single-engineer 권한 거부.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행 (axis 정의 → status 추출 → cross-skill cascade → blocker triage → decision)
- [ ] §6 의 7 출력 섹션 (Summary / Checklist 6 axis / Distribution / Blocker / Yellow Dispositions / Decision / Cascade) 모두 채워짐
- [ ] 6 axis 모두 cover (Engineering / Security & Compliance / Operations / Product / Legal & Comms / Cost & Business)
- [ ] Total row ≥ 17 (axis 당 ≥ 2, 평균 3)
- [ ] 모든 row 에 status + evidence + owner 명시, red/yellow 시 ETA + mitigation 추가
- [ ] §7 의 다른 6 skill 산출물 통합 — canary / flags / rollback / paging / UAT / beta row 매핑
- [ ] Distribution 정량 (green / yellow / red count + %)
- [ ] Decision Recommendation 단일 ("go" / "conditional go" / "no-go") + rationale + final approver
- [ ] §4 posture — 권고 단호, "대부분 OK" 류 hedge 없음
- [ ] §0 anti-pattern 부재 — engineering only / evidence 없음 / owner 없음 / blocker ETA 없음 / pass 정량 없음 / 17 미만 / static 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
