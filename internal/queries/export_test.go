package queries

// Test-only re-exports. Keeps humanDur / failurePercent unexported in the
// public API while still letting external _test packages drive their boundary
// tables directly.

// HumanDurForTest exposes humanDur for table-driven boundary tests.
func HumanDurForTest(ms int64) string { return humanDur(ms) }

// FailurePercentForTest exposes failurePercent for half-up rounding tests.
func FailurePercentForTest(failures, count int64) int {
	return failurePercent(failures, count)
}
