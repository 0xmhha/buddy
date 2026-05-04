package mcp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/wm-it-22-00661/buddy/internal/feature"
)

// ─── feature_list ───────────────────────────────────────────────────────────

type featureListArgs struct {
	Status string `json:"status" jsonschema:"Filter by status: draft, in_progress, done, cancelled. Empty = all."`
}

type featureListResult struct {
	Features []feature.Feature `json:"features"`
}

// ─── feature_get ────────────────────────────────────────────────────────────

type featureGetArgs struct {
	FeatureID string `json:"feature_id" jsonschema:"Feature ID to retrieve."`
}

type featureGetResult struct {
	Feature *feature.Feature `json:"feature"`
}

// ─── feature_upsert ─────────────────────────────────────────────────────────

type featureUpsertArgs struct {
	FeatureID          string           `json:"feature_id"           jsonschema:"Unique feature identifier."`
	Name               string           `json:"name"                 jsonschema:"Short display name."`
	Summary            string           `json:"summary"              jsonschema:"One-paragraph description."`
	Actors             []feature.Actor  `json:"actors"               jsonschema:"List of actor–system-boundary pairs."`
	AcceptanceCriteria []string         `json:"acceptance_criteria"  jsonschema:"Ordered acceptance criteria."`
	TestPlan           feature.TestPlan `json:"test_plan"            jsonschema:"Unit/integration/e2e test descriptions."`
	Status             string           `json:"status"               jsonschema:"draft | in_progress | done | cancelled. Default: draft."`
}

type featureUpsertResult struct {
	FeatureID string `json:"feature_id"`
}

// ─── feature_delete ─────────────────────────────────────────────────────────

type featureDeleteArgs struct {
	FeatureID string `json:"feature_id" jsonschema:"Feature ID to delete."`
}

type featureDeleteResult struct {
	FeatureID string `json:"feature_id"`
}

// ─── feature_search ─────────────────────────────────────────────────────────

type featureSearchArgs struct {
	Query string `json:"query" jsonschema:"Case-insensitive substring to match against name or summary."`
}

type featureSearchResult struct {
	Features []feature.Feature `json:"features"`
}

// ─── registration ────────────────────────────────────────────────────────────

func addFeatureTools(s *mcp.Server, opts Options) {
	fo := feature.Options{DBPath: opts.DBPath}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "feature_list",
		Description: "List features stored in the local buddy DB. Optionally filter by status.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args featureListArgs) (*mcp.CallToolResult, featureListResult, error) {
		features, err := feature.List(ctx, fo, args.Status)
		if err != nil {
			return nil, featureListResult{}, err
		}
		text := renderFeatureTable(features)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, featureListResult{Features: features}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "feature_get",
		Description: "Get a single feature by ID from the local buddy DB.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args featureGetArgs) (*mcp.CallToolResult, featureGetResult, error) {
		f, err := feature.Get(ctx, fo, args.FeatureID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("feature not found: %s", args.FeatureID)}},
				}, featureGetResult{}, nil
			}
			return nil, featureGetResult{}, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: renderFeatureDetail(f)}},
		}, featureGetResult{Feature: &f}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "feature_upsert",
		Description: "Insert or update a feature in the local buddy DB. Updates all fields on conflict.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args featureUpsertArgs) (*mcp.CallToolResult, featureUpsertResult, error) {
		status := args.Status
		if status == "" {
			status = feature.StatusDraft
		}
		f := feature.Feature{
			FeatureID:          args.FeatureID,
			Name:               args.Name,
			Summary:            args.Summary,
			Actors:             args.Actors,
			AcceptanceCriteria: args.AcceptanceCriteria,
			TestPlan:           args.TestPlan,
			Status:             status,
		}
		if err := feature.Upsert(ctx, fo, f); err != nil {
			return nil, featureUpsertResult{}, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("saved feature: %s", args.FeatureID)}},
		}, featureUpsertResult{FeatureID: args.FeatureID}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "feature_delete",
		Description: "Delete a feature from the local buddy DB. Idempotent — returns success even if the feature does not exist.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args featureDeleteArgs) (*mcp.CallToolResult, featureDeleteResult, error) {
		if err := feature.Delete(ctx, fo, args.FeatureID); err != nil {
			return nil, featureDeleteResult{}, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("deleted feature: %s", args.FeatureID)}},
		}, featureDeleteResult{FeatureID: args.FeatureID}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "feature_search",
		Description: "Search features by name or summary substring (case-insensitive).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args featureSearchArgs) (*mcp.CallToolResult, featureSearchResult, error) {
		features, err := feature.Search(ctx, fo, args.Query)
		if err != nil {
			return nil, featureSearchResult{}, err
		}
		text := renderFeatureTable(features)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, featureSearchResult{Features: features}, nil
	})
}

// ─── render helpers ──────────────────────────────────────────────────────────

func renderFeatureTable(features []feature.Feature) string {
	if len(features) == 0 {
		return "No features found."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-20s %-10s %s\n", "ID", "STATUS", "NAME"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	for _, f := range features {
		sb.WriteString(fmt.Sprintf("%-20s %-10s %s\n", f.FeatureID, f.Status, f.Name))
	}
	return sb.String()
}

func renderFeatureDetail(f feature.Feature) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ID:      %s\n", f.FeatureID))
	sb.WriteString(fmt.Sprintf("Name:    %s\n", f.Name))
	sb.WriteString(fmt.Sprintf("Status:  %s\n", f.Status))
	sb.WriteString(fmt.Sprintf("Summary: %s\n", f.Summary))
	if len(f.Actors) > 0 {
		sb.WriteString("Actors:\n")
		for _, a := range f.Actors {
			sb.WriteString(fmt.Sprintf("  - %s (%s)\n", a.ID, a.SystemBoundary))
		}
	}
	if len(f.AcceptanceCriteria) > 0 {
		sb.WriteString("AC:\n")
		for _, c := range f.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("  - %s\n", c))
		}
	}
	return sb.String()
}
