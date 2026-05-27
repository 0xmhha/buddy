# Design Observability — logs / metrics / traces 3-pillar + SLO/SLI

## 1. 목적

`design-system` 안에서 *production observability* 를 사전 설계. **3 pillar** (logs / metrics / traces) 의 책임 / 도구 / cost / retention 결정 + **SLO/SLI** 정의 + **alert 정책** 작성.

`design-data-model` / `design-api-contract` 와 cascade — observability 가 schema / API 흐름 추적의 foundation.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| 시스템 토폴로지 | 선택 | artifact | `derive-system-topology` 산출물 | 없으면 단일 서비스 가정 |
| SLO 목표 | ✅ | knowledge | 사용자 도메인 지식 | "핵심 SLO 지표는? (가용성 99.9%, 응답시간 p99 < 500ms 등)" |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Observability 전략 (logs/metrics/traces/SLO 구성) | artifact | structured YAML | `setup-incident-paging`, `audit-error-budget` |

## 2. 사용 시점

- §3 design-system 안에서 tech stack 결정 (`define-tech-stack`) 후
- `prepare-launch-checklist` 직전 — observability gate
- 사고 (incident) 후 사후 — observability 의 *gap 식별*
- 새 actor / 외부 의존 추가 시 — 추적 영역 확장
- cost 분석 (`audit-cost-efficiency`) 직전 — observability 비용 추정 입력

## 3. 입력

### 필수
- `define-tech-stack` 산출 — 언어 / 프레임워크 / hosting
- `derive-system-topology` 산출 — sub-component 분포
- 사용자 SLA 가설 (있으면) — 가용성 / latency 목표

### 선택
- 기존 incident log (postmortem 기반 *추적 부족 영역* 식별)
- 비용 budget 가설

## 4. Stage 흐름

### Stage 1: 3 pillar 책임 분담

| pillar | 무엇 | 도구 후보 | retention |
|--------|------|---------|---------|
| Logs | structured event (시간 + level + context) | OpenTelemetry / Loki / CloudWatch | 7~30일 (디버그) |
| Metrics | 시계열 수치 (count / latency / error rate) | Prometheus / Datadog / CloudWatch metric | 90일~2년 |
| Traces | request 흐름 (span tree, cross-service) | OpenTelemetry / Jaeger / Tempo | 7~14일 (sampling) |

### Stage 2: SLI / SLO / error budget 정의

각 critical user journey 에:

| 항목 | 정의 |
|------|------|
| SLI | 측정 가능한 신호 (e.g. 99.9% requests 200ms 이내 응답) |
| SLO | 목표 (e.g. 99.9% / month) |
| error budget | 100% - SLO (e.g. 0.1% = 43.2 min / month 다운타임) |

→ `audit-error-budget` skill (§8) 의 입력.

### Stage 3: Alert 정책

| 신호 | trigger | 대응 |
|------|--------|------|
| Error rate spike | 5min window 5%+ | on-call page |
| Latency p99 above SLO | 10min window 200ms+ | warn |
| Error budget burn | 25%/day | review release |

alert 너무 많으면 *fatigue* — top-3 critical 만.

### Stage 4: Cost 추정

3 pillar 각 도구 의 *월 비용 추정*:
- Logs ingestion: $/GB
- Metrics cardinality: $/series
- Traces sample rate × storage

총 observability 비용 가이드: 인프라 비용의 **5~15%**.

### Stage 5: Privacy + secret 처리

logs / traces 에 PII / secret 들어가면 안 됨. *redaction* 정책:
- secret 패턴 (Bearer / api_key / password) auto-redact
- PII 필드 (email / phone) hash or drop

→ `design-secret-management` skill 정합.

## 5. 산출물 형식

```markdown
## Observability Design — {제품}

### 3 pillar
| pillar | 도구 | retention | cost (월) |
|--------|------|---------|---------|

### SLI / SLO / error budget
| journey | SLI | SLO | budget |

### Alert top-3
| 신호 | trigger | 대응 |

### Cost 합계
- Logs: $X / Metrics: $Y / Traces: $Z = $TOTAL (인프라 ~ %)

### PII / secret redaction 정책
- 패턴: ...
```

## 6. 검증

- [ ] 3 pillar 모두 도구 + retention 결정?
- [ ] SLI / SLO / error budget *측정 가능* 형태?
- [ ] Alert top-3 (alert fatigue 회피)?
- [ ] Cost 추정 인프라의 5~15% 이내?
- [ ] PII / secret redaction 정책?

## 7. 다음 phase

- `prepare-launch-checklist` 의 observability gate 입력
- `audit-error-budget` (§8) 의 SLO / budget 입력
- `design-secret-management` 와 redaction 정합 검증

## 8. 참조

- Google SRE Book Ch.3-4 — SLO / error budget
- OpenTelemetry specification — 3 pillar 표준
- Observability Engineering (Charity Majors) — high-cardinality observability
