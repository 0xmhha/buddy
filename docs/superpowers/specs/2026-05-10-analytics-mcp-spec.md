# analytics-mcp — Spec

> **Date**: 2026-05-10 (W4-2.1 phase accepted 2026-05-11)
> **Status**: **Accepted (v0.2.0 — phase W4-2.1 stub published)**. 7 tool registration + friend-tone "backend not configured" 응답 + race-clean test 까지 ship. 본격 adapter 구현 (W4-2.2 custom SQL → W4-2.3 handler 본격) 은 다음 cycle.
> **Track**: plugin buddy (charter §3 의 4 자산 중 *MCP* 영역)
> **Related**:
> - [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §2.4.2 그룹 3 MCP 결정
> - missing-skills-inventory note §2.3 D-C C2 분할 진입 (file removed in v0.2.0 doc cleanup; 결정 요지: §8 Cluster F 7 skill 작성 후 trigger 활성)
> - [`plugin/skills/analyze-feature-adoption/PROCEDURE.md`](../../../plugin/skills/analyze-feature-adoption/PROCEDURE.md), [`analyze-user-cohort`](../../../plugin/skills/analyze-user-cohort/PROCEDURE.md), [`analyze-customer-feedback-corpus`](../../../plugin/skills/analyze-customer-feedback-corpus/PROCEDURE.md), [`audit-error-budget`](../../../plugin/skills/audit-error-budget/PROCEDURE.md)

---

## 1. 목적

§8 cluster F 의 7 데이터 분석 skill (`analyze-feature-adoption`, `analyze-user-cohort`, `analyze-actor-failure-rate`, `analyze-cost-anomaly`, `triage-customer-support-ticket`, `analyze-customer-feedback-corpus`, `audit-error-budget`) 이 모두 *외부 데이터 source* 의존. 사용자의 production 데이터에 *MCP server 통한 표준 access* 제공.

본 server 는 *plugin buddy 의 4 자산 중 MCP* 영역. 위 7 skill 의 PROCEDURE.md 가 *MCP tool 호출* 로 funnel / AB / cohort / cost / error budget 데이터 직접 query 가능.

---

## 2. Scope

### 2.1 노출 MCP tool

| tool | 책임 | 매칭 skill |
|------|------|---------|
| `analytics_query_funnel` | AARRR funnel 단계별 conversion 측정 | `analyze-feature-adoption`, `optimize-conversion-funnel` |
| `analytics_query_cohort` | acquisition cohort × retention curve | `analyze-user-cohort` |
| `analytics_query_ab_experiment` | A/B test 결과 (variant + metric + significance) | `analyze-ab-experiment` (구현됨) |
| `analytics_query_actor_failure` | actor 별 failure rate + recovery time | `analyze-actor-failure-rate` |
| `analytics_query_cost` | cost timeline + anomaly detection | `analyze-cost-anomaly`, `audit-cost-efficiency` |
| `analytics_query_slo_burn` | SLO error budget burn rate (multi-window) | `audit-error-budget` |
| `analytics_query_feedback_corpus` | 텍스트 corpus 검색 + 토픽 / sentiment | `analyze-customer-feedback-corpus` |

총 7 tool — 7 skill 1:1 매핑.

### 2.2 Non-goals

- *raw event ingest* — production 의 analytics pipeline 책임 (Mixpanel / Amplitude / 자체)
- *visualization* — MCP 는 data 만, 시각화는 사용자 dashboard 또는 Claude Code 의 markdown 출력
- *write* — 본 server 는 *read-only*. data 쓰기 / 변경 X (security / compliance)

---

## 3. 아키텍처

### 3.1 4 layer

```
[Claude Code session]
         │
         │ MCP stdio JSON-RPC
         ▼
[analytics-mcp server]
         │
         ├─ tool dispatch
         │     analytics_query_funnel
         │     analytics_query_cohort
         │     analytics_query_ab_experiment
         │     analytics_query_actor_failure
         │     analytics_query_cost
         │     analytics_query_slo_burn
         │     analytics_query_feedback_corpus
         │
         │
         ▼
[adapter layer — backend 별]
  ├─ Mixpanel adapter
  ├─ Amplitude adapter
  ├─ Datadog adapter
  ├─ Custom SQL adapter
  ├─ Elasticsearch adapter (text corpus)
  └─ Stripe adapter (cost / billing)
         │
         ▼
[사용자 production data sources]
```

### 3.2 Adapter 패턴

각 backend (Mixpanel / Amplitude / etc) 별 *adapter* :
- 같은 MCP tool signature
- 다른 *internal API* 호출 + 결과 변환
- 사용자 환경 변수 / config 로 backend 선택

→ 사용자가 *어느 analytics provider* 사용 중인지에 따라 adapter activate.

---

## 4. MCP tool signatures

### 4.1 analytics_query_funnel

```typescript
input: {
  product?: string  // 다중 product 시
  funnel: {
    stages: string[]  // e.g. ["signup", "first_action", "retention_d7"]
  }
  time_range: { from: ISODate, to: ISODate }
  segment?: { dimension: string, value: string }
}
output: {
  funnel: Array<{
    stage: string
    count: number
    conversion_from_prior: number  // %
    drop_off_from_prior: number    // %
  }>
  total_users: number
  time_range_used: { from, to }
}
```

### 4.2 analytics_query_cohort

```typescript
input: {
  cohort_dimension: "weekly" | "monthly" | "quarterly"
  retention_metric: "active" | "revenue" | "feature_use"
  segments?: string[]
}
output: {
  cohorts: Array<{
    cohort_id: string  // e.g. "2026-W18"
    size: number
    retention: { D1, D7, D30, D90, D180 }  // %
    ltv?: number  // if revenue metric
    cac?: number  // if attribution available
  }>
}
```

### 4.3 analytics_query_ab_experiment

```typescript
input: {
  experiment_id: string
}
output: {
  experiment: { id, name, started_at, ended_at? }
  variants: Array<{
    variant: string
    sample_size: number
    primary_metric_value: number
    confidence_interval: [number, number]
  }>
  significance: {
    p_value: number
    confidence: number  // %
    statistically_significant: boolean
  }
  guardrail_metrics?: Array<{ metric, change, alert }>
  recommendation: "ship" | "revert" | "continue" | "inconclusive"
}
```

### 4.4 analytics_query_actor_failure

```typescript
input: {
  actor: "user" | "system" | "3rd-party" | "external-tool"
  time_range: { from, to }
}
output: {
  actor: string
  failure_rate: number  // %
  predictability_variance: number
  mttr_minutes: number
  blast_radius: number  // count of affected actors
  trust_score: number  // 4-dim composite (Release It! Nygard)
  recovery_patterns_applied: Array<"retry" | "circuit_breaker" | "fallback" | "bulkhead" | "timeout" | "idempotent">
}
```

### 4.5 analytics_query_cost

```typescript
input: {
  time_range: { from, to }
  drill_down?: "service" | "component" | "region" | "account" | "tag"
  anomaly_detection?: boolean  // if true, return spikes
}
output: {
  total_cost_cents: number
  by_dimension: Record<string, number>
  anomalies?: Array<{
    dimension: string
    spike_at: ISODate
    cost_cents: number
    z_score: number  // how many σ from baseline
  }>
}
```

### 4.6 analytics_query_slo_burn

```typescript
input: {
  sli: string  // e.g. "p99_latency_ms"
  slo_target: number
  time_window: "1h" | "6h" | "1d" | "3d" | "30d"
}
output: {
  sli, slo_target, time_window
  current_value: number
  budget_remaining_pct: number
  burn_rate: number  // multi-window
  alert_level: "info" | "warning" | "critical"
  release_gate_decision: "ship" | "limit" | "freeze" | "rollback"
}
```

### 4.7 analytics_query_feedback_corpus

```typescript
input: {
  source?: "cs_ticket" | "nps_comment" | "app_review" | "interview"
  time_range: { from, to }
  topic_modeling?: boolean
  sentiment_analysis?: boolean
}
output: {
  total_items: number
  topics?: Array<{
    topic: string
    item_count: number
    sentiment: { positive, negative, neutral }
    verbatim_quotes: string[]
  }>
  nps_segments?: {
    promoter: { count, top_topics }
    passive: { count, top_topics }
    detractor: { count, top_topics }
  }
}
```

---

## 5. 구현 — Go MCP SDK

### 5.1 권장 stack

| 영역 | 도구 |
|------|------|
| MCP SDK | `modelcontextprotocol/go-sdk` (또는 `mcp-go` upstream) |
| Transport | stdio (default) — JSON-RPC over stdin/stdout |
| Adapter SDK | per-backend (Mixpanel / Amplitude / Datadog) — 각자 Go client |
| Config | env var + JSON config (`~/.buddy/analytics-mcp.json`) |
| Logging | structured (JSON) — buddy 의 hook_events 와 정합 |
| Auth | per-backend API key (env var) |

### 5.2 Binary entry

```
cmd/buddy-analytics-mcp/main.go  (신규)
```

또는 기존 `cmd/buddy-mcp/` 와 통합 — *복합 MCP server* 가능 (doctor / stats / feature.* 와 한 binary).

→ **결정 권장**: 기존 `cmd/buddy-mcp/` 확장 (single binary, multi-namespace tool — `doctor.*`, `stats.*`, `feature.*`, `analytics.*`).

### 5.3 Migration strategy

기존 `cmd/buddy-mcp/` 의 doctor / stats / feature tool 유지. analytics namespace 추가:

```go
// cmd/buddy-mcp/main.go
server.RegisterTool("analytics_query_funnel", analyticsFunnelHandler)
server.RegisterTool("analytics_query_cohort", analyticsCohortHandler)
// ... 7 tool 모두 등록
```

### 5.4 Adapter 결정 priority

처음 v1 = **single adapter (custom SQL)**. 사용자가 *자체 PostgreSQL / MySQL* 에 events 적재 가정. 후속 v2+ 에서:

- v2: Mixpanel adapter (가장 흔한 SaaS)
- v3: Amplitude adapter
- v4: Datadog adapter (cost / SLO 영역 강함)
- v5+: 사용자 요청 기반 추가

---

## 6. Acceptance — analytics-mcp v0.1.0

| 항목 | gate |
|------|-----|
| 7 tool 모두 MCP register | `claude mcp tools list buddy-mcp` 에 7 tool 등장 |
| Custom SQL adapter | PostgreSQL / MySQL 의 events 테이블 query 정상 |
| Tool signature 정합 | 본 spec §4 의 input/output schema 일치 |
| Error handling | adapter 의 connection / auth fail → friend-tone error message (persona catalog 사용) |
| race-clean test | `go test -race ./cmd/buddy-mcp/...` 통과 |
| Documentation | README + Claude Code MCP install 안내 |
| skill 정합 | 7 skill PROCEDURE.md 의 *MCP tool 호출 example* 추가 |

---

## 7. cli buddy 트랙 분리 (D-C C2)

| MCP server | 트랙 | trigger |
|----------|-----|--------|
| **analytics-mcp** (본 spec) | plugin buddy | §8 cluster F 7 skill 작성 후 — *현재 trigger 발화* |
| **feature-management-mcp** | cli buddy | cli buddy spec 작성 시 (W3-1 done) |

→ feature-management-mcp 는 cli buddy 의 *agent runtime* 이 *agent metadata* 관리 시 활용. 별도 spec.

---

## 8. Implementation phases

| phase | 범위 | 비용 | 상태 |
|-------|------|------|------|
| W4-2.1 | MCP tool registration (7 tool stub) | LOW | ✅ Done (2026-05-11, plugin v0.2.0 commit `bdf957e^..`). `internal/mcp/analytics_tool.go` 신규 + `server.go` 의 `addAnalyticsTools` wire-up |
| W4-2.2 | Custom SQL adapter (PostgreSQL / MySQL) | MED | ⏳ |
| W4-2.3 | 7 tool handler 본격 구현 | HIGH | ⏳ |
| W4-2.4 | 7 PROCEDURE.md 의 *MCP tool 호출 example* 추가 | LOW | ✅ Done (2026-05-11). 10 skill PROCEDURE (primary 7 + secondary 3: optimize-conversion-funnel / audit-cost-efficiency / triage-customer-support-ticket) 끝에 `## MCP integration (analytics-mcp v0.2.0+)` 섹션 일괄 추가 |
| W4-2.5 | error handling + friend-tone i18n | LOW | ✅ W4-2.1 안에서 흡수 (현재 한국어 stub message + `BUDDY_ANALYTICS_BACKEND` env var 안내). i18n split 은 v0.2 i18n sweep 과 함께 |
| W4-2.6 | race-clean test + integration test (synthetic events) | MED | 🟡 부분 (registration + stub behaviour `internal/mcp/analytics_tool_test.go` 4 test 통과. synthetic event integration 은 adapter 구현 후) |
| W4-2.7 | release v0.1.0 | LOW | ⏳ (analytics-mcp 단독 release tag 미발행 — plugin v0.2.0 안에 흡수. 향후 `analytics-mcp-v0.1.0` 별도 tag 후보) |

→ 6 phase × 평균 1~2 week = **2~4 month** estimate. W4-2.1 + W4-2.5 + W4-2.6 부분 = *현재 ship 됨*.

---

## 9. Open questions

| ID | 질문 | 결정 시점 |
|----|------|--------|
| Q-1 | first adapter = custom SQL? Mixpanel? | W4-2.2 시 |
| Q-2 | Cost cents (int64) 는 *어느 통화* default? USD vs locale-별 | W4-2.3 시 |
| Q-3 | NPS / sentiment 분석은 *MCP 안에서* 처리 vs *외부 LLM 위임*? | W4-2.3 시 |
| Q-4 | per-product / multi-product 지원? | W4-2.3 시 |
| Q-5 | rate limit / quota 정책 (외부 backend 의 API 비용 보호) | W4-2.6 시 |

---

## 10. References

- [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §2.4.2 그룹 3
- missing-skills-inventory note §2.3 D-C C2 (removed in v0.2.0 doc cleanup, see git history)
- 7 §8 Cluster F PROCEDURE.md (cascade in)
- `modelcontextprotocol/go-sdk` 또는 `mcp-go` upstream
- 기존 `cmd/buddy-mcp/` (doctor / stats / feature 자산)

---

> 본 spec 은 *Draft*. v1 구현은 plugin buddy v1.1.0 안정화 + cli buddy spec lock-in (W3-1 Accepted) 후 본격 진입.
