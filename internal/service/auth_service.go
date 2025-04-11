package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lordaris/resume_generator/internal/domain"
	"github.com/lordaris/resume_generator/internal/repository"
	"github.com/lordaris/resume_generator/pkg/auth"
	"github.com/lordaris/resume_generator/pkg/security"
	"github.com/rs/zerolog/log"
)

// AuthService errors
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token expired")
	ErrInvalidSession       = errors.New("invalid session")
	ErrPasswordResetExpired = errors.New("password reset expired")
	ErrPasswordResetUsed    = errors.New("password reset already used")
)

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Seconds until access token expires
}

// AuthService handles authentication and authorization
type AuthService struct {
	userRepo domain.UserRepository
	jwt      *auth.JWT
	config   AuthServiceConfig
}

// Logout logs out a user by invalidating their refresh token
func (s *AuthService) Logout(refreshToken string) error {
	// Get session by refresh token
	session, err := s.userRepo.GetSessionByToken(refreshToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// Session not found, consider it already logged out
			return nil
		}
		return err
	}

	// Delete session
	return s.userRepo.DeleteSession(session.ID)
}

// LogoutAll logs out a user from all devices
func (s *AuthService) LogoutAll(userID uuid.UUID) error {
	return s.userRepo.DeleteUserSessions(userID)
}

// RequestPasswordReset generates a password reset token
func (s *AuthService) RequestPasswordReset(email string) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	// Generate reset token
	resetToken, err := s.jwt.GenerateResetToken(user.ID.String(), user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate reset token")
		return "", err
	}

	// Store password reset
	reset := &domain.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(s.config.ResetTokenExpiry),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.CreatePasswordReset(reset); err != nil {
		log.Error().Err(err).Msg("Failed to create password reset")
		return "", err
	}

	return resetToken, nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(resetToken, newPassword string) error {
	// Validate reset token
	claims, err := s.jwt.ValidateResetToken(resetToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return ErrExpiredToken
		}
		return ErrInvalidToken
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid user ID in token")
		return ErrInvalidToken
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Get password reset by token
	reset, err := s.userRepo.GetPasswordResetByToken(resetToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	// Check if reset is expired
	if time.Now().After(reset.ExpiresAt) {
		return ErrPasswordResetExpired
	}

	// Check if reset was already used
	if !reset.UsedAt.IsZero() {
		return ErrPasswordResetUsed
	}

	// Hash new password
	passwordHash, err := security.HashPassword(newPassword, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return err
	}

	// Update user's password
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(user); err != nil {
		log.Error().Err(err).Msg("Failed to update user password")
		return err
	}

	// Mark reset as used
	if err := s.userRepo.MarkPasswordResetUsed(reset.ID); err != nil {
		log.Error().Err(err).Msg("Failed to mark password reset as used")
		// Continue anyway, just log the error
	}

	// Delete all user sessions
	if err := s.userRepo.DeleteUserSessions(user.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete user sessions")
		// Continue anyway, just log the error
	}

	return nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *AuthService) ValidateAccessToken(accessToken string) (*auth.JWTClaims, error) {
	claims, err := s.jwt.ValidateAccessToken(accessToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// AuthServiceConfig contains configuration for the auth service
type AuthServiceConfig struct {
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	ResetTokenExpiry   time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo domain.UserRepository, jwt *auth.JWT, config AuthServiceConfig) *AuthService {
	// Set default values if not provided
	if config.AccessTokenExpiry == 0 {
		config.AccessTokenExpiry = 15 * time.Minute
	}
	if config.RefreshTokenExpiry == 0 {
		config.RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
	}
	if config.ResetTokenExpiry == 0 {
		config.ResetTokenExpiry = 1 * time.Hour
	}

	return &AuthService{
		userRepo: userRepo,
		jwt:      jwt,
		config:   config,
	}
}

// Register registers a new user
func (s *AuthService) Register(email, password, role string) (*domain.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	// Hash password
	passwordHash, err := security.HashPassword(password, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(email, password, userAgent, clientIP string) (*TokenPair, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	match, err := security.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify password")
		return nil, err
	}

	if !match {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate access token")
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		return nil, err
	}

	// Store session
	session := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		ExpiresAt:    time.Now().Add(s.config.RefreshTokenExpiry),
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateSession(session); err != nil {
		log.Error().Err(err).Msg("Failed to create session")
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(refreshToken, userAgent, clientIP string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Invalid user ID in token")
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Get session by refresh token
	session, err := s.userRepo.GetSessionByToken(refreshToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		_ = s.userRepo.DeleteSession(session.ID)
		return nil, ErrExpiredToken
	}

	// Generate new tokens
	newAccessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate access token")
		return nil, err
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		return nil, err
	}

	// Delete old session
	if err := s.userRepo.DeleteSession(session.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete old session")
		// Continue anyway, just log the error
	}

	// Create new session
	newSession := &domain.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		ExpiresAt:    time.Now().Add(s.config.RefreshTokenExpiry),
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateSession(newSession); err != nil {
		log.Error().Err(err).Msg("Failed to create new session")
		return nil, err
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.AccessTokenExpiry.Seconds()),
	}, nil
}
