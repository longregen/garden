package output

import (
	"context"
)

// ContentProcessor defines the interface for processing HTML content
type ContentProcessor interface {
	// ProcessWithLynx processes HTML content using lynx
	ProcessWithLynx(ctx context.Context, htmlContent string) (string, error)

	// ProcessWithReader processes HTML content using reader mode and converts to markdown
	// url parameter is optional and used for attribution in the output
	ProcessWithReader(ctx context.Context, htmlContent []byte, url string) (string, error)

	// ProcessURL fetches and processes content from a URL
	ProcessURL(ctx context.Context, url string) (string, error)
}
