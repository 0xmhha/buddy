package schema

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validBase() HookEventPayload {
	return HookEventPayload{
		Ts:         1700_000_000_000,
		Event:      EventPreToolUse,
		HookName:   "pre-commit",
		DurationMs: 42,
		ExitCode:   0,
	}
}

func TestValidate_AcceptsMinimalPayload(t *testing.T) {
	p := validBase()
	require.NoError(t, p.Validate())
}

func TestValidate_AcceptsFullOptionAPayload(t *testing.T) {
	p := validBase()
	p.SessionID = "sess-abc"
	p.PID = 12345
	p.Cwd = "/tmp/project"
	p.ToolName = "Bash"
	p.ToolArgs = map[string]any{"command": "npm test"}
	p.ModelName = "claude-opus-4-7"
	p.TokenUsage = &TokenUsage{
		InputTokens:       1000,
		OutputTokens:      200,
		CacheReadTokens:   500,
		CacheCreateTokens: 0,
	}
	p.CustomTags = map[string]string{"branch": "main"}
	require.NoError(t, p.Validate())
}

func TestValidate_RejectsUnknownEvent(t *testing.T) {
	p := validBase()
	p.Event = HookEventName("NotARealEvent")
	assert.Error(t, p.Validate())
}

func TestValidate_RejectsNegativeDuration(t *testing.T) {
	p := validBase()
	p.DurationMs = -1
	assert.Error(t, p.Validate())
}

func TestValidate_RejectsEmptyHookName(t *testing.T) {
	p := validBase()
	p.HookName = ""
	assert.Error(t, p.Validate())
}

func TestValidate_RejectsTooLongHookName(t *testing.T) {
	p := validBase()
	p.HookName = strings.Repeat("a", 101)
	assert.Error(t, p.Validate())
}

func TestValidate_RejectsNegativeTokenUsage(t *testing.T) {
	p := validBase()
	p.TokenUsage = &TokenUsage{InputTokens: -1}
	assert.Error(t, p.Validate())
}

func TestIsKnown(t *testing.T) {
	assert.True(t, EventPreToolUse.IsKnown())
	assert.True(t, EventStop.IsKnown())
	assert.False(t, HookEventName("Bogus").IsKnown())
}
