// Package agent implements the cli buddy automation-agent runtime.
//
// Per cli-buddy-spec §3 / §4 (locked in by ADR-005), an agent is a static
// definition (id + name + chain of buddy commands + schedule) whose runtime
// shape is:
//
//   agent.Run() = for each step in spec.Chain:
//                  spawn `claude` subprocess
//                  send "/buddy:<command> <args>"
//                  capture stdout, parse the PROCEDURE output
//                  record run/log rows in SQLite
//
// v0.3.x ships the storage + runtime + Subprocess/Mock executor + on-demand
// `buddy agent run` CLI. Background scheduler (W3-3 follow-on), TUI
// (W3-2), and the reference webtoon agent (W3-6) are subsequent phases.
package agent

import "time"

// Status is the agent's lifecycle state. It is intentionally narrow in v0.3 —
// future statuses (paused, errored, archived) land in W3-3 follow-ons.
type Status string

const (
	StatusIdle    Status = "idle"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

// Agent is the persistent definition stored in the agents table.
type Agent struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	SpecYAML   string    `json:"spec_yaml"`     // original YAML, kept for round-trip / TUI editing
	Schedule   string    `json:"schedule"`      // cron expression, empty = on-demand only
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastRunAt  *time.Time `json:"last_run_at,omitempty"`
}

// AgentSpec is the parsed YAML form. See cli-buddy-spec §2.2 for the canonical
// "webtoon agent" example. v0.3 supports a subset:
//   - id / name        — required
//   - schedule         — optional cron expression (currently informational; the
//                        on-demand `buddy agent run` ignores it). Background
//                        ticking lands in the W3-3 follow-on.
//   - chain            — ordered list of buddy commands to dispatch
//   - retry            — optional retry policy applied uniformly to every step
//   - output           — optional terminal destination (stdout / file). v0.3
//                        writes the JSON result to stdout when omitted.
type AgentSpec struct {
	ID       string         `yaml:"id"`
	Name     string         `yaml:"name"`
	Schedule string         `yaml:"schedule,omitempty"`
	Chain    []ChainStep    `yaml:"chain"`
	Retry    *RetryPolicy   `yaml:"retry,omitempty"`
	Output   *OutputTarget  `yaml:"output,omitempty"`
}

// ChainStep is one rung of an agent's command chain.
type ChainStep struct {
	Command string `yaml:"command"`        // buddy command name (no /buddy: prefix)
	Args    string `yaml:"args,omitempty"` // free-form argument string passed to the command
}

// RetryPolicy is a uniform retry config for every step. v0.3 shipped a
// fixed-delay capped retry; v0.6.x adds exponential backoff with an
// optional cap (BackoffMax).
type RetryPolicy struct {
	MaxAttempts int `yaml:"max_attempts"`
	// BackoffDelay is the *base* sleep between retry attempts. With
	// BackoffStrategy="fixed" (or unset, for v0.6.x backward compat)
	// every retry waits exactly this duration. With
	// BackoffStrategy="exponential" it is the delay after the *first*
	// failure (attempt 1 fail → wait BackoffDelay → retry as attempt 2);
	// subsequent failures double the wait up to BackoffMax.
	BackoffDelay time.Duration `yaml:"backoff_delay,omitempty"`
	// BackoffStrategy selects how BackoffDelay grows across retries.
	// Empty string is treated as "fixed" so existing specs keep their
	// v0.3+ behavior verbatim. Valid values: "fixed" | "exponential".
	BackoffStrategy string `yaml:"backoff_strategy,omitempty"`
	// BackoffMax caps the exponential growth. Zero means uncapped (the
	// growth still terminates when MaxAttempts is reached). Has no
	// effect when BackoffStrategy="fixed".
	BackoffMax time.Duration `yaml:"backoff_max,omitempty"`
}

const (
	// BackoffStrategyFixed makes every retry wait BackoffDelay (the v0.3+
	// default — explicit constant so callers can name it).
	BackoffStrategyFixed = "fixed"
	// BackoffStrategyExponential doubles the wait each failed attempt,
	// capped at BackoffMax (0 = uncapped).
	BackoffStrategyExponential = "exponential"
)

// OutputTarget describes where the final aggregated result goes. v0.3 ships
// stdout and file targets — webhook / API endpoints (spec §2.2 webtoon
// example) land in W3-6.
type OutputTarget struct {
	Type string `yaml:"type"`           // "stdout" | "file"
	Path string `yaml:"path,omitempty"` // file path when Type=="file"
}

// AgentRun is one execution record. Persisted to agent_runs.
type AgentRun struct {
	ID         int64      `json:"id"`
	AgentID    string     `json:"agent_id"`
	StartedAt  time.Time  `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	ExitCode   int        `json:"exit_code"`
	Error      string     `json:"error,omitempty"`
	ResultJSON string     `json:"result_json"`
}

// AgentLog is one line-level log event. Persisted to agent_logs.
type AgentLog struct {
	ID      int64     `json:"id"`
	RunID   int64     `json:"run_id"`
	Ts      time.Time `json:"ts"`
	Level   string    `json:"level"`   // "info" | "warn" | "error" | "debug"
	Message string    `json:"message"`
}

// StepResult is the captured output of a single chain step. The Runtime
// aggregates these into AgentRun.ResultJSON.
type StepResult struct {
	Command  string       `json:"command"`
	Args     string       `json:"args,omitempty"`
	Attempt  int          `json:"attempt"`
	ExitCode int          `json:"exit_code"`
	Stdout   string       `json:"stdout,omitempty"`
	Stderr   string       `json:"stderr,omitempty"`
	Error    string       `json:"error,omitempty"`
	Parsed   ParsedOutput `json:"parsed"`
}
