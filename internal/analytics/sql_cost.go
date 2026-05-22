package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"
)
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
