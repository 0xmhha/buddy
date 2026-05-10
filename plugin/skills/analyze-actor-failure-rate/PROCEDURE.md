# Analyze Actor Failure Rate — actor 별 실패 패턴 + 회복 전략

## 1. 목적

system 의 각 *actor* (user / system / 3rd-party / external-tool) 별 **실패율 측정 + 패턴 분석 + recovery 전략 평가**. agent-evaluation 의 trust-centric 패턴 차용 — *actor 별 trust score* 산출.

`audit-error-budget` (§8) 와 cascade — actor 별 budget burn 분포.

## 2. 사용 시점

- §8 iterate-product 의 *production reliability 분석*
- incident 사후 — 어느 actor 가 *원인 / 영향 받음*
- 외부 의존 추가 후 — 새 actor 의 failure 영향
- 분기 SLO review — actor 별 budget burn
- chaos-test 결과 분석 — failure injection 후 actor 별 recovery

## 3. 입력

### 필수
- `derive-system-topology` 산출 — actor 그래프
- production logs / metrics — actor 별 attribution
- error budget / SLO (`audit-error-budget` 산출)

### 선택
- 사용자 incident report
- 3rd-party SaaS status page (외부 dependency status)

## 4. Stage 흐름

### Stage 1: Actor 별 failure 분류

| actor | failure 종류 |
|-------|----------|
| User (input) | invalid input / abandoned flow / browser error |
| System (internal service) | crash / timeout / OOM / DB connection lost |
| 3rd-party SaaS | API down / rate limit / breaking change |
| External tool (CI / monitoring) | build fail / alert miss |

### Stage 2: 실패율 측정

각 actor × failure 종류 매트릭스:

```
              User    System   3rd-party  External
Jan          0.5%    0.1%     2.0%       0.3%
Feb          0.6%    0.2%     5.0%        ← spike
Mar          0.4%    0.1%     1.8%       0.3%
```

→ *spike* 발견 시 root cause 분석 trigger.

### Stage 3: Trust score (agent-evaluation 패턴)

각 actor 의 *trust score* (input 신뢰성):

| 차원 | 측정 |
|------|------|
| Reliability | 1 - failure rate |
| Predictability | failure 분포의 variance |
| Recovery time | mean time to recovery (MTTR) |
| Blast radius | 실패 시 영향 받는 다른 actor 수 |

→ trust score 종합. 낮은 actor → *우선 fortification* 대상.

### Stage 4: Recovery 전략 평가

각 actor failure 의 recovery 패턴:

| 패턴 | 적용 actor | 효과 |
|------|---------|------|
| Retry (exponential backoff) | 3rd-party / system transient | 일시적 실패 |
| Circuit breaker | 3rd-party 지속 실패 | cascading failure 차단 |
| Fallback (degraded) | 3rd-party / cache | 가용성 우선 |
| Bulkhead (resource isolation) | system | blast radius 통제 |
| Timeout + graceful error | user | UX 보호 |
| Idempotent retry | system / 3rd-party | data integrity |

→ 각 actor 의 recovery 적용 여부 + gap 식별.

### Stage 5: Cascade analysis

actor 간 failure 전파:
- 3rd-party fail → system retry → user timeout → user abandon
- system DB fail → multiple service degraded → user-facing error

→ cascade 차단 patterns (circuit breaker / bulkhead) 의 효과 측정.

## 5. 산출물 형식

```markdown
## Actor Failure Analysis — {분기}

### 실패율 매트릭스
| actor | period | rate | trend |

### Trust score
| actor | reliability | predictability | MTTR | blast | total |

### Recovery 패턴 적용
| actor | retry | circuit | fallback | bulkhead | timeout |

### Cascade
- {trigger} → {affected actors}
- 차단 pattern: ...

### Improvement 우선순위
1. ...
```

## 6. 검증

- [ ] 4 actor 분류 모두 매트릭스 측정?
- [ ] Trust score 4 차원 산출?
- [ ] Recovery 패턴 6 종 actor 별 적용 여부?
- [ ] Cascade pattern 식별 + 차단 검증?
- [ ] Improvement 우선순위 (가장 낮은 trust → 우선)?

## 7. 다음 phase

- `audit-error-budget` (§8) — actor 별 budget burn 입력
- `chaos-test` (§6) — recovery 패턴 검증 실험
- `conduct-postmortem` (구현됨) — incident 학습

## 8. 참조

- Release It! (Michael Nygard) — circuit breaker / bulkhead / timeout patterns
- agent-evaluation (외부 reference, MIT — Kevin + Claude, OMAS v2 trust 패턴)
- Google SRE Book Ch.10 — postmortem / cascade
