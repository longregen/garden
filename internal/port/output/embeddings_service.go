package output

import (
	"context"

	"garden3/internal/domain/entity"
)

// EmbeddingsService defines the interface for generating text embeddings
type EmbeddingsService interface {
	// GetEmbedding generates embeddings for the given text
	// Returns chunked embeddings if text is too large
	GetEmbedding(ctx context.Context, text string) ([]entity.Embedding, error)
}
