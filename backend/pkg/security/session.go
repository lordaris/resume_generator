package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// SessionCookieName is the name of the cookie that stores the session
	SessionCookieName = "session"
)

var (
	// ErrSessionDecryption is returned when session decryption fails
	ErrSessionDecryption = errors.New("session decryption failed")
	// ErrInvalidSession is returned when the session is invalid
	ErrInvalidSession = errors.New("invalid session")
)

// SessionConfig contains configuration options for session management
type SessionConfig struct {
	// Key is the secret key used to encrypt session data (must be 32 bytes for AES-256)
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

// SessionData contains the data stored in a session
type SessionData struct {
	// UserID is the ID of the authenticated user
	UserID string `json:"user_id,omitempty"`
	// Role is the role of the authenticated user
	Role string `json:"role,omitempty"`
	// ExpiresAt is the expiration time of the session
	ExpiresAt time.Time `json:"expires_at"`
	// Custom data map for additional session data
	Data map[string]any `json:"data,omitempty"`
}

// Session provides secure session management with encrypted cookies
type Session struct {
	config SessionConfig
}

// NewSession creates a new session manager
func NewSession(config SessionConfig) (*Session, error) {
	// AES-256 requires a 32-byte key
	if len(config.Key) != 32 {
		return nil, errors.New("session encryption key must be 32 bytes")
	}

	// Set defaults if not provided
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = 86400 // 24 hours
	}
	if config.CookieSameSite == 0 {
		config.CookieSameSite = http.SameSiteLaxMode
	}

	return &Session{
		config: config,
	}, nil
}

// Create creates a new session
func (s *Session) Create(w http.ResponseWriter, sessionData SessionData) error {
	// Set expiration time if not already set
	if sessionData.ExpiresAt.IsZero() {
		sessionData.ExpiresAt = time.Now().Add(time.Duration(s.config.CookieMaxAge) * time.Second)
	}

	// Serialize session data to JSON
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	// Encrypt the session data
	encryptedData, err := s.encrypt(jsonData)
	if err != nil {
		return err
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    encryptedData,
		Path:     s.config.CookiePath,
		Domain:   s.config.CookieDomain,
		MaxAge:   s.config.CookieMaxAge,
		Secure:   s.config.CookieSecure,
		HttpOnly: true,
		SameSite: s.config.CookieSameSite,
	})

	return nil
}

// Get retrieves the session data from the request
func (s *Session) Get(r *http.Request) (SessionData, error) {
	var sessionData SessionData

	// Get the session cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return sessionData, err
	}

	// Decrypt the session data
	jsonData, err := s.decrypt(cookie.Value)
	if err != nil {
		return sessionData, ErrSessionDecryption
	}

	// Deserialize the JSON data
	if err := json.Unmarshal(jsonData, &sessionData); err != nil {
		return sessionData, ErrInvalidSession
	}

	// Check if the session has expired
	if time.Now().After(sessionData.ExpiresAt) {
		return sessionData, errors.New("session expired")
	}

	return sessionData, nil
}

// Clear removes the session
func (s *Session) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     s.config.CookiePath,
		Domain:   s.config.CookieDomain,
		MaxAge:   -1,
		Secure:   s.config.CookieSecure,
		HttpOnly: true,
		SameSite: s.config.CookieSameSite,
	})
}

// encrypt encrypts data using AES-GCM
func (s *Session) encrypt(plaintext []byte) (string, error) {
	// Create a new AES cipher block
	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Base64 encode the ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-GCM
func (s *Session) decrypt(encryptedData string) ([]byte, error) {
	// Base64 decode the ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(s.config.Key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM cipher mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract the nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Middleware provides session middleware that adds session data to the request context
func (s *Session) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get session data
		sessionData, err := s.Get(r)
		if err != nil {
			// If the session is invalid or expired, clear it
			if errors.Is(err, ErrSessionDecryption) || errors.Is(err, ErrInvalidSession) ||
				err.Error() == "session expired" {
				s.Clear(w)
				log.Info().Msg("Cleared invalid session")
			}

			// Continue without session data
			next.ServeHTTP(w, r)
			return
		}

		// Add session data to request context
		ctx := SetSessionContext(r.Context(), sessionData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
