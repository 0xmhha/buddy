// Package format holds tiny, language-neutral display helpers shared by the
// user-facing report packages (diagnose, queries, and future events readers).
//
// These used to live as package-local copies in internal/diagnose and
// internal/queries; they were already drifting (one took int, the other int64)
// when extracted. Centralising them is the only way to keep boundary behaviour
// consistent across reports.
//
// Scope rules:
//   - Pure functions only, no IO, no globals.
//   - No domain knowledge — this package must not import anything from
//     internal/* so report packages can depend on it freely without cycles.
package format

import (
	"strconv"
	"strings"
)

// Duration renders ms as a human-friendly duration. Boundaries:
//
//	[0, 1000)      → "<n>ms"
//	[1000, 60000)  → "<n.n>s"   (rounded to nearest 0.1s)
//	[60000, ∞)     → "<n.n>m"
//
// Bucket selection runs on integer-rounded tenths-of-a-second, not on the
// formatted float, so the [s → m] boundary cannot straddle units. e.g.
// 59,999ms rounds to 60.0s of tenths and is reported as "1.0m" rather than
// "60.0s". We fix one decimal for s/m so columns line up in lists.
//
// Negative inputs are clamped to 0 — the caller's contract is "milliseconds
// elapsed" and a negative duration would only mean upstream measurement bugs;
// rendering "-3ms" would push that confusion onto the user.
func Duration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	if ms < 1000 {
		return strconv.FormatInt(ms, 10) + "ms"
	}
	// Round to nearest 0.1s. tenths >= 600 means the rendered seconds-value
	// would be >= 60.0s, which we promote to the minute branch instead.
	tenthsOfSec := (ms + 50) / 100
	if tenthsOfSec < 600 {
		return strconv.FormatFloat(float64(tenthsOfSec)/10.0, 'f', 1, 64) + "s"
	}
	return strconv.FormatFloat(float64(ms)/60_000.0, 'f', 1, 64) + "m"
}

// Thousands inserts commas as thousands separators. Pure helper, no locale
// awareness — v0.1 only ships ko/en, both of which are fine with commas.
//
// Takes int64 so callers with int values cast at the call site; we picked
// int64 over generics because the only types in play are int and int64 and
// adding constraints buys us nothing for v0.1.
func Thousands(n int64) string {
	if n < 0 {
		return "-" + Thousands(-n)
	}
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if len(s) > pre {
			b.WriteByte(',')
		}
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}
