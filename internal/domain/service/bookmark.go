package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"garden3/internal/port/output"
)

// BookmarkService implements the BookmarkUseCase interface
type BookmarkService struct {
	repo            output.BookmarkRepository
	httpFetcher     output.HTTPFetcher
	embeddingsService output.EmbeddingsService
	aiService       output.AIService
	contentProcessor output.ContentProcessor
}

// NewBookmarkService creates a new bookmark service
func NewBookmarkService(
	repo output.BookmarkRepository,
	httpFetcher output.HTTPFetcher,
	embeddingsService output.EmbeddingsService,
	aiService output.AIService,
	contentProcessor output.ContentProcessor,
) *BookmarkService {
	return &BookmarkService{
		repo:            repo,
		httpFetcher:     httpFetcher,
		embeddingsService: embeddingsService,
		aiService:       aiService,
		contentProcessor: contentProcessor,
	}
}

func (s *BookmarkService) ListBookmarks(ctx context.Context, filters entity.BookmarkFilters) (*input.PaginatedResponse[entity.BookmarkWithTitle], error) {
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	bookmarks, err := s.repo.ListBookmarks(
		ctx,
		filters.CategoryID,
		filters.SearchQuery,
		filters.StartCreationDate,
		filters.EndCreationDate,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookmarks: %w", err)
	}

	total, err := s.repo.CountBookmarks(
		ctx,
		filters.CategoryID,
		filters.SearchQuery,
		filters.StartCreationDate,
		filters.EndCreationDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count bookmarks: %w", err)
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))

	return &input.PaginatedResponse[entity.BookmarkWithTitle]{
		Data:       bookmarks,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}

func (s *BookmarkService) GetRandomBookmark(ctx context.Context) (uuid.UUID, error) {
	id, err := s.repo.GetRandomBookmark(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get random bookmark: %w", err)
	}
	return id, nil
}

func (s *BookmarkService) GetBookmarkDetails(ctx context.Context, bookmarkID uuid.UUID) (*entity.BookmarkDetails, error) {
	details, err := s.repo.GetBookmarkDetails(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark details: %w", err)
	}

	questions, err := s.repo.GetBookmarkQuestions(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark questions: %w", err)
	}

	details.Questions = questions
	return details, nil
}

func (s *BookmarkService) SearchSimilarBookmarks(ctx context.Context, query string, strategy string) ([]entity.BookmarkWithTitle, error) {
	if strategy == "" {
		strategy = "qa-v2-passage"
	}

	embeddings, err := s.embeddingsService.GetEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	results, err := s.repo.SearchSimilarBookmarks(ctx, embeddings[0].Embedding, strategy, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar bookmarks: %w", err)
	}

	return results, nil
}

func (s *BookmarkService) UpdateBookmarkQuestion(ctx context.Context, input entity.UpdateQuestionInput) error {
	bookmark, err := s.repo.GetBookmarkForQuestion(ctx, input.BookmarkID)
	if err != nil {
		return fmt.Errorf("failed to get bookmark: %w", err)
	}

	newContent := fmt.Sprintf("%s?\n%s", input.NewQuestion, input.NewAnswer)

	embeddings, err := s.embeddingsService.GetEmbedding(ctx, newContent)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return fmt.Errorf("no embedding generated")
	}

	err = s.repo.UpdateBookmarkQuestion(ctx, newContent, embeddings[0].Embedding, input.ReferenceID, input.BookmarkID)
	if err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}

	observationData := map[string]interface{}{
		"bookmarkId":       input.BookmarkID.String(),
		"title":            bookmark.Title,
		"summary":          bookmark.Summary,
		"previousQuestion": input.PreviousQuestion,
		"previousAnswer":   input.PreviousAnswer,
		"newQuestion":      input.NewQuestion,
		"newAnswer":        input.NewAnswer,
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	dataJSON, err := json.Marshal(observationData)
	if err != nil {
		return fmt.Errorf("failed to marshal observation: %w", err)
	}

	err = s.repo.CreateObservation(ctx, string(dataJSON), "qa-edit", "user-edit", "edit,question,answer", input.BookmarkID.String())
	if err != nil {
		return fmt.Errorf("failed to create observation: %w", err)
	}

	return nil
}

func (s *BookmarkService) DeleteBookmarkQuestion(ctx context.Context, input entity.DeleteQuestionInput) error {
	bookmark, err := s.repo.GetBookmarkForQuestion(ctx, input.BookmarkID)
	if err != nil {
		return fmt.Errorf("failed to get bookmark: %w", err)
	}

	err = s.repo.DeleteBookmarkQuestion(ctx, input.ReferenceID, input.BookmarkID)
	if err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}

	observationData := map[string]interface{}{
		"bookmarkId": input.BookmarkID.String(),
		"title":      bookmark.Title,
		"summary":    bookmark.Summary,
		"question":   input.Question,
		"answer":     input.Answer,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	dataJSON, err := json.Marshal(observationData)
	if err != nil {
		return fmt.Errorf("failed to marshal observation: %w", err)
	}

	err = s.repo.CreateObservation(ctx, string(dataJSON), "qa-delete", "user-delete", "delete,question,answer", input.BookmarkID.String())
	if err != nil {
		return fmt.Errorf("failed to create observation: %w", err)
	}

	return nil
}

func (s *BookmarkService) FetchBookmarkContent(ctx context.Context, bookmarkID uuid.UUID) (*entity.FetchResult, error) {
	bookmark, err := s.repo.GetBookmark(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	url := sanitizeURL(bookmark.URL)

	fetchCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	response, err := s.httpFetcher.Fetch(fetchCtx, url, 25000)
	if err != nil {
		errorContent := []byte(err.Error())
		storeErr := s.repo.InsertHttpResponse(ctx, bookmarkID, 500, "{}", errorContent, time.Now())
		if storeErr != nil {
			return nil, fmt.Errorf("fetch failed and failed to store error: %v, %w", err, storeErr)
		}
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}

	err = s.repo.InsertHttpResponse(ctx, bookmarkID, response.StatusCode, response.Headers, response.Content, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to store http response: %w", err)
	}

	return &entity.FetchResult{
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
		Content:    response.Content,
		Message:    "Fetch finished",
	}, nil
}

func (s *BookmarkService) ProcessWithLynx(ctx context.Context, bookmarkID uuid.UUID) (*entity.ProcessingResult, error) {
	httpResp, err := s.repo.GetLatestHttpResponse(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get http response: %w", err)
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(httpResp.Headers), &headers); err != nil {
		return nil, fmt.Errorf("failed to parse headers: %w", err)
	}

	contentType := getContentType(headers)
	if !strings.Contains(contentType, "text") && !strings.Contains(contentType, "html") {
		return &entity.ProcessingResult{
			Message: "Content is not HTML and cannot be processed",
		}, nil
	}

	processedContent, err := s.contentProcessor.ProcessWithLynx(ctx, string(httpResp.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to process with lynx: %w", err)
	}

	err = s.repo.InsertProcessedContent(ctx, bookmarkID, "lynx", processedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to store processed content: %w", err)
	}

	return &entity.ProcessingResult{
		Message: "Bookmark processed successfully",
	}, nil
}

func (s *BookmarkService) ProcessWithReader(ctx context.Context, bookmarkID uuid.UUID) (*entity.ProcessingResult, error) {
	httpResp, err := s.repo.GetLatestHttpResponse(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get http response: %w", err)
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(httpResp.Headers), &headers); err != nil {
		return nil, fmt.Errorf("failed to parse headers: %w", err)
	}

	contentType := getContentType(headers)
	if !strings.Contains(contentType, "text") && !strings.Contains(contentType, "html") {
		return &entity.ProcessingResult{
			Message: "Content is not HTML and cannot be processed",
		}, nil
	}

	bookmark, err := s.repo.GetBookmark(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	var processedContent string
	if strings.HasSuffix(sanitizeURL(bookmark.URL), "README.md") {
		processedContent = string(httpResp.Content)
	} else {
		processedContent, err = s.contentProcessor.ProcessWithReader(ctx, httpResp.Content, bookmark.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to process with reader: %w", err)
		}
	}

	err = s.repo.InsertProcessedContent(ctx, bookmarkID, "reader", processedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to store processed content: %w", err)
	}

	return &entity.ProcessingResult{
		Message: "Bookmark processed successfully",
		Content: &processedContent,
	}, nil
}

func (s *BookmarkService) CreateEmbeddingChunks(ctx context.Context, bookmarkID uuid.UUID) (*entity.EmbeddingResult, error) {
	processedContent, err := s.repo.GetProcessedContentByStrategy(ctx, bookmarkID, "reader")
	if err != nil {
		return nil, fmt.Errorf("failed to get processed content: %w", err)
	}

	if processedContent == nil {
		return nil, fmt.Errorf("no processed content found")
	}

	content := *processedContent
	if len(content) > 15000 {
		content = content[:15000]
	}

	embeddings, err := s.embeddingsService.GetEmbedding(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	var ids []uuid.UUID
	maxChunks := 20
	if len(embeddings) > maxChunks {
		embeddings = embeddings[:maxChunks]
	}

	for _, emb := range embeddings {
		id, err := s.repo.CreateEmbeddingChunk(ctx, bookmarkID, emb.Text, "chunked-reader", emb.Embedding)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding chunk: %w", err)
		}
		ids = append(ids, id)
	}

	var warning *string
	if len(embeddings) >= maxChunks {
		msg := "The content was too large, only the first ~10000 characters were processed"
		warning = &msg
	}

	return &entity.EmbeddingResult{
		IDs:     ids,
		Warning: warning,
	}, nil
}

func (s *BookmarkService) CreateSummaryEmbedding(ctx context.Context, bookmarkID uuid.UUID) (*entity.SummaryEmbeddingResult, error) {
	processedContent, err := s.repo.GetProcessedContentByStrategy(ctx, bookmarkID, "reader")
	if err != nil {
		return nil, fmt.Errorf("failed to get processed content: %w", err)
	}

	if processedContent == nil {
		return nil, fmt.Errorf("no processed content found")
	}

	bookmark, err := s.repo.GetBookmark(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	var summaryText string
	var embeddings []entity.Embedding
	wordCount := 300

	for wordCount > 200 {
		summaryText, err = s.aiService.GenerateSummary(ctx, *processedContent, bookmark.URL, wordCount)
		if err != nil {
			return nil, fmt.Errorf("failed to generate summary: %w", err)
		}

		embeddings, err = s.embeddingsService.GetEmbedding(ctx, summaryText)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %w", err)
		}

		if len(embeddings) == 1 {
			break
		}

		wordCount -= 10
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("failed to generate valid embedding")
	}

	id, err := s.repo.CreateEmbeddingChunk(ctx, bookmarkID, embeddings[0].Text, "summary-reader", embeddings[0].Embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to create summary embedding: %w", err)
	}

	var warning *string
	if len(embeddings) > 1 {
		msg := "The content was too large, only the first ~10000 characters were processed"
		warning = &msg
	}

	return &entity.SummaryEmbeddingResult{
		IDs:     []uuid.UUID{id},
		Summary: summaryText,
		Warning: warning,
	}, nil
}

func (s *BookmarkService) GetBookmarkTitle(ctx context.Context, bookmarkID uuid.UUID) (*entity.TitleExtractionResult, error) {
	titleData, err := s.repo.GetBookmarkTitle(ctx, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark title data: %w", err)
	}

	details := entity.BookmarkDetails{
		BookmarkID:   titleData.BookmarkID,
		URL:          titleData.URL,
		CreationDate: titleData.CreationDate,
		Title:        titleData.ExistingTitle,
	}

	if titleData.ExistingTitle != nil {
		return &entity.TitleExtractionResult{
			Data: details,
		}, nil
	}

	if titleData.ReaderTitle != nil {
		err = s.repo.InsertBookmarkTitle(ctx, bookmarkID, *titleData.ReaderTitle, "reader-title")
		if err != nil {
			return nil, fmt.Errorf("failed to insert reader title: %w", err)
		}
		return &entity.TitleExtractionResult{
			Data:  details,
			Title: titleData.ReaderTitle,
		}, nil
	}

	if titleData.RawContent != nil {
		title := extractHTMLTitle(string(titleData.RawContent))
		if title != nil {
			err = s.repo.InsertBookmarkTitle(ctx, bookmarkID, *title, "html-title")
			if err != nil {
				return nil, fmt.Errorf("failed to insert html title: %w", err)
			}
			return &entity.TitleExtractionResult{
				Data:  details,
				Title: title,
			}, nil
		}
	}

	return &entity.TitleExtractionResult{
		Data: details,
	}, nil
}

func (s *BookmarkService) GetMissingHttpResponses(ctx context.Context) ([]entity.Bookmark, error) {
	bookmarks, err := s.repo.GetMissingHttpResponses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get missing http responses: %w", err)
	}
	return bookmarks, nil
}

func (s *BookmarkService) GetMissingReaderContent(ctx context.Context) ([]entity.Bookmark, error) {
	bookmarks, err := s.repo.GetMissingReaderContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get missing reader content: %w", err)
	}
	return bookmarks, nil
}

// Helper functions

func sanitizeURL(url string) string {
	// Simple sanitization - can be enhanced
	return strings.TrimSpace(url)
}

func getContentType(headers map[string]string) string {
	for k, v := range headers {
		if strings.ToLower(k) == "content-type" {
			return strings.ToLower(v)
		}
	}
	return ""
}

func extractHTMLTitle(html string) *string {
	re := regexp.MustCompile(`(?i)<title>([^<]+)</title>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		return &title
	}
	return nil
}
