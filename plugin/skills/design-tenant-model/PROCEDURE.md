# design-tenant-model — multi-tenant 격리 전략 결정

§3 Technical Design phase 의 stage. **multi-tenant SaaS 의 핵심 design decision**. shared row (RLS) vs schema-per-tenant vs DB-per-tenant 3 핵심 모델 + tenant identity propagation + cross-tenant query prevention + onboarding/offboarding cost + scaling implication. 산출물은 model selection + isolation matrix + migration plan + ADR handoff.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Model 결정 deferred** — "나중에 결정" → 첫 enterprise 고객에서 강제 마이그레이션 (수개월 비용).
- **Cross-tenant query 방어 layer 단일** — application-level check 만 → 버그 1 개로 cross-tenant data leak. DB-level RLS + app-level + audit log 3 layer 의무.
- **Tenant id propagation 모호** — JWT claim / header / cookie 중 어느 source 인지 불명. 매 endpoint 마다 다른 source 사용 시 inconsistency.
- **Onboarding cost 무측정** — 신규 tenant 1 개 추가 비용 (provisioning time + storage + monthly $) 정량 부재.
- **Per-tenant customization 결정 deferred** — feature flag / branding / billing tier 의 tenant-scoped vs global 결정 부재.
- **Compliance scope 모호** — SOC 2 / HIPAA / GDPR 의 tenant boundary 가 어디까지인지 불명.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

multi-tenant 격리 전략을 v1 부터 design — 후 도입 비용 큰 결정. 3 모델 trade-off 명시 + selected model 의 isolation guarantee + migration path + scaling cost.

## 2. 사용 시점 (When to invoke)

- §3 design 단계, design-data-model 후 (tenant boundary 가 schema 영향)
- B2B SaaS 시작 시 (B2C 는 single-tenant 또는 simpler)
- enterprise 고객 onboarding 직전 (data isolation 의무 확인)
- compliance audit 진입 (HIPAA / SOC 2 / GDPR — tenant boundary 명시)
- multi-region 결정과 결합 (region-별 tenant 분리)

## 3. 입력 (Inputs)

### 필수
- §3 define-tech-stack ADR (DB / compute 결정)
- §3 design-data-model 산출 (entity + RLS plan)
- §2 feature spec (tenant 의 정의 — company / org / individual)
- 예상 tenant 수 (12-mo / 3-yr)
- per-tenant 평균 user 수 / 평균 data volume

### 선택
- enterprise customer 요구 (data residency / dedicated infra / custom domain)
- compliance scope (HIPAA BAA per tenant)
- existing single-tenant migration 필요 여부

### 입력이 부족할 때 forcing question
- "tenant 정의가 명확한가? 'company' / 'individual' / 'workspace' — entity 단위 명시 없으면 본 skill 무근거."
- "예상 tenant 수가 < 100 인가, > 10k 인가? 차원이 다른 결정 (DB-per-tenant 가능 vs 강제 shared)."
- "enterprise 고객의 dedicated infra 요구 가능성? 'eventually' 답이면 hybrid model design 부담 cover."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — 3 모델 중 1 단호 선택.
- **사용자 입력을 challenge** — "tenant 격리하면 됨" 발화에 "RLS + schema-per + DB-per 어느 layer? application + DB 양쪽?" push.
- **Specificity 강제** — vague "isolated" 거부, "Postgres RLS policy on 5 table + JWT tenant_id claim + audit log 3 layer".
- **Trade-off 명시** — onboarding cost vs isolation guarantee + scaling cost 모두 정량.

도메인 원칙:

1. **3 모델 trade-off 분석 의무** — shared / schema-per / DB-per 모두 검토.
2. **Defense in depth 3 layer** — DB (RLS / schema) + App (JWT claim check) + Audit (cross-tenant access 0 verify).
3. **Tenant identity SSoT** — JWT claim 권장 (HTTP header 는 forgeable).
4. **Onboarding / offboarding cost 정량** — provisioning time + storage cost + monthly $.
5. **Per-tenant customization scope 명시** — feature flag tier / billing / branding / config.
6. **Compliance scope 명시** — HIPAA BAA / SOC 2 audit / GDPR DPO per tenant 범위.

## 5. 단계 (Phases)

### Phase 1. 3 모델 trade-off 분석

> **AI 편향 차단 게이트** — 본 Phase는 Shared (RLS) / Schema-per / DB-per 3 모델을 *기본 후보*로 비교하지만, 도메인에 따라 *추가 orthogonal 옵션*(예: 지역별 격리 / namespace-per-tenant in k8s / hybrid — premium tier만 DB-per) 검토 필요. 직접 또는 `verify-best-alternative` 호출. 3 기본 모델이 *유일한 alternatives*라고 가정하지 말 것 — 추가 후보가 없는지 확인 후 진행. 적용 근거: [`docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md`](../../../../docs/superpowers/specs/2026-05-21-engineering-decision-gate-mapping.md) §3.

| Aspect | Shared (RLS) | Schema-per-tenant | DB-per-tenant |
|--------|--------------|---------------------|------------------|
| Isolation | logical (RLS policy) | strong (schema namespace) | strongest (separate DB instance) |
| Onboarding cost | minimal (1 INSERT tenants row) | medium (CREATE SCHEMA + migrate) | high (RDS provision 5-30 min) |
| Per-tenant migration | 단일 migration 모든 tenant | per-schema migration coordination | 가장 flexible (per-tenant timing) |
| Storage cost | shared, very efficient | separate schema, ~5% overhead | separate DB, instance min cost ($X/mo per tenant) |
| Backup granularity | shared (per-DB) — selective restore 어려움 | per-schema (`pg_dump --schema=`) | per-DB (clean) |
| Cross-tenant analytics | easy (single SELECT) | medium (UNION ALL across schema) | hard (cross-DB query) |
| Compliance | RLS bug = data leak risk | schema isolation reduce risk | strongest, separate audit per DB |
| Max tenant count | 100k+ (RLS scales) | ~1k (Postgres schema limit) | ~100 (RDS instance limit per region without sharding) |
| Cross-tenant migration / consolidation | application-level data move | schema rename + copy | data export/import |

### Phase 2. Selected model + rationale

본 SaaS auth example (5-yr horizon: 100k tenants, B2B, mostly SMB + occasional Enterprise):

| Selected | Rationale |
|----------|-----------|
| **Shared (RLS)** primary, **DB-per-tenant** for Enterprise tier opt-in | (a) 100k tenant scale 만족 (b) onboarding cost minimal (c) Enterprise 고객 dedicated infra 옵션 추가 — hybrid cost 작음 |

| Trigger to re-evaluate | tenant count > 100k OR per-tenant data > 10GB average OR compliance audit demands strict schema-per-tenant OR Enterprise demand dedicated infra > 30% conversion |

### Phase 3. Defense in depth — 3 layer isolation

| Layer | Mechanism | Implementation |
|-------|-----------|------------------|
| **DB layer (primary)** | Postgres RLS policy on 5 table (users, sessions, verification_tokens, password_reset_tokens, audit_log) | per design-data-model. `CREATE POLICY ... USING (tenant_id = current_setting('app.tenant_id'))`. session-scoped `SET app.tenant_id = '...'` on connection. |
| **App layer (secondary)** | Fastify hook reads JWT claim `tenant_id`, sets DB session variable. all queries via Drizzle ORM with `where(eq(table.tenantId, ctx.tenantId))` in addition to RLS. | belt-and-suspenders. RLS bug 회피 + audit log explicit. |
| **Audit layer (tertiary)** | nightly pgTAP test: synthetic Tenant A connection → SELECT all tables → expected 0 rows from Tenant B. CloudWatch metric: `cross_tenant_select_count`. alarm: > 0 → SEV1. | continuous verification. |

JWT claim format:
```json
{
  "iss": "auth.example.com",
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "roles": ["TenantAdmin"],
  "exp": 1234567890
}
```

backend handler 패턴:
```typescript
fastify.addHook('preHandler', async (req) => {
  const claims = await verifyJWT(req.headers.authorization);
  req.tenantId = claims.tenant_id;
  // Drizzle: set DB session variable
  await db.execute(sql`SELECT set_config('app.tenant_id', ${claims.tenant_id}, true)`);
});
```

### Phase 4. Onboarding / offboarding cost + per-tenant customization

#### Onboarding cost (shared model)

| Step | Time | Cost |
|------|------|------|
| INSERT tenants row | < 100ms | minimal |
| Provision LaunchDarkly tenant context | < 5s (API call) | $0 (LD free for small flag count) |
| Setup Stripe customer | < 5s | $0 (Stripe customer free) |
| Send welcome email | async via SQS | $0.0001 (SES) |
| **Total per-tenant onboarding** | **~10s wall-clock** | **<$0.001** |

#### Offboarding cost (per GDPR right to erasure)

| Step | Time |
|------|------|
| `users.deleted_at` set + `email_hash` 보존 | <100ms |
| All sessions revoke | <100ms |
| `audit_log.user_id` NULL set (preserve audit per regulation) | per partition retention drop |
| Stripe customer delete | <5s (API) |
| LaunchDarkly tenant context delete | <5s |
| **Total** | **~10s wall-clock**, retention partition drop async |

#### Per-tenant customization scope

| Customization | Scope | Implementation |
|----------------|-------|------------------|
| Feature flag (release / experiment) | per-tenant via LaunchDarkly target rule | `setup-feature-flags` integration |
| Branding (logo, color) | per-tenant `tenants.branding` JSONB | minimal — Pro+ tier |
| Billing tier (Free / Pro / Enterprise) | `tenants.plan` enum + Stripe subscription | core |
| Custom domain (Enterprise) | per-tenant CNAME + ACM cert | Enterprise tier opt-in (CloudFront alternate domain) |
| Data residency (EU / US) | Enterprise tier — separate region deployment | hybrid model trigger (DB-per-tenant Enterprise 부분 활용) |
| SSO (SAML) | per-tenant WorkOS connection | per design-auth-model Enterprise tier |
| Audit log retention (default 18mo, Enterprise 7y) | `tenants.audit_retention_months` | retention job per-tenant |

### Phase 5. Compliance scope per model

| Compliance | Shared model boundary | Enterprise (DB-per) boundary |
|------------|------------------------|--------------------------------|
| SOC 2 | shared infra, per-tenant audit_log RLS-isolated | separate audit per DB instance |
| HIPAA BAA | shared infra accept BAA scope (separate BAA per HIPAA tenant) | per-DB isolation strong, BAA per instance |
| GDPR | EU tenant 의 data 가 US region 에 저장 가능 (Standard Contractual Clauses 의무) | EU-only deployment 가능 (data residency) |
| PCI-DSS | not applicable (auth service, no card data) | n/a |

### ADR Handoff

다음 단계:
```bash
/buddy:write-adr "Multi-tenant model — Shared (RLS) primary + DB-per-tenant Enterprise opt-in"
```

또는 chain:
```bash
/buddy:chain design-tenant-model,write-adr -- "<project>"
```

## 6. 산출물 형식 (Output format)

```markdown
## design-tenant-model Output — <project name>

### Summary
<3 줄: selected model / 3 layer defense / max tenant scale / hybrid Enterprise opt-in / compliance>

### 3-Model Comparison
| Aspect | Shared (RLS) | Schema-per | DB-per |
|--------|--------------|------------|--------|
| ... | ... | ... | ... |

### Selected Model + Rationale
| Field | Value |
|-------|-------|
| Primary | Shared (RLS) |
| Enterprise opt-in | DB-per-tenant |
| Rationale | 100k tenant scale + minimal onboarding + Enterprise dedicated infra option |
| Re-evaluation trigger | tenant count > 100k OR per-tenant > 10GB OR audit demands schema-per OR Enterprise dedicated > 30% conversion |

### 3-Layer Defense
| Layer | Mechanism | Implementation |
|-------|-----------|------------------|
| DB | Postgres RLS on 5 table | per design-data-model |
| App | Fastify JWT hook + Drizzle `where(tenantId)` | belt-and-suspenders |
| Audit | nightly pgTAP cross-tenant test | CW metric + SEV1 alarm |

### JWT Claim Format
\`\`\`json
{ "tenant_id": "...", "roles": [...], ... }
\`\`\`

### Onboarding / Offboarding
| Operation | Time | Cost |
|-----------|------|------|
| Onboard tenant (shared) | ~10s | <$0.001 |
| Onboard Enterprise (DB-per) | ~30 min (RDS provision) | $X/mo per instance |
| Offboard (GDPR erasure) | ~10s + retention partition drop async | minimal |

### Per-Tenant Customization
| Customization | Scope | Tier |
|----------------|-------|------|
| Feature flag | per-tenant LD target | all |
| Branding | tenants.branding JSONB | Pro+ |
| Custom domain | CloudFront alt-domain | Enterprise |
| Data residency (EU) | DB-per Enterprise | Enterprise |
| SSO SAML | WorkOS per-tenant | Enterprise |

### Compliance Scope
| Compliance | Shared boundary | Enterprise boundary |
|------------|-----------------|----------------------|
| SOC 2 | RLS-isolated audit | per-DB audit |
| HIPAA | shared BAA | per-instance BAA |
| GDPR | SCC required | EU-only deployment |

### ADR Handoff
다음 단계: `/buddy:write-adr "Multi-tenant model — Shared (RLS) + Enterprise DB-per opt-in"`

### Cascade
- **§3 design-data-model**: RLS policy 5 table 정합 (이미 done)
- **§3 design-auth-model**: JWT tenant_id claim + role hierarchy 정합
- **§3 design-api-contract**: per-tenant rate-limit + audit endpoint scope
- **§7 prepare-launch-checklist**: tenant isolation row 입력
- **§9 manage-lifecycle** (deferred): Enterprise tenant migration / archive

### Next Step
<구체 action — 1줄: 예 "RLS pgTAP nightly test 추가 + Enterprise hybrid model 의 DB provisioning Terraform module">
```

## 7. Cross-phase cascade

- **§3 design-data-model**: RLS 정합 (확인 의무)
- **§3 design-auth-model**: JWT tenant_id claim 정합
- **§3 design-api-contract**: per-tenant scope
- **§7 prepare-launch-checklist**: tenant isolation evidence
- **§9 manage-lifecycle**: Enterprise migration / archive

## 8. 다음 skill (next in stage flow)

- `write-adr` — model 영속화
- `prepare-launch-checklist` — readiness gate

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `design-data-model`** (§3) — 그것은 schema + index, 본 skill 은 tenant boundary. RLS 가 두 skill 모두 cover (data-model 이 어느 column 으로 RLS, tenant-model 이 RLS 가 적합한 격리 layer 인지).
- **vs `design-auth-model`** (§3) — 그것은 user identity, 본 skill 은 tenant identity. JWT claim 합쳐 통합.
- **vs `setup-feature-flags`** (§7) — 그것은 flag 자체, 본 skill 은 per-tenant flag scope.

## 10. 중요 규칙

- **3 모델 분석 의무** — single model 결정 거부.
- **Defense in depth 3 layer** — single layer 거부.
- **Tenant identity SSoT** — JWT claim 권장.
- **Onboarding / offboarding 정량** — time + cost 명시.
- **Compliance scope 명시** — SOC 2 / HIPAA / GDPR boundary.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 9 출력 섹션 모두 채워짐
- [ ] 3 model trade-off table (shared / schema-per / DB-per) 9 aspect 모두 비교
- [ ] Selected model 단일 + rationale + re-evaluation trigger
- [ ] 3 layer defense (DB / App / Audit) 모두 mechanism + implementation
- [ ] JWT claim format 명시
- [ ] Onboarding / offboarding time + cost 정량
- [ ] Per-tenant customization 5+ category
- [ ] Compliance scope (SOC 2 / HIPAA / GDPR) 명시
- [ ] §4 posture
- [ ] §0 anti-pattern 부재 — model 결정 / single layer / identity 모호 / cost 무측정 / customization 미정 / compliance 모호 모두 충족
- [ ] **`verify-best-alternative` 1회 이상 호출 완료** — AI 편향 방지 의무. tenant isolation model(shared/schema-per/DB-per) 결정이 *첫 답*이 아니라 다관점 검토 후 최선임을 확인
- [ ] **3+ orthogonal tenant isolation 후보의 rubric 비교 표가 산출물에 존재** — 체크박스만 체크하는 *anti-rationalization 회피* 금지. 기본 3 모델 + 추가 옵션(hybrid/지역별/namespace-per-tenant) 검토 여부 명시

하나라도 no 면 해당 phase 회귀 후 재검증.
