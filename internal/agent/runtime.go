package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
//   - When a step exhausts its retry budget and still fails (non-zero exit
//     or executor error), the chain stops UNLESS the step has
//     `continue_on_fail: true` set. With continue_on_fail, the failure is
//     still recorded in RunResult.Steps but the next step runs anyway.
//     RunResult.ExitCode at the end is the LAST non-zero step exit seen
//     (so a single failing cleanup step at the end still surfaces failure),
//     or 0 when every step succeeded.
//   - The output target (stdout / file / webhook) receives the JSON-
//     serialised RunResult after the chain finishes (regardless of
//     success/failure).
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

	// pendingStep wraps the next ChainStep to run with the cascade depth
	// at which it was queued. Original chain steps are at depth 0;
	// cascaded steps land at depth N+1. The queue lets auto-cascade
	// append a successor without complicating the for-range index.
	type pendingStep struct {
		step  ChainStep
		depth int
	}
	queue := make([]pendingStep, 0, len(spec.Chain))
	for _, s := range spec.Chain {
		queue = append(queue, pendingStep{step: s, depth: 0})
	}

	cascadeEnabled := spec.AutoCascade != nil
	maxCascadeDepth := DefaultCascadeMaxDepth
	if cascadeEnabled && spec.AutoCascade.MaxDepth > 0 {
		maxCascadeDepth = spec.AutoCascade.MaxDepth
	}

	var stepErr error
	stepIdx := 0
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		stepResult, err := r.runOneStep(ctx, runID, stepIdx, p.step, retry)
		stepResult.CascadeDepth = p.depth
		result.Steps = append(result.Steps, stepResult)
		stepIdx++

		stepFailed := err != nil || stepResult.ExitCode != 0
		if stepFailed {
			// Always track the most recent failure as the run-level
			// exit code so even continue_on_fail runs surface failure
			// at the top level.
			result.ExitCode = stepResult.ExitCode
			if result.ExitCode == 0 {
				result.ExitCode = 1 // executor error without exit code = generic failure
			}
			stepErr = err
			if p.step.ContinueOnFail {
				_ = r.store.AppendLog(ctx, runID, "warn",
					fmt.Sprintf("step[%d] %s failed (exit=%d) — continue_on_fail=true, chain continues",
						stepIdx-1, p.step.Command, stepResult.ExitCode))
				continue
			}
			break
		}

		// Success branch: optionally cascade by appending the first
		// parsed §next-phase skill as a new step at depth+1. We do NOT
		// cascade from failed steps (broken §next-phase shouldn't
		// drive a fault path).
		if cascadeEnabled && p.depth < maxCascadeDepth {
			if next := pickCascadeTarget(stepResult.Parsed.NextPhase, spec.BranchHints); next != "" {
				queue = append(queue, pendingStep{
					step:  ChainStep{Command: next},
					depth: p.depth + 1,
				})
				_ = r.store.AppendLog(ctx, runID, "info",
					fmt.Sprintf("step[%d] %s auto-cascade → %s (depth %d)",
						stepIdx-1, p.step.Command, next, p.depth+1))
			}
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

		// Stream sink: forwards every line the executor emits straight into
		// agent_logs so `buddy agent log <id>` shows mid-progress on long
		// steps instead of waiting for the whole step to finish. Stdout
		// lines become info-level; stderr becomes warn — matching the
		// runtime's convention for the post-step ok / warn / error lines.
		sink := func(stream, line string) {
			level := "info"
			if stream == "stderr" {
				level = "warn"
			}
			_ = r.store.AppendLog(ctx, runID, level,
				fmt.Sprintf("step[%d] %s %s: %s", idx, step.Command, stream, line))
		}
		stdout, stderr, code, err := r.executor.Run(ctx, step.Command, step.Args, sink)
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
			// Self-check verdict policy. The verdict is already logged
			// above; the runtime acts on it only when the spec opts in
			// via OnSelfCheckFail. SelfCheckFailContinue (and the
			// empty default) preserves the prior behaviour: log the
			// verdict, return the step as success.
			//
			// Failure surfaces through last.ExitCode (=1) rather than a
			// non-nil error return — same contract as exit-code-fail so
			// the cascade loop's stepFailed detector picks it up
			// uniformly, and Run() keeps its (result, nil) success-on-
			// soft-failure invariant.
			if last.Parsed.SelfCheck.Verdict == SelfCheckFail {
				switch step.OnSelfCheckFail {
				case SelfCheckFailAbort:
					last.ExitCode = 1
					last.Error = "self-check failed (on_self_check_fail=abort)"
					_ = r.store.AppendLog(ctx, runID, "error",
						fmt.Sprintf("step[%d] %s self-check fail → abort", idx, step.Command))
					return last, nil
				case SelfCheckFailRetry:
					last.ExitCode = 1
					last.Error = "self-check failed (on_self_check_fail=retry)"
					_ = r.store.AppendLog(ctx, runID, "warn",
						fmt.Sprintf("step[%d] %s self-check fail → retry (attempt=%d)",
							idx, step.Command, attempt))
					// Skip the success-return path; fall through to the
					// retry-backoff block below for the next attempt.
				default:
					// "" / "continue" — keep v0.5.0+ behaviour: success.
					_ = r.store.AppendLog(ctx, runID, "info",
						fmt.Sprintf("step[%d] %s ok", idx, step.Command))
					return last, nil
				}
			} else {
				_ = r.store.AppendLog(ctx, runID, "info",
					fmt.Sprintf("step[%d] %s ok", idx, step.Command))
				return last, nil
			}
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

// pickCascadeTarget chooses the next-phase skill to auto-cascade into.
//
// Two code paths:
//
//   - No Branches: the PROCEDURE produced an unconditional next-phase
//     section. We take Skills[0] — the common single-sequential-
//     candidate case.
//   - Branches present: the PROCEDURE expresses conditional cascade
//     ("Korea → skill-a / USA → skill-b"). We pick the first branch
//     whose Condition has BranchHints[Condition] == true. Missing
//     hint or hint=false suppresses that branch; if no branch
//     matches, we return "" so the cascade stops rather than
//     silently grabbing the union's first entry.
//
// Returns "" when there is no cascade target.
func pickCascadeTarget(np ParsedOutput_NextPhaseAlias, hints map[string]bool) string {
	if len(np.Branches) > 0 {
		for _, b := range np.Branches {
			if hints[b.Condition] && len(b.Skills) > 0 {
				return b.Skills[0]
			}
		}
		return ""
	}
	if len(np.Skills) == 0 {
		return ""
	}
	return np.Skills[0]
}

// ParsedOutput_NextPhaseAlias is a local alias so the helper signature
// reads at the call site as "what part of the parser output we use"
// without dragging the full NextPhase type name into every line. Kept
// next to pickCascadeTarget because it has no other consumer.
type ParsedOutput_NextPhaseAlias = NextPhase

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
	case "webhook":
		return postWebhook(target, result)
	}
	return fmt.Errorf("agent: unsupported output.type %q", target.Type)
}

// postWebhook serialises the RunResult to JSON and POSTs (or PUTs / PATCHes)
// it to the target URL. Non-2xx responses surface as errors so the caller
// can log them — the runtime itself never blocks step success on output
// dispatch, but the appended log line records the failure for `buddy agent
// log <id>` consumers.
//
// Security note: header values are written literally. The runtime does
// not template `${ENV_VAR}` style references — if a spec needs to inject
// an Authorization secret, the caller is expected to template the YAML
// before `buddy agent create` rather than commit secrets to source.
func postWebhook(target *OutputTarget, result RunResult) error {
	if strings.TrimSpace(target.URL) == "" {
		return errors.New("agent: output.url empty when type='webhook'")
	}
	method := target.Method
	if method == "" {
		method = http.MethodPost
	}
	timeout := target.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("agent: marshal result for webhook: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, target.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("agent: webhook request: %w", err)
	}
	// Set Content-Type unless the spec explicitly overrides it (rare but
	// legal — e.g. a target that wants application/vnd.buddy+json).
	if _, hasContentType := target.Headers["Content-Type"]; !hasContentType {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range target.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("agent: webhook %s %s: %w", method, target.URL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Drain a short prefix of the body so error messages include any
		// machine-readable detail the server wanted to surface (e.g.
		// `{"error":"unauthorised"}`). Cap at 512 bytes so a server
		// returning a 5 MB stack trace doesn't blow up the log line.
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf(
			"agent: webhook %s %s returned %s: %s",
			method, target.URL, resp.Status, strings.TrimSpace(string(preview)))
	}
	return nil
}
