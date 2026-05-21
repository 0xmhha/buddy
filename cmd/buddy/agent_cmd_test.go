package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/agent"
	"github.com/0xmhha/buddy/internal/db"
)

// runAgentEdit drives `buddy agent edit <id>` through cobra with the
// editorRunner seam swapped for a fake that calls editorFn against the
// temp file. The fake observes whatever bytes the cmd wrote to the file
// and may replace them — exactly what a real editor would do.
func runAgentEdit(t *testing.T, dbPath, id string, editorFn func(string) []byte) (string, error) {
	t.Helper()
	originalRunner := editorRunner
	editorRunner = func(_, path string) error {
		current, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(path, editorFn(string(current)), 0o600)
	}
	t.Cleanup(func() { editorRunner = originalRunner })

	cmd := newAgentEditCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--db", dbPath, id})
	err := cmd.Execute()
	return stdout.String(), err
}

const agentEditTestSpec = `id: test-agent
name: "test agent"
schedule: ""
chain:
  - command: define-features
output:
  type: stdout
`

func seedAgent(t *testing.T, dbPath, yamlBody string) *agent.Agent {
	t.Helper()
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	store := agent.NewStore(conn)
	spec, err := agent.ParseSpec([]byte(yamlBody))
	require.NoError(t, err)
	created, err := store.Create(context.Background(), spec, yamlBody)
	require.NoError(t, err)
	return &created
}

// TestAgentEdit_PersistsModifiedSpec — the happy path: editor replaces
// the name field, command parses + saves + reports success.
func TestAgentEdit_PersistsModifiedSpec(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "edit.db")
	seedAgent(t, dbPath, agentEditTestSpec)

	stdout, err := runAgentEdit(t, dbPath, "test-agent", func(orig string) []byte {
		return []byte(strings.Replace(orig, `name: "test agent"`, `name: "renamed via edit"`, 1))
	})
	require.NoError(t, err)
	assert.Contains(t, stdout, "test-agent 저장 완료")

	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	defer conn.Close()
	store := agent.NewStore(conn)
	got, err := store.Get(context.Background(), "test-agent")
	require.NoError(t, err)
	assert.Equal(t, "renamed via edit", got.Name, "the edit must reach the DB")
	assert.Contains(t, got.SpecYAML, "renamed via edit", "spec_yaml must hold the edited body")
}

// TestAgentEdit_RejectsRename — when the editor changes the id field
// (deliberately or by accident) the cmd refuses to save, mirroring the
// TUI's same guard. Renames orphan runs + logs by FK and are not
// supported anywhere in the edit flow.
func TestAgentEdit_RejectsRename(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "rename.db")
	seedAgent(t, dbPath, agentEditTestSpec)

	_, err := runAgentEdit(t, dbPath, "test-agent", func(orig string) []byte {
		return []byte(strings.Replace(orig, "id: test-agent", "id: other-agent", 1))
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rename not allowed")
}

// TestAgentEdit_NotFoundSurfacesAsError — editing a missing agent must
// fail loudly via Store.Get's ErrNotFound; we never reach the editor.
func TestAgentEdit_NotFoundSurfacesAsError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "missing.db")
	// Open + migrate so the agents table exists, but seed nothing.
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	editorCalled := false
	_, err = runAgentEdit(t, dbPath, "does-not-exist", func(orig string) []byte {
		editorCalled = true
		return []byte(orig)
	})
	require.Error(t, err)
	assert.False(t, editorCalled, "Store.Get must fail before we open the editor")
}

// TestResolveEditor_OrdersEnvFallback locks the EDITOR → VISUAL → vi
// preference order, matching the TUI's resolution so users see the same
// editor whether they edit through the TUI or the CLI.
func TestResolveEditor_OrdersEnvFallback(t *testing.T) {
	cases := []struct {
		name   string
		editor string
		visual string
		want   string
	}{
		{"EDITOR wins", "nano", "emacs", "nano"},
		{"VISUAL when EDITOR empty", "", "emacs", "emacs"},
		{"vi when both empty", "", "", "vi"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("EDITOR", tc.editor)
			t.Setenv("VISUAL", tc.visual)
			assert.Equal(t, tc.want, resolveEditor())
		})
	}
}

// Silence the unused-import warning if the test file is ever pared down.
var _ = fmt.Sprintf
