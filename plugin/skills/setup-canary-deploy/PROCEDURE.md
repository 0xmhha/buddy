# setup-canary-deploy — canary stage + metric gate + auto-promote/rollback policy

§7 Release & Beta phase 의 stage. production deploy 직전 / 직후 traffic-level routing 으로 blast radius 를 limit. **단계 비율 + dwell time + metric gate + auto-promote vs auto-rollback policy** 를 hosting platform 별 메커니즘에 매핑. 산출물은 canary plan + implementation hints (platform 별) + metric gate + manual override protocol.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Single-stage canary** — 100% 즉시 promote. blast radius limit 의미 없음. 최소 3 stage (1% / 25% / 100%) 권장.
- **Dwell time 0** — stage 간 즉시 promote. metric stabilization 시간 없으면 noise 가 false positive.
- **Metric gate threshold 미정** — "error rate 모니터" 만, 임계값 없음. 정량 (p99 latency ≤ 500ms, error rate ≤ 0.1%) 강제.
- **Baseline comparison 없음** — 절대값만 보고 "OK" 판정 → trend 변화 (이전 deploy 대비 +20% latency) 못 잡음.
- **Auto-rollback 미설정** — 모든 promotion 이 manual approval → 야간 / 휴일 deploy 시 oncall 부담 폭증.
- **Hosting platform 별 변형 무시** — Fargate 와 Lambda 의 canary 메커니즘 다른데 동일 절차 가정 → implementation 시점 redo.
- **Manual override 권한자 모호** — "필요 시 override" 만, 누가 권한자인지 미정 → incident 시 의사결정 지연.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

production traffic 의 staged rollout 으로 blast radius 를 cohort size 가 아닌 traffic % 로 limit. metric gate 자동화로 야간 / 휴일 deploy 의 oncall 부담 감소.

## 2. 사용 시점 (When to invoke)

- production deploy 직전 (default for new release)
- breaking change 출시 시 (필수)
- SLO sensitivity 높은 endpoint 배포 (login / payment / search)
- 1k+ user system (smaller scale 은 비용 vs benefit balance 검토)
- regulated industry (의료/금융 — staged rollout 이 audit 산출물)

## 3. 입력 (Inputs)

### 필수
- §3 tech stack ADR (hosting platform — Fargate / Lambda / Kubernetes / Vercel / Cloudflare Workers — 별 canary 메커니즘 다름)
- SLO definition (latency / error rate / saturation thresholds)
- traffic 규모 (RPS, daily user) — canary % 의 절대값 결정
- observability stack (metric source, dashboard, alarm)

### 선택
- 이전 release 의 canary metric (baseline comparison)
- business KPI (signup success, conversion) — technical SLO 외 추가 gate
- multi-region deploy 정책 (region-by-region canary 인지 global)
- traffic shaping (sticky session, regional routing)

### 입력이 부족할 때 forcing question
- "hosting platform 이 명시됐나? Fargate weighted target group / Lambda alias weighted / k8s Argo Rollouts 셋의 메커니즘이 완전 다름."
- "metric gate 의 baseline 이 (a) 이전 deploy 절대값 (b) 이전 deploy 대비 비율 변화 (c) static SLO threshold 중 어느 것?"
- "auto-rollback 의 책임자가 sole engineer 인가, oncall rotation 인가? rollback 후 incident channel 자동 alert 인가?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — stage % / dwell time / metric threshold 모두 단호. "적당히" 거부, "1% / 15min, 25% / 1h, 100% / 0" 형식.
- **사용자 입력을 challenge** — "그냥 canary 5% 부터 갈게" 발화에 "왜 1% 가 아니라 5%? blast radius 5x 차이의 근거?" push.
- **Specificity 강제** — "성능 좋으면 promote" 거부, "p99 ≤ 500ms (15min sustained) AND error rate ≤ 0.1% AND saturation < 70% → auto-promote".
- **Auto-rollback bias** — auto-promote 보다 auto-rollback 이 보수적. red metric 시 즉시 stop, manual approval 후 재시도.

도메인 원칙:

1. **단계 ≥ 3** — minimum 1% / 25% / 100%. high-risk 배포는 5 stage (1% / 5% / 25% / 50% / 100%).
2. **Dwell time 의무** — minimum 5 min 1% stage, 15-30 min 25% stage. metric stabilization 시간.
3. **Metric gate 정량** — error rate, p99 latency, saturation, business KPI 모두 threshold 명시.
4. **Baseline comparison + absolute** — both. baseline 만이면 baseline 자체가 bad case 일 때 false positive 못 잡음. absolute 만이면 trend 변화 못 잡음.
5. **Manual override 권한자 명시** — 누가 promotion 강제 / rollback 강제 권한, 의사결정 시간 (가급적 < 5 min).
6. **Platform 별 implementation hint** — Fargate / Lambda / k8s 별 actual 메커니즘 명시 (CloudFormation / Terraform / Argo manifest 예시).

## 5. 단계 (Phases)

### Phase 1. Stage 비율 + dwell time

| Stage | Traffic % | Dwell time | Notes |
|-------|-----------|------------|-------|
| 0 (control) | 0% (existing version 100%) | n/a | baseline establishment |
| 1 | 1% | 15 min | initial blast radius limit |
| 2 | 5% | 30 min | early issue detection (low-volume edge case) |
| 3 | 25% | 1 h | mainstream traffic check |
| 4 | 50% | 30 min | scale validation |
| 5 | 100% | n/a | full promote |

variations:
- **fast canary**: 1% / 5min, 25% / 15min, 100% — patch / hotfix 용
- **standard**: 위 5 stage 표준
- **slow canary**: 1% / 1h, 5% / 4h, 25% / 24h, 50% / 24h, 100% — regulated industry / data-critical

선택 rationale: traffic 규모 + risk tier (breaking / feature / patch) + business hour vs off-hour.

### Phase 2. Metric gate 정의

각 stage 마다 evaluate. green = next stage promote, red = halt + rollback decision.

| Metric | Threshold (absolute) | Threshold (vs baseline) | Window |
|--------|----------------------|--------------------------|--------|
| **Error rate** (5xx) | ≤ 0.1% | ≤ baseline + 0.05% | rolling 5 min |
| **p99 latency** | ≤ SLO (예: 500ms login) | ≤ baseline × 1.10 | rolling 5 min |
| **p50 latency** | ≤ SLO | ≤ baseline × 1.20 | rolling 5 min |
| **Saturation** (CPU / memory / DB conn) | ≤ 70% | ≤ baseline + 10pp | rolling 5 min |
| **Business KPI** (예: signup success rate) | ≥ baseline × 0.95 | n/a | rolling 30 min |
| **Custom domain** (예: argon2 hash latency) | ≤ 200ms | ≤ baseline × 1.15 | rolling 5 min |

baseline source: previous release (last 7 day p99 percentile rolled up).
gate evaluation cadence: per-minute scheduled job (Lambda / cron / Argo Rollouts analysis), aggregate over window.

red metric → halt: NEW deploy 의 traffic 비율을 현 stage 에 freeze, on-call page, manual decision 5 min 내. Decision: rollback (default) / extend dwell + observe / promote anyway (rare, with override approval).

### Phase 3. Auto-promote vs auto-rollback 결정

| Rule | Behavior |
|------|----------|
| All metrics green throughout dwell time | **auto-promote** to next stage |
| 1+ metric red within first 50% of dwell time | **auto-rollback** to previous stage (or 0% if first stage) |
| 1+ metric red within last 50% of dwell time | **halt** + page on-call + 5min manual decision window |
| Metric gate evaluation timeout (observability outage) | **halt** + page on-call (gate uncertainty 자체가 risk) |
| Business KPI red but technical metrics green | **halt** + page on-call (business / engineering joint decision) |

manual override 권한자:
- **Promote anyway**: release manager + on-call engineer 둘 다 approve (2-person rule)
- **Rollback now (skip wait)**: on-call engineer single approve (rollback is conservative)
- **Override threshold**: VP Engineering + release manager joint approval, recorded in incident channel

### Phase 4. Hosting platform 별 implementation

| Platform | Mechanism | Tool / config |
|----------|-----------|---------------|
| **AWS Fargate (ECS) + ALB** | Weighted target group | ALB listener rule with weighted forward to two target groups (green/canary) — Terraform `aws_lb_listener_rule` with `forward.target_group.weight`. Argo Rollouts not native but can be wrapped. |
| **AWS Lambda + API Gateway** | Alias weighted routing | Lambda `additional-version-weights` on alias — `aws lambda update-alias` with `routing-config.additional-version-weights`. |
| **AWS App Runner / EKS / Fargate via App Mesh** | App Mesh virtual router weighted route | service mesh weighted routes — Terraform `aws_appmesh_route`. |
| **Kubernetes** | Argo Rollouts (recommended) or Flagger | `Rollout` CRD with `spec.strategy.canary.steps` — declarative %  + dwell time + analysis template. |
| **Vercel** | not native production-grade canary | requires preview deployments + manual traffic split (limited control). For canary-critical workload, NOT Vercel. |
| **Cloudflare Workers** | gradual deployment via `wrangler deploy --percentage` | `wrangler` CLI — manual % bumps, no native dwell + gate. |
| **Self-hosted (nginx)** | upstream weighted load balancing | nginx `upstream` block with `weight=N` — manual reload. |

본 SaaS auth example 의 cascade decision (define-tech-stack 의 Fargate 채택): **Fargate + ALB weighted target group**. T5.4 Terraform module (decompose-track-to-tasks 산출) 에서 weighted routing 추가.

### Phase 5. Verification protocol

각 stage promote 시 verify:
- SLO compliance (위 metric gate)
- business metric (signup success rate, login success rate, conversion)
- log error pattern drift (new error code 등장 여부)
- downstream dependency health (RDS connection pool, Redis hit rate, SQS depth)

post-100%-promote verify:
- 1 hour smoke test (k6 5 scenario, GA hot path)
- production dashboard 확인
- revert plan 24 hour idle (즉시 rollback 가능 상태 유지)
- 24h 후 cleanup: previous version image 삭제 (storage cost) — 단 7-day delay 권장 (불의의 recall)

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지. 본 skill 은 plan + decision 산출, 실제 deploy 자동화는 §5 build / CI/CD pipeline 에 위임.

```markdown
## setup-canary-deploy Output — <feature name> v<version>

### Summary
<3 줄: platform / stage 수 / total dwell time / auto-rollback policy>

### Stage Schedule
| Stage | Traffic % | Dwell time | Promote criteria | Rollback trigger |
|-------|-----------|------------|------------------|------------------|
| 1 | 1% | 15 min | all metrics green | any red within 7.5 min → auto-rollback, after 7.5 min → halt |
| 2 | 5% | 30 min | ... | ... |
| ... | ... | ... | ... | ... |

### Metric Gate
| Metric | Source | Absolute threshold | Baseline ratio | Window | Owner |
|--------|--------|---------------------|----------------|--------|-------|
| Error rate (5xx) | CloudWatch / OTel | ≤ 0.1% | ≤ baseline + 0.05% | rolling 5 min | platform |
| p99 latency (login) | CloudWatch | ≤ 500ms | ≤ baseline × 1.10 | rolling 5 min | backend |
| ... | ... | ... | ... | ... | ... |

### Promote / Rollback Policy
| Condition | Action | Approver | Decision SLA |
|-----------|--------|----------|--------------|
| All green throughout dwell | auto-promote | — | n/a |
| Red within first 50% dwell | auto-rollback | — | n/a |
| Red within last 50% dwell | halt + page | on-call engineer | 5 min |
| Observability outage | halt + page | on-call engineer | 5 min |
| Business KPI red, tech green | halt + page | release manager + on-call | 10 min |
| Promote anyway override | manual | release manager + on-call (2-person) | n/a |

### Platform Implementation
| Layer | Technology | Config / Tool | Owner |
|-------|------------|---------------|-------|
| Compute | AWS Fargate ECS | Terraform `aws_lb_listener_rule.canary` with weighted forward | platform / W4 |
| Routing | ALB weighted target group | green TG + canary TG, weight per stage | platform |
| Metric source | CloudWatch + OTel → X-Ray | metric stream API + alarm rules | observability / W5 (AI scaffolded) |
| Gate evaluator | Lambda scheduled (1 min) | Python lambda invoking CloudWatch GetMetricStatistics + decision log | platform |
| Promote / rollback executor | GitHub Actions deploy workflow | called from gate evaluator via webhook | platform / W4 |

### Verification per Stage
| Check | Tool | Frequency |
|-------|------|-----------|
| SLO compliance | metric gate (위) | per minute |
| Business metric | analytics dashboard | per stage |
| Log error pattern drift | CloudWatch Logs Insights query | per stage |
| Downstream health | RDS Performance Insights / Redis stats / SQS depth | per stage |
| Post-100% smoke | k6 5 scenario | once at 100% |

### Cascade
- **§7 setup-rollback-runbook**: canary auto-rollback 이 일부 처리, complex case (data corruption, multi-service) 는 rollback runbook 의 manual procedure
- **§7 prepare-launch-checklist**: canary plan 의 status 가 launch checklist 의 row
- **§7 setup-feature-flags**: canary 가 traffic-level, feature-flag 가 code-level — 두 toggle 의 cleanup 정책 정합
- **§8 monitor-regressions**: canary 의 metric history 가 regression baseline 의 출처

### Next Step
<구체 action — 1줄: 예 "T5.4 Terraform module 에 weighted target group + canary TG 추가, gate evaluator Lambda Task 신규 생성">
```

## 7. Cross-phase cascade

- **§7 setup-rollback-runbook**: auto-rollback 의 escape valve, complex rollback (schema migration, multi-service) 의 manual procedure
- **§7 prepare-launch-checklist**: canary plan 가 readiness gate row
- **§7 setup-feature-flags**: traffic-level canary + code-level flag 두 layer 의 cleanup 정합
- **§7 setup-incident-paging**: canary halt → on-call page 의 routing
- **§8 handle-incident**: canary halt 가 incident 화 시 진입점
- **§8 monitor-regressions**: canary metric history → regression baseline

## 8. 다음 skill (next in stage flow)

- `setup-rollback-runbook` — canary 자동 rollback 외 복잡 시나리오의 manual procedure
- `setup-feature-flags` — code-level toggle 인프라
- `prepare-launch-checklist` — readiness gate 통합

권장 chain (Phase 3 §7 cascade):
```
/buddy:chain setup-canary-deploy,setup-feature-flags,setup-rollback-runbook,setup-incident-paging,prepare-launch-checklist -- "<feature name> v<version>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `setup-feature-flags`** (§7) — flags = code branch toggle in same deploy, canary = traffic routing across deploys. flags 는 즉시 rollback (toggle off), canary 는 traffic shift. 두 layer 모두 사용 권장.
- **vs `setup-rollback-runbook`** (§7) — canary auto-rollback 이 1차, runbook 의 manual rollback 이 2차 (data state, schema migration, multi-service 영향).
- **vs `automate-release-tagging`** (§7) — release tagging 은 forward (semver / git tag), canary 는 traffic % rollout 의 control plane.
- **vs `handle-incident`** (§8) — incident 는 production 에서 발생한 issue 대응, canary 는 incident 전 traffic limit. canary halt 는 incident 화 가능.
- **vs `monitor-regressions`** (§8) — 그것은 long-term regression detection, 본 skill 은 short-term (per-deploy) traffic gating.

## 10. 중요 규칙

- **단계 ≥ 3 의무** — single-stage canary 거부.
- **Dwell time > 0 의무** — instant promote 거부.
- **Metric gate 정량 의무** — vague monitoring 거부.
- **Baseline + absolute 둘 다** — single criteria 거부.
- **Auto-rollback 의무** — 모든 promotion 이 manual approval 인 시스템 거부.
- **Manual override 권한자 명시** — 모호한 "필요 시 override" 거부.
- **Platform 별 implementation hint 의무** — generic "canary 적용" 거부, hosting 별 mechanism 명시.
- **Read-only on production code** — 본 skill 은 plan 산출, 실제 deploy 자동화는 §5 build / pipeline 에 위임.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행 (stage 비율 → metric gate → policy → platform → verification)
- [ ] §6 의 6 출력 섹션 (Summary / Stage Schedule / Metric Gate / Promote-Rollback Policy / Platform Implementation / Verification / Cascade) 모두 채워짐
- [ ] Stage Schedule ≥ 3 stage, dwell time > 0 모든 stage
- [ ] Metric Gate ≥ 4 metric (error rate / latency / saturation / business KPI 최소) — 모두 absolute + baseline 둘 다 명시
- [ ] Promote / Rollback Policy 의 5 condition 모두 정의 + auto-rollback default
- [ ] Platform Implementation 이 §3 tech stack 의 hosting platform 과 일치 + tool / config 구체 명시
- [ ] Verification per Stage ≥ 4 check + tool + frequency
- [ ] §4 posture — stage / threshold 단호, "적당히" / "필요 시" hedge 없음
- [ ] §0 anti-pattern 부재 — single-stage / dwell 0 / threshold 미정 / baseline 없음 / auto-rollback 없음 / platform 무시 / override 모호 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
