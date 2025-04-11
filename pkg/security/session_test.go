package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionCreateAndGet(t *testing.T) {
	// Create a session manager
	sessionKey := []byte("test-session-key-must-be-32-bytes!")
	config := SessionConfig{
		Key:            sessionKey,
		CookieSecure:   false, // for testing
		CookieSameSite: http.SameSiteLaxMode,
	}
	session, err := NewSession(config)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	// Create a test session
	sessionData := SessionData{
		UserID:    "user123",
		Role:      "admin",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Data: map[string]interface{}{
			"username": "testuser",
			"email":    "test@example.com",
		},
	}

	// Create a response recorder to capture the cookie
	rec := httptest.NewRecorder()

	// Create the session
	err = session.Create(rec, sessionData)
	if err != nil {
		t.Fatalf("Session creation failed: %v", err)
	}

	// Check that the session cookie was set
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("Session cookie not set")
	}

	// Create a request with the session cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(sessionCookie)

	// Get the session data
	retrievedData, err := session.Get(req)
	if err != nil {
		t.Fatalf("Session retrieval failed: %v", err)
	}

	// Check the retrieved data
	if retrievedData.UserID != sessionData.UserID {
		t.Errorf("Wrong UserID: got %s, want %s", retrievedData.UserID, sessionData.UserID)
	}
	if retrievedData.Role != sessionData.Role {
		t.Errorf("Wrong Role: got %s, want %s", retrievedData.Role, sessionData.Role)
	}
	if retrievedData.Data["username"] != sessionData.Data["username"] {
		t.Errorf("Wrong username: got %v, want %v", retrievedData.Data["username"], sessionData.Data["username"])
	}
}

func TestSessionClear(t *testing.T) {
	// Create a session manager
	sessionKey := []byte("test-session-key-must-be-32-bytes!")
	config := SessionConfig{
		Key:            sessionKey,
		CookieSecure:   false, // for testing
		CookieSameSite: http.SameSiteLaxMode,
	}
	session, err := NewSession(config)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	// Create a test session
	sessionData := SessionData{
		UserID:    "user123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Create a response recorder to capture the cookie
	rec := httptest.NewRecorder()

	// Create the session
	err = session.Create(rec, sessionData)
	if err != nil {
		t.Fatalf("Session creation failed: %v", err)
	}

	// Clear the session
	rec2 := httptest.NewRecorder()
	session.Clear(rec2)

	// Check that the session cookie was cleared
	cookies := rec2.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("Session cookie not found")
	}
	if sessionCookie.Value != "" {
		t.Errorf("Session cookie value should be empty")
	}
	if sessionCookie.MaxAge != -1 {
		t.Errorf("Session cookie MaxAge should be -1, got %d", sessionCookie.MaxAge)
	}
}

func TestSessionMiddleware(t *testing.T) {
	// Create a session manager
	sessionKey := []byte("test-session-key-must-be-32-bytes!")
	config := SessionConfig{
		Key:            sessionKey,
		CookieSecure:   false, // for testing
		CookieSameSite: http.SameSiteLaxMode,
	}
	session, err := NewSession(config)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	// Create a test handler that checks the session
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionData, ok := GetSessionFromContext(r.Context())
		if !ok {
			t.Error("Session data not found in context")
		} else if sessionData.UserID != "user123" {
			t.Errorf("Wrong UserID: got %s, want %s", sessionData.UserID, "user123")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Apply session middleware
	handler := session.Middleware(testHandler)

	// Create a test session
	sessionData := SessionData{
		UserID:    "user123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Create a response recorder to capture the cookie
	rec := httptest.NewRecorder()

	// Create the session
	err = session.Create(rec, sessionData)
	if err != nil {
		t.Fatalf("Session creation failed: %v", err)
	}

	// Extract the session cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("Session cookie not set")
	}

	// Create a request with the session cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(sessionCookie)

	// Test the middleware
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)

	// Check the response
	if rec2.Code != http.StatusOK {
		t.Errorf("Request with valid session should succeed: got %d, want %d", rec2.Code, http.StatusOK)
	}
}

func TestSessionExpired(t *testing.T) {
	// Create a session manager
	sessionKey := []byte("test-session-key-must-be-32-bytes!")
	config := SessionConfig{
		Key:            sessionKey,
		CookieSecure:   false, // for testing
		CookieSameSite: http.SameSiteLaxMode,
	}
	session, err := NewSession(config)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}

	// Create a test session that is already expired
	sessionData := SessionData{
		UserID:    "user123",
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
	}

	// Create a response recorder to capture the cookie
	rec := httptest.NewRecorder()

	// Create the session
	err = session.Create(rec, sessionData)
	if err != nil {
		t.Fatalf("Session creation failed: %v", err)
	}

	// Extract the session cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("Session cookie not set")
	}

	// Create a request with the session cookie
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(sessionCookie)

	// Get the session data
	_, err = session.Get(req)
	if err == nil {
		t.Error("Get should fail for expired session")
	} else if err.Error() != "session expired" {
		t.Errorf("Wrong error: got %v, want %v", err, "session expired")
	}
}
