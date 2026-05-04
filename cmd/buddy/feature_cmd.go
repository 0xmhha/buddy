package main

// feature_cmd.go owns the `buddy feature` subcommand group.
// Subcommands: list, get, upsert, delete, search.
// All read-only subcommands open the DB read-only; upsert/delete open writable.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/wm-it-22-00661/buddy/internal/feature"
	"github.com/wm-it-22-00661/buddy/internal/persona"
)

func newFeatureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feature",
		Short: "로컬 feature registry 조회 및 관리",
	}
	cmd.AddCommand(
		newFeatureListCmd(),
		newFeatureGetCmd(),
		newFeatureUpsertCmd(),
		newFeatureDeleteCmd(),
		newFeatureSearchCmd(),
	)
	return cmd
}

// newFeatureListCmd — buddy feature list [--status <s>] [--db <path>]
func newFeatureListCmd() *cobra.Command {
	var dbFlag, statusFlag string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "feature 목록 (--status 로 필터)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			features, err := feature.List(cmd.Context(), feature.Options{DBPath: dbFlag}, statusFlag)
			if err != nil {
				return featureError(err)
			}
			if len(features) == 0 {
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyFeatureListEmpty))
				return nil
			}
			renderFeatureList(features)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&statusFlag, "status", "", "draft | in_progress | done | cancelled")
	return cmd
}

// newFeatureGetCmd — buddy feature get <feature-id> [--db <path>]
func newFeatureGetCmd() *cobra.Command {
	var dbFlag string
	cmd := &cobra.Command{
		Use:   "get <feature-id>",
		Short: "feature 상세 조회",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := feature.Get(cmd.Context(), feature.Options{DBPath: dbFlag}, args[0])
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return newFriendError(persona.M(persona.KeyFeatureNotFound, args[0]))
				}
				return featureError(err)
			}
			renderFeatureDetail(f)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	return cmd
}

// newFeatureUpsertCmd — buddy feature upsert --id <id> --name <name> [--status <s>] [--summary <text>]
func newFeatureUpsertCmd() *cobra.Command {
	var dbFlag, idFlag, nameFlag, statusFlag, summaryFlag string
	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "feature 저장 (create / update)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if idFlag == "" || nameFlag == "" {
				return newFriendError("buddy: --id 와 --name 이 필요해.")
			}
			if statusFlag == "" {
				statusFlag = feature.StatusDraft
			}
			f := feature.Feature{
				FeatureID:          idFlag,
				Name:               nameFlag,
				Summary:            summaryFlag,
				Status:             statusFlag,
				Actors:             []feature.Actor{},
				AcceptanceCriteria: []string{},
				TestPlan:           feature.TestPlan{},
			}
			if err := feature.Upsert(cmd.Context(), feature.Options{DBPath: dbFlag}, f); err != nil {
				return featureError(err)
			}
			fmt.Fprintln(os.Stderr, persona.M(persona.KeyFeatureUpserted, idFlag))
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	cmd.Flags().StringVar(&idFlag, "id", "", "feature ID (필수)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "feature 이름 (필수)")
	cmd.Flags().StringVar(&statusFlag, "status", "", "draft | in_progress | done | cancelled (기본: draft)")
	cmd.Flags().StringVar(&summaryFlag, "summary", "", "feature 요약 설명")
	return cmd
}

// newFeatureDeleteCmd — buddy feature delete <feature-id> [--db <path>]
func newFeatureDeleteCmd() *cobra.Command {
	var dbFlag string
	cmd := &cobra.Command{
		Use:   "delete <feature-id>",
		Short: "feature 삭제",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := feature.Delete(cmd.Context(), feature.Options{DBPath: dbFlag}, args[0]); err != nil {
				return featureError(err)
			}
			fmt.Fprintln(os.Stderr, persona.M(persona.KeyFeatureDeleted, args[0]))
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	return cmd
}

// newFeatureSearchCmd — buddy feature search <query> [--db <path>]
func newFeatureSearchCmd() *cobra.Command {
	var dbFlag string
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "name/summary 에서 검색 (대소문자 무시)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			features, err := feature.Search(cmd.Context(), feature.Options{DBPath: dbFlag}, args[0])
			if err != nil {
				return featureError(err)
			}
			if len(features) == 0 {
				fmt.Fprintln(os.Stderr, persona.M(persona.KeyFeatureListEmpty))
				return nil
			}
			renderFeatureList(features)
			return nil
		},
	}
	cmd.Flags().StringVar(&dbFlag, "db", "", "buddy DB 경로 (기본: ~/.buddy/buddy.db)")
	return cmd
}

// featureError translates common store errors to friend-tone messages.
func featureError(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return newFriendError(persona.M(persona.KeyFeatureFailed, err))
}

// renderFeatureList prints a compact table: id  status  updated_at  name.
func renderFeatureList(features []feature.Feature) {
	for _, f := range features {
		ts := time.UnixMilli(f.UpdatedAt).UTC().Format("2006-01-02")
		fmt.Printf("%-30s  %-12s  %s  %s\n", f.FeatureID, f.Status, ts, f.Name)
	}
}

// renderFeatureDetail prints a multi-line block for a single feature.
func renderFeatureDetail(f feature.Feature) {
	ts := time.UnixMilli(f.UpdatedAt).UTC().Format(time.RFC3339)
	fmt.Printf("feature_id : %s\n", f.FeatureID)
	fmt.Printf("name       : %s\n", f.Name)
	fmt.Printf("status     : %s\n", f.Status)
	fmt.Printf("updated_at : %s\n", ts)
	if f.Summary != "" {
		fmt.Printf("summary    : %s\n", f.Summary)
	}
	if len(f.Actors) > 0 {
		fmt.Printf("actors     : ")
		for i, a := range f.Actors {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s(%s)", a.ID, a.SystemBoundary)
		}
		fmt.Println()
	}
	if len(f.AcceptanceCriteria) > 0 {
		fmt.Printf("acceptance :\n")
		for _, c := range f.AcceptanceCriteria {
			fmt.Printf("  - %s\n", c)
		}
	}
}
