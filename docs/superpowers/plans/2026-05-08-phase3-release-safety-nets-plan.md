# Phase 3 — §7 Release Safety Nets 7 Stage Skill 작성 Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement task-by-task. Each task = one PROCEDURE.md + command md + catalog row + commit.

**Goal:** §7 Release & Beta phase 의 production safety net 을 구성. `ship-release` orchestrator 가 현재 PR/tagging 자동화까지만 가능 — canary / feature flag / rollback / UAT / beta / launch checklist / incident paging 7 개의 안전망 stage skill 을 채워 상용 배포 직전 risk 를 plan-driven 으로 관리한다.

**Parent plan:** [`2026-05-06-stage-buildout-plan.md`](./2026-05-06-stage-buildout-plan.md) Phase 3.
**SSoT:** [`docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`](../specs/2026-05-06-lifecycle-orchestrator-architecture.md) §3 §7 + §10.
**Predecessor plans:**
- [`2026-05-07-phase1-design-stages-plan.md`](./2026-05-07-phase1-design-stages-plan.md) — PROCEDURE template established.
- [`2026-05-07-phase2-implementation-plan-stages-plan.md`](./2026-05-07-phase2-implementation-plan-stages-plan.md) — pattern reuse precedent.

## Scope (7 skills)

| Skill name | 1줄 용도 | 의존 (입력) |
|-----------|---------|------------|
| `run-uat` | UAT scenario 실행 + go/no-go 판단 + acceptance evidence 수집 | feature spec + acceptance test plan (§4) |
| `run-beta-program` | 클로즈드 beta cohort 운영 + 피드백 수집 + GA gating | UAT pass + cohort 정의 |
| `setup-canary-deploy` | canary stage 비율 + metric gate + auto-promote/rollback policy | tech stack (hosting), SLO |
| `setup-feature-flags` | flag system 설계 + kill switch + targeting + cleanup 정책 | tech stack + flag system 결정 |
| `setup-rollback-runbook` | rollback decision tree + 실행 절차 + verification | deploy strategy + observability |
| `prepare-launch-checklist` | launch readiness gate (security / legal / docs / monitoring 17+ 항목) | quality gate + UAT + beta + canary plan |
| `setup-incident-paging` | on-call rotation + escalation policy + alert wiring + runbook 인덱스 | observability + alert thresholds |

**왜 이 7개:** 사고 발생 비용이 가장 높은 영역. ship-release orchestrator 의 stage 5/6 (UAT, beta) 가 현재 bracketed pending 으로 표시 — orchestrator 가 직접 수행 중인 항목을 stage skill 로 분리하면 (a) 호출 explicit, (b) verification gate enforcement, (c) anti-pattern 방지. 나머지 5개 (canary / flags / rollback / launch / paging) 는 orchestrator 가 아예 다루지 않는 빈 영역.

**Use case cascade 활성화:**
- §6 quality gate pass → **§7 safety net 7 stage** → §8 production traffic 분석
- 본 phase 가 §6 → §8 의 missing safety layer 채움.

## Out of scope

- 실제 외부 SaaS 통합 (Datadog incident, PagerDuty, LaunchDarkly, Statsig API call) — 본 skill 은 plan / decision / runbook 산출, integration 코드는 §5 build-feature 가 별도 task 로 처리
- canary deploy 의 실제 traffic shift 자동화 (CI/CD pipeline level) — runbook 만 작성, 실행은 ship-release 후속 stage
- 신규 phase 추가 (Q7=(b) §7.5 분리) — 현재 §7 안의 stage 로 흡수
- §6 의 보강 audit (load test, a11y, i18n, cost, chaos, mutation) — Phase 7 로 deferred

## 진행 원칙

- **순차 작성** — Task 3.1 (run-uat) 가 첫 skill, ship-release orchestrator 의 bracketed pending 을 가장 먼저 해소. 후속 6 skill 은 안전망 dimension 별 독립 (서로 입력 의존 약함) 이라 순서 자유, 본 plan 의 Task 번호 = 추천 순서.
- **각 skill = 단일 commit** (per task)
- **PROCEDURE template 재사용** — Phase 1 의 §0 STOP / §4 posture / §6 explicit output / §11 verification gate 4 enhancement 그대로 적용. PROCEDURE 본문은 frontmatter 없이 §0 ~ §11 의 12 sections.
- **side-effecting 표현 금지** — 본 skill 들은 모두 plan / runbook / checklist / decision 산출 (read-only). 실제 deploy / paging / flag toggle 은 별도 §5 build / §7 ship-release 후속.

---

## PROCEDURE template

Phase 1 plan 의 [common template](2026-05-07-phase1-design-stages-plan.md#proceduremd-공통-템플릿) 그대로 사용. 모든 §7 skill 은 frontmatter 없이 §0 ~ §11 의 12 sections 를 가진다:

- §0 STOP — anti-pattern 5+ enumeration
- §1 목적
- §2 사용 시점 (When to invoke)
- §3 입력 (필수 / 선택 / forcing question)
- §4 핵심 원칙 (posture + 도메인 원칙)
- §5 단계 (Phases)
- §6 산출물 형식 (structured output 강제)
- §7 Cross-phase cascade
- §8 다음 skill
- §9 다른 skill 과의 경계
- §10 중요 규칙
- §11 Verification gate (self-check)

---

## Tasks

### Task 3.1: `run-uat` — UAT orchestration

**Files:**
- Create: `plugin/skills/run-uat/PROCEDURE.md`
- Create: `plugin/commands/run-uat.md`
- Modify: `plugin/skills/router/references/skill-catalog.md` (§7 row 추가)
- Modify: `plugin/skills/ship-release/PROCEDURE.md` (stage 5 의 bracketed pending → [Done] 표기 + 호출 라인)

**Skill 정의:**
- description: "UAT scenario 실행 + go/no-go 판단 — acceptance test plan 의 critical flow 를 stakeholder 가 verify 한 결과 + critical bug 0 개 gate."
- when-to-use: §6 quality gate pass 직후 / production deploy 전 stakeholder sign-off 단계 / acceptance criteria 변경 후 재실행
- inputs: feature spec (per-actor acceptance), §4 define-acceptance-test-plan 산출물, UAT cohort 정의 (PM + 1 ~ 3 designated user)
- outputs: UAT scenario 표 (각 시나리오 / 실행자 / 결과 / evidence), critical bug 리스트, go/no-go 판단 + rationale, acceptance evidence (스크린샷 / 로그 link)
- 충돌 방지: vs `run-browser-qa` (§6 — automated browser QA, 본 skill 은 human acceptance) / vs `run-beta-program` (그것은 실 사용자 cohort, 본 skill 은 designated stakeholder)

**§5 phase 골자:**
1. **UAT scenario 정의** — define-acceptance-test-plan 의 cross-actor flow 8 + per-actor critical 5 → ~13 scenario
2. **실행자 배정** — PM / Tenant Admin / 1~3 designated user, 시나리오 별 owner
3. **Evidence 수집 protocol** — 각 시나리오의 step-by-step 실행 + screenshot/recording + 시간 기록
4. **Critical bug triage** — severity 분류 (blocker / major / minor) + 수정 sprint 결정 + go/no-go 영향
5. **Go/No-go decision** — pass criteria: 0 blocker + 0 major (또는 같은 sprint 내 fix 가능 명시) + UAT 참여자 ≥ 80% sign-off

**Acceptance:** 표준 (PROCEDURE/command 신규 + catalog/orchestrator 갱신 + CI 88→89)

**Commit:** `feat(skill): add run-uat §7 stage skill`

---

### Task 3.2: `run-beta-program` — 클로즈드 beta cohort 운영

**Files:**
- Create: `plugin/skills/run-beta-program/PROCEDURE.md`
- Create: `plugin/commands/run-beta-program.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/ship-release/PROCEDURE.md` (stage 6 의 bracketed pending 해소)

**Skill 정의:**
- description: "클로즈드 beta cohort (5-20 early adopter) 운영 + structured 피드백 수집 + GA gating — 실 traffic 안전망의 마지막 layer 전 단계."
- when-to-use: UAT pass 후 GA 직전 / 신규 critical feature / API breaking change 출시 / 산업 규제 (의료/금융) 영역
- inputs: UAT 산출물, beta cohort 정의 (사용자 selection criteria + 규모), beta 기간 (1-4 주), 피드백 채널 (Slack / form / interview)
- outputs: cohort table (참여자 + segment + 동의 status), structured feedback corpus, critical issue 리스트 + triage, GA go/no-go + condition
- 충돌 방지: vs `run-uat` (UAT = designated stakeholder, beta = real users) / vs `analyze-customer-feedback-corpus` (§8 — production-scale corpus 분석, 본 skill 은 closed beta period 한정)

**§5 phase 골자:**
1. **Cohort selection** — segment criteria (size / industry / risk tolerance / NDA status) + recruitment + onboarding
2. **Beta period 설계** — 기간 (1/2/4 주 옵션) + milestone (week 1: setup, week 2: usage, week 3+: feedback)
3. **Feedback channel 정의** — 동기 (interview, NPS) + 비동기 (form, Slack DM) + 자동 (telemetry opt-in)
4. **Structured triage** — feedback 분류 (bug / feature / UX / business model), severity, GA blocker 판단
5. **GA gating decision** — pass criteria: critical bug 0, beta NPS ≥ +20 (또는 dispute resolution), participation rate ≥ 60%

**Acceptance:** 표준 + cohort consent / NDA template 참조 (별도 파일 생성 안 함, plan 안에 형식만 명시)

**Commit:** `feat(skill): add run-beta-program §7 stage skill`

---

### Task 3.3: `setup-canary-deploy` — canary rollout strategy

**Files:**
- Create: `plugin/skills/setup-canary-deploy/PROCEDURE.md`
- Create: `plugin/commands/setup-canary-deploy.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/ship-release/PROCEDURE.md` (7-3 GA Release 섹션에 stage 추가)

**Skill 정의:**
- description: "canary deploy 단계 비율 + metric gate + auto-promote / rollback 정책 — staged rollout 으로 blast radius 제한."
- when-to-use: production deploy 직전 / breaking change 출시 / SLO sensitivity 높은 endpoint 배포 / 1k+ user system
- inputs: tech stack ADR (hosting platform — Fargate / Lambda / Kubernetes 별 canary 메커니즘 다름), SLO (latency / error rate / saturation), traffic 규모, observability stack
- outputs: canary stage table (1% → 5% → 25% → 100% 단계 비율 + dwell time), metric gate (어느 metric 이 어느 threshold 위반 시 stop / rollback), auto-promote condition, manual override protocol
- 충돌 방지: vs `setup-feature-flags` (flags = code-level toggle in same deploy, canary = traffic-level routing across deploys) / vs `setup-rollback-runbook` (canary 가 자동, rollback 이 escape valve)

**§5 phase 골자:**
1. **Stage 비율 + dwell time** — 1% (15min) → 5% (1h) → 25% (4h) → 100%, 각 stage 의 traffic 노출량 + 모니터링 시간
2. **Metric gate 정의** — error rate, p99 latency, saturation, business KPI (예: signup success rate). 각 metric 의 threshold + 비교 baseline (이전 deploy 또는 control group)
3. **Auto-promote vs auto-rollback 결정** — green metric → next stage, red metric → halt + page on-call. 단계별 manual override 권한자.
4. **Hosting platform 별 implementation** — ECS+ALB weighted target group / Lambda alias weighted routing / k8s Argo Rollouts / Kong canary 등
5. **Verification protocol** — canary 단계마다 SLO compliance 확인 + business metric 검증 + drift detection

**Acceptance:** 표준 + canary stage table 의 비율 + dwell time + metric gate 구체 명시 (synthetic example: SaaS auth 의 p99 login 500ms threshold + 0.1% error rate)

**Commit:** `feat(skill): add setup-canary-deploy §7 stage skill`

---

### Task 3.4: `setup-feature-flags` — flag system 설계

**Files:**
- Create: `plugin/skills/setup-feature-flags/PROCEDURE.md`
- Create: `plugin/commands/setup-feature-flags.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/ship-release/PROCEDURE.md`

**Skill 정의:**
- description: "feature flag system 설계 + kill switch + targeting rule + flag lifecycle (cleanup) 정책."
- when-to-use: A/B test 또는 staged rollout 도입 시 / risky feature merge-then-launch 패턴 / kill switch 필요 (incident response) / dark launch
- inputs: flag system 결정 (LaunchDarkly / Statsig / Unleash / 자체 구현 — 사용자 결정), targeting axes (user / tenant / region / cohort), 운영 빈도, lifecycle 정책
- outputs: flag inventory schema (name, type, default, owner, sunset date), targeting rule 형식, kill switch 패턴, cleanup 정책 (stale flag 식별 + removal SLA), governance (누가 toggle 권한)
- 충돌 방지: vs `setup-canary-deploy` (canary = traffic routing, flags = code branch) / vs `design-ab-experiment` (§8 — A/B 통계 분석 설계, 본 skill 은 flag 인프라 설계) / vs `compose-safety-mode` (그것은 build-time safety, 본 skill 은 runtime feature toggle)

**§5 phase 골자:**
1. **Flag taxonomy** — release flag (limited TTL) / experiment flag (A/B) / ops flag (kill switch, indefinite) / permission flag (long-term)
2. **Targeting rule 모델** — 단순 (% rollout) / segment (user attribute) / hierarchical (tenant → user override) — 시스템 capability 매핑
3. **Kill switch protocol** — 어떤 flag 가 kill switch 자격, 활성 누구 권한, 활성화 후 SLA, 활성화 시 metric monitor
4. **Cleanup policy** — release flag 의 max TTL (예: 90 일), stale flag 자동 alert, removal PR 생성 트리거
5. **Governance** — flag 추가 시 owner / 만료 / 영향 범위 명시 의무, 정기 audit (분기)

**Acceptance:** 표준 + flag inventory schema YAML 예시 포함 + cleanup SLA 정량

**Commit:** `feat(skill): add setup-feature-flags §7 stage skill`

---

### Task 3.5: `setup-rollback-runbook` — rollback decision tree + 절차

**Files:**
- Create: `plugin/skills/setup-rollback-runbook/PROCEDURE.md`
- Create: `plugin/commands/setup-rollback-runbook.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/ship-release/PROCEDURE.md`

**Skill 정의:**
- description: "rollback decision tree (언제 rollback / 언제 forward fix) + 실행 절차 + verification — incident response 의 핵심 도구."
- when-to-use: production deploy 직전 / 신규 service launch / migration 포함 deploy / 신규 데이터 schema 변경
- inputs: deploy strategy (canary / blue-green / rolling), schema migration 유무, observability stack (어떤 signal 로 rollback trigger 판단), data state 관리 정책
- outputs: rollback decision tree (signal → action), 실행 step-by-step (image revert / migration revert / cache invalidation / config rollback / DNS rollback), verification protocol (rollback 후 어떻게 health 확인), data integrity 처리 (forward-only migration 의 경우)
- 충돌 방지: vs `setup-canary-deploy` (canary 가 자동 rollback 일부 처리, 본 skill 은 manual + complex case) / vs `handle-incident` (§8 — incident triage, 본 skill 은 rollback 도구 사전 준비) / vs `automate-release-tagging` (그것은 forward, 본 skill 은 backward)

**§5 phase 골자:**
1. **Rollback trigger signal** — error rate spike, p99 latency, business KPI drop, security alert. 각 signal 의 threshold + observation window.
2. **Decision tree** — signal 발생 → forward fix possible? (HOTFIX in <30min) → no → rollback. data corruption 의심? → freeze + investigate. partial degradation only? → feature flag kill.
3. **실행 step-by-step** — pre-rollback (notify on-call, snapshot logs), execute (image revert / Argo rollout undo / Lambda alias swap / DNS), verify (smoke test + SLO check), post-rollback (incident channel update, postmortem trigger)
4. **Schema migration handling** — forward-only migrations: rollback impossible without data loss → expand-contract 강제 (이미 design-data-model 에서 권장). 잘못된 migration deploy 시: backfill 또는 reverse migration. 본 skill 이 expand-contract pattern 의 deploy-time enforcement 명시.
5. **Verification + post-mortem trigger** — rollback 성공 = SLO recovery within X min, post-mortem 자동 trigger (24h 안에 conduct-postmortem 호출)

**Acceptance:** 표준 + decision tree mermaid + step-by-step 의 platform 별 (Fargate / Lambda / k8s) 변형 명시

**Commit:** `feat(skill): add setup-rollback-runbook §7 stage skill`

---

### Task 3.6: `prepare-launch-checklist` — launch readiness gate

**Files:**
- Create: `plugin/skills/prepare-launch-checklist/PROCEDURE.md`
- Create: `plugin/commands/prepare-launch-checklist.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/ship-release/PROCEDURE.md` (Go-Live Readiness Checklist 섹션을 본 skill 호출로 대체 또는 보강)

**Skill 정의:**
- description: "launch readiness 17+ 항목 gate (security / legal / docs / monitoring / rollback / paging / cost / a11y / i18n / SLA / pricing / support / messaging) — GA 직전 final check."
- when-to-use: GA 직전 / 신규 product launch / 신규 region expansion / breaking change major version
- inputs: §6 quality gate 결과, §7 의 다른 6 skill 산출물 (canary / flags / rollback / UAT / beta / paging), legal/compliance review status, marketing readiness
- outputs: launch checklist (17+ row), 각 row 의 status (green / yellow / red) + owner + evidence link, blocker 리스트 + ETA, go/no-go 권고
- 충돌 방지: vs orchestrator 의 inline checklist (그것은 7 항목 high-level, 본 skill 은 17+ 의 세분화 + ownership + evidence) / vs `setup-quality-gates` (quality = automated CI, launch = cross-functional sign-off)

**§5 phase 골자:**
1. **Category 정의** — 6 axis: Engineering (5 항목) / Security & Compliance (4) / Operations (4) / Product (3) / Legal & Comms (4) / Cost & Business (3) — 각 axis 내부 세분
2. **Per-row status 추출** — 각 항목 의 evidence (test report / sign-off / runbook link / SLA doc) 와 owner. 자동 추출 가능한 것 (CI status, k6 result) 과 manual sign-off 분리.
3. **Cross-skill cascade** — 본 skill 이 §7 의 다른 6 skill 의 산출물을 input 으로 받아 통합. canary plan, rollback runbook, paging, UAT, beta 의 status 가 row 로 매핑.
4. **Blocker triage** — red row 의 blocker / mitigation / ETA. yellow row 의 risk acceptance.
5. **Go/no-go 권고** — pass criteria: green ≥ 90%, red 0, yellow ≤ 10% with documented dispositions.

**Acceptance:** 표준 + 17+ row checklist template (synthetic SaaS auth 예시로 ~20 row)

**Commit:** `feat(skill): add prepare-launch-checklist §7 stage skill`

---

### Task 3.7: `setup-incident-paging` — on-call + escalation

**Files:**
- Create: `plugin/skills/setup-incident-paging/PROCEDURE.md`
- Create: `plugin/commands/setup-incident-paging.md`
- Modify: `plugin/skills/router/references/skill-catalog.md`
- Modify: `plugin/skills/ship-release/PROCEDURE.md`

**Skill 정의:**
- description: "on-call rotation + escalation policy + alert wiring + runbook 인덱스 — production incident 의 first response 구조."
- when-to-use: production launch 직전 / SLO 정의 직후 / 신규 oncall 멤버 onboarding / paging tool 변경
- inputs: 팀 composition (on-call 가능 인원), SLO + alert thresholds (§6 또는 observability 산출), incident severity 분류, paging tool (PagerDuty / Opsgenie / VictorOps / 자체)
- outputs: rotation schedule (primary / secondary / weekend / holiday), escalation policy (10 min ack → secondary → 20 min → manager), severity 분류표 (SEV1-4), alert routing (어느 alert → 어느 oncall layer), runbook 인덱스 (alert 별 runbook link)
- 충돌 방지: vs `handle-incident` (§8 — actual triage 후 절차, 본 skill 은 paging infrastructure 사전 준비) / vs `conduct-postmortem` (§8 — post-incident 회고, 본 skill 은 pre-incident readiness)

**§5 phase 골자:**
1. **Rotation 설계** — primary / secondary 구분, shift 길이 (24h / 7day), holiday 처리, backup pool, fairness 검증 (per-person hours/quarter)
2. **Severity 정의 + escalation** — SEV1 (full outage / data loss): page within 5 min, manager 즉시 / SEV2 (degraded): page primary, ack 10 min / SEV3 (minor): in-hours only / SEV4 (informational): no page
3. **Alert routing matrix** — alert source (CloudWatch / Datadog / OTel) → severity inference rule → paging layer. flap suppression / dedup window.
4. **Runbook 인덱스** — 각 alert 마다 1 runbook link (handle-incident 의 input). orphan alert (runbook 없는 alert) 0 maintain.
5. **Drill** — 월 1회 paging drill (synthetic alert), 분기 1회 chaos drill, 모든 oncall 멤버가 6 개월 내 first-page 경험. drill 결과로 escalation rule 보정.

**Acceptance:** 표준 + rotation schedule + severity matrix + alert routing 예시

**Commit:** `feat(skill): add setup-incident-paging §7 stage skill`

---

### Task 3.8: `ship-release` orchestrator wiring

**Files:**
- Modify: `plugin/skills/ship-release/PROCEDURE.md` (stage 흐름 갱신, 7 신규 stage 통합)

**Background:** 현재 orchestrator 는 9 stage 인데 그 중 stage 5 (`run-uat`) + stage 6 (`run-beta-program`) 가 bracketed pending. Phase 3 완료 시 두 stage 가 [Done] + 추가 5 skill (`setup-canary-deploy`, `setup-feature-flags`, `setup-rollback-runbook`, `prepare-launch-checklist`, `setup-incident-paging`) 도 stage 흐름 안에 통합되어야 함.

**Decision:** orchestrator 의 9 stage 구조를 다음과 같이 갱신:

```
ship-release (7단계 phase orchestrator)
├── 7-1 단계 Release Preparation
│   ├── stage 1: setup-quality-gates    [Done]
│   ├── stage 2: write-changelog        [Done]
│   ├── stage 3: sync-release-docs      [Done]
│   └── stage 4: auto-create-pr         [Done]
├── 7-2 단계 Pre-Launch Safety Nets (Phase 3 신규)
│   ├── stage 5: setup-feature-flags    [Done] flag 인프라 사전 설계
│   ├── stage 6: setup-canary-deploy    [Done] canary 단계 + metric gate
│   ├── stage 7: setup-rollback-runbook [Done] rollback decision tree
│   ├── stage 8: setup-incident-paging  [Done] on-call rotation + alert wiring
│   └── stage 9: prepare-launch-checklist [Done] 17+ readiness gate
├── 7-3 단계 Beta / UAT
│   ├── stage 10: run-uat               [Done] designated stakeholder UAT
│   └── stage 11: run-beta-program      [Done] 클로즈드 beta cohort
└── 7-4 단계 GA Release
    ├── stage 12: automate-release-tagging  [Done]
    ├── stage 13: guard-destructive-commands [Done]
    └── stage 14: compose-safety-mode       [Done]
```

stage 수: 9 → 14. ordering rationale: safety net 인프라 (5-9) 가 UAT/beta (10-11) 전에 ready 되어야 — UAT 시점에 rollback runbook + paging 이 작동 검증되어야 GA 안전.

**Acceptance:** ship-release/PROCEDURE.md 의 stage 흐름 표 갱신, Go-Live Readiness Checklist 섹션이 prepare-launch-checklist 호출로 redirect

**Commit:** `feat(skill): wire 7 new Phase 3 §7 safety net skills into ship-release orchestrator`

---

### Task 3.9: v1.0.4 release

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json` (version: 1.0.3 → 1.0.4)
- Modify: `.claude-plugin/marketplace.json` (version sync, description 업데이트 — "78 procedures" → "95 procedures")
- Modify: `CHANGELOG.md` (1.0.4 entry 추가)
- Modify: `scripts/test-router-wireup.sh` (PROCEDURE.md count check 88 → 95)

**Background:** Phase 3 의 7 skill 추가로 PROCEDURE.md 88 → 95, command md 40 → 47. version 1.0.3 → 1.0.4 (MINOR 후보 — 신규 7 skill 은 backward-compat 한 추가, breaking 없음).

**Steps:**
- [ ] plugin.json + marketplace.json version 일치 (1.0.4)
- [ ] CHANGELOG 의 1.0.4 entry: 7 신규 skill 명 + ship-release orchestrator wiring + count 갱신
- [ ] CI smoke test 의 PROCEDURE count 95 로 갱신
- [ ] `claude plugin validate` 통과
- [ ] `bash scripts/test-router-wireup.sh` 10/10 통과

**Acceptance:**
- `find plugin/skills -name PROCEDURE.md | wc -l` = 95
- `find plugin/commands -name "*.md" | wc -l` ≥ 47
- CI script 통과
- plugin validate 통과

**Commit:** `release: v1.0.4 (Phase 3 §7 release safety nets — 7 new stage skills)`

---

## Verification (Phase 3 종료 시)

- [ ] 7 신규 skill 모두 PROCEDURE.md 존재 (frontmatter 없음, §0~§11 12 sections)
- [ ] 7 신규 command md 가 router single-mode dispatch 호출 패턴 일치
- [ ] skill-catalog.md §7 section 에 7 row 추가
- [ ] ship-release orchestrator 의 stage 14 모두 [Done] 표기
- [ ] spec §3 §7 row 의 status 가 [Partial] → [Done] (또는 일부 Pending 잔여 시 명시)
- [ ] CI script PROCEDURE count 95 통과
- [ ] plugin.json validate 통과
- [ ] `Makefile install-plugin` (또는 사용자 install 경로) 동작 확인 (manual 또는 CI)

## Rollback

각 7 skill task = 단일 commit, 단순 `git revert` 로 안전 롤백. orchestrator wiring (Task 3.8) 도 단일 파일 수정. v1.0.4 release commit 만 별도 (version bump + CHANGELOG + count) — release commit revert 시 plugin 은 1.0.3 으로 복귀하지만 신규 PROCEDURE 파일은 잔존 (정합성 문제 — orphaned skill). 따라서 rollback 순서 권장: release revert → orchestrator revert → 7 skill commits revert (역순).

## 다음 plan (Phase 3 완료 후)

[parent plan](./2026-05-06-stage-buildout-plan.md) 의 Phase 4 (§6 use-case 테스트 stage 2 skills: `test-per-actor-use-case`, `test-cross-actor-flow`) 가 다음 plan 작성 대상. Phase 4 가 Q8=(a) cascade 의 마지막 puzzle piece — §2 use case → §3 system → §4 actor track → §5 build → **§6 actor-별 테스트** 로 흐름 닫힘.
