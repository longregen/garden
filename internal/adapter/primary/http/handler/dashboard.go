package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"garden3/internal/port/input"
)

type DashboardHandler struct {
	useCase input.DashboardUseCase
}

func NewDashboardHandler(useCase input.DashboardUseCase) *DashboardHandler {
	return &DashboardHandler{
		useCase: useCase,
	}
}

func (h *DashboardHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/dashboard", func(r chi.Router) {
		r.Get("/stats", h.GetStats)
	})
}

// GetStats godoc
// @Summary Get dashboard statistics
// @Description Get comprehensive statistics for contacts, sessions, bookmarks, browser history, and recent items
// @Tags dashboard
// @Success 200 {object} entity.DashboardStats
// @Router /api/dashboard/stats [get]
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.useCase.GetStats(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
