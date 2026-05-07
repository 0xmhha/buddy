# design-api-contract — REST/GraphQL/RPC 계약 + actor 경계

API 계약은 actor 간 / system 간 경계를 명시화한 산출물이다. 잘못된 contract 는 client 영향 + versioning 부담 + 호환성 문제로 운영 비용을 증폭시킨다. 본 skill 은 **API style 선택 (REST/GraphQL/gRPC) + resource ↔ use case 매핑 + schema 정의 + error taxonomy + versioning 정책 + contract test 전략** 을 강제해 contract-first 를 보장한다. tech stack + data model 결정 후 호출된다.

## 0. STOP — 시작 전 읽기

이 skill 은 다음 anti-pattern 들을 방지하기 위한 절차다. 산출물에 다음이 발견되면 검증 실패로 §5 로 돌아가 보강한다:

- **Style 선택 무근거** — REST 인지 GraphQL 인지 그냥 "익숙해서" 결정. use case 적합도 매트릭스 누락.
- **Resource = DB table 1:1 매핑** — API resource 와 DB entity 가 자동 동일하다고 가정. aggregation / projection / view 검토 누락.
- **Error taxonomy 부재** — 200/4xx/5xx 만 정하고 도메인 에러 분류 안 함. client 가 매번 error message 파싱 강요.
- **Versioning 정책 미결정** — "나중에 v2 만들면 되지" 식. breaking change 정의·deprecation 일정·sunset 정책 누락.
- **Contract test 전략 없음** — schema validation 만으로 끝내고 consumer-driven contract test 누락. client 깨질 때 알 수 없음.
- **Auth / rate-limiting 무규명** — endpoint 단위가 아니라 system 단위로만 정의. resource 별 권한 결정 누수.

§5 의 모든 phase 누락 없이 수행하라.

## 1. 목적

API 계약 의사결정 누수 — "REST 가 표준이니까" / "schema 는 코드에서" / "에러는 그냥 500 던지면 돼" — 를 evidence-based contract-first 결정으로 강제 변환. actor use case → resource → endpoint → schema 흐름이 끊기지 않게 보장.

## 2. 사용 시점 (When to invoke)

- `define-tech-stack` (backend framework) + `design-data-model` (entity) 결정 후
- 신규 외부 통합 추가 (3rd-party client / SaaS partner)
- breaking change 감지 / API redesign 강제
- 신규 client (mobile / web / AI agent) 추가로 contract 재평가
- contract drift incident (client 와 server schema 불일치) 후 재설계

## 3. 입력 (Inputs)

### 필수
- `define-tech-stack` 산출물 (backend framework)
- `design-data-model` 산출물 (entity / 관계)
- actor list + use case 매핑 (§2 의 산출물)
- expected client 종류 (web / mobile / AI agent / 3rd-party)
- versioning 정책 가설 (URL versioning / header / GraphQL field deprecation)

### 선택
- 기존 API (있다면 backward-compat 강제)
- 외부 partner contract (있다면 호환성 강제)
- SLA / SLO (latency, uptime per endpoint)
- compliance 요구 (PII redaction, GDPR field-level access)

### 입력이 부족할 때 forcing question
- "client 가 서로 다른 view 를 원하면? GraphQL 이 적합한지 vs REST + multiple endpoints 비교"
- "현재 use case 중 highest QPS 와 most complex aggregation 은? gRPC vs REST 의 차이가 의미 있는 지점"
- "client 가 새 field 를 무시하는가, 거부하는가? schema evolution 정책에 영향"
- "auth model 은? OAuth2 / JWT / API key / session cookie 중 client 별 어느 것?"

## 4. 핵심 원칙 (Principles + Posture)

이 skill 의 운영 posture:

- **입장 취함, hedge 금지** — "REST vs GraphQL 모두 가능" 거부. use case 매트릭스로 한쪽 추천 + 변경 조건 명시.
- **사용자 입력을 challenge** — "REST 로 가자" 발화에 그대로 따르지 말고 "이 use case (예: nested data, multiple views) 에 GraphQL 이 더 fit 한지 검토했나?" 로 push.
- **Specificity 강제** — "good error handling" 거부. error code, retry-after header, idempotency key 같이 구체.
- **Contract-first bias** — schema 가 code 에서 도출되는 게 아니라 schema 가 정의되고 code 가 따라간다. 명시적으로 OpenAPI / GraphQL SDL / proto 작성.

도메인 원칙:

1. **Actor = API client 식별 단위** — actor 별로 어떤 endpoint 호출하는지 매핑. 권한 / rate-limit 도 actor 단위.
2. **Resource ≠ DB table** — aggregation / projection 가능. 1:1 매핑은 가설이지 default 아님.
3. **Error taxonomy 는 도메인 분류** — 4xx/5xx 만으로 부족. domain error code (예: `INVENTORY_INSUFFICIENT`, `PAYMENT_DECLINED`) 명시.
4. **Versioning 은 정책** — URL prefix / header / SDL deprecation 중 선택 + breaking change 정의 + sunset 일정.
5. **Contract test = schema validation + consumer-driven** — schema 만 검증하면 부족. consumer 기대 행동까지 검증.

## 5. 단계 (Phases)

### Phase 1. API style 선택

다음 매트릭스로 평가:

| Style | Best fit | 단점 | Mature ecosystem |
|-------|----------|------|------------------|
| REST | resource-centric, public API, cacheable, simple clients | over-fetching, multiple round-trips | ✓ |
| GraphQL | nested data, multi-view clients, evolution-tolerant | caching 복잡, security (cost analysis), N+1 | ✓ |
| gRPC | service-to-service, low latency, strict typing | browser support 제한, debugging 어려움 | ✓ |
| Webhook / event | async push, fire-and-forget | reliability 보장 어려움 | ✓ |

추천 + 변경 조건 명시.

### Phase 2. Resource / operation 매핑

actor use case → API operation 매핑 표:

| Actor | Use case | API resource / mutation | Method (REST) / Type (GQL) | Auth required |
|-------|----------|-------------------------|----------------------------|--------------|
| User | 회원가입 | POST /users | REST POST | none |
| User | 프로필 조회 | GET /users/me | REST GET | bearer |
| ... | ... | ... | ... | ... |

### Phase 3. Schema 정의

각 resource / mutation 의 request / response schema:

```yaml
# OpenAPI 예시 또는 GraphQL SDL 또는 proto
POST /users:
  request:
    email: string (email format)
    password: string (min 12)
  response 201:
    id: uuid
    email: string
    created_at: datetime
  response 400:
    error: VALIDATION_ERROR | EMAIL_EXISTS
    details: ...
```

### Phase 4. Error taxonomy + Versioning + Auth

- **Error taxonomy**: HTTP status family + domain error code + machine-readable details
- **Versioning 정책**: URL prefix (예: /v1/) / header (예: API-Version) / GraphQL field @deprecated + sunset 일정
- **Auth 방식**: OAuth2 (어느 grant) / JWT (어느 algorithm) / API key (rotation 정책) — actor 별

### Phase 5. Contract test 전략 + ADR handoff

- **Schema validation**: OpenAPI validator / GraphQL introspection / proto compile
- **Consumer-driven contract**: Pact 또는 동등 — consumer 가 expected interaction 정의, provider 가 검증
- **Breaking change 감지**: schema diff tool (oasdiff / graphql-inspector) CI integration

마지막에 `write-adr` 호출 권장: API style 선택 / versioning 정책 / auth 결정의 ADR.

## 6. 산출물 형식 (Output format)

> **Note**: structured 출력 강제. prose 변환 금지.

```markdown
## design-api-contract Output — <project / domain>

### Summary
<3 줄: 선택된 style / 핵심 resource 수 / 가장 큰 versioning risk>

### Style Decision
| Considered | Score (1-10) | Reason |
|-----------|--------------|--------|
| REST | ... | ... |
| GraphQL | ... | ... |
| gRPC | ... | ... |
| Webhook | ... | ... |

Selected: <style> — <1줄 근거>
Change conditions: <어떤 signal 시 재평가>

### Actor → Operation Map
| Actor | Use case | Operation | Method/Type | Auth |
|-------|----------|-----------|-------------|------|
| ... | ... | ... | ... | ... |

(use case 모두 매핑 — 누락 시 §11 검증 실패)

### Schema Excerpts
\`\`\`yaml or graphql or proto
<핵심 endpoint / type 의 schema>
\`\`\`

(at minimum 3 endpoint / type)

### Error Taxonomy
| HTTP status | Domain code | When | Details schema |
|-------------|-------------|------|----------------|
| 400 | VALIDATION_ERROR | input format invalid | { field, reason } |
| 409 | DUPLICATE_RESOURCE | unique constraint violation | { resource, field } |
| 503 | DEPENDENCY_UNAVAILABLE | downstream service down | { service, retry_after_s } |
| ... | ... | ... | ... |

### Versioning Policy
| Aspect | Decision |
|--------|----------|
| Strategy | URL prefix / header / GQL deprecation |
| Breaking change definition | <명시> |
| Deprecation timeline | <예: 6mo notice + 6mo dual-run> |
| Sunset notification | <예: header + email + docs banner> |

### Contract Test Strategy
| Layer | Tool | Frequency |
|-------|------|-----------|
| Schema validation | <tool> | per PR |
| Consumer-driven | <tool, 예: Pact> | per consumer PR |
| Breaking change diff | <oasdiff 등> | per PR + scheduled |

### ADR Handoff
다음 단계: `/buddy:write-adr "<title>"` 또는 chain `/buddy:chain design-api-contract,write-adr -- "<endpoint set>"`.

### Next Step
<구체 action — 1줄: 예 "design-data-model 으로 backfill 한 entity ↔ resource 매핑이 일치하는지 verify-quality 단계에서 contract test 작성">
```

## 7. Cross-phase cascade

- **§4 plan-build**: API endpoint 별 implementation task 가 actor track 분배의 단위
- **§5 build-feature**: contract-first 라 schema 가 먼저, code generation (OpenAPI codegen / proto compile) 이 build 의 일부
- **§6 verify-quality**: contract test (schema validation + consumer-driven) 가 quality gate 의 일부
- **§7 ship-release**: breaking change 감지가 release readiness gate
- **§8 iterate-product**: API metric (per-endpoint p99, error rate by domain code) 이 health monitoring 입력

## 8. 다음 skill (next in stage flow)

- `write-adr` — 본 skill 의 결정 영속화
- (별도 plan) `design-event-schema` — async / webhook 설계
- (별도 plan) `design-auth-model` — auth 결정이 본 skill 보다 deep 한 경우

권장 chain: `define-tech-stack → design-data-model → design-api-contract → write-adr`

## 9. 다른 skill 과의 경계 (충돌 방지)

- **vs `define-tech-stack`** — define-tech-stack 은 backend framework 선택 (Spring vs FastAPI). 본 skill 은 그 위의 contract 설계.
- **vs `design-data-model`** — design-data-model 은 DB 모양. 본 skill 은 API 모양. resource 가 entity 와 다를 수 있음 (aggregation / projection).
- **vs `design-event-schema`** (미구현) — 본 skill 은 sync request/response. event-schema 는 async event payload. 둘 다 필요한 경우 본 skill 후 별도 호출.
- **vs `design-mcp-server`** — design-mcp-server 는 MCP 특수 (tool schema, transport). 본 skill 은 일반 client-facing API. system 이 MCP 도 expose 하면 본 skill 후 design-mcp-server 가 후속.
- **vs `consult-design-system`** — consult-design-system 은 UI design system, 무관.

## 10. 중요 규칙

- **Schema-first** — code 가 schema 를 정의하지 않고, schema 가 code 를 정의. OpenAPI / SDL / proto 가 source-of-truth.
- **Versioning 정책 누락 금지** — 결정 안 하면 첫 breaking change 에서 무계획 대응.
- **Auth 는 endpoint 단위 결정** — system-wide 만으로 부족.
- **Contract test 전략 누락 금지** — consumer 가 깨지는지 알 수 없는 system 은 production-grade 아님.
- **Read-only on production** — code 변경 안 함.

## 11. Verification gate — 완료 선언 전 self-check

- [ ] §5 의 5 phase 누락 없이 실행
- [ ] §6 의 7 출력 섹션 (Summary / Style Decision / Actor-Operation Map / Schema Excerpts / Error Taxonomy / Versioning / Contract Test / ADR Handoff / Next Step) 모두 채워짐
- [ ] Style Decision 표가 ≥ 3 styles 평가 + 선택 + 변경 조건
- [ ] Actor → Operation Map 이 actor-use case 매핑의 모든 entry 포함
- [ ] Schema Excerpts 가 ≥ 3 endpoint / type 포함
- [ ] Error Taxonomy 가 ≥ 5 domain error code (HTTP status 4 family 만으론 부족)
- [ ] Versioning Policy 가 strategy + breaking change 정의 + deprecation timeline + sunset 모두 명시
- [ ] Contract Test Strategy 가 ≥ 2 layer (schema + consumer) + tool + frequency
- [ ] §4 posture 적용 — hedge 표현 없음
- [ ] §0 anti-pattern 들이 산출물에 등장하지 않음
- [ ] ADR handoff 라인 명시

하나라도 no 면 해당 phase 로 돌아가 보강 후 재검증.
