package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

// apiTokenURL is where Atlassian API tokens are created and managed.
const apiTokenURL = "https://id.atlassian.com/manage-profile/security/api-tokens"

// writeTokenSetupHelp explains, before prompting, how to create a Bitbucket
// API token that actually works (the common pitfall is a token without scopes).
func writeTokenSetupHelp(w io.Writer) {
	fmt.Fprintln(w, "bb signs in with a Bitbucket API token (recommended, one-time setup).")
	fmt.Fprintln(w, "  1. Create one at: "+apiTokenURL)
	fmt.Fprintln(w, "     Use \"Create API token with scopes\" and select Bitbucket.")
	fmt.Fprintln(w, "  2. Grant scopes: Account (read) at minimum, plus Repositories /")
	fmt.Fprintln(w, "     Pull requests / Pipelines / Issues as you need them.")
	fmt.Fprintln(w, "  3. Sign in with your Atlassian account email and the token below.")
	fmt.Fprintln(w)
}

// consumerSetupURL returns the OAuth-consumer settings URL for a workspace, or a
// placeholder URL when the workspace is unknown.
func consumerSetupURL(workspace string) string {
	if workspace == "" {
		return "https://bitbucket.org/<workspace>/workspace/settings/api"
	}
	return "https://bitbucket.org/" + workspace + "/workspace/settings/api"
}

// writeConsumerSetupHelp explains the one-time OAuth consumer creation for
// browser login, and points to the simpler API-token path.
func writeConsumerSetupHelp(w io.Writer, workspace string) {
	fmt.Fprintln(w, "Browser login needs a one-time OAuth consumer (Bitbucket has no PKCE).")
	fmt.Fprintln(w, "Tip: for a simpler login, run `bb auth login` to use an API token instead.")
	fmt.Fprintln(w, "  1. Open: "+consumerSetupURL(workspace)+"  -> Add consumer")
	fmt.Fprintln(w, "  2. Callback URL:  "+oauthRedirectURI)
	fmt.Fprintln(w, "  3. Check \"This is a private consumer\".")
	fmt.Fprintln(w, "  4. Permissions: Account, Repositories, Pull requests, Pipelines, Issues.")
	fmt.Fprintln(w, "  5. Save, then paste the Key and Secret below.")
	fmt.Fprintln(w)
}

// loginError turns a credential-verification failure into actionable guidance.
// A 401 from Bitbucket almost always means the API token lacks Bitbucket scopes.
func loginError(err error) error {
	var apiErr *bitbucket.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == 401 {
		return fmt.Errorf(
			"token verification failed (401: %s)\n"+
				"This usually means the API token has no Bitbucket scopes. Recreate it with\n"+
				"\"Create API token with scopes\" (select Bitbucket, grant at least Account: read)\n"+
				"and sign in with your Atlassian account email.\n"+
				"Manage tokens: %s",
			apiErr.Message, apiTokenURL)
	}
	return fmt.Errorf("token verification failed: %w", err)
}
