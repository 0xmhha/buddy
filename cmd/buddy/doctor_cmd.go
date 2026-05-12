package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/diagnose"
)

// newDoctorCmd wires the read-only health snapshot. Render output goes to
// stdout (it is the user-facing report, not log noise). Exit code is 0 when
// the report is healthy, 1 otherwise — matches m4-plan §Task 2.
//
// M5 T3: thresholds (HookTimeoutMs, HookSlowMs, HookFailRatePct, OutboxBacklog)
// are now read from ~/.buddy/config.json via loadEffectiveConfig. A missing
// config file falls back to spec defaults silently. Pass --config <path> to
// point at a different file.
func newDoctorCmd() *cobra.Command {
	var (
		dbFlag     string
		pidFlag    string
		configFlag string
	)
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "hook health 즉시 진단 (read-only, daemon 의존 없음)",
		RunE: func(_ *cobra.Command, _ []string) error {
			eff, err := loadEffectiveConfig(configFlag)
			if err != nil {
				return err
			}
			pidFile, err := resolvePIDFile(pidFlag, dbFlag)
			if err != nil {
				return err
			}
			rep, err := diagnose.Check(buildDoctorOptions(dbFlag, pidFile, eff))
			if err != nil {
				return err
			}
			rep.Render(os.Stdout)
			if !rep.Healthy {
				return errUnhealthy
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&pidFlag, "pid", "", "PID 파일 경로 (기본: <db dir>/daemon.pid)")
	cmd.Flags().StringVar(&configFlag, "config", "", "config 파일 경로 (기본: ~/.buddy/config.json)")
	return cmd
}
