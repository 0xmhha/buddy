package persona

// enCatalog returns the English templates. v0.2 i18n sweep is filling this
// map incrementally; missing keys still fall back to ko (see persona.go's
// "Locale fallback" rationale).
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
//   - Verb shape (%s, %d) preserved 1:1 from ko.
//
// When this map covers every Key, swap the fallback test
// TestKO_HasEntryForEveryKey for a parallel TestEN_HasEntryForEveryKey.
func enCatalog() map[Key]string {
	return map[Key]string{
		// Sample tracer-bullet — 5 keys spanning install, daemon, and doctor.
		// Used to lock in the friend-tone English voice before the rest of
		// the catalog is filled in. Keep these in sync with ko.go if the
		// canonical Korean strings ever change.
		KeyInstallDone:            "buddy: hooked up. i'll keep an eye on things now.",
		KeyInstallNoOp:            "buddy: already hooked up. nothing changed.",
		KeyDaemonStarted:          "buddy: daemon started (pid %d).",
		KeyDoctorAllHealthy:       "everything looks fine.\n",
		KeyDoctorDaemonNotRunning: "daemon isn't running. you can start it with 'buddy daemon start'.",
	}
}
