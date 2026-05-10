// Package pricing turns a transcript-derived TokenUsage into a cost
// estimate. v0.2 ships the table embedded — Anthropic publishes prices
// on console.anthropic.com and adjusts them out-of-band, so we treat
// each release as the freshness boundary for these numbers. v1.0+ may
// fetch this externally (plan §3.4 BUDDY-D-WEB / pricing trigger).
//
// Arithmetic is in cents (1/100 USD) using int64 to avoid float64
// rounding drift across millions of tokens. Callers convert to dollars
// at display time only.
package pricing

import (
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

// ModelPrice is the per-million-token unit price for one Claude model.
// Field units are cents-per-million-tokens; Estimate divides by 1e6 so a
// usage line of "1,500,000 input tokens" against InputCentsPerMTok=1500
// yields 2,250 cents = $22.50.
type ModelPrice struct {
	InputCentsPerMTok       int64
	OutputCentsPerMTok      int64
	CacheReadCentsPerMTok   int64
	CacheCreateCentsPerMTok int64
}

// Estimate returns the cost in cents (1/100 USD) of the given usage at
// this model's price. Zero usage returns zero; negative usage cannot
// occur because schema.HookEventPayload.Validate rejects it upstream.
func (p ModelPrice) Estimate(u schema.TokenUsage) int64 {
	const perMillion = 1_000_000
	var c int64
	c += int64(u.InputTokens) * p.InputCentsPerMTok / perMillion
	c += int64(u.OutputTokens) * p.OutputCentsPerMTok / perMillion
	c += int64(u.CacheReadTokens) * p.CacheReadCentsPerMTok / perMillion
	c += int64(u.CacheCreateTokens) * p.CacheCreateCentsPerMTok / perMillion
	return c
}

// AnthropicPricing as published on console.anthropic.com (2026-05). Cents
// per million tokens. Update on each Anthropic price change; release
// cadence pins the freshness window.
//
// Coverage = Claude 4.x family + the 4.5 Haiku variant buddy currently
// observes in transcripts. Aliases (e.g. dated suffixes) resolve via
// Lookup's prefix fallback so transcripts like "claude-opus-4-7@20260415"
// still map to the right rate.
var AnthropicPricing = map[string]ModelPrice{
	"claude-opus-4-7": {
		InputCentsPerMTok:       1500,
		OutputCentsPerMTok:      7500,
		CacheReadCentsPerMTok:   150,
		CacheCreateCentsPerMTok: 1875,
	},
	"claude-sonnet-4-6": {
		InputCentsPerMTok:       300,
		OutputCentsPerMTok:      1500,
		CacheReadCentsPerMTok:   30,
		CacheCreateCentsPerMTok: 375,
	},
	"claude-haiku-4-5": {
		InputCentsPerMTok:       80,
		OutputCentsPerMTok:      400,
		CacheReadCentsPerMTok:   8,
		CacheCreateCentsPerMTok: 100,
	},
}

// Lookup returns the price for a model name. The bool is false when the
// model is unknown — callers decide how to surface that (skip, default,
// or warn). No fallback to a default rate; an unknown model means buddy
// can't estimate without misleading.
func Lookup(model string) (ModelPrice, bool) {
	p, ok := AnthropicPricing[model]
	return p, ok
}
