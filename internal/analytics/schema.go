package analytics

import (
	"context"
	"database/sql"
	"fmt"
)

// Schema is the SQLite-flavoured DDL for the analytics-mcp reference adapter.
//
// PostgreSQL / MySQL portability:
//   - Timestamps are TEXT (ISO-8601 UTC). Portable across all three.
//   - Surrogate keys are INTEGER PRIMARY KEY AUTOINCREMENT. PostgreSQL/MySQL
//     adapters substitute GENERATED ALWAYS AS IDENTITY / AUTO_INCREMENT
//     during their own Migrate.
//   - No JSON column type — properties live in a TEXT field. Queries can
//     pattern-match if needed; structured filters use dedicated columns.
//
// Spec §8 reference impl is SQLite. Other backends layer on top of the
// same Adapter interface with their own DDL.
const Schema = `
-- events: unified event stream backing funnel / cohort / feature-adoption.
-- Funnels query by event_name in stage order. Cohorts use a "signup" event
-- (or whichever event_name is configured as the cohort anchor) to bucket
-- users, then count subsequent activity within retention windows.
CREATE TABLE IF NOT EXISTS events (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  event_name  TEXT NOT NULL,
  user_id     TEXT,
  occurred_at TEXT NOT NULL,
  properties  TEXT,
  value_cents INTEGER,
  segment_dim TEXT,
  segment_val TEXT
);
CREATE INDEX IF NOT EXISTS idx_events_name_time      ON events(event_name, occurred_at);
CREATE INDEX IF NOT EXISTS idx_events_user_time      ON events(user_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_events_segment_value  ON events(segment_dim, segment_val);

-- A/B experiment header: experiment_id is the natural key the analyst uses.
CREATE TABLE IF NOT EXISTS ab_experiments (
  experiment_id  TEXT PRIMARY KEY,
  name           TEXT NOT NULL,
  started_at     TEXT NOT NULL,
  ended_at       TEXT,
  primary_metric TEXT NOT NULL
);

-- A/B variant assignment per user. Sample size = COUNT(user_id) per variant.
CREATE TABLE IF NOT EXISTS ab_assignments (
  experiment_id TEXT NOT NULL,
  user_id       TEXT NOT NULL,
  variant       TEXT NOT NULL,
  assigned_at   TEXT NOT NULL,
  PRIMARY KEY (experiment_id, user_id)
);

-- A/B metric observations: primary + guardrail metrics, one row per user-metric.
-- value is REAL because conversions / ratios / latency live alongside counts.
CREATE TABLE IF NOT EXISTS ab_metric_observations (
  experiment_id TEXT NOT NULL,
  user_id       TEXT NOT NULL,
  metric_name   TEXT NOT NULL,
  metric_value  REAL NOT NULL,
  observed_at   TEXT NOT NULL,
  is_primary    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_abmo_lookup ON ab_metric_observations(experiment_id, metric_name);

-- Failures / incidents driving the actor_failure tool. recovered_at NULL =
-- still ongoing; MTTR is recovered_at - occurred_at when both present.
CREATE TABLE IF NOT EXISTS failures (
  id               INTEGER PRIMARY KEY AUTOINCREMENT,
  actor_type       TEXT NOT NULL,
  occurred_at      TEXT NOT NULL,
  recovered_at     TEXT,
  blast_radius     INTEGER NOT NULL DEFAULT 0,
  recovery_pattern TEXT
);
CREATE INDEX IF NOT EXISTS idx_failures_actor_time ON failures(actor_type, occurred_at);

-- Cost records: one row per service/component/region/account/tag bucket per
-- time grain (typically daily). cost_cents is INTEGER to avoid FP rounding.
CREATE TABLE IF NOT EXISTS cost_records (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  service     TEXT NOT NULL,
  component   TEXT,
  region      TEXT,
  account     TEXT,
  tag         TEXT,
  cost_cents  INTEGER NOT NULL,
  recorded_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cost_time ON cost_records(recorded_at);

-- SLO observations: one row per SLI per measurement timestamp. Aggregation
-- (mean / p99) happens at query time using the time_window argument.
CREATE TABLE IF NOT EXISTS slo_observations (
  sli         TEXT NOT NULL,
  observed_at TEXT NOT NULL,
  value       REAL NOT NULL,
  PRIMARY KEY (sli, observed_at)
);

-- Feedback corpus: NPS comments, CS tickets, app reviews, interview notes.
-- topic/sentiment columns are optional pre-classified labels; the adapter
-- can fall back to keyword clustering when they are NULL.
CREATE TABLE IF NOT EXISTS feedback_items (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  source      TEXT NOT NULL,
  occurred_at TEXT NOT NULL,
  body        TEXT NOT NULL,
  nps_score   INTEGER,
  topic       TEXT,
  sentiment   TEXT
);
CREATE INDEX IF NOT EXISTS idx_feedback_source_time ON feedback_items(source, occurred_at);
`

// Migrate applies Schema to the given *sql.DB. Idempotent — every statement
// uses CREATE IF NOT EXISTS, so calling Migrate at process start is safe.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, Schema); err != nil {
		return fmt.Errorf("analytics: apply schema: %w", err)
	}
	return nil
}
