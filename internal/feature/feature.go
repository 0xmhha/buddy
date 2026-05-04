package feature

// Feature represents a planned or in-progress product feature in the local buddy DB.
type Feature struct {
	FeatureID          string   `json:"feature_id"`
	Name               string   `json:"name"`
	Summary            string   `json:"summary"`
	Actors             []Actor  `json:"actors"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	TestPlan           TestPlan `json:"test_plan"`
	Status             string   `json:"status"`
	UpdatedAt          int64    `json:"updated_at"` // unix ms
}

// Actor is a system-boundary participant in a feature's actor-track plan.
type Actor struct {
	ID             string `json:"id"`
	SystemBoundary string `json:"system_boundary"`
}

// TestPlan holds the structured test cases for a feature, grouped by scope.
type TestPlan struct {
	Unit        []string `json:"unit"`
	Integration []string `json:"integration"`
	E2E         []string `json:"e2e"`
}

// Status constants for Feature.Status.
const (
	StatusDraft      = "draft"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
	StatusCancelled  = "cancelled"
)
