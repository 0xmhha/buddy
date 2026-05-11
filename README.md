# Buddy

A reliability and observability control plane for [Claude Code](https://claude.ai/code) sessions.

Buddy wraps your Claude Code hooks, validates state schemas, and surfaces failures before they silently accumulate — plus a Claude Code plugin with 99 slash commands and 148 skills covering the full product development lifecycle, dispatched through a single auto-loaded `router` skill.

```
              ┌──────────────────────────┐
  you ──────▶ │      buddy (control)     │ ──▶  Claude Code sessions
              │  hooks · state · tasks   │       (1, 2, 3, ... N)
              └──────────────────────────┘
```

---

## Two tracks

This repo contains two independent assets that share the `buddy` name. The canonical definitions and responsibility boundaries live in [`docs/two-tracks-charter.md`](./docs/two-tracks-charter.md); the one-line summary:

| Track | One-line definition |
|-------|---------------------|
| **plugin buddy** | A Claude Code plugin — a unified catalog of skills, MCP servers, agents, and hooks that supports the full product lifecycle (idea → business validation → app/web decision → design → spec → build → automated tests → deploy → A/B → growth → marketing → maintenance) creatively, efficiently, and reliably. |
| **cli buddy** | A TUI tool that embeds plugin buddy to manage automation agents — create, run, stop, and configure agents that drive long-running automated work (e.g. a "draw a webtoon" agent that organizes a story world, plans the daily episode, generates art, and publishes to a target service on schedule). |

cli buddy *embeds* plugin buddy; plugin buddy stands alone. The Features table below describes what is actually shipped today, which is plugin buddy in full and a small slice of cli buddy (the v0.1.0 hook-reliability monitor — one sub-feature of the broader cli buddy goal).

---

## Features

| Area | Track | What it does |
|------|-------|-------------|
| **Claude Code plugin** | plugin buddy | 9-phase lifecycle orchestrator, 99 `/buddy:*` commands, 148 skills behind one router |
| **Hook reliability** | cli buddy (v0.1.0) | Wraps Claude Code hooks; surfaces silent failures with structured logs |
| **State schema** | cli buddy (v0.1.0) | Zod-validated JSON state prevents corruption and schema drift |
| **Task retry** | cli buddy (v0.1.0) | WAL-backed outbox ensures failed tasks are replayed, not dropped |
| **Observability** | cli buddy (v0.1.0) | Unified token/cost/session/hook status via a single `stats` command |
| **TUI agent manager** | cli buddy (planned) | Create / run / stop / configure automation agents that embed plugin buddy. Not yet implemented — see charter §3 |

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
VERSION=0.5.0

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
gh release download v0.5.0 --repo 0xmhha/buddy \
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
# buddy  0.5.0  installed
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

Once the plugin is installed, 99 slash commands are available in any Claude Code session, all dispatched through the single auto-loaded `router` skill.

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
| §3 | `/buddy:explore-design-variants`, `/buddy:define-tech-stack`, `/buddy:design-data-model`, `/buddy:design-api-contract`, `/buddy:design-event-schema`, `/buddy:design-auth-model`, `/buddy:design-tenant-model`, `/buddy:map-use-cases-to-infra`, `/buddy:derive-system-topology`, `/buddy:write-adr` |
| §4 | `/buddy:decompose-feature-to-actor-tracks`, `/buddy:decompose-track-to-tasks`, `/buddy:map-task-dependencies`, `/buddy:plan-parallel-execution`, `/buddy:define-acceptance-test-plan`, `/buddy:estimate-build-timeline` |
| §5 | `/buddy:build-with-tdd`, `/buddy:diagnose-bug`, `/buddy:dispatch-parallel-agents` |
| §6 | `/buddy:audit-security`, `/buddy:audit-accessibility`, `/buddy:audit-cost-efficiency`, `/buddy:run-load-test`, `/buddy:measure-code-health`, `/buddy:test-per-actor-use-case`, `/buddy:test-cross-actor-flow` |
| §7 | `/buddy:auto-create-pr`, `/buddy:setup-quality-gates`, `/buddy:setup-canary-deploy`, `/buddy:setup-feature-flags`, `/buddy:setup-rollback-runbook`, `/buddy:run-uat`, `/buddy:run-beta-program`, `/buddy:prepare-launch-checklist`, `/buddy:setup-incident-paging` |
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

Portions of the Claude Code plugin skills (`plugin/skills/`) are derived from, inspired by, or reference the following MIT-licensed projects:

| Project | Author / Copyright | Used by |
|---------|-------------------|---------|
| [mattpocock/skills](https://github.com/mattpocock/skills) | Matt Pocock | `define-product-spec` (define-product-context + write-prd absorption), `review-engineering` (review-code-architecture absorption) |
| [gstack](https://github.com/garrytan/gstack) | Garry Tan | early plugin scaffolding inspiration |
| marketingskills | Corey Haines, 2025 | `analyze-competition-and-substitutes`, `optimize-conversion-funnel` (5 CRO sub), `draft-marketing-copy`, `plan-marketing-channel`, `audit-seo-aso`, `automate-marketing-content` (3 sub), `analyze-feature-adoption`, `analyze-user-cohort`, `analyze-customer-feedback-corpus`, `conduct-customer-interview`, `map-customer-segments` |
| designer-skills | MC Dean, 2026 | `apply-design-system`, `audit-ui-quality`, `prototype-from-spec`, `design-interaction-pattern`, `design-accessibility-baseline`, `conduct-customer-interview` (design-research) |
| make-interfaces-feel-better | (MIT) | `audit-ui-quality` (micro-detail patterns) |
| agent-evaluation | Kevin + Claude, 2026 (OMAS v2) | `audit-test-coverage-meaningful`, `analyze-actor-failure-rate` (input-vs-output trust scoring) |
| humanizer | Siqi Chen, 2025 | `analyze-customer-feedback-corpus` (AI-text inverse pattern) |
| [superpowers](https://github.com/obra/superpowers) | Jesse Vincent, 2025 | `docs/superpowers/` directory naming + composable-skill + router-instruction pattern. See [ADR-003](./docs/superpowers/decisions/2026-05-10-superpowers-attribution.md). |
| gpt-researcher | (referenced) | `conduct-customer-interview` (automation aid) |
| Korean legal cluster | varies (MIT) | Korea cluster deferred (`consult-korea-legal-context` etc) — `ai-professional-replacement-legal-exploration_skill`, `korean-legal-guide_skill`, `patent-application-drafting_skill`, `KESE-KIT` |

These works are used and modified in accordance with their respective MIT licenses. No verbatim code or text was adopted in any case — all buddy implementations are independently authored, with the upstream projects providing pattern inspiration, naming conventions, or domain framing only. Full license texts are reproduced in [`NOTICE`](./NOTICE).

---

## License

Copyright 2026 mhha

Licensed under the Apache License, Version 2.0. See [`LICENSE`](./LICENSE) for the full text.

Third-party components are listed in [`NOTICE`](./NOTICE).
