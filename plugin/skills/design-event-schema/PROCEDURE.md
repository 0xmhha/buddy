# design-event-schema — async event schema-first design

§3 Technical Design phase 의 stage. **design-api-contract 의 sync-only gap 을 채우는 async layer**. SQS / Kafka / Webhook / EventBridge / Pub-Sub 의 event payload schema-first 설계 + producer/consumer contract + versioning + dead-letter handling + idempotency. 산출물은 event inventory + schema registry + delivery guarantee + drift detection 전략.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Schema-less event** — JSON blob 으로 시작 → consumer 마다 다른 가정, 첫 schema 변경에서 모두 깨짐.
- **Versioning 정책 없음** — event v1 → v2 시 호환성 전략 부재. dual-publish vs upgrade-in-place 결정 필요.
- **DLQ 전략 부재** — consumer 실패 시 어디로 가는지 불명. message loss / infinite retry storm risk.
- **Idempotency 없는 consumer** — at-least-once delivery 인데 dedupe 안 함 → side effect 중복 (이메일 2 회 발송 등).
- **Producer / Consumer 양방향 contract 안 정의** — schema 만 있고 ordering / partitioning / retention 결정 없음.
- **Sync 와 통합 design 부재** — sync API 와 async event 가 같은 도메인 데이터 만지는데 design 분리.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

async event 의 schema-first contract. design-api-contract 가 sync layer 결정한 후 본 skill 이 async layer 결정 — 두 layer 모두 schema-driven 으로 통일.

## 2. 사용 시점 (When to invoke)

- design-api-contract 후 (sync layer 결정 후 async layer)
- async event 가 도메인의 핵심인 경우 (event-sourcing / CQRS / SNS-fanout)
- 신규 message broker 도입 시 (SQS → Kafka migration 등)
- consumer 추가 / 변경 시 contract 검증
- DLQ / retry storm incident 후 보강

## 3. 입력 (Inputs)

### 필수
- §3 design-api-contract 산출물 (sync layer 결정)
- §3 define-tech-stack ADR (queue / event broker 결정 — SQS / Kafka / EventBridge)
- §3 derive-system-topology (async edge enumeration)
- §2 feature spec 의 async event use case (signup → email 등)

### 선택
- 기존 event registry (Confluent Schema Registry / AWS Glue Schema Registry)
- compliance 의무 (event 에 PII 포함 시 retention / encryption)
- multi-region event (cross-region replication 결정)

### 입력이 부족할 때 forcing question
- "broker 결정됐나? SQS vs Kafka vs EventBridge — define-tech-stack ADR 입력. 미결정 시 본 skill 무근거."
- "ordering 의무 있나? FIFO 필요한 use case (signup→verify) 와 unordered OK 인 use case (audit log) 분리 필요."
- "consumer 의 idempotency key source 가 명시됐나? Idempotency-Key header / event id / business key 중 어느 것?"

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — broker / schema format / versioning 단호.
- **사용자 입력을 challenge** — "SQS 쓸게" 발화에 "ordering 필요? FIFO vs standard?" push.
- **Specificity 강제** — vague "event publish" 거부, "EmailEnqueued event with `to_email_hash` partition key, FIFO group, 14d retention".
- **Contract-first** — schema 가 code 보다 먼저, JSON Schema / Protobuf / Avro 중 결정.

도메인 원칙:

1. **Event taxonomy** — domain event / integration event / command 분류.
2. **Schema format 결정** — JSON Schema (낮은 진입장벽), Protobuf (성능 + cross-lang), Avro (Kafka 표준).
3. **Versioning 정책** — additive-only (backward-compat) vs breaking (new event type).
4. **Delivery guarantee 명시** — at-most-once / at-least-once / exactly-once.
5. **Idempotency key source 명시** — event 마다.
6. **DLQ + retry policy** — dead letter 처리 + retry exponential backoff.
7. **Ordering 결정** — FIFO + partition key vs unordered.

## 5. 단계 (Phases)

### Phase 1. Event taxonomy + inventory

| Type | Definition | Use case |
|------|------------|----------|
| **Domain event** | 시스템 안의 사실 발생 ("UserRegistered", "PasswordChanged") | event-sourcing source-of-truth |
| **Integration event** | 외부 시스템과 통신 ("EmailEnqueued", "WebhookReceived") | sync-async bridge |
| **Command** | 의도 명령 ("SendEmail", "RevokeSession") | task queue, intent-driven |

본 SaaS auth example inventory:

| Event ID | Type | Producer | Consumer(s) |
|----------|------|----------|-------------|
| `EmailEnqueued` | integration | auth-api (signup/reset/resend) | email-worker |
| `EmailBounced` | integration (incoming) | SES → SNS → SQS | auth-api (mark verified=NULL) |
| `EmailComplaint` | integration (incoming) | SES → SNS → SQS | auth-api (suppression list) |
| `UserRegistered` | domain (future) | auth-api signup handler | analytics-worker, audit-writer |
| `PasswordChanged` | domain (future) | auth-api reset/change handler | audit-writer, security-monitor |
| `SessionRevoked` | domain (future) | auth-api revoke handler | audit-writer, real-time-notify |

총 6 event (현재 v1 = 3 implemented, future = 3 planned).

### Phase 2. Schema format + registry 결정

| Aspect | Decision |
|--------|----------|
| Schema format | **JSON Schema** (1차 — SQS / SNS json payload, low entry barrier). v2+ Protobuf 검토 (Kafka 도입 시). |
| Schema registry | **AWS Glue Schema Registry** (managed, free for SQS/SNS, version evolution support) |
| Schema validation | **producer-side** (auth-api Fastify validator with TypeBox + Glue) + **consumer-side** (email-worker validator + version check) |
| Code generation | TypeBox (TS) — interface 자동 생성 |

JSON Schema 예시:

```json
{
  "$id": "https://schema.example.com/EmailEnqueued/v1.json",
  "type": "object",
  "required": ["event_id", "occurred_at", "schema_version", "to_email_hash", "template_id", "tenant_id", "idempotency_key"],
  "properties": {
    "event_id": {"type": "string", "format": "uuid"},
    "occurred_at": {"type": "string", "format": "date-time"},
    "schema_version": {"type": "string", "const": "1.0"},
    "to_email_hash": {"type": "string", "description": "SHA-256 + tenant salt"},
    "to_email_lower": {"type": "string", "format": "email", "description": "PII — masked in logs"},
    "template_id": {"type": "string", "enum": ["signup-verify", "password-reset", "suspicious-login"]},
    "payload": {"type": "object", "additionalProperties": true},
    "tenant_id": {"type": "string", "format": "uuid"},
    "idempotency_key": {"type": "string", "description": "from request Idempotency-Key header"},
    "trace_id": {"type": "string", "description": "OTel trace id propagation"}
  }
}
```

### Phase 3. Versioning 정책

| Change type | Strategy |
|-------------|----------|
| **Additive (new optional field)** | bump minor (`1.0` → `1.1`), backward-compat. consumer 가 unknown field ignore. |
| **Field removal / type change** | bump major (`1.0` → `2.0`), breaking. dual-publish strategy: producer 가 v1 + v2 모두 publish, consumer 가 자기 version subscribe. v1 sunset 9 mo (api-contract 정책 정합). |
| **Semantic change (same field, different meaning)** | new event type 신설 (event id 변경) — 기존 v1 계속 |

producer 가 항상 `schema_version` 필드 포함, consumer 가 그 field 로 routing.

### Phase 4. Delivery guarantee + ordering + idempotency

| Event | Delivery | Ordering | Idempotency key | Retention |
|-------|----------|----------|-----------------|-----------|
| `EmailEnqueued` | at-least-once | FIFO per `to_email_hash` (per-recipient ordering) | `idempotency_key` (from request header) | 14d |
| `EmailBounced` | at-least-once | unordered | `event_id` (SES message id) | 14d |
| `EmailComplaint` | at-least-once | unordered | `event_id` | 14d |
| `UserRegistered` (future) | at-least-once | FIFO per `tenant_id` (per-tenant ordering) | `event_id` | 30d |
| `PasswordChanged` (future) | at-least-once | unordered (different from registration ordering) | `event_id` + `user_id` composite | 90d (compliance audit) |
| `SessionRevoked` (future) | at-least-once | unordered | `session_id` | 14d |

at-least-once 결정 근거: at-most-once = 중요 event loss risk, exactly-once = SQS 미지원 (Kafka 만 EOS 지원, ROI 부족) → at-least-once + consumer-side dedup.

### Phase 5. DLQ + retry + monitoring

| Layer | Policy |
|-------|--------|
| **Retry** | exponential backoff: 5s, 15s, 1min, 5min, 15min (max 5 retry, 30 min total) |
| **DLQ destination** | per-queue DLQ (예: `email-out-dlq`) — separate queue with 14d retention |
| **DLQ alert** | DLQ depth ≥ 1 → SEV3 paging (per setup-incident-paging matrix) |
| **DLQ replay** | manual review → fix root cause → redrive via SQS console / `aws sqs send-message-batch` |
| **Poison message** | DLQ 에서 발견 시 evidence 보존 (S3 archive) + 사용자 영향 분석 |
| **Monitoring** | CloudWatch metric per queue: depth, age-of-oldest-message, DLQ rate. alarm: depth > 100 / age > 1h / DLQ rate > 1% |

## 6. 산출물 형식 (Output format)

```markdown
## design-event-schema Output — <project name>

### Summary
<3 줄: event count / broker / schema format / delivery guarantee / DLQ count>

### Event Inventory
| Event ID | Type (domain/integration/command) | Producer | Consumer(s) |
|----------|------------------------------------|----------|-------------|
| ... | ... | ... | ... |

### Schema Decision
| Field | Value |
|-------|-------|
| Schema format | JSON Schema (1차) |
| Registry | AWS Glue Schema Registry |
| Validation | producer + consumer 양쪽 |
| Code generation | TypeBox (TS) |

### Schema Sample
\`\`\`json
<JSON Schema sample>
\`\`\`

### Versioning Policy
| Change type | Strategy |
|-------------|----------|
| Additive | minor bump (backward-compat) |
| Breaking | major bump, dual-publish 9 mo |

### Per-Event Contract
| Event | Delivery | Ordering | Idempotency key | Retention |
|-------|----------|----------|-----------------|-----------|
| ... | ... | ... | ... | ... |

### DLQ + Retry
| Aspect | Policy |
|--------|--------|
| Retry | exp backoff 5s/15s/1m/5m/15m (max 5) |
| DLQ | per-queue DLQ, 14d retention |
| DLQ alert | depth ≥ 1 → SEV3 |
| Replay | manual review + redrive |
| Poison | S3 archive + impact analysis |

### Cascade
- **§3 design-api-contract**: sync API 의 Idempotency-Key header → event idempotency_key field
- **§4 decompose-feature-to-actor-tracks**: producer/consumer = cross-actor contract source
- **§6 test-cross-actor-flow**: event flow E2E 검증 (LocalStack SES + elasticmq)
- **§7 setup-rollback-runbook**: poison message 처리 = rollback runbook 의 일부
- **§7 setup-incident-paging**: DLQ depth alert 매핑

### Next Step
<구체 action — 1줄: 예 "AWS Glue Schema Registry 셋업 + producer-side validator 통합 task 생성">
```

## 7. Cross-phase cascade

- **§3 design-api-contract**: sync ↔ async layer 통합
- **§4 decompose-feature-to-actor-tracks**: producer/consumer = cross-track contract
- **§6 test-cross-actor-flow**: async event E2E 검증
- **§7 setup-rollback-runbook**: poison message handling
- **§7 setup-incident-paging**: DLQ alert routing

## 8. 다음 skill (next in stage flow)

- `design-auth-model` (Cluster B) — 다음 SaaS pattern
- `design-tenant-model` (Cluster B)
- `write-adr` — event schema 결정 영속화

권장 chain (Cluster B 일괄):
```
/buddy:chain design-event-schema,design-auth-model,design-tenant-model,write-adr -- "<project>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `design-api-contract`** (§3) — 그것은 sync layer (REST/GraphQL/gRPC), 본 skill 은 async layer (event/queue/webhook). 두 layer 보완.
- **vs `derive-system-topology`** (§3) — 그것은 async-event edge enumeration (시각화), 본 skill 은 그 edge 의 schema + contract detail.
- **vs `setup-feature-flags`** (§7) — 그것은 runtime toggle, 본 skill 은 event schema. ops_email_disable kill switch 가 본 skill 의 event publish 차단.

## 10. 중요 규칙

- **Schema-first 의무** — schema-less JSON blob 거부.
- **Versioning 정책 의무** — additive vs breaking 결정.
- **DLQ 의무** — 모든 queue DLQ 보유.
- **Idempotency key 명시** — at-least-once 의무 (consumer-side dedup).
- **Producer + Consumer 양방향 contract** — single-side 거부.
- **Read-only on production** — 본 skill 은 design 산출, infra 변경 안 함.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 8 출력 섹션 모두 채워짐
- [ ] Event Inventory 의 모든 event 가 type / producer / consumer 명시
- [ ] Schema Decision 의 format / registry / validation / codegen 모두 결정
- [ ] Schema Sample ≥ 1 (JSON Schema or Protobuf or Avro)
- [ ] Versioning Policy additive + breaking 둘 다 정의
- [ ] Per-Event Contract 의 모든 event 에 4 attribute (delivery / ordering / idempotency / retention)
- [ ] DLQ + Retry 5 aspect 모두 정의
- [ ] §4 posture 적용 — 단호
- [ ] §0 anti-pattern 부재 — schema-less / versioning 부재 / DLQ 부재 / idempotency 없음 / single-side / sync 분리 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
