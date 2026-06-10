package bitbucket

// ParseAPIErrorForTest exposes parseAPIError to tests.
func ParseAPIErrorForTest(statusCode int, body []byte) error {
	return parseAPIError(statusCode, body)
}
