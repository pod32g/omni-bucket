package cli

import (
	"strconv"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/config"
	"github.com/pod32g/omni-bucket/internal/output"
	"github.com/spf13/cobra"
)

func newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "issue", Short: "Work with issues"}
	cmd.AddCommand(newIssueListCmd())
	return cmd
}

func newIssueListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			repo, err := resolveRepo(cfg)
			if err != nil {
				return err
			}
			client, err := newClientFn()
			if err != nil {
				return err
			}
			var issues []bitbucket.Issue
			for is, err := range client.Issues.List(cmd.Context(), repo, bitbucket.IssueListOptions{Limit: limit}) {
				if err != nil {
					return err
				}
				issues = append(issues, is)
			}
			if flags.json {
				return output.JSON(cmd.OutOrStdout(), issues)
			}
			rows := make([][]string, 0, len(issues))
			for _, is := range issues {
				rows = append(rows, []string{strconv.Itoa(is.ID), is.State, is.Kind, is.Title})
			}
			return output.Table(cmd.OutOrStdout(), []string{"ID", "STATE", "KIND", "TITLE"}, rows)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "max results (0 = all)")
	return cmd
}
