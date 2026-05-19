# ADR-011 — Release policy: milestone-driven version bumps, not commit-driven

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy + plugin track)
**Supersedes**: —
**Related**: ADR-004 (plugin version reset), ADR-010 (v1.0.0 scope — pending)

## Context

This session (2026-05-17 ~ 2026-05-19) shipped **six releases in three days**:

- v0.7.0 (2026-05-17) — W3-2 follow-on bundle + F5 purge + W3-6 example
- v0.7.1 (2026-05-18) — W3-2 follow-on finish + W3-4 chain control
- v0.7.2 (2026-05-18) — TUI hook-stats pane (A-3.2)
- v0.7.3 (2026-05-19) — ADR-006/007 + HANDOFF sync (doc-only)
- v0.7.4 (2026-05-19) — ADR-008 scope lock-in (doc-only)
- v0.7.5 (2026-05-19) — dogfood guide (doc-only)

The pattern that emerged: *accumulate 1-4 commits → release → push tag → repeat*. The motive was *frequent small patch* cadence — each release was *additively safe* (no breaking changes, all verify gates green), so the per-release blast radius was small.

The user's feedback (2026-05-19): **this is too frequent**. The intent should be *milestone-driven* — version bumps fire when a *named, planned milestone* is achieved, not whenever the working tree happens to be in a coherent state.

Three of the six releases (v0.7.3 / v0.7.4 / v0.7.5) were **doc-only** with no code path change. Publishing them as separate tags / GitHub releases created noise — a stargazer / contributor / dogfood-er sees six release entries in three days and cannot tell which one is *materially different*.

## Decision

**Version bumps fire only when a named, planned milestone is achieved.**

Concretely:

1. **Doc-only commits do NOT trigger a release.** ADR additions, HANDOFF.md sync, CHANGELOG entries about prior commits, dogfood guide additions — these accumulate in `[Unreleased]` and ride along with the next code-bearing release.
2. **A milestone is something nameable.** Examples that qualify:
   - "W3-2 follow-on cascade Done" (the entire bundle, not each follow-on item)
   - "W3-4 chain control shipped"
   - "Plugin v1.0.0 entry condition B-X closed" (each B-X is a named milestone)
   - "cli buddy F.2 vision area 1 (cross-session monitor) impl Done"
   - "Bug-fix for blocker finding from dogfood cycle-N"
3. **Examples that do NOT qualify**:
   - "Added ADR-006 only" — admin, ride along.
   - "Updated HANDOFF.md" — admin, ride along.
   - "Wrote a new docs/dogfood-guide.md" — admin (the guide describes existing behaviour; users get value from reading it, not from a release tag).
   - "Fixed a typo" — admin.
   - "Refactored an internal function with no external behaviour change" — admin.
4. **At least one of the following must be true for a version bump**:
   - **Code path** changes that a user can observe (CLI flag, MCP tool, spec field, runtime semantic, TUI key binding).
   - **Schema migration** that a user's DB will run on next launch.
   - **Bug fix** for a blocker / high-severity finding from dogfood.
   - **Security fix** (any severity).
5. **One release = one stated milestone in CHANGELOG.** The `## [X.Y.Z]` section opens with "What ships:" + a single milestone name. Doc-only commits ride along under "Also bundled (admin):" — they are not first-class line items.

## Alternatives considered

### Option A — Time-based cadence (rejected)

"Release every 2 weeks regardless of accumulation."

**Why rejected**: arbitrary timing creates releases without a *story*. Half the releases would be empty admin patches; half would group unrelated milestones. Worse than the status quo.

### Option B — Per-commit release (this session's de facto pattern) (rejected)

What just happened: 6 releases in 3 days. Tagged any time the working tree was coherent.

**Why rejected**: rejected by user as "too frequent". Creates noise, dilutes signal, makes "what changed since I last looked?" hard to answer.

### Option C — Milestone-driven (chosen)

Above. Bumps tied to nameable milestones; doc commits ride along.

**Why chosen**: aligns with the user's stated intent. Each release tells a story; doc-only churn doesn't compete for attention.

## Consequences

- **Historical note**: v0.7.0 / v0.7.1 / v0.7.2 each carried a named milestone (W3-2 cascade / W3-2 finish + W3-4 / W3-5 hook-stats) — those qualified. v0.7.3 / v0.7.4 / v0.7.5 carried only ADRs / docs — they would NOT have been released under this policy; the next milestone bump would have rolled them in.
- **Going forward**: `[Unreleased]` is allowed to accumulate doc / ADR / governance entries indefinitely until a code-bearing milestone fires the next release. CHANGELOG `[Unreleased]` becomes a *running ledger* of admin work between milestones.
- **No retroactive un-publishing**: v0.7.3 / v0.7.4 / v0.7.5 are already published and tagged. We don't yank them — but the next release (v0.8.0 or v0.7.6) will be the next *milestone-driven* one, not the next *commit-driven* one.
- **CHANGELOG style change**: each release section opens with the *milestone name* + 1-line description. Code changes listed first; admin (ADR / HANDOFF sync / doc add) listed under "Also bundled (admin):" sub-section.
- **CI / release.yml unchanged**: the tooling already supports any-time tag push. The policy is *when humans choose to push the tag*, not *what the workflow does*. No code change needed.

## Verification

This ADR is structural / governance, not testable via code. Verification is the next release's shape:

- Next release section in CHANGELOG must open with `## [X.Y.Z] — YYYY-MM-DD — <milestone name>` (the milestone name is the new convention).
- If the release contains zero code-path changes, the release is *not legitimate* under this policy and a future ADR amendment is required to explain why.
- Reviewer (or post-release self-audit) checks: "Was there a nameable milestone? Could I write the milestone name in one English noun phrase?" If yes → legitimate. If no → policy violated.

## Trigger to revisit

- If milestone-driven cadence makes a *critical bug fix* wait days for a release, amend to add a "blocker-bug exception" clause.
- If dogfood cycle reveals that doc-only updates need their own visible tag (e.g., contributors prefer to see "v0.7.5 ← guide added" in `gh release list`), amend with a "doc tag" sub-version scheme (e.g., v0.7.2-docs-1).
- For now: assume bundling doc into next code release is fine.

## References

- This session's release log: v0.7.0 ~ v0.7.5 (see `CHANGELOG.md` `[0.7.x]` sections).
- ADR-004 §2.2 (plugin version reset) — discusses *version namespace* but not *cadence*.
- User feedback 2026-05-19: "너무 잦으며, 처음부터 목적을 명확히 하고, 해당 목적이 달성되었을때, 버전을 올리는것이 맞다."
