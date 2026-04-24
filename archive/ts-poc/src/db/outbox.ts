import type { DatabaseSync } from 'node:sqlite'
import type { HookEventPayload } from '../schema/hook-event.js'

export function appendToOutbox(
  db: DatabaseSync,
  payload: HookEventPayload,
): number {
  const stmt = db.prepare(
    'INSERT INTO hook_outbox (ts, payload) VALUES (?, ?)',
  )
  const result = stmt.run(payload.ts, JSON.stringify(payload))
  return Number(result.lastInsertRowid)
}

export interface OutboxRow {
  id: number
  ts: number
  payload: string
}

export function readPendingOutbox(
  db: DatabaseSync,
  limit = 100,
): OutboxRow[] {
  const stmt = db.prepare(
    'SELECT id, ts, payload FROM hook_outbox WHERE consumed_at IS NULL ORDER BY id LIMIT ?',
  )
  return stmt.all(limit) as unknown as OutboxRow[]
}

export function markConsumed(db: DatabaseSync, ids: number[]): void {
  if (ids.length === 0) return
  const placeholders = ids.map(() => '?').join(',')
  const stmt = db.prepare(
    `UPDATE hook_outbox SET consumed_at = ? WHERE id IN (${placeholders})`,
  )
  stmt.run(Date.now(), ...ids)
}
