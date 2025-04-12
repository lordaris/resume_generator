-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Resumes table stores the main resume records
CREATE TABLE resumes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Foreign key to users table
    CONSTRAINT fk_resumes_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

-- Add index on user_id for faster lookups of a user's resumes
CREATE INDEX idx_resumes_user_id ON resumes(user_id);

-- Add comments on table and columns
COMMENT ON TABLE resumes IS 'Stores main resume records';
COMMENT ON COLUMN resumes.id IS 'Unique identifier for the resume';
COMMENT ON COLUMN resumes.user_id IS 'Foreign key to the user who owns this resume';
COMMENT ON COLUMN resumes.created_at IS 'Timestamp when the resume was created';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX IF EXISTS idx_resumes_user_id;
DROP TABLE IF EXISTS resumes;
