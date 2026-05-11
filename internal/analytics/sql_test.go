package analytics

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
)

// newTestAdapter returns a freshly-migrated in-memory SQLite adapter for one
// test. Each test gets its own connection so parallel runs do not collide.
func newTestAdapter(t *testing.T) (*SQLAdapter, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, Migrate(context.Background(), db))
	return NewSQLAdapter(db), db
}

func seedEvents(t *testing.T, db *sql.DB, rows ...[]any) {
	t.Helper()
	for _, r := range rows {
		_, err := db.Exec(
			`INSERT INTO events (event_name, user_id, occurred_at, value_cents, segment_dim, segment_val) VALUES (?, ?, ?, ?, ?, ?)`,
			r...,
		)
		require.NoError(t, err)
	}
}

// ─── 1. funnel ────────────────────────────────────────────────────────────

func TestSQLAdapter_QueryFunnel_BasicLossy(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	tFunc := func(s string) string { return s }
	seedEvents(t, db,
		[]any{"signup", "u1", tFunc("2026-04-10T00:00:00Z"), nil, "", ""},
		[]any{"signup", "u2", tFunc("2026-04-11T00:00:00Z"), nil, "", ""},
		[]any{"signup", "u3", tFunc("2026-04-12T00:00:00Z"), nil, "", ""},
		[]any{"first_action", "u1", tFunc("2026-04-12T00:00:00Z"), nil, "", ""},
		[]any{"first_action", "u2", tFunc("2026-04-13T00:00:00Z"), nil, "", ""},
		[]any{"retention_d7", "u1", tFunc("2026-04-18T00:00:00Z"), nil, "", ""},
	)

	res, err := a.QueryFunnel(context.Background(), FunnelQuery{
		Stages:    []string{"signup", "first_action", "retention_d7"},
		TimeRange: TimeRange{From: parseFor(t, "2026-04-01T00:00:00Z"), To: parseFor(t, "2026-05-01T00:00:00Z")},
	})
	require.NoError(t, err)
	require.Len(t, res.Funnel, 3)
	require.Equal(t, int64(3), res.Funnel[0].Count)
	require.Equal(t, int64(2), res.Funnel[1].Count)
	require.Equal(t, int64(1), res.Funnel[2].Count)
	require.InDelta(t, 66.666, res.Funnel[1].ConversionFromPrior, 0.01)
	require.InDelta(t, 50.0, res.Funnel[2].ConversionFromPrior, 0.01)
	require.Equal(t, int64(3), res.TotalUsers)
}

func TestSQLAdapter_QueryFunnel_SegmentFiltersUsers(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	seedEvents(t, db,
		[]any{"signup", "u_pro", "2026-04-10T00:00:00Z", nil, "plan_tier", "pro"},
		[]any{"signup", "u_free", "2026-04-10T00:00:00Z", nil, "plan_tier", "free"},
		[]any{"first_action", "u_pro", "2026-04-11T00:00:00Z", nil, "plan_tier", "pro"},
	)

	res, err := a.QueryFunnel(context.Background(), FunnelQuery{
		Stages:  []string{"signup", "first_action"},
		Segment: &Segment{Dimension: "plan_tier", Value: "pro"},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.Funnel[0].Count)
	require.Equal(t, int64(1), res.Funnel[1].Count)
}

// ─── 2. cohort ────────────────────────────────────────────────────────────

func TestSQLAdapter_QueryCohort_WeeklyActive(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	// Two users sign up in the same ISO week, one active at D1 and D7.
	seedEvents(t, db,
		[]any{"signup", "u1", "2026-04-07T00:00:00Z", nil, "", ""},
		[]any{"signup", "u2", "2026-04-07T00:00:00Z", nil, "", ""},
		[]any{"feature_used", "u1", "2026-04-07T12:00:00Z", nil, "", ""},
		[]any{"feature_used", "u1", "2026-04-13T12:00:00Z", nil, "", ""}, // within D7 window
		[]any{"feature_used", "u2", "2026-04-08T00:00:00Z", nil, "", ""}, // D1 only
	)

	res, err := a.QueryCohort(context.Background(), CohortQuery{
		CohortDimension: CohortWeekly,
		RetentionMetric: RetentionActive,
	})
	require.NoError(t, err)
	require.Len(t, res.Cohorts, 1)
	require.Equal(t, int64(2), res.Cohorts[0].Size)
	require.InDelta(t, 100.0, res.Cohorts[0].Retention["D1"], 0.01)
	require.InDelta(t, 50.0, res.Cohorts[0].Retention["D7"], 0.01)
}

// ─── 3. A/B experiment ────────────────────────────────────────────────────

func TestSQLAdapter_QueryABExperiment_SignificantShipRecommendation(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	_, err := db.Exec(
		`INSERT INTO ab_experiments (experiment_id, name, started_at, primary_metric) VALUES (?, ?, ?, ?)`,
		"exp_42", "checkout_button_color", "2026-04-01T00:00:00Z", "checkout_rate",
	)
	require.NoError(t, err)

	// Variant A: 1000 users, mean conv ≈ 0.10. Variant B: 1000 users, mean ≈ 0.13.
	// Designed to give a clean ship recommendation.
	for i := 0; i < 1000; i++ {
		uA := "uA_" + itoa(i)
		uB := "uB_" + itoa(i)
		_, err := db.Exec(`INSERT INTO ab_assignments VALUES (?, ?, ?, ?)`,
			"exp_42", uA, "A", "2026-04-01T00:00:00Z")
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO ab_assignments VALUES (?, ?, ?, ?)`,
			"exp_42", uB, "B", "2026-04-01T00:00:00Z")
		require.NoError(t, err)

		valA := 0.0
		if i < 100 {
			valA = 1.0
		}
		valB := 0.0
		if i < 130 {
			valB = 1.0
		}
		_, err = db.Exec(`INSERT INTO ab_metric_observations VALUES (?, ?, ?, ?, ?, ?)`,
			"exp_42", uA, "checkout_rate", valA, "2026-04-02T00:00:00Z", 1)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO ab_metric_observations VALUES (?, ?, ?, ?, ?, ?)`,
			"exp_42", uB, "checkout_rate", valB, "2026-04-02T00:00:00Z", 1)
		require.NoError(t, err)
	}

	res, err := a.QueryABExperiment(context.Background(), ABExperimentQuery{ExperimentID: "exp_42"})
	require.NoError(t, err)
	require.Equal(t, "exp_42", res.ExperimentID)
	require.Equal(t, "checkout_button_color", res.Name)
	require.Len(t, res.Variants, 2)
	require.True(t, res.Significance.StatisticallySignificant, "p=%.4f", res.Significance.PValue)
	require.Equal(t, "ship", res.Recommendation)
}

func TestSQLAdapter_QueryABExperiment_NotFound(t *testing.T) {
	t.Parallel()
	a, _ := newTestAdapter(t)
	_, err := a.QueryABExperiment(context.Background(), ABExperimentQuery{ExperimentID: "missing"})
	require.ErrorIs(t, err, ErrNotFound)
}

// ─── 4. actor failure ─────────────────────────────────────────────────────

func TestSQLAdapter_QueryActorFailure_ComputesTrustScore(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	// 2 failures over 1 day, mean MTTR = 15min, blast_radius 3 and 5.
	_, err := db.Exec(`INSERT INTO failures (actor_type, occurred_at, recovered_at, blast_radius, recovery_pattern) VALUES (?, ?, ?, ?, ?)`,
		"3rd-party", "2026-04-10T00:00:00Z", "2026-04-10T00:10:00Z", 3, "retry")
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO failures (actor_type, occurred_at, recovered_at, blast_radius, recovery_pattern) VALUES (?, ?, ?, ?, ?)`,
		"3rd-party", "2026-04-10T12:00:00Z", "2026-04-10T12:20:00Z", 5, "circuit_breaker")
	require.NoError(t, err)

	res, err := a.QueryActorFailure(context.Background(), ActorFailureQuery{
		Actor:     "3rd-party",
		TimeRange: TimeRange{From: parseFor(t, "2026-04-10T00:00:00Z"), To: parseFor(t, "2026-04-11T00:00:00Z")},
	})
	require.NoError(t, err)
	require.Equal(t, "3rd-party", res.Actor)
	require.InDelta(t, 2.0, res.FailureRate, 0.01)
	require.InDelta(t, 15.0, res.MTTRMinutes, 0.01)
	require.Equal(t, []string{"circuit_breaker", "retry"}, res.RecoveryPatternsApplied)
	require.Greater(t, res.TrustScore, 0.0)
	require.LessOrEqual(t, res.TrustScore, 100.0)
}

// ─── 5. cost ─────────────────────────────────────────────────────────────

func TestSQLAdapter_QueryCost_DrillDownAndAnomaly(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	insert := func(service string, cents int64, at string) {
		_, err := db.Exec(`INSERT INTO cost_records (service, cost_cents, recorded_at) VALUES (?, ?, ?)`,
			service, cents, at)
		require.NoError(t, err)
	}
	for i := 1; i <= 10; i++ {
		day := zeroPad2(i)
		insert("rds", 1000, "2026-04-"+day+"T00:00:00Z")
		insert("ec2", 2000, "2026-04-"+day+"T00:00:00Z")
	}
	// One spike day on rds: 5000 cents.
	insert("rds", 5000, "2026-04-15T00:00:00Z")

	res, err := a.QueryCost(context.Background(), CostQuery{
		TimeRange:        TimeRange{From: parseFor(t, "2026-04-01T00:00:00Z"), To: parseFor(t, "2026-05-01T00:00:00Z")},
		DrillDown:        CostByService,
		AnomalyDetection: true,
	})
	require.NoError(t, err)
	require.Equal(t, int64(35000), res.TotalCostCents) // 10*1000 + 10*2000 + 5000
	require.Equal(t, int64(15000), res.ByDimension["rds"])
	require.Equal(t, int64(20000), res.ByDimension["ec2"])
	require.NotEmpty(t, res.Anomalies)
	var spiked bool
	for _, a := range res.Anomalies {
		if a.Dimension == "rds" && a.CostCents == 5000 {
			spiked = true
		}
	}
	require.True(t, spiked, "expected rds 5000-cent row flagged as anomaly")
}

// ─── 6. SLO burn ─────────────────────────────────────────────────────────

func TestSQLAdapter_QuerySLOBurn_HotWindowGatesRelease(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	// Three observations inside the 1h window averaging 4000ms — with a target
	// of 250ms that gives burn_rate = 16, which lands in the rollback bucket.
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		_, err := db.Exec(`INSERT INTO slo_observations VALUES (?, ?, ?)`,
			"p99_latency_ms",
			now.Add(-time.Duration(10+i*5)*time.Minute).Format(time.RFC3339),
			4000.0,
		)
		require.NoError(t, err)
	}

	res, err := a.QuerySLOBurn(context.Background(), SLOBurnQuery{
		SLI:        "p99_latency_ms",
		SLOTarget:  250,
		TimeWindow: "1h",
	})
	require.NoError(t, err)
	require.Equal(t, "critical", res.AlertLevel)
	require.Equal(t, "rollback", res.ReleaseGateDecision)
	require.Greater(t, res.BurnRate, 14.0)
	require.InDelta(t, 0.0, res.BudgetRemainingPct, 0.01)
}

func TestSQLAdapter_QuerySLOBurn_NoObservationsReturnsNotFound(t *testing.T) {
	t.Parallel()
	a, _ := newTestAdapter(t)
	_, err := a.QuerySLOBurn(context.Background(), SLOBurnQuery{
		SLI: "missing_sli", SLOTarget: 1, TimeWindow: "1h",
	})
	require.ErrorIs(t, err, ErrNotFound)
}

// ─── 7. feedback corpus ─────────────────────────────────────────────────

func TestSQLAdapter_QueryFeedbackCorpus_NPSBands(t *testing.T) {
	t.Parallel()
	a, db := newTestAdapter(t)

	insert := func(source string, body string, nps int64, topic, sentiment string) {
		_, err := db.Exec(
			`INSERT INTO feedback_items (source, occurred_at, body, nps_score, topic, sentiment) VALUES (?, ?, ?, ?, ?, ?)`,
			source, "2026-04-10T00:00:00Z", body, nps, topic, sentiment,
		)
		require.NoError(t, err)
	}
	insert("nps_comment", "love the speed", 10, "performance", "positive")
	insert("nps_comment", "fast enough", 9, "performance", "positive")
	insert("nps_comment", "ok", 7, "ui", "neutral")
	insert("nps_comment", "slow on mobile", 4, "performance", "negative")
	insert("nps_comment", "buggy checkout", 3, "billing", "negative")

	res, err := a.QueryFeedbackCorpus(context.Background(), FeedbackQuery{
		Source:            FeedbackNPSComment,
		TimeRange:         TimeRange{From: parseFor(t, "2026-04-01T00:00:00Z"), To: parseFor(t, "2026-05-01T00:00:00Z")},
		TopicModeling:     true,
		SentimentAnalysis: true,
	})
	require.NoError(t, err)
	require.Equal(t, int64(5), res.TotalItems)

	// Topic modelling: performance has 3 items (highest).
	require.NotEmpty(t, res.Topics)
	require.Equal(t, "performance", res.Topics[0].Topic)
	require.Equal(t, int64(3), res.Topics[0].ItemCount)
	require.Equal(t, int64(2), res.Topics[0].Sentiment["positive"])
	require.Equal(t, int64(1), res.Topics[0].Sentiment["negative"])

	// NPS bands.
	require.NotNil(t, res.NPSSegments)
	require.Equal(t, int64(2), res.NPSSegments["promoter"].Count)
	require.Equal(t, int64(1), res.NPSSegments["passive"].Count)
	require.Equal(t, int64(2), res.NPSSegments["detractor"].Count)
}

// ─── helpers ─────────────────────────────────────────────────────────────

func parseFor(t *testing.T, s string) time.Time {
	t.Helper()
	tt, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return tt
}

// zeroPad2 renders 1..99 with a leading zero when needed — keeps test seed
// dates as valid RFC3339 even for single-digit days.
func zeroPad2(i int) string {
	if i < 10 {
		return "0" + itoa(i)
	}
	return itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	const digits = "0123456789"
	var buf [20]byte
	pos := len(buf)
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = digits[i%10]
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// sanity-check imports used elsewhere
var _ = strings.Contains
