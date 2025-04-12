package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// CSRFTokenLength is the length of the random part of the token
	CSRFTokenLength = 32
	// CSRFCookieName is the name of the cookie that stores the CSRF token
	CSRFCookieName = "csrf_token"
	// CSRFHeaderName is the name of the header that should contain the CSRF token
	CSRFHeaderName = "X-CSRF-Token"
	// CSRFFormField is the name of the form field that can contain the CSRF token
	CSRFFormField = "csrf_token"
)

var (
	// ErrInvalidCSRFToken is returned when the CSRF token is invalid
	ErrInvalidCSRFToken = errors.New("invalid CSRF token")
	// ErrMissingCSRFToken is returned when the CSRF token is missing
	ErrMissingCSRFToken = errors.New("missing CSRF token")
)

// CSRFConfig contains configuration options for CSRF protection
type CSRFConfig struct {
	// Key is the secret key used to sign CSRF tokens
	Key []byte
	// CookieSecure determines if the cookie should be sent only over HTTPS
	CookieSecure bool
	// CookiePath is the path for which the cookie is valid
	CookiePath string
	// CookieDomain is the domain for which the cookie is valid
	CookieDomain string
	// CookieMaxAge is the maximum age of the cookie in seconds
	CookieMaxAge int
	// CookieSameSite determines the SameSite attribute of the cookie
	CookieSameSite http.SameSite
}

// CSRFProtection provides CSRF protection using HMAC-SHA256
type CSRFProtection struct {
	config CSRFConfig
}

// NewCSRFProtection creates a new CSRF protection middleware
func NewCSRFProtection(config CSRFConfig) *CSRFProtection {
	if len(config.Key) == 0 {
		panic("CSRF protection key cannot be empty")
	}

	// Set defaults if not provided
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = 86400 // 24 hours
	}
	if config.CookieSameSite == 0 {
		config.CookieSameSite = http.SameSiteStrictMode
	}

	return &CSRFProtection{
		config: config,
	}
}

// generateToken generates a new CSRF token
func (c *CSRFProtection) generateToken() (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Base64 encode the random bytes
	randomString := base64.StdEncoding.EncodeToString(randomBytes)

	// Create the token payload (random string + timestamp)
	timestamp := time.Now().Unix()
	payload := fmt.Sprintf("%s|%d", randomString, timestamp)

	// Sign the payload using HMAC-SHA256
	h := hmac.New(sha256.New, c.config.Key)
	h.Write([]byte(payload))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Combine the payload and signature
	token := fmt.Sprintf("%s|%s", payload, signature)

	return token, nil
}

// validateToken validates a CSRF token
func (c *CSRFProtection) validateToken(token string) error {
	if token == "" {
		return ErrMissingCSRFToken
	}

	// Split the token into payload and signature
	parts := strings.Split(token, "|")
	if len(parts) != 3 {
		return ErrInvalidCSRFToken
	}

	// Extract the components
	randomStr, timestampStr, receivedSignature := parts[0], parts[1], parts[2]
	payload := fmt.Sprintf("%s|%s", randomStr, timestampStr)

	// Sign the payload using HMAC-SHA256
	h := hmac.New(sha256.New, c.config.Key)
	h.Write([]byte(payload))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Compare the signatures
	if !hmac.Equal([]byte(receivedSignature), []byte(expectedSignature)) {
		return ErrInvalidCSRFToken
	}

	return nil
}

// Middleware provides CSRF protection middleware
func (c *CSRFProtection) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for GET, HEAD, OPTIONS, TRACE methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || r.Method == http.MethodTrace {
			// Generate and set a new CSRF token if not present
			cookie, err := r.Cookie(CSRFCookieName)
			if err != nil || cookie.Value == "" {
				// Generate a new token
				token, err := c.generateToken()
				if err != nil {
					http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
					return
				}

				// Set the token as a cookie
				http.SetCookie(w, &http.Cookie{
					Name:     CSRFCookieName,
					Value:    token,
					Path:     c.config.CookiePath,
					Domain:   c.config.CookieDomain,
					MaxAge:   c.config.CookieMaxAge,
					Secure:   c.config.CookieSecure,
					HttpOnly: true,
					SameSite: c.config.CookieSameSite,
				})
			}
			next.ServeHTTP(w, r)
			return
		}

		// For other methods, validate the CSRF token
		cookie, err := r.Cookie(CSRFCookieName)
		if err != nil || cookie.Value == "" {
			log.Error().Err(err).Msg("CSRF token cookie missing")
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		// Check for the token in the header or form
		// Priority: Header > Form > Cookie (cookie is used as fallback)
		var token string
		if headerToken := r.Header.Get(CSRFHeaderName); headerToken != "" {
			token = headerToken
		} else if formToken := r.FormValue(CSRFFormField); formToken != "" {
			token = formToken
		} else {
			log.Error().Msg("CSRF token not found in header or form")
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		// Validate the token
		if err := c.validateToken(token); err != nil {
			log.Error().Err(err).Msg("CSRF token validation failed")
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
