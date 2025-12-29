package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
)

type CategoryHandler struct {
	useCase input.CategoryUseCase
}

func NewCategoryHandler(useCase input.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{
		useCase: useCase,
	}
}

func (h *CategoryHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/categories", func(r chi.Router) {
		r.Get("/", h.ListCategories)
		r.Post("/merge", h.MergeCategories)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetCategory)
			r.Put("/", h.UpdateCategory)
			r.Post("/sources", h.CreateSource)
		})

		r.Route("/sources/{id}", func(r chi.Router) {
			r.Put("/", h.UpdateSource)
			r.Delete("/", h.DeleteSource)
		})
	})
}

// CategorySourceResponse represents a category source in the API
type CategorySourceResponse struct {
	ID         uuid.UUID       `json:"id"`
	CategoryID uuid.UUID       `json:"category_id"`
	SourceURI  *string         `json:"source_uri"`
	RawSource  json.RawMessage `json:"raw_source"`
}

// CategoryResponse represents a category with sources
type CategoryResponse struct {
	CategoryID uuid.UUID                `json:"category_id"`
	Name       string                   `json:"name"`
	Sources    []CategorySourceResponse `json:"sources"`
}

// SimpleCategoryResponse represents a category without sources
type SimpleCategoryResponse struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
}

// CategoriesListResponse represents the API response for category list with pagination
type CategoriesListResponse struct {
	Data       []CategoryResponse `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int32 `json:"page"`
	TotalPages int32 `json:"totalPages"`
	Limit      int32 `json:"limit"`
	TotalItems int64 `json:"totalItems"`
}

// ListCategories godoc
// @Summary List all categories
// @Description Get all categories with their sources
// @Tags categories
// @Success 200 {object} CategoriesListResponse
// @Router /api/categories [get]
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := h.useCase.ListCategories(ctx)
	if err != nil {
		http.Error(w, "Failed to list categories", http.StatusInternalServerError)
		return
	}

	data := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		sources := make([]CategorySourceResponse, len(cat.Sources))
		for j, src := range cat.Sources {
			sources[j] = CategorySourceResponse{
				ID:         src.ID,
				CategoryID: src.CategoryID,
				SourceURI:  src.SourceURI,
				RawSource:  src.RawSource,
			}
		}

		data[i] = CategoryResponse{
			CategoryID: cat.Category.CategoryID,
			Name:       cat.Category.Name,
			Sources:    sources,
		}
	}

	totalItems := int64(len(categories))
	response := CategoriesListResponse{
		Data: data,
		Pagination: PaginationResponse{
			Page:       1,
			TotalPages: 1,
			Limit:      int32(totalItems),
			TotalItems: totalItems,
		},
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Get a single category by ID
// @Tags categories
// @Param id path string true "Category ID"
// @Success 200 {object} SimpleCategoryResponse
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryIDStr := chi.URLParam(r, "id")

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := h.useCase.GetCategory(ctx, categoryID)
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	response := SimpleCategoryResponse{
		CategoryID: category.CategoryID,
		Name:       category.Name,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update a category's name
// @Tags categories
// @Param id path string true "Category ID"
// @Param body body UpdateCategoryRequest true "Category update data"
// @Success 204
// @Router /api/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryIDStr := chi.URLParam(r, "id")

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.useCase.UpdateCategory(ctx, categoryID, req.Name); err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MergeCategoriesRequest represents the request body for merging categories
type MergeCategoriesRequest struct {
	SourceID uuid.UUID `json:"source_id"`
	TargetID uuid.UUID `json:"target_id"`
}

// MergeCategories godoc
// @Summary Merge categories
// @Description Merge two categories together
// @Tags categories
// @Param body body MergeCategoriesRequest true "Merge request data"
// @Success 204
// @Router /api/categories/merge [post]
func (h *CategoryHandler) MergeCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req MergeCategoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input := entity.MergeCategoriesInput{
		SourceID: req.SourceID,
		TargetID: req.TargetID,
	}

	if err := h.useCase.MergeCategories(ctx, input); err != nil {
		http.Error(w, "Failed to merge categories", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateSourceRequest represents the request body for creating a category source
type CreateSourceRequest struct {
	SourceURI *string         `json:"source_uri"`
	RawSource json.RawMessage `json:"raw_source"`
}

// CreateSource godoc
// @Summary Create category source
// @Description Add a source to a category
// @Tags categories
// @Param id path string true "Category ID"
// @Param body body CreateSourceRequest true "Source data"
// @Success 201 {object} CategorySourceResponse
// @Router /api/categories/{id}/sources [post]
func (h *CategoryHandler) CreateSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	categoryIDStr := chi.URLParam(r, "id")

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req CreateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input := entity.CreateCategorySourceInput{
		CategoryID: categoryID,
		SourceURI:  req.SourceURI,
		RawSource:  req.RawSource,
	}

	source, err := h.useCase.CreateCategorySource(ctx, input)
	if err != nil {
		http.Error(w, "Failed to create category source", http.StatusInternalServerError)
		return
	}

	response := CategorySourceResponse{
		ID:         source.ID,
		CategoryID: source.CategoryID,
		SourceURI:  source.SourceURI,
		RawSource:  source.RawSource,
	}

	httpAdapter.JSON(w, http.StatusCreated, response)
}

// UpdateSourceRequest represents the request body for updating a category source
type UpdateSourceRequest struct {
	SourceURI *string          `json:"source_uri"`
	RawSource *json.RawMessage `json:"raw_source"`
}

// UpdateSource godoc
// @Summary Update category source
// @Description Update a category source
// @Tags categories
// @Param id path string true "Source ID"
// @Param body body UpdateSourceRequest true "Source update data"
// @Success 204
// @Router /api/categories/sources/{id} [put]
func (h *CategoryHandler) UpdateSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceIDStr := chi.URLParam(r, "id")

	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	var req UpdateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	input := entity.UpdateCategorySourceInput{
		SourceURI: req.SourceURI,
		RawSource: req.RawSource,
	}

	if err := h.useCase.UpdateCategorySource(ctx, sourceID, input); err != nil {
		http.Error(w, "Failed to update category source", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteSource godoc
// @Summary Delete category source
// @Description Delete a category source
// @Tags categories
// @Param id path string true "Source ID"
// @Success 204
// @Router /api/categories/sources/{id} [delete]
func (h *CategoryHandler) DeleteSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceIDStr := chi.URLParam(r, "id")

	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	if err := h.useCase.DeleteCategorySource(ctx, sourceID); err != nil {
		http.Error(w, "Failed to delete category source", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
