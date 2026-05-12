package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/hookwrap"
	"github.com/0xmhha/buddy/internal/schema"
)

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
