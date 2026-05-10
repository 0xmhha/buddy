package persona

// enCatalog returns the English templates. v0.2 i18n-1 fills the full set;
// missing keys still fall back to ko via persona.go's "Locale fallback"
// rationale, but with a complete map this path is exercised only by future
// Key additions that haven't yet been translated.
//
// Tone guide (mirror of spec §6.3 friend-tone, applied to English):
//   - Same `buddy: ` prefix as ko.
//   - Plain declarative — no exclamation marks, no "Hello!", no "Awesome!".
//   - Lowercase and contractions where natural ("i'll", "isn't"). The ko
//     catalog uses casual register; the en mirror keeps that same energy
//     instead of translating to formal/business English.
//   - Match the trailing punctuation of the ko entry byte-for-byte (period
//     vs. period+\n etc.) so renderers that write the template directly to
//     an io.Writer keep producing identical layouts.
//   - Verb shape (%s, %d, %v, %q) preserved 1:1 from ko.
//
// TestEN_HasEntryForEveryKey mirrors TestKO_HasEntryForEveryKey to enforce
// completeness now that the map is full.
func enCatalog() map[Key]string {
	return map[Key]string{
		// install / uninstall
		KeyInstallDone:                 "buddy: hooked up. i'll keep an eye on things now.",
		KeyInstallNoOp:                 "buddy: already hooked up. nothing changed.",
		KeyInstallCliwrapWritten:       "buddy: also wrote out cliwrap.yaml (%s).",
		KeyInstallSettingsMissing:      "buddy: can't find ~/.claude/settings.json. is Claude Code installed?",
		KeyInstallBinaryHasSpaces:      "buddy: binary path has spaces. try moving it somewhere else: %s",
		KeyUninstallRestoredFromBackup: "buddy: unhooked. restored from backup.",
		KeyUninstallRemovedWrapping:    "buddy: unhooked. wrapping removed.",
		KeyUninstallNothingRegistered:  "buddy: nothing registered. leaving things alone.",
		KeyUninstallDaemonStopped:      "buddy: daemon stopped too.",
		KeyUninstallDaemonKept:         "buddy: leaving the daemon running (--keep-daemon).",
		KeyUninstallDaemonNotStopping:  "buddy: daemon won't stop, leaving it for now. try 'buddy daemon stop' once.",

		// daemon
		KeyDaemonAlreadyRunning: "buddy: already running (pid %d).",
		KeyDaemonStarted:        "buddy: daemon started (pid %d).",
		KeyDaemonStopSignalSent: "buddy: sent a stop signal to the daemon (pid %d).",
		KeyDaemonNotRunning:     "buddy: no daemon is running.",

		// doctor / health (consumed by internal/diagnose)
		//
		// AllHealthy / IssuesHeader carry trailing newlines because the doctor
		// renderer writes them directly to the io.Writer with no extra Println.
		// Preserving the trailing \n keeps doctor's output byte-for-byte.
		KeyDoctorAllHealthy:       "everything looks fine.\n",
		KeyDoctorIssuesHeader:     "hey, a few things to look at.\n\n",
		KeyDoctorDaemonUnreadable: "couldn't read daemon state (%s): %v",
		KeyDoctorDaemonNotRunning: "daemon isn't running. you can start it with 'buddy daemon start'.",
		KeyDoctorBacklog:          "%s items piled up in the outbox. take a look at the daemon (buddy daemon status).",
		KeyDoctorSlowHook:         "'%s' hook is getting slow. p95 is %s (threshold %s).",
		KeyDoctorFailRate:         "'%s' hook fail rate is %d%%. %d failures out of the last %d.",
		KeyDoctorDBOpenFailed:     "couldn't open the DB (%s): %v",
		KeyDoctorDBMissing:        "DB doesn't exist yet (%s). check that you ran 'buddy install' first.",

		// db / events / stats common
		KeyDBReadFailed: "buddy: couldn't read the DB. has the daemon ever run? (%v)",
		KeyDBOpenFailed: "buddy: couldn't open the DB (%v).",
		KeyDBMissing:    "buddy: DB doesn't exist yet (%s). check that you ran 'buddy install' first.",

		// config CLI
		KeyConfigInvalid:           "buddy: config is invalid:",
		KeyConfigInvalidField:      "  - %s: %s",
		KeyConfigUnknownField:      "buddy: no setting like '%s'. run 'buddy config show' to see the list.",
		KeyConfigReadFailed:        "buddy: couldn't read config (%v).",
		KeyConfigPathUnknown:       "buddy: not sure where the config lives (%v).",
		KeyConfigSaveFailed:        "buddy: config save failed (%v).",
		KeyConfigSetExpectInt:      "buddy: %s should be a number (\"%s\").",
		KeyConfigSetExpectDuration: "buddy: %s should be a duration (\"%s\", like 1s or 500ms).",
		KeyConfigSetParseFailed:    "buddy: couldn't parse %s value (%v).",
		KeyConfigJSONFailed:        "buddy: JSON serialization failed (%v).",

		// config Validate Reason translations
		KeyConfigReasonHookTimeoutOutOfRange:  "needs to be between 100ms and 600s (currently %dms).",
		KeyConfigReasonHookSlowOutOfRange:     "needs to be between 1ms and hookTimeoutMs (currently %dms, timeout %dms).",
		KeyConfigReasonFailRateOutOfRange:     "needs to be between 1 and 100 (currently %d).",
		KeyConfigReasonOutboxBacklogTooSmall:  "needs to be at least 1 (currently %d).",
		KeyConfigReasonNotifyChannelInvalid:   "only \"stderr\" is allowed (currently %q).",
		KeyConfigReasonPollIntervalOutOfRange: "needs to be between 100ms and 60s (currently %s).",
		KeyConfigReasonBatchSizeOutOfRange:    "needs to be between 1 and 100000 (currently %d).",
		KeyConfigReasonPersonaLocaleInvalid:   "needs to be \"ko\" or \"en\" (currently %q).",

		// purge — DryRun/Applied summaries carry trailing newlines because the
		// purge command Fprintf's them directly without an extra Println.
		KeyPurgeBeforeRequired:  "buddy: --before is required (e.g. --before 30d, --before 2026-01-01).",
		KeyPurgeBeforeBadFormat: "buddy: --before format looks off (%v). examples: 30d, 2026-01-01, 2026-01-01T00:00:00Z",
		KeyPurgeFailed:          "buddy: purge failed (%v).",
		KeyPurgeDryRunSummary:   "buddy: dry-run. %d hook_events and %d hook_stats would be deleted.\n",
		KeyPurgeDryRunNudge:     "buddy: to actually delete, add --apply. (outbox stays untouched.)\n",
		KeyPurgeAppliedSummary:  "buddy: deleted %d hook_events and %d hook_stats. (outbox left alone.)\n",

		// events
		KeyEventsFollowFailed: "buddy: events follow failed (%v)",

		// feature CLI
		KeyFeatureUpserted:  "buddy: feature saved (%s).",
		KeyFeatureDeleted:   "buddy: feature deleted (%s).",
		KeyFeatureNotFound:  "buddy: can't find that feature (%s).",
		KeyFeatureListEmpty: "buddy: no features registered.",
		KeyFeatureFailed:    "buddy: feature operation failed (%v).",
	}
}
