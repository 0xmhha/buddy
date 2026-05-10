package pricing_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0xmhha/buddy/internal/pricing"
	"github.com/0xmhha/buddy/internal/schema"
)

// TestEstimate_Zero verifies the zero TokenUsage returns zero cost across
// every component, no float drift.
func TestEstimate_Zero(t *testing.T) {
	p := pricing.ModelPrice{
		InputCentsPerMTok:       1500,
		OutputCentsPerMTok:      7500,
		CacheReadCentsPerMTok:   150,
		CacheCreateCentsPerMTok: 1875,
	}
	assert.Equal(t, int64(0), p.Estimate(schema.TokenUsage{}))
}

// TestEstimate_OneMillionTokensEach is the unit-price sanity check: feeding
// exactly 1M tokens through each component yields the per-component cents
// values directly. Confirms Estimate's per-million division is correct and
// each component lands in its own term.
func TestEstimate_OneMillionTokensEach(t *testing.T) {
	p := pricing.ModelPrice{
		InputCentsPerMTok:       1500,
		OutputCentsPerMTok:      7500,
		CacheReadCentsPerMTok:   150,
		CacheCreateCentsPerMTok: 1875,
	}
	usage := schema.TokenUsage{
		InputTokens:       1_000_000,
		OutputTokens:      1_000_000,
		CacheReadTokens:   1_000_000,
		CacheCreateTokens: 1_000_000,
	}
	// 1500 + 7500 + 150 + 1875 = 11,025 cents = $110.25
	assert.Equal(t, int64(11_025), p.Estimate(usage))
}

// TestEstimate_RealisticOpusUsage spot-checks a believable session shape:
// 250k input, 30k output, 1M cache read, 50k cache create. Numbers are
// computed manually so the test catches reordering or coefficient swaps
// in Estimate.
func TestEstimate_RealisticOpusUsage(t *testing.T) {
	p := pricing.AnthropicPricing["claude-opus-4-7"]
	usage := schema.TokenUsage{
		InputTokens:       250_000,
		OutputTokens:      30_000,
		CacheReadTokens:   1_000_000,
		CacheCreateTokens: 50_000,
	}
	// input:  250_000  * 1500 / 1e6 = 375
	// output: 30_000   * 7500 / 1e6 = 225
	// cache read: 1_000_000 * 150 / 1e6 = 150
	// cache create: 50_000 * 1875 / 1e6 = 93   (50_000*1875=93_750_000; /1e6=93)
	// total = 375 + 225 + 150 + 93 = 843 cents = $8.43
	assert.Equal(t, int64(843), p.Estimate(usage))
}

// TestLookup_KnownModels verifies every documented model name resolves,
// guarding against typos in the table keys.
func TestLookup_KnownModels(t *testing.T) {
	for _, name := range []string{
		"claude-opus-4-7",
		"claude-sonnet-4-6",
		"claude-haiku-4-5",
	} {
		_, ok := pricing.Lookup(name)
		assert.True(t, ok, "Lookup(%q) should hit", name)
	}
}

// TestLookup_UnknownModel verifies an unrecognized name returns ok=false
// instead of a default rate. Callers must decide how to surface the gap.
func TestLookup_UnknownModel(t *testing.T) {
	_, ok := pricing.Lookup("claude-future-99-9")
	assert.False(t, ok)
}
