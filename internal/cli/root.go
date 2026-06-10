package cli

import "github.com/spf13/cobra"

type globalFlags struct {
	json      bool
	repo      string
	workspace string
}

var flags globalFlags

// NewRootCmd builds the full command tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "bb",
		Short:         "A command-line client for Bitbucket Cloud",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVar(&flags.json, "json", false, "output JSON")
	root.PersistentFlags().StringVar(&flags.repo, "repo", "", "repository as workspace/repo")
	root.PersistentFlags().StringVar(&flags.workspace, "workspace", "", "default workspace")
	return root
}
