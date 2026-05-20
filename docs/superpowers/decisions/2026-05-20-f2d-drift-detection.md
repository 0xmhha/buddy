# ADR-017 — F2.D Drift Detection design (W7-4, v1.0 entry C-4 — final C-x)

**Status**: Accepted (2026-05-20)
**Authors**: mhha (cli buddy track)
**Supersedes**: —
**Related**: ADR-009 (cli buddy vision — F2.D area), ADR-010 (v1.0.0 scope — closes C-4), ADR-011 (release policy), ADR-012 (sessions.goal_text substrate), ADR-014 (knowledge retrieval — embedder + cosine substrate), ADR-015 (F2.C Advisor — pipeline)

## Context

ADR-009 declared F2.D as drift detection — "단일 conversation 안 *원래 목적* vs *현재 turn* 의 semantic drift 감지". By v0.12.0 every substrate this requires is already shipped:

- `sessions.goal_text` — original goal extracted by fsLister on first user message (W7-1, ADR-012).
- `internal/knowledge/` chunks — every active session has chunks + embeddings via `buddy knowledge ingest` (W7-3a, ADR-014).
- `knowledge.PythonEmbedder` — local embedding of arbitrary text.
- `knowledge.CosineSimilarity` — vector comparison.
- `internal/advisor/Evaluator` — rule+evidence+dedup+persistence pipeline that auto-flows to CLI / MCP / TUI / notify (W7-3b/W7-5).

→ F2.D 는 **새 데이터 모델 없이** advisor 의 6번째 rule (`KindGoalDrift`) 로 ship 가능. v0.13.0 closes C-4 — 마지막 C-x condition.

## Decision

W7-4 ships **Goal Drift Detection v1.0** as a single new advisor rule + 3 threshold config keys. Per user 3-Q confirm (all defaults accepted):

### Q1 — Embedding cosine similarity (no LLM call)

For each active session, embed `goal_text` and the average of the recent N chunks; compute cosine. `score < threshold` (default 0.4) fires `KindGoalDrift`. Deterministic, offline, reuses W7-3a stack. LLM-judge stays a future trigger (only if user-visible accuracy is insufficient).

### Q2 — Recent N chunks (default 10)

Per-session window over the last N inserted chunks. Average embedding across those chunks → single "current activity" vector. Configurable via `advisor.goalDriftSampleChunks`. Chunk-count window (not time) keeps drift detection meaningful even on bursty / idle sessions.

### Q3 — Surface as advisor rule (no new package)

`ruleGoalDrift` joins the existing 5 rules. `Snapshot.DriftItems []SessionDrift` is filled in `buildSnapshot` (one entry per active session whose goal_text is non-empty AND has ≥ N chunks ingested). The rule walks `DriftItems` and emits one advisory per active session whose score < threshold. Pipeline reuse → automatic flow into:

- `buddy advise` CLI ✅
- `usage_advise` MCP tool ✅
- TUI Usage pane advisory section ✅
- daemon advisorMonitor → notify dispatch ✅

No new MCP tool, no new CLI subcommand, no new table, no migration. Just one rule + one snapshot field + three config keys.

### Data model deltas

```go
// advisor/types.go additions to Thresholds
GoalDriftDisabled     bool
GoalDriftThreshold    float64       // cosine sim below this fires (default 0.40)
GoalDriftSampleChunks int           // recent chunk window (default 10)

// advisor/rules.go: new Snapshot field
type SessionDrift struct {
    SessionID      string
    GoalText       string
    Score          float64    // cosine similarity, 0..1
    SampleChunks   int        // chunks the score was computed over
    WorstChunk     string     // preview for Evidence
}

// new ruleFn
func ruleGoalDrift(t Thresholds, snap Snapshot) *Advisory  // first triggered drift wins (one advisory per Run)
```

### Config keys (3 new)

```yaml
advisorGoalDriftDisabled: false
advisorGoalDriftThreshold: 0.4
advisorGoalDriftSampleChunks: 10
```

Validation: threshold in [0,1], sampleChunks ≥ 1.

### Embedder dependency handling

Goal drift detection requires the Python embedder. When `Knowledge` / `Embedder` is nil OR the embedder fails (ErrEmbedderUnavailable) on the goal_text or chunk embedding, `buildSnapshot` skips drift entirely and ruleGoalDrift returns nil. No noisy "drift unavailable" advisory — silent fall-through. Same graceful degradation pattern as knowledge_query.

## Alternatives considered

### Option A — LLM-judge via Claude API (rejected)

More semantically accurate but introduces API key management + per-call cost + non-determinism. ADR-014 already settled on local stack. v0.13.x can revisit if cosine accuracy is insufficient.

### Option B — Separate `internal/drift/` package + CLI / MCP / TUI surface (rejected)

User Q3 chose the advisor-rule integration. Less code, automatic surface coverage. Drift is fundamentally another "is something off?" signal — fits naturally in advisor's existing severity / dedup / notify pipeline.

### Option C — All-chunks-vs-goal (no recent window) (rejected)

User Q2 chose recent-N window. Whole-session average flattens out the drift signal — by the time the session has 100 chunks even substantial topic shifts produce only small avg-embedding deltas. Recent window keeps sensitivity high.

### Option D — Per-chunk drift score (one advisory per drifted chunk) (rejected)

Too noisy. One advisory per session per evaluator tick is enough; dedup window handles repeat fires.

## Consequences

- **`internal/advisor/types.go`** — 3 new Thresholds fields + `DefaultThresholds()` updates.
- **`internal/advisor/rules.go`** — `SessionDrift` type added to package + `Snapshot.DriftItems` field + `ruleGoalDrift` + `allRules` append + Korean prose template.
- **`internal/advisor/evaluator.go`** — `buildSnapshot` augmentation:
  - Pulls `knowledge.Store` chunks for each active session.
  - Embeds goal_text + recent N chunks' average vector via the wired Embedder.
  - Records `SessionDrift{Score, WorstChunk}` per session.
  - All-or-nothing: when embedder unavailable, DriftItems stays nil and ruleGoalDrift quietly returns nil.
- **`internal/config/config.go`** — 3 new keys + 2 validation rules.
- **`cmd/buddy/loadconfig.go`** — projects new keys onto `daemon.AdvisorMonitorConfig.Thresholds`.
- **`docs/cli-buddy-spec.md` §9 W7** → Done.
- **No new table / no migration / no new CLI subcommand / no new MCP tool / no TUI mode** — strict scope discipline.

## Verification

When W7-4 ships (v0.13.0):

```bash
# After buddy v0.13.0 install + buddy knowledge ingest --all run:

# Manual drift trigger: pick a session whose recent chunks talk
# about something different from goal_text, then:
buddy advise              # KindGoalDrift surfaces if cosine < 0.4
buddy daemon start        # advisorMonitor picks it up next tick

# Pipeline integration:
# - buddy notify status shows the drift advisory being dispatched
# - TUI ModeList banner shows it
# - TUI Usage pane advisory section shows it

go test -race -count=1 ./internal/advisor/...   # ~30 race-clean tests
make verify-versions                              # 5 sources on 0.13.0

# Whole-product status:
# B-1, B-3, B-4, C-1, C-2, C-3, C-5 ✅ (this release adds C-4 ✅)
# Only B-2 (user-paced production dogfood) remains for v1.0.0.
```

## Trigger to revisit

- User reports cosine over-fires (false-positive drift) → tune default threshold (0.4 → 0.3) or move to LLM-judge.
- Korean / mixed-language sessions show systematic drift bias → swap embedding model to a multilingual one (`paraphrase-multilingual-MiniLM-L12-v2`) via `BUDDY_EMBED_MODEL`.
- "Drift advisory triggers but goal_text was wrong from the start" → revisit fsLister goal_text extraction heuristic (W7-1 follow-on).
- v0.13.x users want CLI override for per-session threshold → add `buddy advise --drift-threshold` flag.

## References

- ADR-009 (cli buddy vision) — F2.D area definition.
- ADR-010 (v1.0.0 scope) — C-4 condition this ADR closes.
- ADR-012 (Session Monitor) — `sessions.goal_text` substrate.
- ADR-014 (knowledge retrieval) — embedder + cosine substrate.
- ADR-015 (Advisor) — pipeline reused.
- ADR-016 (Notification) — automatic dispatch reused.
