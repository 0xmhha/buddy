package db

import (
	"database/sql"
	"fmt"
	"time"
)

type migration struct {
	version int
	sql     string
}

// migrations is append-only. Never edit a past entry; add a new one.
var migrations = []migration{
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
	{
		version: 2,
		sql: `
			CREATE TABLE features (
				feature_id          TEXT    PRIMARY KEY,
				name                TEXT    NOT NULL,
				summary             TEXT    NOT NULL DEFAULT '',
				actors              TEXT    NOT NULL DEFAULT '[]',
				acceptance_criteria TEXT    NOT NULL DEFAULT '[]',
				test_plan           TEXT    NOT NULL DEFAULT '{}',
				status              TEXT    NOT NULL DEFAULT 'draft',
				updated_at          INTEGER NOT NULL
			);
			CREATE INDEX idx_features_status  ON features(status);
			CREATE INDEX idx_features_updated ON features(updated_at);
		`,
	},
	{
		version: 3,
		sql: `
			CREATE TABLE sessions (
				id                   TEXT    PRIMARY KEY,
				pid                  INTEGER,
				transcript_path      TEXT    NOT NULL,
				started_at           INTEGER NOT NULL,
				last_active          INTEGER NOT NULL,
				total_input_tokens   INTEGER NOT NULL DEFAULT 0,
				total_output_tokens  INTEGER NOT NULL DEFAULT 0,
				total_cache_read     INTEGER NOT NULL DEFAULT 0,
				total_cache_create   INTEGER NOT NULL DEFAULT 0,
				last_offset          INTEGER NOT NULL DEFAULT 0
			);
			CREATE INDEX idx_sessions_last_active     ON sessions(last_active);
			CREATE INDEX idx_sessions_transcript_path ON sessions(transcript_path);
		`,
	},
	{
		// v4 — cli buddy W3-3 agent runtime tables. Per cli-buddy-spec §3.3
		// + ADR-005 lock-in. agents = static definition, agent_runs = one
		// row per Run(agent) invocation, agent_logs = streaming log events
		// (line-level) so the TUI / future buddy:status can tail without
		// touching the filesystem.
		version: 4,
		sql: `
			CREATE TABLE agents (
				id           TEXT    PRIMARY KEY,
				name         TEXT    NOT NULL,
				spec_yaml    TEXT    NOT NULL,
				schedule     TEXT    NOT NULL DEFAULT '',
				status       TEXT    NOT NULL DEFAULT 'idle',
				created_at   INTEGER NOT NULL,
				updated_at   INTEGER NOT NULL,
				last_run_at  INTEGER
			);
			CREATE INDEX idx_agents_status     ON agents(status);
			CREATE INDEX idx_agents_updated_at ON agents(updated_at);

			CREATE TABLE agent_runs (
				id            INTEGER PRIMARY KEY AUTOINCREMENT,
				agent_id      TEXT    NOT NULL,
				started_at    INTEGER NOT NULL,
				ended_at      INTEGER,
				exit_code     INTEGER NOT NULL DEFAULT 0,
				error         TEXT,
				result_json   TEXT    NOT NULL DEFAULT '{}',
				FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
			);
			CREATE INDEX idx_agent_runs_agent_started ON agent_runs(agent_id, started_at);

			CREATE TABLE agent_logs (
				id         INTEGER PRIMARY KEY AUTOINCREMENT,
				run_id     INTEGER NOT NULL,
				ts         INTEGER NOT NULL,
				level      TEXT    NOT NULL,
				message    TEXT    NOT NULL,
				FOREIGN KEY (run_id) REFERENCES agent_runs(id) ON DELETE CASCADE
			);
			CREATE INDEX idx_agent_logs_run_ts ON agent_logs(run_id, ts);
		`,
	},
	{
		// v5 — F2.A Session Monitor (W7-1) — ADR-012 lock-in. Additive
		// columns on the v3 sessions table:
		//   - ended_at  : NULL while session may be active. Daemon sets it
		//                 when last_active is older than EndedThreshold
		//                 (default 1h). Cleared on resume — no row split.
		//   - goal_text : first user message extracted by fsLister. F2.D
		//                 (W7-4) Drift Detection compares against this.
		//   - metadata  : JSON escape hatch for future per-session fields.
		// All three are backward-compat: existing rows get NULL / "" / "{}".
		version: 5,
		sql: `
			ALTER TABLE sessions ADD COLUMN ended_at  INTEGER;
			ALTER TABLE sessions ADD COLUMN goal_text TEXT    NOT NULL DEFAULT '';
			ALTER TABLE sessions ADD COLUMN metadata  TEXT    NOT NULL DEFAULT '{}';
			CREATE INDEX idx_sessions_ended_at ON sessions(ended_at);
		`,
	},
	{
		// v6 — F2.C Advisory foundation (W7-3a) — ADR-014. The chunks
		// table holds knowledge data extracted from sessions transcripts:
		//   - session_id     : FK back to sessions(id) — cascade delete.
		//   - content        : raw chunk text (user+assistant pair, max
		//                      ~500 tokens). Used as BM25 source.
		//   - token_count    : naive whitespace-split estimate; lets
		//                      retrieval rank by chunk size.
		//   - embedding      : NULL until the Python embedder fills it.
		//                      JSON-encoded float32 array as a BLOB so
		//                      pure-Go SQLite (modernc.org/sqlite) can
		//                      store it without CGo vector extensions.
		//   - created_at     : epoch ms.
		version: 6,
		sql: `
			CREATE TABLE chunks (
				id          INTEGER PRIMARY KEY AUTOINCREMENT,
				session_id  TEXT    NOT NULL,
				content     TEXT    NOT NULL,
				token_count INTEGER NOT NULL,
				embedding   BLOB,
				created_at  INTEGER NOT NULL,
				FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
			);
			CREATE INDEX idx_chunks_session ON chunks(session_id);
			CREATE INDEX idx_chunks_created ON chunks(created_at);
		`,
	},
	{
		// v7 — F2.C Advisor (W7-3b) — ADR-015. The advisories table
		// holds advisor output records for dedup + history + future mute
		// UX. Columns:
		//   - kind          : rule identifier (e.g., "token-spike-day").
		//                     Indexed for dedup-window lookups.
		//   - severity      : "info" | "warn" | "high" (free text — Go
		//                     side enforces the enum via persona.Severity).
		//   - message       : friend-tone Korean prose rendered via
		//                     persona.M at advisory generation time.
		//   - evidence_json : JSON-encoded []EvidenceItem (metric values
		//                     + chunk references that backed the
		//                     advisory). Held as text rather than a
		//                     side table because evidence is only ever
		//                     read whole alongside its advisory.
		//   - created_at    : epoch ms.
		//   - muted         : 0/1. Future mute UX flips this.
		version: 7,
		sql: `
			CREATE TABLE advisories (
				id            INTEGER PRIMARY KEY AUTOINCREMENT,
				kind          TEXT    NOT NULL,
				severity      TEXT    NOT NULL,
				message       TEXT    NOT NULL,
				evidence_json TEXT    NOT NULL DEFAULT '[]',
				created_at    INTEGER NOT NULL,
				muted         INTEGER NOT NULL DEFAULT 0
			);
			CREATE INDEX idx_advisories_created    ON advisories(created_at);
			CREATE INDEX idx_advisories_kind_muted ON advisories(kind, muted);
		`,
	},
}

// RunMigrations applies every migration whose version is greater than the
// highest applied version. Idempotent across opens.
func RunMigrations(conn *sql.DB) error {
	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version    INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL
		);`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	var current sql.NullInt64
	row := conn.QueryRow("SELECT MAX(version) FROM schema_version")
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("read schema_version: %w", err)
	}

	for _, m := range migrations {
		if int64(m.version) <= current.Int64 {
			continue
		}
		if err := applyOne(conn, m); err != nil {
			return fmt.Errorf("apply v%d: %w", m.version, err)
		}
	}
	return nil
}

func applyOne(conn *sql.DB, m migration) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.sql); err != nil {
		return err
	}
	if _, err := tx.Exec(
		"INSERT INTO schema_version (version, applied_at) VALUES (?, ?)",
		m.version, time.Now().UnixMilli(),
	); err != nil {
		return err
	}
	return tx.Commit()
}
