package aggregator_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/aggregator"
	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

const baseTs = int64(1_700_000_000_000) // a known instant

func openTmp(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "buddy.db")
}

func TestBucketStartMs_AlignsToWindowBoundary(t *testing.T) {
	// 5-min window: bucket aligned to 5-minute boundary
	got := aggregator.BucketStartMs(baseTs+123_456, 5)
	// floor to 5-minute boundary in minutes since epoch
	expectedMin := ((baseTs + 123_456) / 60_000 / 5) * 5
	assert.Equal(t, expectedMin*60_000, got)

	// 60-min window
	got = aggregator.BucketStartMs(baseTs+999_999, 60)
	expectedMin = ((baseTs + 999_999) / 60_000 / 60) * 60
	assert.Equal(t, expectedMin*60_000, got)
}

func TestProcessBatch_NoOpWhenEmpty(t *testing.T) {
	conn, err := db.Open(db.Options{Path: openTmp(t)})
	require.NoError(t, err)
	defer conn.Close()

	n, err := aggregator.ProcessBatch(conn, 100)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestProcessBatch_DrainsOutboxIntoEvents(t *testing.T) {
	path := openTmp(t)
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	defer conn.Close()

	// enqueue 3 payloads
	for i := range 3 {
		p := &schema.HookEventPayload{
			Ts:         baseTs + int64(i)*1000,
			Event:      schema.EventPostToolUse,
			HookName:   "hook-x",
			ToolName:   "Bash",
			DurationMs: int64(100 + i*10),
			ExitCode:   0,
		}
		_, err := db.AppendToOutbox(conn, p)
		require.NoError(t, err)
	}

	n, err := aggregator.ProcessBatch(conn, 100)
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	// outbox is now empty (consumed)
	pending, err := db.ReadPendingOutbox(conn, 100)
	require.NoError(t, err)
	assert.Empty(t, pending)

	// 3 hook_events rows present
	var eventCount int
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM hook_events").Scan(&eventCount))
	assert.Equal(t, 3, eventCount)
}

func TestProcessBatch_ComputesStatsForAllWindows(t *testing.T) {
	conn, err := db.Open(db.Options{Path: openTmp(t)})
	require.NoError(t, err)
	defer conn.Close()

	// 5 events in the same 5-min bucket, with known durations
	durations := []int64{10, 20, 30, 40, 50}
	for i, d := range durations {
		_, err := db.AppendToOutbox(conn, &schema.HookEventPayload{
			Ts:         baseTs + int64(i)*1000,
			Event:      schema.EventPostToolUse,
			HookName:   "h1",
			ToolName:   "Read",
			DurationMs: d,
			ExitCode:   0,
		})
		require.NoError(t, err)
	}

	n, err := aggregator.ProcessBatch(conn, 100)
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	// stats present for all 3 windows
	for _, w := range aggregator.StatsWindowsMin {
		var count, failures int64
		var p50, p95, p99 int64
		err := conn.QueryRow(`
			SELECT count, failures, p50_ms, p95_ms, p99_ms
			FROM hook_stats
			WHERE hook_name = ? AND tool_name = ? AND window_min = ?
		`, "h1", "Read", w).Scan(&count, &failures, &p50, &p95, &p99)
		require.NoError(t, err, "window %dmin", w)
		assert.Equal(t, int64(5), count)
		assert.Equal(t, int64(0), failures)
		assert.Equal(t, int64(30), p50, "p50 of [10,20,30,40,50] = 30")
		assert.GreaterOrEqual(t, p95, int64(40))
		assert.LessOrEqual(t, p95, int64(50))
	}
}

func TestProcessBatch_TracksFailures(t *testing.T) {
	conn, err := db.Open(db.Options{Path: openTmp(t)})
	require.NoError(t, err)
	defer conn.Close()

	// 4 events, 2 failing
	for i, code := range []int{0, 1, 0, 1} {
		_, err := db.AppendToOutbox(conn, &schema.HookEventPayload{
			Ts:         baseTs + int64(i)*1000,
			Event:      schema.EventStop,
			HookName:   "h2",
			DurationMs: 50,
			ExitCode:   code,
		})
		require.NoError(t, err)
	}

	_, err = aggregator.ProcessBatch(conn, 100)
	require.NoError(t, err)

	var count, failures int64
	err = conn.QueryRow(`
		SELECT count, failures FROM hook_stats
		WHERE hook_name = ? AND window_min = 5
	`, "h2").Scan(&count, &failures)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)
	assert.Equal(t, int64(2), failures)
}

func TestProcessBatch_PartitionsByToolName(t *testing.T) {
	conn, err := db.Open(db.Options{Path: openTmp(t)})
	require.NoError(t, err)
	defer conn.Close()

	for _, tool := range []string{"Bash", "Bash", "Read", "Read", "Read"} {
		_, err := db.AppendToOutbox(conn, &schema.HookEventPayload{
			Ts: baseTs, Event: schema.EventPreToolUse,
			HookName: "h3", ToolName: tool, DurationMs: 1, ExitCode: 0,
		})
		require.NoError(t, err)
	}

	_, err = aggregator.ProcessBatch(conn, 100)
	require.NoError(t, err)

	rows, err := conn.Query(`
		SELECT tool_name, count FROM hook_stats
		WHERE hook_name = ? AND window_min = 5 ORDER BY tool_name
	`, "h3")
	require.NoError(t, err)
	defer rows.Close()

	got := map[string]int64{}
	for rows.Next() {
		var tool string
		var count int64
		require.NoError(t, rows.Scan(&tool, &count))
		got[tool] = count
	}
	assert.Equal(t, int64(2), got["Bash"])
	assert.Equal(t, int64(3), got["Read"])
}

func TestProcessBatch_HandlesMalformedPayload(t *testing.T) {
	conn, err := db.Open(db.Options{Path: openTmp(t)})
	require.NoError(t, err)
	defer conn.Close()

	// directly inject malformed JSON
	_, err = conn.Exec(
		"INSERT INTO hook_outbox (ts, payload) VALUES (?, ?)",
		baseTs, "{not valid json}",
	)
	require.NoError(t, err)

	n, err := aggregator.ProcessBatch(conn, 100)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "malformed row still counted as processed")

	// no hook_events written
	var c int
	require.NoError(t, conn.QueryRow("SELECT COUNT(*) FROM hook_events").Scan(&c))
	assert.Equal(t, 0, c)

	// outbox row marked consumed (won't block queue)
	pending, err := db.ReadPendingOutbox(conn, 100)
	require.NoError(t, err)
	assert.Empty(t, pending)
}
