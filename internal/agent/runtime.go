package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	maxAttempts := 1
	var backoff time.Duration
	if spec.Retry != nil {
		maxAttempts = spec.Retry.MaxAttempts
		backoff = spec.Retry.BackoffDelay
	}

	var stepErr error
	for i, step := range spec.Chain {
		stepResult, err := r.runOneStep(ctx, runID, i, step, maxAttempts, backoff)
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
func (r *Runtime) runOneStep(ctx context.Context, runID int64, idx int, step ChainStep, maxAttempts int, backoff time.Duration) (StepResult, error) {
	var last StepResult
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		_ = r.store.AppendLog(ctx, runID, "info",
			fmt.Sprintf("step[%d] %s attempt=%d", idx, step.Command, attempt))

		stdout, stderr, code, err := r.executor.Run(ctx, step.Command, step.Args)
		last = StepResult{
			Command: NormalizeCommand(step.Command), Args: step.Args, Attempt: attempt,
			ExitCode: code, Stdout: stdout, Stderr: stderr,
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
			_ = r.store.AppendLog(ctx, runID, "info",
				fmt.Sprintf("step[%d] %s ok", idx, step.Command))
			return last, nil
		}

		if attempt < maxAttempts && backoff > 0 {
			select {
			case <-ctx.Done():
				return last, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return last, lastErr
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
