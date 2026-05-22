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
// `buddy agent run` CLI. Background scheduler, TUI
//, and the reference webtoon agent are subsequent phases.
package agent

import "time"

// Status is the agent's lifecycle state. The set is intentionally narrow —
// future statuses (paused, errored, archived) will land in follow-ons.
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
//                        ticking is implemented separately by the scheduler.
//   - chain            — ordered list of buddy commands to dispatch
//   - retry            — optional retry policy applied uniformly to every step
//   - output           — optional terminal destination (stdout / file). v0.3
//                        writes the JSON result to stdout when omitted.
type AgentSpec struct {
	ID          string              `yaml:"id"`
	Name        string              `yaml:"name"`
	Schedule    string              `yaml:"schedule,omitempty"`
	Chain       []ChainStep         `yaml:"chain"`
	Retry       *RetryPolicy        `yaml:"retry,omitempty"`
	Output      *OutputTarget       `yaml:"output,omitempty"`
	AutoCascade *AutoCascadeConfig  `yaml:"auto_cascade,omitempty"`
	// BranchHints disambiguates conditional next-phase branches at
	// cascade-pick time. Keys are the trimmed LHS prose the parser
	// records in NextPhase.Branches[i].Condition (e.g. "글로벌",
	// "Korea", "USA / EU / 기타"). A true value selects that branch's
	// skills; false explicitly suppresses it. Missing key = unknown
	// preference, which pickCascadeTarget logs and skips rather than
	// silently picking the first one.
	BranchHints map[string]bool `yaml:"branch_hints,omitempty"`
}

// AutoCascadeConfig opts an agent into auto-cascade: after every
// successful step, the runtime looks at the parsed §next-phase block
// (ParsedOutput.NextPhase.Skills) and, if non-empty, appends the first
// listed skill as a new chain step.
//
// A non-nil AutoCascadeConfig enables the feature. The struct can be
// empty (`auto_cascade: {}`) to take the default settings.
//
// Cascading is *strictly* skip-on-failure: a step that exits non-zero
// or returns an executor error does NOT contribute a cascaded successor.
// This avoids the obvious failure-mode of "broken §next-phase cascades
// down a fault path".
type AutoCascadeConfig struct {
	// MaxDepth limits the cascade. Original chain steps are at depth 0;
	// every step appended via §next-phase parsing is at depth N+1.
	// Reaching MaxDepth halts further cascading from that branch (the
	// step still executes). Default 5 when AutoCascadeConfig is non-nil
	// but MaxDepth is zero.
	MaxDepth int `yaml:"max_depth,omitempty"`
}

// DefaultCascadeMaxDepth is the cap applied when auto_cascade is enabled
// without an explicit max_depth. Five covers the §1→§9 happy-path
// orchestrator chain (idea → features → design → plan → build → quality
// → release → operate → lifecycle) with one tier of slack.
const DefaultCascadeMaxDepth = 5

// ChainStep is one rung of an agent's command chain.
type ChainStep struct {
	Command string `yaml:"command"`        // buddy command name (no /buddy: prefix)
	Args    string `yaml:"args,omitempty"` // free-form argument string passed to the command
	// ContinueOnFail, when true, lets the runtime move on to the next
	// step even if this step exhausts its retry budget without success.
	// The failed step is still recorded in RunResult.Steps and the
	// per-step exit code is preserved; what changes is that the chain
	// does NOT short-circuit. Useful for cleanup or best-effort
	// notification steps (e.g. "post-completion webhook" that should
	// not block downstream "release-tag" if it fails).
	//
	// Default (false) preserves the v0.6.x behaviour: first non-zero
	// step after retries stops the chain. The final RunResult.ExitCode
	// is the *last* non-zero step's exit when any step failed (so a
	// single failed cleanup step at the end still surfaces failure),
	// or 0 when every step (including continue_on_fail ones) succeeded.
	ContinueOnFail bool `yaml:"continue_on_fail,omitempty"`
	// OnSelfCheckFail is the policy knob for what the runtime does
	// when the step's parsed §self-check section reports a fail verdict
	// (any unchecked `- [ ]` in the section body). Allowed values:
	//
	//   - "" (default, equivalent to "continue"): runtime preserves the
	//     v0.5.0+ behaviour — the verdict lands in the per-step log line
	//     but the chain continues. RunResult.Steps still records the
	//     verdict for downstream consumers.
	//   - "continue": same as the empty-string default; explicit form.
	//   - "abort": the cascade short-circuits as if the step itself had
	//     failed. RunResult.ExitCode reflects the failure, the cascade
	//     stops here, and ContinueOnFail still applies — a step marked
	//     ContinueOnFail+abort will skip the rest of the chain but the
	//     overall run is not failed by *this* step's verdict alone.
	//   - "retry": treat the verdict as a transient failure and trigger
	//     the agent's RetryPolicy. Bounded by MaxAttempts like exit-code
	//     failures, so a deterministic self-check failure cannot loop
	//     forever.
	//
	// Default ("") preserves the prior behaviour; opt-in is the
	// discipline.
	OnSelfCheckFail string `yaml:"on_self_check_fail,omitempty"`
}

// SelfCheckFailPolicy* are the allowed values of ChainStep.OnSelfCheckFail.
// Validate() in spec.go enforces this enum so a typo in the YAML lands
// as a parse error rather than silently being ignored at runtime.
const (
	SelfCheckFailContinue = "continue"
	SelfCheckFailAbort    = "abort"
	SelfCheckFailRetry    = "retry"
)

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

// OutputTarget describes where the final aggregated result goes. The
// initial release ships stdout, file, and webhook destinations (the spec
// §2.2 webtoon example is the canonical webhook case).
type OutputTarget struct {
	Type string `yaml:"type"`           // "stdout" | "file" | "webhook"
	Path string `yaml:"path,omitempty"` // file path when Type=="file"

	// URL is the POST target when Type=="webhook". Required for that
	// type; ignored otherwise.
	URL string `yaml:"url,omitempty"`
	// Method is the HTTP method for Type=="webhook". Defaults to POST.
	// Useful when the target expects PUT (idempotent uploads) or PATCH.
	Method string `yaml:"method,omitempty"`
	// Headers are extra HTTP headers sent on the webhook request. The
	// runtime sets Content-Type=application/json unless the spec
	// overrides it here. Authorization, X-API-Key, etc. live here.
	// Header values are written literally — secret expansion (e.g.
	// `${BUDDY_TOKEN}`) is intentionally not auto-applied; if a user
	// needs secrets, they should template the spec before `buddy agent
	// create` rather than commit them to a YAML file.
	Headers map[string]string `yaml:"headers,omitempty"`
	// Timeout is the per-request HTTP timeout for Type=="webhook".
	// Defaults to 30s. Long uploads should bump this explicitly.
	Timeout time.Duration `yaml:"timeout,omitempty"`
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
	// CascadeDepth is 0 for original chain steps and N+1 for steps
	// appended via auto-cascade from a step at depth N. Lets downstream
	// callers distinguish "user wrote this step" from "the runtime
	// inferred this step from §next-phase". Always 0 when AutoCascade
	// is disabled.
	CascadeDepth int `json:"cascade_depth,omitempty"`
}
