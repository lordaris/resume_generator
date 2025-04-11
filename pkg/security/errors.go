package security

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
	Code   string `json:"code,omitempty"`
}

// WriteError writes an error response in JSON format
func WriteError(w http.ResponseWriter, status int, message string, code string) {
	response := ErrorResponse{
		Error:  message,
		Status: status,
		Code:   code,
	}

	// Marshal error response to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		// If marshaling fails, write a plain text error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set content type and status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
}

// WriteInvalidRequestError writes a 400 Bad Request error
func WriteInvalidRequestError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message, "INVALID_REQUEST")
}

// WriteUnauthorizedError writes a 401 Unauthorized error
func WriteUnauthorizedError(w http.ResponseWriter) {
	WriteError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
}

// WriteForbiddenError writes a 403 Forbidden error
func WriteForbiddenError(w http.ResponseWriter) {
	WriteError(w, http.StatusForbidden, "Forbidden", "FORBIDDEN")
}

// WriteNotFoundError writes a 404 Not Found error
func WriteNotFoundError(w http.ResponseWriter) {
	WriteError(w, http.StatusNotFound, "Not found", "NOT_FOUND")
}

// WriteMethodNotAllowedError writes a 405 Method Not Allowed error
func WriteMethodNotAllowedError(w http.ResponseWriter) {
	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "METHOD_NOT_ALLOWED")
}

// WriteConflictError writes a 409 Conflict error
func WriteConflictError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, message, "CONFLICT")
}

// WriteTooManyRequestsError writes a 429 Too Many Requests error
func WriteTooManyRequestsError(w http.ResponseWriter) {
	WriteError(w, http.StatusTooManyRequests, "Rate limit exceeded", "RATE_LIMIT_EXCEEDED")
}

// WriteInternalServerError writes a 500 Internal Server Error
func WriteInternalServerError(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_SERVER_ERROR")
}
