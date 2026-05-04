package feature

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wm-it-22-00661/buddy/internal/db"
)

// Options configures store operations.
type Options struct {
	DBPath string // empty → db.DefaultPath()
}

// Upsert inserts or replaces a feature, setting updated_at to now.
func Upsert(ctx context.Context, opts Options, f Feature) error {
	conn, err := db.Open(db.Options{Path: opts.DBPath})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()
	return upsertConn(ctx, conn, f)
}

func upsertConn(ctx context.Context, conn *sql.DB, f Feature) error {
	actors, err := json.Marshal(f.Actors)
	if err != nil {
		return fmt.Errorf("marshal actors: %w", err)
	}
	ac, err := json.Marshal(f.AcceptanceCriteria)
	if err != nil {
		return fmt.Errorf("marshal acceptance_criteria: %w", err)
	}
	tp, err := json.Marshal(f.TestPlan)
	if err != nil {
		return fmt.Errorf("marshal test_plan: %w", err)
	}
	now := time.Now().UnixMilli()
	_, err = conn.ExecContext(ctx, `
		INSERT INTO features
			(feature_id, name, summary, actors, acceptance_criteria, test_plan, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(feature_id) DO UPDATE SET
			name                = excluded.name,
			summary             = excluded.summary,
			actors              = excluded.actors,
			acceptance_criteria = excluded.acceptance_criteria,
			test_plan           = excluded.test_plan,
			status              = excluded.status,
			updated_at          = excluded.updated_at`,
		f.FeatureID, f.Name, f.Summary,
		string(actors), string(ac), string(tp),
		f.Status, now,
	)
	return err
}

// Get returns a single feature by feature_id.
// Returns sql.ErrNoRows when the feature does not exist.
func Get(ctx context.Context, opts Options, featureID string) (Feature, error) {
	conn, err := db.Open(db.Options{Path: opts.DBPath, ReadOnly: true})
	if err != nil {
		return Feature{}, fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()
	return getConn(ctx, conn, featureID)
}

func getConn(ctx context.Context, conn *sql.DB, featureID string) (Feature, error) {
	row := conn.QueryRowContext(ctx, `
		SELECT feature_id, name, summary, actors, acceptance_criteria, test_plan, status, updated_at
		  FROM features
		 WHERE feature_id = ?`, featureID)
	return scanRow(row)
}

// List returns features ordered by updated_at DESC.
// When status is non-empty, only features with that status are returned.
func List(ctx context.Context, opts Options, status string) ([]Feature, error) {
	conn, err := db.Open(db.Options{Path: opts.DBPath, ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()
	return listConn(ctx, conn, status)
}

func listConn(ctx context.Context, conn *sql.DB, status string) ([]Feature, error) {
	var (
		q    string
		args []any
	)
	if status != "" {
		q = `SELECT feature_id, name, summary, actors, acceptance_criteria, test_plan, status, updated_at
			   FROM features WHERE status = ? ORDER BY updated_at DESC`
		args = []any{status}
	} else {
		q = `SELECT feature_id, name, summary, actors, acceptance_criteria, test_plan, status, updated_at
			   FROM features ORDER BY updated_at DESC`
	}
	rows, err := conn.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Feature
	for rows.Next() {
		f, err := scanRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// Delete removes a feature by feature_id.
// Returns nil when the feature does not exist (idempotent).
func Delete(ctx context.Context, opts Options, featureID string) error {
	conn, err := db.Open(db.Options{Path: opts.DBPath})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()
	_, err = conn.ExecContext(ctx, `DELETE FROM features WHERE feature_id = ?`, featureID)
	return err
}

// Search returns features whose name or summary contain query (case-insensitive),
// ordered by updated_at DESC.
func Search(ctx context.Context, opts Options, query string) ([]Feature, error) {
	conn, err := db.Open(db.Options{Path: opts.DBPath, ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	defer conn.Close()

	like := "%" + query + "%"
	rows, err := conn.QueryContext(ctx, `
		SELECT feature_id, name, summary, actors, acceptance_criteria, test_plan, status, updated_at
		  FROM features
		 WHERE LOWER(name)    LIKE LOWER(?)
		    OR LOWER(summary) LIKE LOWER(?)
		 ORDER BY updated_at DESC`, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Feature
	for rows.Next() {
		f, err := scanRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// scanRow reads a *sql.Row into a Feature.
func scanRow(row *sql.Row) (Feature, error) {
	var f Feature
	var actors, ac, tp string
	if err := row.Scan(&f.FeatureID, &f.Name, &f.Summary, &actors, &ac, &tp, &f.Status, &f.UpdatedAt); err != nil {
		return Feature{}, err
	}
	return unmarshal(f, actors, ac, tp)
}

// scanRows reads a *sql.Rows cursor into a Feature.
func scanRows(rows *sql.Rows) (Feature, error) {
	var f Feature
	var actors, ac, tp string
	if err := rows.Scan(&f.FeatureID, &f.Name, &f.Summary, &actors, &ac, &tp, &f.Status, &f.UpdatedAt); err != nil {
		return Feature{}, err
	}
	return unmarshal(f, actors, ac, tp)
}

func unmarshal(f Feature, actors, ac, tp string) (Feature, error) {
	if err := json.Unmarshal([]byte(actors), &f.Actors); err != nil {
		return Feature{}, fmt.Errorf("unmarshal actors: %w", err)
	}
	if err := json.Unmarshal([]byte(ac), &f.AcceptanceCriteria); err != nil {
		return Feature{}, fmt.Errorf("unmarshal acceptance_criteria: %w", err)
	}
	if err := json.Unmarshal([]byte(tp), &f.TestPlan); err != nil {
		return Feature{}, fmt.Errorf("unmarshal test_plan: %w", err)
	}
	return f, nil
}
