package output

import (
	"context"
)

// HTTPFetcher defines the interface for fetching HTTP content
type HTTPFetcher interface {
	// Fetch retrieves content from a URL with timeout
	Fetch(ctx context.Context, url string, timeout int) (*FetchResponse, error)
}

// FetchResponse represents the response from an HTTP fetch operation
type FetchResponse struct {
	StatusCode int32
	Headers    string
	Content    []byte
}
