package knowledge

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/0xmhha/buddy/internal/db"
)

// ErrNotFound mirrors the agent / sessions / feature pkgs so callers can
// distinguish "row missing" from "real DB error".
var ErrNotFound = errors.New("knowledge: not found")

// Store owns the `chunks` table CRUD. Mirrors the sessions.Store shape
// (one method per intent) so callers across the binary look uniform.
type Store struct {
	db db.Conn
}

// NewStore wraps any db.Conn implementation.
func NewStore(conn db.Conn) *Store { return &Store{db: conn} }

// Insert appends a new chunk row. Returns the autoincrement id.
// Caller-supplied CreatedAt is preserved (zero → now()).
func (s *Store) Insert(ctx context.Context, c Chunk) (int64, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	embBlob, err := encodeEmbedding(c.Embedding)
	if err != nil {
		return 0, fmt.Errorf("knowledge: insert: encode embedding: %w", err)
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO chunks (session_id, content, token_count, embedding, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		c.SessionID, c.Content, c.TokenCount, embBlob,
		c.CreatedAt.UTC().UnixMilli(),
	)
	if err != nil {
		return 0, fmt.Errorf("knowledge: insert: %w", err)
	}
	return res.LastInsertId()
}

// UpdateEmbedding sets the embedding column on an existing chunk. The
// embed pipeline runs after Insert (Insert may pass nil Embedding when
// the Python venv is unavailable; later runs fill it).
func (s *Store) UpdateEmbedding(ctx context.Context, chunkID int64, emb []float32) error {
	blob, err := encodeEmbedding(emb)
	if err != nil {
		return fmt.Errorf("knowledge: update embedding: %w", err)
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE chunks SET embedding = ? WHERE id = ?`,
		blob, chunkID)
	if err != nil {
		return fmt.Errorf("knowledge: update embedding: %w", err)
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

// Get returns one chunk by id.
func (s *Store) Get(ctx context.Context, id int64) (Chunk, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, session_id, content, token_count, embedding, created_at
		 FROM chunks WHERE id = ?`, id)
	c, err := scanChunk(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Chunk{}, ErrNotFound
	}
	return c, err
}

// ListBySession returns chunks for a given session in insertion order
// (id ASC). Empty session => no rows.
func (s *Store) ListBySession(ctx context.Context, sessionID string) ([]Chunk, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, content, token_count, embedding, created_at
		 FROM chunks WHERE session_id = ? ORDER BY id ASC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("knowledge: list by session: %w", err)
	}
	defer rows.Close()
	return scanChunkRows(rows)
}

// All returns every chunk in id ASC order. Used by the BM25 indexer +
// vector retrieval — both load the full set into memory and compute over
// it (acceptable at v0.10.0's expected ~10k chunks per ADR-014 §Q2).
func (s *Store) All(ctx context.Context) ([]Chunk, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, content, token_count, embedding, created_at
		 FROM chunks ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("knowledge: all: %w", err)
	}
	defer rows.Close()
	return scanChunkRows(rows)
}

// Count is the cheap stats helper — `buddy knowledge stats` reads it.
func (s *Store) Count(ctx context.Context) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunks`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("knowledge: count: %w", err)
	}
	return n, nil
}

// CountWithEmbedding lets stats report ingest progress (how many of the
// chunks have had the embedding pass run).
func (s *Store) CountWithEmbedding(ctx context.Context) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunks WHERE embedding IS NOT NULL`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("knowledge: count w/embedding: %w", err)
	}
	return n, nil
}

// DeleteBySession removes every chunk attached to a session. Used by
// `buddy knowledge ingest --rebuild` for the targeted session before
// re-chunking, and indirectly by the sessions FK cascade when the
// underlying session row is deleted.
func (s *Store) DeleteBySession(ctx context.Context, sessionID string) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM chunks WHERE session_id = ?`, sessionID)
	if err != nil {
		return 0, fmt.Errorf("knowledge: delete by session: %w", err)
	}
	return res.RowsAffected()
}

// DeleteAll wipes the chunks table. `buddy knowledge ingest --rebuild`
// without --session uses this to start clean.
func (s *Store) DeleteAll(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM chunks`)
	if err != nil {
		return 0, fmt.Errorf("knowledge: delete all: %w", err)
	}
	return res.RowsAffected()
}

// ─── encoding helpers ────────────────────────────────────────────────

// encodeEmbedding serialises a float32 slice into a compact little-endian
// BLOB. Nil input → NULL (database-side). We picked binary over JSON
// here for storage efficiency — float32 × 768 dims = 3 KB binary vs
// ~9 KB JSON. JSON would be marginally easier to inspect but we never
// hand-read embeddings; cosine sim consumes them programmatically.
func encodeEmbedding(emb []float32) (any, error) {
	if emb == nil {
		return nil, nil
	}
	buf := make([]byte, 4*len(emb))
	for i, v := range emb {
		binary.LittleEndian.PutUint32(buf[i*4:i*4+4], math.Float32bits(v))
	}
	return buf, nil
}

// decodeEmbedding is the inverse. Empty / nil blob → nil slice.
// Returns an error if the byte length isn't a multiple of 4.
func decodeEmbedding(blob []byte) ([]float32, error) {
	if len(blob) == 0 {
		return nil, nil
	}
	if len(blob)%4 != 0 {
		// Fall back to JSON in case some future writer uses that shape.
		// Catches both legitimate-but-old data and corrupt rows; the
		// JSON branch errors back to the caller with a clear message.
		var out []float32
		if err := json.Unmarshal(blob, &out); err == nil {
			return out, nil
		}
		return nil, fmt.Errorf("knowledge: decode embedding: unexpected length %d", len(blob))
	}
	n := len(blob) / 4
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		bits := binary.LittleEndian.Uint32(blob[i*4 : i*4+4])
		out[i] = math.Float32frombits(bits)
	}
	return out, nil
}

func scanChunk(row *sql.Row) (Chunk, error) {
	var (
		c         Chunk
		blob      []byte
		createdAt int64
	)
	err := row.Scan(&c.ID, &c.SessionID, &c.Content, &c.TokenCount, &blob, &createdAt)
	if err != nil {
		return Chunk{}, err
	}
	emb, derr := decodeEmbedding(blob)
	if derr != nil {
		return Chunk{}, derr
	}
	c.Embedding = emb
	c.CreatedAt = time.UnixMilli(createdAt).UTC()
	return c, nil
}

func scanChunkRows(rows *sql.Rows) ([]Chunk, error) {
	var out []Chunk
	for rows.Next() {
		var (
			c         Chunk
			blob      []byte
			createdAt int64
		)
		err := rows.Scan(&c.ID, &c.SessionID, &c.Content, &c.TokenCount, &blob, &createdAt)
		if err != nil {
			return nil, err
		}
		emb, derr := decodeEmbedding(blob)
		if derr != nil {
			return nil, derr
		}
		c.Embedding = emb
		c.CreatedAt = time.UnixMilli(createdAt).UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}
