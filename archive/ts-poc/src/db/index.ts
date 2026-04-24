import { createRequire } from 'node:module'
import { mkdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { homedir } from 'node:os'
import { runMigrations } from './migrations.js'

const nodeRequire = createRequire(import.meta.url)
const { DatabaseSync } = nodeRequire('node:sqlite') as typeof import('node:sqlite')

export type DatabaseSync = InstanceType<typeof DatabaseSync>

export interface OpenDbOptions {
  path?: string
  readonly?: boolean
}

export function defaultDbPath(): string {
  return join(homedir(), '.buddy', 'buddy.db')
}

function applyPragma(db: DatabaseSync, statement: string): void {
  db.exec(statement)
}

export function openDb(opts: OpenDbOptions = {}): DatabaseSync {
  const path = opts.path ?? defaultDbPath()
  mkdirSync(dirname(path), { recursive: true })

  const db = new DatabaseSync(path, { readOnly: opts.readonly === true })
  applyPragma(db, 'PRAGMA journal_mode = WAL')
  applyPragma(db, 'PRAGMA synchronous = NORMAL')
  applyPragma(db, 'PRAGMA foreign_keys = ON')

  if (!opts.readonly) {
    runMigrations(db)
  }
  return db
}
