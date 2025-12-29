package input

import (
	"context"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// BookmarkUseCase defines the business operations for bookmarks
type BookmarkUseCase interface {
	// ListBookmarks retrieves paginated and filtered bookmarks
	ListBookmarks(ctx context.Context, filters entity.BookmarkFilters) (*PaginatedResponse[entity.BookmarkWithTitle], error)

	// GetRandomBookmark retrieves a random bookmark ID
	GetRandomBookmark(ctx context.Context) (uuid.UUID, error)

	// GetBookmarkDetails retrieves complete bookmark details with all relations
	GetBookmarkDetails(ctx context.Context, bookmarkID uuid.UUID) (*entity.BookmarkDetails, error)

	// SearchSimilarBookmarks performs vector similarity search
	SearchSimilarBookmarks(ctx context.Context, query string, strategy string) ([]entity.BookmarkWithTitle, error)

	// UpdateBookmarkQuestion updates a Q&A content reference
	UpdateBookmarkQuestion(ctx context.Context, input entity.UpdateQuestionInput) error

	// DeleteBookmarkQuestion deletes a Q&A content reference
	DeleteBookmarkQuestion(ctx context.Context, input entity.DeleteQuestionInput) error

	// FetchBookmarkContent fetches and stores HTTP content for a bookmark
	FetchBookmarkContent(ctx context.Context, bookmarkID uuid.UUID) (*entity.FetchResult, error)

	// ProcessWithLynx processes bookmark content using lynx
	ProcessWithLynx(ctx context.Context, bookmarkID uuid.UUID) (*entity.ProcessingResult, error)

	// ProcessWithReader processes bookmark content using reader mode
	ProcessWithReader(ctx context.Context, bookmarkID uuid.UUID) (*entity.ProcessingResult, error)

	// CreateEmbeddingChunks creates chunked embeddings for bookmark content
	CreateEmbeddingChunks(ctx context.Context, bookmarkID uuid.UUID) (*entity.EmbeddingResult, error)

	// CreateSummaryEmbedding creates a summary and its embedding
	CreateSummaryEmbedding(ctx context.Context, bookmarkID uuid.UUID) (*entity.SummaryEmbeddingResult, error)

	// GetBookmarkTitle extracts and stores the bookmark title
	GetBookmarkTitle(ctx context.Context, bookmarkID uuid.UUID) (*entity.TitleExtractionResult, error)

	// GetMissingHttpResponses retrieves bookmarks without HTTP responses
	GetMissingHttpResponses(ctx context.Context) ([]entity.Bookmark, error)

	// GetMissingReaderContent retrieves bookmarks without reader-processed content
	GetMissingReaderContent(ctx context.Context) ([]entity.Bookmark, error)
}
