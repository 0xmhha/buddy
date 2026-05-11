package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound mirrors the analytics package's sentinel so CLI callers can
// give a "no such agent" message uniformly.
var ErrNotFound = errors.New("agent: not found")

// Store is the SQLite-backed persistence layer. Callers wrap an open *sql.DB
// (the same one buddy uses for hook events / sessions / features). v4 migration
// adds the agents / agent_runs / agent_logs tables.
type Store struct {
	db *sql.DB
}

// NewStore wraps an open DB. The DB is not owned — caller closes it.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// ─── agent CRUD ────────────────────────────────────────────────────────────

// Create inserts a new agent row. spec.ID becomes the primary key, so calling
// Create twice with the same ID is an error (use Upsert if you want replace).
func (s *Store) Create(ctx context.Context, spec AgentSpec, specYAML string) (Agent, error) {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO agents (id, name, spec_yaml, schedule, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		spec.ID, spec.Name, specYAML, spec.Schedule, string(StatusIdle),
		now.UnixMilli(), now.UnixMilli(),
	)
	if err != nil {
		return Agent{}, fmt.Errorf("agent: create: %w", err)
	}
	return Agent{
		ID: spec.ID, Name: spec.Name, SpecYAML: specYAML, Schedule: spec.Schedule,
		Status: StatusIdle, CreatedAt: now, UpdatedAt: now,
	}, nil
}

// Get fetches one agent by ID. Returns ErrNotFound when no row matches.
func (s *Store) Get(ctx context.Context, id string) (Agent, error) {
	var (
		a           Agent
		createdMs   int64
		updatedMs   int64
		lastRunMs   sql.NullInt64
		statusText  string
	)
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, spec_yaml, schedule, status, created_at, updated_at, last_run_at
		FROM agents WHERE id = ?`, id,
	).Scan(&a.ID, &a.Name, &a.SpecYAML, &a.Schedule, &statusText, &createdMs, &updatedMs, &lastRunMs)
	if errors.Is(err, sql.ErrNoRows) {
		return Agent{}, ErrNotFound
	}
	if err != nil {
		return Agent{}, fmt.Errorf("agent: get: %w", err)
	}
	a.Status = Status(statusText)
	a.CreatedAt = time.UnixMilli(createdMs).UTC()
	a.UpdatedAt = time.UnixMilli(updatedMs).UTC()
	if lastRunMs.Valid {
		t := time.UnixMilli(lastRunMs.Int64).UTC()
		a.LastRunAt = &t
	}
	return a, nil
}

// List returns all agents ordered by updated_at descending. Small enough to
// not need pagination in v0.3.
func (s *Store) List(ctx context.Context) ([]Agent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, spec_yaml, schedule, status, created_at, updated_at, last_run_at
		FROM agents ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("agent: list: %w", err)
	}
	defer rows.Close()

	var out []Agent
	for rows.Next() {
		var (
			a          Agent
			createdMs  int64
			updatedMs  int64
			lastRunMs  sql.NullInt64
			statusText string
		)
		if err := rows.Scan(&a.ID, &a.Name, &a.SpecYAML, &a.Schedule, &statusText, &createdMs, &updatedMs, &lastRunMs); err != nil {
			return nil, fmt.Errorf("agent: list scan: %w", err)
		}
		a.Status = Status(statusText)
		a.CreatedAt = time.UnixMilli(createdMs).UTC()
		a.UpdatedAt = time.UnixMilli(updatedMs).UTC()
		if lastRunMs.Valid {
			t := time.UnixMilli(lastRunMs.Int64).UTC()
			a.LastRunAt = &t
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Delete removes an agent and (via FK cascade) all its runs + logs.
func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM agents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("agent: delete: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateStatus transitions an agent's status. last_run_at is bumped to now
// when transitioning to/from StatusRunning so list views show recency.
func (s *Store) UpdateStatus(ctx context.Context, id string, status Status) error {
	now := time.Now().UTC().UnixMilli()
	_, err := s.db.ExecContext(ctx, `
		UPDATE agents
		SET status = ?, updated_at = ?, last_run_at = ?
		WHERE id = ?`,
		string(status), now, now, id,
	)
	if err != nil {
		return fmt.Errorf("agent: update status: %w", err)
	}
	return nil
}

// ─── run + log ────────────────────────────────────────────────────────────

// StartRun inserts a new agent_runs row with ended_at=NULL and returns its ID.
func (s *Store) StartRun(ctx context.Context, agentID string) (int64, error) {
	now := time.Now().UTC().UnixMilli()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO agent_runs (agent_id, started_at) VALUES (?, ?)`,
		agentID, now,
	)
	if err != nil {
		return 0, fmt.Errorf("agent: start run: %w", err)
	}
	return res.LastInsertId()
}

// FinishRun closes out a run with exit code, optional error string, and the
// aggregated result as JSON. The result is whatever the Runtime decided to
// surface (typically a list of StepResult).
func (s *Store) FinishRun(ctx context.Context, runID int64, exitCode int, runErr error, result any) error {
	now := time.Now().UTC().UnixMilli()
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("agent: marshal run result: %w", err)
	}
	errText := ""
	if runErr != nil {
		errText = runErr.Error()
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE agent_runs
		SET ended_at = ?, exit_code = ?, error = ?, result_json = ?
		WHERE id = ?`,
		now, exitCode, nullableString(errText), string(resultBytes), runID,
	)
	if err != nil {
		return fmt.Errorf("agent: finish run: %w", err)
	}
	return nil
}

// AppendLog writes one log line.
func (s *Store) AppendLog(ctx context.Context, runID int64, level, message string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO agent_logs (run_id, ts, level, message) VALUES (?, ?, ?, ?)`,
		runID, time.Now().UTC().UnixMilli(), level, message,
	)
	if err != nil {
		return fmt.Errorf("agent: append log: %w", err)
	}
	return nil
}

// Logs returns up to `limit` log lines for a run, oldest first. Set limit=0
// to fetch all.
func (s *Store) Logs(ctx context.Context, runID int64, limit int) ([]AgentLog, error) {
	q := `SELECT id, run_id, ts, level, message FROM agent_logs WHERE run_id = ? ORDER BY ts ASC`
	args := []any{runID}
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("agent: logs: %w", err)
	}
	defer rows.Close()
	var out []AgentLog
	for rows.Next() {
		var l AgentLog
		var ts int64
		if err := rows.Scan(&l.ID, &l.RunID, &ts, &l.Level, &l.Message); err != nil {
			return nil, err
		}
		l.Ts = time.UnixMilli(ts).UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}

// nullableString turns empty strings into sql.NullString{Valid:false} so the
// agent_runs.error column ends up NULL rather than the literal "".
func nullableString(s string) any {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{Valid: true, String: s}
}
