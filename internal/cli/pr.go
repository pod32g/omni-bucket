package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/config"
	"github.com/pod32g/omni-bucket/internal/output"
	"github.com/spf13/cobra"
)

func newPRCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "pr", Short: "Work with pull requests"}
	cmd.AddCommand(newPRListCmd())
	return cmd
}

func newPRListCmd() *cobra.Command {
	var state string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			repo, err := resolveRepo(cfg)
			if err != nil {
				return err
			}
			upperState := strings.ToUpper(state)
			validStates := map[string]bool{"OPEN": true, "MERGED": true, "DECLINED": true, "SUPERSEDED": true}
			if !validStates[upperState] {
				return fmt.Errorf("unknown state %q; must be one of: open, merged, declined, superseded", state)
			}
			client, err := newClientFn()
			if err != nil {
				return err
			}
			var prs []bitbucket.PullRequest
			for pr, err := range client.PullRequests.List(cmd.Context(), repo, bitbucket.ListOptions{
				State: upperState,
				Limit: limit,
			}) {
				if err != nil {
					return err
				}
				prs = append(prs, pr)
			}
			if flags.json {
				return output.JSON(cmd.OutOrStdout(), prs)
			}
			rows := make([][]string, 0, len(prs))
			for _, pr := range prs {
				rows = append(rows, []string{
					strconv.Itoa(pr.ID),
					pr.State,
					pr.Title,
					pr.Source.Branch.Name + " -> " + pr.Destination.Branch.Name,
					pr.Author.DisplayName,
				})
			}
			return output.Table(cmd.OutOrStdout(), []string{"ID", "STATE", "TITLE", "BRANCH", "AUTHOR"}, rows)
		},
	}
	cmd.Flags().StringVar(&state, "state", "open", "filter by state (open, merged, declined)")
	cmd.Flags().IntVar(&limit, "limit", 0, "max results (0 = all)")
	return cmd
}
