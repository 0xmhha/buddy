package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)
func (a *SQLAdapter) QueryABExperiment(ctx context.Context, q ABExperimentQuery) (ABExperimentResult, error) {
	if q.ExperimentID == "" {
		return ABExperimentResult{}, fmt.Errorf("analytics: ab_experiment requires experiment_id")
	}

	var name, primaryMetric, startedAt string
	var endedAt sql.NullString
	err := a.db.QueryRowContext(ctx, `
		SELECT name, primary_metric, started_at, ended_at
		FROM ab_experiments
		WHERE experiment_id = ?`, q.ExperimentID,
	).Scan(&name, &primaryMetric, &startedAt, &endedAt)
	if err == sql.ErrNoRows {
		return ABExperimentResult{}, ErrNotFound
	}
	if err != nil {
		return ABExperimentResult{}, fmt.Errorf("analytics: ab_experiment header: %w", err)
	}
	startTime, err := parseISO(startedAt)
	if err != nil {
		return ABExperimentResult{}, fmt.Errorf("analytics: ab_experiment started_at: %w", err)
	}
	var endTime *time.Time
	if endedAt.Valid {
		t, err := parseISO(endedAt.String)
		if err == nil {
			endTime = &t
		}
	}

	rows, err := a.db.QueryContext(ctx, `
		SELECT
		  a.variant,
		  COUNT(DISTINCT a.user_id),
		  COALESCE(AVG(o.metric_value), 0),
		  COALESCE(SUM((o.metric_value) * (o.metric_value)), 0),
		  COUNT(o.metric_value)
		FROM ab_assignments a
		LEFT JOIN ab_metric_observations o
		  ON o.experiment_id = a.experiment_id
		  AND o.user_id = a.user_id
		  AND o.metric_name = ?
		WHERE a.experiment_id = ?
		GROUP BY a.variant
		ORDER BY a.variant`, primaryMetric, q.ExperimentID)
	if err != nil {
		return ABExperimentResult{}, fmt.Errorf("analytics: ab_experiment variants: %w", err)
	}
	defer rows.Close()

	type variantStats struct {
		variant string
		n       int64
		mean    float64
		sumSq   float64
		obs     int64
	}
	var stats []variantStats
	for rows.Next() {
		var v variantStats
		if err := rows.Scan(&v.variant, &v.n, &v.mean, &v.sumSq, &v.obs); err != nil {
			return ABExperimentResult{}, err
		}
		stats = append(stats, v)
	}
	if err := rows.Err(); err != nil {
		return ABExperimentResult{}, err
	}

	variants := make([]ABVariant, 0, len(stats))
	for _, s := range stats {
		variance := 0.0
		if s.obs > 1 {
			variance = (s.sumSq - float64(s.obs)*s.mean*s.mean) / float64(s.obs-1)
		}
		ci := 1.96 * math.Sqrt(math.Max(variance, 0)) / math.Sqrt(math.Max(float64(s.obs), 1))
		variants = append(variants, ABVariant{
			Variant:            s.variant,
			SampleSize:         s.n,
			PrimaryMetricValue: s.mean,
			ConfidenceInterval: [2]float64{s.mean - ci, s.mean + ci},
		})
	}

	sig := ABSignificance{Confidence: 95}
	if len(stats) == 2 {
		a0, a1 := stats[0], stats[1]
		if a0.obs > 1 && a1.obs > 1 {
			v0 := (a0.sumSq - float64(a0.obs)*a0.mean*a0.mean) / float64(a0.obs-1)
			v1 := (a1.sumSq - float64(a1.obs)*a1.mean*a1.mean) / float64(a1.obs-1)
			se := math.Sqrt(v0/float64(a0.obs) + v1/float64(a1.obs))
			if se > 0 {
				z := math.Abs(a1.mean-a0.mean) / se
				sig.PValue = 2 * (1 - normalCDF(z))
				sig.StatisticallySignificant = sig.PValue < 0.05
			}
		}
	}

	rec := "inconclusive"
	switch {
	case sig.StatisticallySignificant && len(stats) == 2 && stats[1].mean > stats[0].mean:
		rec = "ship"
	case sig.StatisticallySignificant && len(stats) == 2 && stats[1].mean < stats[0].mean:
		rec = "revert"
	case !sig.StatisticallySignificant && endTime == nil:
		rec = "continue"
	}

	return ABExperimentResult{
		ExperimentID:   q.ExperimentID,
		Name:           name,
		StartedAt:      startTime,
		EndedAt:        endTime,
		Variants:       variants,
		Significance:   sig,
		Recommendation: rec,
	}, nil
}

// normalCDF is the standard-normal cumulative distribution. Abramowitz
// & Stegun rational approximation — accurate to ~7e-8, fast, dependency-free.
func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

// ─── 4. actor failure ──────────────────────────────────────────────────────

// QueryActorFailure aggregates the failures table for one actor type:
// failure_rate (per day), mean MTTR, blast radius p95, and a heuristic
