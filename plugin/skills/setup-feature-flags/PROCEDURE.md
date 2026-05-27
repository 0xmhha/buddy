# setup-feature-flags — flag system + kill switch + targeting + cleanup

§7 Release & Beta phase 의 stage. **code-level toggle 인프라** — release flag (TTL 있는 staged rollout), experiment flag (A/B test), ops flag (kill switch, indefinite), permission flag (long-term entitlement) 4 taxonomy 을 정의 + targeting rule + flag lifecycle (cleanup) 정책 산출. canary deploy (traffic-level) 와 함께 두 layer 의 toggle 인프라 구성. 산출물은 flag inventory schema + targeting model + kill switch protocol + cleanup SLA + governance policy.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Feature 목록 | ✅ | knowledge | 사용자 도메인 지식 | "feature flag로 관리할 기능 목록을 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Feature flag 구성 (kill switch + targeting + lifecycle) | artifact | 설정 파일 | `ship-release` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Flag taxonomy 미분류** — 모든 flag 가 동일하게 취급. release / experiment / ops / permission 분류 안 하면 cleanup SLA / 권한 / TTL 모두 모호.
- **Cleanup policy 없음** — flag 추가만 하고 제거 안 함 → 6 개월 후 stale flag 100+ 개 (codebase 부패).
- **Owner / sunset date 없는 flag** — orphan flag → 누가 toggle 권한, 언제 삭제 할 지 모름.
- **Kill switch 권한 분산** — "누구나 toggle 가능" → incident 시 의도치 않은 활성화 / 비활성화 위험. 권한 명시 의무.
- **Targeting rule 단일** — 단순 % rollout 만 → segment / hierarchical 케이스 지원 못 함, 운영 후 강제 도입 시 마이그레이션 비용.
- **Flag system 결정 없이 시작** — LaunchDarkly / Statsig / Unleash / 자체 구현 중 결정 없이 진행 → 첫 flag 추가 시 설계 redo.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

code branch toggle 의 인프라화. canary (traffic-level) 와 별도 layer 로 code path 를 runtime 에 toggle 가능하게 만들어 (a) staged rollout 가능 (b) kill switch 가능 (c) A/B test 가능 (d) 단계별 entitlement 가능. flag 자체의 lifecycle (생성 → targeting → cleanup) 까지 governance.

## 2. 사용 시점 (When to invoke)

- A/B test 도입 시 (experiment flag)
- staged rollout 패턴 채택 시 (release flag)
- kill switch 필요 시 (ops flag — incident response 도구)
- dark launch 패턴 시 (release flag with 0% targeting)
- entitlement / pricing tier 분리 시 (permission flag)
- flag system 변경 (자체 → SaaS 또는 vice versa) 시 마이그레이션

## 3. 입력 (Inputs)

### 필수
- flag system 결정 (사용자 결정 필요): LaunchDarkly / Statsig / Unleash / Flagsmith / 자체 구현 (Postgres + Redis cache) — 5 옵션 중 1
- 운영 빈도 (monthly flag count: low <5, medium 5-50, high 50+)
- targeting axes (user / tenant / region / cohort / version)
- 팀 governance 모델 (어느 role 이 flag toggle 권한)

### 선택
- 기존 flag 재고 (legacy flags 마이그레이션 대상)
- compliance 요구 (GDPR 의 user 별 flag opt-out, audit log 의무)
- regional restrictions (EU only feature 등)
- billing tier 결정 (paid plan 별 entitlement)

### 입력이 부족할 때 forcing question
- "flag system 결정이 안 됐다면 운영 빈도 + 예산 + 팀 capability 로 결정 — LaunchDarkly $X/seat/mo 인지 자체구현 한 번에 끝낼지."
- "monthly flag count 가 low (< 5) 인데 LaunchDarkly 면 over-engineering, high 인데 자체구현이면 under-engineering."
- "flag toggle 권한자가 engineer 만 인가, PM / on-call 도 가능한가? on-call 의 kill switch 권한 없으면 incident response 늦어짐."
- "stale flag 의 max TTL 은? 90 일 / 180 일 / 1 년 — 기준 없으면 cleanup deferred forever."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — flag system / TTL / 권한 모두 단호. "필요하면 자체구현" 거부, "monthly flag < 5 → LaunchDarkly Free / 5-50 → LaunchDarkly Pro / > 50 → Statsig 또는 Unleash self-hosted" 형식.
- **사용자 입력을 challenge** — "모든 feature 를 flag 뒤로" 발화에 "flag 의 기술 부채는? cleanup SLA 는?" push.
- **Specificity 강제** — "필요할 때 cleanup" 거부, "release flag 90 일 TTL, 60 일 시 alert, 90 일 시 auto-PR remove".

도메인 원칙:

1. **Flag taxonomy 4 분류** — release / experiment / ops / permission. 분류 별 TTL / 권한 / cleanup 정책 다름.
2. **Owner + sunset date 의무** — 모든 flag 에 owner (engineer email) + sunset date (release/experiment) 또는 "indefinite" + audit cadence (ops/permission).
3. **Cleanup SLA 정량** — release flag max 90 일, experiment max 30 일 (test 종료 후), ops 분기 audit, permission 연 audit.
4. **Targeting model 계층** — % rollout (단순) / segment (user attribute) / hierarchical (tenant override user). 시스템이 hierarchical 까지 지원하는지 확인.
5. **Kill switch 권한 명시** — on-call engineer 단독 (single-person, fast response) + audit log + 24h post-mortem.
6. **Governance 정기 audit** — 분기 1회 flag inventory review (orphan / stale / 의무 위반).

## 5. 단계 (Phases)

### Phase 1. Flag taxonomy 정의

| Type | Definition | TTL | Owner | Cleanup trigger |
|------|------------|-----|-------|-----------------|
| **Release flag** | staged rollout / dark launch — 새 feature 의 점진적 활성 | **max 90 일** | feature owner (engineer) | 100% rollout 후 14 일 idle → auto-PR remove |
| **Experiment flag** | A/B test — control vs treatment | **max 30 일** post test 종료 | data scientist + PM | test 종료 후 ship/revert decision → 즉시 cleanup |
| **Ops flag** | kill switch — runtime feature disable for incident | **indefinite** | platform team | 분기 audit (사용 history 0 이고 < 1 년 전 추가 시 cleanup 검토) |
| **Permission flag** | entitlement — pricing tier / role-based access | **indefinite** | product / billing | 연 audit (deprecated tier 제거 시 동시 cleanup) |

각 type 별 default 권한, audit cadence, 명명 prefix (예: `release_*`, `exp_*`, `ops_*`, `perm_*`).

### Phase 2. Flag system 선택

| System | Pricing | Capabilities | Maintenance | Best for |
|--------|---------|--------------|-------------|----------|
| **LaunchDarkly** | $X/seat/mo, free starter | comprehensive (targeting / streaming / SDK 다양), audit, governance UI | external SaaS, no self-host | medium-high flag count (5+) + multiple SDK |
| **Statsig** | usage-based, free tier | flag + experiment + analytics 통합 | external SaaS | A/B test 비중 높을 때 |
| **Unleash** | self-host (Apache 2 OSS), Pro ($/team) | targeting / segments / strategies | self-host (Postgres) or SaaS | OSS preference, EU compliance |
| **Flagsmith** | OSS or hosted | similar to Unleash | self-host or SaaS | OSS preference |
| **자체 구현 (Postgres + Redis cache)** | 0 (개발 비용) | basic % rollout + segment | 장기 maintenance 부담 | low flag count (< 5) + 외부 의존 회피 needs |

선택 rationale: monthly flag count + 예산 + compliance + 팀 capability.

본 SaaS auth example (cascade): tech stack 의 AWS 만 + monthly flag 추정 5-10 → **LaunchDarkly Free (or Pro $20/seat/mo for 4 seat = $80/mo)** 권장. 단 EU compliance 시 self-hosted Unleash 검토.

### Phase 3. Targeting rule 모델

| Strategy | Definition | Use case |
|----------|------------|----------|
| **Boolean (on/off)** | global toggle | dark launch, kill switch |
| **Percentage rollout** | random % of all users | release flag staged rollout |
| **User segment (attribute)** | match user attribute (email domain, plan, region) | beta cohort targeting, EU-only feature |
| **Tenant segment** | match tenant attribute (plan, industry) | enterprise-only feature |
| **Hierarchical (tenant → user)** | tenant default + user override | enterprise admin grants per-user |
| **Sticky bucket** | hash(user_id) for consistent assignment | A/B test (no flapping) |
| **Schedule** | time-based activation (start at 2026-XX-XX) | scheduled feature launch |

targeting evaluation:
- client-side SDK or server-side
- streaming updates (LaunchDarkly / Statsig) vs polling (Unleash basic)
- cache TTL (typically 30-300s)

### Phase 4. Kill switch protocol

| Aspect | Decision |
|--------|----------|
| Authorized roles | on-call engineer (primary) + platform lead (secondary) |
| Decision SLA | < 5 min (incident response context) |
| Activation evidence | incident ticket id + reason (1 line) — recorded in flag system audit log + #incidents Slack channel auto-post |
| Post-activation monitoring | metric watch for 30 min (does kill switch actually mitigate? 다른 issue 노출 여부) |
| Post-mortem trigger | 24h 내 conduct-postmortem 자동 호출 (kill switch activation = incident-grade) |
| Re-activation gate | flag system 의 4-eye review (engineer + lead) + condition (root cause fix merged) 후 재활성 |

ops flag inventory (kill switch 자격):
- `ops_signup_disable` — signup 일시 중단 (DB write storm, 외부 API outage)
- `ops_email_disable` — email send 중단 (mail provider 장애)
- `ops_admin_audit_disable` — admin endpoint 일시 차단 (보안 incident)
- (확장 시 동일 패턴)

### Phase 5. Cleanup policy + governance

cleanup automation:

| Trigger | Action | Tool |
|---------|--------|------|
| Release flag 60 일 (75% TTL) | Slack DM to owner | flag system webhook → CI cron |
| Release flag 90 일 (100% TTL) | auto-PR to remove flag (codemod) | dependabot-style PR |
| Experiment flag, test 종료 시 | ship/revert decision required → 14 일 grace → auto-PR | data team + flag system integration |
| Ops flag, 사용 history 0 + 1 년 경과 | 분기 audit 시 review → keep / remove | manual review |
| Permission flag, deprecated tier 제거 시 | 연 audit 동시 cleanup | manual |

governance:
- 분기 audit (PM + platform lead): orphan flag / stale flag / 의무 위반 (no owner / no sunset)
- 연 audit (이상 + product): permission flag deprecation
- 신규 flag 추가 시 PR template 강제 (owner / sunset / type / targeting 모두 채워야 merge 가능 — CI lint)
- flag count dashboard: monthly review 에서 "active vs scheduled-cleanup" ratio 추적 (active > 50% 가 정상)

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## setup-feature-flags Output — <project name>

### Summary
<3 줄: 선택 system / 4 taxonomy / kill switch flag count / cleanup SLA>

### Flag System Decision
| Field | Value |
|-------|-------|
| System | LaunchDarkly Pro (4 seat × $20/mo = $80/mo) |
| Rationale | monthly flag estimate 5-10, multi-SDK (TS / future mobile), audit + governance UI |
| Self-host alternative | Unleash (EU compliance trigger, self-host Postgres) |
| Migration plan (future) | LaunchDarkly export → Unleash import via OpenFeature spec compatibility |

### Flag Taxonomy
| Type | Naming prefix | TTL | Owner role | Cleanup trigger |
|------|---------------|-----|------------|-----------------|
| Release | `release_*` | 90 d | feature owner | 100% rollout + 14 d idle → auto-PR |
| Experiment | `exp_*` | 30 d post test 종료 | data + PM | test 종료 + 14 d → auto-PR |
| Ops | `ops_*` | indefinite | platform team | quarterly audit |
| Permission | `perm_*` | indefinite | product / billing | annual audit |

### Targeting Model
| Strategy | Supported | Use case |
|----------|-----------|----------|
| Boolean | yes | dark launch, kill switch |
| Percentage | yes | release flag staged |
| User segment | yes | beta, EU-only |
| Tenant segment | yes | enterprise-only |
| Hierarchical | yes (tenant default + user override) | enterprise admin per-user |
| Sticky bucket | yes (user_id hash) | A/B test |
| Schedule | yes | scheduled launch |

### Kill Switch Inventory
| Flag | Purpose | Authorized roles | SLA |
|------|---------|------------------|-----|
| `ops_signup_disable` | signup 일시 중단 (DB storm / external outage) | on-call engineer, platform lead | < 5 min |
| `ops_email_disable` | email send 중단 (mail provider outage) | on-call engineer, platform lead | < 5 min |
| `ops_admin_audit_disable` | admin endpoint 차단 (security incident) | on-call engineer, platform lead, security lead | < 5 min |
| (extensible per future need) | | | |

### Kill Switch Activation Protocol
| Step | Action | Evidence |
|------|--------|----------|
| 1 | on-call detects incident, decides kill switch needed | incident ticket created |
| 2 | toggle flag in LaunchDarkly UI | LD audit log auto |
| 3 | post in #incidents Slack | manual + bot |
| 4 | monitor metric for 30 min | dashboard watch |
| 5 | conduct postmortem within 24h | trigger postmortem skill |
| 6 | re-activation: 4-eye review + root cause fix | LD audit + post-mortem doc |

### Cleanup Policy
| Trigger | Action | Tool |
|---------|--------|------|
| Release flag 60 d | Slack DM owner | LD webhook → cron |
| Release flag 90 d | auto-PR remove flag (codemod) | github action |
| Experiment 종료 + 14 d | ship/revert decision → auto-PR | data + LD integration |
| Ops audit (quarterly) | review usage, remove if 0 use + 1 y old | manual |
| Permission audit (annual) | review with billing, remove deprecated tier flags | manual |

### Governance
| Cadence | Owner | Activity |
|---------|-------|----------|
| Per-PR | engineer (CI lint) | flag PR template 강제: type / owner / sunset / targeting 모두 채워야 merge |
| Quarterly | PM + platform lead | flag inventory audit — orphan / stale / violation 식별 |
| Annual | product + billing + platform | permission flag deprecation review |
| Monthly | platform | flag count dashboard review (active vs scheduled-cleanup ratio) |

### Cascade
- **§7 setup-canary-deploy**: traffic-level canary + code-level flag 두 layer 정합 — flag 의 % rollout 과 canary stage % 가 충돌 안 하게
- **§7 run-beta-program**: beta cohort flag (`release_beta_<feature>`) 가 본 skill 의 release type 으로 등록, beta 종료 시 cleanup
- **§7 setup-rollback-runbook**: kill switch (`ops_*`) 가 rollback 의 fast escape valve
- **§7 setup-incident-paging**: ops flag 활성 시 alert routing
- **§8 design-ab-experiment**: experiment flag 가 §8 의 hypothesis test 인프라
- **§8 handle-incident**: kill switch 활성이 incident 의 first-response 도구

### Next Step
<구체 action — 1줄: 예 "LaunchDarkly account 생성 + Pro plan 4 seat → SDK 통합 task 생성 (T2.X)">
```

## 7. Cross-phase cascade

- **§7 setup-canary-deploy**: traffic + flag 두 layer 정합
- **§7 run-beta-program**: beta cohort flag = release type, cleanup integrated
- **§7 setup-rollback-runbook**: kill switch = fast rollback escape
- **§7 setup-incident-paging**: ops flag activation alert routing
- **§8 design-ab-experiment**: experiment flag 인프라
- **§8 handle-incident**: kill switch first-response

## 8. 다음 skill (next in stage flow)

- `setup-canary-deploy` — traffic layer
- `setup-rollback-runbook` — escape valve protocol
- `setup-incident-paging` — kill switch alert routing
- `prepare-launch-checklist` — readiness gate

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `setup-canary-deploy`** (§7) — canary = traffic %, flag = code branch. 두 layer 모두 사용 권장. flag 가 즉시 rollback (toggle off) 가능, canary 는 traffic shift 시간 소요.
- **vs `compose-safety-mode`** (§7) — compose-safety-mode 는 build-time safety hook composition (guard-destructive-commands + freeze-edit-scope), 본 skill 은 runtime feature toggle. 둘 다 safety 이지만 layer 다름.
- **vs `design-ab-experiment`** (§8) — 그것은 통계적 A/B 설계, 본 skill 은 flag 인프라 (실험의 토대).
- **vs `setup-rollback-runbook`** (§7) — kill switch (flag) 가 fast rollback path, runbook 의 manual procedure 가 deep rollback (data state, schema). 두 layer 정합.
- **vs `automate-release-tagging`** (§7) — release tag 가 forward (semver), flag 가 runtime toggle. 직접 conflict 없음.

## 10. 중요 규칙

- **Taxonomy 4 분류 의무** — single category 거부.
- **Owner + sunset date 의무** — orphan flag 거부 (CI lint 강제).
- **Cleanup SLA 정량** — release 90d, experiment 30d post-test, ops/permission audit cadence.
- **Kill switch 권한 명시** — fast response 단일 권한자 + audit log + post-mortem.
- **Targeting hierarchical 지원 검증** — system 이 tenant → user override 가능한지 확인 (LaunchDarkly yes, 자체구현 보통 no).
- **Read-only on production code** — 본 skill 은 인프라 plan, SDK 통합 / 첫 flag 도입은 §5 build 의 task.
- **Governance 정기 audit 의무** — 분기 / 연 cadence 명시.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행 (taxonomy → system 선택 → targeting → kill switch → cleanup + governance)
- [ ] §6 의 8 출력 섹션 (Summary / System Decision / Taxonomy / Targeting / Kill Switch Inventory / Activation Protocol / Cleanup Policy / Governance / Cascade) 모두 채워짐
- [ ] Flag System Decision 의 rationale + alternatives + future migration plan 모두 명시
- [ ] Flag Taxonomy 4 type 모두 정의 (TTL / owner / cleanup)
- [ ] Targeting Model 7 strategy 모두 supported / 부분 / unsupported 명시
- [ ] Kill Switch Inventory ≥ 2 flag, 권한자 명시, SLA < 5 min
- [ ] Kill Switch Activation Protocol 6 step 정의 (detect → toggle → post → monitor → postmortem → re-activate)
- [ ] Cleanup Policy 5 trigger × action × tool 명시 — release 90 d / experiment 30 d / ops 분기 / permission 연 모두 cover
- [ ] Governance 4 cadence (per-PR / quarterly / annual / monthly) 명시
- [ ] §4 posture — system / TTL / 권한 단호, "필요하면" 류 hedge 없음
- [ ] §0 anti-pattern 부재 — taxonomy 미분류 / cleanup 없음 / orphan flag / kill switch 분산 / targeting 단일 / system 미정 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
