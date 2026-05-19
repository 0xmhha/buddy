package advisor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0xmhha/buddy/internal/knowledge"
	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/usage"
)

// Evaluator runs the rule set against live data and emits advisories.
// Combines the four substrates ADR-015 calls out:
//
//   - Usage      : token / session metrics (W7-2).
//   - Sessions   : active-session list for ruleLongSession (W7-1).
//   - Knowledge  : retrieval enrichment per advisory (W7-3a). Optional —
//                  nil store skips evidence enrichment; nil embedder
//                  forces BM25-only.
//   - Advisories : dedup window lookup (this package's Store).
//
// All fields except Advisories are read-only; Advisories.Insert is
// called by Persist (not Run) so callers can preview without writing.
type Evaluator struct {
	Thresholds Thresholds
	Usage      *usage.Service
	Sessions   *sessions.Store
	Knowledge  *knowledge.Store
	Embedder   knowledge.Embedder
	Advisories *Store

	// Now is injectable for tests. Production callers leave nil and
	// the runner uses time.Now().UTC().
	Now func() time.Time
}

func (e *Evaluator) now() time.Time {
	if e.Now != nil {
		return e.Now().UTC()
	}
	return time.Now().UTC()
}

// Run executes every rule against a freshly-built Snapshot and returns
// the advisories that fired AND pass the dedup window check.
//
// Dedup: each advisory's Kind is checked against the advisories table —
// if a row with the same kind was created in the last
// Thresholds.DedupWindow, the new advisory is dropped. This prevents
// noisy repeat firing from the daemon goroutine; on-demand CLI usage
// can opt out via DedupWindow=0 in their config (or by calling rules
// directly).
func (e *Evaluator) Run(ctx context.Context) ([]Advisory, error) {
	if e.Thresholds.Disabled {
		return nil, nil
	}
	t := e.Thresholds.WithDefaults()
	if e.Usage == nil {
		return nil, errors.New("advisor: Usage service is required")
	}

	snap, err := e.buildSnapshot(ctx, t)
	if err != nil {
		return nil, err
	}

	var fired []Advisory
	for _, rule := range allRules() {
		adv := rule(t, snap)
		if adv == nil {
			continue
		}
		if e.skipForDedup(ctx, adv.Kind, t.DedupWindow, snap.Now) {
			continue
		}
		e.enrichWithRetrieval(ctx, adv)
		fired = append(fired, *adv)
	}
	return fired, nil
}

// Persist executes the runner AND writes each advisory to the
// advisories table. Used by the daemon advisorMonitor and CLI
// `--persist`. Returns the written advisories (with ID populated).
func (e *Evaluator) Persist(ctx context.Context) ([]Advisory, error) {
	fired, err := e.Run(ctx)
	if err != nil {
		return nil, err
	}
	if e.Advisories == nil {
		// No store wired — surface the advisories but skip persistence.
		// The daemon should never hit this branch.
		return fired, nil
	}
	out := make([]Advisory, 0, len(fired))
	for _, a := range fired {
		id, err := e.Advisories.Insert(ctx, a)
		if err != nil {
			return out, fmt.Errorf("advisor: persist %s: %w", a.Kind, err)
		}
		a.ID = id
		out = append(out, a)
	}
	return out, nil
}

// buildSnapshot fetches everything the rules need in a single pass.
// Ordering of fetches doesn't matter (they're independent reads); we
// don't parallelise because the cost is ms-level on the expected
// corpus (~1500 sessions / 30d per ADR-013).
func (e *Evaluator) buildSnapshot(ctx context.Context, t Thresholds) (Snapshot, error) {
	now := e.now()
	w24 := usage.TimeWindow{Since: now.Add(-24 * time.Hour), Until: now}
	w7d := usage.TimeWindow{Since: now.Add(-7 * 24 * time.Hour), Until: now}

	spend24, err := e.Usage.QueryTokenSpend(ctx, w24)
	if err != nil {
		return Snapshot{}, err
	}
	stats24, err := e.Usage.QuerySessionStats(ctx, w24)
	if err != nil {
		return Snapshot{}, err
	}
	spend7, err := e.Usage.QueryTokenSpend(ctx, w7d)
	if err != nil {
		return Snapshot{}, err
	}

	var active []sessions.Session
	if e.Sessions != nil {
		// Active = ended_at IS NULL. ruleLongSession iterates them.
		ss, err := e.Sessions.List(ctx, sessions.ListOptions{IncludeEnded: false})
		if err != nil {
			return Snapshot{}, err
		}
		active = ss
	}

	return Snapshot{
		Now:            now,
		Window24h:      w24,
		Spend24h:       spend24,
		Stats24h:       stats24,
		Window7d:       w7d,
		Spend7d:        spend7,
		ActiveSessions: active,
	}, nil
}

// skipForDedup returns true when an advisory of `kind` was created
// inside the dedup window. Zero/negative window disables dedup.
func (e *Evaluator) skipForDedup(ctx context.Context, kind string, window time.Duration, now time.Time) bool {
	if e.Advisories == nil || window <= 0 {
		return false
	}
	_, err := e.Advisories.LastInWindow(ctx, kind, window, now)
	return err == nil
}

// enrichWithRetrieval runs a knowledge query keyed by the advisory's
// Kind + a short hint, appending up to 3 chunk evidence items.
// Failure is silent — advisories are useful even without evidence,
// and the embedder error path is already friend-tone elsewhere.
func (e *Evaluator) enrichWithRetrieval(ctx context.Context, a *Advisory) {
	if e.Knowledge == nil {
		return
	}
	chunks, err := e.Knowledge.All(ctx)
	if err != nil || len(chunks) == 0 {
		return
	}
	idx := knowledge.NewBM25Index(chunks)
	queryText := retrievalQueryFor(a.Kind)
	var emb []float32
	if e.Embedder != nil {
		res, err := e.Embedder.Embed(ctx, []knowledge.EmbedRequest{{ID: 0, Text: queryText}})
		if err == nil && len(res) > 0 {
			emb = res[0].Embedding
		}
	}
	hits := knowledge.HybridSearch(idx, chunks, queryText, emb, 3)
	for _, h := range hits {
		preview := strings.ReplaceAll(h.Content, "\n", " ")
		if len(preview) > 100 {
			preview = preview[:100] + "…"
		}
		a.Evidence = append(a.Evidence, EvidenceItem{
			Type:    "chunk",
			Detail:  preview,
			ChunkID: h.ID,
		})
	}
}

// retrievalQueryFor builds the query string passed to BM25 / vector
// search for a given advisory kind. Keeps the keyword choice in one
// place so persona text and retrieval intent stay aligned. Strings
// are Korean + English mixed — BM25 matches whichever appears in the
// chunk text.
func retrievalQueryFor(kind string) string {
	switch kind {
	case KindTokenSpikeDay, KindTokenDailyCap:
		return "토큰 사용량 프롬프트 효율 token spend prompt"
	case KindLongSession:
		return "세션 길이 작업 분할 long session break down task"
	case KindLowCacheRatio:
		return "프롬프트 캐시 cache hit 재사용 prompt template"
	case KindSessionVolumeDay:
		return "세션 빈도 컨텍스트 유지 session frequency context"
	default:
		return kind
	}
}
