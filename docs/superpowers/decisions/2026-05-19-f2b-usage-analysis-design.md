# ADR-013 — F2.B Usage Analysis design (W7-2, v1.0 entry C-2)

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: —
**Related**: ADR-009 (cli buddy vision expansion — F2.B area), ADR-010 (v1.0.0 scope — C-2 condition), ADR-011 (release policy — milestone-driven), ADR-012 (F2.A Session Monitor — W7-1 substrate)

## Context

ADR-009 declared F2.B Usage Analysis as one of the five `AI-usage coaching` responsibility areas. Current fragments in the repo:

- `internal/sessions/` package + `sessions` table (W7-1 ship, v0.8.0) — fields `id` / `pid` / `transcript_path` / `started_at` / `last_active` / `total_input_tokens` / `total_output_tokens` / `total_cache_read` / `total_cache_create` / `last_offset` / `ended_at` / `goal_text` / `metadata`.
- `internal/analytics/` MCP tools 7개 (`analytics_query_funnel` / `cohort` / `ab_experiment` / `actor_failure` / `cost` / `slo_burn` / `feedback_corpus`) — generic *product analytics* 도메인. AI-usage 분석과 도메인 mismatch.
- No CLI / TUI / MCP surface exposes AI-usage analytic primitives. ~30일 데이터가 누적되어도 사용자가 볼 channel 이 없음.

→ F2.B 는 ~5% (input 인 sessions 테이블만 존재, output surface 전무). W7-2 cycle 이 C-2 를 닫는다.

## Decision

W7-2 ships **Usage Analysis v1.0** with four design choices locked in:

### Q1 — Data source: sessions table only (no transcript JSONL reparsing)

`sessions` 테이블의 컬럼만으로 metric 을 derive 한다.

- token spend, session count, session duration, 시간대 분포, cache hit ratio, active/ended ratio, goal_text 보유율 — 8 metric 이상이 `sessions` 컬럼만으로 가능.
- transcript JSONL 재파싱 (message-length, time-to-first-tool-call, tool-call frequency) 은 v0.10+ 또는 W7-3 Advisory 가 필요 시점에 추가.
- 이유:
  - fsLister 와 reader path 충돌 / 캐시 설계 불필요 — single-source ownership.
  - 비용 낮음 — query 가 SQL aggregate 한 줄로 끝남.
  - sessions 테이블 자체가 W7-1 에서 ~30일 누적될 데이터의 1차 SSoT.

### Q2 — Storage: live aggregation (no derived `usage_metrics` table)

매 query 마다 `sessions` 테이블에서 SQL aggregate 로 계산. 사전 집계 / cache / migration v6 도입하지 않는다.

- 이유:
  - sessions row 수 estimate: 1일 ~50 세션 × 30일 = ~1500 rows. SQLite aggregate 가 ms 단위로 끝남.
  - migration / cache invalidation / daemon goroutine 추가 없이 가볍게 ship.
  - 만약 row 수가 ~100k 단위로 증가하면 그때 derived 테이블 도입 (W7-2.x 또는 v0.10+ patch — trigger 명시).

### Q3 — Surface: CLI + MCP + TUI pane (3-way)

세 channel 모두 v0.9.0 ship 범위에 포함한다.

- **CLI**: `buddy usage today` / `buddy usage session <id>` / `buddy usage skills` / `buddy usage trend` — 사용자가 shell 에서 즉시 확인하는 1차 channel.
- **MCP**: `usage_query_token_spend` / `usage_query_session_stats` / `usage_query_time_distribution` / `usage_query_top_sessions` / `usage_query_overview` — Claude Code 안에서 LLM 이 호출. W7-3 Advisory 가 이 tool 들을 input source 로 사용 예정.
- **TUI**: `buddy tui` 의 새 `Usage` pane — 기존 List / Detail / HookStats / Scheduler / LogTail 과 동급 mode. 토큰 / 시간 분포 viewer.

기존 7 analytics tools 의 repurpose 는 채택하지 않음 — funnel/cohort 의미가 AI-usage 도메인에 어색.

### Q4 — MVP scope: full ship (CLI + MCP + TUI + 5+ metric)

v0.9.0 (W7-2 ship) 에 위 3-way surface + 최소 5 metric 모두 포함.

**제공 metric (live aggregate)**:

1. **Token spend** — 일별 합 (input / output / cache_read / cache_create 분해), 기간 합, 평균/median 세션당.
2. **Session count** — 일별 시작 세션 수, 기간 총합.
3. **Session duration** — `last_active - started_at` 분포 (p50 / p90 / max), ended session 만.
4. **Time-of-day distribution** — `started_at` 의 hour bucket 분포 (peak hours 식별).
5. **Cache hit ratio** — `cache_read / (input_tokens + cache_read + cache_create)`.
6. **Top sessions** — 토큰 사용 top-N 세션 (id / goal_text / total tokens).
7. **Active vs ended** — `ended_at IS NULL` 비율 (현재 살아있는 세션 수).

→ 7 metric. ADR-009 명시 5+ 기준 충족.

**제공 surface**:

- CLI: `buddy usage today|session|skills|trend|overview`
- MCP: 5 tool (token_spend / session_stats / time_distribution / top_sessions / overview)
- TUI: Usage pane (Tab/shift-Tab navigation 으로 진입, summary table 형태)

## Alternatives considered

### Option A — sessions + transcript JSONL 재파싱 (rejected)

message-length, time-to-first-tool-call, tool-call frequency 같은 rich metric 가능. 그러나 fsLister 와 별도 reader path 가 필요하고 cache 설계 부담 — Q2 의 "가벼움" 원칙과 충돌. v0.10+ trigger 로 deferral.

### Option B — derived `usage_metrics` 테이블 + daemon 사전 집계 (rejected)

미래 row 수가 100k+ 가 되면 필요할 수 있으나 현재 ~1500 rows 기준에서는 과잉. migration v6 + invalidation 책임이 v0.9.0 ship 시간을 늘림.

### Option C — 기존 7 analytics tools repurpose (rejected)

`analytics_query_funnel` 을 'skill 단계별 turn count' 로 재사용 등. 그러나 funnel/cohort 의미 (stage 순서, retention metric) 가 AI-usage 도메인에 매핑 어색. 새 `usage_query_*` namespace 가 사용자 mental model 에 더 부합.

### Option D — CLI 만 ship, MCP / TUI 는 v0.10+ (rejected by user)

사용자가 "full — CLI + MCP + TUI pane + 5+ metric" 답변. W7-3 Advisory 가 MCP tool 을 input 으로 쓰는 path 까지 v0.9.0 단계에서 미리 갖추는 것이 W7-3 진입 마찰을 줄임.

## Consequences

- **`internal/usage/` package 신설** — sessions store 를 input 으로 받아 metric 7 종 aggregate. SQL 은 sessions store 의 *_test.go 와 동일한 패턴 (race-clean, single-file store + types + queries).
- **`cmd/buddy/usage_cmd.go` 신설** — `buddy usage today|session|skills|trend|overview` subcommand.
- **`internal/mcp/usage_tool.go` 신설** — 5 MCP tools. `server.go` 에 `addUsageTools(s, opts)` 추가.
- **`internal/tui/model.go` 확장** — Usage mode 추가 (기존 List / Detail / Scheduler / DeleteConfirm / LogTail / HookStats / Edit / Create 옆에 1개 더). Tab navigation 갱신.
- **`internal/mcp/server.go` Options** — `Sessions *sessions.Store` 또는 `Usage usage.Service` 추가 (현재 `Analytics analytics.Adapter` 와 같은 패턴).
- **No migration** — Q2 결정. `sessions` 테이블 그대로 사용.
- **`docs/cli-buddy-spec.md` §9 W7-2 → Done** + W4 phase 표 갱신.
- **`docs/BACKLOG.md` Wave 7 W7-2 → 100%** + Progress 재계산 (4/9 closed = ~44%).
- **`CHANGELOG.md` [0.9.0]** — milestone-driven release, ADR-011 준수.
- **W7-3 Advisory 진입 unblocked** — `usage_query_*` MCP tool 이 LLM 의 input source 가 됨.

## Verification

When W7-2 ships:

```bash
# After buddy v0.9.0 install + daemon 가동 + ~hours of sessions data:

buddy usage today                             # 오늘 token spend / session count
buddy usage session <id>                      # 단일 세션 metric
buddy usage trend --days 7                    # 최근 7일 trend
buddy usage overview                          # 7 metric 한 번에

# MCP 측 (Claude Code 안에서):
# mcp__buddy__usage_query_overview              # JSON aggregate
# mcp__buddy__usage_query_token_spend           # 일별 토큰 분해

buddy tui                                     # Usage pane 진입 가능 (Tab 으로 cycle)

go test -race -count=1 ./internal/usage/...   # ~10+ race-clean tests
go test -race -count=1 ./internal/mcp/...     # usage tool registration tests
make verify-versions                          # 5 sources on 0.9.0
```

## Trigger to revisit

- sessions row 수가 100k+ 으로 증가 → Option B (derived 테이블 + 사전 집계) 도입.
- W7-3 Advisory 가 transcript JSONL 의 message-length / tool-call frequency 를 필요로 함 → Option A (transcript 재파싱) 추가 ADR.
- 사용자가 "이 metric 보다 저 metric 이 더 유용하다" 피드백 → 7 metric 구성 재조정.
- F2.B metric 이 W7-3 Advisory 만의 input 으로 사용되고 사용자가 CLI 를 안 쓰면 → CLI surface 우선순위 downgrade.

## References

- ADR-009 (cli buddy vision expansion) — defines F2.B area.
- ADR-010 (whole-product v1.0.0 scope) — C-2 condition that this ADR closes.
- ADR-011 (release policy) — milestone-driven v0.9.0 framing.
- ADR-012 (F2.A Session Monitor) — W7-1 produced the `sessions` table that this ADR consumes.
- `internal/sessions/store.go` — input data source.
- `internal/db/migrations.go` v5 — `sessions` table schema (W7-1 ship).
- `internal/analytics/` — 도메인이 다른 generic analytics MCP tools (참고용, repurpose 안 함).
