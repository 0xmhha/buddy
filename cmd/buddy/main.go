// Command buddy is the user-facing CLI for the buddy harness control plane.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/config"
	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/persona"
)

// version / gitSHA / buildDate are the three pieces that compose `buddy
// --version`. They are package-level `var` (not `const`) so the release build
// can inject real values via `-ldflags "-X main.gitSHA=... -X main.buildDate=..."`
// — see Makefile `build` target. The defaults below keep an unflagged
// `go build ./cmd/buddy` working, but mark the binary as a dev cut so users
// reporting bugs can tell at a glance.
//
// NOTE: `version` below is the source of truth for the binary's self-reported
// version. The Makefile mirrors it as RELEASE_VERSION (used in release
// artifact filenames like `dist/buddy_0.1.0_linux_amd64`) — when bumping for
// a release, update both. See Makefile §RELEASE_VERSION.
//
// Roadmap §3 M6 T3.
var (
	version   = "0.6.2"
	gitSHA    = "dev"
	buildDate = "unknown"
)

// versionString assembles the spec format
//
//	buddy X.Y.Z (sha=<short>, built=<rfc3339>)
//
// for cobra's `--version` output. Kept as a function (not a const) because
// gitSHA / buildDate are themselves vars set at link time. The "buddy " prefix
// is part of the contract, so newRootCmd overrides cobra's own version
// template to avoid printing "buddy version buddy 0.1.0 ...".
func versionString() string {
	return fmt.Sprintf("buddy %s (sha=%s, built=%s)", version, gitSHA, buildDate)
}

// friendError carries a pre-formatted, friend-tone message that main() prints
// verbatim (no `buddy: ` prefix added) and uses to exit with code 1 instead of
// the generic 2. Use newFriendError to construct.
type friendError struct{ msg string }

func (e *friendError) Error() string { return e.msg }

func newFriendError(msg string) error { return &friendError{msg: msg} }

// errUnhealthy is a sentinel returned by report-driven commands (e.g. doctor)
// that have already printed their full output to stdout and only need main() to
// signal a non-zero exit code. main() recognises it and exits 1 silently — no
// extra "buddy: " line, no message duplication.
var errUnhealthy = errors.New("unhealthy")

// resolvedDBPath returns the user-visible DB path: the explicit --db value if
// non-empty, else the default. Used solely to embed a friendly path in
// db-missing error messages — never propagated to db.Open (which has its own
// default-resolution and is the source of truth for actual file IO).
func resolvedDBPath(dbFlag string) string {
	if dbFlag != "" {
		return dbFlag
	}
	if p, err := db.DefaultPath(); err == nil {
		return p
	}
	return "~/.buddy/buddy.db"
}

// dbMissingFriendError renders the M5 T8 friend-tone message for read-only
// commands (stats, events) when db.Open returns ErrDBMissing. Centralised so
// stats and events stay in sync with each other and with diagnose's wording.
func dbMissingFriendError(dbFlag string) error {
	return newFriendError(persona.M(persona.KeyDBMissing, resolvedDBPath(dbFlag)))
}

func main() {
	root := newRootCmd()
	err := root.ExecuteContext(context.Background())
	if err == nil {
		return
	}
	if errors.Is(err, errUnhealthy) {
		os.Exit(1)
	}
	var fe *friendError
	if errors.As(err, &fe) {
		fmt.Fprintln(os.Stderr, fe.msg)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "buddy: %v\n", err)
	os.Exit(2)
}

// newRootCmd assembles the top-level `buddy` cobra command. Subcommands live
// in sibling `<feature>_cmd.go` files so any one of them can be edited
// without touching this file. The wire-up here is intentionally a flat list
// of constructors — alphabetical-ish by user-facing flow (hook capture →
// configuration → daemon lifecycle → install → reporting → feature/mcp/agent
// extensions).
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "buddy",
		Short:         "Claude Code 옆에서 hook 신뢰성을 지켜주는 친구",
		Version:       versionString(),
		SilenceUsage:  true,
		SilenceErrors: true,
		// PersistentPreRunE: best-effort locale resolution. Errors here are
		// non-fatal — buddy works in ko regardless. The per-subcommand --config
		// flag isn't a persistent flag, so root's PersistentPreRunE can't see
		// it; we read only config.DefaultPath() here. Subcommand-flag-aware
		// locale is a v0.2 deferral.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if path, err := config.DefaultPath(); err == nil {
				if c, err := config.Load(path); err == nil {
					_ = persona.SetLocale(persona.Locale(c.Effective().PersonaLocale))
				}
			}
			return nil
		},
	}
	root.AddCommand(newHookWrapCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newDaemonCmd())
	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newStatsCmd())
	root.AddCommand(newEventsCmd())
	root.AddCommand(newPurgeCmd())
	root.AddCommand(newFeatureCmd())
	root.AddCommand(newMcpCmd())
	root.AddCommand(newAgentCmd())
	// Strip cobra's default "<name> version " prefix: versionString() already
	// starts with "buddy ", and the spec format would otherwise render as
	// "buddy version buddy 0.1.0 (...)". The trailing newline matches cobra's
	// own default template so terminal output is unchanged in feel.
	root.SetVersionTemplate("{{.Version}}\n")
	return root
}
