package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pod32g/omni-bucket/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage authentication"}
	cmd.AddCommand(newAuthLoginCmd(), newAuthStatusCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var email, token string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a Bitbucket API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if email == "" {
				reader := bufio.NewReader(cmd.InOrStdin())
				fmt.Fprint(out, "Bitbucket email: ")
				line, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				email = strings.TrimSpace(line)
			}
			if token == "" {
				fmt.Fprint(out, "API token (input hidden): ")
				tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return err
				}
				fmt.Fprintln(out)
				token = strings.TrimSpace(string(tokenBytes))
			}
			if email == "" {
				return errors.New("email cannot be empty")
			}
			if token == "" {
				return errors.New("token cannot be empty")
			}

			client := newCredClientFn(email, token)
			acct, err := client.CurrentUser(cmd.Context())
			if err != nil {
				return fmt.Errorf("token verification failed: %w", err)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.Email = email
			cfg.Token = token
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(out, "Logged in as %s (%s)\n", acct.DisplayName, acct.Username)
			return nil
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "Bitbucket account email (skips the prompt)")
	cmd.Flags().StringVar(&token, "token", "", "Bitbucket API token (skips the hidden prompt)")
	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the authenticated account",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClientFn()
			if err != nil {
				return err
			}
			acct, err := client.CurrentUser(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s (%s)\n", acct.DisplayName, acct.Username)
			return nil
		},
	}
}
