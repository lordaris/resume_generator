package domain

import (
	"fmt"
)

// Common resume model errors
var (
	ErrInvalidField = fmt.Errorf("invalid field value")
	ErrDateRange    = fmt.Errorf("invalid date range")
)

// ValidationError represents a validation error with a field name and message
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Unwrap returns the underlying error
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}
