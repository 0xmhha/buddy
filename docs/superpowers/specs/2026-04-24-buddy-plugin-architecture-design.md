# Buddy Plugin Architecture — Design Spec

- **Date**: 2026-04-24
- **Status**: Draft (awaiting user review before implementation plan)
- **Scope**: Bootstrap scaffold for a personal Claude Code plugin named `buddy`, distributed via GitHub + `curl | bash`.

## 1. Summary

Buddy is a Claude Code plugin that bundles the user's preferred commands, agents, skills, rules, hooks, and MCP servers into a single GitHub repo installable with one `curl` command. Components are added incrementally over time; the initial deliverable is the **distribution scaffold** itself (install / update / uninstall / doctor scripts + CI release pipeline + empty component directories ready to receive curated content).

The design prioritizes:

- **Zero-friction iteration** — `git pull` at `~/.buddy/` instantly reflects in `~/.claude/` via symlinks.
- **Safe install/uninstall** — metadata file tracks every deployed artefact so removal is exact.
- **Plugin system compatibility** — `.claude-plugin/plugin.json` follows Claude Code's native manifest convention.
- **Hybrid binary strategy** — externally-built MCP binaries committed to `vendor/`; self-built artefacts attached to GitHub Release.

## 2. Goals / Non-goals

**Goals**

1. `curl -fsSL https://raw.githubusercontent.com/<user>/buddy/main/install.sh | bash` installs everything.
2. Support the six component types (commands, agents, skills, rules, hooks, MCP servers).
3. Idempotent install / update / uninstall with dry-run mode.
4. Work on macOS (arm64/x86_64) and Linux (arm64/x86_64).
5. No runtime dependencies beyond `bash`, `git`, `jq`, and `curl` (pre-installed on most dev machines).
6. Self-verifying: every install runs `doctor.sh` to confirm integrity.

**Non-goals**

1. Windows support (WSL is expected if Windows is a target).
2. Multi-tenant / per-project install (single global install to `~/.claude/`).
3. Automated component migration from legacy SuperClaude v3 install (already cleaned out).
4. Pre-populated component set in the initial scaffold — components are added in follow-up work.
5. Rollback automation beyond manually re-running `update.sh` against an older tag.

## 3. Architecture Overview

```
┌────────────────────────────────────────────────────────────────────┐
│  curl | bash (entry point)                                          │
│       │                                                             │
│       ▼                                                             │
│  ~/.buddy/                 (git clone --depth=1 --branch <ref>)     │
│  ├── plugin/               (what gets deployed)                     │
│  │   ├── .claude-plugin/plugin.json                                 │
│  │   ├── commands/  agents/  skills/  rules/  hooks/hooks.json      │
│  │   └── mcp/servers.json  (build artefact, gitignored)             │
│  ├── vendor/<name>/bin/<platform-specific>  (committed MCP bins)    │
│  ├── lib/bash/             (install helpers)                        │
│  └── scripts/ install.sh uninstall.sh update.sh doctor.sh           │
│       │                                                             │
│       │  symlinks + JSON merges                                     │
│       ▼                                                             │
│  ~/.claude/                                                         │
│  ├── commands/buddy/ → ~/.buddy/plugin/commands/                    │
│  ├── agents/buddy-*.md → ~/.buddy/plugin/agents/buddy-*.md          │
│  ├── skills/buddy-*/SKILL.md → ~/.buddy/plugin/skills/*/SKILL.md    │
│  ├── rules/buddy-*.md → ~/.buddy/plugin/rules/buddy-*.md            │
│  ├── settings.json                  (hooks merged, id="buddy:*")    │
│  ├── .mcp.json                      (servers merged, key="buddy-*") │
│  └── .buddy-metadata.json           (install ledger for uninstall)  │
└────────────────────────────────────────────────────────────────────┘
```

The core insight: `~/.buddy/` is the single source of truth. `~/.claude/` holds symlinks + two JSON files that have been patched with buddy-tagged entries. Uninstall consults `~/.claude/.buddy-metadata.json` to reverse those patches precisely.

## 4. Directory Layout (Repo)

```
buddy/
├── plugin/                                  # Deployable payload (self-contained unit)
│   ├── .claude-plugin/
│   │   └── plugin.json                      # Authoritative manifest (paths relative here)
│   ├── commands/                            # /buddy:* slash commands
│   ├── agents/                              # buddy-* subagents
│   ├── skills/                              # buddy-*/SKILL.md (each with optional scripts/references/assets)
│   ├── rules/                               # buddy-*.md
│   ├── hooks/
│   │   └── hooks.json                       # Pre/Post/Stop hook declarations
│   └── mcp/
│       └── servers.json                     # GITIGNORED — built from vendor/*/mcp.config.json
├── vendor/                                  # Externally built binaries, committed
│   ├── README.md                            # Convention: every vendor entry has UPSTREAM.md
│   └── <mcp-name>/
│       ├── UPSTREAM.md                      # Source repo URL, commit SHA, build steps, license
│       ├── bin/
│       │   ├── <name>-x86_64-linux
│       │   ├── <name>-aarch64-linux
│       │   ├── <name>-x86_64-darwin
│       │   └── <name>-aarch64-darwin
│       └── mcp.config.json                  # Fragment merged into plugin/mcp/servers.json
├── lib/                                     # Reusable libraries
│   ├── bash/
│   │   ├── log.sh                           # info/warn/err formatters
│   │   ├── platform.sh                      # os/arch detection + binary path resolution
│   │   ├── json.sh                          # jq-based merge/patch/extract helpers
│   │   └── symlink.sh                       # idempotent symlink creation / verification
│   └── README.md
├── scripts/
│   ├── install.sh                           # Core deploy logic (called by root install.sh)
│   ├── uninstall.sh                         # Metadata-driven removal
│   ├── update.sh                            # git fetch → tag checkout → install.sh --force
│   ├── doctor.sh                            # Health check (7 sections)
│   ├── validate-manifest.sh                 # JSON schema check for plugin.json
│   ├── build-mcp-aggregate.sh               # Merge vendor/*/mcp.config.json → plugin/mcp/servers.json
│   └── release/
│       ├── build.sh                         # Self-built binary producer (no-op in v0.1.0)
│       ├── checksums.sh                     # SHA256SUMS generator
│       └── changelog-extract.sh             # Extract <TAG> section from CHANGELOG.md
├── install.sh                               # curl entry point: git clone + delegate to scripts/install.sh
├── uninstall.sh                             # Thin wrapper delegating to scripts/uninstall.sh
├── VERSION                                  # semver, synced with git tag
├── CHANGELOG.md                             # Keep-a-Changelog format
├── README.md                                # (Existing, extended with plugin usage)
├── LICENSE                                  # (Existing)
├── docs/                                    # (Existing; this spec lives under docs/superpowers/specs/)
├── src/                                     # (Existing user code, unchanged)
└── .github/
    └── workflows/
        ├── ci.yml                           # shellcheck, jq, manifest lint, install/uninstall smoke
        └── release.yml                      # Tag push → multi-platform build + GitHub Release
```

### 4.1 Why `plugin/` as a subfolder

Wrapping deployable content in `plugin/` cleanly separates it from `vendor/`, `lib/`, `scripts/`, and the existing `src/` / `docs/`. The manifest at `plugin/.claude-plugin/plugin.json` uses clean relative paths (`./commands/`, `./agents/`, etc.) rather than `../plugin/commands/`. Claude Code's native plugin loader can therefore read `plugin/` as a self-contained unit if that becomes useful in the future.

### 4.2 Skill layout pattern

Each skill is deployed as a **real directory** at `~/.claude/skills/buddy-<name>/` with a **symlinked SKILL.md** pointing back to `~/.buddy/plugin/skills/buddy-<name>/SKILL.md`. Rationale:

- If a skill adds `scripts/`, `references/`, `assets/` subdirs over time, they live in the source (`~/.buddy/plugin/skills/buddy-<name>/…`) and remain accessible via Claude Code's skill discovery (which reads SKILL.md and resolves sibling paths).
- The real directory at the target (not a directory symlink) lets runtime-generated content (hook outputs, per-invocation state) sit alongside the symlinked SKILL.md without polluting the source repo.

## 5. Component Deploy Mapping

### 5.1 Per-type strategy

| # | Component | Source | Destination | Strategy | Conflict policy |
|---|-----------|--------|-------------|----------|-----------------|
| 1 | commands  | `~/.buddy/plugin/commands/` | `~/.claude/commands/buddy/` | `dir-symlink` (single link) | Refresh link target if already present |
| 2 | agents    | `~/.buddy/plugin/agents/*.md` | `~/.claude/agents/<name>` | `file-symlink` per file | Halt on existing same-named file; `--force` overrides |
| 3 | skills    | `~/.buddy/plugin/skills/<skill>/SKILL.md` | `~/.claude/skills/<skill>/SKILL.md` | `dir-with-symlinked-SKILL.md` | Halt on existing same-named skill; `--force` overrides |
| 4 | rules     | `~/.buddy/plugin/rules/*.md` | `~/.claude/rules/<name>` | `file-symlink` per file | Halt on existing same-named file |
| 5 | hooks     | `~/.buddy/plugin/hooks/hooks.json` | `~/.claude/settings.json` (`hooks` key) | `json-merge` with `id: "buddy:*"` tagging | Replace buddy-tagged entries, preserve others |
| 6 | mcp       | `~/.buddy/plugin/mcp/servers.json` (built from `vendor/*/mcp.config.json`) | `~/.claude/.mcp.json` (`mcpServers` key) | `json-merge` with `buddy-*` server-name prefix | Halt on same-named server not tagged buddy-* |

### 5.2 `buddy-` naming convention

| Type | Example | Rationale |
|------|---------|-----------|
| commands  | `~/.claude/commands/buddy/ship.md` invoked as `/buddy:ship` | Directory namespace auto-isolates |
| agents    | `buddy-planner.md` invoked as `@buddy-planner` | Filename is the invocation name; prefix mandatory |
| skills    | `buddy-tdd/SKILL.md` invoked as `/buddy-tdd` | Directory name is the invocation name; prefix mandatory |
| rules     | `buddy-coding-style.md` | Coexists with existing `coding-style.md`, etc. |
| hooks     | `{"id": "buddy:pretooluse-lint"}` | JSON identifier convention for precise uninstall |
| mcp       | `"buddy-serena": { ... }` | `.mcp.json` conflict prevention |

### 5.3 `install.sh` processing order

```
1. Preflight
   - Verify ~/.claude/ exists (Claude Code installed)
   - Verify git + jq binaries present
   - Detect platform via lib/bash/platform.sh

2. Deploy (dependency order)
   a. rules       (text references, no side effects)
   b. commands    (directory symlink)
   c. agents      (file symlinks)
   d. skills      (dir + SKILL.md symlink)
   e. mcp         (build plugin/mcp/servers.json from vendor/*/mcp.config.json with injected abs paths; merge into ~/.claude/.mcp.json)
   f. hooks       (last — hooks may reference artefacts from a-e)

3. Record
   - Write ~/.claude/.buddy-metadata.json

4. Verify
   - Invoke scripts/doctor.sh; exit non-zero on failure
```

### 5.4 Binary selection logic (mcp step)

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')   # darwin / linux
ARCH_RAW=$(uname -m)                           # x86_64 / arm64 / aarch64
ARCH=$(case "$ARCH_RAW" in arm64|aarch64) echo aarch64 ;; *) echo x86_64 ;; esac)

BIN="$HOME/.buddy/vendor/$NAME/bin/$NAME-$ARCH-$OS"
[ -x "$BIN" ] || { warn "no binary for $NAME on $OS/$ARCH; skipping"; continue; }

jq --arg bin "$BIN" \
   --arg name "buddy-$NAME" \
   '.mcpServers[$name].command = $bin' \
   "$HOME/.claude/.mcp.json" > "$TMP" && mv "$TMP" "$HOME/.claude/.mcp.json"
```

Missing binary for current platform → skip that MCP server with a warning. Overall install does **not** fail (partial installs are permitted).

### 5.5 Install flags

| Flag | Effect |
|------|--------|
| `--force` | Overwrite conflicting targets instead of halting |
| `--copy` | Replace symlinks with `cp -R` (breaks the `~/.buddy/` live link) |
| `--dry-run` | Log planned actions without touching filesystem |
| `--only=<types>` | Install only specified component types (e.g., `--only=skills,rules`) |
| `--yes` | Auto-confirm all prompts (CI-friendly) |

## 6. Manifests & Metadata

### 6.1 `plugin/.claude-plugin/plugin.json` (static, committed)

```json
{
  "name": "buddy",
  "version": "0.1.0",
  "description": "Personal Claude Code plugin — commands, skills, agents, rules, hooks, MCP",
  "author": {
    "name": "<user>",
    "url": "https://github.com/<user>/buddy"
  },
  "commands": "./commands/",
  "agents": "./agents/",
  "skills": "./skills/",
  "rules": "./rules/",
  "hooks": "./hooks/hooks.json",
  "mcpServers": "./mcp/servers.json"
}
```

Notes:

- Paths are relative to the manifest's parent (`plugin/`).
- `version` must match repo-root `VERSION` (CI enforces).
- The `rules` field is not known to be a standard Claude Code manifest field. If the native plugin loader ignores it, behavior is unaffected; rules deployment is handled entirely by `scripts/install.sh`.
- `mcpServers` points to `./mcp/servers.json`, which is a build artefact (gitignored) assembled by `scripts/build-mcp-aggregate.sh`.

### 6.2 `plugin/mcp/servers.json` (build output, gitignored)

Assembled per-machine by walking `vendor/*/mcp.config.json` and injecting platform-appropriate absolute binary paths.

```json
{
  "mcpServers": {
    "buddy-serena": {
      "command": "/Users/<u>/.buddy/vendor/serena/bin/serena-aarch64-darwin",
      "args": ["stdio"],
      "env": {}
    }
  }
}
```

### 6.3 `~/.claude/.buddy-metadata.json` (runtime install ledger)

```json
{
  "schema_version": 1,
  "buddy_version": "0.1.0",
  "installed_at": "2026-04-24T12:34:56Z",
  "updated_at": "2026-04-24T12:34:56Z",
  "source_path": "/Users/<u>/.buddy",
  "install_mode": "symlink",
  "platform": { "os": "darwin", "arch": "aarch64" },
  "components": {
    "commands": {
      "strategy": "dir-symlink",
      "entries": [
        { "target": "/Users/<u>/.claude/commands/buddy",
          "source": "/Users/<u>/.buddy/plugin/commands" }
      ]
    },
    "agents": {
      "strategy": "file-symlink",
      "entries": [
        { "target": "/Users/<u>/.claude/agents/buddy-planner.md",
          "source": "/Users/<u>/.buddy/plugin/agents/buddy-planner.md" }
      ]
    },
    "skills": {
      "strategy": "dir-with-symlinked-SKILL.md",
      "entries": [
        { "target_dir": "/Users/<u>/.claude/skills/buddy-tdd",
          "skill_md_symlink": "/Users/<u>/.claude/skills/buddy-tdd/SKILL.md",
          "source": "/Users/<u>/.buddy/plugin/skills/buddy-tdd/SKILL.md" }
      ]
    },
    "rules": {
      "strategy": "file-symlink",
      "entries": [
        { "target": "/Users/<u>/.claude/rules/buddy-coding-style.md",
          "source": "/Users/<u>/.buddy/plugin/rules/buddy-coding-style.md" }
      ]
    },
    "hooks": {
      "strategy": "json-merge",
      "target_file": "/Users/<u>/.claude/settings.json",
      "ids": ["buddy:pretooluse-lint"]
    },
    "mcp": {
      "strategy": "json-merge",
      "target_file": "/Users/<u>/.claude/.mcp.json",
      "server_names": ["buddy-serena"]
    }
  }
}
```

Principles:

- `schema_version` allows forward-compatible uninstallers.
- `strategy` values are a closed vocabulary (`dir-symlink`, `file-symlink`, `dir-with-symlinked-SKILL.md`, `json-merge`). Uninstall reverses operations per strategy.
- hooks/mcp record **only identifiers**, not full entries — the current `settings.json` / `.mcp.json` is the source of truth at removal time.

## 7. Update / Uninstall / Doctor

### 7.1 Update (`scripts/update.sh`)

```
1. Preflight: metadata present, ~/.buddy is a git repo, git binary available.
2. Fetch: cd ~/.buddy && git fetch --tags; determine LATEST vs CURRENT (VERSION file).
3. Diff preview: list added/removed components vs prior plugin tree.
   Prompt for confirmation unless --yes.
4. Apply: git checkout <LATEST_TAG>; scripts/install.sh --force.
5. Post: metadata.updated_at + buddy_version refreshed; doctor.sh invoked.
```

Key rules:

- **Only tagged releases** are checked out — never a moving branch ref.
- **Removed components** are detected by diffing metadata entries against the new plugin tree and unlinked automatically.
- **Rollback** is manual: `cd ~/.buddy && git checkout <older_tag> && scripts/install.sh --force`.

### 7.2 Uninstall (`scripts/uninstall.sh`)

```
For each component in metadata.components:
  - file-symlink / dir-symlink:
      if target is a symlink AND readlink(target) == entry.source → unlink
      else warn + skip
  - dir-with-symlinked-SKILL.md:
      unlink skill_md_symlink if valid; rmdir target_dir if now empty, else warn
  - json-merge (hooks):
      read target_file; remove array entries whose id ∈ metadata.ids; atomic write
  - json-merge (mcp):
      read target_file; remove mcpServers keys ∈ metadata.server_names; atomic write

Then delete ~/.claude/.buddy-metadata.json.
Prompt to remove ~/.buddy/ (y/N).
```

Safety:

- Never unlink unless `readlink` confirms original source. Protects manually-edited files.
- `rmdir` only on empty skill target directories. Preserves user-added content.
- `--force` bypasses verification (explicit opt-in; non-recoverable).
- `--dry-run` and `--keep-repo` flags supported.

### 7.3 Doctor (`scripts/doctor.sh`)

Seven check sections with ✅ / ⚠️ / ❌ output:

1. **Environment** — Claude Code installed, git + jq present, platform detected.
2. **Repo state** — `~/.buddy` is a git repo; checked-out tag matches `VERSION`.
3. **Metadata integrity** — metadata exists, schema_version compatible, buddy_version ↔ VERSION match, source_path resolves.
4. **Component health** — for each metadata entry: target exists, symlink valid, SKILL.md parseable (skills), file readable (rules).
5. **Hooks health** — `settings.json` contains the expected buddy-tagged hook IDs; referenced commands exist.
6. **MCP health** — `.mcp.json` contains expected buddy-prefixed servers; binary paths resolve and are executable.
7. **Summary** — counts + exit code (0 all-pass, 1 warnings, 2 failures).

Options:

- `--verbose` (per-check detail)
- `--only=<sections>` (subset)
- `--fix` (auto-repair obvious issues — broken symlinks, missing executable bit)

## 8. Release & CI Pipeline

### 8.1 Entry points

```bash
# Latest main (dev convenience)
curl -fsSL https://raw.githubusercontent.com/<u>/buddy/main/install.sh | bash

# Pinned to a release tag
curl -fsSL https://raw.githubusercontent.com/<u>/buddy/v0.1.0/install.sh | bash

# main's install.sh but clone specific tag
curl -fsSL https://raw.githubusercontent.com/<u>/buddy/main/install.sh | BUDDY_REF=v0.1.0 bash
```

`install.sh` ref resolution:

```bash
BUDDY_REF="${BUDDY_REF:-main}"
BUDDY_REPO="${BUDDY_REPO:-https://github.com/<user>/buddy.git}"

# If BUDDY_REPO points to a local path, use it as source in-place (CI mode).
# Otherwise clone.
if [ -d "$BUDDY_REPO" ]; then
  ln -sfn "$BUDDY_REPO" "$HOME/.buddy"
else
  git clone --depth=1 --branch "$BUDDY_REF" "$BUDDY_REPO" "$HOME/.buddy"
fi
"$HOME/.buddy/scripts/install.sh" "$@"
```

### 8.2 `ci.yml` (PR + push to main)

```yaml
on:
  push: { branches: [main] }
  pull_request: { branches: [main] }

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          find scripts lib/bash install.sh uninstall.sh -type f -name '*.sh' -print0 \
            | xargs -0 shellcheck
      - run: find plugin vendor -name '*.json' -exec jq empty {} \;
      - run: scripts/validate-manifest.sh plugin/.claude-plugin/plugin.json

  smoke-install:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          mkdir -p /tmp/fake-home/.claude
          HOME=/tmp/fake-home BUDDY_REPO=$(pwd) scripts/install.sh --dry-run
      - run: HOME=/tmp/fake-home BUDDY_REPO=$(pwd) scripts/install.sh --yes
      - run: HOME=/tmp/fake-home /tmp/fake-home/.buddy/scripts/doctor.sh
      - run: HOME=/tmp/fake-home /tmp/fake-home/.buddy/scripts/uninstall.sh --yes
```

### 8.3 `release.yml` (tag push `v*.*.*`)

```yaml
on:
  push:
    tags: ['v*.*.*']

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: version ↔ tag
        run: |
          TAG="${GITHUB_REF#refs/tags/v}"
          FILE=$(cat VERSION)
          [ "$TAG" = "$FILE" ] || { echo "tag v$TAG != VERSION $FILE"; exit 1; }
      - name: changelog entry
        run: grep -q "^## \[${GITHUB_REF_NAME#v}\]" CHANGELOG.md

  build:
    needs: validate
    strategy:
      matrix:
        include:
          - { os: ubuntu-latest, target: linux-x86_64 }
          - { os: ubuntu-latest, target: linux-aarch64 }
          - { os: macos-latest,  target: darwin-x86_64 }
          - { os: macos-latest,  target: darwin-aarch64 }
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - run: scripts/release/build.sh --target ${{ matrix.target }}
      - uses: actions/upload-artifact@v4
        if: ${{ hashFiles('dist/*') != '' }}
        with:
          name: buddy-${{ matrix.target }}
          path: dist/

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with: { path: ./artifacts }
      - run: scripts/release/checksums.sh ./artifacts > SHA256SUMS
      - id: notes
        run: scripts/release/changelog-extract.sh "$GITHUB_REF_NAME" > release-notes.md
      - uses: softprops/action-gh-release@v2
        with:
          files: |
            ./artifacts/**
            SHA256SUMS
          body_path: release-notes.md
          prerelease: ${{ contains(github.ref_name, '-rc') || contains(github.ref_name, '-beta') }}
```

### 8.4 `scripts/release/build.sh` (v0.1.0 stub)

```bash
#!/usr/bin/env bash
set -euo pipefail
TARGET="${1:-}"
mkdir -p dist
echo "[build] target=$TARGET"
echo "[build] no self-built artefacts in v$(cat VERSION); skipping."
# Future expansion examples:
#   rustc --target <…> -o dist/buddy-cli-"$TARGET" src/buddy-cli.rs
#   bun build --compile --target=<…> -o dist/buddy-helper-"$TARGET" src/buddy-helper.ts
```

The release workflow succeeds even with no artefacts; when binaries emerge, adding a `build.sh` branch + matrix target is the only change needed.

### 8.5 Release flow (human-run)

```bash
# 1. Edit VERSION and CHANGELOG.md (Unreleased → [x.y.z])
# 2. Commit
git add VERSION CHANGELOG.md
git commit -m "release: v0.2.0"
# 3. Tag + push
git tag v0.2.0
git push origin main --tags
# → release.yml runs, Release is published automatically.
```

## 9. Rationale Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Install location | `~/.buddy/` (clone) | Persistent, gitable, allows `git pull` updates |
| Deploy strategy | symlinks (default) | Edits in `~/.buddy/` reflect immediately; `--copy` escape hatch preserved |
| Namespace | `buddy-` / `buddy:` prefixes | Coexists safely with other plugins and user files |
| Manifest | `plugin/.claude-plugin/plugin.json` | Self-contained `plugin/` unit matches Claude Code native format |
| Metadata | `~/.claude/.buddy-metadata.json` | Install ledger survives `~/.buddy/` removal; drives precise uninstall |
| Binary strategy | hybrid | External MCP binaries committed to `vendor/`; self-built → Release assets |
| Release trigger | explicit `v*.*.*` tag push | Prevents accidental releases; tag is deliberate |
| CI self-test | `BUDDY_REPO=$(pwd)` mode | Full install → doctor → uninstall cycle validated per PR |
| Doctor runs auto | after install/update | Catches problems at the earliest opportunity |

## 10. Open Questions / Future Work

1. **Manifest `rules` field support**: Claude Code native plugin loader's handling of a non-standard `rules` key is not verified. First implementation should confirm behavior; if the loader rejects the manifest, move `rules` out of `plugin.json` and rely solely on `scripts/install.sh`.
2. **Windows / WSL support**: Not in v0.1.0. Scripts use POSIX-only constructs; porting to Windows-native is future work.
3. **Component set**: The initial scaffold ships empty `plugin/commands/`, `plugin/agents/`, etc. Curation of the first component batch is tracked in follow-up plans.
4. **Signed releases**: SHA256SUMS only; cosign / sigstore integration is a future enhancement.
5. **Telemetry / analytics**: Not included. If needed later, would be opt-in and documented.
6. **Multi-profile support**: Single global install. If per-project or per-profile variants emerge, revisit `~/.buddy/` → `~/.buddy/profiles/<name>/` structure.

## 11. Implementation Plan Handoff

After user approval of this spec, invoke `writing-plans` skill to produce a step-by-step implementation plan covering:

- Phase 1: repo skeleton (directories, stubs, VERSION/CHANGELOG/README baseline)
- Phase 2: `lib/bash/` helpers (log, platform, json, symlink)
- Phase 3: `scripts/install.sh` + `scripts/build-mcp-aggregate.sh`
- Phase 4: `scripts/uninstall.sh` + `scripts/doctor.sh` + `scripts/update.sh`
- Phase 5: Root `install.sh` / `uninstall.sh` entry points
- Phase 6: CI workflows (`ci.yml`, `release.yml`) + release helpers
- Phase 7: Integration test on a fresh machine (or fake HOME)
- Phase 8: v0.1.0 release dry-run
