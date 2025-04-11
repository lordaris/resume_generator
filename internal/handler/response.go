package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Status  int                    `json:"-"`
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	if _, err := w.Write(jsonData); err != nil {
		log.Error().Err(err).Msg("Failed to write JSON response")
	}
}

// respondWithError writes a JSON error response
func respondWithError(w http.ResponseWriter, status int, message, code string) {
	response := ErrorResponse{
		Status: status,
		Error:  message,
		Code:   code,
	}
	respondWithJSON(w, status, response)
}

// respondWithValidationError writes a validation error response
func respondWithValidationError(w http.ResponseWriter, validationErrors validator.ValidationErrors) {
	// Create error details
	details := make(map[string]interface{})
	fieldErrors := make(map[string]string)

	for _, err := range validationErrors {
		// Convert field name to JSON field name (usually lowercase)
		field := strings.ToLower(err.Field())
		fieldErrors[field] = getValidationErrorMessage(err)
	}

	details["fields"] = fieldErrors

	// Create error response
	response := ErrorResponse{
		Status:  http.StatusBadRequest,
		Error:   "Validation failed",
		Code:    "VALIDATION_FAILED",
		Details: details,
	}

	respondWithJSON(w, http.StatusBadRequest, response)
}

// getValidationErrorMessage returns a human-readable validation error message
func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + err.Param() + " characters long"
	case "max":
		return "Must be at most " + err.Param() + " characters long"
	case "oneof":
		return "Must be one of: " + err.Param()
	default:
		return "Invalid value"
	}
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// respondWithSuccess writes a JSON success response
func respondWithSuccess(w http.ResponseWriter, message string, data interface{}) {
	response := SuccessResponse{
		Message: message,
		Data:    data,
	}
	respondWithJSON(w, http.StatusOK, response)
}
