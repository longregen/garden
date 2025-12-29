# AI Adapters Documentation

This document provides comprehensive documentation for the AI service adapters in the Garden project. These adapters implement various AI capabilities including text summarization, language model interactions, and embedding generation.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [AI Service Integration](#ai-service-integration)
3. [LLM (Ollama) Integration](#llm-ollama-integration)
4. [Embedding Generation Service](#embedding-generation-service)
5. [Model Configuration](#model-configuration)
6. [Usage Patterns and Examples](#usage-patterns-and-examples)
7. [Error Handling](#error-handling)
8. [Testing Strategies](#testing-strategies)

## Architecture Overview

The AI adapters follow the Hexagonal Architecture (Ports and Adapters) pattern, implementing output port interfaces defined in `internal/port/output/`. This design allows for:

- **Loose coupling**: Business logic is independent of AI service implementations
- **Testability**: Services can be easily mocked or stubbed
- **Flexibility**: Different AI providers can be swapped without changing core logic

### Port Interfaces

Three main port interfaces define the contracts for AI services:

```go
// AIService - For high-level AI operations like summarization
type AIService interface {
    GenerateSummary(ctx context.Context, content, url string, maxWords int) (string, error)
}

// LLMService - For direct language model interactions
type LLMService interface {
    CallLLM(ctx context.Context, prompt string) (string, error)
}

// EmbeddingsService - For generating text embeddings
type EmbeddingsService interface {
    GetEmbedding(ctx context.Context, text string) ([]entity.Embedding, error)
}
```

### Domain Entity

The `Embedding` entity represents a text chunk and its vector representation:

```go
type Embedding struct {
    Text      string    // The text chunk
    Embedding []float32 // Vector representation
}
```

## AI Service Integration

**Location**: `/home/user/garden/internal/adapter/secondary/ai/service.go`

The AI Service adapter provides high-level AI operations, specifically text summarization using Ollama.

### Implementation Details

```go
type Service struct {
    serviceURL string      // Ollama API endpoint
    apiKey     string      // Optional API key for authentication
    client     *http.Client // HTTP client for API calls
}
```

### Constructor

```go
func NewService(serviceURL, apiKey string) *Service
```

**Parameters**:
- `serviceURL`: The Ollama API endpoint (defaults to `http://localhost:11434` if empty)
- `apiKey`: Optional API key for authenticated requests

**Default Configuration**:
- Default URL: `http://localhost:11434`
- Default model: `current-default:latest`

### Summary Generation

The `GenerateSummary` method generates concise summaries of article content:

```go
func (s *Service) GenerateSummary(ctx context.Context, content, url string, maxWords int) (string, error)
```

**Parameters**:
- `ctx`: Context for request cancellation and timeout control
- `content`: The full text content to summarize
- `url`: Source URL (included in prompt for context)
- `maxWords`: Maximum word count for summary (defaults to 400 if 0)

**Process Flow**:
1. Constructs a structured prompt with the article content and URL
2. Sends request to Ollama's `/api/generate` endpoint
3. Receives and parses the response
4. Strips `<think>...</think>` tags from the output
5. Returns cleaned summary text

**Request Format**:
```json
{
    "model": "current-default:latest",
    "prompt": "I have read the following article of url {url}:\n\n\n===\n{content}\n\n===\nNow, what would be your summary of this article? Please use less than {maxWords} words",
    "stream": false
}
```

**Response Format**:
```json
{
    "response": "The summary text...",
    "done": true
}
```

### Special Features

#### Think Tag Stripping

The service automatically removes `<think>...</think>` tags from responses. This is useful when using models that include reasoning steps in their output:

```go
func stripThinkTags(text string) string {
    re := regexp.MustCompile(`<think>[\s\S]*?</think>`)
    cleaned := re.ReplaceAllString(text, "")
    return strings.TrimSpace(cleaned)
}
```

## LLM (Ollama) Integration

**Location**: `/home/user/garden/internal/adapter/secondary/llm/ollama.go`

The LLM service provides direct access to Ollama's language model capabilities for custom prompts.

### Implementation Details

```go
type OllamaService struct {
    baseURL string      // Ollama API base URL
    model   string      // Model identifier
    client  *http.Client // HTTP client
}
```

### Constructor

```go
func NewOllamaService(baseURL, model string) output.LLMService
```

**Parameters**:
- `baseURL`: Ollama API endpoint (defaults to `http://localhost:11434`)
- `model`: Model name (defaults to `current-default:latest`)

**Defaults**:
- Base URL: `http://localhost:11434`
- Model: `current-default:latest`

### Direct LLM Calls

```go
func (s *OllamaService) CallLLM(ctx context.Context, prompt string) (string, error)
```

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `prompt`: The complete prompt text

**Use Cases**:
- Custom AI operations not covered by higher-level services
- Experimental features requiring direct model access
- Fine-tuned prompt engineering
- Multi-turn conversations

**Request Format**:
```json
{
    "model": "current-default:latest",
    "prompt": "Your prompt here...",
    "stream": false
}
```

### Difference from AI Service

| Feature | AI Service | LLM Service |
|---------|-----------|-------------|
| **Purpose** | High-level operations (summarization) | Direct model access |
| **Prompt** | Structured, predefined format | Fully customizable |
| **Processing** | Includes post-processing (think tag removal) | Raw model output |
| **Use Case** | Standard operations | Custom/experimental features |

## Embedding Generation Service

The project includes multiple embedding service implementations to support different deployment scenarios.

### Ollama Embedding Service

**Location**: `/home/user/garden/internal/adapter/secondary/embedding/ollama.go`

Provides two service implementations:

#### 1. OllamaEmbeddingService (Single Embeddings)

For generating embeddings of individual text segments:

```go
type OllamaEmbeddingService struct {
    baseURL string
    model   string
    client  *http.Client
}
```

**Constructor**:
```go
func NewOllamaEmbeddingService(baseURL, model string) output.EmbeddingService
```

**Defaults**:
- Base URL: `http://localhost:11434`
- Model: `nomic-embed-text:latest`

**Method**:
```go
func (s *OllamaEmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error)
```

Returns a single embedding vector as `[]float32`.

#### 2. OllamaEmbeddingsService (Chunked Embeddings)

For handling large texts with automatic chunking:

```go
type OllamaEmbeddingsService struct {
    baseURL   string
    model     string
    client    *http.Client
    chunkSize int  // Default: 8000 characters
}
```

**Constructor**:
```go
func NewOllamaEmbeddingsService(baseURL, model string) output.EmbeddingsService
```

**Method**:
```go
func (s *OllamaEmbeddingsService) GetEmbedding(ctx context.Context, text string) ([]entity.Embedding, error)
```

Returns multiple embeddings (one per chunk) as `[]entity.Embedding`.

### Text Chunking Algorithm

The service intelligently splits large texts:

1. **Size-based chunking**: Splits text into ~8000 character chunks
2. **Sentence-aware**: Attempts to split on sentence boundaries (`.`, `!`, `?`)
3. **Paragraph-aware**: Also splits on newlines for better semantic coherence

**Algorithm Details**:
```go
func (s *OllamaEmbeddingsService) chunkText(text string) []string {
    // If text fits in one chunk, return as-is
    if len(text) <= s.chunkSize {
        return []string{text}
    }

    // Split by sentences
    sentences := splitSentences(text)

    // Group sentences into chunks
    for _, sentence := range sentences {
        if len(currentChunk) + len(sentence) > s.chunkSize && len(currentChunk) > 0 {
            // Start new chunk
            chunks = append(chunks, strings.TrimSpace(currentChunk))
            currentChunk = sentence
        } else {
            // Add to current chunk
            currentChunk += " " + sentence
        }
    }

    return chunks
}
```

### Generic Embedding Service

**Location**: `/home/user/garden/internal/adapter/secondary/embedding/service.go`

A generic HTTP-based embedding service that can work with any API endpoint:

```go
type Service struct {
    serviceURL string  // Custom API endpoint
    apiKey     string  // API authentication key
    client     *http.Client
}
```

**Constructor**:
```go
func NewService(serviceURL, apiKey string) *Service
```

**Request Format**:
```json
{
    "prompt": "Text to embed...",
    "operation": "query: "
}
```

**Response Format** (Array of tuples):
```json
[
    ["chunk 1 text", [0.1, 0.2, 0.3, ...]],
    ["chunk 2 text", [0.4, 0.5, 0.6, ...]]
]
```

This service is ideal for:
- Custom embedding APIs
- Cloud-based embedding services
- Services requiring special authentication

### Stub Embedding Service

**Location**: `/home/user/garden/internal/adapter/secondary/embedding/stub.go`

Placeholder implementations for development and testing:

```go
type StubEmbeddingService struct{}

func (s *StubEmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
    return nil, fmt.Errorf("embedding service not implemented")
}
```

**Use Cases**:
- Development without Ollama/embedding service running
- Testing error handling paths
- Graceful degradation (falls back to text search)

## Model Configuration

### Supported Models

#### LLM Models (Text Generation)
- **Default**: `current-default:latest`
- **Alternatives**: Any Ollama-compatible model
  - `llama2`
  - `mistral`
  - `codellama`
  - `deepseek-coder`
  - Custom fine-tuned models

#### Embedding Models
- **Default**: `nomic-embed-text:latest`
- **Alternatives**:
  - `all-minilm`
  - `mxbai-embed-large`
  - Custom embedding models

### Configuration Options

#### Service URL Configuration

```go
// Default local Ollama instance
service := ai.NewService("", "")

// Custom Ollama instance
service := ai.NewService("http://ollama-server:11434", "")

// With API key (if using authenticated endpoint)
service := ai.NewService("http://api.example.com", "your-api-key")
```

#### Model Selection

```go
// Default model
llmService := llm.NewOllamaService("", "")

// Specific model
llmService := llm.NewOllamaService("", "llama2:latest")

// Custom endpoint and model
llmService := llm.NewOllamaService("http://custom-server:11434", "mistral:latest")
```

#### Embedding Configuration

```go
// Single embedding service with defaults
embeddingService := embedding.NewOllamaEmbeddingService("", "")

// With custom model
embeddingService := embedding.NewOllamaEmbeddingService("", "mxbai-embed-large")

// Chunked embeddings for large texts
embeddingsService := embedding.NewOllamaEmbeddingsService("", "nomic-embed-text:latest")
```

### Environment-based Configuration

Recommended pattern for production deployment:

```go
import "os"

func initializeAIServices() {
    ollamaURL := os.Getenv("OLLAMA_URL")
    if ollamaURL == "" {
        ollamaURL = "http://localhost:11434"
    }

    llmModel := os.Getenv("LLM_MODEL")
    if llmModel == "" {
        llmModel = "current-default:latest"
    }

    embeddingModel := os.Getenv("EMBEDDING_MODEL")
    if embeddingModel == "" {
        embeddingModel = "nomic-embed-text:latest"
    }

    aiService := ai.NewService(ollamaURL, "")
    llmService := llm.NewOllamaService(ollamaURL, llmModel)
    embeddingService := embedding.NewOllamaEmbeddingsService(ollamaURL, embeddingModel)
}
```

## Usage Patterns and Examples

### Example 1: Generating Article Summaries

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "garden3/internal/adapter/secondary/ai"
)

func main() {
    // Initialize AI service
    aiService := ai.NewService("http://localhost:11434", "")

    // Article content
    content := `
    Artificial intelligence has made significant strides in recent years.
    Machine learning models can now perform tasks that were previously
    thought to require human intelligence...
    `

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Generate summary
    summary, err := aiService.GenerateSummary(
        ctx,
        content,
        "https://example.com/ai-article",
        200, // max 200 words
    )
    if err != nil {
        log.Fatalf("Failed to generate summary: %v", err)
    }

    fmt.Printf("Summary: %s\n", summary)
}
```

### Example 2: Direct LLM Interaction

```go
package main

import (
    "context"
    "fmt"
    "log"

    "garden3/internal/adapter/secondary/llm"
)

func main() {
    // Initialize LLM service
    llmService := llm.NewOllamaService("http://localhost:11434", "llama2")

    // Custom prompt
    prompt := `
    Given the following code snippet, explain what it does:

    func fibonacci(n int) int {
        if n <= 1 {
            return n
        }
        return fibonacci(n-1) + fibonacci(n-2)
    }

    Provide a concise explanation.
    `

    ctx := context.Background()

    // Call LLM
    response, err := llmService.CallLLM(ctx, prompt)
    if err != nil {
        log.Fatalf("LLM call failed: %v", err)
    }

    fmt.Printf("Response: %s\n", response)
}
```

### Example 3: Generating Embeddings for Search

```go
package main

import (
    "context"
    "fmt"
    "log"

    "garden3/internal/adapter/secondary/embedding"
)

func main() {
    // Initialize embedding service with chunking support
    embeddingService := embedding.NewOllamaEmbeddingsService(
        "http://localhost:11434",
        "nomic-embed-text:latest",
    )

    // Large document
    document := `
    This is a very large document that will be automatically chunked
    into smaller pieces. Each chunk will get its own embedding vector
    for better semantic search capabilities...
    ` // (imagine this is 20,000 characters)

    ctx := context.Background()

    // Generate embeddings (automatically chunked)
    embeddings, err := embeddingService.GetEmbedding(ctx, document)
    if err != nil {
        log.Fatalf("Failed to generate embeddings: %v", err)
    }

    fmt.Printf("Generated %d embedding chunks\n", len(embeddings))
    for i, emb := range embeddings {
        fmt.Printf("Chunk %d: %d dimensions, text length: %d\n",
            i+1, len(emb.Embedding), len(emb.Text))
    }
}
```

### Example 4: Single Text Embedding

```go
package main

import (
    "context"
    "fmt"
    "log"

    "garden3/internal/adapter/secondary/embedding"
)

func main() {
    // Initialize single embedding service
    embeddingService := embedding.NewOllamaEmbeddingService("", "")

    // Short query text
    query := "What is the capital of France?"

    ctx := context.Background()

    // Generate embedding
    vector, err := embeddingService.GetEmbedding(ctx, query)
    if err != nil {
        log.Fatalf("Failed to generate embedding: %v", err)
    }

    fmt.Printf("Embedding vector dimensions: %d\n", len(vector))
    fmt.Printf("First 5 values: %v\n", vector[:5])
}
```

### Example 5: Using Stub for Development

```go
package main

import (
    "context"
    "log"

    "garden3/internal/adapter/secondary/embedding"
)

func main() {
    // Use stub when Ollama is not available
    embeddingService := embedding.NewStubEmbeddingService()

    ctx := context.Background()

    // This will return an error, allowing graceful fallback
    _, err := embeddingService.GetEmbedding(ctx, "test query")
    if err != nil {
        log.Printf("Embedding not available, falling back to text search: %v", err)
        // Implement fallback logic here
        performTextBasedSearch()
    }
}

func performTextBasedSearch() {
    // Fallback search implementation
}
```

### Example 6: Service Composition in Application

```go
package main

import (
    "garden3/internal/adapter/secondary/ai"
    "garden3/internal/adapter/secondary/llm"
    "garden3/internal/adapter/secondary/embedding"
    "garden3/internal/port/output"
)

// Application holds all AI services
type Application struct {
    aiService        output.AIService
    llmService       output.LLMService
    embeddingService output.EmbeddingsService
}

func NewApplication(ollamaURL string) *Application {
    return &Application{
        aiService:        ai.NewService(ollamaURL, ""),
        llmService:       llm.NewOllamaService(ollamaURL, "current-default:latest"),
        embeddingService: embedding.NewOllamaEmbeddingsService(ollamaURL, "nomic-embed-text:latest"),
    }
}

func (app *Application) ProcessBookmark(ctx context.Context, url, content string) error {
    // Generate summary
    summary, err := app.aiService.GenerateSummary(ctx, content, url, 300)
    if err != nil {
        return err
    }

    // Generate embeddings for search
    embeddings, err := app.embeddingService.GetEmbedding(ctx, content)
    if err != nil {
        // Fall back to text-only processing
        embeddings = nil
    }

    // Store bookmark with summary and embeddings
    return app.storeBookmark(url, content, summary, embeddings)
}

func (app *Application) storeBookmark(url, content, summary string, embeddings []entity.Embedding) error {
    // Implementation details...
    return nil
}
```

## Error Handling

### Common Error Scenarios

#### 1. Service Unavailable

```go
summary, err := aiService.GenerateSummary(ctx, content, url, 400)
if err != nil {
    // Check for connection errors
    if strings.Contains(err.Error(), "connection refused") {
        log.Printf("Ollama service not running: %v", err)
        // Use fallback or return user-friendly message
        return "Summary unavailable - AI service offline"
    }
    return fmt.Errorf("summarization failed: %w", err)
}
```

#### 2. Timeout Handling

```go
// Set appropriate timeout
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

response, err := llmService.CallLLM(ctx, prompt)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Printf("LLM call timed out after 60 seconds")
        // Handle timeout gracefully
    }
    return err
}
```

#### 3. Invalid Response Format

```go
embeddings, err := embeddingService.GetEmbedding(ctx, text)
if err != nil {
    if strings.Contains(err.Error(), "invalid response format") {
        log.Printf("Unexpected API response format: %v", err)
        // Log for debugging, return generic error to user
        return errors.New("embedding generation failed")
    }
    return err
}
```

#### 4. HTTP Error Codes

The services handle various HTTP error codes:

- `400`: Invalid request (check prompt format, model name)
- `401`/`403`: Authentication issues (check API key)
- `404`: Model not found (check model name)
- `500`: Server error (retry with backoff)
- `503`: Service overloaded (implement rate limiting)

### Best Practices

1. **Always use context with timeout**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

2. **Implement graceful degradation**
   ```go
   embeddings, err := embeddingService.GetEmbedding(ctx, text)
   if err != nil {
       log.Printf("Embedding generation failed, using text search: %v", err)
       return performTextSearch(text)
   }
   ```

3. **Log errors with context**
   ```go
   if err != nil {
       log.Printf("Failed to generate summary for URL %s: %v", url, err)
       return err
   }
   ```

4. **Validate inputs before API calls**
   ```go
   if len(content) == 0 {
       return "", errors.New("content cannot be empty")
   }
   if maxWords < 0 {
       maxWords = 400
   }
   ```

## Testing Strategies

### Unit Testing with Mocks

Create mock implementations of the port interfaces:

```go
package mocks

import (
    "context"

    "garden3/internal/domain/entity"
)

type MockAIService struct {
    GenerateSummaryFunc func(ctx context.Context, content, url string, maxWords int) (string, error)
}

func (m *MockAIService) GenerateSummary(ctx context.Context, content, url string, maxWords int) (string, error) {
    if m.GenerateSummaryFunc != nil {
        return m.GenerateSummaryFunc(ctx, content, url, maxWords)
    }
    return "Mock summary", nil
}

type MockLLMService struct {
    CallLLMFunc func(ctx context.Context, prompt string) (string, error)
}

func (m *MockLLMService) CallLLM(ctx context.Context, prompt string) (string, error) {
    if m.CallLLMFunc != nil {
        return m.CallLLMFunc(ctx, prompt)
    }
    return "Mock response", nil
}

type MockEmbeddingsService struct {
    GetEmbeddingFunc func(ctx context.Context, text string) ([]entity.Embedding, error)
}

func (m *MockEmbeddingsService) GetEmbedding(ctx context.Context, text string) ([]entity.Embedding, error) {
    if m.GetEmbeddingFunc != nil {
        return m.GetEmbeddingFunc(ctx, text)
    }
    return []entity.Embedding{
        {
            Text:      text,
            Embedding: []float32{0.1, 0.2, 0.3},
        },
    }, nil
}
```

### Integration Testing

Test against real Ollama instance:

```go
// +build integration

package integration_test

import (
    "context"
    "testing"
    "time"

    "garden3/internal/adapter/secondary/ai"
    "garden3/internal/adapter/secondary/llm"
    "garden3/internal/adapter/secondary/embedding"
)

func TestOllamaIntegration(t *testing.T) {
    // Skip if Ollama not available
    ollamaURL := "http://localhost:11434"

    t.Run("AI Summary Generation", func(t *testing.T) {
        service := ai.NewService(ollamaURL, "")
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        summary, err := service.GenerateSummary(
            ctx,
            "Test content for summarization",
            "http://test.com",
            50,
        )

        if err != nil {
            t.Fatalf("Summary generation failed: %v", err)
        }

        if len(summary) == 0 {
            t.Error("Expected non-empty summary")
        }
    })

    t.Run("LLM Call", func(t *testing.T) {
        service := llm.NewOllamaService(ollamaURL, "")
        ctx := context.Background()

        response, err := service.CallLLM(ctx, "Hello, how are you?")
        if err != nil {
            t.Fatalf("LLM call failed: %v", err)
        }

        if len(response) == 0 {
            t.Error("Expected non-empty response")
        }
    })

    t.Run("Embedding Generation", func(t *testing.T) {
        service := embedding.NewOllamaEmbeddingService(ollamaURL, "")
        ctx := context.Background()

        vector, err := service.GetEmbedding(ctx, "Test text for embedding")
        if err != nil {
            t.Fatalf("Embedding generation failed: %v", err)
        }

        if len(vector) == 0 {
            t.Error("Expected non-empty embedding vector")
        }
    })
}
```

### Using Stub for Testing

```go
package business_test

import (
    "context"
    "testing"

    "garden3/internal/adapter/secondary/embedding"
)

func TestFallbackToTextSearch(t *testing.T) {
    // Use stub to simulate missing embedding service
    embeddingService := embedding.NewStubEmbeddingService()

    ctx := context.Background()
    _, err := embeddingService.GetEmbedding(ctx, "test query")

    if err == nil {
        t.Error("Expected error from stub service")
    }

    // Test that application handles this gracefully
    // and falls back to text search
}
```

### Table-Driven Tests

```go
func TestChunkText(t *testing.T) {
    service := embedding.NewOllamaEmbeddingsService("", "")

    tests := []struct {
        name      string
        input     string
        chunkSize int
        wantCount int
    }{
        {
            name:      "Short text",
            input:     "Hello world",
            chunkSize: 8000,
            wantCount: 1,
        },
        {
            name:      "Long text",
            input:     strings.Repeat("Lorem ipsum. ", 1000),
            chunkSize: 8000,
            wantCount: 2, // Approximate
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            embeddings, err := service.GetEmbedding(ctx, tt.input)

            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }

            if len(embeddings) != tt.wantCount {
                t.Errorf("Expected %d chunks, got %d", tt.wantCount, len(embeddings))
            }
        })
    }
}
```

## Performance Considerations

### 1. Chunking Strategy

For very large documents (>50,000 characters), consider:
- Adjusting chunk size based on model capabilities
- Parallel embedding generation
- Caching frequently accessed embeddings

### 2. Connection Pooling

The services use `http.Client` which maintains connection pools. For high-throughput scenarios:

```go
client := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 3. Request Batching

For multiple embedding requests, consider batching:

```go
// Instead of individual calls
for _, text := range texts {
    embedding, _ := service.GetEmbedding(ctx, text)
}

// Batch process
var wg sync.WaitGroup
results := make(chan result, len(texts))

for _, text := range texts {
    wg.Add(1)
    go func(t string) {
        defer wg.Done()
        emb, err := service.GetEmbedding(ctx, t)
        results <- result{emb, err}
    }(text)
}

wg.Wait()
close(results)
```

### 4. Caching Strategies

Implement caching for repeated queries:

```go
type CachedEmbeddingService struct {
    base  output.EmbeddingService
    cache map[string][]float32
    mu    sync.RWMutex
}

func (s *CachedEmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
    // Check cache
    s.mu.RLock()
    if cached, ok := s.cache[text]; ok {
        s.mu.RUnlock()
        return cached, nil
    }
    s.mu.RUnlock()

    // Generate and cache
    embedding, err := s.base.GetEmbedding(ctx, text)
    if err == nil {
        s.mu.Lock()
        s.cache[text] = embedding
        s.mu.Unlock()
    }

    return embedding, err
}
```

## Conclusion

The AI adapters in this project provide a flexible, testable, and maintainable approach to integrating AI capabilities. By following the hexagonal architecture pattern and implementing well-defined port interfaces, the system can easily adapt to different AI providers and deployment scenarios while maintaining clean separation of concerns.

For additional information or questions, please refer to:
- [Ollama API Documentation](https://github.com/ollama/ollama/blob/main/docs/api.md)
- Project's port interface definitions in `/home/user/garden/internal/port/output/`
- Domain entity definitions in `/home/user/garden/internal/domain/entity/`
