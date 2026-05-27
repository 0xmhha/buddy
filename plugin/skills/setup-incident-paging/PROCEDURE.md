# setup-incident-paging — on-call rotation + escalation + alert wiring

§7 Release & Beta phase 의 stage. **production incident 의 first response 구조** — on-call rotation schedule + severity 분류 + escalation policy + alert routing matrix + runbook 인덱스 + drill cadence 산출. paging tool (PagerDuty / Opsgenie / Splunk OnCall / 자체) 결정 + 운영 정책. 산출물은 rotation table + escalation policy + alert→runbook 매트릭스 + drill plan.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 팀 구조 + 인프라 정보 | ✅ | knowledge | 사용자 도메인 지식 | "on-call 팀 구성과 alert 대상 서비스를 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Incident paging 구성 (rotation + escalation + alert + runbook index) | artifact | structured document | `handle-incident` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Single-person on-call** — 1 명 sole engineer 가 24/7 → burnout + sole-point-of-failure. 최소 4 명 rotation.
- **Severity 분류 없음** — 모든 alert 가 동일 priority → SEV1 (outage) 와 SEV4 (informational) 가 같은 noise level → alert fatigue.
- **Escalation 미정의** — primary 가 ack 안 하면 어떻게 되는지 모름 → incident 30 분 방치 risk.
- **Alert routing 매트릭스 부재** — alert 가 어느 oncall 에게 가는지 무규칙 → wrong-team paging.
- **Runbook 인덱스 없음** — alert 받았는데 어느 runbook 참조할지 모름 → first response 시간 지연.
- **Drill 없음** — paging 시스템이 production incident 시 첫 firing → 작동 안 하면 발견 늦음. 분기 drill 의무.
- **Holiday / timezone 무시** — APAC team 의 oncall 이 NA hour 한정 paging 받으면 SLA 위반.
- **Orphan alert** — alert 정의했지만 runbook 없음, severity 없음 → noise + actionable 안 됨.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

production incident 의 first 5 minutes (detection → ack → triage 시작) 구조화. alert fatigue 방지, fair on-call rotation, escalation 자동화로 SLA preservation.

## 2. 사용 시점 (When to invoke)

- production launch 직전 (필수)
- SLO 정의 직후 — alert threshold 결정 동시
- 신규 oncall 멤버 onboarding 시
- paging tool 변경 시
- post-incident review 의 "paging gap" finding 후 보강
- 팀 expansion / re-org 후 rotation 재설계

## 3. 입력 (Inputs)

### 필수
- 팀 composition (oncall 가능 인원, timezone, language)
- SLO + alert thresholds (`setup-canary-deploy` 의 metric gate, `setup-rollback-runbook` 의 trigger)
- incident severity 정의 (또는 본 skill 에서 정의)
- paging tool 결정 (사용자 결정 필요): PagerDuty / Opsgenie / Splunk OnCall / xMatters / 자체 구현

### 선택
- 이전 paging history (mean ack time, false positive rate)
- compliance 의무 (SOC 2 의 incident response time)
- multi-region team — follow-the-sun rotation
- specialized rotations (security oncall vs platform oncall 분리)

### 입력이 부족할 때 forcing question
- "oncall 가능 인원이 4 명 미만이면 sustainable 한 rotation 불가능 (각자 매주 oncall) — 외부 contractor 또는 partial rotation 결정?"
- "primary timezone 이 어디? follow-the-sun 인지 single-tz 인지에 따라 rotation 전혀 다름."
- "paging tool 결정이 안 됐다면 monthly incident estimate + 예산으로 결정 — PagerDuty $X/seat/mo 인지 self-host 인지."
- "severity SEV1-4 의 정의가 합의됐나? '심각한' vs '경미한' 모호 → SLA 무근거."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — rotation 비율 / SLA / escalation 단호. "필요시 escalation" 거부, "primary 10 min ack 미수령 → secondary auto-page".
- **사용자 입력을 challenge** — "내가 oncall 다 할게" 발화에 "burnout risk + bus factor 1 → unsustainable" push.
- **Specificity 강제** — "monitor alert" 거부, "p99 login > 750ms (5min sustained) → SEV2 → primary backend oncall, runbook: rb-001".
- **Fairness bias** — fairness 측정 (per-person hours/quarter), unequal load 시 rotation 보정.

도메인 원칙:

1. **Rotation min 4 인** — sustainable rotation 의 하한.
2. **Severity 4 분류** — SEV1 (full outage / data loss) / SEV2 (degraded) / SEV3 (minor) / SEV4 (informational). 각 SEV 의 paging behavior 다름.
3. **Escalation auto** — primary 10 min ack 미수령 → secondary, 20 min → manager.
4. **Alert routing 매트릭스 의무** — alert source × severity → routing target 명시.
5. **Runbook 인덱스 의무** — 모든 alert 가 runbook link, orphan alert 0 maintain.
6. **Drill cadence** — 월 1회 paging drill (synthetic alert), 분기 1회 chaos drill, 신규 oncall 6 개월 내 first-page 경험.
7. **Fairness audit** — quarterly per-person hours, primary slot 분포 확인.

## 5. 단계 (Phases)

### Phase 1. Rotation 설계

| Aspect | Decision |
|--------|----------|
| Pool size | minimum 4 (sustainable), recommend 5-8 |
| Shift length | 1 week primary + 1 week secondary (recommended) — 24h shift 는 burnout |
| Coverage | follow-the-sun (multi-region) 또는 single-tz with on-call from home |
| Holiday | holiday calendar respect, swap 가능 with 1 week notice, holiday compensation |
| Weekend | included (production 24/7), weekend pay differential 권장 |
| Manager backup | always 1 manager-level on call (escalation last resort) |

본 SaaS auth example (4 eng + 1 platform lead):
- pool: W1, W2, W4 (W3 부분 — junior, secondary only first 6 mo)
- primary rotation: 1 week, 3 person rotating → 3 weeks idle
- secondary: 1 week shifted offset
- platform lead: always escalation final layer
- timezone: assumed single-tz NA, holiday calendar US

### Phase 2. Severity 정의

| SEV | Definition | Paging behavior | Ack SLA | Resolve SLA |
|-----|------------|-----------------|---------|-------------|
| **SEV1** | full outage / data loss / security breach (현재 + 잠재) | page primary + secondary + manager 즉시 | 5 min | 1 h target, 4 h hard |
| **SEV2** | degraded (subset of users / SLO violation / partial outage) | page primary, ack 10 min → secondary | 10 min | 4 h target, 24 h hard |
| **SEV3** | minor (workaround 가능, non-critical degradation) | in-hours only paging primary, after-hours queued | 1 h business hour | 24 h target, 1 week hard |
| **SEV4** | informational (anomaly detection, capacity warning) | no page, log + dashboard | n/a | tracked in backlog |

severity 결정자: alert source 의 inference rule (예: error rate > 5% AND > 1000 affected user → SEV1) 또는 oncall 의 manual upgrade. SEV downgrade 는 manager approval.

### Phase 3. Escalation policy

```
SEV1:
  T+0:    page primary + secondary + manager (parallel)
  T+5min: if no ack from primary, auto-call (phone) + Slack DM all
  T+15min: if no ack, page exec on-call (VP+)

SEV2:
  T+0:    page primary
  T+10min: if no ack, page secondary
  T+20min: if no ack, page manager
  T+30min: page exec if business-critical

SEV3:
  T+0:    notify primary in-hours (not page)
  T+next-business-day: if not triaged, escalate to manager

SEV4:
  T+0:    notify in shared channel (no page, no DM)
```

ack methods: paging tool ack (in-app), reply via SMS, Slack !ack, voice call accept. dedup: same incident multiple alerts → single page.

### Phase 4. Alert routing 매트릭스

| Alert source | Severity inference | Routing target | Runbook |
|--------------|---------------------|----------------|---------|
| CloudWatch alarm "auth-error-rate-high" | error > 1% (5min) → SEV2, > 5% → SEV1 | backend oncall primary | rb-001-error-rate-spike.md |
| CloudWatch alarm "auth-p99-latency-high" | p99 > SLO × 1.5 (5min) → SEV2, > SLO × 2.5 → SEV1 | backend oncall primary | rb-002-latency-spike.md |
| CloudWatch alarm "rds-connection-saturation" | conn > 80% (10min) → SEV2 | platform oncall (W4 / lead) | rb-010-rds-saturation.md |
| CloudWatch alarm "sqs-dlq-message-count" | DLQ ≥ 1 → SEV3 | backend oncall primary | rb-020-sqs-dlq.md |
| Datadog APM (X-Ray) "downstream-dependency-fail" | error > 0.5% on downstream (5min) → SEV2 | backend oncall primary | rb-030-downstream-fail.md |
| Sentry "uncaught-exception-spike" | > 10x baseline (15min) → SEV2 | backend oncall primary | rb-040-exception-spike.md |
| AWS Health Dashboard | service health degraded → SEV2 (auto) | platform oncall | rb-050-aws-health.md |
| GuardDuty (security) | high finding → SEV1 | security oncall (W2 + lead) | rb-060-guardduty.md |
| Cost anomaly (CloudWatch billing alarm) | spend > budget × 1.5 → SEV2 | platform lead (in-hours only) | rb-070-cost-spike.md |
| LaunchDarkly kill switch activation | manual trigger → SEV decided by activator | post-event tracking only (kill switch already escape valve) | rb-080-kill-switch-activation.md |

flap suppression: 5 min dedup window. Same alert firing 5x in 1h → escalate to noise-pattern review (alert tuning task).

### Phase 5. Runbook 인덱스 + Drill

runbook 인덱스 (모든 alert 가 1 runbook link):
- `docs/runbooks/rb-001-error-rate-spike.md` — first response (verify, scope, mitigate)
- `docs/runbooks/rb-002-latency-spike.md`
- ... (alert source 매트릭스의 각 row)

orphan alert 0 maintain: alert 추가 시 PR template 강제 (runbook link 없이 merge 불가 — CI lint).

drill cadence:

| Drill type | Frequency | Scope |
|------------|-----------|-------|
| **Paging drill** | 월 1회 | synthetic alert → primary ack → escalation 시뮬레이션 (manager 까지 안 감, 5분 내 ack) |
| **Chaos drill** | 분기 1회 | 실제 fault injection (FIS) → real alert fire → real triage → real runbook execution → post-mortem |
| **First-page experience** | 신규 oncall 6 개월 내 | 신규 멤버가 supervised 환경에서 first real page 처리 |
| **Escalation drill** | 반기 1회 | primary 의도적 미응답 → escalation 자동 작동 검증 |
| **Holiday drill** | 연 1회 | holiday calendar respect, swap, fallback 검증 |

drill 결과로 escalation rule / runbook / alert threshold 보정.

fairness audit (quarterly):

| Metric | Target |
|--------|--------|
| Per-person primary hours / quarter | within ± 20% of mean |
| Per-person secondary hours / quarter | within ± 20% of mean |
| Per-person weekend / holiday days | within ± 1 day of mean |
| First-page experience age | < 6 mo for active oncall |

unequal load 시 rotation rebalance.

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## setup-incident-paging Output — <project name>

### Summary
<3 줄: paging tool / pool size / SEV1 ack SLA / drill cadence>

### Paging Tool Decision
| Field | Value |
|-------|-------|
| Tool | PagerDuty Standard ($21/seat/mo × 4 = $84/mo) |
| Rationale | mature ecosystem, AWS integration, mobile app + voice call escalation, SOC 2 compliance |
| Self-host alternative | Splunk OnCall (paid) or self-host (high maintenance) |
| Migration plan (future) | PagerDuty config 은 Terraform-managed (pagerduty provider) → migration possible |

### Rotation Schedule
| Slot | Duration | Pool | Compensation |
|------|----------|------|--------------|
| Primary | 1 week (Mon-Sun) | W1, W2, W4 (3 person, 3-week idle) | base + on-call differential per shift |
| Secondary | 1 week (offset) | W1, W2, W4 (rotation) | base + reduced differential |
| Platform lead (escalation) | always | platform lead | always-on, no differential (role expectation) |
| Holiday | per company calendar, swap allowed with 1-week notice | rotating | holiday differential 2x base shift |

### Severity Definition
| SEV | Definition | Paging | Ack SLA | Resolve SLA |
|-----|------------|--------|---------|-------------|
| SEV1 | full outage / data loss / security breach | primary + secondary + manager 즉시 | 5 min | 1 h target / 4 h hard |
| SEV2 | degraded / SLO violation | primary, escalate 10 min | 10 min | 4 h target / 24 h hard |
| SEV3 | minor (workaround 가능) | in-hours primary | 1 h business | 24 h target / 1 week hard |
| SEV4 | informational | no page | n/a | backlog |

### Escalation Policy
| Severity | T+0 | T+10 min | T+20 min | T+30 min |
|----------|-----|----------|----------|----------|
| SEV1 | primary + secondary + manager (parallel) | auto-call primary + Slack all | exec on-call (VP+) | n/a (already escalated) |
| SEV2 | primary | secondary | manager | exec (if business-critical) |
| SEV3 | primary in-hours | next business day if untriaged → manager | n/a | n/a |
| SEV4 | shared channel notify | n/a | n/a | n/a |

### Alert Routing Matrix
| Alert source | Severity inference | Routing | Runbook |
|--------------|---------------------|---------|---------|
| auth-error-rate-high | error > 1% → SEV2, > 5% → SEV1 | backend oncall primary | rb-001 |
| auth-p99-latency-high | p99 > SLO × 1.5 → SEV2 | backend primary | rb-002 |
| rds-connection-saturation | > 80% → SEV2 | platform oncall | rb-010 |
| sqs-dlq-message-count | ≥ 1 → SEV3 | backend primary | rb-020 |
| ... | ... | ... | ... |

### Runbook Index
| Runbook | Alert | Owner | Last reviewed |
|---------|-------|-------|----------------|
| rb-001-error-rate-spike | auth-error-rate-high | backend lead | 2026-XX-XX |
| rb-002-latency-spike | auth-p99-latency-high | backend lead | 2026-XX-XX |
| ... | ... | ... | ... |

(orphan alert: 0)

### Drill Plan
| Drill type | Frequency | Owner | Last run | Pass criteria |
|------------|-----------|-------|----------|---------------|
| Paging drill | monthly (every 1st Monday) | platform lead | 2026-XX-XX | primary ack ≤ 5 min, escalation auto-fire ≤ 10 min |
| Chaos drill | quarterly | platform + SRE | 2026-Q? | real alert fire → real triage → resolve via runbook → post-mortem |
| First-page experience | per onboarding | manager | per new oncall | new oncall handles first page supervised within 6 mo |
| Escalation drill | semi-annual | platform lead | 2026-H? | primary 의도적 미응답 → escalation 정상 작동 |
| Holiday drill | annual (Q4) | platform lead | 2026-Q4 | holiday calendar respect, swap, fallback 정상 |

### Fairness Audit (quarterly)
| Metric | Target | Last quarter |
|--------|--------|--------------|
| Per-person primary hours | mean ± 20% | (data) |
| Per-person secondary hours | mean ± 20% | (data) |
| Per-person weekend / holiday days | mean ± 1 day | (data) |
| First-page experience age | < 6 mo | (data) |

### Cascade
- **§7 setup-canary-deploy**: canary halt signal 이 본 skill 의 alert 와 통합 routing
- **§7 setup-rollback-runbook**: runbook index 가 본 skill 의 runbook 인덱스와 정합
- **§7 setup-feature-flags**: ops kill switch activation 이 본 skill 의 SEV1 alert 와 routing
- **§7 prepare-launch-checklist**: paging plan 이 readiness gate row
- **§8 handle-incident**: paging 활성 → handle-incident skill 진입점, runbook 이 incident triage 도구

### Next Step
<구체 action — 1줄: 예 "PagerDuty account 생성 + Terraform pagerduty provider 통합 + 첫 paging drill 1주일 내 schedule">
```

## 7. Cross-phase cascade

- **§7 setup-canary-deploy / setup-rollback-runbook / setup-feature-flags**: alert 와 runbook 인덱스 통합
- **§7 prepare-launch-checklist**: paging plan 이 row
- **§8 handle-incident**: paging → handle-incident 진입
- **§8 conduct-postmortem**: drill / real incident 후 lessons 가 본 skill 의 update input

## 8. 다음 skill (next in stage flow)

- `prepare-launch-checklist` — paging plan 통합
- (post-launch) `handle-incident` (§8) — actual incident triage 도구

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `handle-incident`** (§8) — 그것은 actual incident triage 후 절차 (post-page), 본 skill 은 paging infra 사전 준비 (pre-page). 본 skill 이 prerequisite.
- **vs `conduct-postmortem`** (§8) — postmortem 은 post-incident 회고, 본 skill 은 pre-incident readiness. drill 결과가 postmortem 의 일부.
- **vs `setup-rollback-runbook`** (§7) — rollback runbook = 특정 alert 의 procedure, 본 skill 의 runbook 인덱스의 한 entry. 본 skill 이 통합 layer.
- **vs `audit-error-budget`** (§8 — pending) — error budget = SLO breach tracking, 본 skill 은 alert routing. 두 layer 모두 SRE 영역.

## 10. 중요 규칙

- **Rotation pool ≥ 4** — sustainable 한 minimum.
- **Severity 4 분류 의무** — SEV1-4 명확.
- **Escalation auto 의무** — manual escalation 거부.
- **Alert routing 매트릭스 의무** — 모든 alert 가 routing target + runbook link.
- **Runbook 인덱스 의무** — orphan alert 0 maintain (CI lint 강제).
- **Drill 의무** — 월 + 분기 + 반기 + 연 cadence.
- **Fairness audit 의무** — quarterly per-person 분포 검증.
- **Read-only on production** — 본 skill 은 paging infra plan, actual paging tool config 은 §5 build / SRE infra task.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행 (rotation → severity → escalation → alert routing → runbook + drill)
- [ ] §6 의 9 출력 섹션 (Summary / Tool Decision / Rotation / Severity / Escalation / Alert Routing / Runbook Index / Drill / Fairness / Cascade) 모두 채워짐
- [ ] Rotation pool ≥ 4
- [ ] Severity 4 분류 (SEV1-4) 모두 정의 + ack/resolve SLA
- [ ] Escalation policy 모든 SEV 에 T+0 / T+10 / T+20 / T+30 시점 명시
- [ ] Alert Routing Matrix ≥ 7 alert (auth / DB / cache / queue / dependency / exception / cost / security 다양 cover)
- [ ] Runbook Index 모든 alert 가 link 있음 (orphan 0 명시)
- [ ] Drill Plan 5 type (paging / chaos / first-page / escalation / holiday) 모두 cadence + owner + pass criteria
- [ ] Fairness Audit 4 metric 모두 정의
- [ ] §4 posture — 정량 SLA / pool / cadence, "필요시" hedge 없음
- [ ] §0 anti-pattern 부재 — single-person / severity 없음 / escalation 없음 / routing 없음 / runbook 없음 / drill 없음 / timezone 무시 / orphan 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
