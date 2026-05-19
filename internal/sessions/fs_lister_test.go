package sessions

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeTranscript writes a fixture JSONL at projects/<proj>/<sessID>.jsonl
// and returns its absolute path.
func writeTranscript(t *testing.T, projectsRoot, proj, sessID string, lines []string) string {
	t.Helper()
	projDir := filepath.Join(projectsRoot, proj)
	require.NoError(t, os.MkdirAll(projDir, 0o755))
	path := filepath.Join(projDir, sessID+".jsonl")
	body := ""
	for _, ln := range lines {
		body += ln + "\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

// TestFSLister_HappyPath — one assistant message + one user message →
// the session row gets token totals + GoalText extracted from the user
// message + LastActive from the assistant timestamp.
func TestFSLister_HappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	root := t.TempDir()

	const userMsg = `{"type":"user","sessionId":"sess-1","timestamp":"2026-05-19T10:00:00Z","isMeta":false,"message":{"role":"user","content":"design a session monitor"}}`
	const asstMsg = `{"type":"assistant","sessionId":"sess-1","timestamp":"2026-05-19T10:00:05Z","message":{"role":"assistant","usage":{"input_tokens":100,"output_tokens":200,"cache_read_input_tokens":50,"cache_creation_input_tokens":10}}}`
	writeTranscript(t, root, "proj-a", "sess-1", []string{userMsg, asstMsg})

	l := NewFSLister(store)
	l.ProjectsRoot = root
	got, err := l.List(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	sess, err := store.Get(ctx, "sess-1")
	require.NoError(t, err)
	require.Equal(t, "design a session monitor", sess.GoalText)
	require.Equal(t, 100, sess.Usage.InputTokens)
	require.Equal(t, 200, sess.Usage.OutputTokens)
	require.Equal(t, 50, sess.Usage.CacheReadTokens)
	require.Equal(t, 10, sess.Usage.CacheCreateTokens)
	require.False(t, sess.LastActive.IsZero())
}

// TestFSLister_SkipsMetaUserMessages — `isMeta:true` user messages are
// system-injected (local-command-caveat etc.) and must NOT become the
// session's GoalText.
func TestFSLister_SkipsMetaUserMessages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	root := t.TempDir()

	const metaMsg = `{"type":"user","sessionId":"sess-2","timestamp":"2026-05-19T10:00:00Z","isMeta":true,"message":{"role":"user","content":"<local-command-caveat>noise</local-command-caveat>"}}`
	const realMsg = `{"type":"user","sessionId":"sess-2","timestamp":"2026-05-19T10:00:10Z","isMeta":false,"message":{"role":"user","content":"the real goal"}}`
	writeTranscript(t, root, "proj-b", "sess-2", []string{metaMsg, realMsg})

	l := NewFSLister(store)
	l.ProjectsRoot = root
	_, err := l.List(ctx)
	require.NoError(t, err)

	sess, err := store.Get(ctx, "sess-2")
	require.NoError(t, err)
	require.Equal(t, "the real goal", sess.GoalText,
		"isMeta:true messages must be skipped during goal extraction")
}

// TestFSLister_ResumesFromLastOffset — second call only reads bytes
// appended after the first call; token counters accumulate.
func TestFSLister_ResumesFromLastOffset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	root := t.TempDir()

	const userMsg = `{"type":"user","sessionId":"sess-3","timestamp":"2026-05-19T10:00:00Z","isMeta":false,"message":{"role":"user","content":"hello"}}`
	const asst1 = `{"type":"assistant","sessionId":"sess-3","timestamp":"2026-05-19T10:00:05Z","message":{"role":"assistant","usage":{"input_tokens":10,"output_tokens":20,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`
	path := writeTranscript(t, root, "proj-c", "sess-3", []string{userMsg, asst1})

	l := NewFSLister(store)
	l.ProjectsRoot = root
	_, err := l.List(ctx)
	require.NoError(t, err)

	sess1, err := store.Get(ctx, "sess-3")
	require.NoError(t, err)
	require.Equal(t, 10, sess1.Usage.InputTokens)
	require.Equal(t, 20, sess1.Usage.OutputTokens)
	firstOffset := sess1.LastOffset
	require.Greater(t, firstOffset, int64(0))

	// Append another assistant message.
	const asst2 = `{"type":"assistant","sessionId":"sess-3","timestamp":"2026-05-19T10:00:10Z","message":{"role":"assistant","usage":{"input_tokens":7,"output_tokens":3,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	_, err = f.WriteString(asst2 + "\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	_, err = l.List(ctx)
	require.NoError(t, err)

	sess2, err := store.Get(ctx, "sess-3")
	require.NoError(t, err)
	require.Equal(t, 17, sess2.Usage.InputTokens, "tokens must accumulate across ticks")
	require.Equal(t, 23, sess2.Usage.OutputTokens)
	require.Greater(t, sess2.LastOffset, firstOffset, "offset must advance")
}

// TestFSLister_HandlesEmptyProjectsRoot — no projects dir → no error,
// no rows.
func TestFSLister_HandlesEmptyProjectsRoot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)

	l := NewFSLister(store)
	l.ProjectsRoot = "/nonexistent/path/that/should/not/exist"
	got, err := l.List(ctx)
	require.NoError(t, err)
	require.Empty(t, got)
}

// TestFSLister_TruncatedTranscript — file shrunk between ticks (rotated
// / cleared); fsLister resets offset to 0 + re-reads.
func TestFSLister_TruncatedTranscript(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	root := t.TempDir()

	const original = `{"type":"user","sessionId":"sess-4","timestamp":"2026-05-19T10:00:00Z","isMeta":false,"message":{"role":"user","content":"original goal"}}
{"type":"assistant","sessionId":"sess-4","timestamp":"2026-05-19T10:00:05Z","message":{"role":"assistant","usage":{"input_tokens":100,"output_tokens":200,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`
	path := writeTranscript(t, root, "proj-d", "sess-4", []string{original})

	l := NewFSLister(store)
	l.ProjectsRoot = root
	_, err := l.List(ctx)
	require.NoError(t, err)

	sess, err := store.Get(ctx, "sess-4")
	require.NoError(t, err)
	require.Equal(t, "original goal", sess.GoalText)

	// Truncate to a brand-new short content.
	const shorter = `{"type":"user","sessionId":"sess-4","timestamp":"2026-05-19T11:00:00Z","isMeta":false,"message":{"role":"user","content":"rotated"}}`
	require.NoError(t, os.WriteFile(path, []byte(shorter+"\n"), 0o644))

	_, err = l.List(ctx)
	require.NoError(t, err)

	sess, err = store.Get(ctx, "sess-4")
	require.NoError(t, err)
	require.Equal(t, "original goal", sess.GoalText,
		"GoalText stays sticky even on rotation (set in Upsert: CASE WHEN goal_text='' ...)")
}

// TestFSLister_MaxGoalLenTruncates — a pasted-novel first message gets
// truncated at MaxGoalLen.
func TestFSLister_MaxGoalLenTruncates(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	root := t.TempDir()

	long := ""
	for i := 0; i < 1000; i++ {
		long += "X"
	}
	userMsg := `{"type":"user","sessionId":"sess-5","timestamp":"2026-05-19T10:00:00Z","isMeta":false,"message":{"role":"user","content":"` + long + `"}}`
	writeTranscript(t, root, "proj-e", "sess-5", []string{userMsg})

	l := NewFSLister(store)
	l.ProjectsRoot = root
	l.MaxGoalLen = 100
	_, err := l.List(ctx)
	require.NoError(t, err)

	sess, err := store.Get(ctx, "sess-5")
	require.NoError(t, err)
	require.Equal(t, 100, len(sess.GoalText), "GoalText must be capped at MaxGoalLen")
}

// TestFSLister_ContentAsArray — CC sometimes uses content as an array of
// typed parts. Extract text parts only.
func TestFSLister_ContentAsArray(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	root := t.TempDir()

	const arrayContent = `{"type":"user","sessionId":"sess-6","timestamp":"2026-05-19T10:00:00Z","isMeta":false,"message":{"role":"user","content":[{"type":"text","text":"hello from array"},{"type":"tool_result","content":"some tool output"}]}}`
	writeTranscript(t, root, "proj-f", "sess-6", []string{arrayContent})

	l := NewFSLister(store)
	l.ProjectsRoot = root
	_, err := l.List(ctx)
	require.NoError(t, err)

	sess, err := store.Get(ctx, "sess-6")
	require.NoError(t, err)
	require.Equal(t, "hello from array", sess.GoalText)
}

// TestFSLister_MultipleProjectsAndSessions — three projects, each with
// one session, all get rows.
func TestFSLister_MultipleProjectsAndSessions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	root := t.TempDir()

	for i, sid := range []string{"alpha", "beta", "gamma"} {
		_ = i
		writeTranscript(t, root, "proj-"+sid, sid, []string{
			`{"type":"user","sessionId":"` + sid + `","timestamp":"2026-05-19T10:00:00Z","isMeta":false,"message":{"role":"user","content":"goal for ` + sid + `"}}`,
		})
	}

	l := NewFSLister(store)
	l.ProjectsRoot = root
	got, err := l.List(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)
}
