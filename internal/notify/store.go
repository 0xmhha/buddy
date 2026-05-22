package notify

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xmhha/buddy/internal/db"
)

// Store wraps the notification_log table. Append-only on the daemon
// side, read on CLI / TUI / MCP. No update or mute path — once a
// dispatch attempt is logged, its row is immutable.
type Store struct {
	db db.Conn
}

// NewStore wraps any db.Conn implementation.
func NewStore(conn db.Conn) *Store { return &Store{db: conn} }

// Insert appends a single notification_log row. SentAt zero → now().
func (s *Store) Insert(ctx context.Context, row LogRow) (int64, error) {
	if row.SentAt.IsZero() {
		row.SentAt = time.Now().UTC()
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_log (advisory_id, channel, kind, severity, sent_at, outcome, detail)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.AdvisoryID, row.Channel, row.Kind, string(row.Severity),
		row.SentAt.UTC().UnixMilli(), row.Outcome, row.Detail,
	)
	if err != nil {
		return 0, fmt.Errorf("notify: insert log: %w", err)
	}
	return res.LastInsertId()
}

// ListOptions tunes List. Zero-value returns every row newest-first.
type ListOptions struct {
	Channel string    // empty = all channels
	Kinds   []string  // empty = all kinds
	Since   time.Time // zero = all time
	Limit   int       // <=0 = no limit
}

// List returns log rows matching opts, ordered sent_at DESC.
func (s *Store) List(ctx context.Context, opts ListOptions) ([]LogRow, error) {
	q := `SELECT id, advisory_id, channel, kind, severity, sent_at, outcome, detail
	      FROM notification_log`
	var (
		conds []string
		args  []any
	)
	if opts.Channel != "" {
		conds = append(conds, "channel = ?")
		args = append(args, opts.Channel)
	}
	if !opts.Since.IsZero() {
		conds = append(conds, "sent_at >= ?")
		args = append(args, opts.Since.UTC().UnixMilli())
	}
	if len(opts.Kinds) > 0 {
		ph := ""
		for i, k := range opts.Kinds {
			if i > 0 {
				ph += ","
			}
			ph += "?"
			args = append(args, k)
		}
		conds = append(conds, "kind IN ("+ph+")")
	}
	if len(conds) > 0 {
		q += " WHERE " + joinAND(conds)
	}
	q += " ORDER BY sent_at DESC"
	if opts.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("notify: list log: %w", err)
	}
	defer rows.Close()
	var out []LogRow
	for rows.Next() {
		r, err := scanLogRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// LastSent returns the most recent successful "sent" row for the
// given (channel, kind) pair. Used by the dispatcher to decide the
// per-channel dedup-window skip. Returns sql.ErrNoRows when nothing
// has been sent yet (callers check with errors.Is).
func (s *Store) LastSent(ctx context.Context, channel, kind string) (LogRow, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, advisory_id, channel, kind, severity, sent_at, outcome, detail
		FROM notification_log
		WHERE channel = ? AND kind = ? AND outcome = 'sent'
		ORDER BY sent_at DESC
		LIMIT 1`,
		channel, kind,
	)
	r, err := scanLogRowSingle(row)
	return r, err
}

// Count returns total row count (for stats / debugging).
func (s *Store) Count(ctx context.Context) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notification_log`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("notify: count log: %w", err)
	}
	return n, nil
}

// DeleteOlderThan trims log rows. Mirrors the buddy session purge
// shape — analogous `buddy notify purge` could ride on this.
func (s *Store) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM notification_log WHERE sent_at < ?`,
		cutoff.UTC().UnixMilli())
	if err != nil {
		return 0, fmt.Errorf("notify: delete log: %w", err)
	}
	return res.RowsAffected()
}

// ─── helpers ─────────────────────────────────────────────────────────

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

func scanLogRow(rows *sql.Rows) (LogRow, error) {
	var (
		r        LogRow
		severity string
		sentAt   int64
	)
	err := rows.Scan(&r.ID, &r.AdvisoryID, &r.Channel, &r.Kind, &severity, &sentAt, &r.Outcome, &r.Detail)
	if err != nil {
		return LogRow{}, err
	}
	r.Severity = Severity(severity)
	r.SentAt = time.UnixMilli(sentAt).UTC()
	return r, nil
}

func scanLogRowSingle(row *sql.Row) (LogRow, error) {
	var (
		r        LogRow
		severity string
		sentAt   int64
	)
	err := row.Scan(&r.ID, &r.AdvisoryID, &r.Channel, &r.Kind, &severity, &sentAt, &r.Outcome, &r.Detail)
	if err != nil {
		return LogRow{}, err
	}
	r.Severity = Severity(severity)
	r.SentAt = time.UnixMilli(sentAt).UTC()
	return r, nil
}
