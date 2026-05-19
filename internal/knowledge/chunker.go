package knowledge

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultMaxTokens caps single-chunk size — keeps embedding model input
// within typical 512-token windows. Naive whitespace tokenisation
// (overcounts by ~25% vs WordPiece), so picking 500 leaves headroom.
const DefaultMaxTokens = 500

// Chunker turns a session transcript (~/.claude/projects/*/*.jsonl)
// into Chunk records ready for Store.Insert. Stateless — every Extract
// call re-reads from scratch.
type Chunker struct {
	// MaxTokens caps each chunk's token count. Zero => DefaultMaxTokens.
	MaxTokens int
}

// NewChunker returns a Chunker with defaults applied.
func NewChunker() *Chunker {
	return &Chunker{MaxTokens: DefaultMaxTokens}
}

// Extract parses path and returns chunks. SessionID is inferred from
// the filename (CC names files `<sessionId>.jsonl`) when no transcript
// event populates it; otherwise the first observed sessionId wins.
//
// Non-text events (tool_result, tool_use without text part, meta) are
// skipped — we keep the chunk corpus focused on prose the user actually
// typed / saw.
func (c *Chunker) Extract(path string) ([]Chunk, error) {
	maxTok := c.MaxTokens
	if maxTok <= 0 {
		maxTok = DefaultMaxTokens
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("knowledge: chunker open %s: %w", path, err)
	}
	defer f.Close()

	probableID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	sessionID := probableID

	sc := bufio.NewScanner(f)
	// Some transcripts have very long lines (pasted content); raise the
	// buffer to 4 MB so the scanner doesn't truncate.
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var out []Chunk
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev chunkerEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue // skip malformed line; don't abort whole transcript
		}
		if ev.IsMeta {
			continue
		}
		if ev.SessionID != "" {
			sessionID = ev.SessionID
		}
		if ev.Message == nil {
			continue
		}
		// We chunk both roles — user prompts and assistant responses are
		// both useful as knowledge corpus (user prompts show intent,
		// assistant responses show what was produced).
		text := extractTextContent(ev.Message.Content)
		if text == "" {
			continue
		}
		role := ev.Message.Role
		if role == "" {
			role = "unknown"
		}
		// Prefix with role marker so BM25 retrieval can match queries
		// like "what did the user ask about X". Markdown-light to keep
		// the chunk readable when dumped as-is.
		body := "[" + role + "]\n" + text
		for _, piece := range splitByTokens(body, maxTok) {
			out = append(out, Chunk{
				SessionID:  sessionID,
				Content:    piece,
				TokenCount: countTokens(piece),
			})
		}
	}
	if err := sc.Err(); err != nil {
		return out, fmt.Errorf("knowledge: chunker scan: %w", err)
	}
	return out, nil
}

// splitByTokens cuts s into pieces of at most maxTok whitespace tokens.
// Preserves token boundaries (never splits mid-word). For short inputs
// (one piece fits) returns []string{s} so the caller doesn't deal with
// nil checks for the common case.
func splitByTokens(s string, maxTok int) []string {
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return nil
	}
	if len(tokens) <= maxTok {
		return []string{s}
	}
	var out []string
	for i := 0; i < len(tokens); i += maxTok {
		end := i + maxTok
		if end > len(tokens) {
			end = len(tokens)
		}
		out = append(out, strings.Join(tokens[i:end], " "))
	}
	return out
}

// countTokens is the naive whitespace tokeniser the chunker uses for
// TokenCount. Good enough for ranking + sanity stats; embedding model
// real token count will differ but stays within ~25%.
func countTokens(s string) int {
	return len(strings.Fields(s))
}

// ─── JSONL subset ────────────────────────────────────────────────────

// chunkerEvent mirrors the relevant subset of transcript JSONL events.
// Intentionally separate from sessions.transcriptEvent so refactors on
// one package don't ripple to the other; the on-wire shape is stable
// per Claude Code so duplication cost is low.
type chunkerEvent struct {
	Type      string             `json:"type"`
	SessionID string             `json:"sessionId,omitempty"`
	IsMeta    bool               `json:"isMeta,omitempty"`
	Message   *chunkerMessage    `json:"message,omitempty"`
}

type chunkerMessage struct {
	Role    string          `json:"role,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

// extractTextContent flattens message.content into prose. Handles both
// CC shapes: bare string and array of typed parts. Non-text parts
// (tool_use / tool_result) are dropped — they aren't useful corpus.
func extractTextContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var parts []map[string]any
	if err := json.Unmarshal(raw, &parts); err != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		t, _ := p["type"].(string)
		if t != "text" {
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
