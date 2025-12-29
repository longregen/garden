package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
)

type SocialPostHandler struct {
	useCase input.SocialPostUseCase
}

func NewSocialPostHandler(useCase input.SocialPostUseCase) *SocialPostHandler {
	return &SocialPostHandler{
		useCase: useCase,
	}
}

func (h *SocialPostHandler) RegisterRoutes(r chi.Router) {
	// New /api/social routes
	r.Route("/api/social", func(r chi.Router) {
		r.Get("/posts", h.ListPosts)
		r.Get("/posts/{id}", h.GetPost)
		r.Post("/posts", h.CreatePost)
		r.Put("/posts/{id}/status", h.UpdateStatus)
		r.Delete("/posts/{id}", h.DeletePost)
		r.Get("/credentials", h.CheckCredentials)
		r.Put("/twitter/tokens", h.UpdateTwitterTokens)
		r.Post("/twitter/auth", h.InitiateTwitterAuth)
		r.Post("/twitter/callback", h.HandleTwitterCallback)
	})

	// Backwards-compatible /api/microlog aliases
	r.Route("/api/microlog", func(r chi.Router) {
		r.Get("/", h.ListPosts)
		r.Post("/", h.CreatePost)
		r.Get("/{id}", h.GetPost)
		r.Put("/{id}", h.UpdateStatus)
		r.Delete("/{id}", h.DeletePost)
		r.Get("/status", h.CheckCredentials)
		r.Post("/twitter-auth", h.InitiateTwitterAuth)
		r.Post("/twitter-callback", h.HandleTwitterCallback)
	})
}

// ListPosts godoc
// @Summary List social posts
// @Description Get paginated social posts with optional status filter
// @Tags social
// @Param status query string false "Filter by status (pending, completed, partial, failed)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} input.PaginatedResponse[entity.SocialPost]
// @Router /api/social/posts [get]
func (h *SocialPostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters := entity.SocialPostFilters{
		Page:  1,
		Limit: 10,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.ParseInt(pageStr, 10, 32)
		if err != nil {
			httpAdapter.BadRequest(w, err)
			return
		}
		filters.Page = int32(page)
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			httpAdapter.BadRequest(w, err)
			return
		}
		filters.Limit = int32(limit)
	}

	posts, err := h.useCase.ListPosts(ctx, filters)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, posts)
}

// GetPost godoc
// @Summary Get social post by ID
// @Description Get a single social post by its ID
// @Tags social
// @Param id path string true "Post ID (UUID)"
// @Success 200 {object} entity.SocialPost
// @Router /api/social/posts/{id} [get]
func (h *SocialPostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	postID, err := uuid.Parse(idStr)
	if err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	post, err := h.useCase.GetPost(ctx, postID)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}
	if post == nil {
		httpAdapter.NotFound(w)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, post)
}

// CreatePost godoc
// @Summary Create and post to social media
// @Description Create a new social post and publish to Twitter and Bluesky
// @Tags social
// @Param input body entity.CreateSocialPostInput true "Post content"
// @Success 201 {object} entity.PostResult
// @Router /api/social/posts [post]
func (h *SocialPostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input entity.CreateSocialPostInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	result, err := h.useCase.CreatePost(ctx, input)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusCreated, result)
}

// UpdateStatus godoc
// @Summary Update post status
// @Description Update the status of a social post
// @Tags social
// @Param id path string true "Post ID (UUID)"
// @Param input body entity.UpdateStatusInput true "Status update"
// @Success 200
// @Router /api/social/posts/{id}/status [put]
func (h *SocialPostHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	postID, err := uuid.Parse(idStr)
	if err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	var input entity.UpdateStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	if err := h.useCase.UpdateStatus(ctx, postID, input); err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, nil)
}

// DeletePost godoc
// @Summary Delete social post
// @Description Delete a social post by ID
// @Tags social
// @Param id path string true "Post ID (UUID)"
// @Success 204
// @Router /api/social/posts/{id} [delete]
func (h *SocialPostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	postID, err := uuid.Parse(idStr)
	if err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	if err := h.useCase.DeletePost(ctx, postID); err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusNoContent, nil)
}

// CheckCredentials godoc
// @Summary Check social media credentials
// @Description Verify Twitter and Bluesky credentials are valid
// @Tags social
// @Success 200 {object} entity.CredentialsStatus
// @Router /api/social/credentials [get]
func (h *SocialPostHandler) CheckCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status, err := h.useCase.CheckCredentials(ctx)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, status)
}

// UpdateTwitterTokens godoc
// @Summary Update Twitter OAuth tokens
// @Description Store Twitter OAuth tokens from authorization flow
// @Tags social
// @Param tokens body entity.TwitterTokens true "OAuth tokens"
// @Success 200
// @Router /api/social/twitter/tokens [put]
func (h *SocialPostHandler) UpdateTwitterTokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var tokens entity.TwitterTokens
	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	if err := h.useCase.UpdateTwitterTokens(ctx, tokens); err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, nil)
}

// InitiateTwitterAuth godoc
// @Summary Initiate Twitter OAuth flow
// @Description Generate Twitter OAuth authorization URL
// @Tags social
// @Success 200 {object} entity.TwitterAuthURL
// @Router /api/social/twitter/auth [post]
func (h *SocialPostHandler) InitiateTwitterAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authURL, err := h.useCase.InitiateTwitterAuth(ctx)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, authURL)
}

// HandleTwitterCallback godoc
// @Summary Handle Twitter OAuth callback
// @Description Process OAuth callback and exchange code for tokens
// @Tags social
// @Param callback body entity.TwitterCallbackInput true "Callback parameters"
// @Success 200
// @Router /api/social/twitter/callback [post]
func (h *SocialPostHandler) HandleTwitterCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input entity.TwitterCallbackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	if err := h.useCase.HandleTwitterCallback(ctx, input); err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}
