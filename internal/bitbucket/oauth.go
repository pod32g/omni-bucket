package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// tokenEndpoint is the Bitbucket Cloud OAuth 2.0 token endpoint. It is a var so
// tests can point it at an httptest server.
var tokenEndpoint = "https://bitbucket.org/site/oauth2/access_token"

// AuthorizeEndpoint is the Bitbucket Cloud OAuth 2.0 authorization endpoint.
const AuthorizeEndpoint = "https://bitbucket.org/site/oauth2/authorize"

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// oauthErrorResponse is the RFC 6749 token-endpoint error shape.
type oauthErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

// OAuthToken is the result of a successful code exchange or refresh.
type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// postToken performs an OAuth token-endpoint POST with Basic-auth client creds.
func postToken(ctx context.Context, httpClient *http.Client, clientID, clientSecret string, form url.Values) (*tokenResponse, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		var oe oauthErrorResponse
		if json.Unmarshal(data, &oe) == nil && oe.Code != "" {
			return nil, &APIError{StatusCode: resp.StatusCode, Message: oe.Code, Detail: oe.Description}
		}
		return nil, parseAPIError(resp.StatusCode, data)
	}
	var tr tokenResponse
	if err := json.Unmarshal(data, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// ExchangeCode swaps an authorization code for tokens. now may be nil.
func ExchangeCode(ctx context.Context, httpClient *http.Client, clientID, clientSecret, code, redirectURI string, now func() time.Time) (*OAuthToken, error) {
	if now == nil {
		now = time.Now
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	if redirectURI != "" {
		form.Set("redirect_uri", redirectURI)
	}
	tr, err := postToken(ctx, httpClient, clientID, clientSecret, form)
	if err != nil {
		return nil, err
	}
	return &OAuthToken{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Expiry:       expiryFromExpiresIn(now, tr.ExpiresIn),
	}, nil
}

// expiryFromExpiresIn computes an absolute expiry, falling back to one hour when
// the server omits or zeroes expires_in (prevents a refresh-on-every-request loop).
func expiryFromExpiresIn(now func() time.Time, expiresIn int) time.Time {
	if expiresIn > 0 {
		return now().Add(time.Duration(expiresIn) * time.Second)
	}
	return now().Add(time.Hour)
}

const expiryBuffer = 60 * time.Second

// OAuthTokenSource is an Authorizer that attaches a Bearer access token,
// refreshing it via the refresh token when expired or about to expire.
// It is not safe for concurrent use.
type OAuthTokenSource struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	HTTPClient   *http.Client
	OnRefresh    func(accessToken, refreshToken string, expiry time.Time)
	now          func() time.Time
}

// NewOAuthTokenSource builds a token source. httpClient/now may be nil.
func NewOAuthTokenSource(clientID, clientSecret, accessToken, refreshToken string, expiry time.Time, httpClient *http.Client, onRefresh func(string, string, time.Time)) *OAuthTokenSource {
	return &OAuthTokenSource{
		ClientID: clientID, ClientSecret: clientSecret,
		AccessToken: accessToken, RefreshToken: refreshToken,
		Expiry: expiry, HTTPClient: httpClient, OnRefresh: onRefresh,
		now: time.Now,
	}
}

// Authorize ensures a fresh access token and sets the Bearer header.
func (s *OAuthTokenSource) Authorize(ctx context.Context, req *http.Request) error {
	now := s.now
	if now == nil {
		now = time.Now
	}
	if s.AccessToken == "" || !now().Before(s.Expiry.Add(-expiryBuffer)) {
		if err := s.refresh(ctx); err != nil {
			return err
		}
	}
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	return nil
}

func (s *OAuthTokenSource) refresh(ctx context.Context) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", s.RefreshToken)
	tr, err := postToken(ctx, s.HTTPClient, s.ClientID, s.ClientSecret, form)
	if err != nil {
		return fmt.Errorf("oauth token refresh failed; please re-authenticate: %w", err)
	}
	now := s.now
	if now == nil {
		now = time.Now
	}
	s.AccessToken = tr.AccessToken
	if tr.RefreshToken != "" {
		s.RefreshToken = tr.RefreshToken
	}
	s.Expiry = expiryFromExpiresIn(now, tr.ExpiresIn)
	if s.OnRefresh != nil {
		s.OnRefresh(s.AccessToken, s.RefreshToken, s.Expiry)
	}
	return nil
}
