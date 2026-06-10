package bitbucket

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// DefaultBaseURL is the Bitbucket Cloud REST API 2.0 root.
const DefaultBaseURL = "https://api.bitbucket.org/2.0"

// Client talks to the Bitbucket Cloud REST API using API-token Basic auth.
type Client struct {
	BaseURL    string
	Email      string
	Token      string
	HTTPClient *http.Client

	PullRequests *PullRequestsService
}

// NewClient builds a client authenticated with an email + scoped API token.
func NewClient(email, token string) *Client {
	c := &Client{
		BaseURL:    DefaultBaseURL,
		Email:      email,
		Token:      token,
		HTTPClient: &http.Client{},
	}
	c.PullRequests = &PullRequestsService{client: c}
	return c
}

// do executes a request. `pathOrURL` is either an API path beginning with "/"
// (joined to BaseURL) or a full URL (used when following a pagination `next`
// link). When out is non-nil, a successful JSON body is decoded into it.
func (c *Client) do(ctx context.Context, method, pathOrURL string, body io.Reader, out any) error {
	target := pathOrURL
	if !strings.HasPrefix(pathOrURL, "http") {
		target = c.BaseURL + pathOrURL
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Email, c.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, data)
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}
