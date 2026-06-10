package bitbucket

import (
	"context"
	"net/http"
)

// Authorizer attaches credentials to an outgoing request.
type Authorizer interface {
	Authorize(ctx context.Context, req *http.Request) error
}

// BasicAuth authorizes with an email + API token (Bitbucket Basic auth).
type BasicAuth struct {
	Email string
	Token string
}

// Authorize sets HTTP Basic auth credentials.
func (b *BasicAuth) Authorize(ctx context.Context, req *http.Request) error {
	req.SetBasicAuth(b.Email, b.Token)
	return nil
}
