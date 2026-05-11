package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// Executor abstracts "run one buddy command and return its captured output".
// Production uses SubprocessExecutor (spawns `claude` CLI per ADR-005's
// §4.1 option (a) lock-in); tests use MockExecutor.
//
// The interface deliberately mirrors what cli-buddy-spec §4.2 describes
// inside `agent.run()` step (a)..(d): spawn → send → capture → parse. Parsing
// is the Runtime's job; the Executor only owns the spawn+capture half.
type Executor interface {
	// Run executes `<command> "<args>"` against the embedding layer (Claude
	// Code subprocess for SubprocessExecutor) and returns captured stdout +
	// stderr + exit code. Cancellation honours ctx.
	Run(ctx context.Context, command, args string) (stdout string, stderr string, exitCode int, err error)
}

// SubprocessExecutor implements Executor by spawning `claude` once per step
// and piping `/buddy:<command> "<args>"` to its stdin. This is the canonical
// path per ADR-005 §2.2.
//
// v0.3 caveats (tracked as W3-3 follow-ons):
//   - assumes `claude` is on PATH; surfaces a clear error if missing
//   - does not yet stream incremental output to AgentLog — only captures
//     the final stdout/stderr buffers
//   - does not yet parse PROCEDURE §6 self-check inside the captured stdout;
//     callers receive raw output and treat exit_code as success signal
type SubprocessExecutor struct {
	// ClaudeBinary is the path/name used in exec.LookPath. Defaults to "claude".
	ClaudeBinary string
	// ExtraArgs are passed before the dispatch payload. Empty by default.
	ExtraArgs []string
}

// NewSubprocessExecutor constructs the production executor with sane defaults.
func NewSubprocessExecutor() *SubprocessExecutor {
	return &SubprocessExecutor{ClaudeBinary: "claude"}
}

// ErrClaudeMissing is returned when the executor cannot find the claude CLI.
// Callers (CLI subcommands, scheduler) translate it into a friend-tone hint:
// install Claude Code, or set BUDDY_CLAUDE_BIN.
var ErrClaudeMissing = errors.New("agent: claude CLI not found on PATH (install Claude Code or set ClaudeBinary)")

// Run spawns the claude subprocess. The dispatch payload is sent on stdin —
// matching cli-buddy-spec §4.2 step (b).
func (e *SubprocessExecutor) Run(ctx context.Context, command, args string) (string, string, int, error) {
	bin := e.ClaudeBinary
	if bin == "" {
		bin = "claude"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return "", "", -1, ErrClaudeMissing
	}

	payload := fmt.Sprintf("/buddy:%s %s\n", NormalizeCommand(command), args)
	cmd := exec.CommandContext(ctx, bin, e.ExtraArgs...) // #nosec G204 — bin is config-controlled, args are spec-controlled and routed via stdin
	cmd.Stdin = bytes.NewReader([]byte(payload))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			exitCode = ee.ExitCode()
		} else {
			return stdout.String(), stderr.String(), -1, fmt.Errorf("agent: spawn claude: %w", runErr)
		}
	}
	return stdout.String(), stderr.String(), exitCode, nil
}

// MockExecutor is the test double. It records every call so tests can assert
// the sequence and surfaces canned responses keyed by command name.
type MockExecutor struct {
	Responses map[string]MockResponse // keyed by NormalizeCommand(command)
	Calls     []MockCall              // append-only call log
	Default   MockResponse            // returned when Responses has no entry for the command
}

// MockResponse describes the canned reply for one command.
type MockResponse struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// MockCall is one captured invocation. Tests inspect Calls to assert sequence.
type MockCall struct {
	Command string
	Args    string
}

// NewMockExecutor returns a ready-to-use mock with empty maps. Tests fill
// Responses before passing it to the Runtime.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{Responses: map[string]MockResponse{}}
}

// Run records the call and returns the canned response (or Default).
func (m *MockExecutor) Run(_ context.Context, command, args string) (string, string, int, error) {
	m.Calls = append(m.Calls, MockCall{Command: NormalizeCommand(command), Args: args})
	if r, ok := m.Responses[NormalizeCommand(command)]; ok {
		return r.Stdout, r.Stderr, r.ExitCode, r.Err
	}
	return m.Default.Stdout, m.Default.Stderr, m.Default.ExitCode, m.Default.Err
}
