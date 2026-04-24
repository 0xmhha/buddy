// Package adapter translates the Claude Code hook stdin JSON into our schema.
package adapter

import (
	"encoding/json"
	"strings"

	"github.com/wm-it-22-00661/buddy/internal/schema"
)

// ClaudeHookInput is the relevant subset of what Claude Code sends on stdin.
// Unknown fields are tolerated by the JSON decoder by default.
type ClaudeHookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	ToolName       string `json:"tool_name"`
	ToolInput      any    `json:"tool_input"`
}

// Parse tolerates malformed input — wrappers must never crash on a bad stdin.
func Parse(raw string) ClaudeHookInput {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ClaudeHookInput{}
	}
	var in ClaudeHookInput
	if err := json.Unmarshal([]byte(trimmed), &in); err != nil {
		return ClaudeHookInput{}
	}
	return in
}

// AdaptOptions carries the wrapper-supplied measurements that the
// stdin alone can't provide.
type AdaptOptions struct {
	HookName       string
	FallbackEvent  schema.HookEventName
	StartedAt      int64 // unix ms
	FinishedAt     int64 // unix ms
	ExitCode       int
	PID            int
	RecordToolArgs bool
	CustomTags     map[string]string
}

// Build composes a HookEventPayload from stdin + measurements.
func Build(in ClaudeHookInput, opts AdaptOptions) schema.HookEventPayload {
	event := schema.HookEventName(in.HookEventName)
	if !event.IsKnown() {
		event = opts.FallbackEvent
	}

	dur := max(opts.FinishedAt-opts.StartedAt, 0)

	p := schema.HookEventPayload{
		Ts:         opts.FinishedAt,
		Event:      event,
		HookName:   opts.HookName,
		DurationMs: dur,
		ExitCode:   opts.ExitCode,
	}
	if in.SessionID != "" {
		p.SessionID = in.SessionID
	}
	if opts.PID > 0 {
		p.PID = opts.PID
	}
	if in.Cwd != "" {
		p.Cwd = in.Cwd
	}
	if in.ToolName != "" {
		p.ToolName = in.ToolName
	}
	if opts.RecordToolArgs && in.ToolInput != nil {
		p.ToolArgs = in.ToolInput
	}
	if len(opts.CustomTags) > 0 {
		p.CustomTags = opts.CustomTags
	}
	return p
}
