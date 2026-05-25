# ADR-019 — Plugin MCP exposure layer: hybrid distribution model (PATH-first + bundled fallback)

**Status**: Accepted (2026-05-26)
**Authors**: mhha (plugin track)
**Supersedes**: —
**Related**: ADR-003 (external attribution policy), ADR-006 (PROCEDURE form — new artifacts must conform), ADR-010 (v1.0 whole-product scope — this closes one of the four-asset charter gaps), [charter §2.4](../../two-tracks-charter.md) (plugin buddy four-asset promise), [`plugin-skills-engineering-flow.md`](../../plugin-skills-engineering-flow.md) §4 (C1 — Very High priority hole)
**Tags**: mcp-exposure, plugin-track, charter-asset-promise, distribution-model, hybrid-fallback, claude-plugin-spec, four-asset-completion

---

## 1. Context

`docs/two-tracks-charter.md` §2.4 records that a buddy plugin install is supposed to deliver four asset types — skill, MCP, agent, hook — but only skill is wired up today. The plugin/.claude-plugin/plugin.json manifest declares no `mcpServer` field, so installing the plugin gives the user 152 PROCEDURE.md files and 104 slash commands but **zero MCP tools**. Meanwhile, the buddy repository already ships a working MCP server: `cmd/buddy-mcp/main.go` is a standalone binary that speaks the Model Context Protocol over stdio and exposes nine tool surfaces from `internal/mcp/` (advise / analytics / doctor / feature / knowledge / notify / stats / usage / server — roughly 25 tools in total).

The gap is therefore the connecting layer between the existing Go server and the plugin manifest, not the server itself.

Four facts shaped this decision (verified 2026-05-26):

1. `cmd/buddy-mcp/main.go` exists and is built by the existing `make build` flow; it expects `BUDDY_DB` and optionally `BUDDY_ANALYTICS_BACKEND` / `BUDDY_ANALYTICS_DSN` in the environment and speaks stdio.
2. The current registration path is **manual**: `cmd/buddy/mcp_cmd.go` adds a `buddy mcp add/remove` subcommand that wraps `claude mcp add --scope <scope> buddy <binary-path>`. Users who do not know to run `buddy mcp add` get no MCP tools, even with the plugin installed.
3. The Claude Code plugin manifest accepts an `mcpServer` field (singular). The opsin plugin demonstrates the shape: `{"command": "...", "args": [...], "transport": "stdio"}` with `${CLAUDE_PLUGIN_ROOT}` available for path interpolation. The currently published buddy 1.1.0 cached manifest does not include the field.
4. Users of the buddy plugin split into two cohorts with different installation realities. Dogfood users (the maintainer and direct contributors) already have a `buddy` binary on their PATH from `make build`; new users coming through `/plugin install buddy` have no such guarantee.

The cost of the gap is concrete: every published skill in `plugin/skills/router/references/skill-catalog.md` that mentions `buddy mcp serve` (notify_tool, knowledge_tool, advise_tool, usage_tool error messages) is making a promise the install path does not keep.

---

## 2. Decision

Adopt a **hybrid distribution model** that prefers a user's existing `buddy` binary on PATH and falls back to a plugin-bundled platform-specific binary when none is found. Specifically:

- Add an `mcpServer` field to `plugin/.claude-plugin/plugin.json` pointing at a small shell launcher under `plugin/bin/`.
- Ship the launcher (`plugin/bin/buddy-mcp-launcher.sh`) as POSIX-only for now. Windows is out of scope per the existing charter lock-in.
- Ship four platform-specific bundled binaries at `plugin/bin/buddy-mcp-{darwin,linux}-{arm64,amd64}`. Each is the same `cmd/buddy-mcp` build cross-compiled for the target platform.
- The launcher resolves the actual executable in this priority order:
  1. `command -v buddy` on PATH (dogfood users get their freshly built local binary, which keeps `buddy mcp serve` and the plugin MCP server in lockstep).
  2. The bundled `plugin/bin/buddy-mcp-${OS}-${ARCH}` matching the detected platform.
  3. Failing both, exit non-zero with a clear stderr message referencing the install docs.
- Plugin version is bumped on next release (next minor — see ADR-011 milestone-driven cadence).

The implementation steps land in a separate session (see §5 below); this ADR locks the model and the launcher contract.

### Launcher contract (locked)

```sh
#!/usr/bin/env sh
# buddy-mcp-launcher.sh — resolve the MCP server binary at plugin invocation time.
# Priority: PATH-installed buddy > bundled platform binary > fail-with-message.
set -eu

if command -v buddy >/dev/null 2>&1; then
    exec buddy mcp serve "$@"
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    arm64|aarch64) ARCH=arm64 ;;
    x86_64|amd64)  ARCH=amd64 ;;
esac

BUNDLED="${CLAUDE_PLUGIN_ROOT:-}/bin/buddy-mcp-${OS}-${ARCH}"
if [ -x "$BUNDLED" ]; then
    exec "$BUNDLED" "$@"
fi

echo "buddy-mcp: no buddy binary on PATH and no bundled binary at ${BUNDLED:-(unset)}" >&2
echo "Install instructions: https://github.com/0xmhha/buddy#install" >&2
exit 127
```

### Manifest fragment (locked)

```json
"mcpServer": {
    "command": "${CLAUDE_PLUGIN_ROOT}/bin/buddy-mcp-launcher.sh",
    "transport": "stdio"
}
```

The launcher is the public contract; the bundled-binary set, environment variable wiring, and distribution mechanism for the bundle (git-tracked vs release artifact vs Git LFS) are implementation details settled in the next session.

---

## 3. Alternatives considered (rejected)

### Option A — bundled platform-specific binaries (no PATH preference)

Same four bundled binaries, no launcher: the manifest invokes the bundled binary directly with platform suffix resolution. Rejected because it forces dogfood users onto the bundled binary even when their freshly built `buddy` from `make build` is the most current. That subtle divergence (local code on PATH, but a stale bundled binary actually serving MCP requests) would surface as "I fixed it in `buddy` source but the MCP tool still misbehaves," with a long blame-search tail.

### Option B — install hook with download + bundled fallback

A `post-install.sh` hook fetches the matching binary from GitHub releases at install time, with the bundled binary acting as a fallback when the network fails. Rejected for three compounding reasons: post-install scripts violate user expectations about what a `/plugin install` does (they expect a file-copy, not a network call), they require trust the marketplace plugin loader cannot enforce, and the download step is one more failure mode to debug when MCP tools mysteriously stop working. The complexity does not pay for itself versus shipping the bundle inline.

### Option C — external binary requirement only

Keep the current pattern: users must `brew install buddy` (or otherwise put `buddy` on PATH) and the manifest invokes `buddy mcp serve` directly. Rejected because it fails new-user onboarding silently — the plugin would install cleanly, the user would never see an error, and only the next `/buddy:*` command that depended on MCP tools would surface the missing dependency. Onboarding silence is the worst-case failure mode for a plugin that explicitly markets itself as a single-install full-lifecycle product (`plugin.json` description).

The four-asset charter promise is meaningful precisely because it implies "no extra steps." Option C breaks that promise for the cohort that needs it most.

---

## 4. Consequences

### Positive

- Closes the H1/C1 hole identified in [`plugin-skills-engineering-flow.md`](../../plugin-skills-engineering-flow.md) §4 and lifts the four-asset charter promise from 25% (skill only) to 50% (skill + MCP).
- Dogfood users keep the property they have today: the locally built `buddy` binary serves their MCP requests, so changes to `cmd/buddy-mcp` are observable immediately without a plugin rebuild.
- New users get working MCP tools out of the box on macOS and Linux, on both arm64 and amd64.
- The launcher contract is small enough to audit on sight (twenty lines of POSIX shell), which keeps the trust surface narrow.

### Negative

- Plugin payload grows by roughly 38 MB (four ~9.5 MB Go binaries). For a marketplace plugin distributed through git this is non-trivial.
- The bundled binaries drift relative to the user's freshly built `buddy` if PATH lookup ever silently fails (e.g. a malformed PATH that omits the user's bin directory). The launcher's stderr fallback message is the only signal.
- Windows users get no MCP tools at all and the launcher does not attempt to bridge that gap.
- The repo's clone time and disk footprint grow proportionally with the bundled set. Git LFS or release-artifact-only distribution is one mitigation path, both deferred to the implementation session.

### Neutral

- The existing `buddy mcp add/remove` subcommand stays in place; it is now redundant for plugin users but still useful for non-plugin Claude Code users who want buddy-mcp registered into a different scope.
- ADR-011 (milestone-driven release cadence) governs when this lands in a published version; this ADR does not set a release schedule.

---

## 5. Verification

The decision is verified by:

1. **Test gate** — `scripts/test-router-wireup.sh` gains a check that `plugin/.claude-plugin/plugin.json` contains the `mcpServer` field and points at a launcher path under `plugin/bin/`. The launcher itself is executable. `make test-routing` still passes 10/10 (now 11/11 with the new check).
2. **Bundled-binary integrity** — the implementation session adds a `make build-mcp-bundled` target that cross-compiles all four binaries; CI runs it on PRs that touch `cmd/buddy-mcp/` or `internal/mcp/`.
3. **End-to-end dogfood** — after the plugin landing PR merges, a fresh `claude plugin install buddy@buddy` in a clean profile must surface the MCP tools (`/mcp` lists them) without the user running any extra commands. Recorded in `docs/notes/{date}-mcp-exposure-dogfood.md`.
4. **Charter alignment** — `two-tracks-charter.md` §2.4 row for "MCP" flips from "미정" / "미구현" to a concrete reference to this ADR and the launcher contract.

The implementation session is **not** considered complete until items 1, 3, and 4 are checked.

---

## 6. Trigger to revisit

Revisit this ADR when any of the following hold:

- Windows support enters the charter (currently lock-in v1.0+ defer per `HANDOFF.md` §4).
- Plugin payload size becomes a published constraint (e.g. marketplace caps or measured install slowness on slow connections).
- The Claude Code plugin manifest specification adds first-class platform-resolution syntax that makes the shell launcher obsolete.
- The cli buddy track's Wave 7 work surfaces a need for plugin-MCP and cli-MCP to share an active runtime (current design has them as separate processes that happen to share the same on-disk database).
- `cmd/buddy-mcp` grows a non-stdio transport (sse, http) where the launcher contract would need a second resolution step.

A revisit produces either a new ADR (in supersede chain) or a new entry in this ADR's history note. No "soft revisit by edit" — the supersede policy in `decisions/README.md` applies.

---

## 7. References

- [`docs/two-tracks-charter.md`](../../two-tracks-charter.md) §2.4 — four-asset promise (skill / MCP / agent / hook)
- [`docs/plugin-skills-engineering-flow.md`](../../plugin-skills-engineering-flow.md) §4 — C1 Very High priority entry, this ADR closes it
- [`docs/plugin-skills-inventory.md`](../../plugin-skills-inventory.md) §1 — four-asset table (currently shows MCP / agent / hook / rules all empty)
- [`cmd/buddy-mcp/main.go`](../../../cmd/buddy-mcp/main.go) — existing stdio MCP server
- [`cmd/buddy/mcp_cmd.go`](../../../cmd/buddy/mcp_cmd.go) — current manual `buddy mcp add/remove` subcommand (kept for non-plugin users)
- [`internal/mcp/`](../../../internal/mcp/) — nine tool surfaces (~25 tools)
- ADR-003 — attribution policy (launcher script is buddy-authored, no external adoption)
- ADR-006 — new artifacts under plugin/ must conform to PROCEDURE form lints (the launcher is shell, not a PROCEDURE, but the manifest change is subject to the bulk-allowlist gate)
- ADR-010 — v1.0 whole-product scope (this is a post-v1.0 enhancement that lifts the four-asset promise toward fulfilment)
- ADR-011 — milestone-driven release cadence (this ADR does not set a release date)
- opsin plugin manifest (external reference) — `${CLAUDE_PLUGIN_ROOT}` interpolation pattern and `mcpServer` field shape
