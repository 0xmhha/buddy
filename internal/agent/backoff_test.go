package agent

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ─── computeBackoff math ───────────────────────────────────────────────

func TestComputeBackoff_NilPolicyReturnsZero(t *testing.T) {
	t.Parallel()
	require.Equal(t, time.Duration(0), computeBackoff(nil, 1))
}

func TestComputeBackoff_ZeroDelayReturnsZero(t *testing.T) {
	t.Parallel()
	got := computeBackoff(&RetryPolicy{MaxAttempts: 3}, 1)
	require.Equal(t, time.Duration(0), got)
}

func TestComputeBackoff_FixedStrategyReturnsBase(t *testing.T) {
	t.Parallel()
	policy := &RetryPolicy{
		MaxAttempts:     5,
		BackoffDelay:    2 * time.Second,
		BackoffStrategy: BackoffStrategyFixed,
	}
	require.Equal(t, 2*time.Second, computeBackoff(policy, 1))
	require.Equal(t, 2*time.Second, computeBackoff(policy, 2))
	require.Equal(t, 2*time.Second, computeBackoff(policy, 4))
}

func TestComputeBackoff_EmptyStrategyDefaultsToFixed(t *testing.T) {
	t.Parallel()
	policy := &RetryPolicy{
		MaxAttempts:  5,
		BackoffDelay: 750 * time.Millisecond,
		// BackoffStrategy intentionally empty — v0.6.x backward compat
	}
	require.Equal(t, 750*time.Millisecond, computeBackoff(policy, 1))
	require.Equal(t, 750*time.Millisecond, computeBackoff(policy, 3))
}

func TestComputeBackoff_ExponentialDoublesPerAttempt(t *testing.T) {
	t.Parallel()
	policy := &RetryPolicy{
		MaxAttempts:     8,
		BackoffDelay:    100 * time.Millisecond,
		BackoffStrategy: BackoffStrategyExponential,
	}
	require.Equal(t, 100*time.Millisecond, computeBackoff(policy, 1)) // 2^0
	require.Equal(t, 200*time.Millisecond, computeBackoff(policy, 2)) // 2^1
	require.Equal(t, 400*time.Millisecond, computeBackoff(policy, 3)) // 2^2
	require.Equal(t, 800*time.Millisecond, computeBackoff(policy, 4)) // 2^3
}

func TestComputeBackoff_ExponentialCapsAtMax(t *testing.T) {
	t.Parallel()
	policy := &RetryPolicy{
		MaxAttempts:     10,
		BackoffDelay:    100 * time.Millisecond,
		BackoffStrategy: BackoffStrategyExponential,
		BackoffMax:      500 * time.Millisecond,
	}
	require.Equal(t, 100*time.Millisecond, computeBackoff(policy, 1)) // under cap
	require.Equal(t, 200*time.Millisecond, computeBackoff(policy, 2)) // under cap
	require.Equal(t, 400*time.Millisecond, computeBackoff(policy, 3)) // under cap
	require.Equal(t, 500*time.Millisecond, computeBackoff(policy, 4)) // would be 800, capped
	require.Equal(t, 500*time.Millisecond, computeBackoff(policy, 9)) // far past cap
}

func TestComputeBackoff_ExponentialUncappedWhenMaxZero(t *testing.T) {
	t.Parallel()
	policy := &RetryPolicy{
		MaxAttempts:     5,
		BackoffDelay:    time.Second,
		BackoffStrategy: BackoffStrategyExponential,
		// BackoffMax = 0 → no cap
	}
	require.Equal(t, 1*time.Second, computeBackoff(policy, 1))
	require.Equal(t, 16*time.Second, computeBackoff(policy, 5))
}

func TestComputeBackoff_LargeAttemptIsClampedNotOverflow(t *testing.T) {
	t.Parallel()
	policy := &RetryPolicy{
		MaxAttempts:     50,
		BackoffDelay:    time.Second,
		BackoffStrategy: BackoffStrategyExponential,
		BackoffMax:      time.Hour,
	}
	// attempt 40 fail would naïvely be 2^39 seconds (~17000 years).
	// computeBackoff clamps the exponent to 30 internally and the cap kicks
	// in well before that — observe: result must equal BackoffMax exactly.
	require.Equal(t, time.Hour, computeBackoff(policy, 40))
	// And even past the internal clamp, no overflow.
	require.Equal(t, time.Hour, computeBackoff(policy, 100))
}

func TestComputeBackoff_UnknownStrategyFallsThroughToBase(t *testing.T) {
	t.Parallel()
	// spec.go's validation should reject this at parse time, but the runtime
	// helper is also defensive — don't panic, just use the base delay so a
	// run that somehow has a bad strategy still finishes its retry loop.
	policy := &RetryPolicy{
		MaxAttempts:     3,
		BackoffDelay:    250 * time.Millisecond,
		BackoffStrategy: "decorrelated-jitter", // not implemented yet
	}
	require.Equal(t, 250*time.Millisecond, computeBackoff(policy, 1))
	require.Equal(t, 250*time.Millisecond, computeBackoff(policy, 2))
}

// ─── spec parsing of the new fields ─────────────────────────────────────

func TestParseSpec_AcceptsExponentialBackoffWithMax(t *testing.T) {
	t.Parallel()
	src := `
id: backoff-agent
name: "Backoff agent"
chain:
  - command: status
retry:
  max_attempts: 5
  backoff_delay: 500ms
  backoff_strategy: exponential
  backoff_max: 8s
`
	spec, err := ParseSpec([]byte(src))
	require.NoError(t, err)
	require.NotNil(t, spec.Retry)
	require.Equal(t, 5, spec.Retry.MaxAttempts)
	require.Equal(t, 500*time.Millisecond, spec.Retry.BackoffDelay)
	require.Equal(t, BackoffStrategyExponential, spec.Retry.BackoffStrategy)
	require.Equal(t, 8*time.Second, spec.Retry.BackoffMax)
}

func TestParseSpec_AcceptsFixedStrategyAndEmptyMax(t *testing.T) {
	t.Parallel()
	src := `
id: fixed-agent
name: "Fixed"
chain:
  - command: status
retry:
  max_attempts: 3
  backoff_delay: 1s
  backoff_strategy: fixed
`
	spec, err := ParseSpec([]byte(src))
	require.NoError(t, err)
	require.Equal(t, BackoffStrategyFixed, spec.Retry.BackoffStrategy)
	require.Equal(t, time.Duration(0), spec.Retry.BackoffMax,
		"backoff_max should default to zero (uncapped) when omitted")
}

func TestParseSpec_RejectsUnknownBackoffStrategy(t *testing.T) {
	t.Parallel()
	src := `
id: bad-agent
name: "Bad"
chain:
  - command: status
retry:
  max_attempts: 3
  backoff_delay: 1s
  backoff_strategy: gaussian-random
`
	_, err := ParseSpec([]byte(src))
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "backoff_strategy")
}

func TestParseSpec_RejectsInvertedBackoffMax(t *testing.T) {
	t.Parallel()
	src := `
id: inverted-agent
name: "Inverted"
chain:
  - command: status
retry:
  max_attempts: 3
  backoff_delay: 10s
  backoff_strategy: exponential
  backoff_max: 1s
`
	_, err := ParseSpec([]byte(src))
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "backoff_max")
}
