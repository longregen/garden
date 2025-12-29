package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	httpAdapter "garden3/internal/adapter/primary/http"
	"garden3/internal/port/input"
)

// UtilityHandler handles utility-related HTTP requests
type UtilityHandler struct {
	useCase input.UtilityUseCase
}

// NewUtilityHandler creates a new utility handler
func NewUtilityHandler(useCase input.UtilityUseCase) *UtilityHandler {
	return &UtilityHandler{
		useCase: useCase,
	}
}

// RegisterRoutes registers the utility routes
func (h *UtilityHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/debug", h.GetDebugInfo)
	r.Post("/api/conversations/cleanup", h.CleanupStaleConversations)
	r.Post("/api/messages/content", h.GetMessagesContent)
}

// GetDebugInfo returns system debug information
func (h *UtilityHandler) GetDebugInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	debugInfo, err := h.useCase.GetDebugInfo(ctx)
	if err != nil {
		http.Error(w, "Failed to get debug info", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, debugInfo)
}

// CleanupResponse represents the cleanup operation response
type CleanupResponse struct {
	DeletedCount int64 `json:"deleted_count"`
}

// CleanupStaleConversations removes stale sessions
func (h *UtilityHandler) CleanupStaleConversations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	count, err := h.useCase.CleanupStaleConversations(ctx)
	if err != nil {
		http.Error(w, "Failed to cleanup conversations", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, CleanupResponse{DeletedCount: count})
}

// MessagesContentRequest represents the request for fetching messages by IDs
type MessagesContentRequest struct {
	MessageIDs []string `json:"message_ids"`
}

// GetMessagesContent retrieves messages by their IDs
func (h *UtilityHandler) GetMessagesContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req MessagesContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.MessageIDs) == 0 {
		httpAdapter.JSON(w, http.StatusOK, []interface{}{})
		return
	}

	// Parse UUIDs
	messageIDs := make([]uuid.UUID, 0, len(req.MessageIDs))
	for _, idStr := range req.MessageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid message ID: "+idStr, http.StatusBadRequest)
			return
		}
		messageIDs = append(messageIDs, id)
	}

	messages, err := h.useCase.GetMessagesByIDs(ctx, messageIDs)
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	httpAdapter.JSON(w, http.StatusOK, messages)
}
