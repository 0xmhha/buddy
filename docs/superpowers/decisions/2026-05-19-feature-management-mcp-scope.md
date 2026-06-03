# ADR-008 — `feature-management-mcp` scope: minimum-viable CRUD lock-in + naming split from external SaaS reference

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: D-3 / cycle-handoff §4.5 "feature-management-mcp deferred"
**Related**: ADR-005 (cli-buddy-spec lock-in), ADR-002 (roadmap × charter gap)

## Context

Two artefacts in the repo share a "feature-management" name and have caused recurring confusion in handoff docs (cycle-handoff §4.5 marked `feature-management-mcp` as "deferred (trigger: cli buddy W3-3 진입 시 — *현재 trigger 가능*)"):

1. **cli buddy feature registry (in-repo, shipped v0.3.0)**
   - `internal/feature/` package: `Upsert` / `Get` / `List` / `Delete` / `Search`.
   - `internal/mcp/feature_tool.go`: 5 MCP tools (`feature_list` / `feature_get` / `feature_upsert` / `feature_delete` / `feature_search`).
   - `migration v2`: `features` table (feature_id PK, name, summary, actors, acceptance_criteria, test_plan, status, updated_at).
   - `buddy feature` CLI subcommand for the same surface from the shell.
   - Skill catalog row `query-feature-registry`.

2. **`feature-management-saas-mcp` (external reference design, not in-repo)**
   - Referenced from `plugin/skills/design-billing-system/PROCEDURE.md`, `design-embedding-search/PROCEDURE.md`, `design-artifact-storage/PROCEDURE.md`, `design-mcp-server/PROCEDURE.md`.
   - A hypothetical / external SaaS product whose modules (Billing & Licensing / Embedding Knowledge DB / Patch Artifact Storage / etc.) the plugin's design-* skills help users design and implement.
   - **Not part of cli buddy's own runtime.** The skills produce design output; users implement that output in their own SaaS, not in buddy itself.

When cycle-handoff marked `feature-management-mcp` as a deferred item with "trigger: cli buddy W3-3 진입 시 — 현재 trigger 가능", it never specified WHICH of the two artefacts the trigger referred to or what "lock-in" meant for either of them. This ADR resolves both questions.

## Decision

### Part α — minimum-viable CRUD is the spec

The **shipped 5 MCP tools + 1 SQLite table + CLI subcommand** as of v0.3.0 ARE the cli buddy `feature-management-mcp` spec. There is nothing more to design, build, or lock-in at the cli buddy v1.0 horizon.

Concretely:

- The 5 MCP tools (`feature_list` / `feature_get` / `feature_upsert` / `feature_delete` / `feature_search`) form a stable surface. Their `mcp.Tool.Description` lines are the user-facing contract. Renaming or changing their signatures requires a future ADR.
- The `features` table schema (migration v2) is stable. Adding columns is backward-compatible (new migration); removing or renaming columns requires a future ADR.
- The `feature.Feature` struct in `internal/feature/types.go` is the canonical record shape exposed to both MCP and CLI consumers.

This closes the deferred-items entry "feature-management-mcp" from cycle-handoff §4.5. No code changes today.

### Part δ — naming split: cli buddy registry vs external SaaS reference

The two artefacts get distinct, non-overlapping names:

- **cli buddy feature registry** (in-repo, shipped): keep the existing `feature_*` MCP tool names + `buddy feature` CLI. Refer to it as the "cli buddy feature registry" in handoff / charter / spec docs. NEVER abbreviate to "feature-management-mcp" — that name will be reserved for the SaaS reference (see below).
- **`feature-management-saas-mcp`** (external reference, not in-repo): the existing `-saas-mcp` suffix is the disambiguator. The plugin's design-* skills already use the full name when they reference it; this ADR formalizes that "feature-management-mcp" alone is ambiguous and the suffix is required.

This is a *naming hygiene* fix — no skill body changes, no MCP tool renames, no breaking changes. Future docs that need to refer to either artefact must use the full disambiguated name.

### What ABOUT future expansion?

Anything beyond CRUD (semantic search, embedding index, dependency graph, versioning, status workflow, etc.) is a SEPARATE ADR triggered by *concrete user-paced dogfood signal*. Building those features without that signal would be over-engineering — the cli buddy persona is "친구 — silent default", and a sprawling registry violates that.

If a dogfood signal emerges (e.g., "I have 500 features and `feature_search` substring-match is too noisy"), the future ADR will propose the minimum addition that addresses the surfaced pain. Probably it will be embedding search via a new MCP tool, NOT a rewrite of the existing 5.

## Alternatives considered

### Option α only — minimum-viable lock-in, no naming clarification

Lock the cli buddy registry as-is, without addressing the naming ambiguity.

**Why rejected**: The ambiguity has *already* caused confusion (cycle-handoff §4.5 row that says "trigger: cli buddy W3-3 진입 시" referring to an unspecified artefact). Fixing the naming costs zero code; not fixing it perpetuates the confusion.

### Option β — full SaaS spec lock-in

Write a complete spec for `feature-management-saas-mcp` (billing + embedding + artifact + N more modules) inside this repo.

**Why rejected**:
- That SaaS is NOT cli buddy's responsibility — it's a reference design the design-* skills help users build for THEIR own SaaS.
- No dogfood signal for any of the modules. Building blind is over-engineering.
- Multi-cycle effort with low ROI; competes with B-2 (production dogfood) for attention.

### Option γ — defer indefinitely (status quo)

Keep the cycle-handoff "deferred" status, no ADR.

**Why rejected**:
- The user explicitly requested D-3 lock-in. Indefinite defer would be evasion.
- The naming ambiguity is a recurring source of confusion. Even a "we won't expand" decision adds value by killing the open question.

## Consequences

- **D-3 (cycle-handoff §4.5 row) is closed.** The deferred items row for `feature-management-mcp` should be removed from `docs/HANDOFF.md` / `docs/archive/2026-05-11-cycle-handoff.md` (HANDOFF was updated 2026-05-19 to drop this; cycle-handoff is a historical note, archived 2026-06-02, no edit needed).
- The cli buddy `feature_*` MCP surface is frozen at v0.3.0 shape. Adding a new field to `feature.Feature` or a new MCP tool requires a future ADR amending this one.
- The plugin's design-* skills' references to `feature-management-saas-mcp` are now formally OK to keep verbatim — they refer to the external SaaS reference, not to cli buddy itself.
- Future docs that say "feature-management-mcp" alone should be edited to either "cli buddy feature registry" or "feature-management-saas-mcp" depending on intent.

## Verification

```bash
# 1. The 5 MCP tools are registered:
$ grep -E 'Name: *"feature_' internal/mcp/feature_tool.go | wc -l
5

# 2. The `features` table is in migration v2 and unchanged:
$ awk '/version: 2/,/version: 3/' internal/db/migrations.go | grep -c 'CREATE TABLE features'
1

# 3. design-* skills still reference the external SaaS name (not cli buddy registry):
$ grep -rl 'feature-management-saas-mcp' plugin/skills/ | wc -l
4

# 4. cli buddy CLI still exposes `buddy feature ...`:
$ /tmp/buddy feature --help 2>&1 | head -3
```

## Trigger to revisit

A future ADR amends this one when:

- A concrete dogfood signal surfaces (e.g., "substring match is noisy at N features", "I need feature dependency graph for migration planning"). The amending ADR proposes the minimum surface addition.
- The plugin's design-* skills' SaaS reference design changes scope materially (e.g., a new module like `compliance-saas-mcp` joins the family). The amending ADR clarifies whether cli buddy adopts any of it.
- Plugin v2.0 (or some other major bump) is being designed and the registry shape needs a clean break.

Until one of those triggers fires, the cli buddy feature registry stays as-is.
