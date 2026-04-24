package adapter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/adapter"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

func TestParse_EmptyInput(t *testing.T) {
	assert.Equal(t, adapter.ClaudeHookInput{}, adapter.Parse(""))
	assert.Equal(t, adapter.ClaudeHookInput{}, adapter.Parse("   \n  "))
}

func TestParse_TypicalInput(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"session_id":      "sess-1",
		"cwd":             "/tmp/x",
		"hook_event_name": "PreToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": "ls"},
	})
	require.NoError(t, err)

	in := adapter.Parse(string(raw))
	assert.Equal(t, "sess-1", in.SessionID)
	assert.Equal(t, "Bash", in.ToolName)
	assert.Equal(t, "PreToolUse", in.HookEventName)
}

func TestParse_AbsorbsMalformedJSON(t *testing.T) {
	assert.Equal(t, adapter.ClaudeHookInput{}, adapter.Parse("{not json}"))
	assert.Equal(t, adapter.ClaudeHookInput{}, adapter.Parse("null"))
}

func baseOpts() adapter.AdaptOptions {
	return adapter.AdaptOptions{
		HookName:      "pre-commit",
		FallbackEvent: schema.EventPreToolUse,
		StartedAt:     1000,
		FinishedAt:    1250,
		ExitCode:      0,
	}
}

func TestBuild_MinimalFromEmptyInput(t *testing.T) {
	p := adapter.Build(adapter.ClaudeHookInput{}, baseOpts())
	assert.Equal(t, schema.EventPreToolUse, p.Event)
	assert.Equal(t, "pre-commit", p.HookName)
	assert.Equal(t, int64(250), p.DurationMs)
	assert.Equal(t, 0, p.ExitCode)
	assert.Empty(t, p.ToolName)
	assert.Nil(t, p.ToolArgs)
}

func TestBuild_UsesEventFromInputWhenKnown(t *testing.T) {
	p := adapter.Build(
		adapter.ClaudeHookInput{HookEventName: "PostToolUse"}, baseOpts(),
	)
	assert.Equal(t, schema.EventPostToolUse, p.Event)
}

func TestBuild_FallsBackOnUnknownEvent(t *testing.T) {
	p := adapter.Build(
		adapter.ClaudeHookInput{HookEventName: "NotARealEvent"}, baseOpts(),
	)
	assert.Equal(t, schema.EventPreToolUse, p.Event)
}

func TestBuild_OmitsToolArgsByDefault(t *testing.T) {
	p := adapter.Build(adapter.ClaudeHookInput{
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "rm -rf /"},
	}, baseOpts())
	assert.Equal(t, "Bash", p.ToolName)
	assert.Nil(t, p.ToolArgs)
}

func TestBuild_RecordsToolArgsWhenEnabled(t *testing.T) {
	opts := baseOpts()
	opts.RecordToolArgs = true
	p := adapter.Build(adapter.ClaudeHookInput{
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls"},
	}, opts)
	require.NotNil(t, p.ToolArgs)
	m := p.ToolArgs.(map[string]any)
	assert.Equal(t, "ls", m["command"])
}

func TestBuild_IncludesCustomTags(t *testing.T) {
	opts := baseOpts()
	opts.CustomTags = map[string]string{"branch": "main"}
	p := adapter.Build(adapter.ClaudeHookInput{}, opts)
	assert.Equal(t, "main", p.CustomTags["branch"])
}

func TestBuild_ClampsNegativeDurationToZero(t *testing.T) {
	opts := baseOpts()
	opts.StartedAt, opts.FinishedAt = 2000, 1000
	p := adapter.Build(adapter.ClaudeHookInput{}, opts)
	assert.Equal(t, int64(0), p.DurationMs)
}
