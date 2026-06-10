package bitbucket

import (
	"context"
	"iter"
	"net/http"
)

// pages yields every item across all pages, following the opaque `next` URL
// until exhausted. On error it yields the zero value of T together with the
// error and then stops, so callers MUST check the error before using the value.
func pages[T any](ctx context.Context, c *Client, firstPathOrURL string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		next := firstPathOrURL
		for next != "" {
			var page Paginated[T]
			if err := c.do(ctx, http.MethodGet, next, nil, &page); err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, v := range page.Values {
				if !yield(v, nil) {
					return
				}
			}
			next = page.Next
		}
	}
}

// limit wraps a sequence to stop after n successful items.
func limit[T any](seq iter.Seq2[T, error], n int) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		count := 0
		for v, err := range seq {
			if !yield(v, err) {
				return
			}
			if err != nil {
				return
			}
			count++
			if count >= n {
				return
			}
		}
	}
}
