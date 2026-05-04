package main

import (
	"encoding/json"
	"fmt"
	"os"
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
		claudeDirFlag string
		binaryFlag    string
		dbFlag        string
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "~/.claude/settings.json에 buddy MCP 서버를 등록한다",
		RunE: func(_ *cobra.Command, _ []string) error {
			settingsPath, err := resolveSettingsPath(claudeDirFlag)
			if err != nil {
				return err
			}
			mcpBinary, err := resolveMCPBinary(binaryFlag)
			if err != nil {
				return err
			}

			doc, err := readSettingsDoc(settingsPath)
			if err != nil {
				return err
			}

			servers := ensureMap(doc, "mcpServers")
			entry := map[string]any{
				"command": mcpBinary,
			}
			if dbFlag != "" {
				entry["env"] = map[string]any{"BUDDY_DB": dbFlag}
			}

			if _, exists := servers["buddy"]; exists {
				servers["buddy"] = entry
				if err := writeSettingsDoc(settingsPath, doc); err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, "buddy: MCP 서버 설정 업데이트했어.")
				return nil
			}

			servers["buddy"] = entry
			if err := writeSettingsDoc(settingsPath, doc); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "buddy: MCP 서버 등록했어 → %s\n", settingsPath)
			fmt.Fprintln(os.Stderr, "Claude Code를 재시작하면 mcp__buddy__* 도구가 활성화돼.")
			return nil
		},
	}
	cmd.Flags().StringVar(&claudeDirFlag, "claude-dir", "", "Claude Code 설정 디렉터리 (기본: ~/.claude)")
	cmd.Flags().StringVar(&binaryFlag, "buddy-mcp-binary", "", "buddy-mcp 바이너리 절대 경로 (기본: 현재 실행 파일과 같은 디렉터리)")
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db). 비면 BUDDY_DB 환경변수로 넘기지 않음.")
	return cmd
}

func newMcpUninstallCmd() *cobra.Command {
	var claudeDirFlag string
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "~/.claude/settings.json에서 buddy MCP 서버 등록을 제거한다",
		RunE: func(_ *cobra.Command, _ []string) error {
			settingsPath, err := resolveSettingsPath(claudeDirFlag)
			if err != nil {
				return err
			}

			doc, err := readSettingsDoc(settingsPath)
			if err != nil {
				return err
			}

			servers, ok := doc["mcpServers"].(map[string]any)
			if !ok {
				fmt.Fprintln(os.Stderr, "buddy: mcpServers 항목 없어. 이미 제거된 것 같아.")
				return nil
			}
			if _, exists := servers["buddy"]; !exists {
				fmt.Fprintln(os.Stderr, "buddy: buddy MCP 서버가 등록되어 있지 않아.")
				return nil
			}

			delete(servers, "buddy")
			if err := writeSettingsDoc(settingsPath, doc); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "buddy: MCP 서버 등록 제거했어.")
			return nil
		},
	}
	cmd.Flags().StringVar(&claudeDirFlag, "claude-dir", "", "Claude Code 설정 디렉터리 (기본: ~/.claude)")
	return cmd
}

// resolveSettingsPath returns the absolute path to settings.json.
func resolveSettingsPath(claudeDir string) (string, error) {
	if claudeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("user home dir: %w", err)
		}
		claudeDir = filepath.Join(home, ".claude")
	}
	return filepath.Join(claudeDir, "settings.json"), nil
}

// resolveMCPBinary returns the absolute path to the buddy-mcp binary.
// Falls back to looking for buddy-mcp next to the current executable.
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

func readSettingsDoc(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read settings: %w", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse settings.json: %w", err)
	}
	if doc == nil {
		doc = map[string]any{}
	}
	return doc, nil
}

func writeSettingsDoc(path string, doc map[string]any) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write tmp settings: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename settings: %w", err)
	}
	return nil
}

func ensureMap(doc map[string]any, key string) map[string]any {
	if v, ok := doc[key].(map[string]any); ok {
		return v
	}
	m := map[string]any{}
	doc[key] = m
	return m
}
