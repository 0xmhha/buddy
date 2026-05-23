package agent

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// parser.go implements the procedure-output parser: extract structured
// signals from a Claude subprocess's stdout after a buddy command runs.
//
// Two signals matter:
//   - §self-check verdict — the boolean checklist Claude marks with
//     `- [x]` once each item is verified. We compute pass / fail /
//     pending / unknown from the box state, plus per-item detail.
//   - §next-phase skill names — the cascade hint emitted as
//     backtick-wrapped skill identifiers. The Scheduler / `buddy agent
//     run` surface these so the user (or a future cascade runtime) can
//     pick what to invoke next.
//
// Important non-goal: the parser does NOT change retry / fail semantics.
// A failed self-check is surfaced as metadata in StepResult; the run's
// exit_code still comes from the executor. Tightening the loop
// (auto-retry on fail, abort-on-fail, etc.) is a follow-on change
// because the LLM may consistently echo `- [ ]` without genuinely
// completing self-check, and surprise auto-retry would burn token cost.

// SelfCheckVerdict is the parser's overall judgement on the §self-check
// boolean checklist.
type SelfCheckVerdict string

const (
	// SelfCheckPass — every checkbox in the section is `- [x]`.
	SelfCheckPass SelfCheckVerdict = "pass"
	// SelfCheckFail — at least one `- [x]` AND at least one `- [ ]` — the
	// LLM started checking but did not finish.
	SelfCheckFail SelfCheckVerdict = "fail"
	// SelfCheckPending — every checkbox is `- [ ]`. Usually means the
	// procedure was not actually executed (e.g. template echo) or the
	// LLM forgot to mark items. Distinct from Fail so callers can react
	// differently (Pending is often retry-worthy; Fail is not).
	SelfCheckPending SelfCheckVerdict = "pending"
	// SelfCheckUnknown — no §self-check section found in the output.
	// This is the default for orchestrator skills, special skills, and
	// skills that omit the section by design.
	SelfCheckUnknown SelfCheckVerdict = "unknown"
)

// SelfCheckItem is one checklist row.
type SelfCheckItem struct {
	Passed bool   `json:"passed"`
	Text   string `json:"text"`
}

// SelfCheck is the parser's summary plus the raw per-item list.
type SelfCheck struct {
	Verdict SelfCheckVerdict `json:"verdict"`
	Passed  int              `json:"passed"`
	Failed  int              `json:"failed"`
	Total   int              `json:"total"`
	Items   []SelfCheckItem  `json:"items,omitempty"`
}

// NextPhase is the cascade hint extracted from §next-phase. Skills is
// every backtick-wrapped identifier the parser found (the union across
// any conditional branches plus any sequential candidates). Branches
// captures conditional cascade rules ("- 글로벌 → `skill-a`" lines) so
// callers that want to follow the condition rather than the union can
// pick the right target. Raw is the section body verbatim.
type NextPhase struct {
	Skills   []string          `json:"skills,omitempty"`
	Branches []NextPhaseBranch `json:"branches,omitempty"`
	Raw      string            `json:"raw,omitempty"`
}

// NextPhaseBranch is one conditional cascade rule. Condition is the
// trimmed LHS prose (e.g. "글로벌", "Korea", "USA / EU / 기타"); Skills is
// the backtick-extracted RHS (may be empty when the PROCEDURE marks a
// branch as "to be created" without naming an existing skill).
type NextPhaseBranch struct {
	Condition string   `json:"condition"`
	Skills    []string `json:"skills,omitempty"`
}

// ParsedOutput aggregates everything the parser pulls out of one Claude
// stdout buffer. Embedded into StepResult.
type ParsedOutput struct {
	SelfCheck SelfCheck `json:"self_check"`
	NextPhase NextPhase `json:"next_phase"`
}

// section header regex per Form. Matches things like
//   ## 6. 검증 (self-check)
//   ## 6. 검증
//   ## 6. Verification
//   ## 11. Verification gate — 완료 선언 전 self-check
//
// The numbering is intentionally optional ("## 검증" alone also works)
// so unnumbered orchestrator-flavoured sections register.
var (
	selfCheckHeaderRE = regexp.MustCompile(`(?m)^##\s+(?:\d+\.\s+)?(?:검증|Verification|Self-check)`)
	nextPhaseHeaderRE = regexp.MustCompile(`(?m)^##\s+(?:\d+\.\s+)?(?:다음\s+(?:phase|skill)|Next\s+(?:phase|skill))`)
	anySectionHeader  = regexp.MustCompile(`(?m)^##\s+`)
	checkboxRE        = regexp.MustCompile(`^\s*-\s*\[([ xX])\]\s*(.+)$`)
	// backtickRefRE matches `name` and is the simplest reliable skill-
	// identifier extractor. We post-filter to drop trivially non-skill
	// values (those containing whitespace or starting with /).
	backtickRefRE = regexp.MustCompile("`([^`\n]+)`")
	// branchLineRE matches a bullet of the form
	//   - <condition> → <rhs>
	//   - <condition> -> <rhs>
	// where <condition> is short prose (≤80 chars, no opening backtick).
	// The backtick-leading check rejects ordinary sequential bullets like
	// `- ` + "`skill-name` — description with → arrow inside" where the
	// arrow belongs to a description rather than a conditional split.
	branchLineRE = regexp.MustCompile(`^\s*-\s+([^` + "`" + `\n][^\n]*?)\s*(?:→|->)\s*(.+)$`)
)

// ParseClaudeOutput scans raw stdout from a Claude subprocess run of one
// buddy command and returns the structured signals. Empty / unknown
// values are valid — the caller checks Verdict / Skills before using
// them.
//
// The function is forgiving: malformed checkboxes are skipped, missing
// sections produce SelfCheckUnknown / empty NextPhase, and the raw body
// is preserved for callers that want to inspect manually.
func ParseClaudeOutput(stdout string) ParsedOutput {
	return ParsedOutput{
		SelfCheck: parseSelfCheck(stdout),
		NextPhase: parseNextPhase(stdout),
	}
}

func parseSelfCheck(s string) SelfCheck {
	body := extractSection(s, selfCheckHeaderRE)
	if body == "" {
		return SelfCheck{Verdict: SelfCheckUnknown}
	}
	var items []SelfCheckItem
	for _, line := range strings.Split(body, "\n") {
		m := checkboxRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		passed := m[1] == "x" || m[1] == "X"
		items = append(items, SelfCheckItem{Passed: passed, Text: strings.TrimSpace(m[2])})
	}
	sc := SelfCheck{Items: items, Total: len(items)}
	for _, it := range items {
		if it.Passed {
			sc.Passed++
		} else {
			sc.Failed++
		}
	}
	switch {
	case sc.Total == 0:
		sc.Verdict = SelfCheckUnknown
	case sc.Passed == sc.Total:
		sc.Verdict = SelfCheckPass
	case sc.Passed == 0:
		sc.Verdict = SelfCheckPending
	default:
		sc.Verdict = SelfCheckFail
	}
	return sc
}

func parseNextPhase(s string) NextPhase {
	body := extractSection(s, nextPhaseHeaderRE)
	if body == "" {
		return NextPhase{}
	}
	var skills []string
	seen := map[string]bool{}
	for _, m := range backtickRefRE.FindAllStringSubmatch(body, -1) {
		candidate := strings.TrimSpace(m[1])
		if !looksLikeSkillIdent(candidate) || seen[candidate] {
			continue
		}
		seen[candidate] = true
		skills = append(skills, candidate)
	}
	return NextPhase{Skills: skills, Branches: parseBranches(body), Raw: strings.TrimSpace(body)}
}

// parseBranches scans the §next-phase body for conditional cascade
// bullets of the form
//
//	- <condition> → `skill-a`
//	- <condition> → `skill-a` + `skill-b` (optional description)
//	- <condition> → (skill not yet authored — prose only)
//
// Sequential candidate bullets (those that start with a backtick or
// contain no arrow) are intentionally ignored — they belong in Skills,
// not Branches.
func parseBranches(body string) []NextPhaseBranch {
	var out []NextPhaseBranch
	for _, line := range strings.Split(body, "\n") {
		m := branchLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		condition := strings.TrimSpace(m[1])
		rhs := m[2]
		// Skip bullets whose LHS is suspiciously long — those are almost
		// certainly prose descriptions that happen to contain an arrow,
		// not condition labels. Measured in *runes*, not bytes, so the
		// limit doesn't shrink for Hangul/Japanese/Chinese conditions.
		// 30 runes covers "USA / EU / 기타" (12) comfortably without
		// admitting paragraph-length prose.
		if condition == "" || utf8.RuneCountInString(condition) > 30 {
			continue
		}
		var skills []string
		seen := map[string]bool{}
		for _, bm := range backtickRefRE.FindAllStringSubmatch(rhs, -1) {
			cand := strings.TrimSpace(bm[1])
			if !looksLikeSkillIdent(cand) || seen[cand] {
				continue
			}
			seen[cand] = true
			skills = append(skills, cand)
		}
		out = append(out, NextPhaseBranch{Condition: condition, Skills: skills})
	}
	return out
}

// looksLikeSkillIdent filters backtick refs to *kebab-case identifiers*
// that match buddy's skill naming convention. Drops obviously-non-skill
// strings (paths, sentences with spaces, code-style identifiers with /).
func looksLikeSkillIdent(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

// extractSection returns the body of the section whose header matches
// headerRE, ending at the next `## ` line (or EOF). Empty string when
// the section is absent.
func extractSection(s string, headerRE *regexp.Regexp) string {
	loc := headerRE.FindStringIndex(s)
	if loc == nil {
		return ""
	}
	// Body starts on the line AFTER the header line.
	rest := s[loc[1]:]
	if nl := strings.Index(rest, "\n"); nl >= 0 {
		rest = rest[nl+1:]
	}
	// End at the next `## ` section header.
	if next := anySectionHeader.FindStringIndex(rest); next != nil {
		return rest[:next[0]]
	}
	return rest
}
