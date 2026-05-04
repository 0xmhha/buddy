package feature_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/feature"
)

// fixtures -----------------------------------------------------------------

func newTempDBPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "buddy.db")
	conn, err := db.Open(db.Options{Path: path})
	require.NoError(t, err)
	require.NoError(t, conn.Close())
	return path
}

func opts(dbPath string) feature.Options {
	return feature.Options{DBPath: dbPath}
}

func sampleFeature(id string) feature.Feature {
	return feature.Feature{
		FeatureID: id,
		Name:      "Test Feature " + id,
		Summary:   "summary for " + id,
		Actors: []feature.Actor{
			{ID: "backend", SystemBoundary: "backend-api-service"},
		},
		AcceptanceCriteria: []string{"criterion A", "criterion B"},
		TestPlan: feature.TestPlan{
			Unit:        []string{"unit test 1"},
			Integration: []string{"integration test 1"},
			E2E:         []string{"e2e test 1"},
		},
		Status: feature.StatusDraft,
	}
}

// Upsert -------------------------------------------------------------------

func TestUpsert_Insert(t *testing.T) {
	dbPath := newTempDBPath(t)
	f := sampleFeature("feat-001")

	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), f))

	got, err := feature.Get(context.Background(), opts(dbPath), "feat-001")
	require.NoError(t, err)
	assert.Equal(t, f.FeatureID, got.FeatureID)
	assert.Equal(t, f.Name, got.Name)
	assert.Equal(t, f.Summary, got.Summary)
	assert.Equal(t, f.Actors, got.Actors)
	assert.Equal(t, f.AcceptanceCriteria, got.AcceptanceCriteria)
	assert.Equal(t, f.TestPlan, got.TestPlan)
	assert.Equal(t, feature.StatusDraft, got.Status)
	assert.Greater(t, got.UpdatedAt, int64(0))
}

func TestUpsert_UpdateExisting(t *testing.T) {
	dbPath := newTempDBPath(t)
	f := sampleFeature("feat-002")
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), f))

	f.Name = "Updated Name"
	f.Status = feature.StatusInProgress
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), f))

	got, err := feature.Get(context.Background(), opts(dbPath), "feat-002")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", got.Name)
	assert.Equal(t, feature.StatusInProgress, got.Status)
}

// Get ----------------------------------------------------------------------

func TestGet_NotFound(t *testing.T) {
	dbPath := newTempDBPath(t)
	_, err := feature.Get(context.Background(), opts(dbPath), "missing-id")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// List ---------------------------------------------------------------------

func TestList_Empty(t *testing.T) {
	dbPath := newTempDBPath(t)
	features, err := feature.List(context.Background(), opts(dbPath), "")
	require.NoError(t, err)
	assert.Empty(t, features)
}

func TestList_AllStatuses(t *testing.T) {
	dbPath := newTempDBPath(t)
	ids := []string{"feat-A", "feat-B", "feat-C"}
	for _, id := range ids {
		require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), sampleFeature(id)))
	}

	features, err := feature.List(context.Background(), opts(dbPath), "")
	require.NoError(t, err)
	assert.Len(t, features, 3)
}

func TestList_FilterByStatus(t *testing.T) {
	dbPath := newTempDBPath(t)

	draft := sampleFeature("feat-draft")
	draft.Status = feature.StatusDraft
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), draft))

	done := sampleFeature("feat-done")
	done.Status = feature.StatusDone
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), done))

	features, err := feature.List(context.Background(), opts(dbPath), feature.StatusDone)
	require.NoError(t, err)
	require.Len(t, features, 1)
	assert.Equal(t, "feat-done", features[0].FeatureID)
}

// Delete -------------------------------------------------------------------

func TestDelete_ExistingFeature(t *testing.T) {
	dbPath := newTempDBPath(t)
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), sampleFeature("feat-del")))

	require.NoError(t, feature.Delete(context.Background(), opts(dbPath), "feat-del"))

	_, err := feature.Get(context.Background(), opts(dbPath), "feat-del")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDelete_NotFound_Idempotent(t *testing.T) {
	dbPath := newTempDBPath(t)
	// Deleting a non-existent feature must return nil.
	require.NoError(t, feature.Delete(context.Background(), opts(dbPath), "ghost"))
}

// Search -------------------------------------------------------------------

func TestSearch_ByName(t *testing.T) {
	dbPath := newTempDBPath(t)
	f := sampleFeature("feat-search")
	f.Name = "Signup Flow"
	f.Summary = "implements user registration"
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), f))

	results, err := feature.Search(context.Background(), opts(dbPath), "signup")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "feat-search", results[0].FeatureID)
}

func TestSearch_BySummary(t *testing.T) {
	dbPath := newTempDBPath(t)
	f := sampleFeature("feat-search2")
	f.Name = "Billing Integration"
	f.Summary = "stripe payment webhook"
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), f))

	results, err := feature.Search(context.Background(), opts(dbPath), "webhook")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "feat-search2", results[0].FeatureID)
}

func TestSearch_CaseInsensitive(t *testing.T) {
	dbPath := newTempDBPath(t)
	f := sampleFeature("feat-ci")
	f.Name = "OAuth Login"
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), f))

	results, err := feature.Search(context.Background(), opts(dbPath), "oauth")
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearch_NoMatch(t *testing.T) {
	dbPath := newTempDBPath(t)
	require.NoError(t, feature.Upsert(context.Background(), opts(dbPath), sampleFeature("feat-nm")))

	results, err := feature.Search(context.Background(), opts(dbPath), "xyzzy")
	require.NoError(t, err)
	assert.Empty(t, results)
}
