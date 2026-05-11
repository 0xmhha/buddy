package analytics

import (
	"context"
	"errors"
)

// Adapter is the seven-query interface every analytics backend must satisfy.
// The MCP layer (internal/mcp/analytics_tool.go) holds an Adapter and
// translates JSON-RPC arguments to/from these typed requests.
//
// All methods take a context so callers can honour Claude Code's request
// cancellation. Implementations must propagate ctx into their underlying
// storage calls.
type Adapter interface {
	QueryFunnel(ctx context.Context, q FunnelQuery) (FunnelResult, error)
	QueryCohort(ctx context.Context, q CohortQuery) (CohortResult, error)
	QueryABExperiment(ctx context.Context, q ABExperimentQuery) (ABExperimentResult, error)
	QueryActorFailure(ctx context.Context, q ActorFailureQuery) (ActorFailureResult, error)
	QueryCost(ctx context.Context, q CostQuery) (CostResult, error)
	QuerySLOBurn(ctx context.Context, q SLOBurnQuery) (SLOBurnResult, error)
	QueryFeedbackCorpus(ctx context.Context, q FeedbackQuery) (FeedbackResult, error)
}

// ErrNotFound is returned by adapters when a specific entity (an experiment
// ID, an SLI) is missing. The MCP layer turns this into a 200-with-empty
// payload rather than an error so Claude can surface "no data" naturally.
var ErrNotFound = errors.New("analytics: not found")
