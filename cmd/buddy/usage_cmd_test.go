package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/usage"
)

// TestRenderDailyBarChart_HeaderAndScaling locks the bar-chart output:
// a header with the peak total, one row per day, and bar widths
// proportional to the day's share of the peak. The renderer must never
// crash on empty or all-zero input; both paths return "".
func TestRenderDailyBarChart_HeaderAndScaling(t *testing.T) {
	day := func(d int, total int64) usage.DailySpend {
		// Build a series where InputTokens carries the full total so
		// TotalTokens() matches the intended value without spreading
		// across the four token categories.
		return usage.DailySpend{
			Date:        time.Date(2026, 5, 17+d, 0, 0, 0, 0, time.UTC),
			InputTokens: total,
		}
	}

	t.Run("empty series returns empty string", func(t *testing.T) {
		assert.Equal(t, "", renderDailyBarChart(nil))
		assert.Equal(t, "", renderDailyBarChart([]usage.DailySpend{}))
	})

	t.Run("all-zero series returns empty string", func(t *testing.T) {
		assert.Equal(t, "", renderDailyBarChart([]usage.DailySpend{day(0, 0), day(1, 0)}))
	})

	t.Run("scales bars to peak and renders header", func(t *testing.T) {
		out := renderDailyBarChart([]usage.DailySpend{
			day(0, 100),
			day(1, 50),
			day(2, 0),
		})
		require.Contains(t, out, "일별 토큰 추이 (peak 100)")

		// Three data rows: peak day full 30 cells, half day 15 cells,
		// zero day zero cells.
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		require.Len(t, lines, 4, "header + 3 days")

		assert.Contains(t, lines[1], "100")
		assert.Equal(t, 30, strings.Count(lines[1], "█"), "peak day = 30 cells")

		assert.Contains(t, lines[2], "50")
		assert.Equal(t, 15, strings.Count(lines[2], "█"), "half day = 15 cells")

		assert.Contains(t, lines[3], "0")
		assert.Equal(t, 0, strings.Count(lines[3], "█"), "zero day = no cells")
	})
}
