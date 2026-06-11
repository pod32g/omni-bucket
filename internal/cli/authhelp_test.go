package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/pod32g/omni-bucket/internal/bitbucket"
)

func TestLoginError401MentionsScopes(t *testing.T) {
	err := loginError(&bitbucket.APIError{
		StatusCode: 401,
		Message:    "Token is invalid, expired, or not supported for this endpoint.",
	})
	got := err.Error()
	if !strings.Contains(got, "scopes") || !strings.Contains(got, "Account") {
		t.Fatalf("401 error should guide on scopes, got %q", got)
	}
}

func TestLoginErrorNon401IsGeneric(t *testing.T) {
	err := loginError(&bitbucket.APIError{StatusCode: 500, Message: "boom"})
	got := err.Error()
	if !strings.Contains(got, "token verification failed") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "scopes") {
		t.Fatalf("non-401 should not show scope guidance, got %q", got)
	}
}

func TestLoginErrorWrapsPlainError(t *testing.T) {
	err := loginError(errors.New("network down"))
	if !strings.Contains(err.Error(), "network down") {
		t.Fatalf("got %q", err.Error())
	}
}

func TestWriteTokenSetupHelp(t *testing.T) {
	var b strings.Builder
	writeTokenSetupHelp(&b)
	s := b.String()
	for _, want := range []string{apiTokenURL, "with scopes", "Account"} {
		if !strings.Contains(s, want) {
			t.Fatalf("token help missing %q in:\n%s", want, s)
		}
	}
}

func TestWriteConsumerSetupHelp(t *testing.T) {
	var b strings.Builder
	writeConsumerSetupHelp(&b, "myws")
	s := b.String()
	for _, want := range []string{"myws", oauthRedirectURI, "private consumer"} {
		if !strings.Contains(s, want) {
			t.Fatalf("consumer help missing %q in:\n%s", want, s)
		}
	}
}

func TestConsumerSetupURLPlaceholderWhenNoWorkspace(t *testing.T) {
	if got := consumerSetupURL(""); !strings.Contains(got, "<workspace>") {
		t.Fatalf("got %q", got)
	}
	if got := consumerSetupURL("acme"); !strings.Contains(got, "/acme/") {
		t.Fatalf("got %q", got)
	}
}
