# ADR-016 — F2.E Notification design (W7-5, v1.0 entry C-5)

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: —
**Related**: ADR-009 (cli buddy vision — F2.E area), ADR-010 (v1.0.0 scope — closes C-5), ADR-011 (release policy), ADR-015 (F2.C Advisor — input source)

## Context

ADR-009 declared F2.E as the *delivery* layer that surfaces F2.C / F2.D messages to OS-level / out-of-band channels:

> "TUI banner / desktop (`osascript` / `notify-send`) / shell prompt / webhook"

ADR-015 (v0.11.0) shipped the advisor that *produces* advisories and persists them to the `advisories` table. The remaining gap: getting those advisories *in front of the user* without requiring them to run `buddy advise` manually.

Current fragments:

- `internal/agent/runtime.go::postWebhook` — battle-tested HTTP POST/PUT/PATCH + headers + timeout pattern. Reusable as the webhook channel.
- `internal/tui/model.go::ModeList` — list view that runs all the time when TUI is open; natural host for a banner.
- `internal/persona/` — Korean friend-tone catalog. New `KeyNotify*` keys go here.
- `internal/daemon/daemon.go::runAdvisorMonitor` — already fires every `AdvisorPollInterval`; natural hook point for dispatch.

Gap: no `internal/notify/` package, no `notification_log` table, no per-channel config keys, no daemon dispatcher.

## Decision

W7-5 ships **Notification v1.0** with 4 channels + daemon auto-dispatch + per-channel severity floor + per-channel dedup window. v0.12.0 closes whole-product v1.0.0 entry condition **C-5**.

### Q1 — Channels (4-way ship, all per user confirm)

| Channel        | Mechanism                                                                            | Config keys                                                  |
|----------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------|
| **desktop**    | macOS: `osascript -e 'display notification ...'`. Linux: `notify-send`.              | `notify.desktopEnabled`, `notify.desktopSeverityMin`, `notify.desktopDedup` |
| **webhook**    | HTTP POST/PUT/PATCH + custom headers + timeout. Reuses `agent/postWebhook` pattern.  | `notify.webhooks: [{url, method, headers, severity_min, dedup_window}]` |
| **tui-banner** | Recent unsent-or-fresh advisories appear in ModeList's top banner. Read from store.  | `notify.tuiBannerEnabled`, `notify.tuiBannerSeverityMin` |
| **shell**     | `buddy notify --prompt` emits a single-line summary for PS1 integration.             | `notify.shellPromptEnabled`, `notify.shellPromptSeverityMin` |

Desktop binary is detected at startup (`exec.LookPath` for `osascript` / `notify-send`). If missing → desktop channel silently disabled (daemon log records the skip).

### Q2 — Trigger: daemon auto-dispatch only

The daemon's `runAdvisorMonitor` already produces new advisories. After each successful Persist(), `dispatcher.Dispatch(advisories)` runs. No CLI on-demand path in v0.12.0 — keeps the system "one trigger, one place" so users don't have to think about scheduling.

> CLI escape hatches that *do* ship: `buddy notify test --channel <name>` to verify wiring, `buddy notify status` to see what's been sent.

### Q3 — Channel config in `buddy config`

Per user answer. Eight v0.12.0 keys are added to `config.Effective`:

```yaml
notifyDesktopEnabled: true
notifyDesktopSeverityMin: "warn"
notifyDesktopDedup: "1h"

notifyTuiBannerEnabled: true
notifyTuiBannerSeverityMin: "info"

notifyShellPromptEnabled: false
notifyShellPromptSeverityMin: "warn"

notifyWebhooks: []  # [{url, method?, headers?, severityMin?, dedupWindow?}]
```

The webhook list is the only structured field — each entry has its own per-channel severity / dedup. Empty list = no webhook dispatch (default).

### Q4 — Per-channel severity floor + dedup window

Two complementary throttles, both per-channel:

- **Severity floor**: an advisory below the channel's `severityMin` is skipped for that channel only. (Allows "desktop = warn+", "webhook = high only", "tui = info+".)
- **Dedup window**: a `(channel, advisory.Kind)` pair sent inside the dedup window is skipped on subsequent dispatches. Tracked via the new `notification_log` table.

Quiet hours (timezone-aware) intentionally not included in v0.12.0 — bring it back as v0.12.x trigger if users ask.

### Data model

Migration v8 adds `notification_log`:

```sql
CREATE TABLE notification_log (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  advisory_id  INTEGER NOT NULL,
  channel      TEXT    NOT NULL,
  kind         TEXT    NOT NULL,
  severity     TEXT    NOT NULL,
  sent_at      INTEGER NOT NULL,
  outcome      TEXT    NOT NULL DEFAULT 'sent', -- 'sent' | 'skipped-severity' | 'skipped-dedup' | 'error'
  detail       TEXT    NOT NULL DEFAULT '',
  FOREIGN KEY (advisory_id) REFERENCES advisories(id) ON DELETE CASCADE
);
CREATE INDEX idx_notification_log_channel_kind ON notification_log(channel, kind);
CREATE INDEX idx_notification_log_sent_at      ON notification_log(sent_at);
```

`outcome=skipped-*` rows are still written so users can debug "why didn't I get notified".

### Code shape

```
internal/notify/
  types.go        — Notification, Channel, Dispatcher, ChannelConfig
  desktop.go      — DesktopChannel (osascript / notify-send invoker)
  webhook.go      — WebhookChannel (lifted from agent/postWebhook)
  tui_banner.go   — read-side helper for the TUI to pull recent advisories
  shell_prompt.go — single-line formatter for PS1
  store.go        — notification_log CRUD
  dispatcher.go   — orchestrates severity filter + dedup + channel calls
```

CLI:

```
cmd/buddy/notify_cmd.go
  buddy notify test --channel <desktop|webhook|tui-banner|shell>
  buddy notify status [--since DUR] [--channel X]
  buddy notify --prompt [--severity-min warn]   # shell PS1 hook
```

MCP:

```
internal/mcp/notify_tool.go
  notify_status — read-only log query
  notify_test   — dispatch a test notification
```

TUI:

ModeList renders a top banner of up to 3 recent unmuted advisories whose severity ≥ tuiBannerSeverityMin (default info). The banner becomes invisible once all are older than the dedup window. The Usage pane's advisory section (W7-3b) stays — these are complementary: banner is the at-a-glance reminder, Usage pane is the drill-down.

## Alternatives considered

### Option A — Skip TUI banner; rely on Usage pane only (rejected)

Half the user value: a user who never enters Usage pane misses everything. Banner is cheap and the natural counterpart to the ModeList "always on" surface.

### Option B — CLI-trigger dispatch only (rejected)

Forces users to schedule it. Too much friction for the "친구가 옆에서 한 마디" persona.

### Option C — Separate `~/.buddy/notify.yaml` config (rejected)

Splits the config surface. Single config.json keeps `buddy config show` authoritative.

### Option D — Per-channel quiet hours (deferred)

Not in v0.12.0. Future trigger: user feedback "I get pinged at 3am".

### Option E — Email channel (deferred)

SMTP introduces creds management. Webhook → email gateway covers the same use case until demand justifies native SMTP.

## Consequences

- **`internal/db/migrations.go` v8** — `notification_log` table per above.
- **`internal/notify/`** (new package, ~600 LoC, ~20 race-clean tests).
- **`internal/config/`** — 8 new keys + a structured `notifyWebhooks` slice + validation.
- **`cmd/buddy/notify_cmd.go`** — 3 subcommands (`test`, `status`, `--prompt`).
- **`internal/mcp/notify_tool.go`** — 2 tools (`notify_status`, `notify_test`).
- **`internal/tui/model.go`** — ModeList banner section. `cmd/buddy/tui_cmd.go` wires `NotifyFetcher`.
- **`internal/daemon/daemon.go`** — `runAdvisorMonitor` calls `notify.Dispatcher.Dispatch(advisories)` after each Persist. No new goroutine — same cadence as advisor.
- **No external dependency** — desktop channel uses `os/exec`. Webhook channel uses stdlib `net/http`.
- **`docs/cli-buddy-spec.md` §9 W7** → Done.
- **BACKLOG / HANDOFF** — 7/9 closed (78%); only B-2 (user-paced) + C-4 remain.

## Verification

When W7-5 ships (v0.12.0):

```bash
# After buddy v0.12.0 install + advisor running:

buddy notify test --channel desktop       # macOS popup or Linux toast
buddy notify test --channel webhook       # POSTs sample payload to first config'd webhook
buddy notify status --since 24h           # log rows

buddy daemon start                        # advisor + notify dispatch kicks in
sleep 3700 && sqlite3 ~/.buddy/buddy.db \
  'SELECT channel, COUNT(*) FROM notification_log GROUP BY channel'

# Shell prompt:
echo 'export PROMPT="$(buddy notify --prompt) $PROMPT"' >> ~/.zshrc
# next shell shows e.g. "⚠ token spike"

# MCP:
# mcp__buddy__notify_status                # JSON log
# mcp__buddy__notify_test                  # dispatch test

go test -race -count=1 ./internal/notify/... # 20+ race-clean tests
make verify-versions                          # 5 sources on 0.12.0
```

## Trigger to revisit

- User asks for quiet hours → add `notify.quietHoursStart/End` (v0.12.x).
- User wants Slack-specific formatting → add channel sub-type (`notify.webhooks[].format: "slack"`).
- Notification volume becomes noisy → bump dedup defaults / surface per-advisory mute (W7-3b already supports muting at the source).
- Linux variant `kdialog` / `zenity` requested → add to desktop channel auto-detect.

## References

- ADR-009 (cli buddy vision) — F2.E area definition.
- ADR-010 (v1.0.0 scope) — C-5 condition that this ADR closes.
- ADR-015 (F2.C Advisor) — input source for notifications.
- `internal/agent/runtime.go::postWebhook` — the webhook impl that gets lifted into the notify package.
- macOS Notification: `osascript -e 'display notification "msg" with title "title"'`.
- Linux Notification: `notify-send "title" "msg"`.
