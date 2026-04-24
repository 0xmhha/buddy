package aggregator

import "time"

// nowMs is a tiny indirection so tests could pin time later if needed.
func nowMs() int64 { return time.Now().UnixMilli() }
