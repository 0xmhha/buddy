# ADR-007 — Router does not maintain cross-invocation conversational state (supersedes B7 / B-4)

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: B7 / B-4 "router smart-skip session state" (cycle-handoff §4.4)
**Related**: ADR-004 §2.2 condition 4, ADR-006 (B-3 closure)

## Context

The plugin v1.0.0 entry condition list (ADR-004 §2.2) includes:

> **#4** — router smart-skip session state (B7) 해결 또는 supersede

The origin of B7 is the Cycle 1 dogfood result (`docs/notes/2026-05-10-dogfood-result-cycle-1.md`):

> B7 | conceptual | router smart-skip | validate-idea Q4 의 smart-skip 규칙 (L161 *이전 답이 이미 나중 질문 커버했으면 skip*) 이 router 단에선 *대화 메모리 의존* — `/buddy:run validate-idea` 재호출 시 이전 답 유실 | session state 메커니즘 도입 (별도 design 필요) — Cycle 3+ 후보.

Concretely: `plugin/skills/validate-idea/PROCEDURE.md` line 161 reads

> **Smart-skip:** 이전 답이 이미 나중 질문 커버했으면 skip.

This directive tells the LLM, when running the validate-idea flow, to skip a question whose answer the user already gave earlier. Within ONE conversation the LLM applies it naturally (it has the full prior turn context). Across DIFFERENT conversations (e.g., the user closes the chat and runs `/buddy:validate-idea` again the next day) the prior Q/A is not in context, so the LLM re-asks every question. B7 names this "session state loss".

The deferred resolution path was either "add a session-state mechanism" or "supersede". This ADR chooses supersede.

## Decision

**The router does not maintain cross-invocation conversational state, and skill-level smart-skip directives are in-conversation-only by design.**

Concretely:

1. The router's single responsibility is *dispatch*: load `<target>/PROCEDURE.md` and execute. It owns no Q/A history, no per-skill counters, no `~/.buddy/session-state.*` files.
2. Smart-skip in `validate-idea` (and any future skill with similar in-flight optimization hints) is interpreted by the LLM against the *current conversation* turn history. Cross-invocation persistence is the conversation-runtime's job (Claude Code's own context window, `/compact` decisions, `--continue`, etc.), not the router's.
3. The validate-idea PROCEDURE will be clarified to say "in-conversation only" so future readers don't misread the rule as cross-session.

## Alternatives considered

### Option A — Implement persistent session state (rejected)

Add a per-user / per-project Q/A persistence layer (e.g., `~/.buddy/session-state/<skill>/<project>.json` keyed by skill name + project root). Each skill checks it at start; smart-skip applies against persisted answers.

**Why rejected**:
- Scope creep: the router would gain a state-management surface that has nothing to do with dispatch.
- Storage hygiene: per-project Q/A persisted indefinitely raises GDPR-style questions (what's the TTL? does `buddy purge` cover it? can users opt out?).
- Marginal benefit: users who want to skip prior questions can paste their prior answer, say "skip Q1", or just answer the re-asked question in one word.
- Safer default: re-asking on a fresh invocation is the LESS surprising default — it doesn't assume the user's situation is unchanged from days/weeks/months ago.
- If a real need emerges, the existing v0.3.0 SQLite store (`~/.buddy/buddy.db`) and the deferred `feature-management-mcp` (cli buddy spec §6.2) are already the right home — it would be a SEPARATE skill ADR, not a router responsibility.

### Option B — Defer indefinitely (rejected)

Keep B7 / B-4 open as a v1.0.0 entry condition without resolution.

**Why rejected**:
- Indefinite deferral on a *conceptual* gap is worse than an explicit "we won't do this" decision. Future contributors waste time investigating an item that is in fact a non-goal.
- v1.0.0 entry conditions should be either active (working toward) or closed (with rationale). Indefinite-deferred items are technical debt in the planning surface itself.

## Consequences

- **Plugin v1.0.0 entry condition B-4 / B7 is closed.** The remaining entry conditions are:
  - B-1 cli buddy W3 cascade — Done (v0.7.x bundle)
  - B-2 production dogfood — user-paced, AI-unaccelerable
  - B-3 PROCEDURE B6 + `--strict` — Done (ADR-006, v0.7.3 patch line)
  - B-4 router smart-skip session state — **Closed by this ADR**

  Of the four, only B-2 remains open. v1.0.0 is gated on production dogfood now.
- `plugin/skills/validate-idea/PROCEDURE.md` smart-skip line is clarified: "이전 답이 *이 대화 안에서* 이미 나중 질문 커버했으면 skip" (in-conversation only).
- The future-ADR-candidate row "router smart-skip session state (B7)" in `docs/superpowers/decisions/README.md` is marked closed by this ADR.
- A future skill that needs cross-invocation state should propose its own ADR (probably via `feature-management-mcp` or a new skill-local store), NOT route through the router.

## Verification

The decision is structural, not testable via code. Verification is doc-coherence:

- `validate-idea` PROCEDURE line 161 updated with "(in-conversation only)" qualifier.
- ADR Index README updated: ADR-007 row added; the future-ADR-candidate B7 row marked closed.
- `docs/HANDOFF.md` track table no longer references "B7 router session state" as deferred — replaced by reference to this ADR.

## Trigger to revisit

If a future cycle's dogfood surfaces a *concrete* user pain ("I keep re-answering the same validate-idea questions every week and I want X persisted"), open a NEW ADR scoped to the persistence design — do NOT reopen B-7. The new ADR would belong to the skill or feature layer (e.g., feature-management-mcp), not the router.
