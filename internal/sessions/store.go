package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/0xmhha/buddy/internal/schema"
)

// ErrNotFound mirrors the agent / feature packages' sentinel so callers can
// distinguish "session doesn't exist" from "DB had a real problem".
var ErrNotFound = errors.New("session: not found")

// Store wraps the SQLite sessions table. Methods take a *sql.DB so callers
// share the same connection pool used elsewhere in the binary.
type Store struct {
	db *sql.DB
}

// NewStore wraps an open connection.
func NewStore(conn *sql.DB) *Store {
	return &Store{db: conn}
}

// Options is the connection-open shape — mirrors the agent / feature
// packages so the "open store" pattern stays consistent across packages.
type Options struct {
	DBPath string
}

// Upsert inserts the session or updates every field except StartedAt (which
// is set on first observation and preserved across resumes). Used by
// fsLister on each periodic tail pass and by the SessionStart hook.
//
// Concurrency: SQLite ON CONFLICT preserves StartedAt; LastActive /
// LastOffset / Usage / EndedAt / GoalText / Metadata always reflect the
// latest known state. EndedAt=nil clears any prior end marker (session
// resumed).
func (s *Store) Upsert(ctx context.Context, sess Session) error {
	const q = `
		INSERT INTO sessions (
			id, pid, transcript_path, started_at, last_active,
			total_input_tokens, total_output_tokens,
			total_cache_read, total_cache_create,
			last_offset, ended_at, goal_text, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			pid                  = excluded.pid,
			transcript_path      = excluded.transcript_path,
			last_active          = excluded.last_active,
			total_input_tokens   = excluded.total_input_tokens,
			total_output_tokens  = excluded.total_output_tokens,
			total_cache_read     = excluded.total_cache_read,
			total_cache_create   = excluded.total_cache_create,
			last_offset          = excluded.last_offset,
			ended_at             = excluded.ended_at,
			goal_text            = CASE WHEN sessions.goal_text = '' THEN excluded.goal_text ELSE sessions.goal_text END,
			metadata             = excluded.metadata`
	endedAt := nullableTime(sess.EndedAt)
	_, err := s.db.ExecContext(ctx, q,
		sess.ID, sess.PID, sess.TranscriptPath,
		sess.StartedAt.UTC().UnixMilli(),
		sess.LastActive.UTC().UnixMilli(),
		sess.Usage.InputTokens, sess.Usage.OutputTokens,
		sess.Usage.CacheReadTokens, sess.Usage.CacheCreateTokens,
		sess.LastOffset, endedAt, sess.GoalText, sess.Metadata,
	)
	if err != nil {
		return fmt.Errorf("session: upsert: %w", err)
	}
	return nil
}

// Get returns one session row or ErrNotFound. Used by `buddy session show`.
func (s *Store) Get(ctx context.Context, id string) (Session, error) {
	const q = `
		SELECT id, pid, transcript_path, started_at, last_active,
		       total_input_tokens, total_output_tokens,
		       total_cache_read, total_cache_create,
		       last_offset, ended_at, goal_text, metadata
		FROM sessions WHERE id = ?`
	row := s.db.QueryRowContext(ctx, q, id)
	sess, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("session: get: %w", err)
	}
	return sess, nil
}

// ListOptions tunes List. Zero-value returns all sessions (active +
// ended) ordered by last_active DESC.
type ListOptions struct {
	// IncludeEnded controls whether sessions with ended_at IS NOT NULL
	// are returned. Default false — `buddy session list` without --all
	// shows only currently-active sessions.
	IncludeEnded bool
	// Since, when non-zero, filters to sessions whose last_active is
	// after this time. Matches `buddy session list --since 24h` shape.
	Since time.Time
}

// List returns sessions matching opts, sorted last_active DESC.
func (s *Store) List(ctx context.Context, opts ListOptions) ([]Session, error) {
	q := `
		SELECT id, pid, transcript_path, started_at, last_active,
		       total_input_tokens, total_output_tokens,
		       total_cache_read, total_cache_create,
		       last_offset, ended_at, goal_text, metadata
		FROM sessions`
	var (
		conds []string
		args  []any
	)
	if !opts.IncludeEnded {
		conds = append(conds, "ended_at IS NULL")
	}
	if !opts.Since.IsZero() {
		conds = append(conds, "last_active >= ?")
		args = append(args, opts.Since.UTC().UnixMilli())
	}
	if len(conds) > 0 {
		q += " WHERE " + joinAND(conds)
	}
	q += " ORDER BY last_active DESC"
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("session: list: %w", err)
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		sess, err := scanSessionRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

// SetEndedAt marks a session as ended (or, with t=nil, clears the marker
// on resume). Used by the daemon's sessionMonitor goroutine when
// last_active crosses EndedThreshold, and by the same goroutine to clear
// the marker when transcript activity resumes.
func (s *Store) SetEndedAt(ctx context.Context, id string, t *time.Time) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET ended_at = ? WHERE id = ?`,
		nullableTime(t), id)
	if err != nil {
		return fmt.Errorf("session: set ended_at: %w", err)
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

// Delete removes one session row. `buddy session purge --before` calls
// this for each row past the cutoff.
func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("session: delete: %w", err)
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

// PurgeBefore deletes every ENDED session whose last_active is older than
// threshold. Active sessions (ended_at IS NULL) are never purged regardless
// of cutoff — mirrors `buddy agent purge` semantics. Returns the row count
// actually deleted.
func (s *Store) PurgeBefore(ctx context.Context, threshold time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE ended_at IS NOT NULL AND last_active < ?`,
		threshold.UTC().UnixMilli())
	if err != nil {
		return 0, fmt.Errorf("session: purge: %w", err)
	}
	return res.RowsAffected()
}

// CountBefore is the dry-run preview for PurgeBefore. Same WHERE clause;
// returns the count that PurgeBefore would delete.
func (s *Store) CountBefore(ctx context.Context, threshold time.Time) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sessions
		WHERE ended_at IS NOT NULL AND last_active < ?`,
		threshold.UTC().UnixMilli()).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("session: count: %w", err)
	}
	return n, nil
}

// ─── helpers ────────────────────────────────────────────────────────────

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().UnixMilli()
}

func scanSession(row *sql.Row) (Session, error) {
	var (
		sess           Session
		startedAt      int64
		lastActive     int64
		endedAt        sql.NullInt64
	)
	err := row.Scan(
		&sess.ID, &sess.PID, &sess.TranscriptPath,
		&startedAt, &lastActive,
		&sess.Usage.InputTokens, &sess.Usage.OutputTokens,
		&sess.Usage.CacheReadTokens, &sess.Usage.CacheCreateTokens,
		&sess.LastOffset, &endedAt, &sess.GoalText, &sess.Metadata,
	)
	if err != nil {
		return Session{}, err
	}
	sess.StartedAt = time.UnixMilli(startedAt).UTC()
	sess.LastActive = time.UnixMilli(lastActive).UTC()
	if endedAt.Valid {
		t := time.UnixMilli(endedAt.Int64).UTC()
		sess.EndedAt = &t
	}
	return sess, nil
}

func scanSessionRows(rows *sql.Rows) (Session, error) {
	var (
		sess           Session
		startedAt      int64
		lastActive     int64
		endedAt        sql.NullInt64
	)
	err := rows.Scan(
		&sess.ID, &sess.PID, &sess.TranscriptPath,
		&startedAt, &lastActive,
		&sess.Usage.InputTokens, &sess.Usage.OutputTokens,
		&sess.Usage.CacheReadTokens, &sess.Usage.CacheCreateTokens,
		&sess.LastOffset, &endedAt, &sess.GoalText, &sess.Metadata,
	)
	if err != nil {
		return Session{}, err
	}
	sess.StartedAt = time.UnixMilli(startedAt).UTC()
	sess.LastActive = time.UnixMilli(lastActive).UTC()
	if endedAt.Valid {
		t := time.UnixMilli(endedAt.Int64).UTC()
		sess.EndedAt = &t
	}
	return sess, nil
}

func joinAND(conds []string) string {
	switch len(conds) {
	case 0:
		return ""
	case 1:
		return conds[0]
	default:
		out := conds[0]
		for _, c := range conds[1:] {
			out += " AND " + c
		}
		return out
	}
}

// Ensure unused import doesn't break.
var _ = schema.TokenUsage{}
