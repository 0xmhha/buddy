# ADR-009 — cli buddy vision expansion: AI-usage coaching beyond automation agents

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: cli-buddy-spec §1.3 narrow definition ("automation agent 관리" 단독)
**Related**: ADR-005 (cli-buddy-spec lock-in), ADR-010 (v1.0.0 scope — pending)

## Context

The user's stated F.2 alignment check on 2026-05-19 surfaced a missing dimension in the cli buddy identity that v0.1.0 ~ v0.7.5 work has *not* captured:

> "buddy 는 유저가 ai 기반으로(현재는 claude code) 작업을 수행하는 것들을 계속 모니터링하고, 유저의 행동 방식, ai 사용법등에 대해서, 분석하여 ai 를 더 잘사용할 수 있도록 서포트 해주는 툴이되어야 함. cli buddy 스스로는 claude code 기반의 다른 세션들도 지속적으로 모니터링 하면서, 어떻게 사용하는것이 더 유용한지등을 알림으로 인식할수 있도록 해줘야함. 즉, cli 기반의 agent 사용은 유저가 llm에 prompt 로 요청하고, 응답을 보고, 다음을 진행하는등의 작업을 수행하는데, 이 때 전달 받는 응답의 양이 상당하고, 긴 작업 시간을 수반하면, 처음의 목적에서 벗어나기도 하는등 유저(인간)가 상황판단이 어려워지는 경우가 발생하는데, 이를 해소할 수 있도록 지원한다."

cli-buddy-spec §1.3 (Accepted in ADR-005) defines cli buddy as having *four responsibilities*:

> agent 생성 / agent 실행 / agent 종료 / agent 설정 변경

The user's vision adds **five additional responsibilities** that the spec does not currently capture. These are not deferred / "future" features — they are *constitutive* of what cli buddy IS supposed to be.

## Decision

**cli buddy's identity is `automation agent management` + `AI-usage coaching`.** The `automation agent management` half is the four responsibilities already spec'd and shipped (W3-1 ~ W3-6). The `AI-usage coaching` half adds five new areas, named below as F2.A ~ F2.E.

### F2.A — Session Monitor

**Scope**: cli buddy passively observes *all Claude Code sessions* on this machine (this user's). Per-session it tracks lifecycle (start / end), token usage (input / output / cache hits), message count, wall-clock time spent, transcript path. Across sessions it produces a *user-level activity ledger*.

**Current fragments**:
- `internal/sessions/` package + `sessions` table (migration v3) already exists.
- Fields tracked: `id` / `pid` / `transcript_path` / `started_at` / `last_active` / `total_input_tokens` / `total_output_tokens`.

**Gap to user vision**: The current code is a passive registry. It does not *actively* attach to a session (no hook into Claude Code's session lifecycle), and the daemon's role in observation vs reporting is not specified.

**Future trigger**: Claude Code session-lifecycle hook (or equivalent observation channel) becomes stable enough for cli buddy to attach.

### F2.B — Usage Analysis

**Scope**: Given the activity ledger from F2.A, produce analytic summaries: *user's interaction patterns, friction points, time-of-day distribution, tool-call frequency, prompt-style trends*. These are computed locally (privacy-preserving — no upstream call).

**Current fragments**:
- `internal/analytics/` MCP tools 7개 (`analytics_query_funnel` / `cohort` / `ab_experiment` / `actor_failure` / `cost` / `slo_burn` / `feedback_corpus`) shipped in v0.3.0.
- These are SQL-adapter-backed, currently working against a *generic* analytics schema, not yet wired to F2.A's session ledger.

**Gap to user vision**: The 7 analytics tools assume a *product analytics* domain (funnels, cohorts, A/B). They are not yet repurposed for *AI-usage analysis* (token spend per session, message-length distribution, time-to-first-tool-call, hook-failure rate per skill, etc.).

**Future trigger**: F2.A produces enough activity data for analytics to be meaningful (~30 days of real usage).

### F2.C — Advisory

**Scope**: F2.B's analytics → *actionable recommendations*. Examples:
- "This hook fails 12% of the time; consider adding X to config."
- "Your last 5 conversations on `/buddy:concretize-idea` averaged 47 turns; the median is 12. Consider breaking into sub-tasks."
- "You've used 350K tokens today (your weekly average is 200K/day). Consider taking a break."

Format: short Korean prose, friend-tone, *not* a wall of metrics. The persona is "친구가 옆에서 한 마디 해주는".

**Current fragments**: *None*. The 7 analytics tools are read-only — they return data, not advice.

**Future trigger**: F2.B has enough analytic primitives to *derive* recommendations rather than just *report* data.

### F2.D — Drift Detection

**Scope**: Inside a *single long-running conversation* (hours / hundreds of turns), detect *semantic drift* — the current turn's topic vs the conversation's *original stated goal*. Surface as alert: "I notice we started working on X and are now spending most of our time on Y. Is that intentional?"

**Mechanism (provisional)**:
1. At session start (or first user message), extract / persist the *stated goal* (could be via a prompt to the LLM: "What is this session's goal in one sentence?").
2. Every N turns, sample current activity and ask: *does current direction match goal?*
3. If drift detected (low semantic similarity between recent N turns and original goal), surface the alert.

**Current fragments**: *None*. This requires LLM-driven semantic comparison; no precedent in the codebase.

**Future trigger**: F2.A session monitor is mature enough to know *which turn we're on* and has access to the *original goal*.

### F2.E — Notification

**Scope**: Deliver F2.C advisories + F2.D drift alerts via *human-perceptible channel*. Options:
- TUI banner (when `buddy tui` is open)
- Desktop notification (`osascript`-based on macOS, `notify-send` on Linux)
- Periodic terminal "status line" in the user's shell prompt
- Slack / email / webhook (opt-in, configured via `buddy config`)

The notification surface is *separate* from the advisory content — F2.C generates the message, F2.E transports it.

**Current fragments**: *None*. The existing `output.type: webhook` in agent runtime is a *step-output* transport, not a *user-notification* channel. The two are different concerns (machine-to-machine vs machine-to-human).

**Future trigger**: F2.C produces advisories that warrant out-of-band delivery (i.e., user wouldn't naturally see them by opening `buddy tui`).

## Alternatives considered

### Option A — Status quo: cli buddy = automation agent only

Keep the spec's four-responsibility definition. Implement F2.A ~ F2.E as *separate* products (e.g., `buddy monitor`, `buddy advise`, `buddy notify`).

**Why rejected**:
- The user's stated vision explicitly *includes* these as cli buddy responsibilities, not as separate products. Spinning them off would fragment the "친구" persona across multiple binaries.
- F2.A ~ F2.E share state (`~/.buddy/buddy.db`) and the persona / Korean tone with the existing cli buddy surface. Same product.

### Option B — One unified ADR + spec rewrite covering everything in v1.0.0

Mark all of F2.A ~ F2.E as v1.0.0 blockers. Refuse to publish v1.0 until all five are implemented.

**Why rejected**:
- ADR-010 (v1.0.0 scope) is a separate decision pending. This ADR (009) defines the *vision*, not the *v1.0.0 entry conditions*. F2.A ~ F2.E may end up split across multiple major versions.
- Vision documents that lock in *what we'll build* are stronger when separated from milestone documents (*when we'll ship it*).

### Option C — Vision ADR (chosen)

This ADR declares the five areas as *part of cli buddy's identity*, with current fragments + future triggers documented. Per-area design (scope, API, spec field, runtime) is *not* in this ADR; subsequent area-specific ADRs (e.g., ADR-{N} Session Monitor design) cover that as triggers fire.

**Why chosen**:
- Identity declaration is a one-time act. Per-area design can happen incrementally without re-opening the identity question.
- Aligns with how W3-1 ~ W3-6 worked: cli-buddy-spec set the scope, then each W3-x was a separate cycle with its own design + impl.

## Consequences

- **`cli-buddy-spec.md` §1.3** must add F2.A ~ F2.E to the responsibility list. The 4 → 9 expansion.
- **`cli-buddy-spec.md` §6** (charter alignment) must note that cli buddy now serves a *higher-level* purpose (AI-usage coaching) beyond agent management.
- **`cli-buddy-spec.md` §9** phase table gains 5 new W-tracks (W4 ~ W8 by some naming; concrete numbering decided in spec update). Currently W3-1 ~ W3-6 are Done; the new tracks start at 0%.
- **`BACKLOG.md` Wave 7** (new) is "cli buddy F2.A ~ F2.E vision impl" with per-area sub-tasks.
- **Per-area implementation ADRs** become future-ADR candidates in `docs/superpowers/decisions/README.md` "향후 ADR 후보":
  - ADR-{N} F2.A Session Monitor design (trigger: Claude Code session hook stable)
  - ADR-{N} F2.B Usage Analysis schema (trigger: F2.A produces 30 days of data)
  - ADR-{N} F2.C Advisory generation (trigger: F2.B analytics primitives stable)
  - ADR-{N} F2.D Drift Detection design (trigger: F2.A turn-level data available)
  - ADR-{N} F2.E Notification transport (trigger: F2.C advisories warrant delivery)
- **v1.0.0 scope** (ADR-010, pending) must decide whether all five F2 areas are v1.0.0 blockers, or whether some are deferred to v1.x / v2.0. Current strong-prior recommendation: at least F2.A (session monitor) is v1.0.0 — without it, "automation agent management" feels like a tool, not a *companion* (which is the persona).
- **No code change today**: this ADR is identity declaration; subsequent design ADRs decide impl shape.

## Verification

- `cli-buddy-spec.md` §1.3 lists 9 responsibilities (4 existing + 5 new).
- `BACKLOG.md` Wave 7 lists 5 sub-tasks (one per F2 area) with current-fragment links and future-trigger conditions.
- `docs/superpowers/decisions/README.md` "향후 ADR 후보" section adds 5 area-specific ADR candidates.

## Trigger to revisit

- After F2.A (the foundational session monitor) ships, this ADR may need refinement to clarify *how* F2.B / C / D / E build on it.
- If the user's daily Claude Code usage *increases* and the friction the vision describes ("응답 폭증", "긴 작업 시간", "목적 이탈") becomes pressing, the per-area design ADRs become high priority.

## References

- User feedback 2026-05-19 (F.2 alignment check): full quote in §1 Context.
- ADR-005 (cli-buddy-spec lock-in) — the spec section this ADR amends.
- `internal/sessions/` package + `sessions` table (migration v3) — F2.A fragment.
- `internal/analytics/` 7 MCP tools (v0.3.0) — F2.B fragment.
- `~/.claude/CLAUDE.md` "기억해" / "잊어" mechanic — adjacent concept (per-user state across sessions).
