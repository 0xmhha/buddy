package db

import (
	"context"
	"database/sql"
)

// Conn is the narrow contract the store packages need from *sql.DB.
// Stores accept this interface instead of *sql.DB directly so tests
// can substitute an in-memory fake without spinning up the production
// SQLite driver, and a future swap to a different backend can implement
// the same three methods without touching every store constructor.
//
// *sql.DB satisfies this interface natively — every existing call site
// keeps working unchanged when stores migrate to Conn. The interface is
// intentionally minimal (only the three methods every store actually
// uses) so an implementation never has to fake more than it needs.
//
// Transactions remain explicit: stores that need them call BeginTx on
// the concrete *sql.DB before delegating to helpers, so the interface
// stays free of *sql.Tx return types.
type Conn interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
