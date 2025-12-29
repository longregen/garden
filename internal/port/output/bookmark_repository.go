package output

import (
	"context"
	"time"

	"github.com/google/uuid"
	"garden3/internal/domain/entity"
)

// BookmarkRepository defines the data access operations for bookmarks
type BookmarkRepository interface {
	// GetBookmark retrieves a bookmark by ID
	GetBookmark(ctx context.Context, bookmarkID uuid.UUID) (*entity.Bookmark, error)

	// ListBookmarks retrieves filtered and paginated bookmarks
	ListBookmarks(ctx context.Context, categoryID *uuid.UUID, searchQuery *string, startDate *time.Time, endDate *time.Time, limit, offset int32) ([]entity.BookmarkWithTitle, error)

	// CountBookmarks returns the total count of bookmarks matching filters
	CountBookmarks(ctx context.Context, categoryID *uuid.UUID, searchQuery *string, startDate *time.Time, endDate *time.Time) (int64, error)

	// GetRandomBookmark retrieves a random bookmark ID
	GetRandomBookmark(ctx context.Context) (uuid.UUID, error)

	// GetBookmarkDetails retrieves complete bookmark details
	GetBookmarkDetails(ctx context.Context, bookmarkID uuid.UUID) (*entity.BookmarkDetails, error)

	// GetBookmarkQuestions retrieves all Q&A pairs for a bookmark
	GetBookmarkQuestions(ctx context.Context, bookmarkID uuid.UUID) ([]entity.BookmarkQuestion, error)

	// SearchSimilarBookmarks performs vector similarity search
	SearchSimilarBookmarks(ctx context.Context, embedding []float32, strategy string, limit int32) ([]entity.BookmarkWithTitle, error)

	// UpdateBookmarkQuestion updates a Q&A content reference
	UpdateBookmarkQuestion(ctx context.Context, content string, embedding []float32, referenceID, bookmarkID uuid.UUID) error

	// DeleteBookmarkQuestion deletes a Q&A content reference
	DeleteBookmarkQuestion(ctx context.Context, referenceID, bookmarkID uuid.UUID) error

	// GetBookmarkForQuestion retrieves bookmark info for Q&A operations
	GetBookmarkForQuestion(ctx context.Context, bookmarkID uuid.UUID) (*entity.BookmarkDetails, error)

	// CreateObservation creates an observation log entry
	CreateObservation(ctx context.Context, data, observationType, source, tags, ref string) error

	// InsertHttpResponse stores an HTTP response
	InsertHttpResponse(ctx context.Context, bookmarkID uuid.UUID, statusCode int32, headers string, content []byte, fetchDate time.Time) error

	// GetLatestHttpResponse retrieves the most recent HTTP response
	GetLatestHttpResponse(ctx context.Context, bookmarkID uuid.UUID) (*HTTPResponse, error)

	// InsertProcessedContent stores processed content
	InsertProcessedContent(ctx context.Context, bookmarkID uuid.UUID, strategyUsed, processedContent string) error

	// GetProcessedContentByStrategy retrieves processed content by strategy
	GetProcessedContentByStrategy(ctx context.Context, bookmarkID uuid.UUID, strategy string) (*string, error)

	// CreateEmbeddingChunk creates a content reference with embedding
	CreateEmbeddingChunk(ctx context.Context, bookmarkID uuid.UUID, content, strategy string, embedding []float32) (uuid.UUID, error)

	// GetBookmarkTitle retrieves bookmark with title-related data
	GetBookmarkTitle(ctx context.Context, bookmarkID uuid.UUID) (*TitleData, error)

	// InsertBookmarkTitle stores a bookmark title
	InsertBookmarkTitle(ctx context.Context, bookmarkID uuid.UUID, title, source string) error

	// GetMissingHttpResponses retrieves bookmarks without HTTP responses
	GetMissingHttpResponses(ctx context.Context) ([]entity.Bookmark, error)

	// GetMissingReaderContent retrieves bookmarks without reader content
	GetMissingReaderContent(ctx context.Context) ([]entity.Bookmark, error)
}

// HTTPResponse represents an HTTP response from the database
type HTTPResponse struct {
	ResponseID uuid.UUID
	StatusCode int32
	Headers    string
	Content    []byte
	FetchDate  time.Time
}

// TitleData represents bookmark data for title extraction
type TitleData struct {
	BookmarkID   uuid.UUID
	URL          string
	CreationDate time.Time
	ExistingTitle *string
	RawContent   []byte
	ReaderTitle  *string
}
