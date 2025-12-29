package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
)

type BrowserHistoryHandler struct {
	useCase input.BrowserHistoryUseCase
}

func NewBrowserHistoryHandler(useCase input.BrowserHistoryUseCase) *BrowserHistoryHandler {
	return &BrowserHistoryHandler{
		useCase: useCase,
	}
}

func (h *BrowserHistoryHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/history", func(r chi.Router) {
		r.Get("/", h.ListHistory)
		r.Get("/domains", h.TopDomains)
		r.Get("/recent", h.RecentHistory)
	})
}

// ListHistory godoc
// @Summary List browser history
// @Description Get filtered and paginated browser history
// @Tags browser-history
// @Param q query string false "Search query"
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param domain query string false "Filter by domain"
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 10)"
// @Success 200 {object} input.PaginatedResponse[entity.BrowserHistory]
// @Router /api/history [get]
func (h *BrowserHistoryHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filters := entity.BrowserHistoryFilters{
		Page:     1,
		PageSize: 10,
	}

	if q := r.URL.Query().Get("q"); q != "" {
		filters.SearchQuery = &q
	}

	if domain := r.URL.Query().Get("domain"); domain != "" {
		filters.Domain = &domain
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filters.Page = int32(page)
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			filters.PageSize = int32(pageSize)
		}
	}

	result, err := h.useCase.ListHistory(ctx, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// TopDomains godoc
// @Summary Get most visited domains
// @Description Get aggregated visit counts by domain
// @Tags browser-history
// @Param limit query int false "Number of domains to return (default 10)"
// @Success 200 {array} entity.DomainVisitCount
// @Router /api/history/domains [get]
func (h *BrowserHistoryHandler) TopDomains(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := int32(10)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(l)
		}
	}

	domains, err := h.useCase.TopDomains(ctx, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domains)
}

// RecentHistory godoc
// @Summary Get recent browser history
// @Description Get N most recent history entries
// @Tags browser-history
// @Param limit query int false "Number of entries to return (default 20)"
// @Success 200 {array} entity.BrowserHistory
// @Router /api/history/recent [get]
func (h *BrowserHistoryHandler) RecentHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := int32(20)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(l)
		}
	}

	history, err := h.useCase.RecentHistory(ctx, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
