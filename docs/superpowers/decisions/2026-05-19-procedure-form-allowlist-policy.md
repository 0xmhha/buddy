# ADR-006 — B6 PROCEDURE form: bulk-allowlist intentional deviations, --strict for new skills

**Status**: Accepted (2026-05-19)
**Authors**: mhha (cli buddy track)
**Supersedes**: —
**Related**: ADR-004 §2.2 condition 3 (plugin v1.0.0 entry — PROCEDURE B6 unification)

## Context

`scripts/lint-skill-procedure.sh` (B6, shipped v0.2.0) checks every `plugin/skills/*/PROCEDURE.md` against three canonical forms:

- **Form A** — Stage skill, 8-section (목적 / 사용 시점 / 입력 / Stage 흐름 / 산출물 / 검증 / 다음 phase / 참조).
- **Form B** — Principles skill, 8-section (목적 / 사용 시점 / 입력 / 핵심 원칙 / 단계 / 출력 템플릿 / 자매 스킬 / Anti-patterns).
- **Form C** — Phase 1-3 era stage skill, 12-section (STOP / 목적 / 사용 시점 / 입력 / 핵심 원칙 / 단계 / 산출물 / Cross-phase cascade / 다음 skill / 경계 / 중요 규칙 / Verification gate).

A skill passes if any one form matches. 19 entries (9 orchestrators + 10 special skills + `.template`) were ALLOWLISTed at lint introduction.

148 skills scanned at the v0.7.2 baseline: 19 allowlisted, 86 passing one of the three forms, **43 deviating**. The lint ran in report-only mode (exit 0 even with deviations) per ADR-004 §2.2 condition 3, which marked B6 as a plugin v1.0.0 entry condition pending a follow-up decision.

This ADR is that decision.

## Investigation of the 43 deviations (2026-05-19)

Two distinct patterns surfaced when each deviating skill's actual `## ` headers were sampled:

### Pattern 1 — methodology / persona / external-asset skills (39 of 43)

These skills do not open with `## 1. 목적` / `## 2. 사용 시점` / `## 3. 입력` at all. They use prose-driven structure tailored to their content type:

- `consult-codex` — a state machine of `Step 0` / `Step 0.5` / `Step 1` / `Step 2A / 2B / 2C` describing Codex CLI mode-switching.
- `review-engineering` — Eng Manager persona block + `Review Dimensions` (Data Flow / Caching / Concurrency / Performance / Edge Cases / Test Coverage / Architecture Lock-In) + Conversational Recursive Pattern + Opinionated Recommendation Style.
- `conduct-postmortem` — Postmortem prose template (요약 / 영향 / 타임라인 / 근본 원인 / Contributing Factors / 해결 방법 / Action Items) embedded as `## ` headers.
- Same pattern for the rest: `summarize-retro`, `handle-incident`, `generate-improvement-tasks`, the 6 `review-*` reviewers, the 5 `map-*` / `compose-*` / `classify-*` mappers, etc.

Forcing these into Form A would erase the very structure that makes them useful at runtime — the persona blocks become "section bodies" rather than first-class shape.

### Pattern 2 — rich domain-design skills (4 of 43)

`design-billing-system`, `design-claude-hooks`, `review-ai-safety-liability`, `review-terms-policy-readiness` DO open with `## 1. 목적` / `## 2. 사용 시점` / `## 3. 입력`, but their §5 onwards is N-module domain decomposition (e.g., `design-billing-system` has 9 modules: Customer & Subscription / Pricing Catalog / Payment Method / Subscription Lifecycle / Usage Metering & Quota / Point Wallet / Invoice & Receipt / etc.).

Same reasoning: the body shape *is* the domain. A 9-module billing spec collapsed into "## 5. 산출물 형식" loses the per-module guidance.

## Options considered

### Option A — Bulk allowlist + --strict CI on new skills (chosen)

Add all 43 deviating skill names to `ALLOWLIST` in `scripts/lint-skill-procedure.sh`, with comment trail naming the two patterns. Then enable `--strict` mode in the CI workflow so any NEW skill that doesn't match Form A/B/C fails the gate.

**Pros**:
- Preserves intentional historical structure verbatim — no rewrites that destroy domain shape.
- Closes the v1.0.0 entry condition B-3: lint is *enforceable*, not report-only.
- ALLOWLIST entries carry a comment block explaining why they're allowlisted (Pattern 1 vs Pattern 2), so future contributors understand the rule.
- Adding a new allowlist entry requires editing this file with a justification, creating a natural review gate.

**Cons**:
- ALLOWLIST grows from 19 to 62 entries (+43). The list is long but the comment block makes it scannable.
- A bug-fix to Form A/B/C semantics will not retroactively force the allowlisted skills back into compliance. Acceptable — the allowlisted skills are *not* the ones we're trying to lint.

### Option B — Add Form D / Form E regex variants

Recognize Pattern 1 (methodology) as "Form D" and Pattern 2 (rich domain-design) as "Form E" via new regex sets, then keep --strict but with five accepted forms instead of three.

**Pros**:
- The lint actually *enforces* something on Pattern 1 / Pattern 2 skills (e.g., "must have a Persona section").

**Cons**:
- Pattern 1 has no shared structural skeleton — each methodology skill has its own headers (Codex's Step state machine ≠ Eng Manager's Review Dimensions). A single regex set cannot capture both without becoming trivially permissive (e.g., "any `## ` header counts").
- Pattern 2's 9-module body has no universal section names — they're all domain-specific.
- Maintenance burden: every new methodology / design skill requires regex tweaks.

### Option C — Status quo (keep report-only)

Defer the decision indefinitely. Lint runs but never blocks.

**Pros**:
- No code changes today.

**Cons**:
- Plugin v1.0.0 entry condition B-3 remains open indefinitely.
- New contributors get no automatic feedback when their skill doesn't match any canonical form.
- Drift accumulates silently.

## Decision

**Option A.** Bulk allowlist the 43 deviating skills with comment trail; enable `--strict` mode in `.github/workflows/ci.yml`. Any NEW skill must conform to Form A/B/C, OR an ADR amending this decision must explain why a Pattern 1 / Pattern 2 exception is justified.

## Consequences

- `scripts/lint-skill-procedure.sh` ALLOWLIST grows from 19 to 62 entries (one comment block per pattern explains the rationale).
- `.github/workflows/ci.yml` gains a `PROCEDURE.md form gate (--strict)` step that runs `bash scripts/lint-skill-procedure.sh --strict` on every push and PR.
- Plugin v1.0.0 entry condition B-3 (`PROCEDURE 양식 (B6) 통일 + --strict lint enforcement`) is **closed**. The remaining entry conditions are B-2 (production dogfood, user-paced) and B-4 (router smart-skip session state, separate ADR pending).
- Adding a NEW allowlist entry requires editing `scripts/lint-skill-procedure.sh` with a comment justifying which pattern the skill belongs to, plus referencing this ADR or its successor.

## Verification

```bash
$ make test-skill-form
=== Skill PROCEDURE Lint Report ===
Total skills:     148
Allowlist (skip): 62
Pass (8-section): 86
Deviate:          0
Missing file:     0

$ bash scripts/lint-skill-procedure.sh --strict ; echo "exit=$?"
=== Skill PROCEDURE Lint Report ===
Total skills:     148
Allowlist (skip): 62
Pass (8-section): 86
Deviate:          0
Missing file:     0
exit=0
```
