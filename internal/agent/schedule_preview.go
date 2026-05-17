package agent

import (
	"time"

	"github.com/robfig/cron/v3"
)

// SchedulePreview is the read-only "if a scheduler were running, this is when
// the agent would fire next" snapshot used by the TUI scheduler-status pane
// and any future `buddy agent scheduler preview` subcommand. It is decoupled
// from a live *Scheduler — callers do not need to start cron to compute it.
//
// Err captures parse failures (e.g. a typo'd cron string committed to a YAML
// spec). Surfacing it inline lets the TUI render the friendly "(invalid:
// <msg>)" copy next to the offending row rather than collapsing the whole
// pane into an error state.
type SchedulePreview struct {
	Schedule string
	Next     time.Time
	Err      error
}

// PreviewSchedule parses one cron string with the same parser flags the
// scheduler uses (cron.ParseStandard — 5-field cron + descriptors like
// @daily / @every 30s; second-precision is intentionally off, mirroring the
// scheduler.go comment).
//
// An empty schedule string returns a zero SchedulePreview (no error) — the
// caller can interpret that as "on-demand, never fires on its own".
//
// `now` is taken explicitly rather than calling time.Now() inside so the
// helper is deterministic and unit-testable.
func PreviewSchedule(schedule string, now time.Time) SchedulePreview {
	if schedule == "" {
		return SchedulePreview{Schedule: schedule}
	}
	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return SchedulePreview{Schedule: schedule, Err: err}
	}
	return SchedulePreview{
		Schedule: schedule,
		Next:     sched.Next(now),
	}
}
