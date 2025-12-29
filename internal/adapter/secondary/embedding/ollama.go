package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

// OllamaEmbeddingService implements the EmbeddingService interface using Ollama API
type OllamaEmbeddingService struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaEmbeddingService creates a new Ollama embedding service for single embeddings
func NewOllamaEmbeddingService(baseURL, model string) output.EmbeddingService {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text:latest"
	}

	return &OllamaEmbeddingService{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

// ollamaEmbedRequest represents the request payload for Ollama embeddings API
type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbedResponse represents the response from Ollama embeddings API
type ollamaEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// GetEmbedding generates a single embedding vector for the given text
func (s *OllamaEmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Prepare request
	reqBody := ollamaEmbedRequest{
		Model:  s.model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/embeddings", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(ollamaResp.Embedding))
	for i, v := range ollamaResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// OllamaEmbeddingsService implements the EmbeddingsService interface with text chunking
type OllamaEmbeddingsService struct {
	baseURL   string
	model     string
	client    *http.Client
	chunkSize int
}

// NewOllamaEmbeddingsService creates a new Ollama embeddings service for chunked embeddings
func NewOllamaEmbeddingsService(baseURL, model string) output.EmbeddingsService {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text:latest"
	}

	return &OllamaEmbeddingsService{
		baseURL:   baseURL,
		model:     model,
		client:    &http.Client{},
		chunkSize: 8000, // Default chunk size in characters
	}
}

// GetEmbedding generates chunked embeddings for the given text
func (s *OllamaEmbeddingsService) GetEmbedding(ctx context.Context, text string) ([]entity.Embedding, error) {
	// Chunk the text
	chunks := s.chunkText(text)

	// Generate embeddings for each chunk
	embeddings := make([]entity.Embedding, 0, len(chunks))
	for _, chunk := range chunks {
		embedding, err := s.getEmbeddingForChunk(ctx, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to get embedding for chunk: %w", err)
		}

		embeddings = append(embeddings, entity.Embedding{
			Text:      chunk,
			Embedding: embedding,
		})
	}

	return embeddings, nil
}

// getEmbeddingForChunk generates embedding for a single chunk
func (s *OllamaEmbeddingsService) getEmbeddingForChunk(ctx context.Context, text string) ([]float32, error) {
	// Prepare request
	reqBody := ollamaEmbedRequest{
		Model:  s.model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/embeddings", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(ollamaResp.Embedding))
	for i, v := range ollamaResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// chunkText splits text into chunks of approximately chunkSize characters
// Tries to split on sentence boundaries when possible
func (s *OllamaEmbeddingsService) chunkText(text string) []string {
	if len(text) <= s.chunkSize {
		return []string{text}
	}

	chunks := make([]string, 0)
	currentChunk := ""

	// Split by sentences (simple approach using periods, exclamation marks, question marks)
	sentences := splitSentences(text)

	for _, sentence := range sentences {
		// If adding this sentence would exceed chunk size, save current chunk and start new one
		if len(currentChunk)+len(sentence) > s.chunkSize && len(currentChunk) > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = sentence
		} else {
			if len(currentChunk) > 0 {
				currentChunk += " "
			}
			currentChunk += sentence
		}
	}

	// Add the last chunk if not empty
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}

// splitSentences splits text into sentences
func splitSentences(text string) []string {
	sentences := make([]string, 0)
	currentSentence := ""

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		currentSentence += string(runes[i])

		// Check for sentence endings
		if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			// Look ahead to see if this is really end of sentence
			// (not an abbreviation or decimal)
			if i+1 < len(runes) && (runes[i+1] == ' ' || runes[i+1] == '\n' || runes[i+1] == '\r') {
				sentences = append(sentences, strings.TrimSpace(currentSentence))
				currentSentence = ""
			}
		} else if runes[i] == '\n' && len(currentSentence) > 1 {
			// Also split on newlines for paragraph boundaries
			sentences = append(sentences, strings.TrimSpace(currentSentence))
			currentSentence = ""
		}
	}

	// Add any remaining text as the last sentence
	if len(strings.TrimSpace(currentSentence)) > 0 {
		sentences = append(sentences, strings.TrimSpace(currentSentence))
	}

	return sentences
}
