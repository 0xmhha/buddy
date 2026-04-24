import { z } from 'zod'
import type { HookEventName, HookEventPayload } from '../schema/hook-event.js'

// Claude Code가 hook stdin으로 보내는 JSON의 알려진 형태.
// 필드 누락에 관대 (모든 것 optional) — Claude Code 버전 차이를 흡수.
export const ClaudeHookInput = z.object({
  session_id: z.string().optional(),
  transcript_path: z.string().optional(),
  cwd: z.string().optional(),
  hook_event_name: z.string().optional(),
  tool_name: z.string().optional(),
  tool_input: z.unknown().optional(),
}).passthrough()
export type ClaudeHookInput = z.infer<typeof ClaudeHookInput>

const KNOWN_EVENTS: readonly HookEventName[] = [
  'SessionStart',
  'PreToolUse',
  'PostToolUse',
  'Stop',
  'PreCompact',
  'UserPromptSubmit',
] as const

function normalizeEvent(raw: string | undefined, fallback: HookEventName): HookEventName {
  if (raw && (KNOWN_EVENTS as readonly string[]).includes(raw)) {
    return raw as HookEventName
  }
  return fallback
}

export interface AdaptOptions {
  hookName: string
  fallbackEvent: HookEventName
  startedAt: number
  finishedAt: number
  exitCode: number
  pid?: number
  recordToolArgs: boolean
  customTags?: Record<string, string>
}

export function buildPayloadFromHookInput(
  input: ClaudeHookInput,
  opts: AdaptOptions,
): HookEventPayload {
  const payload: HookEventPayload = {
    ts: opts.finishedAt,
    event: normalizeEvent(input.hook_event_name, opts.fallbackEvent),
    hookName: opts.hookName,
    durationMs: Math.max(0, opts.finishedAt - opts.startedAt),
    exitCode: opts.exitCode,
  }

  if (input.session_id) payload.sessionId = input.session_id
  if (opts.pid !== undefined) payload.pid = opts.pid
  if (input.cwd) payload.cwd = input.cwd
  if (input.tool_name) payload.toolName = input.tool_name
  if (opts.recordToolArgs && input.tool_input !== undefined) {
    payload.toolArgs = input.tool_input
  }
  if (opts.customTags) payload.customTags = opts.customTags

  return payload
}

export function parseHookInput(raw: string): ClaudeHookInput {
  const trimmed = raw.trim()
  if (trimmed.length === 0) return {}
  try {
    const parsed = JSON.parse(trimmed)
    return ClaudeHookInput.parse(parsed)
  } catch {
    // 알 수 없는 입력은 빈 객체로 흡수 — wrapper는 절대 깨지지 않음
    return {}
  }
}
