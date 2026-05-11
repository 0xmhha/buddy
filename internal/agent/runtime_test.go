package agent

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"

	"github.com/0xmhha/buddy/internal/db"
)

// newTestStore opens a fresh on-disk SQLite (in-memory + cache=shared is flaky
// across goroutines for FK cascade tests) and runs the buddy migrations so
// the agents / agent_runs / agent_logs tables exist.
func newTestStore(t *testing.T) (*Store, *sql.DB) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "agent_test.db")
	conn, err := db.Open(db.Options{Path: dbPath})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return NewStore(conn), conn
}

const minimalSpecYAML = `
id: hello-agent
name: "Hello agent"
chain:
  - command: status
`

const twoStepSpecYAML = `
id: chain-agent
name: "Chain agent"
chain:
  - command: status
    args: ""
  - command: concretize-idea
    args: "test idea"
`

// ─── spec parsing ─────────────────────────────────────────────────────────

func TestParseSpec_MinimalValid(t *testing.T) {
	t.Parallel()
	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	require.Equal(t, "hello-agent", spec.ID)
	require.Equal(t, "Hello agent", spec.Name)
	require.Len(t, spec.Chain, 1)
	require.Equal(t, "status", spec.Chain[0].Command)
}

func TestParseSpec_RejectsEmptyID(t *testing.T) {
	t.Parallel()
	_, err := ParseSpec([]byte(`name: foo
chain:
  - command: status`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "spec.id is required")
}

func TestParseSpec_RejectsBadIDFormat(t *testing.T) {
	t.Parallel()
	_, err := ParseSpec([]byte(`id: BadID
name: foo
chain:
  - command: status`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "must match")
}

func TestParseSpec_RejectsEmptyChain(t *testing.T) {
	t.Parallel()
	_, err := ParseSpec([]byte(`id: foo
name: bar
chain: []`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "spec.chain must contain at least one step")
}

func TestParseSpec_RejectsFileOutputWithoutPath(t *testing.T) {
	t.Parallel()
	_, err := ParseSpec([]byte(`id: foo
name: bar
chain:
  - command: status
output:
  type: file
`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "output.path is required")
}

func TestNormalizeCommand_StripsBuddyPrefix(t *testing.T) {
	t.Parallel()
	require.Equal(t, "status", NormalizeCommand("/buddy:status"))
	require.Equal(t, "status", NormalizeCommand("/status"))
	require.Equal(t, "status", NormalizeCommand("  status  "))
}

// ─── store CRUD ───────────────────────────────────────────────────────────

func TestStore_CreateGetListDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	created, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)
	require.Equal(t, "hello-agent", created.ID)
	require.Equal(t, StatusIdle, created.Status)

	got, err := store.Get(ctx, "hello-agent")
	require.NoError(t, err)
	require.Equal(t, "Hello agent", got.Name)
	require.Equal(t, StatusIdle, got.Status)

	list, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "hello-agent", list[0].ID)

	require.NoError(t, store.Delete(ctx, "hello-agent"))
	_, err = store.Get(ctx, "hello-agent")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestStore_DeleteMissingReturnsNotFound(t *testing.T) {
	t.Parallel()
	store, _ := newTestStore(t)
	require.ErrorIs(t, store.Delete(context.Background(), "ghost"), ErrNotFound)
}

// ─── runtime ──────────────────────────────────────────────────────────────

func TestRuntime_Run_HappyPathPersistsRunAndLogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, conn := newTestStore(t)

	spec, err := ParseSpec([]byte(twoStepSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, twoStepSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: "현재 phase: §7", ExitCode: 0}
	mock.Responses["concretize-idea"] = MockResponse{Stdout: "Stage 1 ok", ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)
	require.Equal(t, 0, res.ExitCode)
	require.Len(t, res.Steps, 2)
	require.Equal(t, "status", res.Steps[0].Command)
	require.Equal(t, "concretize-idea", res.Steps[1].Command)
	require.Equal(t, []MockCall{{"status", ""}, {"concretize-idea", "test idea"}}, mock.Calls)

	// Agent transitioned idle → done.
	after, err := store.Get(ctx, agent.ID)
	require.NoError(t, err)
	require.Equal(t, StatusDone, after.Status)

	// agent_runs ended_at populated and result_json non-empty.
	var ended sql.NullInt64
	var resultJSON string
	require.NoError(t, conn.QueryRow(
		`SELECT ended_at, result_json FROM agent_runs WHERE id = ?`, res.RunID,
	).Scan(&ended, &resultJSON))
	require.True(t, ended.Valid)
	require.Contains(t, resultJSON, `"command":"status"`)
	require.Contains(t, resultJSON, `"command":"concretize-idea"`)

	// Logs include one start + per-step attempt + per-step ok + one finish.
	logs, err := store.Logs(ctx, res.RunID, 0)
	require.NoError(t, err)
	var startedSeen, finishedSeen bool
	for _, l := range logs {
		if strings.Contains(l.Message, "started") {
			startedSeen = true
		}
		if strings.Contains(l.Message, "finished status=done") {
			finishedSeen = true
		}
	}
	require.True(t, startedSeen, "expected started log line, got %+v", logs)
	require.True(t, finishedSeen, "expected finished log line, got %+v", logs)
}

func TestRuntime_Run_StopsOnNonZeroExitCode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(twoStepSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, twoStepSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: "", ExitCode: 1, Stderr: "boom"}
	mock.Responses["concretize-idea"] = MockResponse{Stdout: "should not run", ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)
	require.NotEqual(t, 0, res.ExitCode)
	require.Len(t, res.Steps, 1, "second step must not execute after first fails")
	require.Len(t, mock.Calls, 1, "only one executor call expected")

	after, err := store.Get(ctx, agent.ID)
	require.NoError(t, err)
	require.Equal(t, StatusFailed, after.Status)
}

func TestRuntime_Run_RetryHonoursMaxAttempts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: retry-agent
name: "Retry agent"
chain:
  - command: status
retry:
  max_attempts: 3
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	// Always-fail executor so retries exhaust.
	mock := NewMockExecutor()
	mock.Default = MockResponse{ExitCode: 7}
	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)
	require.Equal(t, 7, res.ExitCode)
	require.Len(t, mock.Calls, 3, "expected 3 attempts before giving up, got %d", len(mock.Calls))
	require.Equal(t, 3, res.Steps[0].Attempt)
}

func TestRuntime_Run_ExecutorErrorMarksFailed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	bang := errors.New("subprocess refused to start")
	mock.Responses["status"] = MockResponse{Err: bang, ExitCode: -1}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.Error(t, err)
	require.ErrorIs(t, err, bang)
	require.NotEqual(t, 0, res.ExitCode)

	after, err := store.Get(ctx, agent.ID)
	require.NoError(t, err)
	require.Equal(t, StatusFailed, after.Status)
}

// TestRuntime_Run_ParsesSelfCheckAndNextPhase verifies the W3-4 wire-up:
// when the executor returns a Claude-style stdout containing §self-check
// and §next phase sections, the Runtime parses them into StepResult.Parsed
// and writes a self-check log line. v0.3 contract: parse result is metadata
// only — step success still tracks ExitCode.
func TestRuntime_Run_ParsesSelfCheckAndNextPhase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{
		Stdout: "## 5. 산출물 형식\n\n결과 본문\n\n" +
			"## 6. 검증 (self-check)\n\n" +
			"- [x] 항목 1 통과\n" +
			"- [x] 항목 2 통과\n\n" +
			"## 7. 다음 phase\n\n" +
			"- `define-features` 가 다음 entry\n",
		ExitCode: 0,
	}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1)

	parsed := res.Steps[0].Parsed
	require.Equal(t, SelfCheckPass, parsed.SelfCheck.Verdict)
	require.Equal(t, 2, parsed.SelfCheck.Total)
	require.Equal(t, []string{"define-features"}, parsed.NextPhase.Skills)

	// Run logs include the self-check summary line.
	logs, err := store.Logs(ctx, res.RunID, 0)
	require.NoError(t, err)
	var selfCheckLogged bool
	for _, l := range logs {
		if strings.Contains(l.Message, "self-check=pass (2/2 passed)") {
			selfCheckLogged = true
		}
	}
	require.True(t, selfCheckLogged, "expected self-check log line, got %+v", logs)
}

// TestRuntime_Run_ParsesConditionalBranches covers the W3-4 follow-on
// (conditional next-phase parse): when the Claude output describes
// branch-style cascade rules, Runtime persists Branches into StepResult
// and emits one log line per branch instead of (or alongside) the
// candidates union line.
func TestRuntime_Run_ParsesConditionalBranches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{
		Stdout: "## 6. 검증\n\n- [x] decided\n\n" +
			"## 7. 다음 phase\n\n" +
			"- 글로벌 → `review-legal-regulatory`\n" +
			"- Korea → `consult-korea-legal-context` + `review-legal-regulatory`\n" +
			"- USA / EU / 기타 → (template 작성 필요)\n",
		ExitCode: 0,
	}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1)

	parsed := res.Steps[0].Parsed
	require.Len(t, parsed.NextPhase.Branches, 3)
	require.Equal(t, "글로벌", parsed.NextPhase.Branches[0].Condition)
	require.Equal(t, []string{"review-legal-regulatory"}, parsed.NextPhase.Branches[0].Skills)
	require.Equal(t, "Korea", parsed.NextPhase.Branches[1].Condition)
	require.Empty(t, parsed.NextPhase.Branches[2].Skills,
		"no-skill branch keeps condition but empty skills")

	logs, err := store.Logs(ctx, res.RunID, 0)
	require.NoError(t, err)
	var globalLogged, koreaLogged, noSkillLogged bool
	for _, l := range logs {
		switch {
		case strings.Contains(l.Message, `next-phase branch: "글로벌" → review-legal-regulatory`):
			globalLogged = true
		case strings.Contains(l.Message, `next-phase branch: "Korea" → consult-korea-legal-context, review-legal-regulatory`):
			koreaLogged = true
		case strings.Contains(l.Message, `next-phase branch: "USA / EU / 기타" → (no skill)`):
			noSkillLogged = true
		}
	}
	require.True(t, globalLogged, "expected 글로벌 branch log line, got %+v", logs)
	require.True(t, koreaLogged, "expected Korea branch log line, got %+v", logs)
	require.True(t, noSkillLogged, "expected USA/EU/기타 branch log line, got %+v", logs)
}

func TestRuntime_Run_FileOutputWritesResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "result.json")
	yamlSrc := `
id: file-out-agent
name: "File output"
chain:
  - command: status
output:
  type: file
  path: ` + outPath
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: "ok", ExitCode: 0}
	rt := NewRuntime(store, mock)
	_, err = rt.Run(ctx, agent)
	require.NoError(t, err)

	// File exists and contains the marshalled RunResult.
	info, err := stat(outPath)
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(10))
}

// ─── helpers ──────────────────────────────────────────────────────────────

// stat is a thin wrapper around os.Stat so the test file does not have to
// import os for one symbol (keeps the import set minimal + linter-quiet).
func stat(path string) (statResult, error) {
	info, err := osStat(path)
	if err != nil {
		return statResult{}, err
	}
	return statResult{size: info.size(), modTime: info.modTime()}, nil
}

// Indirection layer used by stat() — kept tiny so we can avoid importing
// the os package directly in this file (we only need file existence + size).
type statResult struct {
	size    int64
	modTime time.Time
}

func (s statResult) Size() int64        { return s.size }
func (s statResult) ModTime() time.Time { return s.modTime }
