package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
)
func (a *SQLAdapter) QueryActorFailure(ctx context.Context, q ActorFailureQuery) (ActorFailureResult, error) {
	if q.Actor == "" {
		return ActorFailureResult{}, fmt.Errorf("analytics: actor_failure requires actor")
	}
	from, to := rangeArgs(q.TimeRange)

	rows, err := a.db.QueryContext(ctx, `
		SELECT occurred_at, recovered_at, blast_radius, COALESCE(recovery_pattern, '')
		FROM failures
		WHERE actor_type = ? AND occurred_at >= ? AND occurred_at < ?`,
		string(q.Actor), from, to)
	if err != nil {
		return ActorFailureResult{}, fmt.Errorf("analytics: actor_failure scan: %w", err)
	}
	defer rows.Close()

	var (
		occurrences   int64
		mttrSamples   []float64
		blastSamples  []int64
		patternsSeen  = map[string]struct{}{}
	)
	for rows.Next() {
		var occ string
		var rec sql.NullString
		var blast int64
		var pattern string
		if err := rows.Scan(&occ, &rec, &blast, &pattern); err != nil {
			return ActorFailureResult{}, err
		}
		occurrences++
		blastSamples = append(blastSamples, blast)
		if pattern != "" {
			patternsSeen[pattern] = struct{}{}
		}
		if rec.Valid {
			occT, _ := parseISO(occ)
			recT, _ := parseISO(rec.String)
			if !occT.IsZero() && !recT.IsZero() && recT.After(occT) {
				mttrSamples = append(mttrSamples, recT.Sub(occT).Minutes())
			}
		}
	}
	if err := rows.Err(); err != nil {
		return ActorFailureResult{}, err
	}

	windowDays := 1.0
	if !q.TimeRange.From.IsZero() && !q.TimeRange.To.IsZero() {
		windowDays = q.TimeRange.To.Sub(q.TimeRange.From).Hours() / 24
		if windowDays < 1 {
			windowDays = 1
		}
	}
	failureRate := float64(occurrences) / windowDays

	mttr := mean(mttrSamples)
	variance := variance(mttrSamples)
	blastP95 := percentile(blastSamples, 0.95)

	patterns := make([]string, 0, len(patternsSeen))
	for p := range patternsSeen {
		patterns = append(patterns, p)
	}
	sort.Strings(patterns)

	// Trust score: 4 dimensions weighted to 100.
	// Lower failure_rate / variance / mttr / blast_radius → higher trust.
	trust := 100.0
	trust -= clamp(failureRate*5, 0, 25)        // 25 pts for stability
	trust -= clamp(math.Sqrt(variance)/2, 0, 25) // 25 pts for predictability
	trust -= clamp(mttr/30, 0, 25)               // 25 pts for recoverability (30min = -25)
	trust -= clamp(float64(blastP95)/10, 0, 25)  // 25 pts for blast containment

	return ActorFailureResult{
		Actor:                   string(q.Actor),
		FailureRate:             failureRate,
		PredictabilityVariance:  variance,
		MTTRMinutes:             mttr,
		BlastRadius:             blastP95,
		TrustScore:              clamp(trust, 0, 100),
		RecoveryPatternsApplied: patterns,
	}, nil
}

// ─── 5. cost ───────────────────────────────────────────────────────────────

// QueryCost sums cost_records, optionally broken down by a dimension, and
