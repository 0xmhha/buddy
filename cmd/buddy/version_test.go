package main

import (
	"strings"
	"testing"
)

// TestVersionString_FormatExact locks in the spec format
//
//	`buddy <ver> (sha=<short>, built=<rfc3339>)`
//
// so a future refactor can't drift the output silently. Roadmap §3 M6 T3.
func TestVersionString_FormatExact(t *testing.T) {
	// Set the package vars to known values, restore after.
	oldVer, oldSHA, oldDate := version, gitSHA, buildDate
	defer func() { version, gitSHA, buildDate = oldVer, oldSHA, oldDate }()

	version = "0.1.0"
	gitSHA = "abc1234"
	buildDate = "2026-04-26T07:08:09Z"

	got := versionString()
	want := "buddy 0.1.0 (sha=abc1234, built=2026-04-26T07:08:09Z)"
	if got != want {
		t.Fatalf("versionString = %q, want %q", got, want)
	}
}

// TestVersionString_DefaultsAreSafe guards the unflagged-build path: a plain
// `go build ./cmd/buddy` (no -ldflags) must still produce a non-empty,
// well-prefixed string so `buddy --version` never panics or prints garbage.
func TestVersionString_DefaultsAreSafe(t *testing.T) {
	s := versionString()
	if s == "" {
		t.Fatal("versionString() returned empty")
	}
	if !strings.HasPrefix(s, "buddy ") {
		t.Errorf("versionString() = %q; expected prefix 'buddy '", s)
	}
}
