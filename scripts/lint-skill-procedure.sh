#!/usr/bin/env bash
# lint-skill-procedure.sh — B6 PROCEDURE 양식 inconsistency lint
#
# Stage skill 의 8-section canonical form 강제:
#   1. 목적 / 2. 사용 시점 / 3. 입력 / 4. Stage 흐름 / 5. 산출물 형식 / 6. 검증 / 7. 다음 phase / 8. 참조
#
# Orchestrator 9 (concretize-idea, define-features, design-system, plan-build,
#   build-feature, verify-quality, ship-release, iterate-product, manage-lifecycle)
#   + Special skills (router, status, validate-idea, autoplan, restore-context,
#     save-context, dispatch-parallel-agents) 는 의도된 별도 양식 → allowlist.
#
# Usage:
#   scripts/lint-skill-procedure.sh                    # report only (exit 0 always)
#   scripts/lint-skill-procedure.sh --strict           # exit 1 if any deviation (CI)
#   scripts/lint-skill-procedure.sh --verbose          # print per-skill section presence
#
# ADR: docs/superpowers/decisions/2026-05-11-plugin-version-reset.md §2.2 condition 3

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SKILLS_DIR="${REPO_ROOT}/plugin/skills"

STRICT=0
VERBOSE=0
for arg in "$@"; do
  case "$arg" in
    --strict) STRICT=1 ;;
    --verbose) VERBOSE=1 ;;
    *) echo "unknown flag: $arg" >&2; exit 2 ;;
  esac
done

# Allowlist resolved as part of B-3 (2026-05-19) per ADR-006 — the 43
# deviations that surfaced from the original Form A/B/C scan turned out
# to be intentional historical choices, not authoring drift:
#
#   - Pattern 1 (methodology / persona / external-asset skills) lack the
#     `## 1. 목적` shape entirely. Forcing them into the 8-section template
#     would erase the very structure that makes them useful (Eng Manager
#     persona blocks, conduct-postmortem prose template, Codex mode-switch
#     state machine, etc.).
#
#   - Pattern 2 (rich domain-design skills like design-billing-system)
#     DO open with §1./§2./§3. but use 9-module / N-module bodies after
#     §4. Same reasoning: the body shape *is* the domain.
#
# Adding them by name (rather than relaxing the regex with extra Form D/E
# variants) keeps the lint *enforceable* via --strict for every NEW skill
# while preserving the intentional set verbatim.
#
# Adding a new entry to ALLOWLIST should require an ADR note explaining
# why the skill cannot conform — otherwise just align it with Form A.
ALLOWLIST=(
  router
  .template
  # Orchestrators (9-phase)
  concretize-idea
  define-features
  design-system
  plan-build
  build-feature
  verify-quality
  ship-release
  iterate-product
  manage-lifecycle
  # Special skills (forcing-question / context / dispatch / status)
  status
  validate-idea
  validate-advanced-edge-idea
  autoplan
  restore-context
  save-context
  dispatch-parallel-agents
  detect-install-type
  guide-setup-wizard
  # B-3 (2026-05-19, ADR-006) — methodology / persona skills (Pattern 1)
  analyze-ab-experiment
  analyze-user-funnel
  apply-builder-ethos
  audit-live-devex
  audit-security
  benchmark-llm-models
  classify-qa-tiers
  classify-review-risks
  compose-feature-from-use-cases
  compose-safety-mode
  conduct-postmortem
  consult-codex
  consult-design-system
  critique-plan
  decompose-blocker
  define-feature-spec
  design-ab-experiment
  freeze-edit-scope
  generate-improvement-tasks
  guard-destructive-commands
  handle-incident
  identify-actors
  iterate-fix-verify
  map-actor-use-cases
  map-feature-dependencies
  map-use-case-to-system-boundary
  measure-code-health
  monitor-regressions
  persist-learning-jsonl
  review-architecture
  review-design
  review-devex
  review-engineering
  review-scope
  run-browser-qa
  score-feature-priority
  summarize-retro
  sync-release-docs
  verify-best-alternative
  write-changelog
  # B-3 (2026-05-19, ADR-006) — rich domain-design skills (Pattern 2)
  design-billing-system
  design-claude-hooks
  review-ai-safety-liability
  review-terms-policy-readiness
)

is_allowed() {
  local name="$1"
  for a in "${ALLOWLIST[@]}"; do
    [ "$name" = "$a" ] && return 0
  done
  return 1
}

# 세 canonical form 인정 (역사적 batch 별 진화):
#
# Form A — Stage skill (Skill Completion Cycle Batch 1~7, 8-section):
#   1.목적 / 2.사용 시점 / 3.입력 / 4.Stage 흐름 / 5.산출물 형식 /
#   6.검증 / 7.다음 phase / 8.참조
#
# Form B — Principles skill (methodology + anti-pattern, 8-section):
#   1.목적 / 2.사용 시점 / 3.입력 / 4.핵심 원칙 / 5.단계 /
#   6.출력 템플릿 / 7.자매 스킬 / 8.Anti-patterns
#
# Form C — Phase 1-3 stage skill (Phase 1~5 ext, 12-section richer form):
#   0.STOP / 1.목적 / 2.사용 시점 / 3.입력 / 4.핵심 원칙 / 5.단계(Phases) /
#   6.산출물 형식 / 7.Cross-phase cascade / 8.다음 skill / 9.경계 /
#   10.중요 규칙 / 11.Verification gate
#
# 세 form 모두 §1~§3 공통. §4~ 가 분기. lint 는 *하나의 form* 이라도 매치되면 PASS.
# (향후 B6 follow-up ADR 에서 단일 form 으로 통일 결정 시 lint 도 좁힘.)

declare -A SEC_COMMON=(
  [1]='^## 1\. 목적'
  [2]='^## 2\. 사용 시점'
  [3]='^## 3\. 입력'
)

declare -A SEC_FORM_A=(
  [4]='^## 4\. (Stage|스테이지)'
  [5]='^## 5\. (산출물|Output)'
  [6]='^## 6\. (검증|Verification|Self-check)'
  [7]='^## 7\. 다음 phase'
  [8]='^## 8\. (참조|References)'
)

declare -A SEC_FORM_B=(
  [4]='^## 4\. 핵심 원칙'
  [5]='^## 5\. (단계|Phases)'
  [6]='^## 6\. (출력 템플릿|Output)'
  [7]='^## 7\. (자매 스킬|Sibling)'
  [8]='^## 8\. Anti-patterns'
)

# Form C — 12-section rich form (Phase 1-3 / Phase 5 ext era).
# §4-§6 overlaps Form B (핵심 원칙 / 단계 / 산출물 형식), §7-§11 unique (Cross-phase / 다음 skill / 경계 / 중요 규칙 / Verification gate).
declare -A SEC_FORM_C=(
  [4]='^## 4\. 핵심 원칙'
  [5]='^## 5\. (단계|Phases)'
  [6]='^## 6\. (산출물|Output)'
  [7]='^## 7\. (Cross-phase|크로스-페이즈)'
  [8]='^## 8\. 다음 skill'
  [9]='^## 9\. '
  [10]='^## 10\. '
  [11]='^## 11\. (Verification|검증)'
)

total=0
pass=0
fail=0
allowed=0
missing_procedure=0
deviations=()

for d in "${SKILLS_DIR}"/*/; do
  name="$(basename "$d")"
  [ "$name" = ".template" ] && continue
  total=$((total+1))

  f="${d}PROCEDURE.md"
  if [ ! -f "$f" ]; then
    # router uses SKILL.md, not PROCEDURE.md
    if [ "$name" = "router" ]; then
      allowed=$((allowed+1))
      continue
    fi
    missing_procedure=$((missing_procedure+1))
    deviations+=("MISSING_PROCEDURE: ${name}")
    continue
  fi

  if is_allowed "$name"; then
    allowed=$((allowed+1))
    [ "$VERBOSE" = 1 ] && echo "ALLOWED: ${name} (orchestrator / special)"
    continue
  fi

  # §1~§3 공통 확인
  common_missing=()
  for i in 1 2 3; do
    grep -qE "${SEC_COMMON[$i]}" "$f" || common_missing+=("$i")
  done

  # Form A / B / C — 각 form 별 §4 이후 매치
  form_a_missing=()
  form_b_missing=()
  form_c_missing=()
  for i in 4 5 6 7 8; do
    grep -qE "${SEC_FORM_A[$i]}" "$f" || form_a_missing+=("$i")
    grep -qE "${SEC_FORM_B[$i]}" "$f" || form_b_missing+=("$i")
  done
  for i in 4 5 6 7 8 9 10 11; do
    grep -qE "${SEC_FORM_C[$i]}" "$f" || form_c_missing+=("$i")
  done

  if [ "${#common_missing[@]}" -eq 0 ] && { [ "${#form_a_missing[@]}" -eq 0 ] || [ "${#form_b_missing[@]}" -eq 0 ] || [ "${#form_c_missing[@]}" -eq 0 ]; }; then
    pass=$((pass+1))
    if [ "$VERBOSE" = 1 ]; then
      [ "${#form_a_missing[@]}" -eq 0 ] && echo "PASS: ${name} (Form A)"
      [ "${#form_a_missing[@]}" -ne 0 ] && [ "${#form_b_missing[@]}" -eq 0 ] && echo "PASS: ${name} (Form B)"
      [ "${#form_a_missing[@]}" -ne 0 ] && [ "${#form_b_missing[@]}" -ne 0 ] && [ "${#form_c_missing[@]}" -eq 0 ] && echo "PASS: ${name} (Form C)"
    fi
  else
    fail=$((fail+1))
    detail=""
    [ "${#common_missing[@]}" -gt 0 ] && detail+="common §$(IFS=,; echo "${common_missing[*]}") "
    detail+="(A miss §$(IFS=,; echo "${form_a_missing[*]}") / B miss §$(IFS=,; echo "${form_b_missing[*]}") / C miss §$(IFS=,; echo "${form_c_missing[*]}"))"
    deviations+=("DEVIATE: ${name} — ${detail}")
  fi
done

echo "=== Skill PROCEDURE Lint Report ==="
echo "Total skills:     ${total}"
echo "Allowlist (skip): ${allowed}"
echo "Pass (8-section): ${pass}"
echo "Deviate:          ${fail}"
echo "Missing file:     ${missing_procedure}"
echo ""

if [ "${#deviations[@]}" -gt 0 ]; then
  echo "=== Deviations ==="
  for d in "${deviations[@]}"; do
    echo "  $d"
  done
  echo ""
  echo "Action: align skill PROCEDURE to template at plugin/skills/.template/PROCEDURE.md"
  echo "Or add to ALLOWLIST in scripts/lint-skill-procedure.sh if intentional"
fi

if [ "$STRICT" = 1 ] && [ "${#deviations[@]}" -gt 0 ]; then
  exit 1
fi
exit 0
