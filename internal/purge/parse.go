package purge

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// relativeDays matches "<N>d" where N is a non-negative integer. Anchored
// both ends so "7days" / "-7d" / " 7d " are rejected.
var relativeDays = regexp.MustCompile(`^(\d+)d$`)

// ParseBefore turns one of the user-friendly forms into a UTC time:
//
//   - "2026-04-01"           → midnight UTC of that day
//   - "2026-04-01T00:00:00Z" → literal RFC3339 (any tz; preserved as-is)
//   - "7d"                   → now - 7 days
//
// `now` is injected so tests are deterministic. The cmd layer passes
// time.Now().UTC().
//
// Returned errors are English / machine-shaped on purpose: the cmd layer is
// the i18n boundary and renders friend-tone Korean. See cmd/buddy/purge_cmd.go.
func ParseBefore(s string, now time.Time) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty value")
	}
	if m := relativeDays.FindStringSubmatch(s); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			// Unreachable given the regex, but keep the error path explicit.
			return time.Time{}, fmt.Errorf("parse relative days %q: %w", s, err)
		}
		return now.AddDate(0, 0, -n), nil
	}
	// RFC3339 is checked before the bare-date form so that a user passing
	// "2026-04-01T12:34:56Z" gets the precise time, not a midnight rounding.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		// time.Parse("2006-01-02", ...) returns UTC midnight by default.
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognized date %q (try 30d, 2026-01-01, or 2026-01-01T00:00:00Z)", s)
}
