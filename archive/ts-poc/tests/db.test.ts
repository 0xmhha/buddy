import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { openDb } from '../src/db/index.js'
import {
  appendToOutbox,
  markConsumed,
  readPendingOutbox,
} from '../src/db/outbox.js'
import type { HookEventPayload } from '../src/schema/hook-event.js'

let tmpDir: string

beforeEach(() => {
  tmpDir = mkdtempSync(join(tmpdir(), 'buddy-test-'))
})

afterEach(() => {
  rmSync(tmpDir, { recursive: true, force: true })
})

describe('db bootstrap', () => {
  it('creates schema tables on first open', () => {
    const db = openDb({ path: join(tmpDir, 'buddy.db') })
    const tables = db
      .prepare(
        "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name",
      )
      .all() as { name: string }[]

    const names = tables.map((t) => t.name)
    expect(names).toContain('hook_outbox')
    expect(names).toContain('hook_events')
    expect(names).toContain('hook_stats')
    expect(names).toContain('schema_version')
    db.close()
  })

  it('records schema version 1 after bootstrap', () => {
    const db = openDb({ path: join(tmpDir, 'buddy.db') })
    const row = db
      .prepare('SELECT MAX(version) AS v FROM schema_version')
      .get() as { v: number }
    expect(row.v).toBe(1)
    db.close()
  })

  it('uses WAL journal mode', () => {
    const db = openDb({ path: join(tmpDir, 'buddy.db') })
    const row = db.prepare('PRAGMA journal_mode').get() as { journal_mode: string }
    expect(row.journal_mode).toBe('wal')
    db.close()
  })

  it('is idempotent across opens (no double-migration)', () => {
    const path = join(tmpDir, 'buddy.db')
    const db1 = openDb({ path })
    db1.close()
    const db2 = openDb({ path })
    const count = (db2
      .prepare('SELECT COUNT(*) AS c FROM schema_version')
      .get() as { c: number }).c
    expect(count).toBe(1)
    db2.close()
  })
})

describe('outbox', () => {
  it('appends and reads back pending rows', () => {
    const db = openDb({ path: join(tmpDir, 'buddy.db') })
    const payload: HookEventPayload = {
      ts: Date.now(),
      event: 'PreToolUse',
      hookName: 'pre-commit',
      durationMs: 42,
      exitCode: 0,
      toolName: 'Bash',
    }
    const id = appendToOutbox(db, payload)
    expect(id).toBeGreaterThan(0)

    const pending = readPendingOutbox(db)
    expect(pending.length).toBe(1)
    expect(pending[0]?.id).toBe(id)
    db.close()
  })

  it('marks consumed rows so they are no longer pending', () => {
    const db = openDb({ path: join(tmpDir, 'buddy.db') })
    const payload: HookEventPayload = {
      ts: Date.now(),
      event: 'Stop',
      hookName: 'cleanup',
      durationMs: 1,
      exitCode: 0,
    }
    const id = appendToOutbox(db, payload)
    markConsumed(db, [id])

    const pending = readPendingOutbox(db)
    expect(pending.length).toBe(0)
    db.close()
  })
})
