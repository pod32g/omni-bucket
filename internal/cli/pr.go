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
	cmd.AddCommand(newPRListCmd(), newPRViewCmd(), newPRApproveCmd(), newPRMergeCmd(), newPRCreateCmd())
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

func prRepoAndID(args []string) (string, int, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", 0, err
	}
	repo, err := resolveRepo(cfg)
	if err != nil {
		return "", 0, err
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", 0, fmt.Errorf("invalid pull request id %q", args[0])
	}
	if id <= 0 {
		return "", 0, fmt.Errorf("invalid pull request id %q: must be a positive integer", args[0])
	}
	return repo, id, nil
}

func newPRViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, id, err := prRepoAndID(args)
			if err != nil {
				return err
			}
			client, err := newClientFn()
			if err != nil {
				return err
			}
			pr, err := client.PullRequests.Get(cmd.Context(), repo, id)
			if err != nil {
				return err
			}
			if flags.json {
				return output.JSON(cmd.OutOrStdout(), pr)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "#%d %s [%s]\n%s -> %s\nAuthor: %s\n",
				pr.ID, pr.Title, pr.State,
				pr.Source.Branch.Name, pr.Destination.Branch.Name,
				pr.Author.DisplayName)
			return nil
		},
	}
}

func newPRApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, id, err := prRepoAndID(args)
			if err != nil {
				return err
			}
			client, err := newClientFn()
			if err != nil {
				return err
			}
			if err := client.PullRequests.Approve(cmd.Context(), repo, id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Approved pull request #%d\n", id)
			return nil
		},
	}
}

func newPRMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <id>",
		Short: "Merge a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, id, err := prRepoAndID(args)
			if err != nil {
				return err
			}
			client, err := newClientFn()
			if err != nil {
				return err
			}
			pr, err := client.PullRequests.Merge(cmd.Context(), repo, id)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Merged pull request #%d (state: %s)\n", pr.ID, pr.State)
			return nil
		},
	}
}

func newPRCreateCmd() *cobra.Command {
	var title, source, destination, description string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
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
			pr, err := client.PullRequests.Create(cmd.Context(), repo, bitbucket.CreatePullRequest{
				Title:       title,
				Source:      source,
				Destination: destination,
				Description: description,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created pull request #%d\n", pr.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "pull request title (required)")
	cmd.Flags().StringVar(&source, "source", "", "source branch (required)")
	cmd.Flags().StringVar(&destination, "destination", "", "destination branch")
	cmd.Flags().StringVar(&description, "description", "", "pull request description")
	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("source")
	return cmd
}
