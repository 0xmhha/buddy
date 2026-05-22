package advisor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/0xmhha/buddy/internal/db"
)

// ErrNotFound mirrors the sibling stores so callers can route around
// a missing row distinct from a real DB error.
var ErrNotFound = errors.New("advisor: not found")

// Store wraps the advisories table. Read-heavy + append-mostly —
// CLI / TUI / MCP read; daemon inserts; mute updates muted flag.
type Store struct {
	db db.Conn
}

// NewStore wraps any db.Conn implementation. *sql.DB satisfies the
// interface natively, so existing call sites keep working unchanged.
func NewStore(conn db.Conn) *Store { return &Store{db: conn} }

// Insert appends one advisory. CreatedAt zero → time.Now().UTC().
// Returns the autoincrement id.
func (s *Store) Insert(ctx context.Context, a Advisory) (int64, error) {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	evJSON, err := json.Marshal(a.Evidence)
	if err != nil {
		return 0, fmt.Errorf("advisor: insert: encode evidence: %w", err)
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO advisories (kind, severity, message, evidence_json, created_at, muted)
		VALUES (?, ?, ?, ?, ?, ?)`,
		a.Kind, string(a.Severity), a.Message, string(evJSON),
		a.CreatedAt.UTC().UnixMilli(), boolToInt(a.Muted),
	)
	if err != nil {
		return 0, fmt.Errorf("advisor: insert: %w", err)
	}
	return res.LastInsertId()
}

// Get returns one row by id.
func (s *Store) Get(ctx context.Context, id int64) (Advisory, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, kind, severity, message, evidence_json, created_at, muted
		FROM advisories WHERE id = ?`, id)
	a, err := scanAdvisory(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Advisory{}, ErrNotFound
	}
	return a, err
}

// ListOptions tunes List. Zero-value returns every non-muted advisory
// ordered by created_at DESC.
type ListOptions struct {
	IncludeMuted bool
	Since        time.Time
	Kinds        []string // filter to these kinds; empty = all
}

// List returns advisories matching opts.
func (s *Store) List(ctx context.Context, opts ListOptions) ([]Advisory, error) {
	q := `SELECT id, kind, severity, message, evidence_json, created_at, muted
	      FROM advisories`
	var (
		conds []string
		args  []any
	)
	if !opts.IncludeMuted {
		conds = append(conds, "muted = 0")
	}
	if !opts.Since.IsZero() {
		conds = append(conds, "created_at >= ?")
		args = append(args, opts.Since.UTC().UnixMilli())
	}
	if len(opts.Kinds) > 0 {
		placeholders := ""
		for i := range opts.Kinds {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, opts.Kinds[i])
		}
		conds = append(conds, "kind IN ("+placeholders+")")
	}
	if len(conds) > 0 {
		q += " WHERE " + joinAND(conds)
	}
	q += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("advisor: list: %w", err)
	}
	defer rows.Close()

	var out []Advisory
	for rows.Next() {
		a, err := scanAdvisoryRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// LastInWindow returns the most recent advisory of `kind` within the
// dedup window (now - window, now]. Used by the evaluator to skip
// re-firing the same advisory inside its window.
//
// Returns ErrNotFound when no row matches.
func (s *Store) LastInWindow(ctx context.Context, kind string, window time.Duration, now time.Time) (Advisory, error) {
	if window <= 0 {
		return Advisory{}, ErrNotFound
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id, kind, severity, message, evidence_json, created_at, muted
		FROM advisories
		WHERE kind = ? AND created_at >= ? AND muted = 0
		ORDER BY created_at DESC
		LIMIT 1`,
		kind, now.Add(-window).UTC().UnixMilli(),
	)
	a, err := scanAdvisory(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Advisory{}, ErrNotFound
	}
	return a, err
}

// Mute flips the muted flag on a single advisory.
func (s *Store) Mute(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `UPDATE advisories SET muted = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("advisor: mute: %w", err)
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

// MuteKind mutes every active row of the given kind. Future UX:
// `buddy advise mute <kind>` — schema already supports it; CLI wiring
// is a v0.11.x follow-on.
func (s *Store) MuteKind(ctx context.Context, kind string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE advisories SET muted = 1 WHERE kind = ? AND muted = 0`, kind)
	if err != nil {
		return 0, fmt.Errorf("advisor: mute kind: %w", err)
	}
	return res.RowsAffected()
}

// DeleteOlderThan trims the advisories table. Returns row count
// deleted. Mirrors `buddy session purge` semantics — provides the
// hook for a future `buddy advise purge` CLI.
func (s *Store) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM advisories WHERE created_at < ?`,
		cutoff.UTC().UnixMilli())
	if err != nil {
		return 0, fmt.Errorf("advisor: prune: %w", err)
	}
	return res.RowsAffected()
}

// Count returns the total row count (incl. muted).
func (s *Store) Count(ctx context.Context) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM advisories`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("advisor: count: %w", err)
	}
	return n, nil
}

// ─── helpers ─────────────────────────────────────────────────────────

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func joinAND(conds []string) string {
	if len(conds) == 0 {
		return ""
	}
	out := conds[0]
	for _, c := range conds[1:] {
		out += " AND " + c
	}
	return out
}

func scanAdvisory(row *sql.Row) (Advisory, error) {
	var (
		a       Advisory
		sev     string
		evJSON  string
		created int64
		muted   int
	)
	err := row.Scan(&a.ID, &a.Kind, &sev, &a.Message, &evJSON, &created, &muted)
	if err != nil {
		return Advisory{}, err
	}
	a.Severity = Severity(sev)
	a.CreatedAt = time.UnixMilli(created).UTC()
	a.Muted = muted == 1
	if err := json.Unmarshal([]byte(evJSON), &a.Evidence); err != nil {
		// Fall back to empty rather than fail — evidence corruption
		// shouldn't strand the advisory text.
		a.Evidence = nil
	}
	return a, nil
}

func scanAdvisoryRows(rows *sql.Rows) (Advisory, error) {
	var (
		a       Advisory
		sev     string
		evJSON  string
		created int64
		muted   int
	)
	err := rows.Scan(&a.ID, &a.Kind, &sev, &a.Message, &evJSON, &created, &muted)
	if err != nil {
		return Advisory{}, err
	}
	a.Severity = Severity(sev)
	a.CreatedAt = time.UnixMilli(created).UTC()
	a.Muted = muted == 1
	if err := json.Unmarshal([]byte(evJSON), &a.Evidence); err != nil {
		a.Evidence = nil
	}
	return a, nil
}
