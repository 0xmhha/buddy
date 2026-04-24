import type { DatabaseSync } from 'node:sqlite'

interface Migration {
  version: number
  sql: string
}

const MIGRATIONS: Migration[] = [
  {
    version: 1,
    sql: `
      CREATE TABLE hook_outbox (
        id          INTEGER PRIMARY KEY AUTOINCREMENT,
        ts          INTEGER NOT NULL,
        payload     TEXT    NOT NULL,
        consumed_at INTEGER
      );
      CREATE INDEX idx_outbox_pending
        ON hook_outbox(id) WHERE consumed_at IS NULL;

      CREATE TABLE hook_events (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        ts           INTEGER NOT NULL,
        event        TEXT    NOT NULL,
        hook_name    TEXT    NOT NULL,
        duration_ms  INTEGER NOT NULL,
        exit_code    INTEGER NOT NULL,
        session_id   TEXT,
        pid          INTEGER,
        cwd          TEXT,
        tool_name    TEXT,
        tool_args    TEXT,
        model_name   TEXT,
        token_usage  TEXT,
        custom_tags  TEXT,
        meta         TEXT
      );
      CREATE INDEX idx_events_hook_ts ON hook_events(hook_name, ts);
      CREATE INDEX idx_events_tool_ts ON hook_events(tool_name, ts);

      CREATE TABLE hook_stats (
        hook_name   TEXT    NOT NULL,
        tool_name   TEXT    NOT NULL DEFAULT '',
        window_min  INTEGER NOT NULL,
        ts_bucket   INTEGER NOT NULL,
        count       INTEGER NOT NULL,
        failures    INTEGER NOT NULL,
        p50_ms      INTEGER,
        p95_ms      INTEGER,
        p99_ms      INTEGER,
        PRIMARY KEY (hook_name, tool_name, window_min, ts_bucket)
      );
    `,
  },
]

function runSql(db: DatabaseSync, sql: string): void {
  db.exec(sql)
}

function applyMigration(db: DatabaseSync, m: Migration): void {
  runSql(db, 'BEGIN')
  try {
    runSql(db, m.sql)
    db.prepare(
      'INSERT INTO schema_version (version, applied_at) VALUES (?, ?)',
    ).run(m.version, Date.now())
    runSql(db, 'COMMIT')
  } catch (err) {
    runSql(db, 'ROLLBACK')
    throw err
  }
}

export function runMigrations(db: DatabaseSync): number {
  runSql(
    db,
    `CREATE TABLE IF NOT EXISTS schema_version (
       version INTEGER PRIMARY KEY,
       applied_at INTEGER NOT NULL
     );`,
  )

  const row = db
    .prepare('SELECT MAX(version) AS v FROM schema_version')
    .get() as { v: number | null }
  const current = row.v ?? 0

  let applied = 0
  for (const m of MIGRATIONS) {
    if (m.version > current) {
      applyMigration(db, m)
      applied++
    }
  }
  return applied
}
