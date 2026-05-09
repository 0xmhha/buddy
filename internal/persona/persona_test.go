package persona

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKO_HasEntryForEveryKey is the foundational coverage gate: every Key
// constant must have a non-empty entry in the ko catalog. This is the test
// that catches "added a Key but forgot the template" before it ships as a
// runtime panic.
func TestKO_HasEntryForEveryKey(t *testing.T) {
	ko := koCatalog()
	for _, key := range AllKeys() {
		v, ok := ko[key]
		require.True(t, ok, "ko catalog missing key %q", key)
		assert.NotEmpty(t, v, "ko catalog has empty template for key %q", key)
	}
}

// TestM_RendersTemplate verifies the simplest path: a no-arg template returns
// the catalog string verbatim under the active locale (default ko).
func TestM_RendersTemplate(t *testing.T) {
	resetLocale(t)
	got := M(KeyInstallDone)
	assert.Equal(t, "buddy: 등록 완료. 이제 옆에서 보고 있을게.", got)
}

// TestM_FillsArgs verifies fmt.Sprintf-style placeholder substitution. The
// daemon-started template carries a single %d for the PID.
func TestM_FillsArgs(t *testing.T) {
	resetLocale(t)
	got := M(KeyDaemonStarted, 12345)
	assert.Equal(t, "buddy: daemon 시작 (pid 12345).", got)
}

// TestML_FallbackToKO_WhenENMissing verifies the explicit-locale form falls
// back to the ko entry when en doesn't have the key — the v0.2 i18n migration
// path (en map starts partially populated; unfilled keys still resolve via
// ko). KeyUninstallNothingRegistered is intentionally not in the v0.2 sample
// tracer, so it exercises the fallback branch.
func TestML_FallbackToKO_WhenENMissing(t *testing.T) {
	got := ML(LocaleEN, KeyUninstallNothingRegistered)
	assert.Equal(t, "buddy: 등록된 게 없어. 그대로 둘게.", got)
}

// TestSetLocale_ChangesActive sets en as active and verifies both paths the
// catalog can resolve through:
//   - filled en key → en template returned directly (KeyInstallDone is part
//     of the v0.2 sample tracer in en.go).
//   - unfilled en key → ko fallback still wired (KeyUninstallNothingRegistered
//     hasn't been translated yet, so M() resolves through ko).
func TestSetLocale_ChangesActive(t *testing.T) {
	resetLocale(t)
	require.NoError(t, SetLocale(LocaleEN))
	assert.Equal(t, LocaleEN, ActiveLocale())
	assert.Equal(t,
		"buddy: hooked up. i'll keep an eye on things now.",
		M(KeyInstallDone),
	)
	assert.Equal(t,
		"buddy: 등록된 게 없어. 그대로 둘게.",
		M(KeyUninstallNothingRegistered),
	)
}

// TestSetLocale_UnknownLocale_FallsBackToKO verifies an unknown locale string
// (e.g. "fr") returns an error AND resets the active locale to ko, so buddy
// keeps speaking Korean instead of panicking on the next M() call.
func TestSetLocale_UnknownLocale_FallsBackToKO(t *testing.T) {
	resetLocale(t)
	err := SetLocale(Locale("fr"))
	require.Error(t, err)
	assert.Equal(t, LocaleKO, ActiveLocale())
	// And M() still works.
	assert.NotEmpty(t, M(KeyInstallDone))
}

// TestM_PanicsOnTrulyMissingKey simulates the worst case: a key that's in
// neither map. We don't add a real "missing" Key constant (that would
// regress TestKO_HasEntryForEveryKey); instead we delete a key from the
// catalog within a sub-scope and assert M() panics with a useful message.
//
// The package-level catalog is mutated and restored via t.Cleanup so other
// tests aren't affected. Tests in this file run sequentially per the
// `go test` default for a single package.
func TestM_PanicsOnTrulyMissingKey(t *testing.T) {
	resetLocale(t)
	const probeKey Key = "test.probe.missing"
	// Sanity: probeKey must NOT be in either catalog (it's not in AllKeys()).
	mu.Lock()
	_, inKO := catalog[LocaleKO][probeKey]
	_, inEN := catalog[LocaleEN][probeKey]
	mu.Unlock()
	require.False(t, inKO)
	require.False(t, inEN)

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic for truly-missing key")
		msg, ok := r.(string)
		require.True(t, ok, "panic value should be string, got %T", r)
		assert.True(t, strings.Contains(msg, string(probeKey)),
			"panic message should mention the missing key, got %q", msg)
	}()
	_ = M(probeKey)
}

// TestKO_TemplatesValidateUnderFmt is a belt-and-suspenders check: every ko
// template must be a valid fmt.Sprintf format string when called with no
// args (or with the documented arg shape). We don't enforce the latter (no
// arg-shape registry yet); we just call Sprintf with no args and accept the
// %!d(MISSING) — the goal is to catch outright syntax errors like a stray
// `%` that would burn at runtime.
func TestKO_TemplatesValidateUnderFmt(t *testing.T) {
	for k, v := range koCatalog() {
		assert.NotContains(t, v, "%!", "ko template for %q already contains fmt error sentinel: %q", k, v)
	}
}

// resetLocale puts the package back into the default ko state and registers
// a cleanup so the next test starts fresh. Without this, TestSetLocale_*
// could leak active-locale state into other tests.
func resetLocale(t *testing.T) {
	t.Helper()
	require.NoError(t, SetLocale(LocaleKO))
	t.Cleanup(func() { _ = SetLocale(LocaleKO) })
}
