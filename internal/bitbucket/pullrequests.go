package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
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

// Get returns a single pull request by ID.
func (s *PullRequestsService) Get(ctx context.Context, repo string, id int) (*PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/pullrequests/%d", repo, id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// Approve approves a pull request.
func (s *PullRequestsService) Approve(ctx context.Context, repo string, id int) error {
	path := fmt.Sprintf("/repositories/%s/pullrequests/%d/approve", repo, id)
	return s.client.do(ctx, http.MethodPost, path, nil, nil)
}

// Merge merges a pull request and returns its updated state.
func (s *PullRequestsService) Merge(ctx context.Context, repo string, id int) (*PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/pullrequests/%d/merge", repo, id)
	if err := s.client.do(ctx, http.MethodPost, path, nil, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// CreatePullRequest is the payload for opening a pull request.
type CreatePullRequest struct {
	Title       string
	Source      string // source branch name
	Destination string // destination branch name (optional)
	Description string
}

// Create opens a new pull request.
func (s *PullRequestsService) Create(ctx context.Context, repo string, in CreatePullRequest) (*PullRequest, error) {
	payload := map[string]any{
		"title": in.Title,
		"source": map[string]any{
			"branch": map[string]string{"name": in.Source},
		},
	}
	if in.Destination != "" {
		payload["destination"] = map[string]any{
			"branch": map[string]string{"name": in.Destination},
		}
	}
	if in.Description != "" {
		payload["description"] = in.Description
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/pullrequests", repo)
	if err := s.client.do(ctx, http.MethodPost, path, bytes.NewReader(body), &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}
