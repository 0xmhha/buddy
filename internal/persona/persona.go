// Package persona owns user-facing string templates for buddy.
//
// Design rationale:
//
//   - Centralised catalog. A single audit point for all friend-tone wording.
//     spec §6.3 (침묵 default, 호들갑 X, 이모지 X) lives here, so reviewing
//     the persona once gives you the full sound of the CLI.
//
//   - Typed keys. The Key type catches typos at compile time. Adding a new
//     Key without populating the ko catalog is caught by
//     TestKO_HasEntryForEveryKey — it's a programmer error, not a runtime
//     surprise for users.
//
//   - Locale fallback. Missing keys in non-ko locales fall back to ko, so
//     v0.2's en migration can land incrementally without breaking ko
//     speakers. The en map starts empty in v0.1; that's the i18n토대.
//
//   - Format args via fmt.Sprintf. Templates use Go-style verbs (%s, %d,
//     %v) and callers pass typed args. Keeps the catalog dead simple
//     and Go-idiomatic; no template language to learn.
//
// Scope guard: this package has zero non-stdlib dependencies on purpose. Any
// internal package that wants friend-tone output can import persona without
// risking an import cycle. In particular, internal/diagnose currently imports
// it for doctor's render strings.
//
// Known gaps (v0.2 i18n sweep targets):
//   - queries.ErrInvalidLimit / queries.ErrInvalidWindow carry Korean text in
//     their Error() value; the CLI surfaces them via "buddy: " + err.Error().
//     Migrate to KeyQueriesInvalidLimit / KeyQueriesInvalidWindow when the
//     queries package error type grows a Code field (mirrors the v0.2 plan
//     for config.ValidationError.Reason).
//   - Internal log surfaces (daemon boot, hookwrap stderr) intentionally stay
//     scattered — they are debug surfaces, not friend-tone prompts.
package persona

import (
	"fmt"
	"sync"
)

// Locale is the active language tag.
type Locale string

const (
	LocaleKO Locale = "ko"
	LocaleEN Locale = "en"
)

// Key is a typed template identifier. Use the const declarations below; do
// not pass raw strings — the type prevents typos and lets AllKeys() stay
// authoritative.
type Key string

// Key constants. Naming convention: <module>.<situation>. Keep groupings in
// the same order as AllKeys() below so a human review can spot drift.
const (
	// install / uninstall
	KeyInstallDone                 Key = "install.done"
	KeyInstallNoOp                 Key = "install.noop"
	KeyInstallCliwrapWritten       Key = "install.cliwrap_written"
	KeyInstallSettingsMissing      Key = "install.settings_missing"
	KeyInstallBinaryHasSpaces      Key = "install.binary_has_spaces"
	KeyUninstallRestoredFromBackup Key = "uninstall.restored_from_backup"
	KeyUninstallRemovedWrapping    Key = "uninstall.removed_wrapping"
	KeyUninstallNothingRegistered  Key = "uninstall.nothing_registered"
	KeyUninstallDaemonStopped      Key = "uninstall.daemon_stopped"
	KeyUninstallDaemonKept         Key = "uninstall.daemon_kept"
	KeyUninstallDaemonNotStopping  Key = "uninstall.daemon_not_stopping"

	// daemon
	KeyDaemonAlreadyRunning Key = "daemon.already_running"
	KeyDaemonStarted        Key = "daemon.started"
	KeyDaemonStopSignalSent Key = "daemon.stop_signal_sent"
	KeyDaemonNotRunning     Key = "daemon.not_running"

	// doctor / health (rendered by internal/diagnose; the package imports
	// persona directly because it has no other dependents and no cycle risk)
	KeyDoctorAllHealthy       Key = "doctor.all_healthy"
	KeyDoctorIssuesHeader     Key = "doctor.issues_header"
	KeyDoctorDaemonUnreadable Key = "doctor.daemon_unreadable"
	KeyDoctorDaemonNotRunning Key = "doctor.daemon_not_running"
	KeyDoctorBacklog          Key = "doctor.backlog"
	KeyDoctorSlowHook         Key = "doctor.slow_hook"
	KeyDoctorFailRate         Key = "doctor.fail_rate"
	KeyDoctorDBOpenFailed     Key = "doctor.db_open_failed"
	KeyDoctorDBMissing        Key = "doctor.db_missing"

	// db / events / stats common — wording shared between cmd/buddy stats
	// and events read-only paths so they stay in lockstep.
	KeyDBReadFailed Key = "db.read_failed"
	KeyDBOpenFailed Key = "db.open_failed"
	KeyDBMissing    Key = "db.missing"

	// config CLI
	KeyConfigInvalid           Key = "config.invalid"
	KeyConfigInvalidField      Key = "config.invalid_field"       // %s = field, %s = reason
	KeyConfigUnknownField      Key = "config.unknown_field"       // %s = field
	KeyConfigReadFailed        Key = "config.read_failed"         // %v = err
	KeyConfigPathUnknown       Key = "config.path_unknown"        // %v = err
	KeyConfigSaveFailed        Key = "config.save_failed"         // %v = err
	KeyConfigSetExpectInt      Key = "config.set.expect_int"      // %s = field, %s = raw
	KeyConfigSetExpectDuration Key = "config.set.expect_duration" // %s = field, %s = raw
	KeyConfigSetParseFailed    Key = "config.set.parse_failed"    // %s = field, %v = err
	KeyConfigJSONFailed        Key = "config.json_failed"         // %v = err

	// config Validate Reason translations — declared for v0.2 (T2 deferred
	// Important #2). Not yet wired into translateConfigError; the catalog
	// is in place so the v0.2 sweep can flip the switch without churning the
	// persona files.
	KeyConfigReasonHookTimeoutOutOfRange  Key = "config.reason.hook_timeout_out_of_range"  // %d
	KeyConfigReasonHookSlowOutOfRange     Key = "config.reason.hook_slow_out_of_range"     // %d, %d
	KeyConfigReasonFailRateOutOfRange     Key = "config.reason.fail_rate_out_of_range"     // %d
	KeyConfigReasonOutboxBacklogTooSmall  Key = "config.reason.outbox_backlog_too_small"   // %d
	KeyConfigReasonNotifyChannelInvalid   Key = "config.reason.notify_channel_invalid"     // %s
	KeyConfigReasonPollIntervalOutOfRange Key = "config.reason.poll_interval_out_of_range" // %s
	KeyConfigReasonBatchSizeOutOfRange    Key = "config.reason.batch_size_out_of_range"    // %d
	KeyConfigReasonPersonaLocaleInvalid   Key = "config.reason.persona_locale_invalid"     // %s

	// purge
	KeyPurgeBeforeRequired  Key = "purge.before_required"
	KeyPurgeBeforeBadFormat Key = "purge.before_bad_format" // %v
	KeyPurgeFailed          Key = "purge.failed"            // %v
	KeyPurgeDryRunSummary   Key = "purge.dry_run_summary"   // %d, %d
	KeyPurgeDryRunNudge     Key = "purge.dry_run_nudge"
	KeyPurgeAppliedSummary  Key = "purge.applied_summary" // %d, %d

	// events follow markers
	KeyEventsFollowFailed Key = "events.follow_failed" // %v

	// feature CLI
	KeyFeatureUpserted  Key = "feature.upserted"   // %s = feature_id
	KeyFeatureDeleted   Key = "feature.deleted"     // %s = feature_id
	KeyFeatureNotFound  Key = "feature.not_found"   // %s = feature_id
	KeyFeatureListEmpty Key = "feature.list_empty"
	KeyFeatureFailed    Key = "feature.failed" // %v = err
)

// AllKeys returns every Key constant in declaration order. Used by the
// catalog-coverage test (TestKO_HasEntryForEveryKey) and intended to be the
// single source of truth for "what keys exist".
//
// IMPORTANT: this slice is the single source of truth for
// TestKO_HasEntryForEveryKey. A new Key added to the const block but NOT
// added here weakens the gate — koCatalog() could be missing an entry for
// it without any test failing (the runtime panic in M() would still trigger,
// but only when that code path is exercised). Adding new keys to BOTH the
// const block and this slice is mandatory.
func AllKeys() []Key {
	return []Key{
		// install / uninstall
		KeyInstallDone,
		KeyInstallNoOp,
		KeyInstallCliwrapWritten,
		KeyInstallSettingsMissing,
		KeyInstallBinaryHasSpaces,
		KeyUninstallRestoredFromBackup,
		KeyUninstallRemovedWrapping,
		KeyUninstallNothingRegistered,
		KeyUninstallDaemonStopped,
		KeyUninstallDaemonKept,
		KeyUninstallDaemonNotStopping,

		// daemon
		KeyDaemonAlreadyRunning,
		KeyDaemonStarted,
		KeyDaemonStopSignalSent,
		KeyDaemonNotRunning,

		// doctor
		KeyDoctorAllHealthy,
		KeyDoctorIssuesHeader,
		KeyDoctorDaemonUnreadable,
		KeyDoctorDaemonNotRunning,
		KeyDoctorBacklog,
		KeyDoctorSlowHook,
		KeyDoctorFailRate,
		KeyDoctorDBOpenFailed,
		KeyDoctorDBMissing,

		// db / events / stats common
		KeyDBReadFailed,
		KeyDBOpenFailed,
		KeyDBMissing,

		// config CLI
		KeyConfigInvalid,
		KeyConfigInvalidField,
		KeyConfigUnknownField,
		KeyConfigReadFailed,
		KeyConfigPathUnknown,
		KeyConfigSaveFailed,
		KeyConfigSetExpectInt,
		KeyConfigSetExpectDuration,
		KeyConfigSetParseFailed,
		KeyConfigJSONFailed,

		// config validate reasons
		KeyConfigReasonHookTimeoutOutOfRange,
		KeyConfigReasonHookSlowOutOfRange,
		KeyConfigReasonFailRateOutOfRange,
		KeyConfigReasonOutboxBacklogTooSmall,
		KeyConfigReasonNotifyChannelInvalid,
		KeyConfigReasonPollIntervalOutOfRange,
		KeyConfigReasonBatchSizeOutOfRange,
		KeyConfigReasonPersonaLocaleInvalid,

		// purge
		KeyPurgeBeforeRequired,
		KeyPurgeBeforeBadFormat,
		KeyPurgeFailed,
		KeyPurgeDryRunSummary,
		KeyPurgeDryRunNudge,
		KeyPurgeAppliedSummary,

		// events
		KeyEventsFollowFailed,

		// feature CLI
		KeyFeatureUpserted,
		KeyFeatureDeleted,
		KeyFeatureNotFound,
		KeyFeatureListEmpty,
		KeyFeatureFailed,
	}
}

// mu guards `current` and `catalog`. Reads (M, ML, ActiveLocale) take RLock;
// writes (SetLocale) take Lock. Tests may read or temporarily mutate the
// catalog under Lock to exercise fallback / panic paths; production code
// never mutates after init.
var (
	mu      sync.RWMutex
	current = LocaleKO
	catalog = map[Locale]map[Key]string{
		LocaleKO: koCatalog(),
		LocaleEN: enCatalog(), // empty in v0.1 — see package doc.
	}
)

// SetLocale sets the active locale. Safe to call from multiple goroutines.
// If the locale is unknown, falls back to ko and returns an error so a
// caller wanting to log "unknown locale, defaulting to ko" can do so. Buddy
// itself ignores the error — the fallback is the contract.
func SetLocale(loc Locale) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := catalog[loc]; !ok {
		current = LocaleKO
		return fmt.Errorf("unknown locale %q; using ko", loc)
	}
	current = loc
	return nil
}

// ActiveLocale returns the locale currently used by M().
func ActiveLocale() Locale {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// M renders the template under the active locale. Falls back to ko when the
// active-locale catalog is missing the key. Panics only when the key is
// missing in BOTH catalogs — that's a programmer error caught by tests.
func M(key Key, args ...any) string {
	mu.RLock()
	loc := current
	mu.RUnlock()
	return ML(loc, key, args...)
}

// ML renders for an explicit locale. Used by tests and for any future
// per-call locale override (e.g. one report rendered in en even though the
// CLI is otherwise running ko). Same fallback rules as M.
func ML(loc Locale, key Key, args ...any) string {
	mu.RLock()
	defer mu.RUnlock()
	if tpl, ok := catalog[loc][key]; ok && tpl != "" {
		return fmt.Sprintf(tpl, args...)
	}
	if loc != LocaleKO {
		if tpl, ok := catalog[LocaleKO][key]; ok && tpl != "" {
			return fmt.Sprintf(tpl, args...)
		}
	}
	panic(fmt.Sprintf("persona: missing template for key %q in any locale", key))
}
