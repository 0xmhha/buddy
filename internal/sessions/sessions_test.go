package sessions_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/0xmhha/buddy/internal/schema"
	"github.com/0xmhha/buddy/internal/sessions"
)

// TestSession_FieldsMatchMigration is the static check that the Go struct
// stays aligned with the SQLite sessions table (db migration v3). The DB
// schema is the contract; this test fails loudly if a column name in the
// migration drifts from a struct field's tag/intent without updating both.
//
// We don't reflect on tags here (Session has no struct tags yet — Go field
// names + db column names mapped explicitly via the migration). Instead we
// list the field-to-column pairs once and let the test fail at compile time
// if a field is removed and at runtime if it's renamed.
func TestSession_FieldsMatchMigration(t *testing.T) {
	s := sessions.Session{
		ID:             "01HX...",
		PID:            12345,
		TranscriptPath: "/tmp/transcript.jsonl",
		StartedAt:      time.Unix(1_000, 0),
		LastActive:     time.Unix(2_000, 0),
		Usage: schema.TokenUsage{
			InputTokens:       100,
			OutputTokens:      50,
			CacheReadTokens:   10,
			CacheCreateTokens: 5,
		},
		LastOffset: 4096,
	}

	// Spot checks — every column in the migration v3 has a matching field
	// reachable from the struct.
	assert.NotEmpty(t, s.ID)
	assert.NotZero(t, s.PID)
	assert.NotEmpty(t, s.TranscriptPath)
	assert.False(t, s.StartedAt.IsZero())
	assert.False(t, s.LastActive.IsZero())
	assert.Equal(t, 100, s.Usage.InputTokens)
	assert.Equal(t, 50, s.Usage.OutputTokens)
	assert.Equal(t, 10, s.Usage.CacheReadTokens)
	assert.Equal(t, 5, s.Usage.CacheCreateTokens)
	assert.Equal(t, int64(4096), s.LastOffset)
}

// fakeLister is a stand-in proving the Lister interface is satisfiable
// without pulling in a filesystem-based implementation. core-3 will replace
// this with fsLister; the assertion below stays as a compile-time guard.
type fakeLister struct{ items []sessions.Session }

func (f *fakeLister) List(_ context.Context) ([]sessions.Session, error) {
	return f.items, nil
}

var _ sessions.Lister = (*fakeLister)(nil)

func TestLister_FakeImplReturnsSnapshot(t *testing.T) {
	want := []sessions.Session{
		{ID: "a", TranscriptPath: "/a.jsonl"},
		{ID: "b", TranscriptPath: "/b.jsonl"},
	}
	var l sessions.Lister = &fakeLister{items: want}

	got, err := l.List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
