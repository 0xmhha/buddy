# design-data-model — 스키마 / 마이그레이션 / 인덱싱 결정

데이터 모델 결정은 production 운영 시 변경 비용이 매우 큰 영역이다. 잘못된 스키마 / 인덱스 / migration 전략은 downtime · 데이터 손실 · 재작성 비용으로 환산된다. 본 skill 은 **entity 관계 매핑 + read/write 패턴 분류 + normalization 결정 + 인덱스 전략 + zero-downtime migration 강제** 절차로 production 변경 비용을 사전 평가한다. tech stack 결정 (`define-tech-stack`) 다음 단계로 호출된다.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지하기 위한 절차다. 산출물에 다음이 발견되면 검증 실패로 간주하고 §5 로 돌아가 보강한다:

- **Read/write 패턴 분류 생략** — OLTP 인지 analytical 인지, hot path 인지 cold 인지 명시 안 하고 스키마 짜는 것. 인덱스 전략이 무근거가 된다.
- **3NF rubber-stamp** — 무조건 정규화하거나 무조건 denormalize. 결정엔 read pattern 근거 + cost trade-off 필요.
- **Migration plan 부재** — 스키마 변경의 zero-downtime / forward-compatible 전략 없으면 production 적용 불가.
- **Index 전략을 "필요시 추가" 로 미룸** — hot path query plan 을 미리 trace 하지 않으면 운영 후 N+1 / full scan 으로 고통.
- **Scale 가정 명시 안 함** — "rows per day", "GB per month" 추정 없이 스키마 결정. 1년 후 partitioning / sharding 강제 마이그레이션.

§5 의 모든 phase 를 누락 없이 수행하라.

## 1. 목적

데이터 모델 의사결정 누수 — "쿼리 짜다 보면 알게 되겠지" / "Postgres 니까 그냥 정규화" / "ORM 이 알아서 해주겠지" — 를 evidence-based 결정으로 강제 변환. entity / 관계 / read pattern / write pattern / scale 을 빠짐없이 분해해 운영 변경 비용을 production 전에 노출시킨다.

## 2. 사용 시점 (When to invoke)

- `define-tech-stack` 으로 DB 결정 직후 (chain 권장)
- 신규 entity 또는 sub-domain 추가 시
- read pattern 변화 (예: dashboard 추가) 로 인덱스 재설계 필요
- write throughput 임계 도달로 partitioning / sharding 검토
- migration 실패 incident 후 재설계
- multi-tenant model 도입 / 변경 시

## 3. 입력 (Inputs)

### 필수
- `define-tech-stack` 산출물 (선택된 DB)
- entity 초안 (1줄 명사 list 라도)
- 예상 read/write pattern (간단한 user story 수준)
- scale 추정: rows/sec, GB total, 성장률 (월/연)

### 선택
- 기존 brownfield schema (있다면 호환성 강제 조건)
- 산업 규제 요구사항 (PII / GDPR / HIPAA 등 retention/encryption)
- read replicas / multi-region 요구
- analytical workload (BI dashboard 등) 동거 여부

### 입력이 부족할 때 forcing question
- "이 entity 의 access pattern 1순위는? 단일 lookup, range scan, full-text search, aggregation? 그 비율은?"
- "write rate 가 peak 일 때 RPS 추정? hot row (예: 인기 게시물) 의 write 집중도는?"
- "1년 후 데이터 크기 추정? 그 시점에 이 스키마가 partitioning 없이 견디나?"
- "관계가 있다면 cascading 삭제 / soft delete 정책? GDPR 삭제 요구 처리는?"

## 4. 핵심 원칙 (Principles + Posture)

이 skill 의 운영 posture:

- **입장 취함, hedge 금지** — "정규화 vs denormalize, 둘 다 가능" 거부. read/write 비율 + scale 근거로 한쪽 선택 + 변경 조건 명시.
- **사용자 입력을 challenge** — "id, name, email 만 있는 user table" 발화에 그대로 따르지 말고 "tenant_id 는? created_at index 는? soft delete column 은?" 으로 push.
- **Specificity 강제** — "scalable" 거부. RPS / GB / index size MB / query time p99 ms 로 변환.
- **Migration safety bias** — 모든 schema 변경은 zero-downtime + forward-compatible 가능한지 먼저 평가. 안 되면 그 trade-off 명시.

도메인 원칙:

1. **Read pattern 이 인덱스 결정** — write-heavy → minimal index. read-heavy → covering index. 둘 다 → workload 분리.
2. **Normalization decision 은 trade-off** — 3NF 가 default 지만 hot path read 는 denormalize 가치 있음. 명시적으로 결정.
3. **Migration plan 은 schema 변경의 일부** — 새 column / index / table 만 만드는 게 아니라 어떻게 deploy 하는지 (expand-contract / online migration) 까지 결정.
4. **Index 는 cost** — disk, write amplification, vacuum/maintenance 비용. covering / partial index 로 최소화.
5. **PII / encryption 은 schema 단계에서 결정** — column 레벨 암호화, hash 컬럼, retention column 을 schema 에 미리 박는다. 후속 추가는 마이그레이션 비용 큼.

## 5. 단계 (Phases)

### Phase 1. Entity 식별 + 관계 매핑

각 entity 의 attribute 와 관계를 ER 다이어그램 또는 표로 명시:
- entity 이름
- primary key (single column / composite)
- 외래 관계 (1:1, 1:N, N:M)
- 핵심 attribute (type, nullable, constraint)
- soft delete / audit column (created_at, updated_at, deleted_at)

### Phase 2. Read/write 패턴 분류

각 entity 또는 query 에 대해 분류:

| 차원 | 옵션 |
|------|------|
| Workload type | OLTP / OLAP / mixed |
| Access pattern | single-row lookup / range scan / aggregation / full-text search |
| Hot vs cold | hot path (frequent) / warm / cold (archive) |
| Read/write ratio | read-heavy (>10:1) / balanced / write-heavy |
| Latency target | p99 ms (예: 50ms) |
| RPS estimate | now / 1y / 3y |

### Phase 3. Normalization decision

> **AI 편향 차단 게이트** — 본 Phase 시작 전 *3개 이상의 normalization 전략 후보*(예: 3NF / denormalized read-model / event-sourced / document) + *PK/index 전략* 후보 발산. 직접 또는 `verify-best-alternative` 호출. *첫 답 commit 금지*. 산출물에 3+ 후보의 *(write 단순성·read 효율·migration 비용·schema evolution)* rubric 비교 존재해야 §11 통과. 적용 근거: [`docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md`](../../../../docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md) §3.

각 entity 별로 결정:

| Entity | Default | Decision | Reason |
|--------|---------|----------|--------|
| <name> | 3NF | normalized / denormalized / hybrid | 1줄 근거 (read pattern + scale) |

denormalize 한 경우 update consistency 전략 (cache invalidation, materialized view refresh, CQRS 등) 명시.

### Phase 4. Index 전략

각 entity 의 hot path query 를 enumerate 후 index 설계:

| Query | Used columns | Index type | Justification |
|-------|--------------|-----------|---------------|
| <SQL or pseudo> | <cols> | btree / hash / GIN / GiST / partial / covering | <왜 이 index> |

trade-off 명시: write penalty (per-write 추가 비용) vs read gain.

### Phase 5. Migration plan + ADR handoff

Schema 변경의 deploy 전략:

- **Expand-contract pattern**: 새 column 추가 → application 양쪽 호환 → old column drop
- **Online migration**: 인덱스 concurrent build, table rewrite avoidance
- **Backfill 전략**: batch size, throttling, idempotency
- **Rollback 가능성**: 각 단계가 reversible 한지

마지막에 `write-adr` skill 호출 권장: schema 결정 + normalization 결정 + index 전략을 ADR 양식으로 영속화.

## 6. 산출물 형식 (Output format)

> **Note**: Opus 4.7 / Sonnet 4.6 default 는 prose. 다음 구조를 **명시적으로 요구**해야 structured 출력.

다음 형식으로 출력하라 (요약 / prose 변환 금지, 모든 섹션 채우기 강제):

```markdown
## design-data-model Output — <project / sub-domain name>

### Summary
<3 줄: workload type / 핵심 entity 수 / 가장 큰 migration risk 1개>

### Entities
| Entity | PK | Relations | Key Attributes | Audit Columns |
|--------|-----|-----------|----------------|---------------|
| ... | ... | ... | ... | ... |

### Read/Write Pattern Classification
| Entity / Query | Workload | Access Pattern | Hot/Cold | R/W Ratio | p99 Target | RPS (now/1y/3y) |
|----------------|----------|----------------|----------|-----------|-----------|------------------|
| ... | ... | ... | ... | ... | ... | ... |

### Normalization Decisions
| Entity | Decision | Reason | Consistency Strategy (if denormalized) |
|--------|----------|--------|----------------------------------------|
| ... | ... | ... | ... |

### Index Strategy
| Query | Index | Type | Justification | Write Penalty |
|-------|-------|------|---------------|---------------|
| ... | ... | ... | ... | ... |

### Migration Plan
| Step | Operation | Strategy (expand-contract / online / backfill) | Rollback | Estimated downtime |
|------|-----------|------------------------------------------------|----------|--------------------|
| 1 | ... | ... | ... | ... |

### Risks
| # | Risk | Likelihood (1-5) | Impact (PM / data loss / downtime) | Mitigation | Trigger to re-evaluate |
|---|------|-------------------|-------------------------------------|------------|------------------------|
| 1 | ... | ... | ... | ... | ... |

### ADR Handoff
다음 단계: `/buddy:write-adr "<title>"` 또는 chain `/buddy:chain design-data-model,write-adr -- "<feature>"`.

### Next Step
<구체 action — 1줄: 예 "design-api-contract 호출해 API resource 와 schema entity 매핑 시작">
```

## 7. Cross-phase cascade

- **§3 design-api-contract**: entity 가 API resource 와 어떻게 매핑되는지 (1:1 또는 aggregation) 결정 입력
- **§4 plan-build**: schema migration 이 deploy task graph 의 critical path 항목
- **§5 build-feature**: ORM model / repository layer 구현의 schema 입력
- **§6 verify-quality**: test fixture / test DB seeding 전략 입력
- **§7 ship-release**: migration 의 deploy 순서 / rollback 절차 입력
- **§8 iterate-product**: index 효과 측정 / hot path 변경 감지 metric 정의

## 8. 다음 skill (next in stage flow)

- `write-adr` — 본 skill 의 결정 영속화
- `design-api-contract` — entity → API resource 매핑
- (별도 plan) `design-event-schema` — async event / queue payload 설계 시
- (별도 plan) `design-tenant-model` — multi-tenant 결정 시

권장 chain: `define-tech-stack → design-data-model → write-adr → design-api-contract`

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `define-tech-stack`** — define-tech-stack 은 DB *선택* (Postgres / MongoDB / DynamoDB). 본 skill 은 그 위의 *스키마 / 인덱스 / migration* 설계. 본 skill 이 후속.
- **vs `design-api-contract`** — 본 skill 은 DB 모양 (table / collection). design-api-contract 는 API 모양 (REST resource / GraphQL type). 둘은 매핑되지만 같지 않음 (예: API resource 가 여러 DB row aggregation).
- **vs `design-event-schema`** (미구현) — 본 skill 은 영속 데이터 schema. design-event-schema 는 transient event payload schema. 둘 다 필요한 경우 본 skill 후 별도 호출.
- **vs `design-tenant-model`** (미구현) — multi-tenant 결정 (shared DB row-level vs schema-per-tenant vs DB-per-tenant) 은 본 skill 의 입력. tenant 모델 미정 시 본 skill 진행 전 먼저 결정.

## 10. 중요 규칙

- **Read-only on production data** — 코드/DB 수정 안 함. 결정·DDL draft·migration plan 만 산출.
- **Migration plan 누락 금지** — schema 변경 결정은 deploy plan 까지 포함해야 완료.
- **Scale 가정 명시 의무** — "rows / GB / RPS" 추정 없이 스키마 결정 금지.
- **Index 결정 근거 명시** — 모든 index 에 hot query + write penalty trade-off 1줄.
- **GDPR / PII 처리 column 레벨 결정** — schema 단계에서 encryption / retention column 결정.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 7 출력 섹션 (Summary / Entities / R-W Pattern / Normalization / Index / Migration / Risks / ADR Handoff / Next Step) 모두 채워짐
- [ ] Entities 표가 모든 필수 entity 포함, audit column (created_at, updated_at) 명시
- [ ] R/W Pattern 분류표가 hot path 모든 query 포함
- [ ] Normalization decision 근거에 "read pattern + scale" 둘 다 등장
- [ ] Index Strategy 표가 hot query 별 index + write penalty 평가
- [ ] Migration Plan 의 모든 step 이 expand-contract 또는 online 명시, downtime 추정
- [ ] Risks 표가 ≥ 3 row, 정량 impact (PM / data loss / downtime) 포함
- [ ] §4 posture 적용 — hedge 표현 없음, "scalable / fast" 류 카테고리 답변 없음
- [ ] §0 anti-pattern 들이 산출물에 등장하지 않음
- [ ] ADR handoff 라인 명시
- [ ] **`verify-best-alternative` 1회 이상 호출 완료** — AI 편향 방지 의무. normalization / index 전략 / migration 접근이 *첫 답*이 아니라 다관점 검토 후 최선임을 확인
- [ ] **3+ orthogonal normalization 후보의 rubric 비교 표가 산출물에 존재** — 체크박스만 체크하는 *anti-rationalization 회피* 금지. Phase 3의 entity별 결정 + Phase 4의 index 전략 비교 표로 증명

하나라도 no 면 해당 phase 로 돌아가 보강 후 재검증.
