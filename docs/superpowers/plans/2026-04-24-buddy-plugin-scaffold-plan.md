# Buddy Plugin Scaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the distribution scaffold for the `buddy` Claude Code plugin (install/update/uninstall/doctor + CI), producing a v0.1.0 release with empty payload slots ready to receive curated components.

**Architecture:** Single GitHub repo cloned to `~/.buddy/` by a `curl | bash` installer. Components deploy to `~/.claude/` via symlinks (commands/agents/skills/rules) and jq-based JSON merges (hooks/mcp). A runtime metadata ledger (`~/.claude/.buddy-metadata.json`) drives precise uninstall.

**Tech Stack:** Bash 4+, jq, git, bats-core (testing), GitHub Actions (CI).

**Spec:** `docs/superpowers/specs/2026-04-24-buddy-plugin-architecture-design.md`

---

## Prerequisites

Before starting:
- `bash >= 4.0` (macOS ships 3.2; install `brew install bash` if needed — but scripts target POSIX-ish bash4; note in README).
- `jq >= 1.6`
- `bats-core >= 1.10` for tests. On macOS: `brew install bats-core`. On Ubuntu: `sudo apt-get install bats`. Document in README.
- Git repo at `/Users/wm-it-22-00661/Work/github/study/ai/buddy` already initialized with `main` branch.

All paths below are relative to repo root unless noted.

---

## Phase 1: Repo Skeleton

### Task 1.1: Directory Structure + .gitignore

**Files:**
- Create: `.gitignore`
- Create: empty marker files in each component directory (so git tracks the structure)

- [ ] **Step 1: Write `.gitignore`**

Create `.gitignore`:
```
# Build artefacts
plugin/mcp/servers.json
dist/
SHA256SUMS

# Editor cruft
.DS_Store
*.swp
*.swo
.idea/
.vscode/

# Test scratch
/tmp-home/
/tmp-test-home/
```

- [ ] **Step 2: Create directory skeleton with .gitkeep markers**

Run:
```bash
cd /Users/wm-it-22-00661/Work/github/study/ai/buddy
mkdir -p plugin/{commands,agents,skills,rules,hooks,mcp,.claude-plugin}
mkdir -p vendor lib/bash scripts/release .github/workflows tests/bash
touch plugin/commands/.gitkeep plugin/agents/.gitkeep plugin/skills/.gitkeep \
      plugin/rules/.gitkeep plugin/mcp/.gitkeep vendor/.gitkeep \
      scripts/release/.gitkeep .github/workflows/.gitkeep
```

- [ ] **Step 3: Verify structure**

Run: `find . -type d -not -path '*/\.git*' -not -path '*/node_modules*' | sort`

Expected (subset):
```
.
./.github
./.github/workflows
./docs
./docs/superpowers
./docs/superpowers/plans
./docs/superpowers/specs
./lib
./lib/bash
./plugin
./plugin/.claude-plugin
./plugin/agents
./plugin/commands
./plugin/hooks
./plugin/mcp
./plugin/rules
./plugin/skills
./scripts
./scripts/release
./tests
./tests/bash
./vendor
```

- [ ] **Step 4: Commit**

```bash
git add .gitignore plugin vendor lib scripts .github tests
git commit -m "scaffold: repo skeleton for buddy plugin"
```

---

### Task 1.2: VERSION + CHANGELOG

**Files:**
- Create: `VERSION`
- Create: `CHANGELOG.md`

- [ ] **Step 1: Write VERSION**

Create `VERSION` (single line, no trailing newline concerns):
```
0.1.0
```

- [ ] **Step 2: Write CHANGELOG.md**

Create `CHANGELOG.md`:
```markdown
# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-04-24

### Added
- Initial repository scaffold.
- `install.sh` / `uninstall.sh` / `update.sh` / `doctor.sh` lifecycle scripts.
- `lib/bash/` helpers: log, platform detection, jq JSON merge, idempotent symlink.
- `plugin/.claude-plugin/plugin.json` manifest.
- CI workflows: lint + smoke install (ci.yml), tag-triggered release (release.yml).
- Empty component slots: commands, agents, skills, rules, hooks, mcp.

### Notes
- No components shipped in this version. Component content arrives in follow-up releases.
```

- [ ] **Step 3: Commit**

```bash
git add VERSION CHANGELOG.md
git commit -m "chore: add VERSION and CHANGELOG baseline"
```

---

### Task 1.3: Plugin Manifest + Payload Placeholders

**Files:**
- Create: `plugin/.claude-plugin/plugin.json`
- Create: `plugin/hooks/hooks.json`
- Create: `vendor/README.md`
- Create: `lib/README.md`

- [ ] **Step 1: Write plugin.json**

Create `plugin/.claude-plugin/plugin.json`:
```json
{
  "name": "buddy",
  "version": "0.1.0",
  "description": "Personal Claude Code plugin — commands, skills, agents, rules, hooks, MCP",
  "author": {
    "name": "buddy-author",
    "url": "https://github.com/buddy-author/buddy"
  },
  "commands": "./commands/",
  "agents": "./agents/",
  "skills": "./skills/",
  "rules": "./rules/",
  "hooks": "./hooks/hooks.json",
  "mcpServers": "./mcp/servers.json"
}
```

(Author fields are placeholder — update to your GitHub handle before first release.)

- [ ] **Step 2: Write empty hooks.json**

Create `plugin/hooks/hooks.json`:
```json
{
  "hooks": []
}
```

- [ ] **Step 3: Write vendor convention doc**

Create `vendor/README.md` with the following content (use your editor or `cat > vendor/README.md <<'EOF' ... EOF`):

````markdown
# vendor/

Pre-built MCP server binaries committed to this repo (hybrid binary strategy — see spec §4).

## Layout per vendor entry

```
vendor/<name>/
├── UPSTREAM.md               # source repo URL, commit SHA, build steps, license
├── bin/
│   ├── <name>-x86_64-linux
│   ├── <name>-aarch64-linux
│   ├── <name>-x86_64-darwin
│   └── <name>-aarch64-darwin
└── mcp.config.json           # fragment merged into plugin/mcp/servers.json at install time
```

## `mcp.config.json` fragment format

```json
{
  "mcpServers": {
    "<server-key>": {
      "command": "{{BUDDY_BIN}}",
      "args": [],
      "env": {}
    }
  }
}
```

The `{{BUDDY_BIN}}` placeholder is replaced at install time with the absolute
path of the platform-matched binary in `vendor/<name>/bin/`.

The server key should be prefixed `buddy-` (e.g., `buddy-serena`) to avoid
collisions in `~/.claude/.mcp.json`.

## `UPSTREAM.md` template

```
# <name>

- **Source repo:** <url>
- **Commit:** <sha>
- **Version tag:** <tag or N/A>
- **License:** <spdx>
- **Build command:** <exact command used>
- **Build date:** <YYYY-MM-DD>
```

Every vendor entry MUST have a complete `UPSTREAM.md`. CI does not yet enforce
this; add a lint step once the first vendor entry lands.
````

- [ ] **Step 4: Write lib README**

Create `lib/README.md`:
```markdown
# lib/

Reusable libraries used by scripts in `scripts/`.

- `bash/log.sh` — `info()`, `warn()`, `err()` log helpers (stderr, color-aware).
- `bash/platform.sh` — `detect_os()`, `detect_arch()`, `resolve_binary()`.
- `bash/json.sh` — jq-based merge/patch helpers for `settings.json` and `.mcp.json`.
- `bash/symlink.sh` — idempotent `link_file()`, `link_dir()`, `verify_link()`, `unlink_if_owned()`.

Sourced from orchestration scripts via `source "$BUDDY_ROOT/lib/bash/<name>.sh"`.

Keep functions pure (no global side effects beyond explicit `set -eu` locally);
orchestration scripts control process-wide settings.
```

- [ ] **Step 5: Commit**

```bash
git add plugin/.claude-plugin/plugin.json plugin/hooks/hooks.json vendor/README.md lib/README.md
git commit -m "scaffold: plugin manifest and payload placeholders"
```

---

### Task 1.4: Bats Test Harness Setup

**Files:**
- Create: `tests/bash/test_helper.bash`
- Create: `tests/bash/README.md`

- [ ] **Step 1: Write test_helper.bash**

Create `tests/bash/test_helper.bash`:
```bash
# shellcheck shell=bash
# Common bats test helpers. Source from individual .bats files.

TESTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$( cd "$TESTS_DIR/../.." && pwd )"
LIB_DIR="$REPO_ROOT/lib/bash"
SCRIPTS_DIR="$REPO_ROOT/scripts"

export BUDDY_ROOT="$REPO_ROOT"

make_tmp_home() {
  local dir
  dir="$(mktemp -d -t buddy-test-home.XXXXXX)"
  mkdir -p "$dir/.claude"
  echo "$dir"
}

cleanup_tmp_home() {
  local dir="$1"
  [ -n "$dir" ] && [ -d "$dir" ] && rm -rf "$dir"
}
```

- [ ] **Step 2: Write tests/bash README**

Create `tests/bash/README.md`:
```markdown
# tests/bash/

Bats-core tests for `lib/bash/*.sh` helpers and script smoke tests.

## Running

```bash
# Run all tests
bats tests/bash/

# Run one file
bats tests/bash/test_log.bats

# Verbose
bats --verbose-run tests/bash/test_log.bats
```

## Installing bats-core

- macOS: `brew install bats-core`
- Ubuntu: `sudo apt-get install -y bats`
- From source: see <https://github.com/bats-core/bats-core>

## Conventions

- Every helper function in `lib/bash/` has at least one test.
- `test_helper.bash` provides `make_tmp_home` / `cleanup_tmp_home` for fake-HOME tests.
- Tests must be runnable without a real `~/.claude/` present.
```

- [ ] **Step 3: Write sanity test**

Create `tests/bash/test_harness.bats`:
```bash
#!/usr/bin/env bats

load test_helper

@test "REPO_ROOT resolves to repo" {
  [ -f "$REPO_ROOT/VERSION" ]
}

@test "LIB_DIR resolves to lib/bash" {
  [ -d "$LIB_DIR" ]
}

@test "make_tmp_home creates .claude dir" {
  local home
  home="$(make_tmp_home)"
  [ -d "$home/.claude" ]
  cleanup_tmp_home "$home"
  [ ! -d "$home" ]
}
```

- [ ] **Step 4: Run harness test**

Run: `bats tests/bash/test_harness.bats`
Expected: `3 tests, 0 failures`.

If bats is not installed, the test will fail with `bats: command not found`; install bats-core per `tests/bash/README.md` and re-run.

- [ ] **Step 5: Commit**

```bash
git add tests/bash/
git commit -m "test: add bats test harness"
```

---

## Phase 2: Library Helpers (TDD)

### Task 2.1: lib/bash/log.sh

**Files:**
- Test: `tests/bash/test_log.bats`
- Create: `lib/bash/log.sh`

- [ ] **Step 1: Write failing tests**

Create `tests/bash/test_log.bats`:
```bash
#!/usr/bin/env bats

load test_helper

setup() {
  source "$LIB_DIR/log.sh"
}

@test "info writes to stderr with [info] prefix" {
  run bash -c "source '$LIB_DIR/log.sh'; info 'hello' 2>&1 1>/dev/null"
  [ "$status" -eq 0 ]
  [[ "$output" == *"[info]"* ]]
  [[ "$output" == *"hello"* ]]
}

@test "warn writes to stderr with [warn] prefix" {
  run bash -c "source '$LIB_DIR/log.sh'; warn 'careful' 2>&1 1>/dev/null"
  [ "$status" -eq 0 ]
  [[ "$output" == *"[warn]"* ]]
  [[ "$output" == *"careful"* ]]
}

@test "err writes to stderr with [err] prefix and returns 1" {
  run bash -c "source '$LIB_DIR/log.sh'; err 'nope' 2>&1 1>/dev/null; echo EXIT=$?"
  [ "$status" -eq 0 ]
  [[ "$output" == *"[err]"* ]]
  [[ "$output" == *"nope"* ]]
  [[ "$output" == *"EXIT=1"* ]]
}

@test "info does not write to stdout" {
  run bash -c "source '$LIB_DIR/log.sh'; info 'stdout-check' 2>/dev/null"
  [ "$status" -eq 0 ]
  [ -z "$output" ]
}

@test "NO_COLOR=1 suppresses ANSI escapes" {
  run bash -c "NO_COLOR=1 source '$LIB_DIR/log.sh'; info 'plain' 2>&1 1>/dev/null"
  [ "$status" -eq 0 ]
  [[ "$output" != *$'\033['* ]]
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `bats tests/bash/test_log.bats`
Expected: All 5 tests fail (`log.sh` does not exist yet).

- [ ] **Step 3: Implement log.sh**

Create `lib/bash/log.sh`:
```bash
# shellcheck shell=bash
# Logging helpers. Write to stderr. Color when stderr is a TTY and NO_COLOR is unset.

if [ -t 2 ] && [ -z "${NO_COLOR:-}" ]; then
  _BUDDY_C_RESET=$'\033[0m'
  _BUDDY_C_INFO=$'\033[36m'
  _BUDDY_C_WARN=$'\033[33m'
  _BUDDY_C_ERR=$'\033[31m'
else
  _BUDDY_C_RESET=''
  _BUDDY_C_INFO=''
  _BUDDY_C_WARN=''
  _BUDDY_C_ERR=''
fi

info() { printf '%s[info]%s %s\n' "$_BUDDY_C_INFO" "$_BUDDY_C_RESET" "$*" >&2; }
warn() { printf '%s[warn]%s %s\n' "$_BUDDY_C_WARN" "$_BUDDY_C_RESET" "$*" >&2; }
err()  { printf '%s[err]%s %s\n'  "$_BUDDY_C_ERR"  "$_BUDDY_C_RESET" "$*" >&2; return 1; }
```

- [ ] **Step 4: Run tests — expect pass**

Run: `bats tests/bash/test_log.bats`
Expected: `5 tests, 0 failures`.

- [ ] **Step 5: Commit**

```bash
git add lib/bash/log.sh tests/bash/test_log.bats
git commit -m "feat: add lib/bash/log.sh with tests"
```

---

### Task 2.2: lib/bash/platform.sh

**Files:**
- Test: `tests/bash/test_platform.bats`
- Create: `lib/bash/platform.sh`

- [ ] **Step 1: Write failing tests**

Create `tests/bash/test_platform.bats`:
```bash
#!/usr/bin/env bats

load test_helper

setup() {
  source "$LIB_DIR/platform.sh"
}

@test "detect_os returns darwin or linux" {
  run detect_os
  [ "$status" -eq 0 ]
  [[ "$output" == "darwin" ]] || [[ "$output" == "linux" ]]
}

@test "detect_arch returns aarch64 or x86_64" {
  run detect_arch
  [ "$status" -eq 0 ]
  [[ "$output" == "aarch64" ]] || [[ "$output" == "x86_64" ]]
}

@test "detect_arch normalizes arm64 to aarch64" {
  run bash -c "source '$LIB_DIR/platform.sh'; _BUDDY_UNAME_M=arm64 detect_arch"
  [ "$status" -eq 0 ]
  [ "$output" = "aarch64" ]
}

@test "resolve_binary returns path when binary exists" {
  local home; home="$(make_tmp_home)"
  local vendor="$home/vendor/mytool"
  mkdir -p "$vendor/bin"
  local os; os="$(detect_os)"
  local arch; arch="$(detect_arch)"
  local bin="$vendor/bin/mytool-$arch-$os"
  echo '#!/bin/sh' > "$bin"
  chmod +x "$bin"
  run resolve_binary "$vendor" mytool
  [ "$status" -eq 0 ]
  [ "$output" = "$bin" ]
  cleanup_tmp_home "$home"
}

@test "resolve_binary returns non-zero when binary missing" {
  local home; home="$(make_tmp_home)"
  mkdir -p "$home/vendor/mytool/bin"
  run resolve_binary "$home/vendor/mytool" mytool
  [ "$status" -ne 0 ]
  cleanup_tmp_home "$home"
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `bats tests/bash/test_platform.bats`
Expected: all fail (function not defined).

- [ ] **Step 3: Implement platform.sh**

Create `lib/bash/platform.sh`:
```bash
# shellcheck shell=bash
# Platform detection and binary resolution.

detect_os() {
  local s
  s="$(uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')"
  case "$s" in
    darwin) echo darwin ;;
    linux)  echo linux ;;
    *) echo "unknown-$s"; return 1 ;;
  esac
}

detect_arch() {
  local m="${_BUDDY_UNAME_M:-$(uname -m 2>/dev/null)}"
  case "$m" in
    arm64|aarch64) echo aarch64 ;;
    x86_64|amd64)  echo x86_64 ;;
    *) echo "unknown-$m"; return 1 ;;
  esac
}

# resolve_binary <vendor-dir> <binary-name>
# Looks for <vendor-dir>/bin/<name>-<arch>-<os>. Echoes absolute path on success.
resolve_binary() {
  local vendor="$1" name="$2"
  local os arch
  os="$(detect_os)" || return 1
  arch="$(detect_arch)" || return 1
  local candidate="$vendor/bin/$name-$arch-$os"
  if [ -x "$candidate" ]; then
    echo "$candidate"
    return 0
  fi
  return 1
}
```

- [ ] **Step 4: Run tests — expect pass**

Run: `bats tests/bash/test_platform.bats`
Expected: `5 tests, 0 failures`.

- [ ] **Step 5: Commit**

```bash
git add lib/bash/platform.sh tests/bash/test_platform.bats
git commit -m "feat: add lib/bash/platform.sh with tests"
```

---

### Task 2.3: lib/bash/json.sh

**Files:**
- Test: `tests/bash/test_json.bats`
- Create: `lib/bash/json.sh`

- [ ] **Step 1: Write failing tests**

Create `tests/bash/test_json.bats`:
```bash
#!/usr/bin/env bats

load test_helper

setup() {
  source "$LIB_DIR/json.sh"
  TMP_HOME="$(make_tmp_home)"
}

teardown() {
  cleanup_tmp_home "$TMP_HOME"
}

@test "json_ensure_file creates empty object when missing" {
  local f="$TMP_HOME/a.json"
  json_ensure_file "$f"
  [ -f "$f" ]
  run jq -e '. == {}' "$f"
  [ "$status" -eq 0 ]
}

@test "json_ensure_file leaves existing content alone" {
  local f="$TMP_HOME/a.json"
  echo '{"keep":"me"}' > "$f"
  json_ensure_file "$f"
  run jq -e '.keep == "me"' "$f"
  [ "$status" -eq 0 ]
}

@test "json_merge_mcp adds server with command path" {
  local f="$TMP_HOME/.mcp.json"
  json_ensure_file "$f"
  json_merge_mcp "$f" "buddy-serena" "/opt/serena/bin/serena" '["stdio"]' '{}'
  run jq -er '.mcpServers."buddy-serena".command' "$f"
  [ "$status" -eq 0 ]
  [ "$output" = "/opt/serena/bin/serena" ]
  run jq -er '.mcpServers."buddy-serena".args | length' "$f"
  [ "$output" = "1" ]
}

@test "json_remove_mcp deletes buddy-tagged server" {
  local f="$TMP_HOME/.mcp.json"
  echo '{"mcpServers":{"buddy-x":{"command":"/x"},"other":{"command":"/y"}}}' > "$f"
  json_remove_mcp "$f" "buddy-x"
  run jq -er '.mcpServers."buddy-x" // "absent"' "$f"
  [ "$output" = "absent" ]
  run jq -er '.mcpServers.other.command' "$f"
  [ "$output" = "/y" ]
}

@test "json_merge_hook appends hook with id" {
  local f="$TMP_HOME/settings.json"
  json_ensure_file "$f"
  local hook='{"id":"buddy:test","event":"PreToolUse","matcher":"Edit","command":"echo"}'
  json_merge_hook "$f" "$hook"
  run jq -er '.hooks | length' "$f"
  [ "$output" = "1" ]
  run jq -er '.hooks[0].id' "$f"
  [ "$output" = "buddy:test" ]
}

@test "json_merge_hook replaces entry with same id" {
  local f="$TMP_HOME/settings.json"
  echo '{"hooks":[{"id":"buddy:test","command":"old"}]}' > "$f"
  local hook='{"id":"buddy:test","command":"new"}'
  json_merge_hook "$f" "$hook"
  run jq -er '.hooks | length' "$f"
  [ "$output" = "1" ]
  run jq -er '.hooks[0].command' "$f"
  [ "$output" = "new" ]
}

@test "json_remove_hooks_by_id strips matching entries" {
  local f="$TMP_HOME/settings.json"
  echo '{"hooks":[{"id":"buddy:a"},{"id":"other"},{"id":"buddy:b"}]}' > "$f"
  json_remove_hooks_by_id "$f" "buddy:a" "buddy:b"
  run jq -er '.hooks | length' "$f"
  [ "$output" = "1" ]
  run jq -er '.hooks[0].id' "$f"
  [ "$output" = "other" ]
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `bats tests/bash/test_json.bats`
Expected: all fail (functions not defined).

- [ ] **Step 3: Implement json.sh**

Create `lib/bash/json.sh`:
```bash
# shellcheck shell=bash
# jq-based JSON helpers for settings.json and .mcp.json manipulation.
# All writes are atomic (temp file + mv). All functions assume jq is on PATH.

_json_write_atomic() {
  local target="$1" content="$2"
  local tmp
  tmp="$(mktemp "${target}.XXXXXX")"
  printf '%s' "$content" > "$tmp"
  mv "$tmp" "$target"
}

# json_ensure_file <path>
# Creates file with '{}' if missing. Validates jq-parseability if existing.
json_ensure_file() {
  local f="$1"
  if [ ! -f "$f" ]; then
    mkdir -p "$(dirname "$f")"
    printf '{}' > "$f"
    return 0
  fi
  jq empty "$f" >/dev/null
}

# json_merge_mcp <file> <server-key> <command-path> <args-json-array> <env-json-object>
json_merge_mcp() {
  local f="$1" key="$2" cmd="$3" args="$4" envj="$5"
  local out
  out="$(jq \
    --arg key "$key" \
    --arg cmd "$cmd" \
    --argjson args "$args" \
    --argjson env "$envj" \
    '(.mcpServers // {}) as $s
     | .mcpServers = ($s + {($key): {"command":$cmd,"args":$args,"env":$env}})' \
    "$f")"
  _json_write_atomic "$f" "$out"
}

# json_remove_mcp <file> <server-key>
json_remove_mcp() {
  local f="$1" key="$2"
  local out
  out="$(jq --arg key "$key" 'if .mcpServers then .mcpServers |= del(.[$key]) else . end' "$f")"
  _json_write_atomic "$f" "$out"
}

# json_merge_hook <file> <hook-json-object>
# If a hook with same id already exists, replace it; otherwise append.
json_merge_hook() {
  local f="$1" hook_json="$2"
  local out
  out="$(jq \
    --argjson h "$hook_json" \
    '(.hooks // []) as $arr
     | ($h.id) as $id
     | .hooks = ((if $id then ($arr | map(select(.id != $id))) else $arr end) + [$h])' \
    "$f")"
  _json_write_atomic "$f" "$out"
}

# json_remove_hooks_by_id <file> <id> [<id>...]
json_remove_hooks_by_id() {
  local f="$1"; shift
  local ids_json
  ids_json="$(printf '%s\n' "$@" | jq -R . | jq -s .)"
  local out
  out="$(jq \
    --argjson ids "$ids_json" \
    'if .hooks then .hooks |= map(select((.id // "") as $id | ($ids | index($id) | not))) else . end' \
    "$f")"
  _json_write_atomic "$f" "$out"
}
```

- [ ] **Step 4: Run tests — expect pass**

Run: `bats tests/bash/test_json.bats`
Expected: `7 tests, 0 failures`.

- [ ] **Step 5: Commit**

```bash
git add lib/bash/json.sh tests/bash/test_json.bats
git commit -m "feat: add lib/bash/json.sh with tests"
```

---

### Task 2.4: lib/bash/symlink.sh

**Files:**
- Test: `tests/bash/test_symlink.bats`
- Create: `lib/bash/symlink.sh`

- [ ] **Step 1: Write failing tests**

Create `tests/bash/test_symlink.bats`:
```bash
#!/usr/bin/env bats

load test_helper

setup() {
  source "$LIB_DIR/symlink.sh"
  TMP_HOME="$(make_tmp_home)"
  SRC="$TMP_HOME/src.txt"
  echo "hello" > "$SRC"
}

teardown() {
  cleanup_tmp_home "$TMP_HOME"
}

@test "link_file creates symlink" {
  local tgt="$TMP_HOME/tgt.txt"
  link_file "$SRC" "$tgt"
  [ -L "$tgt" ]
  [ "$(readlink "$tgt")" = "$SRC" ]
}

@test "link_file is idempotent when target already matches" {
  local tgt="$TMP_HOME/tgt.txt"
  link_file "$SRC" "$tgt"
  link_file "$SRC" "$tgt"
  [ -L "$tgt" ]
}

@test "link_file fails on existing regular file without --force" {
  local tgt="$TMP_HOME/tgt.txt"
  echo "other" > "$tgt"
  run link_file "$SRC" "$tgt"
  [ "$status" -ne 0 ]
  [ -f "$tgt" ]
  [ ! -L "$tgt" ]
}

@test "link_file --force replaces existing regular file" {
  local tgt="$TMP_HOME/tgt.txt"
  echo "other" > "$tgt"
  link_file --force "$SRC" "$tgt"
  [ -L "$tgt" ]
}

@test "link_file replaces existing stale symlink" {
  local tgt="$TMP_HOME/tgt.txt"
  ln -s "/nonexistent" "$tgt"
  link_file "$SRC" "$tgt"
  [ -L "$tgt" ]
  [ "$(readlink "$tgt")" = "$SRC" ]
}

@test "link_dir creates directory symlink" {
  mkdir -p "$TMP_HOME/srcdir"
  link_dir "$TMP_HOME/srcdir" "$TMP_HOME/tgtdir"
  [ -L "$TMP_HOME/tgtdir" ]
}

@test "verify_link returns 0 when link matches" {
  local tgt="$TMP_HOME/tgt.txt"
  link_file "$SRC" "$tgt"
  run verify_link "$tgt" "$SRC"
  [ "$status" -eq 0 ]
}

@test "verify_link returns non-zero when link target differs" {
  local tgt="$TMP_HOME/tgt.txt"
  ln -s "/other" "$tgt"
  run verify_link "$tgt" "$SRC"
  [ "$status" -ne 0 ]
}

@test "unlink_if_owned removes symlink when target matches" {
  local tgt="$TMP_HOME/tgt.txt"
  link_file "$SRC" "$tgt"
  unlink_if_owned "$tgt" "$SRC"
  [ ! -e "$tgt" ]
}

@test "unlink_if_owned skips non-owned paths" {
  local tgt="$TMP_HOME/tgt.txt"
  ln -s "/other" "$tgt"
  run unlink_if_owned "$tgt" "$SRC"
  [ "$status" -ne 0 ]
  [ -L "$tgt" ]
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `bats tests/bash/test_symlink.bats`
Expected: all fail.

- [ ] **Step 3: Implement symlink.sh**

Create `lib/bash/symlink.sh`:
```bash
# shellcheck shell=bash
# Idempotent symlink utilities.

# link_file [--force] <source> <target>
# Creates a symlink at <target> pointing to <source>. Absolute paths recommended.
# Exit codes: 0 success, 1 target exists and is not a matching symlink (without --force).
link_file() {
  local force=0
  if [ "${1:-}" = "--force" ]; then force=1; shift; fi
  local src="$1" tgt="$2"
  mkdir -p "$(dirname "$tgt")"
  if [ -L "$tgt" ]; then
    local current
    current="$(readlink "$tgt")"
    if [ "$current" = "$src" ]; then
      return 0
    fi
    # stale or mismatched symlink — replace
    rm -f "$tgt"
  elif [ -e "$tgt" ]; then
    if [ "$force" -eq 1 ]; then
      rm -rf "$tgt"
    else
      return 1
    fi
  fi
  ln -s "$src" "$tgt"
}

# link_dir [--force] <source-dir> <target-dir>
# Same semantics as link_file, but source must be a directory.
link_dir() {
  local force=0
  if [ "${1:-}" = "--force" ]; then force=1; shift; fi
  local src="$1" tgt="$2"
  [ -d "$src" ] || return 1
  if [ "$force" -eq 1 ]; then
    link_file --force "$src" "$tgt"
  else
    link_file "$src" "$tgt"
  fi
}

# verify_link <target> <expected-source>
# Returns 0 if <target> is a symlink whose readlink matches <expected-source>.
verify_link() {
  local tgt="$1" expected="$2"
  [ -L "$tgt" ] || return 1
  local actual
  actual="$(readlink "$tgt")"
  [ "$actual" = "$expected" ]
}

# unlink_if_owned <target> <expected-source>
# Removes <target> only if it is a symlink pointing at <expected-source>.
# Prevents accidental removal of user-modified files.
unlink_if_owned() {
  local tgt="$1" expected="$2"
  if verify_link "$tgt" "$expected"; then
    rm -f "$tgt"
    return 0
  fi
  return 1
}
```

- [ ] **Step 4: Run tests — expect pass**

Run: `bats tests/bash/test_symlink.bats`
Expected: `10 tests, 0 failures`.

- [ ] **Step 5: Commit**

```bash
git add lib/bash/symlink.sh tests/bash/test_symlink.bats
git commit -m "feat: add lib/bash/symlink.sh with tests"
```

---

## Phase 3: Install Scripts

### Task 3.1: scripts/validate-manifest.sh

**Files:**
- Test: `tests/bash/test_validate_manifest.bats`
- Create: `scripts/validate-manifest.sh`

- [ ] **Step 1: Write failing tests**

Create `tests/bash/test_validate_manifest.bats`:
```bash
#!/usr/bin/env bats

load test_helper

setup() {
  TMP_HOME="$(make_tmp_home)"
}

teardown() {
  cleanup_tmp_home "$TMP_HOME"
}

@test "validate-manifest accepts a minimal valid manifest" {
  local f="$TMP_HOME/plugin.json"
  cat > "$f" <<EOF
{
  "name": "buddy",
  "version": "0.1.0",
  "description": "d",
  "commands": "./commands/",
  "agents": "./agents/",
  "skills": "./skills/",
  "hooks": "./hooks/hooks.json",
  "mcpServers": "./mcp/servers.json"
}
EOF
  run "$SCRIPTS_DIR/validate-manifest.sh" "$f"
  [ "$status" -eq 0 ]
}

@test "validate-manifest rejects missing name" {
  local f="$TMP_HOME/plugin.json"
  echo '{"version":"0.1.0"}' > "$f"
  run "$SCRIPTS_DIR/validate-manifest.sh" "$f"
  [ "$status" -ne 0 ]
  [[ "$output" == *"name"* ]]
}

@test "validate-manifest rejects non-semver version" {
  local f="$TMP_HOME/plugin.json"
  echo '{"name":"x","version":"hello","description":"d"}' > "$f"
  run "$SCRIPTS_DIR/validate-manifest.sh" "$f"
  [ "$status" -ne 0 ]
  [[ "$output" == *"version"* ]]
}

@test "validate-manifest rejects malformed JSON" {
  local f="$TMP_HOME/plugin.json"
  echo 'not json' > "$f"
  run "$SCRIPTS_DIR/validate-manifest.sh" "$f"
  [ "$status" -ne 0 ]
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `bats tests/bash/test_validate_manifest.bats`
Expected: all fail (script missing).

- [ ] **Step 3: Implement validate-manifest.sh**

Create `scripts/validate-manifest.sh`:
```bash
#!/usr/bin/env bash
# Validate plugin/.claude-plugin/plugin.json structure.
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=../lib/bash/log.sh
source "$SCRIPT_DIR/../lib/bash/log.sh"

if [ $# -ne 1 ]; then
  err "usage: $0 <path-to-plugin.json>"
fi

FILE="$1"

if [ ! -f "$FILE" ]; then
  err "manifest not found: $FILE"
fi

if ! jq empty "$FILE" >/dev/null 2>&1; then
  err "manifest is not valid JSON: $FILE"
fi

REQUIRED_KEYS=(name version description)
for k in "${REQUIRED_KEYS[@]}"; do
  if ! jq -e --arg k "$k" 'has($k)' "$FILE" >/dev/null; then
    err "missing required key: $k"
  fi
done

VERSION="$(jq -r '.version' "$FILE")"
if ! echo "$VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$'; then
  err "version is not semver: $VERSION"
fi

info "manifest valid: $FILE (name=$(jq -r '.name' "$FILE"), version=$VERSION)"
```

- [ ] **Step 4: Make executable and run tests**

Run:
```bash
chmod +x scripts/validate-manifest.sh
bats tests/bash/test_validate_manifest.bats
```
Expected: `4 tests, 0 failures`.

- [ ] **Step 5: Validate actual manifest**

Run: `scripts/validate-manifest.sh plugin/.claude-plugin/plugin.json`
Expected: `[info] manifest valid: ...`.

- [ ] **Step 6: Commit**

```bash
git add scripts/validate-manifest.sh tests/bash/test_validate_manifest.bats
git commit -m "feat: add scripts/validate-manifest.sh with tests"
```

---

### Task 3.2: scripts/build-mcp-aggregate.sh

**Files:**
- Test: `tests/bash/test_build_mcp_aggregate.bats`
- Create: `scripts/build-mcp-aggregate.sh`

- [ ] **Step 1: Write failing tests**

Create `tests/bash/test_build_mcp_aggregate.bats`:
```bash
#!/usr/bin/env bats

load test_helper

setup() {
  TMP_REPO="$(mktemp -d -t buddy-repo.XXXXXX)"
  mkdir -p "$TMP_REPO/vendor" "$TMP_REPO/plugin/mcp" "$TMP_REPO/lib/bash" "$TMP_REPO/scripts"
  cp "$LIB_DIR/"*.sh "$TMP_REPO/lib/bash/"
  cp "$SCRIPTS_DIR/build-mcp-aggregate.sh" "$TMP_REPO/scripts/"
}

teardown() {
  rm -rf "$TMP_REPO"
}

_vendor() {
  # _vendor <name> [os] [arch]
  local name="$1" os="${2:-linux}" arch="${3:-x86_64}"
  local vdir="$TMP_REPO/vendor/$name"
  mkdir -p "$vdir/bin"
  echo '#!/bin/sh' > "$vdir/bin/$name-$arch-$os"
  chmod +x "$vdir/bin/$name-$arch-$os"
  cat > "$vdir/mcp.config.json" <<EOF
{
  "mcpServers": {
    "buddy-$name": {
      "command": "{{BUDDY_BIN}}",
      "args": ["stdio"],
      "env": {}
    }
  }
}
EOF
}

@test "build-mcp-aggregate emits empty servers when no vendor entries" {
  run "$TMP_REPO/scripts/build-mcp-aggregate.sh" "$TMP_REPO" "$TMP_REPO/plugin/mcp/servers.json"
  [ "$status" -eq 0 ]
  run jq -r '.mcpServers | length' "$TMP_REPO/plugin/mcp/servers.json"
  [ "$output" = "0" ]
}

@test "build-mcp-aggregate picks up vendor entry and injects binary path" {
  local os; os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  local m; m="$(uname -m)"
  local arch; case "$m" in arm64|aarch64) arch=aarch64;; *) arch=x86_64;; esac
  _vendor "serena" "$os" "$arch"
  run "$TMP_REPO/scripts/build-mcp-aggregate.sh" "$TMP_REPO" "$TMP_REPO/plugin/mcp/servers.json"
  [ "$status" -eq 0 ]
  run jq -er '.mcpServers."buddy-serena".command' "$TMP_REPO/plugin/mcp/servers.json"
  [[ "$output" == *"vendor/serena/bin/serena-$arch-$os"* ]]
}

@test "build-mcp-aggregate skips vendor with no matching platform binary" {
  _vendor "nomatch" "other-os" "other-arch"
  run "$TMP_REPO/scripts/build-mcp-aggregate.sh" "$TMP_REPO" "$TMP_REPO/plugin/mcp/servers.json"
  [ "$status" -eq 0 ]
  run jq -er '.mcpServers."buddy-nomatch" // "absent"' "$TMP_REPO/plugin/mcp/servers.json"
  [ "$output" = "absent" ]
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `bats tests/bash/test_build_mcp_aggregate.bats`
Expected: all fail.

- [ ] **Step 3: Implement build-mcp-aggregate.sh**

Create `scripts/build-mcp-aggregate.sh`:
```bash
#!/usr/bin/env bash
# Aggregate vendor/*/mcp.config.json fragments into a single servers.json
# with absolute binary paths substituted for the {{BUDDY_BIN}} placeholder.
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=../lib/bash/log.sh
source "$SCRIPT_DIR/../lib/bash/log.sh"
# shellcheck source=../lib/bash/platform.sh
source "$SCRIPT_DIR/../lib/bash/platform.sh"

BUDDY_ROOT="${1:-$SCRIPT_DIR/..}"
OUTPUT="${2:-$BUDDY_ROOT/plugin/mcp/servers.json}"

mkdir -p "$(dirname "$OUTPUT")"
RESULT='{"mcpServers":{}}'

if [ ! -d "$BUDDY_ROOT/vendor" ]; then
  echo "$RESULT" | jq . > "$OUTPUT"
  info "no vendor/ directory — empty servers.json written to $OUTPUT"
  exit 0
fi

for vendor_dir in "$BUDDY_ROOT/vendor"/*/; do
  [ -d "$vendor_dir" ] || continue
  local_name="$(basename "$vendor_dir")"
  [ "$local_name" = ".gitkeep" ] && continue

  cfg="$vendor_dir/mcp.config.json"
  if [ ! -f "$cfg" ]; then
    warn "$local_name: no mcp.config.json — skipping"
    continue
  fi

  if ! bin="$(resolve_binary "$vendor_dir" "$local_name")"; then
    warn "$local_name: no binary for current platform — skipping"
    continue
  fi

  # Replace {{BUDDY_BIN}} in the vendor config with the absolute binary path.
  resolved="$(jq --arg bin "$bin" \
    '.mcpServers |= with_entries(.value.command = $bin)' "$cfg")"

  # Merge into RESULT
  RESULT="$(jq -s '.[0] * .[1]' \
    <(echo "$RESULT") <(echo "$resolved"))"

  info "registered MCP: $local_name ($bin)"
done

echo "$RESULT" | jq . > "$OUTPUT"
info "aggregate written: $OUTPUT"
```

- [ ] **Step 4: Make executable and run tests**

Run:
```bash
chmod +x scripts/build-mcp-aggregate.sh
bats tests/bash/test_build_mcp_aggregate.bats
```
Expected: `3 tests, 0 failures`.

- [ ] **Step 5: Commit**

```bash
git add scripts/build-mcp-aggregate.sh tests/bash/test_build_mcp_aggregate.bats
git commit -m "feat: add scripts/build-mcp-aggregate.sh with tests"
```

---

### Task 3.3: scripts/install.sh (core)

**Files:**
- Create: `scripts/install.sh`

This task creates the install.sh with enough logic to deploy rules, commands, agents, and skills. MCP and hooks come in Task 3.4.

- [ ] **Step 1: Write install.sh (partial — components a–d)**

Create `scripts/install.sh`:
```bash
#!/usr/bin/env bash
# buddy plugin installer. Idempotent. Deploys components from BUDDY_ROOT/plugin/
# to ~/.claude/ via symlinks (or copies with --copy).
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUDDY_ROOT="${BUDDY_ROOT_OVERRIDE:-$( cd "$SCRIPT_DIR/.." && pwd )}"
# shellcheck source=../lib/bash/log.sh
source "$BUDDY_ROOT/lib/bash/log.sh"
# shellcheck source=../lib/bash/platform.sh
source "$BUDDY_ROOT/lib/bash/platform.sh"
# shellcheck source=../lib/bash/json.sh
source "$BUDDY_ROOT/lib/bash/json.sh"
# shellcheck source=../lib/bash/symlink.sh
source "$BUDDY_ROOT/lib/bash/symlink.sh"

# Defaults
CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
INSTALL_MODE=symlink
FORCE=0
DRY_RUN=0
YES=0
ONLY=""
METADATA="$CLAUDE_DIR/.buddy-metadata.json"

usage() {
  cat <<EOF
Usage: $0 [--copy] [--force] [--dry-run] [--yes] [--only=types]

Options:
  --copy           Use cp -R instead of symlinks (breaks live link to BUDDY_ROOT)
  --force          Overwrite conflicting files/links
  --dry-run        Log planned actions without touching filesystem
  --yes            Auto-confirm prompts
  --only=<types>   Comma-separated subset: rules,commands,agents,skills,mcp,hooks
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --copy)      INSTALL_MODE=copy ;;
    --force)     FORCE=1 ;;
    --dry-run)   DRY_RUN=1 ;;
    --yes)       YES=1 ;;
    --only=*)    ONLY="${1#--only=}" ;;
    -h|--help)   usage; exit 0 ;;
    *) err "unknown flag: $1" ;;
  esac
  shift
done

want() {
  # want <type> → 0 if ONLY is empty or contains <type>
  local t="$1"
  [ -z "$ONLY" ] && return 0
  [[ ",$ONLY," == *",$t,"* ]]
}

action() {
  # Wrap a destructive command so it respects --dry-run
  if [ "$DRY_RUN" -eq 1 ]; then
    info "[dry-run] $*"
    return 0
  fi
  "$@"
}

deploy_file() {
  # deploy_file <src> <tgt>
  local src="$1" tgt="$2"
  case "$INSTALL_MODE" in
    symlink)
      if [ "$FORCE" -eq 1 ]; then
        action link_file --force "$src" "$tgt"
      else
        if ! action link_file "$src" "$tgt"; then
          err "conflict at $tgt (use --force to override)"
        fi
      fi
      ;;
    copy)
      action mkdir -p "$(dirname "$tgt")"
      action cp -f "$src" "$tgt"
      ;;
  esac
}

deploy_dir() {
  # deploy_dir <src-dir> <tgt-dir>
  local src="$1" tgt="$2"
  case "$INSTALL_MODE" in
    symlink)
      if [ "$FORCE" -eq 1 ]; then
        action link_dir --force "$src" "$tgt"
      else
        if ! action link_dir "$src" "$tgt"; then
          err "conflict at $tgt (use --force to override)"
        fi
      fi
      ;;
    copy)
      action mkdir -p "$tgt"
      action cp -Rf "$src/." "$tgt/"
      ;;
  esac
}

# ---- Preflight ----
info "buddy install starting (mode=$INSTALL_MODE, dry-run=$DRY_RUN)"
[ -d "$CLAUDE_DIR" ] || err "Claude Code directory not found at $CLAUDE_DIR"
command -v git >/dev/null || err "git is required"
command -v jq  >/dev/null || err "jq is required"

OS="$(detect_os)"; ARCH="$(detect_arch)"
info "platform: $OS/$ARCH"

# ---- Metadata skeleton ----
META_JSON='{
  "schema_version": 1,
  "buddy_version": "'"$(cat "$BUDDY_ROOT/VERSION")"'",
  "installed_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",
  "updated_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",
  "source_path": "'"$BUDDY_ROOT"'",
  "install_mode": "'"$INSTALL_MODE"'",
  "platform": {"os":"'"$OS"'","arch":"'"$ARCH"'"},
  "components": {}
}'

add_entry() {
  # add_entry <type> <strategy> <entry-json>
  local t="$1" s="$2" e="$3"
  META_JSON="$(jq \
    --arg t "$t" --arg s "$s" --argjson e "$e" \
    '(.components[$t] //= {"strategy":$s,"entries":[]})
     | .components[$t].entries += [$e]' \
    <<<"$META_JSON")"
}

# ---- Deploy: rules ----
if want rules; then
  info "deploying rules..."
  mkdir -p "$CLAUDE_DIR/rules"
  src_dir="$BUDDY_ROOT/plugin/rules"
  for f in "$src_dir"/*.md; do
    [ -f "$f" ] || continue
    name="$(basename "$f")"
    case "$name" in
      buddy-*.md) ;;
      *) warn "rules: $name does not use buddy- prefix — skipping"; continue ;;
    esac
    tgt="$CLAUDE_DIR/rules/$name"
    deploy_file "$f" "$tgt"
    add_entry rules file-symlink \
      "{\"target\":\"$tgt\",\"source\":\"$f\"}"
    info "  + $name"
  done
fi

# ---- Deploy: commands (single dir symlink) ----
if want commands; then
  info "deploying commands..."
  mkdir -p "$CLAUDE_DIR/commands"
  src_dir="$BUDDY_ROOT/plugin/commands"
  tgt="$CLAUDE_DIR/commands/buddy"
  deploy_dir "$src_dir" "$tgt"
  add_entry commands dir-symlink \
    "{\"target\":\"$tgt\",\"source\":\"$src_dir\"}"
  info "  + $tgt -> $src_dir"
fi

# ---- Deploy: agents ----
if want agents; then
  info "deploying agents..."
  mkdir -p "$CLAUDE_DIR/agents"
  src_dir="$BUDDY_ROOT/plugin/agents"
  for f in "$src_dir"/*.md; do
    [ -f "$f" ] || continue
    name="$(basename "$f")"
    case "$name" in
      buddy-*.md) ;;
      *) warn "agents: $name does not use buddy- prefix — skipping"; continue ;;
    esac
    tgt="$CLAUDE_DIR/agents/$name"
    deploy_file "$f" "$tgt"
    add_entry agents file-symlink \
      "{\"target\":\"$tgt\",\"source\":\"$f\"}"
    info "  + $name"
  done
fi

# ---- Deploy: skills (real dir + SKILL.md symlink) ----
if want skills; then
  info "deploying skills..."
  mkdir -p "$CLAUDE_DIR/skills"
  src_root="$BUDDY_ROOT/plugin/skills"
  for skill_dir in "$src_root"/*/; do
    [ -d "$skill_dir" ] || continue
    name="$(basename "$skill_dir")"
    case "$name" in
      buddy-*) ;;
      *) warn "skills: $name does not use buddy- prefix — skipping"; continue ;;
    esac
    src_md="$skill_dir/SKILL.md"
    [ -f "$src_md" ] || { warn "skills: $name has no SKILL.md — skipping"; continue; }
    tgt_dir="$CLAUDE_DIR/skills/$name"
    tgt_md="$tgt_dir/SKILL.md"
    action mkdir -p "$tgt_dir"
    deploy_file "$src_md" "$tgt_md"
    add_entry skills dir-with-symlinked-SKILL.md \
      "{\"target_dir\":\"$tgt_dir\",\"skill_md_symlink\":\"$tgt_md\",\"source\":\"$src_md\"}"
    info "  + $name"
  done
fi

# ---- Placeholders for mcp/hooks — filled in Task 3.4 ----
# (later tasks add deploy blocks here)

# ---- Write metadata ----
if [ "$DRY_RUN" -eq 0 ]; then
  echo "$META_JSON" | jq . > "$METADATA"
  info "metadata written: $METADATA"
else
  info "[dry-run] metadata would be written to $METADATA"
fi

info "buddy install complete."
```

- [ ] **Step 2: Make executable**

Run: `chmod +x scripts/install.sh`

- [ ] **Step 3: Smoke run in dry-run against fake HOME**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-test.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" \
  scripts/install.sh --dry-run
```
Expected: `[info] buddy install starting (mode=symlink, dry-run=1)`, followed by per-type deploy messages, ending with `[info] buddy install complete.`. Clean exit code 0.

Cleanup: `rm -rf "$TMP_HOME"`.

- [ ] **Step 4: Commit**

```bash
git add scripts/install.sh
git commit -m "feat: add scripts/install.sh (rules/commands/agents/skills)"
```

---

### Task 3.4: install.sh — MCP + hooks + metadata completion

**Files:**
- Modify: `scripts/install.sh`

- [ ] **Step 1: Insert MCP deploy block**

Edit `scripts/install.sh`. Replace the line:
```
# ---- Placeholders for mcp/hooks — filled in Task 3.4 ----
# (later tasks add deploy blocks here)
```

With:
```bash
# ---- Deploy: mcp ----
if want mcp; then
  info "deploying mcp..."
  "$BUDDY_ROOT/scripts/build-mcp-aggregate.sh" "$BUDDY_ROOT" "$BUDDY_ROOT/plugin/mcp/servers.json" >/dev/null
  agg="$BUDDY_ROOT/plugin/mcp/servers.json"
  mcp_file="$CLAUDE_DIR/.mcp.json"
  json_ensure_file "$mcp_file"
  names="$(jq -r '.mcpServers | keys[]' "$agg" 2>/dev/null || true)"
  server_names=()
  if [ -n "$names" ]; then
    while IFS= read -r key; do
      cmd="$(jq -er --arg k "$key" '.mcpServers[$k].command' "$agg")"
      args="$(jq -c --arg k "$key" '.mcpServers[$k].args // []' "$agg")"
      envj="$(jq -c --arg k "$key" '.mcpServers[$k].env  // {}' "$agg")"
      if [ "$DRY_RUN" -eq 1 ]; then
        info "[dry-run] merge mcp server $key ($cmd)"
      else
        json_merge_mcp "$mcp_file" "$key" "$cmd" "$args" "$envj"
        info "  + $key"
      fi
      server_names+=("$key")
    done <<< "$names"
  fi
  if [ "${#server_names[@]}" -gt 0 ]; then
    ids_json="$(printf '%s\n' "${server_names[@]}" | jq -R . | jq -s .)"
    META_JSON="$(jq \
      --arg target "$mcp_file" \
      --argjson names "$ids_json" \
      '.components.mcp = {"strategy":"json-merge","target_file":$target,"server_names":$names}' \
      <<<"$META_JSON")"
  fi
fi

# ---- Deploy: hooks ----
if want hooks; then
  info "deploying hooks..."
  hooks_src="$BUDDY_ROOT/plugin/hooks/hooks.json"
  settings="$CLAUDE_DIR/settings.json"
  json_ensure_file "$settings"
  ids=()
  if [ -f "$hooks_src" ]; then
    count="$(jq -r '.hooks | length // 0' "$hooks_src")"
    i=0
    while [ "$i" -lt "$count" ]; do
      hook="$(jq -c --argjson i "$i" '.hooks[$i]' "$hooks_src")"
      id="$(jq -r '.id // ""' <<<"$hook")"
      case "$id" in
        buddy:*) ;;
        *) warn "hooks: entry without buddy:* id — skipping"; i=$((i+1)); continue ;;
      esac
      if [ "$DRY_RUN" -eq 1 ]; then
        info "[dry-run] merge hook $id"
      else
        json_merge_hook "$settings" "$hook"
        info "  + $id"
      fi
      ids+=("$id")
      i=$((i+1))
    done
  fi
  if [ "${#ids[@]}" -gt 0 ]; then
    ids_json="$(printf '%s\n' "${ids[@]}" | jq -R . | jq -s .)"
    META_JSON="$(jq \
      --arg target "$settings" \
      --argjson ids "$ids_json" \
      '.components.hooks = {"strategy":"json-merge","target_file":$target,"ids":$ids}' \
      <<<"$META_JSON")"
  fi
fi
```

- [ ] **Step 2: Smoke run in fake HOME (end-to-end dry)**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-test.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" \
  scripts/install.sh --dry-run
```
Expected: sections for `rules`, `commands`, `agents`, `skills`, `mcp`, `hooks` appear; `[dry-run]` prefixed entries logged where applicable.

Cleanup: `rm -rf "$TMP_HOME"`.

- [ ] **Step 3: Real smoke run (no dry-run) in fake HOME**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-test.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" \
  scripts/install.sh --yes
cat "$TMP_HOME/.claude/.buddy-metadata.json" | jq .
```
Expected: metadata file present with `schema_version: 1`, `buddy_version: 0.1.0`, empty `components` (no content in plugin/ yet).

Cleanup: `rm -rf "$TMP_HOME"`.

- [ ] **Step 4: Commit**

```bash
git add scripts/install.sh
git commit -m "feat: install.sh deploys mcp and hooks, writes metadata"
```

---

## Phase 4: Uninstall / Doctor / Update

### Task 4.1: scripts/doctor.sh

**Files:**
- Create: `scripts/doctor.sh`

- [ ] **Step 1: Write doctor.sh**

Create `scripts/doctor.sh`:
```bash
#!/usr/bin/env bash
# Health check for buddy plugin install.
set -uo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUDDY_ROOT="${BUDDY_ROOT_OVERRIDE:-$( cd "$SCRIPT_DIR/.." && pwd )}"
# shellcheck source=../lib/bash/log.sh
source "$BUDDY_ROOT/lib/bash/log.sh"
# shellcheck source=../lib/bash/symlink.sh
source "$BUDDY_ROOT/lib/bash/symlink.sh"

CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
METADATA="$CLAUDE_DIR/.buddy-metadata.json"
VERBOSE=0
FIX=0
ONLY=""

PASS=0; WARN=0; FAIL=0

while [ $# -gt 0 ]; do
  case "$1" in
    --verbose|-v) VERBOSE=1 ;;
    --fix)        FIX=1 ;;
    --only=*)     ONLY="${1#--only=}" ;;
    -h|--help)    echo "Usage: $0 [--verbose] [--fix] [--only=sections]"; exit 0 ;;
    *) err "unknown flag: $1" ;;
  esac
  shift
done

want() { [ -z "$ONLY" ] || [[ ",$ONLY," == *",$1,"* ]]; }
ok()   { PASS=$((PASS+1)); [ "$VERBOSE" -eq 1 ] && echo "  ✅ $*"; }
warnv(){ WARN=$((WARN+1)); echo "  ⚠️  $*"; }
fail() { FAIL=$((FAIL+1)); echo "  ❌ $*"; }

# [1] Environment
if want env; then
  echo "[1] Environment"
  [ -d "$CLAUDE_DIR" ] && ok "CLAUDE_DIR=$CLAUDE_DIR" || fail "CLAUDE_DIR missing: $CLAUDE_DIR"
  command -v git >/dev/null && ok "git available" || fail "git missing"
  command -v jq  >/dev/null && ok "jq available"  || fail "jq missing"
fi

# [2] Repo state
if want repo; then
  echo "[2] Repo state"
  if [ -d "$BUDDY_ROOT/.git" ]; then
    ok "BUDDY_ROOT is a git repo"
    current="$(git -C "$BUDDY_ROOT" describe --tags --always 2>/dev/null || echo unknown)"
    ok "checkout: $current"
  else
    warnv "BUDDY_ROOT is not a git repo ($BUDDY_ROOT)"
  fi
  if [ -f "$BUDDY_ROOT/VERSION" ]; then
    ok "VERSION=$(cat "$BUDDY_ROOT/VERSION")"
  else
    fail "VERSION file missing"
  fi
fi

# [3] Metadata integrity
if want metadata; then
  echo "[3] Metadata integrity"
  if [ -f "$METADATA" ]; then
    ok "metadata present"
    if jq empty "$METADATA" 2>/dev/null; then ok "metadata is valid JSON"
    else fail "metadata is invalid JSON"; fi
    schema="$(jq -r '.schema_version // 0' "$METADATA")"
    [ "$schema" = "1" ] && ok "schema_version=1" || fail "unsupported schema_version: $schema"
    metaver="$(jq -r '.buddy_version // ""' "$METADATA")"
    fileve="$(cat "$BUDDY_ROOT/VERSION" 2>/dev/null || echo "")"
    [ "$metaver" = "$fileve" ] && ok "version match: $metaver" || warnv "metadata version ($metaver) != VERSION ($fileve)"
    src="$(jq -r '.source_path // ""' "$METADATA")"
    [ -d "$src" ] && ok "source_path resolves: $src" || fail "source_path missing: $src"
  else
    warnv "no metadata at $METADATA (buddy not installed here)"
  fi
fi

# [4] Component health
if want components && [ -f "$METADATA" ]; then
  echo "[4] Component health"
  for type in commands agents skills rules; do
    len="$(jq -r --arg t "$type" '.components[$t].entries | length // 0' "$METADATA")"
    [ "$len" = "0" ] && continue
    echo "  ($type: $len)"
    i=0
    while [ "$i" -lt "$len" ]; do
      entry="$(jq -c --arg t "$type" --argjson i "$i" '.components[$t].entries[$i]' "$METADATA")"
      case "$type" in
        commands|agents|rules)
          tgt="$(jq -r '.target' <<<"$entry")"
          src="$(jq -r '.source' <<<"$entry")"
          if verify_link "$tgt" "$src"; then
            ok "$type: $(basename "$tgt")"
          else
            fail "$type link invalid: $tgt"
            if [ "$FIX" -eq 1 ] && [ -e "$src" ]; then
              rm -f "$tgt"; ln -s "$src" "$tgt" && ok "  fixed: $tgt"
            fi
          fi
          ;;
        skills)
          sl="$(jq -r '.skill_md_symlink' <<<"$entry")"
          sr="$(jq -r '.source' <<<"$entry")"
          if verify_link "$sl" "$sr"; then
            ok "skills: $(basename "$(dirname "$sl")")"
          else
            fail "skill link invalid: $sl"
            if [ "$FIX" -eq 1 ] && [ -f "$sr" ]; then
              rm -f "$sl"; ln -s "$sr" "$sl" && ok "  fixed: $sl"
            fi
          fi
          ;;
      esac
      i=$((i+1))
    done
  done
fi

# [5] Hooks health
if want hooks && [ -f "$METADATA" ]; then
  echo "[5] Hooks health"
  settings="$(jq -r '.components.hooks.target_file // ""' "$METADATA")"
  if [ -n "$settings" ] && [ -f "$settings" ]; then
    expected="$(jq -r '.components.hooks.ids | length // 0' "$METADATA")"
    actual="$(jq -r '[.hooks[]? | select(.id // "" | startswith("buddy:"))] | length' "$settings")"
    [ "$expected" = "$actual" ] && ok "hook count match: $actual" || warnv "hooks expected=$expected actual=$actual"
  else
    ok "no hooks registered"
  fi
fi

# [6] MCP health
if want mcp && [ -f "$METADATA" ]; then
  echo "[6] MCP health"
  mcp_file="$(jq -r '.components.mcp.target_file // ""' "$METADATA")"
  if [ -n "$mcp_file" ] && [ -f "$mcp_file" ]; then
    while IFS= read -r name; do
      [ -z "$name" ] && continue
      cmd="$(jq -r --arg k "$name" '.mcpServers[$k].command // ""' "$mcp_file")"
      if [ -x "$cmd" ]; then
        ok "$name: $cmd"
      else
        fail "$name: binary not executable or missing: $cmd"
        if [ "$FIX" -eq 1 ] && [ -f "$cmd" ]; then
          chmod +x "$cmd" && ok "  fixed perms: $cmd"
        fi
      fi
    done < <(jq -r '.components.mcp.server_names[]?' "$METADATA")
  else
    ok "no mcp servers registered"
  fi
fi

# [7] Summary
echo
echo "Summary: ✅ $PASS   ⚠️ $WARN   ❌ $FAIL"
[ "$FAIL" -gt 0 ] && exit 2
[ "$WARN" -gt 0 ] && exit 1
exit 0
```

- [ ] **Step 2: Make executable**

Run: `chmod +x scripts/doctor.sh`

- [ ] **Step 3: Smoke run against a freshly-installed fake HOME**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-test.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" scripts/install.sh --yes
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" scripts/doctor.sh --verbose
```
Expected: Sections [1]-[6] print with ✅ entries; Summary shows 0 failures. Exit code 0.

Cleanup: `rm -rf "$TMP_HOME"`.

- [ ] **Step 4: Commit**

```bash
git add scripts/doctor.sh
git commit -m "feat: add scripts/doctor.sh (7-section health check)"
```

---

### Task 4.2: scripts/uninstall.sh

**Files:**
- Create: `scripts/uninstall.sh`

- [ ] **Step 1: Write uninstall.sh**

Create `scripts/uninstall.sh`:
```bash
#!/usr/bin/env bash
# Metadata-driven uninstall. Reverses strategies recorded at install time.
set -uo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUDDY_ROOT="${BUDDY_ROOT_OVERRIDE:-$( cd "$SCRIPT_DIR/.." && pwd )}"
# shellcheck source=../lib/bash/log.sh
source "$BUDDY_ROOT/lib/bash/log.sh"
# shellcheck source=../lib/bash/json.sh
source "$BUDDY_ROOT/lib/bash/json.sh"
# shellcheck source=../lib/bash/symlink.sh
source "$BUDDY_ROOT/lib/bash/symlink.sh"

CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
METADATA="$CLAUDE_DIR/.buddy-metadata.json"
DRY_RUN=0
FORCE=0
YES=0
KEEP_REPO=0

usage() {
  cat <<EOF
Usage: $0 [--dry-run] [--force] [--yes] [--keep-repo]
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run)   DRY_RUN=1 ;;
    --force)     FORCE=1 ;;
    --yes)       YES=1 ;;
    --keep-repo) KEEP_REPO=1 ;;
    -h|--help)   usage; exit 0 ;;
    *) err "unknown flag: $1" ;;
  esac
  shift
done

act() {
  if [ "$DRY_RUN" -eq 1 ]; then info "[dry-run] $*"; return 0; fi
  "$@"
}

if [ ! -f "$METADATA" ]; then
  info "no metadata at $METADATA — nothing to uninstall"
  exit 0
fi

if [ "$YES" -eq 0 ]; then
  counts="$(jq -r '
    [
      (.components.commands.entries | length // 0 | tostring) + " commands",
      (.components.agents.entries   | length // 0 | tostring) + " agents",
      (.components.skills.entries   | length // 0 | tostring) + " skills",
      (.components.rules.entries    | length // 0 | tostring) + " rules",
      (.components.hooks.ids        | length // 0 | tostring) + " hooks",
      (.components.mcp.server_names | length // 0 | tostring) + " mcp servers"
    ] | join(", ")' "$METADATA")"
  echo "This will remove: $counts"
  read -r -p "Proceed? [y/N] " ans
  case "$ans" in y|Y|yes|YES) ;; *) info "cancelled"; exit 0 ;; esac
fi

# file/dir symlinks (commands, agents, rules)
for t in commands agents rules; do
  len="$(jq -r --arg t "$t" '.components[$t].entries | length // 0' "$METADATA")"
  [ "$len" = "0" ] && continue
  i=0
  while [ "$i" -lt "$len" ]; do
    entry="$(jq -c --arg t "$t" --argjson i "$i" '.components[$t].entries[$i]' "$METADATA")"
    tgt="$(jq -r '.target' <<<"$entry")"
    src="$(jq -r '.source' <<<"$entry")"
    if [ "$FORCE" -eq 1 ]; then
      act rm -f "$tgt"
      info "  removed (forced): $tgt"
    elif act unlink_if_owned "$tgt" "$src" 2>/dev/null; then
      info "  removed: $tgt"
    else
      warn "$t: skip $tgt (not owned or missing)"
    fi
    i=$((i+1))
  done
done

# skills
len="$(jq -r '.components.skills.entries | length // 0' "$METADATA")"
if [ "$len" != "0" ]; then
  i=0
  while [ "$i" -lt "$len" ]; do
    entry="$(jq -c --argjson i "$i" '.components.skills.entries[$i]' "$METADATA")"
    sl="$(jq -r '.skill_md_symlink' <<<"$entry")"
    src="$(jq -r '.source' <<<"$entry")"
    tdir="$(jq -r '.target_dir' <<<"$entry")"
    if [ "$FORCE" -eq 1 ]; then
      act rm -f "$sl"
    elif ! act unlink_if_owned "$sl" "$src" 2>/dev/null; then
      warn "skills: skip $sl (not owned)"
    fi
    # rmdir if empty
    if [ -d "$tdir" ] && [ -z "$(ls -A "$tdir" 2>/dev/null)" ]; then
      act rmdir "$tdir"
      info "  removed skill dir: $tdir"
    else
      [ -d "$tdir" ] && warn "skills: dir not empty — keeping: $tdir"
    fi
    i=$((i+1))
  done
fi

# hooks
hook_target="$(jq -r '.components.hooks.target_file // ""' "$METADATA")"
if [ -n "$hook_target" ] && [ -f "$hook_target" ]; then
  mapfile -t hook_ids < <(jq -r '.components.hooks.ids[]?' "$METADATA")
  if [ "${#hook_ids[@]}" -gt 0 ]; then
    act json_remove_hooks_by_id "$hook_target" "${hook_ids[@]}"
    info "  removed ${#hook_ids[@]} hook(s) from $hook_target"
  fi
fi

# mcp
mcp_target="$(jq -r '.components.mcp.target_file // ""' "$METADATA")"
if [ -n "$mcp_target" ] && [ -f "$mcp_target" ]; then
  while IFS= read -r name; do
    [ -z "$name" ] && continue
    act json_remove_mcp "$mcp_target" "$name"
    info "  removed mcp: $name"
  done < <(jq -r '.components.mcp.server_names[]?' "$METADATA")
fi

# metadata
act rm -f "$METADATA"
info "metadata removed: $METADATA"

# offer repo removal
if [ "$KEEP_REPO" -eq 0 ] && [ -d "$HOME/.buddy" ] && [ "$YES" -eq 0 ]; then
  read -r -p "Remove ~/.buddy/? [y/N] " ans
  case "$ans" in
    y|Y|yes|YES) act rm -rf "$HOME/.buddy"; info "removed ~/.buddy" ;;
    *) info "keeping ~/.buddy" ;;
  esac
fi

info "buddy uninstall complete."
```

- [ ] **Step 2: Make executable**

Run: `chmod +x scripts/uninstall.sh`

- [ ] **Step 3: Full install → uninstall cycle in fake HOME**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-test.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" scripts/install.sh --yes
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" scripts/uninstall.sh --yes --keep-repo
ls "$TMP_HOME/.claude/" 2>/dev/null
```
Expected: `.claude/` remains but no `.buddy-metadata.json`, no `commands/buddy/`, no `buddy-*` files. Exit code 0.

Cleanup: `rm -rf "$TMP_HOME"`.

- [ ] **Step 4: Commit**

```bash
git add scripts/uninstall.sh
git commit -m "feat: add scripts/uninstall.sh (metadata-driven removal)"
```

---

### Task 4.3: scripts/update.sh

**Files:**
- Create: `scripts/update.sh`

- [ ] **Step 1: Write update.sh**

Create `scripts/update.sh`:
```bash
#!/usr/bin/env bash
# Update buddy to latest tag and re-install.
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BUDDY_ROOT="${BUDDY_ROOT_OVERRIDE:-$( cd "$SCRIPT_DIR/.." && pwd )}"
# shellcheck source=../lib/bash/log.sh
source "$BUDDY_ROOT/lib/bash/log.sh"

CLAUDE_DIR="${CLAUDE_DIR:-$HOME/.claude}"
METADATA="$CLAUDE_DIR/.buddy-metadata.json"
YES=0
DRY_RUN=0

while [ $# -gt 0 ]; do
  case "$1" in
    --yes)     YES=1 ;;
    --dry-run) DRY_RUN=1 ;;
    -h|--help) echo "Usage: $0 [--yes] [--dry-run]"; exit 0 ;;
    *) err "unknown flag: $1" ;;
  esac
  shift
done

[ -f "$METADATA" ] || err "buddy not installed (no $METADATA). Run install.sh first."
[ -d "$BUDDY_ROOT/.git" ] || err "$BUDDY_ROOT is not a git repo. Reinstall via curl."

cd "$BUDDY_ROOT"

info "fetching latest tags..."
git fetch --tags --quiet

CURRENT="$(cat VERSION)"
LATEST_TAG="$(git tag --list 'v*.*.*' --sort=-v:refname | head -n1)"
if [ -z "$LATEST_TAG" ]; then
  err "no release tags found"
fi
LATEST="${LATEST_TAG#v}"

if [ "$CURRENT" = "$LATEST" ]; then
  info "already up-to-date: v$CURRENT"
  exit 0
fi

info "update: v$CURRENT → $LATEST_TAG"

if [ "$YES" -eq 0 ]; then
  read -r -p "Proceed? [y/N] " ans
  case "$ans" in y|Y|yes|YES) ;; *) info "cancelled"; exit 0 ;; esac
fi

if [ "$DRY_RUN" -eq 1 ]; then
  info "[dry-run] would checkout $LATEST_TAG and run install.sh --force"
  exit 0
fi

git checkout -q "$LATEST_TAG"
info "checked out $LATEST_TAG"

CLAUDE_DIR="$CLAUDE_DIR" BUDDY_ROOT_OVERRIDE="$BUDDY_ROOT" \
  "$BUDDY_ROOT/scripts/install.sh" --force --yes

# refresh updated_at
jq --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
   --arg v "$LATEST" \
   '.updated_at=$ts | .buddy_version=$v' "$METADATA" \
   > "$METADATA.tmp" && mv "$METADATA.tmp" "$METADATA"

CLAUDE_DIR="$CLAUDE_DIR" BUDDY_ROOT_OVERRIDE="$BUDDY_ROOT" \
  "$BUDDY_ROOT/scripts/doctor.sh"

info "buddy update complete: $LATEST_TAG"
```

- [ ] **Step 2: Make executable**

Run: `chmod +x scripts/update.sh`

- [ ] **Step 3: Verify dry-run behavior when no tags exist**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-test.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" scripts/install.sh --yes
CLAUDE_DIR="$TMP_HOME/.claude" BUDDY_ROOT_OVERRIDE="$(pwd)" scripts/update.sh --yes --dry-run || true
```
Expected: either "no release tags found" (pre-v0.1.0 tag) or "already up-to-date" (if tag exists). Exit code either 0 or 1.

Cleanup: `rm -rf "$TMP_HOME"`.

- [ ] **Step 4: Commit**

```bash
git add scripts/update.sh
git commit -m "feat: add scripts/update.sh (tag-based upgrade)"
```

---

## Phase 5: Root Entry Points

### Task 5.1: Root install.sh and uninstall.sh

**Files:**
- Create: `install.sh` (root)
- Create: `uninstall.sh` (root)

- [ ] **Step 1: Write root install.sh**

Create `install.sh` (in repo root):
```bash
#!/usr/bin/env bash
# buddy plugin — curl | bash entry point.
# Clones ~/.buddy and delegates to scripts/install.sh.
set -euo pipefail

BUDDY_REF="${BUDDY_REF:-main}"
BUDDY_REPO="${BUDDY_REPO:-https://github.com/buddy-author/buddy.git}"
BUDDY_HOME="${BUDDY_HOME:-$HOME/.buddy}"

info() { printf '[buddy-install] %s\n' "$*" >&2; }
err()  { printf '[buddy-install][err] %s\n' "$*" >&2; exit 1; }

command -v git >/dev/null || err "git is required"
command -v jq  >/dev/null || err "jq is required (brew install jq / apt-get install jq)"

# If BUDDY_REPO points to a local path (CI self-test), link it in place.
if [ -d "$BUDDY_REPO" ]; then
  info "local repo mode: linking $BUDDY_REPO → $BUDDY_HOME"
  rm -rf "$BUDDY_HOME"
  ln -sfn "$BUDDY_REPO" "$BUDDY_HOME"
else
  if [ -d "$BUDDY_HOME/.git" ]; then
    info "existing clone at $BUDDY_HOME — fetching $BUDDY_REF"
    git -C "$BUDDY_HOME" fetch --tags --quiet
    git -C "$BUDDY_HOME" checkout -q "$BUDDY_REF"
  else
    info "cloning $BUDDY_REPO @ $BUDDY_REF → $BUDDY_HOME"
    git clone --depth=1 --branch "$BUDDY_REF" "$BUDDY_REPO" "$BUDDY_HOME"
  fi
fi

info "delegating to $BUDDY_HOME/scripts/install.sh $*"
exec "$BUDDY_HOME/scripts/install.sh" "$@"
```

- [ ] **Step 2: Write root uninstall.sh**

Create `uninstall.sh` (in repo root):
```bash
#!/usr/bin/env bash
# Thin wrapper that locates scripts/uninstall.sh.
set -euo pipefail

BUDDY_HOME="${BUDDY_HOME:-$HOME/.buddy}"
if [ -x "$BUDDY_HOME/scripts/uninstall.sh" ]; then
  exec "$BUDDY_HOME/scripts/uninstall.sh" "$@"
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
if [ -x "$SCRIPT_DIR/scripts/uninstall.sh" ]; then
  exec "$SCRIPT_DIR/scripts/uninstall.sh" "$@"
fi

echo "[buddy-uninstall][err] could not find scripts/uninstall.sh" >&2
exit 1
```

- [ ] **Step 3: Make executable**

Run: `chmod +x install.sh uninstall.sh`

- [ ] **Step 4: Smoke run root install.sh in local-repo mode**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-test.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"
HOME="$TMP_HOME" BUDDY_REPO="$(pwd)" ./install.sh --yes
ls -la "$TMP_HOME/.buddy" "$TMP_HOME/.claude/.buddy-metadata.json"
```
Expected: `$TMP_HOME/.buddy` is a symlink to repo; metadata present.

Cleanup: `rm -rf "$TMP_HOME"`.

- [ ] **Step 5: Commit**

```bash
git add install.sh uninstall.sh
git commit -m "feat: add root install.sh / uninstall.sh entry points"
```

---

## Phase 6: CI + Release Helpers

### Task 6.1: Release helper scripts

**Files:**
- Create: `scripts/release/build.sh`
- Create: `scripts/release/checksums.sh`
- Create: `scripts/release/changelog-extract.sh`

- [ ] **Step 1: Write build.sh (no-op stub)**

Create `scripts/release/build.sh`:
```bash
#!/usr/bin/env bash
# Self-built binary producer. v0.1.0: no-op.
set -euo pipefail

TARGET=""
while [ $# -gt 0 ]; do
  case "$1" in
    --target) TARGET="$2"; shift 2 ;;
    --target=*) TARGET="${1#--target=}"; shift ;;
    *) shift ;;
  esac
done

mkdir -p dist
echo "[build] target=${TARGET:-<unspecified>}"
echo "[build] no self-built artefacts in v$(cat VERSION); skipping."
```

- [ ] **Step 2: Write checksums.sh**

Create `scripts/release/checksums.sh`:
```bash
#!/usr/bin/env bash
# Generate SHA256 manifest for files under a directory.
# Usage: checksums.sh <dir>
set -euo pipefail

DIR="${1:-./artifacts}"
if [ ! -d "$DIR" ]; then
  exit 0  # empty output is valid
fi

# Prefer shasum (macOS) then sha256sum (linux)
if command -v sha256sum >/dev/null; then
  (cd "$DIR" && find . -type f -not -name SHA256SUMS | sort | xargs -r sha256sum)
elif command -v shasum >/dev/null; then
  (cd "$DIR" && find . -type f -not -name SHA256SUMS | sort | xargs -r shasum -a 256)
else
  echo "[checksums] no sha256 tool found" >&2
  exit 1
fi
```

- [ ] **Step 3: Write changelog-extract.sh**

Create `scripts/release/changelog-extract.sh`:
```bash
#!/usr/bin/env bash
# Extract the section for a given tag (vX.Y.Z) from CHANGELOG.md.
# Usage: changelog-extract.sh vX.Y.Z [CHANGELOG_PATH]
set -euo pipefail

TAG="${1:?usage: $0 vX.Y.Z [changelog-path]}"
FILE="${2:-CHANGELOG.md}"
VER="${TAG#v}"

awk -v ver="$VER" '
  BEGIN { inside=0 }
  /^## \[/ {
    if (inside) exit
    if ($0 ~ "\\[" ver "\\]") { inside=1; print; next }
  }
  inside { print }
' "$FILE"
```

- [ ] **Step 4: Make executable and test**

Run:
```bash
chmod +x scripts/release/build.sh scripts/release/checksums.sh scripts/release/changelog-extract.sh
scripts/release/build.sh --target linux-x86_64
scripts/release/changelog-extract.sh v0.1.0
```
Expected: build.sh prints skip message; changelog-extract emits the `## [0.1.0]` section.

- [ ] **Step 5: Commit**

```bash
git add scripts/release/
git commit -m "feat: add release helpers (build/checksums/changelog-extract)"
```

---

### Task 6.2: .github/workflows/ci.yml

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write ci.yml**

Create `.github/workflows/ci.yml`:
```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: install tools
        run: sudo apt-get update && sudo apt-get install -y shellcheck jq bats
      - name: shellcheck
        run: |
          find scripts lib/bash install.sh uninstall.sh -type f \
               \( -name '*.sh' -o -name 'install.sh' -o -name 'uninstall.sh' \) -print0 \
            | xargs -0 shellcheck -x
      - name: jq syntax
        run: find plugin vendor -name '*.json' -print0 | xargs -0 -r -n1 jq empty
      - name: manifest schema
        run: scripts/validate-manifest.sh plugin/.claude-plugin/plugin.json
      - name: bats unit tests
        run: bats tests/bash/

  smoke-install:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: install tools
        run: sudo apt-get update && sudo apt-get install -y jq
      - name: prepare fake HOME
        run: |
          mkdir -p /tmp/fake-home/.claude
          chmod +x install.sh uninstall.sh scripts/*.sh scripts/release/*.sh
      - name: dry-run
        run: HOME=/tmp/fake-home BUDDY_REPO="$(pwd)" ./install.sh --dry-run
      - name: real install
        run: HOME=/tmp/fake-home BUDDY_REPO="$(pwd)" ./install.sh --yes
      - name: doctor
        run: HOME=/tmp/fake-home /tmp/fake-home/.buddy/scripts/doctor.sh
      - name: uninstall
        run: HOME=/tmp/fake-home /tmp/fake-home/.buddy/scripts/uninstall.sh --yes --keep-repo
      - name: confirm clean removal
        run: |
          test ! -f /tmp/fake-home/.claude/.buddy-metadata.json
          test ! -e /tmp/fake-home/.claude/commands/buddy
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add lint + smoke-install workflow"
```

---

### Task 6.3: .github/workflows/release.yml

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Write release.yml**

Create `.github/workflows/release.yml`:
```yaml
name: release

on:
  push:
    tags: ['v*.*.*']

jobs:
  validate:
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.v.outputs.version }}
    steps:
      - uses: actions/checkout@v4
      - id: v
        name: tag ↔ VERSION
        run: |
          TAG="${GITHUB_REF#refs/tags/v}"
          FILE=$(cat VERSION)
          [ "$TAG" = "$FILE" ] || { echo "tag v$TAG != VERSION $FILE"; exit 1; }
          echo "version=$TAG" >> "$GITHUB_OUTPUT"
      - name: CHANGELOG has section
        run: grep -q "^## \[${GITHUB_REF_NAME#v}\]" CHANGELOG.md

  build:
    needs: validate
    strategy:
      fail-fast: false
      matrix:
        include:
          - { os: ubuntu-latest, target: linux-x86_64 }
          - { os: ubuntu-latest, target: linux-aarch64 }
          - { os: macos-latest,  target: darwin-x86_64 }
          - { os: macos-latest,  target: darwin-aarch64 }
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - name: build
        run: scripts/release/build.sh --target ${{ matrix.target }}
      - name: upload artefact (if any)
        if: hashFiles('dist/*') != ''
        uses: actions/upload-artifact@v4
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
        continue-on-error: true
      - name: checksums
        run: |
          mkdir -p ./artifacts
          scripts/release/checksums.sh ./artifacts > SHA256SUMS
          cat SHA256SUMS
      - name: extract changelog
        run: scripts/release/changelog-extract.sh "$GITHUB_REF_NAME" > release-notes.md
      - name: create release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            ./artifacts/**
            SHA256SUMS
          body_path: release-notes.md
          prerelease: ${{ contains(github.ref_name, '-rc') || contains(github.ref_name, '-beta') }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add tag-triggered release workflow"
```

---

## Phase 7: Integration Test

### Task 7.1: Local end-to-end verification

- [ ] **Step 1: Run full cycle locally in fake HOME**

Run:
```bash
TMP_HOME="$(mktemp -d -t buddy-e2e.XXXXXX)"
mkdir -p "$TMP_HOME/.claude"

# Install
HOME="$TMP_HOME" BUDDY_REPO="$(pwd)" ./install.sh --yes

# Verify metadata
jq . "$TMP_HOME/.claude/.buddy-metadata.json"

# Doctor
HOME="$TMP_HOME" "$TMP_HOME/.buddy/scripts/doctor.sh" --verbose

# Re-install (force) — should be idempotent
HOME="$TMP_HOME" BUDDY_REPO="$(pwd)" ./install.sh --force --yes

# Uninstall
HOME="$TMP_HOME" "$TMP_HOME/.buddy/scripts/uninstall.sh" --yes --keep-repo

# Verify cleanup
test ! -f "$TMP_HOME/.claude/.buddy-metadata.json" && echo "metadata removed: OK"
test ! -e "$TMP_HOME/.claude/commands/buddy" && echo "commands removed: OK"

# Cleanup
rm -rf "$TMP_HOME"
```
Expected: every verification line prints OK or a valid metadata JSON. No errors.

- [ ] **Step 2: Run all bats tests**

Run: `bats tests/bash/`
Expected: all tests pass.

- [ ] **Step 3: Run shellcheck**

Run:
```bash
find scripts lib/bash install.sh uninstall.sh -type f \
     \( -name '*.sh' -o -name 'install.sh' -o -name 'uninstall.sh' \) -print0 \
  | xargs -0 shellcheck -x
```
Expected: exit code 0, no warnings.

- [ ] **Step 4: Validate manifest**

Run: `scripts/validate-manifest.sh plugin/.claude-plugin/plugin.json`
Expected: `[info] manifest valid`.

- [ ] **Step 5: No commit needed**

This task is verification only. Any fixes surface issues to address in a prior task and commit there.

---

## Phase 8: v0.1.0 Release Dry-Run

### Task 8.1: Tag and verify release workflow

This task tags v0.1.0 **locally** to preview behavior. Push to remote is a separate, explicit action.

- [ ] **Step 1: Verify clean working tree**

Run: `git status`
Expected: no uncommitted changes. If there are, commit them first.

- [ ] **Step 2: Confirm VERSION + CHANGELOG consistency**

Run:
```bash
cat VERSION
grep -q "^## \[0.1.0\]" CHANGELOG.md && echo "CHANGELOG section present: OK"
```
Expected: `0.1.0` and `CHANGELOG section present: OK`.

- [ ] **Step 3: Tag v0.1.0 locally (not pushed yet)**

Run: `git tag v0.1.0`

- [ ] **Step 4: Verify CI workflow files parse**

Run:
```bash
find .github/workflows -name '*.yml' -print0 | while IFS= read -r -d '' f; do
  python3 -c "import yaml,sys; yaml.safe_load(open('$f'))" && echo "yaml ok: $f"
done
```
Expected: all `yaml ok:` lines; exit code 0.

- [ ] **Step 5: Preview release-notes extraction**

Run: `scripts/release/changelog-extract.sh v0.1.0`
Expected: output includes `## [0.1.0] - 2026-04-24` header and the Added/Notes subsections.

- [ ] **Step 6: Ready for remote push (do NOT push in this plan execution)**

The tag is local. To publish:
```bash
git push origin main
git push origin v0.1.0
```

These two commands trigger `release.yml` on GitHub. This plan stops before pushing — final release is a human decision, performed outside the scope of this scaffold build.

Report completion. The scaffold is ready.

- [ ] **Step 7: If anything failed, roll back the local tag**

```bash
git tag -d v0.1.0  # only if the scaffold needs more work before release
```

---

## Appendix — File inventory created by this plan

| Path | Purpose |
|------|---------|
| `.gitignore` | Ignore build artefacts and editor cruft |
| `VERSION` | Semver source of truth |
| `CHANGELOG.md` | Release notes (Keep-a-Changelog) |
| `plugin/.claude-plugin/plugin.json` | Claude Code plugin manifest |
| `plugin/hooks/hooks.json` | Hook declarations (empty in v0.1.0) |
| `plugin/{commands,agents,skills,rules,mcp}/.gitkeep` | Empty component slots |
| `vendor/README.md` | Convention for externally-built MCP binaries |
| `lib/README.md` | Library index |
| `lib/bash/log.sh` | Log helpers |
| `lib/bash/platform.sh` | OS/arch detection + binary resolution |
| `lib/bash/json.sh` | jq-based merge/patch helpers |
| `lib/bash/symlink.sh` | Idempotent symlink utilities |
| `scripts/validate-manifest.sh` | Manifest schema check |
| `scripts/build-mcp-aggregate.sh` | vendor/* → plugin/mcp/servers.json builder |
| `scripts/install.sh` | Core installer |
| `scripts/uninstall.sh` | Metadata-driven uninstaller |
| `scripts/update.sh` | Tag-based upgrade |
| `scripts/doctor.sh` | Health check |
| `scripts/release/build.sh` | Self-built binary producer (stub in v0.1.0) |
| `scripts/release/checksums.sh` | SHA256SUMS generator |
| `scripts/release/changelog-extract.sh` | Release-notes extractor |
| `install.sh` | Root entry (curl \| bash target) |
| `uninstall.sh` | Root uninstall wrapper |
| `tests/bash/test_helper.bash` | Bats test helpers |
| `tests/bash/test_*.bats` | Unit tests for lib/bash + script smoke |
| `.github/workflows/ci.yml` | Lint + smoke install on PR/push |
| `.github/workflows/release.yml` | Tag-triggered GitHub Release |
