# Buddy

A reliability and observability control plane for [Claude Code](https://claude.ai/code) sessions.

Buddy wraps your Claude Code hooks, validates state schemas, and surfaces failures before they silently accumulate — plus a Claude Code plugin with 26 slash commands and 77 skills covering the full product development lifecycle.

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
| **Claude Code plugin** | 9-phase lifecycle orchestrator, 26 `/buddy:*` commands, 77 skills |

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

Once the plugin is installed, the following commands are available in any Claude Code session:

| Phase | Command |
|-------|---------|
| §1 Idea | `/buddy:concretize-idea`, `/buddy:validate-idea` |
| §2 Features | `/buddy:define-features`, `/buddy:map-actor-use-cases` |
| §3 Design | `/buddy:design-system`, `/buddy:explore-design-variants` |
| §4 Plan | `/buddy:plan-build`, `/buddy:autoplan` |
| §5 Build | `/buddy:build-feature`, `/buddy:build-with-tdd` |
| §6 Quality | `/buddy:verify-quality`, `/buddy:audit-security` |
| §7 Release | `/buddy:ship-release`, `/buddy:auto-create-pr` |
| §8 Operate | `/buddy:iterate-product`, `/buddy:analyze-ab-experiment` |
| §9 Lifecycle | `/buddy:manage-lifecycle` |

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
3. Keep changes focused — one logical change per PR.
4. Open a pull request with a clear description of the problem and solution.

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
