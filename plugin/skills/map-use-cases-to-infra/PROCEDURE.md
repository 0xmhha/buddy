# map-use-cases-to-infra — actor system boundary → infra component 매핑

§3 Technical Design phase 의 cascade bridge stage. **Q8=(a) cascade 의 §2 → §3 transition 의 explicit layer**. §2 feature spec 의 actor × use case × system boundary 를 §3 의 actual infra component (compute / DB / cache / queue / CDN / observability / IAM) 로 매핑. silent gap 채움 — 이 layer 없으면 §3 design-system 이 actor model 과 disconnect 된 채 진행. 산출물은 actor → infra component bidirectional 매트릭스 + 책임 분담 + cross-actor infra dependency.


## Input Requirements

| Input | Required | Type | Source | 미제공 시 |
|-------|----------|------|--------|----------|
| Use case map | ✅ | artifact | `map-actor-use-cases` 산출물 | 먼저 `/buddy:map-actor-use-cases` 를 실행하세요 |
| System boundary map | ✅ | artifact | `map-use-case-to-system-boundary` 산출물 | 먼저 `/buddy:map-use-case-to-system-boundary` 를 실행하세요 |

## Output Contract

| Output | Type | Format | Consumers |
|--------|------|--------|-----------|
| Infra mapping (use case별 실행 인프라 매핑) | artifact | structured YAML | `derive-system-topology` |

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지한다. 발견 시 §5 회귀:

- **Actor → infra 모호 매핑** — "user actor 는 frontend" 로 끝. 어느 infra component (compute / CDN / DB) 가 어느 actor 의 use case 를 cover 하는지 1:N / N:1 명시 필요.
- **Cross-actor infra 누락** — actor A 의 use case 가 actor B 의 infra (예: shared DB / queue) 에 의존하는데 명시 안 됨 → §4 actor track 분해 시 cross-track edge 누락.
- **Use case 단위 누락** — actor 단위만 매핑하고 use case 단위 깊이 부재 → 같은 actor 의 다른 use case 가 다른 infra 사용 (예: signup 은 SQS, login 은 cache) silent.
- **Define-tech-stack 결정 무시** — 본 skill 이 §3 design 의 첫 번째 layer 인데 define-tech-stack ADR-0001 결과 (Postgres / Fargate / SQS) 를 입력으로 안 받으면 actor 매핑이 stack-agnostic 추상화 (가치 낮음).
- **Bidirectional matrix 부재** — actor → infra 만 보고 infra → actor (역방향) 안 보면 어느 infra 가 단일 actor 위해 존재하는지 (over-engineering risk) 식별 불가.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

§2 feature spec (actor / use case / system boundary) 와 §3 define-tech-stack ADR (Postgres / Fargate / SQS / etc.) 사이의 mapping layer 채움. Q8=(a) cascade 의 §3 진입 시 actor model 정합성 유지 — 이후 design-data-model / design-api-contract / decompose-feature-to-actor-tracks 가 본 매핑을 입력으로 사용.

## 2. 사용 시점 (When to invoke)

- §3 design-system 진입 직후, define-tech-stack 결정 후 (chain 권장)
- 신규 actor 추가 / use case 추가 시 매핑 갱신
- infra component 추가 / 변경 (예: 신규 cache layer 도입) 시
- §4 plan-build 의 decompose-feature-to-actor-tracks 가 actor track 분해 시 cross-actor edge 가 정합 안 될 때 → 본 skill 회귀 trigger

## 3. 입력 (Inputs)

### 필수
- §2 feature spec — actor list + per-actor use case + system boundary
- §3 `define-tech-stack` ADR — selected infra component (compute / DB / cache / queue / CDN / observability / IAM / CI/CD)
- (해당 시) §3 `derive-system-topology` 산출물 — actor 그래프 + 시스템 토폴로지

### 선택
- 기존 brownfield infra (마이그레이션 시 매핑 강제 조건)
- 산업 규제 의무 (data residency, encryption-at-rest scope, audit log boundary)
- multi-region 결정 (region-별 infra split 매핑 필요 시)

### 입력이 부족할 때 forcing question
- "define-tech-stack ADR 가 확정됐나? 미확정 시 본 skill 의 mapping 이 추상 — 가치 낮음. ADR 먼저."
- "actor list 가 §2 에서 enumerate 됐나? 빠진 actor (예: ops / SRE / auditor) 발견 시 본 skill 결과 무효."
- "cross-actor shared infra (예: shared DB / message bus) 의 ownership 결정 됐나? unclear ownership = §4 plan 의 cross-track contract 무근거."

## 4. 핵심 원칙 (Principles + Posture)

운영 posture:

- **입장 취함, hedge 금지** — actor → infra 매핑 단호. "frontend 는 CDN 이 어느정도 필요" 거부, "User actor 의 7 use case (signup / login / verify / reset / me / logout / GDPR) 모두 CDN + ALB + Fargate(API) + Postgres + Redis(session) + SQS(email enqueue) 6 component 사용" 명시.
- **사용자 입력을 challenge** — "actor 3 개" 발화에 "ops actor / SRE / auditor 는? compliance 영역 actor 누락 자주 발생" push.
- **Specificity 강제** — vague "uses DB" 거부, "users table SELECT (login) + INSERT (signup) + UPDATE (me edit)" 같이 use case × CRUD 단위.
- **Bidirectional check 의무** — actor → infra + infra → actor 양방향, single-actor infra 발견 시 over-engineering flag.

도메인 원칙:

1. **Actor × use case × infra component 3D 매트릭스** — 모든 use case 가 어느 infra component 사용하는지, 누락 0.
2. **Stack-aware** — define-tech-stack ADR 의 actual component 이름 사용 (Postgres / Fargate / SQS) — generic ("DB / compute / queue") 거부.
3. **Bidirectional matrix** — actor → infra + infra → actor 양방향. single-actor infra 발견 시 over-engineering flag.
4. **Cross-actor infra dependency** — shared component (DB / queue / cache) 의 ownership + 접근 패턴 명시 (RW vs R-only).
5. **Compliance scope mapping** — encryption-at-rest / audit log / RLS 영역이 어느 actor × infra 조합에 적용되는지 명시.

## 5. 단계 (Phases)

### Phase 1. Actor × use case enumeration

§2 feature spec 의 actor list × per-actor use case 모두 enumerate:

| Actor | Use case | Source (feature spec) |
|-------|----------|------------------------|
| User (end-user) | signup | feature spec UC-001 |
| User | login | UC-002 |
| User | verify-email | UC-003 |
| User | password-reset | UC-004 |
| User | profile-edit | UC-005 |
| User | gdpr-erasure | UC-006 |
| Tenant Admin | view-audit-log | UC-007 |
| Tenant Admin | manage-users | UC-008 |
| 3rd-party (SES) | send-email | infra interaction |
| 3rd-party (SES) | bounce-handle | infra interaction |
| Ops / SRE | incident-response | (often missing — explicit add) |
| Auditor (compliance) | review-audit-log | (often missing — explicit add) |

orphan actor / orphan use case = §2 feature spec 보강 회귀 trigger.

### Phase 2. Infra component enumeration

§3 define-tech-stack ADR 의 actual component:

| Component | Specific (ADR-0001) | Purpose |
|-----------|---------------------|---------|
| Compute | AWS Fargate ECS | API + email worker |
| Frontend hosting | CloudFront + S3 | Next.js 14 static + SSR |
| DB | Postgres 16 RDS | OLTP transactional |
| Cache | Redis 7 ElastiCache | session + rate-limit |
| Queue | AWS SQS (FIFO + standard) | email + audit event |
| Email service | AWS SES | transactional email |
| Auth (intra) | JWT + argon2id (in-process) | token + password |
| CDN | CloudFront | static assets |
| Observability | OpenTelemetry → X-Ray + CloudWatch | trace + metric + log |
| IAM | AWS IAM + Secrets Manager | per-task least-priv + secret rotation |
| CI/CD | GitHub Actions w/ OIDC | build + deploy |
| Feature flag | LaunchDarkly Pro | release/experiment/ops/permission flag |
| Paging | PagerDuty Pro | on-call alert routing |

### Phase 3. Forward matrix — actor × use case → infra components

| Actor | Use case | Infra component | Access pattern |
|-------|----------|-----------------|----------------|
| User | signup | CloudFront + Fargate API + Postgres (W) + SQS (W) + SES (out) + Redis (idem-cache W) + JWT (out) | end-to-end (W-heavy) |
| User | login | CloudFront + Fargate API + Postgres (R, lazy rehash W) + Redis (rate-limit R+W, session W) + JWT (out) | end-to-end (R-heavy) |
| User | verify-email | CloudFront + Fargate API + Postgres (token consume W + users update W) + JWT (out) | shorter chain |
| User | password-reset | CloudFront + Fargate API + Postgres (W) + SQS (W) + SES (out) + Redis (rate-limit) + JWT (out, post-confirm) | similar to signup |
| User | profile-edit | CloudFront + Fargate API + Postgres (W) + JWT (in) | API + DB |
| User | gdpr-erasure | Fargate API + Postgres (audit user_id NULL set, sessions revoke) + JWT (in, re-auth) | compliance scope |
| Tenant Admin | view-audit-log | Fargate API + Postgres (R, RLS-scoped) + JWT (in, role=admin) | RLS-critical |
| Tenant Admin | manage-users | Fargate API + Postgres (W, RLS-scoped) + JWT (in) | RLS + admin scope |
| 3rd-party (SES outgoing) | send-email | SQS (R, consumer) + SES (out) + Postgres (email_send_log W) | async event |
| 3rd-party (SES incoming) | bounce-handle | SNS + SQS (W) + Postgres (users.email_verified W) | async event |
| Ops / SRE | incident-response | CloudWatch (R) + PagerDuty + LaunchDarkly (kill switch W) + AWS Console (audit) | observability + control plane |
| Auditor (compliance) | review-audit-log | Postgres audit_log (R, time-range scan) + CloudWatch logs (R, retention) | compliance read-only |

### Phase 4. Reverse matrix — infra component → actor

bidirectional check — single-actor infra 발견 시 over-engineering flag:

| Infra | Actor coverage | Single-actor flag |
|-------|----------------|---------------------|
| CloudFront + S3 | User (모든 use case) | n/a (User-only by design — frontend) |
| Fargate ECS API | User + Tenant Admin + 3rd-party (consumer task) | shared (correct) |
| Postgres 16 RDS | User (R+W) + Tenant Admin (R+W RLS) + 3rd-party (W via consumer) + Auditor (R) | shared (correct) |
| Redis ElastiCache | User (session R+W + rate-limit R+W) | **single-actor flag — User only**. Tenant Admin / 3rd-party 사용 안 함. acceptable (session/rate-limit 본질적 user-scope) — over-engineering 아님 |
| SQS | User (signup/reset producer) + 3rd-party (consumer) | shared (correct) |
| SES | 3rd-party (out direct) | **single-actor flag** — but SES 자체가 3rd-party actor 의 일부 — by design |
| JWT (in-process) | User + Tenant Admin (auth header in/out) | shared |
| LaunchDarkly | Ops/SRE (toggle) + User (flag-evaluated) | shared (control plane + runtime) |
| PagerDuty | Ops/SRE only | **single-actor — by design** (SRE plane) |
| Auditor 의 audit_log read | Auditor only | **single-actor — by design** (compliance plane) |

flag 결과: 4 single-actor infra 모두 "by design" 정당화. over-engineering 0.

### Phase 5. Cross-actor shared infra ownership + dependency

| Shared infra | Actor list | Ownership | Cross-actor edge implication (§4 input) |
|--------------|------------|-----------|-------------------------------------------|
| Postgres (users / sessions / verification_tokens / password_reset_tokens / audit_log) | User + Tenant Admin + 3rd-party (consumer W) + Auditor | **DBA / platform team** | RLS policy (T3.3) 가 cross-actor isolation 강제 — User actor 가 다른 tenant 의 row 못 읽음 |
| SQS (email-out) | User (producer via Fargate) + 3rd-party (consumer) | platform team | EmailEnqueued event schema 가 cross-actor contract — `design-event-schema` skill (Cluster B) 보강 후보 |
| Redis (session) | User only | platform team | n/a (no cross-actor) |
| LaunchDarkly | Ops/SRE (governance) + User (runtime evaluation) | platform team + Ops | ops_* kill switch 가 User-facing 영향 — incident response cross-actor 활성 |
| audit_log table | User (W via Backend) + Tenant Admin (R) + Auditor (R-only export) | DBA + compliance | RLS + retention policy — compliance scope cross-actor |

compliance scope mapping:
- **Encryption-at-rest**: Postgres + S3 + ElastiCache 모두 KMS encryption (User + Tenant Admin + Auditor 영향)
- **RLS**: Postgres 의 5 tenant-scoped table (users / sessions / verification_tokens / password_reset_tokens / audit_log) — User + Tenant Admin + Auditor 의 cross-tenant 접근 차단
- **Audit log retention**: 18 mo (auditor 요구) — `audit_log` partitioned, retention policy 자동 drop
- **GDPR erasure scope**: User actor 의 `users` row + `audit_log.user_id` NULL set, `email_hash` 보존 — Auditor 가 erasure history 추적 가능

## 6. 산출물 형식 (Output format)

> structured 출력 강제, prose 변환 금지.

```markdown
## map-use-cases-to-infra Output — <project name>

### Summary
<3 줄: actor count / use case count / infra component count / cross-actor shared count / over-engineering flag>

### Actor × Use Case
| Actor | Use case | Source |
|-------|----------|--------|
| ... | ... | ... |

### Infra Component (from define-tech-stack ADR)
| Component | Specific | Purpose |
|-----------|----------|---------|
| ... | ... | ... |

### Forward Matrix (Actor × Use Case → Infra)
| Actor | Use case | Infra components | Access pattern |
|-------|----------|------------------|----------------|
| ... | ... | ... | ... |

### Reverse Matrix (Infra → Actor)
| Infra | Actor coverage | Single-actor flag |
|-------|----------------|---------------------|
| ... | ... | ... |

### Cross-Actor Shared Infra
| Shared infra | Actor list | Ownership | Cross-actor edge implication (§4 input) |
|--------------|------------|-----------|-------------------------------------------|
| ... | ... | ... | ... |

### Compliance Scope Mapping
| Compliance area | Infra scope | Actor scope |
|-----------------|-------------|-------------|
| Encryption-at-rest | ... | ... |
| RLS | ... | ... |
| Audit retention | ... | ... |
| GDPR erasure | ... | ... |

### Cascade
- **§3 design-data-model**: shared DB scope + RLS 정합 입력
- **§3 design-api-contract**: actor × use case 가 endpoint 매핑 입력
- **§3 derive-system-topology**: actor 그래프 + infra 매핑이 토폴로지 자동 도출 입력 (또는 본 skill 의 입력)
- **§4 decompose-feature-to-actor-tracks**: shared infra ownership 이 cross-track contract 의 source

### Next Step
<구체 action — 1줄: 예 "derive-system-topology 호출 또는 design-data-model 진입">
```

## 7. Cross-phase cascade

- **§3 design-data-model**: shared DB scope + RLS 정합
- **§3 design-api-contract**: actor × use case → endpoint
- **§3 derive-system-topology**: actor 그래프 + infra 매핑 → 토폴로지
- **§4 decompose-feature-to-actor-tracks**: shared infra ownership → cross-track contract
- **§6 test-cross-actor-flow**: cross-actor shared infra → cross-actor flow E2E coverage

## 8. 다음 skill (next in stage flow)

- `derive-system-topology` — 본 skill 의 산출물 위에서 actor 그래프 자동 도출
- `design-data-model` — shared DB scope + RLS 정합
- `design-api-contract` — actor × use case → endpoint

권장 chain (Q8=(a) cascade §3):
```
/buddy:chain define-tech-stack,map-use-cases-to-infra,derive-system-topology,design-data-model,design-api-contract,write-adr -- "<project>"
```

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `define-tech-stack`** (§3) — 그것은 어떤 component (Postgres vs DynamoDB) 결정, 본 skill 은 결정된 component 가 어느 actor × use case 를 cover. 본 skill 이 후속.
- **vs `derive-system-topology`** (§3) — 본 skill 은 actor × use case → infra mapping (의도), 다음 skill 은 actor 그래프 → topology 자동 도출 (구조). 두 skill 이 짝.
- **vs `design-data-model`** (§3) — 본 skill 은 shared DB scope (어느 actor 가 어느 table 사용), design-data-model 은 그 위의 schema 설계. 본 skill 이 먼저.
- **vs `decompose-feature-to-actor-tracks`** (§4) — 본 skill 은 §3 mapping (어떤 infra), 그것은 §4 actor track (누가 implement). 본 skill 이 §3, 후속이 §4.

## 10. 중요 규칙

- **모든 use case 매핑 의무** — orphan use case = §2 회귀.
- **Stack-aware** — generic component 명 거부, define-tech-stack ADR 의 specific 명 사용.
- **Bidirectional matrix 의무** — single-actor infra flag 검증.
- **Cross-actor shared ownership 명시** — DBA / platform / Ops 중 책임자.
- **Compliance scope mapping 의무** — encryption / RLS / retention / GDPR 4 영역.
- **Read-only on production code** — 본 skill 은 mapping 산출, infra 변경 안 함.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 8 출력 섹션 (Summary / Actor × Use Case / Infra / Forward Matrix / Reverse Matrix / Cross-Actor Shared / Compliance Scope / Cascade) 모두 채워짐
- [ ] Actor × Use Case 가 §2 feature spec 의 모든 actor + use case 매핑 (orphan 0, ops/auditor 누락 검증)
- [ ] Infra Component 가 define-tech-stack ADR 의 specific 명 사용 (generic 거부)
- [ ] Forward Matrix 의 모든 use case 가 infra components + access pattern 명시
- [ ] Reverse Matrix 의 single-actor infra 모두 "by design" 정당화 또는 over-engineering flag
- [ ] Cross-Actor Shared Infra 의 ownership + cross-actor edge implication 명시
- [ ] Compliance Scope Mapping 4 영역 (encryption / RLS / retention / GDPR) 모두 명시
- [ ] §4 posture 적용 — 매핑 단호, "어느 정도" hedge 없음
- [ ] §0 anti-pattern 부재 — 모호 매핑 없음 / cross-actor 명시 / use case 단위 깊이 / stack-aware / bidirectional 모두 충족

하나라도 no 면 해당 phase 회귀 후 재검증.
