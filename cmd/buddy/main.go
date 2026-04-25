// Command buddy is the user-facing CLI for the buddy harness control plane.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/wm-it-22-00661/buddy/internal/daemon"
	"github.com/wm-it-22-00661/buddy/internal/db"
	"github.com/wm-it-22-00661/buddy/internal/diagnose"
	"github.com/wm-it-22-00661/buddy/internal/hookwrap"
	"github.com/wm-it-22-00661/buddy/internal/install"
	"github.com/wm-it-22-00661/buddy/internal/queries"
	"github.com/wm-it-22-00661/buddy/internal/schema"
)

const version = "0.0.1"

// friendError carries a pre-formatted, friend-tone message that main() prints
// verbatim (no `buddy: ` prefix added) and uses to exit with code 1 instead of
// the generic 2. Use newFriendError to construct.
type friendError struct{ msg string }

func (e *friendError) Error() string { return e.msg }

func newFriendError(msg string) error { return &friendError{msg: msg} }

// errUnhealthy is a sentinel returned by report-driven commands (e.g. doctor)
// that have already printed their full output to stdout and only need main() to
// signal a non-zero exit code. main() recognises it and exits 1 silently — no
// extra "buddy: " line, no message duplication.
var errUnhealthy = errors.New("unhealthy")

func main() {
	root := newRootCmd()
	err := root.ExecuteContext(context.Background())
	if err == nil {
		return
	}
	if errors.Is(err, errUnhealthy) {
		os.Exit(1)
	}
	var fe *friendError
	if errors.As(err, &fe) {
		fmt.Fprintln(os.Stderr, fe.msg)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "buddy: %v\n", err)
	os.Exit(2)
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "buddy",
		Short:         "Claude Code 옆에서 hook 신뢰성을 지켜주는 친구",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newHookWrapCmd())
	root.AddCommand(newDaemonCmd())
	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newStatsCmd())
	root.AddCommand(newEventsCmd())
	return root
}

// newEventsCmd wires the read-only hook_events tail. Output is structured
// (one line per event) on stdout — debug surface, not friend tone. With
// --follow the command installs a signal-aware context so Ctrl-C / SIGTERM
// cleanly stops the polling loop and the command writes friend-tone start /
// end markers to stderr (the only friendly touch on this command). See
// m4-plan §Task 4.
func newEventsCmd() *cobra.Command {
	var (
		dbFlag     string
		hookFlag   string
		limitFlag  int
		followFlag bool
	)
	cmd := &cobra.Command{
		Use:   "events",
		Short: "최근 hook_events tail (read-only, --follow 으로 실시간)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := queries.EventsOptions{
				DBPath:     dbFlag,
				HookFilter: hookFlag,
				Limit:      limitFlag,
			}
			if followFlag {
				ctx, stop := signal.NotifyContext(cmd.Context(),
					syscall.SIGINT, syscall.SIGTERM)
				defer stop()
				if err := queries.Follow(ctx, opts, os.Stdout); err != nil {
					if errors.Is(err, queries.ErrInvalidLimit) {
						return newFriendError("buddy: " + err.Error())
					}
					return newFriendError(fmt.Sprintf(
						"buddy: events follow 실패 (%v)", err))
				}
				return nil
			}
			res, err := queries.RunEvents(opts)
			if err != nil {
				if errors.Is(err, queries.ErrInvalidLimit) {
					return newFriendError("buddy: " + err.Error())
				}
				return newFriendError(fmt.Sprintf(
					"buddy: DB를 못 읽었어. daemon이 한 번이라도 돈 적 있어? (%v)", err))
			}
			res.RenderLines(os.Stdout)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&hookFlag, "hook", "", "특정 hook 이름으로 필터 (대소문자 무시)")
	cmd.Flags().IntVar(&limitFlag, "limit", 20, "표시할 최근 event 개수 (기본 20)")
	cmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "새 event를 1초 간격으로 따라가기")
	return cmd
}

// newStatsCmd wires the read-only hook_stats report. Output goes to stdout
// (it's the user-facing report, not log noise). Exit is always 0 unless we
// fail to open the DB or the user passes a bad --window — both surface as
// friendError values that main() prints and exits 1 on. See m4-plan §Task 3.
func newStatsCmd() *cobra.Command {
	var (
		dbFlag     string
		windowFlag string
		byToolFlag bool
		hookFlag   string
	)
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "최근 hook 통계 (read-only, daemon 의존 없음)",
		RunE: func(_ *cobra.Command, _ []string) error {
			res, err := queries.Run(queries.Options{
				DBPath:     dbFlag,
				Window:     windowFlag,
				ByTool:     byToolFlag,
				HookFilter: hookFlag,
			})
			if err != nil {
				if errors.Is(err, queries.ErrInvalidWindow) {
					return newFriendError("buddy: " + err.Error())
				}
				// Any other failure is a DB-side problem (open or query). Match
				// doctor's wording so the two read-only commands feel consistent.
				return newFriendError(fmt.Sprintf(
					"buddy: DB를 못 읽었어. daemon이 한 번이라도 돈 적 있어? (%v)", err))
			}
			res.Render(os.Stdout)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&windowFlag, "window", "1h", "집계 윈도우: 5m | 1h | 24h")
	cmd.Flags().BoolVar(&byToolFlag, "by-tool", false, "tool 단위로 분리해서 보기")
	cmd.Flags().StringVar(&hookFlag, "hook", "", "특정 hook 이름으로 필터")
	return cmd
}

// newDoctorCmd wires the read-only health snapshot. Render output goes to
// stdout (it is the user-facing report, not log noise). Exit code is 0 when
// the report is healthy, 1 otherwise — matches m4-plan §Task 2.
func newDoctorCmd() *cobra.Command {
	var (
		dbFlag  string
		pidFlag string
	)
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "hook health 즉시 진단 (read-only, daemon 의존 없음)",
		RunE: func(_ *cobra.Command, _ []string) error {
			pidFile, err := resolvePIDFile(pidFlag, dbFlag)
			if err != nil {
				return err
			}
			rep, err := diagnose.Check(diagnose.Options{
				DBPath:  dbFlag,
				PIDFile: pidFile,
			})
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
	return cmd
}

func newInstallCmd() *cobra.Command {
	var (
		claudeDirFlag string
		buddyDirFlag  string
		binaryFlag    string
		dbFlag        string
		withCliwrap   bool
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Claude Code settings.json의 hook들을 buddy로 감싼다",
		RunE: func(_ *cobra.Command, _ []string) error {
			res, err := install.Install(install.Options{
				ClaudeDir:   claudeDirFlag,
				BuddyDir:    buddyDirFlag,
				BuddyBinary: binaryFlag,
				DBPath:      dbFlag,
				WithCliwrap: withCliwrap,
			})
			if err != nil {
				return translateInstallError(err)
			}
			if res.NoOp {
				fmt.Fprintln(os.Stderr, "buddy: 이미 등록되어 있어. 변화 없음.")
			} else {
				fmt.Fprintln(os.Stderr, "buddy: 등록 완료. 이제 옆에서 보고 있을게.")
			}
			if res.CliwrapWritten {
				fmt.Fprintf(os.Stderr, "buddy: cliwrap.yaml 도 써뒀어 (%s).\n", res.CliwrapPath)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&claudeDirFlag, "claude-dir", "", "Claude Code 설정 디렉터리 (기본: ~/.claude)")
	cmd.Flags().StringVar(&buddyDirFlag, "buddy-dir", "", "buddy 작업 디렉터리 (기본: ~/.buddy)")
	cmd.Flags().StringVar(&binaryFlag, "buddy-binary", "", "buddy 바이너리 절대 경로 (기본: 현재 실행 파일)")
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: <buddy-dir>/buddy.db). 이 경로로 DB가 만들어지고 cliwrap.yaml 에도 들어가. 이후 daemon/doctor/stats/events 에는 같은 --db 를 직접 줘야 해.")
	cmd.Flags().BoolVar(&withCliwrap, "with-cliwrap", false, "cliwrap.yaml 도 함께 생성")
	return cmd
}

// translateInstallError maps install/uninstall sentinel errors to friend-tone
// friendError values that main() prints verbatim and exits 1 on.
func translateInstallError(err error) error {
	var spaceErr *install.BinaryPathSpaceError
	switch {
	case errors.Is(err, install.ErrSettingsMissing):
		return newFriendError("buddy: ~/.claude/settings.json 이 안 보여. Claude Code 설치되어 있어?")
	case errors.As(err, &spaceErr):
		return newFriendError(fmt.Sprintf(
			"buddy: 바이너리 경로에 공백이 있어. 다른 경로로 옮겨봐: %s", spaceErr.Path))
	}
	return err
}

func newUninstallCmd() *cobra.Command {
	var (
		claudeDirFlag string
		buddyDirFlag  string
		binaryFlag    string
	)
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "settings.json의 buddy hook wrapping 을 제거 (백업 우선 복원)",
		RunE: func(_ *cobra.Command, _ []string) error {
			res, err := install.Uninstall(install.Options{
				ClaudeDir:   claudeDirFlag,
				BuddyDir:    buddyDirFlag,
				BuddyBinary: binaryFlag,
			})
			if err != nil {
				return translateInstallError(err)
			}
			switch {
			case res.RestoredFromBackup:
				fmt.Fprintln(os.Stderr, "buddy: 해제 완료. 백업에서 복원했어.")
			case res.Unwrapped > 0:
				fmt.Fprintln(os.Stderr, "buddy: 해제 완료. wrapping 제거했어.")
			default:
				fmt.Fprintln(os.Stderr, "buddy: 등록된 게 없어. 그대로 둘게.")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&claudeDirFlag, "claude-dir", "", "Claude Code 설정 디렉터리 (기본: ~/.claude)")
	cmd.Flags().StringVar(&buddyDirFlag, "buddy-dir", "", "buddy 작업 디렉터리 (기본: ~/.buddy)")
	cmd.Flags().StringVar(&binaryFlag, "buddy-binary", "", "buddy 바이너리 절대 경로 (기본: 현재 실행 파일)")
	return cmd
}

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

func newDaemonRunCmd() *cobra.Command {
	var (
		dbFlag    string
		pollFlag  time.Duration
		batchFlag int
		pidFlag   string
	)
	cmd := &cobra.Command{
		Use:   "run",
		Short: "foreground 실행 (cli-wrapper supervise용)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return daemon.Run(cmd.Context(), daemon.Config{
				DBPath:       dbFlag,
				PIDFile:      pidFlag,
				PollInterval: pollFlag,
				BatchSize:    batchFlag,
			})
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&pidFlag, "pid", "", "PID 파일 경로 (기본: <db dir>/daemon.pid)")
	cmd.Flags().DurationVar(&pollFlag, "poll", time.Second, "outbox poll 간격")
	cmd.Flags().IntVar(&batchFlag, "batch", 500, "한 tick에 처리할 outbox row 상한")
	return cmd
}

func newDaemonStartCmd() *cobra.Command {
	var (
		dbFlag   string
		pidFlag  string
		pollFlag time.Duration
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
				fmt.Fprintf(os.Stderr, "buddy: 이미 실행 중이야 (pid %d).\n", st.PID)
				return nil
			}
			return spawnDetached(dbFlag, pidFile, pollFlag)
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로")
	cmd.Flags().StringVar(&pidFlag, "pid", "", "PID 파일 경로")
	cmd.Flags().DurationVar(&pollFlag, "poll", time.Second, "outbox poll 간격")
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
				fmt.Fprintln(os.Stderr, "buddy: 실행 중인 daemon이 없어.")
				return nil
			}
			if err := daemon.Stop(pidFile); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "buddy: daemon에 종료 신호 보냈어 (pid %d).\n", st.PID)
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

func spawnDetached(dbFlag, pidFile string, poll time.Duration) error {
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
	cmd := exec.Command(self, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	pid, err := startAndDetach(cmd)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "buddy: daemon 시작 (pid %d).\n", pid)
	return nil
}

// keep context alive in case cmd switches off ExecuteContext later
var _ = context.Background

func newHookWrapCmd() *cobra.Command {
	var (
		eventFlag  string
		dbFlag     string
		recordArgs bool
		tagFlags   []string
	)
	cmd := &cobra.Command{
		Use:   "hook-wrap <hook-name> [-- <original-command...>]",
		Short: "Claude Code hook을 감싸 실행하고 outbox에 기록한다",
		Long: `Claude Code hook을 감싸 실행한다.
stdin/stdout/stderr/exit code를 그대로 전달하며, 실행 결과를 buddy outbox에 기록한다.

invariant 7가지 (v0.1 spec §7.1):
  1. wrapper는 stdout에 자기 출력을 절대 쓰지 않는다 (LLM 컨텍스트 오염 방지).
  2. child stdout/stderr는 buffering 없이 부모 stdio에 직접 pipe (streaming).
  3. stdin은 buddy가 한 번 buffering한 뒤 child에 다시 흘려준다.
  4. exit code 그대로 통과 (signal 종료는 128+sigNo).
  5. outbox 기록 실패는 hook의 exit code를 바꾸지 않는다.
  6. <original-command>가 비면 monitoring-only 모드, exit 0.
  7. malformed input은 흡수, wrapper 자체는 절대 깨지지 않는다.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hookName := args[0]
			command := args[1:]

			res, err := hookwrap.Run(cmd.Context(), hookwrap.Options{
				HookName:       hookName,
				Command:        command,
				FallbackEvent:  schema.HookEventName(eventFlag),
				DBPath:         dbFlag,
				RecordToolArgs: recordArgs,
				CustomTags:     parseTags(tagFlags),
			})
			if err != nil {
				return err
			}
			os.Exit(res.ExitCode)
			return nil // unreachable
		},
	}
	cmd.Flags().StringVar(&eventFlag, "event", "PreToolUse",
		"기본 event (stdin이 비었을 때 fallback)")
	cmd.Flags().StringVar(&dbFlag, "db", "",
		"buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().BoolVar(&recordArgs, "record-tool-args", false,
		"tool_input을 outbox에 기록 (default off, privacy)")
	cmd.Flags().StringSliceVar(&tagFlags, "tag", nil,
		"customTags 추가. 형식: key=value. 반복 가능.")
	return cmd
}

func parseTags(raw []string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, item := range raw {
		eq := strings.Index(item, "=")
		if eq <= 0 {
			continue
		}
		k := strings.TrimSpace(item[:eq])
		v := strings.TrimSpace(item[eq+1:])
		if k != "" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
