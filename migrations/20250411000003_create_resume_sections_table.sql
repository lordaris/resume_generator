-- +goose Up
-- Replace with separate tables for each section type

-- Create table for personal info
CREATE TABLE personal_info (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    street TEXT,
    city TEXT,
    country TEXT,
    job_title TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_personal_info_resume FOREIGN KEY (resume_id)
        REFERENCES resumes(id) ON DELETE CASCADE
);

-- Create table for education entries
CREATE TABLE education (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL,
    institution TEXT NOT NULL,
    location TEXT,
    degree TEXT NOT NULL,
    field TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_education_resume FOREIGN KEY (resume_id)
        REFERENCES resumes(id) ON DELETE CASCADE
);

-- Create table for experience entries
CREATE TABLE experience (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL,
    employer TEXT NOT NULL,
    job_title TEXT NOT NULL,
    location TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_experience_resume FOREIGN KEY (resume_id)
        REFERENCES resumes(id) ON DELETE CASCADE
);

-- Create table for skills
CREATE TABLE skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    proficiency INT CHECK (proficiency BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_skills_resume FOREIGN KEY (resume_id)
        REFERENCES resumes(id) ON DELETE CASCADE
);

-- Create table for projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    repo_url TEXT,
    demo_url TEXT,
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_projects_resume FOREIGN KEY (resume_id)
        REFERENCES resumes(id) ON DELETE CASCADE
);

-- Create table for project technologies (many-to-many)
CREATE TABLE project_technologies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL,
    technology TEXT NOT NULL,
    CONSTRAINT fk_project_technologies_project FOREIGN KEY (project_id)
        REFERENCES projects(id) ON DELETE CASCADE
);

-- Create table for certifications
CREATE TABLE certifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL,
    name TEXT NOT NULL,
    issuer TEXT NOT NULL,
    issue_date DATE NOT NULL,
    expiry_date DATE,
    credential_id TEXT,
    url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_certifications_resume FOREIGN KEY (resume_id)
        REFERENCES resumes(id) ON DELETE CASCADE
);

-- Create indexes for faster lookups
CREATE INDEX idx_personal_info_resume_id ON personal_info(resume_id);
CREATE INDEX idx_education_resume_id ON education(resume_id);
CREATE INDEX idx_experience_resume_id ON experience(resume_id);
CREATE INDEX idx_skills_resume_id ON skills(resume_id);
CREATE INDEX idx_projects_resume_id ON projects(resume_id);
CREATE INDEX idx_certifications_resume_id ON certifications(resume_id);


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP INDEX IF EXISTS idx_certifications_resume_id;
DROP INDEX IF EXISTS idx_projects_resume_id;
DROP INDEX IF EXISTS idx_skills_resume_id;
DROP INDEX IF EXISTS idx_experience_resume_id;
DROP INDEX IF EXISTS idx_education_resume_id;
DROP INDEX IF EXISTS idx_personal_info_resume_id;

DROP TABLE IF EXISTS project_technologies;
DROP TABLE IF EXISTS certifications;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS skills;
DROP TABLE IF EXISTS experience;
DROP TABLE IF EXISTS education;
DROP TABLE IF EXISTS personal_info;
