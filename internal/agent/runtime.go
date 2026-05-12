package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Runtime is the engine that walks an agent's spec.Chain step-by-step,
// delegating each step to the Executor, and persists the result to the
// Store. One Runtime instance is safe to reuse across many Run calls.
type Runtime struct {
	store    *Store
	executor Executor
}

// NewRuntime wires a Runtime to its dependencies. SubprocessExecutor is the
// production default; tests pass a MockExecutor.
func NewRuntime(store *Store, executor Executor) *Runtime {
	return &Runtime{store: store, executor: executor}
}

// RunResult is the aggregated outcome of a Run call. It is also what gets
// JSON-serialised into agent_runs.result_json.
type RunResult struct {
	RunID    int64        `json:"run_id"`
	AgentID  string       `json:"agent_id"`
	Steps    []StepResult `json:"steps"`
	ExitCode int          `json:"exit_code"`
}

// Run executes the agent's chain. Each step is retried up to spec.Retry.MaxAttempts
// (default 1, i.e. no retry) with an optional fixed backoff between attempts.
// Run is synchronous; cancellation honours ctx (the in-flight subprocess
// receives the cancel via exec.CommandContext).
//
// Semantics:
//   - First non-zero exit code OR executor error in any step short-circuits
//     the chain. Subsequent steps are not invoked.
//   - The output target (stdout / file) receives the JSON-serialised
//     RunResult after the chain finishes (regardless of success/failure).
//   - The agents row is transitioned: idle → running → done|failed.
func (r *Runtime) Run(ctx context.Context, agent Agent) (RunResult, error) {
	spec, err := ParseSpec([]byte(agent.SpecYAML))
	if err != nil {
		return RunResult{}, fmt.Errorf("runtime: parse agent.SpecYAML: %w", err)
	}

	if err := r.store.UpdateStatus(ctx, agent.ID, StatusRunning); err != nil {
		return RunResult{}, err
	}
	runID, err := r.store.StartRun(ctx, agent.ID)
	if err != nil {
		_ = r.store.UpdateStatus(ctx, agent.ID, StatusFailed)
		return RunResult{}, err
	}

	result := RunResult{RunID: runID, AgentID: agent.ID}
	_ = r.store.AppendLog(ctx, runID, "info", fmt.Sprintf("agent %q started (%d steps)", agent.ID, len(spec.Chain)))

	retry := spec.Retry

	var stepErr error
	for i, step := range spec.Chain {
		stepResult, err := r.runOneStep(ctx, runID, i, step, retry)
		result.Steps = append(result.Steps, stepResult)
		if err != nil || stepResult.ExitCode != 0 {
			stepErr = err
			result.ExitCode = stepResult.ExitCode
			if result.ExitCode == 0 {
				result.ExitCode = 1 // executor error without exit code = generic failure
			}
			break
		}
	}

	finalStatus := StatusDone
	if stepErr != nil || result.ExitCode != 0 {
		finalStatus = StatusFailed
	}

	if err := r.store.FinishRun(ctx, runID, result.ExitCode, stepErr, result); err != nil {
		_ = r.store.UpdateStatus(ctx, agent.ID, StatusFailed)
		return result, err
	}
	if err := r.store.UpdateStatus(ctx, agent.ID, finalStatus); err != nil {
		return result, err
	}

	if spec.Output != nil {
		if err := writeOutput(spec.Output, result); err != nil {
			_ = r.store.AppendLog(ctx, runID, "warn", fmt.Sprintf("write output: %v", err))
		}
	}

	_ = r.store.AppendLog(ctx, runID, "info", fmt.Sprintf("agent %q finished status=%s exit=%d", agent.ID, finalStatus, result.ExitCode))
	return result, stepErr
}

// runOneStep handles the retry loop for a single chain step. It records every
// attempt as its own log line so post-hoc analysis sees the full picture.
// The retry argument may be nil — that's interpreted as "no retry" (one
// attempt, no backoff). When non-nil, computeBackoff selects the sleep
// duration for each retry based on the policy's strategy + base + max.
func (r *Runtime) runOneStep(ctx context.Context, runID int64, idx int, step ChainStep, retry *RetryPolicy) (StepResult, error) {
	maxAttempts := 1
	if retry != nil {
		maxAttempts = retry.MaxAttempts
	}
	var last StepResult
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		_ = r.store.AppendLog(ctx, runID, "info",
			fmt.Sprintf("step[%d] %s attempt=%d", idx, step.Command, attempt))

		stdout, stderr, code, err := r.executor.Run(ctx, step.Command, step.Args)
		last = StepResult{
			Command: NormalizeCommand(step.Command), Args: step.Args, Attempt: attempt,
			ExitCode: code, Stdout: stdout, Stderr: stderr,
			Parsed: ParseClaudeOutput(stdout),
		}
		lastErr = err
		if err != nil {
			last.Error = err.Error()
			_ = r.store.AppendLog(ctx, runID, "error",
				fmt.Sprintf("step[%d] %s attempt=%d error: %v", idx, step.Command, attempt, err))
		} else if code != 0 {
			_ = r.store.AppendLog(ctx, runID, "warn",
				fmt.Sprintf("step[%d] %s attempt=%d exit=%d", idx, step.Command, attempt, code))
		} else {
			// Surface the parsed §self-check verdict in the run log so
			// `buddy agent log <id>` (future) and live tail show the
			// quality signal alongside the exit code. v0.3 does NOT
			// flip step success to "failure" based on the verdict —
			// that's a follow-on once we have dogfood signal that the
			// LLM consistently fills the checkboxes.
			if last.Parsed.SelfCheck.Verdict != SelfCheckUnknown {
				_ = r.store.AppendLog(ctx, runID, "info",
					fmt.Sprintf("step[%d] %s self-check=%s (%d/%d passed)",
						idx, step.Command, last.Parsed.SelfCheck.Verdict,
						last.Parsed.SelfCheck.Passed, last.Parsed.SelfCheck.Total))
			}
			if len(last.Parsed.NextPhase.Skills) > 0 {
				_ = r.store.AppendLog(ctx, runID, "info",
					fmt.Sprintf("step[%d] %s next-phase candidates: %s",
						idx, step.Command, strings.Join(last.Parsed.NextPhase.Skills, ", ")))
			}
			if len(last.Parsed.NextPhase.Branches) > 0 {
				// One log line per branch keeps each conditional rule on
				// its own grep-able line — important for the future
				// cascade engine, which will pick *one* branch based on
				// the run's environment / inputs rather than fanning out
				// to the Skills union.
				for _, b := range last.Parsed.NextPhase.Branches {
					rhs := strings.Join(b.Skills, ", ")
					if rhs == "" {
						rhs = "(no skill)"
					}
					_ = r.store.AppendLog(ctx, runID, "info",
						fmt.Sprintf("step[%d] %s next-phase branch: %q → %s",
							idx, step.Command, b.Condition, rhs))
				}
			}
			_ = r.store.AppendLog(ctx, runID, "info",
				fmt.Sprintf("step[%d] %s ok", idx, step.Command))
			return last, nil
		}

		if attempt < maxAttempts {
			sleep := computeBackoff(retry, attempt)
			if sleep <= 0 {
				continue
			}
			_ = r.store.AppendLog(ctx, runID, "info",
				fmt.Sprintf("step[%d] %s backoff %s before attempt %d",
					idx, step.Command, sleep, attempt+1))
			select {
			case <-ctx.Done():
				return last, ctx.Err()
			case <-time.After(sleep):
			}
		}
	}
	return last, lastErr
}

// computeBackoff selects the wait duration before the next retry of a step,
// given which attempt just failed (1-indexed). Returns 0 to skip the wait
// (no policy / zero base / unreachable retry count).
//
// "fixed" (default) returns BackoffDelay verbatim — every retry waits the
// same. "exponential" doubles per failed attempt, capped at BackoffMax when
// BackoffMax > 0. The doubling is implemented as a *Duration multiply*
// (BackoffDelay * 2^k) rather than a shift, with an overflow-safe cap: the
// cap kicks in well before time.Duration's int64 range overflows in
// realistic configurations (BackoffMax ≤ a few minutes).
func computeBackoff(retry *RetryPolicy, attemptJustFailed int) time.Duration {
	if retry == nil || retry.BackoffDelay <= 0 || attemptJustFailed < 1 {
		return 0
	}
	switch retry.BackoffStrategy {
	case "", BackoffStrategyFixed:
		return retry.BackoffDelay
	case BackoffStrategyExponential:
		// attempt 1 fail → multiplier 1 (BackoffDelay)
		// attempt 2 fail → multiplier 2 (2 × BackoffDelay)
		// attempt 3 fail → multiplier 4 (4 × BackoffDelay)
		// Cap the shift exponent so a misconfigured very-large MaxAttempts
		// can't push 2^k past int64 wraparound. 30 doublings = 2^30 ≈ 1e9;
		// even with BackoffDelay=1s that's already ~34 years, well beyond
		// any realistic BackoffMax.
		exp := min(attemptJustFailed-1, 30)
		multiplier := time.Duration(1) << exp
		wait := retry.BackoffDelay * multiplier
		if retry.BackoffMax > 0 && wait > retry.BackoffMax {
			return retry.BackoffMax
		}
		return wait
	}
	// Unknown strategy passes validation if and only if spec.go missed it.
	// Falling through to the base delay is safer than panicking inside the
	// retry loop.
	return retry.BackoffDelay
}

func writeOutput(target *OutputTarget, result RunResult) error {
	switch target.Type {
	case "", "stdout":
		// Stdout output is the CLI subcommand's responsibility (it has the
		// real os.Stdout). Runtime treats stdout as the implicit default and
		// emits no side effect here.
		return nil
	case "file":
		if target.Path == "" {
			return errors.New("agent: output.path empty")
		}
		// Best-effort mkdir for relative or nested target paths so users can
		// point at "logs/{id}/result.json" style outputs without setup.
		if dir := filepath.Dir(target.Path); dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", dir, err)
			}
		}
		f, err := os.Create(target.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		return writeJSON(f, result)
	}
	return fmt.Errorf("agent: unsupported output.type %q", target.Type)
}
