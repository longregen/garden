package http

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func Error(w http.ResponseWriter, status int, err error) {
	JSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: err.Error(),
	})
}

func NotFound(w http.ResponseWriter) {
	JSON(w, http.StatusNotFound, ErrorResponse{
		Error: "Not Found",
	})
}

func BadRequest(w http.ResponseWriter, err error) {
	Error(w, http.StatusBadRequest, err)
}

func InternalError(w http.ResponseWriter, err error) {
	Error(w, http.StatusInternalServerError, err)
}
