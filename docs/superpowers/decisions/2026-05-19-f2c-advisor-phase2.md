# ADR-015 — F2.C Advisor design (W7-3b, v1.0 entry C-3) — Phase 2/3

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: —
**Related**: ADR-009 (cli buddy vision), ADR-010 (v1.0.0 scope — closes C-3), ADR-011 (release policy), ADR-012 (F2.A substrate), ADR-013 (F2.B substrate), ADR-014 (F2.C Phase 1 foundation substrate)

## Context

ADR-014 split F2.C into 3 phases and shipped Phase 1 (knowledge retrieval foundation) as v0.10.0. Phase 2 is the advisor itself — the layer that *uses* the retrieval primitive + W7-2's usage metric to produce **actionable friend-tone Korean advisories**. v0.11.0 closes whole-product v1.0.0 entry condition **C-3**.

User feedback (2026-05-19 F2.C Q1 follow-up) framed the value:

> "토큰을 비효율적으로 사용한다면, prompt 사용과 skill 사용등에 있어서 추천 제안을 해줄수 있다는 것이야. ... 유저의 사용 패턴을 통해, 반복적인 작업에 대해서는 skill 을 generate 하여, 해당 스킬을 사용하도록 해줄수도 있어."

Phase 2 covers the *recommendation generation*. Phase 3 (v0.12.0, skill autogen) is when the recommendation becomes auto-generated skill spec.

Current fragments after v0.10.0:

- `internal/usage/Service` — 7 metric primitives.
- `internal/knowledge/{Store, BM25Index, HybridSearch, Embedder}` — retrieval.
- `internal/diagnose/Doctor` — proven `Diagnostic + Report + Thresholds` pattern.
- `internal/persona/` — friend-tone Korean catalog with key-typed templates.
- `internal/config/` — config.yaml + validation; supports threshold-style keys.

Gap: nothing combines them. No advisor data model, no rules, no advisor CLI / MCP / TUI surface, no advisory persistence.

## Decision

W7-3b ships **Advisor v1.0** locking in four design choices per user confirm:

### Q1 — Combinator: metric is trigger, retrieval is evidence

The advisor walks rules over the live usage snapshot. Each rule decides "this metric crossed the threshold" → spawn an advisory. The advisor then runs a retrieval query keyed by the advisory's kind (e.g. `"token spike high token usage"`) against the knowledge store; top-3 hits attach as `Evidence` items so the advisory's prose can reference *past sessions where a similar pattern showed up*.

Why this combinator: deterministic decision logic (rules are testable in pure Go) + retrieval enriches without changing the verdict. The advisor remains explainable — every advisory's `Evidence` field shows *which metric value triggered* and *which past chunks are similar*. Retrieval failure (Python embedder unavailable) degrades gracefully: advisory still fires, evidence list just contains the metric row.

### Q2 — Rule definition: `buddy config` thresholds (5+ keys)

The threshold knobs live in `~/.buddy/config.yaml` per the user's explicit answer. Five v0.11.0 rules ship:

| Rule kind            | Config key                      | Default | Trigger                                                                                            |
|----------------------|---------------------------------|---------|----------------------------------------------------------------------------------------------------|
| `token-spike-day`    | `advisor.token_spike_ratio`     | 1.5     | Last-24h spend / last-7d daily-avg ≥ ratio                                                         |
| `long-session`       | `advisor.long_session_hours`    | 4       | Any active session whose `last_active - started_at` ≥ N hours                                      |
| `low-cache-ratio`    | `advisor.low_cache_pct`         | 30      | Last-24h cache-hit-ratio percent < N                                                               |
| `session-volume-day` | `advisor.session_volume_per_day`| 20      | Sessions started in last 24h > N (suggests fragmented work)                                        |
| `token-daily-cap`    | `advisor.token_daily_threshold` | 500000  | Last-24h total tokens > N (cost / break suggestion)                                                |

Plus operational:

- `advisor.disabled` (bool, default false)
- `advisor.dedup_window` (duration, default 24h) — same kind doesn't re-fire inside this window
- `advisor.poll_interval` (duration, default 1h) — daemon cadence

### Q3 — Trigger surfaces: all four

Per user answer (CLI + TUI + MCP + daemon):

- **CLI**: `buddy advise [--since DUR] [--persist]` — on-demand evaluation. `--persist` writes to advisories table. Default prints fresh advisories only.
- **TUI**: ModeUsage pane gains an "조언" section under existing metric blocks. AdvisorFetcher hooks the same fresh-conn-per-call pattern as UsageFetcher / HookStatsFetcher.
- **MCP**: `usage_advise(since?, k?)` returns `Advisory[]` for LLM consumption. W7-5 Notification ships will pull from this.
- **daemon**: new `advisorMonitor` goroutine. Polls every `advisor.poll_interval`, generates advisories, writes new ones to advisories table (skipping rows that hit the dedup window). Disabled via `advisor.disabled`.

### Q4 — Output model: Advisory struct + advisories table (migration v7)

```go
type Severity string
const (
  SeverityInfo  Severity = "info"
  SeverityWarn  Severity = "warn"
  SeverityHigh  Severity = "high"
)

type Advisory struct {
  ID        int64
  Kind      string         // rule kind, e.g. "token-spike-day"
  Severity  Severity
  Message   string         // friend-tone Korean prose, rendered via persona.M
  Evidence  []EvidenceItem // metric + chunk references
  CreatedAt time.Time
  Muted     bool
}

type EvidenceItem struct {
  Type    string  // "metric" | "chunk"
  Detail  string  // human-readable; metric value or chunk preview
  ChunkID int64   // 0 when Type == "metric"
}
```

Migration v7:

```sql
CREATE TABLE advisories (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  kind          TEXT NOT NULL,
  severity      TEXT NOT NULL,
  message       TEXT NOT NULL,
  evidence_json TEXT NOT NULL DEFAULT '[]',
  created_at    INTEGER NOT NULL,
  muted         INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_advisories_created    ON advisories(created_at);
CREATE INDEX idx_advisories_kind_muted ON advisories(kind, muted);
```

## Alternatives considered

### Option A — Inline ephemeral advisories (no advisories table, rejected)

Cheaper to ship, but user answer explicitly chose "daemon 주기적 생성 + persistence (advisories 테이블)". Persistence enables dedup + future mute UX + W7-5 Notification delivery records.

### Option B — Single combined `usage_advise` MCP tool that returns prose+rationale (rejected)

Returns a single advisory blob. Loses the structured `Evidence[]` shape that makes the advisory inspectable. The chosen `Advisory[]` shape lets W7-5 Notification dispatch per-severity and lets the TUI sort/filter.

### Option C — LLM-driven prose generation via Claude API (rejected by ADR-014)

ADR-014 already settled this: local stack only. Phase 2 sticks to template-based prose + retrieval evidence; LLM-driven prose is a Phase-2.x trigger (only if user-visible quality is insufficient).

### Option D — Rule definition inline in code (rejected by user Q2 answer)

Hardcoded threshold pattern (doctor.go style) was the recommended option. User picked the config-driven path so threshold tuning doesn't require recompiles. Implementation effort is small — config.go already has the validation framework + 8 reason keys.

## Consequences

- **`internal/db/migrations.go` v7** — `advisories` table per above.
- **`internal/advisor/`** (new package):
  - `types.go` — `Advisory`, `Severity`, `EvidenceItem`.
  - `rules.go` — 5 rule funcs (one per kind). Each takes `Thresholds` + `usage.Overview` + (optional) retrieval handle and returns `*Advisory` or nil.
  - `evaluator.go` — `Evaluator{Thresholds, Usage, Knowledge}` with `Run(ctx) → []Advisory`. Composes rule walk + retrieval evidence enrichment + dedup against the live advisories table.
  - `store.go` — `Store` with `Insert / List / Get / Mute / DeleteOlderThan`.
- **`internal/persona/`** — 5 new `KeyAdvisor*` template keys (one per rule). koCatalog populated; en gets the fallback path.
- **`internal/config/`** — 8 new keys (5 rule thresholds + disabled + dedup_window + poll_interval). Validation: positive floats / non-negative durations.
- **`cmd/buddy/advise_cmd.go`** — `buddy advise [--since DUR] [--persist] [--all-history] [--mute KIND]`.
- **`internal/mcp/advise_tool.go`** — `usage_advise` MCP tool. server.go wires it via `Options.Advisor`.
- **`internal/tui/model.go`** — ModeUsage pane gains advisory section + AdvisorFetcher closure. `cmd/buddy/tui_cmd.go` wires.
- **`internal/daemon/daemon.go`** — `runAdvisorMonitor` goroutine alongside session monitor + outbox aggregator.
- **`docs/cli-buddy-spec.md` §9 W6** — Phase 2 marked Done; Phase 3 (skill-gen) remains.
- **`docs/BACKLOG.md`** — Wave 7 W7-3b ✅; progress 5/9 → **6/9 (67%)** with C-3 closed.
- **`docs/HANDOFF.md`** — v1.0.0 entry table updated.

## Verification

When W7-3b ships (v0.11.0):

```bash
# After buddy v0.11.0 install + sessions accumulated + knowledge ingested:

buddy advise                                  # current advisories
buddy advise --persist                        # writes to advisories table
buddy advise --since 168h --all-history       # last week's persisted advisories

# MCP (Claude Code 안):
# mcp__buddy__usage_advise                     # Advisory[] JSON

buddy daemon start                            # advisor monitor goroutine kicks in
sleep 3700 && sqlite3 ~/.buddy/buddy.db 'SELECT COUNT(*) FROM advisories'
                                              # rows accumulated by polling

buddy tui                                     # U → Usage pane shows 조언 section

go test -race -count=1 ./internal/advisor/...  # ~15+ race-clean tests
make verify-versions                            # 5 sources on 0.11.0
```

## Trigger to revisit

- Template-based prose quality is poor → consider LLM-driven prose (Claude API option) as v0.11.x.
- Rule set of 5 is too narrow → add domain-specific rules (e.g. "session goal vs current activity drift" — that's actually F2.D territory; if it bleeds in, refactor).
- advisories table growth → add `buddy advise purge` (analogous to `buddy session purge`) as v0.11.x.
- User wants per-rule mute UX → add `buddy advise mute <kind>` (the schema already has the `muted` column).

## References

- ADR-009 (cli buddy vision) — F2.C area definition.
- ADR-010 (v1.0.0 scope) — C-3 condition that this ADR closes.
- ADR-014 (F2.C Phase 1) — retrieval substrate.
- User feedback 2026-05-19 (F2.C Q1 + advisor 4-Q) — full quotes in ADR-014 §Context.
- `internal/diagnose/doctor.go` — the pattern this advisor extends.
