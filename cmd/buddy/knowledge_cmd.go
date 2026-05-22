package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/knowledge"
	"github.com/0xmhha/buddy/internal/sessions"
)

// newKnowledgeCmd wires `buddy knowledge ...` per ADR-014. Surfaces:
//
//   - ingest  : transcript JSONL → chunks (+ optional embedding pass).
//   - query   : BM25 / vector / hybrid retrieval against the corpus.
//   - stats   : chunk count / embedding coverage.
//
// Phase 2 (v0.11.0) will add `buddy advise` consuming this surface.
func newKnowledgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Local knowledge retrieval over your Claude Code sessions (ADR-014)",
	}
	cmd.AddCommand(
		newKnowledgeIngestCmd(),
		newKnowledgeQueryCmd(),
		newKnowledgeStatsCmd(),
	)
	return cmd
}

func openKnowledgeStore(dbFlag string) (*knowledge.Store, *sessions.Store, func(), error) {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open db: %w", err)
	}
	return knowledge.NewStore(conn), sessions.NewStore(conn), func() { _ = conn.Close() }, nil
}

// ─── ingest ────────────────────────────────────────────────────────────

func newKnowledgeIngestCmd() *cobra.Command {
	var (
		dbFlag      string
		sessionID   string
		all         bool
		rebuild     bool
		skipEmbed   bool
		pythonBin   string
		scriptPath  string
	)
	c := &cobra.Command{
		Use:   "ingest",
		Short: "Chunk sessions transcripts and (optionally) compute embeddings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if sessionID == "" && !all {
				return errors.New("--session <id> 또는 --all 필요해")
			}
			kstore, sstore, closer, err := openKnowledgeStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			var targets []sessions.Session
			if sessionID != "" {
				s, err := sstore.Get(ctx, sessionID)
				if err != nil {
					return err
				}
				targets = []sessions.Session{s}
			} else {
				targets, err = sstore.List(ctx, sessions.ListOptions{IncludeEnded: true})
				if err != nil {
					return err
				}
			}
			if len(targets) == 0 {
				fmt.Println("ingest 대상 세션이 없어. `buddy session list --all` 로 확인해줘.")
				return nil
			}

			chunker := knowledge.NewChunker()
			var totalNew int
			for _, s := range targets {
				if rebuild {
					n, _ := kstore.DeleteBySession(ctx, s.ID)
					if n > 0 {
						fmt.Printf("buddy: %s 의 기존 chunk %d개 삭제 (rebuild)\n", short(s.ID), n)
					}
				}
				chunks, err := chunker.Extract(s.TranscriptPath)
				if err != nil {
					fmt.Printf("buddy: %s transcript 읽기 실패: %v\n", short(s.ID), err)
					continue
				}
				if len(chunks) == 0 {
					continue
				}
				var ids []int64
				for _, ch := range chunks {
					id, err := kstore.Insert(ctx, ch)
					if err != nil {
						return fmt.Errorf("insert chunk for %s: %w", s.ID, err)
					}
					ids = append(ids, id)
				}
				totalNew += len(chunks)
				fmt.Printf("buddy: %s → %d chunks 추가\n", short(s.ID), len(chunks))

				if skipEmbed {
					continue
				}
				if err := embedNew(ctx, kstore, chunks, ids, pythonBin, scriptPath); err != nil {
					fmt.Printf("buddy: embedding pass 건너뜀: %v\n", err)
					// Stop further embed attempts — same env will fail.
					skipEmbed = true
				}
			}
			fmt.Printf("buddy: ingest 완료 — 신규 chunk 합계 %d개\n", totalNew)
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().StringVar(&sessionID, "session", "", "target a single session id")
	c.Flags().BoolVar(&all, "all", false, "ingest every observed session")
	c.Flags().BoolVar(&rebuild, "rebuild", false, "delete existing chunks for the target before re-chunking")
	c.Flags().BoolVar(&skipEmbed, "skip-embed", false, "skip the Python embedding pass (BM25 only)")
	c.Flags().StringVar(&pythonBin, "python", "", "python interpreter (default python3)")
	c.Flags().StringVar(&scriptPath, "embed-script", "", "path to embed.py (default ./scripts/embed.py or BUDDY_EMBED_SCRIPT env)")
	return c
}

// embedNew runs the embed pipeline over the just-inserted chunks and
// writes the resulting vectors back. Errors collapse to
// ErrEmbedderUnavailable so the CLI prints one friendly note and
// continues with BM25-only behaviour.
func embedNew(ctx context.Context, store *knowledge.Store, chunks []knowledge.Chunk, ids []int64, pythonBin, scriptPath string) error {
	if len(chunks) == 0 {
		return nil
	}
	embedder := &knowledge.PythonEmbedder{PythonBin: pythonBin, ScriptPath: scriptPath}
	reqs := make([]knowledge.EmbedRequest, 0, len(chunks))
	for i, ch := range chunks {
		reqs = append(reqs, knowledge.EmbedRequest{ID: ids[i], Text: ch.Content})
	}
	results, err := embedder.Embed(ctx, reqs)
	if err != nil {
		return err
	}
	for _, r := range results {
		if err := store.UpdateEmbedding(ctx, r.ID, r.Embedding); err != nil {
			return fmt.Errorf("update embedding %d: %w", r.ID, err)
		}
	}
	return nil
}

// ─── query ─────────────────────────────────────────────────────────────

func newKnowledgeQueryCmd() *cobra.Command {
	var (
		dbFlag      string
		limit       int
		channel     string
		pythonBin   string
		scriptPath  string
	)
	c := &cobra.Command{
		Use:   "query <text>",
		Short: "Retrieve top-N chunks for a text query (BM25 / vector / hybrid)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			text := strings.Join(args, " ")
			kstore, _, closer, err := openKnowledgeStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()

			all, err := kstore.All(ctx)
			if err != nil {
				return err
			}
			if len(all) == 0 {
				fmt.Println("chunks 가 비어 있어. `buddy knowledge ingest --all` 먼저 실행해줘.")
				return nil
			}

			idx := knowledge.NewBM25Index(all)
			var hits []knowledge.ScoredChunk
			switch channel {
			case "bm25":
				hits = idx.Search(text, limit)
			case "vector":
				emb, err := embedQuery(ctx, text, pythonBin, scriptPath)
				if err != nil {
					return err
				}
				hits = knowledge.VectorSearch(all, emb, limit)
			default: // "hybrid"
				emb, err := embedQuery(ctx, text, pythonBin, scriptPath)
				if err != nil {
					fmt.Printf("buddy: embedding 사용 못 함 (%v) — BM25 만 사용\n", err)
					emb = nil
				}
				hits = knowledge.HybridSearch(idx, all, text, emb, limit)
			}
			if len(hits) == 0 {
				fmt.Println("매칭된 chunk 가 없어.")
				return nil
			}
			fmt.Println(renderKnowledgeHits(hits))
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().IntVar(&limit, "k", 5, "max chunks to return")
	c.Flags().StringVar(&channel, "channel", "hybrid", "retrieval channel: hybrid | bm25 | vector")
	c.Flags().StringVar(&pythonBin, "python", "", "python interpreter (default python3)")
	c.Flags().StringVar(&scriptPath, "embed-script", "", "path to embed.py (default ./scripts/embed.py or BUDDY_EMBED_SCRIPT env)")
	return c
}

// embedQuery is a one-shot embedder for a single query string.
func embedQuery(ctx context.Context, text, pythonBin, scriptPath string) ([]float32, error) {
	embedder := &knowledge.PythonEmbedder{PythonBin: pythonBin, ScriptPath: scriptPath}
	res, err := embedder.Embed(ctx, []knowledge.EmbedRequest{{ID: 0, Text: text}})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, errors.New("empty embedding result")
	}
	return res[0].Embedding, nil
}

// ─── stats ─────────────────────────────────────────────────────────────

func newKnowledgeStatsCmd() *cobra.Command {
	var dbFlag string
	c := &cobra.Command{
		Use:   "stats",
		Short: "Show chunk count, embedding coverage, last ingest time",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			kstore, _, closer, err := openKnowledgeStore(dbFlag)
			if err != nil {
				return err
			}
			defer closer()
			total, err := kstore.Count(ctx)
			if err != nil {
				return err
			}
			withEmb, err := kstore.CountWithEmbedding(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("chunks         : %d\n", total)
			fmt.Printf("with embedding : %d (%s)\n", withEmb, pctStr(withEmb, total))
			fmt.Printf("BM25 only      : %d\n", total-withEmb)
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	return c
}

func pctStr(num, denom int64) string {
	if denom == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.0f%%", float64(num)/float64(denom)*100)
}

// ─── render ───────────────────────────────────────────────────────────

func renderKnowledgeHits(hits []knowledge.ScoredChunk) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-6s %-10s %-8s %s\n", "RANK", "CHANNEL", "SCORE", "CONTENT (first 80 chars)")
	for i, h := range hits {
		preview := strings.ReplaceAll(h.Content, "\n", " ")
		if len(preview) > 80 {
			preview = preview[:80] + "…"
		}
		fmt.Fprintf(&b, "%-6d %-10s %-8.4f %s\n", i+1, h.Channel, h.Score, preview)
	}
	return b.String()
}

func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
