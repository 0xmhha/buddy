package queries

import (
	"context"
	"database/sql"
)

// Test-only re-exports. Keeps failurePercent unexported in the public API
// while still letting external _test packages drive its boundary table directly.
//
// HumanDurForTest used to live here, but the duration formatter moved to
// internal/format and is tested there directly.

// FailurePercentForTest exposes failurePercent for half-up rounding tests.
func FailurePercentForTest(failures, count int64) int {
	return failurePercent(failures, count)
}

// TailFunc mirrors the signature of tailQuery so tests can substitute a fake
// tail implementation via SwapFollowTailForTest. Mirrors the real signature
// 1:1 so a swapped function plugs in without adapter glue.
type TailFunc func(ctx context.Context, conn *sql.DB, hookFilter string, sinceTs, sinceID int64, limit int) ([]Event, error)

// SwapFollowTailForTest swaps the package-private followTailFn that Follow
// invokes on every tick. Returns a restore func the caller defers. Used by
// tests that need to simulate persistent DB failures or count tick calls
// without standing up a real corrupted SQLite file mid-loop.
func SwapFollowTailForTest(fn TailFunc) func() {
	prev := followTailFn
	followTailFn = fn
	return func() { followTailFn = prev }
}
