package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/daemon"
	"github.com/0xmhha/buddy/internal/db"
	"github.com/0xmhha/buddy/internal/persona"
)

func newDaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "백그라운드 집계 daemon 관리 (run/start/stop/status)",
	}
	cmd.AddCommand(
		newDaemonRunCmd(),
		newDaemonStartCmd(),
		newDaemonStopCmd(),
		newDaemonStatusCmd(),
	)
	return cmd
}

// newDaemonRunCmd wires `buddy daemon run`. pollInterval / batchSize
// are read from ~/.buddy/config.json (spec defaults: 1s / 500). Explicit
// --poll / --batch flags still win — the precedence is flag > config > default.
// Zero is the "use config / default" sentinel for the flags, so the help text
// reflects that rather than the old hard-coded 1s / 500.
func newDaemonRunCmd() *cobra.Command {
	var (
		dbFlag     string
		pollFlag   time.Duration
		batchFlag  int
		pidFlag    string
		configFlag string
	)
	cmd := &cobra.Command{
		Use:   "run",
		Short: "foreground 실행 (cli-wrapper supervise용)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			eff, err := loadEffectiveConfig(configFlag)
			if err != nil {
				return err
			}
			return daemon.Run(cmd.Context(),
				resolveDaemonRunConfig(dbFlag, pidFlag, pollFlag, batchFlag, eff))
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&pidFlag, "pid", "", "PID 파일 경로 (기본: <db dir>/daemon.pid)")
	cmd.Flags().DurationVar(&pollFlag, "poll", 0, "outbox poll 간격 (기본: config 또는 1s)")
	cmd.Flags().IntVar(&batchFlag, "batch", 0, "한 tick에 처리할 outbox row 상한 (기본: config 또는 500)")
	cmd.Flags().StringVar(&configFlag, "config", "", "config 파일 경로 (기본: ~/.buddy/config.json)")
	return cmd
}

// newDaemonStartCmd spawns `buddy daemon run` detached. Flags here mirror
// `daemon run` so the user can express the same intent at start time; we
// forward them as argv to the spawned child, which then loads config itself.
// --config and --batch are propagated alongside --poll.
func newDaemonStartCmd() *cobra.Command {
	var (
		dbFlag     string
		pidFlag    string
		pollFlag   time.Duration
		batchFlag  int
		configFlag string
	)
	cmd := &cobra.Command{
		Use:   "start",
		Short: "background로 daemon 띄우기 (detach fork)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pidFile, err := resolvePIDFile(pidFlag, dbFlag)
			if err != nil {
				return err
			}
			st, _ := daemon.CheckStatus(pidFile)
			if st.Running {
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyDaemonAlreadyRunning, st.PID))
				return nil
			}
			return spawnDetached(dbFlag, pidFile, pollFlag, batchFlag, configFlag)
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로")
	cmd.Flags().StringVar(&pidFlag, "pid", "", "PID 파일 경로")
	cmd.Flags().DurationVar(&pollFlag, "poll", 0, "outbox poll 간격 (기본: config 또는 1s)")
	cmd.Flags().IntVar(&batchFlag, "batch", 0, "한 tick에 처리할 outbox row 상한 (기본: config 또는 500)")
	cmd.Flags().StringVar(&configFlag, "config", "", "config 파일 경로 (기본: ~/.buddy/config.json)")
	return cmd
}

func newDaemonStopCmd() *cobra.Command {
	var pidFlag, dbFlag string
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "실행 중인 daemon에 SIGTERM",
		RunE: func(_ *cobra.Command, _ []string) error {
			pidFile, err := resolvePIDFile(pidFlag, dbFlag)
			if err != nil {
				return err
			}
			st, err := daemon.CheckStatus(pidFile)
			if err != nil {
				return err
			}
			if !st.Running {
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyDaemonNotRunning))
				return nil
			}
			if err := daemon.Stop(pidFile); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, persona.M(persona.KeyDaemonStopSignalSent, st.PID))
			return nil
		},
	}
	cmd.Flags().StringVar(&pidFlag, "pid", "", "PID 파일 경로")
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로")
	return cmd
}

func newDaemonStatusCmd() *cobra.Command {
	var pidFlag, dbFlag string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "daemon 실행 여부 확인",
		RunE: func(_ *cobra.Command, _ []string) error {
			pidFile, err := resolvePIDFile(pidFlag, dbFlag)
			if err != nil {
				return err
			}
			st, err := daemon.CheckStatus(pidFile)
			if err != nil {
				return err
			}
			if st.Running {
				fmt.Printf("running (pid %d)\n", st.PID)
			} else {
				fmt.Println("not running")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&pidFlag, "pid", "", "PID 파일 경로")
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로")
	return cmd
}

func resolvePIDFile(pidFlag, dbFlag string) (string, error) {
	if pidFlag != "" {
		return pidFlag, nil
	}
	if dbFlag != "" {
		return defaultPIDFromDB(dbFlag), nil
	}
	d, err := db.DefaultPath()
	if err != nil {
		return "", err
	}
	return defaultPIDFromDB(d), nil
}

func defaultPIDFromDB(dbPath string) string {
	dir := dbPath
	for i := len(dir) - 1; i >= 0; i-- {
		if dir[i] == '/' {
			dir = dir[:i]
			break
		}
	}
	if dir == dbPath {
		dir = "."
	}
	return dir + "/daemon.pid"
}

// spawnDetached launches `buddy daemon run` as a detached child. Each flag
// is forwarded only when set (non-zero / non-empty) so the child can fall
// back to its own config / spec defaults — the parent does NOT pre-resolve
// poll / batch here, that resolution happens once on the child via
// loadEffectiveConfig.
func spawnDetached(dbFlag, pidFile string, poll time.Duration, batch int, configFlag string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate self: %w", err)
	}
	args := []string{"daemon", "run"}
	if dbFlag != "" {
		args = append(args, "--db", dbFlag)
	}
	if pidFile != "" {
		args = append(args, "--pid", pidFile)
	}
	if poll > 0 {
		args = append(args, "--poll", poll.String())
	}
	if batch > 0 {
		args = append(args, "--batch", strconv.Itoa(batch))
	}
	if configFlag != "" {
		args = append(args, "--config", configFlag)
	}
	cmd := exec.Command(self, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	pid, err := startAndDetach(cmd)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, persona.M(persona.KeyDaemonStarted, pid))
	return nil
}
