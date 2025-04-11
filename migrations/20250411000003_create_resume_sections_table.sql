-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Resume sections table stores different sections of a resume with typed JSONB content
CREATE TABLE resume_sections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL,
    type VARCHAR(15) NOT NULL,
    
    -- JSONB columns for different section types
    personal_data JSONB,
    education_data JSONB,
    experience_data JSONB,
    skill_data JSONB,
    project_data JSONB,
    certification_data JSONB,
    
    -- Check constraint to ensure type is one of the allowed values
    CONSTRAINT check_section_type CHECK (type IN ('personal', 'education', 'experience', 'skill', 'project', 'certification')),
    
    -- Check constraint to ensure only the appropriate data column is used based on type
    CONSTRAINT check_personal_data CHECK (type != 'personal' OR personal_data IS NOT NULL),
    CONSTRAINT check_education_data CHECK (type != 'education' OR education_data IS NOT NULL),
    CONSTRAINT check_experience_data CHECK (type != 'experience' OR experience_data IS NOT NULL),
    CONSTRAINT check_skill_data CHECK (type != 'skill' OR skill_data IS NOT NULL),
    CONSTRAINT check_project_data CHECK (type != 'project' OR project_data IS NOT NULL),
    CONSTRAINT check_certification_data CHECK (type != 'certification' OR certification_data IS NOT NULL),
    
    -- Foreign key to resumes table
    CONSTRAINT fk_resume_sections_resume FOREIGN KEY (resume_id)
        REFERENCES resumes(id) ON DELETE CASCADE
);

-- Add index on resume_id for faster lookups of sections for a specific resume
CREATE INDEX idx_resume_sections_resume_id ON resume_sections(resume_id);

-- Add index on type for queries that filter by section type
CREATE INDEX idx_resume_sections_type ON resume_sections(type);

-- Add combined index on resume_id and type for queries that look for specific sections of a resume
CREATE INDEX idx_resume_sections_resume_id_type ON resume_sections(resume_id, type);

-- Create GIN indexes on JSONB columns for efficient querying
CREATE INDEX idx_resume_sections_personal_data ON resume_sections USING GIN (personal_data);
CREATE INDEX idx_resume_sections_education_data ON resume_sections USING GIN (education_data);
CREATE INDEX idx_resume_sections_experience_data ON resume_sections USING GIN (experience_data);
CREATE INDEX idx_resume_sections_skill_data ON resume_sections USING GIN (skill_data);
CREATE INDEX idx_resume_sections_project_data ON resume_sections USING GIN (project_data);
CREATE INDEX idx_resume_sections_certification_data ON resume_sections USING GIN (certification_data);

-- Add comments on table and columns
COMMENT ON TABLE resume_sections IS 'Stores different sections of a resume with typed JSON content';
COMMENT ON COLUMN resume_sections.id IS 'Unique identifier for the resume section';
COMMENT ON COLUMN resume_sections.resume_id IS 'Foreign key to the resume this section belongs to';
COMMENT ON COLUMN resume_sections.type IS 'Type of resume section (personal, education, experience, skill, project, certification)';
COMMENT ON COLUMN resume_sections.personal_data IS 'JSONB data for personal information section: {first_name, last_name, email, phone, address, job_title}';
COMMENT ON COLUMN resume_sections.education_data IS 'JSONB data for education section: {institution, location, degree, start_date, end_date}';
COMMENT ON COLUMN resume_sections.experience_data IS 'JSONB data for experience section: {employer, title, start_date, end_date, location, description}';
COMMENT ON COLUMN resume_sections.skill_data IS 'JSONB data for skill section: {category, name, proficiency}';
COMMENT ON COLUMN resume_sections.project_data IS 'JSONB data for project section: {name, technologies, repo_url, description}';
COMMENT ON COLUMN resume_sections.certification_data IS 'JSONB data for certification section: {name, issuer, issue_date, expiry_date, credential_id}';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX IF EXISTS idx_resume_sections_certification_data;
DROP INDEX IF EXISTS idx_resume_sections_project_data;
DROP INDEX IF EXISTS idx_resume_sections_skill_data;
DROP INDEX IF EXISTS idx_resume_sections_experience_data;
DROP INDEX IF EXISTS idx_resume_sections_education_data;
DROP INDEX IF EXISTS idx_resume_sections_personal_data;
DROP INDEX IF EXISTS idx_resume_sections_resume_id_type;
DROP INDEX IF EXISTS idx_resume_sections_type;
DROP INDEX IF EXISTS idx_resume_sections_resume_id;
DROP TABLE IF EXISTS resume_sections;
