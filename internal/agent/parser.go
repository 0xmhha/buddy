package agent

import (
	"regexp"
	"strings"
)

// parser.go implements W3-4 of the cli-buddy-spec: extract structured
// signals from a Claude subprocess's stdout after a buddy command runs.
//
// Two signals matter for v0.3:
//   - §self-check verdict — the boolean checklist that PROCEDURE.md §6 /
//     §11 (Form A / C) asks Claude to mark with `- [x]` once each item is
//     verified. We compute pass / fail / pending / unknown from the box
//     state, plus per-item detail.
//   - §next-phase skill names — the cascade hint PROCEDURE.md §7 / §8
//     emits as backtick-wrapped skill identifiers. The Scheduler /
//     `buddy agent run` surface these so the user (or a future cascade
//     runtime) can pick what to invoke next.
//
// Important non-goal: the parser does NOT change retry / fail semantics
// in v0.3. A failed self-check is surfaced as metadata in StepResult;
// the run's exit_code still comes from the executor. Tightening the
// loop (auto-retry on fail, abort-on-fail, etc.) is a follow-on cycle
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
// every backtick-wrapped identifier the parser found; Raw is the
// section body verbatim so callers can surface conditional branches
// ("if A → X / if B → Y") that the parser does not interpret.
type NextPhase struct {
	Skills []string `json:"skills,omitempty"`
	Raw    string   `json:"raw,omitempty"`
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
	return NextPhase{Skills: skills, Raw: strings.TrimSpace(body)}
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
