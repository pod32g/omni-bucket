package bitbucket

// Paginated wraps Bitbucket Cloud's list responses. `Next` is an opaque URL
// that must be followed verbatim, never reconstructed by the client.
type Paginated[T any] struct {
	Size     int    `json:"size"`
	Page     int    `json:"page"`
	PageLen  int    `json:"pagelen"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Values   []T    `json:"values"`
}
