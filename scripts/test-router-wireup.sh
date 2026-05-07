#!/usr/bin/env bash
# test-router-wireup.sh — verify the router skill wire-up invariants.
#
# This script enforces structural invariants of the buddy plugin's router-based
# dispatch model so that future PRs cannot accidentally break the routing
# topology (e.g. by reintroducing per-command skill stubs, leaving orphaned
# commands without a target PROCEDURE, or putting commands at a path Claude
# Code does not auto-discover from).
#
# Source of truth for slash commands is plugin/commands/*.md (auto-discovered).
# plugin.json must NOT carry a `commands` field — Claude Code's plugin schema
# rejects it (see commit log for the validation failure that motivated this).
#
# Run manually:    bash scripts/test-router-wireup.sh
# Run via Make:    make test-routing
#
# Exit codes:
#   0 — all invariants hold
#   1 — at least one invariant violated (details printed above the summary)

set -euo pipefail

# Always run from repo root regardless of how the script is invoked.
cd "$(git rev-parse --show-toplevel)"

PLUGIN_JSON="plugin/.claude-plugin/plugin.json"
SKILLS_DIR="plugin/skills"
COMMANDS_DIR="plugin/commands"
ROUTER_SKILL="plugin/skills/router/SKILL.md"

# Composition commands are exempt from the "single mode" invariant — their
# target PROCEDURE is determined dynamically from $ARGUMENTS, not hardcoded.
COMPOSITION_COMMANDS=("run" "chain" "parallel")

PASS_COUNT=0
FAIL_COUNT=0
FAILURES=()

pass() {
    echo "OK: $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo "FAIL: $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    FAILURES+=("$1")
}

is_composition() {
    local name="$1"
    local c
    for c in "${COMPOSITION_COMMANDS[@]}"; do
        if [ "$name" = "$c" ]; then
            return 0
        fi
    done
    return 1
}

# --- Check 1: SKILL.md count -------------------------------------------------
# Exactly one SKILL.md should exist (the router) — all other skills use
# PROCEDURE.md. Any extra SKILL.md likely means a per-skill stub was
# reintroduced, which breaks the lazy-load model.
skill_md_files=$(find "$SKILLS_DIR" -name SKILL.md)
skill_md_count=$(printf '%s\n' "$skill_md_files" | grep -c . || true)
if [ "$skill_md_count" = "1" ] && [ "$skill_md_files" = "$ROUTER_SKILL" ]; then
    pass "exactly 1 SKILL.md exists and it is router/SKILL.md"
else
    fail "expected exactly 1 SKILL.md ($ROUTER_SKILL); found $skill_md_count: $skill_md_files"
fi

# --- Check 2: PROCEDURE.md count --------------------------------------------
# 79 PROCEDURE.md files cover the full lifecycle catalog (78 baseline + 1
# Phase 1 Task 1.1 define-tech-stack). Drift here means a skill was added or
# removed without updating tests or docs. Bump this count when Phase 1 Tasks
# 1.2 / 1.3 / 1.4 land (→ 80, 81, 82).
procedure_count=$(find "$SKILLS_DIR" -name PROCEDURE.md | wc -l | tr -d ' ')
if [ "$procedure_count" = "84" ]; then
    pass "PROCEDURE.md count is 84"
else
    fail "expected 84 PROCEDURE.md files, found $procedure_count"
fi

# --- Check 3: plugin.json must NOT declare a `commands` field ----------------
# Claude Code's plugin schema does not accept a commands array (verified via
# `claude plugin validate`). Slash commands are auto-discovered from
# plugin/commands/*.md — adding commands to plugin.json breaks installation.
if jq -e 'has("commands")' "$PLUGIN_JSON" > /dev/null 2>&1; then
    fail "plugin.json must not contain a 'commands' field — slash commands are auto-discovered from $COMMANDS_DIR/"
else
    pass "plugin.json has no 'commands' field (slash commands auto-discovered)"
fi

# --- Check 4: command md count >= 30 (defensive lower bound) ----------------
# Current public surface is 30; this guards against accidental command
# deletion. Adding new commands is fine.
command_md_count=$(find "$COMMANDS_DIR" -mindepth 1 -maxdepth 1 -name "*.md" | wc -l | tr -d ' ')
if [ "$command_md_count" -ge 30 ]; then
    pass "command md count is $command_md_count (>= 30)"
else
    fail "command md count is $command_md_count, expected >= 30"
fi

# --- Check 5: command md files live at $COMMANDS_DIR/<name>.md (no nesting) -
# Nested directories (e.g. plugin/commands/buddy/<name>.md) cause Claude Code
# to surface commands as /<plugin>:<dir>:<name> instead of /<plugin>:<name>.
nested_count=$(find "$COMMANDS_DIR" -mindepth 2 -name "*.md" | wc -l | tr -d ' ')
if [ "$nested_count" = "0" ]; then
    pass "no nested command md files (slash names won't get extra namespace prefix)"
else
    fail "$nested_count command md files are nested under $COMMANDS_DIR/ subdirs:"
    find "$COMMANDS_DIR" -mindepth 2 -name "*.md"
fi

# --- Check 6: every non-composition command md targets an existing PROCEDURE
# Each single-mode command md must declare both `mode: \`single\`` and
# `target PROCEDURE: \`<X>\``, and the referenced PROCEDURE.md must exist.
single_mode_failures=()
while IFS= read -r md_file; do
    name=$(basename "$md_file" .md)
    if is_composition "$name"; then
        continue
    fi
    if ! grep -qF 'mode: `single`' "$md_file"; then
        single_mode_failures+=("$name: missing 'mode: \`single\`' literal")
        continue
    fi
    target=$(grep -E 'target PROCEDURE: `[^`]+`' "$md_file" | head -1 \
        | sed -E 's/.*target PROCEDURE: `([^`]+)`.*/\1/' || true)
    if [ -z "$target" ]; then
        single_mode_failures+=("$name: missing 'target PROCEDURE: \`<X>\`' line")
        continue
    fi
    if [ "$target" = "\$ARGUMENTS" ]; then
        # composition-style dynamic target — should have been filtered above
        continue
    fi
    if [ ! -f "$SKILLS_DIR/$target/PROCEDURE.md" ]; then
        single_mode_failures+=("$name: target PROCEDURE '$target' not found at $SKILLS_DIR/$target/PROCEDURE.md")
    fi
done < <(find "$COMMANDS_DIR" -mindepth 1 -maxdepth 1 -name "*.md")

if [ "${#single_mode_failures[@]}" -eq 0 ]; then
    pass "every non-composition command md targets an existing PROCEDURE"
else
    fail "single-mode command target issues:"
    for msg in "${single_mode_failures[@]}"; do
        echo "    - $msg"
    done
fi

# --- Check 7: composition commands have correct mode keywords ---------------
declare -a composition_failures=()
for entry in "run:single" "chain:chain" "parallel:parallel"; do
    name="${entry%%:*}"
    expected_mode="${entry##*:}"
    md_file="$COMMANDS_DIR/$name.md"
    if [ ! -f "$md_file" ]; then
        composition_failures+=("$name.md missing")
        continue
    fi
    if ! grep -qF "mode: \`$expected_mode\`" "$md_file"; then
        composition_failures+=("$name.md missing 'mode: \`$expected_mode\`'")
    fi
done

if [ "${#composition_failures[@]}" -eq 0 ]; then
    pass "composition commands (run/chain/parallel) have correct mode keywords"
else
    fail "composition command issues:"
    for msg in "${composition_failures[@]}"; do
        echo "    - $msg"
    done
fi

# --- Check 8: every command md invokes the router skill ---------------------
# A command md whose body does not invoke `router` would either route
# through a different skill or fall through to fuzzy auto-matching. Ensures
# the dispatch contract is uniform.
no_router_failures=()
while IFS= read -r md_file; do
    if ! grep -qE '`?router`?' "$md_file"; then
        no_router_failures+=("$md_file")
    fi
done < <(find "$COMMANDS_DIR" -mindepth 1 -maxdepth 1 -name "*.md")

if [ "${#no_router_failures[@]}" -eq 0 ]; then
    pass "every command md invokes the 'router' skill"
else
    fail "${#no_router_failures[@]} command md files don't reference 'router':"
    for f in "${no_router_failures[@]}"; do
        echo "    - $f"
    done
fi

# --- Check 9: router/SKILL.md description length is bounded -----------------
# Sanity bound to prevent reverting to per-skill description bloat. Current
# length is ~170 chars; cap at 250 leaves headroom for legitimate edits.
desc_line=$(grep -E '^description:' "$ROUTER_SKILL" | head -1)
desc_len=${#desc_line}
if [ "$desc_len" -le 250 ] && [ "$desc_len" -gt 0 ]; then
    pass "router/SKILL.md description line is $desc_len chars (<= 250)"
else
    fail "router/SKILL.md description length $desc_len out of bounds (1..250)"
fi

# --- Check 10: PROCEDURE.md files have no skill-shaped frontmatter ----------
# Auto-discovery is keyed off YAML frontmatter shape (name + description).
# PROCEDURE.md files must NOT carry frontmatter so that only router/SKILL.md
# is auto-discovered. Re-introducing frontmatter here would revive the
# routing collision and the illusory token-reduction regression.
bad_proc=()
while IFS= read -r f; do
    if [ "$(head -1 "$f")" = "---" ]; then
        bad_proc+=("$f")
    fi
done < <(find "$SKILLS_DIR" -name PROCEDURE.md)

if [ "${#bad_proc[@]}" -eq 0 ]; then
    pass "no PROCEDURE.md has skill-shaped YAML frontmatter (auto-discovery off)"
else
    fail "${#bad_proc[@]} PROCEDURE.md files still have YAML frontmatter (auto-discovery active):"
    for p in "${bad_proc[@]}"; do
        echo "    - $p"
    done
fi

# --- Summary -----------------------------------------------------------------
TOTAL=$((PASS_COUNT + FAIL_COUNT))
echo ""
echo "$PASS_COUNT/$TOTAL checks passed"

if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
fi
exit 0
