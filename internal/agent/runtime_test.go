package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// TestParseSpec_ReferenceWebtoonAgentValidates pins the reference
// example to ParseSpec so a future schema change can't silently break
// the shipped example. The example sits in examples/webtoon-agent/spec.yaml
// and exercises every v0.6.x cli buddy capability (retry exponential,
// webhook output, multi-step chain). Path is relative to this test file.
func TestParseSpec_ReferenceWebtoonAgentValidates(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("../../examples/webtoon-agent/spec.yaml")
	require.NoError(t, err, "the reference webtoon agent example must remain on disk")

	spec, err := ParseSpec(body)
	require.NoError(t, err, "examples/webtoon-agent/spec.yaml must always parse cleanly")
	require.Equal(t, "webtoon-publish", spec.ID)
	require.Equal(t, "0 3 * * *", spec.Schedule)
	require.Len(t, spec.Chain, 4)

	require.NotNil(t, spec.Retry)
	require.Equal(t, BackoffStrategyExponential, spec.Retry.BackoffStrategy)
	require.Equal(t, 2*time.Minute, spec.Retry.BackoffMax)

	require.NotNil(t, spec.Output)
	require.Equal(t, "webhook", spec.Output.Type)
	require.Contains(t, spec.Output.URL, "https://")
	require.Equal(t, "POST", spec.Output.Method)
	require.Equal(t, 90*time.Second, spec.Output.Timeout)
	require.NotEmpty(t, spec.Output.Headers["Authorization"])
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

// TestRuntime_Run_ParsesSelfCheckAndNextPhase verifies the parser wire-up:
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

// TestRuntime_Run_ParsesConditionalBranches covers the follow-on parser
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

// TestRuntime_Run_StreamsStdoutLinesToAgentLogs covers Tier 1.5 streaming
// log capture: each line emitted by the executor surfaces in agent_logs in
// the order it was produced, with one log entry per line. This is what
// makes `buddy agent log <id>` show mid-progress on long-running steps
// instead of only the final post-step summary.
func TestRuntime_Run_StreamsStdoutLinesToAgentLogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{
		Stdout:   "compiling…\nlinking…\nready\n",
		Stderr:   "deprecation: foo() will be removed\n",
		ExitCode: 0,
	}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1)

	logs, err := store.Logs(ctx, res.RunID, 0)
	require.NoError(t, err)

	var stdoutLines []string
	var stderrLines []string
	for _, l := range logs {
		switch {
		case strings.Contains(l.Message, " stdout: "):
			stdoutLines = append(stdoutLines, l.Message)
		case strings.Contains(l.Message, " stderr: "):
			stderrLines = append(stderrLines, l.Message)
		}
	}

	require.Len(t, stdoutLines, 3, "every stdout line should produce one log entry")
	require.Contains(t, stdoutLines[0], "stdout: compiling…")
	require.Contains(t, stdoutLines[1], "stdout: linking…")
	require.Contains(t, stdoutLines[2], "stdout: ready")

	require.Len(t, stderrLines, 1)
	require.Contains(t, stderrLines[0], "stderr: deprecation: foo() will be removed")

	// StepResult.Stdout / Stderr still hold the full captured output so the
	// post-step JSON-encoded result keeps the same shape it had in v0.6.x.
	require.Equal(t, "compiling…\nlinking…\nready\n", res.Steps[0].Stdout)
	require.Equal(t, "deprecation: foo() will be removed\n", res.Steps[0].Stderr)
}

// TestRuntime_Run_StreamingHasCorrectLevels asserts that stdout lines are
// recorded as info while stderr lines are recorded as warn — matching the
// runtime's pre-existing convention for the post-step ok / warn / error
// summary lines so log consumers can filter consistently.
func TestRuntime_Run_StreamingHasCorrectLevels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	spec, err := ParseSpec([]byte(minimalSpecYAML))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, minimalSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{
		Stdout:   "hello\n",
		Stderr:   "uh oh\n",
		ExitCode: 0,
	}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)

	logs, err := store.Logs(ctx, res.RunID, 0)
	require.NoError(t, err)

	var stdoutLevel, stderrLevel string
	for _, l := range logs {
		if strings.Contains(l.Message, "stdout: hello") {
			stdoutLevel = l.Level
		}
		if strings.Contains(l.Message, "stderr: uh oh") {
			stderrLevel = l.Level
		}
	}
	require.Equal(t, "info", stdoutLevel)
	require.Equal(t, "warn", stderrLevel)
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

// TestRuntime_Run_WebhookOutputPostsResult covers the Tier 1.8 happy path:
// a httptest.Server receives the agent's RunResult as JSON, the runtime
// sets Content-Type=application/json by default, and the configured
// custom Authorization header round-trips intact.
func TestRuntime_Run_WebhookOutputPostsResult(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	var (
		mu          sync.Mutex
		gotMethod   string
		gotPath     string
		gotCT       string
		gotAuth     string
		gotPayload  RunResult
		callCount   int
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotPayload)
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)

	yamlSrc := `
id: webhook-agent
name: "Webhook output"
chain:
  - command: status
output:
  type: webhook
  url: ` + srv.URL + `/buddy-result
  headers:
    Authorization: "Bearer test-token"
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: "step done", ExitCode: 0}
	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, callCount, "webhook should be hit exactly once per run")
	require.Equal(t, "POST", gotMethod, "default method is POST")
	require.Equal(t, "/buddy-result", gotPath)
	require.Equal(t, "application/json", gotCT, "Content-Type defaults to application/json")
	require.Equal(t, "Bearer test-token", gotAuth, "custom Authorization header round-trips")
	require.Equal(t, res.RunID, gotPayload.RunID, "received body parses as the runtime's RunResult")
	require.Equal(t, "webhook-agent", gotPayload.AgentID)
	require.Equal(t, 0, gotPayload.ExitCode)
}

// TestRuntime_Run_WebhookOutputCustomMethodAndContentType verifies the spec
// can override the HTTP verb (PUT) and the Content-Type header — useful
// for targets that expect idempotent uploads or a custom media type.
func TestRuntime_Run_WebhookOutputCustomMethodAndContentType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	var (
		mu        sync.Mutex
		gotMethod string
		gotCT     string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	yamlSrc := `
id: webhook-put-agent
name: "Webhook PUT"
chain:
  - command: status
output:
  type: webhook
  url: ` + srv.URL + `
  method: PUT
  headers:
    Content-Type: application/vnd.buddy+json
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{ExitCode: 0}
	rt := NewRuntime(store, mock)
	_, err = rt.Run(ctx, agent)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, "PUT", gotMethod)
	require.Equal(t, "application/vnd.buddy+json", gotCT, "explicit Content-Type wins over the default")
}

// TestRuntime_Run_WebhookOutputNon2xxLogged ensures the runtime surfaces a
// non-2xx response as a warn-level log line on the run (the step itself
// still succeeded, but the output dispatch did not). Body preview should
// appear in the log so users can diagnose without re-running.
func TestRuntime_Run_WebhookOutputNon2xxLogged(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"webhook target down"}`))
	}))
	t.Cleanup(srv.Close)

	yamlSrc := `
id: webhook-fail-agent
name: "Webhook 500"
chain:
  - command: status
output:
  type: webhook
  url: ` + srv.URL + `
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	agent, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{ExitCode: 0}
	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, agent)
	require.NoError(t, err, "webhook failure does not fail the step itself")
	require.Equal(t, 0, res.ExitCode)

	logs, err := store.Logs(ctx, res.RunID, 0)
	require.NoError(t, err)
	var found bool
	for _, l := range logs {
		if strings.Contains(l.Message, "write output") && strings.Contains(l.Message, "500") {
			found = true
			break
		}
	}
	require.True(t, found, "expected a warn log line mentioning the 500 status, got %+v", logs)
}

// TestParseSpec_RejectsWebhookWithoutURL covers the spec-side guardrail:
// a webhook target without a URL is a configuration error, caught at
// ParseSpec time rather than at the first agent run.
func TestParseSpec_RejectsWebhookWithoutURL(t *testing.T) {
	t.Parallel()
	yamlSrc := `
id: bad-webhook
name: "missing URL"
chain:
  - command: status
output:
  type: webhook
`
	_, err := ParseSpec([]byte(yamlSrc))
	require.Error(t, err)
	require.Contains(t, err.Error(), "output.url")
}

// TestParseSpec_RejectsWebhookBadScheme catches typos in the URL — only
// http:// and https:// are honoured, so a missing scheme or one of
// ftp:// / file:// / javascript: never reaches the HTTP client.
func TestParseSpec_RejectsWebhookBadScheme(t *testing.T) {
	t.Parallel()
	yamlSrc := `
id: ftp-webhook
name: "wrong scheme"
chain:
  - command: status
output:
  type: webhook
  url: ftp://example.com/upload
`
	_, err := ParseSpec([]byte(yamlSrc))
	require.Error(t, err)
	require.Contains(t, err.Error(), "http:// or https://")
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

// ─── continue_on_fail (P2-1) ──────────────────────────────────────────

// TestRuntime_Run_ContinueOnFail_AdvancesPastFailedStep — when the first
// step has continue_on_fail=true and fails (non-zero exit), the chain
// must continue to step 2 instead of short-circuiting. Both steps end
// up in RunResult.Steps; mock.Calls records both executor invocations.
func TestRuntime_Run_ContinueOnFail_AdvancesPastFailedStep(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: cleanup-then-release
name: "Cleanup then release"
chain:
  - command: status
    args: ""
    continue_on_fail: true
  - command: concretize-idea
    args: "after cleanup"
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{ExitCode: 7, Stderr: "cleanup hiccup"}
	mock.Responses["concretize-idea"] = MockResponse{ExitCode: 0, Stdout: "ok"}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, 2, "step 2 must run despite step 1 failure")
	require.Equal(t, 7, res.Steps[0].ExitCode, "step 1's failure exit preserved in Steps")
	require.Equal(t, 0, res.Steps[1].ExitCode)
	require.Equal(t, []MockCall{
		{"status", ""},
		{"concretize-idea", "after cleanup"},
	}, mock.Calls)
}

// TestRuntime_Run_ContinueOnFail_FinalExitCodeIsLastNonZero — the
// run-level exit code surfaces failure even when continue_on_fail
// stopped the chain from short-circuiting. The contract is "last
// non-zero step exit wins" so a single failing cleanup step at the end
// is not silently swallowed.
func TestRuntime_Run_ContinueOnFail_FinalExitCodeIsLastNonZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: failing-cleanup
name: "Failing cleanup"
chain:
  - command: status
    args: ""
    continue_on_fail: true
  - command: concretize-idea
    args: "ok step"
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{ExitCode: 9, Stderr: "boom"}
	mock.Responses["concretize-idea"] = MockResponse{ExitCode: 0, Stdout: "ok"}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Equal(t, 9, res.ExitCode,
		"run-level exit must surface the failed step's exit code even though chain continued")

	after, err := store.Get(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, StatusFailed, after.Status,
		"agent status must reflect overall failure when any step failed")
}

// TestRuntime_Run_ContinueOnFail_DefaultFalseStillStops — a step
// without continue_on_fail behaves exactly like v0.6.x: failure
// short-circuits the chain. This guards backward compatibility — every
// existing spec must keep its v0.6.x semantics verbatim.
func TestRuntime_Run_ContinueOnFail_DefaultFalseStillStops(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	// twoStepSpecYAML has neither step marked continue_on_fail.
	spec, err := ParseSpec([]byte(twoStepSpecYAML))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, twoStepSpecYAML)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{ExitCode: 3, Stderr: "fail"}
	mock.Responses["concretize-idea"] = MockResponse{ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1, "default behavior must short-circuit after first failure")
	require.Len(t, mock.Calls, 1)
	require.Equal(t, 3, res.ExitCode)
}

// ─── auto-cascade (P2-2) ──────────────────────────────────────────────

// procWithNextPhase returns a stdout body that looks like a buddy
// PROCEDURE output with a §next-phase block listing the given skill
// as a backtick-wrapped identifier. This is the minimum substring the
// parser needs to populate ParsedOutput.NextPhase.Skills.
func procWithNextPhase(skill string) string {
	return "## 6. 검증 (self-check)\n" +
		"- [x] all good\n\n" +
		"## 7. 다음 phase\n\n" +
		"- `" + skill + "`\n"
}

// TestRuntime_Run_AutoCascade_AppendsParsedNextPhase — when auto_cascade
// is enabled and a step's parsed §next-phase has a skill, the runtime
// appends it as a new chain step. Two-step original chain + one
// cascade = three executor calls total.
func TestRuntime_Run_AutoCascade_AppendsParsedNextPhase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: cascade-agent
name: "Cascade demo"
auto_cascade: {}
chain:
  - command: status
    args: ""
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithNextPhase("concretize-idea"), ExitCode: 0}
	mock.Responses["concretize-idea"] = MockResponse{Stdout: "", ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, 2, "auto-cascade must append a step from §next-phase")
	require.Equal(t, "status", res.Steps[0].Command)
	require.Equal(t, 0, res.Steps[0].CascadeDepth, "original step is depth 0")
	require.Equal(t, "concretize-idea", res.Steps[1].Command)
	require.Equal(t, 1, res.Steps[1].CascadeDepth, "cascaded step is depth 1")
	require.Len(t, mock.Calls, 2)
}

// TestRuntime_Run_AutoCascade_RespectsMaxDepth — every cascade can
// produce one more cascade, but the runtime stops appending once depth
// reaches MaxDepth. Final step at MaxDepth still executes.
func TestRuntime_Run_AutoCascade_RespectsMaxDepth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: deep-cascade
name: "Deep cascade"
auto_cascade:
  max_depth: 2
chain:
  - command: status
    args: ""
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	// Every command emits a §next-phase pointing to the next.
	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithNextPhase("concretize-idea"), ExitCode: 0}
	mock.Responses["concretize-idea"] = MockResponse{Stdout: procWithNextPhase("define-features"), ExitCode: 0}
	mock.Responses["define-features"] = MockResponse{Stdout: procWithNextPhase("design-system"), ExitCode: 0}
	mock.Responses["design-system"] = MockResponse{Stdout: procWithNextPhase("plan-build"), ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	// Original at depth 0 → cascade to depth 1 → cascade to depth 2 → STOP.
	// (depth 2 step runs, but its §next-phase is NOT followed.)
	require.Len(t, res.Steps, 3, "max_depth=2 must cap the chain at depth 0/1/2")
	require.Equal(t, 0, res.Steps[0].CascadeDepth)
	require.Equal(t, 1, res.Steps[1].CascadeDepth)
	require.Equal(t, 2, res.Steps[2].CascadeDepth)
	require.Equal(t, []MockCall{
		{"status", ""},
		{"concretize-idea", ""},
		{"define-features", ""},
	}, mock.Calls)
}

// TestRuntime_Run_AutoCascade_NoEffectWhenDisabled — without
// auto_cascade in the spec, parsed §next-phase is logged but does NOT
// drive new chain steps (preserves v0.6.x backward compat).
func TestRuntime_Run_AutoCascade_NoEffectWhenDisabled(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	// Same as the AppendsParsedNextPhase fixture but WITHOUT auto_cascade.
	yamlSrc := `
id: no-cascade
name: "No cascade"
chain:
  - command: status
    args: ""
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithNextPhase("concretize-idea"), ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1, "auto-cascade disabled must not add steps")
	require.Len(t, mock.Calls, 1)
}

// TestRuntime_Run_AutoCascade_SkipsFromFailedStep — a failed step does
// NOT cascade. Even with `continue_on_fail: true` to keep the chain
// alive, the runtime does not auto-append §next-phase from a fault path.
func TestRuntime_Run_AutoCascade_SkipsFromFailedStep(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: fail-no-cascade
name: "Fail no cascade"
auto_cascade: {}
chain:
  - command: status
    args: ""
    continue_on_fail: true
  - command: concretize-idea
    args: "follow-up"
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	// status FAILS but still emits a §next-phase. Must not be followed.
	mock.Responses["status"] = MockResponse{Stdout: procWithNextPhase("define-features"), ExitCode: 5}
	mock.Responses["concretize-idea"] = MockResponse{Stdout: "", ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, 2, "chain proceeds via continue_on_fail but NO cascade from failed step")
	require.Equal(t, "status", res.Steps[0].Command)
	require.Equal(t, "concretize-idea", res.Steps[1].Command)
	// define-features was never invoked.
	for _, c := range mock.Calls {
		require.NotEqual(t, "define-features", c.Command,
			"failed step must not cascade into §next-phase target")
	}
}

// TestRuntime_Run_AutoCascade_NoSkillsNoEffect — auto_cascade enabled
// but the step's stdout has no §next-phase (or it's empty) → no
// cascade. Common case for terminal phases like ship-release.
func TestRuntime_Run_AutoCascade_NoSkillsNoEffect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: terminal-step
name: "Terminal step"
auto_cascade: {}
chain:
  - command: status
    args: ""
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: "no §next-phase block here", ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, 1, "no §next-phase → no cascade")
}

// TestRuntime_Run_AutoCascade_DefaultMaxDepth — auto_cascade with empty
// `{}` (no max_depth) uses DefaultCascadeMaxDepth. Build a chain that
// would cascade indefinitely if unbounded; verify it stops exactly at
// the default cap.
func TestRuntime_Run_AutoCascade_DefaultMaxDepth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: default-depth
name: "Default depth"
auto_cascade: {}
chain:
  - command: status
    args: ""
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	// Self-referential cascade — every step says "cascade to self". We
	// can't have skill names collide for the parser ("status" multiple
	// times is fine — each invocation re-runs the same MockResponse).
	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithNextPhase("status"), ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, DefaultCascadeMaxDepth+1,
		"default cap allows depth 0..DefaultCascadeMaxDepth inclusive")
	require.Equal(t, DefaultCascadeMaxDepth, res.Steps[len(res.Steps)-1].CascadeDepth)
}

// TestRuntime_Run_ContinueOnFail_AllPassExitZero — when continue_on_fail
// is set but the step actually succeeds (and downstream steps also
// succeed), the run-level exit must be 0 — no false-positive failure.
func TestRuntime_Run_ContinueOnFail_AllPassExitZero(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: best-effort-all-ok
name: "Best effort all ok"
chain:
  - command: status
    args: ""
    continue_on_fail: true
  - command: concretize-idea
    args: "ok"
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{ExitCode: 0}
	mock.Responses["concretize-idea"] = MockResponse{ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Equal(t, 0, res.ExitCode)
	after, err := store.Get(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, StatusDone, after.Status)
}

// ─── branch hints ─────────────────────────────────────────────────────

// procWithBranches mirrors procWithNextPhase but emits two conditional
// branches, exercising the parser's NextPhase.Branches surface that
// the BranchHints disambiguates.
func procWithBranches() string {
	return "## 6. 검증\n\n- [x] decided\n\n" +
		"## 7. 다음 phase\n\n" +
		"- Korea → `consult-korea-legal-context`\n" +
		"- 글로벌 → `review-legal-regulatory`\n"
}

// TestPickCascadeTarget_NoBranchesFallsBackToFirstSkill — the
// no-branches path stays identical to the original behaviour: Skills[0].
func TestPickCascadeTarget_NoBranchesFallsBackToFirstSkill(t *testing.T) {
	t.Parallel()
	np := NextPhase{Skills: []string{"a", "b"}}
	require.Equal(t, "a", pickCascadeTarget(np, nil))
	require.Equal(t, "a", pickCascadeTarget(np, map[string]bool{"unrelated": true}))
}

// TestPickCascadeTarget_HintSelectsMatchingBranch — when Branches are
// present, only a true hint for a Condition picks that branch's first
// skill. Branch order in the source is preserved.
func TestPickCascadeTarget_HintSelectsMatchingBranch(t *testing.T) {
	t.Parallel()
	np := NextPhase{
		Branches: []NextPhaseBranch{
			{Condition: "Korea", Skills: []string{"consult-korea-legal-context"}},
			{Condition: "글로벌", Skills: []string{"review-legal-regulatory"}},
		},
	}
	require.Equal(t, "consult-korea-legal-context",
		pickCascadeTarget(np, map[string]bool{"Korea": true}))
	require.Equal(t, "review-legal-regulatory",
		pickCascadeTarget(np, map[string]bool{"글로벌": true}))
}

// TestPickCascadeTarget_NoMatchingHintReturnsEmpty — missing hint, false
// hint, and empty-skill branch all suppress the cascade so the runtime
// stops rather than silently picking the union's first entry.
func TestPickCascadeTarget_NoMatchingHintReturnsEmpty(t *testing.T) {
	t.Parallel()
	np := NextPhase{
		Branches: []NextPhaseBranch{
			{Condition: "Korea", Skills: []string{"consult-korea-legal-context"}},
			{Condition: "글로벌", Skills: []string{"review-legal-regulatory"}},
		},
	}
	require.Equal(t, "", pickCascadeTarget(np, nil))
	require.Equal(t, "", pickCascadeTarget(np, map[string]bool{"USA": true}))
	require.Equal(t, "", pickCascadeTarget(np, map[string]bool{"Korea": false, "글로벌": false}))
}

// TestRuntime_Run_BranchHintsDriveCascade — end-to-end: a spec with
// branch_hints set picks the matching branch and the cascade only
// executes that path, never the other branch.
func TestRuntime_Run_BranchHintsDriveCascade(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)

	yamlSrc := `
id: hinted-cascade
name: "Hinted cascade"
auto_cascade: {}
branch_hints:
  Korea: true
chain:
  - command: status
    args: ""
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithBranches(), ExitCode: 0}
	mock.Responses["consult-korea-legal-context"] = MockResponse{Stdout: "", ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Len(t, res.Steps, 2, "Korea hint must trigger the Korea-branch cascade")
	require.Equal(t, "consult-korea-legal-context", res.Steps[1].Command)
}

// ─── on_self_check_fail policy ────────────────────────────────────────

// procWithSelfCheckFail emits a PROCEDURE body whose §self-check has
// an unchecked checkbox so the parser records SelfCheckFail. No
// next-phase block — cascade is irrelevant to this policy test.
func procWithSelfCheckFail() string {
	return "## 6. 검증 (self-check)\n" +
		"- [x] some checks pass\n" +
		"- [ ] one critical check fails\n"
}

// TestSpec_OnSelfCheckFail_RejectsUnknownValue — a typo in the YAML
// must surface at parse time, not silently default at runtime.
func TestSpec_OnSelfCheckFail_RejectsUnknownValue(t *testing.T) {
	t.Parallel()
	yamlSrc := `
id: typo-policy
name: "typo policy"
chain:
  - command: status
    on_self_check_fail: maybeRetry
`
	_, err := ParseSpec([]byte(yamlSrc))
	require.Error(t, err)
	require.Contains(t, err.Error(), "on_self_check_fail")
}

// TestRuntime_Run_SelfCheckFail_ContinueIsDefault — empty policy
// preserves v0.5.0+ behaviour: the verdict is logged but the step
// returns success and the run completes.
func TestRuntime_Run_SelfCheckFail_ContinueIsDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)
	yamlSrc := `
id: default-policy
name: "default policy"
chain:
  - command: status
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithSelfCheckFail(), ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.Equal(t, 0, res.ExitCode, "self-check fail with default policy must not flip the run")
	require.Len(t, mock.Calls, 1, "no retry, no extra calls")
}

// TestRuntime_Run_SelfCheckFail_AbortStopsChain — abort policy turns
// a verdict-fail step into an immediate cascade halt.
func TestRuntime_Run_SelfCheckFail_AbortStopsChain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)
	yamlSrc := `
id: abort-policy
name: "abort policy"
chain:
  - command: status
    on_self_check_fail: abort
  - command: status
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithSelfCheckFail(), ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err, "Run returns nil; failure surfaces via ExitCode")
	require.NotEqual(t, 0, res.ExitCode, "abort must surface as a non-zero run exit code")
	require.Len(t, res.Steps, 1, "the second step must not run after abort")
}

// TestRuntime_Run_SelfCheckFail_RetryHonoursMaxAttempts — retry policy
// triggers the step's RetryPolicy on a verdict fail, just like an
// exit-code failure. After max_attempts the run still surfaces failure.
func TestRuntime_Run_SelfCheckFail_RetryHonoursMaxAttempts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, _ := newTestStore(t)
	yamlSrc := `
id: retry-policy
name: "retry policy"
retry:
  max_attempts: 3
chain:
  - command: status
    on_self_check_fail: retry
`
	spec, err := ParseSpec([]byte(yamlSrc))
	require.NoError(t, err)
	a, err := store.Create(ctx, spec, yamlSrc)
	require.NoError(t, err)

	mock := NewMockExecutor()
	mock.Responses["status"] = MockResponse{Stdout: procWithSelfCheckFail(), ExitCode: 0}

	rt := NewRuntime(store, mock)
	res, err := rt.Run(ctx, a)
	require.NoError(t, err)
	require.NotEqual(t, 0, res.ExitCode, "exhausted retries on self-check fail must surface failure")
	require.Equal(t, 3, len(mock.Calls), "max_attempts=3 with verdict-fail every time = 3 calls")
}

// Boundary tests for computeBackoff. The exponential branch has two
// guards a previous coverage audit flagged as one-sided: the shift
// clamp (attempt-1 ≤ 30) and the BackoffMax cap (wait > BackoffMax).
// The pairs below pin both sides so a future mutation that flips a
// strict inequality to a non-strict one breaks at least one assertion.

func TestComputeBackoff_ExponentialCapAtExactMax(t *testing.T) {
	t.Parallel()
	// BackoffDelay=1s, exponent attempt-1 = 1 → wait = 2s.
	// BackoffMax = 2s exactly. Current code uses `wait > BackoffMax` so
	// equality keeps wait — the cap only kicks in strictly above max.
	policy := &RetryPolicy{
		BackoffStrategy: BackoffStrategyExponential,
		BackoffDelay:    time.Second,
		BackoffMax:      2 * time.Second,
	}
	require.Equal(t, 2*time.Second, computeBackoff(policy, 2),
		"wait == BackoffMax is allowed; strict-greater drives the cap")

	// attempt=3 → wait would be 4s; cap to 2s.
	require.Equal(t, 2*time.Second, computeBackoff(policy, 3),
		"wait > BackoffMax is clamped to BackoffMax")
}

func TestComputeBackoff_ExponentClampAtThirty(t *testing.T) {
	t.Parallel()
	// The clamp on (attempt-1, 30) prevents 2^k from running away when
	// MaxAttempts is misconfigured. At attempt 31 the shift exponent
	// must equal the attempt-32 result — same 2^30 multiplier.
	policy := &RetryPolicy{
		BackoffStrategy: BackoffStrategyExponential,
		BackoffDelay:    time.Nanosecond,
		// No BackoffMax so the cap doesn't mask the clamp behaviour.
	}
	at31 := computeBackoff(policy, 31)
	at32 := computeBackoff(policy, 32)
	require.Equal(t, at31, at32,
		"attempt >= 31 must reuse the 2^30 multiplier (shift clamp)")
	// And the value itself: 2^30 nanoseconds.
	require.Equal(t, time.Duration(1<<30)*time.Nanosecond, at31)
}

func TestComputeBackoff_GuardsBelowAttemptOne(t *testing.T) {
	t.Parallel()
	policy := &RetryPolicy{
		BackoffStrategy: BackoffStrategyExponential,
		BackoffDelay:    time.Second,
	}
	require.Equal(t, time.Duration(0), computeBackoff(policy, 0),
		"attempt < 1 short-circuits to zero")
	require.Equal(t, time.Duration(0), computeBackoff(policy, -1),
		"negative attempt also short-circuits")
	require.Equal(t, time.Duration(0), computeBackoff(nil, 1),
		"nil policy short-circuits to zero")
}
