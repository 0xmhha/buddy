# run-load-test — sustained load + soak + spike + stress 시나리오 실행

§6 verify-quality phase 의 stage. **prepare-launch-checklist Performance row 의 evidence source**. k6 nightly smoke 만으로는 부족 — production-grade launch 직전 sustained load (1h+) + soak (8h+) + spike + stress 4 시나리오 실행 + breaking point 식별 + capacity headroom 측정. 산출물은 시나리오별 metric report + breaking point + capacity plan + acceptance gate decision.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Single 시나리오만 (peak load only)** — sustained / soak / spike 중 1-2 개 누락 → memory leak / connection pool exhaustion / CDN cache miss spike 못 잡음.
- **Production parity 없는 staging 측정** — staging 의 instance size / DB connection / cache 가 production 보다 작으면 결과 무근거.
- **Breaking point 미측정** — SLO 통과만 보고 "성공" → 1.5x / 2x / 5x peak 에서 어디서 깨지는지 모름. capacity plan 무근거.
- **Headroom 정량 없음** — "여유 있음" 만 → numeric headroom % 명시.
- **Monitoring 부재** — k6 client metric 만 보고 server metric (CPU / memory / DB conn / cache hit rate) 안 보면 root cause 불명.
- **Cost 추정 누락** — load test 자체의 AWS 비용 + production scale 비용 추정 안 하면 budget 충돌.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

production launch 직전 capacity + breaking point + headroom 정량 확보. prepare-launch-checklist 의 Performance / Operations row 의 evidence. SLA commitment (예: 99.9% / p99 500ms) 의 근거.

## 2. 사용 시점 (When to invoke)

- production launch 직전 (필수, prepare-launch-checklist Performance row)
- 신규 critical endpoint 추가 후
- 1.5x+ traffic 증가 예상 시 (marketing campaign, news cycle)
- infra 변경 (instance type / DB version / cache layer) 후 regression
- SLA renegotiation / customer commitment 변경 시

## 3. 입력 (Inputs)

### 필수
- §3 SLO definition (latency / availability / throughput thresholds)
- §3 tech stack (compute / DB / cache / queue) — production-parity staging 보유
- §4 acceptance test plan 의 k6 nightly scenario (baseline)
- production 추정 traffic (baseline RPS + peak ratio + growth rate)
- §6 test-cross-actor-flow 의 hot endpoint list

### 선택
- 이전 release 의 load test 결과 (regression 비교)
- compliance 의무 (예: 금융 — 1k TPS 의무)
- multi-region 결정 (region-별 load 분배)
- chaos test 결합 (Cluster A 의 chaos-test deferred)

### 입력이 부족할 때 forcing question
- "production-parity staging 보유? 아니면 staging upgrade 가 prerequisite — load test 무근거."
- "SLO 가 정량 명시됐나? 'fast' 거부, 'p99 login 500ms / error 0.1% / availability 99.9%'."
- "production peak RPS 추정 근거? historical traffic + growth + safety factor (2x) — 미확보 시 마지막 부분 결정."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — pass/fail/breaking-point 단호. "거의 충분" 거부.
- **사용자 입력을 challenge** — "load 통과" 발화에 "어느 시나리오? sustained / soak / spike 모두?" push.
- **Specificity 강제** — vague "성능 OK" 거부, "1h sustained 100 RPS, p99 login 488ms vs 500ms SLO, error 0.05%, RDS conn 65% saturation, breaking point 380 RPS at conn pool exhaustion".
- **Production-parity bias** — staging size = production size 강제, downsize 시 결과 무근거 명시.

도메인 원칙:

1. **4 시나리오 의무** — sustained / soak / spike / stress.
2. **Server + client metric 동시 측정** — k6 client metric + CloudWatch server metric (CPU / memory / DB conn / cache hit / GC pause).
3. **Breaking point 측정** — 시작점부터 monotonic 증가, 첫 SLO 위반 지점 = breaking point.
4. **Headroom 정량** — `(breaking point - peak) / peak × 100%` — minimum 50% 권장.
5. **Cost transparent** — load test AWS cost + extrapolated production cost 모두 보고.

## 5. 단계 (Phases)

### Phase 1. 시나리오 정의 (4)

| 시나리오 | 목적 | 부하 패턴 | 시간 |
|----------|------|-----------|------|
| **sustained** | normal peak 지속 가능 검증 | 100 RPS 고정 (1.5x baseline 추정) | 1 hour |
| **soak** | memory leak / resource exhaustion 검출 | 50 RPS 고정 (baseline) | 8 hour (overnight) |
| **spike** | sudden burst 회복 | baseline 50 RPS → 30s 동안 500 RPS spike → baseline 복귀 → 5 min 안정 | 30 min total |
| **stress** | breaking point 식별 | 0 RPS → linear ramp 1000 RPS over 1 hour | 1 hour |

각 시나리오의 hot endpoint coverage: signup / login / refresh / verify-email / `/me` (5 endpoint, k6 baseline scenario 재사용).

### Phase 2. Production-parity staging 검증

| Component | Production | Staging | Parity |
|-----------|------------|---------|--------|
| Fargate auth-api | 4 vCPU × 8GB × 4 task | 4 vCPU × 8GB × 4 task | ✓ |
| RDS Postgres 16 | db.r6g.large (multi-AZ) | db.r6g.large (single-AZ) | △ (multi-AZ 차이만) |
| ElastiCache Redis 7 | cache.r7g.large × 2 | cache.r7g.medium × 1 | ✗ (downsize) |
| ALB | production tier | production tier | ✓ |
| SQS | unlimited (managed) | unlimited (managed) | ✓ |

**Redis downsize** flag — staging 결과의 cache layer breaking point 가 prod 보다 빨리 발생 가능. extrapolate 시 보정 (Redis ops/sec 비율로 scale).

### Phase 3. Tool stack + 실행

| Tool | 용도 |
|------|------|
| **k6** | load generator (cloud-distributed for stress 1000 RPS) — k6 cloud or self-host 4 worker EC2 |
| **CloudWatch** | server-side metric (CPU / memory / DB conn / cache hit / 5xx rate) |
| **OTel + X-Ray** | trace level latency breakdown (slow operation 식별) |
| **RDS Performance Insights** | DB query 별 latency / connection wait |
| **k6 cloud dashboard** | client-side aggregate metric + percentile |

실행 절차:
```bash
# sustained
k6 run --vus 100 --duration 1h scenarios/sustained.js
# soak (overnight, separate run)
k6 run --vus 50 --duration 8h scenarios/soak.js
# spike
k6 run --stage 5m:50,30s:500,5m:50 scenarios/spike.js
# stress (breaking point search)
k6 run --stage 60m:1000 scenarios/stress.js
```

### Phase 4. Metric collection + breaking point 식별

| 시나리오 | Client metric (k6) | Server metric (CloudWatch) | Pass criteria |
|----------|---------------------|------------------------------|---------------|
| sustained | p99 login 488ms / error 0.05% | CPU 65% / RDS conn 65% / Redis hit 98% / GC pause < 50ms | p99 ≤ SLO + saturation < 70% + no error spike |
| soak | p99 stable across 8h / error stable | memory steady (no leak) / GC pause stable | no monotonic memory growth, no GC pause growth, no connection leak |
| spike | recovery within 30s post-spike | brief CPU spike to 90%, recovery | recovery time < 30s, no cascading failure |
| stress | breaking point 측정 | 첫 SLO 위반 = breaking point | RPS 380+ 권장 (3.8x peak baseline) |

breaking point identification:
- p99 latency > SLO × 1.5 (sustained 1 min) = breaking point candidate
- error rate > 1% (sustained 30s) = hard breaking point
- saturation > 90% (CPU / DB conn) = soft breaking point (degradation imminent)

본 SaaS auth example breaking point: **stress test RPS 420 at RDS connection pool exhaustion** (prod connection pool 100 + per-task 20 connections × 4 tasks = 80, saturation hits at ~420 RPS due to query mix p95 50ms × 80 = 1600 query-ms / sec / 80 conn = 20 RPS/conn × 80 = 1600 RPS theoretical, but query queue + commit lag → 420 actual).

### Phase 5. Headroom + capacity plan + cost

| Metric | Value |
|--------|-------|
| Production peak RPS (estimated) | 100 (12-mo target) |
| Breaking point RPS | 420 (stress test) |
| **Headroom** | (420 - 100) / 100 = **320%** ✓ (target ≥ 50%) |
| Capacity at 1k RPS | requires 2x Fargate tasks + RDS read replica + connection pool tuning to 200 |
| Capacity at 10k RPS | requires write sharding (per design-data-model risk #1) |

**Cost (load test AWS bill estimate):**
- sustained 1h × 4 worker EC2 (k6 self-host): ~$1
- soak 8h × 4 worker EC2: ~$8
- spike + stress: ~$3
- staging Fargate + RDS uplift during test: ~$10
- **Total: ~$22 per full load test run**

**Cost (production scaled to 1k RPS):**
- Fargate 8 task × 4vCPU × 8GB ≈ $400/mo (vs 100 RPS $200/mo)
- RDS r6g.xlarge multi-AZ ≈ $700/mo (vs $300/mo)
- Redis cluster mode 4 shard ≈ $400/mo
- **Total ~$1,500/mo at 1k RPS** (vs prod budget $2,500/mo — within budget with 40% headroom)

acceptance gate:
- **pass**: 4 시나리오 통과 + breaking point ≥ 3x peak + headroom ≥ 50% + cost projection within budget
- **conditional pass**: 1 시나리오 marginal (예: spike recovery 35s vs 30s target) + mitigation plan + ETA
- **fail**: breaking point < 2x peak OR memory leak detected OR cost projection > budget

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## run-load-test Output — <project name> v<version>

### Summary
<3 줄: 4 시나리오 결과 / breaking point / headroom % / cost / decision>

### Scenario Results
| 시나리오 | k6 client metric | Server metric | Status |
|----------|---------------------|----------------|--------|
| sustained 1h × 100 RPS | p99 488ms / err 0.05% | CPU 65% / RDS conn 65% / Redis 98% | pass |
| soak 8h × 50 RPS | p99 stable / no error growth | memory steady / no leak | pass |
| spike (50→500→50 RPS) | recovery 28s | brief CPU 88% / no cascading | pass |
| stress (linear 0→1000 RPS) | breaking at 420 RPS | RDS conn pool exhaustion | breaking point identified |

### Production-Parity Staging Check
| Component | Parity status | Adjustment |
|-----------|---------------|------------|
| ... | ✓ / △ / ✗ | ... |

### Breaking Point + Headroom
| Field | Value |
|-------|-------|
| Production peak (12-mo) | 100 RPS |
| Breaking point | 420 RPS (RDS conn pool) |
| **Headroom** | 320% (target ≥ 50%) |
| Headroom 1k RPS scale-up | requires 2x Fargate + RDS read replica + conn pool 200 |
| Headroom 10k RPS scale-up | requires write sharding |

### Cost Report
| Item | Cost |
|------|------|
| Load test AWS bill (4 시나리오) | ~$22 |
| Production at 1k RPS (extrapolated) | ~$1,500/mo (within $2,500 budget) |
| Production at 10k RPS (extrapolated) | requires sharding — separate budget review |

### Acceptance Gate
| Field | Value |
|-------|-------|
| Decision | pass / conditional / fail |
| Rationale | 4 시나리오 통과 + breaking point 3.2x + headroom 320% + budget 60% utilized |
| Mitigation (if conditional) | n/a |

### Cascade
- **§7 prepare-launch-checklist**: Performance row green 입력
- **§7 setup-canary-deploy**: breaking point + headroom = canary metric gate baseline
- **§8 monitor-regressions**: scenario metric = production regression baseline
- **§3 design-data-model** (필요 시): write sharding plan trigger (10k RPS scale-up)

### Next Step
<구체 action — 1줄: 예 "audit-accessibility 진입 + canary deploy plan finalize">
```

## 7. Cross-phase cascade

- **§7 prepare-launch-checklist**: Performance row green
- **§7 setup-canary-deploy**: breaking point baseline → canary metric gate
- **§8 monitor-regressions**: scenario metric → production regression baseline
- **§3 design-data-model**: scale-up trigger (sharding plan)

## 8. 다음 skill (next in stage flow)

- `audit-accessibility` (§6 Cluster A) — 다음 launch readiness 보강
- `audit-cost-efficiency` (§6 Cluster A) — economics 정량
- `prepare-launch-checklist` (§7) — readiness gate 통합

권장 chain (Cluster A 일괄):
```
/buddy:chain run-load-test,audit-accessibility,audit-cost-efficiency,prepare-launch-checklist -- "<project> v<version>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `define-acceptance-test-plan`** (§4) — 그것은 plan + k6 nightly scenario 정의, 본 skill 은 production-grade launch 직전 4 시나리오 sustained/soak/spike/stress 실행. 본 skill 이 후속.
- **vs `monitor-regressions`** (§8) — 그것은 production long-term, 본 skill 은 pre-launch one-time. 본 skill 결과가 baseline.
- **vs `setup-canary-deploy`** (§7) — 그것은 deploy-time traffic gate, 본 skill 은 capacity 사전 측정. 본 skill 의 breaking point 가 canary metric gate 의 input.
- **vs `chaos-test`** (deferred Cluster A) — 본 skill 은 load 만, chaos-test 는 fault injection. 함께 쓰면 production-grade resiliency 검증.

## 10. 중요 규칙

- **4 시나리오 의무** — single scenario 거부.
- **Production-parity 의무** — downsize staging 시 명시 + extrapolate 보정.
- **Breaking point 측정 의무** — SLO 통과만으로 끝내지 않음.
- **Headroom 정량** — minimum 50% target.
- **Cost transparent** — test cost + production scaled cost.
- **Read-only on production** — staging 만 측정.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 6 출력 섹션 (Summary / Scenario Results / Parity / Breaking Point / Cost / Acceptance Gate / Cascade) 모두 채워짐
- [ ] 4 시나리오 (sustained / soak / spike / stress) 모두 실행
- [ ] Production-Parity Staging Check 의 모든 component 명시 (downsize 시 보정 명시)
- [ ] Breaking Point 식별 + RPS 정량 + root cause 명시
- [ ] Headroom % 정량 (target ≥ 50%)
- [ ] Cost report 의 test cost + production scaled cost
- [ ] Acceptance Gate decision 단일 (pass/conditional/fail) + rationale
- [ ] §4 posture — 정량 단호, hedge 없음
- [ ] §0 anti-pattern 부재 — single scenario / parity 무시 / breaking point 부재 / headroom 없음 / monitoring 부재 / cost 누락 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
