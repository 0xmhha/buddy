package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/wm-it-22-00661/buddy/internal/schema"
)

// AppendToOutbox writes the payload to hook_outbox synchronously.
// Caller must validate the payload first; this layer trusts its input.
func AppendToOutbox(conn *sql.DB, p *schema.HookEventPayload) (int64, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}
	res, err := conn.Exec(
		"INSERT INTO hook_outbox (ts, payload) VALUES (?, ?)",
		p.Ts, string(raw),
	)
	if err != nil {
		return 0, fmt.Errorf("insert outbox: %w", err)
	}
	return res.LastInsertId()
}

// OutboxRow is the raw shape of a pending outbox entry.
type OutboxRow struct {
	ID      int64
	Ts      int64
	Payload string
}

// ReadPendingOutbox returns up to limit oldest-first pending rows.
func ReadPendingOutbox(conn *sql.DB, limit int) ([]OutboxRow, error) {
	rows, err := conn.Query(
		"SELECT id, ts, payload FROM hook_outbox WHERE consumed_at IS NULL ORDER BY id LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query outbox: %w", err)
	}
	defer rows.Close()

	var out []OutboxRow
	for rows.Next() {
		var r OutboxRow
		if err := rows.Scan(&r.ID, &r.Ts, &r.Payload); err != nil {
			return nil, fmt.Errorf("scan outbox: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// MarkConsumed sets consumed_at on the given ids. No-op for empty input.
func MarkConsumed(conn *sql.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, time.Now().UnixMilli())
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	q := "UPDATE hook_outbox SET consumed_at = ? WHERE id IN (" +
		strings.Join(placeholders, ",") + ")"
	if _, err := conn.Exec(q, args...); err != nil {
		return fmt.Errorf("update outbox: %w", err)
	}
	return nil
}
