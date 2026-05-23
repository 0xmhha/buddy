package knowledge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// EmbedRequest is one input row sent to the Python embedder.
type EmbedRequest struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

// EmbedResult is one output row received from the embedder.
type EmbedResult struct {
	ID        int64     `json:"id"`
	Embedding []float32 `json:"embedding"`
}

// ErrEmbedderUnavailable signals the Python binary or sentence-transformers
// dependency is missing. CLI / MCP callers translate this to a friend-tone
// "Python venv 가 안 보여..." message so the user knows where to go next.
var ErrEmbedderUnavailable = errors.New("knowledge: embedder unavailable (python or sentence-transformers missing)")

// Embedder is the abstract embed pipeline. PythonEmbedder is the only
// production implementation; tests can inject MockEmbedder without
// touching subprocess machinery.
type Embedder interface {
	Embed(ctx context.Context, reqs []EmbedRequest) ([]EmbedResult, error)
}

// PythonEmbedder spawns scripts/embed.py once per batch and pipes
// JSON Lines in/out. Reusing the process across the whole batch
// amortises the sentence-transformers model load (~2-5s) over the
// batch's cost.
//
// Per-call subprocess lifetime: started by Embed, killed when Embed
// returns. We do not keep a long-lived daemon — embed traffic is
// bursty (ingest run) and a daemon would just hold 80MB of model
// RAM between bursts.
type PythonEmbedder struct {
	// PythonBin is the python interpreter. Empty => "python3" in PATH.
	PythonBin string
	// ScriptPath is the embed.py path. Empty => env BUDDY_EMBED_SCRIPT,
	// or fall back to "scripts/embed.py" relative to cwd.
	ScriptPath string
}

// NewPythonEmbedder applies the standard defaults.
func NewPythonEmbedder() *PythonEmbedder {
	return &PythonEmbedder{}
}

func (e *PythonEmbedder) pythonBin() string {
	if e.PythonBin != "" {
		return e.PythonBin
	}
	return "python3"
}

func (e *PythonEmbedder) scriptPath() (string, error) {
	if e.ScriptPath != "" {
		return e.ScriptPath, nil
	}
	if env := os.Getenv("BUDDY_EMBED_SCRIPT"); env != "" {
		return env, nil
	}
	// Fall back to scripts/embed.py relative to cwd. The CLI caller is
	// expected to be run from the repo root for now; a future change will
	// move the script under buddy's install dir for general-use deployments.
	abs, err := filepath.Abs("scripts/embed.py")
	if err != nil {
		return "", err
	}
	return abs, nil
}

// Embed runs the batch end-to-end. Returns embeddings keyed by request
// id. Empty reqs => no subprocess spawn.
func (e *PythonEmbedder) Embed(ctx context.Context, reqs []EmbedRequest) ([]EmbedResult, error) {
	if len(reqs) == 0 {
		return nil, nil
	}
	script, err := e.scriptPath()
	if err != nil {
		return nil, fmt.Errorf("knowledge: resolve embed script: %w", err)
	}
	if _, err := os.Stat(script); err != nil {
		return nil, fmt.Errorf("knowledge: embed script not found at %s: %w", script, ErrEmbedderUnavailable)
	}

	cmd := exec.CommandContext(ctx, e.pythonBin(), script)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("knowledge: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("knowledge: stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("knowledge: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		// exec.ErrNotFound or similar surfaces here on bare-metal envs
		// without python3 in PATH — collapse to the friendly sentinel.
		return nil, fmt.Errorf("%w: %v", ErrEmbedderUnavailable, err)
	}

	// Writer goroutine: stream all requests, then close stdin so the
	// Python side hits EOF and exits cleanly.
	writeErr := make(chan error, 1)
	go func() {
		defer close(writeErr)
		enc := json.NewEncoder(stdin)
		for _, r := range reqs {
			if err := enc.Encode(r); err != nil {
				writeErr <- err
				_ = stdin.Close()
				return
			}
		}
		_ = stdin.Close()
	}()

	// Reader: streamed JSON Lines.
	results := make([]EmbedResult, 0, len(reqs))
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // embedding lines can be ~4KB; pad headroom
	for scanner.Scan() {
		var r EmbedResult
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		results = append(results, r)
	}
	scanErr := scanner.Err()

	// Wait drains stderr + waits for process exit.
	stderrBytes, _ := io.ReadAll(stderr)
	waitErr := cmd.Wait()

	if werr := <-writeErr; werr != nil {
		return nil, fmt.Errorf("knowledge: write batch: %w", werr)
	}
	if scanErr != nil {
		return results, fmt.Errorf("knowledge: read embeddings: %w", scanErr)
	}
	if waitErr != nil {
		// Parse the stderr JSON error if present so the caller sees a
		// usable message ("sentence-transformers not installed ..."),
		// not just "exit status 2".
		var errPayload struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(stderrBytes, &errPayload)
		if errPayload.Error != "" {
			return results, fmt.Errorf("%w: %s", ErrEmbedderUnavailable, errPayload.Error)
		}
		return results, fmt.Errorf("%w: %v (stderr: %s)", ErrEmbedderUnavailable, waitErr, string(stderrBytes))
	}
	return results, nil
}

// MockEmbedder is the test-side Embedder that returns canned vectors.
// CLI / MCP tests use this to verify wiring without spawning Python.
type MockEmbedder struct {
	// Vectors maps request text → embedding. Unmapped text returns nil
	// embedding (callers should treat that as "no vector for this row").
	Vectors map[string][]float32
	// Err is returned verbatim if set, before any vector lookup.
	Err error
}

// Embed satisfies the Embedder contract.
func (m *MockEmbedder) Embed(_ context.Context, reqs []EmbedRequest) ([]EmbedResult, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	out := make([]EmbedResult, 0, len(reqs))
	for _, r := range reqs {
		if v, ok := m.Vectors[r.Text]; ok {
			out = append(out, EmbedResult{ID: r.ID, Embedding: v})
		}
	}
	return out, nil
}
