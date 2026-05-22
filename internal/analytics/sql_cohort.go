package analytics

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)
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
