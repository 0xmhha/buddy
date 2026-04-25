// Package db owns the SQLite-backed buddy state.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// ErrDBMissing means the requested DB file (or its parent directory) does not
// exist on a read-only open. Read-only opens never create the file — without
// this sentinel, modernc.org/sqlite would surface SQLite's cryptic
// "out of memory (14)" (missing parent dir) or, after lazily creating an empty
// file in mode=ro, "no such table: hook_outbox" on first query. Callers
// (doctor, stats, events) translate this into a friend-tone Korean message.
var ErrDBMissing = errors.New("buddy DB missing")

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
	// single KindDBOpen issue. Stat first so we can return a clean sentinel
	// (ErrDBMissing) instead of letting modernc.org/sqlite surface either
	// "out of memory (14)" (parent dir gone) or "no such table" (file lazily
	// created empty under mode=ro). One Stat covers both cases — fs.ErrNotExist
	// triggers for missing leaf or any missing ancestor.
	if opts.ReadOnly {
		if _, err := os.Stat(opts.Path); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("%s: %w", opts.Path, ErrDBMissing)
			}
			return nil, fmt.Errorf("stat %s: %w", opts.Path, err)
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(opts.Path), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", filepath.Dir(opts.Path), err)
		}
	}

	// busy_timeout makes concurrent opens (e.g. daemon writer + doctor reader)
	// wait for the WAL/header lock instead of failing with "database is locked".
	dsn := opts.Path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
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
