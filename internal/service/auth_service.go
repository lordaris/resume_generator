package service

import (
	"time"

	"github.com/lordaris/resume_generator/pkg/auth"
)

// SetupAuth creates and configures the auth service and JWT handler
func SetupAuth(config AuthConfig) (*AuthService, *auth.JWT) {
	// Create JWT handler
	jwtConfig := auth.JWTConfig{
		Secret:             config.JWTSecret,
		AccessTokenExpiry:  config.AccessTokenExpiry,
		RefreshTokenExpiry: config.RefreshTokenExpiry,
		ResetTokenExpiry:   config.ResetTokenExpiry,
		Issuer:             config.Issuer,
		Audience:           config.Audience,
	}
	jwtHandler := auth.NewJWT(jwtConfig)

	// Create auth service
	authServiceConfig := AuthServiceConfig{
		AccessTokenExpiry:  config.AccessTokenExpiry,
		RefreshTokenExpiry: config.RefreshTokenExpiry,
		ResetTokenExpiry:   config.ResetTokenExpiry,
	}
	authService := NewAuthService(config.Repository, jwtHandler, authServiceConfig)

	return authService, jwtHandler
}

// AuthConfig contains all configuration for auth components
type AuthConfig struct {
	// Repository is the user repository
	Repository UserRepository
	// JWTSecret is the secret key for signing JWT tokens
	JWTSecret string
	// AccessTokenExpiry is the duration after which an access token expires
	AccessTokenExpiry time.Duration
	// RefreshTokenExpiry is the duration after which a refresh token expires
	RefreshTokenExpiry time.Duration
	// ResetTokenExpiry is the duration after which a password reset token expires
	ResetTokenExpiry time.Duration
	// Issuer is the JWT issuer
	Issuer string
	// Audience is the JWT audience
	Audience string
}

// UserRepository is an interface for user repository methods used in auth configuration
type UserRepository interface {
	// User operations
	CreateUser(user *User) error
	GetUserByID(id string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error

	// Session operations
	CreateSession(session *Session) error
	GetSessionByToken(token string) (*Session, error)
	DeleteSession(id string) error
	DeleteUserSessions(userID string) error

	// Password reset operations
	CreatePasswordReset(reset *PasswordReset) error
	GetPasswordResetByToken(token string) (*PasswordReset, error)
	MarkPasswordResetUsed(id string) error
}

// User represents a user in the auth system
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
}

// Session represents a user session
type Session struct {
	ID           string
	UserID       string
	RefreshToken string
	UserAgent    string
	ClientIP     string
	ExpiresAt    time.Time
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
}
