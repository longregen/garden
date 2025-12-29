package output

import (
	"context"
)

// LLMService defines the interface for calling Language Learning Models
type LLMService interface {
	// CallLLM sends a prompt to the LLM and returns the complete response
	CallLLM(ctx context.Context, prompt string) (string, error)
}
