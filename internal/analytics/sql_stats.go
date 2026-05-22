package analytics

import (
	"math"
	"sort"
)
func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func variance(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := mean(xs)
	s := 0.0
	for _, x := range xs {
		d := x - m
		s += d * d
	}
	return s / float64(len(xs)-1)
}

func percentile(xs []int64, p float64) int64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]int64(nil), xs...)
	slicesSortInt64(sorted)
	idx := int(math.Round(p * float64(len(sorted)-1)))
	idx = max(idx, 0)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// slicesSortInt64 is a tiny shim over sort.Slice. We keep it as a named
// helper so that swapping to slices.Sort once go.mod targets the appropriate
// Go version is a one-line change.
func slicesSortInt64(xs []int64) {
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
