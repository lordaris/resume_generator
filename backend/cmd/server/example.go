package main

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/lordaris/resume_generator/pkg/security"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Example of using the security middleware
func setupSecurityMiddleware() (http.Handler, error) {
	// Generate random keys for demo purposes
	// In production, these should come from config and be persistent
	csrfKey := make([]byte, 32)
	if _, err := rand.Read(csrfKey); err != nil {
		return nil, err
	}

	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, err
	}

	// Set up CSRF protection
	csrfConfig := security.CSRFConfig{
		Key:            csrfKey,
		CookieSecure:   true,
		CookieSameSite: http.SameSiteStrictMode,
	}
	csrfProtection := security.NewCSRFProtection(csrfConfig)

	// Set up session management
	sessionConfig := security.SessionConfig{
		Key:            sessionKey,
		CookieSecure:   true,
		CookieSameSite: http.SameSiteLaxMode,
	}
	session, err := security.NewSession(sessionConfig)
	if err != nil {
		return nil, err
	}

	// Set up Redis client for rate limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	// Set up rate limiter
	rateLimiterConfig := security.RateLimiterConfig{
		Redis:    redisClient,
		Limit:    100,         // 100 requests
		Interval: time.Minute, // per minute
	}
	rateLimiter := security.NewRateLimiter(rateLimiterConfig)

	// Set up request validator
	validatorConfig := security.ValidationConfig{
		AllowedContentTypes: []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		},
		MaxBodySize:  1 * 1024 * 1024, // 1MB
		StrictPolicy: true,
	}
	validator := security.NewValidator(validatorConfig)

	// Create middleware chain
	securityMiddleware := security.SecurityMiddleware(csrfProtection, session, rateLimiter, validator)

	// Example handler
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Hello, secure world!"}`))
	})

	// Apply middleware chain to handler
	return securityMiddleware(helloHandler), nil
}

func exampleSetupRouter() {
	// Create router
	mux := http.NewServeMux()

	// Set up security middleware
	secureHandler, err := setupSecurityMiddleware()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set up security middleware")
	}

	// Add routes
	mux.Handle("GET /api/hello", secureHandler)

	// Define login handler that uses password verification
	loginHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			security.WriteInvalidRequestError(w, "Invalid form data")
			return
		}

		// Get username and password
		username := r.FormValue("username")
		password := r.FormValue("password")

		// In a real application, you would fetch the password hash from the database
		// Here we just use a hardcoded hash for demonstration
		hashedPassword := "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$hash"

		// Verify password
		match, err := security.VerifyPassword(password, hashedPassword)
		if err != nil {
			log.Error().Err(err).Msg("Password verification error")
			security.WriteInternalServerError(w)
			return
		}

		if !match {
			security.WriteUnauthorizedError(w)
			return
		}

		// Password is correct, create a session
		sessionConfig := security.SessionConfig{
			Key:          []byte("your-32-byte-session-encryption-key"),
			CookieSecure: true,
		}
		session, _ := security.NewSession(sessionConfig)

		// Create session data
		sessionData := security.SessionData{
			UserID:    "user123",
			Role:      "user",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Data: map[string]any{
				"username": username,
			},
		}

		// Store session
		if err := session.Create(w, sessionData); err != nil {
			log.Error().Err(err).Msg("Failed to create session")
			security.WriteInternalServerError(w)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"message":"Login successful"}`))
	})

	// Add login route
	mux.Handle("POST /api/login", loginHandler)

	// Start server
	log.Info().Msg("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
