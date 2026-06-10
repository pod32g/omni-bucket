package bitbucket

import (
	"context"
	"fmt"
	"iter"
	"net/url"
)

// RepositoriesService groups repository endpoints.
type RepositoriesService struct {
	client *Client
}

// Repository is a Bitbucket repository (subset of fields).
type Repository struct {
	FullName    string `json:"full_name"`
	Name        string `json:"name"`
	IsPrivate   bool   `json:"is_private"`
	Description string `json:"description"`
	UpdatedOn   string `json:"updated_on"`
	MainBranch  struct {
		Name string `json:"name"`
	} `json:"mainbranch"`
}

// RepoListOptions controls repository listing.
type RepoListOptions struct {
	Limit int // max items; 0 = all
}

// List yields repositories in a workspace.
func (s *RepositoriesService) List(ctx context.Context, workspace string, opts RepoListOptions) iter.Seq2[Repository, error] {
	q := url.Values{}
	q.Set("pagelen", "100")
	path := fmt.Sprintf("/repositories/%s?%s", workspace, q.Encode())
	seq := pages[Repository](ctx, s.client, path)
	if opts.Limit > 0 {
		seq = limit(seq, opts.Limit)
	}
	return seq
}
