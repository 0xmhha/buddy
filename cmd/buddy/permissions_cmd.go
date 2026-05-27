package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xmhha/buddy/internal/permissions"
	"github.com/0xmhha/buddy/internal/persona"
)

func newPermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permissions",
		Short: "subagent 권한 관리",
	}
	cmd.AddCommand(newPermissionsCheckCmd())
	cmd.AddCommand(newPermissionsInjectCmd())
	return cmd
}

func newPermissionsCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "subagent에 필요한 Edit 권한 확인",
		RunE: func(_ *cobra.Command, _ []string) error {
			result, err := permissions.Check()
			if err != nil {
				return err
			}
			if result.Healthy() {
				fmt.Println(persona.M(persona.KeyPermissionsAllPresent))
				return nil
			}
			fmt.Print(persona.M(persona.KeyPermissionsMissing, len(result.Missing)))
			fmt.Print(permissions.FormatMissing(result.Missing))
			fmt.Println()
			fmt.Println(persona.M(persona.KeyPermissionsInjectHint))
			return nil
		},
	}
}

func newPermissionsInjectCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "inject",
		Short: "subagent에 필요한 Edit 권한을 settings.local.json에 추가",
		RunE: func(_ *cobra.Command, _ []string) error {
			result, err := permissions.Check()
			if err != nil {
				return err
			}
			if result.Healthy() {
				fmt.Println(persona.M(persona.KeyPermissionsAllPresent))
				return nil
			}

			fmt.Print(persona.M(persona.KeyPermissionsInjectPreview, result.Path))
			fmt.Print(permissions.FormatMissing(result.Missing))
			fmt.Println()
			fmt.Println(persona.M(persona.KeyPermissionsInjectScope))

			if !force {
				fmt.Print(persona.M(persona.KeyPermissionsConfirm))
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println(persona.M(persona.KeyPermissionsCancelled))
					return nil
				}
			}

			if err := permissions.Inject(result.Missing); err != nil {
				return fmt.Errorf("inject: %w", err)
			}
			fmt.Print(persona.M(persona.KeyPermissionsInjected, len(result.Missing)))
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "확인 없이 바로 적용")
	return cmd
}
