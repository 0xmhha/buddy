package advisor

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/knowledge"
	"github.com/0xmhha/buddy/internal/schema"
	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/usage"
)

// newRunner builds a fully-wired Evaluator over an isolated on-disk DB
// plus pre-seeded sessions. Returns the runner + raw stores so tests
// can inject more data.
func newRunner(t *testing.T) (*Evaluator, *sessions.Store, *Store) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "advisor-eval.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	ss := sessions.NewStore(conn)
	as := NewStore(conn)
	return &Evaluator{
		Thresholds: DefaultThresholds(),
		Usage:      usage.NewService(conn),
		Sessions:   ss,
		Advisories: as,
	}, ss, as
}

func seedSession(t *testing.T, ss *sessions.Store, id string, started time.Time, durMs int64, usage schema.TokenUsage, ended bool) {
	t.Helper()
	sess := sessions.Session{
		ID: id, PID: 1, TranscriptPath: "/tmp/" + id + ".jsonl",
		StartedAt: started, LastActive: started.Add(time.Duration(durMs) * time.Millisecond),
		Usage: usage, LastOffset: 1024, Metadata: "{}",
	}
	require.NoError(t, ss.Upsert(context.Background(), sess))
	if ended {
		t2 := sess.LastActive
		require.NoError(t, ss.SetEndedAt(context.Background(), id, &t2))
	}
}

func TestEvaluator_DisabledReturnsNil(t *testing.T) {
	t.Parallel()
	r, _, _ := newRunner(t)
	r.Thresholds.Disabled = true
	got, err := r.Run(context.Background())
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestEvaluator_NoDataFiresNothing(t *testing.T) {
	t.Parallel()
	r, _, _ := newRunner(t)
	got, err := r.Run(context.Background())
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestEvaluator_TokenSpikeFires(t *testing.T) {
	t.Parallel()
	r, ss, _ := newRunner(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	r.Now = func() time.Time { return now }
	// 7d baseline: 7 sessions × 100k input each = 700k total → avg 100k/day.
	for i := 0; i < 7; i++ {
		seedSession(t, ss, "wk-"+string(rune('a'+i)),
			now.Add(time.Duration(-(i+1)*24)*time.Hour), 60_000,
			schema.TokenUsage{InputTokens: 100_000}, true)
	}
	// Today: a single session with 500k input → 5x avg
	seedSession(t, ss, "today",
		now.Add(-1*time.Hour), 60_000,
		schema.TokenUsage{InputTokens: 500_000}, false)

	got, err := r.Run(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	var fired bool
	for _, a := range got {
		if a.Kind == KindTokenSpikeDay {
			fired = true
			require.Equal(t, SeverityHigh, a.Severity)
		}
	}
	require.True(t, fired, "token spike rule must fire on 5x baseline")
}

func TestEvaluator_LongSessionFires(t *testing.T) {
	t.Parallel()
	r, ss, _ := newRunner(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	r.Now = func() time.Time { return now }
	// Active session running 6h (default threshold 4h → fires).
	seedSession(t, ss, "longone",
		now.Add(-6*time.Hour),
		int64(6*time.Hour/time.Millisecond),
		schema.TokenUsage{InputTokens: 1}, false)
	got, err := r.Run(ctx)
	require.NoError(t, err)
	var fired bool
	for _, a := range got {
		if a.Kind == KindLongSession {
			fired = true
		}
	}
	require.True(t, fired)
}

func TestEvaluator_DedupSkipsRecent(t *testing.T) {
	t.Parallel()
	r, ss, as := newRunner(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	r.Now = func() time.Time { return now }

	// Seed conditions that fire token-daily-cap (1M > 500k default).
	seedSession(t, ss, "big",
		now.Add(-1*time.Hour), 60_000,
		schema.TokenUsage{InputTokens: 1_000_000}, false)

	// Pre-populate an advisories row of the same kind within window.
	_, err := as.Insert(ctx, Advisory{
		Kind: KindTokenDailyCap, Severity: SeverityWarn,
		Message: "previously fired", CreatedAt: now.Add(-1 * time.Hour),
	})
	require.NoError(t, err)

	got, err := r.Run(ctx)
	require.NoError(t, err)
	for _, a := range got {
		require.NotEqual(t, KindTokenDailyCap, a.Kind,
			"dedup must skip kinds inside DedupWindow")
	}
}

func TestEvaluator_PersistWritesRows(t *testing.T) {
	t.Parallel()
	r, ss, as := newRunner(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	r.Now = func() time.Time { return now }
	seedSession(t, ss, "big",
		now.Add(-1*time.Hour), 60_000,
		schema.TokenUsage{InputTokens: 1_000_000}, false)
	got, err := r.Persist(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	for _, a := range got {
		require.NotZero(t, a.ID, "Persist returns rows with populated id")
	}
	count, _ := as.Count(ctx)
	require.Equal(t, int64(len(got)), count)
}

func TestEvaluator_EnrichWithRetrieval_AppendsChunks(t *testing.T) {
	t.Parallel()
	r, ss, _ := newRunner(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	r.Now = func() time.Time { return now }
	seedSession(t, ss, "big",
		now.Add(-1*time.Hour), 60_000,
		schema.TokenUsage{InputTokens: 1_000_000}, false)

	// Wire knowledge store with relevant chunks. BM25 only (no embedder).
	// Re-use the same DB the evaluator opened.
	kstore := knowledge.NewStore(r.Advisories.db)
	_, err := kstore.Insert(ctx, knowledge.Chunk{
		SessionID: "big", Content: "프롬프트 효율 token spend 최적화 사례",
		TokenCount: 6,
	})
	require.NoError(t, err)
	r.Knowledge = kstore

	got, err := r.Run(ctx)
	require.NoError(t, err)
	var saw bool
	for _, a := range got {
		if a.Kind != KindTokenDailyCap {
			continue
		}
		saw = true
		var hasChunk bool
		for _, ev := range a.Evidence {
			if ev.Type == "chunk" {
				hasChunk = true
				require.NotZero(t, ev.ChunkID)
			}
		}
		require.True(t, hasChunk, "knowledge retrieval evidence must be appended")
	}
	require.True(t, saw)
}
