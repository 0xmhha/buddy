package main

import (
	"fmt"
	"os/exec"
)

// startAndDetach starts cmd and detaches it from the parent so the spawned
// child outlives this process. It returns the child's PID captured BEFORE
// Release, because (*os.Process).Release sets Pid = -1 on Unix on success
// (see Go stdlib os/exec_unix.go: (*Process).release zeroes the public Pid
// field). Reading cmd.Process.Pid after Release would always yield -1, which
// is the original M5 T7 bug — see docs/roadmap.md §M5 T7.
//
// The helper is split out from spawnDetached so a unit test can lock in the
// "PID captured pre-Release" invariant without exercising the full daemon
// lifecycle.
func startAndDetach(cmd *exec.Cmd) (int, error) {
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("spawn: %w", err)
	}
	pid := cmd.Process.Pid // capture BEFORE Release; Release() sets Pid = -1 on Unix.
	if err := cmd.Process.Release(); err != nil {
		return 0, fmt.Errorf("detach: %w", err)
	}
	return pid, nil
}
