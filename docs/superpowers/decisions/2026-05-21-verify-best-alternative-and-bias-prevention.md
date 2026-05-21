# ADR-018 — verify-best-alternative rename + AI-bias prevention forced gate + decompose-blocker addition

**Status**: Accepted (2026-05-21)
**Authors**: mhha (plugin track)
**Supersedes**: —
**Related**: ADR-003 (superpowers attribution — sister `decompose-blocker` inspired-by), ADR-006 (B6 PROCEDURE form — new skill must conform), ADR-007 (router no cross-invocation state — decompose-blocker honours), ADR-010 (v1.0 scope — none of this changes entry conditions)

## Context

`SKILLS_ANALYSIS § A` (cross-skill engineering process audit, 2026-05-20) surfaced three independent gaps in the plugin track that were rejecting reviewers:

1. **`explore-design-variants` was being misread as design-only.** The maintainer themselves needed to dig into the body to remember the skill's actual purpose — preventing AI commit-to-first-answer bias on engineering decisions, applicable to architecture, naming, algorithm, API shape, data model, etc. A model reviewing the catalog made the same mistake. The title, catalog description, and persona paragraph all under-represented the intent.
2. **The bias-prevention machinery was dormant by default.** `explore-design-variants` had to be invoked manually. Whenever a user or AI made a `design-*` decision without thinking to call it, the bias guard did nothing. The most valuable property of the skill (catching first-answer bias in routine decisions) was paywalled behind user awareness.
3. **No skill covered code-work stuck states.** When a user said "I don't know where to start looking" or "no next action is visible," `diagnose-bug` assumed a reproducible symptom, `verify-best-alternative` assumed candidate options existed, and `concretize-idea` was a §1 lifecycle stage entry. The ad-hoc decomposition / entry-point search slot was empty.

These three failures were independent but reinforcing: a hidden bias guard couldn't help when users were stuck and didn't even know what to ask.

## Decision

Ship three coordinated changes as a single coherent intervention on the engineering-process cluster.

### 1. Rename `explore-design-variants` → `verify-best-alternative`

The new name leads with the **goal** ("verify the chosen path is actually best") instead of the **mechanism** ("explore design variants"). Intent-forward naming so future readers don't fall into the same trap. The persona paragraph is rewritten to state the failure mode being prevented — first-answer commit bias from the model's training distribution — and to list explicit application domains (architecture, naming, algorithm, API shape, data model, auth model, tenant model, secret strategy, ...). The catalog description gains an `[AI bias prevention]` marker so anyone scanning the catalog sees the purpose without opening the body.

Rename cascades through 9 files:
- `plugin/skills/verify-best-alternative/PROCEDURE.md` (was `explore-design-variants/PROCEDURE.md`)
- `plugin/commands/verify-best-alternative.md`
- `plugin/skills/router/references/skill-catalog.md`
- `plugin/skills/router/references/routing-rules.md`
- `plugin/skills/design-system/PROCEDURE.md` (sister-skill reference)
- `plugin/skills/define-tech-stack/PROCEDURE.md` (boundary clause)
- `plugin/skills/compose-safety-mode/PROCEDURE.md` (YELLOW invoke list)
- `docs/superpowers/specs/2026-05-06-lifecycle-orchestrator-architecture.md`
- `scripts/lint-skill-procedure.sh` (allowlist; resorts under `v`)

A historical-reference note in the renamed PROCEDURE.md preserves the prior name for grep continuity.

### 2. Wire `verify-best-alternative` as a forced gate across §3 design skills

Add a verification-gate checklist item to each leaf §3 design skill requiring one or more invocations of `verify-best-alternative` before the design can be declared complete:

- `define-tech-stack`
- `design-api-contract`
- `design-data-model`
- `design-event-schema`
- `design-auth-model`
- `design-tenant-model`
- `design-secret-management`

The orchestrator `design-system` already invokes the skill at Stage 2 topology decisions (updated to use the new name), so coverage is complete at both orchestrator and leaf levels. Each leaf checklist item explains *what* the guard protects against in that skill's domain (stack lock-in / API style choice / normalisation strategy / schema format / auth mechanism / tenant isolation model / secret-store choice).

§5 development skills are intentionally **excluded** from the gate. `build-with-tdd` and `refactor-with-rename-trace` execute decisions made upstream rather than make new ones; adding the gate there would be ceremony without payoff (and `build-with-tdd` already imposes its own discipline cycle that the second gate would interfere with).

### 3. New skill: `decompose-blocker`

A Cross-cutting Utility skill targeting the ad-hoc stuck state during code work — explicitly outside any single lifecycle phase. Procedure:

1. Classify the user's utterance into fact / guess / unknown three-way.
2. Pick decomposition axes (4D = where/when/what/why for systems work, binary search of code path for logic work, 5 Whys for root-cause work, fishbone for multi-factor work, trade-off matrix for option-selection work). Axis choice is **declared explicitly** to make the procedure reproducible.
3. Build a known-unknowns map with answer-cost per axis. "Unknown" is a first-class member, not a dead-end — often the most important first action candidate is "go gather the missing dimension."
4. Ask the user at most three forcing questions, starting from the cheapest high-information axis. Three is a hard cap; further questions belong to the next cycle.
5. Compress hypotheses to three or fewer, each with one-line evidence + one-line disproof.
6. Surface action candidates with a cost vs information-yield matrix, **in a single unit** (information-gathering / hypothesis-test / code-modification). Mixed-unit candidate lists make user selection impossible; the unit is declared and held constant within the cycle.
7. Confirm and dispatch — the cycle ends at action selection, and the next skill (`diagnose-bug` / `iterate-fix-verify` / `build-with-tdd` / `verify-best-alternative`) takes over execution.

Boundary clauses distinguish `decompose-blocker` from neighbouring skills: `diagnose-bug` assumes a repro, `verify-best-alternative` assumes candidate options, `concretize-idea` is a §1 lifecycle stage. `decompose-blocker` is for the case where none of those preconditions hold.

The skill is the first end-to-end dogfood case for `write-a-skill`. Step 2 RED dispatched a subagent with description only against a 504-timeout stuck scenario; the run surfaced five ambiguities (undeclared decomposition prior, mixed action-candidate units, unbounded question count, undistinguished too-wide vs too-narrow stuck shapes, undefined termination condition). Those became the five primary `§8 Anti-patterns` the body explicitly closes. Step 9 REFACTOR re-ran the same scenario with the finished body and the procedure passed in one cycle — anti-patterns blocked, default assumptions answered by explicit lines, output yaml schema accommodated the simulation result with no missing fields.

NOTICE updated with an `inspired-by` attribution per ADR-003 §2.4 acknowledging superpowers `brainstorming`'s forcing-question + design-thinking influence on `decompose-blocker`'s interaction pattern. No verbatim adoption — `decompose-blocker` targets a narrower scope (in-flight code-work stuck state vs idea-to-design pipeline) and uses different decomposition primitives (3-classification, 4D axes, cost-information matrix) that `brainstorming` does not have.

## Consequences

**Positive.**
- The bias-prevention property of `verify-best-alternative` is no longer paywalled behind user awareness. Every leaf §3 design decision now triggers the guard automatically as part of the verification gate. The maintainer no longer has to remember to invoke it.
- The catalog and routing layers correctly describe the skill's purpose to both future maintainers and to the model reading the catalog for dispatch. The "even the author forgot what this does" failure mode is closed.
- `decompose-blocker` fills the unbounded slot before `diagnose-bug` — stuck-state debugging now has a documented entry procedure. Step 2 RED + Step 9 REFACTOR was a useful dogfood of `write-a-skill` and surfaced no new defects in either.
- All three changes carry forward unchanged through commit splits and cross-session edits — the path is auditable.

**Negative.**
- §3 design skills now have one more checklist item to satisfy. For genuinely-cheap decisions (e.g. a stack with a single viable option due to constraints) the gate is overhead. The wording of each leaf checklist allows "1 or more invocations" so a single invocation that confirms the obvious answer satisfies the gate cheaply.
- `decompose-blocker` overlaps in surface area with `concretize-idea` and `diagnose-bug`. The boundary clauses in `§7 자매 스킬` are the primary defence; if users systematically misroute, expect a follow-up boundary-clarification PR.
- The rename creates a transitional period where users with cached `/buddy:explore-design-variants` muscle memory will hit a missing command. The historical-reference note in the renamed body and the lifecycle-architecture spec's `formerly` annotation cover documentation discoverability; the slash-command itself is gone.

**Neutral.**
- No new tables, no migrations, no new MCP tools, no new CLI subcommands, no new TUI surfaces. The intervention is entirely in skill-cluster documents + the lint allowlist.
- Routing-count rolls: `§5.4` common-tools `4 → 5`, plugin slash-command total `100 → 101`, skills total `149 → 150`.

## Alternatives Considered

**Description-only change (rejected).** Updating just the catalog description for `explore-design-variants` would fix the at-a-glance comprehension but leave the title and persona paragraph misleading. Both the maintainer and a reviewing model failed precisely because the title was the first signal they consumed. A description-only fix does not interrupt that priority order.

**No forced gate, rely on user awareness (rejected).** This is the status-quo failure mode. The skill's whole value proposition is fighting a default that the user/model has by definition — first-answer bias. Treating it as opt-in re-creates the gap.

**`brainstorm` as the new skill name instead of `decompose-blocker` (rejected).** `brainstorm` is too broad. It collides with the superpowers `brainstorming` skill's idea-to-design pipeline scope, which is not what the new skill does. `decompose-blocker` names the activity (problem decomposition when blocked) instead of the genre (creative exploration); the activity framing is more discriminating against neighbouring skills and uses buddy's existing `decompose-*` verb family.

**Split into three separate ADRs (rejected).** The rename, the gate wiring, and the new skill are independent decisions and could be three ADRs. They are bundled into one because the motivation is shared (the engineering-process audit surfaced all three together) and because the gate decision (#2) is meaningless without the rename (#1) being committed first — splitting risks intermediate inconsistent states in audit history.

## Verification

- `make test-routing` exits 0 on the post-merge tree (router dispatch, no conflicts).
- `plugin/skills/router/references/skill-catalog.md` grep for `verify-best-alternative` returns 1 hit (the catalog entry); for `decompose-blocker` returns 1 hit. `explore-design-variants` grep returns 0 hits in current files (only in this ADR and the historical-reference note inside the renamed PROCEDURE).
- Each of the 7 leaf §3 skill `§11 Verification gate` sections contains the `verify-best-alternative` checklist line.
- `decompose-blocker` was developed end-to-end through `write-a-skill` Step 1~10. RED + GREEN + REFACTOR cycle logs persisted in the subagent run that originally produced the body and the matching one-cycle-pass run that validated it.
- `scripts/lint-skill-procedure.sh` allowlist contains `verify-best-alternative` (alphabetically sorted between `sync-release-docs` and `write-changelog`) and `decompose-blocker` (between `critique-plan` and `define-feature-spec`).

## Trigger to revisit

- A user reports systematically routing to `decompose-blocker` when they meant `concretize-idea` or vice versa — boundary-clarification follow-up.
- A leaf §3 design skill gets shipped where the `verify-best-alternative` gate is genuinely inapplicable — refine the gate's "1 or more invocations" allowance to an explicit exemption clause.
- A `write-a-skill` execution surfaces an ambiguity in `decompose-blocker`'s body that this ADR's evidence (single-cycle Step 9 pass) failed to predict.
- `explore-design-variants` keeps appearing in user transcripts after 60 days post-rename — extend the historical-reference note or add a slash-command alias for the cached muscle memory.

## References

- `SKILLS_ANALYSIS.md § A` (cross-skill engineering process audit, `/Users/wm-it-22-00661/Work/github/study/ai/skill/`) — original gap identification (not in this repo)
- `plugin/skills/verify-best-alternative/PROCEDURE.md` — renamed body
- `plugin/skills/decompose-blocker/PROCEDURE.md` — new skill body
- `plugin/skills/write-a-skill/PROCEDURE.md` — `decompose-blocker` was authored through this skill
- ADR-003 `superpowers-attribution.md` — `decompose-blocker`'s NOTICE attribution follows §2.4 inspired-by classification
- Commit `182c219` — Prong 1-3 implementation
- Commit `98a91b1` — Prong 1-3 split-out partner (W4-6/W4-7 closure separated for clean attribution)
