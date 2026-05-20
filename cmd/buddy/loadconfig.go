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

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/config"
	"github.com/0xmhha/buddy/internal/daemon"
	"github.com/0xmhha/buddy/internal/diagnose"
	"github.com/0xmhha/buddy/internal/notify"
	"github.com/0xmhha/buddy/internal/persona"
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
		// W7-1 (ADR-012) session monitor wiring from config. No CLI flag
		// override yet — config-driven only. Disabled=true skips the
		// goroutine entirely; defaults wire to 30s / 1h via config.Defaults.
		SessionMonitor: daemon.SessionMonitorConfig{
			Disabled:       eff.SessionMonitorDisabled,
			PollInterval:   eff.SessionMonitorPollInterval,
			EndedThreshold: eff.SessionMonitorEndedThreshold,
		},
		// W7-3b (ADR-015) advisor monitor wiring. Same config-driven
		// pattern. Thresholds.PollInterval governs cadence (default 1h).
		Advisor: daemon.AdvisorMonitorConfig{
			Disabled: eff.AdvisorDisabled,
			Thresholds: advisor.Thresholds{
				Disabled:            eff.AdvisorDisabled,
				TokenSpikeRatio:     eff.AdvisorTokenSpikeRatio,
				LongSessionHours:    eff.AdvisorLongSessionHours,
				LowCachePct:         eff.AdvisorLowCachePct,
				SessionVolumePerDay: eff.AdvisorSessionVolumePerDay,
				TokenDailyThreshold: eff.AdvisorTokenDailyThreshold,
				DedupWindow:         eff.AdvisorDedupWindow,
				PollInterval:        eff.AdvisorPollInterval,
			},
			NotifyChannels: buildNotifyChannelSpecs(eff),
		},
	}
}

// buildNotifyChannelSpecs projects the Effective config into the
// transport-agnostic daemon spec list. Empty slice = no channels
// configured / all disabled → daemon skips notify dispatch.
func buildNotifyChannelSpecs(eff config.Effective) []daemon.NotifyChannelSpec {
	var out []daemon.NotifyChannelSpec
	if eff.NotifyDesktopEnabled {
		out = append(out, daemon.NotifyChannelSpec{
			Kind:        notify.ChannelDesktop,
			SeverityMin: eff.NotifyDesktopSeverityMin,
			DedupWindow: eff.NotifyDesktopDedup,
		})
	}
	if eff.NotifyTUIBannerEnabled {
		out = append(out, daemon.NotifyChannelSpec{
			Kind:        notify.ChannelTUIBanner,
			SeverityMin: eff.NotifyTUIBannerSeverityMin,
		})
	}
	if eff.NotifyShellPromptEnabled {
		out = append(out, daemon.NotifyChannelSpec{
			Kind:        notify.ChannelShell,
			SeverityMin: eff.NotifyShellPromptSeverityMin,
		})
	}
	for _, w := range eff.NotifyWebhooks {
		out = append(out, daemon.NotifyChannelSpec{
			Kind:        notify.ChannelWebhook,
			SeverityMin: w.SeverityMin,
			DedupWindow: w.DedupWindow,
			URL:         w.URL,
			Method:      w.Method,
			Headers:     w.Headers,
			Timeout:     w.Timeout,
		})
	}
	return out
}
