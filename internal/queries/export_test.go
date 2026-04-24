package queries

// Test-only re-exports. Keeps failurePercent unexported in the public API
// while still letting external _test packages drive its boundary table directly.
//
// HumanDurForTest used to live here, but the duration formatter moved to
// internal/format and is tested there directly.

// FailurePercentForTest exposes failurePercent for half-up rounding tests.
func FailurePercentForTest(failures, count int64) int {
	return failurePercent(failures, count)
}
