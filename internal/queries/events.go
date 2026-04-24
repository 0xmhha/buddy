// events.go produces `buddy events`: a tail of the hook_events table for
// debugging, optionally followed in real time. Output is structured (one line
// per event, space-separated columns) — this surface is debug-style, not
// friend-tone (per v0.1-spec §6.3 and m4-plan Task 4). The only friend-tone
// touches are the start/end markers Follow writes to stderr.
//
// Like stats.go, every function here is pure: open DB read-only, query, shape
// rows. The CLI layer (cmd/buddy) owns process exit codes and IO routing.

package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/format"
)

// defaultEventsLimit is the row cap when EventsOptions.Limit is unset (0).
// Matches `tail`'s habit of printing the last N lines so the user gets useful
// output from a bare `buddy events` invocation.
const defaultEventsLimit = 20

// defaultPollInterval is how often Follow re-queries for new events when
// EventsOptions.PollInterval is unset. 1s matches the daemon poll cadence so
// follow lag stays predictable relative to ingestion.
const defaultPollInterval = time.Second

// followMaxConsecutiveErrors caps how many back-to-back poll errors Follow
// will tolerate before surfacing the error and aborting. Without a cap, a
// permanently broken DB (file deleted/corrupted/permissions revoked) spins
// forever at 1Hz spamming stderr — see m4-plan §Task 4 review fix.
// Five gives transient blips (daemon restart, brief lock contention) room to
// recover while still failing fast on permanent breakage.
const followMaxConsecutiveErrors = 5

// ErrInvalidLimit is the sentinel returned when EventsOptions.Limit is
// negative. Limit == 0 still means "use the default" (defaultEventsLimit) —
// consistent with the common UX expectation that 0 = unset on integer flags.
// Carries a friend-tone Korean message that the CLI layer surfaces verbatim.
var ErrInvalidLimit = errors.New("--limit 은 1 이상이어야 해.")

// EventsOptions configure RunEvents and Follow.
type EventsOptions struct {
	// DBPath is the absolute path to the buddy SQLite store. Empty means
	// db.DefaultPath() (i.e. ~/.buddy/buddy.db).
	DBPath string
	// HookFilter, when non-empty, restricts results to a single hook_name,
	// matched case-insensitively (same convention as stats.HookFilter).
	HookFilter string
	// Limit caps the row count of the initial fetch. Zero means
	// defaultEventsLimit (20). Negative returns ErrInvalidLimit at the
	// boundary (RunEvents / Follow entry).
	Limit int
	// Since, when > 0, restricts results to events with ts strictly greater
	// than this unix-ms value. Used by Follow to fetch only new events on
	// each poll tick.
	Since int64
	// PollInterval is the cadence at which Follow re-queries. Zero means
	// defaultPollInterval (1s). Exposed for fast-test paths.
	PollInterval time.Duration
	// Stderr is where Follow writes friend-tone start/end markers. Zero
	// means os.Stderr. Exposed so tests can capture markers without
	// touching the global stderr.
	Stderr io.Writer
}

// Event is one row in the events output. Empty SessionID/ToolName render as
// "-" placeholders so columns line up regardless of which fields a hook event
// carried.
type Event struct {
	ID         int64
	Ts         int64 // unix ms
	Event      string
	HookName   string
	DurationMs int64
	ExitCode   int
	SessionID  string
	ToolName   string
}

// EventsResult is the full snapshot ready for RenderLines. Events are in
// chronological order (oldest first) so the output reads top-to-bottom like
// `tail` — the SQL query orders DESC for the LIMIT then we reverse.
type EventsResult struct {
	Events []Event
}

// RunEvents validates options, opens the DB read-only, runs the tail query,
// and returns rows in chronological order.
func RunEvents(opts EventsOptions) (EventsResult, error) {
	limit, err := resolveLimit(opts.Limit)
	if err != nil {
		return EventsResult{}, err
	}

	conn, err := db.Open(db.Options{Path: opts.DBPath, ReadOnly: true})
	if err != nil {
		return EventsResult{}, fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	events, err := tailQuery(context.Background(), conn, opts.HookFilter, opts.Since, 0, limit)
	if err != nil {
		return EventsResult{}, fmt.Errorf("query hook_events: %w", err)
	}
	return EventsResult{Events: events}, nil
}

// resolveLimit applies the Limit boundary policy: negative is invalid (fail
// fast at the boundary), zero means "use the default," positive passes through.
// Centralised so RunEvents and Follow share the exact same rule.
func resolveLimit(raw int) (int, error) {
	if raw < 0 {
		return 0, ErrInvalidLimit
	}
	if raw == 0 {
		return defaultEventsLimit, nil
	}
	return raw, nil
}

// tailQuery builds and runs the tail SELECT against an existing conn. We do
// all filtering in SQL so LIMIT is meaningful: a Go-side filter after a LIMIT
// would silently drop matching rows when the limit is hit by non-matching ones
// first.
//
// Cursor semantics: (sinceTs, sinceID) is a lexicographic position. Rows with
// (ts, id) > (sinceTs, sinceID) are returned. This prevents Follow from
// dropping events that share a unix-ms with the last seen row — the daemon
// batches inserts so same-millisecond rows are realistic, and a strict
// `WHERE ts > ?` cursor would silently lose them.
//
// ORDER BY ts DESC, id DESC + LIMIT + reverse-in-Go gives us "the most recent
// N events in chronological order." Reversing 20 rows in Go is cheaper than a
// subquery. Using QueryContext lets ctx cancellation interrupt a long-running
// query (matters for Follow when the user Ctrl-Cs mid-tick).
func tailQuery(ctx context.Context, conn *sql.DB, hookFilter string, sinceTs, sinceID int64, limit int) ([]Event, error) {
	var (
		clauses []string
		args    []any
	)
	if hookFilter != "" {
		clauses = append(clauses, "LOWER(hook_name) = LOWER(?)")
		args = append(args, hookFilter)
	}
	if sinceTs > 0 {
		// Lexicographic cursor on (ts, id): strictly greater than the last
		// seen row. Equivalent to "(ts, id) > (sinceTs, sinceID)" but
		// expressed in standard SQL for SQLite portability.
		clauses = append(clauses, "(ts > ? OR (ts = ? AND id > ?))")
		args = append(args, sinceTs, sinceTs, sinceID)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	q := `
		SELECT id, ts, event, hook_name, duration_ms, exit_code,
		       COALESCE(session_id, '') AS session_id,
		       COALESCE(tool_name, '')  AS tool_name
		  FROM hook_events
		  ` + where + `
		 ORDER BY ts DESC, id DESC
		 LIMIT ?`
	args = append(args, limit)

	rs, err := conn.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	out := make([]Event, 0, limit)
	for rs.Next() {
		var e Event
		if err := rs.Scan(&e.ID, &e.Ts, &e.Event, &e.HookName,
			&e.DurationMs, &e.ExitCode, &e.SessionID, &e.ToolName); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	// Reverse to chronological order (oldest first) so the tail reads
	// top-to-bottom oldest-to-newest, matching the default `tail` behavior.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// RenderLines writes one structured line per event to w. No header, no
// friend-tone — this is a debug surface. Columns are separated by 2 spaces
// (not tabs) so the output stays grep-friendly. Empty SessionID/ToolName
// render as "-" so columns are visually present in every row.
//
// Format:
//
//	<RFC3339 ts UTC>  <hook>  <event>  <tool>  exit=<n>  dur=<human>  sess=<short>
func (r EventsResult) RenderLines(w io.Writer) {
	for _, e := range r.Events {
		ts := time.UnixMilli(e.Ts).UTC().Format(time.RFC3339)
		tool := e.ToolName
		if tool == "" {
			tool = "-"
		}
		sess := shortSession(e.SessionID)
		_, _ = fmt.Fprintf(w, "%s  %s  %s  %s  exit=%d  dur=%s  sess=%s\n",
			ts, e.HookName, e.Event, tool, e.ExitCode,
			format.Duration(e.DurationMs), sess)
	}
}

// shortSession returns the first 8 chars of id, or "-" when id is empty.
// 8 chars is enough to disambiguate sessions in a tail without dominating the
// line; matches the convention used by `git log --abbrev=8`.
func shortSession(id string) string {
	if id == "" {
		return "-"
	}
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// followTailFn is the indirection Follow uses for each tail query. Real code
// always points it at tailQuery; tests swap it (via export_test.go) to
// simulate persistent DB failures without needing to corrupt a real SQLite
// file mid-tick. Package-private and only mutated from tests.
var followTailFn = tailQuery

// Follow streams events to w until ctx is cancelled. On entry it writes a
// friend-tone start marker to opts.Stderr, fetches the initial tail (same as
// RunEvents), then polls every PollInterval for events newer than the last
// seen (ts, id) cursor. On ctx cancel it writes a friend-tone end marker and
// returns nil.
//
// Connection lifecycle: Follow opens the DB once at start and reuses the
// handle across every tick. The previous design called RunEvents per tick,
// which opened+closed a fresh handle ~3,600 times per hour — wasteful and
// noisy in /proc.
//
// Error handling: a single poll error is logged to stderr and the loop
// continues (transient blips like daemon restart shouldn't kill the tail).
// After followMaxConsecutiveErrors back-to-back failures we surface the last
// error and return — a permanently broken DB (file deleted, corrupted,
// permissions revoked) shouldn't spin forever.
//
// Polling pattern (vs sqlite update_hook or fts5 triggers): the daemon writes
// a few rows per second at most, and the user's terminal can only consume that
// fast anyway. A simple ticker keeps the implementation single-goroutine and
// race-free; ctx.Done() is the only stop signal we need.
func Follow(ctx context.Context, opts EventsOptions, w io.Writer) error {
	limit, err := resolveLimit(opts.Limit)
	if err != nil {
		return err
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	interval := opts.PollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}

	conn, err := db.Open(db.Options{Path: opts.DBPath, ReadOnly: true})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	_, _ = fmt.Fprintln(stderr, "(따라가는 중. Ctrl-C로 멈춰.)")
	// End marker on every exit path (including error returns) so the user
	// sees a clean close even when the loop aborts on persistent failure.
	defer func() { _, _ = fmt.Fprintln(stderr, "(끝.)") }()

	// Initial fetch: same shape as a bare `buddy events` invocation so the
	// entry behaviour matches one-shot mode exactly.
	initial, err := followTailFn(ctx, conn, opts.HookFilter, opts.Since, 0, limit)
	if err != nil {
		return fmt.Errorf("query hook_events: %w", err)
	}
	(EventsResult{Events: initial}).RenderLines(w)

	sinceTs, sinceID := opts.Since, int64(0)
	for _, e := range initial {
		if e.Ts > sinceTs || (e.Ts == sinceTs && e.ID > sinceID) {
			sinceTs, sinceID = e.Ts, e.ID
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	consecErrs := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			next, err := followTailFn(ctx, conn, opts.HookFilter, sinceTs, sinceID, limit)
			if err != nil {
				// ctx cancellation surfaces as a query error; treat it as
				// a normal shutdown rather than counting it toward the
				// abort threshold.
				if ctx.Err() != nil {
					return nil
				}
				consecErrs++
				_, _ = fmt.Fprintf(stderr, "buddy: events poll 실패 (%v)\n", err)
				if consecErrs >= followMaxConsecutiveErrors {
					return fmt.Errorf("events poll: %d consecutive failures: %w", consecErrs, err)
				}
				continue
			}
			consecErrs = 0
			(EventsResult{Events: next}).RenderLines(w)
			for _, e := range next {
				if e.Ts > sinceTs || (e.Ts == sinceTs && e.ID > sinceID) {
					sinceTs, sinceID = e.Ts, e.ID
				}
			}
		}
	}
}
