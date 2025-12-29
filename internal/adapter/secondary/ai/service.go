package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Service implements the output.AIService interface
type Service struct {
	serviceURL string
	apiKey     string
	client     *http.Client
}

// NewService creates a new AI service
func NewService(serviceURL, apiKey string) *Service {
	if serviceURL == "" {
		serviceURL = "http://localhost:11434"
	}
	return &Service{
		serviceURL: serviceURL,
		apiKey:     apiKey,
		client:     &http.Client{},
	}
}

// ollamaGenerateRequest represents the request payload for Ollama generate API
type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaGenerateResponse represents the response from Ollama generate API
type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (s *Service) GenerateSummary(ctx context.Context, content, url string, maxWords int) (string, error) {
	if maxWords == 0 {
		maxWords = 400
	}

	// Build the prompt using the same format as the base project
	prompt := fmt.Sprintf("I have read the following article of url %s:\n\n\n===\n%s\n\n===\nNow, what would be your summary of this article? Please use less than %d words", url, content, maxWords)

	// Prepare request
	reqBody := ollamaGenerateRequest{
		Model:  "current-default:latest",
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/api/generate", s.serviceURL)
	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI service error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Strip <think> tags from the response
	summary := stripThinkTags(ollamaResp.Response)

	return summary, nil
}

// stripThinkTags removes <think>...</think> tags from the text
func stripThinkTags(text string) string {
	// Use regex to remove <think>...</think> blocks (including content inside)
	// This matches the TypeScript regex: .replace(/<think>[\s\S]*?<\/think>/, "")
	re := regexp.MustCompile(`<think>[\s\S]*?</think>`)
	cleaned := re.ReplaceAllString(text, "")

	// Trim whitespace from the result (equivalent to .trim() in JavaScript)
	return strings.TrimSpace(cleaned)
}
