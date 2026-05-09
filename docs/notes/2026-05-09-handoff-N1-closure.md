# N-1 closure handoff (2026-05-09)

> **Predecessor:** [`2026-05-08-handoff-context-bloat-investigation.md`](./2026-05-08-handoff-context-bloat-investigation.md) (left N-1 as the next-session top priority).
> **Status:** **N-1 closed — applied + verified at runtime.**
> **HEAD at closure:** `ca99762` on `origin/main`.

---

## 1. What was the request

> "buddy 플러그인이 추가한 skill 들이 매 prompt 마다 모두 로드되어 context 를 대량 사용하는 문제가 있어. indexing 을 통해 개선 방안을 검토해."

## 2. Outcome

| Question | Answer |
|----------|--------|
| Why does buddy bloat context? | Each command's `description` was loaded into baseline context for every prompt (verified via `code.claude.com/docs/en/skills`). 57 commands × ~85 chars/desc = ~4,887 chars / ~1,950 tokens/prompt. |
| Best lever found? | Frontmatter flag `disable-model-invocation: true` removes a skill's description from baseline while keeping `/buddy:<name>` user invocation working. |
| What was applied? | Quick Win Z: flag added to all 57 `plugin/commands/*.md`. Router skill (`plugin/skills/router/SKILL.md`) intentionally left as default (model-invocable) so natural-language routing still flows through it. |
| How is the decision recorded? | [ADR-001](../superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md) — context, decision, 4 alternatives, verification, future revisits. |

## 3. Commits this session (5)

| SHA | Subject |
|-----|---------|
| `2e94e04` | docs: replace machine-specific absolute paths with repo-relative refs |
| `7ce5c85` | docs(notes): record N-1 skill context bloat audit measurements |
| `350f2e3` | docs(commands): shorten top-10 longest descriptions (Quick Win A) |
| `ebce0b9` | docs(notes): record N-1 audit addendum verifying hypothesis A |
| `ca99762` | feat(commands): disable model invocation for all 57 dispatch commands |

## 4. Verification evidence

### 4.1 Tier-1 (static)

- `claude plugin marketplace add 0xmhha/buddy` → success
- `claude plugin install buddy@buddy` → installed at `~/.claude/plugins/marketplaces/buddy/plugin/` (v1.0.8, enabled)
- 57/57 installed `commands/*.md` files contain `disable-model-invocation: true` (`grep -L` count = 0)
- `diff -r plugin/commands/ ~/.claude/plugins/marketplaces/buddy/plugin/commands/` → identical
- Router SKILL.md untouched (still default, still model-invocable)

### 4.2 Tier-2 (runtime, after `/reload-plugins`)

After `/reload-plugins`, the Claude Code session's available-skills system-reminder lists **`buddy:router` only**. None of the 57 `/buddy:<name>` commands appear in the auto-loaded skill catalog.

This is the predicted outcome:

- ✅ Descriptions removed from baseline context (Quick Win Z works)
- ✅ Router remains model-invocable (natural-language routing entry point intact)
- ✅ User invocation path (`/buddy:<name>` via `/` autocomplete) is unaffected — these commands still appear in the autocomplete menu, just not in Claude's context

Estimated saving: ~1,725 tokens per prompt (4,319 chars after Quick Win A / ~2.5 chars per token). Buddy's share of the default 8,000-char `SLASH_COMMAND_TOOL_CHAR_BUDGET` drops from ~54% to ~0%.

## 5. Documentation surface change

- New: [`docs/notes/audits/2026-05-09-skill-context-bloat-audit.md`](./audits/2026-05-09-skill-context-bloat-audit.md) — measurement, hypothesis verification, Quick Win re-evaluation table.
- New: [`docs/superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md`](../superpowers/decisions/2026-05-09-buddy-commands-disable-model-invocation.md) — ADR-001.
- New: this note.
- Predecessor handoff (`2026-05-08-...`) intentionally **not** modified — it remains the historical record at the time it was written.

## 6. Next-session entry point

The original handoff §4 listed `N-1 → A-4 → Phase 5 residual → A-2 → B-1 → Cluster F/G`. With N-1 closed, the next live priority is **A-4 (plugin dogfood — install buddy on a real SaaS project)**.

A-4 unblocks the re-prioritization of the remaining 37 skills in §A-1 (now 29 after Phase 5 ext, see predecessor §3.1).

Recommended reading order for the next session:

1. This note (5 minutes).
2. Predecessor handoff §3 (state at v1.0.8) and §4.2 (residual cluster table). Skip §2 (it's now this note).
3. ADR-001 only if touching plugin command frontmatter conventions.

## 7. Deferred (low priority, not blocking)

| Item | Reason for deferral |
|------|---------------------|
| Quick Win C (body slim) | No baseline impact; only reduces cumulative cost across many invocations within one session. Re-evaluate if long-running session traces show measurable bloat. |
| CONTRIBUTING.md / lint enforcement of `disable-model-invocation: true` | Future-prevention only. Not urgent until first contributor PR adds a new command. |
| HANDOFF.md / README.md SSoT refresh (v1.0.5 → v1.0.8 + Quick Win Z) | Predecessor §8 already marked these stale; bundling with the next active doc edit is cheaper than refreshing now. |
| Module path drift (`wm-it-22-00661/buddy` → `0xmhha/buddy` in `go.mod`) | Predecessor §4.6 — separate Go-track concern. |

## 8. Fact-based summary

<Fact-based Answer>
- **Fact:**
  - Five commits pushed to `origin/main` ending at `ca99762`.
  - `claude plugin install buddy@buddy` succeeds; v1.0.8 enabled with all 57 commands flagged.
  - After `/reload-plugins`, only `buddy:router` is in the session's available-skills system-reminder; the 57 `/buddy:<name>` entries are absent (descriptions correctly removed from baseline).
  - ADR-001 records the decision with 4 rejected alternatives.

- **Your Opinion:**
  - **High prediction:** N-1 is fully resolved at the level of the original request. Token cost is no longer governed by `commands/*.md` description length, only by the router skill's single description.
  - **Mid prediction:** Quick Win C remains worth doing once invocation-cost measurement is feasible (multi-invocation session traces).
  - **Low prediction:** Some users on the latest Claude Code may already see different baseline budget behavior (e.g., dynamic 1% rule); re-measure if the budget mechanism changes.
  - **None.**
</Fact-based Answer>
