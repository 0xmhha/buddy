// Package schema defines the canonical hook event payload.
// 1:1 port of the v0.1 spec §6.1 (option A).
package schema

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// HookEventName enumerates the Claude Code hook events buddy understands.
type HookEventName string

const (
	EventSessionStart     HookEventName = "SessionStart"
	EventPreToolUse       HookEventName = "PreToolUse"
	EventPostToolUse      HookEventName = "PostToolUse"
	EventStop             HookEventName = "Stop"
	EventPreCompact       HookEventName = "PreCompact"
	EventUserPromptSubmit HookEventName = "UserPromptSubmit"
	EventNotification     HookEventName = "Notification"
	EventSubagentStop     HookEventName = "SubagentStop"
	EventSessionEnd       HookEventName = "SessionEnd"
)

// KnownEvents is the canonical, ordered list of hook events.
// install.hookEvents derives from this list, so adding a new event
// here automatically widens install's wrap coverage.
var KnownEvents = []HookEventName{
	EventSessionStart, EventPreToolUse, EventPostToolUse,
	EventStop, EventPreCompact, EventUserPromptSubmit,
	EventNotification, EventSubagentStop, EventSessionEnd,
}

// IsKnown reports whether n is one of the canonical hook event names.
func (n HookEventName) IsKnown() bool {
	return slices.Contains(KnownEvents, n)
}

// TokenUsage mirrors Claude Code's per-message usage object.
type TokenUsage struct {
	InputTokens       int `json:"inputTokens"`
	OutputTokens      int `json:"outputTokens"`
	CacheReadTokens   int `json:"cacheReadTokens"`
	CacheCreateTokens int `json:"cacheCreateTokens"`
}

// HookEventPayload is what buddy writes to the outbox per hook execution.
// Field-by-field 1:1 with the TS Zod schema in archive/ts-poc.
type HookEventPayload struct {
	Ts         int64             `json:"ts"`
	Event      HookEventName     `json:"event"`
	HookName   string            `json:"hookName"`
	DurationMs int64             `json:"durationMs"`
	ExitCode   int               `json:"exitCode"`

	SessionID  string            `json:"sessionId,omitempty"`
	PID        int               `json:"pid,omitempty"`
	Cwd        string            `json:"cwd,omitempty"`

	ToolName   string            `json:"toolName,omitempty"`
	ToolArgs   any               `json:"toolArgs,omitempty"`
	ModelName  string            `json:"modelName,omitempty"`
	TokenUsage *TokenUsage       `json:"tokenUsage,omitempty"`
	CustomTags map[string]string `json:"customTags,omitempty"`

	Meta map[string]any `json:"meta,omitempty"`
}

// Validate checks the structural invariants the v0.1 spec promises.
// We validate at boundaries (before outbox write); internal code may trust
// already-validated payloads.
func (p *HookEventPayload) Validate() error {
	var errs []string
	if p.Ts <= 0 {
		errs = append(errs, "ts must be positive")
	}
	if !p.Event.IsKnown() {
		errs = append(errs, fmt.Sprintf("unknown event %q", p.Event))
	}
	if p.HookName == "" {
		errs = append(errs, "hookName must not be empty")
	}
	if len(p.HookName) > 100 {
		errs = append(errs, "hookName exceeds 100 chars")
	}
	if p.DurationMs < 0 {
		errs = append(errs, "durationMs must be >= 0")
	}
	if p.PID < 0 {
		errs = append(errs, "pid must be >= 0")
	}
	if p.TokenUsage != nil {
		if p.TokenUsage.InputTokens < 0 ||
			p.TokenUsage.OutputTokens < 0 ||
			p.TokenUsage.CacheReadTokens < 0 ||
			p.TokenUsage.CacheCreateTokens < 0 {
			errs = append(errs, "tokenUsage values must be >= 0")
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New("HookEventPayload invalid: " + strings.Join(errs, ", "))
}
