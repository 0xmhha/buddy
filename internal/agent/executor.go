package agent

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// LogSink is the streaming callback Runtime supplies to executors so that
// each line emitted by the child process can be persisted in real time
// (rather than only after the whole step completes). stream is the literal
// "stdout" or "stderr". sink == nil means "do not stream" — Run still
// returns the full captured strings via its return values, so callers that
// did not opt in keep their v0.6.x behavior byte-identically.
type LogSink func(stream, line string)

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
	//
	// sink, when non-nil, is invoked once per output line (newline-trimmed)
	// as the child process emits it — enabling `buddy agent log <id>` to
	// surface mid-progress on long-running steps. Passing nil disables
	// streaming and matches the v0.6.x behavior exactly.
	Run(ctx context.Context, command, args string, sink LogSink) (stdout string, stderr string, exitCode int, err error)
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
//
// When sink != nil, stdout / stderr are scanned line-by-line as the child
// emits them: each line is forwarded to sink immediately and also collected
// into the returned buffers (so callers that read the full strings get the
// same content they always did). When sink == nil, the implementation falls
// back to bulk buffer capture — byte-identical to the v0.6.x behavior.
func (e *SubprocessExecutor) Run(ctx context.Context, command, args string, sink LogSink) (string, string, int, error) {
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

	if sink == nil {
		// Fast path: no streaming requested. Keep the previous bulk-buffer
		// behavior exactly so v0.6.x callers see no change.
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		exitCode, exitErr := translateExitCode(cmd.Run())
		return stdout.String(), stderr.String(), exitCode, exitErr
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", -1, fmt.Errorf("agent: stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		// stdoutPipe was acquired before this failed — close it explicitly
		// so the pipe FD doesn't leak until GC.
		_ = stdoutPipe.Close()
		return "", "", -1, fmt.Errorf("agent: stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		// Both pipes were acquired but the spawn failed; close them so the
		// FDs are released immediately (cmd.Wait() would close them after
		// a successful Start, but we never get there on this path).
		_ = stdoutPipe.Close()
		_ = stderrPipe.Close()
		return "", "", -1, fmt.Errorf("agent: spawn claude: %w", err)
	}

	var stdout, stderr bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go streamLines(stdoutPipe, &stdout, sink, "stdout", &wg)
	go streamLines(stderrPipe, &stderr, sink, "stderr", &wg)
	wg.Wait()

	exitCode, exitErr := translateExitCode(cmd.Wait())
	return stdout.String(), stderr.String(), exitCode, exitErr
}

// streamLines runs in a goroutine, reading newline-delimited lines from r,
// invoking sink for each, and appending the line (with its trailing newline
// re-attached so the buffer round-trips the original byte content) to buf.
// Scanner uses a 1 MiB max line size so JSON-formatted PROCEDURE outputs
// with embedded artefacts are not split mid-record; longer lines fall back
// to multi-chunk emission.
//
// r is *not* explicitly closed here — exec.Cmd.Wait() closes the pipe
// automatically once the child exits, and the caller (Run) calls Wait
// after this goroutine joins. An explicit Close would be redundant.
func streamLines(r io.ReadCloser, buf *bytes.Buffer, sink LogSink, stream string, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	// Use a large-ish max line size so a JSON-formatted PROCEDURE output
	// (could easily exceed 64KB with embedded artefacts) is not split mid-
	// record. 1 MiB matches the limit applied to log retention elsewhere.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(line)
		buf.WriteByte('\n')
		sink(stream, line)
	}
	// Scanner errors are surfaced indirectly: a truncated read shows up as a
	// missing trailing line in the captured buffer. Surfacing them to the
	// caller would race with cmd.Wait() and complicate Run's return contract;
	// the executor treats them as best-effort streaming and falls back on
	// exit code for success/failure.
	_ = scanner.Err()
}

// translateExitCode converts os/exec's run-error into the (exitCode, error)
// pair Run's contract expects: a non-zero child exit becomes (code, nil)
// instead of an error, while a spawn/transport failure becomes (-1, wrapped
// error). Used by both the fast path (bulk capture) and the streaming path.
func translateExitCode(runErr error) (int, error) {
	if runErr == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(runErr, &ee) {
		return ee.ExitCode(), nil
	}
	return -1, fmt.Errorf("agent: spawn claude: %w", runErr)
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

// Run records the call and returns the canned response (or Default). When
// sink != nil the mock simulates streaming by splitting the canned Stdout /
// Stderr on newlines and emitting one sink call per line — matching the
// SubprocessExecutor's contract closely enough that Runtime-level tests can
// assert mid-progress log behavior against the mock alone.
func (m *MockExecutor) Run(_ context.Context, command, args string, sink LogSink) (string, string, int, error) {
	m.Calls = append(m.Calls, MockCall{Command: NormalizeCommand(command), Args: args})
	resp := m.Default
	if r, ok := m.Responses[NormalizeCommand(command)]; ok {
		resp = r
	}
	if sink != nil {
		for _, line := range splitLines(resp.Stdout) {
			sink("stdout", line)
		}
		for _, line := range splitLines(resp.Stderr) {
			sink("stderr", line)
		}
	}
	return resp.Stdout, resp.Stderr, resp.ExitCode, resp.Err
}

// splitLines is the mock-side counterpart to bufio.Scanner: returns the
// newline-delimited lines of s with their trailing newline stripped. Empty
// trailing lines (e.g. "a\nb\n" → ["a", "b"]) are dropped so the mock
// doesn't synthesise a phantom blank line per response.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\n")
	// Drop the final empty element produced when s ends with "\n".
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
