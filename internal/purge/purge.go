// Package purge implements the buddy retention sweep: delete old hook_events
// and hook_stats rows.
//
// CRITICAL invariant (v0.1-spec §4 invariant 1, roadmap §M5 T4):
//
//	hook_outbox는 절대 건드리지 않는다.
//
// Outbox is the sync-write WAL the hook wrapper relies on. The daemon drains
// it. If purge truncated outbox mid-write, hook events would be silently lost.
// This package therefore intentionally contains NO SQL that references
// hook_outbox — search for "outbox" in this file and you will only find this
// comment. That is structural enforcement of the invariant: a future
// contributor adding `WHERE ... outbox ...` here must first delete this
// note, which should make the regression obvious in review.
package purge

import (
	"database/sql"
	"fmt"
)

// Result reports how many rows would be (DryRun) or were (Apply) deleted per
// table. Outbox is intentionally absent.
type Result struct {
	Events int64 // rows in hook_events with ts < cutoff
	Stats  int64 // rows in hook_stats with ts_bucket < cutoff
}

// Options configure a Run.
type Options struct {
	// BeforeMillis is the unix-millisecond cutoff. Rows with ts (events) /
	// ts_bucket (stats) strictly less than this are eligible for deletion.
	// The boundary is `< BeforeMillis`, never `<=`: a row at exactly the
	// cutoff is preserved.
	BeforeMillis int64
	// DryRun, when true, runs SELECT COUNT(*) per table and returns without
	// any DELETE. The default-safe path the CLI uses unless the user passes
	// --apply.
	DryRun bool
}

// Run computes (and optionally executes) the purge.
//
// On DryRun=true: SELECT COUNT(*) per table; no row is deleted; hook_outbox
// is never queried.
//
// On DryRun=false: both DELETE statements run inside a single transaction.
// On any error (including a transient SQLite lock), the transaction rolls
// back so the DB is left untouched. hook_outbox is never read or mutated.
func Run(conn *sql.DB, opts Options) (Result, error) {
	if opts.DryRun {
		return countOnly(conn, opts.BeforeMillis)
	}
	return deleteInTx(conn, opts.BeforeMillis)
}

func countOnly(conn *sql.DB, before int64) (Result, error) {
	var res Result
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM hook_events WHERE ts < ?", before,
	).Scan(&res.Events); err != nil {
		return Result{}, fmt.Errorf("count events: %w", err)
	}
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM hook_stats WHERE ts_bucket < ?", before,
	).Scan(&res.Stats); err != nil {
		return Result{}, fmt.Errorf("count stats: %w", err)
	}
	return res, nil
}

func deleteInTx(conn *sql.DB, before int64) (Result, error) {
	tx, err := conn.Begin()
	if err != nil {
		return Result{}, fmt.Errorf("begin: %w", err)
	}
	// Safe no-op after a successful Commit (sql.Tx documents Rollback as
	// idempotent post-Commit). Guarantees the DB is untouched on any
	// pre-Commit error path.
	defer func() { _ = tx.Rollback() }()

	eventsRes, err := tx.Exec("DELETE FROM hook_events WHERE ts < ?", before)
	if err != nil {
		return Result{}, fmt.Errorf("delete events: %w", err)
	}
	statsRes, err := tx.Exec("DELETE FROM hook_stats WHERE ts_bucket < ?", before)
	if err != nil {
		return Result{}, fmt.Errorf("delete stats: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit: %w", err)
	}

	var out Result
	out.Events, _ = eventsRes.RowsAffected()
	out.Stats, _ = statsRes.RowsAffected()
	return out, nil
}
