# ADR-001 — Disable model invocation for all `plugin/commands/*.md`

> **Status:** Accepted
> **Date:** 2026-05-09
> **Deciders:** buddy maintainer
> **Tags:** plugin-architecture, context-cost, routing
> **Related:** [N-1 audit](../../notes/audits/2026-05-09-skill-context-bloat-audit.md), handoff §2 (`docs/notes/2026-05-08-handoff-context-bloat-investigation.md`)

## Context

`buddy` plugin v1.0.8 ships **57 slash commands** under `plugin/commands/*.md`. Every command file is a thin dispatcher whose body invokes the `router` skill (`plugin/skills/router/SKILL.md`) with a `target PROCEDURE` name; the router then loads the corresponding `plugin/skills/<target>/PROCEDURE.md` and executes it. Body shape is identical across all 57 files (router invocation block).

Per Claude Code's plugin/skills specification (verified 2026-05-09 via `code.claude.com/docs/en/skills`):

> "skill descriptions are loaded into context so Claude knows what's available, but full skill content only loads when invoked"

So every command's `description` field consumes baseline context tokens for **every prompt in every session**, regardless of whether the user invokes the command.

The N-1 audit measured the cost: 57 descriptions sum to ~4,300 chars (~1,725 tokens at ~2.5 chars/token), which occupies ~54% of the default 8,000-char `SLASH_COMMAND_TOOL_CHAR_BUDGET`. The cost is structural and recurring, not one-time.

The same spec also documents a frontmatter flag that flips the loading rule:

> "`disable-model-invocation: true` — Description not in context, full skill loads when you invoke. Only you can invoke the skill."

A command with this flag remains visible in the `/` autocomplete menu and remains user-invocable via `/buddy:<name>`, but its description is **not** placed in baseline context, and Claude can no longer auto-invoke it from natural language.

### Why this fits buddy's design

- buddy commands are user-explicit dispatch endpoints. The intended routing entry-point for natural language is the `router` skill, not individual commands.
- The router skill keeps its default (model-invocable) behavior, so Claude can still route conversational input through `router` when appropriate.
- No production buddy workflow relies on Claude auto-invoking a specific `/buddy:<name>` from natural language; the `chain`/`parallel`/`single` mode dispatch happens inside the router.

The previous behavior (descriptions in baseline) was therefore paying a recurring token cost for a capability — Claude auto-invoking individual commands — that buddy's design does not actually use.

## Decision

Add `disable-model-invocation: true` to the YAML frontmatter of **all 57 files** under `plugin/commands/*.md`.

The `router` skill at `plugin/skills/router/SKILL.md` is **not** modified — it remains model-invocable so Claude can route natural-language input through it.

Frontmatter shape after this change:

```yaml
---
description: <short tagline shown in / menu>
argument-hint: "<arg description>"
disable-model-invocation: true
---
```

This convention applies to **every future command added under `plugin/commands/`**, not just the current 57.

## Consequences

### Positive

- **~1,725 token savings per prompt** (estimate, ~4,300 chars / 2.5). Frees ~54% of the default `SLASH_COMMAND_TOOL_CHAR_BUDGET`, leaving budget for other plugins or for a wider context window.
- **Zero user-facing behavior change**: `/buddy:<name>` autocomplete and execution work identically. Users explicitly type the command via `/`, which is buddy's intended invocation path.
- **Aligns code with stated design**: makes the "router is the routing entry point" principle enforceable rather than aspirational.
- **Eliminates description-budget contention**: in environments running multiple plugins, buddy no longer eats most of the shared budget.

### Negative

- **Claude can no longer auto-invoke specific commands** from natural language (e.g., user says "I have an idea" → Claude can no longer fire `/buddy:concretize-idea` directly). Mitigation: the `router` skill remains model-invocable and is the correct entry point for natural-language routing. If a future requirement needs Claude to auto-fire individual commands, it must be raised as a new ADR superseding this one.
- **Convention enforcement is manual**. New commands added without the flag will silently re-introduce the cost. Mitigation: document in plugin contributing guidance and consider a CI lint check.

### Neutral

- Description text remains useful as an autocomplete tagline, so length still matters for the `/` menu UX even though it no longer affects baseline tokens.
- Body content (the router invocation block) still loads on invocation and persists in conversation. Body slim (Quick Win C in the audit) remains a separate, lower-priority optimization.

## Alternatives Considered

### Alternative 1: Description shortening only (Quick Win A/B)

Trim descriptions to ~50 chars each (~750 tokens saved). **Rejected** because it does not address the root cause (descriptions in baseline) and still leaks tokens to the budget. Some Quick Win A trims have already been applied (commit `350f2e3`) and are absorbed by this decision rather than rolled back; they continue to provide a leaner `/` menu.

### Alternative 2: Body slim only (Quick Win C)

Reduce each command body to a 1-line invocation. **Rejected as primary** because bodies are not in baseline context (per spec), so this doesn't reduce per-prompt cost. Retained as a future low-priority optimization for cumulative invocation cost within long sessions.

### Alternative 3: Apply `user-invocable: false` instead

The `user-invocable: false` flag hides skills from the `/` menu while keeping Claude able to invoke them. **Rejected** because it inverts buddy's actual usage pattern: users explicitly type `/buddy:<name>`, while Claude routes through `router`. We need the opposite restriction.

### Alternative 4: Raise `SLASH_COMMAND_TOOL_CHAR_BUDGET`

Set the env var so all descriptions fit without compromise. **Rejected** because it (a) requires every user to set the env var, (b) does not reduce token cost, only the truncation cliff, and (c) is host-controlled and could change without notice.

## Verification

Post-application checks (manual, 2026-05-09):

- All 57 `plugin/commands/*.md` contain `disable-model-invocation: true` (`grep -l '^disable-model-invocation: true$' plugin/commands/*.md | wc -l` → 57).
- Router SKILL.md unchanged.
- `/buddy:status` dry-run via `/` menu lists the command with its description tagline (autocomplete unchanged).
- A natural-language prompt that previously could trigger `/buddy:<name>` auto-invocation now goes through the router skill or returns a normal response — desired outcome.

## Future revisits

Re-evaluate this decision if any of the following becomes true:

- Claude Code spec changes how `disable-model-invocation` works (e.g., still indexes name+description even when set).
- buddy adds a workflow that legitimately needs Claude to auto-invoke a specific command from natural language.
- Token-budget pressure on the host shifts (e.g., `SLASH_COMMAND_TOOL_CHAR_BUDGET` default increases significantly).

## Index

This is **ADR-001** for buddy. Future ADRs continue numbering in this directory.
