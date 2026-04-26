package persona

// koCatalog returns the Korean templates. Strings here are the canonical
// friend-tone wording — spec §6.3 (침묵 default, 호들갑 X, 이모지 X). When
// migrating an existing callsite, copy the string verbatim (down to the
// trailing punctuation) so byte-level test fixtures keep passing.
//
// fmt.Sprintf verbs in templates:
//   - %s   string
//   - %q   quoted string (escapes embedded quotes)
//   - %d   int
//   - %v   any (used for errors and durations)
//
// The arg shape is documented per-key in persona.go's const block.
func koCatalog() map[Key]string {
	return map[Key]string{
		// install / uninstall
		KeyInstallDone:                 "buddy: 등록 완료. 이제 옆에서 보고 있을게.",
		KeyInstallNoOp:                 "buddy: 이미 등록되어 있어. 변화 없음.",
		KeyInstallCliwrapWritten:       "buddy: cliwrap.yaml 도 써뒀어 (%s).",
		KeyInstallSettingsMissing:      "buddy: ~/.claude/settings.json 이 안 보여. Claude Code 설치되어 있어?",
		KeyInstallBinaryHasSpaces:      "buddy: 바이너리 경로에 공백이 있어. 다른 경로로 옮겨봐: %s",
		KeyUninstallRestoredFromBackup: "buddy: 해제 완료. 백업에서 복원했어.",
		KeyUninstallRemovedWrapping:    "buddy: 해제 완료. wrapping 제거했어.",
		KeyUninstallNothingRegistered:  "buddy: 등록된 게 없어. 그대로 둘게.",
		KeyUninstallDaemonStopped:      "buddy: daemon도 같이 멈췄어.",
		KeyUninstallDaemonKept:         "buddy: daemon은 그대로 둘게 (--keep-daemon).",
		KeyUninstallDaemonNotStopping:  "buddy: daemon이 안 멈춰서 그냥 두고 갈게. 'buddy daemon stop' 한 번 해줘.",

		// daemon
		KeyDaemonAlreadyRunning: "buddy: 이미 실행 중이야 (pid %d).",
		KeyDaemonStarted:        "buddy: daemon 시작 (pid %d).",
		KeyDaemonStopSignalSent: "buddy: daemon에 종료 신호 보냈어 (pid %d).",
		KeyDaemonNotRunning:     "buddy: 실행 중인 daemon이 없어.",

		// doctor / health (consumed by internal/diagnose)
		//
		// AllHealthy / IssuesHeader carry trailing newlines because the doctor
		// renderer writes them directly to the io.Writer with no extra Println.
		// Preserving the trailing \n keeps doctor's output byte-for-byte.
		KeyDoctorAllHealthy:       "모두 정상이야.\n",
		KeyDoctorIssuesHeader:     "어, 몇 가지 봐줄 게 있어.\n\n",
		KeyDoctorDaemonUnreadable: "daemon 상태를 못 읽었어 (%s): %v",
		KeyDoctorDaemonNotRunning: "daemon이 실행 중이 아니야. 'buddy daemon start'로 띄울 수 있어.",
		KeyDoctorBacklog:          "outbox에 %s개 쌓였어. daemon 한 번 봐줘 (buddy daemon status).",
		KeyDoctorSlowHook:         "'%s' hook이 좀 느려졌어. p95가 %s (기준 %s).",
		KeyDoctorFailRate:         "'%s' hook 실패율이 %d%% 야. 최근 %d번 중 %d번 실패.",
		KeyDoctorDBOpenFailed:     "DB를 못 열었어 (%s): %v",
		KeyDoctorDBMissing:        "DB가 아직 없어 (%s). 먼저 'buddy install' 했는지 확인해줘.",

		// db / events / stats common — match existing wording exactly so the
		// stats and events read-only paths produce identical user output.
		KeyDBReadFailed: "buddy: DB를 못 읽었어. daemon이 한 번이라도 돈 적 있어? (%v)",
		KeyDBOpenFailed: "buddy: DB를 못 열었어 (%v).",
		KeyDBMissing:    "buddy: DB가 아직 없어 (%s). 먼저 'buddy install' 했는지 확인해줘.",

		// config CLI
		KeyConfigInvalid:           "buddy: 설정이 잘못됐어:",
		KeyConfigInvalidField:      "  - %s: %s",
		KeyConfigUnknownField:      "buddy: '%s' 같은 설정은 없어. 'buddy config show'로 목록 봐줘.",
		KeyConfigReadFailed:        "buddy: 설정 못 읽었어 (%v).",
		KeyConfigPathUnknown:       "buddy: config 경로를 모르겠어 (%v).",
		KeyConfigSaveFailed:        "buddy: config 저장 실패 (%v).",
		KeyConfigSetExpectInt:      "buddy: %s 값은 숫자여야 해 (\"%s\").",
		KeyConfigSetExpectDuration: "buddy: %s 는 duration 형식이어야 해 (\"%s\", 예: 1s, 500ms).",
		KeyConfigSetParseFailed:    "buddy: %s 값을 못 읽었어 (%v).",
		KeyConfigJSONFailed:        "buddy: JSON 직렬화 실패 (%v).",

		// config Validate Reason translations — declared and tested for v0.2,
		// not yet wired into translateConfigError. See the TODO in
		// cmd/buddy/config_cmd.go pointing at these keys.
		KeyConfigReasonHookTimeoutOutOfRange:  "100ms부터 600초까지 잡아야 해 (지금 %dms).",
		KeyConfigReasonHookSlowOutOfRange:     "1ms부터 hookTimeoutMs 까지 (지금 %dms, timeout %dms).",
		KeyConfigReasonFailRateOutOfRange:     "1부터 100까지여야 해 (지금 %d).",
		KeyConfigReasonOutboxBacklogTooSmall:  "1 이상이어야 해 (지금 %d).",
		KeyConfigReasonNotifyChannelInvalid:   "\"stderr\" 만 가능해 (지금 %q).",
		KeyConfigReasonPollIntervalOutOfRange: "100ms부터 60초까지여야 해 (지금 %s).",
		KeyConfigReasonBatchSizeOutOfRange:    "1부터 100000까지여야 해 (지금 %d).",
		KeyConfigReasonPersonaLocaleInvalid:   "\"ko\" 또는 \"en\" 이어야 해 (지금 %q).",

		// purge — DryRun/Applied summaries carry trailing newlines because the
		// purge command Fprintf's them directly without an extra Println.
		KeyPurgeBeforeRequired:  "buddy: --before 가 필요해 (예: --before 30d, --before 2026-01-01).",
		KeyPurgeBeforeBadFormat: "buddy: --before 형식이 이상해 (%v). 예: 30d, 2026-01-01, 2026-01-01T00:00:00Z",
		KeyPurgeFailed:          "buddy: purge 실패 (%v).",
		KeyPurgeDryRunSummary:   "buddy: dry-run. %d개 hook_events, %d개 hook_stats 가 삭제 대상이야.\n",
		KeyPurgeDryRunNudge:     "buddy: 진짜 지우려면 --apply 추가해줘. (outbox는 안 건드려.)\n",
		KeyPurgeAppliedSummary:  "buddy: %d개 hook_events, %d개 hook_stats 삭제했어. (outbox는 그대로.)\n",

		// events
		KeyEventsFollowFailed: "buddy: events follow 실패 (%v)",
	}
}
