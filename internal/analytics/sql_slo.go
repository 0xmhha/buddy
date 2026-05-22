package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)
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
