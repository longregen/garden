package handler

import (
	"net/http"
	"strconv"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/port/input"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TagHandler struct {
	useCase input.TagUseCase
}

func NewTagHandler(useCase input.TagUseCase) *TagHandler {
	return &TagHandler{
		useCase: useCase,
	}
}

func (h *TagHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/tags", func(r chi.Router) {
		r.Get("/", h.ListTags)
		r.Get("/{name}", h.GetTag)
		r.Get("/{name}/items", h.GetItemsByTag)
		r.Delete("/{name}", h.DeleteTag)
		r.Post("/{itemId}/{tagName}", h.AddTag)
		r.Delete("/{itemId}/{tagName}", h.RemoveTag)
	})
}

// TagResponse represents the API response for a single tag
type TagResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Created      int64  `json:"created"`
	Modified     int64  `json:"modified"`
	LastActivity *int64 `json:"lastActivity,omitempty"`
	UsageCount   *int64 `json:"usageCount,omitempty"`
}

// TagsListResponse represents the API response for tag list
type TagsListResponse struct {
	Tags []TagResponse `json:"tags"`
}

// GetTag godoc
// @Summary Get tag by name
// @Description Get a single tag
// @Tags tags
// @Param name path string true "Tag Name"
// @Success 200 {object} TagResponse
// @Router /api/tags/{name} [get]
func (h *TagHandler) GetTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tagName := chi.URLParam(r, "name")

	tag, err := h.useCase.GetTag(ctx, tagName)
	if err != nil {
		httpAdapter.JSON(w, http.StatusNotFound, map[string]string{
			"message": "Tag not found",
		})
		return
	}

	response := TagResponse{
		ID:           tag.ID.String(),
		Name:         tag.Name,
		Created:      tag.Created,
		Modified:     tag.Modified,
		LastActivity: tag.LastActivity,
	}

	httpAdapter.JSON(w, http.StatusOK, response)
}

// ListTags godoc
// @Summary List all tags
// @Description Get all tags, optionally with usage counts
// @Tags tags
// @Param includeUsage query boolean false "Include usage counts"
// @Success 200 {object} TagsListResponse
// @Router /api/tags [get]
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	includeUsageStr := r.URL.Query().Get("includeUsage")
	includeUsage := includeUsageStr == "true"

	tags, err := h.useCase.ListAllTags(ctx, includeUsage)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to list tags",
		})
		return
	}

	responses := make([]TagResponse, 0, len(tags))
	for _, tag := range tags {
		response := TagResponse{
			ID:           tag.ID.String(),
			Name:         tag.Name,
			Created:      tag.Created,
			Modified:     tag.Modified,
			LastActivity: tag.LastActivity,
		}
		if includeUsage {
			response.UsageCount = &tag.UsageCount
		}
		responses = append(responses, response)
	}

	httpAdapter.JSON(w, http.StatusOK, TagsListResponse{Tags: responses})
}

// GetItemsByTag godoc
// @Summary Get items by tag
// @Description Get all items with a specific tag
// @Tags tags
// @Param name path string true "Tag Name"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} ItemsListResponse
// @Router /api/tags/{name}/items [get]
func (h *TagHandler) GetItemsByTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tagName := chi.URLParam(r, "name")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := int32(1)
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	limit := int32(20)
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	result, err := h.useCase.GetItemsByTag(ctx, tagName, page, limit)
	if err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to get items by tag",
		})
		return
	}

	items := make([]ItemListItemResponse, 0, len(result.Data))
	for _, item := range result.Data {
		items = append(items, ItemListItemResponse{
			ID:       item.ID.String(),
			Title:    item.Title,
			Tags:     item.Tags,
			Created:  item.Created,
			Modified: item.Modified,
		})
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

// AddTagRequest represents the request body for adding a tag
type AddTagRequest struct {
	ItemID  string `json:"itemId"`
	TagName string `json:"tagName"`
}

// AddTag godoc
// @Summary Add tag to item
// @Description Add a tag to an item
// @Tags tags
// @Param itemId path string true "Item ID"
// @Param tagName path string true "Tag Name"
// @Success 200 {object} map[string]string
// @Router /api/tags/{itemId}/{tagName} [post]
func (h *TagHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemId")
	tagName := chi.URLParam(r, "tagName")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	if err := h.useCase.AddTag(ctx, itemID, tagName); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to add tag",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Tag added successfully",
	})
}

// RemoveTag godoc
// @Summary Remove tag from item
// @Description Remove a tag from an item
// @Tags tags
// @Param itemId path string true "Item ID"
// @Param tagName path string true "Tag Name"
// @Success 200 {object} map[string]string
// @Router /api/tags/{itemId}/{tagName} [delete]
func (h *TagHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemIDStr := chi.URLParam(r, "itemId")
	tagName := chi.URLParam(r, "tagName")

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		httpAdapter.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid item ID",
		})
		return
	}

	if err := h.useCase.RemoveTag(ctx, itemID, tagName); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to remove tag",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Tag removed successfully",
	})
}

// DeleteTag godoc
// @Summary Delete tag
// @Description Delete a tag and all its associations
// @Tags tags
// @Param name path string true "Tag Name"
// @Success 200 {object} map[string]string
// @Router /api/tags/{name} [delete]
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tagName := chi.URLParam(r, "name")

	if err := h.useCase.DeleteTag(ctx, tagName); err != nil {
		httpAdapter.JSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Failed to delete tag",
		})
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{
		"message": "Tag deleted successfully",
	})
}
