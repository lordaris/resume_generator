package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lordaris/resume_generator/internal/domain"
	"github.com/rs/zerolog/log"
)

// Common errors
var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record already exists")
)

// PostgresUserRepository implements the UserRepository interface using PostgreSQL
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// CreateUser creates a new user
func (r *PostgresUserRepository) CreateUser(user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	// Set default values if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.Role == "" {
		user.Role = "user" // Default role
	}
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	var id uuid.UUID
	err := r.db.QueryRow(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)
	if err != nil {
		// Check for duplicate email
		if isDuplicateKeyError(err) {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to create user: duplicate email")
			return ErrConflict
		}
		log.Error().Err(err).Msg("Failed to create user")
		return err
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *PostgresUserRepository) GetUserByID(id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.Get(&user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to get user by ID")
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *PostgresUserRepository) GetUserByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.Get(&user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates a user
func (r *PostgresUserRepository) UpdateUser(user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, role = $3, updated_at = $4
		WHERE id = $5
	`

	// Update the updated_at timestamp
	user.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		// Check for duplicate email
		if isDuplicateKeyError(err) {
			log.Error().Err(err).Str("email", user.Email).Msg("Failed to update user: duplicate email")
			return ErrConflict
		}
		log.Error().Err(err).Msg("Failed to update user")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteUser deletes a user
func (r *PostgresUserRepository) DeleteUser(id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to delete user")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// CreateSession creates a new session
func (r *PostgresUserRepository) CreateSession(session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, user_agent, client_ip, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	// Set default values if not provided
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}

	var id uuid.UUID
	err := r.db.QueryRow(
		query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.ClientIP,
		session.ExpiresAt,
		session.CreatedAt,
	).Scan(&id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create session")
		return err
	}

	return nil
}

// GetSessionByID retrieves a session by ID
func (r *PostgresUserRepository) GetSessionByID(id uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, client_ip, expires_at, created_at
		FROM sessions
		WHERE id = $1
	`

	var session domain.Session
	err := r.db.Get(&session, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to get session by ID")
		return nil, err
	}

	return &session, nil
}

// GetSessionByToken retrieves a session by refresh token
func (r *PostgresUserRepository) GetSessionByToken(token string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, client_ip, expires_at, created_at
		FROM sessions
		WHERE refresh_token = $1
	`

	var session domain.Session
	err := r.db.Get(&session, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to get session by token")
		return nil, err
	}

	return &session, nil
}

// DeleteSession deletes a session
func (r *PostgresUserRepository) DeleteSession(id uuid.UUID) error {
	query := `
		DELETE FROM sessions
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		log.Error().Err(err).Str("session_id", id.String()).Msg("Failed to delete session")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteUserSessions deletes all sessions for a user
func (r *PostgresUserRepository) DeleteUserSessions(userID uuid.UUID) error {
	query := `
		DELETE FROM sessions
		WHERE user_id = $1
	`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to delete user sessions")
		return err
	}

	return nil
}

// CreatePasswordReset creates a new password reset
func (r *PostgresUserRepository) CreatePasswordReset(reset *domain.PasswordReset) error {
	query := `
		INSERT INTO password_resets (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	// Set default values if not provided
	if reset.ID == uuid.Nil {
		reset.ID = uuid.New()
	}
	now := time.Now()
	if reset.CreatedAt.IsZero() {
		reset.CreatedAt = now
	}

	var id uuid.UUID
	err := r.db.QueryRow(
		query,
		reset.ID,
		reset.UserID,
		reset.Token,
		reset.ExpiresAt,
		reset.CreatedAt,
	).Scan(&id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create password reset")
		return err
	}

	return nil
}

// GetPasswordResetByToken retrieves a password reset by token
func (r *PostgresUserRepository) GetPasswordResetByToken(token string) (*domain.PasswordReset, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, used_at
		FROM password_resets
		WHERE token = $1
	`

	var reset domain.PasswordReset
	err := r.db.Get(&reset, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to get password reset by token")
		return nil, err
	}

	return &reset, nil
}

// MarkPasswordResetUsed marks a password reset as used
func (r *PostgresUserRepository) MarkPasswordResetUsed(id uuid.UUID) error {
	query := `
		UPDATE password_resets
		SET used_at = $1
		WHERE id = $2
	`

	now := time.Now()
	result, err := r.db.Exec(query, now, id)
	if err != nil {
		log.Error().Err(err).Str("reset_id", id.String()).Msg("Failed to mark password reset as used")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get rows affected")
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteExpiredPasswordResets deletes expired password resets
func (r *PostgresUserRepository) DeleteExpiredPasswordResets() error {
	query := `
		DELETE FROM password_resets
		WHERE expires_at < $1
		OR used_at IS NOT NULL
	`

	now := time.Now()
	_, err := r.db.Exec(query, now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete expired password resets")
		return err
	}

	return nil
}

// Helper functions

// isDuplicateKeyError checks if an error is a duplicate key error
func isDuplicateKeyError(err error) bool {
	// PostgreSQL error code for unique_violation is 23505
	// This is not an elegant solution, but it works
	return err != nil && err.Error() != "" &&
		(err.Error() == "pq: duplicate key value violates unique constraint" ||
			err.Error() == "ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
}
