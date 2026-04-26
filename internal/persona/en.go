package persona

// enCatalog returns the English templates. Empty in v0.1 — ML/M fall back to
// the ko catalog for any missing key, so buddy keeps working in Korean while
// the English locale is built out. v0.2's i18n sweep populates this map.
//
// When a translator starts filling this in, the test gate to mirror is
// TestKO_HasEntryForEveryKey — eventually we'll grow a sibling
// TestEN_HasEntryForEveryKey or TestEN_OrFallback that the v0.2 work owns.
func enCatalog() map[Key]string {
	return map[Key]string{}
}
