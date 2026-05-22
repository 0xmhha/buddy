# define-tech-stack — 기술 스택 결정과 락인 영향 평가

기술 스택 결정 (language / framework / DB / cache / queue / hosting / observability / CI) 은 락인 영향이 매우 큰 다년 단위 의사결정이다. 잘못된 선택의 비용은 마이그레이션·재작성으로 환산되어 수개월~수년의 엔지니어링을 잠식한다. 본 skill 은 **8 차원 분해 + 차원별 alternatives 평가 + cross-차원 호환성 매트릭스 + 5년 lock-in 정량 평가** 를 강제해 의사결정 누수를 차단한다. 산출물은 표 + risk register + ADR draft 로 다음 단계 (`write-adr`) 에 연결된다.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지하기 위한 절차다. 산출물에 다음이 발견되면 검증 실패로 간주하고 §5 로 돌아가 보강한다:

- **단일 후보만 평가** — "Postgres 쓰면 됩니다" 식. 차원당 최소 3 alternatives 비교 필수.
- **카테고리 답변 채택** — "modern, scalable, fast" 류. 숫자 (latency, RPS, $/month) 와 조건 (어떤 워크로드에서) 으로 강제 변환.
- **사용자 입력 무비판 채택** — 사용자가 "Node.js + MongoDB" 라 명시해도 그 선택의 lock-in 과 alternatives 를 challenge 하지 않으면 skill 의 가치가 무너진다.
- **Lock-in cost 정성 평가만** — "high / medium / low" 로 끝내지 말 것. "Postgres → DynamoDB 마이그레이션 시 약 6 person-month + downtime risk" 같은 정량 추정 필수.
- **Cross-차원 충돌 무시** — 예: "Bun + Deno 라이브러리" 처럼 incompatible 조합. 매트릭스 검증 필수.

§5 의 모든 phase 를 누락 없이 수행하라. skip 시 산출물의 신뢰도가 무너진다.

## 1. 목적

기술 스택 결정의 의사결정 누수 — "그냥 익숙한 거" / "팀이 좋아하는 거" / "최신이라 멋있어 보이는 거" — 를 evidence-based decision 으로 강제 변환한다. 결정 자체는 사용자/팀이 하지만, 본 skill 은:

1. 차원을 빠짐없이 분해 (8+ dimensions)
2. 차원당 alternatives 강제 비교 (≥3)
3. cross-차원 충돌 노출
4. 5년 lock-in 정량 시나리오 작성
5. ADR draft 로 결정 영속화 준비

까지 강제한다.

## 2. 사용 시점 (When to invoke)

- PRD 확정 후 §3 Technical Design 첫 단계로 진입
- 신규 service / sub-system 추가 시 stack 결정
- 기술 부채 누적으로 stack 일부 재평가 (예: SQL → NoSQL migration 검토)
- 신규 팀원 onboarding 비용이 hiring pool 부족으로 임계 도달 시
- production 운영 1+ 년 후 unit economics 재계산 / cost spike 시 hosting 재평가

## 3. 입력 (Inputs)

### 필수
- Feature backlog (또는 PRD) — 어떤 워크로드가 예상되는지
- 팀 보유 expertise (현재 팀이 익숙한 stack)
- 운영 환경 제약 (cloud provider, on-prem, regulated industry 등)
- Budget (월 운영비 상한, 인건비)
- Latency / throughput / availability SLO

### 선택
- 기존 brownfield code (있다면 호환성 강제 조건)
- comparable companies 의 공개 stack 정보
- 향후 12~24 개월 hiring plan (팀 확장 시 hiring pool 영향)

### 입력이 부족할 때 forcing question
- "예상 일일 active user / RPS / data growth rate 는 얼마인가? 추정 근거는?"
- "팀이 production 운영 경험 있는 stack 은? 새로 배워야 하는 stack 으로 가면 학습 비용 + 초기 incident 빈도가 어디까지 허용 가능한가?"
- "이 시스템이 5년 후에도 같은 형태일 가능성? 아니면 1~2년 안에 큰 pivot 가능성? lock-in 허용 범위가 다르다."
- "Cost SLO 가 latency SLO 와 충돌하면 (예: cold start vs always-on) 어느 것을 우선?"

## 4. 핵심 원칙 (Principles + Posture)

이 skill 의 운영 posture:

- **입장 취함, hedge 금지** — 차원별 추천을 명시. "Postgres vs MySQL 둘 다 가능" 같은 fence-sitting 거부. "이 워크로드 (write-heavy + JSON column) 에는 Postgres 권장. 마음 바꿀 조건: write throughput 이 50K/sec 초과 시 ScyllaDB 재평가" 같은 형식.
- **사용자 입력을 challenge** — "Node.js 가 좋다" 발화에 그대로 따르지 말고 "왜 Node? 팀 expertise 외 객관 근거? Latency SLO 와 single-thread bottleneck 검토했나?" 로 push.
- **Specificity 강제** — "fast", "scalable", "modern" 거부. p99 latency (ms), RPS, scale 한계 (rows, GB), churn 비용 ($) 으로 변환.
- **Boring-by-default bias** — "최신 / hype" 보다 "proven 5+ 년 production 운영 사례" 우선. novelty 채택 시 분명한 근거 + risk mitigation 명시.

도메인 원칙:

1. **Lock-in 은 비용으로 환산** — "high lock-in" 만으론 부족. "vendor X 떠나려면 Y person-month + Z downtime risk" 형태로 정량화.
2. **Hiring pool 은 실제 시장 데이터** — "hire 가능한 engineer 수" 를 LinkedIn / Stack Overflow Survey 등 public data 로 추정. "우리 회사가 채용 가능한 비율" 까지 보정.
3. **Compatibility matrix 는 필수** — language ↔ framework, framework ↔ DB driver, DB ↔ ORM, hosting ↔ runtime 의 known 호환성 / known 충돌 명시.
4. **Cost model 은 unit economics 와 결합** — $/req, $/GB stored, $/GB transferred 로 분해. SLO 와 trade-off 매트릭스 작성.
5. **5년 horizon 시나리오 ≥ 2개** — best-case (모든 가정 성립) / worst-case (가정 깨짐 + migration 강제). 각 시나리오의 stack 적합도 평가.

## 5. 단계 (Phases)

### Phase 1. Stack 차원 분해

다음 8+ 차원을 빠짐없이 enumerate. 각 차원이 의사결정 단위.

| Dimension | 결정해야 할 것 | 예시 후보 |
|-----------|--------------|---------|
| Language | primary backend / frontend / scripting | Go / Rust / Python / TS / Java |
| Backend framework | request handling, ORM, middleware | Gin / Fastify / FastAPI / Spring |
| Frontend framework | rendering, routing, state | Next.js / Remix / SvelteKit / Native |
| Primary DB | OLTP / general persistence | Postgres / MySQL / MongoDB / DynamoDB |
| Cache | hot path acceleration | Redis / Memcached / Hazelcast |
| Queue / async | background jobs, event flow | Kafka / RabbitMQ / SQS / NATS |
| Hosting / runtime | compute layer | AWS Fargate / GCP Cloud Run / Vercel / k8s |
| Observability | logging / metrics / tracing | Datadog / Grafana stack / New Relic / OpenTelemetry |
| CI/CD | build / test / deploy | GitHub Actions / GitLab CI / Buildkite |

추가 차원 (해당 시): 검색 (Elasticsearch / Meilisearch), 분석 DB (ClickHouse / BigQuery), feature flag (LaunchDarkly / Unleash), secret manager (Vault / AWS SM).

각 차원이 빠지지 않았는지 사용자와 함께 cross-check.

### Phase 2. 차원별 ≥ 3 후보 + 5축 scoring

각 차원에 대해 최소 3 개의 alternatives 를 쓰고 다음 5축 으로 scoring (1~10):

| Axis | 설명 |
|------|------|
| Fit | 본 워크로드 / SLO 적합도 |
| Maturity | production-grade 운영 사례 + 안정성 |
| Hiring pool | 시장에서 뽑을 수 있는 engineer 밀도 |
| Cost | $/req or $/month + 운영 인건비 |
| Lock-in | exit cost (낮을수록 좋음 → 높은 점수가 낮은 lock-in) |

scoring 은 evidence-based 여야 한다. 각 점수 옆에 1줄 근거.

### Phase 3. Cross-차원 호환성 매트릭스

후보 조합의 known 호환성 / 충돌을 매트릭스로 정리:

```
              Postgres   MongoDB   DynamoDB
Node.js       ✓ (pg)    ✓ (driver) ✓ (sdk)
Go            ✓ (pgx)   ✓ (mongo-go-driver) ✓ (aws-sdk-go)
Python        ✓ (asyncpg) ✓ (motor) ✓ (aioboto3)

              GitHub Actions   GitLab CI   Buildkite
AWS Fargate   ✓               ✓           ✓
Vercel        △ (제한)         ✗           ✗
GCP Cloud Run ✓               ✓           ✓
```

명시적 ✗ 또는 △ 가 있는 조합은 제외 또는 추가 검토 사유 명시.

### Phase 4. 선택 + 결정 근거 작성

> **AI 편향 차단 게이트** — 본 Phase 시작 전 *3개 이상의 orthogonal한 stack 후보*를 발산시킨다. 발산은 직접 또는 `verify-best-alternative` 호출로. *첫 답으로 commit 금지*. Phase 4 산출물에 3개 이상 후보의 *rubric 비교 표*(언어/framework·DB·hosting·observability·5년 lock-in)가 나란히 존재해야 §11 검증 게이트 통과. 체크박스만 체크하는 회피는 *anti-rationalization 위반*. 적용 근거: [`docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md`](../../../../docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md) §3.

차원별로 선택을 확정. 각 결정에 다음 형태로 근거 명시:

> Selected: <option>
> Why: <1-2 문장 evidence-based reason>
> Alternatives rejected: <other2-3 후보 + rejection reason 1줄>
> Change conditions: <어떤 measurable signal 이 발견되면 재평가>

### Phase 5. 5년 lock-in risk register + ADR handoff

각 결정의 5년 시나리오 작성:

- **Best-case**: 가정대로 → 적합도 유지
- **Worst-case**: scale / cost / vendor 변화 → 마이그레이션 강제. 그 비용을 person-month + downtime risk + 데이터 마이그레이션 복잡도로 정량 추정.

마지막에 `write-adr` skill 호출 권장 (chain): 본 skill 의 출력을 ADR draft 의 Context / Decision / Consequences / Alternatives 섹션 입력으로 사용.

## 6. 산출물 형식 (Output format)

> **Note**: Opus 4.7 / Sonnet 4.6 default 는 prose 출력. 다음 구조를 **명시적으로 요구**해야 모델이 structured 출력함.

다음 형식으로 출력하라 (요약 / prose 변환 금지, 모든 섹션 채우기 강제):

```markdown
## define-tech-stack Output — <project name>

### Summary
<3 줄: 어떤 워크로드 / 핵심 결정 3개 / 가장 큰 lock-in 1개>

### Dimensions Evaluated
| # | Dimension | Selected | Alternatives Considered | Rationale | Lock-in Cost (5y est.) |
|---|-----------|----------|-------------------------|-----------|------------------------|
| 1 | Language | ... | ... / ... / ... | ... | ... person-month |
| 2 | Backend framework | ... | ... / ... / ... | ... | ... |
| ... | ... | ... | ... | ... | ... |

(8+ rows minimum)

### Cross-dimension Compatibility
| Combination | Status | Note |
|-------------|--------|------|
| <lang> × <framework> | ✓ / △ / ✗ | ... |
| <framework> × <DB driver> | ... | ... |

(at minimum: language × framework, framework × DB, DB × ORM, hosting × runtime)

### Risk Register (5-year horizon)
| # | Risk | Likelihood (1-5) | Impact ($/PM) | Mitigation | Trigger to re-evaluate |
|---|------|-------------------|---------------|------------|------------------------|
| 1 | <risk> | ... | ... | ... | <measurable signal> |
| 2 | ... | ... | ... | ... | ... |

(at minimum 5 rows for 8+ dimensions)

### ADR Handoff
다음 단계: `/buddy:write-adr "<title>"` 를 호출해 본 산출물의 Decision / Alternatives / Consequences 섹션을 ADR 형식으로 영속화. 또는 chain:
`/buddy:chain define-tech-stack,write-adr -- "<feature>"`

### Next Step
<구체 action — 1줄: 예 "design-data-model 호출해 Postgres 스키마 설계 시작">
```

## 7. Cross-phase cascade

본 skill 의 산출물이 후속 phase 에 미치는 영향:

- **§3 design-data-model**: 선택된 DB 가 schema 설계 옵션을 제약 (Postgres → relational, DynamoDB → key-value patterns)
- **§3 design-api-contract**: 선택된 backend framework 가 typical API style 결정 (Spring → REST, Apollo → GraphQL, gRPC stack)
- **§4 plan-build**: 선택된 stack 별 hiring pool 이 actor track 분해 시 worker 가용성 제약
- **§6 verify-quality**: 선택된 observability stack 이 test infra 결정 (Datadog APM 통합 여부 등)
- **§7 ship-release**: hosting / CI 선택이 deploy 전략 (canary, feature flag) 의 옵션 공간 정의

## 8. 다음 skill (next in stage flow)

- `write-adr` — 본 skill 의 결정을 표준 ADR 양식으로 영속화 (강력 권장)
- `design-data-model` — DB 결정 후 schema / migration / index 설계
- `design-api-contract` — backend framework 결정 후 endpoint contract 설계

권장 chain: `define-tech-stack → write-adr → design-data-model → design-api-contract` (의존 순서대로)

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `consult-design-system`** — 이 skill 은 backend / infra / data stack 결정. consult-design-system 은 UI design system (typography, color, spacing) 결정. 영역 분리.
- **vs `design-mcp-server`** — 본 skill 은 일반 stack. design-mcp-server 는 MCP server 특수 설계 (tool exposure, transport). MCP server 가 system 의 일부라면 본 skill 이 stack 결정 후 design-mcp-server 가 후속 invoke.
- **vs `design-data-model`** — 본 skill 은 DB *선택* (Postgres 인지 MongoDB 인지). design-data-model 은 선택된 DB 의 *스키마 설계* (table / index / migration). 본 skill 이 먼저.
- **vs `design-api-contract`** — 본 skill 은 backend framework *선택* (Spring 인지 FastAPI 인지). design-api-contract 는 endpoint *contract 설계* (REST / GraphQL / RPC). 본 skill 이 먼저.
- **vs `verify-best-alternative`** — verify-best-alternative 는 *AI 편향 방지 다관점 검토 메커니즘* (모든 엔지니어링 결정 sub-step). 본 skill 은 stack 결정의 *전체 절차* — 본 skill *내부에서* stack 후보 평가 시 verify-best-alternative를 의무 호출.

## 10. 중요 규칙

- **Read-only on production state** — 코드·infra 수정 안 함. 결정과 추천만 산출.
- **Evidence > opinion** — 모든 점수 / 결정에 근거 1줄 이상.
- **No silent assumption** — 사용자 입력의 결정 영향이 큰 가정 (예: "회사가 AWS 만 쓴다") 은 명시적으로 confirm.
- **Cross-dimension matrix 누락 금지** — 8+ 차원의 모든 가능 조합을 evaluate 안 해도 되지만, 위험한 조합 (라이브러리 호환성, runtime 호환성) 은 반드시 매트릭스로 검증.
- **Lock-in 정량화 의무** — "high / medium / low" 만으로 끝내면 acceptance fail.

## 11. Verification gate — 완료 선언 전 self-check

다음 체크가 모두 yes 여야 절차 완료 보고:

- [ ] §5 의 5 phase 가 누락 없이 실행됨
- [ ] §6 의 4 출력 섹션 (Summary / Dimensions / Compatibility / Risk Register / ADR Handoff / Next Step) 모두 채워짐
- [ ] Dimensions Evaluated 표가 ≥ 8 row
- [ ] 각 dimension 에 ≥ 3 alternatives 가 검토됨
- [ ] Cross-dimension Compatibility 매트릭스가 ≥ 4 조합 평가
- [ ] Risk Register 가 ≥ 5 row, 각 row 에 정량 impact (PM 또는 $) + 재평가 trigger 포함
- [ ] §4 의 posture (입장 / specificity / challenge / boring-by-default) 가 산출물에 적용됨 — hedge 표현 없음
- [ ] §0 의 5 anti-pattern 들이 산출물에 등장하지 않음
- [ ] ADR handoff 라인이 출력에 명시됨
- [ ] **`verify-best-alternative` 1회 이상 호출 완료** — AI 편향 방지 의무. stack 결정(언어/프레임워크/DB/runtime/hosting)이 *첫 답*이 아니라 다관점 검토 후 *어떤 관점에서 봐도 최선*임을 확인. 자동 dispatch 또는 명시 호출 — `/buddy:verify-best-alternative "stack 결정 컨텍스트"`
- [ ] **3+ orthogonal 후보의 rubric 비교 표가 산출물에 존재** — 체크박스만 체크하는 *anti-rationalization 회피* 금지. Phase 4 산출물의 dimensions evaluated 표 또는 별도 rubric 표로 증명

하나라도 no 면 해당 phase 로 돌아가 보강 후 재검증. 사용자에게 incomplete 산출물을 "충분하다" 고 보고하지 말 것.
