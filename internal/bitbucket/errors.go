package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError is a non-2xx response from the Bitbucket API.
type APIError struct {
	StatusCode int
	Message    string
	Detail     string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("bitbucket: %d: %s (%s)", e.StatusCode, e.Message, e.Detail)
	}
	return fmt.Sprintf("bitbucket: %d: %s", e.StatusCode, e.Message)
}

type errorEnvelope struct {
	Type  string `json:"type"`
	Error struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
	} `json:"error"`
}

// parseAPIError builds an APIError from a status code and response body.
func parseAPIError(statusCode int, body []byte) *APIError {
	e := &APIError{StatusCode: statusCode}
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error.Message != "" {
		e.Message = env.Error.Message
		e.Detail = env.Error.Detail
		return e
	}
	e.Message = strings.TrimSpace(string(body))
	if e.Message == "" {
		e.Message = http.StatusText(statusCode)
	}
	return e
}
