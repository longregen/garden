package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"garden3/internal/domain/entity"
	"garden3/internal/port/input"
	"github.com/go-chi/chi/v5"
)

type SearchHandler struct {
	useCase input.SearchUseCase
}

func NewSearchHandler(useCase input.SearchUseCase) *SearchHandler {
	return &SearchHandler{
		useCase: useCase,
	}
}

func (h *SearchHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/search", func(r chi.Router) {
		r.Get("/", h.SearchAll)
		r.Post("/advanced", h.AdvancedSearch)
	})
}

// SearchAll godoc
// @Summary Unified search across all content
// @Description Search across contacts, conversations, bookmarks, browser history, and notes
// @Tags search
// @Param q query string false "Search query (legacy)"
// @Param query query string false "Search query (new)"
// @Param limit query int false "Result limit (default 50)"
// @Param exact_match_weight query number false "Exact match weight (default 5.0)"
// @Param similarity_weight query number false "Similarity weight (default 2.0)"
// @Param levenshteinWeight query number false "Similarity weight (legacy, same as similarity_weight)"
// @Param recency_weight query number false "Recency weight (default 1.0)"
// @Success 200 {array} entity.UnifiedSearchResult
// @Router /api/search [get]
func (h *SearchHandler) SearchAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Support both old (?q=) and new (?query=) parameter names
	query := r.URL.Query().Get("query")
	if query == "" {
		query = r.URL.Query().Get("q")
	}

	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	limit := int32(50)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(l)
		}
	}

	var weights *entity.SearchWeights
	if r.URL.Query().Has("exact_match_weight") || r.URL.Query().Has("similarity_weight") || r.URL.Query().Has("levenshteinWeight") || r.URL.Query().Has("recency_weight") {
		defaultWeights := entity.DefaultSearchWeights()
		weights = &defaultWeights

		if emw := r.URL.Query().Get("exact_match_weight"); emw != "" {
			if w, err := strconv.ParseFloat(emw, 64); err == nil {
				weights.ExactMatchWeight = w
			}
		}

		// Support both old (levenshteinWeight) and new (similarity_weight) parameter names
		// Old name takes precedence if both are provided
		sw := r.URL.Query().Get("levenshteinWeight")
		if sw == "" {
			sw = r.URL.Query().Get("similarity_weight")
		}
		if sw != "" {
			if w, err := strconv.ParseFloat(sw, 64); err == nil {
				weights.SimilarityWeight = w
			}
		}

		if rw := r.URL.Query().Get("recency_weight"); rw != "" {
			if w, err := strconv.ParseFloat(rw, 64); err == nil {
				weights.RecencyWeight = w
			}
		}
	}

	results, err := h.useCase.SearchAll(ctx, query, weights, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// AdvancedSearch godoc
// @Summary Advanced LLM-powered search
// @Description Searches bookmarks using vector similarity and synthesizes an answer using an LLM
// @Tags search
// @Accept json
// @Produce json
// @Param body body object{query=string} true "Search query"
// @Success 200 {object} entity.AdvancedSearchResult
// @Router /api/search/advanced [post]
func (h *SearchHandler) AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req struct {
		Query interface{} `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Handle different query formats
	var queryString string
	switch v := req.Query.(type) {
	case string:
		queryString = v
	case map[string]interface{}:
		// JSON query object - stringify it
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			http.Error(w, "Failed to process query object", http.StatusBadRequest)
			return
		}
		queryString = string(jsonBytes)
	default:
		http.Error(w, "Query must be a string or object", http.StatusBadRequest)
		return
	}

	if queryString == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Perform advanced search
	result, err := h.useCase.AdvancedSearch(ctx, queryString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
