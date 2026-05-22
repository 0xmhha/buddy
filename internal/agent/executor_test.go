package agent

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// ─── NewSubprocessExecutor defaults ────────────

// NewSubprocessExecutor must default ExtraArgs to ["--print"]. Without it
// the spawned `claude` process drops into an interactive REPL and the run
// never finalises (agent_runs.ended_at stays NULL, agents.status stays
// "running"). Dogfood surfaced this and the test locks the
// default in so a future refactor cannot silently drop --print without
// also updating this assertion.
func TestNewSubprocessExecutor_DefaultsToPrintMode(t *testing.T) {
	t.Parallel()
	e := NewSubprocessExecutor()
	require.Equal(t, "claude", e.ClaudeBinary,
		"default binary should remain `claude` (LookPath-resolved)")
	require.Equal(t, []string{"--print"}, e.ExtraArgs,
		"default ExtraArgs must contain --print for headless spawn")
}

// ─── MockExecutor sink emit ─────────────────────────────────────────────

func TestMockExecutor_NilSinkSkipsStreaming(t *testing.T) {
	t.Parallel()
	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{
		Stdout:   "line a\nline b\n",
		ExitCode: 0,
	}
	stdout, stderr, code, err := mock.Run(context.Background(), "status", "", nil)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	require.Equal(t, "line a\nline b\n", stdout)
	require.Equal(t, "", stderr)
}

func TestMockExecutor_NonNilSinkReceivesEachLine(t *testing.T) {
	t.Parallel()
	mock := NewMockExecutor()
	mock.Responses["build"] = MockResponse{
		Stdout:   "starting\ncompiling\nlinking\n",
		Stderr:   "warning: deprecated flag\n",
		ExitCode: 0,
	}
	var mu sync.Mutex
	var captured []string
	sink := func(stream, line string) {
		mu.Lock()
		defer mu.Unlock()
		captured = append(captured, stream+":"+line)
	}
	stdout, stderr, code, err := mock.Run(context.Background(), "build", "", sink)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	require.Equal(t,
		[]string{
			"stdout:starting",
			"stdout:compiling",
			"stdout:linking",
			"stderr:warning: deprecated flag",
		},
		captured,
		"sink should emit stdout lines first then stderr lines, each once, in order")

	// Full strings still come back so existing v0.6.x callers keep their
	// behavior.
	require.Equal(t, "starting\ncompiling\nlinking\n", stdout)
	require.Equal(t, "warning: deprecated flag\n", stderr)
}

func TestMockExecutor_EmptyStreamYieldsNoSinkCall(t *testing.T) {
	t.Parallel()
	mock := NewMockExecutor()
	mock.Responses["noop"] = MockResponse{Stdout: "", Stderr: "", ExitCode: 0}
	var calls int
	sink := func(stream, line string) { calls++ }
	_, _, _, err := mock.Run(context.Background(), "noop", "", sink)
	require.NoError(t, err)
	require.Zero(t, calls, "empty stdout/stderr must not synthesise a phantom blank line")
}

func TestMockExecutor_TrailingNewlineDoesNotDoubleEmit(t *testing.T) {
	t.Parallel()
	mock := NewMockExecutor()
	// "a\nb\n" vs "a\nb" must yield the same number of sink calls — the
	// trailing newline is a terminator, not a record.
	mock.Responses["x"] = MockResponse{Stdout: "a\nb\n"}
	mock.Responses["y"] = MockResponse{Stdout: "a\nb"}

	var countX, countY int
	sinkX := func(string, string) { countX++ }
	sinkY := func(string, string) { countY++ }

	_, _, _, _ = mock.Run(context.Background(), "x", "", sinkX)
	_, _, _, _ = mock.Run(context.Background(), "y", "", sinkY)
	require.Equal(t, 2, countX)
	require.Equal(t, 2, countY)
	require.Equal(t, countX, countY)
}

// ─── splitLines unit ─────────────────────────────────────────────────────

func TestSplitLines_HandlesAllShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"a\n", []string{"a"}},
		{"a\nb", []string{"a", "b"}},
		{"a\nb\n", []string{"a", "b"}},
		{"\n", []string{""}}, // single newline → one empty record (caller's stdout is literally "\n")
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, splitLines(tc.in), "splitLines(%q)", tc.in)
	}
}

// ─── SubprocessExecutor real subprocess streaming ────────────────────────

// TestSubprocessExecutor_StreamsLinesViaSink exercises the real subprocess
// path against /bin/sh, which is universally available on macOS/Linux CI
// runners. It validates that sink fires for each printed line *before* the
// command exits, by feeding the lines to the sink synchronously through a
// channel that the test reads after Run returns.
//
// We don't time-assert mid-progress (CI scheduler noise would make that
// flaky) — we only verify the *contract*: each printed line surfaces via
// sink, in stdout / stderr order, and the captured full strings still
// reflect the same content.
func TestSubprocessExecutor_StreamsLinesViaSink(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("/bin/sh"); err != nil {
		t.Skip("/bin/sh not available; skipping subprocess streaming test")
	}

	// Configure SubprocessExecutor to run /bin/sh directly. We bypass the
	// normal "claude subprocess" path because no claude CLI is on the CI
	// runner; the streaming logic itself is what we're testing.
	exe := &SubprocessExecutor{
		ClaudeBinary: "/bin/sh",
		ExtraArgs:    []string{"-c", `printf 'one\ntwo\nthree\n'; printf 'err1\nerr2\n' 1>&2`},
	}

	var mu sync.Mutex
	var got []string
	sink := func(stream, line string) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, stream+":"+line)
	}

	stdout, stderr, code, err := exe.Run(context.Background(), "noop", "", sink)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	mu.Lock()
	stdoutLines := filterStream(got, "stdout")
	stderrLines := filterStream(got, "stderr")
	mu.Unlock()

	require.Equal(t,
		[]string{"stdout:one", "stdout:two", "stdout:three"},
		stdoutLines,
		"every stdout line should surface via sink in order")
	require.Equal(t,
		[]string{"stderr:err1", "stderr:err2"},
		stderrLines,
		"every stderr line should surface via sink in order")

	// Captured full strings still hold the same content (caller contract).
	require.Equal(t, "one\ntwo\nthree\n", stdout)
	require.Equal(t, "err1\nerr2\n", stderr)
}

// TestSubprocessExecutor_NilSinkBulkCapture verifies the fast path
// (sink == nil) — the executor should fall back to bytes.Buffer capture
// and return the same content as the streaming path would, byte-for-byte.
func TestSubprocessExecutor_NilSinkBulkCapture(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("/bin/sh"); err != nil {
		t.Skip("/bin/sh not available")
	}
	exe := &SubprocessExecutor{
		ClaudeBinary: "/bin/sh",
		ExtraArgs:    []string{"-c", `printf 'first\nsecond\n'`},
	}
	stdout, stderr, code, err := exe.Run(context.Background(), "noop", "", nil)
	require.NoError(t, err)
	require.Equal(t, 0, code)
	require.Equal(t, "first\nsecond\n", stdout)
	require.Equal(t, "", stderr)
}

// TestSubprocessExecutor_NonZeroExitCodeIsCaptured makes sure the streaming
// path still translates an exit error into (code, nil) like the bulk path.
func TestSubprocessExecutor_NonZeroExitCodeIsCaptured(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("/bin/sh"); err != nil {
		t.Skip("/bin/sh not available")
	}
	exe := &SubprocessExecutor{
		ClaudeBinary: "/bin/sh",
		ExtraArgs:    []string{"-c", `printf 'partial output\n'; exit 7`},
	}
	sink := func(string, string) {}
	stdout, _, code, err := exe.Run(context.Background(), "noop", "", sink)
	require.NoError(t, err, "non-zero exit should not be an error")
	require.Equal(t, 7, code)
	require.Equal(t, "partial output\n", stdout)
}

func filterStream(events []string, stream string) []string {
	var out []string
	prefix := stream + ":"
	for _, e := range events {
		if strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	return out
}
