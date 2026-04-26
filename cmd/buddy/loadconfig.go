package main

// loadconfig.go is the M5 T3 integration boundary between the cmd layer and
// the internal/config package. The two policy consumers — doctor (diagnose)
// and daemon — DO NOT import internal/config: they keep accepting plain
// typed fields (Thresholds / Config) so their unit tests stay free of file
// IO. cmd/buddy is the single place that translates "user's config.json"
// into those typed views.
//
// Precedence (matches user intuition + the design note in the T3 brief):
//   explicit CLI flag > config file value > spec default
//
// The "spec default" tier is provided by config.Defaults() — which is itself
// the same source of truth as diagnose.DefaultThresholds() and
// daemon.Config.Defaults(). All three trace back to v0.1-spec §6.2.

import (
	"errors"
	"time"

	"github.com/wm-it-22-00661/buddy/internal/config"
	"github.com/wm-it-22-00661/buddy/internal/daemon"
	"github.com/wm-it-22-00661/buddy/internal/diagnose"
	"github.com/wm-it-22-00661/buddy/internal/persona"
)

// loadEffectiveConfig loads ~/.buddy/config.json (or path) and returns its
// Effective view.
//
// Behavior:
//   - Missing file ⇒ config.Defaults(), no error. This is the normal first-run
//     state and must NOT degrade the user experience.
//   - Parse / validation failure ⇒ friendError with a per-field bullet list
//     (via translateConfigError) so the user can fix it without grepping
//     stack traces.
//
// Pass an empty path to use config.DefaultPath().
func loadEffectiveConfig(path string) (config.Effective, error) {
	c, err := config.Load(path)
	if err != nil {
		return config.Effective{}, translateConfigLoadError(err)
	}
	return c.Effective(), nil
}

// translateConfigLoadError mirrors config_cmd.go's translateConfigError but
// stays usable when the underlying error is already wrapped by config.Load
// ("invalid config <path>: <ValidationError|MultiError>"). errors.As walks
// the chain, so the bullet-list rendering still triggers.
func translateConfigLoadError(err error) error {
	var ve *config.ValidationError
	var multi *config.MultiError
	switch {
	case errors.As(err, &multi):
		// translateConfigError returns a *friendError wrapping the bullet
		// list; reuse it verbatim so doctor / daemon CLI surfaces stay in
		// lockstep with `buddy config show`.
		return translateConfigError(multi)
	case errors.As(err, &ve):
		return translateConfigError(ve)
	}
	return newFriendError(persona.M(persona.KeyConfigReadFailed, err))
}

// buildDoctorOptions packages an Effective into the diagnose.Options the
// doctor command passes to diagnose.Check.
//
// Split out so a unit test can lock in the threshold mapping without
// constructing a fake cobra.Command tree. If a future field on Effective
// gains a doctor-side meaning, this is the single place that grows.
func buildDoctorOptions(dbFlag, pidFile string, eff config.Effective) diagnose.Options {
	return diagnose.Options{
		DBPath:  dbFlag,
		PIDFile: pidFile,
		Thresholds: diagnose.Thresholds{
			HookTimeoutMs:   eff.HookTimeoutMs,
			HookSlowMs:      eff.HookSlowMs,
			HookFailRatePct: eff.HookFailRatePct,
			OutboxBacklog:   eff.OutboxBacklog,
		},
	}
}

// resolveDaemonRunConfig applies CLI-flag-overrides-config precedence for the
// `buddy daemon run` knobs and returns a daemon.Config ready for daemon.Run.
//
// pollFlag/batchFlag are the cobra flag values; they win over `eff` whenever
// they are non-zero. A zero pollFlag (0s) or zero batchFlag (0) is the agreed
// "use config / spec default" sentinel — see the flag definitions in
// newDaemonRunCmd which now default to 0 instead of 1s/500.
//
// Split out for the same reason as buildDoctorOptions: a unit test can pin the
// precedence rules without spinning up a real daemon.
func resolveDaemonRunConfig(dbFlag, pidFile string, pollFlag time.Duration, batchFlag int, eff config.Effective) daemon.Config {
	poll := pollFlag
	if poll == 0 {
		poll = eff.PollInterval
	}
	batch := batchFlag
	if batch == 0 {
		batch = eff.BatchSize
	}
	return daemon.Config{
		DBPath:       dbFlag,
		PIDFile:      pidFile,
		PollInterval: poll,
		BatchSize:    batch,
	}
}
