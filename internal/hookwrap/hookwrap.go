// Package hookwrap implements the `buddy hook-wrap` command.
//
// Invariants (v0.1 spec §7.1):
//  1. stdout silence — wrapper never writes to stdout itself.
//  2. streaming passthrough — child stdout/stderr inherit parent fds.
//  3. stdin replay — buffered once, then forwarded to child.
//  4. exit code passthrough — signals encoded as 128+sig.
//  5. outbox failure does not break the hook.
//  6. empty child command = monitoring-only, exit 0.
//  7. malformed input is absorbed; wrapper itself never crashes.
package hookwrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/wm-it-22-00661/buddy/internal/adapter"
	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

const fallbackEvent = schema.EventPreToolUse

// Options carries everything Run needs from the CLI layer.
type Options struct {
	HookName       string
	Command        []string
	FallbackEvent  schema.HookEventName
	DBPath         string
	RecordToolArgs bool
	CustomTags     map[string]string

	// Injectable for tests; default to real os.Stdin/Stderr.
	Stdin  io.Reader
	Stderr io.Writer
}

// Result is what the CLI uses to set its own exit code.
type Result struct {
	ExitCode int
	OutboxID int64 // 0 if outbox write failed
}

// Run executes a hook with the wrapper invariants.
func Run(ctx context.Context, opts Options) (Result, error) {
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.FallbackEvent == "" {
		opts.FallbackEvent = fallbackEvent
	}

	startedAt := time.Now().UnixMilli()

	// 1. buffer stdin once — wrapper inspects it and replays to child.
	rawIn, err := io.ReadAll(opts.Stdin)
	if err != nil {
		// Inv 7: never crash on stdin issue. Treat as empty input.
		rawIn = nil
	}
	parsed := adapter.Parse(string(rawIn))

	// 2. spawn child (if any). stdio passthrough for streaming.
	exitCode := 0
	var childPID int
	if len(opts.Command) > 0 {
		exitCode, childPID = runChild(ctx, opts.Command, rawIn, opts.Stderr)
	}

	finishedAt := time.Now().UnixMilli()

	// 3. build payload.
	payload := adapter.Build(parsed, adapter.AdaptOptions{
		HookName:       opts.HookName,
		FallbackEvent:  opts.FallbackEvent,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
		ExitCode:       exitCode,
		PID:            childPID,
		RecordToolArgs: opts.RecordToolArgs,
		CustomTags:     opts.CustomTags,
	})

	// 4. outbox write — failure here MUST NOT change the hook's exit code (Inv 5).
	id, writeErr := writeOutbox(opts.DBPath, &payload)
	if writeErr != nil {
		fmt.Fprintf(opts.Stderr, "buddy: outbox write failed (%v)\n", writeErr)
	}

	return Result{ExitCode: exitCode, OutboxID: id}, nil
}

func runChild(
	ctx context.Context,
	command []string,
	stdin []byte,
	wrapperStderr io.Writer,
) (exitCode, pid int) {
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdinPipe, pipeErr := cmd.StdinPipe()
	if pipeErr != nil {
		fmt.Fprintf(wrapperStderr, "buddy: stdin pipe: %v\n", pipeErr)
		return 127, 0
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(wrapperStderr, "buddy: spawn failed: %v\n", err)
		return 127, 0
	}
	pid = cmd.Process.Pid

	// Replay stdin to child, then close.
	if len(stdin) > 0 {
		_, _ = stdinPipe.Write(stdin)
	}
	_ = stdinPipe.Close()

	if err := cmd.Wait(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			st := ee.ProcessState
			if st.Exited() {
				return st.ExitCode(), pid
			}
			// Signal-terminated: encode as 128 + signum (POSIX shell convention).
			if ws, ok := st.Sys().(interface{ Signal() int }); ok {
				return 128 + ws.Signal(), pid
			}
			return 128, pid
		}
		fmt.Fprintf(wrapperStderr, "buddy: child wait: %v\n", err)
		return 127, pid
	}
	return 0, pid
}

func writeOutbox(dbPath string, p *schema.HookEventPayload) (int64, error) {
	conn, err := db.Open(db.Options{Path: dbPath})
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	if err := p.Validate(); err != nil {
		return 0, fmt.Errorf("validate payload: %w", err)
	}
	return db.AppendToOutbox(conn, p)
}
