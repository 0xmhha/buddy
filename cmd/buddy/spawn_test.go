package main

import (
	"os/exec"
	"runtime"
	"testing"
)

// TestStartAndDetach_CapturesPIDBeforeRelease seals the M5 T7 regression:
// spawnDetached used to read cmd.Process.Pid AFTER Release(), and Release()
// on Unix sets Pid = -1, so the user always saw "buddy: daemon 시작 (pid -1)".
//
// This test asserts two things:
//  1. The PID returned by startAndDetach is a real, positive PID — not -1.
//  2. cmd.Process.Pid AFTER startAndDetach returns is -1, i.e. Release() did
//     run. This is the structural guarantee: if a future refactor moves the
//     Pid read back below Release(), the first assertion will fail; if a
//     refactor drops Release() entirely (so the bug "appears" fixed for the
//     wrong reason — leaking the os.Process), the second assertion fails.
func TestStartAndDetach_CapturesPIDBeforeRelease(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Release() semantics differ on Windows; the bug we're sealing is
		// Unix-specific (Pid = -1 post-Release). buddy's daemon path is
		// Unix-only in practice.
		t.Skip("Unix-specific Release() behavior")
	}

	// Pick an inert child that exits quickly so the test doesn't leave a
	// stray process around. /bin/sleep 0.05 is portable across macOS/Linux
	// and finishes well before the test does.
	cmd := exec.Command("/bin/sleep", "0.05")

	pid, err := startAndDetach(cmd)
	if err != nil {
		t.Fatalf("startAndDetach: %v", err)
	}
	if pid <= 0 {
		t.Fatalf("expected positive PID captured before Release, got %d", pid)
	}
	if pid == -1 {
		t.Fatalf("PID is -1 — Release() ran before capture (M5 T7 regression)")
	}

	// Structural seal: Release() must have zeroed the public Pid field.
	// If this fails, either Release() was skipped (process leak) or stdlib
	// behavior changed; either way we want to know.
	if cmd.Process.Pid != -1 {
		t.Fatalf("expected cmd.Process.Pid = -1 after Release(), got %d", cmd.Process.Pid)
	}
}
