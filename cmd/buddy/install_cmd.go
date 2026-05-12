package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/install"
	"github.com/0xmhha/buddy/internal/persona"
)

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
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyInstallNoOp))
			} else {
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyInstallDone))
			}
			if res.CliwrapWritten {
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyInstallCliwrapWritten, res.CliwrapPath))
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
		return newFriendError(persona.M(persona.KeyInstallSettingsMissing))
	case errors.As(err, &spaceErr):
		return newFriendError(persona.M(persona.KeyInstallBinaryHasSpaces, spaceErr.Path))
	}
	return err
}

func newUninstallCmd() *cobra.Command {
	var (
		claudeDirFlag string
		buddyDirFlag  string
		binaryFlag    string
		dbFlag        string
		keepDaemon    bool
	)
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "settings.json의 buddy hook wrapping 을 제거 (백업 우선 복원)",
		RunE: func(_ *cobra.Command, _ []string) error {
			res, err := install.Uninstall(install.Options{
				ClaudeDir:   claudeDirFlag,
				BuddyDir:    buddyDirFlag,
				BuddyBinary: binaryFlag,
				DBPath:      dbFlag,
				KeepDaemon:  keepDaemon,
			})
			if err != nil {
				return translateInstallError(err)
			}
			switch {
			case res.RestoredFromBackup:
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyUninstallRestoredFromBackup))
			case res.Unwrapped > 0:
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyUninstallRemovedWrapping))
			default:
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyUninstallNothingRegistered))
			}
			// M5 T9: friend-tone note about daemon disposition. The Uninstall
			// call already attempted (or skipped) the stop based on
			// KeepDaemon — we just report what happened.
			if res.DaemonWasRunning {
				switch {
				case res.DaemonStopped:
					fmt.Fprintln(os.Stderr, persona.M(persona.KeyUninstallDaemonStopped))
				case keepDaemon:
					fmt.Fprintln(os.Stderr, persona.M(persona.KeyUninstallDaemonKept))
				default:
					fmt.Fprintln(os.Stderr, persona.M(persona.KeyUninstallDaemonNotStopping))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&claudeDirFlag, "claude-dir", "", "Claude Code 설정 디렉터리 (기본: ~/.claude)")
	cmd.Flags().StringVar(&buddyDirFlag, "buddy-dir", "", "buddy 작업 디렉터리 (기본: ~/.buddy)")
	cmd.Flags().StringVar(&binaryFlag, "buddy-binary", "", "buddy 바이너리 절대 경로 (기본: 현재 실행 파일)")
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: <buddy-dir>/buddy.db). daemon PID 위치 추론에 사용.")
	cmd.Flags().BoolVar(&keepDaemon, "keep-daemon", false, "daemon이 떠있어도 자동 stop 하지 않음")
	return cmd
}
