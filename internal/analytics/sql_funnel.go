package analytics

import (
	"context"
	"fmt"
	"strings"
)
// ─── 1. funnel ─────────────────────────────────────────────────────────────

// QueryFunnel returns one row per stage with a distinct-user count and the
// pairwise conversion / drop-off vs the previous stage.
//
// v1 semantics: "loose funnel" — each stage counts distinct users who fired
// that event within the time window and the (optional) segment. Ordered
// funnel (user must have fired stage[i-1] strictly before stage[i]) is a
// follow-up flag once a real user requests it.
func (a *SQLAdapter) QueryFunnel(ctx context.Context, q FunnelQuery) (FunnelResult, error) {
	if len(q.Stages) == 0 {
		return FunnelResult{}, fmt.Errorf("analytics: funnel requires at least one stage")
	}

	from, to := rangeArgs(q.TimeRange)
	stages := make([]FunnelStage, 0, len(q.Stages))
	var prior int64
	for i, stage := range q.Stages {
		clauses := []string{"event_name = ?", "occurred_at >= ?", "occurred_at < ?"}
		args := []any{stage, from, to}
		segmentClause(q.Segment, &clauses, &args)

		query := "SELECT COUNT(DISTINCT user_id) FROM events WHERE " + strings.Join(clauses, " AND ")
		var count int64
		if err := a.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
			return FunnelResult{}, fmt.Errorf("analytics: funnel stage %d: %w", i, err)
		}

		var conv, drop float64
		if i == 0 {
			conv = 100
			drop = 0
		} else if prior > 0 {
			conv = float64(count) / float64(prior) * 100
			drop = 100 - conv
		}
		stages = append(stages, FunnelStage{
			Stage:               stage,
			Count:               count,
			ConversionFromPrior: conv,
			DropOffFromPrior:    drop,
		})
		prior = count
	}

	total := int64(0)
	if len(stages) > 0 {
		total = stages[0].Count
	}
	return FunnelResult{
		Funnel:        stages,
		TotalUsers:    total,
		TimeRangeUsed: q.TimeRange,
	}, nil
}

// ─── 2. cohort ─────────────────────────────────────────────────────────────

// retentionWindows is the canonical set of retention buckets the funnel skill
// expects (D1, D7, D30, D90, D180). These are days *since* the user's signup
// anchor event. ordered = render-friendly listing.
var retentionWindows = []struct {
	label string
	days  int
}{
	{"D1", 1},
	{"D7", 7},
	{"D30", 30},
	{"D90", 90},
	{"D180", 180},
}

// QueryCohort groups users by their first "signup" event into weekly /
// monthly / quarterly buckets and reports D1/D7/D30/D90/D180 active rates.
//
// "Active" is one or more events after the anchor; revenue retention narrows
// to events with value_cents > 0; feature_use is identical to active in v1
// because the events stream is the only signal (callers tag the metric for
