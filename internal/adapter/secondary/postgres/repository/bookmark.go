package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"garden3/internal/adapter/secondary/postgres/generated/db"
	"garden3/internal/domain/entity"
	"garden3/internal/port/output"
)

// BookmarkRepository implements the output.BookmarkRepository interface
type BookmarkRepository struct {
	pool *pgxpool.Pool
}

// NewBookmarkRepository creates a new bookmark repository
func NewBookmarkRepository(pool *pgxpool.Pool) *BookmarkRepository {
	return &BookmarkRepository{
		pool: pool,
	}
}

func (r *BookmarkRepository) GetBookmark(ctx context.Context, bookmarkID uuid.UUID) (*entity.Bookmark, error) {
	queries := db.New(r.pool)
	dbBookmark, err := queries.GetBookmark(ctx, bookmarkID)
	if err != nil {
		return nil, err
	}

	return &entity.Bookmark{
		BookmarkID:   dbBookmark.BookmarkID,
		URL:          dbBookmark.Url,
		CreationDate: dbBookmark.CreationDate.Time,
	}, nil
}

func (r *BookmarkRepository) ListBookmarks(
	ctx context.Context,
	categoryID *uuid.UUID,
	searchQuery *string,
	startDate *time.Time,
	endDate *time.Time,
	limit, offset int32,
) ([]entity.BookmarkWithTitle, error) {
	queries := db.New(r.pool)

	var categoryIDVal uuid.UUID
	if categoryID != nil {
		categoryIDVal = *categoryID
	}

	var searchQueryVal string
	if searchQuery != nil {
		searchQueryVal = *searchQuery
	}

	var startDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}

	var endDatePg pgtype.Timestamp
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}

	dbBookmarks, err := queries.ListBookmarks(ctx, db.ListBookmarksParams{
		Column1: categoryIDVal,
		Column2: searchQueryVal,
		Column3: startDatePg,
		Column4: endDatePg,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}

	bookmarks := make([]entity.BookmarkWithTitle, len(dbBookmarks))
	for i, dbBookmark := range dbBookmarks {
		bookmarks[i] = entity.BookmarkWithTitle{
			BookmarkID:   dbBookmark.BookmarkID,
			URL:          dbBookmark.Url,
			CreationDate: dbBookmark.CreationDate.Time,
			Title:        dbBookmark.Title,
		}
	}

	return bookmarks, nil
}

func (r *BookmarkRepository) CountBookmarks(
	ctx context.Context,
	categoryID *uuid.UUID,
	searchQuery *string,
	startDate *time.Time,
	endDate *time.Time,
) (int64, error) {
	queries := db.New(r.pool)

	var categoryIDVal uuid.UUID
	if categoryID != nil {
		categoryIDVal = *categoryID
	}

	var searchQueryVal string
	if searchQuery != nil {
		searchQueryVal = *searchQuery
	}

	var startDatePg pgtype.Timestamp
	if startDate != nil {
		startDatePg = pgtype.Timestamp{Time: *startDate, Valid: true}
	}

	var endDatePg pgtype.Timestamp
	if endDate != nil {
		endDatePg = pgtype.Timestamp{Time: *endDate, Valid: true}
	}

	count, err := queries.CountBookmarks(ctx, db.CountBookmarksParams{
		Column1: categoryIDVal,
		Column2: searchQueryVal,
		Column3: startDatePg,
		Column4: endDatePg,
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *BookmarkRepository) GetRandomBookmark(ctx context.Context) (uuid.UUID, error) {
	queries := db.New(r.pool)
	id, err := queries.GetRandomBookmark(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *BookmarkRepository) GetBookmarkDetails(ctx context.Context, bookmarkID uuid.UUID) (*entity.BookmarkDetails, error) {
	queries := db.New(r.pool)
	dbDetails, err := queries.GetBookmarkDetails(ctx, bookmarkID)
	if err != nil {
		return nil, err
	}

	var rawSource *string
	if len(dbDetails.RawSource) > 0 {
		s := string(dbDetails.RawSource)
		rawSource = &s
	}

	return &entity.BookmarkDetails{
		BookmarkID:    dbDetails.BookmarkID,
		URL:           dbDetails.Url,
		CreationDate:  dbDetails.CreationDate.Time,
		CategoryName:  dbDetails.CategoryName,
		SourceURI:     dbDetails.SourceUri,
		RawSource:     rawSource,
		Title:         dbDetails.Title,
		LynxContent:   dbDetails.LynxContent,
		ReaderContent: dbDetails.ReaderContent,
		Summary:       dbDetails.Summary,
		StatusCode:    dbDetails.StatusCode,
		Headers:       dbDetails.Headers,
		HTTPContent:   dbDetails.HttpContent,
		FetchDate:     &dbDetails.FetchDate.Time,
	}, nil
}

func (r *BookmarkRepository) GetBookmarkQuestions(ctx context.Context, bookmarkID uuid.UUID) ([]entity.BookmarkQuestion, error) {
	queries := db.New(r.pool)

	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}
	dbQuestions, err := queries.GetBookmarkQuestions(ctx, bookmarkIDPg)
	if err != nil {
		return nil, err
	}

	questions := make([]entity.BookmarkQuestion, len(dbQuestions))
	for i, dbQ := range dbQuestions {
		content := ""
		if dbQ.Content != nil {
			content = *dbQ.Content
		}
		questions[i] = entity.BookmarkQuestion{
			ID:      dbQ.ID,
			Content: content,
		}
	}

	return questions, nil
}

func (r *BookmarkRepository) SearchSimilarBookmarks(
	ctx context.Context,
	embedding []float32,
	strategy string,
	limit int32,
) ([]entity.BookmarkWithTitle, error) {
	queries := db.New(r.pool)

	// Convert embedding to pgvector.Vector
	embeddingVec := pgvector.NewVector(embedding)

	dbBookmarks, err := queries.SearchSimilarBookmarks(ctx, db.SearchSimilarBookmarksParams{
		Strategy: &strategy,
		Column2:  &embeddingVec,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	bookmarks := make([]entity.BookmarkWithTitle, len(dbBookmarks))
	for i, dbBookmark := range dbBookmarks {
		bookmarks[i] = entity.BookmarkWithTitle{
			BookmarkID:   dbBookmark.BookmarkID,
			URL:          dbBookmark.Url,
			CreationDate: dbBookmark.CreationDate.Time,
			Title:        dbBookmark.Title,
			Summary:      dbBookmark.Summary,
		}
	}

	return bookmarks, nil
}

func (r *BookmarkRepository) UpdateBookmarkQuestion(
	ctx context.Context,
	content string,
	embedding []float32,
	referenceID, bookmarkID uuid.UUID,
) error {
	queries := db.New(r.pool)

	// Convert embedding to pgvector.Vector
	embeddingVec := pgvector.NewVector(embedding)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}

	return queries.UpdateBookmarkQuestion(ctx, db.UpdateBookmarkQuestionParams{
		Content:    &content,
		Column2:    &embeddingVec,
		ID:         referenceID,
		BookmarkID: bookmarkIDPg,
	})
}

func (r *BookmarkRepository) DeleteBookmarkQuestion(ctx context.Context, referenceID, bookmarkID uuid.UUID) error {
	queries := db.New(r.pool)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}
	return queries.DeleteBookmarkQuestion(ctx, db.DeleteBookmarkQuestionParams{
		ID:         referenceID,
		BookmarkID: bookmarkIDPg,
	})
}

func (r *BookmarkRepository) GetBookmarkForQuestion(ctx context.Context, bookmarkID uuid.UUID) (*entity.BookmarkDetails, error) {
	queries := db.New(r.pool)
	dbBookmark, err := queries.GetBookmarkForQuestion(ctx, bookmarkID)
	if err != nil {
		return nil, err
	}

	return &entity.BookmarkDetails{
		BookmarkID: dbBookmark.BookmarkID,
		Title:      dbBookmark.Title,
		Summary:    dbBookmark.Summary,
	}, nil
}

func (r *BookmarkRepository) CreateObservation(ctx context.Context, data, observationType, source, tags, ref string) error {
	queries := db.New(r.pool)

	refUUID, err := uuid.Parse(ref)
	if err != nil {
		return err
	}
	refPg := pgtype.UUID{Bytes: refUUID, Valid: true}

	return queries.CreateObservation(ctx, db.CreateObservationParams{
		Data:   []byte(data),
		Type:   &observationType,
		Source: &source,
		Tags:   &tags,
		Ref:    refPg,
	})
}

func (r *BookmarkRepository) InsertHttpResponse(
	ctx context.Context,
	bookmarkID uuid.UUID,
	statusCode int32,
	headers string,
	content []byte,
	fetchDate time.Time,
) error {
	queries := db.New(r.pool)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}
	return queries.InsertHttpResponse(ctx, db.InsertHttpResponseParams{
		BookmarkID: bookmarkIDPg,
		StatusCode: &statusCode,
		Headers:    &headers,
		Content:    content,
		FetchDate:  pgtype.Timestamp{Time: fetchDate, Valid: true},
	})
}

func (r *BookmarkRepository) GetLatestHttpResponse(ctx context.Context, bookmarkID uuid.UUID) (*output.HTTPResponse, error) {
	queries := db.New(r.pool)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}
	dbResp, err := queries.GetLatestHttpResponse(ctx, bookmarkIDPg)
	if err != nil {
		return nil, err
	}

	var statusCode int32
	if dbResp.StatusCode != nil {
		statusCode = *dbResp.StatusCode
	}

	var headers string
	if dbResp.Headers != nil {
		headers = *dbResp.Headers
	}

	return &output.HTTPResponse{
		ResponseID: dbResp.ResponseID,
		StatusCode: statusCode,
		Headers:    headers,
		Content:    dbResp.Content,
		FetchDate:  dbResp.FetchDate.Time,
	}, nil
}

func (r *BookmarkRepository) InsertProcessedContent(
	ctx context.Context,
	bookmarkID uuid.UUID,
	strategyUsed, processedContent string,
) error {
	queries := db.New(r.pool)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}
	return queries.InsertProcessedContent(ctx, db.InsertProcessedContentParams{
		BookmarkID:       bookmarkIDPg,
		StrategyUsed:     &strategyUsed,
		ProcessedContent: &processedContent,
	})
}

func (r *BookmarkRepository) GetProcessedContentByStrategy(
	ctx context.Context,
	bookmarkID uuid.UUID,
	strategy string,
) (*string, error) {
	queries := db.New(r.pool)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}
	dbContent, err := queries.GetProcessedContentByStrategy(ctx, db.GetProcessedContentByStrategyParams{
		BookmarkID:   bookmarkIDPg,
		StrategyUsed: &strategy,
	})
	if err != nil {
		return nil, err
	}

	return dbContent.ProcessedContent, nil
}

func (r *BookmarkRepository) CreateEmbeddingChunk(
	ctx context.Context,
	bookmarkID uuid.UUID,
	content, strategy string,
	embedding []float32,
) (uuid.UUID, error) {
	queries := db.New(r.pool)

	// Convert embedding to pgvector.Vector
	embeddingVec := pgvector.NewVector(embedding)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}

	id, err := queries.CreateEmbeddingChunk(ctx, db.CreateEmbeddingChunkParams{
		BookmarkID: bookmarkIDPg,
		Content:    &content,
		Strategy:   &strategy,
		Column4:    &embeddingVec,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *BookmarkRepository) GetBookmarkTitle(ctx context.Context, bookmarkID uuid.UUID) (*output.TitleData, error) {
	queries := db.New(r.pool)
	dbTitle, err := queries.GetBookmarkTitle(ctx, bookmarkID)
	if err != nil {
		return nil, err
	}

	// Convert reader title: empty string means no title extracted
	var readerTitle *string
	if dbTitle.ReaderTitle != "" {
		readerTitle = &dbTitle.ReaderTitle
	}

	return &output.TitleData{
		BookmarkID:    dbTitle.BookmarkID,
		URL:           dbTitle.Url,
		CreationDate:  dbTitle.CreationDate.Time,
		ExistingTitle: dbTitle.ExistingTitle,
		RawContent:    dbTitle.RawContent,
		ReaderTitle:   readerTitle,
	}, nil
}

func (r *BookmarkRepository) InsertBookmarkTitle(ctx context.Context, bookmarkID uuid.UUID, title, source string) error {
	queries := db.New(r.pool)
	bookmarkIDPg := pgtype.UUID{Bytes: bookmarkID, Valid: true}
	return queries.InsertBookmarkTitle(ctx, db.InsertBookmarkTitleParams{
		BookmarkID: bookmarkIDPg,
		Title:      &title,
		Source:     &source,
	})
}

func (r *BookmarkRepository) GetMissingHttpResponses(ctx context.Context) ([]entity.Bookmark, error) {
	queries := db.New(r.pool)
	dbBookmarks, err := queries.GetMissingHttpResponses(ctx)
	if err != nil {
		return nil, err
	}

	bookmarks := make([]entity.Bookmark, len(dbBookmarks))
	for i, dbBookmark := range dbBookmarks {
		bookmarks[i] = entity.Bookmark{
			BookmarkID:   dbBookmark.BookmarkID,
			URL:          dbBookmark.Url,
			CreationDate: dbBookmark.CreationDate.Time,
		}
	}

	return bookmarks, nil
}

func (r *BookmarkRepository) GetMissingReaderContent(ctx context.Context) ([]entity.Bookmark, error) {
	queries := db.New(r.pool)
	dbBookmarks, err := queries.GetMissingReaderContent(ctx)
	if err != nil {
		return nil, err
	}

	bookmarks := make([]entity.Bookmark, len(dbBookmarks))
	for i, dbBookmark := range dbBookmarks {
		bookmarks[i] = entity.Bookmark{
			BookmarkID:   dbBookmark.BookmarkID,
			URL:          dbBookmark.Url,
			CreationDate: dbBookmark.CreationDate.Time,
		}
	}

	return bookmarks, nil
}

// Helper function to convert float32 slice to pgvector format
func embeddingToString(embedding []float32) string {
	// This will be replaced by proper pgvector handling in sqlc
	// For now, return a placeholder
	return ""
}
