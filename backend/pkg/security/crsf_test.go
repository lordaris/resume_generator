package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFGenerateAndValidateToken(t *testing.T) {
	// Create a CSRF protection instance
	csrfKey := []byte("test-csrf-key-must-be-32-bytes-long!")
	config := CSRFConfig{
		Key:            csrfKey,
		CookieSecure:   true,
		CookieSameSite: http.SameSiteStrictMode,
	}
	csrf := NewCSRFProtection(config)

	// Generate a token
	token, err := csrf.generateToken()
	if err != nil {
		t.Fatalf("Token generation failed: %v", err)
	}

	// Check token format
	parts := strings.Split(token, "|")
	if len(parts) != 3 {
		t.Fatalf("Token should have 3 parts, got %d", len(parts))
	}

	// Validate the token
	err = csrf.validateToken(token)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	// Validate with invalid token
	err = csrf.validateToken("invalid-token")
	if err == nil {
		t.Errorf("Validation should fail with invalid token")
	}

	// Validate with modified token
	modifiedToken := strings.Replace(token, parts[0], "modified", 1)
	err = csrf.validateToken(modifiedToken)
	if err == nil {
		t.Errorf("Validation should fail with modified token")
	}

	// Validate with empty token
	err = csrf.validateToken("")
	if err == nil {
		t.Errorf("Validation should fail with empty token")
	}
}

func TestCSRFMiddleware(t *testing.T) {
	// Create a CSRF protection instance
	csrfKey := []byte("test-csrf-key-must-be-32-bytes-long!")
	config := CSRFConfig{
		Key:            csrfKey,
		CookieSecure:   false, // for testing
		CookieSameSite: http.SameSiteStrictMode,
	}
	csrf := NewCSRFProtection(config)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply CSRF middleware
	handler := csrf.Middleware(testHandler)

	// Test GET request (should set a CSRF token cookie)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("GET request failed: %d", rec.Code)
	}

	// Extract CSRF token from cookie
	cookies := rec.Result().Cookies()
	var csrfToken string
	for _, cookie := range cookies {
		if cookie.Name == CSRFCookieName {
			csrfToken = cookie.Value
			break
		}
	}
	if csrfToken == "" {
		t.Fatalf("CSRF token cookie not set")
	}

	// Test POST request without CSRF token (should fail)
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: csrfToken})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusForbidden {
		t.Errorf("POST without token should fail: got %d, want %d", rec.Code, http.StatusForbidden)
	}

	// Test POST request with valid CSRF token in header
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: csrfToken})
	req.Header.Set(CSRFHeaderName, csrfToken)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("POST with valid token should succeed: got %d, want %d", rec.Code, http.StatusOK)
	}

	// Test POST request with valid CSRF token in form
	formData := "csrf_token=" + csrfToken
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: csrfToken})
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("POST with valid form token should succeed: got %d, want %d", rec.Code, http.StatusOK)
	}

	// Test POST request with invalid CSRF token
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: csrfToken})
	req.Header.Set(CSRFHeaderName, "invalid-token")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusForbidden {
		t.Errorf("POST with invalid token should fail: got %d, want %d", rec.Code, http.StatusForbidden)
	}
}
