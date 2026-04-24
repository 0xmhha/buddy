import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { Readable } from 'node:stream'
import { runHookWrap } from '../src/cli/hook-wrap.js'
import { openDb } from '../src/db/index.js'

let tmpDir: string
let dbPath: string

beforeEach(() => {
  tmpDir = mkdtempSync(join(tmpdir(), 'buddy-hookwrap-'))
  dbPath = join(tmpDir, 'buddy.db')
})

afterEach(() => {
  rmSync(tmpDir, { recursive: true, force: true })
})

// stdin을 readable stream으로 교체하기 위한 헬퍼
function withStdin<T>(input: string, fn: () => Promise<T>): Promise<T> {
  const original = process.stdin
  const fake = Readable.from([Buffer.from(input, 'utf8')]) as unknown as typeof process.stdin
  Object.defineProperty(process, 'stdin', { value: fake, configurable: true })
  return fn().finally(() => {
    Object.defineProperty(process, 'stdin', { value: original, configurable: true })
  })
}

describe('hook-wrap', () => {
  it('records a successful hook execution', async () => {
    const stdin = JSON.stringify({
      session_id: 'sess-x',
      cwd: '/tmp/x',
      hook_event_name: 'PreToolUse',
      tool_name: 'Bash',
      tool_input: { command: 'echo hi' },
    })

    const result = await withStdin(stdin, () =>
      runHookWrap({
        hookName: 'pre-commit',
        command: ['node', '-e', 'process.exit(0)'],
        dbPath,
      }),
    )

    expect(result.exitCode).toBe(0)
    expect(result.outboxId).not.toBeNull()

    const db = openDb({ path: dbPath })
    const row = db
      .prepare('SELECT payload FROM hook_outbox WHERE id = ?')
      .get(result.outboxId) as { payload: string }
    db.close()

    const payload = JSON.parse(row.payload)
    expect(payload.exitCode).toBe(0)
    expect(payload.hookName).toBe('pre-commit')
    expect(payload.event).toBe('PreToolUse')
    expect(payload.toolName).toBe('Bash')
    expect(payload.sessionId).toBe('sess-x')
    expect(payload.cwd).toBe('/tmp/x')
    expect(payload.toolArgs).toBeUndefined()  // default off
  })

  it('records a failing hook execution with non-zero exit code', async () => {
    const result = await withStdin('{}', () =>
      runHookWrap({
        hookName: 'lint',
        command: ['node', '-e', 'process.exit(7)'],
        dbPath,
      }),
    )

    expect(result.exitCode).toBe(7)

    const db = openDb({ path: dbPath })
    const row = db
      .prepare('SELECT payload FROM hook_outbox WHERE id = ?')
      .get(result.outboxId) as { payload: string }
    db.close()

    expect(JSON.parse(row.payload).exitCode).toBe(7)
  })

  it('records duration > 0 for slow hooks', async () => {
    const result = await withStdin('{}', () =>
      runHookWrap({
        hookName: 'slow',
        command: [
          'node',
          '-e',
          'setTimeout(() => process.exit(0), 80)',
        ],
        dbPath,
      }),
    )

    const db = openDb({ path: dbPath })
    const row = db
      .prepare('SELECT payload FROM hook_outbox WHERE id = ?')
      .get(result.outboxId) as { payload: string }
    db.close()

    const payload = JSON.parse(row.payload)
    expect(payload.durationMs).toBeGreaterThanOrEqual(50)
  })

  it('records toolArgs when recordToolArgs=true', async () => {
    const stdin = JSON.stringify({
      tool_name: 'Bash',
      tool_input: { command: 'ls -la' },
    })
    const result = await withStdin(stdin, () =>
      runHookWrap({
        hookName: 'pre',
        command: ['node', '-e', 'process.exit(0)'],
        dbPath,
        recordToolArgs: true,
      }),
    )

    const db = openDb({ path: dbPath })
    const row = db
      .prepare('SELECT payload FROM hook_outbox WHERE id = ?')
      .get(result.outboxId) as { payload: string }
    db.close()

    const payload = JSON.parse(row.payload)
    expect(payload.toolArgs).toEqual({ command: 'ls -la' })
  })

  it('still records when child command is empty (monitoring-only mode)', async () => {
    const stdin = JSON.stringify({ hook_event_name: 'Stop' })
    const result = await withStdin(stdin, () =>
      runHookWrap({
        hookName: 'noop',
        command: [],
        dbPath,
      }),
    )
    expect(result.exitCode).toBe(0)
    expect(result.outboxId).not.toBeNull()
  })

  it('returns exit code 127 when spawn fails', async () => {
    const result = await withStdin('{}', () =>
      runHookWrap({
        hookName: 'broken',
        command: ['/nonexistent/binary-' + Math.random()],
        dbPath,
      }),
    )
    expect(result.exitCode).toBe(127)
  })

  it('forwards customTags to outbox', async () => {
    const result = await withStdin('{}', () =>
      runHookWrap({
        hookName: 'tagged',
        command: ['node', '-e', 'process.exit(0)'],
        dbPath,
        customTags: { branch: 'feature/x', exp: 'A' },
      }),
    )

    const db = openDb({ path: dbPath })
    const row = db
      .prepare('SELECT payload FROM hook_outbox WHERE id = ?')
      .get(result.outboxId) as { payload: string }
    db.close()

    const payload = JSON.parse(row.payload)
    expect(payload.customTags).toEqual({ branch: 'feature/x', exp: 'A' })
  })
})
