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
