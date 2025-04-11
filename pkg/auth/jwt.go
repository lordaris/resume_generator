package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// JWT claim keys
const (
	// TokenTypeAccess is the token type for access tokens
	TokenTypeAccess = "access"
	// TokenTypeRefresh is the token type for refresh tokens
	TokenTypeRefresh = "refresh"
	// TokenTypeReset is the token type for password reset tokens
	TokenTypeReset = "reset"
)

// JWT claim errors
var (
	// ErrTokenExpired is returned when the token has expired
	ErrTokenExpired = errors.New("token expired")
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrWrongTokenType is returned when the token type is not as expected
	ErrWrongTokenType = errors.New("wrong token type")
)

// JWTClaims defines custom claims for JWT
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTConfig contains JWT configuration
type JWTConfig struct {
	// Secret is the JWT signing key
	Secret string
	// AccessTokenExpiry is the duration after which an access token expires
	AccessTokenExpiry time.Duration
	// RefreshTokenExpiry is the duration after which a refresh token expires
	RefreshTokenExpiry time.Duration
	// ResetTokenExpiry is the duration after which a password reset token expires
	ResetTokenExpiry time.Duration
	// Issuer is the token issuer
	Issuer string
	// Audience is the token audience
	Audience string
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour, // 7 days
		ResetTokenExpiry:   1 * time.Hour,
		Issuer:             "resume_generator",
		Audience:           "resume_generator_users",
	}
}

// JWT handles JWT token generation and validation
type JWT struct {
	config JWTConfig
}

// NewJWT creates a new JWT handler
func NewJWT(config JWTConfig) *JWT {
	if config.Secret == "" {
		panic("JWT secret is required")
	}

	// Apply defaults for zero values
	if config.AccessTokenExpiry == 0 {
		config.AccessTokenExpiry = DefaultJWTConfig().AccessTokenExpiry
	}
	if config.RefreshTokenExpiry == 0 {
		config.RefreshTokenExpiry = DefaultJWTConfig().RefreshTokenExpiry
	}
	if config.ResetTokenExpiry == 0 {
		config.ResetTokenExpiry = DefaultJWTConfig().ResetTokenExpiry
	}
	if config.Issuer == "" {
		config.Issuer = DefaultJWTConfig().Issuer
	}
	if config.Audience == "" {
		config.Audience = DefaultJWTConfig().Audience
	}

	return &JWT{
		config: config,
	}
}

// GenerateAccessToken generates a new access token
func (j *JWT) GenerateAccessToken(userID, email, role string) (string, error) {
	return j.generateToken(userID, email, role, TokenTypeAccess, j.config.AccessTokenExpiry)
}

// GenerateRefreshToken generates a new refresh token
func (j *JWT) GenerateRefreshToken(userID, email, role string) (string, error) {
	return j.generateToken(userID, email, role, TokenTypeRefresh, j.config.RefreshTokenExpiry)
}

// GenerateResetToken generates a new password reset token
func (j *JWT) GenerateResetToken(userID, email string) (string, error) {
	return j.generateToken(userID, email, "", TokenTypeReset, j.config.ResetTokenExpiry)
}

// generateToken is a helper function to generate JWT tokens
func (j *JWT) generateToken(userID, email, role, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.config.Issuer,
			Subject:   userID,
			Audience:  []string{j.config.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.config.Secret))
	if err != nil {
		log.Error().Err(err).Msg("Failed to sign JWT token")
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ParseToken parses and validates a JWT token
func (j *JWT) ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (j *JWT) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, ErrWrongTokenType
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (j *JWT) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrWrongTokenType
	}

	return claims, nil
}

// ValidateResetToken validates a password reset token
func (j *JWT) ValidateResetToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeReset {
		return nil, ErrWrongTokenType
	}

	return claims, nil
}
