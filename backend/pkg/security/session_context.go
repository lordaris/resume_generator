package security

import (
	"context"
)

// contextKey is a private type for context keys
type contextKey int

const (
	// sessionContextKey is the context key for session data
	sessionContextKey contextKey = iota
)

// SetSessionContext adds session data to the context
func SetSessionContext(ctx context.Context, sessionData SessionData) context.Context {
	return context.WithValue(ctx, sessionContextKey, sessionData)
}

// GetSessionFromContext retrieves session data from the context
func GetSessionFromContext(ctx context.Context) (SessionData, bool) {
	sessionData, ok := ctx.Value(sessionContextKey).(SessionData)
	return sessionData, ok
}

// IsAuthenticated checks if the user in the context is authenticated
func IsAuthenticated(ctx context.Context) bool {
	sessionData, ok := GetSessionFromContext(ctx)
	return ok && sessionData.UserID != ""
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	sessionData, ok := GetSessionFromContext(ctx)
	if !ok {
		return ""
	}
	return sessionData.UserID
}

// GetUserRole retrieves the user role from the context
func GetUserRole(ctx context.Context) string {
	sessionData, ok := GetSessionFromContext(ctx)
	if !ok {
		return ""
	}
	return sessionData.Role
}

// HasRole checks if the user in the context has the specified role
func HasRole(ctx context.Context, role string) bool {
	sessionData, ok := GetSessionFromContext(ctx)
	return ok && sessionData.Role == role
}
