import { spawn } from 'node:child_process'
import { Buffer } from 'node:buffer'
import { openDb, defaultDbPath } from '../db/index.js'
import { appendToOutbox } from '../db/outbox.js'
import {
  buildPayloadFromHookInput,
  parseHookInput,
} from '../adapter/claude-hook-input.js'
import type { HookEventName } from '../schema/hook-event.js'

const FALLBACK_EVENT: HookEventName = 'PreToolUse'

export interface HookWrapOptions {
  hookName: string
  command: string[]            // 원본 hook (argv). 비어 있으면 no-op (모니터링만)
  fallbackEvent?: HookEventName
  dbPath?: string
  recordToolArgs?: boolean
  customTags?: Record<string, string>
}

export interface HookWrapResult {
  exitCode: number
  outboxId: number | null      // null이면 outbox 기록 실패 (그래도 hook은 통과)
}

// stdin 전체를 동기적으로 읽기 (작은 JSON 한 번)
async function readAllStdin(): Promise<string> {
  const chunks: Buffer[] = []
  for await (const chunk of process.stdin) {
    chunks.push(chunk as Buffer)
  }
  return Buffer.concat(chunks).toString('utf8')
}

export async function runHookWrap(opts: HookWrapOptions): Promise<HookWrapResult> {
  const startedAt = Date.now()

  // 1. stdin을 buffer로 한 번 읽음 — buddy는 hook input을 검사도 하고
  //    원본 hook에도 다시 넘겨야 하므로
  const rawInput = await readAllStdin()
  const parsed = parseHookInput(rawInput)

  // 2. 원본 hook 실행 (있을 때만). stdio passthrough로 streaming 보장.
  let exitCode = 0
  let childPid: number | undefined

  if (opts.command.length > 0) {
    const [cmd, ...args] = opts.command
    if (cmd === undefined) {
      throw new Error('hook-wrap: command[0] is undefined')
    }
    const child = spawn(cmd, args, {
      stdio: ['pipe', 'inherit', 'inherit'],
    })
    childPid = child.pid

    // child stdin에 원본 input을 다시 흘려준다
    if (child.stdin) {
      child.stdin.end(rawInput)
    }

    exitCode = await new Promise<number>((resolve) => {
      child.on('error', () => resolve(127))   // spawn 자체 실패
      child.on('exit', (code, signal) => {
        if (signal) resolve(128 + (signalToNumber(signal) ?? 0))
        else resolve(code ?? 0)
      })
    })
  }

  const finishedAt = Date.now()

  // 3. payload 빌드 + outbox 기록 (실패해도 wrapper는 hook의 exit code를 통과)
  const payload = buildPayloadFromHookInput(parsed, {
    hookName: opts.hookName,
    fallbackEvent: opts.fallbackEvent ?? FALLBACK_EVENT,
    startedAt,
    finishedAt,
    exitCode,
    pid: childPid,
    recordToolArgs: opts.recordToolArgs ?? false,
    customTags: opts.customTags,
  })

  let outboxId: number | null = null
  try {
    const db = openDb({ path: opts.dbPath ?? defaultDbPath() })
    outboxId = appendToOutbox(db, payload)
    db.close()
  } catch (err) {
    // 절대 hook을 깨지 않는다. stderr에 한 줄만.
    process.stderr.write(
      `buddy: outbox write failed (${(err as Error).message ?? 'unknown'})\n`,
    )
  }

  return { exitCode, outboxId }
}

const SIGNAL_TABLE: Record<string, number> = {
  SIGHUP: 1,
  SIGINT: 2,
  SIGQUIT: 3,
  SIGKILL: 9,
  SIGTERM: 15,
}
function signalToNumber(sig: NodeJS.Signals): number | undefined {
  return SIGNAL_TABLE[sig]
}
