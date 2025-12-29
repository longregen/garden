package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"garden3/internal/domain/entity"
)

// Service implements the output.EmbeddingsService interface
type Service struct {
	serviceURL string
	apiKey     string
	client     *http.Client
}

// NewService creates a new embeddings service
func NewService(serviceURL, apiKey string) *Service {
	return &Service{
		serviceURL: serviceURL,
		apiKey:     apiKey,
		client:     &http.Client{},
	}
}

// embeddingRequest matches the expected API request format
type embeddingRequest struct {
	Prompt    string `json:"prompt"`
	Operation string `json:"operation"`
}

// embeddingResponse represents a tuple of [text, embedding]
type embeddingResponse [][]interface{}

func (s *Service) GetEmbedding(ctx context.Context, text string) ([]entity.Embedding, error) {
	// Prepare request with default operation prefix
	reqBody := embeddingRequest{
		Prompt:    text,
		Operation: "query: ",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service error: status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - expecting array of [text, embedding[]] tuples
	var rawResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to entity.Embedding format
	embeddings := make([]entity.Embedding, 0, len(rawResp))
	for _, item := range rawResp {
		if len(item) != 2 {
			return nil, fmt.Errorf("invalid response format: expected [text, embedding] tuple")
		}

		// Extract text
		text, ok := item[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid response format: first element should be string")
		}

		// Extract embedding array
		embeddingSlice, ok := item[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response format: second element should be array")
		}

		// Convert to []float32
		embedding := make([]float32, len(embeddingSlice))
		for i, v := range embeddingSlice {
			switch val := v.(type) {
			case float64:
				embedding[i] = float32(val)
			case float32:
				embedding[i] = val
			default:
				return nil, fmt.Errorf("invalid embedding value type at index %d", i)
			}
		}

		embeddings = append(embeddings, entity.Embedding{
			Text:      text,
			Embedding: embedding,
		})
	}

	return embeddings, nil
}
