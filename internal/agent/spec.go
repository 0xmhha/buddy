package agent

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseSpec unmarshals YAML into an AgentSpec and runs structural validation.
// Round-trip preservation: the original YAML stays in Agent.SpecYAML — this
// function never re-serialises it.
func ParseSpec(yamlBytes []byte) (AgentSpec, error) {
	var spec AgentSpec
	if err := yaml.Unmarshal(yamlBytes, &spec); err != nil {
		return AgentSpec{}, fmt.Errorf("agent: parse yaml: %w", err)
	}
	if err := spec.Validate(); err != nil {
		return AgentSpec{}, err
	}
	return spec, nil
}

// idPattern enforces the kebab-case / lowercase identifier shape used across
// buddy (matches the skill / command naming convention).
var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

// Validate checks the structural invariants of a parsed spec. Returns the
// *first* problem found; callers don't need a list because lint passes fix
// one thing at a time.
func (s AgentSpec) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return errors.New("agent: spec.id is required")
	}
	if !idPattern.MatchString(s.ID) {
		return fmt.Errorf("agent: spec.id %q must match %s", s.ID, idPattern)
	}
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("agent: spec.name is required")
	}
	if len(s.Chain) == 0 {
		return errors.New("agent: spec.chain must contain at least one step")
	}
	for i, step := range s.Chain {
		if strings.TrimSpace(step.Command) == "" {
			return fmt.Errorf("agent: spec.chain[%d].command is required", i)
		}
		// Buddy convention: command names match skill directory names
		// (kebab-case, no /buddy: prefix). We accept a leading slash for
		// usability but strip it during Run.
	}
	if s.Retry != nil && s.Retry.MaxAttempts < 1 {
		return errors.New("agent: spec.retry.max_attempts must be >= 1 when retry is set")
	}
	if s.Output != nil {
		switch s.Output.Type {
		case "stdout", "":
			// stdout is the default; empty is treated as stdout.
		case "file":
			if strings.TrimSpace(s.Output.Path) == "" {
				return errors.New("agent: spec.output.path is required when output.type='file'")
			}
		default:
			return fmt.Errorf("agent: spec.output.type %q is unsupported (want stdout|file)", s.Output.Type)
		}
	}
	return nil
}

// NormalizeCommand strips a leading "/buddy:" or "/" prefix so a user pasting
// the slash form from their CLI works without surprise.
func NormalizeCommand(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	cmd = strings.TrimPrefix(cmd, "/buddy:")
	cmd = strings.TrimPrefix(cmd, "/")
	return cmd
}
