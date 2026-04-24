package diagnose

// Test-only re-exports. Lives in *_test.go so it ships only with tests.
// Keeps humanDur unexported in the public API while making it directly
// table-testable from the external diagnose_test package.

// HumanDurForTest exposes humanDur for table-driven boundary tests.
func HumanDurForTest(ms int64) string { return humanDur(ms) }
