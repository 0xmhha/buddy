import { describe, expect, it } from 'vitest'
import { HookEventPayload } from '../src/schema/hook-event.js'

describe('HookEventPayload', () => {
  it('accepts a minimal valid payload', () => {
    const result = HookEventPayload.safeParse({
      ts: Date.now(),
      event: 'PreToolUse',
      hookName: 'pre-commit',
      durationMs: 42,
      exitCode: 0,
    })
    expect(result.success).toBe(true)
  })

  it('accepts a full payload with option-A fields', () => {
    const result = HookEventPayload.safeParse({
      ts: Date.now(),
      event: 'PostToolUse',
      hookName: 'lint',
      durationMs: 120,
      exitCode: 0,
      sessionId: 'sess-abc',
      pid: 12345,
      cwd: '/tmp/project',
      toolName: 'Bash',
      toolArgs: { command: 'npm test' },
      modelName: 'claude-opus-4-7',
      tokenUsage: {
        inputTokens: 1000,
        outputTokens: 200,
        cacheReadTokens: 500,
        cacheCreateTokens: 0,
      },
      customTags: { branch: 'main', experiment: 'A' },
    })
    expect(result.success).toBe(true)
  })

  it('rejects unknown event names', () => {
    const result = HookEventPayload.safeParse({
      ts: Date.now(),
      event: 'NotARealEvent',
      hookName: 'x',
      durationMs: 0,
      exitCode: 0,
    })
    expect(result.success).toBe(false)
  })

  it('rejects negative duration', () => {
    const result = HookEventPayload.safeParse({
      ts: Date.now(),
      event: 'Stop',
      hookName: 'x',
      durationMs: -1,
      exitCode: 0,
    })
    expect(result.success).toBe(false)
  })

  it('rejects empty hook name', () => {
    const result = HookEventPayload.safeParse({
      ts: Date.now(),
      event: 'Stop',
      hookName: '',
      durationMs: 0,
      exitCode: 0,
    })
    expect(result.success).toBe(false)
  })
})
