// fs_lister.go — the filesystem-scanning Lister. Scans
// ~/.claude/projects/*/*.jsonl, tails each file from its last_offset,
// and updates token / activity / goal_text in the sessions table.

package sessions

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FSLister walks Claude Code's per-project transcript directories and
// upserts session rows. Stateless across calls — every call re-scans the
// filesystem; resumption between calls is via sessions.last_offset.
type FSLister struct {
	store *Store
	// ProjectsRoot is the directory containing per-project subdirectories
	// holding .jsonl transcripts. Defaults to ~/.claude/projects.
	ProjectsRoot string
	// MaxGoalLen caps GoalText extraction so a pasted-novel first message
	// doesn't blow up the row. Default 500 chars.
	MaxGoalLen int
}

// NewFSLister constructs a Lister rooted at the user's Claude Code
// project directory. Empty ProjectsRoot resolves to $HOME/.claude/projects.
func NewFSLister(store *Store) *FSLister {
	return &FSLister{
		store:        store,
		ProjectsRoot: defaultProjectsRoot(),
		MaxGoalLen:   500,
	}
}

// List walks ProjectsRoot, tails every JSONL it finds, upserts each
// session, and returns the in-memory snapshot. Matches the Lister
// interface contract from sessions.go.
//
// Errors on individual files are logged-via-error-wrap but do not abort
// the walk — a corrupt transcript shouldn't block monitoring of the
// rest.
func (l *FSLister) List(ctx context.Context) ([]Session, error) {
	if l.ProjectsRoot == "" {
		l.ProjectsRoot = defaultProjectsRoot()
	}
	if l.MaxGoalLen == 0 {
		l.MaxGoalLen = 500
	}

	entries, err := os.ReadDir(l.ProjectsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("fs lister: read projects root: %w", err)
	}

	var out []Session
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		projDir := filepath.Join(l.ProjectsRoot, e.Name())
		files, err := os.ReadDir(projDir)
		if err != nil {
			continue // project dir unreadable; skip silently
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			sess, err := l.processTranscript(ctx, filepath.Join(projDir, f.Name()))
			if err != nil {
				continue // bad file; skip
			}
			if sess.ID == "" {
				continue // no usable event found
			}
			out = append(out, sess)
		}
	}
	return out, nil
}

// processTranscript reads (or resumes reading) one transcript JSONL.
// On success, the corresponding row in sessions is Upserted and the
// in-memory Session is returned. If the row already exists, prior
// counters are added to (not replaced).
func (l *FSLister) processTranscript(ctx context.Context, path string) (Session, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Session{}, err
	}

	// Derive a probable session ID from filename (CC names files
	// `<sessionId>.jsonl`); the first event with a sessionId field
	// confirms / corrects this.
	probableID := strings.TrimSuffix(filepath.Base(path), ".jsonl")

	// Look up existing row (if any) to resume from last_offset and
	// preserve started_at / goal_text.
	prior, err := l.store.Get(ctx, probableID)
	priorExisted := err == nil
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Session{}, fmt.Errorf("fs lister: load prior: %w", err)
	}

	// If the file shrunk (transcript rotated / cleared), reset offset.
	startOffset := prior.LastOffset
	if startOffset > info.Size() {
		startOffset = 0
	}
	// Nothing new — skip parse but still touch last_active = mtime.
	if priorExisted && startOffset == info.Size() {
		return prior, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer f.Close()
	if startOffset > 0 {
		if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
			return Session{}, err
		}
	}

	sess := prior
	sess.TranscriptPath = path
	if !priorExisted {
		sess.StartedAt = info.ModTime().UTC()
		sess.ID = probableID
	}

	scanner := bufio.NewScanner(f)
	// Default Scanner buffer (64KB) may be too small for big system-injected
	// messages. Bump to 4MB to cover normal CC payloads.
	scanner.Buffer(make([]byte, 0, 1<<20), 4<<20)

	var (
		consumed int64
		lastTs   time.Time
	)
	for scanner.Scan() {
		line := scanner.Bytes()
		consumed += int64(len(line)) + 1 // +1 for the newline
		var ev transcriptEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue // malformed line; skip
		}
		if ev.SessionID != "" {
			sess.ID = ev.SessionID
		}
		if ts, ok := parseTimestamp(ev.Timestamp); ok {
			lastTs = ts
		}
		switch ev.Type {
		case "assistant":
			if ev.Message != nil && ev.Message.Usage != nil {
				u := ev.Message.Usage
				sess.Usage.InputTokens += u.InputTokens
				sess.Usage.OutputTokens += u.OutputTokens
				sess.Usage.CacheReadTokens += u.CacheReadInputTokens
				sess.Usage.CacheCreateTokens += u.CacheCreationInputTokens
			}
		case "user":
			// Goal extraction: first non-meta user message becomes GoalText.
			if sess.GoalText == "" && !ev.IsMeta && ev.Message != nil {
				if text := extractTextContent(ev.Message.Content); text != "" {
					if len(text) > l.MaxGoalLen {
						text = text[:l.MaxGoalLen]
					}
					sess.GoalText = text
				}
			}
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		// Buffer-exceeded or other read error. Save what we have; the
		// next tick will retry from the new offset.
		_ = err
	}

	sess.LastOffset = startOffset + consumed
	if !lastTs.IsZero() {
		sess.LastActive = lastTs
	} else if sess.LastActive.IsZero() {
		sess.LastActive = info.ModTime().UTC()
	}
	if sess.StartedAt.IsZero() {
		sess.StartedAt = sess.LastActive
	}
	if sess.Metadata == "" {
		sess.Metadata = "{}"
	}

	if err := l.store.Upsert(ctx, sess); err != nil {
		return Session{}, err
	}
	return sess, nil
}

// ─── JSONL shape (subset we read) ──────────────────────────────────────

// transcriptEvent is the subset of fields fsLister actually reads from
// one JSONL line. We do not capture every CC field — Forward-compat with
// CC adding fields is via json.Decoder ignoring unknown keys.
type transcriptEvent struct {
	Type      string             `json:"type"`
	SessionID string             `json:"sessionId,omitempty"`
	Timestamp string             `json:"timestamp,omitempty"`
	IsMeta    bool               `json:"isMeta,omitempty"`
	Message   *transcriptMessage `json:"message,omitempty"`
}

type transcriptMessage struct {
	Role    string          `json:"role,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
	Usage   *transcriptUsage `json:"usage,omitempty"`
}

type transcriptUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// extractTextContent flattens message.content to a plain string. CC uses
// two shapes:
//   - string ("hello")
//   - array of typed parts ([{type:"text", text:"hello"}, ...])
// Returns "" when no plain text can be extracted (e.g., tool_result only).
func extractTextContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// Try string first.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	// Try array of typed parts.
	var parts []map[string]any
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if t, ok := p["type"].(string); !ok || t != "text" {
			continue
		}
		if text, ok := p["text"].(string); ok {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(text)
		}
	}
	return strings.TrimSpace(b.String())
}

// parseTimestamp is forgiving — CC uses ISO 8601 (RFC 3339) but we accept
// any common variant. Returns (zero, false) when un-parseable.
func parseTimestamp(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// defaultProjectsRoot returns $HOME/.claude/projects.
func defaultProjectsRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}
