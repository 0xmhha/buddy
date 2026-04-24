// Package db owns the SQLite-backed buddy state.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DefaultPath returns the standard buddy DB location: ~/.buddy/buddy.db.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".buddy", "buddy.db"), nil
}

// Options configure how the DB is opened.
type Options struct {
	Path     string
	ReadOnly bool
}

// Open prepares the buddy DB. On a writable open, schema migrations run
// to completion before the handle is returned. WAL mode is always enabled.
func Open(opts Options) (*sql.DB, error) {
	if opts.Path == "" {
		p, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		opts.Path = p
	}
	// Read-only opens never create the parent directory: that would be a write
	// side effect on a "read-only" code path, and read-driven callers (doctor)
	// rely on a missing-DB open failing loudly so they can collapse it into a
	// single KindDBOpen issue.
	if !opts.ReadOnly {
		if err := os.MkdirAll(filepath.Dir(opts.Path), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", filepath.Dir(opts.Path), err)
		}
	}

	dsn := opts.Path + "?_pragma=foreign_keys(1)"
	if opts.ReadOnly {
		dsn += "&mode=ro"
	}

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", opts.Path, err)
	}
	// modernc.org/sqlite supports a single connection well; cap the pool
	// so PRAGMA journal_mode = WAL applies consistently and writes serialize.
	conn.SetMaxOpenConns(1)

	if !opts.ReadOnly {
		if _, err := conn.Exec("PRAGMA journal_mode = WAL"); err != nil {
			conn.Close()
			return nil, fmt.Errorf("set WAL: %w", err)
		}
		if _, err := conn.Exec("PRAGMA synchronous = NORMAL"); err != nil {
			conn.Close()
			return nil, fmt.Errorf("set synchronous: %w", err)
		}
		if err := RunMigrations(conn); err != nil {
			conn.Close()
			return nil, fmt.Errorf("migrations: %w", err)
		}
	}
	return conn, nil
}
