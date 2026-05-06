#!/usr/bin/env bash
# test-router-wireup.sh — verify the router skill wire-up invariants.
#
# This script enforces structural invariants of the buddy plugin's router-based
# dispatch model so that future PRs cannot accidentally break the routing
# topology (e.g. by reintroducing per-command skill stubs or by leaving
# orphaned commands without a target PROCEDURE).
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

# Aggregate failures so we can report all problems in one run instead of
# bailing on the first one. Each FAIL line is printed immediately for
# locality with its check; the summary shows the total at the end.
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
# 78 PROCEDURE.md files cover the full lifecycle catalog. Drift here means
# a skill was added/removed without updating tests or docs.
procedure_count=$(find "$SKILLS_DIR" -name PROCEDURE.md | wc -l | tr -d ' ')
if [ "$procedure_count" = "78" ]; then
    pass "PROCEDURE.md count is 78"
else
    fail "expected 78 PROCEDURE.md files, found $procedure_count"
fi

# --- Check 3: All plugin.json commands route through router -----------------
# A command with skill != "router" would bypass the dispatcher and break the
# uniform routing model. Should always be 0.
non_router_count=$(jq '[.commands[] | select(.skill != "router")] | length' "$PLUGIN_JSON")
if [ "$non_router_count" = "0" ]; then
    pass "all plugin.json commands route through 'router' skill"
else
    fail "expected 0 non-router commands; found $non_router_count"
    jq -r '.commands[] | select(.skill != "router") | "  - \(.name) -> \(.skill)"' "$PLUGIN_JSON" || true
fi

# --- Check 4: plugin.json command count >= 30 (defensive lower bound) -------
# The current public surface is 30; this guards against accidental command
# deletion. Adding new commands is fine.
command_count=$(jq '.commands | length' "$PLUGIN_JSON")
if [ "$command_count" -ge 30 ]; then
    pass "plugin.json command count is $command_count (>= 30)"
else
    fail "plugin.json command count is $command_count, expected >= 30"
fi

# --- Check 5: every plugin.json command has a corresponding md file ---------
missing_files=()
while IFS= read -r name; do
    if [ ! -f "$COMMANDS_DIR/$name.md" ]; then
        missing_files+=("$name")
    fi
done < <(jq -r '.commands[].name' "$PLUGIN_JSON")

if [ "${#missing_files[@]}" -eq 0 ]; then
    pass "every plugin.json command has a matching md file in $COMMANDS_DIR/"
else
    fail "missing command md files for: ${missing_files[*]}"
fi

# --- Check 6: every single-mode command targets an existing PROCEDURE ------
# Composition commands (run/chain/parallel) are exempt — they resolve targets
# dynamically. All others must declare a literal `target PROCEDURE: \`X\``
# and `mode: \`single\`` and the referenced PROCEDURE.md must exist.
single_mode_failures=()
while IFS= read -r name; do
    if is_composition "$name"; then
        continue
    fi
    md_file="$COMMANDS_DIR/$name.md"
    if [ ! -f "$md_file" ]; then
        # already reported in check 5
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
    # The run.md command uses `$ARGUMENTS` as a dynamic placeholder — already
    # filtered above, but defensively skip non-static targets.
    if [ "$target" = "\$ARGUMENTS" ]; then
        continue
    fi
    if [ ! -f "$SKILLS_DIR/$target/PROCEDURE.md" ]; then
        single_mode_failures+=("$name: target PROCEDURE '$target' not found at $SKILLS_DIR/$target/PROCEDURE.md")
    fi
done < <(jq -r '.commands[].name' "$PLUGIN_JSON")

if [ "${#single_mode_failures[@]}" -eq 0 ]; then
    pass "every single-mode command targets an existing PROCEDURE"
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

# --- Check 8: router/SKILL.md description length is bounded ------------------
# Sanity bound to prevent reverting to per-skill description bloat. Current
# length is ~170 chars; cap at 250 leaves headroom for legitimate edits.
desc_line=$(grep -E '^description:' "$ROUTER_SKILL" | head -1)
desc_len=${#desc_line}
if [ "$desc_len" -le 250 ] && [ "$desc_len" -gt 0 ]; then
    pass "router/SKILL.md description line is $desc_len chars (<= 250)"
else
    fail "router/SKILL.md description length $desc_len out of bounds (1..250)"
fi

# --- Check 9: description drift between plugin.json and md frontmatter ------
# Each plugin.json command's description must byte-match the corresponding
# md file's frontmatter `description:` value. Catches Task 0.1-style drift
# from sneaking back in.
drift_failures=()
while IFS= read -r name; do
    md_file="$COMMANDS_DIR/$name.md"
    if [ ! -f "$md_file" ]; then
        continue
    fi
    json_desc=$(jq -r --arg n "$name" '.commands[] | select(.name == $n) | .description' "$PLUGIN_JSON")
    # Extract the description line from the md frontmatter, stripping the
    # `description: ` prefix and surrounding double quotes if present.
    md_desc=$(awk '
        /^---[[:space:]]*$/ { in_fm = !in_fm; next }
        in_fm && /^description:/ {
            sub(/^description:[[:space:]]*/, "")
            sub(/^"/, ""); sub(/"$/, "")
            print
            exit
        }
    ' "$md_file")
    if [ "$json_desc" != "$md_desc" ]; then
        drift_failures+=("$name: plugin.json='$json_desc' vs md='$md_desc'")
    fi
done < <(jq -r '.commands[].name' "$PLUGIN_JSON")

if [ "${#drift_failures[@]}" -eq 0 ]; then
    pass "plugin.json descriptions match md frontmatter descriptions"
else
    fail "description drift detected:"
    for msg in "${drift_failures[@]}"; do
        echo "    - $msg"
    done
fi

# --- Check 10: PROCEDURE.md files have no skill-shaped frontmatter ----------
# Auto-discovery is keyed off YAML frontmatter shape (name + description).
# After Phase 0', PROCEDURE.md files must NOT carry frontmatter so that only
# router/SKILL.md is auto-discovered. Re-introducing frontmatter here would
# revive the routing collision and the illusory token-reduction regression.
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
