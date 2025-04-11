package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lordaris/resume_generator/internal/service"
	"github.com/lordaris/resume_generator/pkg/auth"
	"github.com/rs/zerolog/log"
)

// contextKey is a type for context keys
type contextKey int

const (
	// userContextKey is the key for the user in the context
	userContextKey contextKey = iota
	// claimsContextKey is the key for the JWT claims in the context
	claimsContextKey
)

// AuthMiddleware extracts and validates JWT tokens from requests
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// AuthRequired middleware checks for a valid JWT token and injects user info into the context
func (m *AuthMiddleware) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := extractTokenFromHeader(r)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized", "INVALID_TOKEN")
			return
		}

		// Validate token
		claims, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			if errors.Is(err, service.ErrExpiredToken) {
				respondWithError(w, http.StatusUnauthorized, "Token expired", "TOKEN_EXPIRED")
				return
			}
			respondWithError(w, http.StatusUnauthorized, "Invalid token", "INVALID_TOKEN")
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), claimsContextKey, claims)

		// Continue with the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware checks if the user has the required role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := r.Context().Value(claimsContextKey).(*auth.JWTClaims)
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
				return
			}

			// Check role
			if claims.Role != role {
				respondWithError(w, http.StatusForbidden, "Forbidden", "FORBIDDEN")
				return
			}

			// Continue with the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// SessionLogger middleware logs user activity
type SessionLogger struct {
	// In a real application, this would often write to a database
	// For this example, we'll just log to stdout
}

// NewSessionLogger creates a new session logger
func NewSessionLogger() *SessionLogger {
	return &SessionLogger{}
}

// LogActivity middleware logs user activity
func (l *SessionLogger) LogActivity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Get claims from context if available
		var userID, email, role string
		if claims, ok := r.Context().Value(claimsContextKey).(*auth.JWTClaims); ok {
			userID = claims.UserID
			email = claims.Email
			role = claims.Role
		}

		// Create a custom response writer to capture the status code
		rw := &responseWriter{w, http.StatusOK}

		// Process the request
		next.ServeHTTP(rw, r)

		// Log the activity
		duration := time.Since(startTime)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.statusCode).
			Str("user_id", userID).
			Str("email", email).
			Str("role", role).
			Str("ip", getClientIP(r)).
			Str("user_agent", r.UserAgent()).
			Dur("duration", duration).
			Msg("API request")
	})
}

// Helper functions

// extractTokenFromHeader extracts the token from the Authorization header
func extractTokenFromHeader(r *http.Request) (string, error) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header")
	}

	// Check if it's a Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// getClientIP gets the client IP address
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, we want the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Get IP from remote address
	ip := r.RemoteAddr
	// Remove port if present
	if index := strings.IndexByte(ip, ':'); index != -1 {
		ip = ip[:index]
	}
	return ip
}

// GetUserIDFromContext gets the user ID from the context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	claims, ok := ctx.Value(claimsContextKey).(*auth.JWTClaims)
	if !ok {
		return uuid.Nil, errors.New("no claims in context")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// GetClaimsFromContext gets the JWT claims from the context
func GetClaimsFromContext(ctx context.Context) (*auth.JWTClaims, error) {
	claims, ok := ctx.Value(claimsContextKey).(*auth.JWTClaims)
	if !ok {
		return nil, errors.New("no claims in context")
	}

	return claims, nil
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
