package bitbucket

import (
	"context"
	"fmt"
	"iter"
	"net/url"
)

// IssuesService groups issue tracker endpoints.
type IssuesService struct {
	client *Client
}

// Issue is a Bitbucket issue (subset of fields).
type Issue struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	State    string `json:"state"`
	Kind     string `json:"kind"`
	Priority string `json:"priority"`
	Reporter struct {
		DisplayName string `json:"display_name"`
	} `json:"reporter"`
}

// IssueListOptions controls issue listing.
type IssueListOptions struct {
	Limit int // max items; 0 = all
}

// List yields issues for "workspace/repo".
func (s *IssuesService) List(ctx context.Context, repo string, opts IssueListOptions) iter.Seq2[Issue, error] {
	q := url.Values{}
	q.Set("pagelen", "100")
	path := fmt.Sprintf("/repositories/%s/issues?%s", repo, q.Encode())
	seq := pages[Issue](ctx, s.client, path)
	if opts.Limit > 0 {
		seq = limit(seq, opts.Limit)
	}
	return seq
}
