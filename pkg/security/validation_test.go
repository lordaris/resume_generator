package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateContentType(t *testing.T) {
	// Create a validator with default allowed content types
	config := ValidationConfig{
		AllowedContentTypes: []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		},
	}
	validator := NewValidator(config)

	// Test cases
	testCases := []struct {
		name        string
		method      string
		contentType string
		shouldError bool
	}{
		{
			name:        "Valid JSON",
			method:      http.MethodPost,
			contentType: "application/json",
			shouldError: false,
		},
		{
			name:        "Valid Form",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			shouldError: false,
		},
		{
			name:        "Valid Multipart",
			method:      http.MethodPost,
			contentType: "multipart/form-data; boundary=something",
			shouldError: false,
		},
		{
			name:        "Invalid Content Type",
			method:      http.MethodPost,
			contentType: "text/plain",
			shouldError: true,
		},
		{
			name:        "Empty Content Type",
			method:      http.MethodPost,
			contentType: "",
			shouldError: false, // Empty is allowed for requests without body
		},
		{
			name:        "Invalid Format",
			method:      http.MethodPost,
			contentType: "invalid-format",
			shouldError: true,
		},
		{
			name:        "GET Request",
			method:      http.MethodGet,
			contentType: "text/plain", // Would be invalid for POST, but GET is allowed
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", nil)
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			err := validator.ValidateContentType(req)

			if tc.shouldError && err == nil {
				t.Errorf("Expected error for content type %q, got nil", tc.contentType)
			} else if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error for content type %q: %v", tc.contentType, err)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	// Create a validator
	validator := NewValidator(ValidationConfig{StrictPolicy: true})

	// Test cases
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Plain Text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "Script Tag",
			input:    "Hello <script>alert('XSS')</script> World",
			expected: "Hello  World",
		},
		{
			name:     "Style Tag",
			input:    "Hello <style>body { color: red; }</style> World",
			expected: "Hello  World",
		},
		{
			name:     "Iframe Tag",
			input:    "Hello <iframe src=\"evil.html\"></iframe> World",
			expected: "Hello  World",
		},
		{
			name:     "Allowed HTML with Strict Policy",
			input:    "Hello <b>Bold</b> <i>Italic</i> World",
			expected: "Hello Bold Italic World",
		},
		{
			name:     "HTML Attributes",
			input:    "<a href=\"https://example.com\" onclick=\"alert('XSS')\">Link</a>",
			expected: "Link",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.SanitizeHTML(tc.input)
			if result != tc.expected {
				t.Errorf("Sanitize: got %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestValidationMiddleware(t *testing.T) {
	// Create a validator
	config := ValidationConfig{
		AllowedContentTypes: []string{"application/json"},
		MaxBodySize:         100, // Very small for testing
	}
	validator := NewValidator(config)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply validation middleware
	handler := validator.Middleware(testHandler)

	// Test valid request
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"key\":\"value\"}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("Valid request failed: got %d, want %d", rec.Code, http.StatusOK)
	}

	// Test invalid content type
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("text data"))
	req.Header.Set("Content-Type", "text/plain")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Invalid content type: got %d, want %d", rec.Code, http.StatusUnsupportedMediaType)
	}

	// Test body too large
	largeBody := strings.Repeat("a", 200) // Larger than our 100 byte limit
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Body too large: got %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}

	// Test GET request (should bypass validation)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "text/plain") // Would be invalid for POST
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("GET request failed: got %d, want %d", rec.Code, http.StatusOK)
	}
}
