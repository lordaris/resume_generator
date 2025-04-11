-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Sessions table to store user sessions and refresh tokens
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    refresh_token TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    client_ip TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_sessions_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- Create unique index on refresh_token for faster lookups and to ensure uniqueness
CREATE UNIQUE INDEX idx_sessions_refresh_token ON sessions(refresh_token);

-- Create index on expires_at for cleanup tasks
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Add comments
COMMENT ON TABLE sessions IS 'Stores user sessions and refresh tokens';
COMMENT ON COLUMN sessions.id IS 'Unique identifier for the session';
COMMENT ON COLUMN sessions.user_id IS 'Reference to the user this session belongs to';
COMMENT ON COLUMN sessions.refresh_token IS 'JWT refresh token for the session';
COMMENT ON COLUMN sessions.user_agent IS 'User agent string from the client';
COMMENT ON COLUMN sessions.client_ip IS 'IP address of the client';
COMMENT ON COLUMN sessions.expires_at IS 'Expiration time for the session';
COMMENT ON COLUMN sessions.created_at IS 'Time when the session was created';

-- Password resets table to store password reset requests
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    
    CONSTRAINT fk_password_resets_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

-- Create unique index on token for faster lookups and to ensure uniqueness
CREATE UNIQUE INDEX idx_password_resets_token ON password_resets(token);

-- Create index on user_id for faster lookups
CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);

-- Create index on expires_at for cleanup tasks
CREATE INDEX idx_password_resets_expires_at ON password_resets(expires_at);

-- Add comments
COMMENT ON TABLE password_resets IS 'Stores password reset requests';
COMMENT ON COLUMN password_resets.id IS 'Unique identifier for the password reset request';
COMMENT ON COLUMN password_resets.user_id IS 'Reference to the user this password reset request belongs to';
COMMENT ON COLUMN password_resets.token IS 'JWT token for the password reset request';
COMMENT ON COLUMN password_resets.expires_at IS 'Expiration time for the password reset request';
COMMENT ON COLUMN password_resets.created_at IS 'Time when the password reset request was created';
COMMENT ON COLUMN password_resets.used_at IS 'Time when the password reset request was used (NULL if not used)';

-- Add role column to users table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';
        COMMENT ON COLUMN users.role IS 'User role (e.g., user, admin)';
    END IF;
END$$;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX IF EXISTS idx_password_resets_expires_at;
DROP INDEX IF EXISTS idx_password_resets_user_id;
DROP INDEX IF EXISTS idx_password_resets_token;
DROP TABLE IF EXISTS password_resets;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_refresh_token;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS sessions;

-- Remove the role column from users table (if added by this migration)
ALTER TABLE users DROP COLUMN IF EXISTS role;
