package bitbucket

import (
	"context"
	"fmt"
	"iter"
	"net/url"
)

// PullRequestsService groups pull request endpoints.
type PullRequestsService struct {
	client *Client
}

// PullRequest is a Bitbucket pull request (subset of fields used by the CLI).
type PullRequest struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Author struct {
		DisplayName string `json:"display_name"`
	} `json:"author"`
	Source struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"destination"`
}

// ListOptions controls pull request listing.
type ListOptions struct {
	State string // OPEN, MERGED, DECLINED, SUPERSEDED; empty = API default
	Limit int    // max items; 0 = all
}

// List yields pull requests for "workspace/repo".
func (s *PullRequestsService) List(ctx context.Context, repo string, opts ListOptions) iter.Seq2[PullRequest, error] {
	q := url.Values{}
	q.Set("pagelen", "100")
	if opts.State != "" {
		q.Set("state", opts.State)
	}
	path := fmt.Sprintf("/repositories/%s/pullrequests?%s", repo, q.Encode())
	seq := pages[PullRequest](ctx, s.client, path)
	if opts.Limit > 0 {
		seq = limit(seq, opts.Limit)
	}
	return seq
}
