package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ItemHandler struct {
	useCase    input.ItemUseCase
	tagUseCase input.TagUseCase
}

func NewItemHandler(useCase input.ItemUseCase, tagUseCase input.TagUseCase) *ItemHandler {
	return &ItemHandler{
		useCase:    useCase,
		tagUseCase: tagUseCase,
	}
}

func (h *ItemHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/items", func(r chi.Router) {
		r.Get("/", h.ListItems)
		r.Post("/", h.CreateItem)
		r.Get("/search", h.SearchItems)
		r.Get("/{id}", h.GetItem)
		r.Put("/{id}", h.UpdateItem)
		r.Delete("/{id}", h.DeleteItem)
		r.Get("/{id}/tags", h.GetItemTags)
		r.Put("/{id}/tags", h.UpdateItemTags)
		// Backwards-compatible routes for legacy tag endpoints
		r.Post("/{id}/tags", h.AddTagLegacy)
		r.Delete("/{id}/tags", h.RemoveTagLegacy)
	})
}

// ItemResponse represents the API response for a single item
type ItemResponse struct {
	ID       string   `json:"id"`
	Title    *string  `json:"title"`
	Contents *string  `json:"contents"`
	Tags     []string `json:"tags"`
	Created  int64    `json:"created"`
	Modified int64    `json:"modified"`
}

// ItemsListResponse represents the API response for item list with pagination
type ItemsListResponse struct {
	Data       []ItemListItemResponse `json:"data"`
	Pagination PaginationResponse     `json:"pagination"`
}

// ItemListItemResponse represents an item in list view
type ItemListItemResponse struct {
	ID       string   `json:"id"`
	Title    *string  `json:"title"`
	Tags     []string `json:"tags"`
	Created  int64    `json:"created"`
	Modified int64    `json:"modified"`
}

// GetItem godoc
// @Summary Get item by ID
// @Description Get a single item with tags
// @Tags items
// @Param id path string true "Item ID"
// @Success 200 {object} ItemResponse
// @Router /api/items/{id} [get]
func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	fullItem, err := h.useCase.GetItem(ctx, itemID)
	if err != nil {
		httpAdapter.JSON(w, http.StatusNotFound, map[string]string{
			"message": "Item not found",
		})
		return
	}

	response := ItemResponse{
		ID:       fullItem.Item.ID.String(),
		Title:    fullItem.Item.Title,
		Contents: fullItem.Item.Contents,
		Tags:     fullItem.Tags,
		Created:  fullItem.Item.Created,
		Modified: fullItem.Item.Modified,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ListItems godoc
// @Summary List items
// @Description Get paginated list of items
// @Tags items
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(10)
// @Success 200 {object} ItemsListResponse
// @Router /api/items [get]
func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := int32(1)
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	limit := int32(10)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = int32(l)
		}
	}

	result, err := h.useCase.ListItems(ctx, page, limit)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Internal Server Error",
			"message": "Failed to list items",
		})
		return
	}

	items := make([]ItemListItemResponse, len(result.Data))
	for i, item := range result.Data {
		items[i] = ItemListItemResponse{
			ID:       item.ID.String(),
			Title:    item.Title,
			Tags:     item.Tags,
			Created:  item.Created,
			Modified: item.Modified,
		}
	}

	response := ItemsListResponse{
		Data: items,
		Pagination: PaginationResponse{
			Page:       result.Page,
			TotalPages: result.TotalPages,
			Limit:      result.PageSize,
			TotalItems: result.Total,
		},
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// CreateItem godoc
// @Summary Create item
// @Description Create a new item with tags
// @Tags items
// @Param item body entity.CreateItemInput true "Item data"
// @Success 201 {object} map[string]interface{}
// @Router /api/items [post]
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input entity.CreateItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation Error",
			"details": []map[string]string{{"message": "Invalid request body"}},
		})
		return
	}

	// Validation
	if input.Title == "" {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation Error",
			"details": []map[string]string{{"message": "Title is required"}},
		})
		return
	}

	if input.Contents == "" {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation Error",
			"details": []map[string]string{{"message": "Contents are required"}},
		})
		return
	}

	if len(input.Tags) == 0 {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation Error",
			"details": []map[string]string{{"message": "At least one tag is required"}},
		})
		return
	}

	fullItem, err := h.useCase.CreateItem(ctx, input)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Internal Server Error",
			"message": "Failed to create item",
		})
		return
	}

	response := map[string]interface{}{
		"data": ItemResponse{
			ID:       fullItem.Item.ID.String(),
			Title:    fullItem.Item.Title,
			Contents: fullItem.Item.Contents,
			Tags:     fullItem.Tags,
			Created:  fullItem.Item.Created,
			Modified: fullItem.Item.Modified,
		},
		"message": "Item created successfully",
	}

	httpAdapter.JSON(w, http.StatusCreated, response)
}

// UpdateItem godoc
// @Summary Update item
// @Description Update an item's information
// @Tags items
// @Param id path string true "Item ID"
// @Param item body entity.UpdateItemInput true "Item data"
// @Success 200 {object} ItemResponse
// @Router /api/items/{id} [put]
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	var input entity.UpdateItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid request body",
		})
		return
	}

	// Validation: at least one field must be provided
	if input.Title == nil && input.Contents == nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Title or contents are required",
		})
		return
	}

	fullItem, err := h.useCase.UpdateItem(ctx, itemID, input)
	if err != nil {
		httpAdapter.JSON(w, http.StatusNotFound, map[string]string{
			"message": "Item not found",
		})
		return
	}

	response := map[string]interface{}{
		"id":       fullItem.Item.ID.String(),
		"title":    fullItem.Item.Title,
		"contents": fullItem.Item.Contents,
		"tags":     fullItem.Tags,
		"created":  fullItem.Item.Created,
		"modified": fullItem.Item.Modified,
		"message":  "Item modified successfully",
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// DeleteItem godoc
// @Summary Delete item
// @Description Delete an item and all related data
// @Tags items
// @Param id path string true "Item ID"
// @Success 200 {object} map[string]string
// @Router /api/items/{id} [delete]
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	if err := h.useCase.DeleteItem(ctx, itemID); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to delete item",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Item deleted successfully",
	})
}

// GetItemTags godoc
// @Summary Get item tags
// @Description Get all tags for a specific item
// @Tags items
// @Param id path string true "Item ID"
// @Success 200 {array} string
// @Router /api/items/{id}/tags [get]
func (h *ItemHandler) GetItemTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	tags, err := h.useCase.GetItemTags(ctx, itemID)
	if err != nil {
		httpAdapter.JSON(w, http.StatusNotFound, map[string]string{
			"message": "Item not found",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, tags)
}

// UpdateItemTags godoc
// @Summary Update item tags
// @Description Update the tags for a specific item
// @Tags items
// @Param id path string true "Item ID"
// @Param tags body []string true "Tags"
// @Success 200 {object} map[string]string
// @Router /api/items/{id}/tags [put]
func (h *ItemHandler) UpdateItemTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	var tags []string
	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid request body",
		})
		return
	}

	if err := h.useCase.UpdateItemTags(ctx, itemID, tags); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to update tags",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Tags updated successfully",
	})
}

// SearchItems godoc
// @Summary Search items
// @Description Perform vector similarity search on items
// @Tags items
// @Param q query string true "Search query (embedding as JSON array)"
// @Success 200 {object} map[string]interface{}
// @Router /api/items/search [get]
func (h *ItemHandler) SearchItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queryEmbeddingStr := r.URL.Query().Get("q")
	if queryEmbeddingStr == "" {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation Error",
			"details": []map[string]string{{"message": "Search query is required"}},
		})
		return
	}

	// Parse the embedding from JSON array
	var queryEmbedding []float32
	if err := json.Unmarshal([]byte(queryEmbeddingStr), &queryEmbedding); err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation Error",
			"details": []map[string]string{{"message": "Invalid embedding format"}},
		})
		return
	}

	results, err := h.useCase.SearchItems(ctx, queryEmbedding)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Internal Server Error",
			"message": "Failed to search items",
		})
		return
	}

	items := make([]ItemListItemResponse, len(results))
	for i, item := range results {
		items[i] = ItemListItemResponse{
			ID:       item.ID.String(),
			Title:    item.Title,
			Tags:     item.Tags,
			Created:  item.Created,
			Modified: item.Modified,
		}
	}

	response := map[string]interface{}{
		"data": items,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// AddTagRequest represents the legacy request body for adding a tag
type AddTagLegacyRequest struct {
	Tag string `json:"tag"`
}

// AddTagLegacy godoc
// @Summary Add tag to item (legacy)
// @Description Backwards-compatible endpoint: POST /api/items/{id}/tags with body { tag: string }
// @Tags items
// @Param id path string true "Item ID"
// @Param body body AddTagLegacyRequest true "Tag name"
// @Success 200 {object} map[string]string
// @Router /api/items/{id}/tags [post]
func (h *ItemHandler) AddTagLegacy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	var req AddTagLegacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid request body",
		})
		return
	}

	if req.Tag == "" {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Tag name is required",
		})
		return
	}

	if err := h.tagUseCase.AddTag(ctx, itemID, req.Tag); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to add tag",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Tag added successfully",
	})
}

// RemoveTagLegacy godoc
// @Summary Remove tag from item (legacy)
// @Description Backwards-compatible endpoint: DELETE /api/items/{id}/tags?tagname={name}
// @Tags items
// @Param id path string true "Item ID"
// @Param tagname query string true "Tag name"
// @Success 200 {object} map[string]string
// @Router /api/items/{id}/tags [delete]
func (h *ItemHandler) RemoveTagLegacy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "id")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	tagName := r.URL.Query().Get("tagname")
	if tagName == "" {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Tag name is required",
		})
		return
	}

	if err := h.tagUseCase.RemoveTag(ctx, itemID, tagName); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to remove tag",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Tag removed successfully",
	})
}
