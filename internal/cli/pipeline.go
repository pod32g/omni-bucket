package cli

import (
	"strconv"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
	"github.com/pod32g/omni-bucket/internal/config"
	"github.com/pod32g/omni-bucket/internal/output"
	"github.com/spf13/cobra"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "pipeline", Short: "Work with pipelines"}
	cmd.AddCommand(newPipelineListCmd())
	return cmd
}

func newPipelineListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipeline runs",
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
			var runs []bitbucket.Pipeline
			for p, err := range client.Pipelines.List(cmd.Context(), repo, bitbucket.PipelineListOptions{Limit: limit}) {
				if err != nil {
					return err
				}
				runs = append(runs, p)
			}
			if flags.json {
				return output.JSON(cmd.OutOrStdout(), runs)
			}
			rows := make([][]string, 0, len(runs))
			for _, p := range runs {
				status := p.State.Name
				if p.State.Result.Name != "" {
					status = p.State.Result.Name
				}
				rows = append(rows, []string{strconv.Itoa(p.BuildNumber), status, p.CreatedOn})
			}
			return output.Table(cmd.OutOrStdout(), []string{"BUILD", "STATUS", "CREATED"}, rows)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "max results (0 = all)")
	return cmd
}
