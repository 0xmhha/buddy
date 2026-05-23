package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/advisor"
	"github.com/0xmhha/buddy/internal/config"
	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/knowledge"
	"github.com/0xmhha/buddy/internal/sessions"
	"github.com/0xmhha/buddy/internal/usage"
)

// newAdviseCmd wires `buddy advise ...`. Subcommands:
//
//   - (root)    : current advisories (run rules now, optionally persist)
//   - history   : list persisted advisories
//   - mute      : flip muted flag on one advisory or all of a kind
func newAdviseCmd() *cobra.Command {
	var (
		dbFlag      string
		configFlag  string
		persist     bool
		allHistory  bool
		sinceStr    string
		pythonBin   string
		scriptPath  string
		muteID      int64
		muteKind    string
	)
	c := &cobra.Command{
		Use:   "advise",
		Short: "Friend-tone advisories generated from usage + knowledge",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Mute path forks first — no evaluation needed.
			if muteID > 0 || muteKind != "" {
				return handleMute(ctx, dbFlag, muteID, muteKind)
			}

			if allHistory {
				return showHistory(ctx, dbFlag, sinceStr)
			}

			runner, closer, err := buildEvaluator(dbFlag, configFlag, pythonBin, scriptPath)
			if err != nil {
				return err
			}
			defer closer()

			var advs []advisor.Advisory
			if persist {
				advs, err = runner.Persist(ctx)
			} else {
				advs, err = runner.Run(ctx)
			}
			if err != nil {
				return err
			}
			if len(advs) == 0 {
				fmt.Println("buddy: 지금은 특별히 알릴 조언이 없어. 다른 작업도 해봐.")
				return nil
			}
			fmt.Println(renderAdvisories(advs))
			if persist {
				fmt.Printf("\nbuddy: %d 개 advisor 행을 advisories 테이블에 기록했어.\n", len(advs))
			}
			return nil
		},
	}
	c.Flags().StringVar(&dbFlag, "db", "", "path to buddy.db (default ~/.buddy/buddy.db)")
	c.Flags().StringVar(&configFlag, "config", "", "path to config.json (default ~/.buddy/config.json)")
	c.Flags().BoolVar(&persist, "persist", false, "write fresh advisories to the advisories table")
	c.Flags().BoolVar(&allHistory, "all-history", false, "list persisted advisories instead of running rules")
	c.Flags().StringVar(&sinceStr, "since", "", "with --all-history: lookback duration (e.g. 24h, 7d)")
	c.Flags().StringVar(&pythonBin, "python", "", "python interpreter for retrieval embedder (default python3)")
	c.Flags().StringVar(&scriptPath, "embed-script", "", "path to embed.py (default ./scripts/embed.py)")
	c.Flags().Int64Var(&muteID, "mute", 0, "mute a single advisory by id")
	c.Flags().StringVar(&muteKind, "mute-kind", "", "mute every advisory of the given kind")
	return c
}

// buildEvaluator opens buddy.db + loads the user config and returns a
// fully-wired advisor runner. closer drains both opens.
func buildEvaluator(dbFlag, configFlag, pythonBin, scriptPath string) (*advisor.Evaluator, func(), error) {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	closer := func() { _ = conn.Close() }

	t := advisor.DefaultThresholds()
	if cfg, err := loadConfigForAdvise(configFlag); err == nil {
		eff := cfg.Effective()
		t = advisor.Thresholds{
			Disabled:              eff.AdvisorDisabled,
			TokenSpikeRatio:       eff.AdvisorTokenSpikeRatio,
			LongSessionHours:      eff.AdvisorLongSessionHours,
			LowCachePct:           eff.AdvisorLowCachePct,
			SessionVolumePerDay:   eff.AdvisorSessionVolumePerDay,
			TokenDailyThreshold:   eff.AdvisorTokenDailyThreshold,
			GoalDriftDisabled:     eff.AdvisorGoalDriftDisabled,
			GoalDriftThreshold:    eff.AdvisorGoalDriftThreshold,
			GoalDriftSampleChunks: eff.AdvisorGoalDriftSampleChunks,
			DedupWindow:           eff.AdvisorDedupWindow,
			PollInterval:          eff.AdvisorPollInterval,
		}.WithDefaults()
	}

	runner := &advisor.Evaluator{
		Thresholds: t,
		Usage:      usage.NewService(conn),
		Sessions:   sessions.NewStore(conn),
		Knowledge:  knowledge.NewStore(conn),
		Embedder:   &knowledge.PythonEmbedder{PythonBin: pythonBin, ScriptPath: scriptPath},
		Advisories: advisor.NewStore(conn),
	}
	return runner, closer, nil
}

// loadConfigForAdvise picks the explicit --config path or falls back to
// the default location. Silently returns zero Config + error on read
// failure — the runner uses DefaultThresholds() in that case.
func loadConfigForAdvise(path string) (config.Config, error) {
	if path == "" {
		p, err := config.DefaultPath()
		if err != nil {
			return config.Config{}, err
		}
		path = p
	}
	return config.Load(path)
}

func showHistory(ctx context.Context, dbFlag, sinceStr string) error {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return err
	}
	defer conn.Close()
	store := advisor.NewStore(conn)

	opts := advisor.ListOptions{IncludeMuted: true}
	if sinceStr != "" {
		dur, err := parseDurationDays(sinceStr)
		if err != nil {
			return fmt.Errorf("--since: %w", err)
		}
		opts.Since = time.Now().UTC().Add(-dur)
	}
	advs, err := store.List(ctx, opts)
	if err != nil {
		return err
	}
	if len(advs) == 0 {
		fmt.Println("buddy: 기록된 advisor 행이 없어. `buddy advise --persist` 로 첫 행을 만들어봐.")
		return nil
	}
	fmt.Println(renderAdvisoryHistory(advs))
	return nil
}

func handleMute(ctx context.Context, dbFlag string, muteID int64, muteKind string) error {
	conn, err := db.Open(db.Options{Path: dbFlag})
	if err != nil {
		return err
	}
	defer conn.Close()
	store := advisor.NewStore(conn)

	if muteID > 0 {
		if err := store.Mute(ctx, muteID); err != nil {
			if errors.Is(err, advisor.ErrNotFound) {
				return fmt.Errorf("advisor id=%d 가 없어", muteID)
			}
			return err
		}
		fmt.Printf("buddy: advisor #%d 음소거 했어.\n", muteID)
		return nil
	}
	n, err := store.MuteKind(ctx, muteKind)
	if err != nil {
		return err
	}
	fmt.Printf("buddy: kind=%q 의 %d 행을 음소거 했어.\n", muteKind, n)
	return nil
}

// ─── render ──────────────────────────────────────────────────────────

func renderAdvisories(advs []advisor.Advisory) string {
	var b strings.Builder
	for i, a := range advs {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%s · %s]\n  %s\n", a.Kind, a.Severity, a.Message)
		if len(a.Evidence) > 0 {
			fmt.Fprintln(&b, "  근거:")
			for _, ev := range a.Evidence {
				if ev.Type == "chunk" && ev.ChunkID > 0 {
					fmt.Fprintf(&b, "    · [chunk #%d] %s\n", ev.ChunkID, ev.Detail)
				} else {
					fmt.Fprintf(&b, "    · [%s] %s\n", ev.Type, ev.Detail)
				}
			}
		}
	}
	return b.String()
}

func renderAdvisoryHistory(advs []advisor.Advisory) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-6s %-22s %-6s %-19s %s\n",
		"ID", "KIND", "SEV", "CREATED", "MESSAGE")
	for _, a := range advs {
		muted := ""
		if a.Muted {
			muted = " (muted)"
		}
		msg := a.Message
		if len(msg) > 60 {
			msg = msg[:60] + "…"
		}
		fmt.Fprintf(&b, "%-6d %-22s %-6s %-19s %s%s\n",
			a.ID, a.Kind, a.Severity,
			a.CreatedAt.Local().Format("2006-01-02 15:04"),
			msg, muted)
	}
	return b.String()
}
