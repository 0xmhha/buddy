# ADR-012 — F2.A Session Monitor design (W7-1, v1.0 entry C-1)

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: —
**Related**: ADR-009 (cli buddy vision expansion — F2.A area), ADR-010 (v1.0.0 scope — C-1 condition), ADR-011 (release policy — milestone-driven)

## Context

ADR-009 declared F2.A Session Monitor as one of the five `AI-usage coaching` responsibility areas. Current fragments in the repo:

- `internal/sessions/sessions.go`: `Session` struct + `Lister` interface (skeleton only).
- `internal/db/migrations.go` v3: `sessions` table — fields `id` / `pid` / `transcript_path` / `started_at` / `last_active` / `total_input_tokens` / `total_output_tokens` / `total_cache_read` / `total_cache_create` / `last_offset`.
- A v0.2 control plane plan (referenced as `BUDDY-D-MULTI-IMPL` in `sessions.go` line 45) that was *deferred indefinitely* — the actual filesystem-scanning `fsLister` impl was never written.
- No CLI / TUI / daemon integration. Nothing currently writes to the `sessions` table.

→ F2.A is ~10% (schema + types only). The W7-1 cycle is what closes C-1.

## Decision

W7-1 ships **Session Monitor v1.0** with four design choices locked in:

### Q1 — Hybrid observation mechanism

- **SessionStart hook** (registered via `buddy install` already supports this) → calls a new entry point `buddy session register --id <ID> --transcript-path <path>` synchronously. Real-time registration of session existence.
- **fsLister** runs as the daemon's session-monitor goroutine (and on-demand by CLI) → tails each registered session's transcript JSONL, updates `total_*_tokens`, `last_active`, `last_offset`.
- Fallback: if hook is not installed (or fired before buddy install), `fsLister` periodic scan of `~/.claude/projects/*/` discovers the session and registers it on first observation. **buddy works for unhooked sessions too**.

Both paths converge on the same `sessions` table row (UNIQUE on `id`).

### Q2 — Schema extension: additive migration v5

Add 3 columns to `sessions`:

```sql
ended_at INTEGER,                  -- NULL until session is observed as ended;
                                   -- set when last_active is older than EndedThreshold (default 1h)
goal_text TEXT NOT NULL DEFAULT '', -- extracted from the first user message;
                                   -- F2.D Drift Detection will compare against this
metadata TEXT NOT NULL DEFAULT '{}' -- JSON escape hatch for future per-session fields
                                   -- (e.g., project_root, claude_version, custom tags)
```

`ended_at` is NULL while the session may still be active; the daemon sets it when `last_active` crosses `EndedThreshold` (default 1h, configurable via `buddy config set session-monitor.ended-threshold 30m`). Setting `ended_at` is *advisory*; if the session resumes activity, `ended_at` is cleared and the row stays single (no row split).

`goal_text` is populated by `fsLister` on first tail — the first user message in the transcript becomes the goal. F2.D (W7-4) will use this. If the first message is `/clear` or empty, `goal_text` stays empty.

`metadata` is a JSON object (`{}` default). Reserved for future per-session fields without further migrations.

### Q3 — CLI surface: list + show + purge

```bash
buddy session list                            # tabular: id (short) / project / started / last_active / tokens
buddy session list --all                      # include ended sessions
buddy session list --since 24h                # filter by last_active
buddy session show <id>                       # detail: full path, full metadata, goal_text, all counters
buddy session purge --before <duration>       # retention; ended sessions only
buddy session register --id <id> --transcript-path <path>   # hook entry point (internal)
```

Same pattern as `buddy agent {list,show,purge}`. `buddy session register` is hidden from `--help` (it's a hook-callable entry, not a user-facing command).

### Q4 — Daemon role: background poll + on-demand CLI

- **Daemon side**: existing `buddy daemon run` gains a new goroutine `sessionMonitor` that runs `fsLister` every `PollInterval` (default 30s, configurable). Discovers new sessions + updates tracked sessions + sets `ended_at` on stale.
- **CLI side**: `buddy session list/show` calls `fsLister` once on-demand if daemon is not running OR if `--refresh` flag is passed. If daemon is running, CLI reads from `sessions` table directly (no double-scan).
- Daemon enabled by default (matches v0.1 hook-monitor behavior). Disable: `buddy config set session-monitor.enabled false`.

## Minimum-viable scope for W7-1 ship (C-1 close)

| Item | In | Out (deferred) |
|------|----|----|
| `internal/sessions/fs_lister.go` (Hybrid impl: fsLister + JSONL tail + goal extraction) | ✅ | — |
| `internal/sessions/store.go` (Upsert / Get / List / SetEndedAt / Delete) | ✅ | — |
| migration v5 (3 new columns) | ✅ | — |
| `cmd/buddy/session_cmd.go` (list / show / purge / register) | ✅ | — |
| `buddy daemon` 의 sessionMonitor goroutine wiring | ✅ | — |
| `buddy config` 의 session-monitor.* keys (ended-threshold / poll-interval / enabled) | ✅ | — |
| Tests: race-clean (fsLister parsing JSONL fixtures, Store CRUD, lifecycle) | ✅ | — |
| TUI pane (`buddy tui` ModeSessions, `S` key) | ❌ | Wave 4 follow-on |
| Hook auto-registration via `buddy install` enhancement | ❌ | W7-1.1 follow-on patch (after dogfood signal) |
| Multi-machine session sync | ❌ | Wave 6 (D-2 멀티-머신 deferred) |

## Alternatives considered

### Q1 alt: fsLister only (rejected)
Polling-only feels safe but adds CC-format-change risk + no real-time start detection. Hybrid is strict superset.

### Q1 alt: Hook only (rejected)
Locks out unhooked users; small group but matters for first-time / quick-test users. Hybrid keeps fallback.

### Q2 alt: ended_at only (rejected)
Drift detection (F2.D / W7-4) needs `goal_text` anyway; not building it now means another migration soon.

### Q2 alt: no schema change (rejected)
`goal_text` cannot be derived from current fields; F2.D would block W7-4 on a schema change anyway.

### Q3 alt: list-only minimum (rejected)
Without `show`, users can't see the full transcript path / token totals; opacity makes session monitor feel half-baked.

### Q3 alt: full incl. live tail (rejected for v1)
`buddy session tail <id>` would mirror `buddy events --follow` but per-session. Useful but not v1-blocking; ship after dogfood confirms users want it.

### Q4 alt: on-demand only (rejected)
Without background poll, "session ended 2h ago" detection is impossible (no one to call). Background is essential for `ended_at` semantics.

### Q4 alt: opt-in flag (rejected)
The user's stated vision (ADR-009 F.2) is *passive observation* by default — opt-in defeats the persona ("친구가 옆에서 지켜보고 있다").

## Consequences

- **migration v5** lands in `internal/db/migrations.go`. Tests gate on `make verify-versions` + race-clean.
- **buddy install** stays unchanged for v1.0 of session monitor; hook auto-registration is a v1.0+ patch.
- **schema_version** bumps 4 → 5. Backward compat: existing v4 DBs run the v5 migration on next launch (additive, no data loss).
- **CLI surface** gains 4 commands (`list` / `show` / `purge` / `register`). `register` is hidden.
- **`buddy config` keys**: `session-monitor.enabled` (bool, default true) / `session-monitor.poll-interval` (duration, default 30s) / `session-monitor.ended-threshold` (duration, default 1h).
- **`buddy daemon`** lifecycle gains the sessionMonitor goroutine; daemon stop / restart honors it.
- **Test coverage target**: `internal/sessions/` package goes from 2 tests to ~15 (fsLister JSONL parsing, Store CRUD, ended_at lifecycle, goal extraction, race-clean concurrent observe + query).
- **Plugin v1.0.0 entry condition C-1 closes** when W7-1 ships and a real Claude Code session has been observed end-to-end (start → tail → ended) on the developer's machine.
- **Per ADR-011 release policy**: next release will be **v0.8.0** (milestone = W7-1 Session Monitor ship + C-1 close). Doc-only intermediate commits stay in `[Unreleased]`.

## Verification

When W7-1 ships:

```bash
# After buddy v0.8.0 install + buddy daemon start:
# (open a Claude Code session, send a message, wait 30s)

buddy session list                            # the active session appears
buddy session show <id>                       # goal_text populated from first user message
buddy session list --all                      # ended sessions appear after 1h inactive
buddy session purge --before 7d               # retention works

go test -race -count=1 ./internal/sessions/...   # ~15 race-clean tests
make verify-versions                             # 5 sources on 0.8.0
```

## Trigger to revisit

- If Claude Code transcript JSONL format changes incompatibly → patch `fsLister` parser, possibly new ADR if format shift is structural.
- If users complain about polling overhead → switch hook path to per-tool-call incremental update (W7-1.x patch).
- After F2.D (W7-4) ships → revisit if `goal_text` extraction logic should move to that layer.

## References

- ADR-009 (cli buddy vision expansion) — defines F2.A area.
- ADR-010 (whole-product v1.0.0 scope) — C-1 condition that this ADR will close.
- ADR-011 (release policy) — milestone-driven v0.8.0 framing.
- `internal/sessions/sessions.go` — current Session struct + Lister interface.
- `internal/db/migrations.go` v3 — current `sessions` table.
