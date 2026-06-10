package cli

import (
	"errors"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/config"
	"github.com/pod32g/omni-bucket/internal/output"
	"github.com/spf13/cobra"
)

func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "repo", Short: "Work with repositories"}
	cmd.AddCommand(newRepoListCmd())
	return cmd
}

func newRepoListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories in a workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			_, _, ws := cfg.Resolved()
			if flags.workspace != "" {
				ws = flags.workspace
			}
			if ws == "" {
				return errors.New("no workspace; use --workspace or set a default workspace")
			}
			client, err := newClientFn()
			if err != nil {
				return err
			}
			var repos []bitbucket.Repository
			for repo, err := range client.Repositories.List(cmd.Context(), ws, bitbucket.RepoListOptions{Limit: limit}) {
				if err != nil {
					return err
				}
				repos = append(repos, repo)
			}
			if flags.json {
				return output.JSON(cmd.OutOrStdout(), repos)
			}
			rows := make([][]string, 0, len(repos))
			for _, repo := range repos {
				visibility := "public"
				if repo.IsPrivate {
					visibility = "private"
				}
				rows = append(rows, []string{repo.FullName, visibility, repo.MainBranch.Name})
			}
			return output.Table(cmd.OutOrStdout(), []string{"REPO", "VISIBILITY", "MAIN"}, rows)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "max results (0 = all)")
	return cmd
}
