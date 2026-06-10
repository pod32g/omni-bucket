package bitbucket

import (
	"context"
	"fmt"
	"iter"
	"net/url"
)

// PipelinesService groups pipeline endpoints.
type PipelinesService struct {
	client *Client
}

// Pipeline is a Bitbucket Pipelines run (subset of fields).
type Pipeline struct {
	UUID        string `json:"uuid"`
	BuildNumber int    `json:"build_number"`
	State       struct {
		Name   string `json:"name"`
		Result struct {
			Name string `json:"name"`
		} `json:"result"`
	} `json:"state"`
	CreatedOn string `json:"created_on"`
}

// PipelineListOptions controls pipeline listing.
type PipelineListOptions struct {
	Limit int // max items; 0 = all
}

// List yields pipeline runs for "workspace/repo", newest first.
func (s *PipelinesService) List(ctx context.Context, repo string, opts PipelineListOptions) iter.Seq2[Pipeline, error] {
	q := url.Values{}
	q.Set("pagelen", "100")
	q.Set("sort", "-created_on")
	path := fmt.Sprintf("/repositories/%s/pipelines/?%s", repo, q.Encode())
	seq := pages[Pipeline](ctx, s.client, path)
	if opts.Limit > 0 {
		seq = limit(seq, opts.Limit)
	}
	return seq
}
