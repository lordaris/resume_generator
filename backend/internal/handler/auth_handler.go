package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/lordaris/resume_generator/internal/service"
	"github.com/lordaris/resume_generator/pkg/security"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validate
	rateLimiter *security.RateLimiter
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, redisClient *redis.Client) *AuthHandler {
	// Create rate limiter for auth endpoints
	rateLimiterConfig := security.RateLimiterConfig{
		Redis:    redisClient,
		Limit:    100,         // 100 requests
		Interval: time.Minute, // per minute
	}

	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
		rateLimiter: security.NewRateLimiter(rateLimiterConfig),
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Role     string `json:"role" validate:"omitempty,oneof=user admin"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// PasswordResetRequestRequest represents a password reset request request
type PasswordResetRequestRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}

// RegisterHandler handles user registration
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if ok := h.applyRateLimit(w, r); !ok {
		return
	}

	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		RespondWithValidationError(w, validationErrors)
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "user"
	}

	// Register user
	user, err := h.authService.Register(req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			RespondWithError(w, http.StatusConflict, "User with this email already exists", "USER_EXISTS")
			return
		}
		log.Error().Err(err).Msg("Failed to register user")
		RespondWithError(w, http.StatusInternalServerError, "Failed to register user", "REGISTRATION_FAILED")
		return
	}

	// Return success response
	RespondWithJSON(w, http.StatusCreated, map[string]any{
		"message": "User registered successfully",
		"user_id": user.ID,
	})
}

// LoginHandler handles user login
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if ok := h.applyRateLimit(w, r); !ok {
		return
	}

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		RespondWithValidationError(w, validationErrors)
		return
	}

	// Get client info
	userAgent := r.UserAgent()
	clientIP := getClientIP(r)

	// Login user
	tokens, err := h.authService.Login(req.Email, req.Password, userAgent, clientIP)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			// Return same error for invalid email or password to prevent user enumeration
			RespondWithError(w, http.StatusUnauthorized, "Invalid email or password", "INVALID_CREDENTIALS")
			return
		}
		log.Error().Err(err).Msg("Failed to login user")
		RespondWithError(w, http.StatusInternalServerError, "Failed to login user", "LOGIN_FAILED")
		return
	}

	// Return tokens
	RespondWithJSON(w, http.StatusOK, tokens)
}

// RefreshTokenHandler handles token refresh
func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if ok := h.applyRateLimit(w, r); !ok {
		return
	}

	// Parse request body
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		RespondWithValidationError(w, validationErrors)
		return
	}

	// Get client info
	userAgent := r.UserAgent()
	clientIP := getClientIP(r)

	// Refresh token
	tokens, err := h.authService.RefreshToken(req.RefreshToken, userAgent, clientIP)
	if err != nil {
		status := http.StatusUnauthorized
		code := "INVALID_TOKEN"
		message := "Invalid refresh token"

		if errors.Is(err, service.ErrExpiredToken) {
			code = "TOKEN_EXPIRED"
			message = "Refresh token expired"
		} else if errors.Is(err, service.ErrInvalidSession) {
			code = "INVALID_SESSION"
			message = "Invalid session"
		}

		RespondWithError(w, status, message, code)
		return
	}

	// Return tokens
	RespondWithJSON(w, http.StatusOK, tokens)
}

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if ok := h.applyRateLimit(w, r); !ok {
		return
	}

	// Parse request body
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		RespondWithValidationError(w, validationErrors)
		return
	}

	// Logout user
	if err := h.authService.Logout(req.RefreshToken); err != nil {
		log.Error().Err(err).Msg("Failed to logout user")
		RespondWithError(w, http.StatusInternalServerError, "Failed to logout user", "LOGOUT_FAILED")
		return
	}

	// Return success response
	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "User logged out successfully",
	})
}

// RequestPasswordResetHandler handles password reset requests
func (h *AuthHandler) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if ok := h.applyRateLimit(w, r); !ok {
		return
	}

	// Parse request body
	var req PasswordResetRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		RespondWithValidationError(w, validationErrors)
		return
	}

	// Request password reset
	resetToken, err := h.authService.RequestPasswordReset(req.Email)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			// Always return success even if user doesn't exist to prevent user enumeration
			RespondWithJSON(w, http.StatusOK, map[string]any{
				"message": "Password reset instructions sent if email exists",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to request password reset")
		RespondWithError(w, http.StatusInternalServerError, "Failed to request password reset", "PASSWORD_RESET_FAILED")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Password reset instructions sent",
		"token":   resetToken,
	})
}

// ResetPasswordHandler handles password reset
func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Apply rate limiting
	if ok := h.applyRateLimit(w, r); !ok {
		return
	}

	// Parse request body
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		RespondWithValidationError(w, validationErrors)
		return
	}

	// Reset password
	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		status := http.StatusBadRequest
		code := "PASSWORD_RESET_FAILED"
		message := "Failed to reset password"

		if errors.Is(err, service.ErrInvalidToken) {
			code = "INVALID_TOKEN"
			message = "Invalid reset token"
		} else if errors.Is(err, service.ErrExpiredToken) {
			code = "TOKEN_EXPIRED"
			message = "Reset token expired"
		} else if errors.Is(err, service.ErrPasswordResetUsed) {
			code = "TOKEN_USED"
			message = "Reset token already used"
		}

		RespondWithError(w, status, message, code)
		return
	}

	// Return success response
	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Password reset successfully",
	})
}

// Helper functions

// applyRateLimit applies rate limiting to a request
func (h *AuthHandler) applyRateLimit(w http.ResponseWriter, r *http.Request) bool {
	count, err := h.rateLimiter.CheckRateLimit(r.Context(), r)
	if err != nil {
		if errors.Is(err, security.ErrRateLimitExceeded) {
			RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded", "RATE_LIMIT_EXCEEDED")
			return false
		}
		log.Error().Err(err).Msg("Rate limiting error")
	}

	// Add rate limit headers with correct calculation
	w.Header().Set("X-RateLimit-Limit", "100")
	remaining := max(100-count, 0)
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("X-RateLimit-Reset", "60")

	return true
}
