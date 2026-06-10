package bitbucket

import (
	"context"
	"net/http"
)

// Account is the authenticated Bitbucket user.
type Account struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	UUID        string `json:"uuid"`
}

// CurrentUser returns the account for the current credentials (GET /user).
func (c *Client) CurrentUser(ctx context.Context) (*Account, error) {
	var a Account
	if err := c.do(ctx, http.MethodGet, "/user", nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
