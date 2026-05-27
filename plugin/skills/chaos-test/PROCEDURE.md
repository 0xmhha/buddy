# Chaos Test — failure injection (network / pod / latency / dependency)

## 1. 목적

Production-like 환경에서 *의도적으로 실패 유발* — *blast radius / recovery time / data integrity* 검증. **chaos engineering** (Casey Rosenthal) 의 4 원칙 적용.

`run-load-test` (§6, 구현됨) 와 책임 분리 — load test 는 *부하* 측정, chaos test 는 *실패 회복* 측정.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 대상 시스템 + 실패 가설 | ✅ | knowledge | 사용자 도메인 지식 | "어떤 실패 시나리오를 테스트하나요? 가설을 알려주세요." |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Chaos test report (가설 검증 결과 + blast radius) | artifact | structured report | `conduct-postmortem` |

## 2. 사용 시점

- §6 verify-quality 의 *high-availability 영역* skill
- 분기 / 반기 chaos game day (운영 팀 합동)
- 신규 의존 (3rd-party API / DB / cache) 추가 후 — 의존 실패 시 영향 검증
- production incident 후 — 같은 실패 *재현* + recovery 검증
- SLO threshold (`design-observability` 산출) 달성 가능 검증

## 3. 입력

### 필수
- `derive-system-topology` 산출 — failure injection 가능 영역 식별
- `design-observability` 의 SLO / error budget — chaos 결과 정량 평가 baseline
- production-like 환경 (staging 또는 sandboxed prod)

### 선택
- 이전 chaos test log (drift 추적)
- 사용자 SLA contractual obligation (외부 약속 baseline)

## 4. Stage 흐름

### Stage 1: 4 원칙 (Chaos Engineering manifesto)

| 원칙 | 적용 |
|------|------|
| 1. Steady state hypothesis 정의 | "정상 상태" 측정 가능 metric (e.g. p99 latency < 200ms, error rate < 0.1%) |
| 2. 실세계 event 가정 | server crash / network partition / DB slow / dependency timeout |
| 3. Production *like* 환경 | staging 우선, 점진적 production (controlled blast radius) |
| 4. Steady-state 깨지면 즉시 중단 | automated abort — manual intervention 보호 |

### Stage 2: Failure injection 종류

| 종류 | 도구 | 영향 |
|------|------|------|
| **Network** | Chaos Mesh / Pumba / iptables | latency / packet loss / partition |
| **Pod / process** | Chaos Mesh / kube-monkey | kill / restart / pause |
| **CPU / memory** | stress-ng / Chaos Mesh | resource exhaustion |
| **Dependency** | toxiproxy / Chaos Mesh HTTPChaos | 3rd-party slow / fail / timeout |
| **DB** | Chaos Mesh / pg_proxy | slow query / connection pool exhaustion |
| **Time** | libfaketime | clock skew / leap second |

### Stage 3: Blast radius 통제

| 단계 | radius | 환경 |
|------|------|------|
| 1 | 단일 pod | dev |
| 2 | 전체 service (1 region) | staging |
| 3 | 1 region 전체 | staging |
| 4 | production (1% traffic) | production canary |
| 5 | production (점진 확대) | production controlled |

→ 단계 4-5 는 *automated abort* + *runbook* 필수.

### Stage 4: Hypothesis-driven 실험

각 chaos test 는 *가설* 명시:

```yaml
hypothesis: "DB primary 가 30s 응답 지연 시, application 은 read replica 로 자동 fallback 하고 p99 latency < 500ms 유지"

steady_state:
  - metric: p99 latency
    threshold: < 200ms
  - metric: error rate
    threshold: < 0.1%

experiment:
  type: "DB primary 30s slow"
  duration: 5min
  blast_radius: "single replica"

abort_conditions:
  - p99 latency > 1000ms (5x baseline)
  - error rate > 5%
  - on-call alert triggered

expected_recovery:
  - within 60s of experiment end
  - no data loss
```

### Stage 5: Result analysis

| 항목 | 측정 |
|------|------|
| Hypothesis 통과? | yes / no |
| Recovery time | actual vs expected |
| Error budget burn | % of monthly budget |
| 발견된 issue | list with severity |
| Runbook gap | 부재한 절차 |

→ 발견된 issue 는 *즉시 fix* 또는 *runbook 추가*. *학습 = chaos test 의 main 산출*.

### Stage 6: Game day (조직 차원)

분기 / 반기 *예정된 chaos session* :
- 운영 팀 + dev 팀 합동
- *모르는* 실패 (random injection) 가정
- on-call 대응 시뮬레이션
- 사후 *blameless postmortem* — runbook 갱신

## 5. 산출물 형식

```markdown
## Chaos Test — {experiment 이름}

### Hypothesis
- steady state: ...
- experiment: ...
- abort conditions: ...

### Result
- hypothesis 통과: yes / no
- recovery time: actual / expected
- error budget burn: X%

### 발견 issue
| severity | issue | fix plan |

### Runbook gap
- ...

### Game day log (조직 차원)
- 참가 팀 / 진행 / 학습
```

## 6. 검증

- [ ] 4 원칙 모두 적용 (steady state / 실세계 event / production-like / abort)?
- [ ] Failure injection 종류 ≥3 영역 검증?
- [ ] Blast radius 단계 (1~5) 명시?
- [ ] Hypothesis-driven (가설 + abort condition + expected recovery) ?
- [ ] Result analysis 정량 (recovery time / budget burn) ?
- [ ] Game day 분기/반기 진행?

## 7. 다음 phase

- `audit-error-budget` (§8) 의 burn rate 입력
- `design-observability` 의 SLO 정합 — 깨진 가설 시 SLO 재정의
- `conduct-postmortem` (구현됨) — chaos 학습 영속화

## 8. 참조

- Chaos Engineering (Casey Rosenthal + Nora Jones) — 4 원칙
- principlesofchaos.org — chaos manifesto
- Chaos Mesh (CNCF) / Gremlin — 도구 reference
