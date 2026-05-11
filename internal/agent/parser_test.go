package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ─── self-check ──────────────────────────────────────────────────────────

func TestParseSelfCheck_AllPassedYieldsPassVerdict(t *testing.T) {
	t.Parallel()
	out := `
Some preamble.

## 6. 검증 (self-check)

- [x] 첫 번째 항목 완료
- [x] 두 번째 항목 완료
- [x] 세 번째 항목 완료

## 7. 다음 phase
something
`
	got := ParseClaudeOutput(out).SelfCheck
	require.Equal(t, SelfCheckPass, got.Verdict)
	require.Equal(t, 3, got.Total)
	require.Equal(t, 3, got.Passed)
	require.Equal(t, 0, got.Failed)
}

func TestParseSelfCheck_MixedYieldsFailVerdict(t *testing.T) {
	t.Parallel()
	out := `
## 6. 검증

- [x] item A 통과
- [ ] item B 미완료
- [x] item C 통과
`
	got := ParseClaudeOutput(out).SelfCheck
	require.Equal(t, SelfCheckFail, got.Verdict)
	require.Equal(t, 3, got.Total)
	require.Equal(t, 2, got.Passed)
	require.Equal(t, 1, got.Failed)
	require.Equal(t, "item B 미완료", got.Items[1].Text)
}

func TestParseSelfCheck_AllUncheckedYieldsPending(t *testing.T) {
	t.Parallel()
	out := `
## 6. 검증 (self-check)

- [ ] never started
- [ ] also never started
`
	got := ParseClaudeOutput(out).SelfCheck
	require.Equal(t, SelfCheckPending, got.Verdict)
	require.Equal(t, 2, got.Total)
	require.Equal(t, 0, got.Passed)
}

func TestParseSelfCheck_FormCVerificationGate(t *testing.T) {
	t.Parallel()
	out := `
## 11. Verification gate — 완료 선언 전 self-check

다음 체크가 모두 yes 여야 절차 완료 보고:

- [x] §5 phase 모두 실행됨
- [x] §6 출력 섹션 모두 채워짐
- [x] Dimensions Evaluated 표가 8 row
- [ ] Risk Register 가 5 row
`
	got := ParseClaudeOutput(out).SelfCheck
	require.Equal(t, SelfCheckFail, got.Verdict)
	require.Equal(t, 4, got.Total)
	require.Equal(t, 3, got.Passed)
}

func TestParseSelfCheck_NoSectionYieldsUnknown(t *testing.T) {
	t.Parallel()
	out := `
Just some prose without any self-check section. Lorem ipsum.
- [x] this checkbox is outside any §검증 section
`
	got := ParseClaudeOutput(out).SelfCheck
	require.Equal(t, SelfCheckUnknown, got.Verdict)
	require.Equal(t, 0, got.Total)
	require.Empty(t, got.Items)
}

func TestParseSelfCheck_CaseInsensitiveCapitalXIsPassed(t *testing.T) {
	t.Parallel()
	out := `
## 6. Verification

- [X] capital X
- [x] lowercase x
`
	got := ParseClaudeOutput(out).SelfCheck
	require.Equal(t, SelfCheckPass, got.Verdict)
	require.Equal(t, 2, got.Passed)
}

// ─── next-phase ──────────────────────────────────────────────────────────

func TestParseNextPhase_ExtractsSkillIdents(t *testing.T) {
	t.Parallel()
	out := `
## 7. 다음 phase

본 skill 산출이 *region cluster trigger* 결과에 따라 다음 분기:

- 글로벌 → ` + "`review-legal-regulatory`" + ` (region-agnostic frame)
- Korea → ` + "`consult-korea-legal-context`" + ` + ` + "`review-legal-regulatory`" + `
- USA → 별 cluster
`
	got := ParseClaudeOutput(out).NextPhase
	require.Equal(t, []string{"review-legal-regulatory", "consult-korea-legal-context"}, got.Skills)
	require.Contains(t, got.Raw, "region-agnostic frame")
}

func TestParseNextPhase_FiltersNonSkillBackticks(t *testing.T) {
	t.Parallel()
	out := `
## 8. 다음 skill (next in stage flow)

- ` + "`plan-build`" + ` is the orchestrator
- ` + "`docs/something.md`" + ` should be filtered (has slash)
- ` + "`a sentence with spaces`" + ` should be filtered (has space)
- ` + "`UPPERCASE`" + ` should be filtered (capitals)
`
	got := ParseClaudeOutput(out).NextPhase
	require.Equal(t, []string{"plan-build"}, got.Skills)
}

func TestParseNextPhase_AbsentSectionYieldsEmpty(t *testing.T) {
	t.Parallel()
	out := `
## 6. 검증

- [x] something passed
`
	got := ParseClaudeOutput(out).NextPhase
	require.Empty(t, got.Skills)
	require.Empty(t, got.Raw)
}

func TestParseNextPhase_DeduplicatesRepeatedRefs(t *testing.T) {
	t.Parallel()
	out := `
## 다음 phase

- ` + "`design-system`" + ` is the entry
- See ` + "`design-system`" + ` repeated
- And ` + "`design-system`" + ` again
`
	got := ParseClaudeOutput(out).NextPhase
	require.Equal(t, []string{"design-system"}, got.Skills)
}

// ─── integration ─────────────────────────────────────────────────────────

func TestParseClaudeOutput_BothSectionsPopulated(t *testing.T) {
	t.Parallel()
	out := `
# Decide Target Market — 한국어 SaaS

## 5. 산출물 형식

(...산출물 내용...)

## 6. 검증 (self-check)

- [x] Stage 1 matrix 가 3 시장 평가됐는가?
- [x] Stage 2 결정이 3 옵션 중 하나로 명시됐는가?
- [x] ADR 작성됐는가?

## 7. 다음 phase

- 글로벌 → ` + "`review-legal-regulatory`" + `
- Korea → ` + "`consult-korea-legal-context`" + `
`
	parsed := ParseClaudeOutput(out)
	require.Equal(t, SelfCheckPass, parsed.SelfCheck.Verdict)
	require.Equal(t, 3, parsed.SelfCheck.Total)
	require.Len(t, parsed.NextPhase.Skills, 2)
	require.Equal(t, "review-legal-regulatory", parsed.NextPhase.Skills[0])
	require.Equal(t, "consult-korea-legal-context", parsed.NextPhase.Skills[1])
}

func TestParseClaudeOutput_EmptyStringIsHarmless(t *testing.T) {
	t.Parallel()
	parsed := ParseClaudeOutput("")
	require.Equal(t, SelfCheckUnknown, parsed.SelfCheck.Verdict)
	require.Empty(t, parsed.NextPhase.Skills)
	require.Empty(t, parsed.NextPhase.Branches)
}

// ─── conditional branches ────────────────────────────────────────────────

func TestParseBranches_SingleConditional(t *testing.T) {
	t.Parallel()
	out := `
## 7. 다음 phase

- 글로벌 → ` + "`review-legal-regulatory`" + ` (region-agnostic frame)
`
	got := ParseClaudeOutput(out).NextPhase
	require.Len(t, got.Branches, 1)
	require.Equal(t, "글로벌", got.Branches[0].Condition)
	require.Equal(t, []string{"review-legal-regulatory"}, got.Branches[0].Skills)
}

func TestParseBranches_MultiConditionalWithDescription(t *testing.T) {
	t.Parallel()
	out := `
## 7. 다음 phase

본 skill 산출이 *region cluster trigger* 결과에 따라 다음 분기:

- 글로벌 → ` + "`review-legal-regulatory`" + ` (region-agnostic frame)
- Korea → ` + "`consult-korea-legal-context`" + ` + ` + "`review-legal-regulatory`" + `
- USA / EU / 기타 → (template 작성 필요 — trigger 발생 시 즉시 신규 skill)
`
	got := ParseClaudeOutput(out).NextPhase
	require.Len(t, got.Branches, 3)
	require.Equal(t, "글로벌", got.Branches[0].Condition)
	require.Equal(t, []string{"review-legal-regulatory"}, got.Branches[0].Skills)

	require.Equal(t, "Korea", got.Branches[1].Condition)
	require.Equal(t,
		[]string{"consult-korea-legal-context", "review-legal-regulatory"},
		got.Branches[1].Skills)

	require.Equal(t, "USA / EU / 기타", got.Branches[2].Condition)
	require.Empty(t, got.Branches[2].Skills, "no-skill branch keeps condition but empty Skills")
}

func TestParseBranches_AsciiArrowAlsoMatches(t *testing.T) {
	t.Parallel()
	out := `
## 다음 phase

- staging -> ` + "`run-uat`" + `
- production -> ` + "`setup-canary-deploy`" + `
`
	got := ParseClaudeOutput(out).NextPhase
	require.Len(t, got.Branches, 2)
	require.Equal(t, "staging", got.Branches[0].Condition)
	require.Equal(t, []string{"run-uat"}, got.Branches[0].Skills)
	require.Equal(t, "production", got.Branches[1].Condition)
	require.Equal(t, []string{"setup-canary-deploy"}, got.Branches[1].Skills)
}

func TestParseBranches_SequentialOnlyHasEmptyBranches(t *testing.T) {
	t.Parallel()
	out := `
## 8. 다음 skill (next in stage flow)

- ` + "`write-adr`" + ` — 본 skill 의 결정 영속화
- ` + "`design-data-model`" + ` — entity → API resource 매핑
- ` + "`design-api-contract`" + ` — schema-first design
`
	got := ParseClaudeOutput(out).NextPhase
	require.Equal(t,
		[]string{"write-adr", "design-data-model", "design-api-contract"},
		got.Skills,
		"sequential bullets still feed the union Skills list")
	require.Empty(t, got.Branches,
		"backtick-leading bullets must not register as conditional branches even when their description contains →")
}

func TestParseBranches_RejectsLongConditionLine(t *testing.T) {
	t.Parallel()
	// A description-style bullet that happens to contain an arrow mid-sentence,
	// without a leading backtick, must not be misclassified as a branch.
	out := `
## 다음 phase

- 본 skill 산출이 매우 길고 자세한 prose 설명이며 그 안에 entity → API resource 매핑 의미의 화살표가 포함되어 있는 줄은 분기 정보가 아니라 권장 흐름 설명이다 → ` + "`design-system`" + `
`
	got := ParseClaudeOutput(out).NextPhase
	require.Empty(t, got.Branches, "long LHS prose (>80 chars) is not a condition label")
	require.Equal(t, []string{"design-system"}, got.Skills, "skill still feeds the union list via backtickRefRE")
}

func TestParseBranches_RejectsBacktickLeadingLine(t *testing.T) {
	t.Parallel()
	// Bullet whose LHS literally starts with a backtick — the arrow is inside
	// the description, not a condition split.
	out := `
## 다음 phase

- ` + "`design-data-model`" + ` — entity → API resource 매핑
`
	got := ParseClaudeOutput(out).NextPhase
	require.Empty(t, got.Branches)
	require.Equal(t, []string{"design-data-model"}, got.Skills)
}

func TestParseClaudeOutput_BranchesAlongsideSkills(t *testing.T) {
	t.Parallel()
	out := `
## 6. 검증

- [x] 결정됐는가

## 7. 다음 phase

- 글로벌 → ` + "`review-legal-regulatory`" + `
- Korea → ` + "`consult-korea-legal-context`" + `
`
	parsed := ParseClaudeOutput(out)
	require.Equal(t, SelfCheckPass, parsed.SelfCheck.Verdict)
	require.Len(t, parsed.NextPhase.Skills, 2, "union Skills covers all branches")
	require.Len(t, parsed.NextPhase.Branches, 2, "Branches preserves the conditional structure")
	require.Equal(t, "글로벌", parsed.NextPhase.Branches[0].Condition)
	require.Equal(t, "Korea", parsed.NextPhase.Branches[1].Condition)
}
