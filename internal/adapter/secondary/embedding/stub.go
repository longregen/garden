package embedding

import (
	"context"
	"fmt"

	"garden3/internal/domain/entity"
)

// StubEmbeddingService is a placeholder implementation that returns empty embeddings
// In production, this should call an actual embedding API (e.g., OpenAI, Cohere, etc.)
// It implements both output.EmbeddingService and output.EmbeddingsService
type StubEmbeddingService struct{}

// NewStubEmbeddingService creates a new stub embedding service
func NewStubEmbeddingService() *StubEmbeddingService {
	return &StubEmbeddingService{}
}

// GetEmbeddingVector implements output.EmbeddingService
// Returns a single embedding vector for the given text
func (s *StubEmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Return error to signal that embedding search should fall back to text search
	// In production, this would call an external embedding API
	return nil, fmt.Errorf("embedding service not implemented")
}

// StubEmbeddingsService wraps StubEmbeddingService to provide chunked embeddings
type StubEmbeddingsService struct {
	base *StubEmbeddingService
}

// NewStubEmbeddingsService creates a new stub embeddings service
func NewStubEmbeddingsService() *StubEmbeddingsService {
	return &StubEmbeddingsService{
		base: NewStubEmbeddingService(),
	}
}

// GetEmbedding implements output.EmbeddingsService
// Returns chunked embeddings for the given text
func (s *StubEmbeddingsService) GetEmbedding(ctx context.Context, text string) ([]entity.Embedding, error) {
	// Return error to signal that embedding search should fall back to text search
	// In production, this would chunk text and call an external embedding API
	return nil, fmt.Errorf("embeddings service not implemented")
}
