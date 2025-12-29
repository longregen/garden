package output

import (
	"context"
)

// AIService defines the interface for AI operations like summarization
type AIService interface {
	// GenerateSummary generates a summary of the given content
	GenerateSummary(ctx context.Context, content, url string, maxWords int) (string, error)
}
