package security

import (
	"net/http"
)

// Middleware represents a middleware handler
type Middleware func(http.Handler) http.Handler

// Chain chains multiple middlewares together
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// ResponseWriter is a wrapper for http.ResponseWriter that captures status code
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs requests and responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal server error","status":500}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// SecurityMiddleware combines all security middleware
func SecurityMiddleware(csrfProtection *CSRFProtection, session *Session, rateLimiter *RateLimiter, validator *Validator) Middleware {
	securityHeaders := SecurityHeaders(DefaultHeadersConfig())

	return Chain(
		RecoveryMiddleware,
		LoggingMiddleware,
		securityHeaders,
		rateLimiter.Middleware,
		validator.Middleware,
		session.Middleware,
		csrfProtection.Middleware,
	)
}
