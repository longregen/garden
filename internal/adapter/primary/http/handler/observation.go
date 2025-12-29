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

type ObservationHandler struct {
	useCase input.ObservationUseCase
}

func NewObservationHandler(useCase input.ObservationUseCase) *ObservationHandler {
	return &ObservationHandler{
		useCase: useCase,
	}
}

func (h *ObservationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/observations", func(r chi.Router) {
		r.Post("/feedback", h.StoreFeedback)
	})
	r.Route("/api/bookmarks/{id}/feedback", func(r chi.Router) {
		r.Get("/", h.GetFeedbackStats)
	})
}

// StoreFeedback godoc
// @Summary Store Q&A feedback
// @Description Store feedback for Q&A and optionally delete the content reference
// @Tags observations
// @Param input body entity.StoreFeedbackInput true "Feedback input"
// @Success 201 {object} entity.Observation
// @Router /api/observations/feedback [post]
func (h *ObservationHandler) StoreFeedback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input entity.StoreFeedbackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	observation, err := h.useCase.StoreFeedback(ctx, input)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusCreated, observation)
}

// GetFeedbackStats godoc
// @Summary Get feedback statistics
// @Description Get aggregated feedback statistics for a bookmark
// @Tags observations
// @Param id path string true "Bookmark ID (UUID)"
// @Success 200 {object} entity.FeedbackStats
// @Router /api/bookmarks/{id}/feedback [get]
func (h *ObservationHandler) GetFeedbackStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	bookmarkID, err := uuid.Parse(idStr)
	if err != nil {
		httpAdapter.BadRequest(w, err)
		return
	}

	stats, err := h.useCase.GetFeedbackStats(ctx, bookmarkID)
	if err != nil {
		httpAdapter.InternalError(w, err)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, stats)
}
