package web

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    any `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSON sends a JSON response with the given status code
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// Success sends a successful JSON response
func Success(w http.ResponseWriter, data any) {
	resp := Response{
		Success: true,
		Data:    data,
	}
	JSON(w, http.StatusOK, resp)
}

// Error sends an error JSON response
func Error(w http.ResponseWriter, status int, message string) {
	resp := Response{
		Success: false,
		Error:   message,
	}
	JSON(w, status, resp)
}
