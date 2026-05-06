---
name: design-mcp-server
description: "MCP(Model Context Protocol) server 설계. tool / resource / prompt 정의, transport(stdio/http/sse), authentication, error handling, schema validation, idempotency, observability를 표준화. feature-management-saas-mcp 같은 외부 통합 서버 구축 시 사용. 트리거: 'MCP 서버 만들자' / 'MCP tool 설계' / 'feature.query 같은 tool' / 'MCP resource 정의' / 'agent에 노출' / 'MCP transport' / 'tool schema 작성'. 입력: 노출할 capability, 클라이언트(Claude Code/Codex/Cursor), auth 정책, persistence 모델. 출력: MCP server spec + tool/resource catalog + transport + auth + error model. 흐름: split-work-into-features → design-mcp-server → build-with-tdd."
type: skill
---

# Design MCP Server — Tool/Resource/Prompt 표준 설계

## 1. 목적

MCP (Model Context Protocol) server를 **standard / interoperable / observable** 형태로 설계한다.

MCP는 Anthropic이 제안한 LLM ↔ 외부 시스템 표준 프로토콜. 서버는 다음 3개를 노출:
1. **Tools**: agent가 호출하는 함수 (input → output)
2. **Resources**: agent가 읽는 read-only 데이터 (URI 기반)
3. **Prompts**: agent가 사용하는 prompt template

이 스킬을 통과한 MCP server는:
- tool 5요소(name, description, input_schema, output_shape, errors) 충족
- 클라이언트 (Claude Code, Codex, Cursor) 호환성 검증
- auth + rate limit + observability 명시
- error code + retry semantic 표준
- idempotency key 정책

## 2. 사용 시점 (When to invoke)

- `feature-management-saas-mcp` 같은 외부 시스템 LLM 통합
- 기존 internal API를 LLM에 노출 (read-only 또는 mutating)
- agent autonomous workflow 구축 (chain of tool calls)
- 다수 클라이언트 (Claude Code + Codex + Cursor) 호환 필요
- offline batch agent (cron + MCP)
- SaaS product에 MCP API 추가 (paid feature)

## 3. 입력 (Inputs)

### 필수
- 노출할 capability 목록 (read / write / search / action)
- target client (Claude Code / Codex / Cursor / 자체 agent)
- auth 정책 (API key / OAuth / token / none)
- persistence 모델 (DB / file / external API)

### 선택
- 기존 REST / GraphQL API (wrapping 시)
- billing / quota 모델 (paid MCP)
- multi-tenant 요구 (organization scope)

### 입력 부족 시 forcing question
- "tool은 read-only야 mutating이야? mutating이면 idempotency key 필요."
- "auth scope 구분돼? read-only token vs full token."
- "쿼리는 expensive해 cheap해? expensive하면 batch / pagination / cache."
- "transport 뭘 쓸 거야? stdio (local) vs http+sse (remote) vs sse (server-push)."

## 4. 핵심 원칙 (Principles)

1. **Tool name은 verb-noun + scope** — `feature.query`, `patch.verify`, `feature.record_adoption`. consistency.
2. **Description은 LLM이 읽음** — agent가 언제 호출할지 판단. 자연어 설명 + example.
3. **Input schema는 strict** — JSON Schema validation. unknown field reject (또는 warn).
4. **Output shape는 안정적** — schema versioning. breaking change → 새 tool name.
5. **Error는 의미적** — HTTP status 모방 (400 / 401 / 403 / 404 / 409 / 429 / 500). agent retry 가능 여부 명시.
6. **Idempotency key는 mutating tool 필수** — agent retry 시 중복 방지.
7. **Pagination은 cursor 기반** — offset 기반은 변경 시 깨짐.
8. **Observability** — tool 호출 로그 (caller, args, latency, error). audit + debug.

## 5. 단계 (Phases)

### Phase 1. Capability Mapping
노출할 capability 목록 → tool / resource / prompt 분류:
- 함수 호출 (input → output) → **tool**
- read-only 데이터 (URI) → **resource**
- prompt template (가변 인자) → **prompt**

### Phase 2. Tool Catalog
각 tool에 대해:
- `name`: verb.noun (예: `feature.query`)
- `description`: 자연어 (LLM이 invoke 판단)
- `input_schema`: JSON Schema (strict)
- `output_shape`: 응답 형식 + error 형식
- `idempotency`: required / optional / not-applicable
- `auth_scope`: read | write | admin
- `rate_limit`: per minute / per day
- `examples`: 호출 / 응답 예시 2-3개

### Phase 3. Resource Catalog
URI 기반 read-only:
- URI scheme (예: `feature://{feature_id}/spec`)
- mime-type (`application/json`, `text/markdown`)
- caching (ETag / Last-Modified)
- access control (visibility)

### Phase 4. Prompt Catalog
- `name`: 식별자
- `description`
- `arguments`: 변수 schema
- `template`: prompt 본문

### Phase 5. Transport Selection
- **stdio**: local subprocess. Claude Code, Codex CLI 일반적.
- **http**: remote SaaS. paid MCP, multi-tenant.
- **sse**: server-push (subscription, long-running).

각 transport에 auth / error / rate limit 매핑.

### Phase 6. Auth Design
- API key (header `Authorization: Bearer <token>`)
- OAuth 2.0 (delegated)
- token scope (organization / project / read / write)
- multi-tenant: token → tenant resolution
- key rotation 정책

### Phase 7. Error Model
표준 error response:
```json
{
  "error": {
    "code": "FEATURE_NOT_FOUND",
    "message": "Feature feat_999 does not exist",
    "retryable": false,
    "details": { ... }
  }
}
```

retryable 여부 명시 (agent가 자동 retry 결정).

### Phase 8. Observability
- 모든 tool 호출 로그 (timestamp, caller, args, output_size, latency, error)
- metric (calls per tool, p50/p95/p99 latency, error rate)
- distributed tracing (correlation ID)
- audit log (auth change, admin action)

## 6. 출력 템플릿 (Output Format)

```yaml
mcp_server:
  name: feature_registry
  version: "1.0.0"
  description: "Feature management SaaS MCP server"
  transport: http  # stdio | http | sse
  base_url: "https://mcp.feature-registry.example.com"

auth:
  schemes:
    - type: bearer_token
      header: Authorization
      format: "Bearer <token>"
  token_scopes:
    - read:features
    - write:features
    - read:patches
    - download:patches
    - admin
  multi_tenant: yes
  tenant_resolution: from_token

tools:
  - name: feature.query
    description: "feature 후보와 유사한 기존 feature를 검색한다. PRD에서 추출된 candidate를 query로 받아 semantic + interface + stack + license + security 가중치로 매치 후보를 반환한다."
    input_schema:
      type: object
      required: [query, target_stack, intended_use]
      properties:
        query: { type: string }
        domain_tags: { type: array, items: { type: string } }
        desired_flow: { type: array, items: { type: string } }
        target_stack: { type: object }
        constraints: { type: object }
        interfaces: { type: object }
        limit: { type: integer, default: 5, max: 50 }
    output_shape:
      matches: array of FeatureMatch
    idempotency: not-applicable  # read-only
    auth_scope: read:features
    rate_limit: 60/min
    examples:
      - input: { query: "email password login" }
        output: { matches: [{ feature_id: "feat_123", ... }] }

  - name: feature.store
    description: "신규 feature 명세를 저장한다."
    input_schema: { ... }
    output_shape: { feature_id, slug, status }
    idempotency: required  # mutating
    auth_scope: write:features
    rate_limit: 30/min

  - name: patch.verify
    description: "patch 다운로드 또는 적용 전 무결성, 서명, 보안, 라이선스 상태를 확인한다."
    auth_scope: read:patches

  - name: patch.download
    description: "권한, 결제, 라이선스 확인 후 patch 다운로드 URL을 발급한다."
    idempotency: required
    auth_scope: download:patches

resources:
  - uri_template: "feature://{feature_id}/spec"
    mime_type: "application/json"
    description: "feature 명세"
  - uri_template: "feature://{feature_id}/implementations"
    mime_type: "application/json"
  - uri_template: "patch://{patch_artifact_id}/verification"
    mime_type: "application/json"

prompts:
  - name: feature_reuse_review
    description: "feature.query 결과를 reuse / adapt / new로 분류"
    arguments:
      - name: query_result
        required: true
    template: |
      다음 매치를 검토하여 reuse_decision을 부여하라:
      {{query_result}}

      gate: license, security, visibility
      score 가중치: ...

error_model:
  - code: FEATURE_NOT_FOUND
    http_status: 404
    retryable: false
  - code: RATE_LIMIT_EXCEEDED
    http_status: 429
    retryable: true
    retry_after_seconds: 60
  - code: AUTH_INVALID
    http_status: 401
    retryable: false
  - code: SCOPE_INSUFFICIENT
    http_status: 403
    retryable: false
  - code: PATCH_VERIFY_FAILED
    http_status: 422
    retryable: false
  - code: SERVER_ERROR
    http_status: 500
    retryable: true

observability:
  log_format: structured_json
  log_fields: [timestamp, caller_id, tool_name, input_size, output_size, latency_ms, error_code]
  metrics:
    - name: tool_calls_total
      labels: [tool_name, status]
    - name: tool_latency_ms
      type: histogram
      labels: [tool_name]
    - name: tool_error_rate
      labels: [tool_name, error_code]
  tracing: opentelemetry
  audit_log:
    enabled: yes
    events: [auth_change, scope_change, admin_action, patch_download]

versioning:
  schema_version: "1.0.0"
  breaking_change_policy: "new tool name (e.g. feature.query.v2)"
  deprecation_window: "180 days"

client_compat:
  - client: claude-code
    transport: stdio | http
    tested: yes
  - client: codex-cli
    transport: stdio
    tested: yes
  - client: cursor
    transport: stdio
    tested: pending
```

## 7. 자매 스킬 (Sibling Skills)

- 앞 단계: `split-work-into-features` — `Skill` tool로 invoke
- 페어: `design-billing-system` — paid MCP tool 결제 처리
- 페어: `audit-security` — auth, input validation, prompt injection
- 다음 단계: `build-with-tdd` — tool 단위 TDD 구현
- 다음 단계: `setup-quality-gates` — schema validation 강제

## 8. Anti-patterns

1. **Tool description LLM 무시 작성** — agent가 invoke 판단 못함. 자연어 + 예시.
2. **Output shape unstable** — 매번 schema 변경. agent retry 어려움. versioning.
3. **Mutating tool에 idempotency 없음** — agent retry 시 중복 생성.
4. **Error code 없이 message만** — agent retry 결정 불가. structured error.
5. **Pagination offset 기반** — 새 record 추가 시 cursor shift. cursor 기반.
6. **Auth scope 단일** — read-only token / full token 분리 안 됨. principle of least privilege.
7. **Observability 사후 추가** — 디자인 단계 누락. 처음부터 structured log + metric + tracing.
8. **Schema validation 약함** — unknown field 통과. agent error 늦게 발견.
