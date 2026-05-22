package main

import (
	"testing"

	"github.com/0xmhha/buddy/internal/feature"
	"github.com/stretchr/testify/require"
)

// validFeatureStatus is the CLI write-boundary guard. The set of allowed
// values must track internal/feature's constants exactly; this test pins
// the contract so a future status addition forces a corresponding update
// here (or removal of the constant fails the test loudly).
func TestValidFeatureStatus_AllowsKnownConstantsOnly(t *testing.T) {
	t.Parallel()

	allowed := []string{
		feature.StatusDraft,
		feature.StatusInProgress,
		feature.StatusDone,
		feature.StatusCancelled,
	}
	for _, s := range allowed {
		require.Truef(t, validFeatureStatus(s),
			"feature.Status%s (%q) must be accepted by the CLI guard", s, s)
	}

	rejected := []string{
		"",
		"todo",
		"in-progress",
		"DONE",
		"draft ",
		"random",
	}
	for _, s := range rejected {
		require.Falsef(t, validFeatureStatus(s),
			"unknown status %q must be rejected so the CLI fails fast", s)
	}
}
