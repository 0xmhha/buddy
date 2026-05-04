package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newMcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "buddy MCP 서버 등록/해제 관리",
	}
	cmd.AddCommand(newMcpInstallCmd())
	cmd.AddCommand(newMcpUninstallCmd())
	return cmd
}

func newMcpInstallCmd() *cobra.Command {
	var (
		binaryFlag string
		dbFlag     string
		scopeFlag  string
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Claude Code CLI에 buddy MCP 서버를 등록한다 (claude mcp add)",
		RunE: func(_ *cobra.Command, _ []string) error {
			mcpBinary, err := resolveMCPBinary(binaryFlag)
			if err != nil {
				return err
			}

			args := []string{"mcp", "add", "--scope", scopeFlag}
			if dbFlag != "" {
				args = append(args, "-e", "BUDDY_DB="+dbFlag)
			}
			args = append(args, "buddy", mcpBinary)

			out, err := runClaude(args...)
			if err != nil {
				return fmt.Errorf("claude mcp add 실패: %w\n%s", err, out)
			}
			fmt.Fprint(os.Stderr, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&binaryFlag, "buddy-mcp-binary", "", "buddy-mcp 바이너리 절대 경로 (기본: 현재 실행 파일과 같은 디렉터리)")
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로. 지정하면 BUDDY_DB 환경변수로 서버에 전달됨.")
	cmd.Flags().StringVar(&scopeFlag, "scope", "user", "등록 범위: user (전역) | local (현재 프로젝트) | project (.mcp.json)")
	return cmd
}

func newMcpUninstallCmd() *cobra.Command {
	var scopeFlag string
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Claude Code CLI에서 buddy MCP 서버 등록을 제거한다 (claude mcp remove)",
		RunE: func(_ *cobra.Command, _ []string) error {
			args := []string{"mcp", "remove", "buddy"}
			if scopeFlag != "" {
				args = append(args, "--scope", scopeFlag)
			}

			out, err := runClaude(args...)
			if err != nil {
				return fmt.Errorf("claude mcp remove 실패: %w\n%s", err, out)
			}
			fmt.Fprint(os.Stderr, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&scopeFlag, "scope", "", "제거할 범위: user | local | project (기본: 존재하는 곳에서 자동 탐색)")
	return cmd
}

// resolveMCPBinary returns the absolute path to the buddy-mcp binary.
// Falls back to buddy-mcp in the same directory as the current executable.
func resolveMCPBinary(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate self: %w", err)
	}
	return filepath.Join(filepath.Dir(self), "buddy-mcp"), nil
}

// runClaude executes the `claude` CLI with the given arguments and returns
// combined stdout+stderr output. Returns an error if the command exits non-zero
// or if the claude binary cannot be found.
func runClaude(args ...string) (string, error) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude CLI를 찾을 수 없어: PATH에 claude가 있는지 확인해 (%w)", err)
	}
	cmd := exec.Command(claudePath, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
