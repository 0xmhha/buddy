package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// SQLAdapter is the SQLite-flavoured reference implementation of Adapter.
// PostgreSQL / MySQL adapters land alongside this one in spec §8 follow-up
// cycles; they reuse the same query shape and only need driver-specific
// placeholder / autoincrement substitution.
//
// SQLAdapter is safe for concurrent use because *sql.DB itself is — every
// method uses ctx-aware ExecContext / QueryContext and never holds state
// between calls.
type SQLAdapter struct {
	db *sql.DB
}

// NewSQLAdapter wraps an open *sql.DB. The caller still owns lifecycle —
// closing the DB closes the adapter.
func NewSQLAdapter(db *sql.DB) *SQLAdapter {
	return &SQLAdapter{db: db}
}

// ─── helpers ───────────────────────────────────────────────────────────────

func iso(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// rangeArgs returns (from, to) ISO timestamps; zero values become open ends.
func rangeArgs(r TimeRange) (string, string) {
	from := "1970-01-01T00:00:00Z"
	to := "9999-12-31T23:59:59Z"
	if !r.From.IsZero() {
		from = iso(r.From)
	}
	if !r.To.IsZero() {
		to = iso(r.To)
	}
	return from, to
}

func parseISO(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// segmentClause appends a "(segment_dim = ? AND segment_val = ?)" filter
// to the given clauses + args slice when seg is non-nil.
func segmentClause(seg *Segment, clauses *[]string, args *[]any) {
	if seg == nil || seg.Dimension == "" {
		return
	}
	*clauses = append(*clauses, "segment_dim = ? AND segment_val = ?")
	*args = append(*args, seg.Dimension, seg.Value)
}

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
// future-proofing).
func (a *SQLAdapter) QueryCohort(ctx context.Context, q CohortQuery) (CohortResult, error) {
	switch q.CohortDimension {
	case CohortWeekly, CohortMonthly, CohortQuarterly:
	default:
		return CohortResult{}, fmt.Errorf("analytics: unsupported cohort_dimension %q (want weekly|monthly|quarterly)", q.CohortDimension)
	}
	switch q.RetentionMetric {
	case RetentionActive, RetentionRevenue, RetentionFeatureUse:
	case "":
		q.RetentionMetric = RetentionActive
	default:
		return CohortResult{}, fmt.Errorf("analytics: unsupported retention_metric %q (want active|revenue|feature_use)", q.RetentionMetric)
	}

	rows, err := a.db.QueryContext(ctx, `
		SELECT user_id, MIN(occurred_at) AS signup_at
		FROM events
		WHERE event_name = 'signup' AND user_id IS NOT NULL
		GROUP BY user_id
	`)
	if err != nil {
		return CohortResult{}, fmt.Errorf("analytics: cohort signup scan: %w", err)
	}

	cohortMembers := make(map[string][]struct {
		UserID    string
		SignupAt  time.Time
	})
	defer rows.Close()
	for rows.Next() {
		var userID, signupAt string
		if err := rows.Scan(&userID, &signupAt); err != nil {
			return CohortResult{}, fmt.Errorf("analytics: cohort scan row: %w", err)
		}
		t, err := parseISO(signupAt)
		if err != nil {
			continue
		}
		cohortID := bucketCohortID(t, q.CohortDimension)
		cohortMembers[cohortID] = append(cohortMembers[cohortID], struct {
			UserID   string
			SignupAt time.Time
		}{userID, t})
	}
	if err := rows.Err(); err != nil {
		return CohortResult{}, fmt.Errorf("analytics: cohort rows.Err: %w", err)
	}

	cohorts := make([]CohortRow, 0, len(cohortMembers))
	for cohortID, members := range cohortMembers {
		size := int64(len(members))
		retention := make(map[string]float64, len(retentionWindows))
		for _, w := range retentionWindows {
			var active int64
			for _, m := range members {
				wStart := m.SignupAt.AddDate(0, 0, w.days-1)
				wEnd := m.SignupAt.AddDate(0, 0, w.days)
				count, err := a.countActiveInWindow(ctx, m.UserID, wStart, wEnd, q.RetentionMetric)
				if err != nil {
					return CohortResult{}, err
				}
				if count > 0 {
					active++
				}
			}
			if size > 0 {
				retention[w.label] = float64(active) / float64(size) * 100
			}
		}
		cohorts = append(cohorts, CohortRow{CohortID: cohortID, Size: size, Retention: retention})
	}
	sort.Slice(cohorts, func(i, j int) bool { return cohorts[i].CohortID < cohorts[j].CohortID })
	return CohortResult{Cohorts: cohorts}, nil
}

func (a *SQLAdapter) countActiveInWindow(ctx context.Context, userID string, from, to time.Time, metric RetentionMetric) (int64, error) {
	clauses := []string{"user_id = ?", "occurred_at >= ?", "occurred_at < ?"}
	args := []any{userID, iso(from), iso(to)}
	if metric == RetentionRevenue {
		clauses = append(clauses, "value_cents > 0")
	}
	var count int64
	err := a.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM events WHERE "+strings.Join(clauses, " AND "),
		args...,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("analytics: cohort active window: %w", err)
	}
	return count, nil
}

func bucketCohortID(t time.Time, dim CohortDimension) string {
	year, week := t.UTC().ISOWeek()
	switch dim {
	case CohortWeekly:
		return fmt.Sprintf("%04d-W%02d", year, week)
	case CohortMonthly:
		return t.UTC().Format("2006-01")
	case CohortQuarterly:
		q := (int(t.UTC().Month())-1)/3 + 1
		return fmt.Sprintf("%04d-Q%d", t.UTC().Year(), q)
	}
	return t.UTC().Format("2006-01-02")
}

// ─── 3. A/B experiment ─────────────────────────────────────────────────────

// QueryABExperiment summarises one experiment: per-variant sample size,
// primary metric mean + 95% normal CI, p-value (Welch's two-sample t-test
// when exactly two variants), and a recommendation derived from significance
// + guardrail breaches.
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
// trust score (4-dimensional composite, max 100 — Release-It! Nygard).
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
// optionally surfaces spike rows whose z-score over the baseline exceeds 2.
func (a *SQLAdapter) QueryCost(ctx context.Context, q CostQuery) (CostResult, error) {
	from, to := rangeArgs(q.TimeRange)

	var total int64
	if err := a.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(cost_cents), 0) FROM cost_records WHERE recorded_at >= ? AND recorded_at < ?",
		from, to,
	).Scan(&total); err != nil {
		return CostResult{}, fmt.Errorf("analytics: cost total: %w", err)
	}

	res := CostResult{TotalCostCents: total, ByDimension: map[string]int64{}}

	if col := costDrillColumn(q.DrillDown); col != "" {
		rows, err := a.db.QueryContext(ctx,
			"SELECT "+col+", SUM(cost_cents) FROM cost_records WHERE recorded_at >= ? AND recorded_at < ? GROUP BY "+col,
			from, to)
		if err != nil {
			return CostResult{}, fmt.Errorf("analytics: cost drill: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var dim sql.NullString
			var sum int64
			if err := rows.Scan(&dim, &sum); err != nil {
				return CostResult{}, err
			}
			label := dim.String
			if !dim.Valid || label == "" {
				label = "(unattributed)"
			}
			res.ByDimension[label] = sum
		}
		if err := rows.Err(); err != nil {
			return CostResult{}, err
		}
	}

	if q.AnomalyDetection {
		anoms, err := a.detectCostAnomalies(ctx, from, to, q.DrillDown)
		if err != nil {
			return CostResult{}, err
		}
		res.Anomalies = anoms
	}
	return res, nil
}

func costDrillColumn(d CostDrillDown) string {
	switch d {
	case CostByService:
		return "service"
	case CostByComponent:
		return "component"
	case CostByRegion:
		return "region"
	case CostByAccount:
		return "account"
	case CostByTag:
		return "tag"
	}
	return ""
}

func (a *SQLAdapter) detectCostAnomalies(ctx context.Context, from, to string, drill CostDrillDown) ([]CostAnomaly, error) {
	col := costDrillColumn(drill)
	if col == "" {
		col = "service" // sensible default when caller omitted drill_down
	}
	rows, err := a.db.QueryContext(ctx,
		"SELECT "+col+", recorded_at, cost_cents FROM cost_records WHERE recorded_at >= ? AND recorded_at < ? ORDER BY recorded_at",
		from, to)
	if err != nil {
		return nil, fmt.Errorf("analytics: cost anomaly scan: %w", err)
	}
	defer rows.Close()

	type row struct {
		dim   string
		at    time.Time
		cents int64
	}
	bucket := map[string][]row{}
	for rows.Next() {
		var dim sql.NullString
		var at string
		var c int64
		if err := rows.Scan(&dim, &at, &c); err != nil {
			return nil, err
		}
		t, err := parseISO(at)
		if err != nil {
			continue
		}
		label := dim.String
		if !dim.Valid || label == "" {
			label = "(unattributed)"
		}
		bucket[label] = append(bucket[label], row{label, t, c})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	anoms := []CostAnomaly{}
	for _, rs := range bucket {
		if len(rs) < 3 {
			continue
		}
		vals := make([]float64, len(rs))
		for i, r := range rs {
			vals[i] = float64(r.cents)
		}
		m, sd := mean(vals), math.Sqrt(variance(vals))
		if sd == 0 {
			continue
		}
		for _, r := range rs {
			z := (float64(r.cents) - m) / sd
			if math.Abs(z) >= 2 {
				anoms = append(anoms, CostAnomaly{Dimension: r.dim, SpikeAt: r.at, CostCents: r.cents, ZScore: z})
			}
		}
	}
	sort.Slice(anoms, func(i, j int) bool { return anoms[i].SpikeAt.Before(anoms[j].SpikeAt) })
	return anoms, nil
}

// ─── 6. SLO burn ───────────────────────────────────────────────────────────

// QuerySLOBurn computes the current SLI mean over time_window and applies
// the Google SRE multi-window burn-rate thresholds against the SLO target.
//
// burn_rate = current_value / slo_target. We assume "higher is worse" SLIs
// (latency, error rate). For "higher is better" SLIs (availability) callers
// can either invert the target or use a separate query type in a follow-up.
//
// Gate thresholds (matching the analyze-error-budget skill defaults):
//   ≥14×  critical → rollback (1h burn rate; budget gone in <2.5h)
//   ≥6×   critical → freeze   (6h burn rate; budget gone in <2.5d)
//   ≥3×   warning  → limit    (1d burn rate)
//   ≥1×   warning  → ship     (still over target but not catastrophic)
//   <1×   info     → ship
func (a *SQLAdapter) QuerySLOBurn(ctx context.Context, q SLOBurnQuery) (SLOBurnResult, error) {
	if q.SLI == "" {
		return SLOBurnResult{}, fmt.Errorf("analytics: slo_burn requires sli")
	}
	if q.SLOTarget <= 0 {
		return SLOBurnResult{}, fmt.Errorf("analytics: slo_burn requires slo_target > 0")
	}
	dur, err := parseTimeWindow(q.TimeWindow)
	if err != nil {
		return SLOBurnResult{}, err
	}

	now := time.Now().UTC()
	windowStart := now.Add(-dur)

	var windowVal sql.NullFloat64
	err = a.db.QueryRowContext(ctx,
		`SELECT AVG(value) FROM slo_observations WHERE sli = ? AND observed_at >= ? AND observed_at < ?`,
		q.SLI, iso(windowStart), iso(now),
	).Scan(&windowVal)
	if err != nil {
		return SLOBurnResult{}, fmt.Errorf("analytics: slo_burn window avg: %w", err)
	}
	if !windowVal.Valid {
		return SLOBurnResult{}, ErrNotFound
	}

	current := windowVal.Float64
	burnRate := current / q.SLOTarget
	budgetRemaining := math.Max(0, 100*(1-burnRate))

	alert := "info"
	gate := "ship"
	switch {
	case burnRate >= 14:
		alert, gate = "critical", "rollback"
	case burnRate >= 6:
		alert, gate = "critical", "freeze"
	case burnRate >= 3:
		alert, gate = "warning", "limit"
	case burnRate >= 1:
		alert, gate = "warning", "ship"
	}

	return SLOBurnResult{
		SLI:                 q.SLI,
		SLOTarget:           q.SLOTarget,
		TimeWindow:          q.TimeWindow,
		CurrentValue:        current,
		BudgetRemainingPct:  budgetRemaining,
		BurnRate:            burnRate,
		AlertLevel:          alert,
		ReleaseGateDecision: gate,
	}, nil
}

func parseTimeWindow(s string) (time.Duration, error) {
	switch s {
	case "1h":
		return time.Hour, nil
	case "6h":
		return 6 * time.Hour, nil
	case "1d":
		return 24 * time.Hour, nil
	case "3d":
		return 3 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("analytics: unsupported time_window %q (want 1h|6h|1d|3d|30d)", s)
}

// ─── 7. feedback corpus ────────────────────────────────────────────────────

// QueryFeedbackCorpus returns total item count, optional topic clustering
// (pre-classified topic column), and optional NPS segmentation. v1 uses
// pre-existing topic/sentiment columns; topic inference from raw text is a
// follow-up (would need an LLM call or an embedded clusterer).
func (a *SQLAdapter) QueryFeedbackCorpus(ctx context.Context, q FeedbackQuery) (FeedbackResult, error) {
	from, to := rangeArgs(q.TimeRange)

	clauses := []string{"occurred_at >= ?", "occurred_at < ?"}
	args := []any{from, to}
	if q.Source != "" {
		clauses = append(clauses, "source = ?")
		args = append(args, string(q.Source))
	}
	whereSQL := strings.Join(clauses, " AND ")

	var total int64
	if err := a.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM feedback_items WHERE "+whereSQL, args...,
	).Scan(&total); err != nil {
		return FeedbackResult{}, fmt.Errorf("analytics: feedback total: %w", err)
	}

	res := FeedbackResult{TotalItems: total}

	if q.TopicModeling {
		topics, err := a.queryFeedbackTopics(ctx, whereSQL, args, q.SentimentAnalysis)
		if err != nil {
			return FeedbackResult{}, err
		}
		res.Topics = topics
	}

	if q.Source == FeedbackNPSComment || q.Source == "" {
		bands, err := a.queryNPSBands(ctx, whereSQL, args)
		if err != nil {
			return FeedbackResult{}, err
		}
		if len(bands) > 0 {
			res.NPSSegments = bands
		}
	}
	return res, nil
}

func (a *SQLAdapter) queryFeedbackTopics(ctx context.Context, whereSQL string, args []any, withSentiment bool) ([]FeedbackTopic, error) {
	rows, err := a.db.QueryContext(ctx,
		"SELECT COALESCE(topic, '(unclassified)'), COALESCE(sentiment, ''), body FROM feedback_items WHERE "+whereSQL,
		args...)
	if err != nil {
		return nil, fmt.Errorf("analytics: feedback topics: %w", err)
	}
	defer rows.Close()

	type acc struct {
		count     int64
		sentiment map[string]int64
		quotes    []string
	}
	by := map[string]*acc{}
	for rows.Next() {
		var topic, sent, body string
		if err := rows.Scan(&topic, &sent, &body); err != nil {
			return nil, err
		}
		a := by[topic]
		if a == nil {
			a = &acc{sentiment: map[string]int64{}}
			by[topic] = a
		}
		a.count++
		if withSentiment && sent != "" {
			a.sentiment[sent]++
		}
		if len(a.quotes) < 3 && body != "" {
			a.quotes = append(a.quotes, body)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	topics := make([]FeedbackTopic, 0, len(by))
	for name, info := range by {
		t := FeedbackTopic{Topic: name, ItemCount: info.count, VerbatimQuotes: info.quotes}
		if withSentiment && len(info.sentiment) > 0 {
			t.Sentiment = info.sentiment
		}
		topics = append(topics, t)
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].ItemCount > topics[j].ItemCount })
	return topics, nil
}

func (a *SQLAdapter) queryNPSBands(ctx context.Context, whereSQL string, args []any) (map[string]NPSBand, error) {
	rows, err := a.db.QueryContext(ctx,
		"SELECT nps_score, COALESCE(topic, '(unclassified)') FROM feedback_items WHERE "+whereSQL+" AND nps_score IS NOT NULL",
		args...)
	if err != nil {
		return nil, fmt.Errorf("analytics: feedback nps bands: %w", err)
	}
	defer rows.Close()

	bands := map[string]*NPSBand{
		"promoter":  {TopTopics: []string{}},
		"passive":   {TopTopics: []string{}},
		"detractor": {TopTopics: []string{}},
	}
	topicCount := map[string]map[string]int64{
		"promoter":  {},
		"passive":   {},
		"detractor": {},
	}
	for rows.Next() {
		var score int64
		var topic string
		if err := rows.Scan(&score, &topic); err != nil {
			return nil, err
		}
		band := classifyNPS(score)
		bands[band].Count++
		topicCount[band][topic]++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := map[string]NPSBand{}
	hadAny := false
	for band, b := range bands {
		if b.Count == 0 {
			continue
		}
		hadAny = true
		type kv struct {
			topic string
			n     int64
		}
		ranked := make([]kv, 0, len(topicCount[band]))
		for t, n := range topicCount[band] {
			ranked = append(ranked, kv{t, n})
		}
		sort.Slice(ranked, func(i, j int) bool { return ranked[i].n > ranked[j].n })
		top := make([]string, 0, 3)
		for i, r := range ranked {
			if i >= 3 {
				break
			}
			top = append(top, r.topic)
		}
		out[band] = NPSBand{Count: b.Count, TopTopics: top}
	}
	if !hadAny {
		return nil, nil
	}
	return out, nil
}

// classifyNPS maps a 0-10 NPS score to the standard 3-band split.
func classifyNPS(score int64) string {
	switch {
	case score >= 9:
		return "promoter"
	case score >= 7:
		return "passive"
	default:
		return "detractor"
	}
}

// ─── tiny stats helpers ────────────────────────────────────────────────────

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func variance(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := mean(xs)
	s := 0.0
	for _, x := range xs {
		d := x - m
		s += d * d
	}
	return s / float64(len(xs)-1)
}

func percentile(xs []int64, p float64) int64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]int64(nil), xs...)
	slicesSortInt64(sorted)
	idx := int(math.Round(p * float64(len(sorted)-1)))
	idx = max(idx, 0)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// slicesSortInt64 is a tiny shim over sort.Slice. We keep it as a named
// helper so that swapping to slices.Sort once go.mod targets the appropriate
// Go version is a one-line change.
func slicesSortInt64(xs []int64) {
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
