# Reference webtoon agent (cli-buddy-spec §2.2 W3-6)

End-to-end example that exercises every cli buddy capability shipped through v0.6.6. The spec walks the plugin's idea-to-build chain on a daily schedule and POSTs the result to a webtoon publishing API.

## What this exercises

| Capability | Where in this spec |
|---|---|
| Cron schedule | `schedule: "0 3 * * *"` |
| Multi-step chain | 4 `command:` entries (concretize-idea → write-prd → design-system → build-feature) |
| Exponential retry backoff (Tier 1.4 v0.6.3) | `retry.backoff_strategy: exponential` + `backoff_max: 2m` |
| Per-line streaming logs (Tier 1.5 v0.6.4) | observable via `buddy agent log webtoon-publish` |
| Scheduler live refresh (Tier 1.6 v0.6.5) | works with `buddy agent scheduler start --refresh 1m` |
| Webhook output (Tier 1.8 v0.6.6) | `output.type: webhook` |
| PROCEDURE §self-check + §next-phase parsing (W3-4 v0.5.0/v0.6.0) | automatic — runtime parses each step's stdout |

## How to use

### 1. Customise the spec

Before creating the agent, edit `spec.yaml`:

- Replace `Authorization: "Bearer REPLACE-WITH-YOUR-TOKEN"` with a real bearer token.
- Replace `url:` with your own publishing endpoint (or run a local httpbin / [webhook.site](https://webhook.site) target while testing).
- Optionally tweak `args:` on each step to your story prompt.

The runtime writes header values **literally** — there is no `${ENV}` expansion. If you want the token from an environment variable, template the YAML in your own shell:

```bash
sed "s|REPLACE-WITH-YOUR-TOKEN|$WEBTOON_API_TOKEN|" spec.yaml > /tmp/spec.yaml
buddy agent create /tmp/spec.yaml
rm /tmp/spec.yaml
```

### 2. Register the agent

```bash
buddy agent create examples/webtoon-agent/spec.yaml
# created agent "webtoon-publish"

buddy agent list
# webtoon-publish   schedule="0 3 * * *"   status=idle
```

### 3. Start the scheduler (or run once on-demand)

**Scheduled run** — the recommended path:

```bash
buddy agent scheduler start --refresh 1m
# scheduler: started (location=Asia/Seoul, entries=1, refresh=1m)
# (waits for 03:00 to tick)
```

**On-demand smoke test** — useful before relying on the cron:

```bash
buddy agent run webtoon-publish
# (streams Claude output of each chain step, then POSTs the result)
```

### 4. Observe progress

While the agent runs (or after), tail the logs:

```bash
buddy agent log webtoon-publish
# Run ID:     1
# Started:    2026-05-12 03:00:01 UTC
# Log:
#     03:00:01  info   agent "webtoon-publish" started (4 steps)
#     03:00:01  info   step[0] concretize-idea attempt=1
#     03:00:03  info   step[0] concretize-idea stdout: ## Stage 1 — refine the seed idea
#     03:00:04  info   step[0] concretize-idea stdout: ...
#     03:00:42  info   step[0] concretize-idea self-check=pass (5/5 passed)
#     03:00:42  info   step[0] concretize-idea next-phase candidates: write-prd
#     03:00:42  info   step[0] concretize-idea ok
#     03:00:42  info   step[1] write-prd attempt=1
#     ...
#     03:14:18  info   agent "webtoon-publish" finished status=done exit=0
```

### 5. Tweak without restart

Thanks to scheduler live refresh (v0.6.5), you can edit, delete, and recreate the agent in a separate shell — the running scheduler picks up the change within `--refresh` interval:

```bash
# shell B — change the schedule to every 12h
buddy agent delete webtoon-publish
# (edit spec.yaml: schedule: "0 */12 * * *")
buddy agent create examples/webtoon-agent/spec.yaml

# shell A (scheduler) logs:
# scheduler: agent "webtoon-publish" unscheduled (entry_id=1 removed from cron)
# scheduler: agent "webtoon-publish" scheduled (cron="0 */12 * * *" entry_id=2)
# scheduler: refresh poll applied (+1 / -1 / ~0)
```

### 6. What the webhook receives

Each successful run POSTs JSON shaped like:

```json
{
  "run_id": 17,
  "agent_id": "webtoon-publish",
  "exit_code": 0,
  "steps": [
    {
      "command": "concretize-idea",
      "args": "Today's webtoon: <theme>",
      "attempt": 1,
      "exit_code": 0,
      "stdout": "## Stage 1 — refine the seed idea\n...",
      "stderr": "",
      "parsed": {
        "self_check": {
          "verdict": "pass",
          "passed": 5,
          "total": 5,
          "items": [...]
        },
        "next_phase": {
          "skills": ["write-prd"],
          "branches": []
        }
      }
    },
    { "command": "write-prd", "..." },
    { "command": "design-system", "..." },
    { "command": "build-feature", "..." }
  ]
}
```

Non-2xx responses surface as a warn-level `agent_logs` line (without flipping the step's exit code), so a webhook outage is observable but does not corrupt the run record.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `agent: claude CLI not found on PATH` | `claude` not installed or not on PATH | Install Claude Code, or set `--claude-binary /path/to/claude` on `buddy agent run` |
| `agent: webhook POST … returned 401` | Authorization header wrong / expired | Update the spec, `buddy agent delete` + `create` again |
| `agent: webhook POST … context deadline exceeded` | Receiver slow, `timeout: 90s` not enough | Raise `output.timeout` in the spec |
| Step retries hit `max_attempts: 5` and fail | Upstream actually down | Check `buddy agent log webtoon-publish` for the per-attempt `stderr:` lines |

## Reference

- [`docs/cli-buddy-spec.md`](../../docs/cli-buddy-spec.md) §2.2 (motivating example)
- [`docs/cli-buddy-spec.md`](../../docs/cli-buddy-spec.md) §9 W3-6 (this implementation)
- [`CHANGELOG.md`](../../CHANGELOG.md) entries for v0.6.3 → v0.6.6 (the dependencies this agent stitches together)
