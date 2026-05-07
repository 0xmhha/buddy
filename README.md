# Buddy

A reliability and observability control plane for [Claude Code](https://claude.ai/code) sessions.

Buddy wraps your Claude Code hooks, validates state schemas, and surfaces failures before they silently accumulate — plus a Claude Code plugin with 30 slash commands and 78 skills covering the full product development lifecycle, dispatched through a single auto-loaded `router` skill.

```
              ┌──────────────────────────┐
  you ──────▶ │      buddy (control)     │ ──▶  Claude Code sessions
              │  hooks · state · tasks   │       (1, 2, 3, ... N)
              └──────────────────────────┘
```

---

## Features

| Area | What it does |
|------|-------------|
| **Hook reliability** | Wraps Claude Code hooks; surfaces silent failures with structured logs |
| **State schema** | Zod-validated JSON state prevents corruption and schema drift |
| **Task retry** | WAL-backed outbox ensures failed tasks are replayed, not dropped |
| **Observability** | Unified token/cost/session/hook status via a single `stats` command |
| **Claude Code plugin** | 9-phase lifecycle orchestrator, 30 `/buddy:*` commands, 78 skills behind one router |

---

## Requirements

- Go 1.22+ (build from source only)
- macOS or Linux (Windows: v1.0+)
- Claude Code (for the plugin)

---

## Installation

### Release binary (recommended)

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
VERSION=0.1.0

curl -fL "https://github.com/0xmhha/buddy/releases/download/v${VERSION}/buddy_${VERSION}_${OS}_${ARCH}" -o buddy
curl -fL "https://github.com/0xmhha/buddy/releases/download/v${VERSION}/SHA256SUMS" -o SHA256SUMS

# Verify checksum manually
shasum -a 256 buddy
grep "buddy_${VERSION}_${OS}_${ARCH}$" SHA256SUMS

chmod +x buddy
sudo mv buddy /usr/local/bin/
buddy --version
```

> **macOS:** Remove the quarantine attribute after download:
> `xattr -d com.apple.quarantine /usr/local/bin/buddy`

Using the `gh` CLI:

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
gh release download v0.1.0 --repo 0xmhha/buddy \
  --pattern "buddy_*_${OS}_${ARCH}" -O buddy
```

### Claude Code plugin

Install directly from GitHub — no local clone needed:

```bash
claude plugin marketplace add 0xmhha/buddy
claude plugin install buddy@buddy
```

Verify:

```bash
claude plugin list
# buddy  1.0.0  installed
```

Uninstall:

```bash
claude plugin uninstall buddy@buddy
claude plugin marketplace remove buddy
```

### Build from source

```bash
git clone https://github.com/0xmhha/buddy.git
cd buddy
make build       # outputs bin/buddy
```

---

## Usage

```bash
# Wrap Claude Code hooks (optionally generate cliwrap.yaml)
buddy install --with-cliwrap

# Start the background daemon
buddy daemon start

# Run diagnostics
buddy doctor

# View token/hook stats
buddy stats --window 1h
buddy stats --by-tool --window 5m

# Inspect and tune configuration
buddy config show
buddy config set hookSlowMs 3000

# Purge old data (dry-run first)
buddy purge --before 30d
buddy purge --before 30d --apply

# Stream raw events for debugging
buddy events --follow

# Remove hooks and stop daemon
buddy daemon stop
buddy uninstall
```

Full CLI reference: [`docs/v0.1-spec.md §7`](./docs/v0.1-spec.md).

### Claude Code plugin — slash commands

Once the plugin is installed, 30 slash commands are available in any Claude Code session, all dispatched through the single auto-loaded `router` skill.

#### Phase orchestrators (9 — pipeline entry points)

| Phase | Command | Purpose |
|-------|---------|---------|
| §1 | `/buddy:concretize-idea`   | Idea → PRD + business viability |
| §2 | `/buddy:define-features`   | Feature backlog + actor / use case mapping |
| §3 | `/buddy:design-system`     | Tech stack ADR + API contract + data model |
| §4 | `/buddy:plan-build`        | Implementation plan + parallel execution graph |
| §5 | `/buddy:build-feature`     | TDD loop + parallel agent dispatch |
| §6 | `/buddy:verify-quality`    | Test / lint / security / compliance gate |
| §7 | `/buddy:ship-release`      | PR + tag + changelog + canary |
| §8 | `/buddy:iterate-product`   | A/B + funnel + incident + improvement backlog |
| §9 | `/buddy:manage-lifecycle`  | Deprecation + migration + EOL |

#### Stage skills (called inside orchestrators or standalone)

| Phase | Commands |
|-------|----------|
| §1 | `/buddy:validate-idea`, `/buddy:validate-advanced-edge-idea`, `/buddy:assess-business-viability`, `/buddy:define-product-spec` |
| §3 | `/buddy:explore-design-variants` |
| §5 | `/buddy:build-with-tdd`, `/buddy:diagnose-bug`, `/buddy:dispatch-parallel-agents` |
| §6 | `/buddy:audit-security`, `/buddy:measure-code-health` |
| §7 | `/buddy:auto-create-pr`, `/buddy:setup-quality-gates` |
| §8 | `/buddy:summarize-retro` |

#### Cross-phase tools

| Command | Purpose |
|---------|---------|
| `/buddy:status`          | Detect current phase from repo artifacts + suggest next command |
| `/buddy:autoplan`        | 4-mode review pipeline (scope / engineering / design / devex) on any plan / PRD / ADR |
| `/buddy:consult-codex`   | Second opinion via external LLM CLI |
| `/buddy:save-context`    | Checkpoint git state + decisions + remaining tasks |
| `/buddy:restore-context` | Restore most recent saved checkpoint |

#### Router dispatch (composition)

| Command | Purpose |
|---------|---------|
| `/buddy:run <skill> [args]`                | Invoke any catalog skill directly (escape hatch for skills without a dedicated command) |
| `/buddy:chain skill1,skill2,... -- args`    | Sequential composition — each skill receives the prior output |
| `/buddy:parallel skill1,skill2,... -- args` | Parallel composition via Agent dispatch, with aggregated results |

#### Architecture

All slash commands route through a single auto-loaded `router` skill (`plugin/skills/router/SKILL.md`). The catalog and routing rules live at `plugin/skills/router/references/skill-catalog.md` and `plugin/skills/router/references/routing-rules.md`. This keeps always-loaded skill metadata to one description (~170 chars) regardless of how many procedures exist — new procedures can be added without inflating session context.

---

## Roadmap

| Version | Focus | Key additions |
|---------|-------|---------------|
| **v0.1** ✓ | Reliability | Hook monitor, Zod state schema, WAL replay |
| v0.2 | Control Plane | Multi-session dashboard, unified token/cost view |
| v0.3 | Orchestration | Task DAG, wave executor, auto-retry |
| v1.0 | Integration | AGENTS.md auto-sync, plugin model, MCP server |

---

## Contributing

Contributions are welcome. Please follow these steps:

1. Fork the repository and create a feature branch from `main`.
2. Run the test suite: `make test` (includes race detector).
3. If your change touches `plugin/`, also run `make test-routing` to verify the router wire-up.
4. Keep changes focused — one logical change per PR.
5. Open a pull request with a clear description of the problem and solution.

For larger changes, open an issue first to discuss the approach.

Bug reports and feature requests: [GitHub Issues](https://github.com/0xmhha/buddy/issues).

---

## Acknowledgments

Portions of the Claude Code plugin skills (`plugin/skills/`) are derived from or inspired by the following MIT-licensed projects:

- **[mattpocock/skills](https://github.com/mattpocock/skills)** — Copyright (c) 2026 Matt Pocock. [MIT License](https://opensource.org/licenses/MIT).
- **[gstack](https://github.com/garrytan/gstack)** — Copyright (c) 2026 Garry Tan. [MIT License](https://opensource.org/licenses/MIT).

These works are used and modified in accordance with their respective MIT licenses. Full license texts are reproduced in [`NOTICE`](./NOTICE).

---

## License

Copyright 2026 mhha

Licensed under the Apache License, Version 2.0. See [`LICENSE`](./LICENSE) for the full text.

Third-party components are listed in [`NOTICE`](./NOTICE).
