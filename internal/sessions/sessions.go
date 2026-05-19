// Package sessions exposes the active Claude Code sessions buddy observes,
// reusing the v0.1 schema's TokenUsage shape so transcript-derived counters
// land in the same fields the hook event pipeline already speaks.
//
// Scope (v0.2): Session struct + Lister interface only. The fsLister
// implementation lives in the sibling file added by core-3.
package sessions

import (
	"context"
	"time"

	"github.com/0xmhha/buddy/internal/schema"
)

// Session represents an active Claude Code session as observed by buddy.
// Fields map 1:1 to the v0.2 sessions table (internal/db/migrations.go
// version 3, extended by v5 — ADR-012):
//
//   - ID              — canonical session id surfaced by Claude Code.
//   - PID             — process id when observable, zero when not.
//   - TranscriptPath  — absolute path to the session's .jsonl transcript.
//   - StartedAt       — first time buddy noticed the session.
//   - LastActive      — last time the transcript advanced.
//   - Usage           — running totals for the session, as far as buddy
//                       has tailed the transcript.
//   - LastOffset      — byte offset the transcript reader resumes from
//                       on the next tick. Persisting it in the row makes
//                       the reader idempotent across daemon restarts.
//   - EndedAt         — nil while session may still be active. Daemon
//                       sets it when LastActive crosses EndedThreshold
//                       (default 1h). Cleared on resume — no row split.
//                       Added by migration v5 (ADR-012).
//   - GoalText        — first user message extracted by fsLister on
//                       the first tail pass. F2.D Drift Detection
//                       (W7-4) compares current activity against this.
//                       "" when no user message yet or extraction
//                       skipped (e.g., first message is /clear).
//                       Added by migration v5 (ADR-012).
//   - Metadata        — JSON object for future per-session fields
//                       without further migrations (project_root,
//                       claude_version, custom tags). "{}" default.
//                       Added by migration v5 (ADR-012).
type Session struct {
	ID             string
	PID            int
	TranscriptPath string
	StartedAt      time.Time
	LastActive     time.Time
	Usage          schema.TokenUsage
	LastOffset     int64
	EndedAt        *time.Time
	GoalText       string
	Metadata       string
}

// Lister enumerates the active Claude Code sessions visible to buddy.
//
// Implementations differ in how they discover sessions — v0.2 has a single
// filesystem-scanning implementation (fsLister, core-3); future
// implementations are deferred to the plan's BUDDY-D-MULTI-IMPL trigger
// (see docs/superpowers/plans/2026-05-09-v02-control-plane-plan.md §3.4).
// Every implementation returns the same Session shape so callers (daemon,
// CLI, dashboard) stay backend-agnostic.
//
// List returns the snapshot of currently active sessions as of the call.
// Callers should not assume freshness stronger than "snapshot at return";
// implementations are free to cache between calls.
type Lister interface {
	List(ctx context.Context) ([]Session, error)
}
