# audit-cost-efficiency — Infracost + $/MAU + unit economics 정량

§6 verify-quality phase 의 stage. **prepare-launch-checklist Cost & Business axis 의 evidence**. Infracost 의 monthly est. + per-component 분해 + $/MAU unit economics + idle waste 식별 + savings opportunity 추천. budget vs actual 비교 + scaling cost projection. 산출물은 cost matrix + waste report + savings recommendation + acceptance gate.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Total monthly cost 만 보기** — $X/mo 만 보고 component breakdown 안 보면 어디 비용이 집중되는지 모름.
- **Production scale projection 부재** — 현재 dev/staging cost 만 보고 launch 시 cost 추정 안 함 → budget 충돌 risk.
- **Unit economics 부재** — $/MAU / $/request / $/transaction 분해 안 하면 pricing 결정 무근거.
- **Idle waste 무검증** — over-provisioned compute / unused EBS / orphan snapshot — 보통 5-15% 비용 차지.
- **Reserved Instance 미검토** — on-demand 만 사용 → 1y RI 시 30-40% 절감 가능.
- **Cross-AZ traffic cost 누락** — VPC 안 traffic 도 cross-AZ 시 $0.01/GB — 100k user 시 $X/mo.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

infra cost 의 정량 + 분해 + scaling projection + waste 식별. business model (pricing tier) 의 unit economics 근거. budget gate enforcement.

## 2. 사용 시점 (When to invoke)

- launch 직전 (필수, prepare-launch-checklist Cost & Business axis)
- pricing tier 결정 / 변경 시
- monthly cost spike 발견 후 (post-incident root cause)
- scaling milestone (10k MAU, 100k MAU) 진입 시 cost projection
- multi-region expansion 결정 시
- compliance 의무 (의료 / 금융 — cost transparency 의무)

## 3. 입력 (Inputs)

### 필수
- §3 tech stack ADR (선택된 component)
- AWS account (또는 GCP / Azure) actual usage data
- Terraform state (Infracost input)
- production user/traffic projection (12-mo / 3-yr)
- pricing model (Free / Pro / Enterprise tier 정의)

### 선택
- 이전 release 의 cost report (regression 비교)
- competitor benchmark ($/MAU industry standard)
- compliance scope (의료 BAA / 금융 audit cost)
- multi-region 결정

### 입력이 부족할 때 forcing question
- "Terraform state 가 commit 됐나? Infracost 입력 prerequisite. 미commit 시 cost projection 무근거."
- "production user/traffic projection 정량? 'eventual scale' 거부, '12-mo target 100k MAU + 1k RPS peak' 명시."
- "pricing tier 의 $/seat 또는 $/MAU 결정됐나? 미결정 시 unit economics 계산 부분 결정."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — pass / fail / waste % 단호. "비용 적당" 거부.
- **사용자 입력을 challenge** — "$X/mo OK" 발화에 "어느 component? scaling? unit economics?" push.
- **Specificity 강제** — vague "cost reasonable" 거부, "Fargate $400 (43%) + RDS $300 (32%) + S3 $80 (9%) + others $150 (16%) = $930/mo, $/MAU $0.0093 vs revenue $0.05/MAU = 81% margin".
- **Waste 식별 의무** — 보통 5-15% 절감 가능.

도메인 원칙:

1. **Component breakdown 의무** — top 5 component 가 90%+ cover.
2. **Scaling projection 3 단계** — current / 10x / 100x scale.
3. **Unit economics** — $/MAU + $/request + $/transaction.
4. **Waste category 5** — over-provisioning / idle / orphan / cross-AZ traffic / data egress.
5. **Savings recommendation 정량** — RI / Savings Plan / Spot / right-sizing.

## 5. 단계 (Phases)

### Phase 1. Infracost monthly estimate

```bash
infracost breakdown --path=infra/terraform/envs/prod
```

| Component | Resource | $/mo |
|-----------|----------|------|
| Fargate ECS | 4 task × 4vCPU × 8GB × 730h | $400 |
| RDS Postgres 16 (db.r6g.large multi-AZ + 100GB storage) | RDS instance + storage | $310 |
| ElastiCache Redis 7 (cache.r7g.large × 2) | Redis cluster | $200 |
| ALB | 1 ALB + 100k LCU/mo | $50 |
| CloudFront + S3 | 1TB egress + 100k req/mo | $90 |
| SQS | 10M req/mo | $4 |
| SES | 100k email/mo | $10 |
| Secrets Manager | 5 secret × 1k API call/mo | $3 |
| Route 53 | 1 hosted zone | $1 |
| CloudWatch | logs 50GB + metrics + alarm | $80 |
| Data transfer (cross-AZ + egress) | 10TB intra + 1TB egress | $130 |
| **Total** | | **~$1,278/mo** |

(within $2,500 prod budget — 49% utilized)

### Phase 2. Component breakdown + pareto

| Component | $/mo | % | Cumulative % |
|-----------|------|---|---------------|
| Fargate | $400 | 31.3% | 31.3% |
| RDS | $310 | 24.3% | 55.6% |
| Redis | $200 | 15.7% | 71.3% |
| Data transfer | $130 | 10.2% | 81.5% |
| CloudFront + S3 | $90 | 7.0% | 88.5% |
| CloudWatch | $80 | 6.3% | 94.8% |
| ALB | $50 | 3.9% | 98.7% |
| Others (SQS / SES / Secrets / Route53) | $18 | 1.4% | 100% |

top 4 component (Fargate / RDS / Redis / Data transfer) = **81.5%** — 우선 최적화 대상.

### Phase 3. Unit economics

| Metric | Value | Calculation |
|--------|-------|-------------|
| **$/MAU @ 10k MAU** (12-mo target) | $0.13 | $1,278 / 10,000 |
| **$/MAU @ 100k MAU** (3-yr target, with scale-up) | scale projection 별도 (Phase 4) | |
| **$/request** | $0.000064 | $1,278 / 20M req/mo (50 RPS avg × 30d × 24h × 3600s × 1.55 amplification factor) |
| **$/transaction** (signup or login) | $0.001 | (signup CPU + DB write + email cost) |

revenue model 가정 (Pro tier $5/mo/user × 30% conversion = $1.5/MAU avg):
- $1,278 cost / 10k MAU = $0.13/MAU
- $1.5 revenue/MAU
- **margin: 91% gross, healthy** ✓

### Phase 4. Scaling projection

| Scale | MAU | RPS peak | Compute | DB | Cache | Total $/mo | $/MAU |
|-------|-----|----------|---------|-----|-------|-------------|-------|
| current | 1k | 10 | $200 (2 task) | $300 | $100 | $700 | $0.70 |
| 12-mo | 10k | 100 | $400 (4 task) | $310 (same) | $200 | $1,278 | $0.13 |
| 3-yr | 100k | 1k | $1,000 (10 task) | $1,200 (multi-AZ + read replica) | $400 (cluster mode 4 shard) | $3,500 | $0.035 |
| 5-yr | 1M | 10k | requires sharding (per data-model risk #1) | $5,000+ (sharded) | $1,500 | ~$10,000 | $0.010 |

note: per design-data-model risk #1, 1k → 10k RPS 구간에서 write sharding 필수 — 본 projection 의 5-yr 가정은 sharding 도입 후.

### Phase 5. Waste + savings recommendation

#### Waste detection

| Waste type | Detected | $/mo waste | % |
|------------|----------|------------|---|
| Over-provisioned Fargate (CPU avg 30%) | yes — current 4 task → 2 task sufficient at off-peak | $200 | 15.7% |
| Idle EBS (orphan from old test instance) | yes — 50GB orphan snapshot 5 EA | $5 | 0.4% |
| Cross-AZ data transfer (RDS multi-AZ replica + Fargate scattered AZ) | yes — minor optimization possible | $30 | 2.3% |
| Under-utilized Redis (cache hit 92%, could be smaller) | flag — investigate cache.r7g.medium | $80 (potential) | 6.3% |
| CloudWatch log retention (default never expires) | yes — 50GB → 90d retention saves | $40 | 3.1% |
| **Total waste** | | **~$355/mo (28%)** | |

#### Savings recommendation

| Recommendation | Savings | Risk | Action |
|----------------|---------|------|--------|
| 1y Reserved Instance for RDS (db.r6g.large) | $120/mo (-39%) | low (1y commitment) | recommended — apply now |
| 1y Compute Savings Plan for Fargate | $100/mo (-25%) | low (1y commit, flexible) | recommended |
| Right-size Fargate (4 → 2 task off-peak via auto-scaling) | $200/mo (-15.7%) | medium (need testing) | recommended after run-load-test breaking point |
| Right-size Redis (r7g.large → r7g.medium) | $80/mo (-6.3%) | medium (cache hit verify) | post-launch monitoring → resize if hit > 95% sustained |
| CloudWatch log retention 90d | $40/mo (-3.1%) | n/a (compliance check) | apply now if no audit retention |
| EBS orphan snapshot cleanup | $5/mo | n/a | apply now |
| **Total potential savings** | **$545/mo (-43%)** | | |

post-savings cost: $1,278 - $545 ≈ **$733/mo** (29% prod budget utilized — comfortable headroom).

acceptance gate:
- **pass**: total cost within budget (50% utilization ≤ healthy ≤ 80%) + waste < 30% + RI/SP recommended + scaling projection within 3-yr budget
- **conditional pass**: 1-2 high-cost component flagged + optimization plan + ETA
- **fail**: cost > 100% budget OR scaling projection diverges (3-yr > 3x current budget)

## 6. 산출물 형식 (Output format)

```markdown
## audit-cost-efficiency Output — <project name> v<version>

### Summary
<3 줄: total $/mo / budget utilization % / $/MAU / waste % / savings opportunity / decision>

### Infracost Monthly Breakdown
| Component | Resource | $/mo | % | Cumulative % |
|-----------|----------|------|---|---------------|
| ... | ... | ... | ... | ... |

### Unit Economics
| Metric | Value | Calculation |
|--------|-------|-------------|
| $/MAU @ 10k | $0.13 | $1,278 / 10k |
| $/request | $0.000064 | total / 20M req |
| revenue $/MAU | $1.5 | Pro tier × 30% conversion |
| **gross margin** | **91%** | (1.5 - 0.13) / 1.5 |

### Scaling Projection
| Scale | MAU | RPS peak | Total $/mo | $/MAU | Note |
|-------|-----|----------|-------------|-------|------|
| current | ... | ... | ... | ... | ... |
| 12-mo | ... | ... | ... | ... | ... |
| 3-yr | ... | ... | ... | ... | requires sharding |

### Waste Detection
| Waste type | Detected | $/mo | Action |
|------------|----------|------|--------|
| ... | yes/no | ... | ... |

### Savings Recommendation
| Recommendation | Savings $/mo | Risk | Recommended? |
|----------------|---------------|------|--------------|
| RDS RI 1y | $120 | low | ✓ apply |
| Compute Savings Plan | $100 | low | ✓ apply |
| ... | ... | ... | ... |

### Budget Status
| Field | Value |
|-------|-------|
| Current cost | $1,278/mo |
| Budget | $2,500/mo |
| Utilization | 51% |
| Post-savings cost | $733/mo |
| Post-savings utilization | 29% |
| Decision | pass / conditional / fail |

### Cascade
- **§7 prepare-launch-checklist**: Cost & Business axis row 입력
- **§8 analyze-cost-anomaly** (deferred Cluster F): 본 baseline → production cost spike detection
- **§3 design-data-model** (5-yr projection 시): write sharding plan trigger

### Next Step
<구체 action — 1줄: 예 "RDS 1y RI + Compute SP 즉시 apply, prepare-launch-checklist 진입">
```

## 7. Cross-phase cascade

- **§7 prepare-launch-checklist**: Cost & Business axis evidence
- **§8 analyze-cost-anomaly**: baseline = production cost spike detection
- **§3 design-data-model**: 5-yr scaling projection → sharding trigger

## 8. 다음 skill (next in stage flow)

- `prepare-launch-checklist` (§7) — Cluster A 3 skill 통합 후 launch readiness gate

권장 chain (Cluster A 일괄):
```
/buddy:chain run-load-test,audit-accessibility,audit-cost-efficiency,prepare-launch-checklist -- "<project> v<version>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `define-tech-stack`** (§3) — 그것은 stack 결정 시 cost dimension 평가, 본 skill 은 implementation 후 actual measurement. 본 skill 이 후속.
- **vs `analyze-cost-anomaly`** (§8 — deferred) — 그것은 production cost spike detection (continuous), 본 skill 은 launch 전 baseline (one-time).
- **vs `audit-error-budget`** (§8 — deferred) — error budget = SLO breach tracking, cost-efficiency = financial. 두 axis.

## 10. 중요 규칙

- **Component breakdown 의무** — total only 거부.
- **Scaling 3 단계** — current / 12-mo / 3-yr.
- **Unit economics 의무** — $/MAU + $/request 분리.
- **Waste 5 category** — single category audit 거부.
- **Savings 정량** — RI / SP / right-sizing 명시.
- **Budget gate 정량** — utilization % 명시.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 7 출력 섹션 모두 채워짐
- [ ] Infracost breakdown 의 top component 90%+ cover
- [ ] Unit economics ($/MAU + $/request + margin) 명시
- [ ] Scaling Projection 3+ 단계 (current / 12-mo / 3-yr+)
- [ ] Waste Detection 5 category 모두 cover (over-prov / idle / orphan / cross-AZ / log retention)
- [ ] Savings Recommendation 의 RI / SP / right-sizing 정량
- [ ] Budget Status 의 utilization % + decision
- [ ] §4 posture — 정량 단호
- [ ] §0 anti-pattern 부재 — total only / projection 부재 / unit economics 부재 / waste 무검증 / RI 미검토 / cross-AZ 누락 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
