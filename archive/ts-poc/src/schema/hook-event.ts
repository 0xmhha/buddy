import { z } from 'zod'

export const HookEventName = z.enum([
  'SessionStart',
  'PreToolUse',
  'PostToolUse',
  'Stop',
  'PreCompact',
  'UserPromptSubmit',
])
export type HookEventName = z.infer<typeof HookEventName>

export const TokenUsage = z.object({
  inputTokens: z.number().int().nonnegative(),
  outputTokens: z.number().int().nonnegative(),
  cacheReadTokens: z.number().int().nonnegative(),
  cacheCreateTokens: z.number().int().nonnegative(),
})
export type TokenUsage = z.infer<typeof TokenUsage>

export const HookEventPayload = z.object({
  ts: z.number().int().positive(),
  event: HookEventName,
  hookName: z.string().min(1).max(100),
  durationMs: z.number().int().min(0),
  exitCode: z.number().int(),

  sessionId: z.string().optional(),
  pid: z.number().int().positive().optional(),
  cwd: z.string().optional(),

  toolName: z.string().optional(),
  toolArgs: z.unknown().optional(),
  modelName: z.string().optional(),
  tokenUsage: TokenUsage.optional(),
  customTags: z.record(z.string()).optional(),

  meta: z.record(z.unknown()).optional(),
})
export type HookEventPayload = z.infer<typeof HookEventPayload>
