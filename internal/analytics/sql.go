package analytics

import (
	"database/sql"
	"time"
)
// SQLAdapter is the SQLite-flavoured reference implementation of Adapter.
// PostgreSQL / MySQL adapters land alongside this one in spec §8 follow-up
// cycles; they reuse the same query shape and only need driver-specific
// placeholder / autoincrement substitution.
//
// SQLAdapter is safe for concurrent use because *sql.DB itself is — every
// method uses ctx-aware ExecContext / QueryContext and never holds state
// between calls.
type SQLAdapter struct {
	db *sql.DB
}

// NewSQLAdapter wraps an open *sql.DB. The caller still owns lifecycle —
// closing the DB closes the adapter.
func NewSQLAdapter(db *sql.DB) *SQLAdapter {
	return &SQLAdapter{db: db}
}

// ─── helpers ───────────────────────────────────────────────────────────────

func iso(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// rangeArgs returns (from, to) ISO timestamps; zero values become open ends.
func rangeArgs(r TimeRange) (string, string) {
	from := "1970-01-01T00:00:00Z"
	to := "9999-12-31T23:59:59Z"
	if !r.From.IsZero() {
		from = iso(r.From)
	}
	if !r.To.IsZero() {
		to = iso(r.To)
	}
	return from, to
}

func parseISO(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// segmentClause appends a "(segment_dim = ? AND segment_val = ?)" filter
// to the given clauses + args slice when seg is non-nil.
func segmentClause(seg *Segment, clauses *[]string, args *[]any) {
	if seg == nil || seg.Dimension == "" {
		return
	}
	*clauses = append(*clauses, "segment_dim = ? AND segment_val = ?")
	*args = append(*args, seg.Dimension, seg.Value)
}

