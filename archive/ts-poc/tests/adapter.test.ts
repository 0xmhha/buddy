import { describe, expect, it } from 'vitest'
import {
  buildPayloadFromHookInput,
  parseHookInput,
} from '../src/adapter/claude-hook-input.js'

describe('parseHookInput', () => {
  it('returns {} for empty input', () => {
    expect(parseHookInput('')).toEqual({})
    expect(parseHookInput('   \n  ')).toEqual({})
  })

  it('parses a typical Claude Code hook input', () => {
    const raw = JSON.stringify({
      session_id: 'sess-1',
      cwd: '/tmp/x',
      hook_event_name: 'PreToolUse',
      tool_name: 'Bash',
      tool_input: { command: 'ls' },
    })
    const parsed = parseHookInput(raw)
    expect(parsed.session_id).toBe('sess-1')
    expect(parsed.tool_name).toBe('Bash')
  })

  it('absorbs malformed JSON instead of throwing', () => {
    expect(parseHookInput('{not json}')).toEqual({})
    expect(parseHookInput('null')).toEqual({})
  })

  it('preserves unknown fields via passthrough', () => {
    const parsed = parseHookInput(
      JSON.stringify({ session_id: 's', extra_future_field: 42 }),
    )
    expect((parsed as Record<string, unknown>).extra_future_field).toBe(42)
  })
})

describe('buildPayloadFromHookInput', () => {
  const baseOpts = {
    hookName: 'pre-commit',
    fallbackEvent: 'PreToolUse' as const,
    startedAt: 1000,
    finishedAt: 1250,
    exitCode: 0,
    recordToolArgs: false,
  }

  it('builds a minimal payload from empty input', () => {
    const payload = buildPayloadFromHookInput({}, baseOpts)
    expect(payload.event).toBe('PreToolUse')
    expect(payload.hookName).toBe('pre-commit')
    expect(payload.durationMs).toBe(250)
    expect(payload.exitCode).toBe(0)
    expect(payload.toolName).toBeUndefined()
    expect(payload.toolArgs).toBeUndefined()
  })

  it('uses event from input when known', () => {
    const payload = buildPayloadFromHookInput(
      { hook_event_name: 'PostToolUse' },
      baseOpts,
    )
    expect(payload.event).toBe('PostToolUse')
  })

  it('falls back when event is unknown', () => {
    const payload = buildPayloadFromHookInput(
      { hook_event_name: 'NotARealEvent' },
      baseOpts,
    )
    expect(payload.event).toBe('PreToolUse')
  })

  it('omits toolArgs by default (privacy)', () => {
    const payload = buildPayloadFromHookInput(
      { tool_name: 'Bash', tool_input: { command: 'rm -rf /' } },
      baseOpts,
    )
    expect(payload.toolName).toBe('Bash')
    expect(payload.toolArgs).toBeUndefined()
  })

  it('records toolArgs when explicitly enabled', () => {
    const payload = buildPayloadFromHookInput(
      { tool_name: 'Bash', tool_input: { command: 'ls' } },
      { ...baseOpts, recordToolArgs: true },
    )
    expect(payload.toolArgs).toEqual({ command: 'ls' })
  })

  it('includes customTags when provided', () => {
    const payload = buildPayloadFromHookInput(
      {},
      { ...baseOpts, customTags: { branch: 'main' } },
    )
    expect(payload.customTags).toEqual({ branch: 'main' })
  })

  it('clamps negative duration to 0', () => {
    const payload = buildPayloadFromHookInput({}, {
      ...baseOpts,
      startedAt: 2000,
      finishedAt: 1000,
    })
    expect(payload.durationMs).toBe(0)
  })
})
