package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

// SearchService implements the search use case
type SearchService struct {
	repo             output.SearchRepository
	embeddingService output.EmbeddingService
	llmService       output.LLMService
	configService    input.ConfigurationUseCase
}

// NewSearchService creates a new search service
func NewSearchService(
	repo output.SearchRepository,
	embeddingService output.EmbeddingService,
	llmService output.LLMService,
	configService input.ConfigurationUseCase,
) *SearchService {
	return &SearchService{
		repo:             repo,
		embeddingService: embeddingService,
		llmService:       llmService,
		configService:    configService,
	}
}

// SearchAll performs a unified search across multiple tables
func (s *SearchService) SearchAll(ctx context.Context, query string, weights *entity.SearchWeights, limit int32) ([]entity.UnifiedSearchResult, error) {
	if weights == nil {
		defaultWeights := entity.DefaultSearchWeights()
		weights = &defaultWeights
	}

	if limit <= 0 {
		limit = 50
	}

	return s.repo.SearchAll(ctx, query, weights.ExactMatchWeight, weights.SimilarityWeight, weights.RecencyWeight, limit)
}

const searchPromptTemplateKey = "search.prompt.template"

const defaultPromptTemplate = `System: You are a helpful assistant that helps the user answering questions based on the provided context. The context is a set of questions and answers generated from the content of bookmarks of the user, including a summary of the source article.
When answering, if any article from the questions and answers is relevant, quote it with a link in markdown using the format [title](url)
If the context article is not relevant, dismiss it.
If you need to refer to "the context", mention "the bookmarks database" instead, for example: "In the bookmarks database there is an article related to..."

Context:
{{range .RetrievedItems}}Title: {{.BookmarkTitle}}
Question: {{.Question}}
Answer: {{.Answer}}
Summary: {{.Summary}}
Url: {{.BookmarkURL}}
{{end}}
User question: {{.UserQuestion}}

Please answer the user's question, and rely as much as possible on the provided context. If the context doesn't contain relevant information, say so.`

// AdvancedSearch performs an LLM-powered search with context from similar bookmarks
func (s *SearchService) AdvancedSearch(ctx context.Context, query string) (*entity.AdvancedSearchResult, error) {
	// Step 1: Get embedding for the query
	embedding, err := s.embeddingService.GetEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}

	// Step 2: Get similar questions from bookmarks
	similarQuestions, err := s.repo.GetSimilarQuestions(ctx, embedding, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar questions: %w", err)
	}

	// Step 3: Get prompt template from configuration
	templateStr, err := s.configService.GetValue(ctx, searchPromptTemplateKey)
	if err != nil || templateStr == nil {
		// Use default template
		defaultTemplate := defaultPromptTemplate
		templateStr = &defaultTemplate
	}

	// Step 4: Process the template
	renderedPrompt, err := processTemplate(*templateStr, query, similarQuestions)
	if err != nil {
		return nil, fmt.Errorf("failed to process template: %w", err)
	}

	// Step 5: Call LLM
	llmResponse, err := s.llmService.CallLLM(ctx, renderedPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// Step 6: Parse response to separate thinking from answer
	thinkingProcess, answer := parseResponse(llmResponse)

	// Step 7: Return structured result
	return &entity.AdvancedSearchResult{
		Query:            query,
		QueryString:      query,
		SimilarQuestions: similarQuestions,
		RenderedPrompt:   renderedPrompt,
		ThinkingProcess:  thinkingProcess,
		FullResponse:     llmResponse,
		Answer:           answer,
	}, nil
}

// processTemplate processes a template string with the query and retrieved items
func processTemplate(templateStr, userQuestion string, retrievedItems []entity.RetrievedItem) (string, error) {
	// Create template data
	data := map[string]interface{}{
		"UserQuestion":   userQuestion,
		"RetrievedItems": retrievedItems,
	}

	// Parse and execute template
	tmpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// parseResponse parses the LLM response to extract thinking process and final answer
func parseResponse(response string) (thinkingProcess string, answer string) {
	// Look for <think>...</think> pattern
	thinkRegex := regexp.MustCompile(`(?s)<think>(.*?)</think>`)
	match := thinkRegex.FindStringSubmatch(response)

	if len(match) > 1 {
		// Extract thinking content
		thinkingProcess = strings.TrimSpace(match[1])

		// Remove the thinking part from the answer
		answer = strings.TrimSpace(thinkRegex.ReplaceAllString(response, ""))
	} else {
		// No thinking process found, return full response as answer
		answer = response
	}

	// Fallback to full response if answer is empty
	if answer == "" {
		answer = response
	}

	return thinkingProcess, answer
}
