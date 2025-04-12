package security

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
)

const (
	// MaxBodySize is the maximum allowed size for request bodies (1MB)
	MaxBodySize = 1 * 1024 * 1024
)

var (
	// ErrInvalidContentType is returned when the Content-Type header is invalid
	ErrInvalidContentType = errors.New("invalid Content-Type")
	// ErrBodyTooLarge is returned when the request body is too large
	ErrBodyTooLarge = errors.New("request body too large")
)

// ValidationConfig contains configuration options for request validation
type ValidationConfig struct {
	// AllowedContentTypes is the list of allowed Content-Type headers
	AllowedContentTypes []string
	// MaxBodySize is the maximum allowed size for request bodies
	MaxBodySize int64
	// StrictPolicy determines if the HTML sanitizer should use a strict policy
	StrictPolicy bool
}

// Validator provides request validation functionality
type Validator struct {
	config        ValidationConfig
	htmlSanitizer *bluemonday.Policy
}

// NewValidator creates a new request validator
func NewValidator(config ValidationConfig) *Validator {
	// Set defaults if not provided
	if len(config.AllowedContentTypes) == 0 {
		config.AllowedContentTypes = []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		}
	}
	if config.MaxBodySize <= 0 {
		config.MaxBodySize = MaxBodySize
	}

	// Create HTML sanitizer policy
	var policy *bluemonday.Policy
	if config.StrictPolicy {
		policy = bluemonday.StrictPolicy()
	} else {
		policy = bluemonday.UGCPolicy()
	}

	return &Validator{
		config:        config,
		htmlSanitizer: policy,
	}
}

// ValidateContentType validates the Content-Type header
func (v *Validator) ValidateContentType(r *http.Request) error {
	// Skip validation for GET, HEAD, OPTIONS, TRACE methods
	if r.Method == http.MethodGet || r.Method == http.MethodHead ||
		r.Method == http.MethodOptions || r.Method == http.MethodTrace {
		return nil
	}

	// Extract content type without parameters
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		// Empty Content-Type is allowed for requests without body
		return nil
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ErrInvalidContentType
	}

	// Check if the content type is allowed
	for _, allowed := range v.config.AllowedContentTypes {
		if mediaType == allowed {
			return nil
		}
		// For multipart/form-data, we check only the prefix since it includes a boundary
		if strings.HasPrefix(allowed, "multipart/") && strings.HasPrefix(mediaType, allowed) {
			return nil
		}
	}

	return ErrInvalidContentType
}

// LimitBodySize limits the size of the request body
func (v *Validator) LimitBodySize(r *http.Request) ([]byte, error) {
	// Skip validation for GET, HEAD, OPTIONS, TRACE methods
	if r.Method == http.MethodGet || r.Method == http.MethodHead ||
		r.Method == http.MethodOptions || r.Method == http.MethodTrace {
		return nil, nil
	}

	// Limit the body size
	r.Body = http.MaxBytesReader(nil, r.Body, v.config.MaxBodySize)

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			return nil, ErrBodyTooLarge
		}
		return nil, err
	}

	return body, nil
}

// SanitizeHTML sanitizes HTML content
func (v *Validator) SanitizeHTML(input string) string {
	return v.htmlSanitizer.Sanitize(input)
}

// SanitizeMap sanitizes HTML content in a map
func (v *Validator) SanitizeMap(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = v.SanitizeHTML(value)
	}
	return result
}

// ValidationMiddleware provides request validation middleware
func (v *Validator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate Content-Type
		if err := v.ValidateContentType(r); err != nil {
			log.Error().Err(err).
				Str("content_type", r.Header.Get("Content-Type")).
				Msg("Invalid Content-Type")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte(`{"error":"Unsupported Content-Type","status":415}`))
			return
		}

		// Only apply body size limit for methods that can have a body
		if r.Method != http.MethodGet && r.Method != http.MethodHead &&
			r.Method != http.MethodOptions && r.Method != http.MethodTrace {

			// Create a copy of the request body for LimitBodySize to read
			originalBody := r.Body
			limitedBody, err := v.LimitBodySize(r)
			if err != nil {
				if errors.Is(err, ErrBodyTooLarge) {
					log.Error().Err(err).Msg("Request body too large")
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					w.Write([]byte(`{"error":"Request body too large","status":413}`))
					return
				}

				log.Error().Err(err).Msg("Error reading request body")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"Error reading request body","status":400}`))
				return
			}

			// Replace the body with the read body for the next handlers
			if limitedBody != nil {
				r.Body = io.NopCloser(strings.NewReader(string(limitedBody)))
			} else {
				r.Body = originalBody
			}
		}

		next.ServeHTTP(w, r)
	})
}
