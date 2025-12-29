package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
)

type BookmarkHandler struct {
	useCase input.BookmarkUseCase
}

func NewBookmarkHandler(useCase input.BookmarkUseCase) *BookmarkHandler {
	return &BookmarkHandler{
		useCase: useCase,
	}
}

func (h *BookmarkHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/bookmarks", func(r chi.Router) {
		r.Get("/", h.ListBookmarks)
		r.Get("/random", h.RandomBookmark)
		r.Get("/search", h.SearchBookmarks)
		r.Get("/missing/http", h.MissingHttp)
		r.Get("/missing/reader", h.MissingReader)

		// Backwards-compatible aliases for missing endpoints
		r.Get("/missing-http", h.MissingHttp)
		r.Get("/missing-reader", h.MissingReader)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetBookmark)
			r.Put("/question", h.UpdateQuestion)
			r.Delete("/question/{refId}", h.DeleteQuestion)
			r.Post("/fetch", h.FetchContent)
			r.Post("/process/lynx", h.ProcessLynx)
			r.Post("/process/reader", h.ProcessReader)
			r.Post("/embeddings", h.CreateEmbeddings)
			r.Post("/summary-embedding", h.CreateSummary)
			r.Get("/title", h.GetTitle)

			// Backwards-compatible aliases for bookmark-specific endpoints
			r.Put("/update-question", h.UpdateQuestion)
			r.Delete("/delete-question", h.DeleteQuestion)
			r.Post("/lynx", h.ProcessLynx)
			r.Post("/reader", h.ProcessReader)
			r.Post("/embed-chunks", h.CreateEmbeddings)
			r.Post("/embed-summary", h.CreateSummary)
		})
	})
}

// ListBookmarks godoc
// @Summary List bookmarks
// @Description Get filtered and paginated bookmarks
// @Tags bookmarks
// @Param categoryId query string false "Category ID"
// @Param searchQuery query string false "Search query"
// @Param startCreationDate query string false "Start creation date"
// @Param endCreationDate query string false "End creation date"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} input.PaginatedResponse[entity.BookmarkWithTitle]
// @Router /api/bookmarks [get]
func (h *BookmarkHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters := entity.BookmarkFilters{
		Page:  1,
		Limit: 10,
	}

	if categoryIDStr := r.URL.Query().Get("categoryId"); categoryIDStr != "" {
		categoryID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		filters.CategoryID = &categoryID
	}

	if searchQuery := r.URL.Query().Get("searchQuery"); searchQuery != "" {
		filters.SearchQuery = &searchQuery
	}

	if startDateStr := r.URL.Query().Get("startCreationDate"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			http.Error(w, "Invalid start creation date", http.StatusBadRequest)
			return
		}
		filters.StartCreationDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("endCreationDate"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			http.Error(w, "Invalid end creation date", http.StatusBadRequest)
			return
		}
		filters.EndCreationDate = &endDate
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			http.Error(w, "Invalid page", http.StatusBadRequest)
			return
		}
		filters.Page = int32(page)
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		filters.Limit = int32(limit)
	}

	result, err := h.useCase.ListBookmarks(ctx, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RandomBookmark godoc
// @Summary Get random bookmark
// @Description Redirect to a random bookmark
// @Tags bookmarks
// @Success 302
// @Router /api/bookmarks/random [get]
func (h *BookmarkHandler) RandomBookmark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bookmarkID, err := h.useCase.GetRandomBookmark(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/api/bookmarks/"+bookmarkID.String(), http.StatusFound)
}

// GetBookmark godoc
// @Summary Get bookmark details
// @Description Get complete bookmark details with all relations
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Success 200 {object} entity.BookmarkDetails
// @Router /api/bookmarks/{id} [get]
func (h *BookmarkHandler) GetBookmark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	details, err := h.useCase.GetBookmarkDetails(ctx, bookmarkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// SearchBookmarks godoc
// @Summary Search bookmarks
// @Description Perform vector similarity search on bookmarks
// @Tags bookmarks
// @Param query query string true "Search query"
// @Param strategy query string false "Search strategy" default(qa-v2-passage)
// @Success 200 {array} entity.BookmarkWithTitle
// @Router /api/bookmarks/search [get]
func (h *BookmarkHandler) SearchBookmarks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	strategy := r.URL.Query().Get("strategy")
	if strategy == "" {
		strategy = "qa-v2-passage"
	}

	results, err := h.useCase.SearchSimilarBookmarks(ctx, query, strategy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// UpdateQuestion godoc
// @Summary Update bookmark Q&A
// @Description Update a question and answer content reference
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Param input body entity.UpdateQuestionInput true "Update input"
// @Success 200 {object} map[string]string
// @Router /api/bookmarks/{id}/question [put]
func (h *BookmarkHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	var input entity.UpdateQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input.BookmarkID = bookmarkID

	if err := h.useCase.UpdateBookmarkQuestion(ctx, input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Question updated successfully",
	})
}

// DeleteQuestion godoc
// @Summary Delete bookmark Q&A
// @Description Delete a question and answer content reference
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Param refId path string true "Reference ID"
// @Param input body entity.DeleteQuestionInput true "Delete input"
// @Success 200 {object} map[string]string
// @Router /api/bookmarks/{id}/question/{refId} [delete]
func (h *BookmarkHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")
	refIDStr := chi.URLParam(r, "refId")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	refID, err := uuid.Parse(refIDStr)
	if err != nil {
		http.Error(w, "Invalid reference ID", http.StatusBadRequest)
		return
	}

	var input entity.DeleteQuestionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input.BookmarkID = bookmarkID
	input.ReferenceID = refID

	if err := h.useCase.DeleteBookmarkQuestion(ctx, input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Question deleted successfully",
	})
}

// FetchContent godoc
// @Summary Fetch bookmark content
// @Description Fetch and store HTTP content for a bookmark
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Success 200 {object} entity.FetchResult
// @Router /api/bookmarks/{id}/fetch [post]
func (h *BookmarkHandler) FetchContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	result, err := h.useCase.FetchBookmarkContent(ctx, bookmarkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ProcessLynx godoc
// @Summary Process with Lynx
// @Description Process bookmark content using lynx
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Success 200 {object} entity.ProcessingResult
// @Router /api/bookmarks/{id}/process/lynx [post]
func (h *BookmarkHandler) ProcessLynx(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	result, err := h.useCase.ProcessWithLynx(ctx, bookmarkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ProcessReader godoc
// @Summary Process with Reader
// @Description Process bookmark content using reader mode
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Success 200 {object} entity.ProcessingResult
// @Router /api/bookmarks/{id}/process/reader [post]
func (h *BookmarkHandler) ProcessReader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	result, err := h.useCase.ProcessWithReader(ctx, bookmarkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CreateEmbeddings godoc
// @Summary Create embeddings
// @Description Create chunked embeddings for bookmark content
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Success 200 {object} entity.EmbeddingResult
// @Router /api/bookmarks/{id}/embeddings [post]
func (h *BookmarkHandler) CreateEmbeddings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	result, err := h.useCase.CreateEmbeddingChunks(ctx, bookmarkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CreateSummary godoc
// @Summary Create summary embedding
// @Description Create a summary and its embedding
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Success 200 {object} entity.SummaryEmbeddingResult
// @Router /api/bookmarks/{id}/summary-embedding [post]
func (h *BookmarkHandler) CreateSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	result, err := h.useCase.CreateSummaryEmbedding(ctx, bookmarkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetTitle godoc
// @Summary Get bookmark title
// @Description Extract and store the bookmark title
// @Tags bookmarks
// @Param id path string true "Bookmark ID"
// @Success 200 {object} entity.TitleExtractionResult
// @Router /api/bookmarks/{id}/title [get]
func (h *BookmarkHandler) GetTitle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bookmarkIDStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(bookmarkIDStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	result, err := h.useCase.GetBookmarkTitle(ctx, bookmarkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// MissingHttp godoc
// @Summary Get bookmarks missing HTTP responses
// @Description Get bookmarks that don't have HTTP responses
// @Tags bookmarks
// @Success 200 {array} entity.Bookmark
// @Router /api/bookmarks/missing/http [get]
func (h *BookmarkHandler) MissingHttp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bookmarks, err := h.useCase.GetMissingHttpResponses(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarks)
}

// MissingReader godoc
// @Summary Get bookmarks missing reader content
// @Description Get bookmarks that don't have reader-processed content
// @Tags bookmarks
// @Success 200 {array} entity.Bookmark
// @Router /api/bookmarks/missing/reader [get]
func (h *BookmarkHandler) MissingReader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bookmarks, err := h.useCase.GetMissingReaderContent(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarks)
}
